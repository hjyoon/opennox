package server

// PlayerAbilityCooldownGet4FBE60 is the native-width replacement for
// GAME.EXE 004FBE60. The observed PlayerInd addresses the same fixed 32-by-6
// signed int32 matrix as the executable; no object pointer is stored in a
// cooldown cell.
//
//go:noinline
func (a *serverAbilities) PlayerAbilityCooldownGet4FBE60(unit *Object, ability Ability) int32 {
	return playerAbilityCooldownGet4FBE60(unit, ability, playerAbilityCooldownGetHooks4FBE60[
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
		loadCooldown: func(index uint8, ability Ability) int32 {
			return a.PlayerAbilityCooldownAt(index, ability)
		},
	})
}
