package legacy

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Replay original damage, FX, audio, and resource service calls through
// the native modifier implementations.
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
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			coeff := math.Float32frombits(in.Value)
			effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{
				AttackPreHit52: server.ModifierEffFnc{Valf: coeff},
				AttackPreDmg64: server.ModifierEffFnc{Valf: coeff},
			})
			target := selfWeaponElectricNativeAlloc(t, server.Object{
				ObjClass: object.Class(in.Class), ObjSubClass: object.SubClass(in.Subclass),
				PosVec: types.Pointf{X: 12.75, Y: 19.25},
				Damage: objectDamageNativeProbePtr(), NetCode: 301,
			})
			owner := selfWeaponElectricNativeAlloc(t, server.Object{NetCode: 269})
			weapon := selfWeaponElectricNativeAlloc(t, server.Object{NetCode: 512})
			rawDamage := selfWeaponElectricNativeAlloc(t, in.Damage)
			source, victim := owner, target
			if in.NilOwner {
				source = nil
			}
			if in.NilTarget {
				victim = nil
			}
			wantOwner, wantTarget, wantWeapon := *owner, *target, *weapon
			mutate := func() {
				if in.Mutate {
					effect.AttackPreHit52.Valf, effect.AttackPreDmg64.Valf = 1, 1
					if in.Which == 0 {
						target.PosVec = types.Pointf{X: 77.5, Y: 88.25}
					}
				}
			}

			got := make([]event, 0, len(row.Events))
			recordAmount := func(name string, amount uint32) {
				got = append(got, event{Service: name, Amount: amount})
			}
			transfer := func(name string, unit *server.Object, amount int32, expected *server.Object) {
				if unit != expected {
					t.Fatalf("%s uses %p, want %p", name, unit, expected)
				}
				recordAmount(name, uint32(amount))
			}
			readAvailable := func(name string, unit *server.Object) uint16 {
				amount := uint16(in.Available)
				transfer(name, unit, int32(amount), target)
				mutate()
				return amount
			}
			runtime := modifierResourceRuntime{
				damage: func(gotTarget, gotOwner, gotWeapon *server.Object, amount int32, kind object.DamageType) {
					if gotTarget != target || gotOwner != source || gotWeapon != weapon || kind != object.DamageElectric {
						t.Fatal("lightning damage changed identity or type")
					}
					recordAmount("damage", uint32(amount))
					mutate()
				},
				pointFX: func(op netmsg.Op, pos *types.Pointf) {
					if op != netmsg.MSG_FX_BLUE_SPARKS || pos != &target.PosVec {
						t.Fatal("lightning effect lost opcode or live position")
					}
					got = append(got, event{
						Service: "point-fx", Position: []uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)},
					})
				},
				audio: func(id sound.ID, unit *server.Object) {
					transfer("audio", unit, int32(id), target)
				},
				getMana:      func(unit *server.Object) uint16 { return readAvailable("get-mana", unit) },
				getHealth:    func(unit *server.Object) uint16 { return readAvailable("get-hp", unit) },
				manaSub:      func(unit *server.Object, amount int32) { transfer("mana-sub", unit, amount, target) },
				manaAdd:      func(unit *server.Object, amount int16) { transfer("mana-add", unit, int32(uint16(amount)), owner) },
				heal:         func(unit *server.Object, amount int32) { transfer("hp-add", unit, amount, owner) },
				healthDamage: func(unit *server.Object, amount int32) { transfer("hp-damage", unit, amount, owner) },
			}
			switch in.Which {
			case 0:
				modifierLightning4E06F0(effect, weapon, source, victim, runtime)
			case 1:
				modifierDrainMana4E0740(effect, source, victim, rawDamage, runtime)
			case 2:
				modifierVampirism4E07C0(effect, source, victim, rawDamage, runtime)
			case 3:
				modifierSympathy4E08E0(effect, source, victim, rawDamage, runtime)
			default:
				t.Fatalf("invalid original resource effect selector %d", in.Which)
			}
			wantCoeff := in.Value
			if in.Mutate && len(got) != 0 {
				wantCoeff = math.Float32bits(1)
				if in.Which == 0 {
					wantTarget.PosVec = types.Pointf{X: 77.5, Y: 88.25}
				}
			}
			if !reflect.DeepEqual(got, row.Events) || !row.Preserved {
				t.Fatalf("native resource events %+v, original %+v", got, row.Events)
			}
			if *rawDamage != in.Damage || *target != wantTarget || *owner != wantOwner || *weapon != wantWeapon ||
				math.Float32bits(effect.AttackPreHit52.Valf) != wantCoeff || math.Float32bits(effect.AttackPreDmg64.Valf) != wantCoeff {
				t.Fatal("resource effect changed input storage")
			}
		})
	}
	if selected != 120 {
		t.Fatalf("original startup observations %d, want 120", selected)
	}
}

// Keep the public pre-damage dispatch, typed C ABI callback, live HP lookup,
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
