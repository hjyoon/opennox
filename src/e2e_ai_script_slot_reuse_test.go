package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestE2EAIScriptSlotReuseModes(t *testing.T) {
	for _, mode := range []string{"Troll/simple", "Troll/linked", "NPC/simple", "NPC/linked"} {
		kind, linked, ok := e2eAIScriptSlotReuseMode(mode)
		if !ok || !strings.HasPrefix(mode, kind+"/") || linked != strings.HasSuffix(mode, "/linked") {
			t.Fatalf("mode %q = %q/%t/%t", mode, kind, linked, ok)
		}
	}
	for _, mode := range []string{"", "NPC", "Player/simple", "Troll/other", "NPC/linked/extra"} {
		if _, _, ok := e2eAIScriptSlotReuseMode(mode); ok {
			t.Fatalf("invalid mode accepted: %q", mode)
		}
	}
}

func TestE2EAIScriptSlotReuseObserversNeverRepairProductionResults(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_script_slot_reuse.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := make(map[string]int)
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		observer := strings.HasPrefix(fn.Name.Name, "observe")
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					calls[sel.Sel.Name]++
					switch sel.Sel.Name {
					case "MonsterActionRefresh50A910", "SetFrame", "IncFrame", "CallUpdate", "SetHealth", "SetMaxHealth":
						t.Errorf("fixture forces production outcome through %s", sel.Sel.Name)
					}
					if observer && (sel.Sel.Name == "Move" || sel.Sel.Name == "Wander" || sel.Sel.Name == "SetPos" || sel.Sel.Name == "MonsterPushAction" || sel.Sel.Name == "ClearActionStack") {
						t.Errorf("observer forces AI progress through %s", sel.Sel.Name)
					}
				}
			}
			if assign, ok := n.(*ast.AssignStmt); ok {
				for _, lhs := range assign.Lhs {
					if index, ok := lhs.(*ast.IndexExpr); ok {
						if sel, ok := index.X.(*ast.SelectorExpr); ok && (sel.Sel.Name == "Args" || sel.Sel.Name == "AIStack") {
							t.Errorf("fixture repairs raw production slot %s", sel.Sel.Name)
						}
					}
					if sel, ok := lhs.(*ast.SelectorExpr); ok {
						switch sel.Sel.Name {
						case "AIStackInd", "Action", "Field5", "PosVec", "CurrentEnemy", "PreferredEnemy":
							t.Errorf("fixture overwrites production AI result %s", sel.Sel.Name)
						}
					}
				}
			}
			return true
		})
	}
	for _, name := range []string{"MonsterPushAction", "Move", "Wander", "Frame", "ByNetCode", "NewWaypoint", "Screen"} {
		if calls[name] == 0 {
			t.Errorf("missing production input or live observer: %s", name)
		}
	}
}

func TestE2EAIScriptSlotReuseScenarioCoverage(t *testing.T) {
	data, err := os.ReadFile("../scripts/e2e/host-game-ai-script-slot-reuse.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Troll/simple", "Troll/linked", "NPC/simple", "NPC/linked", "action: quit"} {
		if strings.Count(string(data), text) != 1 {
			t.Errorf("scenario must contain exactly one %q", text)
		}
	}
	if strings.Count(string(data), "action: check-ai-script-slot-reuse") != 4 {
		t.Fatal("scenario must cover all four native script-dispatch conditions")
	}
}
