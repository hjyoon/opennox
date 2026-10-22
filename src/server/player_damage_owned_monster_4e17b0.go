package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
)

// Script Damage resolves a distinct terminal player owner, while retaining
// the initialized monster as the actual attack. Ordinary native strikes
// keep their existing unit-as-source-and-weapon path.
func playerDamageOwnedMonsterShape4E17B0(source, weapon *Object, _ object.DamageType) bool {
	return source != nil && weapon != nil && source != weapon && source.UpdateData != nil && weapon.UpdateData != nil &&
		source.Class().Has(object.ClassPlayer) && !source.Class().HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand|object.ClassMissile) && source.UpdateDataPlayer().Player != nil &&
		weapon.Class().Has(object.ClassMonster) && !weapon.Class().HasAny(object.ClassPlayer|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}

// Cache target layout/equipment before observer callbacks, but reload live
// attack positions, classes, ownership, inventory and carry at original use.
func playerDamageOwnedMonster4E17B0(
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
			return reject("missing owned monster invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing owned monster invulnerability audio")
			}
			r.Audio(71, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source.FindOwnerChainPlayer() == target && target.Class().Has(object.ClassPlayer) && typ != object.DamageManaBomb {
		return true, false
	}
	// The selected update and player-info dereferences are deliberately at
	// their original use, after the early flags/buff/Coop gates. In particular,
	// an absent live record is not turned into a silently ignored hit.
	if target.UpdateData == nil {
		return reject("missing owned monster target update")
	}
	cached := target.UpdateData
	var weaponFlags, armorFlags uint32
	var absorption float32
	if target.Class().Has(object.ClassPlayer) {
		update := target.UpdateDataPlayer()
		if update.Player == nil {
			return reject("missing owned monster target player info")
		}
		weaponFlags, armorFlags = update.Player.WeaponEquip, update.Player.ArmorEquip
		absorption = math.Float32frombits(update.Field57)
		if update.Player.Field3680&1 != 0 {
			return true, false
		}
	} else {
		if !target.Class().Has(object.ClassMonster) || uint32(target.SubClass())&0x10 == 0 {
			return reject("unsupported owned monster target")
		}
		update := target.UpdateDataMonster()
		weaponFlags, armorFlags = update.WeaponEquipFlags, update.ArmorEquipFlags
		absorption = math.Float32frombits(update.Field518)
	}
	marker, _ := playerDamageCasterImpactMarker4E17B0(target, cached)
	*marker = 0
	if target.Class().Has(object.ClassPlayer) && target.UpdateDataPlayer().Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing owned monster observe clear")
		}
		r.ObserveClear(target)
	}
	// Reflect selects the live attack after observer callbacks. Nonmissile
	// types16/17 have their original point-FX/audio-only reflection branch.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && (weapon.Class().Has(object.ClassMissile) || typ == object.DamageZapRay || typ == object.DamageAirborneElectric) {
		missile := weapon.Class().Has(object.ClassMissile)
		if r.BlockDirection == nil {
			return reject("missing owned monster Reflect direction")
		}
		if r.BlockDirection(target, weapon.PosVec) {
			if missile {
				if r.ProjectileReflect == nil {
					return reject("missing owned monster Reflect service")
				}
				r.ProjectileReflect(weapon, target)
				if uint32(weapon.SubClass())&0x40 == 0 {
					if r.ClearOwner == nil || r.SetOwner == nil {
						return reject("missing owned monster Reflect owner")
					}
					r.ClearOwner(weapon)
					r.SetOwner(target, weapon)
				}
				if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 != 0 {
					if r.ChangeOwner == nil {
						return reject("missing owned monster Reflect change owner")
					}
					r.ChangeOwner(weapon, target)
				}
			}
			if typ == object.DamageZapRay {
				if r.PointFX == nil {
					return reject("missing owned monster Reflect FX")
				}
				r.PointFX(132, target.PosVec)
			}

			if r.Audio == nil {
				return reject("missing owned monster Reflect audio")
			}
			r.Audio(122, target)
			return true, false
		}
	}
	if source != nil {
		pos := weapon.PrevPos
		if r.BlockSourceExcluded == nil {
			return reject("missing owned monster exclusion")
		}
		excluded := r.BlockSourceExcluded(weapon)
		if source != weapon {
			if marker, kind := playerDamageCasterImpactMarker4E17B0(target, cached); marker != nil {
				*marker, *kind = 1, uint32(weapon.TypeInd)
			}
		}
		if !excluded && typ != object.DamageManaBomb {
			if r.BlockDirection == nil {
				return reject("missing owned monster block direction")
			}
			if r.BlockDirection(target, pos) {
				if blocked, h, result := playerDamageCasterImpactBlock4E17B0(target, source, weapon, cached, weaponFlags, armorFlags, damage, typ, r); blocked {
					return h, result
				}
			}
		}
	}
	return playerDamageOwnedMonsterSwitch4E17B0(target, source, weapon, cached, absorption, damage, typ, r)
}

// 004E1E83..004E2040 retains the actual owner and attack through every type.
// The marker base is cached; its layout follows the live class. Carry uses
// the live update, including changes made by electric scaling or armor wear.
func playerDamageOwnedMonsterSwitch4E17B0(target, source, weapon *Object, cached unsafe.Pointer, absorption float32, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0) (bool, bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	physical := typ == object.DamageBlade || typ == object.DamageCrush || typ == object.DamageImpale || typ == object.DamageExplosion || typ == object.DamageBite || typ == object.DamageClaw || typ == object.DamageImpact
	electric := typ == object.DamageElectric || typ == object.DamageAirborneElectric
	wearArmor := physical || electric || typ == object.DamageFlame || typ == object.DamageLava
	if wearArmor && !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported owned monster armor durability")
	}
	effective, wear := damage, damage
	if physical || electric {
		var scaled float32
		if electric {
			if r.ElectricArmorScale == nil {
				return reject("missing owned monster electric armor scale")
			}
			scaled = float32(float64(r.ElectricArmorScale(target)) * float64(damage))
		} else {
			factor := float64(absorption)
			if typ == object.DamageCrush {
				factor *= .5
			}
			scaled = float32((1 - factor) * float64(damage))
		}
		var carry *uint32
		if target.Class().Has(object.ClassPlayer) {
			carry = &target.UpdateDataPlayer().Field21
		} else if target.Class().Has(object.ClassMonster) {
			carry = &target.UpdateDataMonster().Field1
		}
		accumulated := scaled
		if carry != nil {
			accumulated += math.Float32frombits(*carry)
		}
		effective = playerDamageRound4E17B0(accumulated)
		if carry != nil {
			*carry = math.Float32bits(accumulated - float32(effective))
		}
		if !electric {
			wear = damage - effective
		}
	}
	if wearArmor {
		playerDamageApplyArmor4E17B0(target, source, weapon, wear, typ, r)
	}
	if typ >= object.DamageBlade && typ <= object.DamageAirborneElectric && typ != object.DamageDeathMagic {
		if marker, kind := playerDamageCasterImpactMarker4E17B0(target, cached); marker != nil && *marker == 0 {
			*marker, *kind = 2, uint32(typ)
		}
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing owned monster Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing owned monster default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
