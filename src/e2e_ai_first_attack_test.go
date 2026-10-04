package opennox

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EAIFirstAttackDamageSource(t *testing.T) {
	enemy, weapon, missile, unrelated := new(server.Object), new(server.Object), new(server.Object), new(server.Object)
	for _, kind := range []string{"Spider", "Troll", "Urchin", "NPC"} {
		for _, candidate := range []struct {
			name   string
			source *server.Object
		}{
			{"nil", nil}, {"enemy", enemy}, {"equipped-weapon", weapon},
			{"observed-owned-missile", missile}, {"unrelated", unrelated},
		} {
			t.Run(kind+"/"+candidate.name, func(t *testing.T) {
				f := e2eAIFirstAttackFixture{kind: kind, enemy: enemy, weapon: weapon, projectiles: map[*server.Object]bool{missile: true}}
				want := candidate.source == enemy || kind == "NPC" && candidate.source == weapon || kind == "Urchin" && candidate.source == missile
				if got := f.acceptsDamageSource(candidate.source); got != want {
					t.Fatalf("source=%s accepted=%t want=%t", candidate.name, got, want)
				}
			})
		}
	}
	t.Run("unobserved-projectile", func(t *testing.T) {
		f := e2eAIFirstAttackFixture{kind: "Urchin", enemy: enemy}
		if f.acceptsDamageSource(missile) {
			t.Fatal("missile without prior owner observation accepted")
		}
	})
}

func TestE2EAIFirstAttackSchedule(t *testing.T) {
	for _, kind := range []string{"Spider", "Troll", "Urchin", "NPC"} {
		for _, slope := range []string{"ascending", "descending"} {
			for _, lane := range []string{"clear", "off-ray-box"} {
				mode := kind + "/" + slope + "/" + lane
				t.Run(mode, func(t *testing.T) {
					gotKind, descending, obstacle, ok := e2eAIFirstAttackMode(mode)
					if !ok || gotKind != kind || descending != (slope == "descending") || obstacle != (lane == "off-ray-box") {
						t.Fatalf("mode parsed as %q/%t/%t/%t", gotKind, descending, obstacle, ok)
					}
					var sc e2eScenario
					sc.CheckAIFirstAttack(mode, "AI")
					if len(sc.steps) != 3 {
						t.Fatalf("steps=%d want=3", len(sc.steps))
					}
					for suffix, timeout := range map[string]time.Duration{
						" prepare untouched hostile AI": 1200, " acquire and attack without being hit": 600,
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
							t.Fatalf("%q found=%d", suffix, found)
						}
					}
					if dt := sc.steps[2].time - sc.steps[1].time; dt != 12 {
						t.Fatalf("ordinary deletion wait=%d want=12", dt)
					}
				})
			}
		}
	}
}

func TestE2EAIFirstAttackRejectsInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "Spider", "Player/ascending/clear", "Spider/other/clear", "NPC/descending/injected", "Spider/ascending/clear/extra", "spider/ascending/clear"} {
		t.Run(mode, func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid mode was scheduled")
				}
			}()
			sc.CheckAIFirstAttack(mode, "invalid")
		})
	}
}

func TestE2EAIFirstAttackScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-ai-first-attack.yaml"
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
		if step.Action != "check-ai-first-attack" {
			continue
		}
		if _, _, _, ok := e2eAIFirstAttackMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 16 {
		t.Fatalf("modes=%d want=16", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " acquire and attack without being hit") {
			checks++
		}
	}
	if checks != 16 {
		t.Fatalf("unforced live checks=%d want=16", checks)
	}
	for _, action := range []string{"damage-monster", "damage-player", "prepare-urchin-ranged-attack", "cast", "set-monster-health"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("separate result injection %s", action)
		}
	}
}
