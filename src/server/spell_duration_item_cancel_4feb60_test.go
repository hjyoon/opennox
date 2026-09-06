package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	itemCancelDurSpellsItem4FEB60         = uint64(0x100000123)
	itemCancelDurSpellsOwner4FEB60        = uint64(0x200000456)
	itemCancelDurSpellsSameLowOwner4FEB60 = uint64(0x300000456)
)

type itemCancelDurSpellsItemState4FEB60 struct {
	class    uint32
	subclass uint32
}

type itemCancelDurSpellsWorld4FEB60 struct {
	item  uint64
	owner uint64
	items map[uint64]*itemCancelDurSpellsItemState4FEB60

	events    []string
	cancelled []struct {
		spell int32
		owner uint64
	}
	faultAt int
	after   map[string]func()
}

func newItemCancelDurSpellsWorld4FEB60() *itemCancelDurSpellsWorld4FEB60 {
	return &itemCancelDurSpellsWorld4FEB60{
		item:  itemCancelDurSpellsItem4FEB60,
		owner: itemCancelDurSpellsOwner4FEB60,
		items: map[uint64]*itemCancelDurSpellsItemState4FEB60{
			itemCancelDurSpellsItem4FEB60: {
				class: itemCancelDurSpellsClass4FEB60,
				subclass: itemCancelDurSpellsSpell43Bit4FEB60 |
					itemCancelDurSpellsSpell59Bit4FEB60,
			},
		},
		after: make(map[string]func()),
	}
}

func (w *itemCancelDurSpellsWorld4FEB60) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *itemCancelDurSpellsWorld4FEB60) hooks() ItemCancelDurSpellsHooks4FEB60[uint64, uint64] {
	return ItemCancelDurSpellsHooks4FEB60[uint64, uint64]{
		LoadItemArg: func() uint64 {
			item := w.item
			w.observe(fmt.Sprintf("item:%x", item))
			return item
		},
		LoadClass: func(item uint64) uint32 {
			class := w.items[item].class
			w.observe(fmt.Sprintf("class:%x:%x", item, class))
			return class
		},
		LoadSubclass: func(item uint64) uint32 {
			subclass := w.items[item].subclass
			w.observe(fmt.Sprintf("subclass:%x:%x", item, subclass))
			return subclass
		},
		LoadOwnerArg: func() uint64 {
			owner := w.owner
			w.observe(fmt.Sprintf("owner:%x", owner))
			return owner
		},
		Cancel: func(spell int32, owner uint64) {
			w.cancelled = append(w.cancelled, struct {
				spell int32
				owner uint64
			}{spell: spell, owner: owner})
			w.observe(fmt.Sprintf("cancel:%d:%x", spell, owner))
		},
	}
}

func TestItemCancelDurSpells4FEB60LiveSubclassReloadAndCachedOwner(t *testing.T) {
	w := newItemCancelDurSpellsWorld4FEB60()
	w.items[w.item].subclass = itemCancelDurSpellsSpell43Bit4FEB60
	w.after["cancel:43:200000456"] = func() {
		w.items[w.item].subclass |= itemCancelDurSpellsSpell59Bit4FEB60
		w.owner = itemCancelDurSpellsSameLowOwner4FEB60
	}

	ItemCancelDurSpells4FEB60(w.hooks())

	wantEvents := []string{
		"item:100000123",
		"class:100000123:1000",
		"subclass:100000123:40000",
		"owner:200000456",
		"cancel:43:200000456",
		"subclass:100000123:4040000",
		"cancel:59:200000456",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %q, want %q", w.events, wantEvents)
	}
	wantCancelled := []struct {
		spell int32
		owner uint64
	}{
		{spell: 43, owner: itemCancelDurSpellsOwner4FEB60},
		{spell: 59, owner: itemCancelDurSpellsOwner4FEB60},
	}
	if !reflect.DeepEqual(w.cancelled, wantCancelled) {
		t.Fatalf("cancelled = %#v, want %#v", w.cancelled, wantCancelled)
	}
}

func TestItemCancelDurSpells4FEB60FirstCallbackCanSuppressSecond(t *testing.T) {
	w := newItemCancelDurSpellsWorld4FEB60()
	w.after["cancel:43:200000456"] = func() {
		w.items[w.item].subclass &^= itemCancelDurSpellsSpell59Bit4FEB60
	}

	ItemCancelDurSpells4FEB60(w.hooks())

	if len(w.cancelled) != 1 || w.cancelled[0].spell != 43 {
		t.Fatalf("cancelled = %#v, want only spell 43", w.cancelled)
	}
	wantTail := []string{"cancel:43:200000456", "subclass:100000123:40000"}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}

func TestItemCancelDurSpells4FEB60ClassAndSubclassGates(t *testing.T) {
	tests := []struct {
		name       string
		class      uint32
		subclass   uint32
		wantEvents []string
		wantSpells []int32
	}{
		{
			name:       "wrong-class",
			class:      ^itemCancelDurSpellsClass4FEB60,
			subclass:   itemCancelDurSpellsSpell43Bit4FEB60 | itemCancelDurSpellsSpell59Bit4FEB60,
			wantEvents: []string{"item:100000123", "class:100000123:ffffefff"},
		},
		{
			name:       "no-spell-bits",
			class:      itemCancelDurSpellsClass4FEB60,
			wantEvents: []string{"item:100000123", "class:100000123:1000", "subclass:100000123:0", "owner:200000456", "subclass:100000123:0"},
		},
		{
			name:       "spell-43-only",
			class:      itemCancelDurSpellsClass4FEB60,
			subclass:   itemCancelDurSpellsSpell43Bit4FEB60,
			wantEvents: []string{"item:100000123", "class:100000123:1000", "subclass:100000123:40000", "owner:200000456", "cancel:43:200000456", "subclass:100000123:40000"},
			wantSpells: []int32{43},
		},
		{
			name:       "spell-59-only",
			class:      itemCancelDurSpellsClass4FEB60,
			subclass:   itemCancelDurSpellsSpell59Bit4FEB60,
			wantEvents: []string{"item:100000123", "class:100000123:1000", "subclass:100000123:4000000", "owner:200000456", "subclass:100000123:4000000", "cancel:59:200000456"},
			wantSpells: []int32{59},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newItemCancelDurSpellsWorld4FEB60()
			w.items[w.item].class = tc.class
			w.items[w.item].subclass = tc.subclass

			ItemCancelDurSpells4FEB60(w.hooks())

			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, tc.wantEvents)
			}
			var gotSpells []int32
			for _, call := range w.cancelled {
				gotSpells = append(gotSpells, call.spell)
			}
			if !reflect.DeepEqual(gotSpells, tc.wantSpells) {
				t.Fatalf("spells = %v, want %v", gotSpells, tc.wantSpells)
			}
		})
	}
}

func TestItemCancelDurSpells4FEB60NilOwnerIsForwarded(t *testing.T) {
	w := newItemCancelDurSpellsWorld4FEB60()
	w.owner = 0
	w.items[w.item].subclass = itemCancelDurSpellsSpell43Bit4FEB60

	ItemCancelDurSpells4FEB60(w.hooks())

	if len(w.cancelled) != 1 || w.cancelled[0].owner != 0 {
		t.Fatalf("cancelled = %#v, want spell cancellation with nil owner", w.cancelled)
	}
}

func TestItemCancelDurSpells4FEB60DoesNotGuardNilItem(t *testing.T) {
	w := newItemCancelDurSpellsWorld4FEB60()
	w.item = 0
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		ItemCancelDurSpells4FEB60(w.hooks())
	}()
	if recovered == nil {
		t.Fatal("nil item did not fault at the class load")
	}
	if want := []string{"item:0"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want fault prefix %q", w.events, want)
	}
}

func TestItemCancelDurSpells4FEB60FaultPrefixes(t *testing.T) {
	baseline := newItemCancelDurSpellsWorld4FEB60()
	ItemCancelDurSpells4FEB60(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newItemCancelDurSpellsWorld4FEB60()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				ItemCancelDurSpells4FEB60(w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}
