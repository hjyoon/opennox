package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func e2eMeteorPlayerMode(level int, mode string) bool {
	return mode == "npc-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
}

// Reuse only the read-only Meteor creation/audio observations and ordinary
// script/animated cast request. The Troll fixture's prepare, hit, complete and
// cleanup methods are deliberately not used: the damage target is the host.
type e2eMeteorPlayerFixture struct {
	cast               e2eMeteorFixture
	original           types.Pointf
	hostHP, hostMax    uint16
	armor, carry, next float32
	effective          int32
	hit, verified      bool
}

func (f *e2eMeteorPlayerFixture) prepare() {
	c := &f.cast
	c.host = noxServer.Players.HostUnit()
	if c.host == nil || c.host.ControllingPlayer() == nil || c.host.UpdateData == nil ||
		c.host.HealthData == nil || c.host.HealthData.Cur == 0 || c.host.Buffs != 0 ||
		c.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Meteor player damage requires a live unenchanted host"))
		return
	}
	f.original, f.hostHP, f.hostMax = c.host.PosVec, c.host.HealthData.Cur, c.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, c.host.Shape.Circle.R+24, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(c.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	c.caster = noxServer.NewObjectByTypeID("NPC")
	if c.caster == nil {
		e2eError(fmt.Errorf("Meteor player damage requires the stock NPC"))
		return
	}
	asObjectS(c.host).SetPos(origin)
	c.host.VelVec, c.host.ForceVec, c.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	// An unowned stock NPC is an enemy of the player. Host ownership would
	// turn this into a friendly/self cast, so use the normal neutral creation.
	noxServer.CreateObjectAt(c.caster, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	c.target = c.host
	if !e2eObjectInWorld(c.caster) || c.caster.UpdateData == nil || c.caster.HealthData == nil ||
		c.caster.Damage != c.target.Damage || !c.caster.SubClass().AsMonster().Has(object.MonsterNPC) ||
		!noxServer.S().IsEnemyTo(c.target, c.caster.FindOwnerChainPlayer()) {
		e2eError(fmt.Errorf("Meteor player fixture lacks the stock enemy NPC/PlayerDamage callback"))
		return
	}
	// Durable HP, placement and WAIT are setup only. Keep the stock player's
	// armor, resistance, carry and every actual cast/impact result untouched.
	asObjectS(c.host).SetMaxHealth(2000)
	c.caster.UpdateDataMonster().SetAggression(0)
	c.caster.ClearActionStack()
	c.caster.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	if c.caster.Buffs != 0 || len(c.ownedMeteors()) != 0 {
		e2eError(fmt.Errorf("Meteor player baseline is enchanted or owns a Meteor"))
		return
	}
	for item := c.target.FirstItem(); item != nil; item = item.NextItem() {
		if !item.Flags().Has(object.FlagEquipped) || item.InitData == nil ||
			!item.Class().HasAny(object.ClassArmor|object.ClassWeapon|object.ClassWand) {
			continue
		}
		for _, modifier := range item.InitDataModifier().Modifiers {
			if modifier != nil && (modifier.Defend76.Fnc != nil || modifier.Engage112 != nil) {
				e2eError(fmt.Errorf("Meteor player baseline has an equipped protection modifier"))
				return
			}
		}
	}
	c.actualLevel = c.level
	if c.level == 0 {
		c.actualLevel = int(noxServer.SpellPower4FE7B0(spell.SPELL_METEOR, c.caster))
	}
	if c.actualLevel < 1 || c.actualLevel > 5 {
		e2eError(fmt.Errorf("Meteor NPC power is outside 1..5: %d", c.actualLevel))
		return
	}
	c.damage = int32(math.RoundToEven(float64(float32(noxServer.Balance.FloatInd("MeteorDamage", c.actualLevel-1)))))
	if c.damage <= 0 || c.damage >= 2000 {
		e2eError(fmt.Errorf("Meteor stock damage is outside durable player health: %d", c.damage))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(c.host), unsafe.Pointer(c.caster), c.host.UpdateData, c.caster.UpdateData} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Meteor player native unit/update allocation below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(c.observeSound)
	noxServer.TickHook(f.observeHit)
}

func (f *e2eMeteorPlayerFixture) beginCast() {
	ud := f.cast.target.UpdateDataPlayer()
	f.armor, f.carry = math.Float32frombits(ud.Field57), math.Float32frombits(ud.Field21)
	f.cast.beginCast()
}

func (f *e2eMeteorPlayerFixture) observeHit() {
	c := &f.cast
	if !c.active || f.hit || c.impactAudio == 0 {
		return
	}
	if !e2eObjectInWorld(c.target) || c.target.HealthData == nil || c.target.UpdateData == nil {
		e2eError(fmt.Errorf("Meteor durable player target disappeared"))
		return
	}
	if c.target.HealthData.Cur == c.health {
		return
	}
	f.effective, f.next = e2eMagicMissileDamage(c.effective, f.armor, f.carry)
	ud := c.target.UpdateDataPlayer()
	if c.effective <= 0 || f.effective <= 0 || f.effective >= int32(c.health) ||
		c.target.HealthData.Cur != c.health-uint16(f.effective) || c.target.Obj130 != c.caster ||
		c.target.Field131 != uint32(object.DamageExplosion) || ud.Field76 != 2 ||
		ud.Field75 != uint32(object.DamageExplosion) || ud.Field21 != math.Float32bits(f.next) {
		e2eError(fmt.Errorf("Meteor player damage mismatch: HP=%d->%d expected=%d attribution=%p/%p marker=%d/%d carry=%#x/%#x",
			c.health, c.target.HealthData.Cur, c.health-uint16(f.effective), c.target.Obj130, c.caster,
			ud.Field76, ud.Field75, ud.Field21, math.Float32bits(f.next)))
		return
	}
	f.hit = true
	e2eLog.Printf("METEOR PLAYER HIT: mode=%s requested=%d actual=%d HP=%d->%d raw/radial/armored=%d/%d/%d armor=%g carry=%g->%g elapsed=%d natural-NPC=%t",
		c.mode, c.level, c.actualLevel, c.health, c.target.HealthData.Cur, c.damage, c.effective, f.effective,
		f.armor, f.carry, f.next, noxServer.Frame()-c.frame, c.natural)
}

func (f *e2eMeteorPlayerFixture) complete() bool {
	c := &f.cast
	if !f.hit || c.castAudio != 1 || c.impactAudio != 1 ||
		c.level == 0 && (!c.natural || c.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	if e2eFistInWorld(c.meteor, c.wire, c.scriptID) || len(c.ownedMeteors()) != 0 || noxClient.Objs.ByNetCode(uint16(c.wire)) != nil {
		return false
	}
	meter, ready := e2eClientHUDMeter(0)
	// Normal SetMaxHealth setup does not rewrite the HUD's stock maximum.
	// Require the real network-published current HP, not a synthesized delta.
	if !e2eFireballHUDHit(meter, ready, c.health, f.effective, f.hostMax) {
		return false
	}
	f.verified = true
	e2eLog.Printf("METEOR PLAYER COMPLETE: mode=%s requested=%d actual=%d server/client-HP=%d HUD-max=%d world/owned/drawable=removed cast/impact=1/1 natural-NPC=%t",
		c.mode, c.level, c.actualLevel, meter.Current, meter.Maximum, c.natural)
	return true
}

func (f *e2eMeteorPlayerFixture) cleanup() {
	if !f.verified {
		e2eError(fmt.Errorf("Meteor player cleanup cannot restore health before verified damage/HUD"))
		return
	}
	f.cast.active = false
	noxServer.DelayedDelete(f.cast.caster)
	asObjectS(f.cast.host).SetPos(f.original)
	f.cast.host.VelVec, f.cast.host.ForceVec, f.cast.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	// Restoration is strictly after successful natural damage and HUD proof.
	asObjectS(f.cast.host).SetMaxHealth(int(f.hostMax))
	asObjectS(f.cast.host).SetHealth(int(f.hostHP))
}

// All five NPC script powers plus the normal animated CAST_ON_OBJECT entry
// exercise the real nil-weapon Explosion -> player armor -> DefaultDamage ->
// HP/HUD chain. This reserves a normal action, not autonomous AI spell choice.
func (sc *e2eScenario) CheckMeteorPlayerDamage(level int, mode, name string) {
	if !e2eMeteorPlayerMode(level, mode) {
		e2eError(fmt.Errorf("invalid Meteor player scenario %s/%d", mode, level))
		return
	}
	f := &e2eMeteorPlayerFixture{cast: e2eMeteorFixture{level: level, mode: mode}}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" falling drawable", 180, func() bool {
		c := &f.cast
		return c.meteor != nil && e2eFistInWorld(c.meteor, c.wire, c.scriptID) && c.meteor.ZVal > 100 && c.meteor.ZVal < 230 &&
			noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(c.meteor))) != nil
	}, func() {})
	sc.CaptureMagicFrame(name + " actual falling frame")
	sc.addWhen(1, name+" natural player impact and HUD", 300, f.complete, func() {})
	sc.CaptureMagicFrame(name + " actual player impact frame")
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
