package legacy

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestArmorRecalculateOriginal53E300NativeC(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit armor sum probe")
	}
	raw, err := os.ReadFile("testdata/armor_recalculate_53e300_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type item struct {
		Class      uint32 `json:"obj_class"`
		Flags      uint32 `json:"flags"`
		ArmorBits  uint32 `json:"armor_bits"`
		Definition bool   `json:"definition"`
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		Startup   uint32 `json:"actual_startup_control_word"`
		Rows      []struct {
			Name   string `json:"name"`
			Mode   string `json:"mode"`
			Inputs struct {
				Class uint32 `json:"owner_class"`
				Items []item `json:"items"`
			} `json:"inputs"`
			Events      any    `json:"events"`
			PlayerBits  uint32 `json:"player_armor_bits"`
			MonsterBits uint32 `json:"monster_armor_bits"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		!fixture.Unchanged || fixture.Startup != 0xe3f || len(fixture.Rows) != 128 {
		t.Fatal("incomplete unchanged original armor sum observations")
	}
	extract := func(file, signature, next string) string {
		t.Helper()
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(src), signature) != 1 {
			t.Fatalf("nonunique production function in %s", file)
		}
		_, body, _ := strings.Cut(string(src), signature)
		body, _, ok := strings.Cut(body, next)
		if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatalf("missing complete production body boundary in %s", file)
		}
		return signature + body
	}
	body := extract("GAME4_3.c", "int nox_xxx_recalculateArmorVal_53E300(nox_object_t* owner) {", "//----- (0053E3A0)")
	helper := extract("GAME1_1.c", "float nox_double2float(double a1) {", "//----- (00419B10)")
	harness, err := os.ReadFile("testdata/armor_recalculate_53e300_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"// PRODUCTION_BODY_53E300", "// PRODUCTION_DOUBLE2FLOAT_419AF0"} {
		if bytes.Count(harness, []byte(marker)) != 1 {
			t.Fatal("nonunique native harness marker")
		}
	}
	dir := t.TempDir()
	input, binary := filepath.Join(dir, "armor.c"), filepath.Join(dir, "armor")
	combined := strings.Replace(string(harness), "// PRODUCTION_BODY_53E300", body, 1)
	combined = strings.Replace(combined, "// PRODUCTION_DOUBLE2FLOAT_419AF0", helper, 1)
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
	if output, err := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile actual native armor sum: %v\n%s", err, output)
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		selected++
		t.Run(row.Name, func(t *testing.T) {
			args := []string{strconv.FormatUint(uint64(row.Inputs.Class), 10), strconv.Itoa(len(row.Inputs.Items))}
			for _, value := range row.Inputs.Items {
				present := "0"
				if value.Definition {
					present = "1"
				}
				args = append(args, strconv.FormatUint(uint64(value.Class), 10),
					strconv.FormatUint(uint64(value.Flags), 10), strconv.FormatUint(uint64(value.ArmorBits), 10), present)
			}
			output, err := exec.Command(binary, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("actual native armor sum: %v\n%s", err, output)
			}
			var got struct {
				PointerBytes int    `json:"pointer_bytes"`
				High         bool   `json:"high_native_pointers"`
				Canaries     bool   `json:"nonoutput_canaries_unchanged"`
				Events       any    `json:"events"`
				PlayerBits   uint32 `json:"player_armor_bits"`
				MonsterBits  uint32 `json:"monster_armor_bits"`
			}
			if err := json.Unmarshal(output, &got); err != nil {
				t.Fatal(err)
			}
			if got.PointerBytes != 8 || !got.High || !got.Canaries {
				t.Fatal("native pointer/canary contract failed")
			}
			if !reflect.DeepEqual(got.Events, row.Events) || got.PlayerBits != row.PlayerBits || got.MonsterBits != row.MonsterBits {
				t.Fatalf("native armor sum=%s; original player=%08x monster=%08x events=%v",
					output, row.PlayerBits, row.MonsterBits, row.Events)
			}
		})
	}
	if selected != 32 {
		t.Fatalf("original startup contexts=%d", selected)
	}
}
