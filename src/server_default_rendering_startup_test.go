package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// Check the real CLI declaration. A separate real-clock process test loads
// CTF and FlagBall without supplying -serveronly or injecting engine flags.
// The graphical client must retain its false default through the build-tag
// constant, while the dedicated product must skip seat/audio/HUD setup.
func TestServerOnlyCLIDefaultFollowsDedicatedProduct(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	ast.Inspect(f, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || spec.Names[0].Name != "fServer" {
			return true
		}
		count++
		if len(spec.Values) != 1 {
			t.Fatal("server-only CLI declaration must have one initializer")
		}
		call, ok := spec.Values[0].(*ast.CallExpr)
		if !ok || len(call.Args) != 3 {
			t.Fatal("server-only CLI declaration must retain its Bool flag")
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Bool" {
			t.Fatal("server-only CLI declaration no longer uses Bool")
		}
		name, ok := call.Args[0].(*ast.BasicLit)
		if !ok || name.Kind != token.STRING {
			t.Fatal("server-only CLI flag name is not a string")
		}
		value, err := strconv.Unquote(name.Value)
		if err != nil || value != "serveronly" {
			t.Fatal("server-only CLI flag name changed")
		}
		defaultValue, ok := call.Args[1].(*ast.Ident)
		if !ok || defaultValue.Name != "isDedicatedServer" {
			t.Fatal("dedicated product defaults to rendering unless -serveronly is supplied")
		}
		return true
	})
	if count != 1 {
		t.Fatalf("server-only CLI declarations=%d, want 1", count)
	}
}
