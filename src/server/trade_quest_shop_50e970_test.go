package server

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestQuestShopBonusItemTable50E970Sealed(t *testing.T) {
	if got := len(questShopBonusItemTypes50E970); got != 46 {
		t.Fatalf("Quest bonus item count = %d, want 46", got)
	}
	normalized := strings.Join(questShopBonusItemTypes50E970[:], "\x00") + "\x00"
	got := fmt.Sprintf("%x", sha256.Sum256([]byte(normalized)))
	const want = "cb0a3433598109cd654a47b0b2c504f2143eb1239f4128e7b75a2687c5310f35"
	if got != want {
		t.Fatalf("normalized GAME.EXE 005C0540 table SHA-256 = %s, want %s", got, want)
	}
	wantCategories := []uint32{
		8, 8, 8, 8,
		16, 16, 16, 16,
		1, 1, 1, 1, 1, 1,
		4, 4, 4,
	}
	if !reflect.DeepEqual(questShopRewardCategories50E970[:], wantCategories) {
		t.Fatalf("Quest reward categories = %v, want %v", questShopRewardCategories50E970, wantCategories)
	}
}

func TestLoadQuestShopItems50E970ExactSequence(t *testing.T) {
	item, freeItem := alloc.New(Object{})
	defer freeItem()
	marker, freeMarker := alloc.New(Object{})
	defer freeMarker()
	markerData, freeMarkerData := alloc.New(RewardMarkerInitData{})
	defer freeMarkerData()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"item": unsafe.Pointer(item), "marker": unsafe.Pointer(marker), "marker data": unsafe.Pointer(markerData),
		} {
			if uintptr(ptr) <= uintptr(^uint32(0)) {
				t.Fatalf("%s address %#x did not exercise the native high half", name, uintptr(ptr))
			}
		}
	}

	var created []string
	var added []string
	var categories []uint32
	var events []string
	rewardIndex := 0
	loaded, complete := loadQuestShopItems50E970(7, 8, questShopItemLoadHooks50E970{
		newObject: func(typeID string) *Object {
			created = append(created, typeID)
			if typeID == questShopMarkerType50E970 {
				return marker
			}
			return item
		},
		addItem: func(got *Object) bool {
			if got != item {
				t.Fatalf("add item pointer = %p, want native item %p", got, item)
			}
			if rewardIndex == 0 {
				added = append(added, created[len(added)])
			} else {
				added = append(added, fmt.Sprintf("reward-%d", rewardIndex))
			}
			return true
		},
		markerData: func(got *Object) *RewardMarkerInitData {
			if got != marker {
				t.Fatalf("marker-data object = %p, want %p", got, marker)
			}
			return markerData
		},
		activateReward: func(got *Object, stage uint32) *Object {
			if got != marker || stage != 9 {
				t.Fatalf("reward activation = marker %p stage %d, want %p/9", got, stage, marker)
			}
			categories = append(categories, markerData.CategoryMask)
			rewardIndex++
			events = append(events, fmt.Sprintf("reward-%d", rewardIndex))
			return item
		},
		randomInt: func(minimum, maximum int32) int32 {
			if minimum != 0 || maximum != 100 {
				t.Fatalf("bonus RNG bounds = %d..%d, want 0..100", minimum, maximum)
			}
			events = append(events, "random")
			return 90
		},
		delayedDelete: func(got *Object) {
			if got != marker {
				t.Fatalf("deleted marker = %p, want %p", got, marker)
			}
			events = append(events, "delete")
		},
	})
	if !complete || loaded != 64 {
		t.Fatalf("Quest inventory = loaded %d complete %t, want 64/true", loaded, complete)
	}
	wantCreated := append([]string{questShopAnkhType50E970}, questShopBonusItemTypes50E970[:]...)
	wantCreated = append(wantCreated, questShopMarkerType50E970)
	if !reflect.DeepEqual(created, wantCreated) {
		t.Fatalf("created type sequence = %v, want %v", created, wantCreated)
	}
	if !reflect.DeepEqual(categories, questShopRewardCategories50E970[:]) {
		t.Fatalf("reward category sequence = %v, want %v", categories, questShopRewardCategories50E970)
	}
	if len(added) != 64 || added[0] != questShopAnkhType50E970 || added[46] != "WizardHelm" || added[47] != "reward-1" || added[63] != "reward-17" {
		t.Fatalf("added item sequence landmarks/count = %q/%q/%q/%q/%d", added[0], added[46], added[47], added[63], len(added))
	}
	if got := events[len(events)-2:]; !reflect.DeepEqual(got, []string{"random", "delete"}) {
		t.Fatalf("final events = %v, want [random delete]", got)
	}
}

func TestLoadQuestShopItems50E970BonusThresholdAndStageWrap(t *testing.T) {
	for _, tc := range []struct {
		name       string
		draw       int32
		wantLoaded int
		wantCalls  int
		wantLast   uint32
	}{
		{name: "ninety", draw: 90, wantLoaded: 63, wantCalls: 17, wantLast: 4},
		{name: "ninety-one", draw: 91, wantLoaded: 64, wantCalls: 18, wantLast: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marker := &Object{}
			item := &Object{}
			data := &RewardMarkerInitData{}
			var categories []uint32
			var stages []uint32
			deletes := 0
			loaded, complete := loadQuestShopItems50E970(-1, -1, questShopItemLoadHooks50E970{
				newObject: func(typeID string) *Object {
					if typeID == questShopMarkerType50E970 {
						return marker
					}
					return item
				},
				addItem:    func(*Object) bool { return true },
				markerData: func(*Object) *RewardMarkerInitData { return data },
				activateReward: func(_ *Object, uint32Stage uint32) *Object {
					categories = append(categories, data.CategoryMask)
					stages = append(stages, uint32Stage)
					return item
				},
				randomInt:     func(int32, int32) int32 { return tc.draw },
				delayedDelete: func(*Object) { deletes++ },
			})
			if !complete || loaded != tc.wantLoaded || len(categories) != tc.wantCalls || deletes != 1 {
				t.Fatalf("result = loaded %d complete %t calls %d deletes %d, want %d/true/%d/1", loaded, complete, len(categories), deletes, tc.wantLoaded, tc.wantCalls)
			}
			if categories[len(categories)-1] != tc.wantLast {
				t.Fatalf("last category = %d, want %d", categories[len(categories)-1], tc.wantLast)
			}
			for _, stage := range stages {
				if stage != 1 {
					t.Fatalf("wrapped reward stage = %d, want 1", stage)
				}
			}
		})
	}
}

func TestQuestShopRoundFloat32ToInt32_50E970(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{value: 0.5, want: 0},
		{value: 1.5, want: 2},
		{value: 2.5, want: 2},
		{value: -1.5, want: -2},
		{value: float32(math.NaN()), want: math.MinInt32},
		{value: float32(math.Inf(1)), want: math.MinInt32},
		{value: float32(math.Inf(-1)), want: math.MinInt32},
		{value: float32(math.MaxInt32), want: math.MinInt32},
	}
	for _, tc := range tests {
		if got := questShopRoundFloat32ToInt32_50E970(tc.value); got != tc.want {
			t.Errorf("nox_float2int(%v) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestLoadQuestShopItems50E970FailureAndNilRewardSemantics(t *testing.T) {
	marker := &Object{}
	item := &Object{}
	data := &RewardMarkerInitData{}
	addCalls := 0
	rewardCalls := 0
	loaded, complete := loadQuestShopItems50E970(1, 1, questShopItemLoadHooks50E970{
		newObject: func(typeID string) *Object {
			if typeID == questShopMarkerType50E970 {
				return marker
			}
			return item
		},
		addItem: func(*Object) bool {
			addCalls++
			return addCalls != 1
		},
		markerData: func(*Object) *RewardMarkerInitData { return data },
		activateReward: func(*Object, uint32) *Object {
			rewardCalls++
			if rewardCalls == 1 {
				return nil
			}
			return item
		},
		randomInt:     func(int32, int32) int32 { return 0 },
		delayedDelete: func(*Object) {},
	})
	if complete || loaded != 61 || rewardCalls != 17 {
		t.Fatalf("partial result = loaded %d complete %t reward calls %d, want 61/false/17", loaded, complete, rewardCalls)
	}
}
