package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"unsafe"

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

func TestSpellsDurationPlasmaUpdateNativeDispatch(t *testing.T) {
	assertPlasmaNativeDispatch(t, "callUpdate4FEEF0", "SpellPlasmaUpdate531600")
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	r := &server.DurSpell{Field36: 0xabcdef, Field72: 0x123456, Field76: 0x87654321}
	if got := sp.callUpdate4FEEF0(legacy.Get_nox_xxx_plasmaShot_531600(), r); got != 1 {
		t.Fatal(got)
	}
	if r.Field36 != 0xabcdef || r.Field72 != 0x123456 || r.Field76 != 0x87654321 {
		t.Fatal("orphan cancellation changed PE32 scalar slots")
	}
}

func TestSpellsDurationPlasmaDestroyNativeDispatch(t *testing.T) {
	assertPlasmaNativeDispatch(t, "callDestroy4FEDA0", "SpellPlasmaDestroy5319E0")
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	r := &server.DurSpell{Field72: 0x13579}
	data := &server.WandUseData{Flags: 0x87654324}
	wand := &server.Object{}
	wand.UseData.Ptr = unsafe.Pointer(data)
	rt := sp.plasmaRuntime531580()
	rt.StoreWeapon(r, wand)
	sp.callDestroy4FEDA0(legacy.Get_sub_5319E0(), r)
	if rt.LoadWeapon(r) != nil || data.Flags != 0x87654320 || r.Field72 != 0x13579 {
		t.Fatal("native destroy lost flags or retained wand")
	}
}

func TestSpellsDurationPlasmaFreeClearsNativeSidecars(t *testing.T) {
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	r, wand, target := &server.DurSpell{}, &server.Object{}, &server.Object{}
	rt := sp.plasmaRuntime531580()
	rt.StoreWeapon(r, wand)
	rt.StoreRayTarget(r, target)
	if rt.LoadWeapon(r) != wand || rt.LoadRayTarget(r) != target {
		t.Fatal("native pointers lost")
	}
	sp.Free()
	if rt.LoadWeapon(r) != nil || rt.LoadRayTarget(r) != nil {
		t.Fatal("Free retained Plasma sidecars")
	}
}
