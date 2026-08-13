package legacy

import "encoding/binary"

const netPlayerStatusOpcode4D8270 = byte(102)

type netPlayerStatusHooks4D8270[O, U, P any] struct {
	flags      func(O) uint32
	updateData func(O) U
	player     func(U) P
	playerInd  func(P) byte
	send       func(byte, []byte) int
}

// netReportPlayerStatus4D8270 preserves the dependent reads in GAME.EXE
// 004D8270: flags and update-data are cached before the packet is formed,
// while Player and PlayerInd are loaded only at the send boundary.
func netReportPlayerStatus4D8270[O, U, P any](obj O, h netPlayerStatusHooks4D8270[O, U, P]) int {
	flags := h.flags(obj)
	updateData := h.updateData(obj)
	var packet [5]byte
	packet[0] = netPlayerStatusOpcode4D8270
	binary.LittleEndian.PutUint32(packet[1:], flags)
	return h.send(h.playerInd(h.player(updateData)), packet[:])
}
