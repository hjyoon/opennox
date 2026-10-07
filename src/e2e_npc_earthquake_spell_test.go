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

func TestE2ENPCEarthquakeLevelsAndBoundedSchedule(t *testing.T) {
	for level := -1; level <= 6; level++ {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			var sc e2eScenario
			if level < 1 || level > 5 {
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Fatal("invalid NPC Earthquake scheduled gameplay")
					}
				}()
				sc.CheckNPCEarthquakeSpell(level, "invalid")
				return
			}
			sc.CheckNPCEarthquakeSpell(level, "NPC Earthquake")
			if len(sc.steps) != 8 {
				t.Fatalf("steps=%d want=8", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{
				" prepare": 1200, " actual damage and client replay": 300, " real quake packet and natural settling": 300,
			} {
				count := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						count++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Errorf("unbounded or missing NPC observation %s", suffix)
						}
					}
				}
				if count != 1 {
					t.Errorf("NPC observation %s count=%d", suffix, count)
				}
			}
		})
	}
}

func TestE2ENPCEarthquakePublicScenarioAndDispatch(t *testing.T) {
	path := "../scripts/e2e/host-game-npc-earthquake-spell.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	var levels []int
	for _, step := range file.Steps {
		if step.Action == "check-npc-earthquake-spell" {
			levels = append(levels, step.Count)
			if step.Text != "" {
				t.Fatal("NPC world overrides its real cast mode")
			}
		}
	}
	if fmt.Sprint(levels) != "[1 2 3 4 5]" || file.Steps[len(file.Steps)-1].Action != "quit" {
		t.Fatalf("NPC world levels/terminal shutdown=%v", levels)
	}
	var sc e2eScenario
	sc.Load(path)
	if len(sc.steps) < 5*8 {
		t.Fatal("NPC public dispatcher did not schedule all five bounded casts")
	}
}

func TestE2ENPCEarthquakeFixtureDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_npc_earthquake_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{
		"ObjOwner": true, "Obj130": true, "Field131": true, "Frame134": true, "Damage": true,
		"Buffs": true, "Cur": true, "Jiggle12": true, "Field120_1": true, "Field120_2": true,
		"NetCode": true, "ScriptIDVal": true, "Field518": true, "Field1": true, "Field547": true, "Field546": true,
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if assignment, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignment.Lhs {
				ast.Inspect(lhs, func(part ast.Node) bool {
					if field, ok := part.(*ast.SelectorExpr); ok && protected[field.Sel.Name] {
						t.Errorf("NPC fixture writes live outcome %s", field.Sel.Name)
					}
					return true
				})
			}
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if method, ok := call.Fun.(*ast.SelectorExpr); ok {
				switch method.Sel.Name {
				case "CallDamage", "CastEarthquake52DE40", "Nox_xxx_castEquake_52DE40", "Nox_xxx_earthquakeSend_4D9110", "NetSendPacket", "AdjustHP", "BuffApply":
					t.Errorf("NPC fixture bypasses real cast/damage/network: %s", method.Sel.Name)
				}
			}
		}
		return true
	})
}
