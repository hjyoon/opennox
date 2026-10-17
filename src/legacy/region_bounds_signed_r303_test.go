package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Literal bounds follow the original 0x428170 signed JGE comparisons and its
// asymmetric lower-minimum / upper-maximum clamps, including inverted bounds.
func TestRegionBoundsSignedOriginalR303(t *testing.T) {
	fixture, err := os.ReadFile("testdata/region_bounds_signed_r303_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1_2.c")
	if err != nil {
		t.Fatal(err)
	}
	starts := regexp.MustCompile(`(?m)^int sub_428170\([^;\n]*\) \{`).FindAllIndex(source, -1)
	if len(starts) != 1 {
		t.Fatal("nonunique production sub_428170")
	}
	body, _, ok := strings.Cut(string(source[starts[0][0]:]), "\n}\n")
	if !ok {
		t.Fatal("missing production function end")
	}
	program := strings.Replace(string(fixture), "// PRODUCTION_REGION_BOUNDS_BODY", body+"\n}\n", 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "region-bounds-signed")
	args := append(compiler[1:], "-std=c11", "-O2", "-Werror", "-I.", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual region bounds function: %v\n%s", err, out)
	}
	listed, err := exec.Command(bin, "--list").CombinedOutput()
	if err != nil {
		t.Fatalf("list signed bounds cases: %v\n%s", err, listed)
	}
	names := strings.Fields(string(listed))
	if len(names) != 20 {
		t.Fatalf("signed bounds case count=%d, want 20", len(names))
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			if out, err := exec.Command(bin, name).CombinedOutput(); err != nil {
				t.Fatalf("original signed bounds contract: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}

// Exercise the existing Go/CGo boundary with native C-owned input and output.
// These are the same literal original-derived vectors as the C fixture above.
func TestRegionBoundsSignedCGoR303(t *testing.T) {
	const lo, hi = -1 << 31, 1<<31 - 1
	cases := []struct {
		name      string
		endpoints [4]int32 // x0, y0, x1, y1
		want      [4]int32 // xmin, ymin, xmax, ymax
	}{
		{"positive-ordered", [4]int32{46, 69, 92, 115}, [4]int32{46, 69, 92, 115}},
		{"positive-reversed", [4]int32{92, 115, 46, 69}, [4]int32{46, 69, 92, 115}},
		{"negative-x-mixed", [4]int32{-23, 69, 92, 115}, [4]int32{0, 69, 92, 115}},
		{"negative-y-mixed", [4]int32{46, -23, 92, 69}, [4]int32{46, 0, 92, 69}},
		{"negative-both-mixed", [4]int32{-23, -46, 92, 115}, [4]int32{0, 0, 92, 115}},
		{"negative-mixed-reversed", [4]int32{92, 115, -23, -46}, [4]int32{0, 0, 92, 115}},
		{"entirely-negative", [4]int32{-92, -115, -46, -69}, [4]int32{0, 0, -46, -69}},
		{"negative-reversed", [4]int32{-46, -69, -92, -115}, [4]int32{0, 0, -46, -69}},
		{"equal-zero", [4]int32{}, [4]int32{}},
		{"equal-negative", [4]int32{-23, -46, -23, -46}, [4]int32{0, 0, -23, -46}},
		{"upper-boundary", [4]int32{0, 0, 5888, 5888}, [4]int32{0, 0, 5887, 5887}},
		{"upper-last-valid", [4]int32{0, 0, 5887, 5887}, [4]int32{0, 0, 5887, 5887}},
		{"entirely-above-upper", [4]int32{6000, 7000, 8000, 9000}, [4]int32{6000, 7000, 5887, 5887}},
		{"int-min-max", [4]int32{lo, lo, hi, hi}, [4]int32{0, 0, 5887, 5887}},
		{"int-max-min", [4]int32{hi, hi, lo, lo}, [4]int32{0, 0, 5887, 5887}},
		{"equal-int-min", [4]int32{lo, lo, lo, lo}, [4]int32{0, 0, lo, lo}},
		{"equal-int-max", [4]int32{hi, hi, hi, hi}, [4]int32{hi, hi, 5887, 5887}},
		{"equal-above-upper", [4]int32{5888, 5888, 5888, 5888}, [4]int32{5888, 5888, 5887, 5887}},
		{"int-min-and-upper", [4]int32{lo, 69, 92, hi}, [4]int32{0, 69, 92, 5887}},
		{"opposite-extreme-corners", [4]int32{hi, lo, lo, hi}, [4]int32{0, 0, 5887, 5887}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			type inputBox struct {
				Prefix [4]uint32
				Region [8]int32
				Suffix [4]uint32
			}
			type outputBox struct {
				Prefix [4]uint32
				Bounds [4]int32
				Suffix [4]uint32
			}
			canary := [4]uint32{0xa7b8c9da, 0x12345678, 0x87654321, 0xd9e8f7a6}
			input, freeInput := alloc.New(inputBox{})
			defer freeInput()
			*input = inputBox{
				Prefix: canary,
				Region: [8]int32{lo, tc.endpoints[1], tc.endpoints[0], hi, tc.endpoints[2], -7654321, 0x13579bdf, tc.endpoints[3]},
				Suffix: canary,
			}
			output, freeOutput := alloc.New(outputBox{})
			defer freeOutput()
			*output = outputBox{Prefix: canary, Bounds: [4]int32{-99, -99, -99, -99}, Suffix: canary}
			before := *input
			region, bounds := unsafe.Pointer(&input.Region[0]), unsafe.Pointer(&output.Bounds[0])
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(region) <= 1<<32-1 || uintptr(bounds) <= 1<<32-1) {
				t.Fatalf("native C allocations must exceed 4GiB: region=%p bounds=%p", region, bounds)
			}
			Sub_428170(region, bounds)
			if output.Bounds != tc.want {
				t.Errorf("original signed bounds=%v, want %v", output.Bounds, tc.want)
			}
			if *input != before || output.Prefix != canary || output.Suffix != canary {
				t.Fatal("region, ignored DWORDs or allocation canaries changed")
			}
			t.Logf("R303_CGO_CHECK case=%s region=%p bounds=%p actual=%v", tc.name, region, bounds, output.Bounds)
		})
	}
}

// The tile flag already is uint8_t. These cases preserve that existing BYTE
// behavior when the other three bytes of the original DWORD are nonzero.
func TestFloorMapV3FlagByteR303(t *testing.T) {
	fixture, err := os.ReadFile("testdata/floor_map_v3_flag_byte_r303_test.c")
	if err != nil {
		t.Fatal(err)
	}
	var bodies strings.Builder
	for _, part := range []struct{ file, symbol string }{
		{"GAME1_1.c", "nox_xxx_tile_422C10"},
		{"GAME1_2.c", "sub_428170"},
	} {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		starts := regexp.MustCompile(`(?m)^int `+part.symbol+`\([^;\n]*\) \{`).FindAllIndex(source, -1)
		if len(starts) != 1 {
			t.Fatalf("nonunique production function %s", part.symbol)
		}
		body, _, ok := strings.Cut(string(source[starts[0][0]:]), "\n}\n")
		if !ok {
			t.Fatalf("missing production function end: %s", part.symbol)
		}
		bodies.WriteString(body + "\n}\n")
	}
	program := strings.Replace(string(fixture), "// PRODUCTION_V3_FLAG_BYTE_BODIES", bodies.String(), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "floor-v3-flag-byte")
	args := append(compiler[1:], "-std=c11", "-O2", "-Werror", "-Wno-incompatible-pointer-types", "-I.", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual v3 tile codec: %v\n%s", err, out)
	}
	for _, name := range []string{"empty-upper-bytes", "bounds-upper-bytes", "read-upper-bytes"} {
		t.Run(name, func(t *testing.T) {
			if out, err := exec.Command(bin, name).CombinedOutput(); err != nil {
				t.Fatalf("original BYTE contract: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
