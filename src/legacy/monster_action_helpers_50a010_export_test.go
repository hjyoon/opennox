package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterActionHelpers50A010CABIUsesNativePointers(t *testing.T) {
	update := new(server.MonsterUpdateData)
	unit := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	var pin runtime.Pinner
	pin.Pin(update)
	pin.Pin(unit)
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":   uintptr(unsafe.Pointer(unit)),
			"update": uintptr(unsafe.Pointer(update)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	update.AIStackInd = 4
	update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
	update.AIStack[1].Action = uint32(ai.DEPENDENCY_TIME)
	update.AIStack[2].Action = 72
	update.AIStack[3].Action = uint32(ai.DEPENDENCY_OR)
	update.AIStack[4].Action = uint32(ai.ACTION_MOVE_TO)

	if got := monsterActionGetExportCall50A020(unit); got != ai.ACTION_MOVE_TO {
		t.Fatalf("C current action = %s, want MOVE_TO", got)
	}
	if got := monsterActionPreviousExportCall50A040(unit); got != ai.ACTION_ESCORT {
		t.Fatalf("C previous action = %s, want ESCORT", got)
	}
	if got := monsterActionPushIfChangedExportCall50A360(unit, ai.ACTION_MOVE_TO); got != nil {
		t.Fatalf("C same-action push = %p, want nil", got)
	}
	for _, tc := range []struct {
		action int32
		want   int32
	}{
		{action: 39, want: 0},
		{action: 40, want: 1},
		{action: 72, want: 1},
		{action: math.MaxInt32, want: 1},
		{action: math.MinInt32, want: 0},
		{action: -1, want: 0},
	} {
		if got := monsterActionIsConditionExportCall50A010(tc.action); got != tc.want {
			t.Errorf("C IsCondition(%d) = %d, want %d", tc.action, got, tc.want)
		}
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
}
