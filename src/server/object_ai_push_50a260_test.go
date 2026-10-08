package server

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// GAME.EXE 0050A30F/0050A311 only store Action and Field5. In particular,
// reusing idle slot zero does not erase its timer, and unused native pointer
// arguments in the next slot survive until the caller explicitly writes them.
func TestMonsterPushAction50A260PreservesReusedSlotArguments(t *testing.T) {
	for index := -1; index <= 22; index++ {
		t.Run(fmt.Sprintf("index_%d", index), func(t *testing.T) {
			s := unitIdleTestServer515820(t)
			unit := monsterActionTestObject50A910(t)
			unit.serverHandle = s.handle
			update := unit.UpdateDataMonster()
			target := new(Object)
			for i := range update.AIStack {
				update.AIStack[i] = AIStackItem{
					Action: uint32(ai.DEPENDENCY_TIME),
					Args: [4]uintptr{
						uintptr(580 + i), uintptr(0x80000000 + uint32(i)),
						uintptr(unsafe.Pointer(target)), uintptr(0xffffffff),
					},
					Field5: uint32(91 + i),
				}
			}
			update.AIStackInd = int8(index)
			before := update.AIStack
			update.Field2, update.Field67, update.Field74, update.Field91 = 91, 92, 93, 94
			update.Field120_1, update.Field120_2, update.Field120_3 = 1, 2, 3
			update.Field124, update.Field137 = 95, 96
			want := before
			want[index+1].Action = uint32(ai.DEPENDENCY_NOT_CORNERED)
			want[index+1].Field5 = 0

			got := unit.MonsterPushActionImpl(ai.DEPENDENCY_NOT_CORNERED, "fixture", 31)
			if got != &update.AIStack[index+1] || update.AIStackInd != int8(index+1) || update.AIStack != want {
				t.Fatalf("push erased arguments or changed another slot: index %d, got %#v, want %#v", index, got, &want[index+1])
			}
			if update.Field2 != 0 || update.Field67 != 0 || update.Field74 != 0 || update.Field91 != 0 ||
				update.Field120_1 != 0 || update.Field120_2 != 0 || update.Field120_3 != 0 ||
				update.Field124 != s.Frame() || update.Field137 != s.Frame() || !s.AI.StackChanged {
				t.Fatal("original action reset/stack-changed contract was lost")
			}
			runtime.KeepAlive(target)
			runtime.KeepAlive(unit)
		})
	}
}

func TestMonsterPushAction50A260IdleReplacementAndPartialArguments(t *testing.T) {
	for _, mode := range []string{"no_arguments", "one_word", "point_and_object"} {
		t.Run(mode, func(t *testing.T) {
			s := unitIdleTestServer515820(t)
			unit := monsterActionTestObject50A910(t)
			unit.serverHandle = s.handle
			update := unit.UpdateDataMonster()
			target := new(Object)
			args := [4]uintptr{580, 0x80000000, uintptr(unsafe.Pointer(target)), 0xffffffff}
			update.AIStackInd = 0
			update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_IDLE), Args: args, Field5: 1}
			want := args
			var got *AIStackItem
			switch mode {
			case "no_arguments":
				got = unit.MonsterPushAction(ai.DEPENDENCY_NOT_CORNERED)
			case "one_word":
				got = unit.MonsterPushAction(ai.ACTION_WAIT, uint32(689))
				want[0] = 689
			case "point_and_object":
				got = unit.MonsterPushAction(ai.ACTION_MOVE_TO, types.Pointf{X: 1, Y: -2}, target)
				want[0], want[1] = 0x3f800000, 0xc0000000
			}
			if got != &update.AIStack[0] || update.AIStackInd != 0 || got.Args != want || got.Field5 != 0 {
				t.Fatalf("idle replacement/partial stores = %#v at %d, want preserved args %#v and reset Field5", got, update.AIStackInd, want)
			}
			runtime.KeepAlive(target)
			runtime.KeepAlive(unit)
		})
	}
}
