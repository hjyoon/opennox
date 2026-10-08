package opennox

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestE2EGameplayAudioScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-gameplay-audio.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	var phase string
	seen := make(map[string]bool)
	names := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-gameplay-audio" {
			continue
		}
		if step.Text == "" || step.Name == "" || names[step.Name] {
			t.Fatalf("empty/repeated audio observation: %+v", step)
		}
		names[step.Name] = true
		switch step.Mode {
		case 0:
			if phase != "" || seen[step.Text] {
				t.Fatalf("overlapping/repeated phase %q", step.Text)
			}
			phase = step.Text
		case 1:
			if phase != step.Text {
				t.Fatalf("end phase %q without matching baseline %q", step.Text, phase)
			}
			seen[phase] = true
			phase = ""
		default:
			t.Fatalf("invalid audio observation mode %d", step.Mode)
		}
	}
	if phase != "" || len(seen) != 6 {
		t.Fatalf("completed phases=%v active=%q", seen, phase)
	}
	for _, phase := range []string{"walk", "run", "pickup", "drop", "attack", "hurt"} {
		if !seen[phase] {
			t.Fatalf("missing phase %q", phase)
		}
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if names[step.name] {
			checks++
			if step.fnc == nil {
				t.Fatalf("audio observation %q has no live callback", step.name)
			}
		}
	}
	if checks != 12 {
		t.Fatalf("registered live audio observations=%d want=12", checks)
	}
	for _, action := range []string{"damage-monster", "damage-player", "cast", "set-monster-health", "check-native-audio", "screen"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("separate sound/result/pixel injection %s", action)
		}
	}
}
