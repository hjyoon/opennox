package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func TestTrapQuickbarVisibility460B90NativeBodies(t *testing.T) {
	fixture, err := os.ReadFile("testdata/trap_quickbar_visibility_460b90_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for _, part := range []struct {
		file, signature, end, marker string
	}{
		{"testdata/retired-game-sources/GAME2.c.txt", "int sub_460B90(int a1) {", "\n}", "// PRODUCTION_VISIBILITY"},
		{"testdata/retired-game-sources/GAME2_1_trap_quickbar.c.txt", "void sub_461010() {", "\n//----- (00461060)", "// PRODUCTION_CLOSE"},
		{"testdata/retired-game-sources/GAME2_1_trap_quickbar.c.txt", "void sub_461060() {", "\n//----- (004611A0)", "// PRODUCTION_OPEN"},
	} {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), part.signature) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatalf("production signature or marker is not unique: %q", part.signature)
		}
		_, body, _ := strings.Cut(string(source), part.signature)
		body, _, ok := strings.Cut(body, part.end)
		if !ok {
			t.Fatalf("production body boundary is missing: %q", part.signature)
		}
		if part.end == "\n}" {
			body += "\n}"
		}
		if !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatalf("production body is not complete: %q", part.signature)
		}
		program = strings.Replace(program, part.marker, part.signature+body, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "trap-quickbar-visibility")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-Itestdata", "-x", "c", "-", "-o", binary)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native Trap Set bodies: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native Trap Set visibility contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}

func TestTrapQuickbarTogglePreservesWindowOrderAndSounds(t *testing.T) {
	g2qTestIsolate(t)
	root := g2qTestObjects(t, 1)
	replacement := g2qTestObjects(t, 1)
	button := g2qTestObjects(t, 1)
	g2qTestOwner(1, g2qTrap(), root, root, root, root)
	g2qTestPointer(5, replacement, true)
	g2qTestPointer(2, button, true)
	type step struct {
		kind string
		win  unsafe.Pointer
		arg  int32
	}
	var trace []step
	game2QuickbarTestHooks.Hidden = func(win unsafe.Pointer, hidden int32) int32 {
		trace = append(trace, step{"hidden", win, hidden})
		return 0
	}
	game2QuickbarTestHooks.Style = func(win unsafe.Pointer, style int32) int32 {
		trace = append(trace, step{"style", win, style})
		return 0
	}
	game2QuickbarTestHooks.Sound = func(id, volume int32) {
		if volume != 100 {
			t.Fatalf("volume = %d", volume)
		}
		trace = append(trace, step{"sound", nil, id})
	}
	sub_461010()
	sub_461060()
	if state := g2qTestWord(1, 0, false); state != 1 {
		t.Fatalf("opened trap state = %d", state)
	}
	sub_461060()
	sub_461010()
	if state := g2qTestWord(1, 0, false); state != 0 {
		t.Fatalf("closed trap state = %d", state)
	}
	want := []step{
		{"hidden", root, 0}, {"hidden", replacement, 1}, {"style", button, 1}, {"sound", nil, 796},
		{"hidden", root, 1}, {"hidden", replacement, 0}, {"style", button, 0}, {"sound", nil, 797},
	}
	if !reflect.DeepEqual(trace, want) {
		t.Fatalf("trap toggle trace = %+v, want %+v", trace, want)
	}
}
