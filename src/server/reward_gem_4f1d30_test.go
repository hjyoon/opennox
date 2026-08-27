package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type rewardGemTestData4F1D30 struct {
	amount int32
}

type rewardGemTestObject4F1D30 struct {
	name string
	data *rewardGemTestData4F1D30
}

type rewardGemTestRandomCall4F1D30 struct {
	minimum int32
	maximum int32
	path    string
	line    int32
}

type rewardGemTestState4F1D30 struct {
	stage       uint32
	slots       uint32
	draws       []int32
	randomCalls []rewardGemTestRandomCall4F1D30
	created     []string
	events      []string
	factoryNil  bool
	boundMin    [5]int32
	boundMax    [5]int32
}

func newRewardGemTestState4F1D30() *rewardGemTestState4F1D30 {
	state := &rewardGemTestState4F1D30{}
	for index, bounds := range rewardGemGoldAmountBounds4F1D30 {
		state.boundMin[index] = bounds.Minimum
		state.boundMax[index] = bounds.Maximum
	}
	return state
}

func rewardGemTestHooks4F1D30(state *rewardGemTestState4F1D30) rewardGemHooks4F1D30[
	*rewardGemTestObject4F1D30,
	*rewardGemTestData4F1D30,
] {
	return rewardGemHooks4F1D30[*rewardGemTestObject4F1D30, *rewardGemTestData4F1D30]{
		pickSlots: func(stage uint32) uint32 {
			state.events = append(state.events, "slots")
			state.stage = stage
			return state.slots
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			state.events = append(state.events, fmt.Sprintf("rng:%d", line))
			state.randomCalls = append(state.randomCalls, rewardGemTestRandomCall4F1D30{
				minimum: minimum, maximum: maximum, path: path, line: line,
			})
			if len(state.draws) == 0 {
				panic("unexpected RNG call")
			}
			draw := state.draws[0]
			state.draws = state.draws[1:]
			return draw
		},
		createObjectByType: func(name string) *rewardGemTestObject4F1D30 {
			state.events = append(state.events, "create:"+name)
			state.created = append(state.created, name)
			if state.factoryNil {
				return nil
			}
			return &rewardGemTestObject4F1D30{name: name, data: &rewardGemTestData4F1D30{}}
		},
		isNilObject: func(object *rewardGemTestObject4F1D30) bool {
			state.events = append(state.events, "nil")
			return object == nil
		},
		loadInitData: func(object *rewardGemTestObject4F1D30) *rewardGemTestData4F1D30 {
			state.events = append(state.events, "init")
			return object.data
		},
		loadGoldMaximum: func(index rewardGemGoldRange4F1D30) int32 {
			state.events = append(state.events, fmt.Sprintf("max:%d", index))
			return state.boundMax[index]
		},
		loadGoldMinimum: func(index rewardGemGoldRange4F1D30) int32 {
			state.events = append(state.events, fmt.Sprintf("min:%d", index))
			return state.boundMin[index]
		},
		storeGoldAmount: func(data *rewardGemTestData4F1D30, amount int32) {
			state.events = append(state.events, "store")
			data.amount = amount
		},
	}
}

func TestRewardGemEveryGoldRange4F1D30(t *testing.T) {
	tests := []struct {
		slots uint32
		index rewardGemGoldRange4F1D30
		line  int32
	}{
		{slots: 0, index: rewardGemGoldDefault4F1D30, line: 2226},
		{slots: 2, index: rewardGemGoldSlot2_4F1D30, line: 2222},
		{slots: 4, index: rewardGemGoldSlot4_4F1D30, line: 2219},
		{slots: 8, index: rewardGemGoldSlot8_4F1D30, line: 2216},
		{slots: 16, index: rewardGemGoldSlot16_4F1D30, line: 2213},
		{slots: math.MaxUint32, index: rewardGemGoldDefault4F1D30, line: 2226},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("slots_%08x", test.slots), func(t *testing.T) {
			state := newRewardGemTestState4F1D30()
			state.slots = test.slots
			if test.slots >= 4 {
				state.draws = append(state.draws, 90)
			}
			state.draws = append(state.draws, 1, -7)
			got := rewardGem4F1D30(0xfeedbeef, rewardGemTestHooks4F1D30(state))
			if got == nil || got.name != rewardGemGoldChestType4F1D30 || got.data.amount != -7 {
				t.Fatalf("result = %#v, want chest amount -7", got)
			}
			if state.stage != 0xfeedbeef || len(state.draws) != 0 {
				t.Fatalf("stage/remaining draws = %#x/%v", state.stage, state.draws)
			}
			last := state.randomCalls[len(state.randomCalls)-1]
			bounds := rewardGemGoldAmountBounds4F1D30[test.index]
			if last.minimum != bounds.Minimum || last.maximum != bounds.Maximum ||
				last.path != rewardGemRandomPath4F1D30 || last.line != test.line {
				t.Fatalf("amount RNG = %#v, want %d..%d/path/line %d", last, bounds.Minimum, bounds.Maximum, test.line)
			}
			maxEvent := fmt.Sprintf("max:%d", test.index)
			minEvent := fmt.Sprintf("min:%d", test.index)
			maxPos, minPos := -1, -1
			for index, event := range state.events {
				if event == maxEvent {
					maxPos = index
				}
				if event == minEvent {
					minPos = index
				}
			}
			if maxPos < 0 || minPos != maxPos+1 {
				t.Fatalf("bound load order = %v, want max immediately before min", state.events)
			}
		})
	}
}

func TestRewardGemThresholdsAndGoldObjectChoice4F1D30(t *testing.T) {
	gemTests := []struct {
		draw int32
		want string
	}{
		{draw: math.MinInt32, want: rewardGemRubyType4F1D30},
		{draw: 49, want: rewardGemRubyType4F1D30},
		{draw: 50, want: rewardGemEmeraldType4F1D30},
		{draw: 89, want: rewardGemEmeraldType4F1D30},
		{draw: 90, want: rewardGemDiamondType4F1D30},
		{draw: math.MaxInt32, want: rewardGemDiamondType4F1D30},
	}
	for _, test := range gemTests {
		state := newRewardGemTestState4F1D30()
		state.slots = 4
		state.draws = []int32{91, test.draw}
		got := rewardGem4F1D30(7, rewardGemTestHooks4F1D30(state))
		if got == nil || got.name != test.want {
			t.Fatalf("gem draw %d result = %#v, want %s", test.draw, got, test.want)
		}
		if !reflect.DeepEqual(state.events, []string{
			"slots", "rng:2181", "rng:2185", "create:" + test.want,
		}) {
			t.Fatalf("gem events = %v", state.events)
		}
	}

	state := newRewardGemTestState4F1D30()
	state.slots = 1
	state.draws = []int32{2, 100}
	got := rewardGem4F1D30(0, rewardGemTestHooks4F1D30(state))
	if got == nil || got.name != rewardGemGoldPileType4F1D30 || got.data.amount != 100 {
		t.Fatalf("gold pile result = %#v", got)
	}
	if state.randomCalls[0].line != rewardGemGoldTypeLine4F1D30 {
		t.Fatalf("slot below four did not skip gate: %#v", state.randomCalls)
	}
}

func TestRewardGemNilBranchesAndLiveBoundLoads4F1D30(t *testing.T) {
	t.Run("gold nil checks and returns", func(t *testing.T) {
		state := newRewardGemTestState4F1D30()
		state.slots = 2
		state.draws = []int32{1}
		state.factoryNil = true
		if got := rewardGem4F1D30(0, rewardGemTestHooks4F1D30(state)); got != nil {
			t.Fatalf("nil gold result = %#v", got)
		}
		want := []string{"slots", "rng:2197", "create:QuestGoldChest", "nil"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("nil gold events = %v, want %v", state.events, want)
		}
	})

	t.Run("gem returns factory nil without checking", func(t *testing.T) {
		state := newRewardGemTestState4F1D30()
		state.slots = 4
		state.draws = []int32{91, 49}
		state.factoryNil = true
		hooks := rewardGemTestHooks4F1D30(state)
		hooks.isNilObject = func(*rewardGemTestObject4F1D30) bool {
			t.Fatal("gem branch checked nil result")
			return false
		}
		if got := rewardGem4F1D30(0, hooks); got != nil {
			t.Fatalf("nil gem result = %#v", got)
		}
	})

	t.Run("minimum is live after maximum load", func(t *testing.T) {
		state := newRewardGemTestState4F1D30()
		state.slots = 2
		state.draws = []int32{1, 333}
		hooks := rewardGemTestHooks4F1D30(state)
		hooks.loadGoldMaximum = func(index rewardGemGoldRange4F1D30) int32 {
			state.boundMin[index] = 333
			return state.boundMax[index]
		}
		got := rewardGem4F1D30(0, hooks)
		if got == nil || got.data.amount != 333 || state.randomCalls[1].minimum != 333 {
			t.Fatalf("live minimum result/call = %#v/%#v", got, state.randomCalls)
		}
	})
}

func TestRewardGemExactGoldCallbackOrderAndFaultPrefixes4F1D30(t *testing.T) {
	want := []string{
		"slots", "rng:2181", "rng:2197", "create:QuestGoldChest", "nil",
		"init", "max:2", "min:2", "rng:2219", "store",
	}
	run := func(failAt int) (events []string, result *rewardGemTestObject4F1D30, fault any) {
		state := newRewardGemTestState4F1D30()
		state.slots = 4
		state.draws = []int32{90, 1, 400}
		hooks := rewardGemTestHooks4F1D30(state)
		record := func(event string) {
			state.events = append(state.events, event)
			if len(state.events)-1 == failAt {
				panic("fault")
			}
		}
		hooks.pickSlots = func(uint32) uint32 { record("slots"); return 4 }
		hooks.randomInt = func(minimum, maximum int32, _ string, line int32) int32 {
			record(fmt.Sprintf("rng:%d", line))
			draw := state.draws[0]
			state.draws = state.draws[1:]
			return draw
		}
		hooks.createObjectByType = func(name string) *rewardGemTestObject4F1D30 {
			record("create:" + name)
			return &rewardGemTestObject4F1D30{name: name, data: &rewardGemTestData4F1D30{}}
		}
		hooks.isNilObject = func(*rewardGemTestObject4F1D30) bool { record("nil"); return false }
		hooks.loadInitData = func(object *rewardGemTestObject4F1D30) *rewardGemTestData4F1D30 { record("init"); return object.data }
		hooks.loadGoldMaximum = func(rewardGemGoldRange4F1D30) int32 { record("max:2"); return 1000 }
		hooks.loadGoldMinimum = func(rewardGemGoldRange4F1D30) int32 { record("min:2"); return 400 }
		hooks.storeGoldAmount = func(data *rewardGemTestData4F1D30, amount int32) { record("store"); data.amount = amount }
		defer func() { fault = recover(); events = state.events }()
		result = rewardGem4F1D30(0, hooks)
		return events, result, nil
	}

	events, result, fault := run(-1)
	if fault != nil || result == nil || !reflect.DeepEqual(events, want) {
		t.Fatalf("full result/events/fault = %#v/%v/%v, want object/%v/nil", result, events, fault, want)
	}
	for failAt := range want {
		events, result, fault := run(failAt)
		if fault != "fault" || result != nil || !reflect.DeepEqual(events, want[:failAt+1]) {
			t.Fatalf("fault %d result/events/fault = %#v/%v/%v, want nil/%v/fault", failAt, result, events, fault, want[:failAt+1])
		}
	}
}

func TestRewardGemSecondCreatorIsExactWrapper4F1F00(t *testing.T) {
	state := newRewardGemTestState4F1D30()
	state.slots = 2
	state.draws = []int32{1, 200}
	got := rewardGem2_4F1F00(math.MaxUint32, rewardGemTestHooks4F1D30(state))
	if got == nil || got.name != rewardGemGoldChestType4F1D30 || got.data.amount != 200 || state.stage != math.MaxUint32 {
		t.Fatalf("wrapper result/stage = %#v/%#x", got, state.stage)
	}
}
