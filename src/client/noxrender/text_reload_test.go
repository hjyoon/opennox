package noxrender

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/noxfont"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestRenderFontsReloadPreservesHandles(t *testing.T) {
	dir := t.TempDir()
	for _, f := range noxFontFiles {
		if err := os.WriteFile(filepath.Join(dir, f.File+".ttf"), goregular.TTF, 0600); err != nil {
			t.Fatal(err)
		}
	}
	oldData := datapath.Data()
	if !datapath.Found() {
		oldData = ""
	}
	datapath.SetData(dir)
	defer datapath.SetData(oldData)
	handles.Init()
	defer handles.Release()
	var r RenderFonts
	defer r.Free()
	if err := r.Load(0); err != nil {
		t.Fatal(err)
	}
	ptrs := make(map[string]unsafe.Pointer)
	metrics := make(map[string]font.Metrics)
	for _, f := range noxFontFiles {
		ptrs[f.Name] = r.FontPtrByName(f.Name)
		metrics[f.Name] = r.FontByName(f.Name).Metrics()
	}
	verify := func() {
		t.Helper()
		if len(r.byPtr) != len(noxFontFiles) {
			t.Fatalf("reload grew the font handle table: %d", len(r.byPtr))
		}
		for _, f := range noxFontFiles {
			ptr := ptrs[f.Name]
			if r.FontPtrByName(f.Name) != ptr || !handles.IsValid(uintptr(ptr)) {
				t.Errorf("%s: reload changed/invalidated its opaque handle", f.Name)
			}
			face := r.AsFont(ptr)
			if face == nil || face != r.FontByName(f.Name) || face.Metrics() != metrics[f.Name] {
				t.Errorf("%s: retained handle did not resolve to its reloaded face", f.Name)
			}
		}
	}
	verify()
	for _, lang := range []int{0, 6, 8, 0} {
		r.Free()
		for _, ptr := range ptrs {
			if r.AsFont(ptr) != nil {
				t.Fatal("freed handle still resolved to a closed face")
			}
		}
		if err := r.Load(lang); err != nil {
			t.Fatal(err)
		}
		verify()
	}
	// A failure after the default face loads must not lose the saved handles
	// for the later font slots. Restoring the generated file permits a retry.
	r.Free()
	large := filepath.Join(dir, noxfont.LargeFile+".ttf")
	if err := os.WriteFile(large, []byte("invalid generated font"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := r.Load(0); err == nil {
		t.Fatal("invalid font unexpectedly loaded")
	}
	r.Free()
	if err := os.WriteFile(large, goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	if err := r.Load(0); err != nil {
		t.Fatal(err)
	}
	verify()
}

func TestRenderFontsReloadRejectsInvalidCachedHandles(t *testing.T) {
	dir := t.TempDir()
	for _, f := range noxFontFiles {
		if err := os.WriteFile(filepath.Join(dir, f.File+".ttf"), goregular.TTF, 0600); err != nil {
			t.Fatal(err)
		}
	}
	oldData := datapath.Data()
	if !datapath.Found() {
		oldData = ""
	}
	datapath.SetData(dir)
	defer datapath.SetData(oldData)
	handles.Init()
	defer handles.Release()
	var r RenderFonts
	defer r.Free()
	if err := r.Load(0); err != nil {
		t.Fatal(err)
	}
	r.Free()
	// Model a cache from a different handle lifetime with real C-owned
	// addresses outside the live arena. Reinitializing the arena alone may
	// reuse its old numeric range and would not reliably test this branch.
	invalid := make(map[unsafe.Pointer]*fontFile)
	for _, f := range r.byPtr {
		ptr, free := alloc.New(byte(0))
		defer free()
		p := unsafe.Pointer(ptr)
		if handles.IsValid(uintptr(p)) {
			t.Fatal("C allocation unexpectedly belongs to the handle arena")
		}
		invalid[p] = f
	}
	r.byPtr = invalid
	if err := r.Load(0); err != nil {
		t.Fatal(err)
	}
	if len(r.byPtr) != len(noxFontFiles) {
		t.Fatalf("reload retained invalid cached handles: %d", len(r.byPtr))
	}
	for ptr := range invalid {
		if r.AsFont(ptr) != nil {
			t.Error("reload reused an invalid cached handle")
		}
	}
	for _, f := range noxFontFiles {
		ptr := r.FontPtrByName(f.Name)
		if !handles.IsValid(uintptr(ptr)) || r.AsFont(ptr) == nil {
			t.Errorf("%s: reload did not allocate a valid replacement handle", f.Name)
		}
	}
}
