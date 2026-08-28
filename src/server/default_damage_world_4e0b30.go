package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	defaultDamageInvulnerableEnchant4E0B30 = EnchantID(23)
	defaultDamageShockEnchant4E0B30        = EnchantID(22)
	defaultDamageShieldEnchant4E0B30       = EnchantID(26)
	defaultDamageInvisibleEnchant4E0B30    = EnchantID(0)
	defaultDamageVampirismEnchant4E0B30    = EnchantID(13)
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
	FireProtection      func(*Object) float32
	MonsterHasHitSound  func(*Object) bool
	DefaultDamageSound  func(*Object, *Object)
	AdjustFieldGuide    func(*Object, *Object, int32) int32
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

// DefaultDamageFieldGuide4E0B30 restores sub_4FB000/sub_4FB050 for the
// cooperative and Quest call site in GAME.EXE 004E0B30. Guide zero is the
// original invalid sentinel, and the per-player array retains its fixed
// 41-entry semantic layout on native-width hosts.
func (s *Server) DefaultDamageFieldGuide4E0B30(source, target *Object, damage int32) int32 {
	if source == nil || target == nil || !source.Class().Has(object.ClassPlayer) ||
		!target.Class().Has(object.ClassMonster) {
		return damage
	}
	update := source.UpdateDataPlayer()
	if update == nil || update.Player == nil {
		return damage
	}
	typ := s.Types.ByInd(int(target.TypeInd))
	if typ == nil {
		return damage
	}
	guide := 0
	for i := 1; i < len(rewardFieldGuideNames4F0D20); i++ {
		if rewardFieldGuideNames4F0D20[i] == typ.ID() {
			guide = i
			break
		}
	}
	if guide == 0 || update.Player.BeastScrollLvl[guide] == 0 {
		return damage
	}
	value := float32(s.Balance.Float("FieldGuideDamageBonus")*float64(damage) + 0.5)
	return int32(value)
}

// DefaultDamageWorld4E0B30 restores the unmodified world-object Blade branch,
// the monster-on-monster self-weapon BITE branch, and source-less LAVA damage
// to non-unit objects from GAME.EXE 004E0B30 without narrowing Object pointers.
// Player targets use their dedicated damage callback in normal data; other
// protection, modifier, and equipment branches remain visible through
// Unsupported instead of entering the unsafe raw body.
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

	var monsterUpdate *MonsterUpdateData
	if target.Class().Has(object.ClassMonster) {
		monsterUpdate = target.UpdateDataMonster()
		// GAME.EXE clears this latch before checking whether the monster is
		// already dead or whether the incoming damage will be admitted.
		monsterUpdate.Field547 = 0
		if typ == object.DamageFlame || typ == object.DamageLava || typ == object.DamageExplosion ||
			typ == object.DamagePlasma || typ == object.DamageDispelUndead {
			monsterUpdate.StatusFlags |= object.MonStatusOnFire
		}
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
	if target.Class().HasAny(object.MaskUnits) && monsterUpdate == nil {
		return defaultDamageUnsupported4E0B30(runtime, "non-monster unit target", target, source, weapon, damage, typ)
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
	if monsterUpdate != nil {
		if target.HealthData == nil {
			return defaultDamageUnsupported4E0B30(runtime, "monster without health", target, source, weapon, damage, typ)
		}
		playerBlade := source != nil && source.Class().Has(object.ClassPlayer) &&
			weapon != nil && weapon.Class().Has(object.ClassWeapon) && typ == object.DamageBlade
		monsterBite := source != nil && source.Class().Has(object.ClassMonster) && source.UpdateData != nil &&
			weapon == source && typ == object.DamageBite
		if !playerBlade && !monsterBite {
			return defaultDamageUnsupported4E0B30(runtime, "unsupported monster damage shape", target, source, weapon, damage, typ)
		}
		if monsterBite && runtime.MonsterHasHitSound == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing monster hit-sound lookup", target, source, weapon, damage, typ)
		}
		if runtime.IsEnemy == nil || !runtime.IsEnemy(target, source) {
			return true
		}
		// Monster subclass bit 0x10 enters item defense callbacks in the
		// original. Keep it outside this first ordinary-melee admission gate.
		if uint32(target.SubClass())&0x10 != 0 {
			return defaultDamageUnsupported4E0B30(runtime, "monster defense callbacks", target, source, weapon, damage, typ)
		}
		if source.HasEnchant(defaultDamageVampirismEnchant4E0B30) {
			return defaultDamageUnsupported4E0B30(runtime, "Vampirism healing", target, source, weapon, damage, typ)
		}
	}

	lava := typ == object.DamageLava && source == nil && weapon == nil && !target.Class().HasAny(object.MaskUnits)
	if typ != object.DamageBlade && typ != object.DamageBite && !lava {
		return defaultDamageUnsupported4E0B30(runtime, "non-Blade protection", target, source, weapon, damage, typ)
	}
	if lava && runtime.FireProtection == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing fire-protection service", target, source, weapon, damage, typ)
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
	if lava {
		protection := runtime.FireProtection(target)
		if protection != 0 && byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(104, target)
		}
		damage = int32(math.RoundToEven(float64(float32(damage) * (1 - protection))))
		if damage == 0 {
			damage = 1
		}
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
	if (source != nil || lava) && runtime.BuffOff != nil {
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
	if monsterUpdate != nil {
		monsterUpdate.StatusFlags |= object.MonStatusInjured
		if monsterUpdate.Field547 == 0 {
			monsterUpdate.Field547 = 2
			monsterUpdate.Field546 = uint32(typ)
		}
	}

	suppressDamageSound := target == weapon && target.Class().HasAny(object.ClassWeapon|object.ClassWand)
	if !suppressDamageSound && source != nil && source.Class().Has(object.ClassMonster) &&
		source.UpdateData != nil && runtime.MonsterHasHitSound != nil {
		suppressDamageSound = runtime.MonsterHasHitSound(source)
	}
	if !suppressDamageSound && runtime.DefaultDamageSound != nil {
		soundSource := source
		if weapon != nil {
			soundSource = weapon
		}
		runtime.DefaultDamageSound(target, soundSource)
	}
	if monsterUpdate != nil && runtime.AdjustFieldGuide != nil {
		damage = runtime.AdjustFieldGuide(source, target, damage)
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
