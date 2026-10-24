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

func TestArmorPickupOriginal53E7F0NativeC(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit ArmorPickup callback probe")
	}
	raw, err := os.ReadFile("testdata/armor_pickup_53e7f0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		Startup   uint32 `json:"actual_startup_control_word"`
		Rows      []struct {
			Name   string `json:"name"`
			Mode   string `json:"mode"`
			Inputs struct {
				OwnerClass   uint32 `json:"owner_class"`
				ItemSubclass uint32 `json:"item_subclass"`
				PickupReturn int    `json:"pickup_return"`
				OldEquipped  bool   `json:"old_equipped"`
				NewEquipped  bool   `json:"new_equipped"`
				ReplaceAt    string `json:"replace_at"`
				State        int    `json:"state"`
			} `json:"inputs"`
			Events     any    `json:"events"`
			Return     int    `json:"return_value"`
			LiveUpdate string `json:"live_update"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.Startup != 0xe3f || len(fixture.Rows) != 56 {
		t.Fatal("incomplete original complete ArmorPickup observations")
	}
	source, err := os.ReadFile("GAME4_3.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "static int nox_modifier_attrs_have_any_53E7F0(const nox_modifier_attrs_t* attrs) {"
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("nonunique actual production function")
	}
	_, body, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(body, "\n//----- (0053EAE0)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("missing complete production function boundary")
	}
	harness, err := os.ReadFile("testdata/armor_pickup_53e7f0_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// PRODUCTION_BODY_53E7F0"
	if bytes.Count(harness, []byte(marker)) != 1 {
		t.Fatal("nonunique native harness body marker")
	}
	program := strings.Replace(string(harness), marker, signature+body, 1)
	dir := t.TempDir()
	input := filepath.Join(dir, "pickup.c")
	binary := filepath.Join(dir, "pickup")
	if err := os.WriteFile(input, []byte(program), 0600); err != nil {
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
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fno-strict-aliasing", "-fno-strict-overflow", "-D_Static_assert(x,y)=", "-I"+include, input, "-o", binary)
	if output, err := exec.Command(compiler[0], args...).CombinedOutput(); err != nil {
		t.Fatalf("compile actual typed ArmorPickup body: %v\n%s", err, output)
	}
	boolean := func(value bool) int {
		if value {
			return 1
		}
		return 0
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		selected++
		t.Run(row.Name, func(t *testing.T) {
			c := row.Inputs
			replaceAt := map[string]int{"none": 0, "pickup": 1, "sameSubclass": 2, "state": 3}[c.ReplaceAt]
			values := []int64{int64(c.OwnerClass), int64(c.ItemSubclass), int64(c.PickupReturn), int64(boolean(c.OldEquipped)), int64(boolean(c.NewEquipped)), int64(replaceAt), int64(c.State)}
			var args []string
			for _, value := range values {
				args = append(args, strconv.FormatInt(value, 10))
			}
			output, err := exec.Command(binary, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("actual native callback: %v\n%s", err, output)
			}
			var got struct {
				PointerBytes int    `json:"pointer_bytes"`
				High         bool   `json:"high_native_pointers"`
				Canaries     bool   `json:"nonoutput_canaries_unchanged"`
				Events       any    `json:"events"`
				Return       int    `json:"return_value"`
				LiveUpdate   string `json:"live_update"`
			}
			if err := json.Unmarshal(output, &got); err != nil {
				t.Fatal(err)
			}
			if got.PointerBytes != 8 || !got.High || !got.Canaries {
				t.Fatal("native pointer/canary contract failed")
			}
			if !reflect.DeepEqual(got.Events, row.Events) || got.Return != row.Return || got.LiveUpdate != row.LiveUpdate {
				t.Fatalf("actual native ArmorPickup trace/result/update=%s; original events=%v return=%d update=%s", output, row.Events, row.Return, row.LiveUpdate)
			}
		})
	}
	if selected != 14 {
		t.Fatalf("startup contexts=%d", selected)
	}
}
