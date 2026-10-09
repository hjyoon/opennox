package opennox

import (
	"fmt"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eNPCIncomingAttackType(kind string) (object.DamageType, bool) {
	switch kind {
	case "GruntAxe":
		return object.DamageBlade, true
	case "StoneGolem":
		return object.DamageCrush, true
	case "Scorpion":
		return object.DamageImpale, true
	case "Ghost":
		return object.DamageDrain, true
	case "Spider":
		return object.DamageBite, true
	case "VileZombie":
		return object.DamageClaw, true
	}
	return 0, false
}

// A normally player-owned NPC is nearer to the stock enemy than the host.
// The enemy must acquire it itself and deliver three distinct non-poison hits.
func (sc *e2eScenario) CheckNPCIncomingAttack(kind, name string) {
	wantType, ok := e2eNPCIncomingAttackType(kind)
	if !ok {
		e2eError(fmt.Errorf("unsupported NPC incoming attack kind %q", kind))
		return
	}
	var host, target, enemy *server.Object
	var original types.Pointf
	var hostHP, hostMax, previousHP, enemyHP uint16
	var start, lastDamageFrame, enemyIncoming, lastLog uint32
	var hits int
	var acquired, fought bool
	var idleInput e2eLockIdleInput
	sc.addWhen(0, name+" prepare player-owned NPC and untouched enemy", 1200, func() bool {
		idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
		u := noxServer.Players.HostUnit()
		return nox_client_isConnected() && noxClient.ClientPlayerUnit() != nil && u != nil && u.Buffs == 0 && u.Poison540 == 0
	}, func() {
		host = noxServer.Players.HostUnit()
		original, hostHP, hostMax = host.PosVec, host.HealthData.Cur, host.HealthData.Max
		target, enemy = noxServer.NewObjectByTypeID("NPC"), noxServer.NewObjectByTypeID(kind)
		if target == nil || enemy == nil || target.HealthData == nil || enemy.HealthData == nil || target.UpdateData == nil || enemy.UpdateData == nil {
			e2eError(fmt.Errorf("stock NPC/enemy missing: %s", kind))
			return
		}
		from, to, err := e2eAIFirstAttackArena(host, max(target.Shape.Circle.R, enemy.Shape.Circle.R)+12, false)
		if err != nil {
			e2eError(err)
			return
		}
		asObjectS(host).SetPos(to)
		host.VelVec, host.ForceVec, host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		noxServer.CreateObjectAt(target, host, from.Add(to.Sub(from).Mul(0.6)))
		noxServer.CreateObjectAt(enemy, nil, from)
		noxServer.ObjectsAddPending()
		asObjectS(target).SetAggression(0)
		asObjectS(target).SetRetreatLevel(0)
		asObjectS(target).SetMaxHealth(2000)
		asObjectS(host).SetMaxHealth(2000)
		enemy.SetDir(server.DirFromVec(target.PosVec.Sub(enemy.PosVec)))
		ud := enemy.UpdateDataMonster()
		if ud.MonsterDef == nil || object.DamageType(ud.MonsterDef.MeleeAttackDamageType124) != wantType {
			e2eError(fmt.Errorf("stock enemy melee definition mismatch: %s want-type=%d", kind, wantType))
			return
		}
		if !e2eObjectInWorld(target) || !e2eObjectInWorld(enemy) || target.ObjOwner != host || !noxServer.S().IsEnemyTo(enemy, target) || ud.CurrentEnemy != nil || ud.PreferredEnemy != nil || ud.Field282_1 != 0 || enemy.Obj130 != nil || ud.StatusFlags.Has(object.MonStatusInjured) || ud.Aggression <= 0.33 {
			e2eError(fmt.Errorf("NPC/enemy setup is not naturally hostile: %s owner=%p host=%p enemy=%t", kind, target.ObjOwner, host, noxServer.S().IsEnemyTo(enemy, target)))
			return
		}
		previousHP, enemyHP = target.HealthData.Cur, enemy.HealthData.Cur
		lastDamageFrame, enemyIncoming, start = target.Frame134, enemy.Frame134, noxServer.Frame()
		e2eLog.Printf("NPC INCOMING PREPARED: kind=%s NPC=%p enemy=%p owner=%p type=%d pos=%v/%v/%v armor=%g", kind, target, enemy, host, wantType, enemy.PosVec, target.PosVec, host.PosVec, math.Float32frombits(target.UpdateDataMonster().Field518))
	})
	sc.addWhen(1, name+" three natural enemy hits on NPC", 1200, func() bool {
		idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
		ud := enemy.UpdateDataMonster()
		if target.HealthData.Cur == 0 || target.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || enemy.HealthData.Cur != enemyHP || enemy.Frame134 != enemyIncoming || enemy.Obj130 != nil {
			e2eError(fmt.Errorf("NPC incoming fixture died or enemy received an injected/retaliatory hit: %s", kind))
			return true
		}
		acquired = acquired || ud.CurrentEnemy == target
		fought = fought || ud.HasAction(ai.ACTION_FIGHT)
		hp := target.HealthData.Cur
		if hp < previousHP && target.Frame134 != lastDamageFrame && object.DamageType(target.Field131) != object.DamagePoison {
			if !acquired || !fought || target.Obj130 != enemy || object.DamageType(target.Field131) != wantType {
				e2eError(fmt.Errorf("NPC %s damage lacks natural acquisition/identity/type: type=%d source=%p acquired=%t fought=%t", kind, target.Field131, target.Obj130, acquired, fought))
				return true
			}
			hits++
			e2eLog.Printf("NPC INCOMING HIT: kind=%s hit=%d frame=%d NPC-HP=%d->%d type=%d marker=%d/%d source=%p", kind, hits, target.Frame134, previousHP, hp, target.Field131, target.UpdateDataMonster().Field547, target.UpdateDataMonster().Field546, target.Obj130)
		}
		previousHP, lastDamageFrame = hp, target.Frame134
		if noxServer.Frame()-lastLog >= 30 {
			lastLog = noxServer.Frame()
			e2eLog.Printf("NPC INCOMING TICK: kind=%s elapsed=%d hits=%d HP=%d current=%p acquired=%t fought=%t enemy-stack=%v NPC-stack=%v positions=%v/%v", kind, noxServer.Frame()-start, hits, hp, ud.CurrentEnemy, acquired, fought, ud.GetAIStack(), target.UpdateDataMonster().GetAIStack(), enemy.PosVec, target.PosVec)
		}
		if hits < 3 {
			return false
		}
		wire := uint16(noxServer.GetUnitNetCode(target))
		delta, received := legacy.HealthChangeForDrawable(uint32(wire))
		if noxClient.Objs.ByNetCode(wire) == nil || !received || delta >= 0 {
			return false
		}
		path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
		if err != nil {
			e2eError(err)
			return true
		}
		e2eLog.Printf("NPC INCOMING ATTACK PASS: kind=%s type=%d hits=%d NPC-HP=2000->%d client-delta=%d acquired=%t fought=%t enemy-incoming=none path=%s", kind, wantType, hits, hp, delta, acquired, fought, path)
		return true
	}, func() {
		noxServer.DelayedDelete(enemy)
		noxServer.DelayedDelete(target)
		asObjectS(host).SetPos(original)
		host.VelVec, host.ForceVec, host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		noxServer.Server.RemovePoison4EE9D0(host)
		asObjectS(host).SetMaxHealth(int(hostMax))
		asObjectS(host).SetHealth(int(hostHP))
	})
	sc.Wait(12, name+" ordinary deletion cleanup")
}
