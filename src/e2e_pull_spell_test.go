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

func TestE2EPullSpellModes(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player", ""} {
		for level := -1; level <= 6; level++ {
			fromNPC, ok := e2ePullMode(level, direction)
			want := direction == "player-to-npc" && level >= 1 && level <= 5 || direction == "npc-to-player" && level == 0
			if ok != want || ok && fromNPC != (direction == "npc-to-player") {
				t.Fatalf("Pull mode=%s/%d got=%t/%t want ok=%t", direction, level, fromNPC, ok, want)
			}
		}
	}
}

func TestE2EPullSpellSchedule(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player"} {
		level := 0
		if direction == "player-to-npc" {
			level = 3
		}
		var sc e2eScenario
		sc.CheckPullSpell(level, direction, "Pull")
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

func TestE2EPullSpellInvalidFixture(t *testing.T) {
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
			sc.CheckPullSpell(tc.level, tc.dir, "invalid")
		}()
	}
}

func TestE2EPullSpellForcePrediction(t *testing.T) {
	for _, tc := range []struct {
		position types.Pointf
		want     types.Pointf
	}{{types.Ptf(0, 0), types.Ptf(0, 0)}, {types.Ptf(3, 0), types.Ptf(-30.0/3.1, 0)},
		{types.Ptf(0, -3), types.Ptf(0, 30.0/3.1)}, {types.Ptf(601, 0), types.Ptf(0, 0)}} {
		got := e2ePullExpectedForce(types.Ptf(0, 0), tc.position, 1, 1, 1)
		if math.Abs(float64(got.X-tc.want.X)) > 1e-6 || math.Abs(float64(got.Y-tc.want.Y)) > 1e-6 {
			t.Fatalf("%v force=%v want=%v", tc.position, got, tc.want)
		}
	}
	base := e2ePullExpectedForce(types.Ptf(100, 200), types.Ptf(212, 200), 20, 1.00000004, 3)
	halfMass := e2ePullExpectedForce(types.Ptf(100, 200), types.Ptf(212, 200), 10, 1.00000004, 3)
	if base.X >= 0 || base.Y != 0 || halfMass.X != 2*base.X {
		t.Fatalf("attenuation/mass prediction=%v/%v", base, halfMass)
	}
}

func TestE2EPullSpellForcePolarity(t *testing.T) {
	origin := types.Ptf(100, 200)
	for level := int32(1); level <= 5; level++ {
		for _, mass := range []float32{15, 15.6} {
			for _, delta := range []types.Pointf{types.Ptf(3, 4), types.Ptf(112, 0), types.Ptf(-79, 79), types.Ptf(601, 0)} {
				position := origin.Add(delta)
				pull := e2ePullExpectedForce(origin, position, mass, 15, level)
				push := e2ePushExpectedForce(origin, position, mass, 15, level)
				if pull.X != -push.X || pull.Y != -push.Y || pull != (types.Pointf{}) && pull.X*delta.X+pull.Y*delta.Y >= 0 {
					t.Fatalf("level/mass/delta=%d/%g/%v Pull=%v Push=%v", level, mass, delta, pull, push)
				}
			}
		}
	}
}

func TestE2EPullSpellScenario(t *testing.T) {
	path := "../scripts/e2e/host-game-pull-spell.yaml"
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
		if step.Action != "check-pull-spell" {
			continue
		}
		if _, ok := e2ePullMode(step.Count, step.Text); !ok || seen[step.Text][step.Count] {
			t.Fatalf("invalid or repeated Pull scenario=%s/%d", step.Text, step.Count)
		}
		if seen[step.Text] == nil {
			seen[step.Text] = map[int]bool{}
		}
		seen[step.Text][step.Count] = true
	}
	if len(seen["player-to-npc"]) != 5 || len(seen["npc-to-player"]) != 1 {
		t.Fatalf("Pull casts=%v want five player levels and one natural NPC power", seen)
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
			t.Fatalf("Pull scenario has a separate result injection: %s", action)
		}
	}
}
