package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	skeletonDamageInvulnerableSound4E23C0 = 71
	skeletonDamageBlockSound4E23C0        = 878
)

// SkeletonDamageRuntime4E23C0 supplies the engine services used by the
// native-width restoration of GAME.EXE 004E23C0.
type SkeletonDamageRuntime4E23C0 struct {
	Frame     func() uint32
	Audio     func(int, *Object)
	Direction func(types.Pointf, int16, types.Pointf) int32
	Default   DamageFunc
}

func callSpecialDamageDefault4E23C0(
	fnc DamageFunc,
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
) bool {
	if fnc == nil {
		return false
	}
	return fnc(target, source, weapon, damage, typ)
}

// SkeletonDamage4E23C0 preserves the invulnerability and shield-block gates
// of the original SkeletonDamage callback without loading UpdateData through
// a truncated PE32 object address.
func SkeletonDamage4E23C0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime SkeletonDamageRuntime4E23C0,
) bool {
	if target == nil {
		return false
	}
	if target.HasEnchant(ENCHANT_INVULNERABLE) {
		frame := uint32(0)
		if runtime.Frame != nil {
			frame = runtime.Frame()
		}
		if byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(skeletonDamageInvulnerableSound4E23C0, target)
		}
		return true
	}
	if source != nil && runtime.Direction != nil {
		attackPos := source.PrevPos
		if weapon != nil {
			attackPos = weapon.PrevPos
		}
		if runtime.Direction(target.PosVec, int16(target.Direction1), attackPos)&1 != 0 &&
			target.Class().Has(object.ClassMonster) {
			update := target.UpdateDataMonster()
			if update != nil && update.AIStackHead().Type() == ai.ACTION_BLOCK_ATTACK &&
				update.Field120_1 > update.Field120_0>>1 {
				if runtime.Audio != nil {
					runtime.Audio(skeletonDamageBlockSound4E23C0, target)
				}
				return true
			}
		}
	}
	return callSpecialDamageDefault4E23C0(runtime.Default, target, source, weapon, damage, typ)
}

func StoneDamage4E24B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	defaultDamage DamageFunc,
) bool {
	return callSpecialDamageDefault4E23C0(defaultDamage, target, source, weapon, damage, typ)
}

func MechGolemDamage4E24E0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	defaultDamage DamageFunc,
) bool {
	if typ == object.DamageElectric || typ == object.DamageAirborneElectric {
		damage *= 2
	}
	return callSpecialDamageDefault4E23C0(defaultDamage, target, source, weapon, damage, typ)
}

func FlammableDamage4E2520(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	defaultDamage DamageFunc,
) bool {
	if typ == object.DamageFlame || typ == object.DamageLava || typ == object.DamageExplosion {
		damage = 9999999
	}
	return callSpecialDamageDefault4E23C0(defaultDamage, target, source, weapon, damage, typ)
}

func BlackPowderDamage4E2560(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	defaultDamage DamageFunc,
) bool {
	if typ != object.DamageFlame && typ != object.DamageLava &&
		typ != object.DamageCrush && typ != object.DamageBlade {
		return false
	}
	return callSpecialDamageDefault4E23C0(defaultDamage, target, source, weapon, 999999, typ)
}
