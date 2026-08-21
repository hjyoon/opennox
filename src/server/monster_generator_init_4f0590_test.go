package server

import (
	"math"
	"reflect"
	"testing"
)

type monsterGeneratorInitTestUpdate4F0590 struct {
	questSpawnRate [3]uint8
	maxActive      uint8
	guard          uint32
}

type monsterGeneratorInitTestObject4F0590 struct {
	update     *monsterGeneratorInitTestUpdate4F0590
	subClass   uint32
	direction1 uint16
	direction2 uint16
	guard      uint32
}

func TestMonsterGeneratorInit4F0590ExactBalanceKeysAndLowByteStore(t *testing.T) {
	keys := []string{
		"GeneratorMaxActiveCreaturesHigh",
		"GeneratorMaxActiveCreaturesNormal",
		"GeneratorMaxActiveCreaturesLow",
		"GeneratorMaxActiveCreaturesSingular",
	}
	sizes := []int{32, 34, 31, 36}
	values := []float64{255.999, 256.999, -1.999, math.NaN()}
	wants := []uint8{255, 0, 255, 0}
	for selector := range keys {
		t.Run(keys[selector], func(t *testing.T) {
			if len(monsterGeneratorMaxActiveKeys4F0590[selector])+1 != sizes[selector] {
				t.Fatalf("key size including NUL = %d, want %d", len(monsterGeneratorMaxActiveKeys4F0590[selector])+1, sizes[selector])
			}
			update := &monsterGeneratorInitTestUpdate4F0590{
				questSpawnRate: [3]uint8{0xee, uint8(selector), 0xdd},
				maxActive:      0xaa,
				guard:          0xa5a5a5a5,
			}
			unit := &monsterGeneratorInitTestObject4F0590{
				update: update, direction1: 0x1234, direction2: 0xbbbb,
				guard: 0x5a5a5a5a,
			}
			events := make([]string, 0, 9)
			got := monsterGeneratorInit4F0590(unit, monsterGeneratorInitHooks4F0590[
				*monsterGeneratorInitTestObject4F0590,
				*monsterGeneratorInitTestUpdate4F0590,
			]{
				loadUpdateData: func(got *monsterGeneratorInitTestObject4F0590) *monsterGeneratorInitTestUpdate4F0590 {
					events = append(events, "load-update")
					return got.update
				},
				currentQuestGroup: func() uint32 {
					events = append(events, "current-group")
					return 1
				},
				loadQuestSpawnRate: func(got *monsterGeneratorInitTestUpdate4F0590, group uint32) uint8 {
					events = append(events, "load-selector")
					return got.questSpawnRate[group]
				},
				loadBalanceFloat: func(key string) float64 {
					events = append(events, "load-balance")
					if key != keys[selector] {
						t.Fatalf("balance key = %q, want %q", key, keys[selector])
					}
					return values[selector]
				},
				truncQwordLow: func(value float64) int32 {
					events = append(events, "trunc")
					return x87TruncSignedQwordLow566DCC(value)
				},
				storeMaxActive: func(got *monsterGeneratorInitTestUpdate4F0590, value uint8) {
					events = append(events, "store-max")
					got.maxActive = value
				},
				loadObjSubClass: func(got *monsterGeneratorInitTestObject4F0590) uint32 {
					events = append(events, "load-subclass")
					return got.subClass
				},
				directionIndexAngle: func(uint32) uint32 {
					t.Fatal("zero subclass reached direction helper")
					return 0
				},
				loadDirection1: func(got *monsterGeneratorInitTestObject4F0590) uint16 {
					events = append(events, "load-direction-1")
					return got.direction1
				},
				storeDirection2: func(got *monsterGeneratorInitTestObject4F0590, value uint16) {
					events = append(events, "store-direction-2")
					got.direction2 = value
				},
			})

			if got != 0 || update.maxActive != wants[selector] {
				t.Fatalf("return/maxActive = %d/%#x, want 0/%#x", got, update.maxActive, wants[selector])
			}
			wantEvents := []string{
				"load-update", "current-group", "load-selector", "load-balance",
				"trunc", "store-max", "load-subclass", "load-direction-1", "store-direction-2",
			}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
			if unit.direction1 != 0x1234 || unit.direction2 != 0x1234 || unit.guard != 0x5a5a5a5a || update.guard != 0xa5a5a5a5 {
				t.Fatalf("adjacent state changed: unit=%+v update=%+v", *unit, *update)
			}
		})
	}
}

func TestMonsterGeneratorInit4F0590InvalidSelectorSkipsBalanceAndReturnsFullSubclass(t *testing.T) {
	update := &monsterGeneratorInitTestUpdate4F0590{questSpawnRate: [3]uint8{4, 0, 0}, maxActive: 0xaa}
	unit := &monsterGeneratorInitTestObject4F0590{
		update: update, subClass: 0x89abcd00, direction1: 0x4321, direction2: 0xbbbb,
	}
	got := monsterGeneratorInit4F0590(unit, monsterGeneratorInitHooks4F0590[
		*monsterGeneratorInitTestObject4F0590,
		*monsterGeneratorInitTestUpdate4F0590,
	]{
		loadUpdateData: func(got *monsterGeneratorInitTestObject4F0590) *monsterGeneratorInitTestUpdate4F0590 {
			return got.update
		},
		currentQuestGroup:  func() uint32 { return 0 },
		loadQuestSpawnRate: func(got *monsterGeneratorInitTestUpdate4F0590, group uint32) uint8 { return got.questSpawnRate[group] },
		loadBalanceFloat: func(string) float64 {
			t.Fatal("selector greater than three reached balance lookup")
			return 0
		},
		truncQwordLow: func(float64) int32 {
			t.Fatal("selector greater than three reached truncation")
			return 0
		},
		storeMaxActive: func(*monsterGeneratorInitTestUpdate4F0590, uint8) {
			t.Fatal("selector greater than three reached MaxActive store")
		},
		loadObjSubClass: func(got *monsterGeneratorInitTestObject4F0590) uint32 { return got.subClass },
		directionIndexAngle: func(uint32) uint32 {
			t.Fatal("unmatched subclass reached direction helper")
			return 0
		},
		loadDirection1:  func(got *monsterGeneratorInitTestObject4F0590) uint16 { return got.direction1 },
		storeDirection2: func(got *monsterGeneratorInitTestObject4F0590, value uint16) { got.direction2 = value },
	})
	if uint32(got) != unit.subClass || update.maxActive != 0xaa {
		t.Fatalf("return/maxActive = %#x/%#x, want %#x/0xaa", uint32(got), update.maxActive, unit.subClass)
	}
	if unit.direction1 != 0x4321 || unit.direction2 != 0x4321 {
		t.Fatalf("directions = %#x/%#x, want 0x4321/0x4321", unit.direction1, unit.direction2)
	}
}

func TestMonsterGeneratorInit4F0590DirectionBitPriorityAndFullReturn(t *testing.T) {
	tests := []struct {
		name     string
		subClass uint32
		index    uint32
	}{
		{"bit-1", 0x80000001, 0},
		{"bit-1-before-2", 3, 0},
		{"bit-2", 2, 2},
		{"bit-2-before-4", 6, 2},
		{"bit-4", 4, 8},
		{"bit-4-before-8", 12, 8},
		{"bit-8", 8, 6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			update := &monsterGeneratorInitTestUpdate4F0590{questSpawnRate: [3]uint8{0xff, 0xff, 0xff}}
			unit := &monsterGeneratorInitTestObject4F0590{
				update: update, subClass: tc.subClass, direction1: 0xaaaa, direction2: 0xbbbb,
			}
			wantResult := uint32(0x89abc000) | tc.index
			gotIndex := ^uint32(0)
			got := monsterGeneratorInit4F0590(unit, monsterGeneratorInitHooks4F0590[
				*monsterGeneratorInitTestObject4F0590,
				*monsterGeneratorInitTestUpdate4F0590,
			]{
				loadUpdateData: func(got *monsterGeneratorInitTestObject4F0590) *monsterGeneratorInitTestUpdate4F0590 {
					return got.update
				},
				currentQuestGroup:  func() uint32 { return 2 },
				loadQuestSpawnRate: func(got *monsterGeneratorInitTestUpdate4F0590, group uint32) uint8 { return got.questSpawnRate[group] },
				loadObjSubClass:    func(got *monsterGeneratorInitTestObject4F0590) uint32 { return got.subClass },
				directionIndexAngle: func(index uint32) uint32 {
					gotIndex = index
					return wantResult
				},
				storeDirection1: func(got *monsterGeneratorInitTestObject4F0590, value uint16) { got.direction1 = value },
				loadDirection1:  func(got *monsterGeneratorInitTestObject4F0590) uint16 { return got.direction1 },
				storeDirection2: func(got *monsterGeneratorInitTestObject4F0590, value uint16) { got.direction2 = value },
			})
			if uint32(got) != wantResult || gotIndex != tc.index {
				t.Fatalf("return/index = %#x/%d, want %#x/%d", uint32(got), gotIndex, wantResult, tc.index)
			}
			if unit.direction1 != uint16(wantResult) || unit.direction2 != uint16(wantResult) {
				t.Fatalf("directions = %#x/%#x, want %#x", unit.direction1, unit.direction2, uint16(wantResult))
			}
		})
	}
}

func TestMonsterGeneratorInit4F0590CachesUpdateAndReloadsDirection1(t *testing.T) {
	entry := &monsterGeneratorInitTestUpdate4F0590{questSpawnRate: [3]uint8{0, 1, 2}, guard: 0xa5a5a5a5}
	replacement := &monsterGeneratorInitTestUpdate4F0590{questSpawnRate: [3]uint8{3, 3, 3}, maxActive: 0x44, guard: 0x5a5a5a5a}
	unit := &monsterGeneratorInitTestObject4F0590{
		update: entry, subClass: 1, direction1: 0xaaaa, direction2: 0xbbbb,
	}
	events := make([]string, 0, 12)
	got := monsterGeneratorInit4F0590(unit, monsterGeneratorInitHooks4F0590[
		*monsterGeneratorInitTestObject4F0590,
		*monsterGeneratorInitTestUpdate4F0590,
	]{
		loadUpdateData: func(got *monsterGeneratorInitTestObject4F0590) *monsterGeneratorInitTestUpdate4F0590 {
			events = append(events, "load-update")
			return got.update
		},
		currentQuestGroup: func() uint32 {
			events = append(events, "current-group")
			unit.update = replacement
			return 1
		},
		loadQuestSpawnRate: func(got *monsterGeneratorInitTestUpdate4F0590, group uint32) uint8 {
			events = append(events, "load-selector")
			if got != entry {
				t.Fatal("quest callback mutation replaced cached UpdateData")
			}
			return got.questSpawnRate[group]
		},
		loadBalanceFloat: func(key string) float64 {
			events = append(events, "load-balance")
			if key != monsterGeneratorMaxActiveNormalKey4F0590 {
				t.Fatalf("key = %q", key)
			}
			unit.subClass = 4
			return float64(float32(513.75))
		},
		truncQwordLow: func(value float64) int32 {
			events = append(events, "trunc")
			return x87TruncSignedQwordLow566DCC(value)
		},
		storeMaxActive: func(got *monsterGeneratorInitTestUpdate4F0590, value uint8) {
			events = append(events, "store-max")
			got.maxActive = value
		},
		loadObjSubClass: func(got *monsterGeneratorInitTestObject4F0590) uint32 {
			events = append(events, "load-subclass")
			return got.subClass
		},
		directionIndexAngle: func(index uint32) uint32 {
			events = append(events, "direction-angle")
			if index != 8 {
				t.Fatalf("direction index = %d, want 8", index)
			}
			unit.direction1 = 0x1111
			return 0x12345678
		},
		storeDirection1: func(got *monsterGeneratorInitTestObject4F0590, value uint16) {
			events = append(events, "store-direction-1")
			got.direction1 = value
		},
		loadDirection1: func(got *monsterGeneratorInitTestObject4F0590) uint16 {
			events = append(events, "load-direction-1")
			return got.direction1
		},
		storeDirection2: func(got *monsterGeneratorInitTestObject4F0590, value uint16) {
			events = append(events, "store-direction-2")
			got.direction2 = value
		},
	})
	wantEvents := []string{
		"load-update", "current-group", "load-selector", "load-balance", "trunc", "store-max",
		"load-subclass", "direction-angle", "store-direction-1", "load-direction-1", "store-direction-2",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if uint32(got) != 0x12345678 || entry.maxActive != 1 || replacement.maxActive != 0x44 {
		t.Fatalf("return/entry/replacement = %#x/%#x/%#x", uint32(got), entry.maxActive, replacement.maxActive)
	}
	if unit.update != replacement || unit.direction1 != 0x5678 || unit.direction2 != 0x5678 {
		t.Fatalf("unit cache/directions = %p %#x/%#x", unit.update, unit.direction1, unit.direction2)
	}
}

func TestMonsterGeneratorInit4F0590EveryObservableFaultPrefix(t *testing.T) {
	allEvents := []string{
		"load-update", "current-group", "load-selector", "load-balance", "trunc", "store-max",
		"load-subclass", "direction-angle", "store-direction-1", "load-direction-1", "store-direction-2",
	}
	for faultIndex, fault := range allEvents {
		t.Run(fault, func(t *testing.T) {
			update := &monsterGeneratorInitTestUpdate4F0590{questSpawnRate: [3]uint8{0, 0, 0}, maxActive: 0xaa}
			unit := &monsterGeneratorInitTestObject4F0590{
				update: update, subClass: 1, direction1: 0xaaaa, direction2: 0xbbbb,
			}
			events := make([]string, 0, len(allEvents))
			record := func(event string) {
				events = append(events, event)
				if event == fault {
					panic(fault)
				}
			}
			hooks := monsterGeneratorInitHooks4F0590[
				*monsterGeneratorInitTestObject4F0590,
				*monsterGeneratorInitTestUpdate4F0590,
			]{
				loadUpdateData: func(got *monsterGeneratorInitTestObject4F0590) *monsterGeneratorInitTestUpdate4F0590 {
					record("load-update")
					return got.update
				},
				currentQuestGroup: func() uint32 {
					record("current-group")
					return 0
				},
				loadQuestSpawnRate: func(got *monsterGeneratorInitTestUpdate4F0590, group uint32) uint8 {
					record("load-selector")
					return got.questSpawnRate[group]
				},
				loadBalanceFloat: func(string) float64 {
					record("load-balance")
					return 7
				},
				truncQwordLow: func(float64) int32 {
					record("trunc")
					return 7
				},
				storeMaxActive: func(got *monsterGeneratorInitTestUpdate4F0590, value uint8) {
					record("store-max")
					got.maxActive = value
				},
				loadObjSubClass: func(got *monsterGeneratorInitTestObject4F0590) uint32 {
					record("load-subclass")
					return got.subClass
				},
				directionIndexAngle: func(uint32) uint32 {
					record("direction-angle")
					return 0x12345678
				},
				storeDirection1: func(got *monsterGeneratorInitTestObject4F0590, value uint16) {
					record("store-direction-1")
					got.direction1 = value
				},
				loadDirection1: func(got *monsterGeneratorInitTestObject4F0590) uint16 {
					record("load-direction-1")
					return got.direction1
				},
				storeDirection2: func(got *monsterGeneratorInitTestObject4F0590, value uint16) {
					record("store-direction-2")
					got.direction2 = value
				},
			}
			func() {
				defer func() {
					if got := recover(); got != fault {
						t.Fatalf("panic = %v, want %q", got, fault)
					}
				}()
				monsterGeneratorInit4F0590(unit, hooks)
			}()
			if want := allEvents[:faultIndex+1]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
			wantMax := uint8(0xaa)
			if faultIndex > 5 {
				wantMax = 7
			}
			wantDirection1 := uint16(0xaaaa)
			if faultIndex > 8 {
				wantDirection1 = 0x5678
			}
			if update.maxActive != wantMax || unit.direction1 != wantDirection1 || unit.direction2 != 0xbbbb {
				t.Fatalf("state = max %#x directions %#x/%#x, want %#x %#x/0xbbbb", update.maxActive, unit.direction1, unit.direction2, wantMax, wantDirection1)
			}
		})
	}
}

func TestDirectionIndexToAngle509E90PreservesIndexAndFullDword(t *testing.T) {
	calls := 0
	got := directionIndexToAngle509E90(8, directionIndexToAngleHooks509E90{
		loadTable: func(index uint32) uint32 {
			calls++
			if index != 8 {
				t.Fatalf("index = %d, want 8", index)
			}
			return 0x89abcdef
		},
	})
	if got != 0x89abcdef || calls != 1 {
		t.Fatalf("return/calls = %#x/%d", got, calls)
	}
}

func TestX87TruncSignedQwordLow566DCCSharedSemantics(t *testing.T) {
	tests := []struct {
		value float64
		want  int32
	}{
		{123.999, 123},
		{-123.999, -123},
		{0x100000001, 1},
		{-0x100000001, -1},
		{math.Nextafter(0x1p63, 0), -1024},
		{0x1p63, 0},
		{-0x1p63, 0},
		{math.Inf(1), 0},
		{math.Inf(-1), 0},
		{math.NaN(), 0},
	}
	for _, tc := range tests {
		if got := x87TruncSignedQwordLow566DCC(tc.value); got != tc.want {
			t.Errorf("trunc(%v) = %d (%#x), want %d (%#x)", tc.value, got, uint32(got), tc.want, uint32(tc.want))
		}
		if got := goldInitTruncQwordLow4F04B0(tc.value); got != tc.want {
			t.Errorf("Gold wrapper trunc(%v) = %d, want %d", tc.value, got, tc.want)
		}
	}
}
