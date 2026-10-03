package server

// PlayerPoisonReport4D99A7 restores the item-enchantment byte slice of 004D9900
// and the reliable two-byte sender at 004D8840. The update pointer stays
// entry-cached, but Player and the low Field110 byte are reloaded after the send,
// even when it fails. Original object +440 is Field110, not Poison540 (+540).
func (s *Server) PlayerPoisonReport4D99A7(unit *Object, update *PlayerUpdateData) {
	player := update.Player
	current := byte(unit.Field110)
	if player.Field2172 == current {
		return
	}
	recipient := int(player.PlayerInd)
	packet := [2]byte{91, byte(unit.Field110)}
	s.NetSendPacketXxx0(recipient, packet[:], nil, 1)

	player = update.Player
	current = byte(unit.Field110)
	player.Field2172 = current
}
