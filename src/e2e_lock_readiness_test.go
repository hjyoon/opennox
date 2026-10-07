package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// The world gate must observe the ordinary spawn buff expiry, never remove
// an enchantment or modify a frame to get the fixture past its live guard.
func TestE2ELockPreparationWaitsForNaturalBuffExpiry(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_lock_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var prepare *ast.FuncLit
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "CheckLockSpell" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) != 5 {
				return true
			}
			method, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || method.Sel.Name != "addWhen" {
				return true
			}
			name, ok := call.Args[1].(*ast.BinaryExpr)
			if !ok || name.Op != token.ADD {
				return true
			}
			suffix, ok := name.Y.(*ast.BasicLit)
			if ok && suffix.Value == "\" prepare\"" {
				prepare, _ = call.Args[3].(*ast.FuncLit)
			}
			return true
		})
	}
	if prepare == nil {
		t.Fatal("Lock preparation has no observation callback")
	}
	host := ""
	ast.Inspect(prepare.Body, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
			return true
		}
		name, ok := assignment.Lhs[0].(*ast.Ident)
		call, callOK := assignment.Rhs[0].(*ast.CallExpr)
		if !ok || !callOK {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if ok && method.Sel.Name == "HostUnit" && len(call.Args) == 0 {
			host = name.Name
		}
		return true
	})
	if host == "" {
		t.Fatal("Lock preparation does not cache the actual live host")
	}
	nilGate, buffGate := false, false
	var observe func(ast.Expr)
	observe = func(expression ast.Expr) {
		compare, ok := expression.(*ast.BinaryExpr)
		if !ok {
			return
		}
		if compare.Op == token.LAND {
			observe(compare.X)
			observe(compare.Y)
			return
		}
		if name, ok := compare.X.(*ast.Ident); ok && name.Name == host && compare.Op == token.NEQ {
			zero, ok := compare.Y.(*ast.Ident)
			nilGate = nilGate || ok && zero.Name == "nil"
		}
		if field, ok := compare.X.(*ast.SelectorExpr); ok && field.Sel.Name == "Buffs" && compare.Op == token.EQL {
			name, nameOK := field.X.(*ast.Ident)
			zero, zeroOK := compare.Y.(*ast.BasicLit)
			buffGate = buffGate || nameOK && name.Name == host && zeroOK && zero.Kind == token.INT && zero.Value == "0"
		}
	}
	ast.Inspect(prepare.Body, func(node ast.Node) bool {
		if result, ok := node.(*ast.ReturnStmt); ok && len(result.Results) == 1 {
			observe(result.Results[0])
		}
		if call, ok := node.(*ast.CallExpr); ok {
			if method, ok := call.Fun.(*ast.SelectorExpr); ok {
				switch method.Sel.Name {
				case "DisableEnchant", "EnableEnchant", "SetFrame", "SetOwner", "ResetRandom", "Srand":
					t.Errorf("Lock preparation changes a live outcome: %s", method.Sel.Name)
				}
			}
		}
		return true
	})
	if !nilGate || !buffGate {
		t.Fatal("Lock preparation must wait for a nonnil host and naturally cleared Buffs before applying its unchanged live guard")
	}
}
