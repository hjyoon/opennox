package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

// The C leaves already have independent original coverage. Exercise the Go
// adapters actually selected by durability and the exported equipped-item loop.
func TestModifierDefendAdaptersStartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/modifier_defend_multipliers_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		CW        uint16 `json:"actual_startup_control_word"`
		Rows      []struct {
			Function string `json:"function"`
			Name     string `json:"name"`
			Mode     string `json:"mode"`
			Inputs   struct {
				Value       uint32 `json:"value_bits"`
				Coefficient uint32 `json:"coefficient_bits"`
			} `json:"inputs"`
			Value     uint32 `json:"value_bits"`
			Nonoutput bool   `json:"nonoutput_bytes_unchanged"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &data); err != nil {
		t.Fatal(err)
	}
	if data.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !data.Unchanged || data.CW != 0xe3f || len(data.Rows) != 256 {
		t.Fatal("incomplete original multiplier observations")
	}
	fns := playerDamageLateDefendFunctionsNative4E1320()
	selected := 0
	for _, row := range data.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		selected++
		for _, route := range []string{"durability", "equipped-late"} {
			t.Run(fmt.Sprintf("%s/%s/%s", row.Function, route, row.Name), func(t *testing.T) {
				if !row.Nonoutput {
					t.Fatal("original nonoutput bytes changed")
				}
				fn := fns.armorMultiplier
				if row.Function == "DurabilityMultiplierEffect" {
					fn = fns.durabilityMultiplier
				} else if row.Function != "ArmorMultiplierEffect" {
					t.Fatal("unknown original multiplier")
				}
				effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fn, Valf: math.Float32frombits(row.Inputs.Coefficient)}})
				init := selfWeaponElectricNativeAlloc(t, server.ModifierInitData{Modifiers: [4]*server.ModifierEff{nil, nil, effect}})
				item := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)})
				target := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassPlayer, InvFirstItem: item})
				source := selfWeaponElectricNativeAlloc(t, server.Object{})
				weapon := selfWeaponElectricNativeAlloc(t, server.Object{})
				output := selfWeaponElectricNativeAlloc(t, [3]uint32{0xa5a5a5a5, row.Inputs.Value, 0x5a5a5a5a})
				var before [][]byte
				regions := [][]byte{
					unsafe.Slice((*byte)(unsafe.Pointer(effect)), int(unsafe.Sizeof(*effect))),
					unsafe.Slice((*byte)(unsafe.Pointer(init)), int(unsafe.Sizeof(*init))),
					unsafe.Slice((*byte)(unsafe.Pointer(item)), int(unsafe.Sizeof(*item))),
					unsafe.Slice((*byte)(unsafe.Pointer(target)), int(unsafe.Sizeof(*target))),
					unsafe.Slice((*byte)(unsafe.Pointer(source)), int(unsafe.Sizeof(*source))),
					unsafe.Slice((*byte)(unsafe.Pointer(weapon)), int(unsafe.Sizeof(*weapon))),
				}
				for _, region := range regions {
					before = append(before, append([]byte(nil), region...))
					if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&region[0])) <= math.MaxUint32 {
						t.Fatal("expected C-owned records above 4 GiB")
					}
				}
				if route == "durability" {
					if !itemDurabilityApplyDefendNative4E1560(effect, item, target, weapon, source, (*float32)(unsafe.Pointer(&output[1]))) {
						t.Fatal("known durability callback rejected")
					}
				} else if result := Nox_xxx_itemApplyDefendEffect2_4E1320(target, source, weapon, (*int32)(unsafe.Pointer(&output[1])), object.DamageBite); result != 0 {
					t.Fatal("equipped-item loop return changed")
				}
				if output[0] != 0xa5a5a5a5 || output[2] != 0x5a5a5a5a || output[1] != row.Value {
					t.Errorf("native adapter bits=%08x want=%08x canaries=%08x/%08x", output[1], row.Value, output[0], output[2])
				}
				for i, region := range regions {
					if !bytes.Equal(region, before[i]) {
						t.Errorf("native adapter changed nonoutput region %d", i)
					}
				}
			})
		}
	}
	if selected != 64 {
		t.Fatalf("original startup contexts=%d", selected)
	}
}
