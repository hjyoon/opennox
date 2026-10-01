package server

func (s *Server) QuestLivesReport4D9D60(recipient int32, unit *Object) int32 {
	return questLivesReport4D9D60(recipient, unit, questLivesReportHooks4D9D60[*Object, *PlayerUpdateData]{
		netCode: func(unit *Object) uint16 { return uint16(unit.NetCode) },
		update:  func(unit *Object) *PlayerUpdateData { return (*PlayerUpdateData)(unit.UpdateData) },
		lives:   func(update *PlayerUpdateData) uint8 { return uint8(update.ExtraLives) },
		send: func(recipient int32, packet [5]byte) int32 {
			return int32(s.NetSendPacketXxx0(int(recipient), packet[:], nil, 1))
		},
	})
}
