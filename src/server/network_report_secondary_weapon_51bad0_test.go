package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"
)

func TestNetworkReportSecondaryWeaponContractOrder51BAD0(t *testing.T) {
	const (
		owner = uint64(0x7f11223344556677)
		item  = uint64(0x7f88776655443322)
	)
	events := make([]string, 0, 7)
	hooks := networkReportSecondaryWeaponHooks51BAD0[uint64]{
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
		objectFromNetCode: func(code uint32) uint64 {
			events = append(events, "item")
			if code != 0x4567 {
				t.Fatalf("lookup code = %#x", code)
			}
			return item
		},
		report: func(gotOwner, gotItem uint64) {
			events = append(events, "report")
			if gotOwner != owner || gotItem != item {
				t.Fatalf("report = (%#x, %#x), want (%#x, %#x)", gotOwner, gotItem, owner, item)
			}
		},
	}
	if got := networkReportSecondaryWeapon51BAD0(owner, hooks); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	want := []string{"wire", "dynamic", "debug", "high-bit", "item", "report"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkReportSecondaryWeaponZeroAndMissingItem51BAD0(t *testing.T) {
	for _, tc := range []struct {
		name       string
		wireCode   uint16
		resolved   uint64
		wantEvents []string
	}{
		{name: "zero code", wantEvents: []string{"wire", "dynamic", "debug", "report"}},
		{name: "missing item", wireCode: 7, wantEvents: []string{"wire", "dynamic", "debug", "item", "report"}},
		{name: "resolved item", wireCode: 7, resolved: 9, wantEvents: []string{"wire", "dynamic", "debug", "item", "report"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]string, 0, 5)
			hooks := networkReportSecondaryWeaponHooks51BAD0[uint64]{
				loadWireCode:      func() uint16 { events = append(events, "wire"); return tc.wireCode },
				dynamicUnitCode:   func(uint16) uint32 { events = append(events, "dynamic"); return 7 },
				netDebug:          func() bool { events = append(events, "debug"); return false },
				testHighBit:       func(uint16) { t.Fatal("unexpected debug callback") },
				objectFromNetCode: func(uint32) uint64 { events = append(events, "item"); return tc.resolved },
				report: func(owner, item uint64) {
					events = append(events, "report")
					if owner != 3 || item != tc.resolved {
						t.Fatalf("report = (%d, %d), want (3, %d)", owner, item, tc.resolved)
					}
				},
			}
			if got := networkReportSecondaryWeapon51BAD0(uint64(3), hooks); got != 3 {
				t.Fatalf("consumed = %d, want 3", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestNetworkReportSecondaryWeaponNativePointers51BAD0(t *testing.T) {
	const (
		extent  = uint32(0x123)
		netCode = uint32(0x4567)
	)
	owner := &Object{}
	item := &Object{Extent: extent, NetCode: netCode}
	s := &Server{}
	s.Objs.List = item
	packet := &[NetworkReportSecondaryWeaponPacketSize51BAD0]byte{0: 0xe0}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	calls := 0
	got := s.NetworkReportSecondaryWeapon51BAD0(owner, packet, NetworkReportSecondaryWeaponRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		Report: func(gotOwner, gotItem *Object) {
			calls++
			if gotOwner != owner || gotItem != item {
				t.Fatalf("native report = (%p, %p), want (%p, %p)", gotOwner, gotItem, owner, item)
			}
		},
	})
	if got != 3 || calls != 1 {
		t.Fatalf("result = (%d, calls %d), want (3,1)", got, calls)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0))) {
		t.Fatalf("test pointers did not exercise high native halves: owner=%p item=%p", owner, item)
	}
}
