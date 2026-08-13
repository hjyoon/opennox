package server

import (
	"strings"
	"testing"
	"unsafe"
)

func TestDoorCollideNative4E8AC0UnlocksWithNativePointers(t *testing.T) {
	update := &DoorUpdateData{
		LockCode:         2,
		TargetDirection:  24,
		CurrentDirection: 24,
		TileX:            -4,
		TileY:            7,
	}
	door := &Object{UpdateData: unsafe.Pointer(update)}
	unit := new(Object)
	key := new(Object)
	collision := [2]float32{12.5, -3.25}

	var (
		gotRect      doorCollideRect4E8AC0
		gotTarget    DoorTilePoint
		gotSound     uint32
		gotSoundObj  *Object
		gotDeleted   *Object
		questChecks  int
		findKeyCalls int
	)
	unexpected := func(name string) {
		t.Fatalf("unexpected native callback: %s", name)
	}
	doorCollideNative4E8AC0(door, unit, unsafe.Pointer(&collision[0]), doorCollideNativeDeps4E8AC0{
		frame: func() uint32 {
			unexpected("frame")
			return 0
		},
		ticks: func() uint64 {
			unexpected("ticks")
			return 0
		},
		loadFeedbackTicks: func() uint64 {
			unexpected("load feedback ticks")
			return 0
		},
		storeFeedbackTicks: func(uint64) {
			unexpected("store feedback ticks")
		},
		audio: func(id uint32, obj *Object) {
			gotSound, gotSoundObj = id, obj
		},
		priorityMessage: func(*Object, string) {
			unexpected("priority message")
		},
		keyMessage: func(*Object, string, uint8) {
			unexpected("key message")
		},
		findKey: func(gotUnit, gotDoor *Object) *Object {
			findKeyCalls++
			if gotUnit != unit || gotDoor != door {
				t.Fatalf("find key args = %p/%p, want %p/%p", gotUnit, gotDoor, unit, door)
			}
			return key
		},
		questMode: func() bool {
			questChecks++
			return false
		},
		questSync: func(*Object) int32 {
			unexpected("Quest sync")
			return 0
		},
		storeQuestFrame: func(uint32) {
			unexpected("store Quest frame")
		},
		eachObjectInRect: func(rect doorCollideRect4E8AC0, target DoorTilePoint) {
			gotRect, gotTarget = rect, target
		},
		questKeyState: func() int32 {
			unexpected("Quest key state")
			return 0
		},
		delayedDelete: func(obj *Object) {
			gotDeleted = obj
		},
	})

	if findKeyCalls != 1 || questChecks != 1 {
		t.Fatalf("find/Quest calls = %d/%d, want 1/1", findKeyCalls, questChecks)
	}
	if update.LockCode != 0 {
		t.Fatalf("native lock code = %d, want 0", update.LockCode)
	}
	wantRect := doorCollideRect4E8AC0{MinX: -126, MinY: 127, MaxX: -58, MaxY: 195}
	if gotRect != wantRect || gotTarget != (DoorTilePoint{X: -5, Y: 8}) {
		t.Fatalf("native geometry = %#v/%#v, want %#v/{-5 8}", gotRect, gotTarget, wantRect)
	}
	if gotSound != doorCollideSoundUnlock4E8AC0 || gotSoundObj != door || gotDeleted != key {
		t.Fatalf("native effects = sound %d/%p delete %p, want %d/%p/%p",
			gotSound, gotSoundObj, gotDeleted, doorCollideSoundUnlock4E8AC0, door, key)
	}
}

func TestDoorCollideKeyPacket4E8AC0(t *testing.T) {
	const message = doorCollideGateLockedKey4E8AC0
	packet, ok := doorCollideKeyPacket4E8AC0(message, 0xa5)
	if !ok {
		t.Fatal("valid keyed message rejected")
	}
	if packet[0] != 0xf0 || packet[1] != 33 || string(packet[2:2+len(message)]) != message ||
		packet[2+len(message)] != 0 || packet[51] != 0xa5 {
		t.Fatalf("keyed packet = % x", packet)
	}
	if _, ok := doorCollideKeyPacket4E8AC0("", 1); ok {
		t.Fatal("empty keyed message accepted")
	}
	if _, ok := doorCollideKeyPacket4E8AC0(strings.Repeat("x", 49), 1); ok {
		t.Fatal("49-byte keyed message accepted")
	}
	packet, ok = doorCollideKeyPacket4E8AC0(strings.Repeat("y", 48), 0xff)
	if !ok || packet[50] != 0 || packet[51] != 0xff {
		t.Fatalf("48-byte keyed packet boundary = ok:%t tail:% x", ok, packet[48:])
	}
}
