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

	if target.Class().Has(object.ClassPlayer) {
		update := target.UpdateDataPlayer()
		update.Field76 = 0
		if update.Player.ObserveTarget() != nil && runtime.ObserveClear != nil {
			runtime.ObserveClear(target)
		}
	} else {
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
		update = target.UpdateDataPlayer()
		player := update.Player
		if player.ArmorEquip&0x3000000 == 0 {
			return false, false, false
		}
		shieldStance := update.State == PlayerState16
		if !shieldStance && update.State == PlayerState1 && player.WeaponEquip&0x400 == 0 {
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
	if playerTarget {
		update.Field76 = 0
		if update.Player.ObserveTarget() != nil && runtime.ObserveClear != nil {
			runtime.ObserveClear(target)
		}
		if typ == object.DamageImpale && source != nil && weapon != nil && source != weapon {
			// Existing non-PIERCE player slices keep their separate marker contracts.
			update.Field76 = 1
			update.Field75 = uint32(weapon.TypeInd)
		}
	} else {
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
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, source, weapon, damage, typ)
	}
	if quest && runtime.QuestDamageScale == nil {
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
	if quest {
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
	if target.UpdateData == nil || uint32(target.SubClass())&0x10 == 0 {
		return playerDamageUnsupported4E17B0(runtime, "unsupported monster target", target, source, weapon, damage, typ)
	}
	if target.ObjFlags.HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	frame := uint32(0)
	if runtime.Frame != nil {
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
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if runtime.ElectricArmorScale == nil || runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing player electric service", target, source, weapon, damage, typ)
	}
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	update := target.UpdateDataPlayer()
	scaled := float32(float64(runtime.ElectricArmorScale(target)) * float64(damage))
	accumulated := scaled + math.Float32frombits(update.Field21)
	effective := playerDamageRound4E17B0(accumulated)
	if !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	update.Field76 = 0
	if update.Player.ObserveTarget() != nil && runtime.ObserveClear != nil {
		runtime.ObserveClear(target)
	}
	update.Field21 = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, weapon, damage, typ, runtime)
	if update.Field76 == 0 {
		update.Field76 = 2
		// 004E1E49 copies the incoming DWORD type, not float32(type).
		update.Field75 = uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if quest {
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
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, source, weapon, damage, typ)
	}
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	// GreatStaff can reflect missiles in states 13/18/19/20 before this
	// switch. Do not silently turn that separate, unported defense into HP
	// damage, including when its facing predicate has not been evaluated.
	if update.Player.WeaponEquip&0x400 != 0 &&
		(update.State == PlayerState13 || update.State == PlayerState18 || update.State == PlayerState19 || update.State == PlayerState20) {
		return playerDamageUnsupported4E17B0(runtime, "player GreatStaff block", target, source, weapon, damage, typ)
	}
	if applicable, h, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return h, result
	}
	if update.Player.ObserveTarget() != nil {
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
	update.Field76 = 0
	update.Field76 = 1
	update.Field75 = uint32(weapon.TypeInd)
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
	if quest {
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
// missile IMPACT, player/monster-fired missile PIERCE, Berserker Charge CRUSH,
// SentryGlobe ZAP_RAY, world FLAME,
// unarmed player/monster and unit-self-weapon ELECTRIC/AIRBORNE_ELECTRIC,
// and source-less LAVA/POISON branches of
// GAME.EXE 004E17B0 together with their relevant unit-default-damage tails,
// plus the front-facing shield block and the common Quest damage scaling tail,
// and the early Reflect Shield and Coop self-damage gates. It returns
// handled=false before mutation for spell and modifier branches that remain
// separate ports.
func PlayerDamageNative4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if target == nil {
		return playerDamageUnsupported4E17B0(runtime, "non-player target", target, source, weapon, damage, typ)
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
	frame := uint32(0)
	if runtime.Frame != nil {
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
	pierceArmorValue := math.Float32frombits(update.Field57)
	if applicable, handled, result := playerDamageReflectShield4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return handled, result
	}
	if typ == object.DamageImpale && source != nil && source != weapon &&
		source.Class().HasAny(object.ClassPlayer|object.ClassMonster) && source.UpdateData != nil &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
		!defaultDamageAttackQualifies4E1400(source, weapon) {
		// Stock GolemArrow is also WEAPON, subclass 0x10. Admit its
		// ranged predicate without silently including melee Shock shapes.
		// 004E1F84 accepts either unit source and raw signed damage; the
		// positive-only minimum belongs after armor/carry at 004E2011.
		return playerDamageMissilePierce4E17B0(target, source, weapon, update, pierceArmorValue, damage, typ, runtime)
	}
	if damage > 0 && playerDamageElectricShape4E17B0(source, weapon, typ) {
		// Ordinary shield/sword blocks explicitly exclude type 9/17.
		return playerDamageElectricPlayer4E17B0(target, source, weapon, damage, typ, runtime)
	}
	lava := typ == object.DamageLava && damage > 0 && source == nil && weapon == nil
	poison := typ == object.DamagePoison && damage > 0 && source == nil && weapon == nil
	flame := typ == object.DamageFlame && damage > 0 && weapon != nil &&
		(source == nil || source == weapon) && weapon.ObjClass.Has(object.ClassFire)
	bite := typ == object.DamageBite && damage > 0 && source != nil && weapon != nil && source == weapon &&
		source.ObjClass.Has(object.ClassMonster) && source.UpdateData != nil
	missileImpact := typ == object.DamageImpact && damage > 0 && source != nil && weapon != nil && source != weapon &&
		source.ObjClass.Has(object.ClassMonster) && source.UpdateData != nil && weapon.ObjClass.Has(object.ClassMissile)
	playerCharge := typ == object.DamageCrush && damage > 0 && source != nil && source == weapon &&
		source.ObjClass.Has(object.ClassPlayer) && !source.ObjClass.HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand)
	sentryZapRayCandidate := typ == object.DamageZapRay && damage > 0 && source != nil && weapon != nil &&
		source.ObjClass.Has(object.ClassPlayer)
	if sentryZapRayCandidate && runtime.SentryGlobeType == 0 {
		return playerDamageUnsupported4E17B0(runtime, "missing SentryGlobe type", target, source, weapon, damage, typ)
	}
	sentryZapRay := sentryZapRayCandidate && weapon.TypeInd == runtime.SentryGlobeType
	if !flame && !lava && !poison && !bite && !missileImpact && !playerCharge && !sentryZapRay {
		return playerDamageUnsupported4E17B0(runtime, "unsupported player damage shape", target, source, weapon, damage, typ)
	}
	vampirism := (bite || missileImpact || playerCharge || sentryZapRay) && source.HasEnchant(damageVampirismEnchant4E0B30)
	if sentryZapRay {
		// sub_4E1400 is false for the actual SentryGlobe class. Keeping the
		// accepted shape equally narrow avoids silently skipping its separate
		// friendly-hit and Shock-retaliation branches for weapon-like objects.
		if weapon.ObjClass.HasAny(object.ClassMonster | object.ClassWeapon | object.ClassWand) {
			return playerDamageUnsupported4E17B0(runtime, "unexpected SentryGlobe class", target, source, weapon, damage, typ)
		}
	}
	if sentryZapRay || playerCharge {
		// PLAYER charge weapons also make sub_4E1400 false. Only the
		// general owner/friendly-fire gate applies; no melee Shock retaliation.
		if runtime.GameplayFlag1 == nil || runtime.IsEnemy == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing friendly-fire service", target, source, weapon, damage, typ)
		}
		if damage >= 20 && runtime.PlayerSetState == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing player hurt-state service", target, source, weapon, damage, typ)
		}
	}
	if bite || missileImpact || playerCharge || sentryZapRay {
		if applicable, handled, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
			return handled, result
		}
	}
	quest := runtime.QuestMode != nil && runtime.QuestMode()
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
	effective := damage
	remaining := damage
	accumulated := math.Float32frombits(update.Field21)
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
	damageItems := !poison && !sentryZapRay
	if !damageItems {
		// POISON and ZAP_RAY are cases 5 and 16 in the original switch: they
		// change the player damage marker but do not run the armor-durability
		// pass.
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

	update.Field76 = 0
	if player.ObserveTarget() != nil && runtime.ObserveClear != nil {
		runtime.ObserveClear(target)
	}
	if source != nil && weapon != nil && source != weapon {
		update.Field76, update.Field75 = 1, uint32(weapon.TypeInd)
	}
	if armorReduced {
		update.Field21 = math.Float32bits(accumulated - float32(playerDamageRound4E17B0(accumulated)))
	}
	if damageItems {
		playerDamageApplyArmor4E17B0(target, source, weapon, remaining, typ, runtime)
	}
	if update.Field76 == 0 {
		// 004E1E49/004E1EAE/004E1F4D/004E1FE1 copy the raw DWORD,
		// and preserve a hit marker left by an earlier armor callback.
		update.Field76, update.Field75 = 2, uint32(typ)
	}

	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if quest {
		before := effective
		scaled := float32(float64(runtime.QuestDamageScale()) * float64(effective))
		effective = playerDamageRound4E17B0(scaled)
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if (sentryZapRay || playerCharge) && !runtime.GameplayFlag1() {
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
	} else if flame || playerCharge || sentryZapRay {
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
	if (sentryZapRay || playerCharge) && effective >= 20 && update.State != PlayerState1 && update.State != PlayerState15 {
		runtime.PlayerSetState(target, PlayerState30)
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
