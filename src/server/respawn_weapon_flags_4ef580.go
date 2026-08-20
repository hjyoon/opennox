package server

type respawnWeaponFlagsHooks4EF580 struct {
	lookupArmor  func(uint32) int
	lookupWeapon func(uint32) int
	allowed      func(int) bool
}

// respawnWeaponFlags4EF580 preserves GAME.EXE 004EF580. Each lookup result is
// passed immediately to the allowed-type predicate. All eight pairs execute in
// the original order even when an earlier predicate succeeds or fails, so both
// callbacks observe live state at every step.
//
// The original returns BL through AL and its sole caller consumes only that
// byte. A uint8 result therefore preserves the complete public contract without
// carrying the final predicate's otherwise unobserved high EAX bits.
func respawnWeaponFlags4EF580(hooks respawnWeaponFlagsHooks4EF580) uint8 {
	var flags uint8
	if hooks.allowed(hooks.lookupArmor(0x400)) {
		flags = 0x01
	}
	if hooks.allowed(hooks.lookupArmor(0x4)) {
		flags |= 0x02
	}
	if hooks.allowed(hooks.lookupArmor(0x1)) {
		flags |= 0x04
	}
	if hooks.allowed(hooks.lookupWeapon(0x8000)) {
		flags |= 0x08
	}
	if hooks.allowed(hooks.lookupArmor(0x4000)) {
		flags |= 0x10
	}
	if hooks.allowed(hooks.lookupWeapon(0x100)) {
		flags |= 0x20
	}
	if hooks.allowed(hooks.lookupWeapon(0x200)) {
		flags |= 0x40
	}
	if hooks.allowed(hooks.lookupArmor(0x1000000)) {
		flags |= 0x80
	}
	return flags
}
