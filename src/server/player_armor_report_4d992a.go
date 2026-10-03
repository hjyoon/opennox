package server

import (
	"encoding/binary"
	"math"
)

// PlayerArmorReport4D992A restores the armor-delta slice of 004D9900.
// FCOMP's C3 test skips both equal and unordered values, not only equal bits.
// The packet keeps the pre-call binary32 bits; acknowledgement reloads the
// current bits through the entry-cached update pointer even if the send fails.
func (s *Server) PlayerArmorReport4D992A(update *PlayerUpdateData) {
	bits := update.Field57
	current, previous := math.Float32frombits(bits), math.Float32frombits(update.Field58)
	if current == previous || math.IsNaN(float64(current)) || math.IsNaN(float64(previous)) {
		return
	}
	packet := [5]byte{73}
	binary.LittleEndian.PutUint32(packet[1:], bits)
	s.NetSendPacketXxx0(int(update.Player.PlayerInd), packet[:], nil, 1)
	update.Field58 = update.Field57
}
