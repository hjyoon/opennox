package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterActionGuard546010LocationArgumentSlots(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "reused-slot", true: "failed-push"}[missing], func(t *testing.T) {
			unit := guardTestObject546010(t)
			guard := types.Ptf(1406, 1792) // Original War03a:Robber guard point.
			unit.PosVec, unit.PrevPos = guard, guard
			update := unit.UpdateDataMonster()
			update.AIStack[0].SetArgs(guard, uint32(96))
			update.SightRange = 500
			update.CurrentEnemy = &Object{PosVec: types.Ptf(1278, 2147)}
			high := uintptr(0x12345678)
			high <<= 32
			condition := AIStackItem{
				Action: uint32(ai.DEPENDENCY_LOCATION_CLOSER_THAN),
				Args:   [4]uintptr{11, high | 0x87654321, 33, 44}, Field5: 55,
			}
			before := condition
			var records []guardActionRecord546010
			hooks := guardTestHooks546010(702, &records)
			hooks.push = func(action ai.ActionType, args ...any) *AIStackItem {
				records = append(records, guardActionRecord546010{typeID: action})
				if action != ai.DEPENDENCY_LOCATION_CLOSER_THAN {
					return &AIStackItem{}
				}
				if missing {
					return nil
				}
				condition.SetArgs(args...)
				return &condition
			}
			monsterActionGuard546010(unit, hooks)
			want := before
			if !missing {
				// GAME.EXE 00546010 writes item+4, +12 and +16.
				// item+8 is unused and retains its native-width previous value.
				want.Args[0] = uintptr(math.Float32bits(750))
				want.Args[2] = uintptr(math.Float32bits(1406))
				want.Args[3] = uintptr(math.Float32bits(1792))
			}
			if condition != want {
				t.Fatalf("guard location payload = %+v, want original slots %+v (pointer width %d)", condition, want, unsafe.Sizeof(uintptr(0)))
			}
			wantActions := []ai.ActionType{ai.DEPENDENCY_UNDER_ATTACK, ai.DEPENDENCY_LOCATION_CLOSER_THAN, ai.DEPENDENCY_OR, ai.ACTION_FIGHT}
			if len(records) != len(wantActions) {
				t.Fatalf("guard action count = %d, want %d", len(records), len(wantActions))
			}
			for i, action := range wantActions {
				if records[i].typeID != action {
					t.Fatalf("action %d = %s, want %s", i, records[i].typeID, action)
				}
			}
		})
	}
}
