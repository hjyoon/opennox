package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"
)

func TestNetworkTryEquipContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 10)
	hooks := networkTryEquipHooks51BAD0[string, int, string]{
		loadWireCode: func() uint16 { events = append(events, "wire"); return 0x8123 },
		dynamicUnitCode: func(code uint16) uint32 {
			events = append(events, "dynamic")
			if code != 0x8123 {
				t.Fatalf("wire code = %#x", code)
			}
			return 0x4567
		},
		netDebug:    func() bool { events = append(events, "debug"); return true },
		testHighBit: func(uint16) { events = append(events, "high-bit") },
		gameBlocked: func() bool { events = append(events, "blocked"); return false },
		loadPlayer: func(update int) string {
			events = append(events, "player")
			if update != 7 {
				t.Fatalf("update = %d", update)
			}
			return "player"
		},
		loadPlayerStatus: func(string) uint32 { events = append(events, "status"); return 0 },
		findItemByCode: func(unit string, code uint32) string {
			events = append(events, "item")
			if unit != "unit" || code != 0x4567 {
				t.Fatalf("lookup = (%q, %#x)", unit, code)
			}
			return "item"
		},
		tryEquip: func(owner, item string) {
			events = append(events, "equip")
			if owner != "unit" || item != "item" {
				t.Fatalf("equip = (%q, %q)", owner, item)
			}
		},
	}
	if got := networkTryEquip51BAD0("unit", 7, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "blocked", "player", "status", "item", "equip"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTryEquipContractGates51BAD0(t *testing.T) {
	for _, tc := range []struct {
		name    string
		blocked bool
		status  uint32
		want    []string
	}{
		{name: "game blocked", blocked: true, want: []string{"wire", "dynamic", "debug", "blocked"}},
		{name: "player status", status: 2, want: []string{"wire", "dynamic", "debug", "blocked", "player", "status"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]string, 0, 7)
			hooks := networkTryEquipHooks51BAD0[int, int, int]{
				loadWireCode:     func() uint16 { events = append(events, "wire"); return 9 },
				dynamicUnitCode:  func(uint16) uint32 { events = append(events, "dynamic"); return 9 },
				netDebug:         func() bool { events = append(events, "debug"); return false },
				testHighBit:      func(uint16) { t.Fatal("unexpected debug callback") },
				gameBlocked:      func() bool { events = append(events, "blocked"); return tc.blocked },
				loadPlayer:       func(int) int { events = append(events, "player"); return 1 },
				loadPlayerStatus: func(int) uint32 { events = append(events, "status"); return tc.status },
				findItemByCode:   func(int, uint32) int { t.Fatal("gated lookup"); return 0 },
				tryEquip:         func(int, int) { t.Fatal("gated equip") },
			}
			if got := networkTryEquip51BAD0(1, 2, hooks); got != 3 {
				t.Fatalf("consumed = %d, want 3", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestNetworkTryEquipNativePointers51BAD0(t *testing.T) {
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
	packet := &[NetworkTryEquipPacketSize51BAD0]byte{0: 0x75}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	equips := 0
	got := s.NetworkTryEquip51BAD0(unit, update, packet, NetworkTryEquipRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		GameBlocked: func() bool { return false },
		TryEquip: func(gotUnit, gotItem *Object) {
			equips++
			if gotUnit != unit || gotItem != item {
				t.Fatalf("native pointers = (%p, %p), want (%p, %p)", gotUnit, gotItem, unit, item)
			}
		},
	})
	if got != 3 || equips != 1 {
		t.Fatalf("result = (%d, equips %d), want (3,1)", got, equips)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0))) {
		t.Fatalf("test pointers did not exercise high native halves: unit=%p item=%p", unit, item)
	}
}
