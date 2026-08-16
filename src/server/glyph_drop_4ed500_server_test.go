package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func defaultGlyphDropNativeDeps4ED500() glyphDropNativeDeps4ED500 {
	return glyphDropNativeDeps4ED500{
		dropTrap: func(*Object, *Object, *types.Pointf) int32 { return 1 },
		audio:    func(uint32, *Object, int32, uint32) {},
	}
}

func TestGlyphDrop4ED500NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectPos := uintptr(56)
	wantDirection1 := uintptr(124)
	wantDirection2 := uintptr(126)
	wantObjectInitData := uintptr(692)
	wantGlyphSize := uintptr(36)
	wantGlyphPoint := uintptr(28)
	wantSpellArgSize := uintptr(12)
	wantSpellArgPoint := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectPos = 60
		wantDirection1 = 128
		wantDirection2 = 130
		wantObjectInitData = 760
		wantGlyphSize = 40
		wantGlyphPoint = 32
		wantSpellArgSize = 16
		wantSpellArgPoint = 8
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantObjectPos},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection1},
		{"Object.Direction2", unsafe.Offsetof(Object{}.Direction2), wantDirection2},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantObjectInitData},
		{"GlyphInitData size", unsafe.Sizeof(GlyphInitData{}), wantGlyphSize},
		{"GlyphInitData.Spells", unsafe.Offsetof(GlyphInitData{}.Spells), 0},
		{"GlyphInitData.SpellsCnt", unsafe.Offsetof(GlyphInitData{}.SpellsCnt), 20},
		{"GlyphInitData.SpellArg", unsafe.Offsetof(GlyphInitData{}.SpellArg), 24},
		{"GlyphInitData point", unsafe.Offsetof(GlyphInitData{}.SpellArg) + unsafe.Offsetof(SpellAcceptArg{}.Pos), wantGlyphPoint},
		{"SpellAcceptArg size", unsafe.Sizeof(SpellAcceptArg{}), wantSpellArgSize},
		{"SpellAcceptArg.Pos", unsafe.Offsetof(SpellAcceptArg{}.Pos), wantSpellArgPoint},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestGlyphDropNative4ED500TrapRejectsWithoutDereference(t *testing.T) {
	deps := defaultGlyphDropNativeDeps4ED500()
	deps.dropTrap = func(owner, glyph *Object, point *types.Pointf) int32 {
		if owner != nil || glyph != nil || point != nil {
			t.Fatalf("trap args = %p/%p/%p", owner, glyph, point)
		}
		return 0
	}
	deps.audio = func(uint32, *Object, int32, uint32) {
		t.Fatal("rejected drop played audio")
	}
	if got := glyphDropNative4ED500(nil, nil, nil, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestGlyphDropNative4ED500UsesPostTrapInitDataAndNativeFields(t *testing.T) {
	oldData := &GlyphInitData{SpellsCnt: 0x11111111}
	newData := &GlyphInitData{SpellsCnt: 0x22222222}
	owner := &Object{PosVec: types.Ptf(14, -3)}
	glyph := &Object{
		InitData:   unsafe.Pointer(oldData),
		Direction1: Dir16(0xaaaa),
		Direction2: Dir16(0xbbbb),
	}
	point := &types.Pointf{X: 4, Y: 5}
	events := make([]string, 0, 2)
	deps := defaultGlyphDropNativeDeps4ED500()
	deps.dropTrap = func(gotOwner, gotGlyph *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "trap")
		if gotOwner != owner || gotGlyph != glyph || gotPoint != point {
			t.Fatalf("trap args = %p/%p/%p", gotOwner, gotGlyph, gotPoint)
		}
		glyph.InitData = unsafe.Pointer(newData)
		return -1
	}
	deps.audio = func(id uint32, gotGlyph *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 825 || gotGlyph != glyph || kind != 0 || code != 0 {
			t.Fatalf("audio = %d/%p/%d/%d", id, gotGlyph, kind, code)
		}
	}
	if got := glyphDropNative4ED500(owner, glyph, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(events, []string{"trap", "audio"}) {
		t.Fatalf("events = %v", events)
	}
	if oldData.SpellArg.Pos != (types.Pointf{}) || oldData.SpellsCnt != 0x11111111 {
		t.Fatalf("old data changed = %#v", oldData)
	}
	if newData.SpellArg.Pos != *point || newData.SpellsCnt != 0x22222222 {
		t.Fatalf("new data = %#v", newData)
	}
	wantDirection := Dir16(uint16(directionFromVector509ED0(10, -8)))
	if glyph.Direction2 != wantDirection || glyph.Direction1 != wantDirection {
		t.Fatalf("directions = %d/%d, want %d", glyph.Direction2, glyph.Direction1, wantDirection)
	}
}

func TestGlyphDropNative4ED500NilInitDataFaultsAfterTrap(t *testing.T) {
	owner := &Object{}
	glyph := &Object{}
	point := &types.Pointf{}
	deps := defaultGlyphDropNativeDeps4ED500()
	trapCalled := false
	deps.dropTrap = func(*Object, *Object, *types.Pointf) int32 {
		trapCalled = true
		return 1
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil GlyphInitData did not fault")
		}
		if !trapCalled {
			t.Fatal("GlyphInitData fault preceded TrapDrop")
		}
	}()
	glyphDropNative4ED500(owner, glyph, point, deps)
}
