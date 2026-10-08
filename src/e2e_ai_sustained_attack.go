package opennox

import (
	"fmt"
	"strings"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func e2eAISustainedAttackMode(mode string) (kind string, descending, obstacle, ok bool) {
	if k, d, o, valid := e2eAIFirstAttackMode(mode); valid {
		return k, d, o, true
	}
	parts := strings.Split(mode, "/")
	if len(parts) != 3 {
		return "", false, false, false
	}
	switch parts[0] {
	case "OgreWarlord", "OgreBrute", "Scorpion", "VileZombie", "StoneGolem", "MechanicalGolem", "Wasp", "Ghost":
	default:
		return "", false, false, false
	}
	if parts[1] != "ascending" && parts[1] != "descending" || parts[2] != "clear" && parts[2] != "off-ray-box" {
		return "", false, false, false
	}
	return parts[0], parts[1] == "descending", parts[2] == "off-ray-box", true
}

// Observe distinct, real damage events for several attack cycles. Poison DOT
// is deliberately excluded: one Spider bite must not pass this regression.
func (sc *e2eScenario) CheckAISustainedAttack(mode, name string) {
	kind, descending, obstacle, ok := e2eAISustainedAttackMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid sustained AI mode %q", mode))
		return
	}
	f := &e2eAIFirstAttackFixture{mode: mode, kind: kind, descending: descending, obstacle: obstacle}
	var hits int
	var lastDamageFrame uint32
	var previousHP uint16
	var idleInput e2eLockIdleInput
	sc.addWhen(0, name+" prepare untouched hostile AI", 1200, func() bool {
		// Cursor motion alone preserves ordinary input activity. It neither
		// moves nor attacks: otherwise this long matrix crosses the stock
		// 2700-input-tick idle-observer threshold halfway through the run.
		idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
		unit := noxServer.Players.HostUnit()
		return nox_client_isConnected() && unit != nil && noxClient.ClientPlayerUnit() != nil && unit.Buffs == 0 && unit.Poison540 == 0
	}, func() {
		f.prepare()
		previousHP, lastDamageFrame = f.host.HealthData.Cur, f.host.Frame134
	})
	sc.addWhen(1, name+" three separate attacks without retaliation", 900, func() bool {
		idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
		if noxServer.Players.HostUnit() != f.host || f.host.HealthData.Cur == 0 || f.host.Flags().HasAny(object.FlagDestroyed|object.FlagDead) {
			e2eError(fmt.Errorf("host ceased to be live during sustained AI observation: %s flags=%x HP=%d input-delay=%d", mode, f.host.ObjFlags, f.host.HealthData.Cur, noxClient.Inp.SeqDelay()))
			return true
		}
		update := f.enemy.UpdateDataMonster()
		if f.enemy.HealthData.Cur != f.enemyHP || f.enemy.Frame134 != f.incomingFrame || f.enemy.Obj130 != nil || update.StatusFlags.Has(object.MonStatusInjured) {
			e2eError(fmt.Errorf("AI received an incoming hit during sustained attacks: %s", mode))
			return true
		}
		f.acquired = f.acquired || update.CurrentEnemy == f.host
		f.fought = f.fought || update.HasAction(ai.ACTION_FIGHT)
		for _, missile := range noxServer.Objs.AllMissiles() {
			if missile.ObjOwner == f.enemy {
				f.missile = true
				if f.projectiles == nil {
					f.projectiles = make(map[*server.Object]bool)
				}
				f.projectiles[missile] = true
			}
		}
		hp := f.host.HealthData.Cur
		if hp < previousHP && f.host.Frame134 != lastDamageFrame && object.DamageType(f.host.Field131) != object.DamagePoison {
			if !f.acquired || !f.fought || !f.acceptsDamageSource(f.host.Obj130) {
				e2eError(fmt.Errorf("sustained AI damage lacks acquisition/combat/source: %s source=%p", mode, f.host.Obj130))
				return true
			}
			hits++
			e2eLog.Printf("AI SUSTAINED HIT: mode=%s hit=%d frame=%d HP=%d->%d type=%d source=%p stack=%v", mode, hits, f.host.Frame134, previousHP, hp, f.host.Field131, f.host.Obj130, update.GetAIStack())
		}
		previousHP, lastDamageFrame = hp, f.host.Frame134
		if noxServer.Frame()-f.lastLog >= 30 {
			f.lastLog = noxServer.Frame()
			e2eLog.Printf("AI SUSTAINED TICK: mode=%s elapsed=%d hits=%d HP=%d enemy=%p stack=%v anim=%d/%d/%d cooldown=%d", mode, noxServer.Frame()-f.start, hits, hp, update.CurrentEnemy, update.GetAIStack(), update.Field120_1, update.Field120_2, update.Field120_3, update.Field128)
		}
		if hits < 3 {
			return false
		}
		if noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.enemy))) == nil {
			return false
		}
		f.hit = true
		e2eLog.Printf("AI SUSTAINED ATTACK PASS: mode=%s hits=%d HP=2000->%d incoming-hit=none elapsed=%d", mode, hits, hp, noxServer.Frame()-f.start)
		return true
	}, f.cleanup)
	sc.Wait(12, name+" ordinary deletion cleanup")
}
