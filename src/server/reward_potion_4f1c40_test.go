package server

import (
	"math"
	"slices"
	"testing"
)

type rewardPotionTestRow4F1C40 struct {
	name    string
	weight  uint8
	typeInd uint32
	kind    uint8
	slots   uint32
	allowed bool
}

type rewardPotionTestObject4F1C40 struct {
	typeInd uint32
}

type rewardPotionTestState4F1C40 struct {
	rows      []rewardPotionTestRow4F1C40
	pickSlots func(uint32) uint32
	randomInt func(int32, int32) int32
	create    func(uint32) *rewardPotionTestObject4F1C40
}

func rewardPotionTestHooks4F1C40(state *rewardPotionTestState4F1C40) rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40] {
	return rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40]{
		pickSlots: state.pickSlots,
		loadObjectName: func(index int) string {
			return state.rows[index].name
		},
		loadObjectWeight: func(index int) uint8 {
			return state.rows[index].weight
		},
		loadObjectType: func(index int) uint32 {
			return state.rows[index].typeInd
		},
		loadObjectKind: func(index int) uint8 {
			return state.rows[index].kind
		},
		loadObjectSlots: func(index int) uint32 {
			return state.rows[index].slots
		},
		objectTypeAllowed: func(typeInd uint32) bool {
			for index := range state.rows {
				if state.rows[index].typeInd == typeInd {
					return state.rows[index].allowed
				}
			}
			return false
		},
		randomInt:    state.randomInt,
		createObject: state.create,
	}
}

func newRewardPotionTestState4F1C40() *rewardPotionTestState4F1C40 {
	return &rewardPotionTestState4F1C40{
		rows: []rewardPotionTestRow4F1C40{
			{name: "RedPotion", weight: 1, typeInd: 7, kind: rewardPotionObjectKind4F1C40, slots: 1, allowed: true},
			{},
		},
		pickSlots: func(uint32) uint32 { return 1 },
		randomInt: func(minimum, _ int32) int32 {
			return minimum
		},
		create: func(typeInd uint32) *rewardPotionTestObject4F1C40 {
			return &rewardPotionTestObject4F1C40{typeInd: typeInd}
		},
	}
}

func TestRewardPotionDefinitionsAndEverySelection4F1C40(t *testing.T) {
	state := newRewardPotionTestState4F1C40()
	state.rows = make([]rewardPotionTestRow4F1C40, len(rewardObjectDefinitions4F0640))
	for index, row := range rewardObjectDefinitions4F0640 {
		state.rows[index] = rewardPotionTestRow4F1C40{
			name: row.Name, weight: uint8(row.Weight), typeInd: uint32(index + 1),
			kind: uint8(row.Kind), slots: row.Slots, allowed: true,
		}
	}
	wantNames := []string{
		"RedPotion", "BluePotion", "CurePoisonPotion", "HastePotion",
		"InvisibilityPotion", "ShieldPotion", "VampirismPotion",
		"FireProtectPotion", "ShockProtectPotion", "PoisonProtectPotion",
		"InvulnerabilityPotion", "InfravisionPotion",
	}
	wantWeights := []uint8{12, 8, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	for slotIndex := range 5 {
		slots := uint32(1) << slotIndex
		state.pickSlots = func(uint32) uint32 { return slots }
		var total int32
		potionIndex := 0
		for rowIndex, row := range state.rows {
			if row.name == "" {
				break
			}
			if row.kind&rewardPotionObjectKind4F1C40 == 0 || row.slots&slots == 0 {
				continue
			}
			if potionIndex >= len(wantNames) || row.name != wantNames[potionIndex] || row.weight != wantWeights[potionIndex] {
				t.Fatalf("slot %d potion %d row = %q/%d", slotIndex, potionIndex, row.name, row.weight)
			}
			draw := total
			total = rewardPotionAddWeight4F1C40(total, row.weight)
			state.randomInt = func(minimum, maximum int32) int32 {
				if minimum != 0 || maximum != 32 {
					t.Fatalf("slot %d row %d RNG = %d..%d, want 0..32", slotIndex, rowIndex, minimum, maximum)
				}
				return draw
			}
			got := rewardPotion4F1C40(uint32(slotIndex), rewardPotionTestHooks4F1C40(state))
			if got == nil || got.typeInd != row.typeInd {
				t.Fatalf("slot %d row %d result = %#v, want type %d", slotIndex, rowIndex, got, row.typeInd)
			}
			potionIndex++
		}
		if potionIndex != 12 || total != 33 {
			t.Fatalf("slot %d potion count/weight = %d/%d, want 12/33", slotIndex, potionIndex, total)
		}
	}
}

func TestRewardPotionLiveSecondPassAndSelectedTypeReload4F1C40(t *testing.T) {
	state := newRewardPotionTestState4F1C40()
	state.rows = []rewardPotionTestRow4F1C40{
		{name: "First", weight: 2, typeInd: 7, kind: 4, slots: 1, allowed: true},
		{name: "Second", weight: 3, typeInd: 8, kind: 4, slots: 1, allowed: true},
		{},
	}
	state.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 4 {
			t.Fatalf("RNG = %d..%d, want 0..4", minimum, maximum)
		}
		state.rows[0].weight = 5
		return 3
	}
	hooks := rewardPotionTestHooks4F1C40(state)
	hooks.loadObjectWeight = func(index int) uint8 {
		weight := state.rows[index].weight
		if index == 0 && weight == 5 {
			state.rows[0].typeInd = 0xfeedbeef
		}
		return weight
	}
	hooks.objectTypeAllowed = func(uint32) bool { return true }
	got := rewardPotion4F1C40(99, hooks)
	if got == nil || got.typeInd != 0xfeedbeef {
		t.Fatalf("live result = %#v, want selected reloaded type 0xfeedbeef", got)
	}
}

func TestRewardPotionEligibilityShortCircuitsInOriginalOrder4F1C40(t *testing.T) {
	tests := []struct {
		name       string
		row        rewardPotionTestRow4F1C40
		wantEvents []string
	}{
		{
			name:       "kind",
			row:        rewardPotionTestRow4F1C40{name: "NotPotion", weight: 1, typeInd: 7, kind: 2, slots: 1, allowed: true},
			wantEvents: []string{"kind"},
		},
		{
			name:       "slots",
			row:        rewardPotionTestRow4F1C40{name: "WrongSlot", weight: 1, typeInd: 7, kind: 4, slots: 2, allowed: true},
			wantEvents: []string{"kind", "slots"},
		},
		{
			name:       "allowed",
			row:        rewardPotionTestRow4F1C40{name: "Disallowed", weight: 1, typeInd: 7, kind: 4, slots: 1},
			wantEvents: []string{"kind", "slots", "type", "allowed"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newRewardPotionTestState4F1C40()
			state.rows = []rewardPotionTestRow4F1C40{test.row, {}}
			hooks := rewardPotionTestHooks4F1C40(state)
			var events []string
			hooks.loadObjectKind = func(index int) uint8 {
				events = append(events, "kind")
				return state.rows[index].kind
			}
			hooks.loadObjectSlots = func(index int) uint32 {
				events = append(events, "slots")
				return state.rows[index].slots
			}
			hooks.loadObjectType = func(index int) uint32 {
				events = append(events, "type")
				return state.rows[index].typeInd
			}
			hooks.objectTypeAllowed = func(uint32) bool {
				events = append(events, "allowed")
				return false
			}
			hooks.loadObjectWeight = func(int) uint8 {
				t.Fatal("ineligible row loaded weight")
				return 0
			}
			hooks.randomInt = func(int32, int32) int32 {
				t.Fatal("ineligible-only table reached RNG")
				return 0
			}
			if got := rewardPotion4F1C40(0, hooks); got != nil || !slices.Equal(events, test.wantEvents) {
				t.Fatalf("result/events = %#v/%v, want nil/%v", got, events, test.wantEvents)
			}
		})
	}
}

func TestRewardPotionNilBranchesAndSignedArithmetic4F1C40(t *testing.T) {
	if got := rewardPotionAddWeight4F1C40(math.MaxInt32, 1); got != math.MinInt32 {
		t.Fatalf("wrapped add = %d, want %d", got, math.MinInt32)
	}
	if got := rewardPotionAddWeight4F1C40(-1, 1); got != 0 {
		t.Fatalf("wrapped zero add = %d, want 0", got)
	}

	tests := []struct {
		name  string
		setup func(*rewardPotionTestState4F1C40, *rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40])
	}{
		{
			name: "empty entry table",
			setup: func(state *rewardPotionTestState4F1C40, _ *rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40]) {
				state.rows = []rewardPotionTestRow4F1C40{{}}
				state.randomInt = func(int32, int32) int32 { t.Fatal("empty table reached RNG"); return 0 }
			},
		},
		{
			name: "zero eligible total",
			setup: func(state *rewardPotionTestState4F1C40, _ *rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40]) {
				state.rows[0].kind = 2
				state.randomInt = func(int32, int32) int32 { t.Fatal("zero total reached RNG"); return 0 }
			},
		},
		{
			name: "head removed after RNG",
			setup: func(state *rewardPotionTestState4F1C40, _ *rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40]) {
				state.randomInt = func(int32, int32) int32 { state.rows[0].name = ""; return 0 }
			},
		},
		{
			name: "second pass exhaustion",
			setup: func(state *rewardPotionTestState4F1C40, _ *rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40]) {
				state.randomInt = func(int32, int32) int32 { state.rows[0].allowed = false; return 0 }
			},
		},
		{
			name: "selected zero type",
			setup: func(state *rewardPotionTestState4F1C40, hooks *rewardPotionHooks4F1C40[*rewardPotionTestObject4F1C40]) {
				hooks.objectTypeAllowed = func(uint32) bool { return true }
				hooks.loadObjectWeight = func(int) uint8 { state.rows[0].typeInd = 0; return 1 }
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newRewardPotionTestState4F1C40()
			hooks := rewardPotionTestHooks4F1C40(state)
			test.setup(state, &hooks)
			hooks.randomInt = state.randomInt
			hooks.createObject = func(uint32) *rewardPotionTestObject4F1C40 {
				t.Fatal("nil branch created an object")
				return nil
			}
			if got := rewardPotion4F1C40(0, hooks); got != nil {
				t.Fatalf("result = %#v, want nil", got)
			}
		})
	}

	state := newRewardPotionTestState4F1C40()
	state.randomInt = func(int32, int32) int32 { return math.MinInt32 }
	if got := rewardPotion4F1C40(0, rewardPotionTestHooks4F1C40(state)); got == nil || got.typeInd != 7 {
		t.Fatalf("signed negative draw result = %#v, want first row", got)
	}
	state.create = func(uint32) *rewardPotionTestObject4F1C40 { return nil }
	if got := rewardPotion4F1C40(0, rewardPotionTestHooks4F1C40(state)); got != nil {
		t.Fatalf("nil factory result = %#v", got)
	}
}

func TestRewardPotionExactCallbackOrderAndFaultPrefixes4F1C40(t *testing.T) {
	run := func(failAt int) (events []string, result *rewardPotionTestObject4F1C40, panicValue any) {
		record := func(event string) {
			events = append(events, event)
			if len(events)-1 == failAt {
				panic("fault")
			}
		}
		state := newRewardPotionTestState4F1C40()
		hooks := rewardPotionTestHooks4F1C40(state)
		hooks.pickSlots = func(stage uint32) uint32 {
			record("slots")
			if stage != 0xfeedbeef {
				t.Fatalf("stage = %#x, want 0xfeedbeef", stage)
			}
			return 1
		}
		hooks.loadObjectName = func(index int) string { record("name"); return state.rows[index].name }
		hooks.loadObjectKind = func(index int) uint8 { record("kind"); return state.rows[index].kind }
		hooks.loadObjectSlots = func(index int) uint32 { record("row-slots"); return state.rows[index].slots }
		hooks.loadObjectType = func(index int) uint32 { record("type"); return state.rows[index].typeInd }
		hooks.objectTypeAllowed = func(uint32) bool { record("allowed"); return true }
		hooks.loadObjectWeight = func(index int) uint8 { record("weight"); return state.rows[index].weight }
		hooks.randomInt = func(int32, int32) int32 { record("rng"); return 0 }
		hooks.createObject = func(typeInd uint32) *rewardPotionTestObject4F1C40 {
			record("create")
			return &rewardPotionTestObject4F1C40{typeInd: typeInd}
		}
		defer func() { panicValue = recover() }()
		result = rewardPotion4F1C40(0xfeedbeef, hooks)
		return events, result, nil
	}
	want := []string{
		"slots", "name", "kind", "row-slots", "type", "allowed", "weight", "name",
		"rng", "name", "kind", "row-slots", "type", "allowed", "weight", "type", "create",
	}
	events, result, panicValue := run(-1)
	if panicValue != nil || result == nil || result.typeInd != 7 || !slices.Equal(events, want) {
		t.Fatalf("success result/events/panic = %#v/%v/%v, want type 7/%v/nil", result, events, panicValue, want)
	}
	for failAt := range want {
		events, result, panicValue = run(failAt)
		if panicValue == nil || result != nil || !slices.Equal(events, want[:failAt+1]) {
			t.Fatalf("fault %d result/events/panic = %#v/%v/%v, want nil/%v/non-nil", failAt, result, events, panicValue, want[:failAt+1])
		}
	}
}
