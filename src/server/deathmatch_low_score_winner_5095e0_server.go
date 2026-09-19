package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// DeathmatchLowScoreWinnerRuntime5095E0 supplies the two network packet
// services still owned by legacy. All candidate traversal and score/status
// loads use native-width server layouts.
type DeathmatchLowScoreWinnerRuntime5095E0 struct {
	SendTeamWinner   func(*Team, uint8) int32
	SendPlayerWinner func(*Object, uint8) int32
}

func deathmatchLowScoreWinnerNative5095E0(
	firstTeam func() *Team,
	nextTeam func(*Team) *Team,
	firstPlayer func() *Object,
	nextPlayer func(*Object) *Object,
	runtime DeathmatchLowScoreWinnerRuntime5095E0,
) int32 {
	return deathmatchLowScoreWinner5095E0(
		deathmatchLowScoreWinnerHooks5095E0[*Team, *Object, *PlayerUpdateData, *Player]{
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
				return int32(player.Field2140)
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

// DeathmatchLowScoreWinner5095E0 binds GAME.EXE 005095E0 to native Object,
// PlayerUpdateData, Player, ObjectTeam, and Team layouts.
func (s *Server) DeathmatchLowScoreWinner5095E0(runtime DeathmatchLowScoreWinnerRuntime5095E0) int32 {
	return deathmatchLowScoreWinnerNative5095E0(
		s.Teams.First,
		s.Teams.Next,
		s.Players.FirstUnit,
		s.questNextPlayerUnit4DA7F0,
		runtime,
	)
}
