package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type inventoryPutTestPlayer4F3070 struct {
	name       string
	index      uint8
	protect    uint32
	overweight uint32
}

type inventoryPutTestUpdate4F3070 struct {
	name   string
	player *inventoryPutTestPlayer4F3070
}

type inventoryPutTestObject4F3070 struct {
	name     string
	flags    uint32
	class    uint32
	update   *inventoryPutTestUpdate4F3070
	previous *inventoryPutTestObject4F3070
	next     *inventoryPutTestObject4F3070
	first    *inventoryPutTestObject4F3070
	holder   *inventoryPutTestObject4F3070
	owner    *inventoryPutTestObject4F3070
	weight   uint8
	capacity uint16
}

type inventoryPutTestWorld4F3070 struct {
	owner *inventoryPutTestObject4F3070
	item  *inventoryPutTestObject4F3070

	events  []string
	faultAt int

	firstLoads    []*inventoryPutTestObject4F3070
	afterSetOwner func(*inventoryPutTestWorld4F3070)
	afterReport   func(*inventoryPutTestWorld4F3070)
	afterProtect  func(*inventoryPutTestWorld4F3070)
	afterStore    func(*inventoryPutTestWorld4F3070)
}

func inventoryPutObjectName4F3070(object *inventoryPutTestObject4F3070) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func inventoryPutUpdateName4F3070(update *inventoryPutTestUpdate4F3070) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func inventoryPutPlayerName4F3070(player *inventoryPutTestPlayer4F3070) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *inventoryPutTestWorld4F3070) event(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *inventoryPutTestWorld4F3070) hooks() inventoryPutHooks4F3070[
	*inventoryPutTestObject4F3070,
	*inventoryPutTestUpdate4F3070,
	*inventoryPutTestPlayer4F3070,
] {
	return inventoryPutHooks4F3070[
		*inventoryPutTestObject4F3070,
		*inventoryPutTestUpdate4F3070,
		*inventoryPutTestPlayer4F3070,
	]{
		loadFlagsLow: func(object *inventoryPutTestObject4F3070) uint8 {
			w.event("flags:%s", inventoryPutObjectName4F3070(object))
			return uint8(object.flags)
		},
		loadClassLow: func(object *inventoryPutTestObject4F3070) uint8 {
			w.event("class:%s", inventoryPutObjectName4F3070(object))
			return uint8(object.class)
		},
		loadUpdate: func(object *inventoryPutTestObject4F3070) *inventoryPutTestUpdate4F3070 {
			w.event("update:%s", inventoryPutObjectName4F3070(object))
			return object.update
		},
		loadPlayer: func(update *inventoryPutTestUpdate4F3070) *inventoryPutTestPlayer4F3070 {
			w.event("player:%s", inventoryPutUpdateName4F3070(update))
			return update.player
		},
		storeInventoryPrev: func(object, previous *inventoryPutTestObject4F3070) {
			w.event("store-previous:%s:%s", inventoryPutObjectName4F3070(object), inventoryPutObjectName4F3070(previous))
			object.previous = previous
		},
		loadInventoryFirst: func(owner *inventoryPutTestObject4F3070) *inventoryPutTestObject4F3070 {
			first := owner.first
			if len(w.firstLoads) != 0 {
				first = w.firstLoads[0]
				w.firstLoads = w.firstLoads[1:]
			}
			w.event("first:%s=%s", inventoryPutObjectName4F3070(owner), inventoryPutObjectName4F3070(first))
			return first
		},
		storeInventoryNext: func(object, next *inventoryPutTestObject4F3070) {
			w.event("store-next:%s:%s", inventoryPutObjectName4F3070(object), inventoryPutObjectName4F3070(next))
			object.next = next
		},
		storeInventoryFirst: func(owner, first *inventoryPutTestObject4F3070) {
			w.event("store-first:%s:%s", inventoryPutObjectName4F3070(owner), inventoryPutObjectName4F3070(first))
			owner.first = first
		},
		storeInventoryHolder: func(item, holder *inventoryPutTestObject4F3070) {
			w.event("store-holder:%s:%s", inventoryPutObjectName4F3070(item), inventoryPutObjectName4F3070(holder))
			item.holder = holder
		},
		setOwner: func(owner, item *inventoryPutTestObject4F3070) {
			w.event("set-owner:%s:%s", inventoryPutObjectName4F3070(owner), inventoryPutObjectName4F3070(item))
			item.owner = owner
			if w.afterSetOwner != nil {
				w.afterSetOwner(w)
			}
		},
		loadPlayerIndex: func(player *inventoryPutTestPlayer4F3070) uint8 {
			w.event("player-index:%s", inventoryPutPlayerName4F3070(player))
			return player.index
		},
		reportPickup: func(index uint8, item *inventoryPutTestObject4F3070) {
			w.event("report:%d:%s", index, inventoryPutObjectName4F3070(item))
			if w.afterReport != nil {
				w.afterReport(w)
			}
		},
		loadPlayerProtect: func(player *inventoryPutTestPlayer4F3070) uint32 {
			w.event("player-protect:%s", inventoryPutPlayerName4F3070(player))
			return player.protect
		},
		protectItem: func(protect uint32, item *inventoryPutTestObject4F3070) {
			w.event("protect:%d:%s", protect, inventoryPutObjectName4F3070(item))
			if w.afterProtect != nil {
				w.afterProtect(w)
			}
		},
		loadItemWeight: func(item *inventoryPutTestObject4F3070) uint8 {
			w.event("weight:%s", inventoryPutObjectName4F3070(item))
			return item.weight
		},
		loadInventoryNext: func(item *inventoryPutTestObject4F3070) *inventoryPutTestObject4F3070 {
			w.event("next:%s", inventoryPutObjectName4F3070(item))
			return item.next
		},
		loadCarryCapacity: func(owner *inventoryPutTestObject4F3070) uint16 {
			w.event("capacity:%s", inventoryPutObjectName4F3070(owner))
			return owner.capacity
		},
		storePlayerOverweight: func(player *inventoryPutTestPlayer4F3070, value uint32) {
			w.event("store-overweight:%s:%d", inventoryPutPlayerName4F3070(player), value)
			player.overweight = value
			if w.afterStore != nil {
				w.afterStore(w)
			}
		},
		audioEvent: func(id int32, owner *inventoryPutTestObject4F3070, kind int32, code uint32) {
			w.event("audio:%d:%s:%d:%d", id, inventoryPutObjectName4F3070(owner), kind, code)
		},
	}
}

func newInventoryPutFullTrace4F3070() (*inventoryPutTestWorld4F3070, []string) {
	entryUpdate := &inventoryPutTestUpdate4F3070{name: "entry-update"}
	player := &inventoryPutTestPlayer4F3070{name: "cached-player", index: 7, protect: 77}
	replacementPlayer := &inventoryPutTestPlayer4F3070{name: "replacement-player", protect: 1234}
	update := &inventoryPutTestUpdate4F3070{name: "live-update", player: player}
	owner := &inventoryPutTestObject4F3070{name: "owner", update: entryUpdate}
	item := &inventoryPutTestObject4F3070{name: "item"}
	staleHead := &inventoryPutTestObject4F3070{name: "stale-head"}
	reloadedHead := &inventoryPutTestObject4F3070{name: "reloaded-head"}
	weight2 := &inventoryPutTestObject4F3070{name: "weight-2", weight: 5}
	weight1 := &inventoryPutTestObject4F3070{name: "weight-1", weight: 250, next: weight2}
	w := &inventoryPutTestWorld4F3070{
		owner:      owner,
		item:       item,
		firstLoads: []*inventoryPutTestObject4F3070{staleHead, reloadedHead},
	}
	w.afterSetOwner = func(w *inventoryPutTestWorld4F3070) {
		w.owner.class = uint32(inventoryPutPlayerClass4F3070)
		w.owner.update = update
	}
	w.afterReport = func(w *inventoryPutTestWorld4F3070) {
		player.protect = 99
		update.player = replacementPlayer
	}
	w.afterProtect = func(w *inventoryPutTestWorld4F3070) {
		w.owner.first = weight1
		w.owner.capacity = 200
	}
	w.afterStore = func(w *inventoryPutTestWorld4F3070) {
		w.item.class = uint32(inventoryPutAudioClass4F3070)
	}
	want := []string{
		"flags:owner",
		"flags:item",
		"store-previous:item:nil",
		"first:owner=stale-head",
		"store-next:item:stale-head",
		"first:owner=reloaded-head",
		"store-previous:reloaded-head:item",
		"store-first:owner:item",
		"store-holder:item:owner",
		"set-owner:owner:item",
		"class:owner",
		"update:owner",
		"player:live-update",
		"player-index:cached-player",
		"report:7:item",
		"player-protect:cached-player",
		"protect:99:item",
		"first:owner=weight-1",
		"weight:weight-1",
		"next:weight-1",
		"weight:weight-2",
		"next:weight-2",
		"capacity:owner",
		"store-overweight:cached-player:1",
		"class:item",
		"audio:820:owner:0:0",
	}
	return w, want
}

func TestInventoryPut4F3070ExactTraceAndLiveReloads(t *testing.T) {
	w, want := newInventoryPutFullTrace4F3070()
	staleHead, reloadedHead := w.firstLoads[0], w.firstLoads[1]
	inventoryPut4F3070(w.owner, w.item, -1, w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
	if w.item.previous != nil || w.item.next != staleHead || reloadedHead.previous != w.item {
		t.Fatalf("inserted links = previous %p next %p reloaded-prev %p", w.item.previous, w.item.next, reloadedHead.previous)
	}
	if w.item.holder != w.owner || w.item.owner != w.owner {
		t.Fatalf("holder/owner = %p/%p, want %p", w.item.holder, w.item.owner, w.owner)
	}
	if got := w.owner.update.player; got == nil || got.name != "replacement-player" {
		t.Fatalf("live update player = %v, want replacement-player", got)
	}
	if got := w.owner.update; got == nil || got.name != "live-update" {
		t.Fatalf("owner update = %v, want live-update", got)
	}
	if got := w.item.class; uint8(got) != inventoryPutAudioClass4F3070 {
		t.Fatalf("live item class = %#x, want audio bit", got)
	}
}

func TestInventoryPut4F3070NilAndDestroyedGates(t *testing.T) {
	tests := []struct {
		name   string
		owner  *inventoryPutTestObject4F3070
		item   *inventoryPutTestObject4F3070
		events []string
	}{
		{name: "nil owner", item: &inventoryPutTestObject4F3070{name: "item"}},
		{name: "nil item", owner: &inventoryPutTestObject4F3070{name: "owner"}},
		{
			name:   "destroyed owner",
			owner:  &inventoryPutTestObject4F3070{name: "owner", flags: 0xabcdef20},
			item:   &inventoryPutTestObject4F3070{name: "item"},
			events: []string{"flags:owner"},
		},
		{
			name:   "destroyed item",
			owner:  &inventoryPutTestObject4F3070{name: "owner"},
			item:   &inventoryPutTestObject4F3070{name: "item", flags: 0x12340020},
			events: []string{"flags:owner", "flags:item"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := &inventoryPutTestWorld4F3070{owner: test.owner, item: test.item}
			inventoryPut4F3070(w.owner, w.item, 1, w.hooks())
			if !reflect.DeepEqual(w.events, test.events) {
				t.Fatalf("events = %v, want %v", w.events, test.events)
			}
		})
	}
}

func TestInventoryPut4F3070UsesLowBytesOnly(t *testing.T) {
	owner := &inventoryPutTestObject4F3070{name: "owner", flags: 0x20_00, class: 0x04_00}
	item := &inventoryPutTestObject4F3070{name: "item", flags: 0x20_00, class: 0x40_00}
	w := &inventoryPutTestWorld4F3070{owner: owner, item: item}
	inventoryPut4F3070(owner, item, 0, w.hooks())
	if owner.first != item || item.holder != owner {
		t.Fatalf("high-byte flags blocked insertion: first/holder = %p/%p", owner.first, item.holder)
	}
	wantTail := []string{"set-owner:owner:item", "class:owner", "class:item"}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail = %v, want %v", got, wantTail)
	}
}

func TestInventoryPut4F3070LoadsPlayerWhenReportDisabled(t *testing.T) {
	owner := &inventoryPutTestObject4F3070{name: "owner", class: uint32(inventoryPutPlayerClass4F3070)}
	item := &inventoryPutTestObject4F3070{name: "item"}
	w := &inventoryPutTestWorld4F3070{owner: owner, item: item}
	defer func() {
		if recover() == nil {
			t.Fatal("nil UpdateData did not fault with reporting disabled")
		}
		wantTail := []string{"class:owner", "update:owner", "player:nil"}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	}()
	inventoryPut4F3070(owner, item, 0, w.hooks())
}

func TestInventoryPut4F3070NilPlayerFaultDependsOnReport(t *testing.T) {
	for _, test := range []struct {
		name   string
		report int32
		last   string
	}{
		{name: "report", report: 1, last: "player-index:nil"},
		{name: "no report", report: 0, last: "player-protect:nil"},
	} {
		t.Run(test.name, func(t *testing.T) {
			update := &inventoryPutTestUpdate4F3070{name: "update"}
			owner := &inventoryPutTestObject4F3070{name: "owner", class: uint32(inventoryPutPlayerClass4F3070), update: update}
			item := &inventoryPutTestObject4F3070{name: "item"}
			w := &inventoryPutTestWorld4F3070{owner: owner, item: item}
			defer func() {
				if recover() == nil {
					t.Fatal("nil Player did not preserve the original fault")
				}
				if got := w.events[len(w.events)-1]; got != test.last {
					t.Fatalf("last event = %q, want %q", got, test.last)
				}
			}()
			inventoryPut4F3070(owner, item, test.report, w.hooks())
		})
	}
}

func TestInventoryPut4F3070SignedWrappingWeight(t *testing.T) {
	if got, want := addInventoryPutWeight4F3070(math.MaxInt32, 1), int32(math.MinInt32); got != want {
		t.Fatalf("wrapping add = %d, want %d", got, want)
	}
	if got := inventoryPutOverweight4F3070(math.MinInt32, 0); got != 0 {
		t.Fatalf("negative wrapped sum overweight = %d, want 0", got)
	}
	if got := inventoryPutOverweight4F3070(65536, math.MaxUint16); got != 1 {
		t.Fatalf("65536 versus uint16 max = %d, want 1", got)
	}
	if got := inventoryPutOverweight4F3070(math.MaxUint16, math.MaxUint16); got != 0 {
		t.Fatalf("equal capacity = %d, want 0", got)
	}
}

func TestInventoryPut4F3070EveryObservableFaultPrefix(t *testing.T) {
	_, want := newInventoryPutFullTrace4F3070()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w, _ := newInventoryPutFullTrace4F3070()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if expected := want[:faultAt]; !reflect.DeepEqual(w.events, expected) {
					t.Fatalf("events = %v, want %v", w.events, expected)
				}
			}()
			inventoryPut4F3070(w.owner, w.item, -1, w.hooks())
		})
	}
}
