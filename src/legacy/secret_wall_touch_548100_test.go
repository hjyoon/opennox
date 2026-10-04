package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the actual production body and wall layout. Only lookup and audio
// services are intercepted; the secret record comes from the production header.
func TestSecretWallTouch548100NativeBody(t *testing.T) {
	source, err := os.ReadFile("GAME5.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "void sub_548100(int2* a1, nox_object_t* a2) {"
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("production secret-wall touch signature is not unique")
	}
	_, body, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(body, "\n//----- (005481C0)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("production secret-wall touch boundary is missing")
	}
	const wallSignature = "typedef struct nox_wall_native_t {"
	if strings.Count(string(source), wallSignature) != 1 {
		t.Fatal("production native wall layout is not unique")
	}
	_, layout, _ := strings.Cut(string(source), wallSignature)
	layout, _, ok = strings.Cut(layout, "\nstatic nox_collision_hit_t*")
	if !ok || !strings.Contains(layout, "wrong native wall client-data offset") {
		t.Fatal("production native wall layout boundary is missing")
	}
	fixture, err := os.ReadFile("testdata/secret_wall_touch_548100_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for marker, value := range map[string]string{
		"// PRODUCTION_WALL_LAYOUT": wallSignature + layout,
		"// PRODUCTION_BODY_548100": signature + body,
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
	binary := filepath.Join(t.TempDir(), "secret-wall-touch")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-Itestdata", "-x", "c", "-", "-o", binary)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native 00548100 body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native 00548100 touch, audio, and field contract: %v\n%s", err, output)
	}
}
