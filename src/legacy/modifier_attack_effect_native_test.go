package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestModifierAttackEffectsKeepNativePointers(t *testing.T) {
	effect := &server.ModifierEff{
		Attack40:       server.ModifierEffFnc{Valf: 1.75, Val: 11},
		AttackPreHit52: server.ModifierEffFnc{Valf: 2.25, Val: 22},
		AttackPreDmg64: server.ModifierEffFnc{Valf: 3.25, Val: 33},
	}
	damage := float32(12.5)
	projectile := &server.Object{SpeedCur: 8}
	var pin runtime.Pinner
	pin.Pin(effect)
	pin.Pin(&damage)
	pin.Pin(projectile)
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"effect":     unsafe.Pointer(effect),
			"damage":     unsafe.Pointer(&damage),
			"projectile": unsafe.Pointer(projectile),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native high address", name, pointer)
			}
		}
	}

	if got := modifierDamageMultiplierCallNative4E04C0(effect, &damage); got != &damage {
		t.Fatalf("damage multiplier result = %p, want %p", got, &damage)
	}
	if damage != 21.875 {
		t.Fatalf("multiplied damage = %v, want 21.875", damage)
	}
	if got := modifierProjectileSpeedCallNative4E09B0(effect, projectile); got != projectile {
		t.Fatalf("projectile speed result = %p, want %p", got, projectile)
	}
	if projectile.SpeedCur != 14 {
		t.Fatalf("projectile speed = %v, want 14", projectile.SpeedCur)
	}

	if got := int32(nox_modifier_effect_getAttackInt(effect.C())); got != 11 {
		t.Fatalf("attack integer = %d, want 11", got)
	}
	if got := float32(nox_modifier_effect_getAttackFloat(effect.C())); got != 1.75 {
		t.Fatalf("attack float = %v, want 1.75", got)
	}
	if got := int32(nox_modifier_effect_getPreHitInt(effect.C())); got != 22 {
		t.Fatalf("pre-hit integer = %d, want 22", got)
	}
	if got := float32(nox_modifier_effect_getPreHitFloat(effect.C())); got != 2.25 {
		t.Fatalf("pre-hit float = %v, want 2.25", got)
	}
	if got := int32(nox_modifier_effect_getPreDamageInt(effect.C())); got != 33 {
		t.Fatalf("pre-damage integer = %d, want 33", got)
	}
	if got := float32(nox_modifier_effect_getPreDamageFloat(effect.C())); got != 3.25 {
		t.Fatalf("pre-damage float = %v, want 3.25", got)
	}
	runtime.KeepAlive(effect)
	runtime.KeepAlive(projectile)
}
