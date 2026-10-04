package opennox

import (
	"fmt"
	"strings"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eAIFirstAttackMode(mode string) (kind string, descending, obstacle bool, ok bool) {
	parts := strings.Split(mode, "/")
	if len(parts) != 3 {
		return "", false, false, false
	}
	switch parts[0] {
	case "Spider", "Troll", "Urchin", "NPC":
	default:
		return "", false, false, false
	}
	switch parts[1] {
	case "ascending":
	case "descending":
		descending = true
	default:
		return "", false, false, false
	}
	switch parts[2] {
	case "clear":
	case "off-ray-box":
		obstacle = true
	default:
		return "", false, false, false
	}
	return parts[0], descending, obstacle, true
}

// Only prepares ordinary stock placement. Lane selection uses wall traces and
// conservative prop bounds, independently of the vision function under test.
func e2eAIFirstAttackArena(host *server.Object, radius float32, descending bool) (types.Pointf, types.Pointf, error) {
	delta := types.Ptf(128, 96)
	if descending {
		delta.Y = -delta.Y
	}
	side := types.Ptf(-delta.Y, delta.X).Normalize().Mul(radius)
	for ring := 0; ring <= 368; ring += 23 {
		for y := -ring; y <= ring; y += 23 {
			for x := -ring; x <= ring; x += 23 {
				if ring != 0 && x != -ring && x != ring && y != -ring && y != ring {
					continue
				}
				from := host.PosVec.Add(types.Ptf(float32(x), float32(y)))
				to := from.Add(delta)
				if !e2eWarriorLaneClear(host, from, to) || !e2eWarriorLaneClear(host, from.Add(side), to.Add(side)) || !e2eWarriorLaneClear(host, from.Sub(side), to.Sub(side)) {
					continue
				}
				clear := true
				for _, center := range []types.Pointf{from, to} {
					for dir := 0; dir < 256; dir += 16 {
						cosine, sine := server.SinCosDir(byte(dir))
						if !e2eWarriorLaneClear(host, center, center.Add(types.Ptf(radius*cosine, radius*sine))) {
							clear = false
							break
						}
					}
					if !clear {
						break
					}
				}
				if clear {
					return from, to, nil
				}
			}
		}
	}
	return types.Pointf{}, types.Pointf{}, fmt.Errorf("no independent diagonal AI lane near %v", host.PosVec)
}

type e2eAIFirstAttackFixture struct {
	mode, kind, propKind           string
	descending, obstacle           bool
	host, enemy, prop, weapon      *server.Object
	projectiles                    map[*server.Object]bool
	original                       types.Pointf
	hostHP, hostMax, enemyHP       uint16
	start, incomingFrame, lastLog  uint32
	acquired, fought, missile, hit bool
}

func (f *e2eAIFirstAttackFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.Buffs != 0 || f.host.Poison540 != 0 || f.host.ControllingPlayer() == nil {
		e2eError(fmt.Errorf("AI fixture requires unenchanted live host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	f.enemy = noxServer.NewObjectByTypeID(f.kind)
	if f.enemy == nil || f.enemy.UpdateData == nil || f.enemy.HealthData == nil {
		e2eError(fmt.Errorf("missing stock AI type %s", f.kind))
		return
	}
	from, to, err := e2eAIFirstAttackArena(f.host, max(f.host.Shape.Circle.R, f.enemy.Shape.Circle.R)+12, f.descending)
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(to)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.enemy, nil, from)
	noxServer.ObjectsAddPending()
	f.enemy.SetDir(server.DirFromVec(to.Sub(from)))
	ud := f.enemy.UpdateDataMonster()
	defaultAggression := ud.Aggression
	if f.kind == "NPC" {
		// NPC is a stock configurable unit. Use the normal map/script property
		// for a hostile NPC, without selecting an enemy or scheduling combat.
		asObjectS(f.enemy).SetAggression(0.83)
		item := legacy.Nox_xxx_playerRespawnItem_4EF750(f.enemy, "LongSword", nil, 1, 0)
		if item == nil || item.InvHolder != f.enemy || !f.enemy.HasItem(item) || !item.Flags().Has(object.FlagEquipped) && !asObjectS(f.enemy).Equip(item) {
			e2eError(fmt.Errorf("hostile NPC normal equipment setup failed"))
			return
		}
		f.weapon = item
	}
	if f.obstacle {
		for _, typ := range noxServer.Types.List() {
			if typ.Class().HasAny(object.ClassObstacle|object.ClassImmobile) && typ.Flags().Has(object.FlagShadow) && typ.Shape.Kind == server.ShapeKindBox && typ.Shape.Box.W > 0 && typ.Shape.Box.H > 0 && typ.Shape.Box.W <= 36 && typ.Shape.Box.H <= 36 && typ.Update == nil {
				f.propKind = typ.ID()
				f.prop = noxServer.NewObjectByTypeID(f.propKind)
				break
			}
		}
		if f.prop == nil {
			e2eError(fmt.Errorf("no stock small shadow box for AI visibility"))
			return
		}
		// x=1/8, y=1/8 of the canonical bounds is on the INCORRECT mirrored
		// line for descending sight; actual descending sight passes far away.
		// For ascending sight, use the other diagonal's quarter point.
		bounds := types.RectFromPointsf(from, to)
		point := bounds.Min.Add(types.Ptf(16, 12))
		if !f.descending {
			point.Y = bounds.Max.Y - 12
		}
		noxServer.CreateObjectAt(f.prop, nil, point)
		noxServer.ObjectsAddPending()
		if !f.prop.Flags().Has(object.FlagShadow) || f.prop.Shape.Kind != server.ShapeKindBox {
			e2eError(fmt.Errorf("stock shadow box lost definition"))
			return
		}
	}
	asObjectS(f.host).SetMaxHealth(2000)
	f.enemyHP, f.incomingFrame, f.start = f.enemy.HealthData.Cur, f.enemy.Frame134, noxServer.Frame()
	if !e2eObjectInWorld(f.enemy) || ud.CurrentEnemy != nil || ud.PreferredEnemy != nil || ud.Field282_1 != 0 || f.enemy.Obj130 != nil || ud.StatusFlags.Has(object.MonStatusInjured) || !noxServer.S().IsEnemyTo(f.enemy, f.host) || ud.Aggression <= 0.33 {
		e2eError(fmt.Errorf("AI fixture not unforced hostile: mode=%s HP=%d aggr=%g current=%p preferred=%p seen=%d incoming=%p status=%x", f.mode, f.enemyHP, ud.Aggression, ud.CurrentEnemy, ud.PreferredEnemy, ud.Field282_1, f.enemy.Obj130, ud.StatusFlags))
		return
	}
	e2eLog.Printf("AI FIRST ATTACK PREPARED: mode=%s frame=%d enemy=%p player=%p pos=%v->%v default-aggression=%g aggression=%g stack=%v status=%x sight=%g can-see=%t can-interact=%t prop=%p/%s", f.mode, f.start, f.enemy, f.host, from, to, defaultAggression, ud.Aggression, ud.GetAIStack(), ud.StatusFlags, ud.SightRange, noxServer.S().CanSee(f.enemy, f.host, 0), noxServer.S().CanInteract(f.enemy, f.host, 0), f.prop, f.propKind)
}

// DefaultDamage records its nonnil weapon, not necessarily the attacking unit.
// Match the exact ordinary equipped NPC item or an observed enemy-owned missile.
// Keep projectile identities while they are live, avoiding reads of a deleted
// missile after its damage report reaches the client.
func (f *e2eAIFirstAttackFixture) acceptsDamageSource(source *server.Object) bool {
	if source == nil {
		return false
	}
	if source == f.enemy {
		return true
	}
	if f.kind == "NPC" && source == f.weapon {
		return true
	}
	return f.kind == "Urchin" && f.projectiles[source]
}

func (f *e2eAIFirstAttackFixture) observe() bool {
	ud := f.enemy.UpdateDataMonster()
	if f.enemy.HealthData.Cur != f.enemyHP || f.enemy.Frame134 != f.incomingFrame || f.enemy.Obj130 != nil || ud.StatusFlags.Has(object.MonStatusInjured) {
		e2eError(fmt.Errorf("AI received an incoming hit before first attack: %s", f.mode))
		return true
	}
	if ud.CurrentEnemy == f.host {
		f.acquired = true
	}
	if ud.HasAction(ai.ACTION_FIGHT) {
		f.fought = true
	}
	for _, missile := range noxServer.Objs.AllMissiles() {
		if missile.ObjOwner == f.enemy {
			f.missile = true
			if f.projectiles == nil {
				f.projectiles = make(map[*server.Object]bool)
			}
			f.projectiles[missile] = true
		}
	}
	if noxServer.Frame()-f.lastLog >= 30 {
		f.lastLog = noxServer.Frame()
		e2eLog.Printf("AI FIRST ATTACK TICK: mode=%s elapsed=%d acquired=%t fought=%t missile=%t seen=%d current=%p scan=%d action=%v HP=%d/%d can-interact=%t", f.mode, noxServer.Frame()-f.start, f.acquired, f.fought, f.missile, ud.Field282_1, ud.CurrentEnemy, ud.Field302, ud.AIStackHead().Type(), f.host.HealthData.Cur, f.enemy.HealthData.Cur, noxServer.S().CanInteract(f.enemy, f.host, 0))
	}
	if f.host.HealthData.Cur >= 2000 {
		return false
	}
	if !f.acquired || !f.fought || !f.acceptsDamageSource(f.host.Obj130) || f.kind == "Urchin" && !f.missile {
		e2eError(fmt.Errorf("unforced AI damage lacks acquisition/combat/source: mode=%s acquired=%t fought=%t missile=%t source=%p/%p", f.mode, f.acquired, f.fought, f.missile, f.host.Obj130, f.enemy))
		return true
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.enemy)))
	delta, ok := legacy.HealthChangeForDrawable(uint32(noxServer.GetUnitNetCode(f.host)))
	if dr == nil || !ok || delta >= 0 {
		return false
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	f.hit = true
	e2eLog.Printf("AI FIRST ATTACK PASS: mode=%s elapsed=%d actual-HP=2000->%d client-delta=%d acquired=%t fought=%t missile=%t incoming-hit=none damage-object=%p enemy=%p equipped-weapon=%p observed-projectile=%t drawable=%p path=%s", f.mode, noxServer.Frame()-f.start, f.host.HealthData.Cur, delta, f.acquired, f.fought, f.missile, f.host.Obj130, f.enemy, f.weapon, f.projectiles[f.host.Obj130], dr, path)
	return true
}

func (f *e2eAIFirstAttackFixture) cleanup() {
	if !f.hit {
		e2eError(fmt.Errorf("first attack missing: %s", f.mode))
		return
	}
	noxServer.DelayedDelete(f.enemy)
	if f.prop != nil {
		noxServer.DelayedDelete(f.prop)
	}
	// A stock Spider can poison on its actual first hit. Cure only after the
	// damage/client assertions, so its later DOT cannot count as the next AI's hit.
	noxServer.Server.RemovePoison4EE9D0(f.host)
	if f.host.Poison540 != 0 || f.host.HealthData.Field16 != 0 {
		e2eError(fmt.Errorf("AI first-attack ordinary poison cleanup failed: %s", f.mode))
		return
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	asObjectS(f.host).SetMaxHealth(int(f.hostMax))
	asObjectS(f.host).SetHealth(int(f.hostHP))
}

func (sc *e2eScenario) CheckAIFirstAttack(mode, name string) {
	kind, descending, obstacle, ok := e2eAIFirstAttackMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid AI first-attack mode %q", mode))
		return
	}
	f := &e2eAIFirstAttackFixture{mode: mode, kind: kind, descending: descending, obstacle: obstacle}
	sc.addWhen(0, name+" prepare untouched hostile AI", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		return nox_client_isConnected() && unit != nil && noxClient.ClientPlayerUnit() != nil && unit.Buffs == 0 && unit.Poison540 == 0
	}, f.prepare)
	sc.addWhen(1, name+" acquire and attack without being hit", 600, f.observe, f.cleanup)
	sc.Wait(12, name+" ordinary deletion cleanup")
}
