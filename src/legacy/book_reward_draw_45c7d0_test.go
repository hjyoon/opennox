package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBookRewardDraw45C7D0NativeBody(t *testing.T) {
	book, err := os.ReadFile("client__gui__guibook.c")
	if err != nil {
		t.Fatal(err)
	}
	game, err := os.ReadFile("GAME2.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/book_reward_draw_45c7d0_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for _, part := range []struct {
		source, signature, end, marker string
	}{
		{string(book), "int nox_xxx_bookDrawFn_45C7D0(uint32_t* a1) {", "\n//----- (0045D870)", "// PRODUCTION_DRAW_BODY"},
		{string(game), "unsigned char* nox_quickbar_selected_row(void* base) {", "\nunsigned char* nox_quickbar_trap_selected_row", "// PRODUCTION_SELECTED_ROW"},
		{string(game), "void* nox_xxx_book_45DBE0(void* a1, int a2, int a3) {", "\n//----- (0045DC40)", "// PRODUCTION_BOOK_INSERT"},
	} {
		if strings.Count(part.source, part.signature) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatalf("production signature or marker is not unique: %q", part.signature)
		}
		_, body, _ := strings.Cut(part.source, part.signature)
		body, _, ok := strings.Cut(body, part.end)
		if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatalf("production body boundary is missing: %q", part.signature)
		}
		program = strings.Replace(program, part.marker, part.signature+body, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "book-reward-draw")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-Itestdata", "-x", "c", "-", "-o", binary)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native book reward draw bodies: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native 0045C7D0 animation/default-self contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}
