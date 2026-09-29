package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"
)

func TestNetworkTryCreatureCommandContractOrder51BAD0(t *testing.T) {
	const (
		owner    = uint64(0x7f11223344556677)
		creature = uint64(0x7f88776655443322)
	)
	events := make([]string, 0, 10)
	hooks := networkTryCreatureCommandHooks51BAD0[uint64, int, string]{
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
		objectFromNetCode: func(code uint32) uint64 {
			events = append(events, "creature")
			if code != 0x4567 {
				t.Fatalf("lookup code = %#x", code)
			}
			return creature
		},
		loadOrder: func() uint8 {
			events = append(events, "order")
			return 5
		},
		orderUnit: func(gotOwner, gotCreature uint64, order uint32) {
			events = append(events, "dispatch")
			if gotOwner != owner || gotCreature != creature || order != 5 {
				t.Fatalf("dispatch = (%#x, %#x, %d)", gotOwner, gotCreature, order)
			}
		},
	}
	if got := networkTryCreatureCommand51BAD0(owner, 7, hooks); got != 4 {
		t.Fatalf("consumed = %d, want 4", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "player", "status", "creature", "order", "dispatch"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryCreatureCommandGatesAndBroadcast51BAD0(t *testing.T) {
	for _, tc := range []struct {
		name     string
		wireCode uint16
		status   uint32
		creature uint64
		want     []string
		called   bool
		wantUnit uint64
	}{
		{name: "player blocked", wireCode: 8, status: 1, creature: 9, want: []string{"wire", "dynamic", "debug", "player", "status"}},
		{name: "missing creature", wireCode: 8, want: []string{"wire", "dynamic", "debug", "player", "status", "creature"}},
		{name: "specific creature", wireCode: 8, creature: 9, want: []string{"wire", "dynamic", "debug", "player", "status", "creature", "order", "dispatch"}, called: true, wantUnit: 9},
		{name: "all creatures", wireCode: 0, creature: 9, want: []string{"wire", "dynamic", "debug", "player", "status", "order", "dispatch"}, called: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]string, 0, 9)
			calls := 0
			hooks := networkTryCreatureCommandHooks51BAD0[uint64, int, int]{
				loadWireCode:      func() uint16 { events = append(events, "wire"); return tc.wireCode },
				dynamicUnitCode:   func(uint16) uint32 { events = append(events, "dynamic"); return 8 },
				netDebug:          func() bool { events = append(events, "debug"); return false },
				testHighBit:       func(uint16) { t.Fatal("unexpected debug callback") },
				loadPlayer:        func(int) int { events = append(events, "player"); return 1 },
				loadPlayerStatus:  func(int) uint32 { events = append(events, "status"); return tc.status },
				objectFromNetCode: func(uint32) uint64 { events = append(events, "creature"); return tc.creature },
				loadOrder:         func() uint8 { events = append(events, "order"); return 3 },
				orderUnit: func(owner, creature uint64, order uint32) {
					events = append(events, "dispatch")
					calls++
					if owner != 7 || creature != tc.wantUnit || order != 3 {
						t.Fatalf("dispatch = (%d, %d, %d), want (7, %d, 3)", owner, creature, order, tc.wantUnit)
					}
				},
			}
			if got := networkTryCreatureCommand51BAD0(uint64(7), 2, hooks); got != 4 {
				t.Fatalf("consumed = %d, want 4", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
			if got := calls != 0; got != tc.called {
				t.Fatalf("called = %t, want %t", got, tc.called)
			}
		})
	}
}

func TestNetworkTryCreatureCommandNativePointers51BAD0(t *testing.T) {
	const (
		extent  = uint32(0x123)
		netCode = uint32(0x4567)
	)
	owner := &Object{}
	creature := &Object{Extent: extent, NetCode: netCode}
	update := &PlayerUpdateData{Player: &Player{}}
	s := &Server{}
	s.Objs.List = creature
	packet := &[NetworkTryCreatureCommandPacketSize51BAD0]byte{0: 0x78, 3: 5}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	calls := 0
	got := s.NetworkTryCreatureCommand51BAD0(owner, update, packet, NetworkTryCreatureCommandRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		OrderUnit: func(gotOwner, gotCreature *Object, order uint32) {
			calls++
			if gotOwner != owner || gotCreature != creature || order != 5 {
				t.Fatalf("native dispatch = (%p, %p, %d), want (%p, %p, 5)", gotOwner, gotCreature, order, owner, creature)
			}
		},
	})
	if got != 4 || calls != 1 {
		t.Fatalf("result = (%d, calls %d), want (4,1)", got, calls)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(creature)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(update)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(update.Player)) <= uintptr(^uint32(0))) {
		t.Fatalf("test pointers did not exercise high native halves: owner=%p creature=%p update=%p player=%p", owner, creature, update, update.Player)
	}
}
