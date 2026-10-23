package server

import "github.com/opennox/libs/object"

// A direct callback may retain a non-item attack with no source.
// Its prefix skips ordinary source-facing but keeps the actual weapon for
// Reflect, wear and damage attribution. With a nil source the original never dereferences the attack unit update.
func playerDamageSourceLessNonItemShape4E17B0(source, weapon *Object) bool {
	return source == nil && weapon != nil && !weapon.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)
}
