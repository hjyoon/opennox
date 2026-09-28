package server

import (
	"reflect"
	"testing"
)

type weaponCreateTestObject54C710 struct {
	typeInd  uint16
	initData *weaponCreateTestInit54C710
	health   *weaponCreateTestHealth54C710
	class    uint32
	subclass uint32
	useData  *weaponCreateTestUse54C710
}

type weaponCreateTestDefinition54C710 struct {
	durability uint16
}

type weaponCreateTestHealth54C710 struct {
	current uint16
	maximum uint16
}

type weaponCreateTestEffect54C710 struct {
	name string
}

type weaponCreateTestInit54C710 struct {
	modifiers [4]*weaponCreateTestEffect54C710
}

type weaponCreateTestUse54C710 [116]byte

func weaponCreateTestHooksFor54C710(
	obj *weaponCreateTestObject54C710,
	definition *weaponCreateTestDefinition54C710,
	cache *weaponCreateTypeCache54C710,
	quest bool,
) weaponCreateHooks54C710[
	*weaponCreateTestObject54C710,
	*weaponCreateTestDefinition54C710,
	*weaponCreateTestHealth54C710,
	*weaponCreateTestInit54C710,
	*weaponCreateTestEffect54C710,
	*weaponCreateTestUse54C710,
] {
	modifierIDs := map[string]int{
		weaponCreateHeartModifier54C710:     1,
		weaponCreateWierdlingModifier54C710: 2,
		weaponCreateWierdlingSecond54C710:   3,
	}
	modifierNames := map[int]string{
		1: weaponCreateHeartModifier54C710,
		2: weaponCreateWierdlingModifier54C710,
		3: weaponCreateWierdlingSecond54C710,
	}
	return weaponCreateHooks54C710[
		*weaponCreateTestObject54C710,
		*weaponCreateTestDefinition54C710,
		*weaponCreateTestHealth54C710,
		*weaponCreateTestInit54C710,
		*weaponCreateTestEffect54C710,
		*weaponCreateTestUse54C710,
	]{
		loadTypeInd: func(got *weaponCreateTestObject54C710) uint16 {
			return got.typeInd
		},
		loadInitData: func(got *weaponCreateTestObject54C710) *weaponCreateTestInit54C710 {
			return got.initData
		},
		findDefinition: func(uint16) *weaponCreateTestDefinition54C710 {
			return definition
		},
		loadHeartType: func() uint32 {
			return cache.heart
		},
		storeHeartType: func(value uint32) {
			cache.heart = value
		},
		loadWierdlingType: func() uint32 {
			return cache.wierdling
		},
		storeWierdlingType: func(value uint32) {
			cache.wierdling = value
		},
		storeOrbType: func(value uint32) {
			cache.orb = value
		},
		resolveObjectType: func(name string) uint32 {
			return map[string]uint32{
				weaponCreateHeartType54C710:     11,
				weaponCreateWierdlingType54C710: 12,
				weaponCreateOrbType54C710:       13,
			}[name]
		},
		loadHealth: func(got *weaponCreateTestObject54C710) *weaponCreateTestHealth54C710 {
			return got.health
		},
		loadDurability: func(got *weaponCreateTestDefinition54C710) uint16 {
			return got.durability
		},
		storeCurrent: func(health *weaponCreateTestHealth54C710, value uint16) {
			health.current = value
		},
		storeMaximum: func(health *weaponCreateTestHealth54C710, value uint16) {
			health.maximum = value
		},
		loadCurrent: func(health *weaponCreateTestHealth54C710) uint16 {
			return health.current
		},
		loadMaximum: func(health *weaponCreateTestHealth54C710) uint16 {
			return health.maximum
		},
		modifierID: func(name string) int {
			return modifierIDs[name]
		},
		modifierDesc: func(id int) *weaponCreateTestEffect54C710 {
			return &weaponCreateTestEffect54C710{name: modifierNames[id]}
		},
		storeModifier: func(data *weaponCreateTestInit54C710, index int, modifier *weaponCreateTestEffect54C710) {
			data.modifiers[index] = modifier
		},
		loadClass: func(got *weaponCreateTestObject54C710) uint32 {
			return got.class
		},
		loadSubclass: func(got *weaponCreateTestObject54C710) uint32 {
			return got.subclass
		},
		loadUseData: func(got *weaponCreateTestObject54C710) *weaponCreateTestUse54C710 {
			return got.useData
		},
		loadUseByte: func(data *weaponCreateTestUse54C710, offset uintptr) uint8 {
			return data[offset]
		},
		storeUseByte: func(data *weaponCreateTestUse54C710, offset uintptr, value uint8) {
			data[offset] = value
		},
		gameFlag: func(flag uint32) int32 {
			if flag != weaponCreateQuestFlag54C710 {
				panic("unexpected game flag")
			}
			if quest {
				return 1
			}
			return 0
		},
		loadBalance: func(key string) float32 {
			switch key {
			case weaponCreateDurabilityKey54C710:
				return 1.5
			case weaponCreateAmmoQuestKey54C710:
				return 6.5
			case weaponCreateAmmoDefaultKey54C710:
				return 7
			case weaponCreateStaffChargeKey54C710:
				return 1.5
			default:
				panic("unexpected balance key: " + key)
			}
		},
		floatToInt: armorCreateRoundFloat32ToInt32_54C950,
	}
}

func TestWeaponCreate54C710QuestWierdlingAmmoAndStaff(t *testing.T) {
	initialInit := new(weaponCreateTestInit54C710)
	replacementInit := new(weaponCreateTestInit54C710)
	useData := new(weaponCreateTestUse54C710)
	useData[weaponCreateWandCharge54C710] = 3
	useData[weaponCreateWandMaxCharge54C710] = 5
	obj := &weaponCreateTestObject54C710{
		typeInd:  12,
		initData: initialInit,
		health:   new(weaponCreateTestHealth54C710),
		class:    weaponCreateAmmoClass54C710 | weaponCreateStaffClass54C710,
		// Both ammo branches match; the charged branch must win. The staff
		// double-charge bit also exercises the Quest wand path.
		subclass: weaponCreateAmmoChargeMask54C710 | weaponCreateAmmoEmptyMask54C710 |
			weaponCreateDoubleChargeMask54C710,
		useData: useData,
	}
	definition := &weaponCreateTestDefinition54C710{durability: 3}
	cache := new(weaponCreateTypeCache54C710)
	hooks := weaponCreateTestHooksFor54C710(obj, definition, cache, true)

	var resolved []string
	resolve := hooks.resolveObjectType
	hooks.resolveObjectType = func(name string) uint32 {
		resolved = append(resolved, name)
		return resolve(name)
	}
	findDefinition := hooks.findDefinition
	hooks.findDefinition = func(typeInd uint16) *weaponCreateTestDefinition54C710 {
		// GAME.EXE caches InitData before this lookup.
		obj.initData = replacementInit
		return findDefinition(typeInd)
	}
	var stores []struct {
		offset uintptr
		value  uint8
	}
	storeUseByte := hooks.storeUseByte
	hooks.storeUseByte = func(data *weaponCreateTestUse54C710, offset uintptr, value uint8) {
		stores = append(stores, struct {
			offset uintptr
			value  uint8
		}{offset: offset, value: value})
		storeUseByte(data, offset, value)
	}

	weaponCreate54C710(obj, hooks)

	if want := []string{
		weaponCreateHeartType54C710,
		weaponCreateWierdlingType54C710,
		weaponCreateOrbType54C710,
	}; !reflect.DeepEqual(resolved, want) {
		t.Fatalf("resolved types = %v, want %v", resolved, want)
	}
	if *cache != (weaponCreateTypeCache54C710{heart: 11, wierdling: 12, orb: 13}) {
		t.Fatalf("type cache = %+v", *cache)
	}
	if obj.health.current != 4 || obj.health.maximum != 4 {
		t.Fatalf("Quest durability = %d/%d, want 4/4", obj.health.current, obj.health.maximum)
	}
	if got := initialInit.modifiers[2]; got == nil || got.name != weaponCreateWierdlingModifier54C710 {
		t.Fatalf("cached InitData modifier 2 = %+v", got)
	}
	if got := initialInit.modifiers[3]; got == nil || got.name != weaponCreateWierdlingSecond54C710 {
		t.Fatalf("cached InitData modifier 3 = %+v", got)
	}
	if replacementInit.modifiers[2] != nil || replacementInit.modifiers[3] != nil {
		t.Fatalf("replacement InitData was used: %+v", replacementInit.modifiers)
	}
	wantStores := []struct {
		offset uintptr
		value  uint8
	}{
		{weaponCreateAmmoCharge154C710, 6},
		{weaponCreateAmmoField254C710, 0},
		{weaponCreateAmmoCharge054C710, 6},
		{weaponCreateWandMaxCharge54C710, 15},
		{weaponCreateWandCharge54C710, 9},
	}
	if !reflect.DeepEqual(stores, wantStores) {
		t.Fatalf("UseData stores = %v, want %v", stores, wantStores)
	}
}

func TestWeaponCreate54C710DurabilityGatesDoNotReturn(t *testing.T) {
	for _, test := range []struct {
		name       string
		definition *weaponCreateTestDefinition54C710
		health     *weaponCreateTestHealth54C710
	}{
		{name: "nil definition", health: new(weaponCreateTestHealth54C710)},
		{name: "nil health", definition: &weaponCreateTestDefinition54C710{durability: 99}},
	} {
		t.Run(test.name, func(t *testing.T) {
			initData := new(weaponCreateTestInit54C710)
			useData := new(weaponCreateTestUse54C710)
			obj := &weaponCreateTestObject54C710{
				typeInd:  11,
				initData: initData,
				health:   test.health,
				class:    weaponCreateAmmoClass54C710,
				subclass: weaponCreateAmmoChargeMask54C710,
				useData:  useData,
			}
			cache := &weaponCreateTypeCache54C710{heart: 11, wierdling: 12, orb: 13}
			weaponCreate54C710(obj, weaponCreateTestHooksFor54C710(obj, test.definition, cache, false))

			if got := initData.modifiers[2]; got == nil || got.name != weaponCreateHeartModifier54C710 {
				t.Fatalf("modifier after durability gate = %+v", got)
			}
			if useData[0] != 7 || useData[1] != 7 || useData[2] != 0 {
				t.Fatalf("ammo after durability gate = %v, want [7 7 0]", useData[:3])
			}
		})
	}
}

func TestWeaponCreate54C710ReloadsHealthAtEveryAccess(t *testing.T) {
	health := []*weaponCreateTestHealth54C710{
		{current: 100, maximum: 101},
		{current: 110, maximum: 111},
		{current: 3, maximum: 121},
		{current: 130, maximum: 131},
		{current: 140, maximum: 5},
		{current: 150, maximum: 151},
	}
	obj := &weaponCreateTestObject54C710{typeInd: 99, initData: new(weaponCreateTestInit54C710)}
	definition := &weaponCreateTestDefinition54C710{durability: 7}
	cache := &weaponCreateTypeCache54C710{heart: 11, wierdling: 12, orb: 13}
	hooks := weaponCreateTestHooksFor54C710(obj, definition, cache, true)
	healthIndex := 0
	hooks.loadHealth = func(*weaponCreateTestObject54C710) *weaponCreateTestHealth54C710 {
		got := health[healthIndex]
		healthIndex++
		return got
	}
	hooks.loadDurability = func(got *weaponCreateTestDefinition54C710) uint16 {
		value := got.durability
		got.durability = 9
		return value
	}

	weaponCreate54C710(obj, hooks)

	if healthIndex != 6 {
		t.Fatalf("HealthData loads = %d, want 6", healthIndex)
	}
	if health[0].current != 7 || health[1].maximum != 9 {
		t.Fatalf("initial live stores = %d/%d, want 7/9", health[0].current, health[1].maximum)
	}
	if health[3].current != 4 || health[5].maximum != 8 {
		t.Fatalf("scaled live stores = %d/%d, want 4/8", health[3].current, health[5].maximum)
	}
}

func TestWeaponCreate54C710RetainsOriginalUnguardedStores(t *testing.T) {
	t.Run("Oblivion InitData", func(t *testing.T) {
		obj := &weaponCreateTestObject54C710{typeInd: 11}
		cache := &weaponCreateTypeCache54C710{heart: 11, wierdling: 12, orb: 13}
		defer func() {
			if recover() == nil {
				t.Fatal("nil cached InitData did not fault")
			}
		}()
		weaponCreate54C710(obj, weaponCreateTestHooksFor54C710(obj, nil, cache, false))
	})

	t.Run("ammo UseData", func(t *testing.T) {
		obj := &weaponCreateTestObject54C710{
			typeInd:  99,
			initData: new(weaponCreateTestInit54C710),
			class:    weaponCreateAmmoClass54C710,
			subclass: weaponCreateAmmoChargeMask54C710,
		}
		cache := &weaponCreateTypeCache54C710{heart: 11, wierdling: 12, orb: 13}
		defer func() {
			if recover() == nil {
				t.Fatal("nil ammo UseData did not fault")
			}
		}()
		weaponCreate54C710(obj, weaponCreateTestHooksFor54C710(obj, nil, cache, false))
	})
}

func TestWeaponCreateScale54C710UsesBinary32Results(t *testing.T) {
	if got := weaponCreateScale54C710(3, 1.5); got != 4.5 {
		t.Fatalf("durability scale = %v, want 4.5", got)
	}
	if got := weaponCreateScaleByte54C710(5, 1.5); got != 7.5 {
		t.Fatalf("byte scale = %v, want 7.5", got)
	}
	if got := weaponCreateDoubleMultiplier54C710(1.5); got != 3 {
		t.Fatalf("doubled multiplier = %v, want 3", got)
	}
}
