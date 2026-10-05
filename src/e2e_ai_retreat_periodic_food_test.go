package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestE2EAIRetreatPeriodicFoodSchedule(t *testing.T) {
	for _, kind := range []string{"Troll", "NPC"} {
		for _, slope := range []string{"ascending", "descending"} {
			for _, item := range []string{"RedApple", "Meat"} {
				t.Run(kind+"/"+slope+"/"+item, func(t *testing.T) {
					var sc e2eScenario
					sc.CheckAIRetreatPeriodicFood(kind+"/"+slope+"/"+item, "periodic")
					if len(sc.steps) != 3 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
						sc.steps[1].ready == nil || sc.steps[1].waitTimeout != 600 ||
						sc.steps[1].time-sc.steps[0].time != 1 || sc.steps[2].time-sc.steps[1].time != 12 ||
						!strings.HasSuffix(sc.steps[1].name, " observe unforced RETREAT movement periodic consumption") {
						t.Fatal("bounded periodic movement/consumption schedule changed")
					}
				})
			}
		}
	}
}

func TestE2EAIRetreatPeriodicFoodInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "Troll/ascending", "troll/ascending/Meat", "Spider/ascending/Meat", "NPC/other/Meat", "NPC/ascending/Mushroom", "NPC/ascending/Meat/forced", "NPC/ascending/"} {
		t.Run(mode, func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid mode scheduled a periodic probe")
				}
			}()
			sc.CheckAIRetreatPeriodicFood(mode, "invalid")
		})
	}
}

func TestE2EAIRetreatPeriodicFoodScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-ai-retreat-periodic-food.yaml"
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
		if step.Action != "check-ai-retreat-periodic-food" {
			continue
		}
		if _, _, _, ok := e2eAIRetreatFoodMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated periodic mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 8 {
		t.Fatalf("periodic modes=%d want=8", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " observe unforced RETREAT movement periodic consumption") {
			checks++
		}
	}
	if checks != 8 {
		t.Fatalf("periodic observers=%d want=8", checks)
	}
}

func TestE2EAIRetreatPeriodicFoodDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_retreat_periodic_food.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Max": true, "Buffs": true, "HealthData": true, "CurrentEnemy": true,
		"PreferredEnemy": true, "AIStack": true, "AIStackInd": true, "Action": true, "Args": true, "PosVec": true,
		"UpdateData": true, "Frame134": true, "Field2": true, "Field124": true, "Field137": true, "ObjFlags": true, "ObjSubClass": true}
	calls := map[string]bool{"SetHealth": true, "SetMaxHealth": true, "MonsterPushAction": true, "MonsterPopAction": true,
		"SetArgs": true, "MonsterActionRetreat545440": true, "MonsterSearchEdible544A00": true,
		"MonsterActionMoveTo5443F0": true, "MonsterActionPickupObject544B90": true, "UseByNetCode53F8E0": true,
		"PlaceInventory": true, "CallPickup": true, "CallDamage": true, "DrawImageAt": true, "ApplyEnchant": true}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range n.Lhs {
					if sel, ok := lhs.(*ast.SelectorExpr); ok && fields[sel.Sel.Name] {
						t.Errorf("periodic observer writes result %s", sel.Sel.Name)
					}
					if _, ok := lhs.(*ast.StarExpr); ok {
						t.Error("periodic observer writes through a pointer")
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
					if calls[sel.Sel.Name] || sel.Sel.Name == "DoDamage" && fn.Name.Name != "prepare" ||
						sel.Sel.Name == "SetPos" && fn.Name.Name != "prepare" && fn.Name.Name != "cleanup" ||
						sel.Sel.Name == "DelayedDelete" && fn.Name.Name != "cleanup" {
						t.Errorf("periodic observer supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
					}
				}
			}
			return true
		})
	}
}

func TestE2EAIRetreatPeriodicFoodOrdinaryPreparation(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_retreat_periodic_food.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	owned, activeAggression := 0, 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "prepare" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			method, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if method.Sel.Name == "CreateObjectAt" && len(call.Args) == 3 {
				unit, ok := call.Args[0].(*ast.SelectorExpr)
				if ok && unit.Sel.Name == "unit" {
					owner, ok := call.Args[1].(*ast.SelectorExpr)
					if !ok || owner.Sel.Name != "host" {
						t.Error("periodic unit lacks ordinary host ownership")
					} else {
						owned++
					}
				}
			}
			if method.Sel.Name == "SetAggression" && len(call.Args) == 1 {
				value, ok := call.Args[0].(*ast.BasicLit)
				if ok && value.Value == "0.3" {
					activeAggression++
				} else {
					t.Error("periodic probe must activate the original food tail")
				}
			}
			return true
		})
	}
	if owned != 1 || activeAggression != 1 {
		t.Fatalf("ordinary ownership/aggression calls=%d/%d want=1/1", owned, activeAggression)
	}
}
