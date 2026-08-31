package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerAbilityCooldownIndexedSetTestWorld4FC4A0 struct {
	events      []string
	faultAt     int
	playerIndex int32
	ability     Ability
	cooldown    int32
	stores      map[int32]int32
}

func (w *playerAbilityCooldownIndexedSetTestWorld4FC4A0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerAbilityCooldownIndexedSetTestWorld4FC4A0) hooks() playerAbilityCooldownIndexedSetHooks4FC4A0 {
	return playerAbilityCooldownIndexedSetHooks4FC4A0{
		loadPlayerIndexArg: func() int32 {
			value := w.playerIndex
			w.record(fmt.Sprintf("player-index:%d", value))
			return value
		},
		loadAbilityArg: func() Ability {
			value := w.ability
			w.record(fmt.Sprintf("ability:%d", value))
			return value
		},
		loadCooldownArg: func() int32 {
			value := w.cooldown
			w.record(fmt.Sprintf("cooldown:%d", value))
			return value
		},
		storeCooldown: func(flatIndex int32, cooldown int32) {
			w.record(fmt.Sprintf("store:%d=%d", flatIndex, cooldown))
			w.stores[flatIndex] = cooldown
		},
	}
}

func TestPlayerAbilityCooldownIndexedSet4FC4A0OrderWidthsWrapAndReturn(t *testing.T) {
	w := playerAbilityCooldownIndexedSetTestWorld4FC4A0{
		playerIndex: math.MaxInt32,
		ability:     AbilityInfravis,
		cooldown:    math.MinInt32,
		stores:      make(map[int32]int32),
	}

	if got := playerAbilityCooldownIndexedSet4FC4A0(w.hooks()); got != math.MinInt32 {
		t.Fatalf("return = %#08x, want %#08x", uint32(got), uint32(1)<<31)
	}
	if got := w.stores[-1]; got != math.MinInt32 {
		t.Fatalf("stored cooldown = %#08x, want %#08x", uint32(got), uint32(1)<<31)
	}
	want := []string{
		fmt.Sprintf("player-index:%d", int32(math.MaxInt32)),
		fmt.Sprintf("ability:%d", AbilityInfravis),
		fmt.Sprintf("cooldown:%d", int32(math.MinInt32)),
		fmt.Sprintf("store:-1=%d", int32(math.MinInt32)),
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestPlayerAbilityCooldownIndexedSet4FC4A0AllPlayerAbilitySlots(t *testing.T) {
	w := playerAbilityCooldownIndexedSetTestWorld4FC4A0{stores: make(map[int32]int32)}
	hooks := w.hooks()
	for playerIndex := int32(0); playerIndex < abilityRuntimePlayerSlots4FB990; playerIndex++ {
		for ability := AbilityInvalid; ability < AbilityMax; ability++ {
			wantIndex := playerIndex*int32(AbilityMax) + int32(ability)
			wantValue := int32(0x40000000) + wantIndex
			w.playerIndex = playerIndex
			w.ability = ability
			w.cooldown = wantValue
			if got := playerAbilityCooldownIndexedSet4FC4A0(hooks); got != wantValue {
				t.Fatalf("player %d ability %d return = %d, want %d", playerIndex, ability, got, wantValue)
			}
			if got := w.stores[wantIndex]; got != wantValue {
				t.Fatalf("flat index %d stored = %d, want %d", wantIndex, got, wantValue)
			}
		}
	}
	if got, want := len(w.stores), abilityRuntimePlayerSlots4FB990*int(AbilityMax); got != want {
		t.Fatalf("stored words = %d, want %d", got, want)
	}
}

func TestPlayerAbilityCooldownIndexedSet4FC4A0PE32IndexArithmetic(t *testing.T) {
	tests := []struct {
		name        string
		playerIndex int32
		ability     Ability
		want        int32
	}{
		{name: "first", playerIndex: 0, ability: AbilityInvalid, want: 0},
		{name: "last", playerIndex: 31, ability: AbilityInfravis, want: 191},
		{name: "negative-player", playerIndex: -1, ability: AbilityInvalid, want: -6},
		{name: "negative-ability", playerIndex: 1, ability: Ability(-7), want: -1},
		{name: "max-wrap", playerIndex: math.MaxInt32, ability: AbilityInfravis, want: -1},
		{name: "min-player-wrap", playerIndex: math.MinInt32, ability: AbilityInvalid, want: 0},
		{name: "min-ability", playerIndex: 0, ability: Ability(math.MinInt32), want: math.MinInt32},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := playerAbilityCooldownIndexedSetTestWorld4FC4A0{
				playerIndex: test.playerIndex,
				ability:     test.ability,
				cooldown:    123,
				stores:      make(map[int32]int32),
			}
			if got := playerAbilityCooldownIndexedSet4FC4A0(w.hooks()); got != 123 {
				t.Fatalf("return = %d, want 123", got)
			}
			if got := w.stores[test.want]; got != 123 {
				t.Fatalf("flat index %d stored = %d, want 123; stores = %#v", test.want, got, w.stores)
			}
		})
	}
}

func TestPlayerAbilityCooldownIndexedSet4FC4A0FaultPrefixes(t *testing.T) {
	for faultAt, want := range [][]string{
		{"player-index:7"},
		{"player-index:7", "ability:-2"},
		{"player-index:7", "ability:-2", "cooldown:99"},
		{"player-index:7", "ability:-2", "cooldown:99", "store:40=99"},
	} {
		t.Run(fmt.Sprintf("fault-%d", faultAt+1), func(t *testing.T) {
			w := playerAbilityCooldownIndexedSetTestWorld4FC4A0{
				faultAt:     faultAt + 1,
				playerIndex: 7,
				ability:     Ability(-2),
				cooldown:    99,
				stores:      make(map[int32]int32),
			}
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %#v, want %#v", w.events, want)
				}
				if len(w.stores) != 0 {
					t.Fatalf("stores = %#v, want no completed store", w.stores)
				}
			}()
			playerAbilityCooldownIndexedSet4FC4A0(w.hooks())
		})
	}
}
