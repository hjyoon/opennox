package server

import (
	"math"
	"reflect"
	"testing"
)

type playerCheckStrengthTestObject4F3180 struct {
	class   uint32
	typeInd uint16
}

type playerCheckStrengthTestDef4F3180 struct {
	required uint16
}

func TestPlayerCheckStrength4F3180NonPlayerStopsBeforeStrengthAndItem(t *testing.T) {
	player := &playerCheckStrengthTestObject4F3180{class: 0x00000400}
	var events []string
	got := playerCheckStrength4F3180(player, (*playerCheckStrengthTestObject4F3180)(nil), playerCheckStrengthHooks4F3180[
		*playerCheckStrengthTestObject4F3180,
		*playerCheckStrengthTestDef4F3180,
	]{
		loadPlayerClassLow: func(object *playerCheckStrengthTestObject4F3180) uint8 {
			events = append(events, "player-class")
			return uint8(object.class)
		},
		getUnitStrength: func(*playerCheckStrengthTestObject4F3180) int32 {
			t.Fatal("strength reached for non-Player")
			return 0
		},
		loadItemClass: func(*playerCheckStrengthTestObject4F3180) uint32 {
			t.Fatal("item class reached for non-Player")
			return 0
		},
	})
	if got != 0 || !reflect.DeepEqual(events, []string{"player-class"}) {
		t.Fatalf("result/events = %d/%v, want 0/[player-class]", got, events)
	}
}

func TestPlayerCheckStrength4F3180UsesPostStrengthLiveArmorState(t *testing.T) {
	player := &playerCheckStrengthTestObject4F3180{class: 0x80000004}
	item := &playerCheckStrengthTestObject4F3180{class: 0x1000, typeInd: 7}
	def := &playerCheckStrengthTestDef4F3180{required: 99}
	var events []string
	got := playerCheckStrength4F3180(player, item, playerCheckStrengthHooks4F3180[
		*playerCheckStrengthTestObject4F3180,
		*playerCheckStrengthTestDef4F3180,
	]{
		loadPlayerClassLow: func(object *playerCheckStrengthTestObject4F3180) uint8 {
			events = append(events, "player-class")
			return uint8(object.class)
		},
		getUnitStrength: func(gotPlayer *playerCheckStrengthTestObject4F3180) int32 {
			events = append(events, "strength")
			if gotPlayer != player {
				t.Fatalf("strength player = %p, want %p", gotPlayer, player)
			}
			item.class = playerCheckStrengthArmorClass4F3180
			item.typeInd = 0xbeef
			return 42
		},
		loadItemClass: func(gotItem *playerCheckStrengthTestObject4F3180) uint32 {
			events = append(events, "item-class")
			return gotItem.class
		},
		loadItemType: func(gotItem *playerCheckStrengthTestObject4F3180) uint16 {
			events = append(events, "item-type")
			return gotItem.typeInd
		},
		findArmorDef: func(typeInd uint16) *playerCheckStrengthTestDef4F3180 {
			events = append(events, "armor-def")
			if typeInd != 0xbeef {
				t.Fatalf("armor type = %#x, want 0xbeef", typeInd)
			}
			def.required = 42
			return def
		},
		findWeaponDef: func(uint16) *playerCheckStrengthTestDef4F3180 {
			t.Fatal("weapon lookup reached for live armor")
			return nil
		},
		loadRequired: func(gotDef *playerCheckStrengthTestDef4F3180) uint16 {
			events = append(events, "required")
			if gotDef != def {
				t.Fatalf("definition = %p, want %p", gotDef, def)
			}
			return gotDef.required
		},
	})
	wantEvents := []string{"player-class", "strength", "item-class", "item-type", "armor-def", "required"}
	if got != 1 || !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("result/events = %d/%v, want 1/%v", got, events, wantEvents)
	}
}

func TestPlayerCheckStrength4F3180WeaponDefinitionAndNilShortCircuit(t *testing.T) {
	player := &playerCheckStrengthTestObject4F3180{class: 4}
	item := &playerCheckStrengthTestObject4F3180{class: 0xffffffff &^ playerCheckStrengthArmorClass4F3180, typeInd: 0xffff}
	var events []string
	got := playerCheckStrength4F3180(player, item, playerCheckStrengthHooks4F3180[
		*playerCheckStrengthTestObject4F3180,
		*playerCheckStrengthTestDef4F3180,
	]{
		loadPlayerClassLow: func(object *playerCheckStrengthTestObject4F3180) uint8 {
			events = append(events, "player-class")
			return uint8(object.class)
		},
		getUnitStrength: func(*playerCheckStrengthTestObject4F3180) int32 {
			events = append(events, "strength")
			return math.MaxInt32
		},
		loadItemClass: func(object *playerCheckStrengthTestObject4F3180) uint32 {
			events = append(events, "item-class")
			return object.class
		},
		loadItemType: func(object *playerCheckStrengthTestObject4F3180) uint16 {
			events = append(events, "item-type")
			return object.typeInd
		},
		findArmorDef: func(uint16) *playerCheckStrengthTestDef4F3180 {
			t.Fatal("armor lookup reached for weapon class")
			return nil
		},
		findWeaponDef: func(typeInd uint16) *playerCheckStrengthTestDef4F3180 {
			events = append(events, "weapon-def")
			if typeInd != math.MaxUint16 {
				t.Fatalf("weapon type = %#x, want 0xffff", typeInd)
			}
			return nil
		},
		loadRequired: func(*playerCheckStrengthTestDef4F3180) uint16 {
			t.Fatal("required strength read after nil definition")
			return 0
		},
	})
	wantEvents := []string{"player-class", "strength", "item-class", "item-type", "weapon-def"}
	if got != 0 || !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("result/events = %d/%v, want 0/%v", got, events, wantEvents)
	}
}

func TestPlayerCheckStrength4F3180SignedComparison(t *testing.T) {
	tests := []struct {
		name     string
		strength int32
		required uint16
		want     int32
	}{
		{name: "negative versus zero", strength: -1, required: 0, want: 0},
		{name: "zero equality", strength: 0, required: 0, want: 1},
		{name: "below uint16 maximum", strength: math.MaxUint16 - 1, required: math.MaxUint16, want: 0},
		{name: "uint16 maximum equality", strength: math.MaxUint16, required: math.MaxUint16, want: 1},
		{name: "int32 maximum", strength: math.MaxInt32, required: math.MaxUint16, want: 1},
		{name: "int32 minimum", strength: math.MinInt32, required: 0, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			player := &playerCheckStrengthTestObject4F3180{class: 4}
			item := &playerCheckStrengthTestObject4F3180{}
			def := &playerCheckStrengthTestDef4F3180{required: test.required}
			got := playerCheckStrength4F3180(player, item, playerCheckStrengthHooks4F3180[
				*playerCheckStrengthTestObject4F3180,
				*playerCheckStrengthTestDef4F3180,
			]{
				loadPlayerClassLow: func(object *playerCheckStrengthTestObject4F3180) uint8 { return uint8(object.class) },
				getUnitStrength:    func(*playerCheckStrengthTestObject4F3180) int32 { return test.strength },
				loadItemClass:      func(object *playerCheckStrengthTestObject4F3180) uint32 { return object.class },
				loadItemType:       func(object *playerCheckStrengthTestObject4F3180) uint16 { return object.typeInd },
				findWeaponDef:      func(uint16) *playerCheckStrengthTestDef4F3180 { return def },
				loadRequired:       func(def *playerCheckStrengthTestDef4F3180) uint16 { return def.required },
			})
			if got != test.want {
				t.Fatalf("result = %d, want %d", got, test.want)
			}
		})
	}
}

func TestPlayerCheckStrength4F3180FaultOrder(t *testing.T) {
	t.Run("nil player faults before strength and item", func(t *testing.T) {
		var events []string
		defer func() {
			if recover() == nil {
				t.Fatal("nil player class load did not fault")
			}
			if want := []string{"player-class"}; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		}()
		playerCheckStrength4F3180((*playerCheckStrengthTestObject4F3180)(nil), (*playerCheckStrengthTestObject4F3180)(nil), playerCheckStrengthHooks4F3180[
			*playerCheckStrengthTestObject4F3180,
			*playerCheckStrengthTestDef4F3180,
		]{
			loadPlayerClassLow: func(object *playerCheckStrengthTestObject4F3180) uint8 {
				events = append(events, "player-class")
				return uint8(object.class)
			},
			getUnitStrength: func(*playerCheckStrengthTestObject4F3180) int32 {
				events = append(events, "strength")
				return 0
			},
			loadItemClass: func(*playerCheckStrengthTestObject4F3180) uint32 {
				events = append(events, "item-class")
				return 0
			},
		})
	})

	t.Run("nil item faults only after strength", func(t *testing.T) {
		player := &playerCheckStrengthTestObject4F3180{class: 4}
		var events []string
		defer func() {
			if recover() == nil {
				t.Fatal("nil item class load did not fault")
			}
			want := []string{"player-class", "strength", "item-class"}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		}()
		playerCheckStrength4F3180(player, (*playerCheckStrengthTestObject4F3180)(nil), playerCheckStrengthHooks4F3180[
			*playerCheckStrengthTestObject4F3180,
			*playerCheckStrengthTestDef4F3180,
		]{
			loadPlayerClassLow: func(object *playerCheckStrengthTestObject4F3180) uint8 {
				events = append(events, "player-class")
				return uint8(object.class)
			},
			getUnitStrength: func(*playerCheckStrengthTestObject4F3180) int32 {
				events = append(events, "strength")
				return 30
			},
			loadItemClass: func(object *playerCheckStrengthTestObject4F3180) uint32 {
				events = append(events, "item-class")
				return object.class
			},
		})
	})
}
