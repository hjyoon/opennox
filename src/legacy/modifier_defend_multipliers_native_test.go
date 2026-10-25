package legacy

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestModifierDefendMultipliersOriginalNativeC(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit modifier multiplier probe")
	}
	raw, err := os.ReadFile("testdata/modifier_defend_multipliers_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture map[string]any
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	rows := fixture["rows"].([]any)
	if fixture["original_sha256"] != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		fixture["original_unchanged"] != true || fixture["actual_startup_control_word"] != float64(0xe3f) || len(rows) != 256 {
		t.Fatal("incomplete unchanged original multiplier observations")
	}
	source, err := os.ReadFile("GAME3_2.c")
	if err != nil {
		t.Fatal(err)
	}
	const start = "//----- (004E0370)"
	if strings.Count(string(source), start) != 1 {
		t.Fatal("nonunique production multiplier block")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, ok := strings.Cut(body, "//----- (004E03D0)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") ||
		strings.Count(body, "float* sub_4E0370(") != 1 || strings.Count(body, "float* sub_4E0380(") != 1 {
		t.Fatal("missing complete production multiplier functions")
	}
	body = start + body
	conversion, err := os.ReadFile("GAME1_1.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "float nox_double2float(double a1) {"
	if strings.Count(string(conversion), signature) != 1 {
		t.Fatal("nonunique production conversion helper")
	}
	_, helper, _ := strings.Cut(string(conversion), signature)
	helper, _, ok = strings.Cut(helper, "//----- (00419B10)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(helper), "}") {
		t.Fatal("missing complete production conversion helper")
	}
	harness, err := os.ReadFile("testdata/modifier_defend_multipliers_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"// PRODUCTION_DOUBLE2FLOAT_419AF0", "// PRODUCTION_MULTIPLIERS_4E0370_4E0380"} {
		if bytes.Count(harness, []byte(marker)) != 1 {
			t.Fatal("nonunique native harness marker")
		}
	}
	combined := strings.Replace(string(harness), "// PRODUCTION_DOUBLE2FLOAT_419AF0", signature+helper, 1)
	combined = strings.Replace(combined, "// PRODUCTION_MULTIPLIERS_4E0370_4E0380", body, 1)
	dir := t.TempDir()
	input, binary := filepath.Join(dir, "multipliers.c"), filepath.Join(dir, "multipliers")
	if err := os.WriteFile(input, []byte(combined), 0600); err != nil {
		t.Fatal(err)
	}
	include, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fno-strict-aliasing",
		"-fno-strict-overflow", "-D_Static_assert(x,y)=", "-I"+include, input, "-lm", "-o", binary)
	if output, err := exec.Command(compiler[0], args...).CombinedOutput(); err != nil {
		t.Fatalf("compile actual native multipliers: %v\n%s", err, output)
	}
	for which, function := range []string{"ArmorMultiplierEffect", "DurabilityMultiplierEffect"} {
		t.Run(function, func(t *testing.T) {
			selected := 0
			for _, entry := range rows {
				row := entry.(map[string]any)
				if row["function"] != function || row["mode"] != "actual-startup-inherited" {
					continue
				}
				selected++
				t.Run(row["name"].(string), func(t *testing.T) {
					in := row["inputs"].(map[string]any)
					args := []string{strconv.Itoa(which),
						strconv.FormatUint(uint64(in["value_bits"].(float64)), 10),
						strconv.FormatUint(uint64(in["coefficient_bits"].(float64)), 10)}
					output, err := exec.Command(binary, args...).CombinedOutput()
					if err != nil {
						t.Fatalf("actual native multiplier: %v\n%s", err, output)
					}
					var got map[string]any
					if err := json.Unmarshal(output, &got); err != nil {
						t.Fatal(err)
					}
					if got["high_native_pointers"] != true || got["nonoutput_canaries_unchanged"] != true ||
						got["getter_calls"] != float64(1) || got["return_pointer"] != row["return_pointer"] {
						t.Fatalf("native pointer/canary/getter contract failed: %s", output)
					}
					if got["value_bits"] != row["value_bits"] {
						t.Fatalf("actual native multiplier=%s; original value bits=%08x", output, uint32(row["value_bits"].(float64)))
					}
				})
			}
			if selected != 32 {
				t.Fatalf("original startup contexts=%d", selected)
			}
		})
	}
}
