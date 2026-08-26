package server

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

type rewardMarkerTestData4F0720 struct {
	mask       uint32
	chanceMode uint32
}

type rewardMarkerTestObject4F0720 struct {
	initData *rewardMarkerTestData4F0720
	typeInd  uint16
}

func rewardMarkerTestHooks4F0720(
	marker *rewardMarkerTestObject4F0720,
	random func(int32, int32) int32,
	dispatch func(rewardMarkerDispatch4F0720, *rewardMarkerTestObject4F0720, uint32) *int,
) rewardMarkerActivateHooks4F0720[
	*rewardMarkerTestObject4F0720,
	*rewardMarkerTestData4F0720,
	*int,
] {
	return rewardMarkerActivateHooks4F0720[
		*rewardMarkerTestObject4F0720,
		*rewardMarkerTestData4F0720,
		*int,
	]{
		loadCachedPlusType: func() uint32 { return 0x1234 },
		loadInitData:       func(got *rewardMarkerTestObject4F0720) *rewardMarkerTestData4F0720 { return got.initData },
		lookupType: func(string) uint32 {
			panic("populated cache reached type lookup")
		},
		storeCachedPlusType: func(uint32) {
			panic("populated cache reached type-cache store")
		},
		loadTypeInd:      func(got *rewardMarkerTestObject4F0720) uint16 { return got.typeInd },
		loadChanceMode:   func(data *rewardMarkerTestData4F0720) uint32 { return data.chanceMode },
		randomInt:        random,
		loadCategoryMask: func(data *rewardMarkerTestData4F0720) uint32 { return data.mask },
		dispatch:         dispatch,
	}
}

func TestRewardMarkerActivate4F0720SealedCategoryRows(t *testing.T) {
	var raw [64]byte
	var total uint32
	for index, row := range rewardMarkerCategories4F0720 {
		raw[index*8] = row.Weight
		binary.LittleEndian.PutUint32(raw[index*8+4:], row.Field4)
		total += uint32(row.Weight)
	}
	if total != 100 {
		t.Fatalf("all-category weight = %d, want 100", total)
	}
	got := fmt.Sprintf("%x", sha256.Sum256(raw[:]))
	const want = "9243e342f59fc43aee5d96ddf29ea3f71d09d33cf58a2eea3051416ed3538cec"
	if got != want {
		t.Fatalf("normalized GAME.EXE 005B98C4 table SHA-256 = %s, want %s", got, want)
	}
}

func TestRewardMarkerDispatchIndex4F0720ExactSelectorTable(t *testing.T) {
	powers := map[uint32]rewardMarkerDispatch4F0720{
		1:   rewardMarkerDispatchSpellBook4F0720,
		2:   rewardMarkerDispatchAbilityBook4F0720,
		4:   rewardMarkerDispatchFieldGuide4F0720,
		8:   rewardMarkerDispatchWeapon4F0720,
		16:  rewardMarkerDispatchArmor4F0720,
		32:  rewardMarkerDispatchGem4F0720,
		64:  rewardMarkerDispatchPotion4F0720,
		128: rewardMarkerDispatchGem2_4F0720,
	}
	for selector := uint32(0); selector <= 256; selector++ {
		want, ok := powers[selector]
		if !ok {
			want = rewardMarkerDispatchDefaultGem4F0720
		}
		if got := rewardMarkerDispatchIndex4F0720(selector); got != want {
			t.Errorf("selector %#x = dispatch %d, want %d", selector, got, want)
		}
	}
	if got := rewardMarkerDispatchIndex4F0720(^uint32(0)); got != rewardMarkerDispatchDefaultGem4F0720 {
		t.Fatalf("maximum selector dispatch = %d, want default gem", got)
	}
}

func TestRewardMarkerActivate4F0720ChanceThresholdsAndDefaultSkip(t *testing.T) {
	thresholds := map[uint32]int32{1: 75, 2: 50, 3: 25, 4: 5}
	for mode, threshold := range thresholds {
		for _, tc := range []struct {
			name       string
			chanceDraw int32
			wantNil    bool
		}{
			{"equal-passes", threshold, false},
			{"greater-rejects", threshold + 1, true},
		} {
			t.Run(fmt.Sprintf("mode-%d/%s", mode, tc.name), func(t *testing.T) {
				marker := &rewardMarkerTestObject4F0720{initData: &rewardMarkerTestData4F0720{mask: 1, chanceMode: mode}}
				var bounds [][2]int32
				token := 1
				hooks := rewardMarkerTestHooks4F0720(marker, func(minimum, maximum int32) int32 {
					bounds = append(bounds, [2]int32{minimum, maximum})
					if minimum == 0 {
						return tc.chanceDraw
					}
					return 1
				}, func(kind rewardMarkerDispatch4F0720, got *rewardMarkerTestObject4F0720, stage uint32) *int {
					if kind != rewardMarkerDispatchSpellBook4F0720 || got != marker || stage != 7 {
						t.Fatalf("dispatch = %d/%p/%d", kind, got, stage)
					}
					return &token
				})
				got := rewardMarkerActivate4F0720(marker, 7, hooks)
				if (got == nil) != tc.wantNil {
					t.Fatalf("result nil = %v, want %v", got == nil, tc.wantNil)
				}
				wantBounds := [][2]int32{{0, 100}}
				if !tc.wantNil {
					wantBounds = append(wantBounds, [2]int32{1, 16})
				}
				if !reflect.DeepEqual(bounds, wantBounds) {
					t.Fatalf("RNG bounds = %v, want %v", bounds, wantBounds)
				}
			})
		}
	}

	for _, mode := range []uint32{0, 5, 0x80000000, ^uint32(0)} {
		t.Run(fmt.Sprintf("default-%#x", mode), func(t *testing.T) {
			marker := &rewardMarkerTestObject4F0720{initData: &rewardMarkerTestData4F0720{mask: 2, chanceMode: mode}}
			var bounds [][2]int32
			token := 1
			got := rewardMarkerActivate4F0720(marker, 9, rewardMarkerTestHooks4F0720(marker,
				func(minimum, maximum int32) int32 {
					bounds = append(bounds, [2]int32{minimum, maximum})
					return 1
				},
				func(kind rewardMarkerDispatch4F0720, _ *rewardMarkerTestObject4F0720, _ uint32) *int {
					if kind != rewardMarkerDispatchAbilityBook4F0720 {
						t.Fatalf("dispatch = %d, want ability book", kind)
					}
					return &token
				},
			))
			if got != &token || !reflect.DeepEqual(bounds, [][2]int32{{1, 2}}) {
				t.Fatalf("result/bounds = %p/%v, want token/[[1 2]]", got, bounds)
			}
		})
	}
}

func TestRewardMarkerActivate4F0720WeightedUnsignedBoundaries(t *testing.T) {
	tests := []struct {
		draw int32
		want rewardMarkerDispatch4F0720
	}{
		{1, rewardMarkerDispatchSpellBook4F0720},
		{16, rewardMarkerDispatchSpellBook4F0720},
		{17, rewardMarkerDispatchAbilityBook4F0720},
		{18, rewardMarkerDispatchAbilityBook4F0720},
		{19, rewardMarkerDispatchFieldGuide4F0720},
		{20, rewardMarkerDispatchFieldGuide4F0720},
		{21, rewardMarkerDispatchWeapon4F0720},
		{44, rewardMarkerDispatchWeapon4F0720},
		{45, rewardMarkerDispatchArmor4F0720},
		{60, rewardMarkerDispatchArmor4F0720},
		{61, rewardMarkerDispatchGem4F0720},
		{83, rewardMarkerDispatchGem4F0720},
		{84, rewardMarkerDispatchPotion4F0720},
		{99, rewardMarkerDispatchPotion4F0720},
		{100, rewardMarkerDispatchGem2_4F0720},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("draw-%d", tc.draw), func(t *testing.T) {
			marker := &rewardMarkerTestObject4F0720{initData: &rewardMarkerTestData4F0720{mask: 0xff}}
			token := 1
			got := rewardMarkerActivate4F0720(marker, 3, rewardMarkerTestHooks4F0720(marker,
				func(minimum, maximum int32) int32 {
					if minimum != 1 || maximum != 100 {
						t.Fatalf("RNG bounds = %d..%d, want 1..100", minimum, maximum)
					}
					return tc.draw
				},
				func(kind rewardMarkerDispatch4F0720, _ *rewardMarkerTestObject4F0720, stage uint32) *int {
					if kind != tc.want || stage != 3 {
						t.Fatalf("dispatch/stage = %d/%d, want %d/3", kind, stage, tc.want)
					}
					return &token
				},
			))
			if got != &token {
				t.Fatal("selected reward result was not preserved")
			}
		})
	}

	marker := &rewardMarkerTestObject4F0720{initData: &rewardMarkerTestData4F0720{mask: 1}}
	token := 1
	got := rewardMarkerActivate4F0720(marker, 2, rewardMarkerTestHooks4F0720(marker,
		func(int32, int32) int32 { return -1 },
		func(kind rewardMarkerDispatch4F0720, _ *rewardMarkerTestObject4F0720, _ uint32) *int {
			if kind != rewardMarkerDispatchAbilityBook4F0720 {
				t.Fatalf("negative draw dispatch = %d, want stage fallback ability", kind)
			}
			return &token
		},
	))
	if got != &token {
		t.Fatal("unsigned fallback result was not preserved")
	}
}

func TestRewardMarkerActivate4F0720ZeroTotalReturnsBeforeRNG(t *testing.T) {
	marker := &rewardMarkerTestObject4F0720{initData: &rewardMarkerTestData4F0720{mask: 0}}
	got := rewardMarkerActivate4F0720(marker, 1, rewardMarkerTestHooks4F0720(marker,
		func(int32, int32) int32 {
			t.Fatal("zero total reached weighted RNG")
			return 0
		},
		func(rewardMarkerDispatch4F0720, *rewardMarkerTestObject4F0720, uint32) *int {
			t.Fatal("zero total reached dispatch")
			return nil
		},
	))
	if got != nil {
		t.Fatalf("zero-total result = %p, want nil", got)
	}
}

func TestRewardMarkerActivate4F0720ReloadsMaskFromCachedInitData(t *testing.T) {
	entry := &rewardMarkerTestData4F0720{mask: 1}
	replacement := &rewardMarkerTestData4F0720{mask: 128}
	marker := &rewardMarkerTestObject4F0720{initData: entry}
	token := 1
	maskLoads := 0
	hooks := rewardMarkerTestHooks4F0720(marker,
		func(minimum, maximum int32) int32 {
			if minimum != 1 || maximum != 16 {
				t.Fatalf("RNG bounds = %d..%d", minimum, maximum)
			}
			marker.initData = replacement
			return 1
		},
		func(kind rewardMarkerDispatch4F0720, _ *rewardMarkerTestObject4F0720, _ uint32) *int {
			if kind != rewardMarkerDispatchSpellBook4F0720 {
				t.Fatalf("dispatch = %d, want cached-data spell book", kind)
			}
			return &token
		},
	)
	hooks.loadCategoryMask = func(data *rewardMarkerTestData4F0720) uint32 {
		maskLoads++
		if data != entry {
			t.Fatal("live marker InitData replaced the entry-cached pointer")
		}
		return data.mask
	}
	if got := rewardMarkerActivate4F0720(marker, 8, hooks); got != &token || maskLoads != 2 || marker.initData != replacement {
		t.Fatalf("result/mask loads/live data = %p/%d/%p", got, maskLoads, marker.initData)
	}
}

func TestRewardMarkerActivate4F0720MaskMutationUsesAdjustedStageFallback(t *testing.T) {
	data := &rewardMarkerTestData4F0720{mask: 1}
	marker := &rewardMarkerTestObject4F0720{initData: data, typeInd: 0x1234}
	token := 1
	got := rewardMarkerActivate4F0720(marker, 62, rewardMarkerTestHooks4F0720(marker,
		func(minimum, maximum int32) int32 {
			if minimum != 1 || maximum != 16 {
				t.Fatalf("RNG bounds = %d..%d", minimum, maximum)
			}
			data.mask = 0
			return 1
		},
		func(kind rewardMarkerDispatch4F0720, _ *rewardMarkerTestObject4F0720, stage uint32) *int {
			if kind != rewardMarkerDispatchPotion4F0720 || stage != 64 {
				t.Fatalf("dispatch/stage = %d/%d, want potion/64", kind, stage)
			}
			return &token
		},
	))
	if got != &token || data.mask != 0 {
		t.Fatalf("result/mask = %p/%#x", got, data.mask)
	}
}

func TestRewardMarkerActivate4F0720LookupOrderAndWrappingStage(t *testing.T) {
	data := &rewardMarkerTestData4F0720{mask: 128}
	marker := &rewardMarkerTestObject4F0720{initData: data, typeInd: 77}
	events := make([]string, 0, 10)
	cached := uint32(0)
	token := 1
	hooks := rewardMarkerActivateHooks4F0720[
		*rewardMarkerTestObject4F0720,
		*rewardMarkerTestData4F0720,
		*int,
	]{
		loadCachedPlusType: func() uint32 {
			events = append(events, "load-cache")
			return cached
		},
		loadInitData: func(got *rewardMarkerTestObject4F0720) *rewardMarkerTestData4F0720 {
			events = append(events, "load-init")
			return got.initData
		},
		lookupType: func(name string) uint32 {
			events = append(events, "lookup")
			if name != "RewardMarkerPlus" {
				t.Fatalf("lookup name = %q", name)
			}
			return 77
		},
		storeCachedPlusType: func(value uint32) {
			events = append(events, "store-cache")
			cached = value
		},
		loadTypeInd: func(got *rewardMarkerTestObject4F0720) uint16 {
			events = append(events, "load-type")
			return got.typeInd
		},
		loadChanceMode: func(got *rewardMarkerTestData4F0720) uint32 {
			events = append(events, "load-chance")
			return got.chanceMode
		},
		loadCategoryMask: func(got *rewardMarkerTestData4F0720) uint32 {
			events = append(events, "load-mask")
			return got.mask
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "random-weighted")
			if minimum != 1 || maximum != 1 {
				t.Fatalf("RNG bounds = %d..%d", minimum, maximum)
			}
			return 1
		},
		dispatch: func(kind rewardMarkerDispatch4F0720, got *rewardMarkerTestObject4F0720, stage uint32) *int {
			events = append(events, "dispatch")
			if kind != rewardMarkerDispatchGem2_4F0720 || got != marker || stage != 1 {
				t.Fatalf("dispatch = %d/%p/%#x", kind, got, stage)
			}
			return &token
		},
	}
	got := rewardMarkerActivate4F0720(marker, ^uint32(0), hooks)
	wantEvents := []string{
		"load-cache", "load-init", "lookup", "store-cache", "load-type",
		"load-chance", "load-mask", "random-weighted", "load-mask", "dispatch",
	}
	if got != &token || cached != 77 || !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("result/cache/events = %p/%d/%v, want token/77/%v", got, cached, events, wantEvents)
	}
}

func TestRewardMarkerActivate4F0720EveryObservableFaultPrefix(t *testing.T) {
	allEvents := []string{
		"load-cache", "load-init", "lookup", "store-cache", "load-type", "load-chance",
		"random-chance", "load-mask-first", "random-weighted", "load-mask-second", "dispatch",
	}
	for faultIndex, fault := range allEvents {
		t.Run(fault, func(t *testing.T) {
			data := &rewardMarkerTestData4F0720{mask: 1, chanceMode: 1}
			marker := &rewardMarkerTestObject4F0720{initData: data, typeInd: 77}
			events := make([]string, 0, len(allEvents))
			maskLoads := 0
			record := func(event string) {
				events = append(events, event)
				if event == fault {
					panic(fault)
				}
			}
			hooks := rewardMarkerActivateHooks4F0720[
				*rewardMarkerTestObject4F0720,
				*rewardMarkerTestData4F0720,
				*int,
			]{
				loadCachedPlusType: func() uint32 { record("load-cache"); return 0 },
				loadInitData: func(got *rewardMarkerTestObject4F0720) *rewardMarkerTestData4F0720 {
					record("load-init")
					return got.initData
				},
				lookupType: func(string) uint32 { record("lookup"); return 77 },
				storeCachedPlusType: func(uint32) {
					record("store-cache")
				},
				loadTypeInd: func(got *rewardMarkerTestObject4F0720) uint16 {
					record("load-type")
					return got.typeInd
				},
				loadChanceMode: func(got *rewardMarkerTestData4F0720) uint32 {
					record("load-chance")
					return got.chanceMode
				},
				randomInt: func(minimum, maximum int32) int32 {
					if minimum == 0 && maximum == 100 {
						record("random-chance")
						return 0
					}
					record("random-weighted")
					return 1
				},
				loadCategoryMask: func(got *rewardMarkerTestData4F0720) uint32 {
					maskLoads++
					if maskLoads == 1 {
						record("load-mask-first")
					} else {
						record("load-mask-second")
					}
					return got.mask
				},
				dispatch: func(rewardMarkerDispatch4F0720, *rewardMarkerTestObject4F0720, uint32) *int {
					record("dispatch")
					return new(int)
				},
			}
			defer func() {
				if got := recover(); got != fault {
					t.Fatalf("panic = %v, want %q", got, fault)
				}
				want := allEvents[:faultIndex+1]
				if !reflect.DeepEqual(events, want) {
					t.Fatalf("fault prefix = %v, want %v", events, want)
				}
			}()
			rewardMarkerActivate4F0720(marker, 7, hooks)
		})
	}
}
