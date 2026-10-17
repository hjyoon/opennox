package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unsafe"
)

type floorBrushPartR304 struct{ file, symbol string }

// These are local native C functional checks using complete production bodies,
// actual native structs and controlled callbacks. They do not execute GAME.EXE,
// a complete classic mapgen run, or the actual Go/CGo function boundary.
// Original GAME.EXE SHA256:
// 0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a.
// 51D8F0..51D9B6, 51D9C0..51DA6A, 51DA70..51DD4B,
// 51DD50..51DE2E, 51DE30..51DE81, 543680..54371E,
// 5437E0..543BBB, 543BC0..543C4F, 543C50..543E5B,
// 543E60..543EAE and 544310..5443E3 were sealed as complete original bodies.
func floorBrushCompileR304(t *testing.T, group string, parts []floorBrushPartR304) string {
	t.Helper()
	fixture, err := os.ReadFile("testdata/floor_brush_native_r304_test.c")
	if err != nil {
		t.Fatal(err)
	}
	var bodies strings.Builder
	for _, part := range parts {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		// Anchor on a definition, allowing its argument list to span lines.
		// An unanchored name could match a nested call in a condition.
		starts := regexp.MustCompile(`(?m)^(?:int|void)\s+`+regexp.QuoteMeta(part.symbol)+`\([^;{}]*\)\s*\{`).FindAllIndex(source, -1)
		if len(starts) != 1 {
			t.Fatalf("nonunique production function %s", part.symbol)
		}
		body, _, ok := strings.Cut(string(source[starts[0][0]:]), "\n}\n")
		if !ok {
			t.Fatalf("missing production end for %s", part.symbol)
		}
		bodies.WriteString(body + "\n}\n")
	}
	const marker = "// PRODUCTION_FLOOR_BRUSH_BODIES"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("missing or repeated production body marker")
	}
	program := strings.Replace(string(fixture), marker, bodies.String(), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "floor-brush-"+strings.ToLower(group))
	args := append(compiler[1:], "-std=c11", "-O2", "-Werror", "-I.", "-DR304_"+group, "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual floor brush %s functions: %v\n%s", group, err, out)
	}
	return bin
}

func floorBrushRunR304(t *testing.T, bin string, cases []string) {
	t.Helper()
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			if out, err := exec.Command(bin, name).CombinedOutput(); err != nil {
				t.Fatalf("native floor brush contract: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}

func TestFloorBrushNativeR304(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit native pointer regression")
	}
	groups := []struct {
		name  string
		parts []floorBrushPartR304
		cases []string
	}{
		{"CHAIN", []floorBrushPartR304{
			{"GAME4_1.c", "sub_51D8F0"}, {"GAME4_1.c", "sub_51D9C0"}, {"GAME4_2.c", "nox_floorBrushApplyNative_51DA70"},
		}, []string{"descriptor-stack-chain", "descriptor-direct-stack32", "scalar-left", "scalar-right", "dispatch-left", "dispatch-right", "boundary-control", "null-descriptor-control"}},
		{"LAYER", []floorBrushPartR304{
			{"GAME4_3.c", "nox_xxx_tile_543C50"},
		}, []string{"left-head-alignment", "right-head-width", "node-next-append-width", "tail-remove-width", "duplicate-control"}},
		{"FUNCTIONAL", []floorBrushPartR304{
			{"GAME4_1.c", "sub_51D8F0"}, {"GAME4_1.c", "sub_51D9C0"}, {"GAME4_2.c", "nox_floorBrushApplyNative_51DA70"},
			{"GAME4_3.c", "nox_xxx_tile_543C50"}, {"GAME4_3.c", "sub_543E60"},
			{"GAME4_3.c", "nox_xxx_tileSubtile_544310"}, {"GAME4_3.c", "sub_543BC0"},
			{"GAME4_2.c", "sub_51DA70"},
		}, []string{"flag-left", "flag-right", "pattern-negative", "anchor-and-kind", "variation-override", "nil-owned-head", "invalid-layer", "list-sentinels", "merge-append", "merge-map-existing", "merge-unavailable", "merge-coalesce-order", "subtile-producer", "neighbor-writer", "legacy-null-token", "legacy-high-bit-token"}},
		{"FLOOD", []floorBrushPartR304{
			{"GAME4_3.c", "sub_5437E0"}, {"GAME4_2.c", "sub_51DD50"}, {"GAME4_2.c", "sub_51DE30"},
		}, []string{"queue-capacity", "queue-dedup-and-bounds", "queue-pop-width-order", "flood-layer1-order", "flood-layer2-order", "flood-outside-grid"}},
	}
	for _, group := range groups {
		t.Run(strings.ToLower(group.name), func(t *testing.T) {
			floorBrushRunR304(t, floorBrushCompileR304(t, group.name, group.parts), group.cases)
		})
	}
}

// Original 543680 initializes only descriptor offsets0/4/24/28. These two cases
// check the explicit portable definition of the remaining transient state.
// They are not counted as exact original-state equivalence or a stock fix.
func TestFloorBrushDeterminismR304(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit native pointer regression")
	}
	bin := floorBrushCompileR304(t, "DETERMINISM", []floorBrushPartR304{{"GAME4_3.c", "sub_543680"}})
	floorBrushRunR304(t, bin, []string{"flood-brush-defined-state", "flood-brush-skip-and-fix-failure"})
}
