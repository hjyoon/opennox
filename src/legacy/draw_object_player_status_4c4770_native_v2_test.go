package legacy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Compile and execute the actual production C body with native player records.
// The independent oracle is Nox's local/remote observer and invisibility rule;
// it never reads the poisoned legacy byte as player status.
func TestDrawObjectPlayerStatusNative4C4770V2(t *testing.T) {
	file, err := os.ReadFile("GAME3_1.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(file)
	marker := "void nox_xxx_drawObject_4C4770_draw(nox_draw_viewport_t* vp, nox_drawable* dr, void* img) {"
	start := strings.Index(source, marker)
	if start < 0 || strings.Count(source, marker) != 1 {
		t.Fatal("missing or ambiguous production body")
	}
	end := strings.Index(source[start:], "\n}\n")
	if end < 0 {
		t.Fatal("unterminated production body")
	}
	body := source[start : start+end+3]
	harness, err := os.ReadFile("testdata/draw_object_player_status_4c4770_native_v2_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const insertion = "// PRODUCTION_BODY_INSERTION_POINT"
	if strings.Count(string(harness), insertion) != 1 {
		t.Fatal("ambiguous harness insertion")
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "draw-object-player-status")
	args := append(compiler[1:], "-std=c11", "-fshort-wchar", "-I.", "-x", "c", "-", "-lm", "-o", binary)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(strings.Replace(string(harness), insertion, body, 1))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile unmodified production body: %v\n%s", err, output)
	}
	for _, status := range []uint32{0, 0x10, 0x400, 0x10000010, 1, 0x11, 0x401, 0x80000001} {
		for _, poison := range []byte{0, 1, 0xff} {
			for _, local := range []bool{false, true} {
				for _, moving := range []bool{false, true} {
					for _, invisible := range []bool{false, true} {
						name := fmt.Sprintf("status_%08x_legacy_%02x_local_%t_move_%t_invisible_%t", status, poison, local, moving, invisible)
						t.Run(name, func(t *testing.T) {
							boolArg := func(v bool) string {
								if v {
									return "1"
								}
								return "0"
							}
							output, err := exec.Command(binary, strconv.FormatUint(uint64(status), 10), strconv.Itoa(int(poison)), boolArg(local), boolArg(moving), boolArg(invisible)).CombinedOutput()
							if err != nil {
								t.Fatalf("native production draw: %v\n%s", err, output)
							}
							wantDraws, wantAlpha := 1, 255
							observer := status&1 != 0
							if observer && !local || invisible && !observer && !moving {
								wantDraws, wantAlpha = 0, -1
							} else if observer || invisible {
								wantAlpha = 128
							}
							want := fmt.Sprintf("draws=%d alpha=%d ", wantDraws, wantAlpha)
							if !strings.HasPrefix(string(output), want) {
								t.Fatalf("native observer/opacity rule: want %q, got %s", want, output)
							}
							if !strings.Contains(string(output), "pointer_bytes=") {
								t.Fatal("missing native layout diagnostic")
							}
						})
					}
				}
			}
		}
	}
}
