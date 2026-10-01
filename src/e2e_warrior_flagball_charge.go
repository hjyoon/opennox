package opennox

import (
	"fmt"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// This uses the same normal join/player initialization and A-key collision
// fixture as the non-objective test. Ball pickup/drop, owner links, damage,
// force, ability lifetime and HUD values are never injected.
func (sc *e2eScenario) CheckWarriorFlagBallCharge(name string) {
	for attempt := 1; attempt <= 2; attempt++ {
		f := &e2eWarriorChargeCollision{e2eWarriorAbilityFixture: &e2eWarriorAbilityFixture{ability: server.AbilityBerserk}, kind: "player"}
		var ball *server.Object
		var droppedFrame uint32
		var observedFrame uint32
		label := fmt.Sprintf("%s attempt %d", name, attempt)
		sc.addWhen(0, label+" prepare", 1200, func() bool {
			unit := noxServer.Players.HostUnit()
			return unit != nil && unit.HealthData != nil && unit.UpdateData != nil &&
				!unit.HasEnchant(server.ENCHANT_INVULNERABLE) && !unit.HasEnchant(server.ENCHANT_HELD) &&
				noxServer.Abils.GetCooldownForUnit(unit, server.AbilityBerserk) == 0
		}, func() {
			f.prepare()
			if len(f.targets) != 1 {
				return
			}
			target := f.targets[0]
			var team *server.Team
			for it := noxServer.Teams.First(); it != nil; it = noxServer.Teams.Next(it) {
				if it.ID() != f.unit.TeamVal.ID {
					team = it
					break
				}
			}
			var err error
			ball, err = e2eFindObjective("GameBall")
			if err != nil || team == nil || ball.ObjOwner != nil {
				e2eError(fmt.Errorf("FlagBall fixture: team=%p ball=%p err=%v", team, ball, err))
				return
			}
			legacy.Nox_xxx_createAtImpl_4191D0(team.ID(), target.TeamPtr(), 1, int(target.NetCode), 0)
			if !noxServer.Teams.ContainsObject(target.TeamPtr(), team.ID()) || !noxServer.IsEnemyTo(target, f.unit) {
				e2eError(fmt.Errorf("FlagBall target is not linked to the enemy team"))
				return
			}
			asObjectS(target).SetPos(ball.PosVec)
			e2eLog.Printf("FLAGBALL CHARGE APPROACH: attempt=%d ball=%p target=%p team=%d frame=%d", attempt, ball, target, team.ID(), noxServer.Frame())
		})
		sc.addWhen(1, label+" stock ball pickup", 180, func() bool {
			return ball != nil && len(f.targets) == 1 && ball.ObjOwner == f.targets[0] &&
				(*server.GameBallUpdateData4EA800)(ball.UpdateData).Carrier == f.targets[0]
		}, func() {
			target := f.targets[0]
			data := (*server.GameBallUpdateData4EA800)(ball.UpdateData)
			if !ball.Flags().Has(object.FlagNoCollide) || !noxServer.Teams.ContainsObject(ball.TeamPtr(), target.TeamVal.ID) ||
				data.TeamID != uint32(target.TeamVal.ID) || !playerDamageOwnsBallE2E(target, ball) {
				e2eError(fmt.Errorf("FlagBall pickup failed owner/team/flags/carrier validation"))
				return
			}
			// The FlagBall team spawn is beside diagonal walls. Select the
			// open court around the stock ball instead of that spawn lane.
			origin, direction, err := e2eWarriorAbilityArena(ball.PosVec, target.Shape.Circle.R+4, func(from, to types.Pointf) bool {
				return e2eWarriorLaneClear(target, from, to)
			})
			if err != nil {
				e2eError(err)
				return
			}
			f.origin, f.targetOrigin = origin, origin.Add(direction.Mul(112))
			asObjectS(f.unit).SetPos(f.origin)
			f.unit.VelVec, f.unit.ForceVec, f.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
			asObjectS(target).SetPos(f.targetOrigin)
			target.VelVec, target.ForceVec, target.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
			e2eLog.Printf("FLAGBALL CHARGE CARRIED: attempt=%d ball=%p owner=%p carrier=%p team=%d HP=%d frame=%d invulnerable=%t host-pos=%v target-pos=%v", attempt, ball, ball.ObjOwner, data.Carrier, ball.TeamVal.ID, target.HealthData.Cur, noxServer.Frame(), target.HasEnchant(server.ENCHANT_INVULNERABLE), f.unit.PosVec, target.PosVec)
		})
		// Wait out normal spawn protection; never remove it or inject expiry.
		sc.addWhen(0, label+" wait for spawn protection", 1200, func() bool {
			return len(f.targets) == 1 && !f.targets[0].HasEnchant(server.ENCHANT_INVULNERABLE)
		}, func() {
			e2eLog.Printf("FLAGBALL CHARGE TARGET READY: attempt=%d frame=%d host-pos=%v target-pos=%v", attempt, noxServer.Frame(), f.unit.PosVec, f.targets[0].PosVec)
		})
		sc.Wait(4, label+" synchronize fixture")
		sc.add(0, label+" aim by mouse input", f.aim)
		sc.Wait(3, label+" synchronize aim")
		sc.Key(keybind.KeyA, label+" charge by actual keyboard input")
		sc.addWhen(0, label+" server activation", 120, func() bool { return f.record() != nil }, func() {
			f.exec, f.startFrame = f.record(), noxServer.Frame()
			f.deadline = f.exec.Frame
			f.observeHUD()
			e2eLog.Printf("FLAGBALL CHARGE STARTED: attempt=%d frame=%d cooldown=%d", attempt, f.startFrame, noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk))
		})
		sc.addWhen(0, label+" actual damage and drop", 180, func() bool {
			if frame := noxServer.Frame(); observedFrame == 0 || frame-observedFrame >= 30 {
				observedFrame = frame
				target := f.targets[0]
				e2eLog.Printf("FLAGBALL CHARGE OBSERVE: attempt=%d frame=%d state=%d host-HP=%d target-HP=%d owner=%p host-pos=%v target-pos=%v target-flags=%#x target-invulnerable=%t", attempt, frame, f.unit.UpdateDataPlayer().State, f.unit.HealthData.Cur, target.HealthData.Cur, ball.ObjOwner, f.unit.PosVec, target.PosVec, uint32(target.ObjFlags), target.HasEnchant(server.ENCHANT_INVULNERABLE))
			}
			if !f.collided() || ball.ObjOwner != nil {
				return false
			}
			data := (*server.GameBallUpdateData4EA800)(ball.UpdateData)
			return data.Carrier == f.targets[0] && data.TeamID == uint32(f.targets[0].TeamVal.ID) &&
				data.CarrierFrame >= f.startFrame && ball.TeamVal.ID == f.unit.TeamVal.ID && !ball.Flags().Has(object.FlagNoCollide)
		}, func() {
			target := f.targets[0]
			data := (*server.GameBallUpdateData4EA800)(ball.UpdateData)
			if playerDamageOwnsBallE2E(target, ball) || !noxServer.Teams.ContainsObject(ball.TeamPtr(), f.unit.TeamVal.ID) {
				e2eError(fmt.Errorf("dropped FlagBall retained owned-list/team link"))
				return
			}
			droppedFrame = data.CarrierFrame
			e2eLog.Printf("FLAGBALL CHARGE DROPPED: attempt=%d HP=%d->%d ball=%p owner=%p carrier=%p carrier-team=%d ball-team=%d carrier-frame=%d frame=%d force=%v velocity=%v", attempt, f.health, target.HealthData.Cur, ball, ball.ObjOwner, data.Carrier, data.TeamID, ball.TeamVal.ID, data.CarrierFrame, noxServer.Frame(), ball.ForceVec, ball.VelVec)
			// Clear the collision lane without changing the ball. The stock
			// 45-frame previous-carrier gate and normal physics remain active.
			asObjectS(target).SetPos(f.origin.Add(types.Ptf(0, 200)))
		})
		sc.Screen(label + " dropped ball")
		sc.add(0, label+" capture cooldown", func() {
			f.retryFrame, f.retryCooldown = noxServer.Frame(), noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk)
		})
		sc.Key(keybind.KeyA, label+" retry during cooldown")
		sc.Wait(5, label+" process retry")
		sc.add(0, label+" verify retry rejected", func() {
			want := f.retryCooldown - int(noxServer.Frame()-f.retryFrame)
			if got := noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk); want <= 0 || got != want || f.record() != nil && f.record().Frame != f.deadline {
				e2eError(fmt.Errorf("FlagBall charge retry restarted cooldown"))
				return
			}
			e2eLog.Printf("FLAGBALL CHARGE RETRY REJECTED: attempt=%d", attempt)
		})
		sc.addWhen(0, label+" wait for effect end", 1200, f.ended, func() { e2eLog.Printf("FLAGBALL CHARGE ENDED: attempt=%d frame=%d", attempt, noxServer.Frame()) })
		sc.addWhen(0, label+" wait for ready", 1200, func() bool {
			return noxServer.Abils.GetCooldownForUnit(f.unit, server.AbilityBerserk) == 0 && e2eReadAbilityHUD(server.AbilityBerserk).Ready == 1
		}, func() {
			if droppedFrame == 0 {
				e2eError(fmt.Errorf("FlagBall drop not observed"))
				return
			}
			noxServer.PlayerDisconnect(f.targetPlayer, 4)
			e2eLog.Printf("FLAGBALL CHARGE READY: attempt=%d frame=%d", attempt, noxServer.Frame())
		})
		sc.Wait(3, label+" retire fixture")
		// Normal reset creates a fresh stock ball for the second pickup. It
		// does not reset the ability, damage handler, or any team membership.
		if attempt == 1 {
			sc.ResetFlagball(label + " reset stock ball")
			sc.Wait(5, label+" wait for ball replacement")
		}
	}
}

func playerDamageOwnsBallE2E(owner, ball *server.Object) bool {
	for it := owner.Field129; it != nil; it = it.Field128 {
		if it == ball {
			return true
		}
	}
	return false
}
