package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

type rewardWeaponTestObjectRow4F14E0 struct {
	name    string
	weight  uint8
	typeInd uint32
	kind    uint32
	slots   uint32
	allowed bool
}

type rewardWeaponTestModifier4F14E0 struct {
	name          string
	group         uint8
	slots         uint32
	excludeWeapon uint32
	allowWeapons  uint32
	allowPos      uint32
}

type rewardWeaponTestObject4F14E0 struct {
	typeInd   uint32
	class     uint32
	subClass  uint32
	applied   bool
	modifiers [4]*rewardWeaponTestModifier4F14E0
}

type rewardWeaponTestState4F14E0 struct {
	objects       []rewardWeaponTestObjectRow4F14E0
	modifiers     [3][]*rewardWeaponTestModifier4F14E0
	pickSlots     func(uint32) uint32
	randomInt     func(int32, int32) int32
	weaponMask    func(uint32) uint32
	create        func(uint32) *rewardWeaponTestObject4F14E0
	replenishment func() *rewardWeaponTestModifier4F14E0
}

func rewardWeaponTestHooks4F14E0(
	state *rewardWeaponTestState4F14E0,
) rewardWeaponHooks4F14E0[*rewardWeaponTestObject4F14E0, *rewardWeaponTestModifier4F14E0] {
	return rewardWeaponHooks4F14E0[*rewardWeaponTestObject4F14E0, *rewardWeaponTestModifier4F14E0]{
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
		weaponTypeMask: state.weaponMask,
		createObject:   state.create,
		isNilObject: func(object *rewardWeaponTestObject4F14E0) bool {
			return object == nil
		},
		loadObjectClass: func(object *rewardWeaponTestObject4F14E0) uint32 {
			return object.class
		},
		loadObjectSubClass: func(object *rewardWeaponTestObject4F14E0) uint32 {
			return object.subClass
		},
		loadModifierName: func(table rewardWeaponModifierTable4F14E0, index int) string {
			return state.modifiers[table][index].name
		},
		loadModifierValue: func(table rewardWeaponModifierTable4F14E0, index int) *rewardWeaponTestModifier4F14E0 {
			return state.modifiers[table][index]
		},
		loadModifierGroup: func(table rewardWeaponModifierTable4F14E0, index int) uint8 {
			return state.modifiers[table][index].group
		},
		loadModifierSlots: func(table rewardWeaponModifierTable4F14E0, index int) uint32 {
			return state.modifiers[table][index].slots
		},
		loadModifierExcludeWeapon: func(table rewardWeaponModifierTable4F14E0, index int) uint32 {
			return state.modifiers[table][index].excludeWeapon
		},
		modifierAllowWeapons: func(modifier *rewardWeaponTestModifier4F14E0) uint32 {
			return modifier.allowWeapons
		},
		modifierAllowPos: func(modifier *rewardWeaponTestModifier4F14E0) uint32 {
			return modifier.allowPos
		},
		loadReplenishment: state.replenishment,
		applyModifiers: func(object *rewardWeaponTestObject4F14E0, modifiers [4]*rewardWeaponTestModifier4F14E0) {
			object.applied = true
			object.modifiers = modifiers
		},
	}
}

func newRewardWeaponTestState4F14E0() *rewardWeaponTestState4F14E0 {
	state := &rewardWeaponTestState4F14E0{
		objects: []rewardWeaponTestObjectRow4F14E0{
			{name: "Weapon", weight: 1, typeInd: 7, kind: rewardWeaponObjectKind4F14E0, slots: 0x1f, allowed: true},
			{},
		},
		pickSlots: func(uint32) uint32 { return 1 },
		randomInt: func(minimum, _ int32) int32 {
			return minimum
		},
		weaponMask: func(uint32) uint32 { return 0x40 },
		create: func(typeInd uint32) *rewardWeaponTestObject4F14E0 {
			return &rewardWeaponTestObject4F14E0{typeInd: typeInd}
		},
		replenishment: func() *rewardWeaponTestModifier4F14E0 { return nil },
	}
	for table := range state.modifiers {
		state.modifiers[table] = []*rewardWeaponTestModifier4F14E0{{}}
	}
	return state
}

func TestRewardWeaponObjectSelectionMatchesDefinitions4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.objects = make([]rewardWeaponTestObjectRow4F14E0, len(rewardObjectDefinitions4F0640))
	for index, row := range rewardObjectDefinitions4F0640 {
		state.objects[index] = rewardWeaponTestObjectRow4F14E0{
			name: row.Name, weight: uint8(row.Weight), typeInd: uint32(index + 1),
			kind: row.Kind, slots: row.Slots, allowed: true,
		}
	}
	hooks := rewardWeaponTestHooks4F14E0(state)
	for slotIndex := range 5 {
		slots := uint32(1) << slotIndex
		var total int32
		for rowIndex, row := range state.objects {
			if row.name == "" {
				break
			}
			if row.kind&rewardWeaponObjectKind4F14E0 == 0 || row.slots&slots == 0 {
				continue
			}
			draw := total
			total = rewardWeaponAddWeight4F14E0(total, row.weight)
			state.randomInt = func(minimum, maximum int32) int32 {
				if minimum != 0 {
					t.Fatalf("slot %d row %d minimum = %d, want 0", slotIndex, rowIndex, minimum)
				}
				return draw
			}
			hooks.randomInt = state.randomInt
			index, typeInd, ok := rewardWeaponSelectObject4F14E0(9, slots, hooks)
			if !ok || index != rowIndex || typeInd != row.typeInd {
				t.Fatalf("slot %d draw %d selected %d/%d/%v, want %d/%d/true", slotIndex, draw, index, typeInd, ok, rowIndex, row.typeInd)
			}
		}
		if total == 0 {
			t.Fatalf("slot %d has no weapon definitions", slotIndex)
		}
	}
}

func TestRewardWeaponObjectSelectionLivePassAndStageFallback4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.objects = []rewardWeaponTestObjectRow4F14E0{
		{name: "WrongKind", weight: 100, typeInd: 1, kind: 2, slots: 1, allowed: true},
		{name: "WrongSlot", weight: 100, typeInd: 2, kind: 1, slots: 2, allowed: true},
		{name: "Disallowed", weight: 100, typeInd: 3, kind: 1, slots: 1, allowed: false},
		{name: "First", weight: 2, typeInd: 4, kind: 1, slots: 1, allowed: true},
		{name: "Second", weight: 3, typeInd: 5, kind: 1, slots: 1, allowed: true},
		{},
	}
	state.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 4 {
			t.Fatalf("RNG bounds = %d..%d, want 0..4", minimum, maximum)
		}
		state.objects[3].allowed = false
		state.objects[4].weight = 5
		return 2
	}
	index, typeInd, ok := rewardWeaponSelectObject4F14E0(9, 1, rewardWeaponTestHooks4F14E0(state))
	if !ok || index != 4 || typeInd != 5 {
		t.Fatalf("live selection = %d/%d/%v, want 4/5/true", index, typeInd, ok)
	}

	state = newRewardWeaponTestState4F14E0()
	state.randomInt = func(int32, int32) int32 {
		state.objects[0].name = ""
		return 0
	}
	index, typeInd, ok = rewardWeaponSelectObject4F14E0(0x1234, 1, rewardWeaponTestHooks4F14E0(state))
	if !ok || index != 0 || typeInd != 0x1234 {
		t.Fatalf("empty live table fallback = %d/%#x/%v, want 0/0x1234/true", index, typeInd, ok)
	}

	state = newRewardWeaponTestState4F14E0()
	state.randomInt = func(int32, int32) int32 {
		state.objects[0].weight = 0
		return 1
	}
	index, typeInd, ok = rewardWeaponSelectObject4F14E0(0x5678, 1, rewardWeaponTestHooks4F14E0(state))
	if !ok || index != 1 || typeInd != 0x5678 {
		t.Fatalf("sentinel fallback = %d/%#x/%v, want 1/0x5678/true", index, typeInd, ok)
	}

	state = newRewardWeaponTestState4F14E0()
	state.randomInt = func(int32, int32) int32 {
		state.objects[0].weight = 0
		return 1
	}
	if _, _, ok := rewardWeaponSelectObject4F14E0(0, 1, rewardWeaponTestHooks4F14E0(state)); ok {
		t.Fatal("zero stage fallback selected a type")
	}
}

func TestRewardWeaponObjectSelectionEarlyReturns4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.objects[0].name = ""
	state.randomInt = func(int32, int32) int32 {
		t.Fatal("empty table reached RNG")
		return 0
	}
	if _, _, ok := rewardWeaponSelectObject4F14E0(1, 1, rewardWeaponTestHooks4F14E0(state)); ok {
		t.Fatal("empty table selected a type")
	}

	state = newRewardWeaponTestState4F14E0()
	state.objects[0].allowed = false
	state.randomInt = func(int32, int32) int32 {
		t.Fatal("zero total reached RNG")
		return 0
	}
	if _, _, ok := rewardWeaponSelectObject4F14E0(1, 1, rewardWeaponTestHooks4F14E0(state)); ok {
		t.Fatal("zero total selected a type")
	}
}

func TestRewardWeaponWeightWrap4F14E0(t *testing.T) {
	if got, want := rewardWeaponAddWeight4F14E0(math.MaxInt32, 1), int32(math.MinInt32); got != want {
		t.Fatalf("MaxInt32 + 1 = %d, want %d", got, want)
	}
	if got, want := rewardWeaponAddWeight4F14E0(-2, 0xff), int32(253); got != want {
		t.Fatalf("-2 + 255 = %d, want %d", got, want)
	}
}

func TestRewardWeaponModifierSelectionLivePassAndFilters4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.modifiers[rewardWeaponEnchantmentTable4F14E0] = []*rewardWeaponTestModifier4F14E0{
		{name: "WrongSlot", slots: 2, allowWeapons: 0x40, allowPos: 1},
		{name: "WrongWeapon", slots: 1, allowWeapons: 0x20, allowPos: 1},
		{name: "Excluded", slots: 1, allowWeapons: 0x40, excludeWeapon: 0x40, allowPos: 1},
		{name: "WrongPosition", slots: 1, allowWeapons: 0x40, allowPos: 2},
		{name: "FirstPass", slots: 1, allowWeapons: 0x40, allowPos: 1},
		{},
	}
	state.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 0 {
			t.Fatalf("RNG bounds = %d..%d, want 0..0", minimum, maximum)
		}
		state.modifiers[rewardWeaponEnchantmentTable4F14E0][3].allowPos = 1
		state.modifiers[rewardWeaponEnchantmentTable4F14E0][4].allowWeapons = 0
		return 0
	}
	index, ok := rewardWeaponSelectModifier4F14E0(
		rewardWeaponEnchantmentTable4F14E0, 1, 0x40, 1, rewardWeaponTestHooks4F14E0(state),
	)
	if !ok || index != 3 {
		t.Fatalf("selection = %d/%v, want live second-pass row 3", index, ok)
	}
}

func TestRewardWeaponModifierDescriptorFaultOrder4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.modifiers[rewardWeaponPowerTable4F14E0] = []*rewardWeaponTestModifier4F14E0{
		nil, {},
	}
	state.modifiers[rewardWeaponPowerTable4F14E0][1].name = ""
	hooks := rewardWeaponTestHooks4F14E0(state)

	state.modifiers[rewardWeaponPowerTable4F14E0][0] = nil
	hooks.loadModifierName = func(rewardWeaponModifierTable4F14E0, int) string { return "Modifier" }
	hooks.loadModifierSlots = func(rewardWeaponModifierTable4F14E0, int) uint32 { return 0 }
	if rewardWeaponModifierEligible4F14E0(rewardWeaponPowerTable4F14E0, 0, 1, 1, 0, hooks) {
		t.Fatal("wrong-slot nil descriptor was eligible")
	}

	hooks.loadModifierSlots = func(rewardWeaponModifierTable4F14E0, int) uint32 { return 1 }
	defer func() {
		if recover() == nil {
			t.Fatal("eligible nil descriptor did not fault after the slot load")
		}
	}()
	rewardWeaponModifierEligible4F14E0(rewardWeaponPowerTable4F14E0, 0, 1, 1, 0, hooks)
}

func TestRewardWeaponModifierCountRanges4F14E0(t *testing.T) {
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
			got, recognized := rewardWeaponModifierCount4F14E0(tc.slots, func(minimum, maximum int32) int32 {
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

func TestRewardWeaponCategoryMaskBoundaries4F14E0(t *testing.T) {
	tests := []struct {
		count int32
		roll  int32
		want  uint16
	}{
		{1, 20, 4}, {1, 21, 1}, {1, 50, 1}, {1, 51, 2},
		{2, 12, 5}, {2, 13, 6}, {2, 25, 6}, {2, 26, 3},
		{3, 0, 7}, {4, 0, 15}, {5, 0, 0},
	}
	for _, tc := range tests {
		calls := 0
		got := rewardWeaponCategoryMask4F14E0(tc.count, func(minimum, maximum int32) int32 {
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

func TestRewardWeaponPromotion4F14E0(t *testing.T) {
	makeModifier := func(available bool, position uint32) []*rewardWeaponTestModifier4F14E0 {
		if !available {
			return []*rewardWeaponTestModifier4F14E0{{}}
		}
		return []*rewardWeaponTestModifier4F14E0{
			{name: "Modifier", slots: 1, allowWeapons: 0x40, allowPos: position}, {},
		}
	}
	tests := []struct {
		name                     string
		mask                     uint16
		power, material, enchant bool
		want                     uint16
	}{
		{"all_missing_cascade", 1, false, false, false, 0},
		{"power_to_material", 1, false, true, false, 2},
		{"power_material_to_enchant", 1, false, false, true, 4},
		{"keep_all", 7, true, true, true, 7},
		{"promote_fourth", 7, false, true, true, 14},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := newRewardWeaponTestState4F14E0()
			state.modifiers[rewardWeaponPowerTable4F14E0] = makeModifier(tc.power, 0)
			state.modifiers[rewardWeaponMaterialTable4F14E0] = makeModifier(tc.material, 0)
			state.modifiers[rewardWeaponEnchantmentTable4F14E0] = makeModifier(tc.enchant, 1)
			got := rewardWeaponPromoteMask4F14E0(tc.mask, 1, 0x40, rewardWeaponTestHooks4F14E0(state))
			if got != tc.want {
				t.Fatalf("mask = %#x, want %#x", got, tc.want)
			}
		})
	}
}

func TestRewardWeaponEnchantmentSlotBounds4F14E0(t *testing.T) {
	tests := []struct {
		stage            uint32
		minimum, maximum int32
	}{
		{0, 1, 2}, {1, 1, 2}, {2, 1, 2}, {4, 1, 3}, {6, 2, 4},
		{8, 3, 4}, {10, 3, 4}, {math.MaxUint32, 3, 4},
	}
	for _, tc := range tests {
		got := rewardWeaponEnchantmentSlots4F14E0(tc.stage, func(minimum, maximum int32) int32 {
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

func TestRewardWeaponFullSelectionOrder4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.objects = []rewardWeaponTestObjectRow4F14E0{
		{name: "Weapon", weight: 1, typeInd: 7, kind: 1, slots: 16, allowed: true}, {},
	}
	power := &rewardWeaponTestModifier4F14E0{name: "Power", slots: 16, allowWeapons: 0x40}
	material := &rewardWeaponTestModifier4F14E0{name: "Material", slots: 16, allowWeapons: 0x40}
	first := &rewardWeaponTestModifier4F14E0{name: "First", group: 1, slots: 20, allowWeapons: 0x40, allowPos: 1}
	second := &rewardWeaponTestModifier4F14E0{name: "Second", group: 2, slots: 20, allowWeapons: 0x40, allowPos: 2}
	state.modifiers[rewardWeaponPowerTable4F14E0] = []*rewardWeaponTestModifier4F14E0{power, {}}
	state.modifiers[rewardWeaponMaterialTable4F14E0] = []*rewardWeaponTestModifier4F14E0{material, {}}
	state.modifiers[rewardWeaponEnchantmentTable4F14E0] = []*rewardWeaponTestModifier4F14E0{first, second, {}}

	wantBounds := [][2]int32{{0, 0}, {2, 4}, {0, 0}, {0, 0}, {3, 4}, {0, 0}, {3, 4}, {0, 0}}
	results := []int32{0, 4, 0, 0, 4, 0, 4, 0}
	var events []string
	call := 0
	created := &rewardWeaponTestObject4F14E0{typeInd: 7}
	state.pickSlots = func(stage uint32) uint32 {
		events = append(events, "slots")
		if stage != 8 {
			t.Fatalf("stage = %d, want 8", stage)
		}
		return 16
	}
	state.randomInt = func(minimum, maximum int32) int32 {
		events = append(events, "rng")
		if call >= len(wantBounds) || [2]int32{minimum, maximum} != wantBounds[call] {
			t.Fatalf("RNG call %d bounds = %d..%d, want %v", call, minimum, maximum, wantBounds)
		}
		result := results[call]
		call++
		return result
	}
	state.weaponMask = func(typeInd uint32) uint32 {
		events = append(events, "weapon-mask")
		if typeInd != 7 {
			t.Fatalf("weapon mask type = %d, want 7", typeInd)
		}
		return 0x40
	}
	state.create = func(typeInd uint32) *rewardWeaponTestObject4F14E0 {
		events = append(events, "create")
		if typeInd != 7 {
			t.Fatalf("create type = %d, want 7", typeInd)
		}
		return created
	}
	hooks := rewardWeaponTestHooks4F14E0(state)
	baseAllowed := hooks.objectTypeAllowed
	hooks.objectTypeAllowed = func(typeInd uint32) bool {
		events = append(events, "allowed")
		return baseAllowed(typeInd)
	}
	baseClass := hooks.loadObjectClass
	hooks.loadObjectClass = func(object *rewardWeaponTestObject4F14E0) uint32 {
		events = append(events, "class")
		return baseClass(object)
	}
	baseApply := hooks.applyModifiers
	hooks.applyModifiers = func(object *rewardWeaponTestObject4F14E0, modifiers [4]*rewardWeaponTestModifier4F14E0) {
		events = append(events, "apply")
		baseApply(object, modifiers)
	}
	got := rewardWeapon4F14E0(8, hooks)
	if got != created || !created.applied || created.modifiers != [4]*rewardWeaponTestModifier4F14E0{power, material, first, second} {
		t.Fatalf("result/attributes = %p/%v/%#v, want all four native modifiers", got, created.applied, created.modifiers)
	}
	if call != len(wantBounds) {
		t.Fatalf("RNG calls = %d, want %d", call, len(wantBounds))
	}
	wantPrefix := []string{"slots", "allowed", "rng", "allowed", "weapon-mask", "create", "class"}
	if len(events) < len(wantPrefix) || !slices.Equal(events[:len(wantPrefix)], wantPrefix) || events[len(events)-1] != "apply" {
		t.Fatalf("events = %v, want prefix %v and final apply", events, wantPrefix)
	}
}

func TestRewardWeaponReplenishingWand4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.objects[0].name = "#Wand"
	replenishment := &rewardWeaponTestModifier4F14E0{name: "Replenishment1"}
	created := &rewardWeaponTestObject4F14E0{
		typeInd: 7, class: rewardWeaponWandClass4F14E0, subClass: rewardWeaponWandSubMask4F14E0,
	}
	state.create = func(uint32) *rewardWeaponTestObject4F14E0 { return created }
	state.replenishment = func() *rewardWeaponTestModifier4F14E0 { return replenishment }
	rngCalls := 0
	baseRNG := state.randomInt
	state.randomInt = func(minimum, maximum int32) int32 {
		rngCalls++
		return baseRNG(minimum, maximum)
	}
	got := rewardWeapon4F14E0(1, rewardWeaponTestHooks4F14E0(state))
	if got != created || !created.applied || created.modifiers != [4]*rewardWeaponTestModifier4F14E0{nil, nil, replenishment, nil} {
		t.Fatalf("result/attributes = %p/%v/%#v, want replenishment in slot 2", got, created.applied, created.modifiers)
	}
	if rngCalls != 1 {
		t.Fatalf("wand RNG calls = %d, want object draw only", rngCalls)
	}

	state = newRewardWeaponTestState4F14E0()
	state.objects[0].name = "Wand"
	created = &rewardWeaponTestObject4F14E0{
		typeInd: 7, class: rewardWeaponWandClass4F14E0, subClass: rewardWeaponWandSubMask4F14E0,
	}
	state.create = func(uint32) *rewardWeaponTestObject4F14E0 { return created }
	state.replenishment = func() *rewardWeaponTestModifier4F14E0 {
		t.Fatal("plain wand loaded replenishment")
		return nil
	}
	got = rewardWeapon4F14E0(1, rewardWeaponTestHooks4F14E0(state))
	if got != created || created.applied {
		t.Fatal("plain special wand did not return unchanged")
	}
}

func TestRewardWeaponFallbackWandFaultBoundary4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.randomInt = func(int32, int32) int32 {
		state.objects[0].name = ""
		return 0
	}
	state.weaponMask = func(typeInd uint32) uint32 {
		if typeInd != 9 {
			t.Fatalf("fallback type = %d, want 9", typeInd)
		}
		return 0x40
	}
	created := &rewardWeaponTestObject4F14E0{class: rewardWeaponWandClass4F14E0, subClass: rewardWeaponWandSubMask4F14E0}
	state.create = func(uint32) *rewardWeaponTestObject4F14E0 { return created }
	defer func() {
		if recover() == nil || created.applied {
			t.Fatal("fallback wand did not fault on the live empty row name before modifier application")
		}
	}()
	rewardWeapon4F14E0(9, rewardWeaponTestHooks4F14E0(state))
}

func TestRewardWeaponEarlyReturnOrder4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	var events []string
	state.weaponMask = func(uint32) uint32 {
		events = append(events, "weapon-mask")
		return 0x40
	}
	state.create = func(uint32) *rewardWeaponTestObject4F14E0 {
		events = append(events, "create")
		return nil
	}
	if got := rewardWeapon4F14E0(1, rewardWeaponTestHooks4F14E0(state)); got != nil || !slices.Equal(events, []string{"weapon-mask", "create"}) {
		t.Fatalf("nil create result/events = %p/%v", got, events)
	}

	state = newRewardWeaponTestState4F14E0()
	created := &rewardWeaponTestObject4F14E0{}
	state.create = func(uint32) *rewardWeaponTestObject4F14E0 { return created }
	if got := rewardWeapon4F14E0(1, rewardWeaponTestHooks4F14E0(state)); got != created || created.applied {
		t.Fatal("unrecognized slot did not return the unmodified object")
	}
}

func TestRewardWeaponDuplicateEnchantmentGroupSuppressed4F14E0(t *testing.T) {
	state := newRewardWeaponTestState4F14E0()
	state.objects[0].slots = 16
	state.pickSlots = func(uint32) uint32 { return 16 }
	first := &rewardWeaponTestModifier4F14E0{name: "First", group: 3, slots: 20, allowWeapons: 0x40, allowPos: 1}
	second := &rewardWeaponTestModifier4F14E0{name: "Second", group: 3, slots: 20, allowWeapons: 0x40, allowPos: 2}
	state.modifiers[rewardWeaponPowerTable4F14E0] = []*rewardWeaponTestModifier4F14E0{{name: "Power", slots: 16, allowWeapons: 0x40}, {}}
	state.modifiers[rewardWeaponMaterialTable4F14E0] = []*rewardWeaponTestModifier4F14E0{{name: "Material", slots: 16, allowWeapons: 0x40}, {}}
	state.modifiers[rewardWeaponEnchantmentTable4F14E0] = []*rewardWeaponTestModifier4F14E0{first, second, {}}
	results := []int32{0, 4, 0, 0, 4, 0, 4, 0}
	state.randomInt = func(int32, int32) int32 {
		result := results[0]
		results = results[1:]
		return result
	}
	created := &rewardWeaponTestObject4F14E0{}
	state.create = func(uint32) *rewardWeaponTestObject4F14E0 { return created }
	rewardWeapon4F14E0(8, rewardWeaponTestHooks4F14E0(state))
	if !created.applied || created.modifiers[2] != first || created.modifiers[3] != nil || len(results) != 0 {
		t.Fatalf("duplicate group attributes/RNG = %#v/%d, want first only/0", created.modifiers, len(results))
	}
}
