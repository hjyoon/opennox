package legacy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDoubleSpillOriginal419AF0StartupNativeC(t *testing.T) {
	raw, err := os.ReadFile("testdata/float_conversion_widths_419a90_419b10.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Unchanged   bool   `json:"original_unchanged"`
		Helpers     struct {
			Startup uint32 `json:"actual_startup_control_word"`
			Rows    []struct {
				Helper string `json:"helper"`
				Mode   string `json:"mode"`
				Input  string `json:"input_bits"`
				Result string `json:"result_bits"`
			} `json:"rows"`
		} `json:"helpers"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if !fixture.Unchanged || fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Helpers.Startup != 0xe3f || len(fixture.Helpers.Rows) != 984 {
		t.Fatal("incomplete independent original conversion observations")
	}
	var inputs, expected []string
	excludedNaNs := 0
	for _, row := range fixture.Helpers.Rows {
		if row.Helper != "double-to-binary32-spill-and-reload" || row.Mode != "actual-startup-inherited" {
			continue
		}
		bits, err := strconv.ParseUint(row.Input, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		// Keep all original raw NaN rows unchanged. Payload/sign and x87
		// invalid-operation behavior are a separate fidelity boundary.
		if bits&0x7fffffffffffffff > 0x7ff0000000000000 {
			excludedNaNs++
			continue
		}
		inputs = append(inputs, row.Input)
		expected = append(expected, row.Result)
	}
	if len(inputs) != 41 || excludedNaNs != 4 {
		t.Fatal("unexpected finite/infinity conversion coverage")
	}
	raw, err = os.ReadFile("GAME1_1.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	marker := "float nox_double2float(double a1)"
	if strings.Count(source, marker) != 1 {
		t.Fatal("missing or ambiguous production conversion definition")
	}
	start := strings.Index(source, marker)
	end := start + strings.Index(source[start:], "{") + 1
	for depth := 1; depth != 0; end++ {
		if end >= len(source) {
			t.Fatal("unterminated production conversion definition")
		}
		switch source[end] {
		case '{':
			depth++
		case '}':
			depth--
		}
	}
	const harness = `
#include <stdint.h>
#include <inttypes.h>
#include <stdio.h>
#include <string.h>
#include <math.h>
#include <fenv.h>
PRODUCTION_DEFINITION
int main(void) {
    uint64_t bits;
    float (*volatile convert)(double) = nox_double2float;
    while (scanf("%" SCNx64, &bits) == 1) {
        double input;
        memcpy(&input, &bits, sizeof(input));
        int modes[] = {FE_TONEAREST, FE_TOWARDZERO, FE_DOWNWARD, FE_UPWARD};
        for (int i = 0; i < 4; i++) {
            if (fesetround(modes[i])) return 3;
            float value = convert(input);
            uint32_t result;
            memcpy(&result, &value, sizeof(result));
            if (fegetround() != modes[i]) return 4;
            printf("%d %08" PRIx32 "\n", i, result);
        }
    }
    return ferror(stdin) ? 5 : 0;
}
`
	dir := t.TempDir()
	file := filepath.Join(dir, "double-spill.c")
	if err := os.WriteFile(file, []byte(strings.Replace(harness, "PRODUCTION_DEFINITION", source[start:end], 1)), 0600); err != nil {
		t.Fatal(err)
	}
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "cc"
	}
	binary := filepath.Join(dir, "double-spill")
	if output, err := exec.Command(cc, "-std=c11", "-O2", "-fno-inline", file, "-lm", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile exact production conversion: %v\n%s", err, output)
	}
	command := exec.Command(binary)
	command.Stdin = strings.NewReader(strings.Join(inputs, "\n") + "\n")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("native conversion driver: %v\n%s", err, output)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != len(inputs)*4 {
		t.Fatalf("native observation count=%d", len(lines))
	}
	for mode, name := range []string{"nearest", "toward-zero", "downward", "upward"} {
		t.Run("host-"+name, func(t *testing.T) {
			for i, input := range inputs {
				t.Run(input, func(t *testing.T) {
					want := fmt.Sprintf("%d %s", mode, expected[i])
					if got := lines[i*4+mode]; got != want {
						t.Fatalf("original startup conversion: got %s, want %s", got, want)
					}
				})
			}
		})
	}
}
