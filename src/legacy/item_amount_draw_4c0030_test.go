package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the actual dialog drawer and native viewport/drawable/window layouts.
// Only image, position, and child-window services are intercepted.
func TestItemAmountDraw4C0030NativeBody(t *testing.T) {
	source, err := os.ReadFile("GAME3_1.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "int sub_4C0030(nox_window* win, void* draw) {"
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("production item-amount draw signature is not unique")
	}
	_, body, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(body, "\n//----- (004C01C0)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("production item-amount draw boundary is missing")
	}
	const enumStart = "enum {\n\tNOX_ITEM_AMOUNT_IMAGE_BASE,"
	if strings.Count(string(source), enumStart) != 1 {
		t.Fatal("production item-amount image enumeration is not unique")
	}
	_, enum, _ := strings.Cut(string(source), enumStart)
	enum, _, ok = strings.Cut(enum, "\n};")
	if !ok {
		t.Fatal("production item-amount image enumeration boundary is missing")
	}
	fixture, err := os.ReadFile("testdata/item_amount_draw_4c0030_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for marker, value := range map[string]string{
		"// PRODUCTION_IMAGE_ENUM": enumStart + enum + "\n};",
		"// PRODUCTION_DRAW_BODY": signature + body,
	} {
		if strings.Count(program, marker) != 1 {
			t.Fatalf("native fixture marker %q is not unique", marker)
		}
		program = strings.Replace(program, marker, value, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "item-amount-draw")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-Itestdata", "-x", "c", "-", "-o", binary)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native 004C0030 body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native 004C0030 viewport, icon, and overlay contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}
