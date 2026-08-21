package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
)

const playerRespawnPacketType4EFC30 uint8 = uint8(netmsg.MSG_PLAYER_RESPAWN)

type playerRespawnPacketHooks4EFC30[O comparable] struct {
	loadUnitArg      func() O
	loadFrame        func() uint32
	loadNetCode      func(O) uint32
	loadWeaponFlags  func() uint8
	loadKeepItemsArg func() uint8
	sendSequence     func(recipient int32, packet [9]byte, related O, removeIfDisconnected int32) int32
}

// playerRespawnPacket4EFC30 preserves GAME.EXE 004EFC30's observable load,
// packet-store, callback, and return order. The unit argument is cached before
// the live frame read; NetCode contributes only its low 16 bits. Weapon flags
// are read before the keep-items byte, and the sequence-enabled sender's full
// int32 result is returned unchanged.
func playerRespawnPacket4EFC30[O comparable](hooks playerRespawnPacketHooks4EFC30[O]) int32 {
	unit := hooks.loadUnitArg()
	frame := hooks.loadFrame()

	var packet [9]byte
	packet[0] = playerRespawnPacketType4EFC30
	binary.LittleEndian.PutUint32(packet[3:7], frame)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(hooks.loadNetCode(unit)))
	packet[7] = hooks.loadWeaponFlags()
	packet[8] = hooks.loadKeepItemsArg()

	var related O
	return hooks.sendSequence(255, packet, related, 0)
}
