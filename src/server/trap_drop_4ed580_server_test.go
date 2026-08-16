package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func defaultTrapDropNativeDeps4ED580() trapDropNativeDeps4ED580 {
	return trapDropNativeDeps4ED580{
		mapTile:     func(*types.Pointf) int32 { return 0 },
		defaultDrop: func(*Object, *Object, *types.Pointf) int32 { return 1 },
		audio:       func(uint32, *Object, int32, uint32) {},
		setOwner:    func(*Object, *Object) {},
	}
}

func TestTrapDrop4ED580NativeLayout(t *testing.T) {
	wantNetCode := uintptr(36)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantNetCode = 40
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
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

func TestTrapDropNative4ED580ForbiddenTileLoadsNetCodeAfterMap(t *testing.T) {
	owner := &Object{NetCode: 0x11111111}
	glyph := &Object{}
	point := &types.Pointf{X: 3, Y: 4}
	events := make([]string, 0, 2)
	deps := defaultTrapDropNativeDeps4ED580()
	deps.mapTile = func(got *types.Pointf) int32 {
		events = append(events, "map")
		if got != point {
			t.Fatalf("point = %p, want %p", got, point)
		}
		owner.NetCode = 0xfedcba98
		return -1
	}
	deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 {
		t.Fatal("forbidden tile called DefaultDrop")
		return 0
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 925 || gotOwner != owner || kind != 2 || code != 0xfedcba98 {
			t.Fatalf("audio = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	deps.setOwner = func(*Object, *Object) { t.Fatal("forbidden tile set owner") }
	if got := trapDropNative4ED580(owner, glyph, point, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(events, []string{"map", "audio"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestTrapDropNative4ED580DefaultRejectSkipsOwnerTransfer(t *testing.T) {
	owner := &Object{}
	glyph := &Object{}
	point := &types.Pointf{}
	deps := defaultTrapDropNativeDeps4ED580()
	deps.defaultDrop = func(gotOwner, gotGlyph *Object, gotPoint *types.Pointf) int32 {
		if gotOwner != owner || gotGlyph != glyph || gotPoint != point {
			t.Fatalf("default args = %p/%p/%p", gotOwner, gotGlyph, gotPoint)
		}
		return 0
	}
	deps.setOwner = func(*Object, *Object) { t.Fatal("rejected DefaultDrop set owner") }
	if got := trapDropNative4ED580(owner, glyph, point, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestTrapDropNative4ED580SuccessUsesNativeOwnerPointers(t *testing.T) {
	owner := &Object{}
	glyph := &Object{}
	point := &types.Pointf{X: -7, Y: 9}
	events := make([]string, 0, 2)
	deps := defaultTrapDropNativeDeps4ED580()
	deps.defaultDrop = func(gotOwner, gotGlyph *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotGlyph != glyph || gotPoint != point {
			t.Fatalf("default args = %p/%p/%p", gotOwner, gotGlyph, gotPoint)
		}
		return -1
	}
	deps.setOwner = func(gotOwner, gotGlyph *Object) {
		events = append(events, "owner")
		if gotOwner != owner || gotGlyph != glyph {
			t.Fatalf("owner args = %p/%p", gotOwner, gotGlyph)
		}
	}
	if got := trapDropNative4ED580(owner, glyph, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(events, []string{"default", "owner"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestTrapDrop4ED580ServerBindingSetsOwner(t *testing.T) {
	s := &Server{}
	oldOwner := &Object{}
	owner := &Object{}
	glyph := &Object{ObjOwner: oldOwner}
	oldOwner.Field129 = glyph
	runtime := TrapDropRuntime4ED580{
		MapTileAllowTeleport: func(*types.Pointf) int32 { return 0 },
		DefaultDrop:          func(*Object, *Object, *types.Pointf) int32 { return 1 },
	}
	if got := s.TrapDrop4ED580(owner, glyph, &types.Pointf{}, runtime); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if glyph.ObjOwner != owner || owner.Field129 != glyph || oldOwner.Field129 != nil {
		t.Fatalf("ownership = glyph owner %p, new first %p, old first %p", glyph.ObjOwner, owner.Field129, oldOwner.Field129)
	}
}
