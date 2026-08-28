package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupAnkhTradableTestUpdate4F3DD0 struct {
	name       string
	extraLives uint32
}

type pickupAnkhTradableTestObject4F3DD0 struct {
	name     string
	classLow uint8
	update   *pickupAnkhTradableTestUpdate4F3DD0
}

type pickupAnkhTradableTestWorld4F3DD0 struct {
	item        *pickupAnkhTradableTestObject4F3DD0
	events      []string
	faultAt     int
	afterLoad   func(*pickupAnkhTradableTestWorld4F3DD0)
	afterStore  func(*pickupAnkhTradableTestWorld4F3DD0)
	afterDelete func(*pickupAnkhTradableTestWorld4F3DD0)
}

func pickupAnkhTradableTestObjectName4F3DD0(obj *pickupAnkhTradableTestObject4F3DD0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *pickupAnkhTradableTestWorld4F3DD0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupAnkhTradableTestWorld4F3DD0) hooks() pickupAnkhTradableHooks4F3DD0[
	*pickupAnkhTradableTestObject4F3DD0,
	*pickupAnkhTradableTestUpdate4F3DD0,
] {
	return pickupAnkhTradableHooks4F3DD0[
		*pickupAnkhTradableTestObject4F3DD0,
		*pickupAnkhTradableTestUpdate4F3DD0,
	]{
		loadOwnerClassLow: func(owner *pickupAnkhTradableTestObject4F3DD0) uint8 {
			value := owner.classLow
			w.event(fmt.Sprintf("class:%s=%02x", pickupAnkhTradableTestObjectName4F3DD0(owner), value))
			return value
		},
		loadOwnerUpdate: func(owner *pickupAnkhTradableTestObject4F3DD0) *pickupAnkhTradableTestUpdate4F3DD0 {
			update := owner.update
			w.event("update:" + update.name)
			if w.afterLoad != nil {
				w.afterLoad(w)
			}
			return update
		},
		loadExtraLives: func(update *pickupAnkhTradableTestUpdate4F3DD0) uint32 {
			value := update.extraLives
			w.event(fmt.Sprintf("extra:%s=%08x", update.name, value))
			return value
		},
		storeExtraLives: func(update *pickupAnkhTradableTestUpdate4F3DD0, value uint32) {
			w.event(fmt.Sprintf("store:%s=%08x", update.name, value))
			update.extraLives = value
			if w.afterStore != nil {
				w.afterStore(w)
			}
		},
		loadItemArg: func() *pickupAnkhTradableTestObject4F3DD0 {
			w.event("item:" + pickupAnkhTradableTestObjectName4F3DD0(w.item))
			return w.item
		},
		delayedDelete: func(item *pickupAnkhTradableTestObject4F3DD0) {
			w.event("delete:" + pickupAnkhTradableTestObjectName4F3DD0(item))
			if w.afterDelete != nil {
				w.afterDelete(w)
			}
		},
		audio: func(sound uint32, owner *pickupAnkhTradableTestObject4F3DD0, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", sound, pickupAnkhTradableTestObjectName4F3DD0(owner), kind, code))
		},
	}
}

func pickupAnkhTradableSuccessBuild4F3DD0() (
	*pickupAnkhTradableTestWorld4F3DD0,
	*pickupAnkhTradableTestObject4F3DD0,
	*pickupAnkhTradableTestUpdate4F3DD0,
) {
	update := &pickupAnkhTradableTestUpdate4F3DD0{name: "entry-update", extraLives: 9}
	owner := &pickupAnkhTradableTestObject4F3DD0{name: "owner", classLow: 0xa4, update: update}
	item := &pickupAnkhTradableTestObject4F3DD0{name: "item"}
	return &pickupAnkhTradableTestWorld4F3DD0{item: item}, owner, update
}

func pickupAnkhTradableSuccessTrace4F3DD0() []string {
	return []string{
		"class:owner=a4",
		"update:entry-update",
		"extra:entry-update=00000009",
		"store:entry-update=0000000a",
		"item:item",
		"delete:item",
		"audio:1004:owner:0:00000000",
	}
}

func TestPickupAnkhTradable4F3DD0PlayerOrderAndCanonicalOne(t *testing.T) {
	w, owner, update := pickupAnkhTradableSuccessBuild4F3DD0()
	if got := pickupAnkhTradable4F3DD0(owner, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if update.extraLives != 10 {
		t.Fatalf("ExtraLives = %d, want 10", update.extraLives)
	}
	want := pickupAnkhTradableSuccessTrace4F3DD0()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupAnkhTradable4F3DD0NonPlayerStopsAfterClass(t *testing.T) {
	owner := &pickupAnkhTradableTestObject4F3DD0{name: "owner", classLow: 0xfb}
	w := &pickupAnkhTradableTestWorld4F3DD0{}
	if got := pickupAnkhTradable4F3DD0(owner, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"class:owner=fb"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupAnkhTradable4F3DD0WrapsUint32ExtraLives(t *testing.T) {
	w, owner, update := pickupAnkhTradableSuccessBuild4F3DD0()
	update.extraLives = math.MaxUint32
	if got := pickupAnkhTradable4F3DD0(owner, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if update.extraLives != 0 {
		t.Fatalf("ExtraLives = %#x, want 0", update.extraLives)
	}
}

func TestPickupAnkhTradable4F3DD0CachesUpdateAndOwner(t *testing.T) {
	w, owner, entryUpdate := pickupAnkhTradableSuccessBuild4F3DD0()
	replacement := &pickupAnkhTradableTestUpdate4F3DD0{name: "replacement", extraLives: 77}
	replacementOwner := &pickupAnkhTradableTestObject4F3DD0{name: "replacement-owner"}
	currentOwner := owner
	w.afterLoad = func(*pickupAnkhTradableTestWorld4F3DD0) {
		owner.update = replacement
	}
	w.afterDelete = func(*pickupAnkhTradableTestWorld4F3DD0) {
		currentOwner = replacementOwner
	}
	if got := pickupAnkhTradable4F3DD0(owner, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if entryUpdate.extraLives != 10 || replacement.extraLives != 77 {
		t.Fatalf("cached/replacement ExtraLives = %d/%d, want 10/77", entryUpdate.extraLives, replacement.extraLives)
	}
	if currentOwner != replacementOwner {
		t.Fatalf("external owner reference = %p, want %p", currentOwner, replacementOwner)
	}
	want := pickupAnkhTradableSuccessTrace4F3DD0()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupAnkhTradable4F3DD0LoadsItemAfterExtraLivesStore(t *testing.T) {
	w, owner, _ := pickupAnkhTradableSuccessBuild4F3DD0()
	replacement := &pickupAnkhTradableTestObject4F3DD0{name: "late-item"}
	w.afterStore = func(w *pickupAnkhTradableTestWorld4F3DD0) {
		w.item = replacement
	}
	if got := pickupAnkhTradable4F3DD0(owner, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events[4:6], []string{"item:late-item", "delete:late-item"}) {
		t.Fatalf("late item events = %v", w.events[4:6])
	}
}

func TestPickupAnkhTradable4F3DD0ForwardsNilItem(t *testing.T) {
	w, owner, _ := pickupAnkhTradableSuccessBuild4F3DD0()
	w.item = nil
	if got := pickupAnkhTradable4F3DD0(owner, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events[4:6], []string{"item:nil", "delete:nil"}) {
		t.Fatalf("nil item events = %v", w.events[4:6])
	}
}

func TestPickupAnkhTradable4F3DD0FaultPrefixes(t *testing.T) {
	want := pickupAnkhTradableSuccessTrace4F3DD0()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, _ := pickupAnkhTradableSuccessBuild4F3DD0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupAnkhTradable4F3DD0(owner, w.hooks())
		})
	}
}

func TestPickupAnkhTradable4F3DD0HasNoNilGuards(t *testing.T) {
	t.Run("owner", func(t *testing.T) {
		w := &pickupAnkhTradableTestWorld4F3DD0{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil owner did not preserve the class-load fault")
			}
			if len(w.events) != 0 {
				t.Fatalf("events = %v, want none", w.events)
			}
		}()
		pickupAnkhTradable4F3DD0[*pickupAnkhTradableTestObject4F3DD0, *pickupAnkhTradableTestUpdate4F3DD0](nil, w.hooks())
	})

	t.Run("update", func(t *testing.T) {
		owner := &pickupAnkhTradableTestObject4F3DD0{name: "owner", classLow: 4}
		w := &pickupAnkhTradableTestWorld4F3DD0{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not preserve the update dereference fault")
			}
			want := []string{"class:owner=04"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		}()
		pickupAnkhTradable4F3DD0(owner, w.hooks())
	})
}
