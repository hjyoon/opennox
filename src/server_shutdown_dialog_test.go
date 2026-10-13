package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// Dedicated servers still notify remote clients and use the normal shutdown
// acknowledgement/timeout path, but must not load a local graphical dialog.
func TestDedicatedServerShutdownDoesNotLoadLocalDialog(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "gui_quit.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "sub_446190" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("normal server shutdown entry missing")
	}
	var notify, dialog, font, wait token.Pos
	var guardedDialog, guardedFont bool
	ast.Inspect(body, func(node ast.Node) bool {
		if branch, ok := node.(*ast.IfStmt); ok {
			neg, ok := branch.Cond.(*ast.UnaryExpr)
			if ok && neg.Op == token.NOT {
				name, ok := neg.X.(*ast.Ident)
				if ok && name.Name == "isDedicatedServer" {
					ast.Inspect(branch.Body, func(node ast.Node) bool {
						call, ok := node.(*ast.CallExpr)
						if !ok {
							return true
						}
						if fn, ok := call.Fun.(*ast.Ident); ok {
							guardedDialog = guardedDialog || fn.Name == "NewDialogWindow"
							guardedFont = guardedFont || fn.Name == "sub_44A360"
						}
						return true
					})
				}
			}
		}
		if call, ok := node.(*ast.CallExpr); ok {
			if fn, ok := call.Fun.(*ast.SelectorExpr); ok && fn.Sel.Name == "NetSendServerQuit" {
				notify = call.Pos()
			}
			if fn, ok := call.Fun.(*ast.Ident); ok {
				switch fn.Name {
				case "NewDialogWindow":
					dialog = call.Pos()
				case "sub_44A360":
					font = call.Pos()
				}
			}
		}
		if branch, ok := node.(*ast.BranchStmt); ok && branch.Tok == token.GOTO && branch.Label.Name == "LABEL_20" {
			if wait == token.NoPos {
				wait = branch.Pos()
			}
		}
		return true
	})
	if !guardedDialog || !guardedFont {
		t.Fatal("local shutdown dialog and font operation must both exclude dedicated servers")
	}
	if notify == token.NoPos || dialog == token.NoPos || font == token.NoPos || wait == token.NoPos || !(notify < dialog && dialog < font && font < wait) {
		t.Fatal("remote quit notification and the original acknowledgement wait path must be retained")
	}
}
