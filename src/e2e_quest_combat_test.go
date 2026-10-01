package opennox

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/server"
)

func TestE2EQuestCombatSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckQuestMinionCombat("Necromancer", "natural combat")
	for name, timeout := range map[string]time.Duration{
		"natural combat wait for natural AI acquisition":          600,
		"natural combat wait for minion attack and client damage": 900,
		"natural combat defeat minion with real mouse input":      2400,
	} {
		found := 0
		for _, step := range sc.steps {
			if step.name == name {
				found++
				if step.ready == nil || step.waitTimeout != timeout {
					t.Fatalf("unbounded natural combat gate %q", name)
				}
			}
		}
		if found != 1 {
			t.Fatalf("gate %q occurs %d times", name, found)
		}
	}
	for _, name := range []string{"natural combat place only player beside stock minion", "natural combat observe original reward items"} {
		found := false
		for _, step := range sc.steps {
			if step.name == name && step.fnc != nil {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing natural combat observation %q", name)
		}
	}
}

func TestE2EQuestCombatSafeSourceIdentity(t *testing.T) {
	unit, weapon := new(server.Object), new(server.Object)
	unit.InvFirstItem = weapon
	if e2eQuestCombatAttributedTo(nil, unit) || e2eQuestCombatAttributedTo(unit, nil) {
		t.Fatal("nil attribution accepted")
	}
	if !e2eQuestCombatAttributedTo(unit, unit) || !e2eQuestCombatAttributedTo(weapon, unit) {
		t.Fatal("live unit/inventory identity attribution rejected")
	}
}

func TestE2EQuestCombatScenario(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-quest-minion-combat.yaml")
	if err != nil {
		t.Fatal(err)
	}
	scenario := string(raw)
	if strings.Count(scenario, "action: enter-quest-exit") != 4 ||
		!strings.Contains(scenario, "action: check-quest-minion-combat") ||
		!strings.Contains(scenario, "creature: Necromancer") ||
		!strings.Contains(scenario, "map: g_forest") {
		t.Fatal("combat requires a stock stage 5 minion after four Quest transitions")
	}
	for _, action := range []string{"spawn-monster", "set-monster-health", "damage-monster", "call-noxscript-function", "switch-map"} {
		if strings.Contains(scenario, "action: "+action) {
			t.Fatalf("natural combat scenario injects minion behavior: %s", action)
		}
	}
}
