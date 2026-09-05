package server

const (
	castSpellByUserOffensiveFlag4FDD20 = uint32(0x20)
	castSpellByUserTargetedFlag4FDD20  = uint32(0x04)

	castSpellByUserInvisibleEnchant4FDD20    = int32(0)
	castSpellByUserInvulnerableEnchant4FDD20 = int32(23)
	castSpellByUserOvalShieldSpell4FDD20     = int32(67)
)

type castSpellByUserHooks4FDD20[Object comparable, AcceptArg any] struct {
	loadCasterArg func() Object
	loadSpellArg  func() int32

	spellGetPower  func(int32, Object) int32
	spellHasFlags  func(int32, uint32) int32
	disableEnchant func(Object, int32)
	cancelDuration func(int32, Object)

	loadAcceptArg    func() AcceptArg
	loadTarget       func(AcceptArg) Object
	createProjectile func(Object, Object, int32)
	spellAccept      func(int32, Object, Object, Object, AcceptArg, int32) int32
}

// castSpellByUser4FDD20 preserves GAME.EXE 004FDD20's entry-load, callback,
// target-cache, and return-value contract. Caster and spell ID are cached in
// that order, then spell power is always resolved and cached before either
// flag query. Offensive spells disable enchants 0 and 23 and cancel duration
// spell 67 in that exact order. Both flag predicates accept every nonzero
// signed dword.
//
// The accept-argument pointer is loaded only after the targeted flag query.
// Its target is read exactly once and only for a targeted spell. A distinct
// target creates a projectile and returns canonical one; otherwise the signed
// dword returned by spell acceptance is forwarded unchanged. The original has
// no nil, spell-definition, level, or result guards.
func castSpellByUser4FDD20[Object comparable, AcceptArg any](
	hooks castSpellByUserHooks4FDD20[Object, AcceptArg],
) int32 {
	caster := hooks.loadCasterArg()
	spellID := hooks.loadSpellArg()
	power := hooks.spellGetPower(spellID, caster)

	if hooks.spellHasFlags(spellID, castSpellByUserOffensiveFlag4FDD20) != 0 {
		hooks.disableEnchant(caster, castSpellByUserInvisibleEnchant4FDD20)
		hooks.disableEnchant(caster, castSpellByUserInvulnerableEnchant4FDD20)
		hooks.cancelDuration(castSpellByUserOvalShieldSpell4FDD20, caster)
	}

	targeted := hooks.spellHasFlags(spellID, castSpellByUserTargetedFlag4FDD20)
	arg := hooks.loadAcceptArg()
	if targeted != 0 {
		target := hooks.loadTarget(arg)
		if target != caster {
			hooks.createProjectile(caster, target, spellID)
			return 1
		}
	}
	return hooks.spellAccept(spellID, caster, caster, caster, arg, power)
}
