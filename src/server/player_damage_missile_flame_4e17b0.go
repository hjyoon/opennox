package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// SparkExplosionCollide supplies the owning unit and a distinct missile for
// direct FLAME hits, but MapDamageUnits supplies the missile and nil weapon for
// splash. Neither shape qualifies for the original melee/Shock predicate.
func playerDamageMissileFlameShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageFlame || source == nil {
		return false
	}
	missile := source
	if weapon != nil {
		if source == weapon || source.UpdateData == nil || !source.Class().HasAny(object.ClassPlayer|object.ClassMonster) {
			return false
		}
		missile = weapon
	}
	return missile.Class().Has(object.ClassMissile) &&
		!missile.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand)
}

// 004E1DC7 gives FLAME's full signed amount to armor durability. Unlike
// physical damage, it neither absorbs armor nor reads/writes fractional carry.
// Hit markers retain the update base cached before the defense callbacks.
func playerDamageMissileFlameTail4E17B0(
	target, source, weapon *Object, marker, markerType *uint32,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing missile FLAME default damage", target, source, weapon, damage, typ)
	}
	if runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "missing missile FLAME quest scale", target, source, weapon, damage, typ)
	}
	if !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "missile FLAME armor durability callback", target, source, weapon, damage, typ)
	}
	*marker = 0
	if weapon != nil && source != weapon {
		*marker, *markerType = 1, uint32(weapon.TypeInd)
	}
	playerDamageApplyArmor4E17B0(target, source, weapon, damage, typ, runtime)
	if *marker == 0 {
		*marker, *markerType = 2, uint32(typ)
	}
	// 004E2025 tests the live PLAYER class after armor callbacks. NPCs
	// must not inherit the player's GodMode exemption.
	if target.Class().Has(object.ClassPlayer) && runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	effective := damage
	// Quest scaling occurs after durability, before DefaultDamage's fire
	// protection, late Defend, Shield absorption and real HP callback.
	if runtime.QuestMode != nil && runtime.QuestMode() {
		if runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing live missile FLAME quest scale", target, source, weapon, damage, typ)
		}
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(damage)))
		if damage > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

func playerDamageMonsterMissileFlame4E17B0(
	target, source, weapon *Object, update *MonsterUpdateData,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	// 004E1C0F's GreatStaff reflection is a separate defense, not ordinary
	// armor absorption. Never silently replace an unported block with HP loss.
	if update.WeaponEquipFlags&0x400 != 0 && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK {
		return playerDamageUnsupported4E17B0(runtime, "NPC GreatStaff missile FLAME block", target, source, weapon, damage, typ)
	}
	return playerDamageMissileFlameTail4E17B0(target, source, weapon, &update.Field547, &update.Field546, damage, typ, runtime)
}

func playerDamagePlayerMissileFlame4E17B0(
	target, source, weapon *Object, update *PlayerUpdateData,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	// Ordinary shield blocks precede the separate GreatStaff defense and
	// FLAME switch. Reflect Shield has already run in the entry function.
	if applicable, h, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return h, result
	}
	if update.Player.WeaponEquip&0x400 != 0 &&
		(update.State == PlayerState13 || update.State == PlayerState18 || update.State == PlayerState19 || update.State == PlayerState20) {
		return playerDamageUnsupported4E17B0(runtime, "player GreatStaff missile FLAME block", target, source, weapon, damage, typ)
	}
	if update.Player.ObserveTarget() != nil {
		return playerDamageUnsupported4E17B0(runtime, "possessed player missile FLAME", target, source, weapon, damage, typ)
	}
	return playerDamageMissileFlameTail4E17B0(target, source, weapon, &update.Field76, &update.Field75, damage, typ, runtime)
}
