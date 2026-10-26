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

func TestArmorEquipOriginal53E650NativeC(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit ArmorEquip callback probe")
	}
	raw, err := os.ReadFile("testdata/armor_equip_53e650_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		Startup   uint32 `json:"actual_startup_control_word"`
		Rows      []struct {
			Name         string         `json:"name"`
			Mode         string         `json:"mode"`
			Inputs       map[string]any `json:"inputs"`
			Events       any            `json:"events"`
			Return       int            `json:"return_value"`
			OldFlags     uint32         `json:"old_player_flags"`
			NewFlags     uint32         `json:"new_player_flags"`
			ItemFlags    uint32         `json:"item_flags"`
			LiveUpdate   string         `json:"live_update"`
			CachedPlayer string         `json:"cached_update_player"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.Startup != 0xe3f || len(fixture.Rows) != 96 {
		t.Fatal("incomplete unchanged original ArmorEquip observations")
	}
	src, err := os.ReadFile("GAME4_3.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "int nox_xxx_playerEquipArmor_53E650(nox_object_t* owner, nox_object_t* item, int a3, int a4) {"
	if strings.Count(string(src), signature) != 1 {
		t.Fatal("nonunique production function")
	}
	_, body, _ := strings.Cut(string(src), signature)
	body, _, ok := strings.Cut(body, "//----- (0053E7B0)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("missing complete production body boundary")
	}
	harness, err := os.ReadFile("testdata/armor_equip_53e650_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// PRODUCTION_BODY_53E650"
	if bytes.Count(harness, []byte(marker)) != 1 {
		t.Fatal("nonunique native harness marker")
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "equip.c")
	binary := filepath.Join(dir, "equip")
	if err := os.WriteFile(input, []byte(strings.Replace(string(harness), marker, signature+body, 1)), 0600); err != nil {
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
	if output, err := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile actual native ArmorEquip: %v\n%s", err, output)
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		selected++
		t.Run(row.Name, func(t *testing.T) {
			var args []string
			for _, key := range []string{"owner_class", "item_class", "item_flags", "item_subclass", "class_return", "strength_return", "equipped", "equip_flags", "special_return", "npc_return", "play_sound"} {
				var n int64
				switch v := row.Inputs[key].(type) {
				case float64:
					n = int64(v)
				case bool:
					if v {
						n = 1
					}
				default:
					t.Fatalf("unexpected fixture input %s=%v", key, v)
				}
				args = append(args, strconv.FormatInt(n, 10))
			}
			stages := map[string]int{"none": 0, "sameSubclass": 1, "classUse": 2, "strength": 3, "dequip": 4, "armorFlags": 5, "report": 6}
			kinds := map[string]int{"none": 0, "update": 1, "player": 2}
			stage, ok := stages[row.Inputs["replace_at"].(string)]
			if !ok {
				t.Fatal("unknown replacement stage")
			}
			kind, ok := kinds[row.Inputs["replace_kind"].(string)]
			if !ok {
				t.Fatal("unknown replacement kind")
			}
			args = append(args, strconv.Itoa(stage), strconv.Itoa(kind))
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
				OldFlags     uint32 `json:"old_player_flags"`
				NewFlags     uint32 `json:"new_player_flags"`
				ItemFlags    uint32 `json:"item_flags"`
				LiveUpdate   string `json:"live_update"`
				CachedPlayer string `json:"cached_update_player"`
			}
			if err := json.Unmarshal(output, &got); err != nil {
				t.Fatal(err)
			}
			if got.PointerBytes != 8 || !got.High || !got.Canaries {
				t.Fatal("native pointer/canary contract failed")
			}
			if !reflect.DeepEqual(got.Events, row.Events) || got.Return != row.Return || got.OldFlags != row.OldFlags || got.NewFlags != row.NewFlags || got.ItemFlags != row.ItemFlags || got.LiveUpdate != row.LiveUpdate || got.CachedPlayer != row.CachedPlayer {
				t.Fatalf("actual native ArmorEquip=%s; original old/new flags=%#x/%#x events=%v", output, row.OldFlags, row.NewFlags, row.Events)
			}
		})
	}
	if selected != 24 {
		t.Fatalf("original startup contexts=%d", selected)
	}
}
