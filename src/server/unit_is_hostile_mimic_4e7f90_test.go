package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
)

type unitIsHostileMimicCache4E7F90 struct {
	value uint32
}

type unitIsHostileMimicObject4E7F90 struct {
	typeInd  uint16
	classLow uint8
	owner    *unitIsHostileMimicObject4E7F90
}

func unitIsHostileMimicTestHooks4E7F90(
	events *[]string,
	cache *unitIsHostileMimicCache4E7F90,
	enemy func(*unitIsHostileMimicObject4E7F90, *unitIsHostileMimicObject4E7F90) int32,
	quest func() int32,
) unitIsHostileMimicHooks4E7F90[*unitIsHostileMimicObject4E7F90] {
	return unitIsHostileMimicHooks4E7F90[*unitIsHostileMimicObject4E7F90]{
		loadMimicCache: func() uint32 {
			*events = append(*events, "cache")
			return cache.value
		},
		lookupType: func(name string) uint32 {
			*events = append(*events, "lookup:"+name)
			return 0x2468
		},
		storeMimicCache: func(value uint32) {
			*events = append(*events, "store")
			cache.value = value
		},
		isEnemy: func(obj, obj2 *unitIsHostileMimicObject4E7F90) int32 {
			*events = append(*events, "enemy")
			return enemy(obj, obj2)
		},
		isQuest: func() int32 {
			*events = append(*events, "quest")
			return quest()
		},
		loadType: func(obj *unitIsHostileMimicObject4E7F90) uint16 {
			*events = append(*events, "type")
			return obj.typeInd
		},
		loadClassLow: func(obj *unitIsHostileMimicObject4E7F90) uint8 {
			*events = append(*events, "class")
			return obj.classLow
		},
		loadOwner: func(obj *unitIsHostileMimicObject4E7F90) *unitIsHostileMimicObject4E7F90 {
			*events = append(*events, "owner")
			return obj.owner
		},
	}
}

func TestUnitIsHostileMimic4E7F90CachePrecedesNullChecks(t *testing.T) {
	obj2 := &unitIsHostileMimicObject4E7F90{}
	cache := &unitIsHostileMimicCache4E7F90{}
	var events []string
	hooks := unitIsHostileMimicTestHooks4E7F90(&events, cache,
		func(*unitIsHostileMimicObject4E7F90, *unitIsHostileMimicObject4E7F90) int32 {
			t.Fatal("nil input reached enemy callback")
			return 0
		},
		func() int32 {
			t.Fatal("nil input reached Quest callback")
			return 0
		},
	)

	if got := unitIsHostileMimic4E7F90((*unitIsHostileMimicObject4E7F90)(nil), obj2, hooks); got != 0 {
		t.Fatalf("nil first result = %d, want 0", got)
	}
	if want := []string{"cache", "lookup:Mimic", "store"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("nil first events = %v, want %v", events, want)
	}
	if cache.value != 0x2468 {
		t.Fatalf("cache after nil first = %#x, want %#x", cache.value, uint32(0x2468))
	}

	events = nil
	if got := unitIsHostileMimic4E7F90(obj2, (*unitIsHostileMimicObject4E7F90)(nil), hooks); got != 0 {
		t.Fatalf("nil second result = %d, want 0", got)
	}
	if want := []string{"cache"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("nil second events = %v, want %v", events, want)
	}
}

func TestUnitIsHostileMimic4E7F90EnemyUsesZeroAndQuestUsesNonzero(t *testing.T) {
	obj := &unitIsHostileMimicObject4E7F90{}
	obj2 := &unitIsHostileMimicObject4E7F90{}
	for _, enemy := range []int32{-1, 0, 1, 2} {
		t.Run(fmt.Sprintf("enemy_%d", enemy), func(t *testing.T) {
			cache := &unitIsHostileMimicCache4E7F90{value: 7}
			var events []string
			hooks := unitIsHostileMimicTestHooks4E7F90(&events, cache,
				func(*unitIsHostileMimicObject4E7F90, *unitIsHostileMimicObject4E7F90) int32 { return enemy },
				func() int32 { return -1 },
			)
			// The live cache is 7 and the default object type is zero, so a
			// nonzero Quest result still stops at the type comparison.
			got := unitIsHostileMimic4E7F90(obj, obj2, hooks)
			var wantResult int32
			if enemy == 0 {
				wantResult = 1
			}
			if got != wantResult {
				t.Fatalf("result = %d, want %d", got, wantResult)
			}
			want := []string{"cache", "enemy", "quest", "cache", "type"}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestUnitIsHostileMimic4E7F90ReloadsLiveStateInOriginalOrder(t *testing.T) {
	obj := &unitIsHostileMimicObject4E7F90{classLow: uint8(object.ClassPlayer)}
	obj2 := &unitIsHostileMimicObject4E7F90{typeInd: 1}
	cache := &unitIsHostileMimicCache4E7F90{value: 7}
	var events []string
	hooks := unitIsHostileMimicTestHooks4E7F90(&events, cache,
		func(*unitIsHostileMimicObject4E7F90, *unitIsHostileMimicObject4E7F90) int32 {
			cache.value = 8
			return 0
		},
		func() int32 {
			cache.value = 9
			return 2
		},
	)
	cacheLoads := 0
	hooks.loadMimicCache = func() uint32 {
		events = append(events, "cache")
		cacheLoads++
		if cacheLoads == 2 {
			obj2.typeInd = uint16(cache.value)
		}
		return cache.value
	}

	if got := unitIsHostileMimic4E7F90(obj, obj2, hooks); got != 0 {
		t.Fatalf("Quest Mimic exception result = %d, want 0", got)
	}
	want := []string{"cache", "enemy", "quest", "cache", "type", "class", "owner"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if cacheLoads != 2 || cache.value != 9 || obj2.typeInd != 9 {
		t.Fatalf("live reload state = loads %d, cache %#x, type %#x", cacheLoads, cache.value, obj2.typeInd)
	}
}

func TestUnitIsHostileMimic4E7F90ConditionalFieldReads(t *testing.T) {
	owner := &unitIsHostileMimicObject4E7F90{}
	tests := []struct {
		name       string
		quest      int32
		cache      uint32
		typeInd    uint16
		classLow   uint8
		owner      *unitIsHostileMimicObject4E7F90
		want       int32
		wantEvents []string
	}{
		{name: "not quest", quest: 0, cache: 7, want: 1, wantEvents: []string{"cache", "enemy", "quest"}},
		{name: "type mismatch", quest: 1, cache: 7, typeInd: 8, want: 1, wantEvents: []string{"cache", "enemy", "quest", "cache", "type"}},
		{name: "class mismatch", quest: 1, cache: 7, typeInd: 7, want: 1, wantEvents: []string{"cache", "enemy", "quest", "cache", "type", "class"}},
		{name: "owned mimic", quest: 1, cache: 7, typeInd: 7, classLow: uint8(object.ClassPlayer), owner: owner, want: 1, wantEvents: []string{"cache", "enemy", "quest", "cache", "type", "class", "owner"}},
		{name: "unowned mimic", quest: 1, cache: 7, typeInd: 7, classLow: uint8(object.ClassPlayer), want: 0, wantEvents: []string{"cache", "enemy", "quest", "cache", "type", "class", "owner"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := &unitIsHostileMimicObject4E7F90{classLow: tc.classLow}
			obj2 := &unitIsHostileMimicObject4E7F90{typeInd: tc.typeInd, owner: tc.owner}
			cache := &unitIsHostileMimicCache4E7F90{value: tc.cache}
			var events []string
			hooks := unitIsHostileMimicTestHooks4E7F90(&events, cache,
				func(*unitIsHostileMimicObject4E7F90, *unitIsHostileMimicObject4E7F90) int32 { return 0 },
				func() int32 { return tc.quest },
			)
			if got := unitIsHostileMimic4E7F90(obj, obj2, hooks); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestUnitIsHostileMimic4E7F90ZeroLookupStoresAndRepeats(t *testing.T) {
	obj := &unitIsHostileMimicObject4E7F90{classLow: uint8(object.ClassPlayer)}
	obj2 := &unitIsHostileMimicObject4E7F90{}
	cache := &unitIsHostileMimicCache4E7F90{}
	var events []string
	hooks := unitIsHostileMimicTestHooks4E7F90(&events, cache,
		func(*unitIsHostileMimicObject4E7F90, *unitIsHostileMimicObject4E7F90) int32 { return 0 },
		func() int32 { return 1 },
	)
	hooks.lookupType = func(name string) uint32 {
		events = append(events, "lookup:"+name)
		return 0
	}

	for i := 0; i < 2; i++ {
		if got := unitIsHostileMimic4E7F90(obj, obj2, hooks); got != 0 {
			t.Fatalf("call %d result = %d, want 0", i+1, got)
		}
	}
	wantOne := []string{"cache", "lookup:Mimic", "store", "enemy", "quest", "cache", "type", "class", "owner"}
	want := append(append([]string{}, wantOne...), wantOne...)
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if cache.value != 0 {
		t.Fatalf("zero lookup cache = %#x, want zero", cache.value)
	}
}
