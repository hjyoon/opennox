package opennox

import (
	"math"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/opennox/libs/types"
)

func TestE2EManualSaveSnapshot(t *testing.T) {
	want := e2eManualSaveSnapshot{
		Version: 1, Slot: 1, Stage: 1, Class: 0, Strength: 30,
		Map: "war01a", Name: "Jack", Position: types.Pointf{X: 4404.5, Y: 2104.5},
		Health: 20, MaxHealth: 20, Gold: 100, Experience: math.Float32bits(10), Level: 1,
		Inventory: []e2eSavedInventoryItem{
			{Type: "Apple"}, {Type: "Apple"},
			{Type: "ShortSword", Equipped: true, Health: 90, MaxHealth: 100, Modifiers: [4]string{"Steel"}},
		},
	}
	if err := want.compare(want); err != nil {
		t.Fatal(err)
	}
	got := want
	got.Position.X += 1
	if err := want.compare(got); err != nil {
		t.Fatalf("small original spawn displacement was rejected: %v", err)
	}
	for _, tc := range []struct {
		name string
		edit func(*e2eManualSaveSnapshot)
	}{
		{"wrong slot", func(s *e2eManualSaveSnapshot) { s.Slot = 0 }},
		{"wrong map", func(s *e2eManualSaveSnapshot) { s.Map = "war02a" }},
		{"wrong class", func(s *e2eManualSaveSnapshot) { s.Class = 2 }},
		{"wrong player", func(s *e2eManualSaveSnapshot) { s.Name = "Other" }},
		{"health", func(s *e2eManualSaveSnapshot) { s.Health-- }},
		{"mana", func(s *e2eManualSaveSnapshot) { s.Mana++ }},
		{"gold", func(s *e2eManualSaveSnapshot) { s.Gold-- }},
		{"strength", func(s *e2eManualSaveSnapshot) { s.Strength++ }},
		{"experience", func(s *e2eManualSaveSnapshot) { s.Experience++ }},
		{"level", func(s *e2eManualSaveSnapshot) { s.Level++ }},
		{"stage", func(s *e2eManualSaveSnapshot) { s.Stage++ }},
		{"different location", func(s *e2eManualSaveSnapshot) { s.Position.X += 10 }},
		{"nonfinite location", func(s *e2eManualSaveSnapshot) { s.Position.Y = float32(math.NaN()) }},
		{"missing duplicate item", func(s *e2eManualSaveSnapshot) { s.Inventory = s.Inventory[1:] }},
		{"equipment", func(s *e2eManualSaveSnapshot) { s.Inventory[2].Equipped = false }},
		{"durability", func(s *e2eManualSaveSnapshot) { s.Inventory[2].Health-- }},
		{"modifier", func(s *e2eManualSaveSnapshot) { s.Inventory[2].Modifiers[0] = "Gold" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := want
			got.Inventory = slices.Clone(want.Inventory)
			tc.edit(&got)
			if err := want.compare(got); err == nil {
				t.Fatal("save restoration mismatch was accepted")
			}
		})
	}
	for _, edit := range []func(*e2eManualSaveSnapshot){
		func(s *e2eManualSaveSnapshot) { s.Version = 0 },
		func(s *e2eManualSaveSnapshot) { s.Slot = NOX_SAVEGAME_XXX_MAX },
		func(s *e2eManualSaveSnapshot) { s.Slot = 0 },
		func(s *e2eManualSaveSnapshot) { s.Health = 0 },
		func(s *e2eManualSaveSnapshot) { s.Health = s.MaxHealth + 1 },
		func(s *e2eManualSaveSnapshot) { s.Mana = s.MaxMana + 1 },
		func(s *e2eManualSaveSnapshot) { s.Stage = 0 },
		func(s *e2eManualSaveSnapshot) { s.Class = 3 },
		func(s *e2eManualSaveSnapshot) { s.Name = "" },
		func(s *e2eManualSaveSnapshot) { s.Map = "" },
		func(s *e2eManualSaveSnapshot) { s.Inventory = nil },
		func(s *e2eManualSaveSnapshot) { s.Position.X = float32(math.Inf(1)) },
	} {
		invalid := want
		edit(&invalid)
		if err := invalid.compare(invalid); err == nil {
			t.Fatalf("invalid self-consistent expectation was accepted: %+v", invalid)
		}
	}
}

func TestE2EManualSaveSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckManualSaveLoad(1, "manual save")
	for _, name := range []string{
		"verify initial manual slot is empty", "verify first manual file",
		"verify refusal leaves files unchanged", "verify overwritten manual file",
		"prove reload must restore an older location", "verify restored player",
		"record scalar expectation for fresh process", "verify control after live reload",
	} {
		found := 0
		for _, step := range sc.steps {
			if step.name == "manual save "+name {
				found++
				if step.fnc == nil && step.ready == nil {
					t.Fatalf("%s has no assertion", step.name)
				}
			}
		}
		if found != 1 {
			t.Fatalf("%s checks = %d, want 1", name, found)
		}
	}
}

func TestE2EManualSaveScenarios(t *testing.T) {
	for _, name := range []string{"seed-solo-warrior-manual-save.yaml", "solo-warrior-manual-save-load.yaml"} {
		raw, err := os.ReadFile("../scripts/e2e/" + name)
		if err != nil {
			t.Fatal(err)
		}
		scenario := string(raw)
		for _, action := range []string{"set-player-health", "spawn-monster", "call-noxscript-function", "switch-map", "force-coop-autosave", "set-player-stat"} {
			if strings.Contains(scenario, "action: "+action) {
				t.Fatalf("%s injects game behavior: %s", name, action)
			}
		}
		if !strings.Contains(scenario, "slot: 1") {
			t.Fatalf("%s does not test a non-autosave slot", name)
		}
	}
}
