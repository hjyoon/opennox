package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

// The native player damage marker at 004E1EAE copies the damage-type DWORD.
// It is not the IEEE-754 representation of that damage type. The server's
// source-less poison tests separately exercise the actual marker writes.
func e2ePoisonRequireRawDamageType(t *testing.T, expr ast.Expr) {
	t.Helper()
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		t.Fatal("poison expectation must retain the raw damage-type DWORD")
	}
	conversion, ok := call.Fun.(*ast.Ident)
	if !ok || conversion.Name != "uint32" {
		t.Fatal("poison expectation converts the damage type to something other than uint32")
	}
	typ, ok := call.Args[0].(*ast.SelectorExpr)
	if !ok || typ.Sel.Name != "DamagePoison" {
		t.Fatal("poison expectation must use object.DamagePoison")
	}
	pkg, ok := typ.X.(*ast.Ident)
	if !ok || pkg.Name != "object" {
		t.Fatal("poison expectation must use object.DamagePoison")
	}
	if uint32(object.DamagePoison) != 5 || math.Float32bits(float32(object.DamagePoison)) == 5 {
		t.Fatal("raw poison DWORD must be 5, distinct from its float32 encoding")
	}
}

func TestE2EPlayerPoisonDamageUsesRawDWORD(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	checks := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "AssertPlayerPoisonDamage" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			comparison, ok := node.(*ast.BinaryExpr)
			if !ok || comparison.Op != token.NEQ {
				return true
			}
			field, ok := comparison.X.(*ast.SelectorExpr)
			if ok && field.Sel.Name == "Field75" {
				checks++
				e2ePoisonRequireRawDamageType(t, comparison.Y)
			}
			return true
		})
	}
	if checks != 1 {
		t.Fatalf("player poison marker checks=%d, want exactly one raw DWORD check", checks)
	}
}

func TestE2EPoisonSpellDOTUsesRawDWORD(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_poison_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	checks := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "observeDOT" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, lhs := range assignment.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if ok && ident.Name == "wantType" {
					checks++
					e2ePoisonRequireRawDamageType(t, assignment.Rhs[i])
				}
			}
			return true
		})
	}
	if checks != 2 {
		t.Fatalf("poison spell marker checks=%d, want raw DWORD for player and NPC", checks)
	}
}
