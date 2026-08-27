package server

import (
	"reflect"
	"testing"
)

func TestNetworkTryDequipContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 12)
	hooks := networkTryDequipHooks51BAD0[string, int, string]{
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
		loadPlayer: func(update int) string {
			events = append(events, "player")
			if update != 7 {
				t.Fatalf("update = %d", update)
			}
			return "player"
		},
		loadPlayerStatus: func(player string) uint32 {
			events = append(events, "status")
			if player != "player" {
				t.Fatalf("player = %q", player)
			}
			return 0
		},
		findItemByCode: func(unit string, code uint32) string {
			events = append(events, "item")
			if unit != "unit" || code != 0x4567 {
				t.Fatalf("lookup = (%q, %#x)", unit, code)
			}
			return "item"
		},
		loadState:        func(int) uint8 { events = append(events, "state"); return 1 },
		loadItemClass:    func(string) uint32 { events = append(events, "class"); return 0x01000000 },
		loadItemSubclass: func(string) uint32 { events = append(events, "subclass"); return 0x800 },
		tryDequip: func(owner, item string) {
			events = append(events, "dequip")
			if owner != "unit" || item != "item" {
				t.Fatalf("dequip = (%q, %q)", owner, item)
			}
		},
	}
	if got := networkTryDequip51BAD0("unit", 7, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "player", "status", "item", "state", "class", "subclass", "dequip"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryDequipContractGatesAndShortCircuits51BAD0(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    uint32
		item      int
		state     uint8
		class     uint32
		subclass  uint32
		want      []string
		dequipped bool
	}{
		{name: "player status", status: 2, item: 9, want: []string{"wire", "dynamic", "debug", "player", "status"}},
		{name: "missing item", want: []string{"wire", "dynamic", "debug", "player", "status", "item"}},
		{name: "protected item", item: 9, state: 1, class: 0x01000000, subclass: 8, want: []string{"wire", "dynamic", "debug", "player", "status", "item", "state", "class", "subclass"}},
		{name: "other state", item: 9, state: 2, class: 0x01000000, subclass: 8, want: []string{"wire", "dynamic", "debug", "player", "status", "item", "state", "dequip"}, dequipped: true},
		{name: "other class", item: 9, state: 1, subclass: 8, want: []string{"wire", "dynamic", "debug", "player", "status", "item", "state", "class", "dequip"}, dequipped: true},
		{name: "subclass low byte", item: 9, state: 1, class: 0x01000000, subclass: 0x800, want: []string{"wire", "dynamic", "debug", "player", "status", "item", "state", "class", "subclass", "dequip"}, dequipped: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]string, 0, 11)
			calls := 0
			hooks := networkTryDequipHooks51BAD0[int, int, int]{
				loadWireCode:     func() uint16 { events = append(events, "wire"); return 9 },
				dynamicUnitCode:  func(uint16) uint32 { events = append(events, "dynamic"); return 9 },
				netDebug:         func() bool { events = append(events, "debug"); return false },
				testHighBit:      func(uint16) { t.Fatal("unexpected debug callback") },
				loadPlayer:       func(int) int { events = append(events, "player"); return 1 },
				loadPlayerStatus: func(int) uint32 { events = append(events, "status"); return tc.status },
				findItemByCode:   func(int, uint32) int { events = append(events, "item"); return tc.item },
				loadState:        func(int) uint8 { events = append(events, "state"); return tc.state },
				loadItemClass:    func(int) uint32 { events = append(events, "class"); return tc.class },
				loadItemSubclass: func(int) uint32 { events = append(events, "subclass"); return tc.subclass },
				tryDequip:        func(int, int) { events = append(events, "dequip"); calls++ },
			}
			if got := networkTryDequip51BAD0(1, 2, hooks); got != 3 {
				t.Fatalf("consumed = %d, want 3", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
			if got := calls != 0; got != tc.dequipped {
				t.Fatalf("dequip called = %t, want %t", got, tc.dequipped)
			}
		})
	}
}
