package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"
)

func TestNetworkTryUseContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 9)
	hooks := networkTryUseHooks51BAD0[string, int, string]{
		loadWireCode: func() uint16 { events = append(events, "wire"); return 0x8123 },
		dynamicUnitCode: func(code uint16) uint32 {
			events = append(events, "dynamic")
			if code != 0x8123 {
				t.Fatalf("wire code = %#x", code)
			}
			return 0x456789ab
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
		loadPlayerStatus: func(player string) uint32 {
			events = append(events, "status")
			if player != "player" {
				t.Fatalf("player = %q", player)
			}
			return 0
		},
		findItemByCode: func(unit string, code uint32) string {
			events = append(events, "find")
			if unit != "unit" || code != 0x456789ab {
				t.Fatalf("find = (%q, %#x)", unit, code)
			}
			return "item"
		},
		useByNetCode: func(owner, item string) {
			events = append(events, "use")
			if owner != "unit" || item != "item" {
				t.Fatalf("use = (%q, %q)", owner, item)
			}
		},
	}
	if got := networkTryUse51BAD0("unit", 7, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "blocked", "player", "status", "find", "use"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryUseContractGates51BAD0(t *testing.T) {
	tests := []struct {
		name      string
		blocked   bool
		status    uint32
		wantEvent []string
	}{
		{name: "game blocked", blocked: true, wantEvent: []string{"wire", "dynamic", "debug", "blocked"}},
		{name: "player status", status: 3, wantEvent: []string{"wire", "dynamic", "debug", "blocked", "player", "status"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := make([]string, 0, 6)
			hooks := networkTryUseHooks51BAD0[int, int, int]{
				loadWireCode:     func() uint16 { events = append(events, "wire"); return 9 },
				dynamicUnitCode:  func(uint16) uint32 { events = append(events, "dynamic"); return 9 },
				netDebug:         func() bool { events = append(events, "debug"); return false },
				testHighBit:      func(uint16) { t.Fatal("unexpected high-bit callback") },
				gameBlocked:      func() bool { events = append(events, "blocked"); return test.blocked },
				loadPlayer:       func(int) int { events = append(events, "player"); return 1 },
				loadPlayerStatus: func(int) uint32 { events = append(events, "status"); return test.status },
				findItemByCode:   func(int, uint32) int { t.Fatal("gated request searched inventory"); return 0 },
				useByNetCode:     func(int, int) { t.Fatal("gated request invoked Use") },
			}
			if got := networkTryUse51BAD0(1, 2, hooks); got != 3 {
				t.Fatalf("consumed = %d, want 3", got)
			}
			if !reflect.DeepEqual(events, test.wantEvent) {
				t.Fatalf("events = %v, want %v", events, test.wantEvent)
			}
		})
	}
}

func TestNetworkTryUseNativePointers51BAD0(t *testing.T) {
	const extent = uint32(0x123)
	var token byte
	usePointer := unsafe.Pointer(&token)
	unit := &Object{}
	item := &Object{NetCode: 0x456789ab, Extent: extent, Use: UseFuncPtr{Ptr: usePointer}}
	unit.InvFirstItem = item
	update := &PlayerUpdateData{Player: &Player{}}
	s := &Server{}
	s.Objs.List = item
	packet := &[NetworkTryUsePacketSize51BAD0]byte{0: 0x74}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	calls := 0
	objUse.Register(usePointer, func(gotOwner, gotItem *Object) bool {
		calls++
		if gotOwner != unit || gotItem != item {
			t.Fatalf("Use args = (%p, %p), want (%p, %p)", gotOwner, gotItem, unit, item)
		}
		return false
	})
	got := s.NetworkTryUse51BAD0(unit, update, packet, NetworkTryUseRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		GameBlocked: func() bool { return false },
	})
	if got != 3 || calls != 1 {
		t.Fatalf("result = (%d, calls %d), want (3,1)", got, calls)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0)) {
		t.Fatalf("item address %#x did not exercise the high native half", uintptr(unsafe.Pointer(item)))
	}
}
