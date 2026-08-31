package server

import (
	"reflect"
	"testing"
)

func TestNetworkTryCollideContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 12)
	hooks := networkTryCollideHooks51BAD0[string, int, string, int]{
		loadWireCode: func() uint16 {
			events = append(events, "wire")
			return 0x8123
		},
		dynamicUnitCode: func(code uint16) uint32 {
			events = append(events, "dynamic")
			if code != 0x8123 {
				t.Fatalf("wire code = %#x", code)
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
		loadTradeActive: func(update int) bool {
			events = append(events, "trade")
			return update != 7
		},
		loadDialogActive: func(update int) bool {
			events = append(events, "dialog")
			return update != 7
		},
		objectFromNetCode: func(code uint32) string {
			events = append(events, "target")
			if code != 0x4567 {
				t.Fatalf("lookup code = %#x", code)
			}
			return "target"
		},
		loadCollide: func(target string) int {
			events = append(events, "callback")
			if target != "target" {
				t.Fatalf("callback target = %q", target)
			}
			return 99
		},
		callCollide: func(callback int, target, unit string) {
			events = append(events, "collide")
			if callback != 99 || target != "target" || unit != "unit" {
				t.Fatalf("collide = (%d, %q, %q)", callback, target, unit)
			}
		},
	}
	if got := networkTryCollide51BAD0("unit", 7, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "player", "status", "trade", "dialog", "target", "callback", "collide"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryCollideContractGatesAndShortCircuits51BAD0(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   uint32
		trade    bool
		dialog   bool
		target   int
		callback int
		want     []string
		called   bool
	}{
		{name: "player status", status: 2, target: 8, callback: 9, want: []string{"wire", "dynamic", "debug", "player", "status"}},
		{name: "trade", trade: true, target: 8, callback: 9, want: []string{"wire", "dynamic", "debug", "player", "status", "trade"}},
		{name: "dialog", dialog: true, target: 8, callback: 9, want: []string{"wire", "dynamic", "debug", "player", "status", "trade", "dialog"}},
		{name: "missing target", callback: 9, want: []string{"wire", "dynamic", "debug", "player", "status", "trade", "dialog", "target"}},
		{name: "missing callback", target: 8, want: []string{"wire", "dynamic", "debug", "player", "status", "trade", "dialog", "target", "callback"}},
		{name: "collide", target: 8, callback: 9, want: []string{"wire", "dynamic", "debug", "player", "status", "trade", "dialog", "target", "callback", "collide"}, called: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]string, 0, 10)
			calls := 0
			hooks := networkTryCollideHooks51BAD0[int, int, int, int]{
				loadWireCode:      func() uint16 { events = append(events, "wire"); return 8 },
				dynamicUnitCode:   func(uint16) uint32 { events = append(events, "dynamic"); return 8 },
				netDebug:          func() bool { events = append(events, "debug"); return false },
				testHighBit:       func(uint16) { t.Fatal("unexpected debug callback") },
				loadPlayer:        func(int) int { events = append(events, "player"); return 1 },
				loadPlayerStatus:  func(int) uint32 { events = append(events, "status"); return tc.status },
				loadTradeActive:   func(int) bool { events = append(events, "trade"); return tc.trade },
				loadDialogActive:  func(int) bool { events = append(events, "dialog"); return tc.dialog },
				objectFromNetCode: func(uint32) int { events = append(events, "target"); return tc.target },
				loadCollide:       func(int) int { events = append(events, "callback"); return tc.callback },
				callCollide:       func(int, int, int) { events = append(events, "collide"); calls++ },
			}
			if got := networkTryCollide51BAD0(3, 2, hooks); got != 3 {
				t.Fatalf("consumed = %d, want 3", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
			if got := calls != 0; got != tc.called {
				t.Fatalf("collide called = %t, want %t", got, tc.called)
			}
		})
	}
}
