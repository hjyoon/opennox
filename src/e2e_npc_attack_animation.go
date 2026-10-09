package opennox

import (
	"fmt"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
)

func e2eNPCAttackAnimationIndex(item string) (int, bool) {
	switch item {
	case "LongSword":
		return 28, true
	case "StaffWooden":
		return 29, true
	case "WarHammer":
		return 39, true
	}
	return 0, false
}

type e2eNPCAttackCycle struct {
	start                  uint32
	frames, clientFrames   uint64
	hits                   int
	completed, clientFinal bool
}

// Observe ordinary AI attacks and the frame byte received by NPCDraw. No
// attack action, target, stored frame, damage, packet or pixels are supplied.
func (sc *e2eScenario) CheckNPCAttackAnimation(item, name string) {
	animation, ok := e2eNPCAttackAnimationIndex(item)
	if !ok {
		e2eError(fmt.Errorf("unsupported NPC attack-animation item %q", item))
		return
	}
	f := &e2eAIFirstAttackFixture{mode: "NPC/" + item + "/animation", kind: "NPC"}
	var cycles []*e2eNPCAttackCycle
	var frameCount, duration int
	var oldWord uint32
	var previousHP uint16
	var previousDamageFrame uint32
	var active bool
	var idleInput e2eLockIdleInput
	sc.addWhen(0, name+" prepare ordinary NPC weapon", 1200, func() bool {
		idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
		host := noxServer.Players.HostUnit()
		return nox_client_isConnected() && noxClient.ClientPlayerUnit() != nil && host != nil && host.Buffs == 0 && host.Poison540 == 0
	}, func() {
		f.prepare()
		if item != "LongSword" {
			if !legacy.Nox_xxx_playerTryDequip_4F2FB0(f.enemy, f.weapon) {
				e2eError(fmt.Errorf("ordinary NPC dequip failed"))
				return
			}
			noxServer.DelayedDelete(f.weapon)
			f.weapon = legacy.Nox_xxx_playerRespawnItem_4EF750(f.enemy, item, nil, 1, 0)
			if f.weapon == nil || f.weapon.InvHolder != f.enemy || !f.weapon.Flags().Has(object.FlagEquipped) && !asObjectS(f.enemy).Equip(f.weapon) {
				e2eError(fmt.Errorf("ordinary NPC equip failed: %s", item))
				return
			}
		}
		frameCount, duration = noxServer.PlayerAnimFrames(animation)
		dd := (*client.PlayerDrawData)(noxClient.ClientPlayerUnit().DrawData)
		if sub_4FA280(f.enemy.UpdateDataMonster().WeaponEquipFlags&0xfffffffc) != animation || frameCount <= 1 || frameCount > 63 || duration < 0 || dd == nil || dd.Anim[animation].Base.Kind != client.AnimSlave || int(dd.Anim[animation].Base.Cnt40) != frameCount || int(dd.Anim[animation].Base.Val42) != duration {
			e2eError(fmt.Errorf("stock NPC animation data mismatch: %s count=%d duration=%d", item, frameCount, duration))
			return
		}
		oldWord = f.enemy.UpdateDataMonster().Field481
		previousHP, previousDamageFrame = f.host.HealthData.Cur, f.host.Frame134
		active = true
		noxServer.TickHook(func() {
			if !active {
				return
			}
			update := f.enemy.UpdateDataMonster()
			attacking := update.AIStackHead().Type() == ai.ACTION_MELEE_ATTACK && update.AIStackHead().Field5 != 0
			var cycle *e2eNPCAttackCycle
			if len(cycles) != 0 {
				cycle = cycles[len(cycles)-1]
			}
			if attacking && (cycle == nil || cycle.start != f.enemy.Field34) {
				if cycle != nil && !cycle.completed {
					e2eError(fmt.Errorf("NPC %s started its next attack before completion: %+v", item, *cycle))
					return
				}
				cycle = &e2eNPCAttackCycle{start: f.enemy.Field34}
				cycles = append(cycles, cycle)
			}
			if cycle == nil || cycle.completed {
				return
			}
			frame := noxServer.Frame() - 1
			elapsed := frame - cycle.start
			if update.Field481 != oldWord {
				e2eError(fmt.Errorf("NPC %s animation wrote unrelated DWORD +1924", item))
				return
			}
			if attacking {
				want := min(int(elapsed)/(duration+1), frameCount-1)
				if int(update.Field120_1) != want {
					e2eError(fmt.Errorf("NPC %s frame=%d want=%d elapsed=%d", item, update.Field120_1, want, elapsed))
					return
				}
				cycle.frames |= uint64(1) << update.Field120_1
			}
			if hp := f.host.HealthData.Cur; hp < previousHP && f.host.Frame134 != previousDamageFrame && object.DamageType(f.host.Field131) != object.DamagePoison {
				if !f.acceptsDamageSource(f.host.Obj130) {
					e2eError(fmt.Errorf("NPC animation hit has unrelated source %p", f.host.Obj130))
					return
				}
				cycle.hits++
				e2eLog.Printf("NPC ANIMATION ACTUAL HIT: item=%s cycle=%d HP=%d->%d observe-frame=%d hit-frame=%d start=%d server-anim=%d", item, len(cycles), previousHP, hp, frame, f.host.Frame134, cycle.start, update.Field120_1)
				if f.host.Frame134-cycle.start != uint32((frameCount/2)*(duration+1)) || cycle.hits != 1 {
					e2eError(fmt.Errorf("NPC %s hit elapsed=%d midpoint=%d observed-anim=%d hits=%d", item, f.host.Frame134-cycle.start, (frameCount/2)*(duration+1), update.Field120_1, cycle.hits))
					return
				}
			}
			previousHP, previousDamageFrame = f.host.HealthData.Cur, f.host.Frame134
			if elapsed >= uint32(frameCount*(duration+1)) {
				// The real action pop resets +481/+483 after the terminal
				// frame was shown; do not require that completed frame to
				// survive the original MonsterActionReset transition.
				if elapsed != uint32(frameCount*(duration+1)) || attacking || update.Field120_1 != 0 || update.Field120_3 != 0 || cycle.frames != (uint64(1)<<frameCount)-1 || cycle.hits != 1 {
					e2eError(fmt.Errorf("NPC %s incomplete cycle: elapsed=%d attacking=%t frames=%x hits=%d", item, elapsed, attacking, cycle.frames, cycle.hits))
					return
				}
				cycle.completed = true
				e2eLog.Printf("NPC ATTACK CYCLE COMPLETE: item=%s cycle=%d start=%d terminal-frame=%d reset-frame=%d frames=%x hits=%d HP=%d", item, len(cycles), cycle.start, frameCount-1, update.Field120_1, cycle.frames, cycle.hits, f.host.HealthData.Cur)
			}
		})
		e2eLog.Printf("NPC ANIMATION PREPARED: item=%s animation=%d frame-count=%d duration=%d weapon=%p equipment=%x", item, animation, frameCount, duration, f.weapon, f.enemy.UpdateDataMonster().WeaponEquipFlags)
	})
	sc.addWhen(1, name+" three full server and client attack cycles", 1200, func() bool {
		idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
		if f.enemy.HealthData.Cur != f.enemyHP || f.enemy.Frame134 != f.incomingFrame || f.enemy.Obj130 != nil || f.host.HealthData.Cur == 0 {
			e2eError(fmt.Errorf("NPC attack fixture received unrelated damage"))
			return true
		}
		if len(cycles) == 0 {
			return false
		}
		cycle := cycles[len(cycles)-1]
		dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.enemy)))
		if dr != nil && dr.AnimInd >= 1 && dr.AnimInd <= 4 {
			if int(dr.AnimFrameSlave) >= frameCount {
				e2eError(fmt.Errorf("NPC client attack frame out of range: %d", dr.AnimFrameSlave))
				return true
			}
			cycle.clientFrames |= uint64(1) << dr.AnimFrameSlave
			cycle.clientFinal = cycle.clientFinal || int(dr.AnimFrameSlave) == frameCount-1
		}
		if noxServer.Frame()-f.lastLog >= 30 {
			f.lastLog = noxServer.Frame()
			e2eLog.Printf("NPC ANIMATION TICK: item=%s cycles=%d elapsed=%d stack=%v server=%d client=%p frames=%x/%x complete=%t client-final=%t", item, len(cycles), noxServer.Frame()-f.start, f.enemy.UpdateDataMonster().GetAIStack(), f.enemy.UpdateDataMonster().Field120_1, dr, cycle.frames, cycle.clientFrames, cycle.completed, cycle.clientFinal)
		}
		if len(cycles) < 3 || !cycles[2].completed || !cycles[2].clientFinal {
			return false
		}
		for _, cycle := range cycles[:3] {
			if !cycle.completed || !cycle.clientFinal || cycle.clientFrames&1 == 0 || cycle.clientFrames&(uint64(1)<<(frameCount/2)) == 0 {
				e2eError(fmt.Errorf("NPC %s client did not receive initial/midpoint/final frame: %+v", item, *cycle))
				return true
			}
		}
		path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
		if err != nil {
			e2eError(err)
			return true
		}
		f.hit = true
		e2eLog.Printf("NPC ATTACK ANIMATION PASS: item=%s full-cycles=3 frame-count=%d real-midpoint-hits=3 client-initial/midpoint/final=received HP=2000->%d path=%s", item, frameCount, f.host.HealthData.Cur, path)
		return true
	}, func() { active = false; f.cleanup() })
	sc.Wait(12, name+" ordinary deletion cleanup")
}
