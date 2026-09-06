package server

import "unsafe"

// SpellDurationProcessRuntime4FEEF0 supplies the two operations still owned by
// the outer game runtime. Callback and duration-record identities remain
// native-width, while callback results retain the original signed dword ABI.
type SpellDurationProcessRuntime4FEEF0 struct {
	Destroy    func(*DurSpell)
	CallUpdate func(unsafe.Pointer, *DurSpell) int32
}

type spellDurationProcessNativeDeps4FEEF0 struct {
	loadFirst              func() *DurSpell
	loadFlagsLowByte       func(*DurSpell) byte
	loadNext               func(*DurSpell) *DurSpell
	destroy                func(*DurSpell)
	loadCaster             func(*DurSpell) *Object
	loadObjectFlags        func(*Object) uint32
	storeCaster            func(*DurSpell, *Object)
	loadObj12              func(*DurSpell) *Object
	loadObjectFlagsLowByte func(*Object) byte
	storeObj12             func(*DurSpell, *Object)
	loadFlag20             func(*DurSpell) uint32
	loadObj24              func(*DurSpell) *Object
	storeObj24             func(*DurSpell, *Object)
	loadFrame68            func(*DurSpell) uint32
	loadFrame60            func(*DurSpell) uint32
	loadCurrentFrame       func() uint32
	loadUpdate             func(*DurSpell) unsafe.Pointer
	callUpdate             func(unsafe.Pointer, *DurSpell) int32
	cancel                 func(*DurSpell)
}

func spellDurationProcessNative4FEEF0(deps spellDurationProcessNativeDeps4FEEF0) {
	SpellDurationProcess4FEEF0(
		SpellDurationProcessHooks4FEEF0[*DurSpell, *Object, unsafe.Pointer]{
			LoadFirst:              deps.loadFirst,
			LoadFlagsLowByte:       deps.loadFlagsLowByte,
			LoadNext:               deps.loadNext,
			Destroy:                deps.destroy,
			LoadCaster:             deps.loadCaster,
			LoadObjectFlags:        deps.loadObjectFlags,
			StoreCaster:            deps.storeCaster,
			LoadObj12:              deps.loadObj12,
			LoadObjectFlagsLowByte: deps.loadObjectFlagsLowByte,
			StoreObj12:             deps.storeObj12,
			LoadFlag20:             deps.loadFlag20,
			LoadObj24:              deps.loadObj24,
			StoreObj24:             deps.storeObj24,
			LoadFrame68:            deps.loadFrame68,
			LoadFrame60:            deps.loadFrame60,
			LoadCurrentFrame:       deps.loadCurrentFrame,
			LoadUpdate:             deps.loadUpdate,
			CallUpdate:             deps.callUpdate,
			Cancel:                 deps.cancel,
		},
	)
}

func spellDurationProcessServerDeps4FEEF0(
	sp *SpellsDuration,
	runtime SpellDurationProcessRuntime4FEEF0,
) spellDurationProcessNativeDeps4FEEF0 {
	return spellDurationProcessNativeDeps4FEEF0{
		loadFirst: sp.SpellDurationFirst4FE930,
		loadFlagsLowByte: func(record *DurSpell) byte {
			return byte(record.Flags88)
		},
		loadNext: SpellDurationNextNative4FE940,
		destroy:  runtime.Destroy,
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadObjectFlags: func(object *Object) uint32 {
			return uint32(object.ObjFlags)
		},
		storeCaster: func(record *DurSpell, object *Object) {
			record.Caster16 = object
		},
		loadObj12: func(record *DurSpell) *Object {
			return record.Obj12
		},
		loadObjectFlagsLowByte: func(object *Object) byte {
			return byte(object.ObjFlags)
		},
		storeObj12: func(record *DurSpell, object *Object) {
			record.Obj12 = object
		},
		loadFlag20: func(record *DurSpell) uint32 {
			return record.Flag20
		},
		loadObj24: func(record *DurSpell) *Object {
			return record.Obj24
		},
		storeObj24: func(record *DurSpell, object *Object) {
			record.Obj24 = object
		},
		loadFrame68: func(record *DurSpell) uint32 {
			return record.Frame68
		},
		loadFrame60: func(record *DurSpell) uint32 {
			return record.Frame60
		},
		loadCurrentFrame: func() uint32 {
			return sp.s.Frame()
		},
		loadUpdate: func(record *DurSpell) unsafe.Pointer {
			return record.Update
		},
		callUpdate: runtime.CallUpdate,
		cancel: func(record *DurSpell) {
			_ = sp.SpellDurationCancel4FE9D0(record)
		},
	}
}

// SpellDurationProcess4FEEF0 binds GAME.EXE 004FEEF0 to native-width
// *DurSpell links, *Object fields, and callback pointers. It composes the
// restored 004FE930/004FE940 accessors and 004FE9D0 cancellation boundary;
// destruction and callback invocation remain supplied by the outer runtime.
// All decoded callers are Go-owned, so no independent C/CGo ABI is retained.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationProcess4FEEF0(runtime SpellDurationProcessRuntime4FEEF0) {
	spellDurationProcessNative4FEEF0(spellDurationProcessServerDeps4FEEF0(sp, runtime))
}
