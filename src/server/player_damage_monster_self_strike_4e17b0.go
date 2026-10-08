package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// Scorpion/Zombie/Golem and other native strike callbacks pass the monster
// itself as BOTH attacker and weapon, not an inventory weapon or missile.
func playerDamageMonsterSelfStrikeShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	switch typ {
	case object.DamageBlade, object.DamageCrush, object.DamageImpale, object.DamageDrain, object.DamageClaw:
	default:
		return false
	}
	return source != nil && source == weapon && source.UpdateData != nil &&
		source.Class().Has(object.ClassMonster) &&
		!source.Class().HasAny(object.ClassPlayer|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}

// Only BLADE (case 0) reaches these sword/staff blocks. CRUSH, IMPALE and
// CLAW and DRAIN skip them, though all five accept an ordinary shield block.
func playerDamageMonsterSelfBladeBlock4E17B0(target, source, weapon *Object, cached playerDamageGreatSwordContext4E17B0, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0) (applicable, handled, result bool) {
	mask, sound := uint32(0), 0
	state := *cached.state
	if cached.weaponFlags&0x400 != 0 {
		if state != PlayerState13 && state != PlayerState18 && state != PlayerState19 && state != PlayerState20 {
			return false, false, false
		}
		mask, sound = 0x400, 890
	} else if cached.weaponFlags&0x7ff8000 != 0 {
		// 004E1D35 accepts exactly standing 13 or staff-block 21.
		if state != PlayerState13 && state != PlayerState21 {
			return false, false, false
		}
		mask, sound = 0x7ff8000, 894
	} else {
		return false, false, false
	}
	reject := func(reason string) (bool, bool, bool) {
		h, result := playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
		return true, h, result
	}
	if r.Audio == nil || r.PlayerSetState == nil || r.BlockDamagePercent == nil ||
		r.Melee.DamageBlockWeapon == nil || r.Melee.CanDamageBlockWeapon == nil ||
		(sound == 890 && r.Melee.RandomInt == nil) {
		return reject("missing self-weapon BLADE block service")
	}
	r.Audio(sound, target)
	state = PlayerState21
	if sound == 890 {
		state = PlayerState(r.Melee.RandomInt(18, 20))
	}
	r.PlayerSetState(target, state)
	amount := float32(r.BlockDamagePercent() * float64(damage))
	// 004E22A0 reloads inventory only after audio/state/balance callbacks.
	var item *Object
	for next := target.InvFirstItem; next != nil; next = next.InvNextItem {
		if next.Flags().Has(object.FlagEquipped) && uint32(next.SubClass())&mask != 0 {
			item = next
			break
		}
	}
	if item == nil || !r.Melee.CanDamageBlockWeapon(item) {
		return reject("unsupported self-weapon BLADE live block weapon")
	}
	if !r.Melee.DamageBlockWeapon(item, target, source, weapon, amount, typ) && r.Unsupported != nil {
		r.Unsupported("self-weapon BLADE block durability failed", target, source, weapon, damage, typ)
	}
	if item.Flags().Has(object.FlagDestroyed) {
		r.PlayerSetState(target, PlayerState13)
	}
	return true, true, false
}

func playerDamageMonsterSelfStrikeTail4E17B0(target, source, weapon *Object, cached playerDamageGreatSwordContext4E17B0, armorValue float32, damage int32, typ object.DamageType, front bool, r PlayerDamageRuntime4E17B0) (handled, result bool) {
	if front {
		if applicable, h, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, r); applicable {
			return h, result
		}
		if typ == object.DamageBlade {
			if applicable, h, result := playerDamageMonsterSelfBladeBlock4E17B0(target, source, weapon, cached, damage, typ, r); applicable {
				return h, result
			}
		}
	}
	if r.DefaultDamage == nil || typ != object.DamageDrain && !playerDamageArmorReady4E17B0(target, r) {
		return playerDamageUnsupported4E17B0(r, "missing monster self-strike armor/default service", target, source, weapon, damage, typ)
	}
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
		return playerDamageUnsupported4E17B0(r, "unsupported monster self-strike live carry", target, source, weapon, damage, typ)
	}
	effective := damage
	// Ghost DRAIN (case 4 -> 004E1E83) is raw: no armor, carry or wear.
	if typ != object.DamageDrain {
		carry := &target.UpdateDataPlayer().Field21
		absorption := float64(armorValue)
		if typ == object.DamageCrush {
			absorption *= 0.5
		} // 004E1EE8, not full armor.
		scaled := float32((1 - absorption) * float64(damage))
		accumulated := scaled + math.Float32frombits(*carry)
		effective = playerDamageRound4E17B0(accumulated)
		*carry = math.Float32bits(accumulated - float32(effective))
		playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, r)
	}
	if target.Class().Has(object.ClassPlayer) && *cached.marker == 0 {
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
			return playerDamageUnsupported4E17B0(r, "missing monster self-strike live quest service", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
