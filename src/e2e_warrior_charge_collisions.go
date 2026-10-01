package opennox

import (
	"fmt"
	"image"
	"math"
	"sort"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2eWarriorChargeCollision struct {
	*e2eWarriorAbilityFixture
	kind           string
	wall           *server.Wall
	targetPlayer   *server.Player
	expectedDamage uint16
}

// Select an existing stock wall, not an injected CollisionWall or callback.
// The starting circle and approach are clear; the aiming ray crosses the wall.
func e2eWarriorChargeWallLane(unit *server.Object) (types.Pointf, types.Pointf, *server.Wall, error) {
	walls := noxServer.Walls.All()
	original := unit.PosVec
	sort.SliceStable(walls, func(i, j int) bool {
		return walls[i].Pos().Sub(original).Len() < walls[j].Pos().Sub(original).Len()
	})
	radius := unit.Shape.Circle.R + 4
	for _, wall := range walls {
		def := noxServer.Walls.DefByInd(int(wall.Tile1))
		if def == nil || def.Flags32&5 == 0 || wall.Data != nil || wall.Pos().Sub(original).Len() > 512 {
			continue
		}
		for direction := 0; direction < 256; direction += 32 {
			x, y := server.SinCosDir(byte(direction))
			forward := types.Ptf(x, y)
			start := wall.Pos().Sub(forward.Mul(112))
			if !e2eWarriorLaneClear(unit, start, wall.Pos().Sub(forward.Mul(radius+12))) ||
				noxServer.MapTraceRay(start, wall.Pos().Add(forward.Mul(32)), server.MapTraceFlag1) {
				continue
			}
			clear := true
			for around := 0; around < 256; around++ {
				cx, cy := server.SinCosDir(byte(around))
				if !e2eWarriorLaneClear(unit, start, start.Add(types.Ptf(radius*cx, radius*cy))) {
					clear = false
					break
				}
			}
			if clear {
				return start, wall.Pos().Add(forward.Mul(32)), wall, nil
			}
		}
	}
	return types.Pointf{}, types.Pointf{}, nil, fmt.Errorf("no qualifying stock wall lane near %v", original)
}

func (f *e2eWarriorChargeCollision) prepare() {
	f.unit = noxServer.Players.HostUnit()
	unit := f.unit
	if unit == nil || unit.UpdateData == nil || unit.HealthData == nil || unit.ControllingPlayer() == nil ||
		noxServer.Abils.GetCooldownForUnit(unit, server.AbilityBerserk) != 0 || f.record() != nil ||
		unit.ControllingPlayer().SpellLvl[server.AbilityBerserk] == 0 {
		e2eError(fmt.Errorf("charge collision fixture requires a ready Warrior"))
		return
	}
	legacy.Nox_xxx_quickBarSetSpell(int(server.AbilityBerserk), 0)
	var pos, aim types.Pointf
	if f.kind == "wall" {
		var err error
		pos, aim, f.wall, err = e2eWarriorChargeWallLane(unit)
		if err != nil {
			e2eError(err)
			return
		}
	} else {
		pos0, direction, err := e2eWarriorAbilityArena(unit.PosVec, unit.Shape.Circle.R+4, func(from, to types.Pointf) bool {
			return e2eWarriorLaneClear(unit, from, to)
		})
		if err != nil {
			e2eError(err)
			return
		}
		pos, aim = pos0, pos0.Add(direction.Mul(112))
		// Create an initialized server-side player through the normal join and
		// observer-exit services. This is not a second remote client's input test.
		index := ntype.PlayerInd(0)
		for ; index < server.HostPlayerIndex; index++ {
			if !noxServer.Players.ByIndRaw(index).IsActive() {
				break
			}
		}
		if index == server.HostPlayerIndex {
			e2eError(fmt.Errorf("no free player fixture slot"))
			return
		}
		opts := PlayerOpts{Info: *unit.ControllingPlayer().Info(), Screen: image.Pt(1024, 768)}
		opts.Info.SetName("ChargeTarget")
		if noxServer.newPlayer(index, &opts) == 0 {
			e2eError(fmt.Errorf("charge target player join failed"))
			return
		}
		f.targetPlayer = noxServer.Players.ByInd(index)
		noxServer.playerLeaveObserver_4E6AA0(f.targetPlayer)
		target := f.targetPlayer.PlayerUnit
		if target == nil || target.HealthData == nil || target.UpdateData == nil ||
			target.Flags().HasAny(object.FlagNoUpdate|object.FlagNoCollide|object.FlagDead|object.FlagDestroyed) ||
			f.targetPlayer.Field3680&1 != 0 || !noxServer.IsEnemyTo(target, unit) {
			e2eError(fmt.Errorf("charge target player is not a live enemy"))
			return
		}
		asObjectS(target).SetMaxHealth(2000)
		asObjectS(target).SetPos(aim)
		target.VelVec, target.ForceVec, target.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		f.targets, f.health = []*server.Object{target}, target.HealthData.Cur
		armor := math.Float32frombits(target.UpdateDataPlayer().Field57)
		carry := math.Float32frombits(target.UpdateDataPlayer().Field21)
		amount := float32((1 - float64(armor)*0.5) * float64(int32(math.RoundToEven(noxServer.Balance.Float("BerserkerDamage")))))
		f.expectedDamage = uint16(max(1, int32(math.RoundToEven(float64(amount+carry)))))
		e2eLog.Printf("CHARGE PLAYER FIXTURE: index=%d unit=%p update=%p player=%p HP=%d armor=%g expected=%d pos=%v", index, target, target.UpdateData, f.targetPlayer, f.health, armor, f.expectedDamage, aim)
	}
	asObjectS(unit).SetPos(pos)
	unit.VelVec, unit.ForceVec, unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.origin, f.targetOrigin, f.playerHealth = unit.PosVec, aim, unit.HealthData.Cur
	if f.wall != nil {
		f.expectedDamage = uint16(max(1, int32(math.RoundToEven(float64(float32(noxServer.Balance.Float("BerserkerPainRatio")*float64(f.playerHealth)))))))
		e2eLog.Printf("CHARGE WALL FIXTURE: wall=%p grid=%v tile=%d flags=%#x HP=%d expected=%d from=%v aim=%v", f.wall, f.wall.GridPos(), f.wall.Tile1,
			noxServer.Walls.DefByInd(int(f.wall.Tile1)).Flags32, f.playerHealth, f.expectedDamage, pos, aim)
	}
}

func (f *e2eWarriorChargeCollision) collided() bool {
	if f.unit == nil {
		return false
	}
	hud := f.observeHUD()
	if hud.Ready != 0 || !f.hudActiveSeen || f.unit.UpdateDataPlayer().State == server.PlayerState1 {
		return false
	}
	if f.kind == "wall" {
		return f.unit.UpdateDataPlayer().CollisionWall == f.wall && f.unit.HasEnchant(server.ENCHANT_HELD) &&
			f.unit.HealthData.Cur == f.playerHealth-f.expectedDamage
	}
	return len(f.targets) == 1 && e2eObjectInWorld(f.targets[0]) && f.targets[0].HealthData.Cur == f.health-f.expectedDamage &&
		f.targets[0].Obj130 == f.unit && f.targets[0].Field131 == uint32(object.DamageCrush) && f.targets[0].Frame134 >= f.startFrame &&
		f.unit.HealthData.Cur == f.playerHealth && !f.unit.HasEnchant(server.ENCHANT_HELD)
}

// Fixture placement never invokes abilities.Do, the collision callback, damage,
// cooldown setters, HUD setters, or buff expiry. Those run through real A input.
func (sc *e2eScenario) CheckWarriorChargeCollision(kind, name string) {
	if kind != "player" && kind != "wall" {
		panic("charge collision must be player or wall")
	}
	for attempt := 1; attempt <= 2; attempt++ {
		f := &e2eWarriorChargeCollision{e2eWarriorAbilityFixture: &e2eWarriorAbilityFixture{ability: server.AbilityBerserk}, kind: kind}
		label := fmt.Sprintf("%s attempt %d", name, attempt)
		sc.addWhen(0, label+" prepare", 1200, func() bool {
			unit := noxServer.Players.HostUnit()
			return unit != nil && unit.HealthData != nil && unit.UpdateData != nil &&
				!unit.HasEnchant(server.ENCHANT_INVULNERABLE) && !unit.HasEnchant(server.ENCHANT_HELD) &&
				noxServer.Abils.GetCooldownForUnit(unit, server.AbilityBerserk) == 0
		}, f.prepare)
		sc.Wait(3, label+" synchronize fixture")
		sc.add(0, label+" aim by mouse input", f.aim)
		sc.Wait(3, label+" synchronize aim")
		sc.Key(keybind.KeyA, label+" charge by actual keyboard input")
		sc.addWhen(0, label+" server activation", 120, func() bool {
			return f.unit != nil && f.record() != nil && noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk) > 0
		}, func() {
			f.exec, f.startFrame = f.record(), noxServer.Frame()
			f.deadline = f.exec.Frame
			f.observeHUD()
			e2eLog.Printf("CHARGE STARTED: kind=%s attempt=%d frame=%d cooldown=%d", kind, attempt, f.startFrame, noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk))
		})
		sc.addWhen(0, label+" actual collision effect", 180, f.collided, func() {
			targetHP := uint16(0)
			if len(f.targets) != 0 {
				targetHP = f.targets[0].HealthData.Cur
			}
			e2eLog.Printf("CHARGE COLLISION: kind=%s attempt=%d frame=%d player-HP=%d->%d target-HP=%d->%d expected=%d held=%t wall=%p pos=%v->%v", kind, attempt, noxServer.Frame(), f.playerHealth, f.unit.HealthData.Cur,
				f.health, targetHP, f.expectedDamage, f.unit.HasEnchant(server.ENCHANT_HELD), f.unit.UpdateDataPlayer().CollisionWall, f.origin, f.unit.PosVec)
		})
		sc.Screen(label + " collision")
		sc.add(0, label+" capture cooldown", func() {
			f.retryFrame, f.retryCooldown = noxServer.Frame(), noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk)
		})
		sc.Key(keybind.KeyA, label+" retry during cooldown")
		sc.Wait(5, label+" process retry")
		sc.add(0, label+" verify retry rejected", func() {
			want := f.retryCooldown - int(noxServer.Frame()-f.retryFrame)
			if got := noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk); want <= 0 || got != want || f.record() != nil && f.record().Frame != f.deadline {
				e2eError(fmt.Errorf("charge %s retry restarted cooldown: %d want %d", kind, got, want))
				return
			}
			e2eLog.Printf("CHARGE RETRY REJECTED: kind=%s attempt=%d", kind, attempt)
		})
		sc.addWhen(0, label+" wait for effect end", 1200, func() bool { return f.ended() && !f.unit.HasEnchant(server.ENCHANT_HELD) }, func() {
			e2eLog.Printf("CHARGE ENDED: kind=%s attempt=%d frame=%d", kind, attempt, noxServer.Frame())
		})
		sc.addWhen(0, label+" wait for ready", 1200, func() bool {
			return noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk) == 0 && e2eReadAbilityHUD(server.AbilityBerserk).Ready == 1
		}, func() {
			if f.targetPlayer != nil {
				noxServer.PlayerDisconnect(f.targetPlayer, 4)
			}
			e2eLog.Printf("CHARGE READY: kind=%s attempt=%d frame=%d", kind, attempt, noxServer.Frame())
		})
		sc.Wait(3, label+" retire fixture")
	}
}
