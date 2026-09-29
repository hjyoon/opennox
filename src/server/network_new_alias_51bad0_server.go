package server

import "encoding/binary"

// NetworkNewAliasPacketSize51BAD0 is the exact MSG_NEW_ALIAS packet width.
const NetworkNewAliasPacketSize51BAD0 = networkNewAliasPacketSize51BAD0

// NetworkNewAlias51BAD0 binds the alias packet contract to the native Player
// layout on every pointer width.
func (s *Server) NetworkNewAlias51BAD0(
	player *Player,
	packet *[NetworkNewAliasPacketSize51BAD0]byte,
) int32 {
	return networkNewAlias51BAD0(player, networkNewAliasHooks51BAD0[*Player]{
		loadAlias: func() byte {
			return packet[1]
		},
		loadCode: func() uint16 {
			return binary.LittleEndian.Uint16(packet[2:4])
		},
		loadTypeID: func() uint16 {
			return binary.LittleEndian.Uint16(packet[4:6])
		},
		loadDeadline: func() uint32 {
			return binary.LittleEndian.Uint32(packet[6:10])
		},
		storeCode: func(player *Player, alias byte, value uint16) {
			player.NetData16[alias].Field0 = value
		},
		storeTypeID: func(player *Player, alias byte, value uint16) {
			player.NetData16[alias].Field2 = value
		},
		storeDeadline: func(player *Player, alias byte, value uint32) {
			player.NetData16[alias].Frame4 = value
		},
	})
}
