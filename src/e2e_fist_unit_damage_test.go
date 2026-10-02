package opennox

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestE2EFistUnitDamageSchedule(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player"} {
		for level := 1; level <= 5; level++ {
			t.Run(fmt.Sprintf("%s/%d", direction, level), func(t *testing.T) {
				var sc e2eScenario
				sc.CheckFistUnitDamage(level, direction, "Fist unit")
				if len(sc.steps) != 7 {
					t.Fatalf("steps=%d want=7", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare units": 1200, " falling drawable": 120, " natural unit collision and client damage": 180, " natural deletion": 300,
				} {
					found := 0
					for _, step := range sc.steps {
						if strings.HasSuffix(step.name, suffix) {
							found++
							if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
								t.Fatalf("%q has no bounded live check", suffix)
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
}

func TestE2EFistUnitDamageRejectsInvalidFixture(t *testing.T) {
	for _, tc := range []struct {
		level     int
		direction string
	}{{0, "player-to-npc"}, {6, "npc-to-player"}, {1, ""}, {1, "player-to-player"}} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("invalid fixture was scheduled: %+v", tc)
				}
			}()
			sc.CheckFistUnitDamage(tc.level, tc.direction, "invalid")
		}()
	}
}

func TestE2EFistUnitDamageArmorCarryPrediction(t *testing.T) {
	for _, tc := range []struct {
		damage       int32
		armor, carry float32
		wantDamage   int32
		wantCarry    float32
	}{{9, 0.5, 0.4, 7, 0.15}, {5, 1, 0, 2, 0.5}, {7, 1, 0, 4, -0.5}, {1, 1, 0, 1, 0.5}, {100, 0, 0, 100, 0}} {
		got, carry := e2eFistUnitExpectedDamage(tc.damage, tc.armor, tc.carry)
		if got != tc.wantDamage || math.Abs(float64(carry-tc.wantCarry)) > 1e-6 {
			t.Fatalf("prediction=%d/%g want=%d/%g", got, carry, tc.wantDamage, tc.wantCarry)
		}
	}
}

func TestE2EFistUnitDamageScenario(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-game-fist-unit-damage.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-fist-unit-damage" {
			continue
		}
		key := fmt.Sprintf("%s/%d", step.Text, step.Count)
		if _, ok := e2eFistUnitDirection(step.Text); !ok || step.Count < 1 || step.Count > 5 || seen[key] {
			t.Fatalf("invalid or repeated scenario: %s", key)
		}
		seen[key] = true
	}
	if len(seen) != 10 {
		t.Fatalf("unit casts=%d want=10", len(seen))
	}
	var sc e2eScenario
	sc.Load("../scripts/e2e/host-game-fist-unit-damage.yaml")
	loaded := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural unit collision and client damage") {
			loaded++
		}
	}
	if loaded != 10 {
		t.Fatalf("loaded unit collision checks=%d want=10", loaded)
	}
	for _, action := range []string{"set-monster-health", "damage-monster", "damage-player", "cast"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("unit scenario has a separate result injection: %s", action)
		}
	}
}
