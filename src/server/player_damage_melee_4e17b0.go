package server

import "github.com/opennox/libs/object"

// playerDamageMeleeShape4E17B0 identifies the ordinary armed/unarmed unit
// attacks admitted by the melee slice of GAME.EXE 004E17B0/004E0B30. Charge,
// missiles and spell projectiles retain their separately ported paths.
func playerDamageMeleeShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if source == nil || (!source.Class().Has(object.ClassPlayer) &&
		(!source.Class().Has(object.ClassMonster) || source.UpdateData == nil)) {
		return false
	}
	if weapon == nil {
		return (typ == object.DamageClaw || typ == object.DamageCrush) &&
			(source.Class().Has(object.ClassPlayer) || uint32(source.SubClass())&0x10 != 0)
	}
	if source == weapon || weapon.Class().HasAny(object.MaskUnits|object.ClassMissile) ||
		!weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) ||
		(typ != object.DamageBlade && typ != object.DamageCrush) {
		return false
	}
	// A spell staff can qualify only through its native WandUseData flags.
	// Do not enter 004E1400's original nil-fault boundary for malformed data.
	if weapon.Class().Has(object.ClassWand) && uint32(weapon.SubClass())&0x047f0000 != 0 && weapon.UseData.Ptr == nil {
		return false
	}
	return defaultDamageAttackQualifies4E1400(source, weapon)
}

// defaultDamageFriendlyException4E1470 restores the War Hammer exception in
// GAME.EXE 004E1470. This skips only the melee gate, not the earlier campaign
// owner gate in DefaultDamage.
func defaultDamageFriendlyException4E1470(weapon *Object) bool {
	return weapon != nil && weapon.Class().Has(object.ClassWeapon) && uint32(weapon.SubClass())&0x4000 != 0
}
