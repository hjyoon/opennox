package legacy

import (
	"reflect"
	"testing"
)

type unitOwnedTypeCache4E7BE0 struct {
	value uint32
}

type unitOwnedTypeObject4E7BE0 struct {
	typeInd uint16
	first   *unitOwnedTypeObject4E7BE0
	next    *unitOwnedTypeObject4E7BE0
}

func unitOwnedTypeTestHooks4E7BE0(events *[]string) unitOwnedTypeHooks4E7BE0[
	*unitOwnedTypeCache4E7BE0,
	*unitOwnedTypeObject4E7BE0,
] {
	return unitOwnedTypeHooks4E7BE0[*unitOwnedTypeCache4E7BE0, *unitOwnedTypeObject4E7BE0]{
		loadCache: func(cache *unitOwnedTypeCache4E7BE0) uint32 {
			*events = append(*events, "cache")
			return cache.value
		},
		storeCache: func(cache *unitOwnedTypeCache4E7BE0, value uint32) {
			*events = append(*events, "store")
			cache.value = value
		},
		lookupType: func(name string) uint32 {
			*events = append(*events, "lookup:"+name)
			return 0x2468
		},
		loadFirst: func(owner *unitOwnedTypeObject4E7BE0) *unitOwnedTypeObject4E7BE0 {
			*events = append(*events, "first")
			return owner.first
		},
		loadType: func(obj *unitOwnedTypeObject4E7BE0) uint16 {
			*events = append(*events, "type")
			return obj.typeInd
		},
		loadNext: func(obj *unitOwnedTypeObject4E7BE0) *unitOwnedTypeObject4E7BE0 {
			*events = append(*events, "next")
			return obj.next
		},
	}
}

func TestUnitOwnedType4E7BE0CacheMissPrecedesOwnerRead(t *testing.T) {
	cache := &unitOwnedTypeCache4E7BE0{}
	oldItem := &unitOwnedTypeObject4E7BE0{typeInd: 1}
	match := &unitOwnedTypeObject4E7BE0{typeInd: 0x2468}
	owner := &unitOwnedTypeObject4E7BE0{first: oldItem}
	var events []string
	hooks := unitOwnedTypeTestHooks4E7BE0(&events)
	hooks.lookupType = func(name string) uint32 {
		events = append(events, "lookup:"+name)
		owner.first = match
		return 0x2468
	}

	if got := unitIsCrown4E7BE0(cache, owner, hooks); got != 1 {
		t.Fatalf("Crown result = %d, want 1", got)
	}
	want := []string{"cache", "lookup:Crown", "store", "first", "type"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if cache.value != 0x2468 || owner.first != match || oldItem.typeInd != 1 || match.typeInd != 0x2468 {
		t.Fatal("cache or owned objects changed outside the required lookup mutation")
	}
}

func TestUnitOwnedType4E7BE0CacheHitAndMatchSkipCallbacks(t *testing.T) {
	cache := &unitOwnedTypeCache4E7BE0{value: 7}
	item := &unitOwnedTypeObject4E7BE0{typeInd: 7}
	owner := &unitOwnedTypeObject4E7BE0{first: item}
	var events []string
	hooks := unitOwnedTypeTestHooks4E7BE0(&events)
	hooks.lookupType = func(string) uint32 {
		t.Fatal("cache hit performed a lookup")
		return 0
	}
	hooks.storeCache = func(*unitOwnedTypeCache4E7BE0, uint32) {
		t.Fatal("cache hit performed a store")
	}
	hooks.loadNext = func(*unitOwnedTypeObject4E7BE0) *unitOwnedTypeObject4E7BE0 {
		t.Fatal("matching object loaded its successor")
		return nil
	}

	if got := unitIsGameBall4E7C30(cache, owner, hooks); got != 1 {
		t.Fatalf("GameBall result = %d, want 1", got)
	}
	if want := []string{"cache", "first", "type"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if cache.value != 7 || owner.first != item || item.typeInd != 7 {
		t.Fatal("cache hit mutated its inputs")
	}
}

func TestUnitOwnedType4E7BE0TraversalOrder(t *testing.T) {
	third := &unitOwnedTypeObject4E7BE0{typeInd: 9}
	second := &unitOwnedTypeObject4E7BE0{typeInd: 8, next: third}
	first := &unitOwnedTypeObject4E7BE0{typeInd: 7, next: second}
	owner := &unitOwnedTypeObject4E7BE0{first: first}
	cache := &unitOwnedTypeCache4E7BE0{value: 8}
	var events []string
	hooks := unitOwnedTypeTestHooks4E7BE0(&events)

	if got := unitIsCrown4E7BE0(cache, owner, hooks); got != 1 {
		t.Fatalf("middle match result = %d, want 1", got)
	}
	want := []string{"cache", "first", "type", "next", "type"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if first.next != second || second.next != third || third.next != nil {
		t.Fatal("traversal mutated the owned list")
	}

	events = nil
	cache.value = 6
	if got := unitIsCrown4E7BE0(cache, owner, hooks); got != 0 {
		t.Fatalf("missing result = %d, want 0", got)
	}
	want = []string{"cache", "first", "type", "next", "type", "next", "type", "next"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("missing events = %v, want %v", events, want)
	}
}

func TestUnitOwnedType4E7BE0UsesZeroExtendedType(t *testing.T) {
	item := &unitOwnedTypeObject4E7BE0{typeInd: 1}
	owner := &unitOwnedTypeObject4E7BE0{first: item}
	cache := &unitOwnedTypeCache4E7BE0{value: 0x00010001}
	var events []string
	hooks := unitOwnedTypeTestHooks4E7BE0(&events)

	if got := unitIsCrown4E7BE0(cache, owner, hooks); got != 0 {
		t.Fatalf("32-bit cache versus 16-bit type result = %d, want 0", got)
	}
	want := []string{"cache", "first", "type", "next"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitOwnedType4E7BE0ZeroLookupCanMatchAndRepeats(t *testing.T) {
	item := &unitOwnedTypeObject4E7BE0{}
	owner := &unitOwnedTypeObject4E7BE0{first: item}
	cache := &unitOwnedTypeCache4E7BE0{}
	var events []string
	hooks := unitOwnedTypeTestHooks4E7BE0(&events)
	hooks.lookupType = func(name string) uint32 {
		events = append(events, "lookup:"+name)
		return 0
	}

	for i := 0; i < 2; i++ {
		if got := unitIsGameBall4E7C30(cache, owner, hooks); got != 1 {
			t.Fatalf("zero-ID call %d result = %d, want 1", i+1, got)
		}
	}
	wantOne := []string{"cache", "lookup:GameBall", "store", "first", "type"}
	want := append(append([]string{}, wantOne...), wantOne...)
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if cache.value != 0 {
		t.Fatalf("zero lookup cache = %#x, want zero", cache.value)
	}
}

func TestUnitOwnedType4E7BE0NullOwnerFaultOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		initial    uint32
		wantEvents []string
		wantCache  uint32
	}{
		{name: "cache hit", initial: 9, wantEvents: []string{"cache", "first"}, wantCache: 9},
		{name: "cache miss", wantEvents: []string{"cache", "lookup:Crown", "store", "first"}, wantCache: 0x2468},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := &unitOwnedTypeCache4E7BE0{value: tc.initial}
			var events []string
			hooks := unitOwnedTypeTestHooks4E7BE0(&events)
			defer func() {
				if recover() == nil {
					t.Fatal("null owner did not fault")
				}
				if !reflect.DeepEqual(events, tc.wantEvents) {
					t.Fatalf("events = %v, want %v", events, tc.wantEvents)
				}
				if cache.value != tc.wantCache {
					t.Fatalf("cache = %#x, want %#x", cache.value, tc.wantCache)
				}
			}()
			_ = unitIsCrown4E7BE0(cache, (*unitOwnedTypeObject4E7BE0)(nil), hooks)
		})
	}
}

func TestUnitOwnedType4E7BE0EmptyListReturnsZero(t *testing.T) {
	cache := &unitOwnedTypeCache4E7BE0{value: 3}
	owner := &unitOwnedTypeObject4E7BE0{}
	var events []string
	hooks := unitOwnedTypeTestHooks4E7BE0(&events)
	if got := unitIsCrown4E7BE0(cache, owner, hooks); got != 0 {
		t.Fatalf("empty list result = %d, want 0", got)
	}
	if want := []string{"cache", "first"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
