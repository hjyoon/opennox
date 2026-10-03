package server

// PlayerPoisonReport4D99A7 restores the poison-byte delta slice of 004D9900
// and the reliable two-byte sender at 004D8840. The update pointer stays
// entry-cached, but Player and the poison byte are reloaded after the send,
// even when it fails. Poison state/timing/status are not changed by reporting.
func (s *Server) PlayerPoisonReport4D99A7(unit *Object, update *PlayerUpdateData) {
	player := update.Player
	current := unit.Poison540
	if player.Field2172 == current {
		return
	}
	recipient := int(player.PlayerInd)
	packet := [2]byte{91, unit.Poison540}
	s.NetSendPacketXxx0(recipient, packet[:], nil, 1)

	player = update.Player
	current = unit.Poison540
	player.Field2172 = current
}
