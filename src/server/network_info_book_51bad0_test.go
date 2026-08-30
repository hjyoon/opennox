package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"
)

func TestNetworkInfoBookContractSearchAndDefaultOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 13)
	hooks := networkInfoBookHooks51BAD0[string, int]{
		loadWireCode: func() uint16 { events = append(events, "wire"); return 0x8123 },
		dynamicUnitCode: func(code uint16) uint32 {
			events = append(events, "dynamic")
			if code != 0x8123 {
				t.Fatalf("wire code = %#x", code)
			}
			return 0x456789ab
		},
		netDebug:    func() bool { events = append(events, "debug"); return true },
		testHighBit: func(uint16) { events = append(events, "high-bit") },
		findInventory: func(unit string, code uint32) string {
			events = append(events, "inventory")
			if unit != "unit" || code != 0x456789ab {
				t.Fatalf("inventory = (%q, %#x)", unit, code)
			}
			return ""
		},
		findTrade: func(update int, code uint32) string {
			events = append(events, "trade")
			if update != 7 || code != 0x456789ab {
				t.Fatalf("trade = (%d, %#x)", update, code)
			}
			return "item"
		},
		findWorld: func(uint32) string { t.Fatal("trade hit fell through to world"); return "" },
		unitCode: func(item string) uint16 {
			events = append(events, "unit-code")
			if item != "item" {
				t.Fatalf("unit-code item = %q", item)
			}
			return 0xbeef
		},
		loadKind: func() uint8 { events = append(events, "kind"); return 1 },
		loadDefaultInfo: func(item string) uint8 {
			events = append(events, "default")
			if item != "item" {
				t.Fatalf("default item = %q", item)
			}
			return 0xa7
		},
		loadGuideName: func(string) string { t.Fatal("default request loaded guide"); return "" },
		guideID:       func(string) uint8 { t.Fatal("default request resolved guide"); return 0 },
		loadRecipient: func(update int) uint8 {
			events = append(events, "recipient")
			if update != 7 {
				t.Fatalf("recipient update = %d", update)
			}
			return 3
		},
		send: func(recipient uint8, packet [4]byte) {
			events = append(events, "send")
			if recipient != 3 || packet != [4]byte{0xe2, 0xef, 0xbe, 0xa7} {
				t.Fatalf("send = (%d, %#v)", recipient, packet)
			}
		},
	}
	if got := networkInfoBook51BAD0("unit", 7, hooks); got != 4 {
		t.Fatalf("consumed = %d, want 4", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "inventory", "trade", "unit-code", "kind", "default", "recipient", "send"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkInfoBookContractGuideAndMissing51BAD0(t *testing.T) {
	worldCalls := 0
	guideCalls := 0
	sends := 0
	hooks := networkInfoBookHooks51BAD0[int, int]{
		loadWireCode:    func() uint16 { return 9 },
		dynamicUnitCode: func(uint16) uint32 { return 9 },
		netDebug:        func() bool { return false },
		testHighBit:     func(uint16) { t.Fatal("unexpected debug observation") },
		findInventory:   func(int, uint32) int { return 0 },
		findTrade:       func(int, uint32) int { return 0 },
		findWorld: func(code uint32) int {
			worldCalls++
			if code != 9 {
				t.Fatalf("world code = %d", code)
			}
			return 11
		},
		unitCode:        func(int) uint16 { return 0x1234 },
		loadKind:        func() uint8 { return 2 },
		loadDefaultInfo: func(int) uint8 { t.Fatal("guide request loaded default"); return 0 },
		loadGuideName: func(item int) string {
			guideCalls++
			if item != 11 {
				t.Fatalf("guide item = %d", item)
			}
			return "UrchinShaman"
		},
		guideID: func(name string) uint8 {
			if name != "UrchinShaman" {
				t.Fatalf("guide name = %q", name)
			}
			return 40
		},
		loadRecipient: func(int) uint8 { return 5 },
		send: func(recipient uint8, packet [4]byte) {
			sends++
			if recipient != 5 || packet != [4]byte{0xe2, 0x34, 0x12, 40} {
				t.Fatalf("send = (%d, %#v)", recipient, packet)
			}
		},
	}
	if got := networkInfoBook51BAD0(1, 2, hooks); got != 4 || worldCalls != 1 || guideCalls != 1 || sends != 1 {
		t.Fatalf("guide result = (%d, world %d, guide %d, sends %d)", got, worldCalls, guideCalls, sends)
	}

	hooks.findWorld = func(uint32) int { return 0 }
	hooks.unitCode = func(int) uint16 { t.Fatal("missing object loaded unit code"); return 0 }
	hooks.loadKind = func() uint8 { t.Fatal("missing object loaded kind"); return 0 }
	hooks.send = func(uint8, [4]byte) { t.Fatal("missing object sent response") }
	if got := networkInfoBook51BAD0(1, 2, hooks); got != 4 {
		t.Fatalf("missing consumed = %d, want 4", got)
	}
}

func TestNetworkInfoBookNativeTradeAndUseDataPointers51BAD0(t *testing.T) {
	const extent = uint32(0x321)
	unit := &Object{}
	guide := &FieldGuideUseData{}
	guide.SetCreature("UrchinShaman")
	item := &Object{NetCode: 0x4321, Extent: extent, UseData: UseDataPtr{Ptr: unsafe.Pointer(guide)}}
	other := &TradeItem{Item0: &Object{NetCode: 7}}
	node := &TradeItem{Item0: item}
	other.Field8 = node
	update := &PlayerUpdateData{Player: &Player{PlayerInd: 6}, Trade70: &TradeSession{Field20: other}}
	s := &Server{}
	s.Objs.List = item
	packet := &[NetworkInfoBookPacketSize51BAD0]byte{0: 0xe2, 3: 2}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	sends := 0
	got := s.NetworkInfoBook51BAD0(unit, update, packet, NetworkInfoBookRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		Send: func(recipient uint8, response [4]byte) {
			sends++
			if recipient != 6 || response != [4]byte{0xe2, 0x21, 0x43, 40} {
				t.Fatalf("send = (%d, %#v)", recipient, response)
			}
		},
	})
	if got != 4 || sends != 1 {
		t.Fatalf("result = (%d, sends %d), want (4,1)", got, sends)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0)) {
		t.Fatalf("item address %#x did not exercise the high native half", uintptr(unsafe.Pointer(item)))
	}
}
