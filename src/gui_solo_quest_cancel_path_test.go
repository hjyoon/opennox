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

// GAME.EXE 4379F0 clears both Quest mode words before the retained-window
// branch at 4382CE handles request 2 by sending Back twice. The native owner
// destroys its window on entry to character selection, so cancellation must
// reach its existing cleanup before it can recreate the public server browser.
// Real-key and real-Back headless tests separately verify the runtime route.
func TestSoloQuestCancellationEntryPath(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "gui_noxworld.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var entry *ast.FuncDecl
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "noxGameShowGameSelNative" {
			entry = function
			break
		}
	}
	if entry == nil || entry.Body == nil || len(entry.Body.List) == 0 {
		t.Fatal("missing native Nox World entry body")
	}
	compact := func(node ast.Node) string {
		var output bytes.Buffer
		if err := format.Node(&output, fset, node); err != nil {
			t.Fatal(err)
		}
		return strings.Join(strings.Fields(output.String()), "")
	}
	first, ok := entry.Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatal("entry must decide cancellation before GUI/socket services")
	}
	for _, check := range []struct {
		name string
		run  func(*testing.T)
	}{
		{"original request word and value", func(t *testing.T) {
			if got := compact(first.Cond); got != "memmap.Uint32(0x5D4594,1556104)==2" {
				t.Errorf("entry first condition=%s, want original Quest cancellation request 2", got)
			}
		}},
		{"original Quest host flag cleared", func(t *testing.T) {
			if len(first.Body.List) < 1 || compact(first.Body.List[0]) != "noxServer.nox_xxx_setQuest_4D6F60(0)" {
				t.Error("cancellation does not clear the original Quest host mode word")
			}
		}},
		{"original Quest connection flag cleared", func(t *testing.T) {
			if len(first.Body.List) < 2 || compact(first.Body.List[1]) != "legacy.Sub_4D6F80(0)" {
				t.Error("cancellation does not clear the original Quest connection mode word")
			}
		}},
		{"native cleanup owns return", func(t *testing.T) {
			if len(first.Body.List) < 3 || compact(first.Body.List[2]) != "closeNoxWorldToMainMenu()" {
				t.Error("cancellation bypasses native owner/socket cleanup and main-menu return")
			}
		}},
		{"original success return", func(t *testing.T) {
			if len(first.Body.List) < 4 || compact(first.Body.List[3]) != "return1" {
				t.Error("canceled Solo Quest can fall through into server refresh")
			}
		}},
		{"no extra cancellation services", func(t *testing.T) {
			if first.Init != nil || first.Else != nil || len(first.Body.List) != 4 {
				t.Error("cancellation adds services or changes the non-cancelled entry branch")
			}
		}},
		{"automatic Solo Quest host request retained", func(t *testing.T) {
			found := false
			ast.Inspect(entry.Body, func(node ast.Node) bool {
				branch, ok := node.(*ast.IfStmt)
				if ok && compact(branch.Cond) == "memmap.Uint32(0x5D4594,1556104)==1" {
					found = strings.Contains(compact(branch.Body), "noxWorldNative.host(true)")
				}
				return true
			})
			if !found {
				t.Error("original automatic Solo Quest host request 1 was lost")
			}
		}},
	} {
		t.Run(check.name, check.run)
	}
}
