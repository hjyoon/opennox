package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestE2ETelekinesisNPCReleasePrecedesImpactObservation(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_telekinesis_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var observer *ast.FuncDecl
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok && fn.Name.Name == "observeSound" {
			observer = fn
		}
	}
	if observer == nil {
		t.Fatal("missing live sound observer")
	}
	var release *ast.IfStmt
	for _, statement := range observer.Body.List {
		branch, ok := statement.(*ast.IfStmt)
		if !ok {
			continue
		}
		ast.Inspect(branch.Cond, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "GetCastSound" {
				release = branch
			}
			return true
		})
	}
	if release == nil {
		t.Fatal("release must be observed at the spell's actual cast sound")
	}
	seen := make(map[string]bool)
	naturalWrites := 0
	ast.Inspect(release.Body, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			seen[sel.Sel.Name] = true
		}
		if assignment, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignment.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name == "natural" {
					naturalWrites++
				}
			}
		}
		return true
	})
	for _, required := range []string{"AIStackHead", "ArgObj", "ArgU32", "Field120_1", "Field120_2", "MissileAttackFrame216", "UpdateDataSpellProjectile", "Target", "Field0", "Field8", "Spell12", "Level16", "ObjOwner", "Field32"} {
		if !seen[required] {
			t.Errorf("release omitted live animation/projectile observation %s", required)
		}
	}
	if naturalWrites != 1 || seen["observeCreation"] {
		t.Fatal("release must mark one observed animation, without supplying the impact effect")
	}
	allNaturalWrites, guardedCreation := 0, false
	ast.Inspect(observer.Body, func(n ast.Node) bool {
		if assignment, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignment.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name == "natural" {
					allNaturalWrites++
				}
			}
		}
		if block, ok := n.(*ast.BlockStmt); ok {
			guard := false
			for _, statement := range block.List {
				if branch, ok := statement.(*ast.IfStmt); ok {
					if condition, ok := branch.Cond.(*ast.UnaryExpr); ok && condition.Op == token.NOT {
						if sel, ok := condition.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "natural" {
							guard = true
							if _, ok := branch.Body.List[len(branch.Body.List)-1].(*ast.ReturnStmt); !ok {
								t.Error("unobserved release does not stop impact observation")
							}
						}
					}
				}
				if guard {
					ast.Inspect(statement, func(part ast.Node) bool {
						if sel, ok := part.(*ast.SelectorExpr); ok && sel.Sel.Name == "observeCreation" {
							guardedCreation = true
						}
						return true
					})
				}
			}
		}
		return true
	})
	if allNaturalWrites != 1 || !guardedCreation {
		t.Fatal("impact must require the previously observed real release")
	}
}
