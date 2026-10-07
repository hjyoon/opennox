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

func TestE2ETelekinesisModesAndBoundedSchedule(t *testing.T) {
	for _, mode := range []string{"player-script", "npc-animated", "", "npc-direct"} {
		for level := -1; level <= 6; level++ {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				want := mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
				if got := e2eTelekinesisMode(level, mode); got != want {
					t.Fatalf("accepted=%t want=%t", got, want)
				}
				var sc e2eScenario
				if !want {
					defer func() {
						if recover() == nil || len(sc.steps) != 0 {
							t.Fatal("invalid Telekinesis scheduled gameplay")
						}
					}()
					sc.CheckTelekinesisSpell(level, mode, "invalid")
					return
				}
				sc.CheckTelekinesisSpell(level, mode, "Telekinesis")
				if len(sc.steps) != 11 {
					t.Fatalf("steps=%d want=11", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare": 1200, " real cursor movement 1": 180, " real cursor movement 2": 180, " natural expiry": 2000,
				} {
					count := 0
					for _, step := range sc.steps {
						if strings.HasSuffix(step.name, suffix) {
							count++
							if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
								t.Errorf("unbounded/missing observation %s", suffix)
							}
						}
					}
					if count != 1 {
						t.Errorf("observation %s count=%d", suffix, count)
					}
				}
			})
		}
	}
}

func TestE2ETelekinesisPublicScenarioAndDispatch(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-game-telekinesis-spell.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]int)
	for _, step := range file.Steps {
		if step.Action == "check-telekinesis-spell" {
			seen[fmt.Sprintf("%s/%d", step.Text, step.Count)]++
		}
	}
	if len(seen) != 6 || seen["npc-animated/0"] != 1 {
		t.Fatalf("public cases=%v", seen)
	}
	for level := 1; level <= 5; level++ {
		if seen[fmt.Sprintf("player-script/%d", level)] != 1 {
			t.Fatalf("missing/repeated level %d", level)
		}
	}
	if len(file.Steps) == 0 || file.Steps[len(file.Steps)-1].Action != "quit" {
		t.Fatal("public world lacks native terminal shutdown")
	}
	var sc e2eScenario
	sc.Load("../scripts/e2e/host-game-telekinesis-spell.yaml")
	if len(sc.steps) < 6*11 {
		t.Fatal("public dispatcher did not schedule all six bounded casts")
	}
}

func TestE2ETelekinesisFixtureDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_telekinesis_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{
		"CursorVec": true, "Buffs": true, "BuffsDur": true, "BuffsPower": true,
		"ObjOwner": true, "Field32": true, "Field34": true, "NetCode": true,
		"ScriptIDVal": true, "Cur": true, "ManaCur": true,
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if assignment, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignment.Lhs {
				ast.Inspect(lhs, func(part ast.Node) bool {
					if field, ok := part.(*ast.SelectorExpr); ok && protected[field.Sel.Name] {
						t.Errorf("fixture writes live outcome %s", field.Sel.Name)
					}
					return true
				})
			}
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch method.Sel.Name {
		case "BuffApply", "BuffOff", "TelekinesisUpdate53D330", "Nox_xxx_castTelekinesis_52D330":
			t.Errorf("fixture bypasses live cast/update: %s", method.Sel.Name)
		case "DelayedDelete":
			if len(call.Args) == 1 {
				if field, ok := call.Args[0].(*ast.SelectorExpr); ok && field.Sel.Name == "hand" {
					t.Error("fixture forces hand expiry")
				}
			}
		}
		return true
	})
}
