package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
)

func TestUnitBanishNativeLayout5017F0(t *testing.T) {
	wantSize := uintptr(780)
	wantType := uintptr(4)
	wantPosition := uintptr(56)
	wantNext := uintptr(496)
	wantFirst := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantType = 8
		wantPosition = 60
		wantNext = 528
		wantFirst = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.TypeInd size", unsafe.Sizeof(Object{}.TypeInd), 2},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitBanishNative5017F0PreservesPointersAndLivePosition(t *testing.T) {
	replacement := &Object{TypeInd: 8}
	second := &Object{TypeInd: 8}
	first := &Object{TypeInd: 7, InvNextItem: second}
	unit := &Object{InvFirstItem: first, PosVec: types.Ptf(12.5, -7.25)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"unit":        unsafe.Pointer(unit),
			"first":       unsafe.Pointer(first),
			"second":      unsafe.Pointer(second),
			"position":    unsafe.Pointer(&unit.PosVec),
			"replacement": unsafe.Pointer(replacement),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	cache := uint32(7)
	changedPosition := types.Ptf(101.75, -202.5)
	var deleted []*Object
	var fxCode uint8
	var fxPosition *types.Pointf
	unitBanishNative5017F0(unit, unitBanishNativeDeps5017F0{
		loadGlyphCache: func() uint32 { return cache },
		lookupType: func(string) uint32 {
			t.Fatal("nonzero cache unexpectedly resolved")
			return 0
		},
		storeGlyphCache: func(uint32) { t.Fatal("nonzero cache unexpectedly stored") },
		delayedDelete: func(object *Object) {
			deleted = append(deleted, object)
			switch object {
			case first:
				first.InvNextItem = replacement
				cache = 8
			case second:
				unit.PosVec = changedPosition
			}
		},
		sendPointFX: func(code uint8, position *types.Pointf) {
			fxCode = code
			fxPosition = position
		},
	})
	wantDeleted := []*Object{first, second, unit}
	if len(deleted) != len(wantDeleted) {
		t.Fatalf("deleted = %d objects, want %d", len(deleted), len(wantDeleted))
	}
	for index := range wantDeleted {
		if deleted[index] != wantDeleted[index] {
			t.Errorf("deleted[%d] = %p, want %p", index, deleted[index], wantDeleted[index])
		}
	}
	if fxCode != unitBanishBlueSparks5017F0 || fxPosition != &unit.PosVec || *fxPosition != changedPosition {
		t.Fatalf("FX = %#x/%p/%v, want %#x/%p/%v",
			fxCode, fxPosition, *fxPosition, unitBanishBlueSparks5017F0, &unit.PosVec, changedPosition)
	}
	if first.InvNextItem != replacement {
		t.Fatalf("first live successor = %p, want callback replacement %p", first.InvNextItem, replacement)
	}
	runtime.KeepAlive(replacement)
	runtime.KeepAlive(second)
	runtime.KeepAlive(first)
	runtime.KeepAlive(unit)
}

func TestUnitBanishServerDeps5017F0UseDedicatedFixedWidthCache(t *testing.T) {
	s := new(Server)
	s.Types.fast.glyph = 0x1234
	deps := unitBanishServerDeps5017F0(s, UnitBanishRuntime5017F0{})
	if got := deps.loadGlyphCache(); got != 0 {
		t.Fatalf("entry banish cache = %#x, want 0", got)
	}
	deps.storeGlyphCache(math.MaxUint32)
	if got := deps.loadGlyphCache(); got != math.MaxUint32 {
		t.Fatalf("banish cache = %#x, want %#x", got, uint32(math.MaxUint32))
	}
	if s.Types.fast.glyph != 0x1234 {
		t.Fatalf("duration-spell Glyph cache changed = %#x", s.Types.fast.glyph)
	}
	if unsafe.Sizeof(s.Types.fast.banishGlyph5017F0) != 4 {
		t.Fatalf("banish cache size = %d, want 4", unsafe.Sizeof(s.Types.fast.banishGlyph5017F0))
	}
}

func TestServerBanishUnit5017F0RoutesCallbacksWithoutDeathEvent(t *testing.T) {
	item := &Object{TypeInd: 7}
	unit := &Object{TypeInd: 99, InvFirstItem: item, PosVec: types.Ptf(300.5, 400.75)}
	s := new(Server)
	s.Types.fast.banishGlyph5017F0 = 7
	var deleted []*Object
	var fx netmsg.Op
	var position types.Pointf
	s.BanishUnit5017F0(unit, UnitBanishRuntime5017F0{
		DelayedDelete: func(object *Object) {
			deleted = append(deleted, object)
		},
		SendPointFX: func(code netmsg.Op, pos types.Pointf) {
			fx, position = code, pos
		},
	})
	if len(deleted) != 2 || deleted[0] != item || deleted[1] != unit {
		t.Fatalf("deleted = %v, want [%p %p]", deleted, item, unit)
	}
	if fx != netmsg.MSG_FX_BLUE_SPARKS || position != unit.PosVec {
		t.Fatalf("FX = %#x/%v, want %#x/%v", fx, position, netmsg.MSG_FX_BLUE_SPARKS, unit.PosVec)
	}
}
