package legacy

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
)

func TestTeleportToMB4E7190MatchesGAMEEXEGates(t *testing.T) {
	type unit struct {
		anchored      bool
		flags         object.Flags
		class         object.Class
		subclass      object.SubClass
		coopClass     object.Class
		mutateOnCoop  bool
		movementCount int
	}
	tests := []struct {
		name       string
		unit       unit
		quest      bool
		coop       bool
		wantEvents []string
		wantMoves  int
	}{
		{
			name:       "anchored rejects before flags",
			unit:       unit{anchored: true, flags: object.FlagNoUpdate},
			wantEvents: []string{"anchor"},
		},
		{
			name:       "no update rejects before game flags",
			unit:       unit{flags: object.FlagNoUpdate | object.FlagEnabled},
			wantEvents: []string{"anchor", "flags"},
		},
		{
			name: "quest shopkeeper rejects before coop",
			unit: unit{
				class:    object.ClassMonster | object.ClassPlayer,
				subclass: object.SubClass(object.MonsterShopkeeper | object.MonsterNPC),
			},
			quest:      true,
			coop:       true,
			wantEvents: []string{"anchor", "flags", "quest", "class", "subclass"},
		},
		{
			name: "quest ordinary monster moves outside coop",
			unit: unit{
				class:    object.ClassMonster,
				subclass: object.SubClass(object.MonsterNPC),
			},
			quest:      true,
			wantEvents: []string{"anchor", "flags", "quest", "class", "subclass", "coop", "class", "move"},
			wantMoves:  1,
		},
		{
			name:       "coop admits non unit without class read",
			unit:       unit{class: object.ClassSimple},
			coop:       true,
			wantEvents: []string{"anchor", "flags", "quest", "coop", "move"},
			wantMoves:  1,
		},
		{
			name:       "non coop rejects non unit",
			unit:       unit{class: object.ClassMissile},
			wantEvents: []string{"anchor", "flags", "quest", "coop", "class"},
		},
		{
			name:       "non coop admits player with additional bits",
			unit:       unit{class: object.ClassPlayer | object.ClassClientPersist},
			wantEvents: []string{"anchor", "flags", "quest", "coop", "class", "move"},
			wantMoves:  1,
		},
		{
			name: "class is reloaded after coop call",
			unit: unit{
				class:        object.ClassSimple,
				coopClass:    object.ClassMonster,
				mutateOnCoop: true,
			},
			quest:      true,
			wantEvents: []string{"anchor", "flags", "quest", "class", "coop", "class", "move"},
			wantMoves:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := tc.unit
			var events []string
			h := teleportToMB4E7190Hooks[*unit]{
				anchored: func(got *unit) bool {
					events = append(events, "anchor")
					return got.anchored
				},
				flags: func(got *unit) object.Flags {
					events = append(events, "flags")
					return got.flags
				},
				quest: func() bool {
					events = append(events, "quest")
					return tc.quest
				},
				class: func(got *unit) object.Class {
					events = append(events, "class")
					return got.class
				},
				subclass: func(got *unit) object.SubClass {
					events = append(events, "subclass")
					return got.subclass
				},
				coop: func() bool {
					events = append(events, "coop")
					if u.mutateOnCoop {
						u.class = u.coopClass
					}
					return tc.coop
				},
				move: func(got *unit) {
					events = append(events, "move")
					got.movementCount++
				},
			}

			teleportToMB4E7190(&u, h)
			if fmt.Sprint(events) != fmt.Sprint(tc.wantEvents) {
				t.Fatalf("events: got %v, want %v", events, tc.wantEvents)
			}
			if u.movementCount != tc.wantMoves {
				t.Fatalf("move count: got %d, want %d", u.movementCount, tc.wantMoves)
			}
		})
	}
}

func TestTeleportToMB4E7190NilObjectFaultsAfterEnchantCheck(t *testing.T) {
	type unit struct {
		flags object.Flags
	}
	var events []string
	h := teleportToMB4E7190Hooks[*unit]{
		anchored: func(got *unit) bool {
			events = append(events, "anchor")
			_ = got
			return false
		},
		flags: func(got *unit) object.Flags {
			events = append(events, "flags")
			return got.flags
		},
		quest:    func() bool { t.Fatal("quest called after nil fault"); return false },
		class:    func(*unit) object.Class { t.Fatal("class called after nil fault"); return 0 },
		subclass: func(*unit) object.SubClass { t.Fatal("subclass called after nil fault"); return 0 },
		coop:     func() bool { t.Fatal("coop called after nil fault"); return false },
		move:     func(*unit) { t.Fatal("move called after nil fault") },
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teleportToMB4E7190[*unit](nil, h)
	}()
	if recovered == nil {
		t.Fatal("nil object did not fault on flags read")
	}
	if want := []string{"anchor", "flags"}; fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events: got %v, want %v", events, want)
	}
}
