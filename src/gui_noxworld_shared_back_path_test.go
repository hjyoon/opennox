package opennox

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// The unchanged shared MainBG Escape handler and C noxworld controller both
// enter sub_4373A0. A live native owner must use its own cleanup before the
// original Back function can dereference the unused legacy window globals.
// This is source-wiring coverage; real-key headless checks verify the route.
func TestNoxWorldSharedBackEntryPath(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "gui_noxworld.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var entry *ast.FuncDecl
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "sub_4373A0" {
			entry = function
			break
		}
	}
	if entry == nil || entry.Body == nil || len(entry.Body.List) < 2 {
		t.Fatal("missing original Nox World Back body")
	}
	compact := func(node ast.Node) string {
		if node == nil {
			return ""
		}
		var output bytes.Buffer
		if err := format.Node(&output, fset, node); err != nil {
			t.Fatal(err)
		}
		return strings.Join(strings.Fields(output.String()), "")
	}
	first, ok := entry.Body.List[0].(*ast.IfStmt)
	t.Run("native owner precedes legacy globals", func(t *testing.T) {
		if !ok || compact(first.Cond) != "noxWorldNative!=nil" {
			t.Error("shared Escape can enter the unused legacy controller while the native browser is live")
		}
	})
	t.Run("native owner cleanup", func(t *testing.T) {
		if !ok || len(first.Body.List) < 1 || compact(first.Body.List[0]) != "closeNoxWorldToMainMenu()" {
			t.Error("shared Back does not close the native discovery/socket/window owner")
		}
	})
	t.Run("native owner return", func(t *testing.T) {
		if !ok || len(first.Body.List) < 2 || compact(first.Body.List[1]) != "return" {
			t.Error("shared Back can continue into unused legacy window globals")
		}
	})
	t.Run("no extra native services", func(t *testing.T) {
		if !ok || first.Init != nil || first.Else != nil || len(first.Body.List) != 2 {
			t.Error("native Back adds services or changes the nil-owner branch")
		}
	})
	t.Run("legacy context remains after native return", func(t *testing.T) {
		if compact(entry.Body.List[1]) != "c:=noxClient" {
			t.Error("native cleanup is not before the unchanged legacy Back context")
		}
	})
	t.Run("legacy info and create branches retained", func(t *testing.T) {
		body := compact(entry.Body)
		for _, original := range []string{
			"legacy.Get_dword_5d4594_815000()",
			"legacy.Get_nox_game_createOrJoin_815048()==1",
			"legacy.Sub_4375C0(1)",
			"c.nox_game_checkStateSwitch_43C1E0()",
		} {
			if !strings.Contains(body, original) {
				t.Errorf("original legacy Back branch lost: %s", original)
			}
		}
	})
}
