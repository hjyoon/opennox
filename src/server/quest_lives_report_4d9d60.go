package server

import "encoding/binary"

type questLivesReportHooks4D9D60[O, U any] struct {
	netCode func(O) uint16
	update  func(O) U
	lives   func(U) uint8
	send    func(int32, [5]byte) int32
}

// questLivesReport4D9D60 preserves GAME.EXE 004D9D60's exact five-byte
// packet. The low net-code WORD precedes the live UpdateData load, and only
// the low life BYTE is transmitted. There are no class or nil guards.
func questLivesReport4D9D60[O, U any](recipient int32, unit O, h questLivesReportHooks4D9D60[O, U]) int32 {
	code := h.netCode(unit)
	update := h.update(unit)
	packet := [5]byte{0xf0, 4}
	binary.LittleEndian.PutUint16(packet[3:], code)
	packet[2] = h.lives(update)
	return h.send(recipient, packet)
}
