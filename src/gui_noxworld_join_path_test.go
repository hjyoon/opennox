package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// The native browser releases its windows before character selection. Unlike
// the old retained C browser, it must explicitly retain the join frame owner.
func TestNoxWorldJoinRetainsFrameOwnerBeforePlayerSelection(t *testing.T) {
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
	var calls []string
	ast.Inspect(body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if fn, ok := call.Fun.(*ast.Ident); ok {
				calls = append(calls, fn.Name)
			}
		}
		return true
	})
	index := func(name string) int {
		for i, call := range calls {
			if call == name {
				return i
			}
		}
		return -1
	}
	leave, arm, selectPlayer := index("leaveNoxWorldForPlayerSelection"), index("startNoxWorldJoin"), index("showNoxWorldPlayerSelection")
	if leave < 0 || arm <= leave || selectPlayer <= arm {
		t.Fatalf("join loses its frame owner after browser cleanup: calls=%v", calls)
	}
}
