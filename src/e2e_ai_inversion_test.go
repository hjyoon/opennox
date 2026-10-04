package opennox

import (
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EAIInversionSchedule(t *testing.T) {
	for _, kind := range []string{"Wizard", "WizardGreen", "UrchinShaman"} {
		for _, slope := range []string{"ascending", "descending"} {
			mode := kind + "/" + slope
			t.Run(mode, func(t *testing.T) {
				gotKind, descending, ok := e2eAIInversionMode(mode)
				if !ok || gotKind != kind || descending != (slope == "descending") {
					t.Fatal("stock inversion mode was not preserved")
				}
				var sc e2eScenario
				sc.CheckAIInversion(mode, "inversion")
				if len(sc.steps) != 5 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
					sc.steps[2].ready == nil || sc.steps[2].waitTimeout != 1200 ||
					sc.steps[3].ready == nil || sc.steps[3].waitTimeout != 300 ||
					!strings.HasSuffix(sc.steps[1].name, " ordinary incoming magic attack") ||
					!strings.HasSuffix(sc.steps[2].name, " autonomous inversion and return hit") ||
					!strings.HasSuffix(sc.steps[3].name, " natural missile removal") ||
					sc.steps[1].time-sc.steps[0].time != 5 || sc.steps[3].time-sc.steps[2].time != 1 || sc.steps[4].time-sc.steps[3].time != 12 {
					t.Fatal("stock setup/attack/bounded natural observer/cleanup schedule changed")
				}
			})
		}
	}
}

func TestE2EAIInversionRejectsInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "Wizard", "NPC/ascending", "wizard/ascending", "Wizard/other", "Wizard/ascending/forced", "WizardGreen/"} {
		t.Run(strings.ReplaceAll(mode, "/", ":"), func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid mode scheduled a probe")
				}
			}()
			sc.CheckAIInversion(mode, "invalid")
		})
	}
}

func TestE2EAIInversionReflectionIdentityAndNativeFields(t *testing.T) {
	for _, mode := range []string{"valid", "nil missile", "nil update", "wire reuse", "script reuse", "destroyed", "dead", "nonmissile", "nonmagic", "object owner", "update owner", "target", "spell", "lifetime"} {
		t.Run(mode, func(t *testing.T) {
			attacker, defender := new(server.Object), new(server.Object)
			data := &server.MissileUpdateData{Owner: defender, Target: attacker, SpellID: int32(spell.SPELL_MAGIC_MISSILE)}
			missile := &server.Object{
				NetCode: 914, ScriptIDVal: 270, ObjClass: object.ClassMissile,
				ObjSubClass: object.SubClass(object.MissileMagic), ObjOwner: defender,
				Field32: 712, UpdateData: unsafe.Pointer(data),
			}
			switch mode {
			case "nil missile":
				missile = nil
			case "nil update":
				missile.UpdateData = nil
			case "wire reuse":
				missile.NetCode++
			case "script reuse":
				missile.ScriptIDVal++
			case "destroyed":
				missile.ObjFlags = object.FlagDestroyed
			case "dead":
				missile.ObjFlags = object.FlagDead
			case "nonmissile":
				missile.ObjClass = object.ClassMonster
			case "nonmagic":
				missile.ObjSubClass = 0
			case "object owner":
				missile.ObjOwner = attacker
			case "update owner":
				data.Owner = attacker
			case "target":
				data.Target = defender
			case "spell":
				data.SpellID = int32(spell.SPELL_SWAP)
			case "lifetime":
				missile.Field32--
			}
			before := *data
			if got := e2eAIInversionReflection(missile, attacker, defender, 914, 270, 712); got != (mode == "valid") {
				t.Fatalf("reflection accepted=%t mode=%s", got, mode)
			}
			if *data != before {
				t.Fatal("observer supplied missile ownership/target results")
			}
		})
	}
}

func TestE2EAIInversionOwnerList(t *testing.T) {
	first, missile, absent := new(server.Object), new(server.Object), new(server.Object)
	first.Field128 = missile
	owner := &server.Object{Field129: first}
	if !e2eAIInversionOwned(owner, first) || !e2eAIInversionOwned(owner, missile) || e2eAIInversionOwned(owner, absent) {
		t.Fatal("native owner chain identity mismatch")
	}
	missile.Field128 = first
	if e2eAIInversionOwned(owner, absent) {
		t.Fatal("absent object accepted in malformed cyclic owner chain")
	}
}

func TestE2EAIInversionScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-ai-inversion.yaml"
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
		if step.Action != "check-ai-inversion" {
			continue
		}
		if _, _, ok := e2eAIInversionMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated stock inversion mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 6 {
		t.Fatalf("stock inversion modes=%d want=6", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " autonomous inversion and return hit") {
			checks++
		}
	}
	if checks != 6 {
		t.Fatalf("bounded natural observers=%d want=6", checks)
	}
}
