package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// Keep NPC BITE admission separate from the already ported player strike
// shapes. Spider/Wasp, Scorpion/Zombie, Golem and Ghost use the same native
// unit as attacker and weapon, not a distinct inventory weapon or missile.
func playerDamageNPCMonsterSelfStrikeShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if playerDamageMonsterSelfStrikeShape4E17B0(source, weapon, typ) {
		return true
	}
	return typ == object.DamageBite && source != nil && source == weapon && source.UpdateData != nil &&
		source.Class().Has(object.ClassMonster) &&
		!source.Class().HasAny(object.ClassPlayer|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}

func playerDamageNPCMonsterSelfStrikeTail4E17B0(
	target, source, weapon *Object, cached playerDamageGreatSwordContext4E17B0,
	armorValue float32, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	// 004E1C23 and 004E1D16 admit nonmissile sword/staff blocks only for
	// BLADE or IMPACT. Other restored strikes still permit a stance shield.
	defense := cached
	if typ != object.DamageBlade {
		defense.weaponFlags = 0
	}
	if applicable, h, result := playerDamageMonsterImpactBlock4E17B0(target, source, weapon, defense, damage, typ, r); applicable {
		return h, result
	}
	if r.DefaultDamage == nil || typ != object.DamageDrain && !playerDamageArmorReady4E17B0(target, r) {
		return playerDamageUnsupported4E17B0(r, "missing NPC self-strike armor/default service", target, source, weapon, damage, typ)
	}
	if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
		return playerDamageUnsupported4E17B0(r, "unsupported NPC self-strike live carry record", target, source, weapon, damage, typ)
	}
	effective := damage
	// Cases 0/3/8/10 use full cached absorption at 004E1F84; CRUSH
	// uses half at 004E1EE8. Carry and armor wear reload the live update.
	// Ghost DRAIN enters 004E1E83 raw, leaving carry/durability untouched.
	if typ != object.DamageDrain {
		carry := &target.UpdateDataMonster().Field1
		absorption := float64(armorValue)
		if typ == object.DamageCrush {
			absorption *= 0.5
		}
		scaled := float32((1 - absorption) * float64(damage))
		accumulated := scaled + math.Float32frombits(*carry)
		effective = playerDamageRound4E17B0(accumulated)
		*carry = math.Float32bits(accumulated - float32(effective))
		playerDamageApplyArmor4E17B0(target, source, weapon, damage-effective, typ, r)
	}
	if target.Class().Has(object.ClassMonster) && !target.Class().Has(object.ClassPlayer) && *cached.marker == 0 {
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
			return playerDamageUnsupported4E17B0(r, "missing NPC self-strike live quest service", target, source, weapon, damage, typ)
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
