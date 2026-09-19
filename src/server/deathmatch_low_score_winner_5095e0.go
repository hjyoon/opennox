package server

import "math"

const (
	deathmatchLowScoreCompleteFlag5095E0 = uint32(0x08)
	deathmatchLowScoreObserverFlag5095E0 = uint32(0x01)
	deathmatchLowScoreWinnerArg5095E0    = uint8(0x01)
)

// deathmatchLowScoreWinnerHooks5095E0 exposes each live field read and service
// call made by GAME.EXE 005095E0. Object update data is loaded before the team
// test, while Player data and its status/score are not read for teamed units.
type deathmatchLowScoreWinnerHooks5095E0[T comparable, O comparable, U, P any] struct {
	firstTeam        func() T
	nextTeam         func(T) T
	loadTeamScore    func(T) int32
	firstPlayer      func() O
	loadUpdate       func(O) U
	hasTeam          func(O) bool
	loadPlayer       func(U) P
	loadPlayerFlags  func(P) uint32
	loadPlayerScore  func(P) int32
	nextPlayer       func(O) O
	setGameFlags     func(uint32)
	sendTeamWinner   func(T, uint8) int32
	sendPlayerWinner func(O, uint8) int32
}

// deathmatchLowScoreWinner5095E0 preserves the low-score winner resolver at
// GAME.EXE 005095E0. Team Lessons and Player.Field2140 are compared as signed
// 32-bit values. The original keeps separate team and player candidates: the
// first unteamed player tying a team replaces it without declaring a tie, but
// a second player with the same score does declare one. A later lower score
// clears any earlier tie. Game flag 8 is set before every notification.
func deathmatchLowScoreWinner5095E0[T comparable, O comparable, U, P any](
	hooks deathmatchLowScoreWinnerHooks5095E0[T, O, U, P],
) int32 {
	best := int32(math.MaxInt32)
	var teamCandidate T
	var playerCandidate O
	tied := false

	var nilTeam T
	for team := hooks.firstTeam(); team != nilTeam; team = hooks.nextTeam(team) {
		score := hooks.loadTeamScore(team)
		if score <= best {
			tied = score == best && teamCandidate != nilTeam
			best = hooks.loadTeamScore(team)
			teamCandidate = team
		}
	}

	var nilObject O
	for unit := hooks.firstPlayer(); unit != nilObject; unit = hooks.nextPlayer(unit) {
		update := hooks.loadUpdate(unit)
		if hooks.hasTeam(unit) {
			continue
		}
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerFlags(player)&deathmatchLowScoreObserverFlag5095E0 != 0 {
			continue
		}
		score := hooks.loadPlayerScore(player)
		if score <= best {
			tied = score == best && playerCandidate != nilObject
			best = score
			playerCandidate = unit
		}
	}

	hooks.setGameFlags(deathmatchLowScoreCompleteFlag5095E0)
	if tied {
		return hooks.sendTeamWinner(nilTeam, deathmatchLowScoreWinnerArg5095E0)
	}
	if playerCandidate != nilObject {
		return hooks.sendPlayerWinner(playerCandidate, deathmatchLowScoreWinnerArg5095E0)
	}
	if teamCandidate != nilTeam {
		return hooks.sendTeamWinner(teamCandidate, deathmatchLowScoreWinnerArg5095E0)
	}
	return 0
}
