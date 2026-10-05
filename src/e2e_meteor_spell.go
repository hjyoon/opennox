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

func e2eMeteorMode(level int, mode string) bool {
	return mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
}

// Independent calculation of the stock 80/30 radial falloff. This observes the
// actual cast position; it does not replace the NPC's original aim RNG.
func e2eMeteorRadialDamage(raw int32, from, to types.Pointf) (int32, bool) {
	dx, dy := to.X-from.X, to.Y-from.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance > 80 {
		return 0, false
	}
	damage := float32(raw)
	if distance >= 30 {
		damage *= 1 - (distance-30)/50
	}
	return int32(damage), true
}

type e2eMeteorFixture struct {
	level, actualLevel           int
	mode                         string
	host, caster, target, meteor *server.Object
	original, aim                types.Pointf
	wire, frame                  uint32
	scriptID                     int32
	health                       uint16
	damage, effective            int32
	castAudio, impactAudio       int
	active, natural, hit         bool
}

func (f *e2eMeteorFixture) ownedMeteors() []*server.Object {
	var out []*server.Object
	for obj, remaining := f.caster.Field129, 4096; obj != nil; obj, remaining = obj.Field128, remaining-1 {
		if remaining == 0 {
			e2eError(fmt.Errorf("Meteor owned list did not terminate"))
			return nil
		}
		if obj.ObjectTypeC().ID() == "Meteor" {
			out = append(out, obj)
		}
	}
	return out
}

func (f *e2eMeteorFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Buffs != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Meteor requires a live unenchanted host"))
		return
	}
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+24, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.target = noxServer.NewObjectByTypeID("Troll")
	if f.target == nil || f.target.HealthData == nil || f.target.UpdateData == nil {
		e2eError(fmt.Errorf("Meteor requires the stock Troll damage target"))
		return
	}
	f.caster = f.host
	if f.level == 0 {
		f.caster = noxServer.NewObjectByTypeID("NPC")
		if f.caster == nil || f.caster.UpdateData == nil {
			e2eError(fmt.Errorf("Meteor requires the stock NPC caster"))
			return
		}
		// A neutral NPC and Troll are allies under the original owner gate.
		// Use ordinary host ownership for this allied NPC casting fixture;
		// never bypass IsEnemy or supply an expected HP result.
		noxServer.CreateObjectAt(f.caster, f.host, origin)
		asObjectS(f.host).SetPos(origin.Sub(direction.Mul(80)))
	} else {
		asObjectS(f.host).SetPos(origin)
	}
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.target, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	if !noxServer.S().IsEnemyTo(f.target, f.caster.FindOwnerChainPlayer()) {
		e2eError(fmt.Errorf("Meteor fixture does not have an enemy damage target"))
		return
	}
	// Placement, durable target health and ordinary WAIT are fixture setup.
	// Cast power, aim RNG, animation frame, Meteor state and HP results are stock.
	asObjectS(f.target).SetMaxHealth(2000)
	f.target.UpdateDataMonster().SetAggression(0)
	f.target.ClearActionStack()
	f.target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	if f.level == 0 {
		f.caster.UpdateDataMonster().SetAggression(0)
		f.caster.ClearActionStack()
		f.caster.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	}
	if f.target.Buffs != 0 || f.caster.Buffs != 0 || len(f.ownedMeteors()) != 0 {
		e2eError(fmt.Errorf("Meteor baseline is enchanted or already owns a Meteor"))
		return
	}
	f.actualLevel = f.level
	if f.level == 0 {
		f.actualLevel = int(noxServer.SpellPower4FE7B0(spell.SPELL_METEOR, f.caster))
	}
	if f.actualLevel < 1 || f.actualLevel > 5 {
		e2eError(fmt.Errorf("Meteor stock power is outside 1..5: %d", f.actualLevel))
		return
	}
	f.damage = int32(math.RoundToEven(float64(float32(noxServer.Balance.FloatInd("MeteorDamage", f.actualLevel-1)))))
	if f.damage <= 0 || f.damage >= 2000 {
		e2eError(fmt.Errorf("Meteor stock damage is outside durable health: %d", f.damage))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.caster), unsafe.Pointer(f.target), f.caster.UpdateData, f.target.UpdateData} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Meteor actual native unit/update allocation below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.observeHit)
}

func (f *e2eMeteorFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, server.MapTraceFlag1) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Meteor target placement/client publication did not settle"))
		return
	}
	f.health, f.frame, f.active = f.target.HealthData.Cur, noxServer.Frame(), true
	f.caster.SetDir(server.DirFromVec(f.target.PosVec.Sub(f.caster.PosVec)))
	if f.level == 0 {
		// Reserve a normal CAST_ON_OBJECT action. Animation and its original
		// cast-frame gate must execute it; no direct cast or frame is supplied.
		f.caster.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_METEOR), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_METEOR"), f.level, api.toObj(f.caster), e2eFistPosition{f.target.PosVec})
	noxServer.ObjectsAddPending()
	f.observeCreation()
	// A second normal script request must be refused before physics advances.
	api.CastSpellLvl(nsp.Spell("SPELL_METEOR"), f.level, api.toObj(f.caster), e2eFistPosition{f.target.PosVec})
	if owned := f.ownedMeteors(); len(owned) != 1 || owned[0] != f.meteor {
		e2eError(fmt.Errorf("Meteor duplicate script cast changed ownership"))
	}
}

func (f *e2eMeteorFixture) observeCreation() {
	owned := f.ownedMeteors()
	if len(owned) != 1 {
		e2eError(fmt.Errorf("Meteor produced %d owned objects, want one", len(owned)))
		return
	}
	f.meteor = owned[0]
	f.wire, f.scriptID, f.aim = f.meteor.NetCode, f.meteor.ScriptIDVal, f.meteor.PosVec
	data := f.meteor.UpdateDataMeteor()
	if data == nil || data.Damage != f.damage || f.meteor.ObjOwner != f.caster || f.meteor.ZVal != 255 ||
		f.meteor.Field27 != float32(-noxServer.Balance.Float("MeteorSpeed")) || f.meteor.Field5&0x20 == 0 {
		e2eError(fmt.Errorf("Meteor creation mismatch: owner=%p/%p data=%v Z=%g speed=%g field5=%#x",
			f.meteor.ObjOwner, f.caster, data, f.meteor.ZVal, f.meteor.Field27, f.meteor.Field5))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(f.meteor)) <= math.MaxUint32 || uintptr(f.meteor.UpdateData) <= math.MaxUint32) {
		e2eError(fmt.Errorf("Meteor actual object/update pointer below 4 GiB"))
		return
	}
	e2eLog.Printf("METEOR CAST: mode=%s requested=%d actual=%d caster=%p meteor=%p update=%p damage=%d aim=%v frame=%d",
		f.mode, f.level, f.actualLevel, f.caster, f.meteor, f.meteor.UpdateData, f.damage, f.aim, noxServer.Frame())
}

func (f *e2eMeteorFixture) observeSound(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
	if !f.active {
		return
	}
	if id == sound.SoundMeteorCast {
		f.castAudio++
		if f.meteor == nil {
			f.observeCreation()
		}
		if f.castAudio != 1 || kind != 0 || owner != nil || pos != f.aim {
			e2eError(fmt.Errorf("Meteor positional cast sound mismatch: count=%d kind=%d owner=%p pos=%v/%v", f.castAudio, kind, owner, pos, f.aim))
			return
		}
		if f.level == 0 {
			ud, head := f.caster.UpdateDataMonster(), f.caster.UpdateDataMonster().AIStackHead()
			if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
				head.ArgU32(0) != uint32(spell.SPELL_METEOR) || ud.Field120_2 != 0 ||
				uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
				e2eError(fmt.Errorf("Meteor NPC bypassed the natural animation/cast frame"))
				return
			}
			f.natural = true
			e2eLog.Printf("METEOR NPC ANIMATION: frame=%d cast-frame=%d stack-target=%p power=%d", noxServer.Frame(), ud.Field120_1, head.ArgObj(2), f.actualLevel)
		}
	}
	if id == sound.SoundMeteorHit && owner == f.meteor {
		f.impactAudio++
		if f.impactAudio != 1 || kind != 0 || !e2eFistInWorld(f.meteor, f.wire, f.scriptID) || f.meteor.ZVal > 0 {
			e2eError(fmt.Errorf("Meteor impact sound preceded natural landing or repeated"))
			return
		}
		var inside bool
		f.effective, inside = e2eMeteorRadialDamage(f.damage, pos, f.target.PosVec)
		if !inside || f.effective <= 0 {
			e2eError(fmt.Errorf("Meteor stock NPC aim missed the damage fixture: %v/%v", pos, f.target.PosVec))
		}
	}
}

func (f *e2eMeteorFixture) observeHit() {
	if !f.active || f.hit || f.impactAudio == 0 {
		return
	}
	if !e2eObjectInWorld(f.target) || f.target.HealthData == nil {
		e2eError(fmt.Errorf("Meteor durable target disappeared"))
		return
	}
	if f.target.HealthData.Cur == f.health {
		return
	}
	if f.target.HealthData.Cur != f.health-uint16(f.effective) || f.target.Field131 != uint32(object.DamageExplosion) {
		e2eError(fmt.Errorf("Meteor actual explosion damage mismatch: HP=%d->%d want=%d type=%d", f.health, f.target.HealthData.Cur, f.health-uint16(f.effective), f.target.Field131))
		return
	}
	f.hit = true
	e2eLog.Printf("METEOR HIT: mode=%s requested=%d actual=%d HP=%d->%d raw/effective=%d/%d elapsed=%d natural-NPC=%t",
		f.mode, f.level, f.actualLevel, f.health, f.target.HealthData.Cur, f.damage, f.effective, noxServer.Frame()-f.frame, f.natural)
}

func (f *e2eMeteorFixture) complete() bool {
	if !f.hit || f.castAudio != 1 || f.impactAudio != 1 || f.level == 0 && (!f.natural || f.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	if e2eFistInWorld(f.meteor, f.wire, f.scriptID) || len(f.ownedMeteors()) != 0 || noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
		return false
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if dr == nil {
		return false
	}
	if delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32); !ok || int32(delta) != -f.effective {
		return false
	}
	e2eLog.Printf("METEOR COMPLETE: mode=%s requested=%d actual=%d server/client-HP=verified world/owned/drawable=removed cast/impact=1/1 natural-NPC=%t",
		f.mode, f.level, f.actualLevel, f.natural)
	return true
}

func (f *e2eMeteorFixture) cleanup() {
	f.active = false
	noxServer.DelayedDelete(f.target)
	if f.caster != f.host {
		noxServer.DelayedDelete(f.caster)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Five real script-to-position casts and an animated NPC CAST_ON_OBJECT pass
// through normal acceptance, creation, physics, impact and client publication.
// The reserved NPC action tests the reported call chain, not autonomous spell
// selection. Player incantation/mana and every possible damage target are out
// of scope; no spell, HP, packet or pixel result is supplied by this observer.
func (sc *e2eScenario) CheckMeteorSpell(level int, mode, name string) {
	if !e2eMeteorMode(level, mode) {
		e2eError(fmt.Errorf("invalid Meteor scenario %s/%d", mode, level))
		return
	}
	f := &e2eMeteorFixture{level: level, mode: mode}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		// Arena respawn grants five seconds of invulnerability. Observe its
		// normal expiry rather than clearing a production enchant as setup.
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" falling drawable", 180, func() bool {
		return f.meteor != nil && e2eFistInWorld(f.meteor, f.wire, f.scriptID) && f.meteor.ZVal > 100 && f.meteor.ZVal < 230 &&
			noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.meteor))) != nil
	}, func() {
		e2eLog.Printf("METEOR FALL: mode=%s Z=%g velocity=%g drawable=live", mode, f.meteor.ZVal, f.meteor.Field27)
	})
	sc.CaptureMagicFrame(name + " actual falling frame")
	sc.addWhen(1, name+" natural impact and client replay", 300, f.complete, func() {})
	sc.CaptureMagicFrame(name + " actual impact frame")
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
