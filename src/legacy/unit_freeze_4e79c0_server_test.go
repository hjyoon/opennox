package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitFreezeNative4E79C0AlreadyFrozen(t *testing.T) {
	obj := &server.Object{
		ObjClass: object.ClassPlayer,
		ObjFlags: object.Flags(0xa5a50002),
	}
	gate := uint32(9)
	if got := unitFreezeNative4E79C0(obj, 7, &gate, unitFreezeNativeDeps4E79C0{}); got != 2 {
		t.Fatalf("return = %#x, want 2", got)
	}
	if obj.ObjFlags != object.Flags(0xa5a50002) || gate != 9 {
		t.Fatalf("state = (%#x, %#x), want unchanged", obj.ObjFlags, gate)
	}
}

func TestUnitUnfreezeNative4E7A60GateBlocksPlayer(t *testing.T) {
	obj := &server.Object{
		ObjClass: object.ClassPlayer,
		ObjFlags: object.Flags(0x11220002),
	}
	gate := uint32(0x123456a5)
	if got := unitUnfreezeNative4E7A60(obj, 0, &gate, unitFreezeNativeDeps4E79C0{}); got != 0xa5 {
		t.Fatalf("return = %#x, want 0xa5", got)
	}
	if obj.ObjFlags != object.Flags(0x11220002) || gate != 0x123456a5 {
		t.Fatalf("state = (%#x, %#x), want unchanged", obj.ObjFlags, gate)
	}
}

func TestUnitFreezeNative4E79C0DeadMonster(t *testing.T) {
	obj := &server.Object{
		ObjClass: object.ClassMonster,
		ObjFlags: object.Flags(0xabcd80f0),
	}
	gate := uint32(9)
	if got := unitFreezeNative4E79C0(obj, 7, &gate, unitFreezeNativeDeps4E79C0{}); got != 0xf2 {
		t.Fatalf("return = %#x, want 0xf2", got)
	}
	if obj.ObjFlags != object.Flags(0xabcd80f2) || gate != 9 {
		t.Fatalf("state = (%#x, %#x), want frozen dead monster", obj.ObjFlags, gate)
	}
}

func TestUnitFreezeNative4E79C0OwnedMonsterTraversal(t *testing.T) {
	root := &server.Object{ObjClass: object.ClassPlayer}
	child1Update := &server.MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	child1 := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(child1Update)}
	child2Update := &server.MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	child2 := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(child2Update)}
	child3Update := &server.MonsterUpdateData{}
	child3 := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(child3Update)}
	root.Field129 = child1
	child1.Field128 = child2

	gate := uint32(0)
	var events []string
	deps := unitFreezeNativeDeps4E79C0{
		reportStatus:  func(*server.Object) byte { events = append(events, "report"); return 1 },
		setPlayerIdle: func(*server.Object) { events = append(events, "idle") },
		raiseZero:     func(*server.Object) { events = append(events, "raise") },
		resetPaths:    func() { events = append(events, "paths") },
		pushIdle: func(obj *server.Object) byte {
			events = append(events, "push")
			if obj == child1 {
				child1.Field128 = child3
			}
			return 0x55
		},
	}
	if got := unitFreezeNative4E79C0(root, 0x80000001, &gate, deps); got != byte(object.ClassPlayer) {
		t.Fatalf("return = %#x, want player class", got)
	}
	if gate != 0x80000001 || !root.ObjFlags.Has(object.FlagNoUpdate) || !child1.ObjFlags.Has(object.FlagNoUpdate) {
		t.Fatalf("freeze state = (%#x, %#x, %#x), want gate and root/child1 frozen", gate, root.ObjFlags, child1.ObjFlags)
	}
	if child2.ObjFlags.Has(object.FlagNoUpdate) || child3.ObjFlags.Has(object.FlagNoUpdate) {
		t.Fatal("traversal used a stale successor or recursed into an unsummoned monster")
	}
	if want := []string{"report", "idle", "raise", "paths", "push"}; len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	} else {
		for i := range want {
			if events[i] != want[i] {
				t.Fatalf("events = %v, want %v", events, want)
			}
		}
	}
}
