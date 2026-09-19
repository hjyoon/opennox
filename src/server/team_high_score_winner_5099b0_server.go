package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func teamWinnerID5099B0(team *Team, none uint16) uint16 {
	if team == nil {
		return none
	}
	return uint16(uint8(team.ID()))
}

func teamWinnerPacket5099B0(op netmsg.Op, teamID uint16, arg uint8, frame uint32) [8]byte {
	var packet [8]byte
	packet[0] = byte(op)
	binary.LittleEndian.PutUint16(packet[1:], teamID)
	packet[3] = arg
	binary.LittleEndian.PutUint32(packet[4:], frame)
	return packet
}

func (s *Server) teamHighScoreSendFlagBall5099B0(team *Team) int32 {
	packet := teamWinnerPacket5099B0(
		netmsg.MSG_REPORT_FLAG_BALL_WINNER,
		teamWinnerID5099B0(team, 0),
		0,
		s.Frame(),
	)
	return int32(s.NetSendPacketXxx1(255, packet[:], nil, 1))
}

func (s *Server) teamHighScoreSendFlag5099B0(team *Team, arg uint8) int32 {
	packet := teamWinnerPacket5099B0(
		netmsg.MSG_REPORT_FLAG_WINNER,
		teamWinnerID5099B0(team, ^uint16(0)),
		arg,
		s.Frame(),
	)
	return int32(s.NetSendPacketXxx1(255, packet[:], nil, 1))
}

func teamHighScoreWinnerNative5099B0(
	firstTeam func() *Team,
	nextTeam func(*Team) *Team,
	sendFlagBall func(*Team) int32,
	sendFlagWinner func(*Team, uint8) int32,
) int32 {
	return teamHighScoreWinner5099B0(teamHighScoreWinnerHooks5099B0[*Team]{
		firstTeam: firstTeam,
		nextTeam:  nextTeam,
		loadTeamScore: func(team *Team) int32 {
			return team.Lessons
		},
		setGameFlags: func(flags uint32) {
			noxflags.SetGame(noxflags.GameFlag(flags))
		},
		hasGameFlags: func(flags uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flags))
		},
		sendFlagBall:   sendFlagBall,
		sendFlagWinner: sendFlagWinner,
	})
}

// TeamHighScoreWinner5099B0 binds GAME.EXE 005099B0 to native Team pointers
// and emits the winner packet without passing a Team address through C int.
func (s *Server) TeamHighScoreWinner5099B0() int32 {
	return teamHighScoreWinnerNative5099B0(
		s.Teams.First,
		s.Teams.Next,
		s.teamHighScoreSendFlagBall5099B0,
		s.teamHighScoreSendFlag5099B0,
	)
}
