package server

const (
	teamHighScoreCompleteFlag5099B0 = uint32(0x08)
	teamHighScoreWinnerArg5099B0    = uint8(0x01)
	teamHighScoreFlagBallMask5099B0 = uint32(0x40)
)

type teamHighScoreWinnerHooks5099B0[T comparable] struct {
	firstTeam      func() T
	nextTeam       func(T) T
	loadTeamScore  func(T) int32
	setGameFlags   func(uint32)
	hasGameFlags   func(uint32) bool
	sendFlagBall   func(T) int32
	sendFlagWinner func(T, uint8) int32
}

// teamHighScoreWinner5099B0 preserves GAME.EXE 005099B0. Team Lessons are
// signed dwords, -1 is the initial threshold, equal leaders produce a nil
// winner, and the mode test occurs only after game flag 8 is set.
func teamHighScoreWinner5099B0[T comparable](hooks teamHighScoreWinnerHooks5099B0[T]) int32 {
	best := int32(-1)
	var candidate T
	tied := false

	var nilTeam T
	for team := hooks.firstTeam(); team != nilTeam; team = hooks.nextTeam(team) {
		score := hooks.loadTeamScore(team)
		if score >= best {
			tied = score == best && candidate != nilTeam
			best = hooks.loadTeamScore(team)
			candidate = team
		}
	}

	hooks.setGameFlags(teamHighScoreCompleteFlag5099B0)
	if tied {
		candidate = nilTeam
	}
	if hooks.hasGameFlags(teamHighScoreFlagBallMask5099B0) {
		return hooks.sendFlagBall(candidate)
	}
	return hooks.sendFlagWinner(candidate, teamHighScoreWinnerArg5099B0)
}
