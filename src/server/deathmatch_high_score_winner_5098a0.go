package server

import "math"

const (
	deathmatchHighScoreCompleteFlag5098A0 = uint32(0x08)
	deathmatchHighScoreObserverFlag5098A0 = uint32(0x01)
	deathmatchHighScoreWinnerArg5098A0    = uint8(0x01)
)

// deathmatchHighScoreWinnerHooks5098A0 exposes every live field read and
// service call made by GAME.EXE 005098A0. In particular, UpdateData is read
// even for teamed units, while their Player record is not read.
type deathmatchHighScoreWinnerHooks5098A0[T comparable, O comparable, U, P any] struct {
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

// deathmatchHighScoreWinner5098A0 preserves the signed high-score resolver at
// GAME.EXE 005098A0. It deliberately keeps the executable's asymmetric tie
// rule: the first unteamed Player tying a team becomes the Player candidate;
// only a second Player at that score marks the result tied.
func deathmatchHighScoreWinner5098A0[T comparable, O comparable, U, P any](
	hooks deathmatchHighScoreWinnerHooks5098A0[T, O, U, P],
) int32 {
	best := int32(math.MinInt32)
	var teamCandidate T
	var playerCandidate O
	tied := false

	var nilTeam T
	for team := hooks.firstTeam(); team != nilTeam; team = hooks.nextTeam(team) {
		score := hooks.loadTeamScore(team)
		if score >= best {
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
		if hooks.loadPlayerFlags(player)&deathmatchHighScoreObserverFlag5098A0 != 0 {
			continue
		}
		score := hooks.loadPlayerScore(player)
		if score >= best {
			tied = score == best && playerCandidate != nilObject
			best = score
			playerCandidate = unit
		}
	}

	hooks.setGameFlags(deathmatchHighScoreCompleteFlag5098A0)
	if tied {
		return hooks.sendTeamWinner(nilTeam, deathmatchHighScoreWinnerArg5098A0)
	}
	if playerCandidate != nilObject {
		return hooks.sendPlayerWinner(playerCandidate, deathmatchHighScoreWinnerArg5098A0)
	}
	if teamCandidate != nilTeam {
		return hooks.sendTeamWinner(teamCandidate, deathmatchHighScoreWinnerArg5098A0)
	}
	return 0
}
