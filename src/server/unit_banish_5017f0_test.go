package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitBanishObject5017F0 struct {
	name      string
	typeIndex uint16
	first     *unitBanishObject5017F0
	next      *unitBanishObject5017F0
}

type unitBanishWorld5017F0 struct {
	events       []string
	faultAt      int
	cache        uint32
	lookupResult uint32
	onDelete     func(*unitBanishObject5017F0)
}

func (w *unitBanishWorld5017F0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *unitBanishWorld5017F0) hooks() unitBanishHooks5017F0[*unitBanishObject5017F0] {
	return unitBanishHooks5017F0[*unitBanishObject5017F0]{
		loadGlyphCache: func() uint32 {
			value := w.cache
			w.event(fmt.Sprintf("cache:%#x", value))
			return value
		},
		lookupType: func(name string) uint32 {
			w.event("lookup:" + name)
			return w.lookupResult
		},
		storeGlyphCache: func(value uint32) {
			w.event(fmt.Sprintf("store:%#x", value))
			w.cache = value
		},
		loadFirst: func(unit *unitBanishObject5017F0) *unitBanishObject5017F0 {
			w.event("first:" + unit.name)
			return unit.first
		},
		loadNext: func(item *unitBanishObject5017F0) *unitBanishObject5017F0 {
			w.event("next:" + item.name)
			return item.next
		},
		loadTypeIndex: func(item *unitBanishObject5017F0) uint16 {
			w.event("type:" + item.name)
			return item.typeIndex
		},
		delayedDelete: func(object *unitBanishObject5017F0) {
			w.event("delete:" + object.name)
			if w.onDelete != nil {
				w.onDelete(object)
			}
		},
		sendPointFX: func(code uint8, unit *unitBanishObject5017F0) {
			w.event(fmt.Sprintf("fx:%#x:%s", code, unit.name))
		},
	}
}

func newUnitBanishWorld5017F0() (*unitBanishWorld5017F0, *unitBanishObject5017F0) {
	third := &unitBanishObject5017F0{name: "third", typeIndex: 7}
	second := &unitBanishObject5017F0{name: "second", typeIndex: 8, next: third}
	first := &unitBanishObject5017F0{name: "first", typeIndex: 7, next: second}
	unit := &unitBanishObject5017F0{name: "unit", first: first}
	return &unitBanishWorld5017F0{lookupResult: 7}, unit
}

func TestUnitBanish5017F0ZeroCacheExactOrder(t *testing.T) {
	world, unit := newUnitBanishWorld5017F0()
	unitBanish5017F0(unit, world.hooks())
	want := []string{
		"cache:0x0", "lookup:Glyph", "store:0x7", "first:unit",
		"cache:0x7", "next:first", "type:first", "delete:first",
		"cache:0x7", "next:second", "type:second",
		"cache:0x7", "next:third", "type:third", "delete:third",
		"fx:0x81:unit", "delete:unit",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
	if world.cache != 7 {
		t.Fatalf("cache = %#x, want 7", world.cache)
	}
}

func TestUnitBanish5017F0NilUnitStillInitializesCache(t *testing.T) {
	world := &unitBanishWorld5017F0{lookupResult: 9}
	unitBanish5017F0[*unitBanishObject5017F0](nil, world.hooks())
	want := []string{"cache:0x0", "lookup:Glyph", "store:0x9"}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestUnitBanish5017F0UsesWholeCacheAndZeroExtendedType(t *testing.T) {
	item := &unitBanishObject5017F0{name: "item", typeIndex: 7}
	unit := &unitBanishObject5017F0{name: "unit", first: item}
	world := &unitBanishWorld5017F0{cache: 0x00010007, lookupResult: 7}
	unitBanish5017F0(unit, world.hooks())
	want := []string{
		"cache:0x10007", "first:unit", "cache:0x10007", "next:item", "type:item",
		"fx:0x81:unit", "delete:unit",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestUnitBanish5017F0SnapshotsSuccessorButReloadsCache(t *testing.T) {
	replacement := &unitBanishObject5017F0{name: "replacement", typeIndex: 8}
	second := &unitBanishObject5017F0{name: "second", typeIndex: 8}
	first := &unitBanishObject5017F0{name: "first", typeIndex: 7, next: second}
	unit := &unitBanishObject5017F0{name: "unit", first: first}
	world := &unitBanishWorld5017F0{cache: 7}
	world.onDelete = func(object *unitBanishObject5017F0) {
		if object == first {
			first.next = replacement
			world.cache = 8
		}
	}

	unitBanish5017F0(unit, world.hooks())
	want := []string{
		"cache:0x7", "first:unit",
		"cache:0x7", "next:first", "type:first", "delete:first",
		"cache:0x8", "next:second", "type:second", "delete:second",
		"fx:0x81:unit", "delete:unit",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestUnitBanish5017F0ZeroLookupIsNotRetriedWithinCall(t *testing.T) {
	item := &unitBanishObject5017F0{name: "zero", typeIndex: 0}
	unit := &unitBanishObject5017F0{name: "unit", first: item}
	world := &unitBanishWorld5017F0{}
	unitBanish5017F0(unit, world.hooks())
	want := []string{
		"cache:0x0", "lookup:Glyph", "store:0x0", "first:unit",
		"cache:0x0", "next:zero", "type:zero", "delete:zero",
		"fx:0x81:unit", "delete:unit",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestUnitBanish5017F0EveryObservableFaultPrefix(t *testing.T) {
	base, unit := newUnitBanishWorld5017F0()
	unitBanish5017F0(unit, base.hooks())
	want := append([]string(nil), base.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
			world, unit := newUnitBanishWorld5017F0()
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				unitBanish5017F0(unit, world.hooks())
			}()
			if !reflect.DeepEqual(world.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
			}
		})
	}
}
