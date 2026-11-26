package legacy

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

type originalModifierMultiplierFixture struct {
	OriginalSHA256     string `json:"original_sha256"`
	OriginalUnchanged  bool   `json:"original_unchanged"`
	StartupControlWord uint16 `json:"actual_startup_control_word"`
	Rows               []struct {
		Function string
		Name     string
		Mode     string
		Inputs   struct {
			ValueBits       uint32 `json:"value_bits"`
			CoefficientBits uint32 `json:"coefficient_bits"`
		}
		ValueBits     uint32 `json:"value_bits"`
		ReturnPointer string `json:"return_pointer"`
	}
}

func TestModifierReservedCallbacksHaveDistinctCAddresses(t *testing.T) {
	if !cfixture.ModifierReservedCallbacksDistinct() {
		t.Fatal("each reserved modifier must have its own C callback address")
	}
}

func TestModifierDefendMultipliersOriginalNativeC(t *testing.T) {
	replayModifierMultipliers(t, "defend", map[string]func(*server.ModifierEff, *float32) *float32{
		"ArmorMultiplierEffect":      modifierArmorMultiplierCallNative4E0370,
		"DurabilityMultiplierEffect": modifierDurabilityMultiplierCallNative4E0380,
	}, false)
}

func TestModifierAttackMultipliersOriginalNativeC(t *testing.T) {
	replayModifierMultipliers(t, "attack", map[string]func(*server.ModifierEff, *float32) *float32{
		"DamageMultiplierEffect": modifierDamageMultiplierCallNative4E04C0,
	}, true)
}

func replayModifierMultipliers(
	t *testing.T,
	dataset string,
	scalars map[string]func(*server.ModifierEff, *float32) *float32,
	projectileCases bool,
) {
	t.Helper()
	raw, err := os.ReadFile("testdata/modifier_" + dataset + "_multipliers_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture originalModifierMultiplierFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		!fixture.OriginalUnchanged || fixture.StartupControlWord != 0xe3f || len(fixture.Rows) != 256 {
		t.Fatal("incomplete unchanged original multiplier observations")
	}

	effect, freeEffect := alloc.NewHigh(server.ModifierEff{})
	t.Cleanup(freeEffect)
	projectile, freeProjectile := alloc.NewHigh(server.Object{})
	t.Cleanup(freeProjectile)
	output, freeOutput := alloc.NewHigh(struct {
		value float32
		guard uint32
	}{})
	t.Cleanup(freeOutput)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(effect.C()) <= math.MaxUint32 ||
		uintptr(projectile.CObj()) <= math.MaxUint32 || uintptr(unsafe.Pointer(output)) <= math.MaxUint32) {
		t.Fatal("native fixture addresses must exceed 4 GiB")
	}
	counts := make(map[string]int)
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		counts[row.Function]++
		t.Run(row.Function+"/"+row.Name, func(t *testing.T) {
			coefficient := math.Float32frombits(row.Inputs.CoefficientBits)
			value := math.Float32frombits(row.Inputs.ValueBits)
			if row.Function == "ProjectileSpeedEffect" && projectileCases {
				*effect = server.ModifierEff{
					Attack40: server.ModifierEffFnc{Valf: coefficient},
					Defend76: server.ModifierEffFnc{Valf: -6},
				}
				*projectile = server.Object{SpeedCur: value, SpeedBase: 27.5, NetCode: 0x3011}
				if got := modifierProjectileSpeedCallNative4E09B0(effect, projectile); got != projectile ||
					row.ReturnPointer != "projectile" || projectile.SpeedBase != 27.5 || projectile.NetCode != 0x3011 {
					t.Fatal("native projectile callback changed identity or unrelated fields")
				}
				value = projectile.SpeedCur
			} else {
				call := scalars[row.Function]
				if call == nil {
					t.Fatalf("unknown original multiplier %q", row.Function)
				}
				output.value, output.guard = value, 0x61b3a97d
				*effect = server.ModifierEff{
					Attack40: server.ModifierEffFnc{Valf: 0.25},
					Defend76: server.ModifierEffFnc{Valf: coefficient},
				}
				if row.Function == "DamageMultiplierEffect" {
					effect.Attack40.Valf, effect.Defend76.Valf = coefficient, -6
				}
				if got := call(effect, &output.value); got != &output.value || row.ReturnPointer != "value" || output.guard != 0x61b3a97d {
					t.Fatal("native scalar callback changed identity or adjacent data")
				}
				value = output.value
			}
			if bits := math.Float32bits(value); bits != row.ValueBits {
				t.Errorf("native bits %08x, want %08x from GAME.EXE", bits, row.ValueBits)
			}
		})
	}
	for function := range scalars {
		if counts[function] != 32 {
			t.Errorf("%s has %d original startup contexts, want 32", function, counts[function])
		}
	}
	if projectileCases && counts["ProjectileSpeedEffect"] != 32 {
		t.Errorf("projectile speed has %d original startup contexts, want 32", counts["ProjectileSpeedEffect"])
	}
}

func TestModifierNativeMultipliersKeepUnpinnedStorage(t *testing.T) {
	effect := server.ModifierEff{
		Attack40: server.ModifierEffFnc{Valf: 2},
		Defend76: server.ModifierEffFnc{Valf: 0.25},
	}
	projectile := server.Object{SpeedCur: 4}
	server.CallModifierAttackEffect(ProjectileSpeedEffectPointer4E09B0(), &effect, nil, nil, nil, unsafe.Pointer(&projectile))
	if projectile.SpeedCur != 8 {
		t.Fatalf("projectile speed = %g, want 8", projectile.SpeedCur)
	}
	damage := float32(7)
	attack := modifierDamageMultiplierPointer4E04C0()
	server.CallModifierAttackEffect(attack, nil, nil, nil, nil, unsafe.Pointer(&damage))
	if damage != 7 {
		t.Fatalf("missing modifier changed damage to %g", damage)
	}
	server.CallModifierAttackEffect(attack, &effect, nil, nil, nil, nil)
	server.CallModifierAttackEffect(attack, &effect, nil, nil, nil, unsafe.Pointer(&damage))
	if damage != 14 {
		t.Fatalf("multiplied damage = %g, want 14", damage)
	}
	defend := playerDamageLateDefendFunctionsNative4E1320()
	armor, durability := float32(4), float32(4)
	worldPlayerCallDefend415C00(defend.armorMultiplier, &effect, nil, &armor)
	worldPlayerCallDefend415C00(defend.durabilityMultiplier, &effect, nil, &durability)
	if armor != 1 || durability != 7 {
		t.Fatalf("defend multipliers returned armor %g and durability %g, want 1 and 7", armor, durability)
	}
}
