package opennox

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EImpFirstAttack230(t *testing.T) {
	for _, slope := range []string{"ascending", "descending"} {
		for _, lane := range []string{"clear", "off-ray-box"} {
			mode := "Imp/" + slope + "/" + lane
			t.Run(mode, func(t *testing.T) {
				kind, descending, obstacle, ok := e2eAIFirstAttackMode(mode)
				if !ok || kind != "Imp" || descending != (slope == "descending") || obstacle != (lane == "off-ray-box") {
					t.Fatalf("mode=%q parsed as %q/%t/%t/%t", mode, kind, descending, obstacle, ok)
				}
				var sc e2eScenario
				sc.CheckAIFirstAttack(mode, "Imp")
				if len(sc.steps) != 3 {
					t.Fatalf("steps=%d want=3", len(sc.steps))
				}
				for i, timeout := range []time.Duration{1200, 600} {
					if sc.steps[i].ready == nil || sc.steps[i].fnc == nil || sc.steps[i].waitTimeout != timeout {
						t.Fatalf("step %d lost original live check/deadline", i)
					}
				}
				if sc.steps[2].time-sc.steps[1].time != 12 {
					t.Fatal("ordinary deletion cleanup changed")
				}
			})
		}
	}
	t.Run("observed-projectile-source", func(t *testing.T) {
		enemy, missile, unrelated, weapon := new(server.Object), new(server.Object), new(server.Object), new(server.Object)
		f := e2eAIFirstAttackFixture{kind: "Imp", enemy: enemy, weapon: weapon, projectiles: map[*server.Object]bool{missile: true}}
		for _, source := range []*server.Object{nil, enemy, missile, unrelated, weapon} {
			want := source == enemy || source == missile
			if got := f.acceptsDamageSource(source); got != want {
				t.Errorf("source=%p accepted=%t want=%t", source, got, want)
			}
		}
		delete(f.projectiles, missile)
		if f.acceptsDamageSource(missile) {
			t.Fatal("unobserved projectile accepted")
		}
	})
}

func TestE2EImpFirstAttackScenario230(t *testing.T) {
	const path = "../scripts/e2e/host-game-imp-first-attack.yaml"
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
		if step.Action == "check-ai-first-attack" {
			if !strings.HasPrefix(step.Text, "Imp/") || seen[step.Text] {
				t.Fatalf("unexpected/repeated Imp mode %q", step.Text)
			}
			seen[step.Text] = true
		}
		if step.Action != "slow" && step.Action != "wait" && step.Action != "click" && step.Action != "check-ai-first-attack" && step.Action != "quit" {
			t.Fatalf("scenario injects non-input action %q", step.Action)
		}
	}
	for _, slope := range []string{"ascending", "descending"} {
		for _, lane := range []string{"clear", "off-ray-box"} {
			if !seen["Imp/"+slope+"/"+lane] {
				t.Fatalf("missing Imp mode %s/%s", slope, lane)
			}
		}
	}
	if len(seen) != 4 {
		t.Fatalf("modes=%d want=4", len(seen))
	}
}
