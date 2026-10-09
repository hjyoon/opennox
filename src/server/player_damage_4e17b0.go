package server

import (
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	playerDamageInvulnerableEnchant4E17B0 = EnchantID(23)
	playerDamageReflectEnchant4E17B0      = EnchantID(27)
	playerDamageShieldEnchant4E17B0       = EnchantID(26)
	playerDamageInvisibleEnchant4E17B0    = EnchantID(0)
	playerDamageInvulnerableSound4E17B0   = 71
)

// PlayerDamageRuntime4E17B0 contains the services called by the native-width
// PlayerDamage slice. Read-only service checks precede this slice's stores.
// A callback that replaces an item with an unsupported live record is reported
// at that record's use, never retried through a PE32 callback on a 64-bit host.
type PlayerDamageRuntime4E17B0 struct {
	playerPrefix            *playerDamagePrefix4E17B0
	Melee                   PlayerDamageMeleeRuntime4E17B0
	Frame                   func() uint32
	CoopMode                func() bool
	GameplayFlag1           func() bool
	QuestMode               func() bool
	QuestDamageScale        func() float32
	GodMode                 func() bool
	IsEnemy                 func(*Object, *Object) bool
	SentryGlobeType         uint16
	GameBallType            uint16
	GameBallOnDamage        func(*Object, *Object, int32)
	Audio                   func(int, *Object)
	BuffOff                 func(*Object, EnchantID)
	ObserveClear            func(*Object)
	ItemArmorValue          func(*Object) float32
	CanApplyArmorDefend     func(*ModifierEff) bool
	ApplyArmorDefend        func(*ModifierEff, *Object, *Object, *Object, *Object, *float32) bool
	CanDamageArmor          func(*Object) bool
	DamageArmor             func(*Object, *Object, *Object, int32, object.DamageType) bool
	ReportArmorHealth       func(*Object, *Object, uint16, uint16)
	CanApplyLateDefend      func(*ModifierEff) bool
	ApplyLateDefend         func(*ModifierEff, *Object, *Object, *Object, *Object, int32, object.DamageType) int32
	BlockSourceExcluded     func(*Object) bool
	BlockSourceOnlyExcluded func(*Object) bool
	BlockDirection          func(*Object, types.Pointf) bool
	BerserkShieldBlock      func(*Object) bool
	ProjectileReflect       func(*Object, *Object)
	ClearOwner              func(*Object)
	SetOwner                func(*Object, *Object)
	ChangeOwner             func(*Object, *Object)
	PointFX                 func(int, types.Pointf)
	BlockDamagePercent      func() float64
	CanDamageBlockItem      func(*Object) bool
	DamageBlockItem         func(*Object, *Object, *Object, *Object, float32, object.DamageType) bool
	PlayerSetState          func(*Object, PlayerState) bool
	FireProtection          func(*Object) float64
	ElectricArmorScale      func(*Object) float32
	BalanceFloatInd         func(string, int) float64
	AdjustHP                func(*Object, int32)
	VampirismFX             func(int, image.Point, image.Point, uint16)
	PlayerDamageSound       func(*Object, *Object)
	PlayerDamageSoundC      unsafe.Pointer
	ShieldReduce            func(*Object, *int32, object.DamageType, *Object)
	DamageClear             func(*Object, int32)
	DefaultDamage           func(*Object, *Object, *Object, int32, object.DamageType) bool
	Unsupported             func(string, *Object, *Object, *Object, int32, object.DamageType)
}

func playerDamageReflectShield4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	if !target.HasEnchant(playerDamageReflectEnchant4E17B0) {
		return false, false, false
	}
	attack := weapon
	if attack == nil {
		attack = source
	}
	if attack == nil {
		return false, false, false
	}
	missile := attack.ObjClass.Has(object.ClassMissile)
	electric := typ == object.DamageZapRay || typ == object.DamageAirborneElectric
	if !missile && !electric {
		return false, false, false
	}
	if runtime.BlockDirection == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing Reflect Shield direction service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if !runtime.BlockDirection(target, attack.PosVec) {
		return false, false, false
	}
	transferOwner := missile && uint32(attack.ObjSubClass)&0x40 == 0
	changeOwner := missile && uint32(attack.ObjSubClass)&2 != 0
	pointFX := typ == object.DamageZapRay
	if runtime.Audio == nil ||
		(missile && runtime.ProjectileReflect == nil) ||
		(transferOwner && (runtime.ClearOwner == nil || runtime.SetOwner == nil)) ||
		(changeOwner && runtime.ChangeOwner == nil) ||
		(pointFX && runtime.PointFX == nil) {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing Reflect Shield effect service", target, source, weapon, damage, typ)
		return true, handled, result
	}

	if target.Class().Has(object.ClassPlayer) && runtime.playerPrefix == nil {
		update := target.UpdateDataPlayer()
		update.Field76 = 0
		if update.Player.ObserveTarget() != nil && runtime.ObserveClear != nil {
			runtime.ObserveClear(target)
		}
	} else if !target.Class().Has(object.ClassPlayer) {
		// 004E18D1 selects the NPC hit marker at PE32 offset 2188;
		// the player marker/observer fields belong to a different layout.
		target.UpdateDataMonster().Field547 = 0
	}
	if missile {
		runtime.ProjectileReflect(attack, target)
		if transferOwner {
			runtime.ClearOwner(attack)
			runtime.SetOwner(target, attack)
		}
		if changeOwner {
			runtime.ChangeOwner(attack, target)
		}
	}
	if pointFX {
		runtime.PointFX(132, target.PosVec)
	}
	runtime.Audio(122, target)
	return true, true, false
}

func playerDamageShieldItem4E17B0(target *Object) *Object {
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.ObjFlags.Has(object.FlagEquipped) && uint32(item.ObjSubClass)&2 != 0 {
			return item
		}
	}
	return nil
}

func playerDamageShieldBlock4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	playerTarget := target.Class().Has(object.ClassPlayer)
	var update *PlayerUpdateData
	if playerTarget {
		var armorFlags, weaponFlags uint32
		if prefix := runtime.playerPrefix; prefix != nil {
			update = prefix.update
			armorFlags, weaponFlags = prefix.armorFlags, prefix.weaponFlags
		} else {
			update = target.UpdateDataPlayer()
			armorFlags, weaponFlags = update.Player.ArmorEquip, update.Player.WeaponEquip
		}
		if armorFlags&0x3000000 == 0 {
			return false, false, false
		}
		shieldStance := update.State == PlayerState16
		if !shieldStance && update.State == PlayerState1 && weaponFlags&0x400 == 0 {
			if runtime.BerserkShieldBlock == nil {
				handled, result = playerDamageUnsupported4E17B0(runtime, "missing berserker shield service", target, source, weapon, damage, typ)
				return true, handled, result
			}
			shieldStance = runtime.BerserkShieldBlock(target)
		}
		if !shieldStance {
			return false, false, false
		}
	} else {
		// 004E1B1F..004E1B92: ordinary blocks require a source, never
		// intercept MANA_BOMB or electric types 9/17, and select the NPC's
		// action/equipment fields rather than the player state layout.
		if source == nil || typ == object.DamageManaBomb || typ == object.DamageElectric || typ == object.DamageAirborneElectric ||
			target.UpdateDataMonster().ArmorEquipFlags&0x3000000 == 0 || target.MonsterActionGet50A020() != ai.ACTION_BLOCK_ATTACK {
			return false, false, false
		}
	}
	attack := weapon
	if attack == nil {
		attack = source
	}
	if attack == nil {
		return false, false, false
	}
	excluded := runtime.BlockSourceExcluded
	if !playerTarget && weapon == nil {
		// Without a weapon, 004E1ACE..004E1AF4 excludes the three Fists
		// and Meteor only; the two ToxicCloud exclusions belong to a3 != 0.
		excluded = runtime.BlockSourceOnlyExcluded
	}
	if excluded == nil || runtime.BlockDirection == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing shield direction service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if excluded(attack) || !runtime.BlockDirection(target, attack.PrevPos) {
		return false, false, false
	}
	reflectProjectile := attack.ObjClass.Has(object.ClassMissile) && uint32(attack.ObjSubClass)&0x70 == 0
	transferOwner := reflectProjectile && uint32(attack.ObjSubClass)&2 == 0
	if reflectProjectile && runtime.ProjectileReflect == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing projectile reflection service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if transferOwner && (runtime.ClearOwner == nil || runtime.SetOwner == nil) {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing projectile owner service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if runtime.Audio == nil || runtime.BlockDamagePercent == nil || runtime.DamageBlockItem == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing shield damage service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	shield := playerDamageShieldItem4E17B0(target)
	if shield != nil && (runtime.CanDamageBlockItem == nil || !runtime.CanDamageBlockItem(shield) ||
		(playerTarget && runtime.PlayerSetState == nil) || (!playerTarget && runtime.Melee.MonsterPopBlockAction == nil)) {
		handled, result = playerDamageUnsupported4E17B0(runtime, "shield durability callback", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if playerTarget && runtime.playerPrefix == nil {
		update.Field76 = 0
		if update.Player.ObserveTarget() != nil && runtime.ObserveClear != nil {
			runtime.ObserveClear(target)
		}
		if (typ == object.DamageImpale || playerDamageMissileFlameShape4E17B0(source, weapon, typ) ||
			playerDamageMissileExplosionShape4E17B0(source, weapon, typ)) &&
			source != nil && weapon != nil && source != weapon {
			// 004E1A8F records the distinct missile before block audio/wear;
			// Reflect Shield returns earlier with a cleared marker instead.
			update.Field76 = 1
			update.Field75 = uint32(weapon.TypeInd)
		}
	} else if !playerTarget {
		ud := target.UpdateDataMonster()
		ud.Field547 = 0
		if (weapon != nil && source != weapon) || (weapon == nil && (typ == object.DamageClaw || typ == object.DamageCrush)) {
			// 004E1B0D writes the NPC marker/type before block audio or wear.
			ud.Field547, ud.Field546 = 1, uint32(attack.TypeInd)
		}
	}
	runtime.Audio(878, target)
	if reflectProjectile {
		runtime.ProjectileReflect(attack, target)
		if transferOwner && attack.Class().Has(object.ClassMissile) && uint32(attack.SubClass())&2 == 0 {
			runtime.ClearOwner(attack)
			runtime.SetOwner(target, attack)
		}
	}
	// The 004E1BFD call through sub_4E2330 reaches EquipDamage without a selected
	// shield; its nil/no-health guard is a successful wear no-op.
	amount := float32(runtime.BlockDamagePercent() * float64(damage))
	effective := weapon
	if effective == nil {
		effective = source
	}
	if !runtime.DamageBlockItem(shield, target, source, effective, amount, typ) && runtime.Unsupported != nil {
		runtime.Unsupported("shield durability failed", target, source, weapon, damage, typ)
	}
	if shield != nil && shield.ObjFlags.Has(object.FlagDestroyed) {
		if playerTarget {
			runtime.PlayerSetState(target, PlayerState13)
		} else {
			runtime.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, true, false
}

type playerDamageLateDefend4E1320 struct {
	item     *Object
	modifier *ModifierEff
}

func playerDamageUnsupported4E17B0(
	runtime PlayerDamageRuntime4E17B0,
	reason string,
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
) (bool, bool) {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, target, source, weapon, damage, typ)
	}
	return false, false
}

func playerDamageRound4E17B0(value float32) int32 {
	return int32(math.RoundToEven(float64(value)))
}

func playerDamageOwnsType4E17B0(owner *Object, typeInd uint16) bool {
	for obj := owner.Field129; obj != nil; obj = obj.Field128 {
		if obj.TypeInd == typeInd {
			return true
		}
	}
	return false
}

func playerDamagePlanLateDefend4E1320(
	target *Object,
	runtime PlayerDamageRuntime4E17B0,
) ([]playerDamageLateDefend4E1320, bool) {
	var plan []playerDamageLateDefend4E1320
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) ||
			!item.ObjClass.HasAny(object.ClassFlag|object.ClassWeapon|object.ClassArmor|object.ClassWand) ||
			item.InitData == nil {
			continue
		}
		modifiers := item.InitDataModifier().Modifiers
		for i := 2; i < len(modifiers); i++ {
			modifier := modifiers[i]
			if modifier == nil || modifier.Defend76.Fnc == nil {
				continue
			}
			if runtime.CanApplyLateDefend == nil || runtime.ApplyLateDefend == nil ||
				!runtime.CanApplyLateDefend(modifier) {
				return nil, false
			}
			plan = append(plan, playerDamageLateDefend4E1320{item: item, modifier: modifier})
		}
	}
	return plan, true
}

// Admission only: do not retain items or slot contents for the default tail.
// The actual 004E1320 call owns its flags-only traversal and nil-init fault at
// use; a missing init base cannot be inspected by this read-only check.
func playerDamageLateDefendReady4E17B0(target *Object, runtime PlayerDamageRuntime4E17B0) (hasDefend, ready bool) {
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) || item.InitData == nil {
			continue
		}
		initData := item.InitDataModifier()
		for slot := 2; slot < 4; slot++ {
			modifier := initData.Modifiers[slot]
			if modifier == nil || modifier.Defend76.Fnc == nil {
				continue
			}
			hasDefend = true
			if runtime.CanApplyLateDefend == nil || runtime.ApplyLateDefend == nil || !runtime.CanApplyLateDefend(modifier) {
				return hasDefend, false
			}
		}
	}
	return hasDefend, true
}

// Check only callback availability. In particular, do not execute the armor
// lookup/Defend callbacks or snapshot carry, health, or next-item pointers.
// Zero/signed amounts and a zero armor denominator still reach 004E2180.
func playerDamageArmorReady4E17B0(target *Object, runtime PlayerDamageRuntime4E17B0) bool {
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjClass.Has(object.ClassArmor) || !item.ObjFlags.Has(object.FlagEquipped) {
			continue
		}
		if runtime.ItemArmorValue == nil {
			return false
		}
		// The lookup is required even here; EquipDamage owns this no-op.
		if item.HealthData == nil {
			continue
		}
		if item.UpdateData == nil || item.InitData == nil || item.Damage == nil ||
			runtime.CanDamageArmor == nil || !runtime.CanDamageArmor(item) || runtime.DamageArmor == nil {
			return false
		}
		modifier := item.InitDataModifier().Modifiers[1]
		if modifier != nil && modifier.Defend76.Fnc != nil &&
			(runtime.ApplyArmorDefend == nil || runtime.CanApplyArmorDefend == nil || !runtime.CanApplyArmorDefend(modifier)) {
			return false
		}
	}
	return true
}

func playerDamageApplyArmor4E17B0(target, source, weapon *Object, remaining int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0) {
	PlayerDamageItems4E2180(target, source, weapon, remaining, typ, PlayerDamageItemsRuntime4E2180{
		ItemArmorValue: func(item *Object) float64 { return float64(runtime.ItemArmorValue(item)) },
		EquipDamage: func(item, owner, source, effective *Object, amount float32, typ object.DamageType) {
			EquipDamageNative4E16D0(item, owner, source, effective, amount, typ, ItemDurabilityDamageRuntime4E1560{
				ApplyDefend: runtime.ApplyArmorDefend,
				Damage: func(item, source, effective *Object, damage int32, typ object.DamageType) bool {
					// Defend or an earlier item's callback can replace the live
					// damage function after admission. Never fall back to PE32.
					if runtime.CanDamageArmor == nil || !runtime.CanDamageArmor(item) || runtime.DamageArmor == nil {
						if runtime.Unsupported != nil {
							runtime.Unsupported("unsupported live armor damage callback", owner, source, effective, remaining, typ)
						}
						return false
					}
					return runtime.DamageArmor(item, source, effective, damage, typ)
				},
				ReportHealth: runtime.ReportArmorHealth,
				Unsupported: func(reason string, _, owner, source, effective *Object, _ float32, typ object.DamageType) {
					if runtime.Unsupported != nil {
						runtime.Unsupported(reason, owner, source, effective, remaining, typ)
					}
				},
			})
		},
	})
}

// Glyph's CastShock passes its unit caster as both source and weapon. The
// weaponless player/monster shape also supplies spells and Shock retaliation.
// Distinct weapons, wands and missiles still need their separate effect ports.
func playerDamageElectricShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if source == nil || source.UpdateData == nil ||
		(typ != object.DamageElectric && typ != object.DamageAirborneElectric) {
		return false
	}
	if weapon == nil {
		return source.Class().HasAny(object.ClassPlayer | object.ClassMonster)
	}
	return weapon == source && source.Class().HasAny(object.ClassPlayer|object.ClassMonster) &&
		!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)
}

// playerDamageMonsterMissilePierce4E17B0 restores NPC switch case 3 at
// 004E1F84. Absorption uses the armor cached at 004E1898, but the carry and
// armor-durability helper reload the live update data. The cached NPC marker
// identifies the distinct missile before durability, including when a defense
// callback has replaced the live update. GodMode applies only to players.
func playerDamageMonsterMissilePierce4E17B0(
	target, source, weapon *Object, update *MonsterUpdateData, armorValue float32,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, source, weapon, damage, typ)
	}
	// A missing service may reject before stores. With the production scale
	// available, 004E2046 queries Quest only after wear/marker/minimum.
	if runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	scaled := float32((1.0 - float64(armorValue)) * float64(damage))
	live := target.UpdateDataMonster()
	accumulated := scaled + math.Float32frombits(live.Field1)
	effective := playerDamageRound4E17B0(accumulated)
	if !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	update.Field547 = 1
	update.Field546 = uint32(weapon.TypeInd)
	live.Field1 = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, runtime)
	if update.Field547 == 0 {
		update.Field547 = 2
		update.Field546 = uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if runtime.QuestMode != nil && runtime.QuestMode() {
		if runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing live quest damage service", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

// playerDamageMonster4E17B0 restores the native-width monster half of
// PlayerDamage used by scripted NPCs. In particular, War01A's wizard setpiece
// sends Bryan's MorningStar CRUSH and the wizard's weaponless
// AIRBORNE_ELECTRIC hits through this callback. The original accepts monsters
// whose subclass has bit 0x10, shares their fractional-damage accumulator with
// players, damages equipped armor, and then enters DefaultDamage.
func playerDamageMonster4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if defaultDamageCasterImpactShape4E0B30(weapon, typ) && !playerDamageMonsterImpactShape4E17B0(source, weapon, typ) {
		return playerDamageCasterImpact4E17B0(target, source, weapon, damage, typ, runtime)
	}
	if target.UpdateData == nil || uint32(target.SubClass())&0x10 == 0 {
		return playerDamageUnsupported4E17B0(runtime, "unsupported monster target", target, source, weapon, damage, typ)
	}
	if target.ObjFlags.HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	cloudPoison := playerDamageCloudPoisonShape4E17B0(weapon, typ)
	frame := uint32(0)
	if runtime.Frame != nil && (!cloudPoison || target.HasEnchant(playerDamageInvulnerableEnchant4E17B0)) {
		frame = runtime.Frame()
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}

	if source == nil && weapon == nil && typ == object.DamagePoison {
		return playerDamageMonsterPoison4E17B0(target, damage, runtime)
	}
	// 004E187E/004E1898 cache the NPC update and absorption before the
	// defense callbacks. Carry (004E20F0) and armor wear (004E2180) reload
	// their own live update; the hit marker remains on this cached base.
	update := target.UpdateDataMonster()
	armorValue := math.Float32frombits(update.Field518)
	greatSword := playerDamageGreatSwordContext4E17B0{
		weaponFlags: update.WeaponEquipFlags, armorFlags: update.ArmorEquipFlags,
		marker: &update.Field547, markerType: &update.Field546,
	}
	if source == nil && weapon == nil && typ == object.DamageLava || playerDamageWorldImpaleShape4E17B0(source, weapon, typ) {
		return playerDamageNPCEnvironment4E17B0(target, source, weapon, greatSword, armorValue, damage, typ, runtime)
	}
	if cloudPoison {
		return playerDamageMonsterCloudPoison4E17B0(target, source, weapon, update, greatSword, damage, runtime)
	}
	// Meteor's radial call passes its terminal owner (or nil), with no
	// weapon. Case 7 still enters 004E1F84's full armor/carry/wear switch;
	// it is not restricted to missile-shaped sources. Keep this prefix
	// separate so marker reset, cached equipment and the one ordinary
	// facing check precede callbacks, without reselecting live equipment.
	// Powder-barrel breaking objects are SIMPLE|LIGHT world sources,
	// without a unit update record, and use the same case-7 prefix.
	if playerDamageWorldExplosionShape4E17B0(source, weapon, typ) || weapon == nil && typ == object.DamageExplosion &&
		(source == nil || (source.Class().HasAny(object.MaskUnits) &&
			!source.Class().HasAny(object.ClassMissile|object.ClassWeapon|object.ClassWand))) {
		update.Field547 = 0
		if source != nil {
			attackPos := source.PrevPos
			if runtime.BlockSourceOnlyExcluded == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION exclusion", target, source, weapon, damage, typ)
			}
			excluded := runtime.BlockSourceOnlyExcluded(source)
			front := false
			if !excluded {
				if runtime.BlockDirection == nil {
					return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION direction", target, source, weapon, damage, typ)
				}
				front = runtime.BlockDirection(target, attackPos)
			}
			if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
				return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC weapon-less EXPLOSION block record", target, source, weapon, damage, typ)
			}
			if front && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK && greatSword.armorFlags&0x3000000 != 0 {
				if runtime.Audio == nil || runtime.BlockDamagePercent == nil {
					return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION shield effect", target, source, weapon, damage, typ)
				}
				runtime.Audio(878, target)
				// 004E1BA7/004E1BBE read the attack's live class/subclass
				// after audio and reflection, even for a nonmissile entry.
				if source.Class().Has(object.ClassMissile) && uint32(source.SubClass())&0x70 == 0 {
					if runtime.ProjectileReflect == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION reflection", target, source, weapon, damage, typ)
					}
					runtime.ProjectileReflect(source, target)
					if source.Class().Has(object.ClassMissile) && uint32(source.SubClass())&2 == 0 {
						if runtime.ClearOwner == nil || runtime.SetOwner == nil {
							return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION shield owner", target, source, weapon, damage, typ)
						}
						runtime.ClearOwner(source)
						runtime.SetOwner(target, source)
					}
				}
				amount := float32(runtime.BlockDamagePercent() * float64(damage))
				shield := playerDamageShieldItem4E17B0(target)
				if runtime.DamageBlockItem == nil || (shield != nil && (runtime.CanDamageBlockItem == nil || !runtime.CanDamageBlockItem(shield))) {
					return playerDamageUnsupported4E17B0(runtime, "unsupported NPC weapon-less EXPLOSION live shield wear", target, source, weapon, damage, typ)
				}
				if !runtime.DamageBlockItem(shield, target, source, weapon, amount, typ) && runtime.Unsupported != nil {
					runtime.Unsupported("NPC weapon-less EXPLOSION shield wear failed", target, source, weapon, damage, typ)
				}
				if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
					if runtime.Melee.MonsterPopBlockAction == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION broken shield action", target, source, weapon, damage, typ)
					}
					runtime.Melee.MonsterPopBlockAction(target)
				}
				return true, false
			}
		}
		if runtime.DefaultDamage == nil || !playerDamageArmorReady4E17B0(target, runtime) {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION armor/default service", target, source, weapon, damage, typ)
		}
		if runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC weapon-less EXPLOSION quest scale", target, source, weapon, damage, typ)
		}
		if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC weapon-less EXPLOSION carry record", target, source, weapon, damage, typ)
		}
		carry := &target.UpdateDataMonster().Field1
		scaled := float32((1 - float64(armorValue)) * float64(damage))
		accumulated := scaled + math.Float32frombits(*carry)
		effective := playerDamageRound4E17B0(accumulated)
		*carry = math.Float32bits(accumulated - float32(effective))
		playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, runtime)
		if update.Field547 == 0 {
			update.Field547, update.Field546 = 2, uint32(typ)
		}
		if damage > 0 && effective == 0 {
			effective = 1
		}
		if runtime.QuestMode != nil && runtime.QuestMode() {
			if runtime.QuestDamageScale == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing live NPC weapon-less EXPLOSION quest scale", target, source, weapon, damage, typ)
			}
			before := effective
			effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
			if before > 0 && effective < 1 {
				effective = 1
			}
		}
		return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
	}
	// Native monster strikes supply themselves as both source and weapon.
	// Nonmissile strikes skip Reflect, but keep the cached marker clear,
	// PrevPos snapshot and one exclusion/facing check before defenses.
	if playerDamageMonsterImpactShape4E17B0(source, weapon, typ) || playerDamageNPCMonsterSelfStrikeShape4E17B0(source, weapon, typ) {
		update.Field547 = 0
		attackPos := weapon.PrevPos
		if runtime.BlockSourceExcluded == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC self-weapon IMPACT exclusion service", target, source, weapon, damage, typ)
		}
		excluded := runtime.BlockSourceExcluded(weapon)
		if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC self-weapon IMPACT block record", target, source, weapon, damage, typ)
		}
		front := false
		if !excluded {
			if runtime.BlockDirection == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing NPC self-weapon IMPACT direction service", target, source, weapon, damage, typ)
			}
			front = runtime.BlockDirection(target, attackPos)
		}
		if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC self-weapon IMPACT tail record", target, source, weapon, damage, typ)
		}
		runtime.BlockDirection = func(*Object, types.Pointf) bool { return front }
		if typ == object.DamageImpact {
			return playerDamageMonsterImpactTail4E17B0(target, source, weapon, greatSword, armorValue, damage, typ, runtime)
		}
		return playerDamageNPCMonsterSelfStrikeTail4E17B0(target, source, weapon, greatSword, armorValue, damage, typ, runtime)
	}
	// Stock SentryGlobe is SIMPLE|IMMOBILE, not MISSILE or an electric
	// weapon. Its terminal owner is either a unit or the unowned globe
	// itself. Restore this bounded NPC prefix before the generic defenses:
	// 004E18D1 clears the cached marker before Reflect, and 004E1A49
	// snapshots PrevPos before exclusions/one ordinary facing check.
	zapRay := typ == object.DamageZapRay && weapon != nil &&
		weapon.Class().Has(object.ClassSimple) && weapon.Class().Has(object.ClassImmobile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) &&
		(source == weapon || (source != nil && source.UpdateData != nil &&
			source.Class().HasAny(object.MaskUnits) && !source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)))
	if zapRay {
		update.Field547 = 0
		if target.HasEnchant(playerDamageReflectEnchant4E17B0) {
			if runtime.BlockDirection == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY Reflect direction service", target, source, weapon, damage, typ)
			}
			if runtime.BlockDirection(target, weapon.PosVec) {
				// The nonmissile branch was selected before the direction
				// callback; a callback changing class must not reselect it.
				if runtime.PointFX == nil {
					return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY Reflect FX service", target, source, weapon, damage, typ)
				}
				runtime.PointFX(132, target.PosVec)
				if runtime.Audio == nil {
					return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY Reflect audio service", target, source, weapon, damage, typ)
				}
				runtime.Audio(122, target)
				return true, false
			}
		}
		attackPos := weapon.PrevPos
		if runtime.BlockSourceExcluded == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY exclusion service", target, source, weapon, damage, typ)
		}
		excluded := runtime.BlockSourceExcluded(weapon)
		if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC ZAP_RAY marker record", target, source, weapon, damage, typ)
		}
		if source != weapon {
			// 004E1B0D reads the live type after exclusions, but keeps the
			// entry update base, even when a callback replaces UpdateData.
			update.Field547, update.Field546 = 1, uint32(weapon.TypeInd)
		}
		if !excluded {
			if runtime.BlockDirection == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY block direction service", target, source, weapon, damage, typ)
			}
			if runtime.BlockDirection(target, attackPos) {
				if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
					return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC ZAP_RAY block record", target, source, weapon, damage, typ)
				}
				if target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK && greatSword.armorFlags&0x3000000 != 0 {
					if runtime.Audio == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY shield audio service", target, source, weapon, damage, typ)
					}
					runtime.Audio(878, target)
					// 004E1BA7 and 004E1BBE reload class/subclass after the
					// audio/reflection callbacks rather than caching ownership.
					if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&0x70 == 0 {
						if runtime.ProjectileReflect == nil {
							return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY shield reflection service", target, source, weapon, damage, typ)
						}
						runtime.ProjectileReflect(weapon, target)
						if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 == 0 {
							if runtime.ClearOwner == nil || runtime.SetOwner == nil {
								return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY shield owner service", target, source, weapon, damage, typ)
							}
							runtime.ClearOwner(weapon)
							runtime.SetOwner(target, weapon)
						}
					}
					if runtime.BlockDamagePercent == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY shield balance service", target, source, weapon, damage, typ)
					}
					amount := float32(runtime.BlockDamagePercent() * float64(damage))
					// 004E1BFD selects the live inventory shield only after
					// audio/reflection/balance. Nil/no-health wear is a no-op.
					shield := playerDamageShieldItem4E17B0(target)
					if shield != nil && (runtime.CanDamageBlockItem == nil || !runtime.CanDamageBlockItem(shield)) {
						return playerDamageUnsupported4E17B0(runtime, "NPC ZAP_RAY shield durability callback", target, source, weapon, damage, typ)
					}
					if runtime.DamageBlockItem == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing NPC ZAP_RAY shield wear service", target, source, weapon, damage, typ)
					}
					if !runtime.DamageBlockItem(shield, target, source, weapon, amount, typ) && runtime.Unsupported != nil {
						runtime.Unsupported("NPC ZAP_RAY shield durability failed", target, source, weapon, damage, typ)
					}
					if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
						if runtime.Melee.MonsterPopBlockAction == nil || !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
							return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC ZAP_RAY broken-shield action", target, source, weapon, damage, typ)
						}
						runtime.Melee.MonsterPopBlockAction(target)
					}
					return true, false
				}
			}
		}
		if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) ||
			!weapon.Class().Has(object.ClassSimple) || !weapon.Class().Has(object.ClassImmobile) ||
			weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC ZAP_RAY tail record", target, source, weapon, damage, typ)
		}
		// Nonmissile type 16 skips GreatSword at 004E1C23 and enters
		// 004E1E83 raw: no armor, electric scale, carry or hurt state.
		if update.Field547 == 0 {
			update.Field547, update.Field546 = 2, uint32(typ)
		}
		effective := damage
		if runtime.GodMode != nil && runtime.GodMode() && target.Class().Has(object.ClassPlayer) {
			return true, true
		}
		if runtime.QuestMode != nil && runtime.QuestMode() {
			if runtime.QuestDamageScale == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing live NPC ZAP_RAY quest damage service", target, source, weapon, damage, typ)
			}
			effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
			if damage > 0 && effective < 1 {
				effective = 1
			}
		}
		if runtime.DefaultDamage == nil || !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC ZAP_RAY default service/record", target, source, weapon, damage, typ)
		}
		return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
	}
	// 004E18F3..004E19C1 reflects before the damage-type switch, including
	// missiles whose unreflected damage path has not yet been ported.
	if applicable, handled, result := playerDamageReflectShield4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return handled, result
	}
	// The ordinary NPC shield branch also precedes damage-shape admission.
	// Its return suppresses HP/armor damage even for zero/negative hits and
	// missile shapes whose unblocked tails remain separate native ports.
	if applicable, handled, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return handled, result
	}
	if applicable, handled, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, greatSword, damage, typ, runtime); applicable {
		return handled, result
	}
	if playerDamageMissileFlameShape4E17B0(source, weapon, typ) {
		return playerDamageMonsterMissileFlame4E17B0(target, source, weapon, update, damage, typ, runtime)
	}
	if playerDamageMissileExplosionShape4E17B0(source, weapon, typ) {
		return playerDamageMonsterMissileExplosion4E17B0(target, source, weapon, update, armorValue, damage, typ, runtime)
	}
	// HarpoonCollide supplies the owning Player and a distinct HarpoonBolt.
	// Case 11 shares the full-armor/carry tail at 004E1F84 with PIERCE.
	// Stock HarpoonBolt is MISSILE|WEAPON subclass 0x10: 004E1400 excludes
	// ranged weapons, not every weapon-class missile.
	harpoonImpact := typ == object.DamageImpact && source != nil && source != weapon &&
		source.UpdateData != nil && source.Class().Has(object.ClassPlayer) &&
		!source.Class().HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand|object.ClassMissile) &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
		!defaultDamageAttackQualifies4E1400(source, weapon)
	if harpoonImpact {
		return playerDamageMonsterMissilePierce4E17B0(target, source, weapon, update, armorValue, damage, typ, runtime)
	}
	// Match the already restored DefaultDamage PIERCE tail. Stock GolemArrow
	// is MISSILE|WEAPON subclass 0x10, so test 004E1400 rather than rejecting
	// every weapon-class missile. NPC case 3 has no MONSTER-only source
	// test: player-fired missiles use the same cached armor/live carry.
	// Mixed melee/wand/unit projectiles retain separate admission failures.
	if typ == object.DamageImpale && source != nil && source != weapon &&
		source.Class().HasAny(object.ClassPlayer|object.ClassMonster) && source.UpdateData != nil &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
		!defaultDamageAttackQualifies4E1400(source, weapon) {
		return playerDamageMonsterMissilePierce4E17B0(target, source, weapon, update, armorValue, damage, typ, runtime)
	}
	// PlayerCollide passes the charging Warrior as both source and weapon.
	// NPC case 2 at 004E1EE8 shares the CRUSH armor/carry tail with players;
	// it is not restricted to the scripted monster-with-weapon shape.
	playerCharge := typ == object.DamageCrush && source != nil && source == weapon &&
		source.UpdateData != nil && source.Class().Has(object.ClassPlayer) &&
		!source.Class().HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand|object.ClassMissile)
	crush := playerCharge || (typ == object.DamageCrush && source != nil && source.Class().Has(object.ClassMonster) &&
		source.UpdateData != nil && weapon != nil && weapon.Class().Has(object.ClassWeapon))
	electric := playerDamageElectricShape4E17B0(source, weapon, typ)
	if !crush && !electric {
		return playerDamageUnsupported4E17B0(runtime, "unsupported monster damage shape", target, source, weapon, damage, typ)
	}
	// Missing services still fail before tail stores. With the production
	// scale service available, the live Quest flag is read only at 004E2046,
	// after wear and the minimum-damage adjustment, not as an execution plan.
	if runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	if electric && runtime.ElectricArmorScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing electric armor service", target, source, weapon, damage, typ)
	}
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, source, weapon, damage, typ)
	}

	if !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	update.Field547 = 0
	if weapon != nil && weapon != source {
		update.Field547 = 1
		update.Field546 = uint32(weapon.TypeInd)
	} else if weapon == nil && (typ == object.DamageClaw || typ == object.DamageCrush) {
		update.Field547 = 1
		update.Field546 = uint32(source.TypeInd)
	}
	var scaled float32
	if crush {
		scaled = float32((1.0 - float64(armorValue)*0.5) * float64(damage))
	} else {
		scaled = float32(float64(runtime.ElectricArmorScale(target)) * float64(damage))
	}
	// The electric scale callback can replace UpdateData or its carry. Read
	// it after the callback and binary32 spill, just like 004E20F0.
	live := target.UpdateDataMonster()
	accumulated := scaled + math.Float32frombits(live.Field1)
	effective := playerDamageRound4E17B0(accumulated)
	live.Field1 = math.Float32bits(accumulated - float32(effective))
	remaining := damage
	if crush {
		remaining = damage - effective
	}
	playerDamageApplyArmor4E17B0(target, source, weapon, remaining, typ, runtime)
	if update.Field547 == 0 {
		update.Field547 = 2
		update.Field546 = uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if runtime.QuestMode != nil && runtime.QuestMode() {
		if runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing live quest damage service", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

// playerDamageElectricPlayer4E17B0 restores the cases 9/17 at 004E1DF1.
// Electric armor scales HP damage, but 004E1E1E distributes the original
// damage (not the absorbed difference) to equipped armor. The caller has
// already checked the Player, observer, Coop and Reflect Shield gates.
func playerDamageElectricPlayer4E17B0(
	target, source, weapon *Object, damage int32, typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if runtime.ElectricArmorScale == nil || runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing player electric service", target, source, weapon, damage, typ)
	}
	if runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	if !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	var update *PlayerUpdateData
	if runtime.playerPrefix != nil {
		// The entry marker base survives ObserveClear and later callbacks.
		// Do not repeat that prefix using a replacement live observer.
		update = runtime.playerPrefix.update
	} else {
		update = target.UpdateDataPlayer()
		observe := update.Player.ObserveTarget() != nil
		if observe && runtime.ObserveClear == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing player electric ObserveClear service", target, source, weapon, damage, typ)
		}
		update.Field76 = 0
		if observe {
			runtime.ObserveClear(target)
		}
	}
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
		return playerDamageUnsupported4E17B0(runtime, "unsupported live player electric record", target, source, weapon, damage, typ)
	}
	scaled := float32(float64(runtime.ElectricArmorScale(target)) * float64(damage))
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
		return playerDamageUnsupported4E17B0(runtime, "unsupported live player electric carry", target, source, weapon, damage, typ)
	}
	// 004E20F0 reloads the live carry after ElectricArmorScale, not from
	// the cached marker base. 004E2180 similarly reloads live armor data.
	live := target.UpdateDataPlayer()
	accumulated := scaled + math.Float32frombits(live.Field21)
	effective := playerDamageRound4E17B0(accumulated)
	live.Field21 = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, weapon, damage, typ, runtime)
	if target.Class().Has(object.ClassPlayer) && update.Field76 == 0 {
		update.Field76 = 2
		// 004E1E49 copies the incoming DWORD type, not float32(type).
		update.Field75 = uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	// 004E2025 reads GodMode before the live Player-class gate.
	if runtime.GodMode != nil && runtime.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if runtime.QuestMode != nil && runtime.QuestMode() {
		if runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing live quest damage service", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

// playerDamageMissilePierce4E17B0 restores switch case 3 at 004E1F84.
// The armor absorption was captured by the caller before direction/effect
// callbacks; 004E2180 reads the live armor value again for durability. The
// distinct missile marker precedes durability and survives the switch tail.
func playerDamageMissilePierce4E17B0(
	target, source, weapon *Object, update *PlayerUpdateData, armorValue float32,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, source, weapon, damage, typ)
	}
	if runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	if applicable, h, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return h, result
	}
	if runtime.playerPrefix == nil && update.Player.ObserveTarget() != nil {
		// Possession can replace the live update/carry during ObserveClear.
		// That prefix requires its own ordered native slice.
		return playerDamageUnsupported4E17B0(runtime, "possessed player missile PIERCE", target, source, weapon, damage, typ)
	}
	scaled := float32((1.0 - float64(armorValue)) * float64(damage))
	live := target.UpdateDataPlayer()
	accumulated := scaled + math.Float32frombits(live.Field21)
	effective := playerDamageRound4E17B0(accumulated)
	remaining := damage - effective
	if !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	if runtime.playerPrefix == nil {
		update.Field76 = 0
		update.Field76 = 1
		update.Field75 = uint32(weapon.TypeInd)
	}
	live.Field21 = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, weapon, remaining, typ, runtime)
	if update.Field76 == 0 {
		update.Field76 = 2
		update.Field75 = uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if runtime.QuestMode != nil && runtime.QuestMode() {
		if runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing live quest damage service", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

// PlayerDamageNative4E17B0 restores ordinary player/NPC melee, unit-sourced SIMPLE
// CRUSH (including stock Fists), Spider BITE, monster-fired
// missile IMPACT, pure spell-missile IMPACT (including unowned Pixies),
// Troll self-weapon IMPACT, player/monster-fired missile PIERCE, Berserker Charge CRUSH,
// SentryGlobe ZAP_RAY, world FLAME,
// unarmed player/monster and unit-self-weapon ELECTRIC/AIRBORNE_ELECTRIC,
// and source-less LAVA/POISON branches of
// GAME.EXE 004E17B0 together with their relevant unit-default-damage tails,
// plus the front-facing shield block and the common Quest damage scaling tail,
// and the early Reflect Shield and Coop self-damage gates. Unported spell or
// modifier branches return handled=false at their admission/use boundary;
// an already executed entry prefix is not rolled back.
func PlayerDamageNative4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if target == nil {
		return playerDamageUnsupported4E17B0(runtime, "non-player target", target, source, weapon, damage, typ)
	}
	// Keep the stock FIRE|SIMPLE|DANGEROUS route disjoint from the
	// earlier bare-FIRE and NPC SIMPLE shield slices. Class.Has is ANY-bit.
	const stockWorldFlame = object.ClassFire | object.ClassSimple | object.ClassDangerous
	if playerDamageWorldFlameShape4E17B0(weapon, typ) && weapon.Class()&stockWorldFlame == stockWorldFlame {
		return playerDamageWorldFlame4E17B0(target, source, weapon, damage, typ, runtime)
	}
	if target.Class().Has(object.ClassPlayer) && playerDamageWorldImpaleShape4E17B0(source, weapon, typ) {
		return playerDamagePlayerWorldImpale4E17B0(target, source, weapon, damage, typ, runtime)
	}
	if target.Class().Has(object.ClassPlayer) && playerDamagePlayerWeaponlessExplosionShape4E17B0(source, weapon, typ) {
		return playerDamagePlayerWeaponlessExplosion4E17B0(target, source, weapon, damage, typ, runtime)
	}
	// SIMPLE spell missiles (stock Pixie) need the terminal-parent path;
	// keep the already ported bare monster-missile slice below disjoint.
	if target.Class().Has(object.ClassPlayer) && playerDamageSpellMissileImpactShape4E17B0(source, weapon, typ) &&
		weapon.Class().Has(object.ClassSimple) {
		return playerDamagePlayerSpellMissileImpact4E17B0(target, source, weapon, damage, typ, runtime)
	}
	if playerDamageMeleeShape4E17B0(source, weapon, typ) || playerDamageSimpleCrushShape4E17B0(source, weapon, typ) {
		return PlayerDamageMeleeNative4E17B0(target, source, weapon, damage, typ, runtime)
	}
	if target.ObjClass.Has(object.ClassMonster) {
		return playerDamageMonster4E17B0(target, source, weapon, damage, typ, runtime)
	}
	if !target.ObjClass.Has(object.ClassPlayer) || target.UpdateData == nil {
		return playerDamageUnsupported4E17B0(runtime, "non-player target", target, source, weapon, damage, typ)
	}
	if target.ObjFlags.HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	// Stock SentryGlobe is SIMPLE|IMMOBILE. Its terminal owner may be
	// the globe itself, a player or an NPC; case 16 is not positive-only
	// and never depends on a cached SentryGlobe type ID.
	zapRay := typ == object.DamageZapRay && weapon != nil &&
		weapon.Class().Has(object.ClassSimple) && weapon.Class().Has(object.ClassImmobile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) &&
		(source == weapon || (source != nil && source.UpdateData != nil &&
			source.Class().HasAny(object.ClassPlayer|object.ClassMonster) &&
			!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)))
	frame := uint32(0)
	if runtime.Frame != nil && (!zapRay || target.HasEnchant(playerDamageInvulnerableEnchant4E17B0)) {
		frame = runtime.Frame()
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	update := target.UpdateDataPlayer()
	player := update.Player
	if player == nil {
		return playerDamageUnsupported4E17B0(runtime, "nil player info", target, source, weapon, damage, typ)
	}
	if player.Field3680&1 != 0 {
		return true, false
	}
	if runtime.CoopMode != nil && runtime.CoopMode() && source != nil &&
		source.FindOwnerChainPlayer() == target && typ != object.DamageManaBomb {
		return true, false
	}
	if zapRay {
		armorFlags := player.ArmorEquip
		// 004E18C4 clears the entry update base before ObserveClear and
		// Reflect. A callback can replace the live update without moving
		// this cached marker or the entry equipment mask.
		update.Field76 = 0
		if player.ObserveTarget() != nil {
			if runtime.ObserveClear == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY observe service", target, source, weapon, damage, typ)
			}
			runtime.ObserveClear(target)
		}
		if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live player ZAP_RAY observe record", target, source, weapon, damage, typ)
		}
		// The nonmissile Reflect branch is selected before the direction
		// callback. It uses current PosVec, not the later PrevPos snapshot.
		if target.HasEnchant(playerDamageReflectEnchant4E17B0) {
			if weapon.Class().Has(object.ClassMissile) {
				return playerDamageUnsupported4E17B0(runtime, "unsupported live player ZAP_RAY Reflect missile", target, source, weapon, damage, typ)
			}
			if runtime.BlockDirection == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY Reflect direction service", target, source, weapon, damage, typ)
			}
			if runtime.BlockDirection(target, weapon.PosVec) {
				if runtime.PointFX == nil {
					return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY Reflect FX service", target, source, weapon, damage, typ)
				}
				runtime.PointFX(132, target.PosVec)
				if runtime.Audio == nil {
					return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY Reflect audio service", target, source, weapon, damage, typ)
				}
				runtime.Audio(122, target)
				return true, false
			}
		}
		attackPos := weapon.PrevPos
		if runtime.BlockSourceExcluded == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY exclusion service", target, source, weapon, damage, typ)
		}
		excluded := runtime.BlockSourceExcluded(weapon)
		if source != weapon && target.Class().Has(object.ClassPlayer) {
			// 004E1AA2 attributes the live weapon type after exclusions,
			// while retaining the cached entry update address.
			update.Field76, update.Field75 = 1, uint32(weapon.TypeInd)
		}
		if !excluded {
			if runtime.BlockDirection == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY block direction service", target, source, weapon, damage, typ)
			}
			if runtime.BlockDirection(target, attackPos) {
				if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
					return playerDamageUnsupported4E17B0(runtime, "unsupported live player ZAP_RAY block record", target, source, weapon, damage, typ)
				}
				// 004E1B56 uses cached equipment and post-facing cached
				// stance. Type 16 skips the later melee berserker block.
				if update.State == PlayerState16 && armorFlags&0x3000000 != 0 {
					if runtime.Audio == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY shield audio service", target, source, weapon, damage, typ)
					}
					runtime.Audio(878, target)
					// Reload after audio and reflection, as 004E1BA7/1BBE do.
					if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&0x70 == 0 {
						if runtime.ProjectileReflect == nil {
							return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY shield reflection service", target, source, weapon, damage, typ)
						}
						runtime.ProjectileReflect(weapon, target)
						if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 == 0 {
							if runtime.ClearOwner == nil || runtime.SetOwner == nil {
								return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY shield owner service", target, source, weapon, damage, typ)
							}
							runtime.ClearOwner(weapon)
							runtime.SetOwner(target, weapon)
						}
					}
					if runtime.BlockDamagePercent == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY shield balance service", target, source, weapon, damage, typ)
					}
					amount := float32(runtime.BlockDamagePercent() * float64(damage))
					// 004E1BFD selects live inventory after sound/reflection/
					// balance; nil/no-health wear is a successful no-op.
					shield := playerDamageShieldItem4E17B0(target)
					if shield != nil && (runtime.CanDamageBlockItem == nil || !runtime.CanDamageBlockItem(shield)) {
						return playerDamageUnsupported4E17B0(runtime, "player ZAP_RAY shield durability callback", target, source, weapon, damage, typ)
					}
					if runtime.DamageBlockItem == nil {
						return playerDamageUnsupported4E17B0(runtime, "missing player ZAP_RAY shield wear service", target, source, weapon, damage, typ)
					}
					if !runtime.DamageBlockItem(shield, target, source, weapon, amount, typ) && runtime.Unsupported != nil {
						runtime.Unsupported("player ZAP_RAY shield durability failed", target, source, weapon, damage, typ)
					}
					if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
						if runtime.PlayerSetState == nil || !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
							return playerDamageUnsupported4E17B0(runtime, "unsupported live player ZAP_RAY broken-shield state", target, source, weapon, damage, typ)
						}
						runtime.PlayerSetState(target, PlayerState13)
					}
					return true, false
				}
			}
		}
		if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil ||
			!weapon.Class().Has(object.ClassSimple) || !weapon.Class().Has(object.ClassImmobile) ||
			weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live player ZAP_RAY tail record", target, source, weapon, damage, typ)
		}
		// Nonmissile type 16 skips GreatSword and armor/carry. Raw zero
		// stays zero; the later Quest minimum applies only to positive input.
		if update.Field76 == 0 {
			update.Field76, update.Field75 = 2, uint32(typ)
		}
		if runtime.GodMode != nil && runtime.GodMode() && target.Class().Has(object.ClassPlayer) {
			return true, true
		}
		effective := damage
		if runtime.QuestMode != nil && runtime.QuestMode() {
			if runtime.QuestDamageScale == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing live player ZAP_RAY quest damage service", target, source, weapon, damage, typ)
			}
			effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
			if damage > 0 && effective < 1 {
				effective = 1
			}
		}
		if runtime.DefaultDamage == nil || !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live player ZAP_RAY default service/record", target, source, weapon, damage, typ)
		}
		return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
	}
	pierceArmorValue := math.Float32frombits(update.Field57)
	greatSword := playerDamageGreatSwordContext4E17B0{
		weaponFlags: player.WeaponEquip, armorFlags: player.ArmorEquip,
		state: &update.State, marker: &update.Field76, markerType: &update.Field75,
	}
	pierceMissile := playerDamageMissilePierceShape4E17B0(source, weapon, typ)
	flameMissile := playerDamageMissileFlameShape4E17B0(source, weapon, typ)
	explosionMissile := playerDamageMissileExplosionShape4E17B0(source, weapon, typ)
	electricHit := playerDamageElectricShape4E17B0(source, weapon, typ)
	bite := typ == object.DamageBite && damage > 0 && source != nil && weapon != nil && source == weapon &&
		source.ObjClass.Has(object.ClassMonster) && source.UpdateData != nil
	missileImpact := typ == object.DamageImpact && damage > 0 && source != nil && weapon != nil && source != weapon &&
		source.ObjClass.Has(object.ClassMonster) && source.UpdateData != nil && weapon.ObjClass.Has(object.ClassMissile)
	monsterImpact := playerDamageMonsterImpactShape4E17B0(source, weapon, typ)
	monsterSelfStrike := playerDamageMonsterSelfStrikeShape4E17B0(source, weapon, typ)
	playerCharge := typ == object.DamageCrush && damage > 0 && source != nil && source == weapon &&
		source.ObjClass.Has(object.ClassPlayer) && !source.ObjClass.HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand)
	fullArmorHit := bite || missileImpact || monsterImpact || monsterSelfStrike
	armorPrefixHit := fullArmorHit || playerCharge
	nativePrefixHit := armorPrefixHit
	observe := player.ObserveTarget() != nil
	if electricHit || pierceMissile || flameMissile || explosionMissile || nativePrefixHit {
		excluded := runtime.BlockSourceExcluded
		if weapon == nil {
			excluded = runtime.BlockSourceOnlyExcluded
		}
		// The existing normal GreatSword defense can finish before the HP
		// switch without DefaultDamage. Its unblocked tail still checks that
		// service; do not make an unused HP callback a block prerequisite.
		// BITE, monster missile IMPACT and player CRUSH retain their native
		// HP tail. The generic ZAP_RAY prefix returned above.
		defaultRequired := !nativePrefixHit && (electricHit || observe || greatSword.weaponFlags&0x400 == 0)
		if (observe && runtime.ObserveClear == nil) || excluded == nil || runtime.BlockDirection == nil || (defaultRequired && runtime.DefaultDamage == nil) {
			reason := "missing player missile prefix service"
			if electricHit {
				reason = "missing player electric prefix service"
			} else if bite {
				reason = "missing player bite prefix service"
			} else if missileImpact || monsterImpact || monsterSelfStrike {
				reason = "missing player impact prefix service"
			} else if playerCharge {
				reason = "missing player charge prefix service"
			}
			return playerDamageUnsupported4E17B0(runtime, reason, target, source, weapon, damage, typ)
		}
		if electricHit && (runtime.ElectricArmorScale == nil || !playerDamageArmorReady4E17B0(target, runtime)) {
			return playerDamageUnsupported4E17B0(runtime, "missing player electric armor service", target, source, weapon, damage, typ)
		}
		if !electricHit && !nativePrefixHit && !observe && !playerDamageArmorReady4E17B0(target, runtime) {
			// Preserve normal missile admission's read-only armor checks before
			// moving its marker stores into the common entry prefix.
			return playerDamageUnsupported4E17B0(runtime, "missing player missile armor service", target, source, weapon, damage, typ)
		}
		if !nativePrefixHit && runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
			return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
		}
		runtime.playerPrefix = &playerDamagePrefix4E17B0{
			update: update, armorFlags: greatSword.armorFlags, weaponFlags: greatSword.weaponFlags,
		}
		// 004E18C4 clears the cached marker even without possession, before
		// Reflect or exclusion/facing callbacks. ObserveClear is conditional;
		// it can change live records while the marker base stays cached.
		update.Field76 = 0
		if observe {
			runtime.ObserveClear(target)
		}
		if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live possessed player record", target, source, weapon, damage, typ)
		}
	}
	if applicable, handled, result := playerDamageReflectShield4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return handled, result
	}
	prefixFront := false
	if runtime.playerPrefix != nil {
		// 004E1A49 snapshots the weapon position before the six exclusions.
		// Only afterwards does 004E1AA2 attribute a distinct live weapon.
		// Source-only hits use four exclusions at 004E1ACE; self-weapon
		// and source-only electric/player CRUSH hits retain the clear cached marker.
		attack, exclusion := weapon, runtime.BlockSourceExcluded
		if attack == nil {
			attack, exclusion = source, runtime.BlockSourceOnlyExcluded
		}
		pos := attack.PrevPos
		excluded := exclusion(attack)
		if weapon != nil && source != weapon && target.Class().Has(object.ClassPlayer) {
			update.Field76, update.Field75 = 1, uint32(weapon.TypeInd)
		}
		prefixFront = !excluded && runtime.BlockDirection(target, pos)
		// Both specialized defenses consume the original pair's answer;
		// never invoke the external services twice or reset the prefix again.
		runtime.BlockSourceExcluded = func(*Object) bool { return excluded }
		runtime.BlockSourceOnlyExcluded = runtime.BlockSourceExcluded
		runtime.BlockDirection = func(*Object, types.Pointf) bool { return prefixFront }
	}
	if armorPrefixHit && (!target.Class().Has(object.ClassPlayer) || target.UpdateData == nil) {
		// Exclusion/facing can replace the live record. The entry marker and
		// equipment stay cached, but never read a non-player/nil live carry.
		reason := "unsupported live player bite record"
		if missileImpact || monsterImpact || monsterSelfStrike {
			reason = "unsupported live player impact record"
		} else if playerCharge {
			reason = "unsupported live player charge record"
		}
		return playerDamageUnsupported4E17B0(runtime, reason, target, source, weapon, damage, typ)
	}
	if monsterSelfStrike {
		return playerDamageMonsterSelfStrikeTail4E17B0(target, source, weapon, greatSword, pierceArmorValue, damage, typ, prefixFront, runtime)
	}
	if monsterImpact {
		// A Troll's ordinary strike passes itself as both source and weapon.
		// Full armor/carry is shared with PIERCE, but the marker stays clear
		// through defenses and no projectile reflection is applicable.
		return playerDamageMonsterImpactTail4E17B0(target, source, weapon, greatSword, pierceArmorValue, damage, typ, runtime)
	}
	if applicable, handled, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, greatSword, damage, typ, runtime); applicable {
		return handled, result
	}
	if (runtime.playerPrefix != nil && flameMissile) || playerDamageMissileFlameShape4E17B0(source, weapon, typ) {
		return playerDamagePlayerMissileFlame4E17B0(target, source, weapon, update, damage, typ, runtime)
	}
	if (runtime.playerPrefix != nil && explosionMissile) || playerDamageMissileExplosionShape4E17B0(source, weapon, typ) {
		return playerDamagePlayerMissileExplosion4E17B0(target, source, weapon, update, pierceArmorValue, damage, typ, runtime)
	}
	if (runtime.playerPrefix != nil && pierceMissile) || playerDamageMissilePierceShape4E17B0(source, weapon, typ) {
		// Stock GolemArrow is also WEAPON, subclass 0x10. Admit its
		// ranged predicate without silently including melee Shock shapes.
		// 004E1F84 accepts either unit source and raw signed damage; the
		// positive-only minimum belongs after armor/carry at 004E2011.
		return playerDamageMissilePierce4E17B0(target, source, weapon, update, pierceArmorValue, damage, typ, runtime)
	}
	if (runtime.playerPrefix != nil && electricHit) || playerDamageElectricShape4E17B0(source, weapon, typ) {
		// Ordinary shield/sword blocks explicitly exclude type 9/17.
		// 004E1DF1 accepts raw signed damage, including zero. The
		// positive-only minimum belongs after scale/carry/wear at 004E2011;
		// DefaultDamage's later electric protection has its own minimum.
		return playerDamageElectricPlayer4E17B0(target, source, weapon, damage, typ, runtime)
	}
	lava := typ == object.DamageLava && damage > 0 && source == nil && weapon == nil
	poison := typ == object.DamagePoison && damage > 0 && source == nil && weapon == nil
	flame := typ == object.DamageFlame && damage > 0 && weapon != nil &&
		(source == nil || source == weapon) && weapon.ObjClass.Has(object.ClassFire)
	if !flame && !lava && !poison && !bite && !missileImpact && !playerCharge {
		return playerDamageUnsupported4E17B0(runtime, "unsupported player damage shape", target, source, weapon, damage, typ)
	}
	vampirism := (bite || missileImpact || playerCharge) && source.HasEnchant(damageVampirismEnchant4E0B30)
	if playerCharge && prefixFront {
		// 004E1B56 uses the cached equipment masks and post-facing stance,
		// then live inventory. A successful block never needs the HP tail's
		// friendly-fire or hurt services and does not read Quest/GodMode.
		// 004E1B4A skips this branch before any state/layout reads on a rear
		// hit, even if the facing callbacks changed the live target class.
		if applicable, handled, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
			return handled, result
		}
	}
	if playerCharge {
		// PLAYER charge weapons also make sub_4E1400 false. Only the
		// general owner/friendly-fire gate applies; no melee Shock retaliation.
		if runtime.GameplayFlag1 == nil || runtime.IsEnemy == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing friendly-fire service", target, source, weapon, damage, typ)
		}
		if playerCharge && damage >= 20 && runtime.PlayerSetState == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing player hurt-state service", target, source, weapon, damage, typ)
		}
	}
	if bite || missileImpact {
		if applicable, handled, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
			return handled, result
		}
	}
	// With a scale service present, BITE, monster IMPACT, player CRUSH and Sentry read
	// Quest only at 004E2046, after armor/marker/minimum and GodMode.
	// A missing service may still
	// reject the unblocked tail before carry/wear; it is not a block input.
	quest := !nativePrefixHit && runtime.QuestMode != nil && runtime.QuestMode()
	if nativePrefixHit && runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	shielded := !poison && target.HasEnchant(playerDamageShieldEnchant4E17B0) &&
		(typ != object.DamageManaBomb || source != target)
	if shielded && runtime.ShieldReduce == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing Shield reduction service", target, source, weapon, damage, typ)
	}
	if target.DamageSound != nil && target.DamageSound != runtime.PlayerDamageSoundC {
		return playerDamageUnsupported4E17B0(runtime, "custom player damage sound", target, source, weapon, damage, typ)
	}
	hasLateDefend, ok := playerDamageLateDefendReady4E17B0(target, runtime)
	if !ok {
		return playerDamageUnsupported4E17B0(runtime, "late equipped-item defend effect", target, source, weapon, damage, typ)
	}
	if bite && (runtime.IsEnemy == nil || !runtime.IsEnemy(target, source)) {
		return playerDamageUnsupported4E17B0(runtime, "non-enemy source", target, source, weapon, damage, typ)
	}
	if (flame || lava) && runtime.FireProtection == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing source-less damage service", target, source, weapon, damage, typ)
	}
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	if vampirism && (runtime.Audio == nil || runtime.BalanceFloatInd == nil ||
		runtime.AdjustHP == nil || runtime.VampirismFX == nil) {
		return playerDamageUnsupported4E17B0(runtime, "missing Vampirism service", target, source, weapon, damage, typ)
	}

	armorValue := math.Float32frombits(update.Field57)
	carryUpdate := update
	if armorPrefixHit {
		if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			reason := "unsupported live player bite carry"
			if missileImpact {
				reason = "unsupported live player impact carry"
			} else if playerCharge {
				reason = "unsupported live player charge carry"
			}
			return playerDamageUnsupported4E17B0(runtime, reason, target, source, weapon, damage, typ)
		}
		// 004E1F84 consumes entry armor (half at 004E1EE8 for CRUSH);
		// 004E20F0 reloads the live carry,
		// independently of the cached marker/update used by the prefix.
		armorValue, carryUpdate = pierceArmorValue, target.UpdateDataPlayer()
	}
	effective := damage
	remaining := damage
	accumulated := math.Float32frombits(carryUpdate.Field21)
	armorReduced := bite || missileImpact || playerCharge
	if armorReduced {
		absorption := float64(armorValue)
		if playerCharge {
			// Original CRUSH case 004E1EE8 uses half the armor absorption;
			// absorbed damage still goes through the usual durability pass.
			absorption *= 0.5
		}
		armored := float32((1.0 - absorption) * float64(damage))
		accumulated = armored + accumulated
		effective = playerDamageRound4E17B0(accumulated)
		remaining = damage - effective
	}
	damageItems := !poison
	if !damageItems {
		// POISON is case 5: it changes the player damage marker but does
		// not run the armor-durability pass.
		remaining = 0
	}
	if damageItems && !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	if effective == 0 {
		effective = 1
	}
	if runtime.DamageClear == nil || (!poison && !flame && runtime.BuffOff == nil) {
		return playerDamageUnsupported4E17B0(runtime, "missing native damage service", target, source, weapon, damage, typ)
	}
	// DefaultDamage releases a carried GameBall using damage after armor,
	// Quest scaling and late defend, but before Shield absorption. Validate
	// that tail before any stores; callback-dependent adjustments may raise
	// a sub-threshold amount, so they also require the service for carriers.
	// These full-armor hits' late Quest/Defend can introduce a drop after this
	// read-only admission. The final live GameBall guard below owns that amount.
	mayDropBall := effective >= 30 || quest || hasLateDefend || flame || lava
	if mayDropBall && target.Field129 != nil {
		if runtime.GameBallType == 0 {
			return playerDamageUnsupported4E17B0(runtime, "missing GameBall type", target, source, weapon, damage, typ)
		}
		if playerDamageOwnsType4E17B0(target, runtime.GameBallType) {
			if runtime.GameBallOnDamage == nil {
				return playerDamageUnsupported4E17B0(runtime, "GameBall drop", target, source, weapon, damage, typ)
			}
			// The original high-damage release dereferences the attacker's
			// team after detaching the ball. Do not partially mutate a
			// source-less carrier on that still-unsupported faulting path.
			if source == nil {
				return playerDamageUnsupported4E17B0(runtime, "source-less GameBall drop", target, source, weapon, damage, typ)
			}
		}
	}

	if runtime.playerPrefix == nil {
		update.Field76 = 0
		if player.ObserveTarget() != nil && runtime.ObserveClear != nil {
			runtime.ObserveClear(target)
		}
		if source != nil && weapon != nil && source != weapon {
			update.Field76, update.Field75 = 1, uint32(weapon.TypeInd)
		}
	}
	if armorReduced {
		carryUpdate.Field21 = math.Float32bits(accumulated - float32(playerDamageRound4E17B0(accumulated)))
	}
	if damageItems {
		playerDamageApplyArmor4E17B0(target, source, weapon, remaining, typ, runtime)
	}
	if update.Field76 == 0 && (!nativePrefixHit || target.Class().Has(object.ClassPlayer)) {
		// 004E1E49/004E1EAE/004E1F4D/004E1FE1 copy the raw DWORD,
		// and preserve a hit marker left by an earlier armor callback.
		update.Field76, update.Field75 = 2, uint32(typ)
	}

	if runtime.GodMode != nil && runtime.GodMode() && (!nativePrefixHit || target.Class().Has(object.ClassPlayer)) {
		return true, true
	}
	if nativePrefixHit {
		quest = runtime.QuestMode != nil && runtime.QuestMode()
		if quest && runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing live quest damage service", target, source, weapon, damage, typ)
		}
	}
	if quest {
		before := effective
		scaled := float32(float64(runtime.QuestDamageScale()) * float64(effective))
		effective = playerDamageRound4E17B0(scaled)
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if playerCharge && !runtime.GameplayFlag1() {
		owner := source.FindOwnerChainPlayer()
		if owner != nil && owner.Class().HasAny(object.MaskUnits) &&
			!runtime.IsEnemy(target, owner) && (target != owner || quest) {
			return true, true
		}
	}
	if flame || lava || poison {
		// PlayerDamage calls DefaultDamage after the damage-type switch, so
		// the invulnerability gate is observed a second time in the original.
		if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
			if byte(frame)&3 == 0 && runtime.Audio != nil {
				runtime.Audio(playerDamageInvulnerableSound4E17B0, target)
			}
			return true, true
		}
		if flame || lava {
			protectionValue := runtime.FireProtection(target)
			if protectionValue != 0 && byte(frame)&3 == 0 && runtime.Audio != nil {
				runtime.Audio(104, target)
			}
			protection := float32(protectionValue)
			scaled := float32((1.0 - float64(protection)) * float64(effective))
			effective = playerDamageRound4E17B0(scaled)
			if effective == 0 {
				effective = 1
			}
		}
		target.Pos132 = types.Pointf{}
	} else {
		target.Pos132 = weapon.PrevPos
	}
	if !poison && !flame {
		runtime.BuffOff(target, playerDamageInvisibleEnchant4E17B0)
	}
	// 004E2098 reaches DefaultDamage; its 004E0F77 calls 004E1320 here,
	// after armor, Quest scaling, attribution position and BuffOff. Recapture
	// the inventory head now, cache each item's init base, and load slot three
	// and the next link after callbacks rather than executing an eager plan.
	ItemDefendEffects4E1320(target, source, weapon, &effective, int32(typ), ItemDefendEffectsRuntime4E1320{
		ApplyDefend: func(modifier *ModifierEff, item, owner, weapon, attacker *Object, context *[2]int32) {
			// Earlier armor/BuffOff/Defend callbacks may introduce a new
			// function after admission. Report it at use without PE32 fallback,
			// leaving damage unchanged and still visiting supported later slots.
			if runtime.CanApplyLateDefend == nil || runtime.ApplyLateDefend == nil || !runtime.CanApplyLateDefend(modifier) {
				if runtime.Unsupported != nil {
					runtime.Unsupported("unsupported live late equipped-item defend effect", owner, attacker, weapon, context[0], object.DamageType(context[1]))
				}
				return
			}
			context[0] = runtime.ApplyLateDefend(modifier, item, owner, weapon, attacker, context[0], object.DamageType(context[1]))
		},
	})
	target.Obj130 = weapon
	target.Field131 = uint32(typ)
	target.Frame134 = frame

	if lava || poison {
		if runtime.PlayerDamageSound != nil {
			runtime.PlayerDamageSound(target, nil)
		}
	} else if flame || playerCharge {
		if runtime.PlayerDamageSound != nil {
			runtime.PlayerDamageSound(target, weapon)
		}
	} else {
		monsterUpdate := source.UpdateDataMonster()
		monsterHasHitSound := false
		if monsterUpdate.SoundSet122 != nil {
			monsterHasHitSound = *(*uint32)(unsafe.Add(monsterUpdate.SoundSet122, 8*4)) != 0
		}
		if !monsterHasHitSound && runtime.PlayerDamageSound != nil {
			runtime.PlayerDamageSound(target, weapon)
		}
	}
	if vampirism {
		damageApplyVampirism4E0B30(
			source, target, weapon, effective,
			runtime.Audio, runtime.BalanceFloatInd, runtime.AdjustHP, runtime.VampirismFX,
		)
	}
	// A newly installed live Defend can raise damage above the drop threshold
	// even when admission saw no effect. Do not silently skip the required
	// service or enter its source-less fault after the already-executed prefix.
	if effective >= 30 && target.Field129 != nil {
		if runtime.GameBallType == 0 {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live GameBall type", target, source, weapon, effective, typ)
		}
		if playerDamageOwnsType4E17B0(target, runtime.GameBallType) {
			if runtime.GameBallOnDamage == nil {
				return playerDamageUnsupported4E17B0(runtime, "unsupported live GameBall drop", target, source, weapon, effective, typ)
			}
			if source == nil {
				return playerDamageUnsupported4E17B0(runtime, "unsupported live source-less GameBall drop", target, source, weapon, effective, typ)
			}
		}
	}
	if runtime.GameBallOnDamage != nil {
		runtime.GameBallOnDamage(source, target, effective)
	}
	if bite || missileImpact {
		monsterUpdate := source.UpdateDataMonster()
		if monsterUpdate.Field130 == 0 {
			monsterUpdate.Field130 = frame
		}
	}
	if playerCharge && effective >= 20 && target.Class().Has(object.ClassPlayer) {
		// DefaultDamage's 004E1136/004E1147 reload class/update/state after
		// Defend, sound, Vampirism and GameBall. Entry state is not a hurt gate.
		if target.UpdateData == nil {
			reason := "unsupported live player charge hurt record"
			return playerDamageUnsupported4E17B0(runtime, reason, target, source, weapon, effective, typ)
		}
		state := target.UpdateDataPlayer().State
		if state != PlayerState1 && state != PlayerState15 {
			if runtime.PlayerSetState == nil {
				reason := "missing live player charge hurt-state service"
				return playerDamageUnsupported4E17B0(runtime, reason, target, source, weapon, effective, typ)
			}
			runtime.PlayerSetState(target, PlayerState30)
		}
	}
	if shielded {
		shieldSource := weapon
		if shieldSource == nil {
			shieldSource = source
		}
		runtime.ShieldReduce(target, &effective, typ, shieldSource)
		if effective == 0 {
			return true, false
		}
	}
	runtime.DamageClear(target, effective)
	return true, true
}
