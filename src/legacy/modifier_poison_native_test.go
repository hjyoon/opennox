package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestModifierPoisonOriginalNativeC(t *testing.T) {
	blob, err := os.ReadFile("testdata/modifier_poison_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type event struct {
		Service        string   `json:"service"`
		Direction      uint32   `json:"direction_bits32"`
		Maximum        uint32   `json:"maximum_bits32"`
		Increment      int32    `json:"increment"`
		TargetPosition []uint32 `json:"target_position_bits32"`
		SourcePosition []uint32 `json:"source_position_bits32"`
		Index          int      `json:"player_index"`
		Value          int      `json:"value"`
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"actual_startup_control_word"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      []struct {
			Input struct {
				Name             string `json:"name"`
				Class            int    `json:"target_class"`
				State            int    `json:"state"`
				Direction        int    `json:"direction"`
				Power            int32  `json:"power"`
				Facing           int    `json:"facing_return"`
				Activate         int    `json:"activate_return"`
				FacingMutation   int    `json:"facing_mutation"`
				ActivateMutation int    `json:"activate_mutation"`
			} `json:"input"`
			Events    []event `json:"events"`
			Preserved bool    `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || len(fixture.Rows) != 240 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete original poison effect fixture")
	}
	extract := func(file, sig string) string {
		t.Helper()
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		if strings.Count(text, sig) != 1 {
			t.Fatal("nonunique production body")
		}
		begin := strings.Index(text, sig)
		end := begin + strings.Index(text[begin:], "{") + 1
		for depth := 1; depth != 0; end++ {
			if end == len(text) {
				t.Fatal("incomplete body")
			}
			if text[end] == '{' {
				depth++
			}
			if text[end] == '}' {
				depth--
			}
		}
		return text[begin:end]
	}
	harness, err := os.ReadFile("testdata/modifier_poison_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	combined := string(harness)
	for _, item := range [][3]string{
		{"PLAYER", "GAME3_2.c", "static nox_playerInfo* nox_xxx_modifierEffectPlayer_4E04D0("},
		{"INFORM", "GAME3_2.c", "static void nox_xxx_modifierEffectInform_4E04D0("},
		{"POISON", "GAME3_2.c", "void nox_xxx_poisonEffect_4E0850("},
	} {
		marker := "// PRODUCTION_" + item[0]
		if strings.Count(combined, marker) != 1 {
			t.Fatal("nonunique fixture marker")
		}
		combined = strings.Replace(combined, marker, extract(item[1], item[2]), 1)
	}
	dir := t.TempDir()
	input, binary := filepath.Join(dir, "poison.c"), filepath.Join(dir, "poison")
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
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-D_Static_assert(x,y)=", "-I"+include, input, "-lm", "-o", binary)
	if out, err := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile exact poison body: %v\n%s", err, out)
	}
	var stdin strings.Builder
	for _, row := range fixture.Rows {
		i := row.Input
		fmt.Fprintf(&stdin, "%d %d %d %d %d %d %d %d\n", i.Class, i.State, i.Direction, i.Power, i.Facing, i.Activate, i.FacingMutation, i.ActivateMutation)
	}
	command := exec.Command(binary)
	command.Stdin = strings.NewReader(stdin.String())
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("native poison effect: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 240 {
		t.Fatalf("native rows=%d", len(lines))
	}
	for i, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			var got struct {
				Bits      int     `json:"native_pointer_bits"`
				Preserved bool    `json:"nonoutput_bytes_preserved"`
				Events    []event `json:"events"`
			}
			if err := json.Unmarshal([]byte(lines[i]), &got); err != nil {
				t.Fatal(err)
			}
			if got.Bits != strconv.IntSize || !got.Preserved || !row.Preserved || !reflect.DeepEqual(got.Events, row.Events) {
				t.Fatalf("native=%s original=%+v", lines[i], row.Events)
			}
		})
	}
}

// Keep the actual four-slot pre-damage entry, typed C PoisonEffect, CGo export,
// poison cap, logic RNG and poison setter. No activation service is replaced.
func TestModifierPoisonActualPreDamage(t *testing.T) {
	for _, slot := range []int{0, 1, 2, 3} {
		for _, row := range []struct {
			name  string
			start uint8
			power int32
			want  uint8
		}{
			{"first", 0, 4, 1},
			{"increment", 3, 4, 4},
			{"at-cap", 4, 4, 4},
			{"above-cap", 5, 4, 5},
			{"zero-cap", 0, 0, 0},
			{"negative-cap", 0, -1, 0},
			{"full-positive-power", 0, math.MaxInt32, 1},
			{"full-negative-power", 0, math.MinInt32, 0},
		} {
			t.Run(fmt.Sprintf("slot-%d/%s", slot, row.name), func(t *testing.T) {
				srv := npcReflectLegacyServer4E17B0(t)
				health := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 77, Max: 80, Field16: 99})
				target := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassMonster, HealthData: health, Poison540: row.start, Field542: 1234})
				source := selfWeaponElectricNativeAlloc(t, server.Object{})
				effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: itemPreDamageFunctionsNative4E13B0().poison, Val: row.power}})
				data := selfWeaponElectricNativeAlloc(t, server.ModifierInitData{})
				data.Modifiers[slot] = effect
				weapon := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(data)})
				index := srv.Rand.Logic.Index()
				if result := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, nil); result != 0 || target.Poison540 != row.want || srv.Rand.Logic.Index() != index+1 {
					t.Fatalf("actual pre-damage result/poison/RNG=%d/%d/%d", result, target.Poison540, srv.Rand.Logic.Index())
				}
				wantFrame, wantTimer := uint32(99), uint16(1234)
				if row.want != row.start {
					wantTimer = 1000
				}
				if row.start == 0 && row.want > 0 {
					wantFrame = 1400
				}
				if health.Cur != 77 || health.Max != 80 || health.Field16 != wantFrame || target.Field542 != wantTimer || effect.AttackPreDmg64.Val != row.power || data.Modifiers[slot] != effect {
					t.Fatal("actual poison setter, cap, frame or nonoutput state changed")
				}
			})
		}
	}
}
