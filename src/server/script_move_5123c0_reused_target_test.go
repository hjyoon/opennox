package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The original stack push does not clear arguments. Exercise that real reuse
// with C-owned pointers, rather than fresh zero-filled slots or invented addresses.
func TestScriptMoveNative5123C0ReusedObjectTargetFullWidth(t *testing.T) {
	for _, npc := range []bool{false, true} {
		for _, linked := range []bool{false, true} {
			for _, prior := range []ai.ActionType{
				ai.ACTION_ESCORT, ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO,
				ai.ACTION_MISSILE_ATTACK, ai.ACTION_CAST_SPELL_ON_OBJECT,
				ai.ACTION_CAST_DURATION_SPELL, ai.ACTION_FLEE,
			} {
				t.Run(fmt.Sprintf("npc=%t/linked=%t/prior=%s", npc, linked, prior), func(t *testing.T) {
					s := newMoverStateServer(t)
					unit, freeUnit := alloc.New(Object{})
					update, freeUpdate := alloc.New(MonsterUpdateData{})
					target, freeTarget := alloc.New(Object{})
					waypoint, freeWaypoint := alloc.New(Waypoint{})
					t.Cleanup(freeUnit)
					t.Cleanup(freeUpdate)
					t.Cleanup(freeTarget)
					t.Cleanup(freeWaypoint)
					unit.ObjClass = object.ClassMonster
					if npc {
						unit.ObjSubClass = object.SubClass(object.MonsterNPC)
					}
					unit.UpdateData = unsafe.Pointer(update)
					unit.serverHandle = s.handle
					target.ObjClass = object.ClassPlayer
					target.PosVec = types.Pointf{X: 17.5, Y: -9.25}
					waypoint.PosVec = types.Pointf{X: 103.5, Y: -27.25}
					moveIndex := 1
					if linked {
						waypoint.PointsCnt = 1
						moveIndex = 2
					}
					targetPtr := uintptr(unsafe.Pointer(target))
					waypointPtr := uintptr(unsafe.Pointer(waypoint))
					if unsafe.Sizeof(uintptr(0)) == 8 {
						for _, ptr := range []uintptr{uintptr(unsafe.Pointer(unit)), uintptr(unsafe.Pointer(update)), targetPtr, waypointPtr} {
							if ptr <= math.MaxUint32 {
								t.Fatalf("C-owned fixture pointer %#x must be above 4 GiB", ptr)
							}
						}
					}
					unitBefore, targetBefore, waypointBefore := *unit, *target, *waypoint
					for cycle := 0; cycle < 64; cycle++ {
						update.AIStackInd = 0
						update.Field333 = 0xdeadbea5
						update.CurrentEnemy, update.PreferredEnemy = target, target
						for i := range update.AIStack {
							update.AIStack[i] = AIStackItem{
								Action: uint32(ai.ACTION_WAIT),
								Args:   [4]uintptr{targetPtr, waypointPtr, targetPtr, uintptr(0x12340000 + i)},
							}
						}
						if linked {
							unit.MonsterPushAction(ai.ACTION_WAIT)
						}
						old := unit.MonsterPushAction(prior)
						if old != &update.AIStack[moveIndex] || old.ArgObj(2) != target {
							t.Fatal("real stack push did not reuse the pointer-bearing slot")
						}
						want := update.AIStack
						want[0].Action = uint32(ai.ACTION_REPORT)
						want[0].Args[0] = want[0].Args[0]&^uintptr(0xffffffff) | 8
						if linked {
							want[1].Action = uint32(ai.ACTION_ROAM)
							want[1].Args[0] = waypointPtr
							want[1].Args[2] = want[1].Args[2]&^uintptr(0xff) | 0xa5
						}
						want[moveIndex].Action = uint32(ai.ACTION_FAR_MOVE_TO)
						want[moveIndex].Args[0] = want[moveIndex].Args[0]&^uintptr(0xffffffff) | uintptr(math.Float32bits(waypoint.PosVec.X))
						want[moveIndex].Args[1] = want[moveIndex].Args[1]&^uintptr(0xffffffff) | uintptr(math.Float32bits(waypoint.PosVec.Y))
						want[moveIndex].Args[2] = 0

						s.ScriptMoveTo5123C0(unit, waypoint, ScriptMoveRuntime5123C0{})
						move := &update.AIStack[moveIndex]
						// Check before dereferencing: the old partial clear would make
						// refresh read (target & ~0xffffffff) + Object.ObjFlags offset.
						if move.Args[2] != 0 {
							t.Fatalf("cycle %d: cleared target = %#x, want full-width nil; refresh would read flags at %#x (original target %p)", cycle, move.Args[2], move.Args[2]+unsafe.Offsetof(Object{}.ObjFlags), target)
						}
						if update.AIStackInd != int8(moveIndex) || update.AIStack != want {
							t.Fatalf("cycle %d: unrelated reused arguments or stack layout changed", cycle)
						}
						if got := s.MonsterActionRefresh50A910(unit); got != 0 || update.AIStack != want {
							t.Fatalf("cycle %d: production refresh = %d or modified a targetless move", cycle, got)
						}
						if update.CurrentEnemy != target || update.PreferredEnemy != target || update.Field333 != 0xdeadbea5 || !s.AI.StackChanged {
							t.Fatal("move/refresh damaged native enemy, roam flags, or stack-change state")
						}
						if *unit != unitBefore || *target != targetBefore || *waypoint != waypointBefore {
							t.Fatal("move/refresh modified a C-owned object or waypoint")
						}
					}
				})
			}
		}
	}
}
