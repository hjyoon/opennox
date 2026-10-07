package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// 004E1A97/004E1B0D and 004E1FB1 use the live class, but the update base
// cached at 004E184E/004E187E. Do not substitute a callback's new UpdateData
// for hit attribution. Carry and item wear have their own live update reads.
func playerDamageCasterImpactMarker4E17B0(target *Object, cached unsafe.Pointer) (marker, kind *uint32) {
	if target.Class().Has(object.ClassPlayer) {
		update := (*PlayerUpdateData)(cached)
		return &update.Field76, &update.Field75
	}
	if target.Class().Has(object.ClassMonster) {
		update := (*MonsterUpdateData)(cached)
		return &update.Field547, &update.Field546
	}
	return nil, nil
}

// This is the non-item case 11 used by Earthquake, not the existing Troll
// self-strike slice. GAME.EXE does not restrict this switch to MONSTER or
// MISSILE sources: source is the terminal owner and weapon is the full caster.
// The caller admits only the previously missing NPC shapes; an entry-time
// nonmissile can still become a missile during a defense callback.
func playerDamageCasterImpact4E17B0(
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
			return reject("missing caster IMPACT invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing caster IMPACT invulnerability audio")
			}
			r.Audio(71, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source.FindOwnerChainPlayer() == target && target.Class().Has(object.ClassPlayer) {
		return true, false
	}
	// The selected update and player-info dereferences are deliberately at
	// their original use, after the early flags/buff/Coop gates. In particular,
	// an absent live record is not turned into a silently ignored hit.
	cached := target.UpdateData
	var weaponFlags, armorFlags uint32
	var absorption float32
	if target.Class().Has(object.ClassPlayer) {
		update := target.UpdateDataPlayer()
		weaponFlags, armorFlags = update.Player.WeaponEquip, update.Player.ArmorEquip
		absorption = math.Float32frombits(update.Field57)
		if update.Player.Field3680&1 != 0 {
			return true, false
		}
	} else {
		if !target.Class().Has(object.ClassMonster) || uint32(target.SubClass())&0x10 == 0 {
			return reject("unsupported caster IMPACT target")
		}
		update := target.UpdateDataMonster()
		weaponFlags, armorFlags = update.WeaponEquipFlags, update.ArmorEquipFlags
		absorption = math.Float32frombits(update.Field518)
	}
	marker, _ := playerDamageCasterImpactMarker4E17B0(target, cached)
	*marker = 0
	if target.Class().Has(object.ClassPlayer) && target.UpdateDataPlayer().Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing caster IMPACT observe clear")
		}
		r.ObserveClear(target)
	}
	// Ordinarily a nonmissile IMPACT skips Reflect. Read the live attack here,
	// however: Coop/Observe callbacks precede 004E18F5's class test.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && weapon.Class().Has(object.ClassMissile) {
		if r.BlockDirection == nil {
			return reject("missing caster IMPACT Reflect direction")
		}
		if r.BlockDirection(target, weapon.PosVec) {
			if r.ProjectileReflect == nil {
				return reject("missing caster IMPACT Reflect service")
			}
			r.ProjectileReflect(weapon, target)
			if uint32(weapon.SubClass())&0x40 == 0 {
				if r.ClearOwner == nil || r.SetOwner == nil {
					return reject("missing caster IMPACT Reflect owner")
				}
				r.ClearOwner(weapon)
				r.SetOwner(target, weapon)
			}
			if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 != 0 {
				if r.ChangeOwner == nil {
					return reject("missing caster IMPACT Reflect change owner")
				}
				r.ChangeOwner(weapon, target)
			}
			if r.Audio == nil {
				return reject("missing caster IMPACT Reflect audio")
			}
			r.Audio(122, target)
			return true, false
		}
	}
	if source != nil {
		pos := weapon.PrevPos
		if r.BlockSourceExcluded == nil {
			return reject("missing caster IMPACT exclusion")
		}
		excluded := r.BlockSourceExcluded(weapon)
		if source != weapon {
			if marker, kind := playerDamageCasterImpactMarker4E17B0(target, cached); marker != nil {
				*marker, *kind = 1, uint32(weapon.TypeInd)
			}
		}
		if !excluded {
			if r.BlockDirection == nil {
				return reject("missing caster IMPACT block direction")
			}
			if r.BlockDirection(target, pos) {
				if blocked, h, result := playerDamageCasterImpactBlock4E17B0(target, source, weapon, cached, weaponFlags, armorFlags, damage, typ, r); blocked {
					return h, result
				}
			}
		}
	}
	if !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported caster IMPACT armor durability")
	}
	// 004E20F0 selects the live class/update for carry, even if callbacks have
	// replaced the cached NPC record or removed unit class bits altogether.
	accumulated := float32((1 - float64(absorption)) * float64(damage))
	if target.Class().Has(object.ClassPlayer) {
		accumulated += math.Float32frombits(target.UpdateDataPlayer().Field21)
	} else if target.Class().Has(object.ClassMonster) {
		accumulated += math.Float32frombits(target.UpdateDataMonster().Field1)
	}
	effective := playerDamageRound4E17B0(accumulated)
	if target.Class().Has(object.ClassPlayer) {
		target.UpdateDataPlayer().Field21 = math.Float32bits(accumulated - float32(effective))
	} else if target.Class().Has(object.ClassMonster) {
		target.UpdateDataMonster().Field1 = math.Float32bits(accumulated - float32(effective))
	}
	playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, r)
	if marker, kind := playerDamageCasterImpactMarker4E17B0(target, cached); marker != nil && *marker == 0 {
		*marker, *kind = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing live caster IMPACT Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing caster IMPACT default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}

// Preserve the cached equipment selection, but live action, attack class and
// inventory reads. Shield audio precedes reflection; GreatSword reflection
// precedes audio. Both reload subclass/ownership after reflection and select
// the durability item only after audio/action/balance callbacks.
func playerDamageCasterImpactBlock4E17B0(
	target, source, weapon *Object, cached unsafe.Pointer, weaponFlags, armorFlags uint32,
	damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (blocked, handled, result bool) {
	reject := func(reason string) (bool, bool, bool) {
		h, result := playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
		return true, h, result
	}
	player := target.Class().Has(object.ClassPlayer)
	shield := (player && (*PlayerUpdateData)(cached).State == PlayerState16) ||
		(!player && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK)
	shield = shield && armorFlags&0x3000000 != 0
	mask, sound := uint32(2), 878
	if !shield {
		if weaponFlags&0x400 != 0 {
			ready := false
			if player {
				state := (*PlayerUpdateData)(cached).State
				ready = state == PlayerState13 || state == PlayerState18 || state == PlayerState19 || state == PlayerState20
			} else {
				ready = playerDamageMonsterBlockReady534340(target)
			}
			if !ready {
				return false, false, false
			}
			mask, sound = 0x400, 890
		} else {
			if player && (*PlayerUpdateData)(cached).State == PlayerState1 && armorFlags&0x3000000 != 0 {
				if r.BerserkShieldBlock == nil {
					return reject("missing caster IMPACT berserker shield service")
				}
				shield = r.BerserkShieldBlock(target)
			}
			if !shield {
				if weaponFlags&0x7ff8000 == 0 || weapon.Class().Has(object.ClassMissile) {
					return false, false, false
				}
				ready := false
				if player {
					state := (*PlayerUpdateData)(cached).State
					// 004E1D35 accepts the staff's idle/repeated-block states,
					// not GreatSword's randomized states 18/19/20.
					ready = state == PlayerState13 || state == PlayerState21
					if state == PlayerState0 && r.Melee.StaffWalkingBlock != nil {
						ready = r.Melee.StaffWalkingBlock()
					}
				} else {
					ready = playerDamageMonsterBlockReady534340(target)
				}
				if !ready {
					return false, false, false
				}
				mask, sound = 0x7ff8000, 894
			}
		}
	}
	if mask == 0x400 && weapon.Class().Has(object.ClassMissile) {
		if r.ProjectileReflect == nil {
			return reject("missing caster IMPACT GreatSword reflection")
		}
		r.ProjectileReflect(weapon, target)
		if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing caster IMPACT GreatSword owner")
			}
			r.ClearOwner(weapon)
			r.SetOwner(target, weapon)
		}
	}
	if r.Audio == nil {
		return reject("missing caster IMPACT block audio")
	}
	r.Audio(sound, target)
	if mask == 2 && weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&0x70 == 0 {
		if r.ProjectileReflect == nil {
			return reject("missing caster IMPACT shield reflection")
		}
		r.ProjectileReflect(weapon, target)
		if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing caster IMPACT shield owner")
			}
			r.ClearOwner(weapon)
			r.SetOwner(target, weapon)
		}
	}
	if mask != 2 {
		if target.Class().Has(object.ClassPlayer) {
			state := PlayerState21
			if mask == 0x400 {
				if r.Melee.RandomInt == nil {
					return reject("missing caster IMPACT GreatSword random state")
				}
				state = PlayerState(r.Melee.RandomInt(18, 20))
			}
			if r.PlayerSetState == nil {
				return reject("missing caster IMPACT block player state")
			}
			r.PlayerSetState(target, state)
		} else {
			if r.Melee.MonsterBlockAction == nil {
				return reject("missing caster IMPACT block monster action")
			}
			r.Melee.MonsterBlockAction(target)
		}
	}
	if r.BlockDamagePercent == nil {
		return reject("missing caster IMPACT block balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	var item *Object
	for next := target.InvFirstItem; next != nil; next = next.InvNextItem {
		if next.Flags().Has(object.FlagEquipped) && uint32(next.SubClass())&mask != 0 {
			item = next
			break
		}
	}
	apply, admit := r.DamageBlockItem, r.CanDamageBlockItem
	if mask != 2 {
		apply, admit = r.Melee.DamageBlockWeapon, r.Melee.CanDamageBlockWeapon
	}
	if apply == nil || (item != nil && (admit == nil || !admit(item))) {
		return reject("unsupported caster IMPACT live block durability")
	}
	if !apply(item, target, source, weapon, amount, typ) && r.Unsupported != nil {
		r.Unsupported("caster IMPACT block durability failed", target, source, weapon, damage, typ)
	}
	// The original weapon helper dereferences a missing item after its
	// durability call; only the shield helper has the no-item no-op.
	if (mask != 2 || item != nil) && item.ObjFlags.Has(object.FlagDestroyed) {
		if target.Class().Has(object.ClassPlayer) {
			if r.PlayerSetState == nil {
				return reject("missing caster IMPACT broken block player state")
			}
			r.PlayerSetState(target, PlayerState13)
		} else {
			if r.Melee.MonsterPopBlockAction == nil {
				return reject("missing caster IMPACT broken block monster action")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, true, false
}
