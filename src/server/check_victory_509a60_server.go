package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// CheckVictoryRuntime509A60 supplies the legacy configuration and packet
// services around the native-width winner traversal.
type CheckVictoryRuntime509A60 struct {
	ScoreLimit        func(uint16) uint16
	GameplayHasRivals func() bool
	SendTeamWinner    func(*Team, uint8) int32
	SendPlayerWinner  func(*Object, uint8) int32
}

func checkVictoryNative509A60(
	firstTeam func() *Team,
	nextTeam func(*Team) *Team,
	teamByID func(TeamID) *Team,
	firstPlayer func() *Object,
	nextPlayer func(*Object) *Object,
	runtime CheckVictoryRuntime509A60,
) {
	checkVictory509A60(checkVictoryHooks509A60[*Team, *Object, *PlayerUpdateData, *Player]{
		hasGameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		loadGameFlags: func() uint16 {
			return uint16(noxflags.GetGame())
		},
		loadScoreLimit: runtime.ScoreLimit,
		firstTeam:      firstTeam,
		nextTeam:       nextTeam,
		loadTeamScore: func(team *Team) int32 {
			return team.Lessons
		},
		firstPlayer: firstPlayer,
		nextPlayer:  nextPlayer,
		loadUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
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
		loadPlayerDeaths: func(player *Player) uint32 {
			return player.Field2140
		},
		hasTeam: func(unit *Object) bool {
			return unit.TeamVal.Has()
		},
		loadObjectTeamID: func(unit *Object) uint8 {
			return uint8(unit.TeamVal.ID)
		},
		teamByID: func(id uint8) *Team {
			return teamByID(TeamID(id))
		},
		gameplayHasRivals: runtime.GameplayHasRivals,
		setGameFlags: func(flags uint32) {
			noxflags.SetGame(noxflags.GameFlag(flags))
		},
		sendTeamWinner:   runtime.SendTeamWinner,
		sendPlayerWinner: runtime.SendPlayerWinner,
	})
}

// CheckVictory509A60 binds GAME.EXE 00509A60 to native Object,
// PlayerUpdateData, Player, ObjectTeam, and Team layouts.
func (s *Server) CheckVictory509A60(runtime CheckVictoryRuntime509A60) {
	checkVictoryNative509A60(
		s.Teams.First,
		s.Teams.Next,
		s.Teams.ByID,
		s.Players.FirstUnit,
		s.questNextPlayerUnit4DA7F0,
		runtime,
	)
}
