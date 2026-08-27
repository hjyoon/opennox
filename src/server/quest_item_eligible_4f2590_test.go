package server

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type questItemTestModifier4F2590 struct {
	name         string
	allowWeapons uint32
	allowArmor   uint32
	allowPos     uint32
}

type questItemTestModifierData4F2590 struct {
	modifiers [4]*questItemTestModifier4F2590
}

type questItemTestObject4F2590 struct {
	class, subclass         uint32
	typeInd                 uint16
	spell, ability          uint8
	guide                   string
	weaponFlags, armorFlags uint32
	data                    *questItemTestModifierData4F2590
}

type questItemTestObjectRow4F2590 struct {
	name    string
	typeInd uint32
}

type questItemTestScalarRow4F2590 struct {
	id, slots uint32
}

type questItemTestModifierRow4F2590 struct {
	name                        string
	modifier                    *questItemTestModifier4F2590
	excludeArmor, excludeWeapon uint32
}

type questItemTestState4F2590 struct {
	events      []string
	typeIDs     map[string]uint32
	modifierIDs map[string]int32
	modifiers   map[int32]*questItemTestModifier4F2590
	guideIDs    map[string]uint32
	objects     []questItemTestObjectRow4F2590
	spells      []questItemTestScalarRow4F2590
	guides      []questItemTestScalarRow4F2590
	tables      map[questItemModifierTable4F2590][]questItemTestModifierRow4F2590
}

func newQuestItemTestState4F2590() *questItemTestState4F2590 {
	replenishments := [4]*questItemTestModifier4F2590{
		{name: "Replenishment1", allowPos: 3},
		{name: "Replenishment2", allowPos: 3},
		{name: "Replenishment3", allowPos: 3},
		{name: "Replenishment4", allowPos: 3},
	}
	state := &questItemTestState4F2590{
		typeIDs: map[string]uint32{
			"Diamond": 101, "Emerald": 102, "Ruby": 103,
			"SulphorousFlareWand": 104, "StreetSneakers": 105,
			"StreetShirt": 106, "StreetPants": 107,
		},
		modifierIDs: map[string]int32{},
		modifiers:   map[int32]*questItemTestModifier4F2590{},
		guideIDs:    map[string]uint32{"Urchin": 29},
		objects: []questItemTestObjectRow4F2590{
			{name: "Sword", typeInd: 200},
			{},
		},
		spells: []questItemTestScalarRow4F2590{{id: 7, slots: 1}, {}},
		guides: []questItemTestScalarRow4F2590{{id: 29, slots: 1}, {}},
		tables: map[questItemModifierTable4F2590][]questItemTestModifierRow4F2590{},
	}
	for index, modifier := range replenishments {
		id := int32(index + 1)
		state.modifierIDs[modifier.name] = id
		state.modifiers[id] = modifier
	}
	for table := questItemModifierWeaponPower4F2590; table <= questItemModifierEnchantments4F2590; table++ {
		state.tables[table] = []questItemTestModifierRow4F2590{{}}
	}
	return state
}

func (state *questItemTestState4F2590) event(format string, args ...any) {
	state.events = append(state.events, fmt.Sprintf(format, args...))
}

func (state *questItemTestState4F2590) hooks() questItemEligibilityHooks4F2590[
	*questItemTestObject4F2590,
	*questItemTestModifierData4F2590,
	*questItemTestModifier4F2590,
] {
	return questItemEligibilityHooks4F2590[
		*questItemTestObject4F2590,
		*questItemTestModifierData4F2590,
		*questItemTestModifier4F2590,
	]{
		objectTypeID: func(name string) uint32 {
			state.event("type:%s", name)
			return state.typeIDs[name]
		},
		modifierID: func(name string) int32 {
			state.event("modifier-id:%s", name)
			return state.modifierIDs[name]
		},
		modifierDesc: func(id int32) *questItemTestModifier4F2590 {
			state.event("modifier-desc:%d", id)
			return state.modifiers[id]
		},
		isNilModifier: func(modifier *questItemTestModifier4F2590) bool {
			return modifier == nil
		},
		loadClass: func(item *questItemTestObject4F2590) uint32 {
			state.event("class")
			return item.class
		},
		loadSubclass: func(item *questItemTestObject4F2590) uint32 {
			state.event("subclass")
			return item.subclass
		},
		loadTypeInd: func(item *questItemTestObject4F2590) uint16 {
			state.event("type-ind")
			return item.typeInd
		},
		loadUseSpell: func(item *questItemTestObject4F2590) uint8 {
			state.event("use-spell")
			return item.spell
		},
		loadUseGuideName: func(item *questItemTestObject4F2590) string {
			state.event("use-guide")
			return item.guide
		},
		loadUseAbility: func(item *questItemTestObject4F2590) uint8 {
			state.event("use-ability")
			return item.ability
		},
		guideID: func(name string) uint32 {
			state.event("guide-id:%s", name)
			return state.guideIDs[name]
		},
		loadModifierData: func(item *questItemTestObject4F2590) *questItemTestModifierData4F2590 {
			state.event("modifier-data")
			return item.data
		},
		loadModifier: func(data *questItemTestModifierData4F2590, slot int) *questItemTestModifier4F2590 {
			state.event("modifier:%d", slot)
			return data.modifiers[slot]
		},
		loadModifierName: func(modifier *questItemTestModifier4F2590) string {
			state.event("modifier-name:%s", modifier.name)
			return modifier.name
		},
		loadAllowWeapons: func(modifier *questItemTestModifier4F2590) uint32 {
			state.event("allow-weapons:%s", modifier.name)
			return modifier.allowWeapons
		},
		loadAllowArmor: func(modifier *questItemTestModifier4F2590) uint32 {
			state.event("allow-armor:%s", modifier.name)
			return modifier.allowArmor
		},
		loadAllowPos: func(modifier *questItemTestModifier4F2590) uint32 {
			state.event("allow-pos:%s", modifier.name)
			return modifier.allowPos
		},
		loadWeaponEquipFlags: func(item *questItemTestObject4F2590) uint32 {
			state.event("weapon-flags")
			return item.weaponFlags
		},
		loadArmorEquipFlags: func(item *questItemTestObject4F2590) uint32 {
			state.event("armor-flags")
			return item.armorFlags
		},
		loadRewardObjectName: func(index int) string {
			state.event("object-name:%d", index)
			return state.objects[index].name
		},
		loadRewardObjectTypeInd: func(index int) uint32 {
			state.event("object-type:%d", index)
			return state.objects[index].typeInd
		},
		loadRewardSpellID: func(index int) uint32 {
			state.event("spell-id:%d", index)
			return state.spells[index].id
		},
		loadRewardSpellSlots: func(index int) uint32 {
			state.event("spell-slots:%d", index)
			return state.spells[index].slots
		},
		loadRewardGuideID: func(index int) uint32 {
			state.event("guide-row-id:%d", index)
			return state.guides[index].id
		},
		loadRewardGuideSlots: func(index int) uint32 {
			state.event("guide-slots:%d", index)
			return state.guides[index].slots
		},
		loadModifierRowName: func(table questItemModifierTable4F2590, index int) string {
			state.event("row-name:%d:%d", table, index)
			return state.tables[table][index].name
		},
		loadModifierRowModifier: func(table questItemModifierTable4F2590, index int) *questItemTestModifier4F2590 {
			state.event("row-modifier:%d:%d", table, index)
			return state.tables[table][index].modifier
		},
		loadModifierRowExcludeArmor: func(table questItemModifierTable4F2590, index int) uint32 {
			state.event("exclude-armor:%d:%d", table, index)
			return state.tables[table][index].excludeArmor
		},
		loadModifierRowExcludeWeapon: func(table questItemModifierTable4F2590, index int) uint32 {
			state.event("exclude-weapon:%d:%d", table, index)
			return state.tables[table][index].excludeWeapon
		},
	}
}

func questItemTestCall4F2590(
	state *questItemTestState4F2590,
	item *questItemTestObject4F2590,
	cache *questItemEligibilityCache4F2590[*questItemTestModifier4F2590],
) int32 {
	if item.data == nil {
		item.data = &questItemTestModifierData4F2590{}
	}
	return questItemEligible4F2590(item, cache, state.hooks())
}

func TestQuestItemTypeCacheOrderAndRetry4F2590(t *testing.T) {
	state := newQuestItemTestState4F2590()
	state.typeIDs["Diamond"] = 0
	cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
	item := &questItemTestObject4F2590{class: questItemClassRejected4F2590}
	for call := 0; call < 2; call++ {
		if got := questItemTestCall4F2590(state, item, cache); got != 0 {
			t.Fatalf("rejected result = %d, want 0", got)
		}
	}
	var got []string
	for _, event := range state.events {
		if strings.HasPrefix(event, "type:") || event == "class" {
			got = append(got, event)
		}
	}
	wantOnce := []string{
		"type:Diamond", "type:Emerald", "type:Ruby", "type:SulphorousFlareWand",
		"type:StreetSneakers", "type:StreetShirt", "type:StreetPants", "class",
	}
	want := append(append([]string{}, wantOnce...), wantOnce...)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cache retry order = %v, want %v", got, want)
	}
}

func TestQuestItemClassAndSimpleTypeDispatch4F2590(t *testing.T) {
	tests := []struct {
		name string
		item questItemTestObject4F2590
		want int32
	}{
		{name: "rejected wins", item: questItemTestObject4F2590{class: 0x40 | 0x10, subclass: 0x1ff78}},
		{name: "direct subclass allowed", item: questItemTestObject4F2590{class: 0x10, subclass: 0x8}, want: 1},
		{name: "direct subclass denied", item: questItemTestObject4F2590{class: 0x10, subclass: 0x20000}},
		{name: "diamond", item: questItemTestObject4F2590{typeInd: 101}, want: 1},
		{name: "ruby", item: questItemTestObject4F2590{typeInd: 103}, want: 1},
		{name: "ordinary reward", item: questItemTestObject4F2590{typeInd: 200}, want: 1},
		{name: "unknown", item: questItemTestObject4F2590{typeInd: 201}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newQuestItemTestState4F2590()
			cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
			if got := questItemTestCall4F2590(state, &test.item, cache); got != test.want {
				t.Fatalf("eligibility = %d, want %d; events %v", got, test.want, state.events)
			}
		})
	}
}

func TestQuestBookItemEligibilityPathsAndPrecedence4F2700(t *testing.T) {
	tests := []struct {
		name string
		item questItemTestObject4F2590
		want int32
	}{
		{name: "spell", item: questItemTestObject4F2590{subclass: 1, spell: 7}, want: 1},
		{name: "spell absent", item: questItemTestObject4F2590{subclass: 1, spell: 8}},
		{name: "spell precedes guide", item: questItemTestObject4F2590{subclass: 3, spell: 8, guide: "Urchin"}},
		{name: "guide", item: questItemTestObject4F2590{subclass: 2, guide: "Urchin"}, want: 1},
		{name: "guide absent", item: questItemTestObject4F2590{subclass: 2, guide: "Unknown"}},
		{name: "ability one", item: questItemTestObject4F2590{subclass: 4, ability: 1}, want: 1},
		{name: "ability five", item: questItemTestObject4F2590{subclass: 4, ability: 5}, want: 1},
		{name: "ability zero", item: questItemTestObject4F2590{subclass: 4}},
		{name: "ability six", item: questItemTestObject4F2590{subclass: 4, ability: 6}},
		{name: "unsupported", item: questItemTestObject4F2590{subclass: 8}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newQuestItemTestState4F2590()
			if got := questBookItemEligible4F2700(&test.item, state.hooks()); got != test.want {
				t.Fatalf("eligibility = %d, want %d; events %v", got, test.want, state.events)
			}
			if test.item.subclass&1 != 0 {
				spellRow, use := -1, -1
				for index, event := range state.events {
					if event == "spell-id:0" && spellRow < 0 {
						spellRow = index
					}
					if event == "use-spell" && use < 0 {
						use = index
					}
				}
				if spellRow < 0 || use < 0 || spellRow > use {
					t.Fatalf("spell row/use order = %v", state.events)
				}
			}
		})
	}
}

func TestQuestEquipmentModifierValidation4F27A0(t *testing.T) {
	power := &questItemTestModifier4F2590{name: "Power", allowWeapons: 0x10}
	material := &questItemTestModifier4F2590{name: "Material", allowWeapons: 0x10}
	enchant2 := &questItemTestModifier4F2590{name: "Enchant2", allowWeapons: 0x10, allowPos: 1}
	enchant3 := &questItemTestModifier4F2590{name: "Enchant3", allowWeapons: 0x10, allowPos: 2}
	state := newQuestItemTestState4F2590()
	state.tables[questItemModifierWeaponPower4F2590] = []questItemTestModifierRow4F2590{{name: "Power", modifier: power}, {}}
	state.tables[questItemModifierMaterial4F2590] = []questItemTestModifierRow4F2590{{name: "Material", modifier: material}, {}}
	state.tables[questItemModifierEnchantments4F2590] = []questItemTestModifierRow4F2590{
		{name: "Enchant2", modifier: enchant2},
		{name: "Enchant3", modifier: enchant3},
		{},
	}
	item := &questItemTestObject4F2590{
		class:       questItemClassWeapon4F2590,
		typeInd:     200,
		weaponFlags: 0x10,
		data: &questItemTestModifierData4F2590{modifiers: [4]*questItemTestModifier4F2590{
			power, material, enchant2, enchant3,
		}},
	}
	cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
	if got := questItemEligible4F2590(item, cache, state.hooks()); got != 1 {
		t.Fatalf("valid weapon = %d, want 1; events %v", got, state.events)
	}

	state.events = nil
	state.tables[questItemModifierMaterial4F2590][0].excludeWeapon = 0x10
	if got := questItemEligible4F2590(item, cache, state.hooks()); got != 0 {
		t.Fatalf("excluded material = %d, want 0", got)
	}
	for _, event := range state.events {
		if event == "modifier:2" {
			t.Fatalf("material failure did not short-circuit enchantments: %v", state.events)
		}
	}

	state.events = nil
	state.tables[questItemModifierMaterial4F2590][0].excludeWeapon = 0
	power.allowWeapons = 0
	if got := questItemEligible4F2590(item, cache, state.hooks()); got != 0 {
		t.Fatalf("disallowed power = %d, want 0", got)
	}
	for _, event := range state.events {
		if strings.HasPrefix(event, "exclude-weapon:0:") {
			t.Fatalf("allow failure loaded exclusion despite C short circuit: %v", state.events)
		}
	}
}

func TestQuestArmorQualityAndMaterialValidation4F27E0(t *testing.T) {
	quality := &questItemTestModifier4F2590{name: "Quality", allowArmor: 0x20}
	material := &questItemTestModifier4F2590{name: "Material", allowArmor: 0x20}
	state := newQuestItemTestState4F2590()
	state.tables[questItemModifierArmorQuality4F2590] = []questItemTestModifierRow4F2590{{name: "Quality", modifier: quality}, {}}
	state.tables[questItemModifierMaterial4F2590] = []questItemTestModifierRow4F2590{{name: "Material", modifier: material}, {}}
	item := &questItemTestObject4F2590{
		class:      questItemClassArmor4F2590,
		typeInd:    200,
		armorFlags: 0x20,
		data: &questItemTestModifierData4F2590{modifiers: [4]*questItemTestModifier4F2590{
			quality, material, nil, nil,
		}},
	}
	cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
	if got := questItemEligible4F2590(item, cache, state.hooks()); got != 1 {
		t.Fatalf("valid armor = %d, want 1; events %v", got, state.events)
	}
	state.tables[questItemModifierArmorQuality4F2590][0].excludeArmor = 0x20
	if got := questItemEligible4F2590(item, cache, state.hooks()); got != 0 {
		t.Fatalf("excluded quality = %d, want 0", got)
	}
}

func TestQuestReplenishmentRequiresPositionAndHashRewardType4F2960(t *testing.T) {
	state := newQuestItemTestState4F2590()
	replenishment := state.modifiers[state.modifierIDs["Replenishment1"]]
	state.objects = []questItemTestObjectRow4F2590{
		{name: "Sword", typeInd: 200},
		{name: "#ForceWand", typeInd: 300},
		{name: "#OtherWand", typeInd: 301},
		{},
	}
	state.tables[questItemModifierEnchantments4F2590] = []questItemTestModifierRow4F2590{
		{name: "Replenishment1", modifier: replenishment},
		{},
	}
	item := &questItemTestObject4F2590{
		class:       questItemClassWeapon4F2590,
		typeInd:     300,
		weaponFlags: 0x10,
		data: &questItemTestModifierData4F2590{modifiers: [4]*questItemTestModifier4F2590{
			nil, nil, replenishment, nil,
		}},
	}
	cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
	if got := questEnchantmentModifiersEligible4F2960(item, cache, state.hooks()); got != 1 {
		t.Fatalf("hash reward replenishment = %d, want 1; events %v", got, state.events)
	}
	if !strings.Contains(strings.Join(state.events, ","), "object-name:3") {
		t.Fatalf("matching hash row did not continue through sentinel: %v", state.events)
	}
	state.events = nil
	item.typeInd = 302
	if got := questEnchantmentModifiersEligible4F2960(item, cache, state.hooks()); got != 0 {
		t.Fatalf("missing hash type = %d, want 0", got)
	}
	state.events = nil
	item.typeInd = 300
	replenishment.allowPos = 2
	if got := questEnchantmentModifiersEligible4F2960(item, cache, state.hooks()); got != 0 {
		t.Fatalf("wrong slot position = %d, want 0", got)
	}
}

func TestQuestReplenishmentCacheRetriesWhenFirstDescriptorNil4F2960(t *testing.T) {
	state := newQuestItemTestState4F2590()
	state.modifiers[state.modifierIDs["Replenishment1"]] = nil
	item := &questItemTestObject4F2590{
		class: questItemClassWeapon4F2590,
		data:  &questItemTestModifierData4F2590{},
	}
	cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
	for call := 0; call < 2; call++ {
		if got := questEnchantmentModifiersEligible4F2960(item, cache, state.hooks()); got != 1 {
			t.Fatalf("empty modifier call %d = %d, want 1", call, got)
		}
	}
	count := 0
	for _, event := range state.events {
		if strings.HasPrefix(event, "modifier-id:Replenishment") {
			count++
		}
	}
	if count != 8 {
		t.Fatalf("retry modifier lookup count = %d, want 8; events %v", count, state.events)
	}
}

func TestQuestDefaultWandAndStreetClothingPolicies4F2B60(t *testing.T) {
	state := newQuestItemTestState4F2590()
	replenishment := state.modifiers[state.modifierIDs["Replenishment1"]]
	cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
	wand := &questItemTestObject4F2590{
		class:       questItemClassWeapon4F2590,
		typeInd:     104,
		weaponFlags: questItemDefaultWandMask4F2590,
		data: &questItemTestModifierData4F2590{modifiers: [4]*questItemTestModifier4F2590{
			nil, nil, replenishment, nil,
		}},
	}
	if got := questItemEligible4F2590(wand, cache, state.hooks()); got != 1 {
		t.Fatalf("default wand = %d, want 1; events %v", got, state.events)
	}
	wand.data.modifiers[3] = &questItemTestModifier4F2590{name: "Extra"}
	if got := questItemEligible4F2590(wand, cache, state.hooks()); got != 0 {
		t.Fatalf("wand with fourth modifier = %d, want 0", got)
	}
	wand.weaponFlags = 0
	if got := questItemEligible4F2590(wand, cache, state.hooks()); got != 1 {
		t.Fatalf("non-flare weapon mask = %d, want 1", got)
	}

	street := &questItemTestObject4F2590{
		class:      questItemClassArmor4F2590,
		typeInd:    105,
		armorFlags: questItemDefaultArmorMask4F2590,
		data: &questItemTestModifierData4F2590{modifiers: [4]*questItemTestModifier4F2590{
			{name: "UserColor1"}, nil, {name: "uSeRcOlOX"}, nil,
		}},
	}
	if got := questItemEligible4F2590(street, cache, state.hooks()); got != 1 {
		t.Fatalf("Street clothing colors = %d, want 1; events %v", got, state.events)
	}
	street.data.modifiers[2].name = "Material1"
	if got := questItemEligible4F2590(street, cache, state.hooks()); got != 0 {
		t.Fatalf("Street clothing material = %d, want 0", got)
	}
}

func TestQuestItemASCIIUserColorPrefix4F2B60(t *testing.T) {
	for _, test := range []struct {
		value string
		want  bool
	}{
		{value: "UserColo", want: true},
		{value: "usercolor1", want: true},
		{value: "USERCOLO-anything", want: true},
		{value: "UserCol"},
		{value: "UserXolo"},
		{value: "ÜserColor"},
	} {
		if got := questItemASCIIEqualFoldPrefix4F2B60(test.value, "UserColo"); got != test.want {
			t.Fatalf("prefix %q = %t, want %t", test.value, got, test.want)
		}
	}
}

func TestQuestItemNilFaultOccursAfterTypeCache4F2590(t *testing.T) {
	state := newQuestItemTestState4F2590()
	cache := &questItemEligibilityCache4F2590[*questItemTestModifier4F2590]{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not fault")
		}
		if len(state.events) != len(questItemTypeNames4F2590)+1 || state.events[len(state.events)-1] != "class" {
			t.Fatalf("fault prefix = %v, want seven type lookups then class", state.events)
		}
	}()
	questItemEligible4F2590[*questItemTestObject4F2590, *questItemTestModifierData4F2590](nil, cache, state.hooks())
}

func TestQuestEquipmentNilModifierDataFault4F27E0(t *testing.T) {
	state := newQuestItemTestState4F2590()
	item := &questItemTestObject4F2590{class: questItemClassWeapon4F2590}
	defer func() {
		if recover() == nil {
			t.Fatal("nil modifier data did not fault")
		}
		want := []string{"modifier-data", "modifier:0"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("fault prefix = %v, want %v", state.events, want)
		}
	}()
	questFirstModifierEligible4F27E0(item, state.hooks())
}
