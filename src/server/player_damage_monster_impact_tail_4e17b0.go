package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// This nonmissile case 11 uses the cached entry equipment/state/marker and
// the already evaluated exclusion/facing answer. A self-weapon never gets
// the distinct projectile type marker or projectile reflection/ownership.
func playerDamageMonsterImpactBlock4E17B0(
	target, source, weapon *Object, cached playerDamageGreatSwordContext4E17B0,
	damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	if !r.BlockDirection(target, weapon.PrevPos) {
		return false, false, false
	}
	reject := func(reason string) (bool, bool, bool) {
		h, result := playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
		return true, h, result
	}
	player := cached.state != nil
	state := PlayerState(0)
	if player {
		state = *cached.state
	}
	shield := cached.armorFlags&0x3000000 != 0 &&
		((player && state == PlayerState16) || (!player && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK))
	mask, sound := uint32(2), 878
	if !shield && cached.weaponFlags&0x400 != 0 {
		ready := player && (state == PlayerState13 || state == PlayerState18 || state == PlayerState19 || state == PlayerState20)
		if !player {
			ready = playerDamageMonsterBlockReady534340(target)
		}
		if !ready {
			return false, false, false
		}
		mask, sound = 0x400, 890
	} else if !shield {
		if player && state == PlayerState1 && cached.armorFlags&0x3000000 != 0 {
			if r.BerserkShieldBlock == nil {
				return reject("missing self-weapon IMPACT berserker shield service")
			}
			shield = r.BerserkShieldBlock(target)
		}
		if !shield && cached.weaponFlags&0x7ff8000 != 0 {
			ready := player && (state == PlayerState13 || state == PlayerState18 || state == PlayerState19 || state == PlayerState20)
			if player && state == PlayerState0 {
				if r.Melee.StaffWalkingBlock == nil {
					return reject("missing self-weapon IMPACT walking staff service")
				}
				ready = r.Melee.StaffWalkingBlock()
			}
			if !player {
				ready = playerDamageMonsterBlockReady534340(target)
			}
			if !ready {
				return false, false, false
			}
			mask, sound = 0x7ff8000, 894
		} else if !shield {
			return false, false, false
		}
	}
	if r.Audio == nil || r.BlockDamagePercent == nil ||
		(mask != 2 && ((player && (r.PlayerSetState == nil || (sound == 890 && r.Melee.RandomInt == nil))) || (!player && r.Melee.MonsterBlockAction == nil))) {
		return reject("missing self-weapon IMPACT block effect service")
	}
	r.Audio(sound, target)
	if mask != 2 {
		if player {
			state = PlayerState21
			if sound == 890 {
				state = PlayerState(r.Melee.RandomInt(18, 20))
			}
			r.PlayerSetState(target, state)
		} else {
			r.Melee.MonsterBlockAction(target)
		}
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	// 004E22A0/004E2330 reload inventory after audio/state/balance callbacks.
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
		if item == nil {
			return reject("missing self-weapon IMPACT live block weapon")
		}
	}
	if apply == nil || (item != nil && (admit == nil || !admit(item))) {
		return reject("unsupported self-weapon IMPACT live block durability")
	}
	if !apply(item, target, source, weapon, amount, typ) && r.Unsupported != nil {
		r.Unsupported("self-weapon IMPACT block durability failed", target, source, weapon, damage, typ)
	}
	if item != nil && item.Flags().Has(object.FlagDestroyed) {
		if player {
			if r.PlayerSetState == nil {
				return reject("missing self-weapon IMPACT broken block state")
			}
			r.PlayerSetState(target, PlayerState13)
		} else {
			if r.Melee.MonsterPopBlockAction == nil {
				return reject("missing self-weapon IMPACT broken block action")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, true, false
}

// Case 11 shares 004E1F84's full armor/carry/wear switch with case 3, but
// retains the cleared self-weapon marker. Original signed/zero values reach
// the switch; only positive raw/effective values get the later minimum.
func playerDamageMonsterImpactTail4E17B0(
	target, source, weapon *Object, cached playerDamageGreatSwordContext4E17B0,
	armorValue float32, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if applicable, h, result := playerDamageMonsterImpactBlock4E17B0(target, source, weapon, cached, damage, typ, r); applicable {
		return h, result
	}
	if r.DefaultDamage == nil || !playerDamageArmorReady4E17B0(target, r) {
		return playerDamageUnsupported4E17B0(r, "missing self-weapon IMPACT armor/default service", target, source, weapon, damage, typ)
	}
	var carry *uint32
	if cached.state != nil && target.Class().Has(object.ClassPlayer) && target.UpdateData != nil {
		carry = &target.UpdateDataPlayer().Field21
	} else if cached.state == nil && target.Class().Has(object.ClassMonster) && !target.Class().Has(object.ClassPlayer) && target.UpdateData != nil {
		carry = &target.UpdateDataMonster().Field1
	} else {
		return playerDamageUnsupported4E17B0(r, "unsupported self-weapon IMPACT live carry record", target, source, weapon, damage, typ)
	}
	scaled := float32((1 - float64(armorValue)) * float64(damage))
	accumulated := scaled + math.Float32frombits(*carry)
	effective := playerDamageRound4E17B0(accumulated)
	*carry = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, r)
	if *cached.marker == 0 {
		*cached.marker, *cached.markerType = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(r, "missing self-weapon IMPACT live quest service", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
