package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func e2eTriggerGlyphMode(level int, mode string) bool {
	return mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
}

type e2eTriggerGlyphFixture struct {
	level                             int
	mode                              string
	host, caster, selected, foreign   *server.Object
	original, selectedPos, foreignPos types.Pointf
	wire, foreignWire, frame          uint32
	scriptID, foreignScriptID         int32
	castAudio, detonateAudio          int
	active, natural                   bool
}

func (f *e2eTriggerGlyphFixture) ownedGlyphs() []*server.Object {
	var out []*server.Object
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.HasOwner(f.caster) && obj.ObjectTypeC().ID() == "Glyph" {
			out = append(out, obj)
		}
	}
	return out
}

func (f *e2eTriggerGlyphFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Buffs != 0 || f.host.Poison540 != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("TriggerGlyph requires a live unenchanted Wizard host"))
		return
	}
	e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.GetMousePos()})
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+104,
		func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.caster = f.host
	if f.mode == "npc-animated" {
		f.caster = noxServer.NewObjectByTypeID("NPC")
		if f.caster == nil || f.caster.UpdateData == nil {
			e2eError(fmt.Errorf("TriggerGlyph requires the stock NPC caster"))
			return
		}
		// Ordinary ownership and WAIT are setup only. Natural animation and
		// the original cast-frame gate, not this fixture, execute the spell.
		noxServer.CreateObjectAt(f.caster, f.host, origin)
		f.caster.UpdateDataMonster().SetAggression(0)
		f.caster.ClearActionStack()
		f.caster.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		asObjectS(f.host).SetPos(origin.Sub(direction.Mul(80)))
	}
	if len(f.ownedGlyphs()) != 0 || noxServer.Spells.DefByInd(spell.SPELL_TRIGGER_GLYPH).GetCastSound() != sound.SoundTriggerGlyphCast {
		e2eError(fmt.Errorf("TriggerGlyph fixture has existing owned Glyphs or non-stock cast audio"))
		return
	}
	f.selected, f.foreign = noxServer.NewObjectByTypeID("Glyph"), noxServer.NewObjectByTypeID("Glyph")
	if f.selected == nil || f.foreign == nil {
		e2eError(fmt.Errorf("TriggerGlyph requires the stock Glyph type"))
		return
	}
	// The nearer unowned Glyph must survive. The 200-unit separation keeps
	// even diagonal placements outside the original 100-unit chain rectangle.
	// Newly created
	// Glyphs have no stored spells; do not inject a contained-spell outcome.
	noxServer.CreateObjectAt(f.selected, f.caster, origin.Add(direction.Mul(160)))
	noxServer.CreateObjectAt(f.foreign, nil, origin.Sub(direction.Mul(40)))
	noxServer.ObjectsAddPending()
	for _, glyph := range []*server.Object{f.selected, f.foreign} {
		if !e2eObjectInWorld(glyph) || glyph.InitData == nil || glyph.InitDataGlyph().SpellsCnt != 0 ||
			glyph.Flags().Has(object.FlagDestroyed) || glyph.ObjectTypeC().ID() != "Glyph" {
			e2eError(fmt.Errorf("TriggerGlyph did not initialize an empty stock Glyph"))
			return
		}
	}
	if f.selected.ObjOwner != f.caster || f.foreign.ObjOwner != nil || len(f.ownedGlyphs()) != 1 {
		e2eError(fmt.Errorf("TriggerGlyph setup ownership mismatch"))
		return
	}
	f.wire, f.scriptID, f.selectedPos = f.selected.NetCode, f.selected.ScriptIDVal, f.selected.PosVec
	f.foreignWire, f.foreignScriptID, f.foreignPos = f.foreign.NetCode, f.foreign.ScriptIDVal, f.foreign.PosVec
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.caster), unsafe.Pointer(f.selected), unsafe.Pointer(f.foreign), f.selected.InitData, f.foreign.InitData} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("TriggerGlyph fixture pointer below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(f.observeSound)
	e2eLog.Printf("TRIGGER GLYPH SETUP: mode=%s level=%d caster=%p selected=%p init=%p foreign=%p owned=1 stored-spells=0 selected-pos=%v foreign-pos=%v",
		f.mode, f.level, f.caster, f.selected, f.selected.InitData, f.foreign, f.selectedPos, f.foreignPos)
}

func (f *e2eTriggerGlyphFixture) foreignUnchanged() bool {
	return e2eFistInWorld(f.foreign, f.foreignWire, f.foreignScriptID) &&
		f.foreign.ObjOwner == nil && !f.foreign.Flags().Has(object.FlagDestroyed) && f.foreign.PosVec == f.foreignPos &&
		noxClient.Objs.ByNetCode(uint16(f.foreignWire)) != nil
}

func (f *e2eTriggerGlyphFixture) beginCast() {
	if !e2eFistInWorld(f.selected, f.wire, f.scriptID) || f.selected.Flags().Has(object.FlagDestroyed) ||
		f.selected.PosVec != f.selectedPos || noxClient.Objs.ByNetCode(uint16(f.wire)) == nil || !f.foreignUnchanged() {
		e2eError(fmt.Errorf("TriggerGlyph stock objects were not published intact"))
		return
	}
	f.frame, f.active = noxServer.Frame(), true
	if f.mode == "npc-animated" {
		f.caster.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_TRIGGER_GLYPH), 0, f.caster)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_TRIGGER_GLYPH"), f.level, api.toObj(f.caster), api.toObj(f.caster))
}

func (f *e2eTriggerGlyphFixture) observeSound(id sound.ID, kind int, obj *server.Object, pos types.Pointf) {
	if !f.active {
		return
	}
	switch id {
	case sound.SoundTriggerGlyphCast:
		f.castAudio++
		if f.castAudio != 1 || f.detonateAudio != 0 || obj != f.caster || kind != 0 || pos != f.caster.PosVec ||
			!e2eFistInWorld(f.selected, f.wire, f.scriptID) || f.selected.Flags().Has(object.FlagDestroyed) {
			e2eError(fmt.Errorf("TriggerGlyph cast audio did not precede actual death"))
			return
		}
		if f.mode == "npc-animated" {
			ud, head := f.caster.UpdateDataMonster(), f.caster.UpdateDataMonster().AIStackHead()
			if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.caster ||
				head.ArgU32(0) != uint32(spell.SPELL_TRIGGER_GLYPH) || ud.Field120_2 != 0 || uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
				e2eError(fmt.Errorf("TriggerGlyph NPC bypassed natural animation/cast-frame"))
				return
			}
			f.natural = true
			e2eLog.Printf("TRIGGER GLYPH NPC ANIMATION: animation=%d elapsed=%d target=%p", ud.Field120_1, noxServer.Frame()-f.frame, head.ArgObj(2))
		}
	case sound.SoundGlyphDetonate:
		f.detonateAudio++
		if f.castAudio != 1 || f.detonateAudio != 1 || obj != nil || kind != 0 || pos != f.selectedPos ||
			!e2eFistInWorld(f.selected, f.wire, f.scriptID) || !f.selected.Flags().Has(object.FlagDestroyed) {
			e2eError(fmt.Errorf("TriggerGlyph effect was not produced by the actual selected Glyph death"))
			return
		}
		e2eLog.Printf("TRIGGER GLYPH EFFECT: mode=%s level=%d cast-audio=%d detonate-audio=%d selected=%p destroyed=true effect-pos=%v",
			f.mode, f.level, f.castAudio, f.detonateAudio, f.selected, pos)
	}
}

func (f *e2eTriggerGlyphFixture) complete() bool {
	if f.castAudio != 1 || f.detonateAudio != 1 || e2eFistInWorld(f.selected, f.wire, f.scriptID) ||
		noxClient.Objs.ByNetCode(uint16(f.wire)) != nil || len(f.ownedGlyphs()) != 0 ||
		f.mode == "npc-animated" && (!f.natural || f.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	if !f.foreignUnchanged() {
		e2eError(fmt.Errorf("TriggerGlyph selected or chained the nearer unrelated Glyph"))
		return false
	}
	e2eLog.Printf("TRIGGER GLYPH COMPLETE: mode=%s level=%d elapsed=%d audio=1/1 natural-NPC=%t selected-world/owned/drawable=removed foreign-world/drawable=unchanged",
		f.mode, f.level, noxServer.Frame()-f.frame, f.natural)
	return true
}

func (f *e2eTriggerGlyphFixture) castWithoutOwnedGlyph() {
	api := noxServer.noxScriptP()
	level := f.level
	if level == 0 {
		level = 1
	}
	// Ordinary acceptance may play its original fizzle audio. It must not
	// play TriggerGlyph cast/detonate audio or touch the unrelated Glyph.
	api.CastSpellLvl(nsp.Spell("SPELL_TRIGGER_GLYPH"), level, api.toObj(f.caster), api.toObj(f.caster))
}

func (f *e2eTriggerGlyphFixture) checkEmptyCast() {
	if f.castAudio != 1 || f.detonateAudio != 1 || len(f.ownedGlyphs()) != 0 || !f.foreignUnchanged() {
		e2eError(fmt.Errorf("TriggerGlyph empty-owner cast changed audio or unrelated Glyph"))
		return
	}
	e2eLog.Printf("TRIGGER GLYPH EMPTY: mode=%s level=%d additional-trigger/detonate-audio=0 foreign=unchanged", f.mode, f.level)
}

func (f *e2eTriggerGlyphFixture) cleanup() {
	f.active = false
	noxServer.DelayedDelete(f.foreign)
	if f.mode == "npc-animated" {
		noxServer.DelayedDelete(f.caster)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Test selection, real glyph death/point effect and natural removal through
// script acceptance and one animated NPC action. Empty stock Glyphs deliberately
// do not claim contained-spell damage, player incantation/mana or NPC autonomy.
func (sc *e2eScenario) CheckTriggerGlyphSpell(level int, mode, name string) {
	if !e2eTriggerGlyphMode(level, mode) {
		e2eError(fmt.Errorf("invalid TriggerGlyph mode/level %s/%d", mode, level))
		return
	}
	f := &e2eTriggerGlyphFixture{level: level, mode: mode}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual cast and glyph effect", 180, func() bool {
		return f.castAudio == 1 && f.detonateAudio == 1
	}, func() {})
	sc.CaptureMagicFrame(name + " actual glyph effect frame")
	sc.addWhen(1, name+" natural world and client removal", 300, f.complete, func() {})
	sc.add(0, name+" empty-owner cast", f.castWithoutOwnedGlyph)
	sc.Wait(3, name+" empty-owner settle")
	sc.add(0, name+" empty-owner observation", f.checkEmptyCast)
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
