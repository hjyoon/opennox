package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type questThemeUnitTest51A1F0 struct {
	next, update int
	class        uint32
	subclass     uint8
	typeID       uint16
}

type questThemeUpdateTest51A1F0 struct {
	creatures [3]int
	selectors [3]uint8
	maximum   uint8
}

type questThemeTest51A1F0 struct {
	units      map[int]*questThemeUnitTest51A1F0
	updates    map[int]*questThemeUpdateTest51A1F0
	first      int
	stage      uint32
	hec, nec   uint32
	values     map[string]float64
	randoms    []int32
	trace      []string
	deleted    []int
	minions    []int32
	hecSpawn   []int
	necroSpawn []int
	onDelete   func(int)
	onHecSpawn func(int)
	faultAt    int
}

func newQuestThemeTest51A1F0() *questThemeTest51A1F0 {
	return &questThemeTest51A1F0{
		units: make(map[int]*questThemeUnitTest51A1F0), updates: make(map[int]*questThemeUpdateTest51A1F0),
		stage: 1, hec: 101, nec: 102,
		values: map[string]float64{questHardcoreStageKey51A1F0: 10, questMinionsAlwaysKey51A1F0: 10},
	}
}

func (f *questThemeTest51A1F0) record(event string) {
	f.trace = append(f.trace, event)
	if len(f.trace) == f.faultAt {
		panic("injected Quest callback fault")
	}
}

func (f *questThemeTest51A1F0) hooks(t *testing.T) questThemeHooks51A1F0[int, int] {
	t.Helper()
	return questThemeHooks51A1F0[int, int]{
		questStage:       func() uint32 { f.record("stage"); return f.stage },
		balanceFloat:     func(key string) float64 { f.record("balance:" + key); return f.values[key] },
		floatToInt:       func(v float32) int32 { f.record("round"); return questInventoryRoundFloat32ToInt32_4F2C30(v) },
		hecubahType:      func() uint32 { f.record("hec-type"); return f.hec },
		necroType:        func() uint32 { f.record("nec-type"); return f.nec },
		storeHecubahType: func(id uint32) { f.record("store-hec"); f.hec = id },
		storeNecroType:   func(id uint32) { f.record("store-nec"); f.nec = id },
		lookupType: func(name string) uint32 {
			f.record("lookup:" + name)
			if name == questHecubahMarker51A1F0 {
				return 101
			}
			return 102
		},
		first:            func() int { f.record("first"); return f.first },
		next:             func(unit int) int { f.record(fmt.Sprintf("next:%d", unit)); return f.units[unit].next },
		loadClass:        func(unit int) uint32 { f.record(fmt.Sprintf("class:%d", unit)); return f.units[unit].class },
		loadSubclassByte: func(unit int) uint8 { f.record(fmt.Sprintf("subclass:%d", unit)); return f.units[unit].subclass },
		loadType:         func(unit int) uint16 { f.record(fmt.Sprintf("type:%d", unit)); return f.units[unit].typeID },
		storeType:        func(unit int, id uint16) { f.record(fmt.Sprintf("store-type:%d", unit)); f.units[unit].typeID = id },
		loadUpdate:       func(unit int) int { f.record(fmt.Sprintf("update:%d", unit)); return f.units[unit].update },
		loadCreature: func(update int, group int32) int {
			f.record(fmt.Sprintf("creature:%d", update))
			return f.updates[update].creatures[group]
		},
		loadSelector: func(update int, group int32) uint8 {
			f.record(fmt.Sprintf("selector:%d", update))
			return f.updates[update].selectors[group]
		},
		truncQwordLow: func(v float64) int32 { f.record("truncate"); return x87TruncSignedQwordLow566DCC(v) },
		loadMaximum:   func(update int) uint8 { f.record(fmt.Sprintf("maximum:%d", update)); return f.updates[update].maximum },
		storeMaximum: func(update int, v uint8) {
			f.record(fmt.Sprintf("store-maximum:%d", update))
			f.updates[update].maximum = v
		},
		generatorType: func(unit int) int32 { f.record(fmt.Sprintf("generator:%d", unit)); return -1 },
		delete: func(unit int) {
			f.record(fmt.Sprintf("delete:%d", unit))
			f.deleted = append(f.deleted, unit)
			if f.onDelete != nil {
				f.onDelete(unit)
			}
		},
		random: func(lo, hi int32) int32 {
			f.record(fmt.Sprintf("random:%d:%d", lo, hi))
			if len(f.randoms) == 0 {
				t.Fatalf("unexpected random(%d, %d)", lo, hi)
			}
			value := f.randoms[0]
			f.randoms = f.randoms[1:]
			if value < lo || value > hi {
				t.Fatalf("random fixture %d outside %d..%d", value, lo, hi)
			}
			return value
		},
		setMinions: func(value int32) { f.record(fmt.Sprintf("minions:%d", value)); f.minions = append(f.minions, value) },
		spawnHecubah: func(unit int) {
			f.record(fmt.Sprintf("spawn-hec:%d", unit))
			f.hecSpawn = append(f.hecSpawn, unit)
			if f.onHecSpawn != nil {
				f.onHecSpawn(unit)
			}
		},
		spawnNecro: func(unit int) { f.record(fmt.Sprintf("spawn-nec:%d", unit)); f.necroSpawn = append(f.necroSpawn, unit) },
	}
}

func TestQuestTheme51A1F0EmptyInitialCacheOrder(t *testing.T) {
	f := newQuestThemeTest51A1F0()
	f.hec, f.nec = 0, 999
	questTheme51A1F0(-1, f.hooks(t)) // An empty list never accesses the group.
	want := []string{"stage", "balance:QuestHardcoreStage", "round", "hec-type", "lookup:HecubahMarker", "store-hec", "lookup:NecromancerMarker", "store-nec", "first", "minions:0", "first"}
	if !reflect.DeepEqual(f.trace, want) || f.hec != 101 || f.nec != 102 {
		t.Fatalf("trace=%v, cache=%d/%d", f.trace, f.hec, f.nec)
	}
}

func TestQuestTheme51A1F0GeneratorSelectorsAndUnsignedHardcore(t *testing.T) {
	for _, selector := range []uint8{0, 1, 2, 3, 4, 255} {
		for _, hardcore := range []float64{10, -1, math.NaN()} {
			t.Run(fmt.Sprintf("selector=%d/hardcore=%v", selector, hardcore), func(t *testing.T) {
				f := newQuestThemeTest51A1F0()
				f.first = 1
				f.units[1] = &questThemeUnitTest51A1F0{class: 0x20000, typeID: 9, update: 11}
				f.updates[11] = &questThemeUpdateTest51A1F0{creatures: [3]int{0, 0, 7}, selectors: [3]uint8{99, 99, selector}, maximum: 130}
				f.values[questHardcoreStageKey51A1F0] = hardcore
				for _, key := range monsterGeneratorMaxActiveKeys4F0590 {
					f.values[key] = 130.9
				}
				h := f.hooks(t)
				stageCalls := 0
				h.questStage = func() uint32 {
					f.record("stage")
					stageCalls++
					if stageCalls == 1 {
						return 1
					}
					return 10
				}
				questTheme51A1F0(2, h)
				wantMaximum := uint8(130)
				if selector != 3 && hardcore == 10 {
					wantMaximum = 4
				}
				if f.updates[11].maximum != wantMaximum || f.units[1].typeID != math.MaxUint16 || stageCalls != 2 {
					t.Fatalf("maximum/type/stage calls = %d/%d/%d", f.updates[11].maximum, f.units[1].typeID, stageCalls)
				}
				selectorReads, maximumBalanceReads := 0, 0
				for _, event := range f.trace {
					if event == "selector:11" {
						selectorReads++
					}
					for _, key := range monsterGeneratorMaxActiveKeys4F0590 {
						if event == "balance:"+key {
							maximumBalanceReads++
						}
					}
				}
				wantSelectorReads := 1
				if hardcore == 10 {
					wantSelectorReads = 2
				}
				wantBalanceReads := 0
				if selector <= 3 {
					wantBalanceReads = 1
				}
				if selectorReads != wantSelectorReads || maximumBalanceReads != wantBalanceReads || !reflect.DeepEqual(f.minions, []int32{0}) {
					t.Fatalf("selector/balance/minions = %d/%d/%v", selectorReads, maximumBalanceReads, f.minions)
				}
			})
		}
	}
}

func TestQuestTheme51A1F0CachesUpdateAndSuccessorReloadsCreature(t *testing.T) {
	f := newQuestThemeTest51A1F0()
	f.first = 1
	f.units[1] = &questThemeUnitTest51A1F0{next: 2, update: 11, class: 0x20000, typeID: 9}
	f.units[2] = &questThemeUnitTest51A1F0{update: 12, class: 0x20000, typeID: 9}
	f.updates[11] = &questThemeUpdateTest51A1F0{creatures: [3]int{7}, maximum: 23}
	f.updates[12] = &questThemeUpdateTest51A1F0{maximum: 43}
	f.updates[13] = &questThemeUpdateTest51A1F0{maximum: 63}
	f.values[monsterGeneratorMaxActiveHighKey4F0590] = 130.9
	h := f.hooks(t)
	h.balanceFloat = func(key string) float64 {
		f.record("balance:" + key)
		if key == monsterGeneratorMaxActiveHighKey4F0590 {
			f.units[1].next, f.units[1].update = 0, 13
			f.updates[11].creatures[0], f.updates[11].selectors[0] = 8, 3
			f.stage = 10
		}
		return f.values[key]
	}
	questTheme51A1F0(0, h)
	want := []string{
		"stage", "balance:QuestHardcoreStage", "round", "hec-type", "first", "next:1", "class:1", "hec-type", "type:1",
		"update:1", "creature:11", "selector:11", "balance:GeneratorMaxActiveCreaturesHigh", "truncate", "store-maximum:11",
		"stage", "selector:11", "creature:11", "generator:8", "store-type:1", "next:2", "class:2", "hec-type", "type:2",
		"update:2", "creature:12", "delete:2", "minions:0", "first", "next:1", "type:1", "hec-type", "nec-type", "type:1",
	}
	if !reflect.DeepEqual(f.trace, want) || !reflect.DeepEqual(f.deleted, []int{2}) || f.updates[11].maximum != 130 || f.updates[13].maximum != 63 {
		t.Fatalf("trace=%v deleted=%v maxima=%d/%d", f.trace, f.deleted, f.updates[11].maximum, f.updates[13].maximum)
	}
}

func TestQuestTheme51A1F0ExitRetentionCachesSuccessor(t *testing.T) {
	f := newQuestThemeTest51A1F0()
	f.first = 1
	for i := 1; i <= 3; i++ {
		f.units[i] = &questThemeUnitTest51A1F0{next: i + 1, class: 0x100020, subclass: 1, typeID: 9}
	}
	f.units[3].next = 0
	f.randoms = []int32{1}
	f.onDelete = func(unit int) { f.units[unit].next = 0 }
	questTheme51A1F0(0, f.hooks(t))
	if !reflect.DeepEqual(f.deleted, []int{1, 3}) || len(f.randoms) != 0 {
		t.Fatalf("deleted exits=%v, unused randoms=%v", f.deleted, f.randoms)
	}
}

func TestQuestTheme51A1F0MinionSignedStageAndRandomGates(t *testing.T) {
	for _, tc := range []struct {
		stage   uint32
		random  int32
		enabled bool
	}{
		{4, 0, false}, {5, 0, true}, {6, 0, false}, {7, 49, false}, {7, 50, true}, {8, 0, false},
		{10, 0, true}, {math.MaxUint32, 0, false}, {0x80000000, 0, false},
	} {
		t.Run(fmt.Sprintf("%d/%d", tc.stage, tc.random), func(t *testing.T) {
			f := newQuestThemeTest51A1F0()
			f.stage = tc.stage
			if tc.random != 0 {
				f.randoms = []int32{tc.random}
			}
			questTheme51A1F0(0, f.hooks(t))
			want := []int32{0}
			if tc.enabled {
				want = append(want, 1)
			}
			if !reflect.DeepEqual(f.minions, want) || len(f.randoms) != 0 {
				t.Fatalf("minions/randoms=%v/%v", f.minions, f.randoms)
			}
		})
	}
}

func TestQuestTheme51A1F0ExitMarkerDoesNotEnableSpawnPass(t *testing.T) {
	f := newQuestThemeTest51A1F0()
	f.stage, f.first = 5, 1
	f.units[1] = &questThemeUnitTest51A1F0{class: 0x20, subclass: 1, typeID: 101}
	questTheme51A1F0(0, f.hooks(t))
	if !reflect.DeepEqual(f.minions, []int32{0, 1}) || len(f.hecSpawn) != 0 || !reflect.DeepEqual(f.deleted, []int{1}) {
		t.Fatalf("minions/spawns/deletions=%v/%v/%v", f.minions, f.hecSpawn, f.deleted)
	}
}

func TestQuestTheme51A1F0SpawnPassReloadsTypeAndSuccessor(t *testing.T) {
	f := newQuestThemeTest51A1F0()
	f.stage, f.first = 5, 1
	f.units[1] = &questThemeUnitTest51A1F0{typeID: 101, next: 2}
	f.units[2] = &questThemeUnitTest51A1F0{typeID: 102}
	f.units[3] = &questThemeUnitTest51A1F0{typeID: 102}
	f.randoms = []int32{1, 50, 49}
	f.onHecSpawn = func(unit int) { f.units[unit].typeID, f.units[unit].next = 102, 3 }
	f.onDelete = func(unit int) { f.units[unit].next = 0 }
	questTheme51A1F0(0, f.hooks(t))
	if !reflect.DeepEqual(f.hecSpawn, []int{1}) || !reflect.DeepEqual(f.necroSpawn, []int{1}) ||
		!reflect.DeepEqual(f.deleted, []int{1, 3}) || len(f.randoms) != 0 {
		t.Fatalf("Hecubah/Necro/deleted/randoms=%v/%v/%v/%v", f.hecSpawn, f.necroSpawn, f.deleted, f.randoms)
	}
}

func TestQuestTheme51A1F0CleanupReloadsTypeBetweenDeletes(t *testing.T) {
	f := newQuestThemeTest51A1F0()
	f.first = 1
	f.units[1] = &questThemeUnitTest51A1F0{typeID: 101, next: 2}
	f.units[2] = &questThemeUnitTest51A1F0{typeID: 102}
	f.onDelete = func(unit int) { f.units[unit].typeID, f.units[unit].next = 102, 0 }
	questTheme51A1F0(0, f.hooks(t))
	if !reflect.DeepEqual(f.deleted, []int{1, 1, 2}) {
		t.Fatalf("live cleanup deletes = %v", f.deleted)
	}
}

func TestQuestTheme51A1F0EveryObservableFaultPrefix(t *testing.T) {
	setup := func() *questThemeTest51A1F0 {
		f := newQuestThemeTest51A1F0()
		f.hec, f.nec, f.stage, f.first = 0, 0, 5, 1
		f.units[1] = &questThemeUnitTest51A1F0{typeID: 101, next: 2}
		f.units[2] = &questThemeUnitTest51A1F0{typeID: 102}
		f.randoms = []int32{1, 50}
		return f
	}
	baseline := setup()
	questTheme51A1F0(0, baseline.hooks(t))
	for failAt := 1; failAt <= len(baseline.trace); failAt++ {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			f := setup()
			f.faultAt = failAt
			defer func() {
				if recover() == nil || !reflect.DeepEqual(f.trace, baseline.trace[:failAt]) {
					t.Fatalf("fault prefix %d = %v, want %v", failAt, f.trace, baseline.trace[:failAt])
				}
			}()
			questTheme51A1F0(0, f.hooks(t))
		})
	}
}
