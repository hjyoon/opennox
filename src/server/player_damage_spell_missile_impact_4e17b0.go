package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// PixieCollide (004EA080) supplies its live terminal parent and the Pixie
// itself as weapon, with IMPACT/type 11. An unowned Pixie is its own parent.
// This admission describes pure spell missiles, not a cached Pixie type ID;
// weapon/wand/unit shapes keep their separate melee, Shock and modifier paths.
func playerDamageSpellMissileImpactShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageImpact || source == nil || weapon == nil ||
		!weapon.Class().Has(object.ClassMissile) ||
		weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand) {
		return false
	}
	return source == weapon || (source.UpdateData != nil &&
		source.Class().HasAny(object.MaskUnits) &&
		!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile))
}

// 004E17B0..004E1DBC selects the terminal parent/weapon pair BEFORE the
// player prefix. Retain entry equipment, absorption and marker addresses;
// ObserveClear may replace the live player update. Pure Pixie IMPACT/type 11
// reaches the ordinary full-armor switch, not BITE's specialized HP tail.
func playerDamagePlayerSpellMissileImpact4E17B0(
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
			return reject("missing spell-missile IMPACT invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing spell-missile IMPACT invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source.FindOwnerChainPlayer() == target && target.Class().Has(object.ClassPlayer) {
		return true, false
	}
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
		return reject("unsupported spell-missile IMPACT entry player record")
	}
	update := target.UpdateDataPlayer()
	if update.Player == nil {
		return reject("missing spell-missile IMPACT player info")
	}
	weaponFlags, armorFlags := update.Player.WeaponEquip, update.Player.ArmorEquip
	absorption := math.Float32frombits(update.Field57)
	if update.Player.Field3680&1 != 0 {
		return true, false
	}
	update.Field76 = 0
	if update.Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing spell-missile IMPACT observe service")
		}
		r.ObserveClear(target)
	}
	r.playerPrefix = &playerDamagePrefix4E17B0{update: update, weaponFlags: weaponFlags, armorFlags: armorFlags}
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && weapon.Class().Has(object.ClassMissile) {
		if r.BlockDirection == nil {
			return reject("missing spell-missile IMPACT Reflect direction")
		}
		if r.BlockDirection(target, weapon.PosVec) {
			if r.ProjectileReflect == nil {
				return reject("missing spell-missile IMPACT Reflect effect")
			}
			r.ProjectileReflect(weapon, target)
			// 004E193C/1956/195C reload after reflect and ownership calls.
			if uint32(weapon.SubClass())&0x40 == 0 {
				if r.ClearOwner == nil || r.SetOwner == nil {
					return reject("missing spell-missile IMPACT Reflect ownership")
				}
				r.ClearOwner(weapon)
				r.SetOwner(target, weapon)
			}
			if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 != 0 {
				if r.ChangeOwner == nil {
					return reject("missing spell-missile IMPACT Reflect ChangeOwner")
				}
				r.ChangeOwner(weapon, target)
			}
			if r.Audio == nil {
				return reject("missing spell-missile IMPACT Reflect audio")
			}
			r.Audio(122, target)
			return true, false
		}
	}
	// 004E1A49 snapshots PrevPos before exclusion. The distinct marker uses
	// the cached update and the LIVE weapon type after that callback.
	pos := weapon.PrevPos
	if r.BlockSourceExcluded == nil {
		return reject("missing spell-missile IMPACT exclusion")
	}
	excluded := r.BlockSourceExcluded(weapon)
	if source != weapon {
		if !target.Class().Has(object.ClassPlayer) {
			return reject("unsupported live spell-missile IMPACT attribution record")
		}
		update.Field76, update.Field75 = 1, uint32(weapon.TypeInd)
	}
	front := false
	if !excluded {
		if r.BlockDirection == nil {
			return reject("missing spell-missile IMPACT block direction")
		}
		front = r.BlockDirection(target, pos)
	}
	if front {
		if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil ||
			!playerDamageSpellMissileImpactShape4E17B0(source, weapon, typ) {
			return reject("unsupported live spell-missile IMPACT block record")
		}
		if armorFlags&0x3000000 != 0 && update.State == PlayerState16 {
			return playerDamageSpellMissileImpactShield4E17B0(target, source, weapon, damage, typ, r)
		}
		cached := playerDamageGreatSwordContext4E17B0{
			weaponFlags: weaponFlags, armorFlags: armorFlags,
			state: &update.State, marker: &update.Field76, markerType: &update.Field75,
		}
		// The external facing/exclusion pair has already run exactly once.
		r.BlockSourceExcluded = func(*Object) bool { return excluded }
		r.BlockDirection = func(*Object, types.Pointf) bool { return front }
		if applicable, h, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, cached, damage, typ, r); applicable {
			return h, result
		}
		if weaponFlags&0x400 == 0 && update.State == PlayerState1 && armorFlags&0x3000000 != 0 {
			if r.BerserkShieldBlock == nil {
				return reject("missing spell-missile IMPACT berserker shield service")
			}
			if r.BerserkShieldBlock(target) {
				return playerDamageSpellMissileImpactShield4E17B0(target, source, weapon, damage, typ, r)
			}
		}
	}
	return playerDamageSpellMissileImpactTail4E17B0(target, source, weapon, update, absorption, damage, typ, r)
}

// 004E1B98..004E1BFD emits audio before live projectile tests, and selects
// the live shield inventory only AFTER the balance callback. Bit 2 Pixies
// retain ownership in this block, unlike the earlier Reflect Shield branch.
func playerDamageSpellMissileImpactShield4E17B0(
	target, source, weapon *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if r.Audio == nil {
		return reject("missing spell-missile IMPACT shield audio")
	}
	r.Audio(878, target)
	if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&0x70 == 0 {
		if r.ProjectileReflect == nil {
			return reject("missing spell-missile IMPACT shield reflection")
		}
		r.ProjectileReflect(weapon, target)
		if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing spell-missile IMPACT shield ownership")
			}
			r.ClearOwner(weapon)
			r.SetOwner(target, weapon)
		}
	}
	if r.BlockDamagePercent == nil {
		return reject("missing spell-missile IMPACT shield balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	shield := playerDamageShieldItem4E17B0(target)
	if r.DamageBlockItem == nil || (shield != nil && (r.CanDamageBlockItem == nil || !r.CanDamageBlockItem(shield))) {
		return reject("unsupported live spell-missile IMPACT shield durability")
	}
	if !r.DamageBlockItem(shield, target, source, weapon, amount, typ) && r.Unsupported != nil {
		r.Unsupported("spell-missile IMPACT shield durability failed", target, source, weapon, damage, typ)
	}
	if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
		if r.PlayerSetState == nil || !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return reject("unsupported live spell-missile IMPACT broken-shield state")
		}
		r.PlayerSetState(target, PlayerState13)
	}
	return true, false
}

// Case 11 enters 004E1F84: binary32 absorption, LIVE carry, armor wear,
// cached marker, positive-only minimum, LIVE GodMode/Quest, DefaultDamage.
// Do not reinterpret a player's cached marker base as an NPC after a callback.
func playerDamageSpellMissileImpactTail4E17B0(
	target, source, weapon *Object, update *PlayerUpdateData, absorption float32,
	damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil ||
		!playerDamageSpellMissileImpactShape4E17B0(source, weapon, typ) {
		return reject("unsupported live spell-missile IMPACT armor/carry record")
	}
	if !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported spell-missile IMPACT armor durability")
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
		return reject("unsupported live spell-missile IMPACT cached marker layout")
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing live spell-missile IMPACT Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing spell-missile IMPACT default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
