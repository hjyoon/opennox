package server

type playerAbilityCooldownSetHooks4FBEA0[U any, P comparable] struct {
	loadNetCode     func(U) uint32
	playerByNetCode func(uint32) P
	loadPlayerIndex func(P) uint8
	storeCooldown   func(uint8, Ability, int32)
}

// playerAbilityCooldownSet4FBEA0 preserves GAME.EXE 004FBEA0. The original
// routine reads the unit's full 32-bit NetCode before resolving the live
// Player. A missing Player returns zero without a store; otherwise PlayerInd
// is loaded as an unsigned byte and selects one of six signed 32-bit cooldown
// words. A successful store returns the supplied 32-bit value in EAX.
//
// Ability is deliberately not validated. Production callers use slots 1..5,
// while the executable performs its signed dword index arithmetic directly.
// Likewise, there is no nil-unit guard before the first field load.
func playerAbilityCooldownSet4FBEA0[U any, P comparable](
	unit U,
	ability Ability,
	cooldown int32,
	hooks playerAbilityCooldownSetHooks4FBEA0[U, P],
) int32 {
	netCode := hooks.loadNetCode(unit)
	player := hooks.playerByNetCode(netCode)
	var zeroPlayer P
	if player == zeroPlayer {
		return 0
	}
	playerIndex := hooks.loadPlayerIndex(player)
	hooks.storeCooldown(playerIndex, ability, cooldown)
	return cooldown
}
