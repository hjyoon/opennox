package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// The client role gates CONNECT_RESULT_OK's thing/renderer initialization.
// A dedicated-server connection must clear Host without clearing Client.
// Actual map entry and input are exercised separately by the real-clock,
// headless client/dedicated-server diagnostic, without E2E timer overrides.
func TestNoxWorldJoinKeepsClientModeForSessionInitialization(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "gui_noxworld.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "beginNoxWorldJoin" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("native join entry missing")
	}
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || fn.Sel.Name != "UnsetGame" {
			return true
		}
		pkg, ok := fn.X.(*ast.Ident)
		if !ok || pkg.Name != "noxflags" {
			return true
		}
		for _, arg := range call.Args {
			ast.Inspect(arg, func(node ast.Node) bool {
				flag, ok := node.(*ast.SelectorExpr)
				if ok && flag.Sel.Name == "GameClient" {
					t.Error("join clears the client role: CONNECT_RESULT_OK skips session initialization before reading ObjectTOC")
				}
				return true
			})
		}
		return true
	})
}
