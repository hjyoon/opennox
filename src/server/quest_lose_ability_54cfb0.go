package server

import "encoding/binary"

type questLoseAbilityHooks54CFB0[O, U, P comparable] struct {
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

// questLoseAbility54CFB0 restores GAME.EXE 0054CFB0, including the cached
// Player, two live level scans, the empty 1..0 RNG call, and the low EAX byte.
func questLoseAbility54CFB0[O, U, P comparable](unit O, h questLoseAbilityHooks54CFB0[O, U, P]) int8 {
	player := h.loadPlayer(h.loadUpdate(unit))
	class := h.loadClass(player)
	if class != 0 {
		return int8(class)
	}
	var count int32
	for id := int32(0); id < 6; id++ {
		if h.loadLevel(player, id) != 0 && h.eligible(id) != 0 {
			count++
		}
	}
	result := h.randomInt(1, count)
	selected := result
	ordinal := int32(1)
	for id := int32(1); id < 6; id++ {
		if h.loadLevel(player, id) == 0 {
			continue
		}
		result = h.eligible(id)
		if result == 0 {
			continue
		}
		if ordinal == selected {
			h.storeLevel(player, id, 0)
			recipient := h.loadIndex(player)
			packet := [4]byte{0xf0, 0x12}
			binary.LittleEndian.PutUint16(packet[2:], uint16(id))
			return int8(h.sendPacket(recipient, packet))
		}
		ordinal++
	}
	return int8(result)
}
