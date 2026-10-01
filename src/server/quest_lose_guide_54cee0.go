package server

import "encoding/binary"

type questLoseGuideHooks54CEE0[O, U, P comparable] struct {
	loadUpdate func(O) U
	loadPlayer func(U) P
	loadClass  func(P) uint8
	loadLevel  func(P, int32) uint32
	storeLevel func(P, int32, uint32)
	eligible   func(int32) int32
	randomInt  func(int32, int32) int32
	loadIndex  func(P) uint8
	sendPacket func(uint8, [4]byte) int32
}

// questLoseGuide54CEE0 restores GAME.EXE 0054CEE0's cached Player and two
// live scans of exactly-one guide DWORDs. Empty sets still call RNG(0, -1).
func questLoseGuide54CEE0[O, U, P comparable](unit O, h questLoseGuideHooks54CEE0[O, U, P]) {
	player := h.loadPlayer(h.loadUpdate(unit))
	if h.loadClass(player) != 2 {
		return
	}
	var count int32
	for id := int32(0); id < 41; id++ {
		if h.loadLevel(player, id) == 1 && h.eligible(id) != 0 {
			count++
		}
	}
	selected := h.randomInt(0, count-1)
	var ordinal int32
	for id := int32(0); id < 41; id++ {
		if h.loadLevel(player, id) != 1 || h.eligible(id) == 0 {
			continue
		}
		if ordinal == selected {
			h.storeLevel(player, id, 0)
			recipient := h.loadIndex(player)
			packet := [4]byte{0xf0, 0x13}
			binary.LittleEndian.PutUint16(packet[2:], uint16(id))
			h.sendPacket(recipient, packet)
			return
		}
		ordinal++
	}
}
