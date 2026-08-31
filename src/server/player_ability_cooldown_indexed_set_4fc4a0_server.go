package server

// PlayerAbilityCooldownIndexedSet4FC4A0 is the native-width replacement for
// GAME.EXE 004FC4A0. It preserves the original flat PE32 index calculation and
// stores into the fixed 32-by-6 signed int32 cooldown matrix shared by the
// runtime updater, execution paths, and unit-based cooldown accessors.
//
// The executable had no callers or stored entrypoint for this routine, so this
// native method intentionally has no public C ABI wrapper. Out-of-matrix flat
// indices trip Go's array bounds check instead of corrupting adjacent memory.
//
//go:noinline
func (a *serverAbilities) PlayerAbilityCooldownIndexedSet4FC4A0(
	playerIndex int32,
	ability Ability,
	cooldown int32,
) int32 {
	return playerAbilityCooldownIndexedSet4FC4A0(playerAbilityCooldownIndexedSetHooks4FC4A0{
		loadPlayerIndexArg: func() int32 {
			return playerIndex
		},
		loadAbilityArg: func() Ability {
			return ability
		},
		loadCooldownArg: func() int32 {
			return cooldown
		},
		storeCooldown: func(flatIndex int32, cooldown int32) {
			row := flatIndex / int32(AbilityMax)
			column := Ability(flatIndex % int32(AbilityMax))
			a.cooldowns[row][column] = cooldown
		},
	})
}
