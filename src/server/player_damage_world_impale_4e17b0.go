package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// DamageCollide supplies a world hazard and its terminal parent. Stock Spike
// is IMMOBILE, not SIMPLE; SpikeBlock is SIMPLE, and its immobile variant is
// not. Their common admission is DANGEROUS/non-unit IMPALE (case 3), without a
// type ID or positive-only gate. Unit/weapon/wand/missile shapes stay disjoint.
func playerDamageWorldImpaleShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageImpale || source == nil || weapon == nil ||
		!weapon.Class().Has(object.ClassDangerous) ||
		weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) {
		return false
	}
	return source == weapon || (source.UpdateData != nil && source.Class().HasAny(object.MaskUnits) &&
		!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile))
}

// DamageCollide's case 3 shares the full-armor switch with case 11, but a
// world spike is neither a reflectable missile nor a sword/staff attack.
// Cache the entry update/equipment/absorption, retaining live callback reads.
func playerDamagePlayerWorldImpale4E17B0(
	target, source, weapon *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if target.Flags().HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if r.Frame == nil {
			return reject("missing world IMPALE invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing world IMPALE invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source.FindOwnerChainPlayer() == target && target.Class().Has(object.ClassPlayer) {
		return true, false
	}
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
		return reject("unsupported world IMPALE entry player record")
	}
	update := target.UpdateDataPlayer()
	if update.Player == nil {
		return reject("missing world IMPALE player info")
	}
	weaponFlags, armorFlags := update.Player.WeaponEquip, update.Player.ArmorEquip
	absorption := math.Float32frombits(update.Field57)
	if update.Player.Field3680&1 != 0 {
		return true, false
	}
	update.Field76 = 0
	if update.Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing world IMPALE observe service")
		}
		r.ObserveClear(target)
	}
	// Reflect27 skips a non-missile IMPALE hazard. A callback replacing it
	// with a missile belongs to that separate prefix, not a PE32 fallback.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && weapon.Class().Has(object.ClassMissile) {
		return reject("unsupported live world IMPALE Reflect projectile")
	}
	// 004E1A49 snapshots PrevPos before exclusion; attribution uses the
	// cached player update and weapon's live type after the exclusion call.
	pos := weapon.PrevPos
	if r.BlockSourceExcluded == nil {
		return reject("missing world IMPALE exclusion")
	}
	excluded := r.BlockSourceExcluded(weapon)
	if source != weapon {
		if !target.Class().Has(object.ClassPlayer) {
			return reject("unsupported live world IMPALE attribution record")
		}
		update.Field76, update.Field75 = 1, uint32(weapon.TypeInd)
	}
	front := false
	if !excluded {
		if r.BlockDirection == nil {
			return reject("missing world IMPALE block direction")
		}
		front = r.BlockDirection(target, pos)
	}
	if front {
		if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil ||
			!playerDamageWorldImpaleShape4E17B0(source, weapon, typ) {
			return reject("unsupported live world IMPALE block record")
		}
		// Both case 3 and case 11 enter the same ordinary shield branch
		// 004E1B98. Its audio/live-projectile/balance/live-shield order is
		// already restored; do not duplicate or change that function body.
		if armorFlags&0x3000000 != 0 && update.State == PlayerState16 {
			return playerDamageSpellMissileImpactShield4E17B0(target, source, weapon, damage, typ, r)
		}
		// The great-sword and staff alternatives cannot block a non-missile
		// case-3 attack. The original berserker shield exception still can.
		if weaponFlags&0x400 == 0 && update.State == PlayerState1 && armorFlags&0x3000000 != 0 {
			if r.BerserkShieldBlock == nil {
				return reject("missing world IMPALE berserker shield service")
			}
			if r.BerserkShieldBlock(target) {
				return playerDamageSpellMissileImpactShield4E17B0(target, source, weapon, damage, typ, r)
			}
		}
	}
	return playerDamageWorldImpaleTail4E17B0(target, source, weapon, update, absorption, damage, typ, r)
}

// 004E1F84..004E2098: binary32 absorption + LIVE carry, armor wear,
// cached marker, positive-only minimum, then LIVE GodMode/Quest and HP.
func playerDamageWorldImpaleTail4E17B0(
	target, source, weapon *Object, update *PlayerUpdateData, absorption float32,
	damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil ||
		!playerDamageWorldImpaleShape4E17B0(source, weapon, typ) {
		return reject("unsupported live world IMPALE armor/carry record")
	}
	if !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported world IMPALE armor durability")
	}
	scaled := float32((1.0 - float64(absorption)) * float64(damage))
	live := target.UpdateDataPlayer()
	accumulated := scaled + math.Float32frombits(live.Field21)
	effective := playerDamageRound4E17B0(accumulated)
	live.Field21 = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, r)
	if target.Class().Has(object.ClassPlayer) {
		if update.Field76 == 0 {
			update.Field76, update.Field75 = 2, uint32(typ)
		}
	} else if target.Class().Has(object.ClassMonster) {
		return reject("unsupported live world IMPALE cached marker layout")
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing live world IMPALE Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing world IMPALE default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
