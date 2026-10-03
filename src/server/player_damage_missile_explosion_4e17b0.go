package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// BoomCollide sends a distinct Magic Missile and its owning unit for the
// direct hit; MapDamageUnits sends the missile with a nil weapon for splash.
// Pure spell missiles do not enter the original melee/Shock predicate.
func playerDamageMissileExplosionShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageExplosion || source == nil {
		return false
	}
	missile := source
	if weapon != nil {
		if source == weapon || source.UpdateData == nil || !source.Class().HasAny(object.ClassPlayer|object.ClassMonster) {
			return false
		}
		missile = weapon
	}
	return missile.Class().Has(object.ClassMissile) && !missile.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand)
}

// GAME.EXE case 7 enters 004E1F84, unlike FLAME's raw durability tail.
// Absorption is cached before defenses; 004E20F0 reloads live carry and
// 004E2180 reloads live armor. Durability precedes minimum, GodMode and Quest.
func playerDamageMissileExplosionTail4E17B0(
	target, source, weapon *Object, marker, markerType *uint32, armorValue float32,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing missile EXPLOSION default damage", target, source, weapon, damage, typ)
	}
	if runtime.QuestDamageScale == nil && runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "missing missile EXPLOSION quest scale", target, source, weapon, damage, typ)
	}
	if !playerDamageArmorReady4E17B0(target, runtime) {
		return playerDamageUnsupported4E17B0(runtime, "missile EXPLOSION armor durability callback", target, source, weapon, damage, typ)
	}
	if target.UpdateData == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing live missile EXPLOSION update", target, source, weapon, damage, typ)
	}
	var carry *uint32
	if target.Class().Has(object.ClassPlayer) {
		carry = &target.UpdateDataPlayer().Field21
	} else if target.Class().Has(object.ClassMonster) {
		carry = &target.UpdateDataMonster().Field1
	}
	scaled := float32((1.0 - float64(armorValue)) * float64(damage))
	accumulated := scaled
	if carry != nil {
		accumulated += math.Float32frombits(*carry)
	}
	effective := playerDamageRound4E17B0(accumulated)
	*marker = 0
	if weapon != nil && source != weapon {
		*marker, *markerType = 1, uint32(weapon.TypeInd)
	}
	if carry != nil {
		*carry = math.Float32bits(accumulated - float32(effective))
	}
	playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, runtime)
	if *marker == 0 {
		*marker, *markerType = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if target.Class().Has(object.ClassPlayer) && runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if runtime.QuestMode != nil && runtime.QuestMode() {
		if runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing live missile EXPLOSION quest scale", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

func playerDamageMonsterMissileExplosion4E17B0(
	target, source, weapon *Object, update *MonsterUpdateData, armorValue float32,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	return playerDamageMissileExplosionTail4E17B0(target, source, weapon, &update.Field547, &update.Field546, armorValue, damage, typ, runtime)
}

func playerDamagePlayerMissileExplosion4E17B0(
	target, source, weapon *Object, update *PlayerUpdateData, armorValue float32,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if applicable, h, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return h, result
	}
	if update.Player.WeaponEquip&0x400 != 0 &&
		(update.State == PlayerState13 || update.State == PlayerState18 || update.State == PlayerState19 || update.State == PlayerState20) {
		return playerDamageUnsupported4E17B0(runtime, "player GreatStaff missile EXPLOSION block", target, source, weapon, damage, typ)
	}
	if update.Player.ObserveTarget() != nil {
		return playerDamageUnsupported4E17B0(runtime, "possessed player missile EXPLOSION", target, source, weapon, damage, typ)
	}
	return playerDamageMissileExplosionTail4E17B0(target, source, weapon, &update.Field76, &update.Field75, armorValue, damage, typ, runtime)
}
