package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerRemoveSpawnedStuffMatchesGAMEEXEContract(t *testing.T) {
	tests := []struct {
		name        string
		ownerClass  object.Class
		playerClass player.Class
		wantPrefix  []string
	}{
		{name: "non-player skips class cleanup", ownerClass: object.ClassMonster},
		{name: "warrior skips class cleanup", ownerClass: object.ClassPlayer, playerClass: player.Warrior},
		{name: "wizard cleanup runs first", ownerClass: object.ClassPlayer, playerClass: player.Wizard, wantPrefix: []string{"wizard"}},
		{name: "conjurer cleanup runs first", ownerClass: object.ClassPlayer, playerClass: player.Conjurer, wantPrefix: []string{"conjurer", "conjurer", "conjurer"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner := &server.Object{ObjClass: tc.ownerClass}
			if tc.ownerClass.Has(object.ClassPlayer) {
				pl := &server.Player{}
				pl.Info().SetPlayerClass(tc.playerClass)
				owner.UpdateData = unsafe.Pointer(&server.PlayerUpdateData{Player: pl})
			}
			missile := &server.Object{ObjClass: object.ClassMissile, TypeInd: 10}
			disallowed := &server.Object{ObjClass: object.ClassSimple, TypeInd: 11}
			allowed := &server.Object{ObjClass: object.ClassSimple, TypeInd: 12}
			missile.Field128 = disallowed
			disallowed.Field128 = allowed
			owner.Field129 = missile

			names := map[*server.Object]string{
				missile:    "missile",
				disallowed: "disallowed",
				allowed:    "allowed",
			}
			var events []string
			hooks := playerRemoveSpawnedStuffHooks{
				firstObject:  func() *server.Object { t.Fatal("unexpected wizard implementation"); return nil },
				glyphTypeInd: func() int { t.Fatal("unexpected glyph lookup"); return 0 },
				monitored: func(*server.Object, *server.Object) bool {
					t.Fatal("unexpected conjurer implementation")
					return false
				},
				pointFx: func(*server.Object) { t.Fatal("unexpected point effect") },
				typeAllowed: func(typeInd int) bool {
					events = append(events, fmt.Sprintf("allowed:%d", typeInd))
					return typeInd == 12
				},
				delayedDelete: func(unit *server.Object) {
					events = append(events, "delete:"+names[unit])
					unit.Field128 = nil // The original saves the next owned link before deletion.
				},
			}
			hooks.firstObject = func() *server.Object {
				events = append(events, "wizard")
				return nil
			}
			hooks.glyphTypeInd = func() int { return 77 }
			hooks.monitored = func(*server.Object, *server.Object) bool {
				events = append(events, "conjurer")
				return false
			}

			playerRemoveSpawnedStuff_4E5AD0(owner, hooks)

			want := []string{
				"delete:missile",
				"allowed:11",
				"delete:disallowed",
				"allowed:12",
			}
			want = append(append([]string(nil), tc.wantPrefix...), want...)
			if fmt.Sprint(events) != fmt.Sprint(want) {
				t.Fatalf("events: got %v, want %v", events, want)
			}
			if owner.Field129 != missile || allowed.Field128 != nil {
				t.Fatal("cleanup changed state outside the delayed-delete callbacks")
			}
		})
	}
}

func TestPlayerRemoveWizardSpawnedStuffMatchesGAMEEXEContract(t *testing.T) {
	owner := &server.Object{}
	unownedGlyph := &server.Object{TypeInd: 77}
	ownedOther := &server.Object{TypeInd: 76, ObjOwner: owner}
	destroyedGlyph := &server.Object{TypeInd: 77, ObjOwner: owner, ObjFlags: object.FlagDestroyed}
	ownedGlyph := &server.Object{TypeInd: 77, ObjOwner: owner}
	nestedGlyph := &server.Object{TypeInd: 77, ObjOwner: &server.Object{ObjOwner: owner}}
	unownedGlyph.ObjNext = ownedOther
	ownedOther.ObjNext = destroyedGlyph
	destroyedGlyph.ObjNext = ownedGlyph
	ownedGlyph.ObjNext = nestedGlyph

	names := map[*server.Object]string{ownedGlyph: "owned", nestedGlyph: "nested"}
	var events []string
	hooks := playerRemoveSpawnedStuffHooks{
		firstObject: func() *server.Object {
			events = append(events, "first")
			return unownedGlyph
		},
		glyphTypeInd: func() int {
			events = append(events, "glyph")
			return 77
		},
		pointFx: func(unit *server.Object) {
			events = append(events, "fx:"+names[unit])
		},
		delayedDelete: func(unit *server.Object) {
			events = append(events, "delete:"+names[unit])
		},
	}

	playerRemoveWizardSpawnedStuff_4E5F40(owner, hooks)

	want := []string{"glyph", "first", "fx:owned", "delete:owned", "fx:nested", "delete:nested"}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events: got %v, want %v", events, want)
	}
}

func TestPlayerRemoveConjurerSpawnedStuffMatchesGAMEEXEContract(t *testing.T) {
	owner := &server.Object{}
	monitored := &server.Object{}
	unmonitored := &server.Object{}
	item1 := &server.Object{}
	item2 := &server.Object{}
	item1.InvNextItem = item2
	monitored.InvFirstItem = item1
	monitored.Field128 = unmonitored
	owner.Field129 = monitored

	names := map[*server.Object]string{
		monitored:   "monitored",
		unmonitored: "unmonitored",
		item1:       "item1",
		item2:       "item2",
	}
	var events []string
	hooks := playerRemoveSpawnedStuffHooks{
		monitored: func(gotOwner, unit *server.Object) bool {
			if gotOwner != owner {
				t.Fatalf("owner: got %p, want %p", gotOwner, owner)
			}
			events = append(events, "monitor:"+names[unit])
			return unit == monitored
		},
		pointFx: func(unit *server.Object) {
			events = append(events, "fx:"+names[unit])
		},
		delayedDelete: func(unit *server.Object) {
			events = append(events, "delete:"+names[unit])
			if unit == item1 {
				unit.InvNextItem = nil // The original saves the next inventory link first.
			}
			if unit == monitored {
				unit.Field128 = nil // The original saves the next owned link first.
			}
		},
	}

	playerRemoveConjurerSpawnedStuff_4E5FC0(owner, hooks)

	want := []string{
		"monitor:monitored",
		"delete:item1",
		"delete:item2",
		"fx:monitored",
		"delete:monitored",
		"monitor:unmonitored",
	}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events: got %v, want %v", events, want)
	}
}
