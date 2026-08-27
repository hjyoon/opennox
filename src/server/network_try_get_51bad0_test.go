package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestNetworkTryGetContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 20)
	next := map[string]string{"first": "second"}
	weights := map[string]uint8{"first": 3, "second": 4, "item": 5}
	hooks := networkTryGetHooks51BAD0[string, int, string]{
		loadWireCode: func() uint16 { events = append(events, "wire"); return 0x8123 },
		dynamicUnitCode: func(code uint16) uint32 {
			events = append(events, "dynamic")
			if code != 0x8123 {
				t.Fatalf("wire code = %#x", code)
			}
			return 0x4567
		},
		netDebug: func() bool { events = append(events, "debug"); return true },
		testHighBit: func(code uint16) {
			events = append(events, "high-bit")
			if code != 0x8123 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		gameBlocked: func() bool { events = append(events, "blocked"); return false },
		loadPlayer: func(update int) string {
			events = append(events, "player")
			if update != 7 {
				t.Fatalf("update = %d", update)
			}
			return "player"
		},
		loadPlayerStatus: func(string) uint32 { events = append(events, "status"); return 0 },
		loadTradeActive:  func(int) bool { events = append(events, "trade"); return false },
		loadDialogActive: func(int) bool { events = append(events, "dialog"); return false },
		loadUnitFlagsLow: func(unit string) uint8 {
			events = append(events, "flags")
			if unit != "unit" {
				t.Fatalf("unit = %q", unit)
			}
			return 0
		},
		objectFromNetCode: func(code uint32) string {
			events = append(events, "object")
			if code != 0x4567 {
				t.Fatalf("object code = %#x", code)
			}
			return "item"
		},
		loadInventoryFirst: func(string) string { events = append(events, "first"); return "first" },
		loadInventoryNext: func(item string) string {
			events = append(events, "next:"+item)
			return next[item]
		},
		loadWeight: func(item string) uint8 {
			events = append(events, "weight:"+item)
			return weights[item]
		},
		loadCarryCapacity: func(string) uint16 { events = append(events, "capacity"); return 12 },
		pickup: func(unit, item string) {
			events = append(events, "pickup")
			if unit != "unit" || item != "item" {
				t.Fatalf("pickup = (%q, %q)", unit, item)
			}
		},
		carryingTooMuch: func(string) { t.Fatal("unexpected carrying-too-much callback") },
	}
	if got := networkTryGet51BAD0("unit", 7, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	want := []string{
		"wire", "dynamic", "debug", "high-bit", "blocked", "player", "status", "trade", "dialog", "flags", "object",
		"first", "weight:first", "next:first", "weight:second", "next:second", "weight:item", "capacity", "pickup",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryGetContractBlockedBeforePointerLoads51BAD0(t *testing.T) {
	events := make([]string, 0, 5)
	hooks := networkTryGetHooks51BAD0[int, int, int]{
		loadWireCode:    func() uint16 { events = append(events, "wire"); return 9 },
		dynamicUnitCode: func(uint16) uint32 { events = append(events, "dynamic"); return 9 },
		netDebug:        func() bool { events = append(events, "debug"); return false },
		testHighBit:     func(uint16) { t.Fatal("unexpected high-bit callback") },
		gameBlocked:     func() bool { events = append(events, "blocked"); return true },
		loadPlayer:      func(int) int { t.Fatal("blocked game loaded player"); return 0 },
	}
	if got := networkTryGet51BAD0(1, 2, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	if want := []string{"wire", "dynamic", "debug", "blocked"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryGetNativePointersAndCapacity51BAD0(t *testing.T) {
	wantWeight := uintptr(488)
	wantCapacity := uintptr(490)
	wantHolder := uintptr(492)
	wantNext := uintptr(496)
	wantFirst := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantWeight = 516
		wantCapacity = 518
		wantHolder = 520
		wantNext = 528
		wantFirst = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.Weight", unsafe.Offsetof(Object{}.Weight), wantWeight},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), wantCapacity},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), wantHolder},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}

	const extent = uint32(0x123)
	first := &Object{Weight: 2}
	second := &Object{Weight: 3}
	first.InvNextItem = second
	item := &Object{NetCode: 0x4567, Extent: extent, Weight: 4}
	unit := &Object{InvFirstItem: first, CarryCapacity: 9}
	update := &PlayerUpdateData{Player: &Player{}}
	s := &Server{}
	s.Objs.List = item
	packet := &[NetworkTryGetPacketSize51BAD0]byte{0: 0x73}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	pickups := 0
	tooHeavy := 0
	got := s.NetworkTryGet51BAD0(unit, update, packet, NetworkTryGetRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		GameBlocked: func() bool { return false },
		Pickup: func(gotUnit, gotItem *Object) {
			pickups++
			if gotUnit != unit || gotItem != item {
				t.Fatalf("native pointers = (%p, %p), want (%p, %p)", gotUnit, gotItem, unit, item)
			}
		},
		CarryingTooMuch: func(*Object) { tooHeavy++ },
	})
	if got != 3 || pickups != 1 || tooHeavy != 0 {
		t.Fatalf("result = (%d, pickup %d, heavy %d), want (3,1,0)", got, pickups, tooHeavy)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0)) {
		t.Fatalf("item address %#x did not exercise the high native half", uintptr(unsafe.Pointer(item)))
	}

	unit.CarryCapacity = 8
	got = s.NetworkTryGet51BAD0(unit, update, packet, NetworkTryGetRuntime51BAD0{
		NetDebug:    func() bool { return false },
		TestHighBit: func(uint16) {},
		GameBlocked: func() bool { return false },
		Pickup:      func(*Object, *Object) { pickups++ },
		CarryingTooMuch: func(gotUnit *Object) {
			tooHeavy++
			if gotUnit != unit {
				t.Fatalf("heavy unit = %p", gotUnit)
			}
		},
	})
	if got != 3 || pickups != 1 || tooHeavy != 1 {
		t.Fatalf("heavy result = (%d, pickup %d, heavy %d), want (3,1,1)", got, pickups, tooHeavy)
	}
}

func TestNetworkTryGetNativeGates51BAD0(t *testing.T) {
	packet := &[NetworkTryGetPacketSize51BAD0]byte{0: 0x73}
	binary.LittleEndian.PutUint16(packet[1:3], 9)
	item := &Object{NetCode: 9, Weight: 1}
	tests := []struct {
		name    string
		blocked bool
		status  uint32
		trade   bool
		dialog  bool
		flags   object.Flags
	}{
		{name: "game blocked", blocked: true},
		{name: "player status", status: 1},
		{name: "trade", trade: true},
		{name: "dialog", dialog: true},
		{name: "no update", flags: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit := &Object{CarryCapacity: 1, ObjFlags: test.flags}
			update := &PlayerUpdateData{Player: &Player{Field3680: test.status}}
			if test.trade {
				update.Trade70 = &TradeSession{}
			}
			if test.dialog {
				update.DialogWith = &Object{}
			}
			s := &Server{}
			s.Objs.List = item
			pickups := 0
			got := s.NetworkTryGet51BAD0(unit, update, packet, NetworkTryGetRuntime51BAD0{
				NetDebug:        func() bool { return false },
				TestHighBit:     func(uint16) {},
				GameBlocked:     func() bool { return test.blocked },
				Pickup:          func(*Object, *Object) { pickups++ },
				CarryingTooMuch: func(*Object) { t.Fatal("unexpected heavy callback") },
			})
			if got != 3 || pickups != 0 {
				t.Fatalf("result = (%d, pickups %d), want (3,0)", got, pickups)
			}
		})
	}
}
