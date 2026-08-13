package server

import (
	"fmt"
	"reflect"
	"testing"
)

type monsterMarkUpdateObject4E8020 struct {
	name             string
	field35, field36 uint32
}

type monsterMarkUpdatePlayer4E8020 struct {
	name string
	ind  uint8
	unit *monsterMarkUpdateObject4E8020
	next *monsterMarkUpdatePlayer4E8020
}

func monsterMarkUpdateTestHooks4E8020(
	events *[]string,
	first **monsterMarkUpdatePlayer4E8020,
	hostile func(*monsterMarkUpdateObject4E8020, *monsterMarkUpdateObject4E8020) int32,
) monsterMarkUpdateHooks4E8020[*monsterMarkUpdateObject4E8020, *monsterMarkUpdatePlayer4E8020] {
	return monsterMarkUpdateHooks4E8020[*monsterMarkUpdateObject4E8020, *monsterMarkUpdatePlayer4E8020]{
		firstPlayer: func() *monsterMarkUpdatePlayer4E8020 {
			*events = append(*events, "first")
			return *first
		},
		loadPlayerInd: func(player *monsterMarkUpdatePlayer4E8020) uint8 {
			*events = append(*events, "index:"+player.name)
			return player.ind
		},
		loadPlayerUnit: func(player *monsterMarkUpdatePlayer4E8020) *monsterMarkUpdateObject4E8020 {
			*events = append(*events, "unit:"+player.name)
			return player.unit
		},
		isHostile: func(unit, obj *monsterMarkUpdateObject4E8020) int32 {
			*events = append(*events, "hostile:"+unit.name)
			return hostile(unit, obj)
		},
		loadField36: func(obj *monsterMarkUpdateObject4E8020) uint32 {
			*events = append(*events, "load36")
			return obj.field36
		},
		loadField35: func(obj *monsterMarkUpdateObject4E8020) uint32 {
			*events = append(*events, "load35")
			return obj.field35
		},
		storeField36: func(obj *monsterMarkUpdateObject4E8020, value uint32) {
			*events = append(*events, fmt.Sprintf("store36:%#x", value))
			obj.field36 = value
		},
		storeField35: func(obj *monsterMarkUpdateObject4E8020, value uint32) {
			*events = append(*events, fmt.Sprintf("store35:%#x", value))
			obj.field35 = value
		},
		nextPlayer: func(player *monsterMarkUpdatePlayer4E8020) *monsterMarkUpdatePlayer4E8020 {
			*events = append(*events, "next:"+player.name)
			return player.next
		},
	}
}

func TestMonsterMarkUpdate4E8020EmptyListDoesNotTouchObject(t *testing.T) {
	var first *monsterMarkUpdatePlayer4E8020
	var events []string
	monsterMarkUpdate4E8020((*monsterMarkUpdateObject4E8020)(nil), monsterMarkUpdateTestHooks4E8020(
		&events, &first, func(*monsterMarkUpdateObject4E8020, *monsterMarkUpdateObject4E8020) int32 {
			t.Fatal("empty list reached hostility callback")
			return 0
		},
	))
	if want := []string{"first"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMonsterMarkUpdate4E8020NilUnitOrderMaskedShiftAndCachedLoads(t *testing.T) {
	obj := &monsterMarkUpdateObject4E8020{name: "obj", field35: 3, field36: 3}
	player := &monsterMarkUpdatePlayer4E8020{name: "player", ind: 32}
	first := player
	var events []string
	hooks := monsterMarkUpdateTestHooks4E8020(&events, &first,
		func(*monsterMarkUpdateObject4E8020, *monsterMarkUpdateObject4E8020) int32 {
			t.Fatal("nil unit reached hostility callback")
			return 0
		},
	)
	baseStore36 := hooks.storeField36
	hooks.storeField36 = func(obj *monsterMarkUpdateObject4E8020, value uint32) {
		baseStore36(obj, value)
		obj.field35 = 0xffffffff
	}

	monsterMarkUpdate4E8020(obj, hooks)
	if obj.field36 != 2 || obj.field35 != 2 {
		t.Fatalf("masks = (%#x, %#x), want (2, 2)", obj.field35, obj.field36)
	}
	want := []string{
		"first", "index:player", "unit:player", "load36", "load35",
		"store36:0x2", "store35:0x2", "next:player",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMonsterMarkUpdate4E8020ExactOneAndConditionalField35(t *testing.T) {
	unit := &monsterMarkUpdateObject4E8020{name: "unit"}
	tests := []struct {
		name        string
		hostile     int32
		field36     uint32
		wantField35 uint32
		wantField36 uint32
		wantMiddle  []string
	}{
		{name: "negative with absent bit", hostile: -1, wantField35: 0x10, wantField36: 0, wantMiddle: []string{"hostile:unit", "load36"}},
		{name: "zero with present bit", hostile: 0, field36: 8, wantField35: 0x18, wantField36: 0, wantMiddle: []string{"hostile:unit", "load36", "store36:0x0", "load35", "store35:0x18"}},
		{name: "exact one with absent bit", hostile: 1, wantField35: 0x18, wantField36: 8, wantMiddle: []string{"hostile:unit", "load36", "store36:0x8", "load35", "store35:0x18"}},
		{name: "exact one with present bit", hostile: 1, field36: 8, wantField35: 0x10, wantField36: 8, wantMiddle: []string{"hostile:unit", "load36"}},
		{name: "two with absent bit", hostile: 2, wantField35: 0x10, wantField36: 0, wantMiddle: []string{"hostile:unit", "load36"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := &monsterMarkUpdateObject4E8020{name: "obj", field35: 0x10, field36: tc.field36}
			player := &monsterMarkUpdatePlayer4E8020{name: "player", ind: 3, unit: unit}
			first := player
			var events []string
			monsterMarkUpdate4E8020(obj, monsterMarkUpdateTestHooks4E8020(
				&events, &first, func(*monsterMarkUpdateObject4E8020, *monsterMarkUpdateObject4E8020) int32 {
					return tc.hostile
				},
			))
			if obj.field35 != tc.wantField35 || obj.field36 != tc.wantField36 {
				t.Fatalf("masks = (%#x, %#x), want (%#x, %#x)", obj.field35, obj.field36, tc.wantField35, tc.wantField36)
			}
			want := append([]string{"first", "index:player", "unit:player"}, tc.wantMiddle...)
			want = append(want, "next:player")
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestMonsterMarkUpdate4E8020UsesLiveCallbackStateAndSuccessor(t *testing.T) {
	obj := &monsterMarkUpdateObject4E8020{name: "obj", field35: 0x10}
	unit := &monsterMarkUpdateObject4E8020{name: "unit"}
	stale := &monsterMarkUpdatePlayer4E8020{name: "stale"}
	replacement := &monsterMarkUpdatePlayer4E8020{name: "replacement", ind: 4}
	player := &monsterMarkUpdatePlayer4E8020{name: "player", ind: 3, unit: unit, next: stale}
	first := player
	var events []string
	hooks := monsterMarkUpdateTestHooks4E8020(&events, &first,
		func(_ *monsterMarkUpdateObject4E8020, marked *monsterMarkUpdateObject4E8020) int32 {
			marked.field36 = 8
			player.next = replacement
			return 0
		},
	)
	baseStore36 := hooks.storeField36
	hooks.storeField36 = func(obj *monsterMarkUpdateObject4E8020, value uint32) {
		baseStore36(obj, value)
		obj.field35 = 0x40
	}

	monsterMarkUpdate4E8020(obj, hooks)
	if obj.field36 != 0 || obj.field35 != 0x48 {
		t.Fatalf("masks = (%#x, %#x), want (0x48, 0)", obj.field35, obj.field36)
	}
	if containsMonsterMarkEvent4E8020(events, "index:stale") || !containsMonsterMarkEvent4E8020(events, "index:replacement") {
		t.Fatalf("successor events = %v, want live replacement only", events)
	}
	wantPrefix := []string{
		"first", "index:player", "unit:player", "hostile:unit", "load36",
		"store36:0x0", "load35", "store35:0x48", "next:player",
	}
	if len(events) < len(wantPrefix) || !reflect.DeepEqual(events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("event prefix = %v, want %v", events, wantPrefix)
	}
}

func TestMonsterMarkUpdate4E8020NilObjectFaultsAfterPlayerFields(t *testing.T) {
	player := &monsterMarkUpdatePlayer4E8020{name: "player", ind: 7}
	first := player
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nonempty list with nil object returned without a panic")
		}
		want := []string{"first", "index:player", "unit:player", "load36"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	monsterMarkUpdate4E8020((*monsterMarkUpdateObject4E8020)(nil), monsterMarkUpdateTestHooks4E8020(
		&events, &first, func(*monsterMarkUpdateObject4E8020, *monsterMarkUpdateObject4E8020) int32 { return 0 },
	))
}

func containsMonsterMarkEvent4E8020(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
