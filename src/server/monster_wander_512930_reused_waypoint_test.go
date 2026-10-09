package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestMonsterWanderNative512930ReusedWaypointFullWidth(t *testing.T) {
	for _, npc := range []bool{false, true} {
		t.Run(fmt.Sprintf("npc=%t", npc), func(t *testing.T) {
			s := newMoverStateServer(t)
			unit, freeUnit := alloc.New(Object{})
			update, freeUpdate := alloc.New(MonsterUpdateData{})
			waypoint, freeWaypoint := alloc.New(Waypoint{})
			t.Cleanup(freeUnit)
			t.Cleanup(freeUpdate)
			t.Cleanup(freeWaypoint)
			unit.ObjClass = object.ClassMonster
			if npc {
				unit.ObjSubClass = object.SubClass(object.MonsterNPC)
			}
			unit.UpdateData = unsafe.Pointer(update)
			unit.serverHandle = s.handle
			unitPtr := uintptr(unsafe.Pointer(unit))
			waypointPtr := uintptr(unsafe.Pointer(waypoint))
			if unsafe.Sizeof(uintptr(0)) == 8 && (unitPtr <= math.MaxUint32 || waypointPtr <= math.MaxUint32 || uintptr(unsafe.Pointer(update)) <= math.MaxUint32) {
				t.Fatal("C-owned unit, update, and waypoint must have native addresses above 4 GiB")
			}
			unitBefore, waypointBefore := *unit, *waypoint
			for cycle := 0; cycle < 64; cycle++ {
				update.AIStackInd = 0
				update.Field333 = 0xdeadbea5
				for i := range update.AIStack {
					update.AIStack[i] = AIStackItem{
						Action: uint32(ai.ACTION_WAIT),
						Args:   [4]uintptr{unitPtr, waypointPtr, unitPtr, uintptr(0x43210000 + i)},
					}
				}
				old := unit.MonsterPushAction(ai.ACTION_ROAM, unsafe.Pointer(waypoint))
				if old != &update.AIStack[1] || old.Args[0] != waypointPtr {
					t.Fatal("real stack push did not store the native waypoint in the reused slot")
				}
				want := update.AIStack
				want[0].Action = uint32(ai.ACTION_REPORT)
				want[0].Args[0] = want[0].Args[0]&^uintptr(0xffffffff) | 10
				want[1].Action = uint32(ai.ACTION_ROAM)
				want[1].Args[0] = 0
				want[1].Args[2] = want[1].Args[2]&^uintptr(0xff) | 0xa5

				s.ScriptMonsterRoam512930(unit)
				if got := update.AIStack[1].Args[0]; got != 0 {
					t.Fatalf("cycle %d: cleared waypoint = %#x, want full-width nil (original waypoint %p)", cycle, got, waypoint)
				}
				if update.AIStackInd != 1 || update.AIStack != want {
					t.Fatalf("cycle %d: unrelated reused arguments or stack layout changed", cycle)
				}
				if got := s.MonsterActionRefresh50A910(unit); got != 0 || update.AIStack != want {
					t.Fatalf("cycle %d: production refresh = %d or modified the roam stack", cycle, got)
				}
				if update.Field333 != 0xdeadbea5 || !s.AI.StackChanged || *unit != unitBefore || *waypoint != waypointBefore {
					t.Fatal("wander/refresh modified native object, waypoint, roam flags, or stack-change state")
				}
			}
		})
	}
}
