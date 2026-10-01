package server

import "encoding/binary"

type questLoseSpellHooks54CE00[O, U, P comparable] struct {
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

// questLoseSpell54CE00 restores GAME.EXE 0054CE00: cache Player once,
// count and then select using two live scans, including the empty 0..-1 RNG
// call. The original eligibility predicate protects the four starter spells.
func questLoseSpell54CE00[O, U, P comparable](unit O, h questLoseSpellHooks54CE00[O, U, P]) {
	player := h.loadPlayer(h.loadUpdate(unit))
	class := h.loadClass(player)
	if class != 2 && class != 1 {
		return
	}
	var count int32
	for id := int32(0); id < 137; id++ {
		if h.loadLevel(player, id) != 0 && h.eligible(id) != 0 {
			count++
		}
	}
	selected := h.randomInt(0, count-1)
	var ordinal int32
	for id := int32(0); id < 137; id++ {
		if h.loadLevel(player, id) == 0 || h.eligible(id) == 0 {
			continue
		}
		if ordinal == selected {
			h.storeLevel(player, id, 0)
			recipient := h.loadIndex(player)
			packet := [4]byte{0xf0, 0x11}
			binary.LittleEndian.PutUint16(packet[2:], uint16(id))
			h.sendPacket(recipient, packet)
			return
		}
		ordinal++
	}
}
