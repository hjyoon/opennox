package server

import "github.com/opennox/opennox/v1/common/ntype"

func (a *serverAbilities) abilityRuntimeUnit4FBE60(fallback *Object, index uint8) *Object {
	if player := a.s.Players.ByIndRaw(ntype.PlayerInd(index)); player != nil && player.PlayerUnit != nil {
		return player.PlayerUnit
	}
	return fallback
}

// PlayerAbilityCooldownGet4FBE60 is the native-width replacement for
// GAME.EXE 004FBE60. The PE32 matrix is represented by the existing per-unit
// ability runtime map; the observed PlayerInd is resolved back to its current
// native Object without placing a pointer in a 32-bit cooldown cell.
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
			runtimeUnit := a.abilityRuntimeUnit4FBE60(unit, index)
			runtime := a.ByUnit[runtimeUnit]
			if runtime == nil {
				return 0
			}
			return int32(runtime.Cooldowns[ability])
		},
	})
}
