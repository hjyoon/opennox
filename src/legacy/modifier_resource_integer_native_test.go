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

// Compile complete production bodies. Expectations are observations of the
// unchanged original executable, with damage/FX/audio/resource services declared.
func TestModifierResourceIntegerOriginalNativeC(t *testing.T) {
	blob, err := os.ReadFile("testdata/modifier_resource_integer_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type event struct {
		Service  string   `json:"service"`
		Amount   uint32   `json:"amount_bits32"`
		Position []uint32 `json:"position_bits32"`
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"actual_startup_control_word"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      []struct {
			Input struct {
				Which     int    `json:"which"`
				Name      string `json:"name"`
				Value     uint32 `json:"coefficient_bits32"`
				Damage    int32  `json:"damage"`
				Available uint32 `json:"available"`
				Mutate    bool   `json:"mutate"`
				NilOwner  bool   `json:"nil_owner"`
				NilTarget bool   `json:"nil_target"`
				Class     uint32 `json:"target_class"`
				Subclass  uint32 `json:"target_subclass"`
			} `json:"input"`
			Mode      string  `json:"mode"`
			Events    []event `json:"events"`
			Preserved bool    `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || len(fixture.Faults) != 0 || len(fixture.Rows) != 240 {
		t.Fatal("incomplete independent original modifier observations")
	}
	extract := func(file, signature string) string {
		t.Helper()
		blob, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(blob)
		if strings.Count(text, signature) != 1 {
			t.Fatalf("nonunique production function %q", signature)
		}
		begin := strings.Index(text, signature)
		end := begin + strings.Index(text[begin:], "{") + 1
		for depth := 1; depth != 0; end++ {
			if end == len(text) {
				t.Fatal("incomplete production body")
			}
			switch text[end] {
			case '{':
				depth++
			case '}':
				depth--
			}
		}
		return text[begin:end]
	}
	harness, err := os.ReadFile("testdata/modifier_resource_integer_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	combined := string(harness)
	for _, item := range [][3]string{
		{"DOUBLE2INT", "GAME1_1.c", "int nox_double2int(double a1)"},
		{"MULTIPLY", "GAME3_2.c", "static double nox_modifierMulChop53_4E0397("},
		{"LIGHTNING", "GAME3_2.c", "void nox_xxx_lightngEffect_4E06F0("},
		{"DRAIN_MANA", "GAME3_2.c", "void nox_xxx_drainMEffect_4E0740("},
		{"VAMPIRISM", "GAME3_2.c", "void nox_xxx_vampirismEffect_4E07C0("},
		{"SYMPATHY", "GAME3_2.c", "void nox_xxx_sympathyEffect_4E08E0("},
	} {
		marker := "// PRODUCTION_" + item[0]
		if strings.Count(combined, marker) != 1 {
			t.Fatal("nonunique native fixture marker")
		}
		combined = strings.Replace(combined, marker, extract(item[1], item[2]), 1)
	}
	dir := t.TempDir()
	input, binary := filepath.Join(dir, "effects.c"), filepath.Join(dir, "effects")
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
		t.Fatalf("compile actual modifier bodies: %v\n%s", err, out)
	}
	var stdin strings.Builder
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		flag := func(v bool) int {
			if v {
				return 1
			}
			return 0
		}
		i := row.Input
		fmt.Fprintf(&stdin, "%d %d %d %d %d %d %d %d %d\n", i.Which, i.Value, i.Damage, i.Available, flag(i.Mutate), flag(i.NilOwner), flag(i.NilTarget), i.Class, i.Subclass)
	}
	command := exec.Command(binary)
	command.Stdin = strings.NewReader(stdin.String())
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("actual modifier execution: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 120 {
		t.Fatalf("native observations = %d, want 120", len(lines))
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		line := lines[selected]
		selected++
		t.Run(row.Input.Name, func(t *testing.T) {
			var got struct {
				Bits      int     `json:"native_pointer_bits"`
				Preserved bool    `json:"nonoutput_bytes_preserved"`
				Events    []event `json:"events"`
			}
			if err := json.Unmarshal([]byte(line), &got); err != nil {
				t.Fatal(err)
			}
			if got.Bits != strconv.IntSize || !got.Preserved || !row.Preserved || !reflect.DeepEqual(got.Events, row.Events) {
				t.Fatalf("native=%s; unchanged original=%+v", line, row.Events)
			}
		})
	}
}

// Keep the public pre-damage dispatch, typed C callback, live HP lookup,
// UnitAdjustHP and real UnitSetHP. No HP or damage handler is substituted.
func TestModifierResourceVampirismActualHP(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	for _, row := range []struct {
		name        string
		coefficient uint32
		damage      int32
		want        uint16
	}{
		{"ordinary", math.Float32bits(.25), 16, 204},
		{"zero-minimum", 0, 1, 201},
		{"qword-low-wrap", math.Float32bits(65536), 65536, 201},
		{"qword-invalid", 0x7f800000, 1, 201},
	} {
		t.Run(row.name, func(t *testing.T) {
			ownerHP := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 200, Max: 500, Field2: 200})
			targetHP := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 200, Max: 200, Field2: 200})
			owner := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassMonster, HealthData: ownerHP})
			target := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassMonster, HealthData: targetHP})
			weapon := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassWeapon})
			effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: itemPreDamageFunctionsNative4E13B0().vampirism, Valf: math.Float32frombits(row.coefficient)}})
			damage := selfWeaponElectricNativeAlloc(t, row.damage)
			if unsafe.Sizeof(uintptr(0)) > 4 {
				for _, ptr := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(target), unsafe.Pointer(weapon), unsafe.Pointer(effect), unsafe.Pointer(damage), unsafe.Pointer(ownerHP), unsafe.Pointer(targetHP)} {
					if uintptr(ptr) <= math.MaxUint32 {
						t.Fatal("expected native C-owned pointers above 4 GiB")
					}
				}
			}
			beforeOwner, beforeTarget, beforeWeapon, beforeHP := *owner, *target, *weapon, *targetHP
			// The real 004E4560 HP setter marks the owner for synchronization.
			beforeOwner.Field38 = math.MaxUint32
			itemPreDamageApplyNative4E13B0(effect, weapon, owner, target, damage)
			if ownerHP.Cur != row.want || *targetHP != beforeHP || *damage != row.damage || *owner != beforeOwner || *target != beforeTarget || *weapon != beforeWeapon {
				t.Fatalf("actual vampirism HP=%d, want %d; target HP=%d raw damage=%d", ownerHP.Cur, row.want, targetHP.Cur, *damage)
			}
			t.Logf("actual pre-damage -> Vampirism -> UnitAdjustHP -> UnitSetHP: coefficient=%08x raw=%d HP=200->%d", row.coefficient, row.damage, ownerHP.Cur)
		})
	}
}
