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

func TestE2ELockLevelsAndBoundedSchedule(t *testing.T) {
	for level := -1; level <= 6; level++ {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			want := level >= 1 && level <= 5
			if e2eLockLevel(level) != want {
				t.Fatal("Lock level validation changed")
			}
			var sc e2eScenario
			if !want {
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Fatal("invalid Lock level scheduled gameplay")
					}
				}()
				sc.CheckLockSpell(level, "invalid")
				return
			}
			sc.CheckLockSpell(level, "Lock")
			if len(sc.steps) != 12 {
				t.Fatalf("steps=%d want=12", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{" prepare": 1200, " player audio": 120, " refresh audio": 120, " natural expiry": 4000, " NPC audio": 120, " deleted units": 120} {
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

func TestE2ELockPublicScenarioAndDispatch(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-game-lock-spell.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[int]int)
	for _, step := range file.Steps {
		if step.Action == "check-lock-spell" {
			seen[step.Count]++
		}
	}
	if len(seen) != 5 || file.Steps[len(file.Steps)-1].Action != "quit" {
		t.Fatal("public Lock world lacks five levels or native shutdown")
	}
	for level := 1; level <= 5; level++ {
		if seen[level] != 1 {
			t.Fatalf("missing/repeated level %d", level)
		}
	}
	var sc e2eScenario
	sc.Load("../scripts/e2e/host-game-lock-spell.yaml")
	if len(sc.steps) < 5*12 {
		t.Fatal("public dispatcher did not schedule all bounded casts")
	}
}

func TestE2ELockFixtureDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_lock_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{"ObjOwner": true, "Field34": true, "Field32": true, "Cur": true, "ManaCur": true, "Buffs": true, "BuffsDur": true, "BuffsPower": true, "NetCode": true, "ScriptIDVal": true}
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
		if call, ok := n.(*ast.CallExpr); ok {
			if method, ok := call.Fun.(*ast.SelectorExpr); ok {
				switch method.Sel.Name {
				case "CastLock52CE90", "Nox_xxx_castLock_52CE90", "EventObj", "NetSendPacketXxx0", "SetOwner", "SetFrame", "ResetRandom", "Srand":
					t.Errorf("fixture bypasses normal live game API or supplies an outcome: %s", method.Sel.Name)
				}
			}
		}
		return true
	})
}
