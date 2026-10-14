package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the actual death-message and packet-ring bodies against native
// Player headers and the production UTF-16 formatter. Lookup/localization/
// console are controlled services; this is not a full network or GUI test.
func TestClientDeathMessage4952E0NativeBodies(t *testing.T) {
	fixture, err := os.ReadFile("testdata/client_death_message_4952e0_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for _, part := range []struct {
		file, start, end, marker string
	}{
		{"client__network__deathmsg.c", "int sub_4952E0(uint16_t* a1) {", "", "// PRODUCTION_DEATH_MESSAGE"},
		{"GAME2_3.c", "static uint16_t nox_deathmsg_read_u16(const uint8_t* data) {", "//----- (00495430)", "// PRODUCTION_DEATH_PACKET"},
	} {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), part.start) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatalf("ambiguous production boundary: %s", part.start)
		}
		_, body, _ := strings.Cut(string(source), part.start)
		if part.end != "" {
			var found bool
			body, _, found = strings.Cut(body, part.end)
			if !found {
				t.Fatalf("missing production end: %s", part.end)
			}
		}
		if !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatalf("incomplete production body: %s", part.file)
		}
		program = strings.Replace(program, part.marker, part.start+body, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-death-message")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char",
		"-Wall", "-Wextra", "-Werror", "-Wno-sign-compare", "-Itestdata",
		"-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile production death-message bodies: %v\n%s", err, output)
	}
	if output, err := exec.Command(bin).CombinedOutput(); err != nil {
		t.Fatalf("native death-message contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}
