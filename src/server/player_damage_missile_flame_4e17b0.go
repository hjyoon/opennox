package server

import (
	"github.com/opennox/libs/object"
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
	// The entry already applied the real GreatSword defense. A rear hit or
	// an ineligible NPC action reaches FLAME's original raw-durability tail.
	return playerDamageMissileFlameTail4E17B0(target, source, weapon, &update.Field547, &update.Field546, damage, typ, runtime)
}

func playerDamagePlayerMissileFlame4E17B0(
	target, source, weapon *Object, update *PlayerUpdateData,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	// Reflect Shield and GreatSword already ran in the entry. A real shield
	// posture still precedes FLAME's damage switch.
	if applicable, h, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return h, result
	}
	if update.Player.ObserveTarget() != nil {
		return playerDamageUnsupported4E17B0(runtime, "possessed player missile FLAME", target, source, weapon, damage, typ)
	}
	return playerDamageMissileFlameTail4E17B0(target, source, weapon, &update.Field76, &update.Field75, damage, typ, runtime)
}
