package server

// PlayerAbilityCooldownSet4FBEA0 is the native-width replacement for
// GAME.EXE 004FBEA0. The observed PlayerInd addresses the same fixed 32-by-6
// signed int32 matrix as the executable.
//
//go:noinline
func (a *serverAbilities) PlayerAbilityCooldownSet4FBEA0(unit *Object, ability Ability, cooldown int32) int32 {
	return playerAbilityCooldownSet4FBEA0(unit, ability, cooldown, playerAbilityCooldownSetHooks4FBEA0[
		*Object, *Player,
	]{
		loadNetCode: func(unit *Object) uint32 {
			return unit.NetCode
		},
		playerByNetCode: func(netCode uint32) *Player {
			return a.s.Players.ByID(int(netCode))
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		storeCooldown: func(index uint8, ability Ability, cooldown int32) {
			a.SetPlayerAbilityCooldownAt(index, ability, cooldown)
		},
	})
}
