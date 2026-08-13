package legacy

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitBecomePetNative4E7B00UsesNativeFieldsAndCachedUpdate(t *testing.T) {
	first := &server.Player{PlayerInd: 0x81}
	second := &server.Player{PlayerInd: 0xfe}
	replacement := &server.Player{PlayerInd: 0x44}
	cachedUpdate := &server.PlayerUpdateData{Player: first}
	owner := &server.Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   0x11223344,
		UpdateData: unsafe.Pointer(cachedUpdate),
	}
	pet := &server.Object{
		TypeInd:     0x2468,
		ObjSubClass: object.SubClass(0xa5a50001),
		ObjFlags:    0x55667788,
	}
	var events []string
	deps := unitPetNativeDeps4E7B00{
		monitor: func(ind byte, obj *server.Object) {
			events = append(events, "monitor")
			if ind != 0x81 || obj != pet {
				t.Fatalf("monitor = (%#x, %p), want (0x81, %p)", ind, obj, pet)
			}
			owner.UpdateData = unsafe.Pointer(&server.PlayerUpdateData{Player: replacement})
			cachedUpdate.Player = second
		},
		mark: func(ind byte, obj *server.Object, flags uint32) {
			events = append(events, "mark")
			if ind != 0xfe || obj != pet || flags != 1 {
				t.Fatalf("mark = (%#x, %p, %d), want (0xfe, %p, 1)", ind, obj, flags, pet)
			}
		},
		setOwner: func(gotOwner, obj *server.Object) {
			events = append(events, "set-owner")
			if gotOwner != owner || obj != pet {
				t.Fatalf("set owner = (%p, %p), want (%p, %p)", gotOwner, obj, owner, pet)
			}
			obj.ObjOwner = gotOwner
		},
	}

	unitBecomePetNative4E7B00(owner, pet, deps)
	if want := []string{"monitor", "mark", "set-owner"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if got := uint32(pet.ObjSubClass); got != 0xa5a50081 {
		t.Fatalf("pet subclass = %#08x, want 0xa5a50081", got)
	}
	if pet.ObjOwner != owner || pet.TypeInd != 0x2468 || pet.ObjFlags != 0x55667788 || owner.ObjFlags != 0x11223344 {
		t.Fatal("native pet transition changed an unrelated field")
	}
}

func TestUnitBecomeEnemyNative4E7B60UsesNativeFieldsAndCachedUpdate(t *testing.T) {
	first := &server.Player{PlayerInd: 0x80}
	second := &server.Player{PlayerInd: 0xff}
	replacement := &server.Player{PlayerInd: 0x33}
	cachedUpdate := &server.PlayerUpdateData{Player: first}
	owner := &server.Object{UpdateData: unsafe.Pointer(cachedUpdate), ObjFlags: 0x11223344}
	pet := &server.Object{
		TypeInd:     0x1357,
		ObjSubClass: object.SubClass(0x5a5a80ff),
		ObjFlags:    0x88776655,
		ObjOwner:    owner,
	}
	var events []string
	deps := unitPetNativeDeps4E7B00{
		unmonitor: func(ind byte, obj *server.Object) {
			events = append(events, "unmonitor")
			if ind != 0x80 || obj != pet {
				t.Fatalf("unmonitor = (%#x, %p), want (0x80, %p)", ind, obj, pet)
			}
			owner.UpdateData = unsafe.Pointer(&server.PlayerUpdateData{Player: replacement})
			cachedUpdate.Player = second
		},
		unmark: func(ind byte, obj *server.Object, flags uint32) {
			events = append(events, "unmark")
			if ind != 0xff || obj != pet || flags != 1 {
				t.Fatalf("unmark = (%#x, %p, %d), want (0xff, %p, 1)", ind, obj, flags, pet)
			}
		},
		clearOwner: func(obj *server.Object) {
			events = append(events, "clear-owner")
			if obj != pet {
				t.Fatalf("clear owner object = %p, want %p", obj, pet)
			}
			obj.ObjOwner = nil
		},
	}

	unitBecomeEnemyNative4E7B60(owner, pet, deps)
	if want := []string{"unmonitor", "unmark", "clear-owner"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if got := uint32(pet.ObjSubClass); got != 0x5a5a807f {
		t.Fatalf("pet subclass = %#08x, want 0x5a5a807f", got)
	}
	if pet.ObjOwner != nil || pet.TypeInd != 0x1357 || pet.ObjFlags != 0x88776655 || owner.ObjFlags != 0x11223344 {
		t.Fatal("native enemy transition changed an unrelated field")
	}
}

func TestUnitBecomePetNative4E7B00NilGuards(t *testing.T) {
	unitBecomePetNative4E7B00(nil, &server.Object{}, unitPetNativeDeps4E7B00{})
	unitBecomePetNative4E7B00(&server.Object{}, nil, unitPetNativeDeps4E7B00{})
}

func TestUnitBecomeEnemyNative4E7B60NilOwnerFaultsFirst(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner returned without a panic")
		}
	}()
	unitBecomeEnemyNative4E7B60(nil, nil, unitPetNativeDeps4E7B00{})
}

func TestNetMonitorCreatureNative4D9250PacketsAndReloadsHealth(t *testing.T) {
	initialHealth := &server.HealthData{Cur: 0x1111, Max: 0x2222}
	reloadedHealth := &server.HealthData{Cur: 0x3333, Max: 0x4444}
	obj := &server.Object{
		TypeInd:    0x5678,
		NetCode:    0x1234,
		HealthData: initialHealth,
	}
	var packets [][]byte
	var calls [][4]int
	s := &server.Server{
		NetSendPacketXxx: func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
			if related != nil {
				t.Fatalf("related object = %p, want nil", related)
			}
			packets = append(packets, append([]byte(nil), packet...))
			calls = append(calls, [4]int{recipient, remove, sequence, len(packets)})
			if len(packets) == 1 {
				obj.HealthData = reloadedHealth
			}
			return len(packets) * 11
		},
	}

	if got := netMonitorCreatureNative4D9250(s, 0xfe, obj); got != 22 {
		t.Fatalf("monitor result = %d, want 22", got)
	}
	wantPackets := [][]byte{
		{0xdb, 0x34, 0x12, 0x78, 0x56},
		{0xdd, 0x34, 0x12, 0x33, 0x33, 0x44, 0x44},
	}
	if !reflect.DeepEqual(packets, wantPackets) {
		t.Fatalf("packets = %v, want %v", packets, wantPackets)
	}
	wantCalls := [][4]int{{0xfe, 1, 1, 1}, {0xfe, 1, 1, 2}}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", calls, wantCalls)
	}
	if initialHealth.Cur != 0x1111 || initialHealth.Max != 0x2222 {
		t.Fatal("monitor changed the initial health record")
	}
}

func TestNetMonitorCreatureNative4D9250WithoutHealthReturnsZero(t *testing.T) {
	obj := &server.Object{TypeInd: 0x5678, NetCode: 0x1234}
	var packets [][]byte
	s := &server.Server{
		NetSendPacketXxx: func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
			if recipient != 0xfe || related != nil || remove != 1 || sequence != 1 {
				t.Fatalf("send args = (%#x, %p, %d, %d), want (0xfe, nil, 1, 1)", recipient, related, remove, sequence)
			}
			packets = append(packets, append([]byte(nil), packet...))
			return 0x55
		},
	}

	if got := netMonitorCreatureNative4D9250(s, 0xfe, obj); got != 0 {
		t.Fatalf("monitor result = %#x, want 0", got)
	}
	if want := [][]byte{{0xdb, 0x34, 0x12, 0x78, 0x56}}; !reflect.DeepEqual(packets, want) {
		t.Fatalf("packets = %v, want %v", packets, want)
	}
}

func TestNetUnmonitorCreatureNative4D92A0Packet(t *testing.T) {
	obj := &server.Object{NetCode: 0x2468}
	var gotPacket []byte
	s := &server.Server{
		NetSendPacketXxx: func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
			if recipient != 0x81 || related != nil || remove != 1 || sequence != 1 {
				t.Fatalf("send args = (%#x, %p, %d, %d), want (0x81, nil, 1, 1)", recipient, related, remove, sequence)
			}
			gotPacket = append([]byte(nil), packet...)
			return 0x55
		},
	}

	if got := netUnmonitorCreatureNative4D92A0(s, 0x81, obj); got != 0x55 {
		t.Fatalf("unmonitor result = %#x, want 0x55", got)
	}
	if want := []byte{0xdc, 0x68, 0x24}; !reflect.DeepEqual(gotPacket, want) {
		t.Fatalf("packet = %v, want %v", gotPacket, want)
	}
}
