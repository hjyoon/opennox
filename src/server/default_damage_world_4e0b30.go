package server

import (
	"image"
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
	defaultDamageInvulnerableSound4E0B30   = 71
	defaultDamageShockSound4E0B30          = 135
	defaultDamageShockBalance4E0B30        = "ShockDamage"
	defaultDamageShockBalanceIndex4E0B30   = 4
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
	FireProtection      func(*Object) float64
	ElectricProtection  func(*Object) float64
	MonsterHasHitSound  func(*Object) bool
	DefaultDamageSound  func(*Object, *Object)
	AdjustFieldGuide    func(*Object, *Object, int32) int32
	BalanceFloatInd     func(string, int) float64
	CallDamage          func(*Object, *Object, *Object, int32, object.DamageType) bool
	PlayerSetState      func(*Object, PlayerState) bool
	AdjustHP            func(*Object, int32)
	VampirismFX         func(int, image.Point, image.Point, uint16)
	ShieldReduce        func(*Object, *int32, object.DamageType, *Object)
	CanApplyPreDamage   func(*ModifierEff) bool
	ApplyPreDamage      func(*ModifierEff, *Object, *Object, *Object, *int32)
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

func defaultDamageWeaponPreDamageModifiers4E0B30(weapon *Object) [4]*ModifierEff {
	if weapon == nil || weapon.InitData == nil ||
		!weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) {
		return [4]*ModifierEff{}
	}
	return (*ModifierInitData)(weapon.InitData).Modifiers
}

// defaultDamageAttackQualifies4E1400 restores GAME.EXE 004E1400 without
// reading Object or WandUseData through their PE32 offsets. The helper is used
// by DefaultDamage's friendly-hit and Shock predicates in the original. The
// Shock call site always supplies a non-nil weapon, but retaining the no-weapon
// branch here documents the complete predicate and makes later callers safe to
// migrate without reintroducing an ABI32 call.
func defaultDamageAttackQualifies4E1400(source, weapon *Object) bool {
	if weapon == nil {
		if source.Class().Has(object.ClassPlayer) {
			return true
		}
		return source.Class().Has(object.ClassMonster) && uint32(source.SubClass())&0x10 != 0
	}

	class := weapon.Class()
	subclass := uint32(weapon.SubClass())
	if class.Has(object.ClassWand) {
		if subclass&0x047f0000 == 0 {
			return true
		}
		// GAME.EXE reads the low byte of WandUseData.Flags at offset 96.
		// AsWand deliberately preserves the original nil-fault boundary.
		if weapon.UseData.AsWand().Flags&2 != 0 {
			return true
		}
	} else if class.Has(object.ClassWeapon) && subclass&0x047f00fe == 0 {
		return true
	}
	return uint8(class)&uint8(object.ClassMonster) != 0
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

// DefaultDamageWorld4E0B30 restores the unmodified world-object damage branch,
// player melee and unarmed electric spells against ordinary monsters, monster
// and source-less scripted electric damage against ordinary monsters, missile
// IMPACT and Magic Missile EXPLOSION against ordinary monsters, and the
// monster-on-monster self-weapon BITE branch from GAME.EXE 004E0B30 without
// narrowing Object pointers.
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
	monsterElectric := monsterUpdate != nil && weapon == nil && (source == nil || source.Class().HasAny(object.ClassPlayer|object.ClassMonster)) &&
		(typ == object.DamageElectric || typ == object.DamageAirborneElectric)
	selfSourcedMissileImpact := monsterUpdate != nil && source != nil && source == weapon &&
		source.Class().Has(object.ClassMissile) && !source.Class().HasAny(object.MaskUnits) && typ == object.DamageImpact
	playerFiredMissileImpact := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassPlayer) &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) && typ == object.DamageImpact
	monsterFiredMissileImpact := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassMonster) &&
		source.UpdateData != nil && weapon != nil && weapon.Class().Has(object.ClassMissile) && typ == object.DamageImpact
	missileImpact := selfSourcedMissileImpact || playerFiredMissileImpact || monsterFiredMissileImpact
	playerFiredMissileExplosion := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassPlayer) &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) && typ == object.DamageExplosion
	missileSourcedExplosion := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassMissile) &&
		!source.Class().HasAny(object.MaskUnits) && weapon == nil && typ == object.DamageExplosion
	missileExplosion := playerFiredMissileExplosion || missileSourcedExplosion
	missileDamage := missileImpact || missileExplosion
	if monsterUpdate != nil {
		if target.HealthData == nil {
			return defaultDamageUnsupported4E0B30(runtime, "monster without health", target, source, weapon, damage, typ)
		}
		// GAME.EXE 004E0EA1 handles WEAPON|WAND (0x1001000) alike,
		// while 004E0EF5 identifies a no-weapon hit by damage type 10.
		playerMelee := source != nil && source.Class().Has(object.ClassPlayer) &&
			((weapon != nil && weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) && typ == object.DamageBlade) ||
				(weapon == nil && typ == object.DamageClaw))
		monsterBite := source != nil && source.Class().Has(object.ClassMonster) && source.UpdateData != nil &&
			weapon == source && typ == object.DamageBite
		if !playerMelee && !monsterBite && !missileDamage && !monsterElectric {
			return defaultDamageUnsupported4E0B30(runtime, "unsupported monster damage shape", target, source, weapon, damage, typ)
		}
		// This monster subclass ignores both electric damage types.
		if monsterElectric && uint32(target.SubClass())&0x800 != 0 {
			return true
		}
		if monsterBite && runtime.MonsterHasHitSound == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing monster hit-sound lookup", target, source, weapon, damage, typ)
		}
		// The original's friendly-hit gate does not apply when the weapon is
		// a missile (sub_4E1400 returns false for this class).
		if source != nil && !missileDamage && (runtime.IsEnemy == nil || !runtime.IsEnemy(target, source)) {
			return true
		}
	}
	vampirism := source != nil && target.Class().HasAny(object.MaskUnits) &&
		source.HasEnchant(damageVampirismEnchant4E0B30)
	if vampirism && (runtime.Audio == nil || runtime.BalanceFloatInd == nil ||
		runtime.AdjustHP == nil || runtime.VampirismFX == nil) {
		return defaultDamageUnsupported4E0B30(runtime, "missing Vampirism service", target, source, weapon, damage, typ)
	}

	shockRetaliates := source != nil && weapon != nil &&
		target.HasEnchant(defaultDamageShockEnchant4E0B30) &&
		source.Class().HasAny(object.ClassPlayer|object.ClassMonster) &&
		defaultDamageAttackQualifies4E1400(source, weapon)
	if shockRetaliates {
		if runtime.Audio == nil || runtime.BuffOff == nil || runtime.BalanceFloatInd == nil ||
			runtime.CallDamage == nil ||
			(source.Class().Has(object.ClassPlayer) && runtime.PlayerSetState == nil) {
			return defaultDamageUnsupported4E0B30(runtime, "missing Shock retaliation service", target, source, weapon, damage, typ)
		}
		runtime.Audio(defaultDamageShockSound4E0B30, source)
		runtime.BuffOff(target, defaultDamageShockEnchant4E0B30)
		shockDamage := playerCollideRound4E8460(float32(runtime.BalanceFloatInd(
			defaultDamageShockBalance4E0B30, defaultDamageShockBalanceIndex4E0B30,
		)))
		_ = runtime.CallDamage(source, target, nil, shockDamage, object.DamageElectric)
		if source.Class().Has(object.ClassPlayer) {
			_ = runtime.PlayerSetState(source, PlayerState23)
		}
	}

	if monsterUpdate != nil {
		// Monster subclass bit 0x10 enters item defense callbacks in the
		// original. Keep it outside the ordinary-monster admission gate.
		if uint32(target.SubClass())&0x10 != 0 {
			return defaultDamageUnsupported4E0B30(runtime, "monster defense callbacks", target, source, weapon, damage, typ)
		}
	}

	nonUnit := !target.Class().HasAny(object.MaskUnits)
	sourceLessLava := typ == object.DamageLava && source == nil && weapon == nil && nonUnit
	if typ != object.DamageBlade && typ != object.DamageClaw && typ != object.DamageBite && !missileDamage && !nonUnit && !monsterElectric {
		return defaultDamageUnsupported4E0B30(runtime, "unsupported protection branch", target, source, weapon, damage, typ)
	}
	fireProtected := typ == object.DamageFlame || typ == object.DamageLava || typ == object.DamageExplosion
	electricProtected := typ == object.DamageElectric || typ == object.DamageAirborneElectric
	if fireProtected && runtime.FireProtection == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing fire-protection service", target, source, weapon, damage, typ)
	}
	if electricProtected && runtime.ElectricProtection == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing electric-protection service", target, source, weapon, damage, typ)
	}
	shielded := target.HasEnchant(defaultDamageShieldEnchant4E0B30) && typ != object.DamagePoison &&
		(typ != object.DamageManaBomb || source != target)
	if shielded && runtime.ShieldReduce == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing Shield reduction service", target, source, weapon, damage, typ)
	}
	preDamageModifiers := defaultDamageWeaponPreDamageModifiers4E0B30(weapon)
	for _, modifier := range preDamageModifiers {
		if modifier == nil || modifier.AttackPreDmg64.Fnc == nil {
			continue
		}
		if runtime.CanApplyPreDamage == nil || runtime.ApplyPreDamage == nil ||
			!runtime.CanApplyPreDamage(modifier) {
			return defaultDamageUnsupported4E0B30(runtime, "weapon pre-damage modifiers", target, source, weapon, damage, typ)
		}
	}
	if target.DamageSound != nil && target.DamageSound != runtime.DefaultDamageSoundC {
		return defaultDamageUnsupported4E0B30(runtime, "custom damage sound", target, source, weapon, damage, typ)
	}
	if fireProtected {
		protectionValue := runtime.FireProtection(target)
		if protectionValue != 0 && byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(104, target)
		}
		// GAME.EXE compares the binary64 return before spilling it to v46,
		// then evaluates the damage expression in binary64 and spills v42 to
		// binary32 immediately before FISTP.
		protection := float32(protectionValue)
		scaled := float32((1.0 - float64(protection)) * float64(damage))
		damage = int32(math.RoundToEven(float64(scaled)))
		if damage == 0 {
			damage = 1
		}
	}
	if electricProtected {
		protectionValue := runtime.ElectricProtection(target)
		if protectionValue != 0 && byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(108, target)
		}
		protection := float32(protectionValue)
		scaled := float32((1.0 - float64(protection)) * float64(damage))
		damage = int32(math.RoundToEven(float64(scaled)))
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
	if (source != nil || sourceLessLava) && runtime.BuffOff != nil {
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
	for _, modifier := range preDamageModifiers {
		if modifier != nil && modifier.AttackPreDmg64.Fnc != nil {
			runtime.ApplyPreDamage(modifier, weapon, source, target, &damage)
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
	if vampirism {
		damageApplyVampirism4E0B30(
			source, target, weapon, damage,
			runtime.Audio, runtime.BalanceFloatInd, runtime.AdjustHP, runtime.VampirismFX,
		)
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
	if shielded {
		shieldSource := weapon
		if shieldSource == nil {
			shieldSource = source
		}
		runtime.ShieldReduce(target, &damage, typ, shieldSource)
		if damage == 0 {
			return false
		}
	}
	if runtime.DamageClear != nil {
		runtime.DamageClear(target, damage)
	}
	return true
}
