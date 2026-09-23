package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func pickupObjectMonster544B90(target *Object) (*Object, *AIStackItem) {
	update := &MonsterUpdateData{AIStackInd: 0}
	unit := &Object{
		ObjClass:   object.ClassMonster,
		PosVec:     types.Ptf(10, 20),
		UpdateData: unsafe.Pointer(update),
	}
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_PICKUP_OBJECT)}
	update.AIStack[0].SetArgs(target)
	return unit, &update.AIStack[0]
}

func TestMonsterActionPickupObject544B90ReloadsAndUsesHealthPotion(t *testing.T) {
	target := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodSimple), PosVec: types.Ptf(20, 20)}
	reloaded := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodHealthPotion)}
	unit, head := pickupObjectMonster544B90(target)
	var events []string
	got := monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
		canInteract: func(gotUnit, gotTarget *Object, flags int) bool {
			if gotUnit != unit || gotTarget != target || flags != 0 {
				t.Fatalf("interaction = %p/%p/%d", gotUnit, gotTarget, flags)
			}
			events = append(events, "interact")
			return true
		},
		placeInventory: func(gotUnit, gotTarget *Object, arg3, arg4 int) bool {
			if gotUnit != unit || gotTarget != target || arg3 != 1 || arg4 != 1 {
				t.Fatalf("place = %p/%p/%d/%d", gotUnit, gotTarget, arg3, arg4)
			}
			events = append(events, "place")
			head.SetArgs(reloaded)
			return false // GAME.EXE ignores this result.
		},
		useByNetCode: func(gotUnit, gotTarget *Object) int32 {
			if gotUnit != unit || gotTarget != reloaded {
				t.Fatalf("use = %p/%p, want unit/reloaded target", gotUnit, gotTarget)
			}
			events = append(events, "use")
			return 0
		},
		pop: func() int {
			events = append(events, "pop")
			return 37
		},
	})
	if got != 37 {
		t.Fatalf("result = %d, want pop result 37", got)
	}
	if want := []string{"interact", "place", "use", "pop"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMonsterActionPickupObject544B90AdmissionAndCompletion(t *testing.T) {
	tests := []struct {
		name       string
		target     *Object
		interacts  bool
		wantEvents []string
	}{
		{name: "missing target", wantEvents: []string{"pop"}},
		{
			name:       "range boundary excluded",
			target:     &Object{ObjClass: object.ClassFood, PosVec: types.Ptf(85, 20)},
			interacts:  true,
			wantEvents: []string{"pop"},
		},
		{
			name:       "blocked target",
			target:     &Object{ObjClass: object.ClassFood, PosVec: types.Ptf(20, 20)},
			wantEvents: []string{"interact", "pop"},
		},
		{
			name:       "ordinary food",
			target:     &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodSimple), PosVec: types.Ptf(20, 20)},
			interacts:  true,
			wantEvents: []string{"interact", "place", "pop"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit, _ := pickupObjectMonster544B90(tc.target)
			var events []string
			monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
				canInteract: func(*Object, *Object, int) bool {
					events = append(events, "interact")
					return tc.interacts
				},
				placeInventory: func(*Object, *Object, int, int) bool {
					events = append(events, "place")
					return true
				},
				useByNetCode: func(*Object, *Object) int32 {
					events = append(events, "use")
					return 1
				},
				pop: func() int {
					events = append(events, "pop")
					return 0
				},
			})
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestMonsterActionPickupObject544B90RejectsWrongAction(t *testing.T) {
	unit, head := pickupObjectMonster544B90(&Object{})
	head.Action = uint32(ai.ACTION_MOVE_TO)
	popped := false
	if got := monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
		pop: func() int { popped = true; return 0 },
	}); got != 0 || popped {
		t.Fatalf("wrong action = result %d, popped %t", got, popped)
	}
}
