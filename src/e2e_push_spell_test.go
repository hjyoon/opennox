package opennox

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/types"
	"gopkg.in/yaml.v2"
)

func TestE2EPushSpellSchedule(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player"} {
		level := 0
		if direction == "player-to-npc" {
			level = 3
		}
		var sc e2eScenario
		sc.CheckPushSpell(level, direction, "Push")
		if len(sc.steps) != 5 {
			t.Fatalf("%s steps=%d want=5", direction, len(sc.steps))
		}
		for suffix, timeout := range map[string]time.Duration{" prepare": 1200, " actual force and client movement": 300} {
			found := 0
			for _, step := range sc.steps {
				if strings.HasSuffix(step.name, suffix) {
					found++
					if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
						t.Fatalf("%q has no bounded check", suffix)
					}
				}
			}
			if found != 1 {
				t.Fatalf("%q found=%d", suffix, found)
			}
		}
	}
}

func TestE2EPushSpellInvalidFixture(t *testing.T) {
	for _, tc := range []struct {
		level int
		dir   string
	}{{0, "player-to-npc"}, {6, "player-to-npc"}, {1, "npc-to-player"}, {-1, "npc-to-player"}, {1, ""}} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("invalid fixture scheduled: %+v", tc)
				}
			}()
			sc.CheckPushSpell(tc.level, tc.dir, "invalid")
		}()
	}
}

func TestE2EPushSpellForcePrediction(t *testing.T) {
	for _, tc := range []struct {
		position types.Pointf
		want     types.Pointf
	}{{types.Ptf(0, 0), types.Ptf(0, 0)}, {types.Ptf(3, 0), types.Ptf(30.0/3.1, 0)},
		{types.Ptf(0, -3), types.Ptf(0, -30.0/3.1)}, {types.Ptf(601, 0), types.Ptf(0, 0)}} {
		got := e2ePushExpectedForce(types.Ptf(0, 0), tc.position, 1, 1, 1)
		if math.Abs(float64(got.X-tc.want.X)) > 1e-6 || math.Abs(float64(got.Y-tc.want.Y)) > 1e-6 {
			t.Fatalf("%v force=%v want=%v", tc.position, got, tc.want)
		}
	}
	base := e2ePushExpectedForce(types.Ptf(100, 200), types.Ptf(212, 200), 20, 1.00000004, 3)
	halfMass := e2ePushExpectedForce(types.Ptf(100, 200), types.Ptf(212, 200), 10, 1.00000004, 3)
	if base.X <= 0 || base.Y != 0 || halfMass.X != 2*base.X {
		t.Fatalf("attenuation/mass prediction=%v/%v", base, halfMass)
	}
}

func TestE2EPushSpellScenario(t *testing.T) {
	path := "../scripts/e2e/host-game-push-spell.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := map[string]map[int]bool{}
	for _, step := range file.Steps {
		if step.Action != "check-push-spell" {
			continue
		}
		if _, ok := e2ePushMode(step.Count, step.Text); !ok || seen[step.Text][step.Count] {
			t.Fatalf("invalid or repeated Push scenario=%s/%d", step.Text, step.Count)
		}
		if seen[step.Text] == nil {
			seen[step.Text] = map[int]bool{}
		}
		seen[step.Text][step.Count] = true
	}
	if len(seen["player-to-npc"]) != 5 || len(seen["npc-to-player"]) != 1 {
		t.Fatalf("Push casts=%v want five player levels and one natural NPC power", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	loaded := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " actual force and client movement") {
			loaded++
		}
	}
	if loaded != 6 {
		t.Fatalf("loaded checks=%d want=6", loaded)
	}
	for _, action := range []string{"cast", "damage-monster", "damage-player", "set-monster-health"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("Push scenario has a separate result injection: %s", action)
		}
	}
}
