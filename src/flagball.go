package opennox

import "github.com/opennox/opennox/v1/common/memmap"

const (
	gameBallStatusBlob4E8290   uintptr = 0x5D4594
	gameBallStatusOffset4E8290 uintptr = 1567736

	teamFlagStatusBlob4E82C0 uintptr = 0x5D4594
)

func (s *Server) setGameBallStatus4E8290(state uint8, netCode uint16) int32 {
	record := gameBallStatusRecord4E8310()
	return gameBallStatusNative4E8290(record, state, netCode, func(recipient int32, state uint8, netCode uint16) int32 {
		return s.Nox_xxx_netSendBallStatus_4D95F0(recipient, state, netCode)
	})
}

func sub_4E8290(state uint8, netCode uint16) int32 {
	return noxServer.setGameBallStatus4E8290(state, netCode)
}

func gameBallStatusRecord4E8310() *gameBallStatusRecord4E8290 {
	return memmap.PtrT[gameBallStatusRecord4E8290](gameBallStatusBlob4E8290, gameBallStatusOffset4E8290)
}

func sub_4E8310() *gameBallStatusRecord4E8290 {
	return gameBallStatusGetter4E8310(gameBallStatusRecord4E8310())
}

func (s *Server) setTeamFlagStatus4E82C0(teamID, status, flagIndex uint8, carrierNetCode uint16) int32 {
	record := memmap.PtrT[teamFlagStatusRecord4E82C0](
		teamFlagStatusBlob4E82C0,
		teamFlagStatusRecordOffset4E82C0(teamID),
	)
	return teamFlagStatusNative4E82C0(record, teamID, status, flagIndex, carrierNetCode,
		func(recipient int32, teamID, status, flagIndex uint8, carrierNetCode uint16) int32 {
			return s.Nox_xxx_netSendFlagStatus_4D95A0(recipient, teamID, status, flagIndex, carrierNetCode)
		})
}

func sub_4E82C0(teamID, status, flagIndex uint8, carrierNetCode uint16) int32 {
	return noxServer.setTeamFlagStatus4E82C0(teamID, status, flagIndex, carrierNetCode)
}
