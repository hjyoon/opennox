package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type glyphInventoryCountObject4EF6F0 struct {
	name    string
	typeInd uint16
	first   *glyphInventoryCountObject4EF6F0
	next    *glyphInventoryCountObject4EF6F0
}

type glyphInventoryCountWorld4EF6F0 struct {
	events       []string
	faultAt      int
	cache        uint32
	lookupResult uint32
	onEvent      func(string)
}

func (w *glyphInventoryCountWorld4EF6F0) event(event string) {
	w.events = append(w.events, event)
	if w.onEvent != nil {
		w.onEvent(event)
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *glyphInventoryCountWorld4EF6F0) hooks() glyphInventoryCountHooks4EF6F0[*glyphInventoryCountObject4EF6F0] {
	return glyphInventoryCountHooks4EF6F0[*glyphInventoryCountObject4EF6F0]{
		loadCache: func() uint32 {
			value := w.cache
			w.event(fmt.Sprintf("cache:%#x", value))
			return value
		},
		lookupType: func(name string) uint32 {
			w.event("lookup:" + name)
			return w.lookupResult
		},
		storeCache: func(value uint32) {
			w.event(fmt.Sprintf("store:%#x", value))
			w.cache = value
		},
		loadFirst: func(owner *glyphInventoryCountObject4EF6F0) *glyphInventoryCountObject4EF6F0 {
			name := "nil"
			if owner != nil {
				name = owner.name
			}
			w.event("first:" + name)
			if owner == nil {
				return nil
			}
			return owner.first
		},
		loadType: func(item *glyphInventoryCountObject4EF6F0) uint16 {
			w.event("type:" + item.name)
			return item.typeInd
		},
		loadNext: func(item *glyphInventoryCountObject4EF6F0) *glyphInventoryCountObject4EF6F0 {
			w.event("next:" + item.name)
			return item.next
		},
	}
}

func newGlyphInventoryCountWorld4EF6F0() (
	*glyphInventoryCountWorld4EF6F0,
	*glyphInventoryCountObject4EF6F0,
) {
	third := &glyphInventoryCountObject4EF6F0{name: "third", typeInd: 8}
	second := &glyphInventoryCountObject4EF6F0{name: "destroyed", typeInd: 7, next: third}
	first := &glyphInventoryCountObject4EF6F0{name: "first", typeInd: 7, next: second}
	owner := &glyphInventoryCountObject4EF6F0{name: "owner", first: first}
	return &glyphInventoryCountWorld4EF6F0{lookupResult: 7}, owner
}

func TestGlyphInventoryCount4EF6F0ZeroCacheExactOrder(t *testing.T) {
	world, owner := newGlyphInventoryCountWorld4EF6F0()
	if got := glyphInventoryCount4EF6F0(owner, world.hooks()); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	want := []string{
		"cache:0x0", "lookup:Glyph", "store:0x7", "first:owner",
		"cache:0x7", "type:first", "next:first",
		"cache:0x7", "type:destroyed", "next:destroyed",
		"cache:0x7", "type:third", "next:third",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
	if world.cache != 7 {
		t.Fatalf("cache = %#x, want 7", world.cache)
	}
}

func TestGlyphInventoryCount4EF6F0NonzeroWholeCacheSkipsLookup(t *testing.T) {
	item := &glyphInventoryCountObject4EF6F0{name: "item", typeInd: 7}
	owner := &glyphInventoryCountObject4EF6F0{name: "owner", first: item}
	world := &glyphInventoryCountWorld4EF6F0{cache: 0x00010007, lookupResult: 7}

	if got := glyphInventoryCount4EF6F0(owner, world.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"cache:0x10007", "first:owner", "cache:0x10007", "type:item", "next:item"}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestGlyphInventoryCount4EF6F0ZeroLookupIsStoredWithoutSameCallRetry(t *testing.T) {
	item := &glyphInventoryCountObject4EF6F0{name: "zero", typeInd: 0}
	owner := &glyphInventoryCountObject4EF6F0{name: "owner", first: item}
	world := &glyphInventoryCountWorld4EF6F0{}

	if got := glyphInventoryCount4EF6F0(owner, world.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"cache:0x0", "lookup:Glyph", "store:0x0", "first:owner",
		"cache:0x0", "type:zero", "next:zero",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestGlyphInventoryCount4EF6F0UsesLiveCacheAndSuccessor(t *testing.T) {
	replacement := &glyphInventoryCountObject4EF6F0{name: "replacement", typeInd: 9}
	stale := &glyphInventoryCountObject4EF6F0{name: "stale", typeInd: 8}
	first := &glyphInventoryCountObject4EF6F0{name: "first", typeInd: 8, next: stale}
	owner := &glyphInventoryCountObject4EF6F0{name: "owner", first: first}
	world := &glyphInventoryCountWorld4EF6F0{cache: 7}
	world.onEvent = func(event string) {
		switch event {
		case "first:owner":
			world.cache = 8
		case "type:first":
			world.cache = 9
			first.next = replacement
		}
	}

	if got := glyphInventoryCount4EF6F0(owner, world.hooks()); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	want := []string{
		"cache:0x7", "first:owner",
		"cache:0x8", "type:first", "next:first",
		"cache:0x9", "type:replacement", "next:replacement",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestGlyphInventoryCount4EF6F0DoesNotGuardNilOwner(t *testing.T) {
	world := &glyphInventoryCountWorld4EF6F0{cache: 7}
	if got := glyphInventoryCount4EF6F0[*glyphInventoryCountObject4EF6F0](nil, world.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"cache:0x7", "first:nil"}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestGlyphInventoryCount4EF6F0EveryObservableFaultPrefix(t *testing.T) {
	base, owner := newGlyphInventoryCountWorld4EF6F0()
	glyphInventoryCount4EF6F0(owner, base.hooks())
	want := append([]string(nil), base.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
			world, owner := newGlyphInventoryCountWorld4EF6F0()
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				glyphInventoryCount4EF6F0(owner, world.hooks())
			}()
			if !reflect.DeepEqual(world.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
			}
		})
	}
}

func TestIncrementGlyphInventoryCount4EF6F0WrapsInt32(t *testing.T) {
	if got := incrementGlyphInventoryCount4EF6F0(math.MaxInt32); got != math.MinInt32 {
		t.Fatalf("increment(MaxInt32) = %d, want %d", got, int32(math.MinInt32))
	}
}
