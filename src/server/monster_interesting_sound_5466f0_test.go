package server

import (
	"testing"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterInterestingSoundCall5466F0 struct {
	action ai.ActionType
	args   []any
}

func interestingSoundTestHooks5466F0(calls *[]monsterInterestingSoundCall5466F0) monsterInterestingSoundHooks5466F0 {
	return monsterInterestingSoundHooks5466F0{
		frame:    func() uint32 { return 100 },
		tickRate: func() uint32 { return 30 },
		tileAt:   func(types.Pointf) int { return 0 },
		pathReachable: func(*Object, *types.Pointf) bool {
			return true
		},
		trace:  func(types.Pointf, types.Pointf) bool { return true },
		random: func(int, int) int { return 45 },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			*calls = append(*calls, monsterInterestingSoundCall5466F0{action: action, args: args})
			return &AIStackItem{}
		},
	}
}

func interestingSoundTestObject5466F0(t *testing.T) *Object {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.Field97 = 1
	update.Field101 = 90
	update.Field99X = 12.5
	update.Field99Y = -4.25
	unit.PosVec = types.Pointf{X: 1, Y: 2}
	return unit
}

func TestMonsterInterestingSound5466F0DirectPath(t *testing.T) {
	unit := interestingSoundTestObject5466F0(t)
	var calls []monsterInterestingSoundCall5466F0
	hooks := interestingSoundTestHooks5466F0(&calls)
	if got := monsterInterestingSound5466F0(unit, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []ai.ActionType{
		ai.DEPENDENCY_NO_INTERESTING_SOUND,
		ai.DEPENDENCY_NO_VISIBLE_ENEMY,
		ai.ACTION_WAIT,
		ai.ACTION_FACE_LOCATION,
	}
	if len(calls) != len(want) {
		t.Fatalf("push count = %d, want %d", len(calls), len(want))
	}
	for i := range want {
		if calls[i].action != want[i] {
			t.Errorf("push %d = %v, want %v", i, calls[i].action, want[i])
		}
	}
	if got := calls[2].args[0].(uint32); got != 145 {
		t.Errorf("wait deadline = %d, want 145", got)
	}
	if got := calls[3].args[0].(types.Pointf); got != (types.Pointf{X: 12.5, Y: -4.25}) {
		t.Errorf("face position = %v", got)
	}
	if unit.UpdateDataMonster().Field97 != 0 {
		t.Fatal("interesting sound flag was not consumed")
	}
}

func TestMonsterInterestingSound5466F0BlockedPath(t *testing.T) {
	unit := interestingSoundTestObject5466F0(t)
	var calls []monsterInterestingSoundCall5466F0
	hooks := interestingSoundTestHooks5466F0(&calls)
	hooks.trace = func(types.Pointf, types.Pointf) bool { return false }
	if got := monsterInterestingSound5466F0(unit, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []ai.ActionType{
		ai.DEPENDENCY_NO_INTERESTING_SOUND,
		ai.DEPENDENCY_NO_VISIBLE_ENEMY,
		ai.DEPENDENCY_NOT_FRUSTRATED,
		ai.DEPENDENCY_LOCATION_IS_SAFE,
		ai.ACTION_MOVE_TO,
	}
	if len(calls) != len(want) {
		t.Fatalf("push count = %d, want %d", len(calls), len(want))
	}
	for i := range want {
		if calls[i].action != want[i] {
			t.Errorf("push %d = %v, want %v", i, calls[i].action, want[i])
		}
	}
	if got := calls[4].args[1].(int); got != 0 {
		t.Errorf("move target object = %d, want 0", got)
	}
}

func TestMonsterInterestingSound5466F0Gates(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		unit := interestingSoundTestObject5466F0(t)
		unit.UpdateDataMonster().Field97 = 0
		var calls []monsterInterestingSoundCall5466F0
		if got := monsterInterestingSound5466F0(unit, interestingSoundTestHooks5466F0(&calls)); got != 0 || len(calls) != 0 {
			t.Fatalf("result/calls = %d/%d, want 0/0", got, len(calls))
		}
	})
	t.Run("expired", func(t *testing.T) {
		unit := interestingSoundTestObject5466F0(t)
		unit.UpdateDataMonster().Field101 = 10
		var calls []monsterInterestingSoundCall5466F0
		if got := monsterInterestingSound5466F0(unit, interestingSoundTestHooks5466F0(&calls)); got != 0 || len(calls) != 0 {
			t.Fatalf("result/calls = %d/%d, want 0/0", got, len(calls))
		}
	})
	t.Run("lava", func(t *testing.T) {
		unit := interestingSoundTestObject5466F0(t)
		var calls []monsterInterestingSoundCall5466F0
		hooks := interestingSoundTestHooks5466F0(&calls)
		hooks.tileAt = func(types.Pointf) int { return 6 }
		if got := monsterInterestingSound5466F0(unit, hooks); got != 1 || len(calls) != 0 || unit.UpdateDataMonster().Field97 != 0 {
			t.Fatalf("result/calls/flag = %d/%d/%d, want 1/0/0", got, len(calls), unit.UpdateDataMonster().Field97)
		}
	})
}
