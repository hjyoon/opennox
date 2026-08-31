package server

// PlayerAbilityCooldownSet4FBEA0 is the native-width replacement for
// GAME.EXE 004FBEA0. The PE32 matrix is represented by the existing per-unit
// ability runtime map; the observed PlayerInd is resolved back to its current
// native Object before the selected signed 32-bit value is stored.
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
			runtimeUnit := a.abilityRuntimeUnitByPlayerIndex(unit, index)
			runtime := a.GetFor(runtimeUnit)
			runtime.Cooldowns[ability] = int(cooldown)
		},
	})
}
