package opennox

import (
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EAIBlinkSchedule(t *testing.T) {
	for _, kind := range []string{"Troll", "Urchin", "NPC"} {
		for _, slope := range []string{"ascending", "descending"} {
			mode := kind + "/" + slope
			t.Run(mode, func(t *testing.T) {
				got, descending, ok := e2eAIBlinkMode(mode)
				if !ok || got != kind || descending != (slope == "descending") {
					t.Fatal("MainAI Blink mode changed")
				}
				var sc e2eScenario
				sc.CheckAIBlink(mode, "Blink")
				if len(sc.steps) != 4 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
					sc.steps[1].ready == nil || sc.steps[1].waitTimeout != 600 ||
					sc.steps[2].ready == nil || sc.steps[2].waitTimeout != 600 ||
					!strings.HasSuffix(sc.steps[1].name, " natural first attack then configure Blink ability") ||
					!strings.HasSuffix(sc.steps[2].name, " autonomous MainAI Blink and natural teleport") ||
					sc.steps[1].time-sc.steps[0].time != 1 || sc.steps[2].time-sc.steps[1].time != 1 ||
					sc.steps[3].time-sc.steps[2].time != 12 {
					t.Fatal("bounded natural first-attack/Blink observer schedule changed")
				}
			})
		}
	}
}

func TestE2EAIBlinkInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "Troll", "Spider/ascending", "troll/ascending", "NPC/other", "NPC/ascending/forced"} {
		t.Run(mode, func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid mode scheduled a MainAI probe")
				}
			}()
			sc.CheckAIBlink(mode, "invalid")
		})
	}
}

func TestE2EAIBlinkDurationIdentityAndNativeCallbacks(t *testing.T) {
	for _, mode := range []string{"valid", "nil", "spell", "source", "caster", "anchor", "glyph", "target", "clock", "create", "update"} {
		t.Run(mode, func(t *testing.T) {
			target, other := new(server.Object), new(server.Object)
			record := &server.DurSpell{Spell: uint32(spell.SPELL_BLINK), Obj12: target, Caster16: target,
				Target48: target, Frame68: 201, Create: legacy.Get_nox_xxx_spellBlink2_530310(), Update: legacy.Get_nox_xxx_spellBlink1_530380()}
			switch mode {
			case "nil":
				record = nil
			case "spell":
				record.Spell = uint32(spell.SPELL_SWAP)
			case "source":
				record.Obj12 = other
			case "caster":
				record.Caster16 = other
			case "anchor":
				record.Obj24 = other
			case "glyph":
				record.Flag20 = 1
			case "target":
				record.Target48 = other
			case "clock":
				record.Frame68++
			case "create":
				record.Create = nil
			case "update":
				record.Update = nil
			}
			var before server.DurSpell
			if record != nil {
				before = *record
			}
			if got := e2eAIBlinkDuration(record, target, 200); got != (mode == "valid") {
				t.Fatalf("duration accepted=%t mode=%s", got, mode)
			}
			if record != nil && *record != before {
				t.Fatal("duration observer altered the result")
			}
		})
	}
}

func TestE2EAIBlinkScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-ai-blink.yaml"
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
		if step.Action != "check-ai-blink" {
			continue
		}
		if _, _, ok := e2eAIBlinkMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated Blink mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 6 {
		t.Fatalf("MainAI Blink modes=%d want=6", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " autonomous MainAI Blink and natural teleport") {
			checks++
		}
	}
	if checks != 6 {
		t.Fatalf("natural Blink observers=%d want=6", checks)
	}
}
