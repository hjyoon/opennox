package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EAIRetreatFoodSchedule(t *testing.T) {
	for _, kind := range []string{"Troll", "NPC"} {
		for _, slope := range []string{"ascending", "descending"} {
			for _, item := range []string{"RedApple", "Meat"} {
				mode := kind + "/" + slope + "/" + item
				t.Run(mode, func(t *testing.T) {
					gotKind, gotItem, descending, ok := e2eAIRetreatFoodMode(mode)
					if !ok || gotKind != kind || gotItem != item || descending != (slope == "descending") {
						t.Fatal("RETREAT food mode changed")
					}
					var sc e2eScenario
					sc.CheckAIRetreatFood(mode, "food")
					if len(sc.steps) != 3 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
						sc.steps[1].ready == nil || sc.steps[1].waitTimeout != 600 ||
						sc.steps[1].time-sc.steps[0].time != 1 || sc.steps[2].time-sc.steps[1].time != 12 ||
						!strings.HasSuffix(sc.steps[1].name, " observe unforced RETREAT movement PICKUP consumption") {
						t.Fatal("bounded food movement/consumption schedule changed")
					}
				})
			}
		}
	}
}

func TestE2EAIRetreatFoodInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "Troll/ascending", "troll/ascending/Meat", "Spider/ascending/Meat", "NPC/other/Meat", "NPC/ascending/Mushroom", "NPC/ascending/Meat/forced", "NPC/ascending/"} {
		t.Run(mode, func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid food mode scheduled a probe")
				}
			}()
			sc.CheckAIRetreatFood(mode, "invalid")
		})
	}
}

func TestE2EAIRetreatFoodMoveIdentity(t *testing.T) {
	for _, mode := range []string{"valid", "nil-update", "nil-food", "no-head", "short-stack", "invalid-index", "no-retreat", "wrong-dependency", "wrong-head", "visible-other", "pickup-other", "move-other", "visible-x", "move-y"} {
		t.Run(mode, func(t *testing.T) {
			food, other := new(server.Object), new(server.Object)
			pos := types.Ptf(300, 400)
			update := &server.MonsterUpdateData{AIStackInd: 6}
			update.AIStack[0].Action = uint32(ai.DEPENDENCY_NOT_CORNERED)
			for i, action := range []ai.ActionType{ai.ACTION_RETREAT, ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO} {
				update.AIStack[i+1].Action = uint32(action)
			}
			update.AIStack[4].SetArgs(pos, food)
			update.AIStack[5].SetArgs(food)
			update.AIStack[6].SetArgs(pos, food)
			switch mode {
			case "nil-update":
				update = nil
			case "nil-food":
				food = nil
			case "no-head":
				update.AIStackInd = -1
			case "short-stack":
				update.AIStackInd = 4
			case "invalid-index":
				update.AIStackInd = int8(len(update.AIStack))
			case "no-retreat":
				update.AIStack[1].Action = uint32(ai.ACTION_FIGHT)
			case "wrong-dependency":
				update.AIStack[3].Action = uint32(ai.DEPENDENCY_TIME)
			case "wrong-head":
				update.AIStack[6].Action = uint32(ai.ACTION_FLEE)
			case "visible-other":
				update.AIStack[4].SetArgs(pos, other)
			case "pickup-other":
				update.AIStack[5].SetArgs(other)
			case "move-other":
				update.AIStack[6].SetArgs(pos, other)
			case "visible-x":
				update.AIStack[4].SetArgs(types.Ptf(301, 400), food)
			case "move-y":
				update.AIStack[6].SetArgs(types.Ptf(300, 401), food)
			}
			var before server.MonsterUpdateData
			if update != nil {
				before = *update
			}
			if got := e2eAIRetreatFoodMoveStack(update, food, pos); got != (mode == "valid") {
				t.Fatalf("food movement identity accepted=%t mode=%s", got, mode)
			}
			if update != nil && *update != before {
				t.Fatal("movement observer altered action/update state")
			}
			runtime.KeepAlive(food)
			runtime.KeepAlive(other)
		})
	}
}

func TestE2EAIRetreatFoodScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-ai-retreat-food.yaml"
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
		if step.Action != "check-ai-retreat-food" {
			continue
		}
		if _, _, _, ok := e2eAIRetreatFoodMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated food mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 8 {
		t.Fatalf("food modes=%d want=8", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " observe unforced RETREAT movement PICKUP consumption") {
			checks++
		}
	}
	if checks != 8 {
		t.Fatalf("food observers=%d want=8", checks)
	}
}

func TestE2EAIRetreatFoodDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_retreat_food.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Max": true, "Buffs": true, "HealthData": true, "CurrentEnemy": true,
		"PreferredEnemy": true, "AIStack": true, "AIStackInd": true, "Action": true, "Args": true, "PosVec": true,
		"UpdateData": true, "Frame134": true, "Field2": true, "Field124": true, "Field137": true}
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
						t.Errorf("food observer writes result %s", sel.Sel.Name)
					}
					if _, ok := lhs.(*ast.StarExpr); ok {
						t.Error("food observer writes through a pointer")
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
					if calls[sel.Sel.Name] || sel.Sel.Name == "DoDamage" && fn.Name.Name != "prepare" ||
						sel.Sel.Name == "SetPos" && fn.Name.Name != "prepare" && fn.Name.Name != "cleanup" ||
						sel.Sel.Name == "DelayedDelete" && fn.Name.Name != "cleanup" {
						t.Errorf("food observer supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
					}
				}
			}
			return true
		})
	}
}

func TestE2EAIRetreatFoodPreparesOrdinaryOwnership(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_retreat_food.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ownedCreates := 0
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
			if !ok || method.Sel.Name != "CreateObjectAt" || len(call.Args) != 3 {
				return true
			}
			unit, ok := call.Args[0].(*ast.SelectorExpr)
			if !ok || unit.Sel.Name != "unit" {
				return true
			}
			owner, ok := call.Args[1].(*ast.SelectorExpr)
			if !ok || owner.Sel.Name != "host" {
				t.Error("food fixture must create the unit through ordinary host ownership")
				return true
			}
			ownedCreates++
			return true
		})
	}
	if ownedCreates != 1 {
		t.Fatalf("ordinary owned unit creations=%d want=1", ownedCreates)
	}
}

func TestE2EAIRetreatFoodIndependentRegeneration(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		start, frame, injury, fps uint32
		maximum, initial, want    uint16
		valid                     bool
	}{
		{"no-tick", 580, 580, 0, 30, 80, 75, 75, true},
		{"troll-before-gate", 580, 603, 0, 30, 80, 75, 75, true},
		{"troll-first-gate", 580, 604, 0, 30, 80, 75, 76, true},
		{"troll-between-gates", 580, 670, 0, 30, 80, 75, 76, true},
		{"troll-second-gate", 580, 671, 0, 30, 80, 75, 77, true},
		{"troll-clamp", 580, 980, 0, 30, 80, 75, 80, true},
		{"npc-before-gate", 580, 612, 0, 30, 150, 145, 145, true},
		{"npc-first-gate", 580, 613, 0, 30, 150, 145, 146, true},
		{"npc-second-gate", 580, 649, 0, 30, 150, 145, 147, true},
		{"script-injury-no-new-pause", 580, 604, 0, 30, 80, 75, 76, true},
		{"ordinary-injury-pause", 580, 604, 580, 30, 80, 75, 75, true},
		{"ordinary-injury-later-gate", 580, 671, 580, 30, 80, 75, 76, true},
		{"last-pause-frame", 576, 604, 573, 30, 80, 75, 75, true},
		{"first-unpaused-gate", 576, 604, 572, 30, 80, 75, 76, true},
		{"alternate-fps", 580, 603, 0, 60, 80, 75, 75, true},
		{"alternate-fps-gate", 580, 676, 0, 60, 80, 75, 76, true},
		{"full-health", 580, 670, 0, 30, 80, 80, 80, true},
		{"multiple-per-frame", 580, 613, 580, 30, 30000, 29790, 29800, true},
		{"multiple-clamp", 580, 613, 580, 30, 30000, 29999, 30000, true},
		{"wrapped-clock", math.MaxUint32 - 1, 1, math.MaxUint32 - 30, 30, 30000, 29990, 29995, true},
		{"bounded-last-frame", 580, 1180, 0, 30, 80, 75, 80, true},
		{"past-bound", 580, 1181, 0, 30, 80, 75, 0, false},
		{"backwards-clock", 580, 579, 0, 30, 80, 75, 0, false},
		{"zero-fps", 580, 603, 0, 0, 80, 75, 0, false},
		{"zero-maximum", 580, 603, 0, 30, 0, 0, 0, false},
		{"invalid-initial", 580, 603, 0, 30, 80, 81, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, valid := e2eAIRetreatFoodRegenerationHP(tc.start, tc.frame, tc.injury, tc.fps, tc.maximum, tc.initial)
			if got != tc.want || valid != tc.valid {
				t.Fatalf("regeneration HP=%d/%t want=%d/%t", got, valid, tc.want, tc.valid)
			}
		})
	}
}
