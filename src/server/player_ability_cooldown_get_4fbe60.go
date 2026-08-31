package server

type playerAbilityCooldownGetHooks4FBE60[U any, P comparable] struct {
	loadNetCode     func(U) uint32
	playerByNetCode func(uint32) P
	loadPlayerIndex func(P) uint8
	loadCooldown    func(uint8, Ability) int32
}

// playerAbilityCooldownGet4FBE60 preserves GAME.EXE 004FBE60. The original
// routine reads the unit's full 32-bit NetCode before resolving the live
// Player. A missing Player returns zero; otherwise PlayerInd is loaded as an
// unsigned byte and selects one of six signed 32-bit cooldown words.
//
// Ability is deliberately not validated. Production callers use slots 1..5,
// while the executable performs its signed dword index arithmetic directly.
// Likewise, there is no nil-unit guard before the first field load.
func playerAbilityCooldownGet4FBE60[U any, P comparable](
	unit U,
	ability Ability,
	hooks playerAbilityCooldownGetHooks4FBE60[U, P],
) int32 {
	netCode := hooks.loadNetCode(unit)
	player := hooks.playerByNetCode(netCode)
	var zeroPlayer P
	if player == zeroPlayer {
		return 0
	}
	playerIndex := hooks.loadPlayerIndex(player)
	return hooks.loadCooldown(playerIndex, ability)
}
