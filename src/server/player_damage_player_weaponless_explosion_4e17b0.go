package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// Meteor's radial damage supplies its terminal unit owner (or nil), with
// no weapon. Keep missile splash, distinct weapons and ordinary NPC targets
// in their existing prefixes. Do not require update data before the original
// NoUpdate/invulnerability/Coop gates, which never dereference that record.
func playerDamagePlayerWeaponlessExplosionShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageExplosion && weapon == nil &&
		(source == nil || (source.Class().HasAny(object.MaskUnits) &&
			!source.Class().HasAny(object.ClassMissile|object.ClassWeapon|object.ClassWand)))
}

// GAME.EXE 004E17B0: cache entry player update/equipment/absorption, clear
// its marker before ObserveClear, then use four source-only exclusions and
// a pre-exclusion PrevPos snapshot. Case 7 has no source-type hit marker.
func playerDamagePlayerWeaponlessExplosion4E17B0(
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
			return reject("missing weapon-less player EXPLOSION invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing weapon-less player EXPLOSION invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source != nil &&
		source.FindOwnerChainPlayer() == target && target.Class().Has(object.ClassPlayer) {
		return true, false
	}
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
		return reject("unsupported weapon-less EXPLOSION entry player record")
	}
	update := target.UpdateDataPlayer()
	if update.Player == nil {
		return reject("missing weapon-less EXPLOSION player info")
	}
	weaponFlags, armorFlags := update.Player.WeaponEquip, update.Player.ArmorEquip
	absorption := math.Float32frombits(update.Field57)
	if update.Player.Field3680&1 != 0 {
		return true, false
	}
	update.Field76 = 0
	if update.Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing weapon-less EXPLOSION observe service")
		}
		r.ObserveClear(target)
	}
	// Pure unit Explosion does not enter Reflect27. A callback replacing
	// its source with a missile needs that separate live prefix, not PE32.
	if source != nil && target.HasEnchant(playerDamageReflectEnchant4E17B0) && source.Class().Has(object.ClassMissile) {
		return reject("unsupported live weapon-less EXPLOSION Reflect missile")
	}
	if source != nil {
		pos := source.PrevPos
		if r.BlockSourceOnlyExcluded == nil {
			return reject("missing weapon-less EXPLOSION source-only exclusion")
		}
		excluded := r.BlockSourceOnlyExcluded(source)
		front := false
		if !excluded {
			if r.BlockDirection == nil {
				return reject("missing weapon-less EXPLOSION block direction")
			}
			front = r.BlockDirection(target, pos)
		}
		if front {
			if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil ||
				!playerDamagePlayerWeaponlessExplosionShape4E17B0(source, weapon, typ) {
				return reject("unsupported live weapon-less EXPLOSION block record")
			}
			if armorFlags&0x3000000 != 0 && update.State == PlayerState16 {
				return playerDamagePlayerWeaponlessExplosionShield4E17B0(target, source, damage, typ, r)
			}
			// Great-sword/staff alternatives cannot block a unit case-7
			// attack. The original berserker shield exception still can.
			if weaponFlags&0x400 == 0 && update.State == PlayerState1 && armorFlags&0x3000000 != 0 {
				if r.BerserkShieldBlock == nil {
					return reject("missing weapon-less EXPLOSION berserker shield service")
				}
				if r.BerserkShieldBlock(target) {
					return playerDamagePlayerWeaponlessExplosionShield4E17B0(target, source, damage, typ, r)
				}
			}
		}
	}
	return playerDamagePlayerWeaponlessExplosionTail4E17B0(target, source, update, absorption, damage, typ, r)
}

// 004E1B98: audio precedes live source class/subclass and reflection; reload
// again before transferring ownership. Retain a nil weapon in shield wear.
func playerDamagePlayerWeaponlessExplosionShield4E17B0(
	target, source *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, nil, damage, typ)
	}
	if r.Audio == nil {
		return reject("missing weapon-less EXPLOSION shield audio")
	}
	r.Audio(878, target)
	if source.Class().Has(object.ClassMissile) && uint32(source.SubClass())&0x70 == 0 {
		if r.ProjectileReflect == nil {
			return reject("missing weapon-less EXPLOSION shield reflection")
		}
		r.ProjectileReflect(source, target)
		if source.Class().Has(object.ClassMissile) && uint32(source.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing weapon-less EXPLOSION shield ownership")
			}
			r.ClearOwner(source)
			r.SetOwner(target, source)
		}
	}
	if r.BlockDamagePercent == nil {
		return reject("missing weapon-less EXPLOSION shield balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	shield := playerDamageShieldItem4E17B0(target)
	if r.DamageBlockItem == nil || (shield != nil && (r.CanDamageBlockItem == nil || !r.CanDamageBlockItem(shield))) {
		return reject("unsupported live weapon-less EXPLOSION shield durability")
	}
	if !r.DamageBlockItem(shield, target, source, nil, amount, typ) && r.Unsupported != nil {
		r.Unsupported("weapon-less EXPLOSION shield durability failed", target, source, nil, damage, typ)
	}
	if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
		if r.PlayerSetState == nil || !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return reject("unsupported live weapon-less EXPLOSION broken-shield state")
		}
		r.PlayerSetState(target, PlayerState13)
	}
	return true, false
}

// Case 7's 004E1F84 tail: cached binary32 absorption, LIVE carry and armor
// wear, cached marker, positive-only minimum, then LIVE GodMode/Quest/HP.
func playerDamagePlayerWeaponlessExplosionTail4E17B0(
	target, source *Object, update *PlayerUpdateData, absorption float32,
	damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, nil, damage, typ)
	}
	worldExplosion := playerDamageWorldExplosionShape4E17B0(source, nil, typ)
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil ||
		(!worldExplosion && !playerDamagePlayerWeaponlessExplosionShape4E17B0(source, nil, typ)) ||
		(!worldExplosion && source != nil && source.UpdateData == nil) {
		return reject("unsupported live weapon-less EXPLOSION armor/carry record")
	}
	if !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported weapon-less EXPLOSION armor durability")
	}
	scaled := float32((1.0 - float64(absorption)) * float64(damage))
	live := target.UpdateDataPlayer()
	accumulated := scaled + math.Float32frombits(live.Field21)
	effective := playerDamageRound4E17B0(accumulated)
	live.Field21 = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, nil, damage-effective, typ, r)
	if target.Class().Has(object.ClassPlayer) {
		if update.Field76 == 0 {
			update.Field76, update.Field75 = 2, uint32(typ)
		}
	} else if target.Class().Has(object.ClassMonster) {
		return reject("unsupported live weapon-less EXPLOSION cached marker layout")
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing live weapon-less EXPLOSION Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing weapon-less EXPLOSION default damage")
	}
	return true, r.DefaultDamage(target, source, nil, effective, typ)
}
