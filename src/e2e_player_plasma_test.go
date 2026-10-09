package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestE2EPlayerPlasmaUsesRealInputWithoutInjectedResults(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_player_plasma.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := make(map[string]int)
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			var name string
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				name = fun.Name
			case *ast.SelectorExpr:
				name = fun.Sel.Name
			}
			calls[name]++
			if name == "CallDamage" || name == "SpellAccept4FD400" || name == "NewRaw" || name == "CancelSpell" || name == "SpellCancelDurSpell4FEB10" || name == "handleDurationRayPacketNative48EA70" || strings.HasPrefix(name, "SpellPlasma") || name == "SetPlayerState" {
				t.Errorf("E2E injects production result through %s", name)
			}
		}
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "data" {
						t.Errorf("E2E overwrites stock wand data.%s", sel.Sel.Name)
					}
				}
			}
		}
		return true
	})
	for _, name := range []string{"e2eQueueInput", "Input", "Screen", "e2ePlayerPlasmaRecord", "e2ePlayerPlasmaRay", "addWhen", "IsEnemyTo"} {
		if calls[name] == 0 {
			t.Errorf("missing real input/outcome observer: %s", name)
		}
	}
}

func TestE2EPlayerPlasmaStockEquipmentAndEndConditions(t *testing.T) {
	data, err := os.ReadFile("../scripts/e2e/host-wizard-plasma.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, required := range []string{"item: OblivionOrb", "action: grant-item", "action: click-inventory-item", "action: check-player-plasma", "action: quit"} {
		if !strings.Contains(s, required) {
			t.Errorf("scenario missing %s", required)
		}
	}
	if strings.Index(s, "action: click-inventory-item") > strings.Index(s, "action: check-player-plasma") {
		t.Fatal("observer precedes actual equipment input")
	}
	data, err = os.ReadFile("e2e_player_plasma.go")
	if err != nil {
		t.Fatal(err)
	}
	s = string(data)
	for _, required := range []string{"data.Charge != 250", "data.MaxCharge != 250", "data.Charge == 0", "data.Flags&4 == 0", "natural stock charge exhaustion", "no damage after natural stop", "durationRayTargets[record] == nil", "plasmaWeapons[record] == nil"} {
		if !strings.Contains(s, required) {
			t.Errorf("missing stock lifecycle check: %s", required)
		}
	}
}
