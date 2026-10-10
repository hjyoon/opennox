package server

import (
	"math"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// This predicate shares only the armed qualifier with BLADE. The executed
// prefix, armor wear and DefaultDamage always retain incoming EXPLOSION (7).
// Units, missiles, unowned attacks and non-qualifying ranged wands keep their
// separately admitted paths; no object name or cached type ID selects this one.
func playerDamageArmedExplosionShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageExplosion && source != nil && weapon != nil &&
		!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile) &&
		playerDamageMeleeShape4E17B0(source, weapon, object.DamageBlade)
}

// GAME.EXE 004E17B0: cache the entry equipment/absorption/marker base, then
// reset its marker before ObserveClear. Only carry and durability use live
// update/inventory records. Missing services are reported at their executed
// use, without undoing earlier stores or retrying a PE32 damage callback.
func playerDamageArmedExplosion4E17B0(
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
			return reject("missing armed EXPLOSION invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing armed EXPLOSION invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source.FindOwnerChainPlayer() == target && target.Class().Has(object.ClassPlayer) {
		return true, false
	}
	player := target.Class().Has(object.ClassPlayer)
	if target.UpdateData == nil || (!player && (!target.Class().Has(object.ClassMonster) || uint32(target.SubClass())&0x10 == 0)) {
		return reject("unsupported armed EXPLOSION entry update")
	}
	var marker, markerType *uint32
	var cachedState *PlayerState
	var absorption float32
	var armorFlags, weaponFlags uint32
	if player {
		u := target.UpdateDataPlayer()
		if u.Player == nil {
			return reject("missing armed EXPLOSION player info")
		}
		armorFlags, weaponFlags = u.Player.ArmorEquip, u.Player.WeaponEquip
		absorption = math.Float32frombits(u.Field57)
		marker, markerType, cachedState = &u.Field76, &u.Field75, &u.State
		if u.Player.Field3680&1 != 0 {
			return true, false
		}
	} else {
		u := target.UpdateDataMonster()
		armorFlags, weaponFlags = u.ArmorEquipFlags, u.WeaponEquipFlags
		absorption = math.Float32frombits(u.Field518)
		marker, markerType = &u.Field547, &u.Field546
	}
	*marker = 0
	if player && target.UpdateDataPlayer().Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing armed EXPLOSION observe service")
		}
		r.ObserveClear(target)
	}
	// Pure armed case 7 skips Reflect27. If an observer callback introduces
	// MISSILE, preserve the branch's current-position and post-reflect loads.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && weapon.Class().Has(object.ClassMissile) {
		if r.BlockDirection == nil {
			return reject("missing armed EXPLOSION Reflect direction")
		}
		if r.BlockDirection(target, weapon.PosVec) {
			if r.ProjectileReflect == nil {
				return reject("missing armed EXPLOSION Reflect service")
			}
			r.ProjectileReflect(weapon, target)
			if uint32(weapon.SubClass())&0x40 == 0 {
				if r.ClearOwner == nil || r.SetOwner == nil {
					return reject("missing armed EXPLOSION Reflect ownership")
				}
				r.ClearOwner(weapon)
				r.SetOwner(target, weapon)
			}
			if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 != 0 {
				if r.ChangeOwner == nil {
					return reject("missing armed EXPLOSION Reflect owner change")
				}
				r.ChangeOwner(weapon, target)
			}
			if r.Audio == nil {
				return reject("missing armed EXPLOSION Reflect audio")
			}
			r.Audio(122, target)
			return true, false
		}
	}
	sameLayout := func() bool {
		return target.Class().Has(object.ClassPlayer) == player && (player || target.Class().Has(object.ClassMonster))
	}
	// 004E1A49 snapshots PrevPos before six-type exclusions. Attribution
	// instead reads the live weapon type afterwards, even for an excluded hit.
	pos := weapon.PrevPos
	if r.BlockSourceExcluded == nil {
		return reject("missing armed EXPLOSION weapon exclusion")
	}
	excluded := r.BlockSourceExcluded(weapon)
	if !sameLayout() {
		return reject("unsupported live armed EXPLOSION cached marker layout")
	}
	*marker, *markerType = 1, uint32(weapon.TypeInd)
	if !excluded {
		if r.BlockDirection == nil {
			return reject("missing armed EXPLOSION block direction")
		}
		if r.BlockDirection(target, pos) {
			if !sameLayout() {
				return reject("unsupported live armed EXPLOSION block layout")
			}
			shield := false
			if player {
				shield = *cachedState == PlayerState16
			} else {
				if target.UpdateData == nil {
					return reject("missing live armed EXPLOSION NPC action update")
				}
				shield = target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK
			}
			if shield && armorFlags&0x3000000 != 0 {
				return playerDamageArmedExplosionShield4E17B0(target, source, weapon, damage, typ, r)
			}
			// A pure case-7 attack cannot use the GreatSword/staff branch.
			// A callback-created missile needs that separate live weapon block.
			if weaponFlags&0x400 != 0 && weapon.Class().Has(object.ClassMissile) {
				return reject("unsupported live armed EXPLOSION GreatSword missile")
			}
			// Retain OpenNox's optional gameex berserker-shield extension.
			// Stock GAME.EXE has no such branch; its replay never enables it.
			if player && weaponFlags&0x400 == 0 && *cachedState == PlayerState1 && armorFlags&0x3000000 != 0 &&
				r.BerserkShieldBlock != nil && r.BerserkShieldBlock(target) {
				return playerDamageArmedExplosionShield4E17B0(target, source, weapon, damage, typ, r)
			}
		}
	}
	if !sameLayout() || target.UpdateData == nil {
		return reject("unsupported live armed EXPLOSION carry update")
	}
	var carry *uint32
	if player {
		carry = &target.UpdateDataPlayer().Field21
	} else {
		carry = &target.UpdateDataMonster().Field1
	}
	scaled := float32((1.0 - float64(absorption)) * float64(damage))
	accumulated := scaled + math.Float32frombits(*carry)
	effective := playerDamageRound4E17B0(accumulated)
	*carry = math.Float32bits(accumulated - float32(effective))
	if !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported live armed EXPLOSION armor durability")
	}
	playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, r)
	if !sameLayout() {
		return reject("unsupported live armed EXPLOSION armor marker layout")
	}
	if *marker == 0 {
		*marker, *markerType = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing live armed EXPLOSION Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing armed EXPLOSION default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}

// 004E1B98..004E1BFD: sound precedes live missile reflection and ownership;
// balance precedes the live first-equipped-shield lookup. A missing shield
// retains the original no-op wear call, then reports the PE32 nil fault safely.
func playerDamageArmedExplosionShield4E17B0(
	target, source, weapon *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if r.Audio == nil {
		return reject("missing armed EXPLOSION shield audio")
	}
	r.Audio(878, target)
	if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&0x70 == 0 {
		if r.ProjectileReflect == nil {
			return reject("missing armed EXPLOSION shield reflection")
		}
		r.ProjectileReflect(weapon, target)
		if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing armed EXPLOSION shield ownership")
			}
			r.ClearOwner(weapon)
			r.SetOwner(target, weapon)
		}
	}
	if r.BlockDamagePercent == nil {
		return reject("missing armed EXPLOSION shield balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	shield := playerDamageShieldItem4E17B0(target)
	if r.DamageBlockItem == nil || (shield != nil && shield.HealthData != nil &&
		(r.CanDamageBlockItem == nil || !r.CanDamageBlockItem(shield))) {
		return reject("unsupported live armed EXPLOSION shield durability")
	}
	if !r.DamageBlockItem(shield, target, source, weapon, amount, typ) && r.Unsupported != nil {
		r.Unsupported("armed EXPLOSION shield durability failed", target, source, weapon, damage, typ)
	}
	if shield == nil {
		return reject("original armed EXPLOSION nil shield fault")
	}
	if shield.Flags().Has(object.FlagDestroyed) {
		if target.Class().Has(object.ClassPlayer) {
			if r.PlayerSetState == nil || target.UpdateData == nil {
				return reject("unsupported live armed EXPLOSION broken-shield player")
			}
			r.PlayerSetState(target, PlayerState13)
		} else if target.Class().Has(object.ClassMonster) {
			if r.Melee.MonsterPopBlockAction == nil {
				return reject("missing armed EXPLOSION broken-shield NPC action")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, false
}
