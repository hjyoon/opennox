package opennox

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestE2EPurchasedWolfSchedule(t *testing.T) {
	var sc e2eScenario
	sc.PrepareWar03bWolfPurchase("prepare")
	sc.OpenHenrickDialog("first dialog")
	sc.ClickNPCDialogYes("first Yes")
	sc.AssertWar03bWolfPurchase(1, "first purchase")
	sc.OpenHenrickDialog("second dialog")
	sc.ClickNPCDialogYes("second Yes")
	sc.AssertWar03bWolfPurchase(2, "second purchase")
	sc.ContactPurchasedWolfExit("War03c", "stock exit")
	sc.AssertPurchasedWolfTransition("preserved")
	for name, timeout := range map[string]time.Duration{
		"first dialog wait for stock Yes button":           300,
		"second dialog wait for stock Yes button":          300,
		"first purchase synchronize purchase":              300,
		"second purchase synchronize purchase":             300,
		"second purchase capture native client identities": 300,
		"stock exit wait for destination map":              2400,
	} {
		found := 0
		for _, step := range sc.steps {
			if step.name == name {
				found++
				if step.ready == nil || step.waitTimeout != timeout {
					t.Fatalf("unbounded gate %q", name)
				}
			}
		}
		if found != 1 {
			t.Fatalf("gate %q occurs %d times", name, found)
		}
	}
	for _, name := range []string{"first Yes", "second Yes", "preserved", "preserved check purchased identities and no duplicate wolves"} {
		found := false
		for _, step := range sc.steps {
			if step.name == name && step.fnc != nil {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing input/native assertion %q", name)
		}
	}
}

func TestE2EPurchasedWolfScenario(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/solo-warrior-purchased-wolf-transition.yaml")
	if err != nil {
		t.Fatal(err)
	}
	scenario := string(raw)
	for _, action := range []string{"prepare-war03b-wolf-purchase", "open-henrick-dialog", "click-npc-dialog-yes", "assert-war03b-wolf-purchase", "contact-purchased-wolf-exit", "assert-purchased-wolf-transition"} {
		if !strings.Contains(scenario, "action: "+action) {
			t.Fatalf("missing action %s", action)
		}
	}
	if strings.Count(scenario, "action: click-npc-dialog-yes") != 2 || strings.Count(scenario, "action: assert-purchased-wolf-transition") != 4 ||
		!strings.Contains(scenario, "map: War03c") || !strings.Contains(scenario, "map: War03b") || strings.Count(scenario, "dt: 240") != 2 {
		t.Fatal("scenario must buy two wolves and check immediate/delayed migration on fresh and saved maps")
	}
	for _, action := range []string{"create-transition-summons", "create-transition-spell-pets", "spawn-monster", "call-noxscript-function", "enter-pet-transition-exit"} {
		if strings.Contains(scenario, "action: "+action) {
			t.Fatalf("purchased-wolf scenario substitutes/injects pets: %s", action)
		}
	}
}
