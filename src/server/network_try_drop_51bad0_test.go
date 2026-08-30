package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestNetworkTryDropContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 14)
	unit := "unit"
	update := 7
	hooks := networkTryDropHooks51BAD0[string, int, string]{
		loadWireCode: func() uint16 {
			events = append(events, "wire")
			return 0x8123
		},
		dynamicUnitCode: func(code uint16) uint32 {
			events = append(events, "dynamic")
			if code != 0x8123 {
				t.Fatalf("wire code = %#x, want 0x8123", code)
			}
			return 0x4567
		},
		netDebug: func() bool {
			events = append(events, "debug")
			return true
		},
		testHighBit: func(code uint16) {
			events = append(events, "high-bit")
			if code != 0x8123 {
				t.Fatalf("debug code = %#x, want 0x8123", code)
			}
		},
		loadPlayer: func(got int) string {
			events = append(events, "player")
			if got != update {
				t.Fatalf("update = %d, want %d", got, update)
			}
			return "player"
		},
		loadPlayerStatus: func(player string) uint32 {
			events = append(events, "status")
			return 0
		},
		loadTradeActive: func(int) bool {
			events = append(events, "trade")
			return false
		},
		loadDialogActive: func(int) bool {
			events = append(events, "dialog")
			return false
		},
		loadUnitFlagsLow: func(got string) uint8 {
			events = append(events, "flags")
			if got != unit {
				t.Fatalf("unit = %q, want %q", got, unit)
			}
			return 0
		},
		findItemByCode: func(got string, code uint32) string {
			events = append(events, "item")
			if got != unit || code != 0x4567 {
				t.Fatalf("item lookup = (%q, %#x), want (%q, 0x4567)", got, code, unit)
			}
			return "item"
		},
		loadDestinationX: func() uint16 {
			events = append(events, "x")
			return 0x1234
		},
		loadDestinationY: func() uint16 {
			events = append(events, "y")
			return 0xabcd
		},
		drop: func(gotUnit, item string, point *types.Pointf) {
			events = append(events, "drop")
			if gotUnit != unit || item != "item" || *point != (types.Pointf{X: 0x1234, Y: 0xabcd}) {
				t.Fatalf("drop = (%q, %q, %+v)", gotUnit, item, *point)
			}
		},
	}
	if got := networkTryDrop51BAD0(unit, update, hooks); got != 7 {
		t.Fatalf("consumed = %d, want 7", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "player", "status", "trade", "dialog", "flags", "item", "x", "y", "drop"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryDropNativePointersAndLayout51BAD0(t *testing.T) {
	wantObjectFlags := uintptr(16)
	wantObjectNetCode := uintptr(36)
	wantInventoryNext := uintptr(496)
	wantInventoryFirst := uintptr(504)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantTrade := uintptr(280)
	wantDialog := uintptr(284)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectFlags = 20
		wantObjectNetCode = 40
		wantInventoryNext = 528
		wantInventoryFirst = 544
		wantUpdateSize = 656
		wantPlayer = 336
		wantTrade = 344
		wantDialog = 352
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjectFlags},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantObjectNetCode},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantInventoryNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantInventoryFirst},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.Trade70", unsafe.Offsetof(PlayerUpdateData{}.Trade70), wantTrade},
		{"PlayerUpdateData.DialogWith", unsafe.Offsetof(PlayerUpdateData{}.DialogWith), wantDialog},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}

	const (
		extent  = uint32(0x123)
		netCode = uint32(0x4567)
	)
	item := &Object{NetCode: netCode}
	unit := &Object{InvFirstItem: item}
	update := &PlayerUpdateData{Player: &Player{}}
	extentObject := &Object{Extent: extent, NetCode: netCode}
	s := &Server{}
	s.Objs.List = extentObject
	packet := &[NetworkTryDropPacketSize51BAD0]byte{0: 0x72}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)
	binary.LittleEndian.PutUint16(packet[3:5], 321)
	binary.LittleEndian.PutUint16(packet[5:7], 654)

	debugCalls := 0
	dropCalls := 0
	got := s.NetworkTryDrop51BAD0(unit, update, packet, NetworkTryDropRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			debugCalls++
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		Drop: func(gotUnit, gotItem *Object, point *types.Pointf) {
			dropCalls++
			if gotUnit != unit || gotItem != item {
				t.Fatalf("native pointer round trip = (%p, %p), want (%p, %p)", gotUnit, gotItem, unit, item)
			}
			if *point != (types.Pointf{X: 321, Y: 654}) {
				t.Fatalf("point = %+v, want (321,654)", *point)
			}
		},
	})
	if got != 7 || debugCalls != 1 || dropCalls != 1 {
		t.Fatalf("result = (%d, debug %d, drop %d), want (7,1,1)", got, debugCalls, dropCalls)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0)) {
		t.Fatalf("test item address %#x did not exercise the high native half", uintptr(unsafe.Pointer(item)))
	}
}

func TestNetworkTryDropNativeGates51BAD0(t *testing.T) {
	packet := &[NetworkTryDropPacketSize51BAD0]byte{0: 0x72}
	binary.LittleEndian.PutUint16(packet[1:3], 9)
	item := &Object{NetCode: 9}
	tests := []struct {
		name   string
		status uint32
		trade  bool
		dialog bool
		flags  uint32
	}{
		{name: "player status", status: 1},
		{name: "trade", trade: true},
		{name: "dialog", dialog: true},
		{name: "no update", flags: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit := &Object{InvFirstItem: item, ObjFlags: object.Flags(test.flags)}
			update := &PlayerUpdateData{Player: &Player{Field3680: test.status}}
			if test.trade {
				update.Trade70 = &TradeSession{}
			}
			if test.dialog {
				update.DialogWith = &Object{}
			}
			drops := 0
			got := (&Server{}).NetworkTryDrop51BAD0(unit, update, packet, NetworkTryDropRuntime51BAD0{
				NetDebug:    func() bool { return false },
				TestHighBit: func(uint16) {},
				Drop:        func(*Object, *Object, *types.Pointf) { drops++ },
			})
			if got != 7 || drops != 0 {
				t.Fatalf("result = (%d, drops %d), want (7,0)", got, drops)
			}
		})
	}
}
