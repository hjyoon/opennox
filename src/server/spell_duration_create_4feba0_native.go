package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellDurationCreateRuntime4FEBA0 supplies the operations that remain owned
// by the outer game runtime. Callback and object identities stay native-width;
// callback results and audio arguments retain their original dword contracts.
type SpellDurationCreateRuntime4FEBA0 struct {
	DestroySpell func(*DurSpell)
	CallCreate   func(unsafe.Pointer, *DurSpell) int32
	AudioEvent   func(sound.ID, *Object, int, uint32)
}

type spellDurationCreateNativeDeps4FEBA0 struct {
	loadGlyphCache  func() uint32
	lookupGlyph     func() uint32
	storeGlyphCache func(uint32)
	findDuplicate   func(int32, *Object) int32
	cancelFor       func(int32, *Object)
	beforeCreate    func()
	newRecord       func() *DurSpell
	loadFrame       func() uint32
	addRecord       func(*DurSpell)
	spellHasFlags   func(int32, uint32) int32
	spellAudio      func(int32, int32) int32
	audioEvent      func(int32, *Object, int32, int32)
	callCreate      func(unsafe.Pointer, *DurSpell) int32
	cancelSpell     func(*DurSpell)
}

func spellDurationCreateNative4FEBA0(
	spellID int32,
	second, third, fourth *Object,
	arg *SpellAcceptArg,
	level int32,
	create, update, destroy unsafe.Pointer,
	duration uint32,
	deps spellDurationCreateNativeDeps4FEBA0,
) int32 {
	return SpellDurationCreate4FEBA0(SpellDurationCreateHooks4FEBA0[
		*DurSpell,
		*Object,
		*SpellAcceptArg,
		unsafe.Pointer,
	]{
		LoadGlyphCache:  deps.loadGlyphCache,
		LookupGlyph:     deps.lookupGlyph,
		StoreGlyphCache: deps.storeGlyphCache,
		LoadFourthArg: func() *Object {
			return fourth
		},
		LoadThirdArg: func() *Object {
			return third
		},
		LoadObjectFlags: func(object *Object) uint32 {
			return uint32(object.ObjFlags)
		},
		LoadObjectType: func(object *Object) uint16 {
			return object.TypeInd
		},
		LoadSpellArg: func() int32 {
			return spellID
		},
		FindDuplicate: deps.findDuplicate,
		CancelFor:     deps.cancelFor,
		BeforeCreate:  deps.beforeCreate,
		NewRecord:     deps.newRecord,
		LoadLevelArg: func() int32 {
			return level
		},
		LoadSecondArg: func() *Object {
			return second
		},
		StoreSpell: func(record *DurSpell, value uint32) {
			record.Spell = value
		},
		StoreLevel: func(record *DurSpell, value uint32) {
			record.Level = value
		},
		StoreCaster: func(record *DurSpell, value *Object) {
			record.Caster16 = value
		},
		StoreSource: func(record *DurSpell, value *Object) {
			record.Obj12 = value
		},
		StoreSub108: func(record, value *DurSpell) {
			record.Sub108 = value
		},
		StoreSub104: func(record, value *DurSpell) {
			record.Sub104 = value
		},
		StoreMode: func(record *DurSpell, value uint32) {
			record.Flag20 = value
		},
		StoreAnchor: func(record *DurSpell, value *Object) {
			record.Obj24 = value
		},
		LoadPositionX: func(object *Object) float32 {
			return object.PosVec.X
		},
		StorePositionX: func(record *DurSpell, value float32) {
			record.Pos.X = value
		},
		LoadPositionY: func(object *Object) float32 {
			return object.PosVec.Y
		},
		LoadAcceptArg: func() *SpellAcceptArg {
			return arg
		},
		StoreField36: func(record *DurSpell, value uint32) {
			record.Field36 = value
		},
		StorePositionY: func(record *DurSpell, value float32) {
			record.Pos.Y = value
		},
		LoadTarget: func(arg *SpellAcceptArg) *Object {
			return arg.Obj
		},
		LoadCreateArg: func() unsafe.Pointer {
			return create
		},
		StoreTarget: func(record *DurSpell, value *Object) {
			record.Target48 = value
		},
		LoadAcceptX: func(arg *SpellAcceptArg) float32 {
			return arg.Pos.X
		},
		StoreAcceptX: func(record *DurSpell, value float32) {
			record.Pos2.X = value
		},
		LoadUpdateArg: func() unsafe.Pointer {
			return update
		},
		LoadAcceptY: func(arg *SpellAcceptArg) float32 {
			return arg.Pos.Y
		},
		StoreCreate: func(record *DurSpell, value unsafe.Pointer) {
			record.Create = value
		},
		StoreAcceptY: func(record *DurSpell, value float32) {
			record.Pos2.Y = value
		},
		LoadDestroyArg: func() unsafe.Pointer {
			return destroy
		},
		StoreUpdate: func(record *DurSpell, value unsafe.Pointer) {
			record.Update = value
		},
		StoreDestroy: func(record *DurSpell, value unsafe.Pointer) {
			record.Destroy = value
		},
		LoadFrame: deps.loadFrame,
		LoadDurationArg: func() uint32 {
			return duration
		},
		StoreFrame60: func(record *DurSpell, value uint32) {
			record.Frame60 = value
		},
		StoreFrame64: func(record *DurSpell, value uint32) {
			record.Frame64 = value
		},
		StoreFlagsLowByte: func(record *DurSpell, value byte) {
			record.Flags88 = record.Flags88&^0xff | uint32(value)
		},
		StoreFrame68: func(record *DurSpell, value uint32) {
			record.Frame68 = value
		},
		AddRecord:     deps.addRecord,
		SpellHasFlags: deps.spellHasFlags,
		SpellAudio:    deps.spellAudio,
		AudioEvent:    deps.audioEvent,
		CallCreate:    deps.callCreate,
		CancelSpell:   deps.cancelSpell,
	})
}

func spellDurationCreateServerDeps4FEBA0(
	sp *SpellsDuration,
	runtime SpellDurationCreateRuntime4FEBA0,
) spellDurationCreateNativeDeps4FEBA0 {
	return spellDurationCreateNativeDeps4FEBA0{
		loadGlyphCache: func() uint32 {
			return uint32(sp.s.Types.fast.glyph)
		},
		lookupGlyph: func() uint32 {
			return uint32(sp.s.Types.IndByID("Glyph"))
		},
		storeGlyphCache: func(value uint32) {
			sp.s.Types.fast.glyph = int(value)
		},
		findDuplicate: func(spellID int32, caster *Object) int32 {
			if sp.Sub4FEE50(spell.ID(spellID), caster) {
				return 1
			}
			return 0
		},
		cancelFor: func(spellID int32, caster *Object) {
			sp.SpellCancelDurSpell4FEB10(spellID, caster)
		},
		beforeCreate: func() {
			sp.SpellDurationCleanupTraversal4FED70(runtime.DestroySpell)
		},
		newRecord: sp.SpellDurationNew4FE950,
		loadFrame: sp.s.Frame,
		addRecord: sp.SpellDurationInsert4FED40,
		spellHasFlags: func(spellID int32, mask uint32) int32 {
			if sp.s.Spells.Flags(spell.ID(spellID)).Has(things.SpellFlags(mask)) {
				return 1
			}
			return 0
		},
		spellAudio: func(spellID, selector int32) int32 {
			return int32(sp.s.Spells.DefByInd(spell.ID(spellID)).GetAudio(int(selector)))
		},
		audioEvent: func(id int32, object *Object, kind, code int32) {
			runtime.AudioEvent(sound.ID(id), object, int(kind), uint32(code))
		},
		callCreate: runtime.CallCreate,
		cancelSpell: func(record *DurSpell) {
			sp.SpellDurationCancel4FE9D0(record)
		},
	}
}

// SpellDurationCreate4FEBA0 binds GAME.EXE 004FEBA0 to native-width duration
// records, objects, acceptance arguments, and callback pointers. Spell, level,
// type, flags, frame, duration, and callback-result values keep their exact
// fixed-width contracts. The adapter deliberately preserves the executable's
// nil-object fault boundaries instead of adding Go-level guards.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationCreate4FEBA0(
	spellID int32,
	second, third, fourth *Object,
	arg *SpellAcceptArg,
	level int32,
	create, update, destroy unsafe.Pointer,
	duration uint32,
	runtime SpellDurationCreateRuntime4FEBA0,
) int32 {
	return spellDurationCreateNative4FEBA0(
		spellID,
		second,
		third,
		fourth,
		arg,
		level,
		create,
		update,
		destroy,
		duration,
		spellDurationCreateServerDeps4FEBA0(sp, runtime),
	)
}
