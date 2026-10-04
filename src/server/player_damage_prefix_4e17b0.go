package server

import "github.com/opennox/libs/object"

// playerDamagePrefix4E17B0 is the entry state retained by 004E184A..004E18F0.
// ObserveClear runs after the marker reset and equipment/absorption reads,
// before any defense. Later marker/state accesses use this update base even
// when a callback replaces the unit's live update or controlling player.
// A non-nil context means the prefix has already run, not that the live
// player still has (or no longer has) an observation target.
type playerDamagePrefix4E17B0 struct {
	update                  *PlayerUpdateData
	armorFlags, weaponFlags uint32
}

// Keep admission identical to the already-restored PIERCE/DefaultDamage tail.
// Stock GolemArrow is MISSILE|WEAPON subclass 0x10, not a melee weapon.
func playerDamageMissilePierceShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageImpale && source != nil && source != weapon &&
		source.Class().HasAny(object.ClassPlayer|object.ClassMonster) && source.UpdateData != nil &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
		!defaultDamageAttackQualifies4E1400(source, weapon)
}
