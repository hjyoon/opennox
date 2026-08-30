package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestNetworkInventoryFailContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 5)
	position := &types.Pointf{X: 11, Y: 22}
	hooks := networkInventoryFailHooks51BAD0[string]{
		loadCode: func() uint16 { events = append(events, "code"); return 0xabcd },
		findItem: func(unit string, code uint32) string {
			events = append(events, "find")
			if unit != "unit" || code != 0xabcd {
				t.Fatalf("find = (%q, %#x)", unit, code)
			}
			return "item"
		},
		loadPosition: func(unit string) *types.Pointf {
			events = append(events, "position")
			if unit != "unit" {
				t.Fatalf("position unit = %q", unit)
			}
			return position
		},
		drop: func(unit, item string, point *types.Pointf) {
			events = append(events, "drop")
			if unit != "unit" || item != "item" || point != position {
				t.Fatalf("drop = (%q, %q, %p)", unit, item, point)
			}
		},
		carryingHeavy: func(unit string) {
			events = append(events, "message")
			if unit != "unit" {
				t.Fatalf("message unit = %q", unit)
			}
		},
	}
	if got := networkInventoryFail51BAD0("unit", hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	want := []string{"code", "find", "position", "drop", "message"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkInventoryFailMissingItemConsumesPacket51BAD0(t *testing.T) {
	events := make([]string, 0, 2)
	hooks := networkInventoryFailHooks51BAD0[int]{
		loadCode:      func() uint16 { events = append(events, "code"); return 9 },
		findItem:      func(int, uint32) int { events = append(events, "find"); return 0 },
		loadPosition:  func(int) *types.Pointf { t.Fatal("missing item loaded position"); return nil },
		drop:          func(int, int, *types.Pointf) { t.Fatal("missing item dropped") },
		carryingHeavy: func(int) { t.Fatal("missing item reported capacity") },
	}
	if got := networkInventoryFail51BAD0(1, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	if want := []string{"code", "find"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkInventoryFailNativePointers51BAD0(t *testing.T) {
	unit := &Object{PosVec: types.Pointf{X: 7, Y: 13}}
	item := &Object{NetCode: 0xfedcba98}
	unit.InvFirstItem = item
	packet := &[NetworkInventoryFailPacketSize51BAD0]byte{0: 0xf1}
	binary.LittleEndian.PutUint16(packet[1:3], 0xba98)

	drops := 0
	messages := 0
	s := &Server{}
	got := s.NetworkInventoryFail51BAD0(unit, packet, NetworkInventoryFailRuntime51BAD0{
		Drop: func(gotUnit, gotItem *Object, point *types.Pointf) {
			drops++
			if gotUnit != unit || gotItem != item || point != &unit.PosVec {
				t.Fatalf("drop = (%p, %p, %p), want (%p, %p, %p)", gotUnit, gotItem, point, unit, item, &unit.PosVec)
			}
		},
		CarryingTooMuch: func(gotUnit *Object) {
			messages++
			if gotUnit != unit {
				t.Fatalf("message unit = %p, want %p", gotUnit, unit)
			}
		},
	})
	if got != 3 || drops != 0 || messages != 0 {
		t.Fatalf("full-code mismatch = (%d, drops %d, messages %d), want (3,0,0)", got, drops, messages)
	}

	item.NetCode = 0xba98
	got = s.NetworkInventoryFail51BAD0(unit, packet, NetworkInventoryFailRuntime51BAD0{
		Drop: func(gotUnit, gotItem *Object, point *types.Pointf) {
			drops++
			if gotUnit != unit || gotItem != item || point != &unit.PosVec {
				t.Fatalf("drop = (%p, %p, %p), want (%p, %p, %p)", gotUnit, gotItem, point, unit, item, &unit.PosVec)
			}
		},
		CarryingTooMuch: func(*Object) { messages++ },
	})
	if got != 3 || drops != 1 || messages != 1 {
		t.Fatalf("native result = (%d, drops %d, messages %d), want (3,1,1)", got, drops, messages)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0)) {
		t.Fatalf("item address %#x did not exercise the high native half", uintptr(unsafe.Pointer(item)))
	}
}
