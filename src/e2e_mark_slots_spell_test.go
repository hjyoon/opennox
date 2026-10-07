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

func TestE2EMarkSlotsLevelsAndBoundedSchedule(t *testing.T) {
	for level := -1; level <= 6; level++ {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			var sc e2eScenario
			if level < 1 || level > 5 {
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Fatal("invalid named Mark level scheduled gameplay")
					}
				}()
				sc.CheckMarkSlotsSpell(level, "invalid")
				return
			}
			sc.CheckMarkSlotsSpell(level, "named Mark")
			if len(sc.steps) != 24 {
				t.Fatalf("steps=%d want=24", len(sc.steps))
			}
			for suffix, expected := range map[string]struct {
				timeout time.Duration
				count   int
			}{
				" prepare": {1200, 1}, " published": {120, 8}, " NPC no-op observation": {120, 1}, " deleted units": {120, 1},
			} {
				count := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						count++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != expected.timeout {
							t.Errorf("unbounded/missing observation %s", suffix)
						}
					}
				}
				if count != expected.count {
					t.Errorf("observation %s count=%d want=%d", suffix, count, expected.count)
				}
			}
		})
	}
}

func TestE2EMarkSlotsPublicScenarioAndDispatch(t *testing.T) {
	const path = "../scripts/e2e/host-game-mark-slots-spell.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[int]int)
	for _, step := range file.Steps {
		if step.Action == "check-mark-slots-spell" {
			seen[step.Count]++
		}
	}
	if len(seen) != 5 || file.Steps[len(file.Steps)-1].Action != "quit" {
		t.Fatal("public named Mark world lacks five levels or native shutdown")
	}
	for level := 1; level <= 5; level++ {
		if seen[level] != 1 {
			t.Fatalf("missing/repeated named Mark level %d", level)
		}
	}
	var sc e2eScenario
	sc.Load(path)
	if len(sc.steps) < 5*24 {
		t.Fatal("public dispatcher did not schedule all bounded named casts")
	}
}

func TestE2EMarkSlotsFixtureDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_mark_slots_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{"ObjOwner": true, "Field29": true, "Field34": true, "Field32": true, "Field39": true, "Cur": true, "ManaCur": true, "Buffs": true, "BuffsDur": true, "BuffsPower": true, "NetCode": true, "ScriptIDVal": true}
	forbidden := map[string]bool{"CastMarkSlot52CBD0": true, "Sub_52CBD0": true, "EventObj": true, "NetSendPacketXxx0": true, "SetOwner": true, "SetFrame": true, "ResetRandom": true, "Srand": true, "audioEventObj": true}
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
			var method string
			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				method = fun.Sel.Name
			case *ast.Ident:
				method = fun.Name
			}
			if forbidden[method] {
				t.Errorf("fixture bypasses normal live game API or supplies an outcome: %s", method)
			}
		}
		return true
	})
}
