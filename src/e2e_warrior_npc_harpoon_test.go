package opennox

import (
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/server"
)

func TestE2EWarriorNPCHarpoonSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckWarriorAbility(server.AbilityHarpoon, "harpoon", "npc")
	for suffix, timeout := range map[string]time.Duration{
		" prepare": 1200, " wait for server activation": 120, " wait for real gameplay effect": 240,
		" wait for gameplay end": 1200, " wait for cooldown ready report": 1200,
	} {
		count := 0
		for _, step := range sc.steps {
			if strings.HasSuffix(step.name, suffix) {
				count++
				if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
					t.Fatalf("unbounded/incomplete gate: %+v", step)
				}
			}
		}
		if count != 2 {
			t.Fatalf("gate %q count=%d, want two cycles", suffix, count)
		}
	}
	for _, suffix := range []string{" aim by real mouse input", " activate by real keyboard input", " repeat key during cooldown", " verify cooldown did not restart"} {
		count := 0
		for _, step := range sc.steps {
			if strings.HasSuffix(step.name, suffix) && step.fnc != nil {
				count++
			}
		}
		if count != 2 {
			t.Fatalf("input/retry %q count=%d", suffix, count)
		}
	}
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " prepared") || strings.HasSuffix(step.name, " effect") && step.ready == nil {
			t.Fatalf("NPC diagnostic must not create a screenshot baseline: %q", step.name)
		}
	}
}

func TestE2EWarriorNPCHarpoonRejectsInvalidKind(t *testing.T) {
	for _, tc := range []struct {
		ability server.Ability
		kinds   []string
	}{
		{server.AbilityBerserk, []string{"npc"}}, {server.AbilityHarpoon, []string{"NPC"}},
		{server.AbilityHarpoon, []string{"monster"}}, {server.AbilityHarpoon, []string{"npc", "npc"}},
	} {
		t.Run(tc.ability.String()+strings.Join(tc.kinds, "-"), func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid kind must fail before scheduling mutation")
				}
			}()
			sc.CheckWarriorAbility(tc.ability, "invalid", tc.kinds...)
		})
	}
}
