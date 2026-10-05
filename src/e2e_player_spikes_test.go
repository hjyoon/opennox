package opennox

import (
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestE2EPlayerSpikeSchedule(t *testing.T) {
	for _, tc := range []struct {
		typeID   string
		damage   uint8
		switched bool
	}{
		{"Spike", 2, true}, {"PeriodicSpike", 2, true},
		{"SpikeBlock", 3, false}, {"SpikeBlockImmobile", 3, false},
		{"RotatingSpikes", 8, false}, {"RotatingSpikesImmobile", 8, false},
	} {
		t.Run(tc.typeID, func(t *testing.T) {
			damage, switched, ok := e2ePlayerSpikeKind(tc.typeID)
			if !ok || damage != tc.damage || switched != tc.switched {
				t.Fatalf("kind=%d/%t/%t want=%d/%t/true", damage, switched, ok, tc.damage, tc.switched)
			}
			var sc e2eScenario
			sc.CheckPlayerSpikeCollision(tc.typeID, tc.typeID)
			if len(sc.steps) != 7 {
				t.Fatalf("steps=%d want=7", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{
				" prepare stock hazard": 1200, " actual collision damage": 120, " client damage": 120,
			} {
				found := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						found++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Fatalf("%q is not a bounded live check", suffix)
						}
					}
				}
				if found != 1 {
					t.Fatalf("%q found=%d want=1", suffix, found)
				}
			}
		})
	}
}

func TestE2EPlayerSpikeRejectsInvalidKind(t *testing.T) {
	for _, typeID := range []string{"", "SpikeBlockInjected", "NPC"} {
		t.Run(typeID, func(t *testing.T) {
			if _, _, ok := e2ePlayerSpikeKind(typeID); ok {
				t.Fatal("unknown stock fixture accepted")
			}
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid kind was scheduled")
				}
			}()
			sc.CheckPlayerSpikeCollision(typeID, "invalid")
		})
	}
}

func TestE2EPlayerSpikeScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-player-spikes.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-player-spike-collision" {
			continue
		}
		if _, _, ok := e2ePlayerSpikeKind(step.Item); !ok || seen[step.Item] {
			t.Fatalf("invalid/repeated stock fixture %q", step.Item)
		}
		seen[step.Item] = true
	}
	if len(seen) != 6 {
		t.Fatalf("types=%d want=6", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " actual collision damage") {
			checks++
		}
	}
	if checks != 6 {
		t.Fatalf("live collision checks=%d want=6", checks)
	}
	for _, action := range []string{"damage-player", "damage-monster", "projectile-fx", "cast", "set-player-health", "set-monster-health"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("separate result injection %s", action)
		}
	}
}
