package server

import "github.com/opennox/libs/object"

// playerDamageSimpleCrushShape4E17B0 admits a distinct SIMPLE-class CRUSH
// object owned by a unit. In stock thing.bin all three Fists are SIMPLE, not
// MISSILE or WEAPON. This class makes 004E1400 false, so the default tail has
// neither the ordinary-melee friendly gate nor Shock retaliation. The earlier
// campaign owner gate and PlayerDamage's type-based block exclusions remain.
func playerDamageSimpleCrushShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageCrush && source != nil && source != weapon &&
		(source.Class().Has(object.ClassPlayer) ||
			(source.Class().Has(object.ClassMonster) && source.UpdateData != nil)) &&
		weapon != nil && weapon.Class() == object.ClassSimple
}
