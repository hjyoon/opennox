package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestE2EMeteorShowerDurableHostSetup(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_meteor_shower_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	host, target := 0, 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			set, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || set.Sel.Name != "SetMaxHealth" {
				return true
			}
			if fn.Name.Name != "prepare" || len(call.Args) != 1 {
				t.Fatal("MeteorShower changes durability after setup")
			}
			amount, ok := call.Args[0].(*ast.BasicLit)
			if !ok || amount.Kind != token.INT || amount.Value != "30000" {
				t.Fatal("MeteorShower durability does not cover the complete natural shower")
			}
			adapter, ok := set.X.(*ast.CallExpr)
			if !ok || len(adapter.Args) != 1 {
				t.Fatal("MeteorShower durability lacks the ordinary object adapter")
			}
			unit, ok := adapter.Args[0].(*ast.SelectorExpr)
			if !ok {
				t.Fatal("MeteorShower durability does not target a fixture unit")
			}
			switch unit.Sel.Name {
			case "host":
				host++
			case "target":
				target++
			default:
				t.Fatalf("unexpected durable fixture unit %s", unit.Sel.Name)
			}
			return true
		})
	}
	if host != 1 || target != 1 {
		t.Fatalf("durable setup host/target=%d/%d, want 1/1; real self-explosion can kill the host", host, target)
	}
}
