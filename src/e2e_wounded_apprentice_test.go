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

func TestE2EWoundedApprenticePublicScenario(t *testing.T) {
	const path = "../scripts/e2e/solo-wizard-wounded-apprentice.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	checks, openingDone := 0, 0
	for _, step := range file.Steps {
		if step.Action == "check-wounded-apprentice" {
			checks++
			if step.Name == "" {
				t.Fatal("unnamed apprentice lifecycle")
			}
		}
		if step.Action == "click-npc-dialog-done" {
			openingDone++
		}
		if step.Action == "set-map" || step.Action == "call-noxscript-function" || step.Action == "screen" {
			t.Fatalf("public scenario bypass or new golden: %s", step.Action)
		}
	}
	if checks != 1 || openingDone != 1 {
		t.Fatal("missing real solo opening/apprentice lifecycle")
	}
	var sc e2eScenario
	sc.Load(path)
	death, replacement, client := 0, 0, 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " stock dialog death") {
			death++
		}
		if strings.HasSuffix(step.name, " stock corpse replacement") {
			replacement++
		}
		if strings.HasSuffix(step.name, " client corpse") {
			client++
		}
	}
	if death != 1 || replacement != 1 || client != 1 {
		t.Fatal("public Load dispatch omitted apprentice observations")
	}
}

func TestE2EWoundedApprenticeBoundedLifecycle(t *testing.T) {
	var sc e2eScenario
	sc.CheckWoundedApprentice("Wiz01a")
	bounded := 0
	for _, step := range sc.steps {
		if step.ready == nil {
			continue
		}
		bounded++
		if step.waitTimeout != 300 && step.waitTimeout != 1200 || step.fnc == nil {
			t.Fatal("unbounded apprentice observation")
		}
	}
	if bounded != 5 {
		t.Fatalf("bounded observations=%d", bounded)
	}
}

func TestE2EWoundedApprenticeDoesNotSupplyDeath(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_wounded_apprentice.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := map[string]bool{
		"Damage": true, "DoDamage": true, "CallDamage": true, "SetHP": true,
		"NewObjectByTypeID": true, "CreateObjectAt": true, "DelayedDelete": true,
		"ScriptSetDialog": true, "ScriptCancelDialog": true, "SetRaw": true,
		"SetInterp": true, "SetHidden": true, "SetFlags": true, "FlagsEnable": true,
	}
	fields := map[string]bool{
		"HealthData": true, "Cur": true, "Buffs": true, "ObjFlags": true,
		"PosVec": true, "UpdateData": true, "DialogStartFunc": true, "DialogEndFunc": true,
		"DialogWith": true, "AnimInd": true, "Args": true, "Vars": true,
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if _, ok := lhs.(*ast.StarExpr); ok {
					t.Error("observer writes native memory")
				}
				ast.Inspect(lhs, func(n ast.Node) bool {
					if sel, ok := n.(*ast.SelectorExpr); ok && fields[sel.Sel.Name] {
						t.Errorf("observer writes outcome %s", sel.Sel.Name)
					}
					return true
				})
			}
		case *ast.CallExpr:
			if fun, ok := n.Fun.(*ast.SelectorExpr); ok && forbidden[fun.Sel.Name] {
				t.Errorf("observer supplies result via %s", fun.Sel.Name)
			}
			if fun, ok := n.Fun.(*ast.SelectorExpr); ok && fun.Sel.Name == "SetPos" {
				if len(n.Args) != 1 {
					t.Error("unexpected position fixture")
				}
				base, ok := fun.X.(*ast.CallExpr)
				if !ok || len(base.Args) != 1 {
					t.Error("position fixture is not host-only")
					break
				}
				unit, ok := base.Args[0].(*ast.Ident)
				if !ok || unit.Name != "host" {
					t.Error("observer moves NPC/corpse")
				}
			}
			if fun, ok := n.Fun.(*ast.SelectorExpr); ok && fun.Sel.Name == "CallByIndex" {
				if arg, ok := n.Args[0].(*ast.Ident); !ok || arg.Name != "room" {
					t.Error("observer calls death/dialog outcome script directly")
				}
			}
		}
		return true
	})
}
