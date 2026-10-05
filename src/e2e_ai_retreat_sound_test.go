package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// BuffApply4FF380 emits spell-audio selector 1 after applying the buff, while
// MonsterActionCast5413B0 is still on the real attack frame. The observer must
// follow that stock definition's on-sound, not the separate cast-sound slot.
func TestE2EAIRetreatObservesStockBuffOnSound(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_retreat.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	assignments := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "configure" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
				return true
			}
			lhs, ok := assign.Lhs[0].(*ast.SelectorExpr)
			if !ok || lhs.Sel.Name != "castSound" {
				return true
			}
			assignments++
			call, ok := assign.Rhs[0].(*ast.CallExpr)
			if !ok || len(call.Args) != 0 {
				t.Fatal("RETREAT audio must come from the stock spell definition")
			}
			getter, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || getter.Sel.Name != "GetOnSound" {
				t.Fatal("RETREAT observer waits for cast-sound instead of the actual BuffApply selector-1 on-sound")
			}
			definition, ok := getter.X.(*ast.CallExpr)
			if !ok || len(definition.Args) != 1 {
				t.Fatal("RETREAT audio must use the selected stock spell")
			}
			lookup, ok := definition.Fun.(*ast.SelectorExpr)
			if !ok || lookup.Sel.Name != "DefByInd" {
				t.Fatal("RETREAT audio bypasses the real spell definition")
			}
			id, ok := definition.Args[0].(*ast.Ident)
			if !ok || id.Name != "e2eAIRetreatSpell" {
				t.Fatal("RETREAT audio observes a different spell")
			}
			return true
		})
	}
	if assignments != 1 {
		t.Fatalf("RETREAT stock buff-on sound assignments=%d, want 1", assignments)
	}
}
