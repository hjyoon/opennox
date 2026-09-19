package server

const (
	checkVictoryCompleteFlag509A60    = uint32(0x08)
	checkVictoryCoopTeamFlag509A60    = uint32(0x200)
	checkVictoryEliminationFlag509A60 = uint32(0x400)
	checkVictoryObserverFlag509A60    = uint32(0x01)
	checkVictoryWinnerArg509A60       = uint8(0x00)
)

// checkVictoryHooks509A60 exposes the live field reads and service calls made
// by GAME.EXE 00509A60. In particular, update data and Player are loaded before
// the observer/score tests, and score fields are not read for observers.
type checkVictoryHooks509A60[T comparable, O comparable, U, P any] struct {
	hasGameFlag       func(uint32) bool
	loadGameFlags     func() uint16
	loadScoreLimit    func(uint16) uint16
	firstTeam         func() T
	nextTeam          func(T) T
	loadTeamScore     func(T) int32
	firstPlayer       func() O
	nextPlayer        func(O) O
	loadUpdate        func(O) U
	loadPlayer        func(U) P
	loadPlayerFlags   func(P) uint32
	loadPlayerScore   func(P) int32
	loadPlayerDeaths  func(P) uint32
	hasTeam           func(O) bool
	loadObjectTeamID  func(O) uint8
	teamByID          func(uint8) T
	gameplayHasRivals func() bool
	setGameFlags      func(uint32)
	sendTeamWinner    func(T, uint8) int32
	sendPlayerWinner  func(O, uint8) int32
}

// checkVictory509A60 preserves the mode-specific winner check at GAME.EXE
// 00509A60. Elimination requires every surviving candidate to be either on one
// common team or the sole unteamed unit. Other deathmatch modes give a team
// reaching the score limit precedence over a Player reaching it. Game flag 8
// is set before the winner packet in every successful branch.
func checkVictory509A60[T comparable, O comparable, U, P any](
	hooks checkVictoryHooks509A60[T, O, U, P],
) {
	var nilTeam T
	var nilObject O

	if hooks.hasGameFlag(checkVictoryEliminationFlag509A60) {
		limit := uint32(hooks.loadScoreLimit(hooks.loadGameFlags()))
		if limit == 0 {
			return
		}

		var teamCandidate T
		var playerCandidate O
		for unit := hooks.firstPlayer(); unit != nilObject; unit = hooks.nextPlayer(unit) {
			update := hooks.loadUpdate(unit)
			player := hooks.loadPlayer(update)
			if hooks.loadPlayerFlags(player)&checkVictoryObserverFlag509A60 != 0 ||
				hooks.loadPlayerDeaths(player) >= limit {
				continue
			}
			if hooks.hasTeam(unit) {
				team := hooks.teamByID(hooks.loadObjectTeamID(unit))
				if teamCandidate != nilTeam {
					if teamCandidate != team {
						return
					}
				} else {
					teamCandidate = team
				}
				continue
			}
			if playerCandidate != nilObject || teamCandidate != nilTeam {
				return
			}
			playerCandidate = unit
		}

		if !hooks.gameplayHasRivals() {
			return
		}
		hooks.setGameFlags(checkVictoryCompleteFlag509A60)
		if teamCandidate != nilTeam {
			hooks.sendTeamWinner(teamCandidate, checkVictoryWinnerArg509A60)
		} else {
			hooks.sendPlayerWinner(playerCandidate, checkVictoryWinnerArg509A60)
		}
		return
	}

	if hooks.hasGameFlag(checkVictoryCoopTeamFlag509A60) {
		return
	}
	limit := int32(hooks.loadScoreLimit(hooks.loadGameFlags()))
	if limit == 0 {
		return
	}

	for team := hooks.firstTeam(); team != nilTeam; team = hooks.nextTeam(team) {
		if hooks.loadTeamScore(team) >= limit {
			hooks.setGameFlags(checkVictoryCompleteFlag509A60)
			hooks.sendTeamWinner(team, checkVictoryWinnerArg509A60)
			return
		}
	}
	for unit := hooks.firstPlayer(); unit != nilObject; unit = hooks.nextPlayer(unit) {
		update := hooks.loadUpdate(unit)
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerFlags(player)&checkVictoryObserverFlag509A60 == 0 &&
			hooks.loadPlayerScore(player) >= limit {
			hooks.setGameFlags(checkVictoryCompleteFlag509A60)
			hooks.sendPlayerWinner(unit, checkVictoryWinnerArg509A60)
			return
		}
	}
}
