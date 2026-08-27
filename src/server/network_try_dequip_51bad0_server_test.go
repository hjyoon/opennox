package server

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestNetworkTryDequipNativePointers51BAD0(t *testing.T) {
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
	packet := &[NetworkTryDequipPacketSize51BAD0]byte{0: 0x76}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	dequips := 0
	got := s.NetworkTryDequip51BAD0(unit, update, packet, NetworkTryDequipRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		TryDequip: func(gotUnit, gotItem *Object) {
			dequips++
			if gotUnit != unit || gotItem != item {
				t.Fatalf("native pointers = (%p, %p), want (%p, %p)", gotUnit, gotItem, unit, item)
			}
		},
	})
	if got != 3 || dequips != 1 {
		t.Fatalf("result = (%d, dequips %d), want (3,1)", got, dequips)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(unit)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(update)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(update.Player)) <= uintptr(^uint32(0))) {
		t.Fatalf("test pointers did not exercise high native halves: unit=%p item=%p update=%p player=%p", unit, item, update, update.Player)
	}
}

func TestNetworkTryDequipNativeProtectedItem51BAD0(t *testing.T) {
	item := &Object{NetCode: 7, ObjClass: 0x01000000, ObjSubClass: 0x108}
	unit := &Object{InvFirstItem: item}
	update := &PlayerUpdateData{Player: &Player{}, State: PlayerState1}
	s := &Server{}
	packet := &[NetworkTryDequipPacketSize51BAD0]byte{0x76, 7, 0}
	called := false
	if got := s.NetworkTryDequip51BAD0(unit, update, packet, NetworkTryDequipRuntime51BAD0{
		NetDebug:    func() bool { return false },
		TestHighBit: func(uint16) { t.Fatal("unexpected debug callback") },
		TryDequip:   func(*Object, *Object) { called = true },
	}); got != 3 {
		t.Fatalf("consumed = %d, want 3", got)
	}
	if called {
		t.Fatal("protected state-one quest item was dequipped")
	}
}
