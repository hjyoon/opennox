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

const floorBrushReferenceR304 = "testdata/floor_brush_native_r304_reference.c"

// These are local native C functional checks using complete compatibility
// bodies, actual native structs and controlled callbacks. They do not execute GAME.EXE,
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
			t.Fatalf("nonunique fixture function %s in %s", part.symbol, part.file)
		}
		body, _, ok := strings.Cut(string(source[starts[0][0]:]), "\n}\n")
		if !ok {
			t.Fatalf("missing fixture function end for %s", part.symbol)
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
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
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
			{floorBrushReferenceR304, "nox_floorBrushApplyNative_51DA70"},
		}, []string{"descriptor-direct-stack32", "scalar-left", "scalar-right", "dispatch-left", "dispatch-right"}},
		{"LAYER", []floorBrushPartR304{
			{floorBrushReferenceR304, "nox_xxx_tile_543C50"},
		}, []string{"left-head-alignment", "right-head-width", "node-next-append-width", "tail-remove-width", "duplicate-control"}},
		{"FUNCTIONAL", []floorBrushPartR304{
			{floorBrushReferenceR304, "nox_floorBrushApplyNative_51DA70"},
			{floorBrushReferenceR304, "nox_xxx_tile_543C50"}, {floorBrushReferenceR304, "sub_543E60"},
			{"GAME4_3.c", "nox_xxx_tileSubtile_544310"}, {"GAME4_3.c", "sub_543BC0"},
			{floorBrushReferenceR304, "sub_51DA70"},
		}, []string{"flag-left", "flag-right", "pattern-negative", "anchor-and-kind", "variation-override", "nil-owned-head", "invalid-layer", "list-sentinels", "merge-append", "merge-map-existing", "merge-unavailable", "merge-coalesce-order", "subtile-producer", "neighbor-writer", "legacy-null-token", "legacy-high-bit-token"}},
		{"FLOOD", []floorBrushPartR304{
			{"GAME4_3.c", "sub_5437E0"}, {"GAME4_2.c", "sub_51DD50"}, {"GAME4_2.c", "sub_51DE30"},
		}, []string{"queue-capacity", "queue-dedup-and-bounds", "queue-pop-width-order", "flood-layer1-order", "flood-layer2-order", "flood-outside-grid"}},
	}
	for _, group := range groups {
		t.Run(strings.ToLower(group.name), func(t *testing.T) {
			floorBrushRunR304(t, floorBrushCompileR304(t, group.name, group.parts), group.cases)
			if group.name == "CHAIN" {
				testFloorBrushCoordinateChainR304(t)
			}
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

func testFloorBrushCoordinateChainR304(t *testing.T) {
	point := mapgenWorldPoint{x: 448.5, y: 494.5}
	t.Run("descriptor-stack-chain", func(t *testing.T) {
		want := mapgenFloorBrush{tileID: 7, variation: 3, anchorX: -1, anchorY: -1, edgeID: 255}
		place := func(x, y, dx, dy int32, brush *mapgenFloorBrush) int32 {
			return mapgenApplyTileQuarter(x, y, dx, dy, brush, func(col, row int32, got *mapgenFloorBrush, layer int32) int32 {
				if col != 9 || row != 11 || layer != 1 || got == nil || *got != want {
					t.Fatalf("floor brush advanced to %d/%d layer %d with %+v, expected 9/11 layer 1 with %+v",
						col, row, layer, got, want)
				}
				return 1
			})
		}
		if status := mapgenReadTileCoordinate(&point, 7, 3, place); status != 1 {
			t.Fatalf("brush placement returned %d, want 1", status)
		}
	})
	t.Run("boundary-control", func(t *testing.T) {
		apply := func(int32, int32, *mapgenFloorBrush, int32) int32 {
			t.Fatal("the rejected boundary reached the floor brush")
			return -1
		}
		if status := mapgenApplyTileQuarter(1, 10, 0, 0, nil, apply); status != 0 {
			t.Fatalf("boundary returned %d, want 0", status)
		}
	})
	t.Run("null-descriptor-control", func(t *testing.T) {
		calls := 0
		place := func(x, y, dx, dy int32, brush *mapgenFloorBrush) int32 {
			return mapgenApplyTileQuarter(x, y, dx, dy, brush, func(col, row int32, got *mapgenFloorBrush, layer int32) int32 {
				calls++
				if col != 9 || row != 11 || layer != 1 || got != nil {
					t.Fatalf("nil descriptor advanced to %d/%d layer %d with %+v", col, row, layer, got)
				}
				return 69
			})
		}
		if status := mapgenReadTileCoordinate(&point, 255, 3, place); status != 69 || calls != 1 {
			t.Fatalf("nil descriptor returned %d with %d calls", status, calls)
		}
	})
}
