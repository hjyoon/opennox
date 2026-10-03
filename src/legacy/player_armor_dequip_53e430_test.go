package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the unchanged production body, not a second implementation, against
// the real native-width structs. Only its external services are intercepted.
func TestPlayerArmorDequip53E430NativeBody(t *testing.T) {
	source, err := os.ReadFile("GAME4_3.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "int sub_53E430(nox_object_t* owner, nox_object_t* item, int a3, int a4) {"
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("production armor dequip signature is not unique")
	}
	_, body, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(body, "\n//----- (0053E520)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("production armor dequip boundary is missing")
	}
	fixture, err := os.ReadFile("testdata/player_armor_dequip_53e430_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// PRODUCTION_BODY_53E430"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("native fixture body marker is not unique")
	}
	program := strings.Replace(string(fixture), marker, signature+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "player-armor-dequip")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-Itestdata", "-x", "c", "-", "-o", binary)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native 0053E430 body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native 0053E430 callback and field contract: %v\n%s", err, output)
	}
}
