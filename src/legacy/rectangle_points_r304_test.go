package legacy

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// GAME.EXE 5245A0/524610 pass contiguous stack x/y fields and use FADD
// binary32 operands 583B9C=41821B8F and 583B94=42021B8F before each store.
// The literal negative/tie vectors distinguish those operands from truncated
// decimal double literals. Return checks preserve the existing C declaration;
// they are not a claim about an original void-return calling convention.
func TestFloorBrushProducerPointsOriginalR304(t *testing.T) {
	fixture, err := os.ReadFile("testdata/rectangle_points_r304_test.c")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	var bodies strings.Builder
	for _, name := range []string{"sub_5245A0", "sub_524610"} {
		starts := regexp.MustCompile(`(?m)^float\* `+name+`\([^;\n]*\) \{`).FindAllIndex(data, -1)
		if len(starts) != 1 {
			t.Fatalf("nonunique production definition %s", name)
		}
		body, _, ok := strings.Cut(string(data[starts[0][0]:]), "\n}\n")
		if !ok {
			t.Fatalf("missing production end for %s", name)
		}
		bodies.WriteString(body + "\n}\n")
	}
	const marker = "// PRODUCTION_RECTANGLE_POINT_BODIES"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("nonunique fixture insertion marker")
	}
	program := strings.Replace(string(fixture), marker, bodies.String(), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "rectangle-points")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-inline", "-Werror", "-Wno-int-to-pointer-cast", "-x", "c", "-", "-o", bin)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual floor coordinate producers: %v\n%s", err, out)
	}
	for _, name := range []string{"rectangle-two-by-two", "column-three", "rectangle-zero-width", "rectangle-zero-height", "rectangle-negative-width", "rectangle-negative-height", "column-zero", "column-negative"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if out, err := exec.CommandContext(ctx, bin, name).CombinedOutput(); err != nil {
				t.Fatalf("floor coordinate callbacks: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
