package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	defaultDamageInvulnerableEnchant4E0B30 = EnchantID(23)
	defaultDamageShockEnchant4E0B30        = EnchantID(22)
	defaultDamageShieldEnchant4E0B30       = EnchantID(26)
	defaultDamageInvisibleEnchant4E0B30    = EnchantID(0)
	defaultDamageInvulnerableSound4E0B30   = 71
)

// DefaultDamageWorldRuntime4E0B30 isolates the services used by the
// native-width world-object slice of GAME.EXE 004E0B30. Unsupported reports
// a branch that must be ported before it can be executed safely on a 64-bit
// host; the caller must not fall back to the ABI32 body after that point.
type DefaultDamageWorldRuntime4E0B30 struct {
	Frame               func() uint32
	GameplayFlag1       func() bool
	QuestMode           func() bool
	IsZombie            func(*Object) bool
	IsEnemy             func(*Object, *Object) bool
	Audio               func(int, *Object)
	BuffOff             func(*Object, EnchantID)
	DefaultDamageSound  func(*Object, *Object)
	DamageClear         func(*Object, int32)
	DefaultDamageSoundC unsafe.Pointer
	Unsupported         func(reason string, target, source, weapon *Object, damage int32, typ object.DamageType)
}

func defaultDamageUnsupported4E0B30(
	runtime DefaultDamageWorldRuntime4E0B30,
	reason string,
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, target, source, weapon, damage, typ)
	}
	// The original callback reports success for nearly every early-exit path.
	// Returning success is safer than invoking its pointer-truncating ABI32
	// body, while Unsupported keeps the missing branch visible in test logs.
	return true
}

func defaultDamageWeaponHasPreDamageModifiers4E0B30(weapon *Object) bool {
	if weapon == nil || weapon.InitData == nil ||
		!weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) {
		return false
	}
	for _, modifier := range (*ModifierInitData)(weapon.InitData).Modifiers {
		if modifier != nil && modifier.AttackPreDmg64.Fnc != nil {
			return true
		}
	}
	return false
}

// DefaultDamageWorld4E0B30 restores the unmodified world-object Blade branch
// of GAME.EXE 004E0B30 without narrowing Object pointers. Player and Monster
// targets use their dedicated damage callbacks in normal data; if a data file
// assigns DefaultDamage to one of them, the unsupported hook exposes that
// separate porting task instead of entering the unsafe raw body.
func DefaultDamageWorld4E0B30(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime DefaultDamageWorldRuntime4E0B30,
) bool {
	if target == nil {
		return true
	}
	frame := uint32(0)
	if runtime.Frame != nil {
		frame = runtime.Frame()
	}
	if target.HasEnchant(defaultDamageInvulnerableEnchant4E0B30) {
		if byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(defaultDamageInvulnerableSound4E0B30, target)
		}
		return true
	}

	if target.Class().HasAny(object.MaskUnits) && !target.ObjFlags.Has(object.FlagDead) {
		return defaultDamageUnsupported4E0B30(runtime, "unit target", target, source, weapon, damage, typ)
	}
	if target.ObjFlags.Has(object.FlagDead) {
		zombie := false
		if runtime.IsZombie != nil {
			zombie = runtime.IsZombie(target)
		}
		if !zombie {
			return true
		}
		if weapon != nil {
			target.Obj130 = weapon
		} else {
			target.Obj130 = source
		}
		target.Field131 = uint32(typ)
		target.Frame134 = frame
		return true
	}

	gameplay := runtime.GameplayFlag1 != nil && runtime.GameplayFlag1()
	if !gameplay && source != nil {
		owner := source.FindOwnerChainPlayer()
		if owner != nil && owner.Class().HasAny(object.MaskUnits) {
			enemy := runtime.IsEnemy != nil && runtime.IsEnemy(target, owner)
			quest := runtime.QuestMode != nil && runtime.QuestMode()
			if !enemy && (target != owner || quest) {
				return true
			}
		}
	}
	if target.ObjFlags.Has(object.FlagNoUpdate) {
		return true
	}

	if typ != object.DamageBlade {
		return defaultDamageUnsupported4E0B30(runtime, "non-Blade protection", target, source, weapon, damage, typ)
	}
	if source != nil && target.HasEnchant(defaultDamageShockEnchant4E0B30) {
		return defaultDamageUnsupported4E0B30(runtime, "Shock retaliation", target, source, weapon, damage, typ)
	}
	if target.HasEnchant(defaultDamageShieldEnchant4E0B30) {
		return defaultDamageUnsupported4E0B30(runtime, "Shield reduction", target, source, weapon, damage, typ)
	}
	if defaultDamageWeaponHasPreDamageModifiers4E0B30(weapon) {
		return defaultDamageUnsupported4E0B30(runtime, "weapon pre-damage modifiers", target, source, weapon, damage, typ)
	}
	if target.DamageSound != nil && target.DamageSound != runtime.DefaultDamageSoundC {
		return defaultDamageUnsupported4E0B30(runtime, "custom damage sound", target, source, weapon, damage, typ)
	}
	if source == nil {
		target.Pos132 = types.Pointf{}
	} else if weapon != nil {
		if weapon.Class().HasAny(object.ClassWeapon | object.ClassWand) {
			target.Pos132 = source.PrevPos
		} else {
			target.Pos132 = weapon.PrevPos
		}
	} else {
		target.Pos132 = source.PrevPos
	}
	if source != nil && runtime.BuffOff != nil {
		// GAME.EXE calls BuffOff even when INVSIBILITY is not currently set.
		runtime.BuffOff(target, defaultDamageInvisibleEnchant4E0B30)
	}
	if weapon != nil {
		target.Obj130 = weapon
	} else {
		target.Obj130 = source
	}
	target.Field131 = uint32(typ)
	target.Frame134 = frame

	soundSource := source
	if weapon != nil {
		soundSource = weapon
	}
	if runtime.DefaultDamageSound != nil {
		runtime.DefaultDamageSound(target, soundSource)
	}

	if source != nil {
		monster := source
		if !monster.Class().Has(object.ClassMonster) {
			monster = source.ObjOwner
		}
		if monster != nil && monster.Class().Has(object.ClassMonster) &&
			runtime.IsEnemy != nil && runtime.IsEnemy(target, monster) {
			update := monster.UpdateDataMonster()
			if update.Field130 == 0 {
				update.Field130 = frame
			}
		}
	}
	if runtime.DamageClear != nil {
		runtime.DamageClear(target, damage)
	}
	return true
}
