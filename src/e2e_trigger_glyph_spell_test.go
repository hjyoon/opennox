package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestE2ETriggerGlyphModesAndBoundedSchedule(t *testing.T) {
	for _, tc := range []struct {
		mode  string
		level int
		valid bool
	}{
		{"player-script", 1, true}, {"player-script", 2, true}, {"player-script", 3, true},
		{"player-script", 4, true}, {"player-script", 5, true}, {"npc-animated", 0, true},
		{"player-script", -1, false}, {"player-script", 0, false}, {"player-script", 6, false},
		{"npc-animated", -1, false}, {"npc-animated", 1, false}, {"player-script-typo", 1, false},
	} {
		t.Run(fmt.Sprintf("%s_%d", tc.mode, tc.level), func(t *testing.T) {
			var sc e2eScenario
			if !tc.valid {
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Fatal("invalid TriggerGlyph scheduled gameplay")
					}
				}()
				sc.CheckTriggerGlyphSpell(tc.level, tc.mode, "invalid")
				return
			}
			sc.CheckTriggerGlyphSpell(tc.level, tc.mode, "TriggerGlyph")
			if len(sc.steps) != 11 {
				t.Fatalf("steps=%d want=11", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{
				" prepare": 1200, " actual cast and glyph effect": 180, " natural world and client removal": 300,
			} {
				count := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						count++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Errorf("unbounded or missing TriggerGlyph observation %s", suffix)
						}
					}
				}
				if count != 1 {
					t.Errorf("TriggerGlyph observation %s count=%d", suffix, count)
				}
			}
		})
	}
}

func TestE2ETriggerGlyphPublicScenarioAndDispatch(t *testing.T) {
	path := "../scripts/e2e/host-game-trigger-glyph-spell.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	var casts []string
	for _, step := range file.Steps {
		if step.Action == "check-trigger-glyph-spell" {
			casts = append(casts, fmt.Sprintf("%s/%d", step.Text, step.Count))
		}
	}
	if fmt.Sprint(casts) != "[player-script/1 player-script/2 player-script/3 player-script/4 player-script/5 npc-animated/0]" ||
		file.Steps[len(file.Steps)-1].Action != "quit" {
		t.Fatalf("TriggerGlyph public casts/terminal shutdown=%v", casts)
	}
	var sc e2eScenario
	sc.Load(path)
	if len(sc.steps) < 6*11 {
		t.Fatal("public dispatcher did not schedule all six bounded casts")
	}
}

func TestE2ETriggerGlyphFixtureDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_trigger_glyph_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{
		"ObjOwner": true, "ObjNext": true, "Field128": true, "Field129": true, "InitData": true,
		"Spells": true, "SpellsCnt": true, "SpellArg": true, "Update": true, "Damage": true,
		"Buffs": true, "Cur": true, "HealthData": true, "ObjFlags": true, "Field120_1": true, "Field120_2": true,
		"NetCode": true, "ScriptIDVal": true, "Direction1": true, "Direction2": true, "PosVec": true,
	}
	checkWrite := func(lhs ast.Node) {
		ast.Inspect(lhs, func(part ast.Node) bool {
			if field, ok := part.(*ast.SelectorExpr); ok && protected[field.Sel.Name] {
				t.Errorf("Glyph fixture writes live outcome %s", field.Sel.Name)
			}
			return true
		})
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				checkWrite(lhs)
			}
		case *ast.IncDecStmt:
			checkWrite(n.X)
		case *ast.CallExpr:
			name := ""
			switch fn := n.Fun.(type) {
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			case *ast.Ident:
				name = fn.Name
			}
			switch name {
			case "triggerTrap", "castGlyph", "castDetonateGlyphs", "Sub_52CCD0", "CastTriggerGlyph52CCD0",
				"Nox_xxx_dieGlyph_54DF30", "nox_xxx_dieGlyph_54DF30", "CallDamage", "AdjustHP", "BuffApply",
				"SetMaxHealth", "SetHealth", "ObjSetOwner", "NetSendPacket", "Nox_xxx_netSendPointFx_522FF0":
				t.Errorf("Glyph fixture bypasses real cast/effect/removal/network: %s", name)
			}
		}
		return true
	})
}
