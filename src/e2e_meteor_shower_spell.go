package opennox

import (
	"fmt"
	"math"
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

type e2eMeteorShowerProjectile struct {
	object   *server.Object
	wire     uint32
	scriptID int32
	impact   bool
}

type e2eMeteorShowerFixture struct {
	units                       e2eMeteorFixture
	shower                      *server.Object
	wire, birth, expired        uint32
	lastHealFrame               uint32
	scriptID                    int32
	projectiles                 []*e2eMeteorShowerProjectile
	health                      uint16
	expected, healed, lastDelta int32
	castAudio, impacts, hits    int
	active, natural, validated  bool
}

func (f *e2eMeteorShowerFixture) prepare() {
	// Reuse only ordinary placement, host ownership, durable health and WAIT.
	// The single-Meteor fixture's result observers remain inactive.
	f.units.prepare()
	if f.units.level == 0 {
		f.units.actualLevel = int(noxServer.SpellPower4FE7B0(spell.SPELL_METEOR_SHOWER, f.units.caster))
	}
	if f.units.actualLevel < 1 || f.units.actualLevel > 5 {
		e2eError(fmt.Errorf("MeteorShower stock power outside 1..5: %d", f.units.actualLevel))
		return
	}
	f.units.damage = int32(math.RoundToEven(float64(float32(noxServer.Balance.FloatInd("MeteorDamage", f.units.actualLevel-1)))))
	// A five-second shower has many independently randomized impacts. This
	// is fixture durability, not an expected or supplied damage/HP result.
	asObjectS(f.units.target).SetMaxHealth(30000)
	// Nil-weapon player Explosion now applies real self splash as well. Give
	// the isolated host durable starting HP, never replenish it during a cast.
	asObjectS(f.units.host).SetMaxHealth(30000)
	lastHostHP := f.units.host.HealthData.Cur
	if ext := f.units.target.GetExt(); ext.HealthRegenToMax > 0 || ext.HealthRegenPerFrame >= 0 {
		e2eError(fmt.Errorf("MeteorShower fixture unexpectedly overrides ordinary monster regeneration"))
		return
	}
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.observeTick)
	noxServer.TickHook(func() {
		if !f.active || f.units.host.HealthData.Cur == lastHostHP {
			return
		}
		current := f.units.host.HealthData.Cur
		if current < lastHostHP {
			e2eLog.Printf("METEOR SHOWER HOST HIT: mode=%s level=%d HP=%d->%d type=%d source=%p frame=%d",
				f.units.mode, f.units.actualLevel, lastHostHP, current, f.units.host.Field131, f.units.host.Obj130, noxServer.Frame())
		}
		lastHostHP = current
	})
}

func (f *e2eMeteorShowerFixture) beginCast() {
	u := &f.units
	if !e2eObjectInWorld(u.target) || u.target.Buffs != 0 || u.caster.Buffs != 0 ||
		!noxServer.MapTraceRay(u.caster.PosVec, u.target.PosVec, server.MapTraceFlag1) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(u.target))) == nil {
		e2eError(fmt.Errorf("MeteorShower target placement/client publication did not settle"))
		return
	}
	f.health, f.active = u.target.HealthData.Cur, true
	u.caster.SetDir(server.DirFromVec(u.target.PosVec.Sub(u.caster.PosVec)))
	if u.level == 0 {
		// The normal monster action must reach its stock animation cast frame.
		u.caster.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_METEOR_SHOWER), 0, u.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_METEOR_SHOWER"), u.level, api.toObj(u.caster), e2eFistPosition{u.target.PosVec})
	noxServer.ObjectsAddPending()
	f.observeCreation()
}

func (f *e2eMeteorShowerFixture) observeCreation() {
	if f.shower != nil {
		return
	}
	var found []*server.Object
	for obj, remaining := f.units.caster.Field129, 4096; obj != nil; obj, remaining = obj.Field128, remaining-1 {
		if remaining == 0 {
			e2eError(fmt.Errorf("MeteorShower owned list did not terminate"))
			return
		}
		if obj.ObjectTypeC().ID() == "MeteorShower" {
			found = append(found, obj)
		}
	}
	if len(found) != 1 {
		e2eError(fmt.Errorf("MeteorShower produced %d owned sources, want one", len(found)))
		return
	}
	f.shower = found[0]
	f.wire, f.scriptID, f.birth = f.shower.NetCode, f.shower.ScriptIDVal, f.shower.Field32
	data := f.shower.UpdateDataMeteor()
	if data == nil || data.Damage != f.units.damage || f.shower.ObjOwner != f.units.caster ||
		f.shower.ZVal != 0 || f.shower.Field27 != 0 || f.shower.Field5&0x20 != 0 {
		e2eError(fmt.Errorf("MeteorShower source mismatch: owner=%p/%p damage=%v Z=%g speed=%g field5=%#x",
			f.shower.ObjOwner, f.units.caster, data, f.shower.ZVal, f.shower.Field27, f.shower.Field5))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(f.shower)) <= math.MaxUint32 || uintptr(f.shower.UpdateData) <= math.MaxUint32) {
		e2eError(fmt.Errorf("MeteorShower actual source/update allocation below 4 GiB"))
		return
	}
	e2eLog.Printf("METEOR SHOWER CAST: mode=%s requested=%d actual=%d caster=%p shower=%p update=%p damage=%d aim=%v birth=%d",
		f.units.mode, f.units.level, f.units.actualLevel, f.units.caster, f.shower, f.shower.UpdateData, data.Damage, f.shower.PosVec, f.birth)
}

func (f *e2eMeteorShowerFixture) recordProjectiles() {
	for _, obj := range noxServer.Objs.AllObjects() {
		if obj.ObjOwner != f.shower || obj.ObjectTypeC().ID() != "Meteor" {
			continue
		}
		known := false
		for _, p := range f.projectiles {
			if p.object == obj && p.wire == obj.NetCode && p.scriptID == obj.ScriptIDVal {
				known = true
				break
			}
		}
		if known {
			continue
		}
		data := obj.UpdateDataMeteor()
		if data == nil || data.Damage != f.units.damage || !(obj.ZVal > 0 && obj.ZVal <= 255) ||
			obj.Field27 != -8 || obj.Field5&0x20 == 0 {
			e2eError(fmt.Errorf("MeteorShower falling child mismatch: data=%v Z=%g speed=%g field5=%#x", data, obj.ZVal, obj.Field27, obj.Field5))
			return
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("MeteorShower actual child/update allocation below 4 GiB"))
			return
		}
		f.projectiles = append(f.projectiles, &e2eMeteorShowerProjectile{object: obj, wire: obj.NetCode, scriptID: obj.ScriptIDVal})
		e2eLog.Printf("METEOR SHOWER CHILD: mode=%s index=%d meteor=%p update=%p source=%p Z=%g speed=%g frame=%d",
			f.units.mode, len(f.projectiles), obj, obj.UpdateData, f.shower, obj.ZVal, obj.Field27, noxServer.Frame())
	}
}

func (f *e2eMeteorShowerFixture) observeSound(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
	if !f.active {
		return
	}
	if id == sound.SoundMeteorShowerCast {
		f.castAudio++
		f.observeCreation()
		if f.castAudio != 1 || kind != 0 || owner != f.units.caster || pos != f.units.caster.PosVec {
			e2eError(fmt.Errorf("MeteorShower owner-attached cast audio mismatch: count=%d kind=%d owner=%p/%p pos=%v", f.castAudio, kind, owner, f.units.caster, pos))
			return
		}
		if f.units.level == 0 {
			ud := f.units.caster.UpdateDataMonster()
			head := ud.AIStackHead()
			if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.units.target ||
				head.ArgU32(0) != uint32(spell.SPELL_METEOR_SHOWER) || ud.Field120_2 != 0 ||
				uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
				e2eError(fmt.Errorf("MeteorShower NPC bypassed the natural animation/cast frame"))
				return
			}
			f.natural = true
			e2eLog.Printf("METEOR SHOWER NPC ANIMATION: frame=%d cast-frame=%d stack-target=%p power=%d", noxServer.Frame(), ud.Field120_1, head.ArgObj(2), f.units.actualLevel)
		}
	}
	if id != sound.SoundMeteorHit {
		return
	}
	for _, p := range f.projectiles {
		if owner != p.object || !e2eFistInWorld(p.object, p.wire, p.scriptID) {
			continue
		}
		if p.impact || kind != 0 || p.object.ZVal > 0 || !e2eObjectInWorld(f.units.target) {
			e2eError(fmt.Errorf("MeteorShower child impact repeated or preceded natural landing"))
			return
		}
		// Regeneration may run before this impact in the same normal tick.
		// Observe it before the damage callback resets the injury timestamp.
		f.observeHealing()
		p.impact = true
		f.impacts++
		// Independent stock radial calculation with the actual live aim and
		// ordinary map occlusion. No RNG, aim, HP or damage packet is supplied.
		effective := e2eMeteorShowerImpactDamage(f.units.damage, pos, f.units.target.PosVec,
			noxServer.MapTraceRay(pos, f.units.target.PosVec, server.MapTraceFlag1))
		if effective > 0 {
			f.expected += effective
			f.lastDelta = effective
			f.hits++
		}
		e2eLog.Printf("METEOR SHOWER IMPACT: mode=%s count=%d effective=%d sum=%d pos=%v frame=%d", f.units.mode, f.impacts, effective, f.expected, pos, noxServer.Frame())
		return
	}
}

// Independent model of the existing unmodified classic monster regeneration:
// three minutes to full health, one-second injury pause and actual frame gate.
func e2eMeteorShowerRegenAmount(frame, injury, fps uint32, maximum, current int32) int32 {
	if fps == 0 || maximum <= 0 || current >= maximum || frame-injury <= fps {
		return 0
	}
	interval := 180 * uint64(fps)
	each, amount := uint32(interval/uint64(maximum)), int32(1)
	if each == 0 {
		each = 1
		amount = int32(uint64(maximum) / interval)
	}
	if frame%each != 0 {
		return 0
	}
	if amount > maximum-current {
		amount = maximum - current
	}
	return amount
}

func (f *e2eMeteorShowerFixture) observeHealing() {
	u := f.units.target
	before := int32(f.health) - f.expected + f.healed
	gain := int32(u.HealthData.Cur) - before
	if gain == 0 {
		return
	}
	want := e2eMeteorShowerRegenAmount(noxServer.Frame(), u.Frame134, noxServer.TickRate(), int32(u.HealthData.Max), before)
	if gain <= 0 || gain != want || f.lastHealFrame == noxServer.Frame() {
		e2eError(fmt.Errorf("MeteorShower HP change is not an ordinary regen tick: HP=%d/%d gain=%d want=%d frame=%d injury=%d",
			before, u.HealthData.Cur, gain, want, noxServer.Frame(), u.Frame134))
		return
	}
	f.healed += gain
	f.lastHealFrame = noxServer.Frame()
	e2eLog.Printf("METEOR SHOWER REGEN: mode=%s HP=%d->%d gain=%d healed=%d frame=%d injury=%d", f.units.mode, before, u.HealthData.Cur, gain, f.healed, noxServer.Frame(), u.Frame134)
}

func (f *e2eMeteorShowerFixture) observeTick() {
	if !f.active || f.shower == nil {
		return
	}
	f.recordProjectiles()
	if !e2eObjectInWorld(f.units.target) || f.units.target.HealthData == nil || f.expected >= int32(f.health)+f.healed {
		e2eError(fmt.Errorf("MeteorShower durable target disappeared or exhausted"))
		return
	}
	f.observeHealing()
	if got := f.units.target.HealthData.Cur; int32(got) != int32(f.health)-f.expected+f.healed ||
		f.expected != 0 && f.units.target.Field131 != uint32(object.DamageExplosion) {
		e2eError(fmt.Errorf("MeteorShower actual damage differs from observed impacts: HP=%d->%d sum=%d type=%d", f.health, got, f.expected, f.units.target.Field131))
		return
	}
	if f.expired == 0 && !e2eFistInWorld(f.shower, f.wire, f.scriptID) {
		f.expired = noxServer.Frame()
		elapsed, lifetime := f.expired-f.birth, 5*noxServer.TickRate()
		if elapsed < lifetime || elapsed > lifetime+2 {
			e2eError(fmt.Errorf("MeteorShower natural lifetime=%d want=%d..%d", elapsed, lifetime, lifetime+2))
			return
		}
		e2eLog.Printf("METEOR SHOWER EXPIRED: mode=%s elapsed=%d lifetime=%d children=%d", f.units.mode, elapsed, lifetime, len(f.projectiles))
	}
}

func (f *e2eMeteorShowerFixture) fallingVisible() bool {
	for _, p := range f.projectiles {
		if e2eFistInWorld(p.object, p.wire, p.scriptID) && p.object.ZVal > 100 && p.object.ZVal < 230 &&
			noxClient.Objs.ByNetCode(uint16(p.wire)) != nil {
			return true
		}
	}
	return false
}

func (f *e2eMeteorShowerFixture) complete() bool {
	if f.expired == 0 || f.castAudio != 1 || len(f.projectiles) == 0 || f.impacts != len(f.projectiles) || f.hits == 0 ||
		f.units.level == 0 && (!f.natural || f.units.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	for _, p := range f.projectiles {
		// Source expiry transfers surviving children to the caster normally.
		// Never follow/dereference a freed retained source or child pointer.
		if e2eFistInWorld(p.object, p.wire, p.scriptID) || noxClient.Objs.ByNetCode(uint16(p.wire)) != nil {
			return false
		}
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.units.target)))
	if dr == nil {
		return false
	}
	if delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32); !ok || int32(delta) != -f.lastDelta {
		return false
	}
	f.validated = true
	e2eLog.Printf("METEOR SHOWER COMPLETE: mode=%s requested=%d actual=%d HP=%d->%d damage/healed=%d/%d hits=%d children/impacts=%d/%d client-last-delta=%d world/drawable=removed natural-NPC=%t",
		f.units.mode, f.units.level, f.units.actualLevel, f.health, f.units.target.HealthData.Cur, f.expected, f.healed, f.hits, len(f.projectiles), f.impacts, -f.lastDelta, f.natural)
	return true
}

func (f *e2eMeteorShowerFixture) cleanup() {
	if !f.validated {
		e2eError(fmt.Errorf("MeteorShower fixture cleanup before natural completion"))
		return
	}
	f.active = false
	f.units.cleanup()
}

// Actual script-to-position and animated NPC casts exercise native acceptance,
// the five-second source, randomized child production, physics, explosion HP,
// client replay and natural removal. Reserved NPC actions are not proof of
// autonomous spell selection; player mana/incantation input is not simulated.
func (sc *e2eScenario) CheckMeteorShowerSpell(level int, mode, name string) {
	if !e2eMeteorMode(level, mode) {
		e2eError(fmt.Errorf("invalid MeteorShower scenario %s/%d", mode, level))
		return
	}
	f := &e2eMeteorShowerFixture{units: e2eMeteorFixture{level: level, mode: mode}}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" falling drawable", 180, f.fallingVisible, func() {
		e2eLog.Printf("METEOR SHOWER FALL: mode=%s requested=%d actual=%d drawable=live", mode, level, f.units.actualLevel)
	})
	sc.CaptureMagicFrame(name + " actual falling frame")
	sc.addWhen(1, name+" natural shower expiry and client replay", 600, f.complete, func() {})
	sc.CaptureMagicFrame(name + " actual final impact frame")
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
