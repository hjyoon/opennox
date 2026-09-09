package server

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterActionDodgeTestUnit544640(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	unit.SpeedBase = 1
	unit.SpeedCur = 3
	unit.PosVec = types.Ptf(100, 200)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_DODGE)
	update.AIStack[0].SetArgs(types.Ptf(112, 216))
	update.MonsterDef = &MonsterDef{RunMultiplier96: 1.75}
	return unit
}

func TestMonsterActionDodge544640StationaryPopsBeforeEnchant(t *testing.T) {
	unit := monsterActionDodgeTestUnit544640(t)
	unit.SpeedBase = math.Float32frombits(0x3c23d709) // immediately below 0.01f
	var events []string
	if !monsterActionDodge544640(unit, monsterActionDodgeHooks544640{
		hasEnchant: func(EnchantID) bool {
			events = append(events, "enchant")
			return false
		},
		pop: func() int {
			events = append(events, "pop")
			return 0
		},
	}) {
		t.Fatal("valid dodge action was rejected")
	}
	if !reflect.DeepEqual(events, []string{"pop"}) {
		t.Fatalf("events = %v, want [pop]", events)
	}
}

func TestMonsterActionDodge544640EnchantOrderAndShortCircuit(t *testing.T) {
	wantOrder := []EnchantID{ENCHANT_CONFUSED, ENCHANT_HELD, ENCHANT_CHARMING}
	for stop := range wantOrder {
		unit := monsterActionDodgeTestUnit544640(t)
		before := *unit
		var got []EnchantID
		pops := 0
		monsterActionDodge544640(unit, monsterActionDodgeHooks544640{
			hasEnchant: func(enchant EnchantID) bool {
				got = append(got, enchant)
				return enchant == wantOrder[stop]
			},
			pop: func() int { pops++; return 0 },
		})
		if !reflect.DeepEqual(got, wantOrder[:stop+1]) {
			t.Fatalf("stop %d: enchant order = %v, want %v", stop, got, wantOrder[:stop+1])
		}
		if pops != 0 || *unit != before {
			t.Fatalf("stop %d: enchanted dodge changed state or popped", stop)
		}
	}
}

func TestMonsterActionDodge544640DistanceBranches(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target types.Pointf
		pop    bool
	}{
		{name: "below-eight", target: types.Ptf(107, 200), pop: true},
		{name: "at-least-eight-after-epsilon", target: types.Ptf(108, 200), pop: false},
		{name: "unordered", target: types.Ptf(float32(math.NaN()), 200), pop: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := monsterActionDodgeTestUnit544640(t)
			unit.UpdateDataMonster().AIStack[0].SetArgs(tc.target)
			pops := 0
			monsterActionDodge544640(unit, monsterActionDodgeHooks544640{
				hasEnchant: func(EnchantID) bool { return false },
				pop:        func() int { pops++; return 0 },
			})
			if (pops != 0) != tc.pop {
				t.Fatalf("pop calls = %d, want pop=%v", pops, tc.pop)
			}
		})
	}
}

func TestMonsterActionDodge544640UpdatesNativeVelocity(t *testing.T) {
	unit := monsterActionDodgeTestUnit544640(t)
	pops := 0
	monsterActionDodge544640(unit, monsterActionDodgeHooks544640{
		hasEnchant: func(EnchantID) bool { return false },
		pop:        func() int { pops++; return 0 },
	})
	if pops != 0 {
		t.Fatalf("pop calls = %d, want 0", pops)
	}
	if got := math.Float32bits(unit.SpeedCur); got != 0x40a80000 {
		t.Fatalf("SpeedCur bits = %#08x, want 0x40a80000", got)
	}
	if got := math.Float32bits(unit.VelVec.X); got != 0x40499958 {
		t.Fatalf("VelVec.X bits = %#08x, want 0x40499958", got)
	}
	if got := math.Float32bits(unit.VelVec.Y); got != 0x4086663b {
		t.Fatalf("VelVec.Y bits = %#08x, want 0x4086663b", got)
	}
	if uintptr(unit.UpdateData) <= uintptr(^uint32(0)) {
		t.Fatalf("UpdateData pointer = %#x, want native high address", uintptr(unit.UpdateData))
	}
}

func TestMonsterActionDodge544640RejectsMalformedAdmission(t *testing.T) {
	unit := monsterActionDodgeTestUnit544640(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_HUNT)
	called := false
	if monsterActionDodge544640(unit, monsterActionDodgeHooks544640{
		hasEnchant: func(EnchantID) bool { called = true; return false },
		pop:        func() int { called = true; return 0 },
	}) {
		t.Fatal("wrong action head was accepted")
	}
	if called {
		t.Fatal("malformed admission invoked hooks")
	}
}
