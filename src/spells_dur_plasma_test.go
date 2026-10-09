package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func assertPlasmaNativeDispatch(t *testing.T, dispatcher, native string) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "spells_dur.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != dispatcher {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == native {
					found = true
				}
			}
			return true
		})
	}
	if !found {
		t.Fatalf("%s still sends Plasma's native record through the PE32 C callback; missing %s", dispatcher, native)
	}
}

func TestSpellsDurationPlasmaCreateNativeDispatch(t *testing.T) {
	// Fail safely on the unfixed source, before a SIGSEGV or a cgo Go-pointer
	// violation can terminate the process. Then exercise the actual dispatcher.
	assertPlasmaNativeDispatch(t, "callCreate4FEBA0", "SpellPlasmaCreate531580")
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	r := &server.DurSpell{Field72: -1, Field76: 999, Target48: &server.Object{}}
	if got := sp.callCreate4FEBA0(legacy.Get_nox_xxx_plasmaSmth_531580(), r); got != 0 {
		t.Fatal(got)
	}
	if r.Field72 != 0 || r.Field76 != 0 || r.Target48 != nil {
		t.Fatal("native record initialization missing")
	}
}
