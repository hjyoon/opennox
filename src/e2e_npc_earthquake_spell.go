package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
)

// Keep the earlier Wolf/DefaultDamage world unchanged. This separate setup
// exercises the stock NPC's PlayerDamage callback, which rejects an unported
// caster IMPACT even when the earlier DefaultDamage world passes.
func (f *e2eEarthquakeFixture) prepareNPCTarget() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Buffs != 0 || f.host.Poison540 != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("NPC Earthquake requires a live unenchanted Wizard host"))
		return
	}
	e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.GetMousePos()})
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+24,
		func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.caster, f.target = f.host, noxServer.NewObjectByTypeID("NPC")
	if f.target == nil {
		e2eError(fmt.Errorf("NPC Earthquake requires the stock NPC type"))
		return
	}
	noxServer.CreateObjectAt(f.target, nil, origin.Add(direction.Mul(64)))
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.target) || f.target.HealthData == nil || f.target.UpdateData == nil || f.target.Damage == nil ||
		f.target.Damage != f.target.ObjectTypeC().Damage || f.target.Damage != f.host.Damage ||
		!f.target.Class().Has(object.ClassMonster) || uint32(f.target.SubClass())&0x10 == 0 {
		e2eError(fmt.Errorf("NPC Earthquake requires the initialized stock PlayerDamage callback"))
		return
	}
	// Durable HP, placement and WAIT are setup only. Keep the stock damage
	// callback, armor/carry, enemy query, hit attribution and client packets.
	asObjectS(f.target).SetMaxHealth(2000)
	f.target.UpdateDataMonster().SetAggression(0)
	f.target.ClearActionStack()
	f.target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	if !noxServer.S().IsEnemyTo(f.target, f.caster.FindOwnerChainPlayer()) || f.target.Buffs != 0 ||
		f.target.Flags().Has(object.FlagAirborne) || noxClient.Viewport().Jiggle12 != 0 {
		e2eError(fmt.Errorf("NPC Earthquake requires a grounded unenchanted enemy and settled viewport"))
		return
	}
	if radius := float32(noxServer.Balance.Float("EarthquakeRange")); radius <= 64 {
		e2eError(fmt.Errorf("stock EarthquakeRange does not cover the NPC fixture"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.caster), unsafe.Pointer(f.target),
			f.caster.UpdateData, f.target.UpdateData, unsafe.Pointer(f.target.HealthData)} {
			if uintptr(pointer) <= math.MaxUint32 {
				e2eError(fmt.Errorf("NPC Earthquake fixture pointer below 4 GiB: %p", pointer))
				return
			}
		}
	}
	update := f.target.UpdateDataMonster()
	e2eLog.Printf("EARTHQUAKE NPC TARGET: level=%d caster=%p target=%p update=%p health=%p stock-type=NPC callback=%p host-callback=%p armor=%g carry=%g",
		f.level, f.caster, f.target, update, f.target.HealthData, f.target.Damage, f.host.Damage,
		math.Float32frombits(update.Field518), math.Float32frombits(update.Field1))
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.observeTick)
}

// Five ordinary script casts must reach real NPC HP and exactly the matching
// remote client damage number. Shared observers also require the real quake
// packet to arrive and settle; no autonomous casting or mana cost is simulated.
func (sc *e2eScenario) CheckNPCEarthquakeSpell(level int, name string) {
	if level < 1 || level > 5 {
		e2eError(fmt.Errorf("invalid NPC Earthquake level %d", level))
		return
	}
	f := &e2eEarthquakeFixture{level: level, mode: "player-script-npc"}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepareNPCTarget)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual damage and client replay", 300, func() bool {
		if !f.clientHit() {
			return false
		}
		drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
		delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32)
		if !ok || int32(delta) != int32(f.target.HealthData.Cur)-int32(f.health) {
			e2eError(fmt.Errorf("NPC Earthquake client damage differs from actual HP loss: %d/%d->%d",
				delta, f.health, f.target.HealthData.Cur))
		}
		return true
	}, func() {
		drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
		delta, _ := legacy.HealthChangeForDrawable(drawable.NetCode32)
		e2eLog.Printf("EARTHQUAKE NPC CLIENT DAMAGE: level=%d drawable=%p delta=%d", level, drawable, delta)
	})
	sc.CaptureMagicFrame(name + " actual damage frame")
	sc.addWhen(1, name+" real quake packet and natural settling", 300, f.complete, func() {})
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
