package opennox

import (
	"fmt"
	"math"
	"math/big"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eToxicCloudMode(level int, mode string) bool {
	return mode == "player-script" && level >= 1 && level <= 5 ||
		(mode == "script-pos-pos" || mode == "npc-animated") && level == 0
}

func e2eToxicCloudLifetime(lifetime float32, fps uint32) (int32, bool) {
	if math.IsNaN(float64(lifetime)) || math.IsInf(float64(lifetime), 0) {
		return 0, false
	}
	a := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(float64(lifetime))
	b := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetInt64(int64(int32(fps)))
	product := new(big.Float).SetPrec(53).SetMode(big.ToZero).Mul(a, b)
	value, _ := new(big.Float).SetPrec(24).SetMode(big.ToZero).Set(product).Int(nil)
	if !value.IsInt64() || value.Int64() <= 0 || value.Int64() > 1800 {
		return 0, false
	}
	return int32(value.Int64()), true
}

type e2eToxicCloudFixture struct {
	level                         int
	mode                          string
	host, caster, target, cloud   *server.Object
	original, source, aim         types.Pointf
	wire, frame                   uint32
	scriptID                      int32
	duration, previousDuration    int32
	health                        uint16
	active, natural, hit, removed bool
	castAudio                     int
}

func (f *e2eToxicCloudFixture) ownedClouds() []*server.Object {
	var out []*server.Object
	for obj, remaining := f.caster.Field129, 4096; obj != nil; obj, remaining = obj.Field128, remaining-1 {
		if remaining == 0 {
			e2eError(fmt.Errorf("Toxic Cloud owned list did not terminate"))
			return nil
		}
		if typ := obj.ObjectTypeC(); typ != nil && typ.ID() == "ToxicCloud" {
			out = append(out, obj)
		}
	}
	return out
}

func (f *e2eToxicCloudFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Buffs != 0 || f.host.Poison540 != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Toxic Cloud requires a live unenchanted host"))
		return
	}
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+24, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	f.source = origin
	// The stock Troll has IMMUNE_POISON (0x200). Use an ordinary stock
	// Wolf and verify its live subtype, without changing immunity flags.
	f.target = noxServer.NewObjectByTypeID("Wolf")
	if f.target == nil || f.target.HealthData == nil || f.target.UpdateData == nil ||
		!f.target.Class().Has(object.ClassMonster) || f.target.SubClass().Has(0x200) {
		e2eError(fmt.Errorf("Toxic Cloud requires a stock poison-susceptible Wolf target"))
		return
	}
	fmt.Printf("TOXIC CLOUD TARGET: type=%s class=%#x subclass=%#x poison-immune=false\n", f.target.ObjectTypeC().ID(), uint32(f.target.Class()), uint32(f.target.SubClass()))
	f.caster = f.host
	switch f.mode {
	case "script-pos-pos":
		f.caster = nox_xxx_imagCasterUnit_1569664
		if f.caster == nil {
			e2eError(fmt.Errorf("Toxic Cloud requires the normal ImaginaryCaster"))
			return
		}
	case "npc-animated":
		f.caster = noxServer.NewObjectByTypeID("NPC")
		if f.caster == nil || f.caster.UpdateData == nil {
			e2eError(fmt.Errorf("Toxic Cloud requires the stock NPC caster"))
			return
		}
		// Ordinary ownership makes this NPC allied to the host and hostile to
		// the unowned Wolf. No enemy or damage outcome is supplied.
		noxServer.CreateObjectAt(f.caster, f.host, origin)
		f.caster.UpdateDataMonster().SetAggression(0)
		f.caster.ClearActionStack()
		f.caster.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	}
	asObjectS(f.host).SetPos(origin)
	if f.mode != "player-script" {
		asObjectS(f.host).SetPos(origin.Sub(direction.Mul(80)))
	}
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.target, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	// Placement, target durability and WAIT are setup. Never write a cloud,
	// duration, aim RNG, cast-frame, poison, HP result, packet or drawable.
	asObjectS(f.target).SetMaxHealth(2000)
	f.target.UpdateDataMonster().SetAggression(0)
	f.target.ClearActionStack()
	f.target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	f.aim = f.target.PosVec
	if f.mode != "script-pos-pos" && !noxServer.S().IsEnemyTo(f.target, f.caster.FindOwnerChainPlayer()) {
		e2eError(fmt.Errorf("Toxic Cloud fixture has no enemy target"))
		return
	}
	if len(f.ownedClouds()) != 0 || f.target.Poison540 != 0 {
		e2eError(fmt.Errorf("Toxic Cloud baseline already has cloud/poison"))
		return
	}
	var ok bool
	f.duration, ok = e2eToxicCloudLifetime(float32(noxServer.Balance.Float("ToxicCloudLifetime")), noxServer.TickRate())
	if !ok {
		e2eError(fmt.Errorf("stock ToxicCloudLifetime is outside bounded 1..1800-frame fixture"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.caster), unsafe.Pointer(f.target), f.target.UpdateData, unsafe.Pointer(f.target.HealthData)} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Toxic Cloud native fixture below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.observeTick)
}

func (f *e2eToxicCloudFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Toxic Cloud target was not published"))
		return
	}
	f.health, f.frame, f.active = f.target.HealthData.Cur, noxServer.Frame(), true
	api := noxServer.noxScriptP()
	switch f.mode {
	case "npc-animated":
		f.caster.SetDir(server.DirFromVec(f.target.PosVec.Sub(f.caster.PosVec)))
		// The normal animation and cast-frame gate execute this action.
		f.caster.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_TOXIC_CLOUD), 0, f.target)
		return
	case "script-pos-pos":
		// Same normal position-to-position script API as the crash report;
		// its real ImaginaryCaster and default power are not replaced.
		api.CastSpell(nsp.Spell("SPELL_TOXIC_CLOUD"), e2eFistPosition{f.source}, e2eFistPosition{f.aim})
	case "player-script":
		api.CastSpellLvl(nsp.Spell("SPELL_TOXIC_CLOUD"), f.level, api.toObj(f.caster), e2eFistPosition{f.aim})
	}
	noxServer.ObjectsAddPending()
	owned := f.ownedClouds()
	if len(owned) != 1 {
		e2eError(fmt.Errorf("Toxic Cloud produced %d owned clouds", len(owned)))
		return
	}
	f.observeCreation(owned[0])
}

func (f *e2eToxicCloudFixture) observeCreation(cloud *server.Object) {
	if f.cloud != nil {
		if f.cloud != cloud {
			e2eError(fmt.Errorf("Toxic Cloud created a second object"))
		}
		return
	}
	data := (*server.ToxicCloudUpdateData)(cloud.UpdateData)
	if cloud.ObjectTypeC().ID() != "ToxicCloud" || cloud.ObjOwner != f.caster || data == nil || data.Duration != f.duration {
		e2eError(fmt.Errorf("Toxic Cloud creation mismatch: cloud=%p owner=%p/%p data=%v expected-duration=%d", cloud, cloud.ObjOwner, f.caster, data, f.duration))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(cloud)) <= math.MaxUint32 || uintptr(cloud.UpdateData) <= math.MaxUint32) {
		e2eError(fmt.Errorf("Toxic Cloud object/update below 4 GiB"))
		return
	}
	f.cloud, f.wire, f.scriptID, f.previousDuration = cloud, cloud.NetCode, cloud.ScriptIDVal, data.Duration
	e2eLog.Printf("TOXIC CLOUD CAST: mode=%s level=%d caster=%p cloud=%p update=%p owner=%p duration=%d position=%v", f.mode, f.level, f.caster, cloud, cloud.UpdateData, cloud.ObjOwner, data.Duration, cloud.PosVec)
}

func (f *e2eToxicCloudFixture) observeSound(id sound.ID, kind int, cloud *server.Object, _ types.Pointf) {
	if !f.active || id != sound.SoundToxicCloudCast {
		return
	}
	f.castAudio++
	if kind != 0 || f.castAudio != 1 || cloud == nil {
		e2eError(fmt.Errorf("Toxic Cloud object cast audio mismatch"))
		return
	}
	f.observeCreation(cloud)
	if f.mode == "npc-animated" {
		ud, head := f.caster.UpdateDataMonster(), f.caster.UpdateDataMonster().AIStackHead()
		if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target || head.ArgU32(0) != uint32(spell.SPELL_TOXIC_CLOUD) || ud.Field120_2 != 0 || uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
			e2eError(fmt.Errorf("Toxic Cloud NPC bypassed natural animation/cast-frame"))
			return
		}
		f.natural = true
		e2eLog.Printf("TOXIC CLOUD NPC ANIMATION: animation=%d elapsed=%d target=%p", ud.Field120_1, noxServer.Frame()-f.frame, head.ArgObj(2))
	}
}

func (f *e2eToxicCloudFixture) observeTick() {
	if !f.active || f.cloud == nil {
		return
	}
	if e2eFistInWorld(f.cloud, f.wire, f.scriptID) {
		remaining := (*server.ToxicCloudUpdateData)(f.cloud.UpdateData).Duration
		if remaining < 0 || remaining > f.duration || remaining > f.previousDuration || f.previousDuration-remaining > 1 {
			e2eError(fmt.Errorf("Toxic Cloud duration did not count down naturally: %d->%d", f.previousDuration, remaining))
			return
		}
		f.previousDuration = remaining
	} else if !f.removed {
		// Deferred deletion may remove the zero-duration record before this
		// hook. Never dereference its retained address after world removal.
		if f.previousDuration > 1 {
			e2eError(fmt.Errorf("Toxic Cloud disappeared before duration expiry: %d", f.previousDuration))
			return
		}
		f.removed = true
		e2eLog.Printf("TOXIC CLOUD EXPIRED: mode=%s level=%d last-duration=%d elapsed=%d", f.mode, f.level, f.previousDuration, noxServer.Frame()-f.frame)
	}
	// Source-less periodic poison also lowers HP, but carries no cloud
	// attribution. Require this live cloud so that timer damage alone
	// cannot satisfy the instant world-POISON observation.
	if !f.hit && e2eObjectInWorld(f.target) && f.target.Obj130 == f.cloud && f.target.HealthData.Cur < f.health {
		if f.target.HealthData.Cur == 0 || f.target.Field131 != uint32(object.DamagePoison) {
			e2eError(fmt.Errorf("Toxic Cloud target was damaged by another type"))
			return
		}
		f.hit = true
		e2eLog.Printf("TOXIC CLOUD HIT: mode=%s level=%d cloud-attribution=%p HP=%d->%d poison=%d elapsed=%d", f.mode, f.level, f.target.Obj130, f.health, f.target.HealthData.Cur, f.target.Poison540, noxServer.Frame()-f.frame)
	}
}

func (f *e2eToxicCloudFixture) clientHit() bool {
	if !f.hit {
		return false
	}
	drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if drawable == nil {
		return false
	}
	delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32)
	return ok && delta < 0
}

func (f *e2eToxicCloudFixture) complete() bool {
	if !f.removed || !f.hit || f.castAudio != 1 || f.mode == "npc-animated" && (!f.natural || f.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	if len(f.ownedClouds()) != 0 || noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
		return false
	}
	e2eLog.Printf("TOXIC CLOUD COMPLETE: mode=%s level=%d duration=%d actual-HP=%d->%d audio=1 natural-NPC=%t world/owned/drawable=removed", f.mode, f.level, f.duration, f.health, f.target.HealthData.Cur, f.natural)
	return true
}

func (f *e2eToxicCloudFixture) cleanup() {
	f.active = false
	noxServer.DelayedDelete(f.target)
	if f.mode == "npc-animated" {
		noxServer.DelayedDelete(f.caster)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Real position-to-position/script object casts and an animated NPC action
// pass through stock acceptance, cloud creation/update, damage and expiry.
// This does not simulate player incantation/mana or autonomous spell choice.
func (sc *e2eScenario) CheckToxicCloudSpell(level int, mode, name string) {
	if !e2eToxicCloudMode(level, mode) {
		e2eError(fmt.Errorf("invalid Toxic Cloud mode/level %s/%d", mode, level))
		return
	}
	f := &e2eToxicCloudFixture{level: level, mode: mode}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" published cloud", 180, func() bool {
		return f.cloud != nil && e2eFistInWorld(f.cloud, f.wire, f.scriptID) && noxClient.Objs.ByNetCode(uint16(f.wire)) != nil
	}, func() {})
	sc.CaptureMagicFrame(name + " actual cloud frame")
	sc.addWhen(1, name+" actual damage and client replay", 300, f.clientHit, func() {})
	sc.CaptureMagicFrame(name + " actual damage frame")
	sc.addWhen(1, name+" natural expiry", 2100, f.complete, func() {})
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
