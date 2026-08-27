package server

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/player"
)

type secondaryWeaponTestObject53AB90 struct {
	name   string
	update *secondaryWeaponTestUpdate53AB90
}

type secondaryWeaponTestUpdate53AB90 struct {
	player *secondaryWeaponTestPlayer53AB90
	stored *secondaryWeaponTestObject53AB90
}

type secondaryWeaponTestPlayer53AB90 struct {
	name  string
	class player.Class
	index byte
}

func TestSecondaryWeaponReport53AB90OrderAndShortCircuit(t *testing.T) {
	owner := &secondaryWeaponTestObject53AB90{name: "owner"}
	item := &secondaryWeaponTestObject53AB90{name: "item"}
	firstPlayer := &secondaryWeaponTestPlayer53AB90{name: "first", class: player.Warrior, index: 3}
	secondPlayer := &secondaryWeaponTestPlayer53AB90{name: "second", class: player.Wizard, index: 7}
	update := &secondaryWeaponTestUpdate53AB90{player: firstPlayer}
	owner.update = update

	tests := []struct {
		name          string
		item          *secondaryWeaponTestObject53AB90
		classAllowed  bool
		strongEnough  bool
		mutatePlayer  bool
		want          []string
		wantClear     byte
		wantClearSeen bool
	}{
		{name: "nil item", want: []string{"update:owner", "store:<nil>"}},
		{
			name: "class rejected reloads player and skips strength", item: item,
			want:         []string{"update:owner", "player:first", "class:first", "class-use:item:0", "player:second", "index:second", "clear:7", "store:item"},
			mutatePlayer: true, wantClear: 7, wantClearSeen: true,
		},
		{
			name: "strength rejected", item: item, classAllowed: true,
			want:      []string{"update:owner", "player:first", "class:first", "class-use:item:0", "strength:owner:item", "player:first", "index:first", "clear:3", "store:item"},
			wantClear: 3, wantClearSeen: true,
		},
		{
			name: "valid", item: item, classAllowed: true, strongEnough: true,
			want: []string{"update:owner", "player:first", "class:first", "class-use:item:0", "strength:owner:item", "store:item"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			update.player = firstPlayer
			update.stored = nil
			var events []string
			var clearIndex byte
			clearSeen := false
			secondaryWeaponReport53AB90(owner, tc.item, secondaryWeaponHooks53AB90[
				*secondaryWeaponTestObject53AB90,
				*secondaryWeaponTestUpdate53AB90,
				*secondaryWeaponTestPlayer53AB90,
			]{
				loadUpdate: func(object *secondaryWeaponTestObject53AB90) *secondaryWeaponTestUpdate53AB90 {
					events = append(events, "update:"+object.name)
					return object.update
				},
				loadPlayer: func(got *secondaryWeaponTestUpdate53AB90) *secondaryWeaponTestPlayer53AB90 {
					events = append(events, "player:"+got.player.name)
					return got.player
				},
				loadPlayerClass: func(got *secondaryWeaponTestPlayer53AB90) player.Class {
					events = append(events, "class:"+got.name)
					return got.class
				},
				classCanUseItem: func(got *secondaryWeaponTestObject53AB90, class player.Class) bool {
					events = append(events, "class-use:"+got.name+":"+string(rune(class+'0')))
					if tc.mutatePlayer {
						update.player = secondPlayer
					}
					return tc.classAllowed
				},
				checkStrength: func(gotOwner, gotItem *secondaryWeaponTestObject53AB90) bool {
					events = append(events, "strength:"+gotOwner.name+":"+gotItem.name)
					return tc.strongEnough
				},
				loadPlayerIndex: func(got *secondaryWeaponTestPlayer53AB90) byte {
					events = append(events, "index:"+got.name)
					return got.index
				},
				clearClient: func(index byte) {
					events = append(events, "clear:"+string(rune(index+'0')))
					clearIndex, clearSeen = index, true
				},
				store: func(got *secondaryWeaponTestUpdate53AB90, stored *secondaryWeaponTestObject53AB90) {
					name := "<nil>"
					if stored != nil {
						name = stored.name
					}
					events = append(events, "store:"+name)
					got.stored = stored
				},
			})
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %#v, want %#v", events, tc.want)
			}
			if clearSeen != tc.wantClearSeen || clearIndex != tc.wantClear {
				t.Fatalf("clear = (%t,%d), want (%t,%d)", clearSeen, clearIndex, tc.wantClearSeen, tc.wantClear)
			}
			if update.stored != tc.item {
				t.Fatalf("stored = %p, want %p", update.stored, tc.item)
			}
		})
	}
}

func TestSecondaryWeaponReport53AB90NilOwnerReadsNothing(t *testing.T) {
	secondaryWeaponReport53AB90[*secondaryWeaponTestObject53AB90, int, int](nil, new(secondaryWeaponTestObject53AB90), secondaryWeaponHooks53AB90[*secondaryWeaponTestObject53AB90, int, int]{
		loadUpdate: func(*secondaryWeaponTestObject53AB90) int {
			t.Fatal("nil owner read update")
			return 0
		},
	})
}

func TestSecondaryWeaponReport53AB90CachesUpdateAcrossCallbacks(t *testing.T) {
	oldUpdate := &secondaryWeaponTestUpdate53AB90{player: &secondaryWeaponTestPlayer53AB90{class: player.Warrior}}
	newUpdate := &secondaryWeaponTestUpdate53AB90{}
	owner := &secondaryWeaponTestObject53AB90{name: "owner", update: oldUpdate}
	item := &secondaryWeaponTestObject53AB90{name: "item"}
	secondaryWeaponReport53AB90(owner, item, secondaryWeaponHooks53AB90[
		*secondaryWeaponTestObject53AB90,
		*secondaryWeaponTestUpdate53AB90,
		*secondaryWeaponTestPlayer53AB90,
	]{
		loadUpdate:      func(object *secondaryWeaponTestObject53AB90) *secondaryWeaponTestUpdate53AB90 { return object.update },
		loadPlayer:      func(update *secondaryWeaponTestUpdate53AB90) *secondaryWeaponTestPlayer53AB90 { return update.player },
		loadPlayerClass: func(got *secondaryWeaponTestPlayer53AB90) player.Class { return got.class },
		classCanUseItem: func(*secondaryWeaponTestObject53AB90, player.Class) bool {
			owner.update = newUpdate
			return true
		},
		checkStrength: func(*secondaryWeaponTestObject53AB90, *secondaryWeaponTestObject53AB90) bool { return true },
		store: func(update *secondaryWeaponTestUpdate53AB90, stored *secondaryWeaponTestObject53AB90) {
			update.stored = stored
		},
	})
	if oldUpdate.stored != item || newUpdate.stored != nil {
		t.Fatalf("cached/new stores = %p/%p, want %p/nil", oldUpdate.stored, newUpdate.stored, item)
	}
}
