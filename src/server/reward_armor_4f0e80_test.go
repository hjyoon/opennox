package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

type rewardArmorTestObjectRow4F0E80 struct {
	name    string
	weight  uint8
	typeInd uint32
	kind    uint32
	slots   uint32
	allowed bool
}

type rewardArmorTestModifierRow4F0E80 struct {
	name         string
	value        int
	group        uint8
	slots        uint32
	excludeArmor uint32
	allowArmor   uint32
	allowPos     uint32
}

type rewardArmorTestObject4F0E80 struct {
	typeInd   uint32
	applied   bool
	modifiers [4]int
}

type rewardArmorTestState4F0E80 struct {
	objects   []rewardArmorTestObjectRow4F0E80
	modifiers [3][]rewardArmorTestModifierRow4F0E80
	pickSlots func(uint32) uint32
	randomInt func(int32, int32) int32
	armorMask func(uint32) uint32
	create    func(uint32) *rewardArmorTestObject4F0E80
}

func rewardArmorTestHooks4F0E80(state *rewardArmorTestState4F0E80) rewardArmorHooks4F0E80[*rewardArmorTestObject4F0E80, int] {
	return rewardArmorHooks4F0E80[*rewardArmorTestObject4F0E80, int]{
		pickSlots: state.pickSlots,
		randomInt: state.randomInt,
		loadObjectName: func(index int) string {
			return state.objects[index].name
		},
		loadObjectWeight: func(index int) uint8 {
			return state.objects[index].weight
		},
		loadObjectTypeInd: func(index int) uint32 {
			return state.objects[index].typeInd
		},
		loadObjectKind: func(index int) uint32 {
			return state.objects[index].kind
		},
		loadObjectSlots: func(index int) uint32 {
			return state.objects[index].slots
		},
		objectTypeAllowed: func(typeInd uint32) bool {
			for index := range state.objects {
				if state.objects[index].typeInd == typeInd {
					return state.objects[index].allowed
				}
			}
			return false
		},
		armorTypeMask: state.armorMask,
		createObject:  state.create,
		isNilObject: func(object *rewardArmorTestObject4F0E80) bool {
			return object == nil
		},
		loadModifierName: func(table rewardArmorModifierTable4F0E80, index int) string {
			return state.modifiers[table][index].name
		},
		loadModifierValue: func(table rewardArmorModifierTable4F0E80, index int) int {
			return state.modifiers[table][index].value
		},
		loadModifierGroup: func(table rewardArmorModifierTable4F0E80, index int) uint8 {
			return state.modifiers[table][index].group
		},
		loadModifierSlots: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return state.modifiers[table][index].slots
		},
		loadModifierExcludeArmor: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return state.modifiers[table][index].excludeArmor
		},
		loadModifierAllowArmor: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return state.modifiers[table][index].allowArmor
		},
		loadModifierAllowPos: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return state.modifiers[table][index].allowPos
		},
		applyModifiers: func(object *rewardArmorTestObject4F0E80, modifiers [4]int) {
			object.applied = true
			object.modifiers = modifiers
		},
	}
}

func newRewardArmorTestState4F0E80() *rewardArmorTestState4F0E80 {
	state := &rewardArmorTestState4F0E80{
		objects: []rewardArmorTestObjectRow4F0E80{
			{name: "Armor", weight: 1, typeInd: 7, kind: rewardArmorObjectKind4F0E80, slots: 0x1f, allowed: true},
			{},
		},
		pickSlots: func(uint32) uint32 { return 1 },
		randomInt: func(minimum, _ int32) int32 {
			return minimum
		},
		armorMask: func(uint32) uint32 { return 0x40 },
		create: func(typeInd uint32) *rewardArmorTestObject4F0E80 {
			return &rewardArmorTestObject4F0E80{typeInd: typeInd}
		},
	}
	for table := range state.modifiers {
		state.modifiers[table] = []rewardArmorTestModifierRow4F0E80{{}}
	}
	return state
}

func TestRewardArmorObjectSelectionMatchesDefinitions4F0E80(t *testing.T) {
	state := newRewardArmorTestState4F0E80()
	state.objects = make([]rewardArmorTestObjectRow4F0E80, len(rewardObjectDefinitions4F0640))
	for index, row := range rewardObjectDefinitions4F0640 {
		state.objects[index] = rewardArmorTestObjectRow4F0E80{
			name: row.Name, weight: uint8(row.Weight), typeInd: uint32(index + 1),
			kind: row.Kind, slots: row.Slots, allowed: true,
		}
	}
	hooks := rewardArmorTestHooks4F0E80(state)
	for slotIndex := range 5 {
		slots := uint32(1) << slotIndex
		var total int32
		for rowIndex, row := range state.objects {
			if row.name == "" {
				break
			}
			if row.kind&rewardArmorObjectKind4F0E80 == 0 || row.slots&slots == 0 {
				continue
			}
			draw := total
			total = rewardArmorAddWeight4F0E80(total, row.weight)
			state.randomInt = func(minimum, maximum int32) int32 {
				if minimum != 0 {
					t.Fatalf("slot %d row %d minimum = %d, want 0", slotIndex, rowIndex, minimum)
				}
				return draw
			}
			hooks.randomInt = state.randomInt
			got, ok := rewardArmorSelectObject4F0E80(slots, hooks)
			if !ok || got != rowIndex {
				t.Fatalf("slot %d draw %d selected %d/%v, want row %d", slotIndex, draw, got, ok, rowIndex)
			}
		}
		if total == 0 {
			t.Fatalf("slot %d has no armor definitions", slotIndex)
		}
	}
}

func TestRewardArmorObjectSelectionLivePassAndFilters4F0E80(t *testing.T) {
	state := newRewardArmorTestState4F0E80()
	state.objects = []rewardArmorTestObjectRow4F0E80{
		{name: "Weapon", weight: 100, typeInd: 1, kind: 1, slots: 1, allowed: true},
		{name: "WrongSlot", weight: 100, typeInd: 2, kind: 2, slots: 2, allowed: true},
		{name: "Disallowed", weight: 100, typeInd: 3, kind: 2, slots: 1, allowed: false},
		{name: "First", weight: 2, typeInd: 4, kind: 2, slots: 1, allowed: true},
		{name: "Second", weight: 3, typeInd: 5, kind: 2, slots: 1, allowed: true},
		{},
	}
	rngCalls := 0
	state.randomInt = func(minimum, maximum int32) int32 {
		rngCalls++
		if minimum != 0 || maximum != 4 {
			t.Fatalf("RNG bounds = %d..%d, want 0..4", minimum, maximum)
		}
		state.objects[3].allowed = false
		state.objects[4].weight = 5
		return 2
	}
	hooks := rewardArmorTestHooks4F0E80(state)
	index, ok := rewardArmorSelectObject4F0E80(uint32(1), hooks)
	if !ok || index != 4 || rngCalls != 1 {
		t.Fatalf("selection/RNG = %d/%v/%d, want 4/true/1", index, ok, rngCalls)
	}

	state.objects[0].name = ""
	state.randomInt = func(int32, int32) int32 {
		t.Fatal("empty first row reached RNG")
		return 0
	}
	if _, ok := rewardArmorSelectObject4F0E80(uint32(1), rewardArmorTestHooks4F0E80(state)); ok {
		t.Fatal("empty first row selected an object")
	}
}

func TestRewardArmorWeightWrap4F0E80(t *testing.T) {
	if got, want := rewardArmorAddWeight4F0E80(math.MaxInt32, 1), int32(math.MinInt32); got != want {
		t.Fatalf("MaxInt32 + 1 = %d, want %d", got, want)
	}
	if got, want := rewardArmorAddWeight4F0E80(-2, 0xff), int32(253); got != want {
		t.Fatalf("-2 + 255 = %d, want %d", got, want)
	}
}

func TestRewardArmorModifierSelectionLivePassAndFilters4F0E80(t *testing.T) {
	state := newRewardArmorTestState4F0E80()
	state.modifiers[rewardArmorEnchantmentTable4F0E80] = []rewardArmorTestModifierRow4F0E80{
		{name: "WrongSlot", slots: 2, allowArmor: 0x40, allowPos: 1},
		{name: "WrongArmor", slots: 1, allowArmor: 0x20, allowPos: 1},
		{name: "Excluded", slots: 1, allowArmor: 0x40, excludeArmor: 0x40, allowPos: 1},
		{name: "WrongPosition", slots: 1, allowArmor: 0x40, allowPos: 2},
		{name: "FirstPass", slots: 1, allowArmor: 0x40, allowPos: 1},
		{},
	}
	state.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 0 {
			t.Fatalf("RNG bounds = %d..%d, want 0..0", minimum, maximum)
		}
		state.modifiers[rewardArmorEnchantmentTable4F0E80][3].allowPos = 1
		state.modifiers[rewardArmorEnchantmentTable4F0E80][4].allowArmor = 0
		return 0
	}
	index, ok := rewardArmorSelectModifier4F0E80(
		rewardArmorEnchantmentTable4F0E80,
		1,
		0x40,
		1,
		rewardArmorTestHooks4F0E80(state),
	)
	if !ok || index != 3 {
		t.Fatalf("selection = %d/%v, want live second-pass row 3", index, ok)
	}
}

func TestRewardArmorModifierCountRanges4F0E80(t *testing.T) {
	tests := []struct {
		slots      uint32
		minimum    int32
		maximum    int32
		recognized bool
		returned   int32
	}{
		{slots: 1},
		{slots: 2, minimum: 0, maximum: 1, recognized: true, returned: 1},
		{slots: 4, minimum: 0, maximum: 2, recognized: true, returned: 2},
		{slots: 8, minimum: 1, maximum: 3, recognized: true, returned: 3},
		{slots: 16, minimum: 2, maximum: 4, recognized: true, returned: 4},
		{slots: 32},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("slots_%d", tc.slots), func(t *testing.T) {
			calls := 0
			got, recognized := rewardArmorModifierCount4F0E80(tc.slots, func(minimum, maximum int32) int32 {
				calls++
				if minimum != tc.minimum || maximum != tc.maximum {
					t.Fatalf("bounds = %d..%d, want %d..%d", minimum, maximum, tc.minimum, tc.maximum)
				}
				return tc.returned
			})
			if got != tc.returned || recognized != tc.recognized {
				t.Fatalf("result = %d/%v, want %d/%v", got, recognized, tc.returned, tc.recognized)
			}
			wantCalls := 0
			if tc.recognized {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("RNG calls = %d, want %d", calls, wantCalls)
			}
		})
	}
}

func TestRewardArmorCategoryMaskBoundaries4F0E80(t *testing.T) {
	tests := []struct {
		count int32
		roll  int32
		want  uint16
	}{
		{1, 20, 4}, {1, 21, 1}, {1, 50, 1}, {1, 51, 2},
		{2, 12, 5}, {2, 13, 6}, {2, 25, 6}, {2, 26, 3},
		{3, 0, 7}, {4, 0, 15}, {5, 0, 0xbeef},
	}
	for _, tc := range tests {
		calls := 0
		got := rewardArmorCategoryMask4F0E80(tc.count, 0xcafebeef, func(minimum, maximum int32) int32 {
			calls++
			if minimum != 1 || maximum != 100 {
				t.Fatalf("count %d bounds = %d..%d", tc.count, minimum, maximum)
			}
			return tc.roll
		})
		if got != tc.want {
			t.Fatalf("count %d roll %d mask = %#x, want %#x", tc.count, tc.roll, got, tc.want)
		}
		wantCalls := 0
		if tc.count == 1 || tc.count == 2 {
			wantCalls = 1
		}
		if calls != wantCalls {
			t.Fatalf("count %d RNG calls = %d, want %d", tc.count, calls, wantCalls)
		}
	}
}

func TestRewardArmorPromotion4F0E80(t *testing.T) {
	makeModifier := func(available bool, position uint32) []rewardArmorTestModifierRow4F0E80 {
		if !available {
			return []rewardArmorTestModifierRow4F0E80{{}}
		}
		return []rewardArmorTestModifierRow4F0E80{
			{name: "Modifier", slots: 1, allowArmor: 0x40, allowPos: position},
			{},
		}
	}
	tests := []struct {
		name                       string
		mask                       uint16
		quality, material, enchant bool
		want                       uint16
	}{
		{"all_missing_cascade", 1, false, false, false, 0},
		{"quality_to_material", 1, false, true, false, 2},
		{"quality_material_to_enchant", 1, false, false, true, 4},
		{"keep_all", 7, true, true, true, 7},
		{"promote_fourth", 7, false, true, true, 14},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := newRewardArmorTestState4F0E80()
			state.modifiers[rewardArmorQualityTable4F0E80] = makeModifier(tc.quality, 0)
			state.modifiers[rewardArmorMaterialTable4F0E80] = makeModifier(tc.material, 0)
			state.modifiers[rewardArmorEnchantmentTable4F0E80] = makeModifier(tc.enchant, 1)
			got := rewardArmorPromoteMask4F0E80(tc.mask, 1, 0x40, rewardArmorTestHooks4F0E80(state))
			if got != tc.want {
				t.Fatalf("mask = %#x, want %#x", got, tc.want)
			}
		})
	}
}

func TestRewardArmorSecondEnchantmentSlotBounds4F0E80(t *testing.T) {
	tests := []struct {
		stage            uint32
		minimum, maximum int32
	}{
		{0, 1, 2}, {1, 1, 2}, {2, 1, 2}, {4, 1, 3}, {6, 2, 4},
		{8, 3, 4}, {10, 3, 4}, {math.MaxUint32, 3, 4},
	}
	for _, tc := range tests {
		got := rewardArmorSecondEnchantmentSlots4F0E80(tc.stage, func(minimum, maximum int32) int32 {
			if minimum != tc.minimum || maximum != tc.maximum {
				t.Fatalf("stage %#x bounds = %d..%d, want %d..%d", tc.stage, minimum, maximum, tc.minimum, tc.maximum)
			}
			return maximum
		})
		if got != uint32(tc.maximum) {
			t.Fatalf("stage %#x result = %d, want %d", tc.stage, got, tc.maximum)
		}
	}
}

func TestRewardArmorFullModifierPath4F0E80(t *testing.T) {
	state := newRewardArmorTestState4F0E80()
	state.pickSlots = func(stage uint32) uint32 {
		if stage != 8 {
			t.Fatalf("stage = %d, want 8", stage)
		}
		return 16
	}
	state.modifiers[rewardArmorQualityTable4F0E80] = []rewardArmorTestModifierRow4F0E80{
		{name: "Quality", value: 10, slots: 16, allowArmor: 0x40}, {},
	}
	state.modifiers[rewardArmorMaterialTable4F0E80] = []rewardArmorTestModifierRow4F0E80{
		{name: "Material", value: 20, slots: 16, allowArmor: 0x40}, {},
	}
	state.modifiers[rewardArmorEnchantmentTable4F0E80] = []rewardArmorTestModifierRow4F0E80{
		{name: "First", value: 30, group: 1, slots: 16, allowArmor: 0x40, allowPos: 1},
		{name: "Second", value: 40, group: 2, slots: 4, allowArmor: 0x40, allowPos: 2},
		{},
	}
	wantBounds := [][2]int32{{0, 0}, {2, 4}, {0, 0}, {0, 0}, {0, 0}, {3, 4}, {0, 0}}
	results := []int32{0, 4, 0, 0, 0, 4, 0}
	call := 0
	state.randomInt = func(minimum, maximum int32) int32 {
		if call >= len(wantBounds) || [2]int32{minimum, maximum} != wantBounds[call] {
			t.Fatalf("RNG call %d bounds = %d..%d, want sequence %v", call, minimum, maximum, wantBounds)
		}
		result := results[call]
		call++
		return result
	}
	got := rewardArmor4F0E80(uint32(8), rewardArmorTestHooks4F0E80(state))
	if got == nil || got.typeInd != 7 || !got.applied || got.modifiers != [4]int{10, 20, 30, 40} {
		t.Fatalf("result = %#v, want applied [10 20 30 40]", got)
	}
	if call != len(wantBounds) {
		t.Fatalf("RNG calls = %d, want %d", call, len(wantBounds))
	}
}

func TestRewardArmorDuplicateEnchantmentGroup4F0E80(t *testing.T) {
	state := newRewardArmorTestState4F0E80()
	state.pickSlots = func(uint32) uint32 { return 16 }
	state.modifiers[rewardArmorQualityTable4F0E80] = []rewardArmorTestModifierRow4F0E80{{}}
	state.modifiers[rewardArmorMaterialTable4F0E80] = []rewardArmorTestModifierRow4F0E80{{}}
	state.modifiers[rewardArmorEnchantmentTable4F0E80] = []rewardArmorTestModifierRow4F0E80{
		{name: "First", value: 30, group: 7, slots: 16, allowArmor: 0x40, allowPos: 1},
		{name: "Duplicate", value: 40, group: 7, slots: 4, allowArmor: 0x40, allowPos: 2},
		{},
	}
	results := []int32{0, 4, 0, 4, 0}
	state.randomInt = func(int32, int32) int32 {
		result := results[0]
		results = results[1:]
		return result
	}
	hooks := rewardArmorTestHooks4F0E80(state)
	valueLoads := 0
	originalLoad := hooks.loadModifierValue
	hooks.loadModifierValue = func(table rewardArmorModifierTable4F0E80, index int) int {
		valueLoads++
		return originalLoad(table, index)
	}
	got := rewardArmor4F0E80(uint32(8), hooks)
	if got == nil || !got.applied || got.modifiers != [4]int{0, 0, 30, 0} {
		t.Fatalf("result = %#v, want duplicate fourth modifier suppressed", got)
	}
	if valueLoads != 1 || len(results) != 0 {
		t.Fatalf("modifier value loads/remaining RNG = %d/%d, want 1/0", valueLoads, len(results))
	}
}

func TestRewardArmorLiveModifierSelectionStillApplies4F0E80(t *testing.T) {
	state := newRewardArmorTestState4F0E80()
	state.pickSlots = func(uint32) uint32 { return 2 }
	state.modifiers[rewardArmorQualityTable4F0E80] = []rewardArmorTestModifierRow4F0E80{
		{name: "Quality", value: 10, slots: 2, allowArmor: 0x40}, {},
	}
	results := []int32{0, 1, 30, 0}
	state.randomInt = func(int32, int32) int32 {
		result := results[0]
		results = results[1:]
		if len(results) == 0 {
			state.modifiers[rewardArmorQualityTable4F0E80][0].name = ""
		}
		return result
	}
	got := rewardArmor4F0E80(uint32(2), rewardArmorTestHooks4F0E80(state))
	if got == nil || !got.applied || got.modifiers != [4]int{} {
		t.Fatalf("result = %#v, want zero modifiers applied after live exhaustion", got)
	}
}

func TestRewardArmorEarlyReturnAndCallOrder4F0E80(t *testing.T) {
	state := newRewardArmorTestState4F0E80()
	var events []string
	state.pickSlots = func(uint32) uint32 {
		events = append(events, "slots")
		return 1
	}
	state.randomInt = func(minimum, maximum int32) int32 {
		events = append(events, fmt.Sprintf("rng:%d:%d", minimum, maximum))
		return 0
	}
	state.armorMask = func(typeInd uint32) uint32 {
		events = append(events, fmt.Sprintf("armor-mask:%d", typeInd))
		return 0x40
	}
	state.create = func(typeInd uint32) *rewardArmorTestObject4F0E80 {
		events = append(events, fmt.Sprintf("create:%d", typeInd))
		return &rewardArmorTestObject4F0E80{typeInd: typeInd}
	}
	got := rewardArmor4F0E80(uint32(99), rewardArmorTestHooks4F0E80(state))
	if got == nil || got.applied {
		t.Fatalf("slot-one result = %#v, want unmodified object", got)
	}
	want := []string{"slots", "rng:0:0", "armor-mask:7", "create:7"}
	if !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}

	state.create = func(uint32) *rewardArmorTestObject4F0E80 { return nil }
	state.pickSlots = func(uint32) uint32 { return 16 }
	rngCalls := 0
	state.randomInt = func(int32, int32) int32 {
		rngCalls++
		return 0
	}
	if got := rewardArmor4F0E80(uint32(8), rewardArmorTestHooks4F0E80(state)); got != nil || rngCalls != 1 {
		t.Fatalf("nil-create result/RNG = %#v/%d, want nil/one object draw", got, rngCalls)
	}
}
