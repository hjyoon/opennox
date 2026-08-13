package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDoorUpdateDataAndTilePointNativeLayout(t *testing.T) {
	if got := unsafe.Sizeof(DoorUpdateData{}); got != 52 {
		t.Fatalf("DoorUpdateData size = %d, want 52", got)
	}
	if got := unsafe.Offsetof(DoorUpdateData{}.LockCode); got != 1 {
		t.Fatalf("LockCode offset = %d, want 1", got)
	}
	if got := unsafe.Offsetof(DoorUpdateData{}.TileX); got != 16 {
		t.Fatalf("TileX offset = %d, want 16", got)
	}
	if got := unsafe.Offsetof(DoorUpdateData{}.TileY); got != 20 {
		t.Fatalf("TileY offset = %d, want 20", got)
	}
	if got := unsafe.Offsetof(DoorUpdateData{}.QuestSync); got != 48 {
		t.Fatalf("QuestSync offset = %d, want 48", got)
	}
	if got := unsafe.Sizeof(DoorTilePoint{}); got != 8 {
		t.Fatalf("DoorTilePoint size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(DoorTilePoint{}.Y); got != 4 {
		t.Fatalf("DoorTilePoint.Y offset = %d, want 4", got)
	}
}

func TestDoorCloseNative4E8340UsesNamedFieldsOnly(t *testing.T) {
	raw := [52]byte{}
	for i := range raw {
		raw[i] = 0xa5
	}
	update := (*DoorUpdateData)(unsafe.Pointer(&raw[0]))
	update.LockCode = 4
	update.TileX = -17
	update.TileY = math.MaxInt32
	door := &Object{
		ObjClass:   object.ClassDoor | object.ClassClientPersist,
		UpdateData: unsafe.Pointer(update),
	}
	events := []string{}
	doorCloseNative4E8340(door, &DoorTilePoint{X: -17, Y: math.MaxInt32},
		func() int32 {
			events = append(events, "quest")
			if update.LockCode != 0 {
				t.Fatalf("quest observed lock %d, want 0", update.LockCode)
			}
			return -1
		},
		func(got *Object) int32 {
			events = append(events, "sync")
			if got != door {
				t.Fatal("sync received a different door")
			}
			return math.MinInt32
		},
	)
	if len(events) != 2 || events[0] != "quest" || events[1] != "sync" {
		t.Fatalf("events = %v, want [quest sync]", events)
	}
	for i, got := range raw {
		want := byte(0xa5)
		if i == 1 {
			want = 0
		}
		if i >= 16 && i < 24 {
			continue
		}
		if got != want {
			t.Fatalf("raw[%d] = %#x, want %#x", i, got, want)
		}
	}
}

func TestDoorQuestSyncAndExtentPacketNative4E8390(t *testing.T) {
	update := &DoorUpdateData{LockCode: 3, TileX: 11, TileY: 22, QuestSync: 0x7a}
	door := &Object{Extent: 0xa5a5bcde, UpdateData: unsafe.Pointer(update)}
	s := &Server{}
	calls := 0
	s.NetSendPacketXxx = func(recipient int, buf []byte, related *Object, removeIfDisconnected, sequenceEnabled int) int {
		calls++
		if update.QuestSync != 1 {
			t.Fatalf("send observed QuestSync = %d, want 1", update.QuestSync)
		}
		if recipient != 255 || string(buf) != string([]byte{0xf0, 0x0f, 0xde, 0xbc}) || related != nil ||
			removeIfDisconnected != 1 || sequenceEnabled != 0 {
			t.Fatalf("send = (%d, % x, %p, %d, %d)", recipient, buf, related, removeIfDisconnected, sequenceEnabled)
		}
		return math.MinInt32
	}
	if got := s.DoorQuestSync4E8390(door); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if calls != 1 {
		t.Fatalf("send calls = %d, want 1", calls)
	}
	if update.LockCode != 3 || update.TileX != 11 || update.TileY != 22 {
		t.Fatalf("adjacent DoorUpdate fields changed: %#v", update)
	}
}

func TestDoorExtentPacket4D6A20NilObjectFaultsBeforeSend(t *testing.T) {
	s := &Server{NetSendPacketXxx: func(int, []byte, *Object, int, int) int {
		t.Fatal("nil object reached send")
		return 0
	}}
	defer func() {
		if recover() == nil {
			t.Fatal("nil object returned without a panic")
		}
	}()
	s.DoorExtentPacket4D6A20(255, nil)
}
