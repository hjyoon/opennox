package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDoorUpdate53AC50SoundsAndStatus(t *testing.T) {
	s := new(Server)
	departingData := &DoorUpdateData{TargetDirection: 3, SyncedDirection: 3, CurrentDirection: 4}
	departing := &Object{ObjClass: 4, Material: 8, UpdateData: unsafe.Pointer(departingData)}
	var gotSound uint32
	s.DoorUpdate53AC50(departing, DoorUpdateRuntime53AC50{AudioEvent: func(id uint32, got *Object) {
		if got != departing {
			t.Fatal("audio object differs")
		}
		gotSound = id
		departingData.CurrentDirection = 6
	}})
	if gotSound != 245 || departingData.SyncedDirection != 6 || departing.Field38 != ^uint32(0) {
		t.Fatalf("departing = sound %d, sync %d, dirty %#x", gotSound, departingData.SyncedDirection, departing.Field38)
	}

	arrivingData := &DoorUpdateData{TargetDirection: 4, SyncedDirection: 3, CurrentDirection: 4}
	arriving := &Object{ObjClass: 1, UpdateData: unsafe.Pointer(arrivingData)}
	s.Objs.AddToUpdatable(arriving)
	s.DoorUpdate53AC50(arriving, DoorUpdateRuntime53AC50{AudioEvent: func(id uint32, _ *Object) { gotSound = id }})
	if gotSound != 248 || arrivingData.SyncedDirection != 4 || arriving.IsUpdatable != 0 {
		t.Fatalf("arriving = sound %d, sync %d, updatable %d", gotSound, arrivingData.SyncedDirection, arriving.IsUpdatable)
	}

	gotSound = 0
	s.DoorUpdate53AC50(arriving, DoorUpdateRuntime53AC50{AudioEvent: func(id uint32, _ *Object) { gotSound = id }})
	if gotSound != 0 {
		t.Fatalf("stationary door replayed sound %d", gotSound)
	}
}

func TestDoorUpdate53AC50SoundClasses(t *testing.T) {
	tests := []struct {
		class    object.Class
		material uint16
		depart   uint32
		arrive   uint32
	}{
		{4, 8, 245, 246},
		{4, 0, 241, 243},
		{1, 0, 247, 248},
		{0x1000, 0, 1014, 1015},
		{0, 0, 237, 239},
	}
	for _, tc := range tests {
		door := &Object{ObjClass: tc.class, Material: tc.material}
		if got := doorUpdateSound53AC50(door, false); got != tc.depart {
			t.Errorf("class %#x departing sound = %d, want %d", tc.class, got, tc.depart)
		}
		if got := doorUpdateSound53AC50(door, true); got != tc.arrive {
			t.Errorf("class %#x arriving sound = %d, want %d", tc.class, got, tc.arrive)
		}
	}
}

func TestDoorUpdate53AC50MovementAndDeadline(t *testing.T) {
	s := new(Server)
	s.SetTickRate(20)
	data := &DoorUpdateData{TargetDirection: 2, SyncedDirection: 4, CurrentDirection: 4, FractionalDir: 0}
	door := &Object{ObjFlags: object.Flags(0x1000000), UpdateData: unsafe.Pointer(data)}
	var queued, woken int
	runtime := DoorUpdateRuntime53AC50{
		QueueDoor: func(got *DoorUpdateData) {
			if got != data {
				t.Fatal("queued a different update record")
			}
			queued++
		},
		WakeDoor: func(got *Object) {
			if got != door {
				t.Fatal("woke a different door")
			}
			woken++
		},
	}
	s.SetFrame(10)
	s.DoorUpdate53AC50(door, runtime)
	if data.FractionalDir != 0 || queued != 0 || woken != 0 {
		t.Fatalf("at deadline = direction %d, queue %d, wake %d", data.FractionalDir, queued, woken)
	}
	s.SetFrame(11)
	s.DoorUpdate53AC50(door, runtime)
	if data.FractionalDir != 254 || queued != 1 || woken != 1 {
		t.Fatalf("counterclockwise = direction %d, queue %d, wake %d", data.FractionalDir, queued, woken)
	}
	data.CurrentDirection = 20
	data.FractionalDir = 255
	s.DoorUpdate53AC50(door, runtime)
	if data.FractionalDir != 1 || queued != 2 || woken != 2 {
		t.Fatalf("clockwise = direction %d, queue %d, wake %d", data.FractionalDir, queued, woken)
	}
}
