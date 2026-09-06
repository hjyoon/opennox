package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func requireSpellDurationCreateNativePointers4FEBA0(t *testing.T, values ...unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for i, value := range values {
		if value == nil || uintptr(value) <= math.MaxUint32 {
			t.Fatalf("pointer %d = %p, want native address above 4 GiB", i, value)
		}
	}
}

func TestSpellDurationCreate4FEBA0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectType := uintptr(4)
	wantObjectFlags := uintptr(16)
	wantObjectPos := uintptr(56)
	wantDurSize := uintptr(120)
	wantObj12 := uintptr(12)
	wantCaster16 := uintptr(16)
	wantFlag20 := uintptr(20)
	wantObj24 := uintptr(24)
	wantPos := uintptr(28)
	wantField36 := uintptr(36)
	wantTarget48 := uintptr(48)
	wantPos2 := uintptr(52)
	wantFrame60 := uintptr(60)
	wantFlags88 := uintptr(88)
	wantCreate := uintptr(92)
	wantUpdate := uintptr(96)
	wantDestroy := uintptr(100)
	wantSub104 := uintptr(104)
	wantSub108 := uintptr(108)
	wantAcceptPos := uintptr(4)
	wantAcceptSize := uintptr(12)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectType = 8
		wantObjectFlags = 20
		wantObjectPos = 60
		wantDurSize = 184
		wantObj12 = 16
		wantCaster16 = 24
		wantFlag20 = 32
		wantObj24 = 40
		wantPos = 48
		wantField36 = 56
		wantTarget48 = 72
		wantPos2 = 80
		wantFrame60 = 88
		wantFlags88 = 120
		wantCreate = 128
		wantUpdate = 136
		wantDestroy = 144
		wantSub104 = 152
		wantSub108 = 160
		wantAcceptPos = 8
		wantAcceptSize = 16
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantObjectType},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjectFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantObjectPos},
		{"DurSpell size", unsafe.Sizeof(DurSpell{}), wantDurSize},
		{"DurSpell.Spell", unsafe.Offsetof(DurSpell{}.Spell), 4},
		{"DurSpell.Level", unsafe.Offsetof(DurSpell{}.Level), 8},
		{"DurSpell.Obj12", unsafe.Offsetof(DurSpell{}.Obj12), wantObj12},
		{"DurSpell.Caster16", unsafe.Offsetof(DurSpell{}.Caster16), wantCaster16},
		{"DurSpell.Flag20", unsafe.Offsetof(DurSpell{}.Flag20), wantFlag20},
		{"DurSpell.Obj24", unsafe.Offsetof(DurSpell{}.Obj24), wantObj24},
		{"DurSpell.Pos", unsafe.Offsetof(DurSpell{}.Pos), wantPos},
		{"DurSpell.Field36", unsafe.Offsetof(DurSpell{}.Field36), wantField36},
		{"DurSpell.Target48", unsafe.Offsetof(DurSpell{}.Target48), wantTarget48},
		{"DurSpell.Pos2", unsafe.Offsetof(DurSpell{}.Pos2), wantPos2},
		{"DurSpell.Frame60", unsafe.Offsetof(DurSpell{}.Frame60), wantFrame60},
		{"DurSpell.Flags88", unsafe.Offsetof(DurSpell{}.Flags88), wantFlags88},
		{"DurSpell.Create", unsafe.Offsetof(DurSpell{}.Create), wantCreate},
		{"DurSpell.Update", unsafe.Offsetof(DurSpell{}.Update), wantUpdate},
		{"DurSpell.Destroy", unsafe.Offsetof(DurSpell{}.Destroy), wantDestroy},
		{"DurSpell.Sub104", unsafe.Offsetof(DurSpell{}.Sub104), wantSub104},
		{"DurSpell.Sub108", unsafe.Offsetof(DurSpell{}.Sub108), wantSub108},
		{"SpellAcceptArg size", unsafe.Sizeof(SpellAcceptArg{}), wantAcceptSize},
		{"SpellAcceptArg.Obj", unsafe.Offsetof(SpellAcceptArg{}.Obj), 0},
		{"SpellAcceptArg.Pos", unsafe.Offsetof(SpellAcceptArg{}.Pos), wantAcceptPos},
		{"Object pointer width", unsafe.Sizeof(DurSpell{}.Caster16), unsafe.Sizeof(uintptr(0))},
		{"callback pointer width", unsafe.Sizeof(DurSpell{}.Create), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSpellDurationCreateNative4FEBA0PreservesPointersWidthsAndWrites(t *testing.T) {
	second := &Object{PosVec: types.Pointf{X: -1, Y: -2}}
	third := &Object{PosVec: types.Pointf{X: 10.5, Y: 20.25}}
	fourth := &Object{TypeInd: 7, PosVec: types.Pointf{X: 30.75, Y: -40.5}}
	target := new(Object)
	arg := &SpellAcceptArg{Obj: target, Pos: types.Pointf{X: -50.25, Y: 60.5}}
	record := &DurSpell{
		Flags88: 0xa1b2c3d4,
		Sub104:  new(DurSpell),
		Sub108:  new(DurSpell),
	}
	createCell, updateCell, destroyCell := new(byte), new(byte), new(byte)
	create := unsafe.Pointer(createCell)
	update := unsafe.Pointer(updateCell)
	destroy := unsafe.Pointer(destroyCell)
	requireSpellDurationCreateNativePointers4FEBA0(t,
		unsafe.Pointer(second), unsafe.Pointer(third), unsafe.Pointer(fourth), unsafe.Pointer(target),
		unsafe.Pointer(arg), unsafe.Pointer(record), create, update, destroy,
	)

	glyph := uint32(0)
	frames := []uint32{math.MaxUint32 - 1, math.MaxUint32, 2}
	frameIndex := 0
	var events []string
	got := spellDurationCreateNative4FEBA0(
		-17,
		second,
		third,
		fourth,
		arg,
		math.MinInt32,
		create,
		update,
		destroy,
		math.MaxUint32,
		spellDurationCreateNativeDeps4FEBA0{
			loadGlyphCache: func() uint32 {
				events = append(events, "glyph-cache")
				return glyph
			},
			lookupGlyph: func() uint32 {
				events = append(events, "glyph-lookup")
				return 7
			},
			storeGlyphCache: func(value uint32) {
				events = append(events, "glyph-store")
				glyph = value
			},
			findDuplicate: func(int32, *Object) int32 {
				t.Fatal("unexpected duplicate probe")
				return 0
			},
			cancelFor: func(spellID int32, caster *Object) {
				events = append(events, "cancel-for")
				if spellID != -17 || caster != third {
					t.Fatalf("cancel args = %d/%p, want -17/%p", spellID, caster, third)
				}
			},
			beforeCreate: func() {
				events = append(events, "before-create")
			},
			newRecord: func() *DurSpell {
				events = append(events, "new-record")
				return record
			},
			loadFrame: func() uint32 {
				events = append(events, "frame")
				value := frames[frameIndex]
				frameIndex++
				return value
			},
			addRecord: func(value *DurSpell) {
				events = append(events, "add-record")
				if value != record {
					t.Fatalf("added record = %p, want %p", value, record)
				}
			},
			spellHasFlags: func(spellID int32, mask uint32) int32 {
				events = append(events, "spell-flags")
				if spellID != -17 || mask != 4 {
					t.Fatalf("flag args = %d/%#x, want -17/4", spellID, mask)
				}
				return -1
			},
			spellAudio: func(spellID, selector int32) int32 {
				events = append(events, "spell-audio")
				if spellID != -17 || selector != 1 {
					t.Fatalf("audio args = %d/%d, want -17/1", spellID, selector)
				}
				return -123
			},
			audioEvent: func(id int32, caster *Object, kind, code int32) {
				events = append(events, "audio-event")
				if id != -123 || caster != third || kind != 0 || code != 0 {
					t.Fatalf("audio event = %d/%p/%d/%d", id, caster, kind, code)
				}
			},
			callCreate: func(callback unsafe.Pointer, value *DurSpell) int32 {
				events = append(events, "call-create")
				if callback != create || value != record {
					t.Fatalf("create args = %p/%p, want %p/%p", callback, value, create, record)
				}
				return 0
			},
			cancelSpell: func(*DurSpell) {
				t.Fatal("successful create was cancelled")
			},
		},
	)
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantEvents := []string{
		"glyph-cache", "glyph-lookup", "glyph-store", "cancel-for", "before-create", "new-record",
		"glyph-cache", "frame", "frame", "frame", "add-record", "spell-flags", "spell-audio",
		"audio-event", "call-create",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if record.Spell != uint32(0xffffffef) || record.Level != uint32(0x80000000) {
		t.Fatalf("spell/level = %#x/%#x, want 0xffffffef/0x80000000", record.Spell, record.Level)
	}
	if record.Obj12 != second || record.Caster16 != third || record.Flag20 != 1 || record.Obj24 != fourth {
		t.Fatalf("record objects/mode = %p/%p/%d/%p", record.Obj12, record.Caster16, record.Flag20, record.Obj24)
	}
	if record.Pos != fourth.PosVec || record.Field36 != 0 || record.Target48 != target || record.Pos2 != arg.Pos {
		t.Fatalf("record positions/target = %#v/%#x/%p/%#v", record.Pos, record.Field36, record.Target48, record.Pos2)
	}
	if record.Create != create || record.Update != update || record.Destroy != destroy {
		t.Fatalf("record callbacks = %p/%p/%p, want %p/%p/%p",
			record.Create, record.Update, record.Destroy, create, update, destroy)
	}
	if record.Sub104 != nil || record.Sub108 != nil {
		t.Fatalf("record sublinks = %p/%p, want nil/nil", record.Sub104, record.Sub108)
	}
	if record.Frame60 != math.MaxUint32-1 || record.Frame64 != math.MaxUint32 || record.Frame68 != 1 {
		t.Fatalf("record frames = %#x/%#x/%#x", record.Frame60, record.Frame64, record.Frame68)
	}
	if record.Flags88 != 0xa1b2c300 {
		t.Fatalf("record flags = %#x, want low-byte clear 0xa1b2c300", record.Flags88)
	}
	runtime.KeepAlive(second)
	runtime.KeepAlive(third)
	runtime.KeepAlive(fourth)
	runtime.KeepAlive(target)
	runtime.KeepAlive(arg)
	runtime.KeepAlive(record)
	runtime.KeepAlive(createCell)
	runtime.KeepAlive(updateCell)
	runtime.KeepAlive(destroyCell)
}

func TestSpellDurationCreate4FEBA0ServerBinding(t *testing.T) {
	var srv Server
	srv.Types.byID = map[string]*ObjectType{
		"glyph": {ind: 7},
	}
	srv.Spells.byID = map[spell.ID]*SpellDef{
		31: {
			Def:       things.Spell{Flags: things.SpellTargeted},
			CastSound: sound.ID(321),
			OnSound:   sound.ID(654),
		},
	}
	srv.Spells.Dur.init(&srv)
	if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want 1", got)
	}
	t.Cleanup(srv.Spells.Dur.Free)
	srv.SetFrame(1234)
	stale := &DurSpell{Flags88: 0x12345601}
	srv.Spells.Dur.List = stale

	second := new(Object)
	third := &Object{ObjFlags: object.Flags(0), PosVec: types.Pointf{X: 11, Y: 12}}
	fourth := &Object{TypeInd: 7, PosVec: types.Pointf{X: 21, Y: 22}}
	target := new(Object)
	arg := &SpellAcceptArg{Obj: target, Pos: types.Pointf{X: 31, Y: 32}}
	createCell, updateCell, destroyCell := new(byte), new(byte), new(byte)
	create := unsafe.Pointer(createCell)
	update := unsafe.Pointer(updateCell)
	destroy := unsafe.Pointer(destroyCell)
	requireSpellDurationCreateNativePointers4FEBA0(t,
		unsafe.Pointer(second), unsafe.Pointer(third), unsafe.Pointer(fourth), unsafe.Pointer(target),
		unsafe.Pointer(arg), create, update, destroy,
	)

	var destroyed int
	var createRecord *DurSpell
	var audioID sound.ID
	var audioObject *Object
	got := srv.Spells.Dur.SpellDurationCreate4FEBA0(
		31,
		second,
		third,
		fourth,
		arg,
		-2,
		create,
		update,
		destroy,
		9,
		SpellDurationCreateRuntime4FEBA0{
			DestroySpell: func(record *DurSpell) {
				destroyed++
				if record != stale {
					t.Fatalf("destroy record = %p, want stale %p", record, stale)
				}
				srv.Spells.Dur.SpellDurationUnlink4FE900(record)
			},
			CallCreate: func(callback unsafe.Pointer, record *DurSpell) int32 {
				if callback != create {
					t.Fatalf("callback = %p, want %p", callback, create)
				}
				createRecord = record
				return 0
			},
			AudioEvent: func(id sound.ID, object *Object, kind int, code uint32) {
				if kind != 0 || code != 0 {
					t.Fatalf("audio kind/code = %d/%d, want 0/0", kind, code)
				}
				audioID = id
				audioObject = object
			},
		},
	)
	if got != 1 || destroyed != 1 {
		t.Fatalf("result/destroyed = %d/%d, want 1/1", got, destroyed)
	}
	record := srv.Spells.Dur.List
	if record == nil || createRecord != record {
		t.Fatalf("list/create record = %p/%p, want same non-nil record", record, createRecord)
	}
	requireSpellDurationCreateNativePointers4FEBA0(t, unsafe.Pointer(record))
	if srv.Types.fast.glyph != 7 {
		t.Fatalf("Glyph cache = %d, want lookup result 7", srv.Types.fast.glyph)
	}
	if record.Spell != 31 || record.Level != math.MaxUint32-1 || record.Obj12 != second || record.Caster16 != third {
		t.Fatalf("record scalar/object fields = %#x/%#x/%p/%p", record.Spell, record.Level, record.Obj12, record.Caster16)
	}
	if record.Flag20 != 1 || record.Obj24 != fourth || record.Pos != fourth.PosVec || record.Target48 != target || record.Pos2 != arg.Pos {
		t.Fatalf("record anchor/positions = %d/%p/%#v/%p/%#v", record.Flag20, record.Obj24, record.Pos, record.Target48, record.Pos2)
	}
	if record.Frame60 != 1234 || record.Frame64 != 1234 || record.Frame68 != 1243 || record.Flags88 != 0 {
		t.Fatalf("record frames/flags = %d/%d/%d/%#x", record.Frame60, record.Frame64, record.Frame68, record.Flags88)
	}
	if record.Create != create || record.Update != update || record.Destroy != destroy {
		t.Fatalf("record callbacks = %p/%p/%p", record.Create, record.Update, record.Destroy)
	}
	if audioID != 654 || audioObject != third {
		t.Fatalf("audio = %d/%p, want 654/%p", audioID, audioObject, third)
	}
	runtime.KeepAlive(second)
	runtime.KeepAlive(third)
	runtime.KeepAlive(fourth)
	runtime.KeepAlive(target)
	runtime.KeepAlive(arg)
	runtime.KeepAlive(createCell)
	runtime.KeepAlive(updateCell)
	runtime.KeepAlive(destroyCell)
}

func TestSpellDurationCreate4FEBA0ServerBindingAllowsNilCasterGlyph(t *testing.T) {
	var srv Server
	srv.Types.fast.glyph = 7
	srv.Spells.byID = map[spell.ID]*SpellDef{31: {}}
	srv.Spells.Dur.init(&srv)
	if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want 1", got)
	}
	t.Cleanup(srv.Spells.Dur.Free)

	fourth := &Object{TypeInd: 7, PosVec: types.Pointf{X: -7, Y: 8}}
	arg := &SpellAcceptArg{}
	var audioObject *Object
	audioCalls := 0
	got := srv.Spells.Dur.SpellDurationCreate4FEBA0(
		31, nil, nil, fourth, arg, 1, nil, nil, nil, 0,
		SpellDurationCreateRuntime4FEBA0{
			DestroySpell: func(*DurSpell) {
				t.Fatal("destroy callback ran for an empty duration list")
			},
			CallCreate: func(unsafe.Pointer, *DurSpell) int32 {
				t.Fatal("nil callback was invoked")
				return 0
			},
			AudioEvent: func(_ sound.ID, object *Object, _ int, _ uint32) {
				audioCalls++
				audioObject = object
			},
		},
	)
	if got != 1 || srv.Spells.Dur.List == nil {
		t.Fatalf("result/list = %d/%p, want 1/non-nil", got, srv.Spells.Dur.List)
	}
	if record := srv.Spells.Dur.List; record.Caster16 != nil || record.Obj24 != fourth || record.Pos != fourth.PosVec {
		t.Fatalf("nil-caster Glyph record = caster %p, anchor %p, pos %#v", record.Caster16, record.Obj24, record.Pos)
	}
	if audioCalls != 1 || audioObject != nil {
		t.Fatalf("audio calls/caster = %d/%p, want 1/nil", audioCalls, audioObject)
	}
	runtime.KeepAlive(fourth)
	runtime.KeepAlive(arg)
}
