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

func TestModifierSpeedOriginalNativeC(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit speed modifier context")
	}
	raw, err := os.ReadFile("testdata/modifier_speed_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture map[string]any
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	rows := fixture["rows"].([]any)
	if fixture["original_sha256"] != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture["original_unchanged"] != true || fixture["original_inventory_speed_branch_executed"] != true || fixture["raw_rows"] != float64(272) || len(rows) != 272 {
		t.Fatal("incomplete independent original speed/inventory observations")
	}
	source, err := os.ReadFile("GAME3_2.c")
	if err != nil {
		t.Fatal(err)
	}
	production := string(source)
	block := func(start, end string) string {
		if strings.Count(production, start) != 1 || strings.Count(production, end) != 1 {
			t.Fatal("nonunique complete C body boundaries")
		}
		_, body, _ := strings.Cut(production, start)
		body, _, ok := strings.Cut(body, end)
		if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatal("incomplete actual C body")
		}
		return start + body
	}
	inventory := block("//----- (004DFBB0)", "//----- (004DFC30)")
	engage := block("//----- (004DFC30)", "//----- (004DFCA0)")
	disengage := block("//----- (004DFCA0)", "//----- (004DFD10)")
	const sig = "static float nox_modifierSpeedAddChop32_4DFC5B("
	mathBody := ""
	switch strings.Count(production, sig) {
	case 0: // The same regression executes the unchanged production baseline.
	case 1:
		body := production[strings.Index(production, sig):]
		start := strings.IndexByte(body, '{')
		if start < 0 {
			t.Fatal("missing actual math body")
		}
		depth, end := 0, -1
		for i := start; i < len(body); i++ {
			if body[i] == '{' {
				depth++
			}
			if body[i] == '}' {
				depth--
				if depth == 0 {
					end = i + 1
					break
				}
			}
		}
		if end < 0 {
			t.Fatal("unbalanced actual math body")
		}
		mathBody = body[:end]
	default:
		t.Fatal("nonunique actual math definition")
	}
	harness, err := os.ReadFile("testdata/modifier_speed_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	combined := string(harness)
	for marker, body := range map[string]string{"// PRODUCTION_SPEED_MATH": mathBody, "// PRODUCTION_INVENTORY_4DFBB0": inventory, "// PRODUCTION_SPEED_ENGAGE_4DFC30": engage, "// PRODUCTION_SPEED_DISENGAGE_4DFCA0": disengage} {
		if bytes.Count(harness, []byte(marker)) != 1 {
			t.Fatal("nonunique native marker")
		}
		combined = strings.Replace(combined, marker, body, 1)
	}
	dir := t.TempDir()
	input, binary := filepath.Join(dir, "speed.c"), filepath.Join(dir, "speed")
	if err = os.WriteFile(input, []byte(combined), 0600); err != nil {
		t.Fatal(err)
	}
	include, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	cc := strings.Fields(os.Getenv("CC"))
	if len(cc) == 0 {
		cc = []string{"cc"}
	}
	args := append(cc[1:], "-std=c11", "-O2", "-fshort-wchar", "-fno-strict-aliasing", "-fno-strict-overflow", "-D_Static_assert(x,y)=", "-I"+include, input, "-lm", "-o", binary)
	if out, err := exec.Command(cc[0], nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile actual native speed roots: %v\n%s", err, out)
	}
	for which, function := range []string{"SpeedEngage", "SpeedDisengage"} {
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
					shape := 0
					switch in["owner_shape"] {
					case "player":
					case "null":
						shape = 1
					case "nonplayer":
						shape = 2
					default:
						t.Fatal("unknown original owner shape")
					}
					remaining := 0
					if in["remaining_speed_item"] == true {
						remaining = 1
					}
					number := func(k string) string { return strconv.FormatUint(uint64(in[k].(float64)), 10) }
					out, err := exec.Command(binary, strconv.Itoa(which), number("value_bits"), number("coefficient_bits"), number("initial_flags"), strconv.Itoa(shape), strconv.Itoa(remaining), number("remaining_slot"), number("player_index"), number("unused_disengage_float_bits")).CombinedOutput()
					if err != nil {
						t.Fatalf("actual native speed root: %v\n%s", err, out)
					}
					var got map[string]any
					if err = json.Unmarshal(out, &got); err != nil {
						t.Fatal(err)
					}
					if got["high_native_pointers"] != true || got["nonoutput_canaries_unchanged"] != true || got["rounding_mode_unchanged"] != true || got["flags_word"] != row["flags_word"] {
						t.Fatalf("native pointer/canary/flag/rounding contract: %s", out)
					}
					services := row["services"].([]any)
					count := float64(len(services) / 2)
					if got["reports"] != count || got["sounds"] != count || got["float_getters"] != count {
						t.Fatalf("original service/getter count: %s", out)
					}
					fnGets := float64(0)
					if shape == 0 && which == 1 {
						fnGets = 1
					}
					if got["function_getters"] != fnGets {
						t.Fatalf("native full inventory getter count: %s", out)
					}
					if count != 0 {
						if got["packet_bits"] != got["value_bits"] || got["sound"] != services[1].(map[string]any)["sound"] {
							t.Fatalf("original report/store/sound order: %s", out)
						}
					}
					if got["value_bits"] != row["value_bits"] {
						t.Fatalf("actual native speed=%s; original value bits=%08x", out, uint32(row["value_bits"].(float64)))
					}
				})
			}
			if selected != 34 {
				t.Fatalf("original startup contexts=%d", selected)
			}
		})
	}
}
