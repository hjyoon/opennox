package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// DeathmatchHighScoreWinnerRuntime5098A0 supplies the two winner packet
// services whose retained ABI is shared with the low-score resolver.
type DeathmatchHighScoreWinnerRuntime5098A0 = DeathmatchLowScoreWinnerRuntime5095E0

func deathmatchHighScoreWinnerNative5098A0(
	firstTeam func() *Team,
	nextTeam func(*Team) *Team,
	firstPlayer func() *Object,
	nextPlayer func(*Object) *Object,
	runtime DeathmatchHighScoreWinnerRuntime5098A0,
) int32 {
	return deathmatchHighScoreWinner5098A0(
		deathmatchHighScoreWinnerHooks5098A0[*Team, *Object, *PlayerUpdateData, *Player]{
			firstTeam: firstTeam,
			nextTeam:  nextTeam,
			loadTeamScore: func(team *Team) int32 {
				return team.Lessons
			},
			firstPlayer: firstPlayer,
			loadUpdate: func(unit *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(unit.UpdateData)
			},
			hasTeam: func(unit *Object) bool {
				return unit.TeamVal.Has()
			},
			loadPlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			loadPlayerFlags: func(player *Player) uint32 {
				return player.Field3680
			},
			loadPlayerScore: func(player *Player) int32 {
				return player.Lessons
			},
			nextPlayer: nextPlayer,
			setGameFlags: func(flags uint32) {
				noxflags.SetGame(noxflags.GameFlag(flags))
			},
			sendTeamWinner:   runtime.SendTeamWinner,
			sendPlayerWinner: runtime.SendPlayerWinner,
		},
	)
}

// DeathmatchHighScoreWinner5098A0 binds GAME.EXE 005098A0 to native-width
// Team, Object, PlayerUpdateData, and Player layouts.
func (s *Server) DeathmatchHighScoreWinner5098A0(runtime DeathmatchHighScoreWinnerRuntime5098A0) int32 {
	return deathmatchHighScoreWinnerNative5098A0(
		s.Teams.First,
		s.Teams.Next,
		s.Players.FirstUnit,
		s.questNextPlayerUnit4DA7F0,
		runtime,
	)
}
