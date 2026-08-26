package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
)

func TestRewardMarkerActivate4F0720NativeLayoutAndRegistration(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantInitData := uintptr(692)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantInitData = 760
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"RewardMarkerInitData size", unsafe.Sizeof(RewardMarkerInitData{}), 220},
		{"RewardMarkerInitData.CategoryMask", unsafe.Offsetof(RewardMarkerInitData{}.CategoryMask), 0},
		{"RewardMarkerInitData.ChanceMode", unsafe.Offsetof(RewardMarkerInitData{}.ChanceMode), 212},
		{"RewardMarkerInitData.Field216", unsafe.Offsetof(RewardMarkerInitData{}.Field216), 216},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	def, ok := initFuncs["RewardMarkerInit"]
	if !ok {
		t.Fatal("RewardMarkerInit is not registered")
	}
	if def.Func != nil || def.DataSize != unsafe.Sizeof(RewardMarkerInitData{}) || def.DataSize != 220 {
		t.Fatalf("RewardMarkerInit registration = func %p/size %d", def.Func, def.DataSize)
	}
}

func TestRewardMarkerActivateNative4F0720UsesNativePointersAndCachedInitData(t *testing.T) {
	entry := &RewardMarkerInitData{
		CategoryMask: 64,
		ChanceMode:   4,
		Field216:     0xa5a5a5a5,
	}
	replacement := &RewardMarkerInitData{
		CategoryMask: 1,
		ChanceMode:   1,
		Field216:     0x5a5a5a5a,
	}
	marker := &Object{TypeInd: 77, InitData: unsafe.Pointer(entry)}
	created := &Object{NetCode: 0x11223344}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("test marker address %#x does not exercise high native pointers", uintptr(unsafe.Pointer(marker)))
	}

	var cache uint32
	var events []string
	wrong := func(*Object, uint32) *Object {
		t.Fatal("wrong reward creator called")
		return nil
	}
	runtime := RewardMarkerActivateRuntime4F0720{
		SpellBook: wrong, AbilityBook: wrong, FieldGuide: wrong, Weapon: wrong,
		Armor: wrong, Gem: wrong, Gem2: wrong,
		Potion: func(got *Object, stage uint32) *Object {
			events = append(events, "potion")
			if got != marker || stage != 64 {
				t.Fatalf("marker/stage = %p/%d, want %p/64", got, stage, marker)
			}
			return created
		},
	}
	got := rewardMarkerActivateNative4F0720(marker, 62, rewardMarkerActivateNativeDeps4F0720{
		loadCachedPlusType: func() uint32 {
			events = append(events, "load-cache")
			return cache
		},
		lookupType: func(name string) uint32 {
			events = append(events, "lookup:"+name)
			return 77
		},
		storeCachedPlusType: func(value uint32) {
			events = append(events, "store-cache")
			cache = value
		},
		randomInt: func(minimum, maximum int32) int32 {
			if minimum == 0 && maximum == 100 {
				events = append(events, "chance")
				marker.InitData = unsafe.Pointer(replacement)
				return 5
			}
			if minimum != 1 || maximum != 16 {
				t.Fatalf("weighted RNG bounds = %d..%d, want 1..16", minimum, maximum)
			}
			events = append(events, "weighted")
			return 16
		},
		runtime: runtime,
	})
	wantEvents := []string{
		"load-cache", "lookup:RewardMarkerPlus", "store-cache", "chance", "weighted", "potion",
	}
	if got != created || cache != 77 || !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("result/cache/events = %p/%d/%v, want %p/77/%v", got, cache, events, created, wantEvents)
	}
	if marker.InitData != unsafe.Pointer(replacement) || entry.CategoryMask != 64 || entry.ChanceMode != 4 || entry.Field216 != 0xa5a5a5a5 {
		t.Fatalf("cached/live InitData changed unexpectedly: marker=%p entry=%+v", marker.InitData, *entry)
	}
}

func TestRewardMarkerActivate4F0720ServerOwnsDedicatedTypeCache(t *testing.T) {
	typeRecord := &ObjectType{ind: 77, id: "RewardMarkerPlus"}
	srv := new(Server)
	srv.Types.byID = map[string]*ObjectType{"rewardmarkerplus": typeRecord}
	srv.Rand.Logic = prand.New(2011)
	data := &RewardMarkerInitData{CategoryMask: 128}
	marker := &Object{TypeInd: 77, InitData: unsafe.Pointer(data)}
	first := &Object{NetCode: 1}
	second := &Object{NetCode: 2}
	wrong := func(*Object, uint32) *Object {
		t.Fatal("wrong reward creator called")
		return nil
	}
	runtime := RewardMarkerActivateRuntime4F0720{
		SpellBook: wrong, AbilityBook: wrong, FieldGuide: wrong, Weapon: wrong,
		Armor: wrong, Gem: wrong, Potion: wrong,
		Gem2: func(got *Object, stage uint32) *Object {
			if got != marker || stage != 1 {
				t.Fatalf("first marker/stage = %p/%d", got, stage)
			}
			return first
		},
	}
	if got := srv.RewardMarkerActivate4F0720(marker, math.MaxUint32, runtime); got != first {
		t.Fatalf("first result = %p, want %p", got, first)
	}
	if srv.Types.fast.rewardMarkerPlus != 77 {
		t.Fatalf("dedicated cache = %d, want 77", srv.Types.fast.rewardMarkerPlus)
	}

	delete(srv.Types.byID, "rewardmarkerplus")
	data.CategoryMask = 1
	runtime.Gem2 = wrong
	runtime.SpellBook = func(got *Object, stage uint32) *Object {
		if got != marker || stage != 2 {
			t.Fatalf("second marker/stage = %p/%d", got, stage)
		}
		return second
	}
	if got := srv.RewardMarkerActivate4F0720(marker, 0, runtime); got != second {
		t.Fatalf("second result = %p, want %p", got, second)
	}
	if srv.Types.fast.rewardMarkerPlus != 77 {
		t.Fatalf("cached type was lost after source map removal: %d", srv.Types.fast.rewardMarkerPlus)
	}
}

func TestRewardMarkerActivateNative4F0720FaultBoundaries(t *testing.T) {
	t.Run("nil-marker-before-lookup", func(t *testing.T) {
		loads := 0
		lookups := 0
		defer func() {
			if recover() == nil {
				t.Fatal("nil marker did not fault on InitData load")
			}
			if loads != 1 || lookups != 0 {
				t.Fatalf("cache loads/lookups = %d/%d, want 1/0", loads, lookups)
			}
		}()
		rewardMarkerActivateNative4F0720(nil, 0, rewardMarkerActivateNativeDeps4F0720{
			loadCachedPlusType: func() uint32 { loads++; return 0 },
			lookupType:         func(string) uint32 { lookups++; return 1 },
		})
	})

	t.Run("nil-init-after-cache-and-type", func(t *testing.T) {
		lookups := 0
		stores := 0
		randomCalls := 0
		marker := &Object{TypeInd: 77}
		defer func() {
			if recover() == nil {
				t.Fatal("nil InitData did not fault on ChanceMode load")
			}
			if lookups != 1 || stores != 1 || randomCalls != 0 {
				t.Fatalf("lookup/store/random = %d/%d/%d, want 1/1/0", lookups, stores, randomCalls)
			}
		}()
		rewardMarkerActivateNative4F0720(marker, 0, rewardMarkerActivateNativeDeps4F0720{
			loadCachedPlusType:  func() uint32 { return 0 },
			lookupType:          func(string) uint32 { lookups++; return 77 },
			storeCachedPlusType: func(uint32) { stores++ },
			randomInt:           func(int32, int32) int32 { randomCalls++; return 0 },
		})
	})
}
