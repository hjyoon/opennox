package server

type playerAbilityCooldownIndexedSetHooks4FC4A0 struct {
	loadPlayerIndexArg func() int32
	loadAbilityArg     func() Ability
	loadCooldownArg    func() int32
	storeCooldown      func(int32, int32)
}

// playerAbilityCooldownIndexedSet4FC4A0 preserves GAME.EXE 004FC4A0. The
// original routine reads the full signed player-index and ability dwords,
// computes the flat word index as ability + 6*playerIndex with PE32 wrapping,
// and only then reads the full signed cooldown dword. It stores that complete
// 32-bit value once and returns the same EAX bit pattern.
//
// No index is narrowed or validated here. The original executable addresses
// the cooldown matrix directly, so malformed inputs may select storage outside
// its normal 32-player by six-ability domain.
func playerAbilityCooldownIndexedSet4FC4A0(
	hooks playerAbilityCooldownIndexedSetHooks4FC4A0,
) int32 {
	playerIndex := hooks.loadPlayerIndexArg()
	ability := hooks.loadAbilityArg()
	threePlayers := uint32(playerIndex) * 3
	flatIndex := int32(uint32(ability) + 2*threePlayers)
	cooldown := hooks.loadCooldownArg()
	hooks.storeCooldown(flatIndex, cooldown)
	return cooldown
}
