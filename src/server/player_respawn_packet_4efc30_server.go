package server

type playerRespawnPacketNativeDeps4EFC30 struct {
	loadFrame       func() uint32
	loadWeaponFlags func() uint8
	sendSequence    func(recipient int32, packet []byte, related *Object, removeIfDisconnected int32) int32
}

func playerRespawnPacketNative4EFC30(
	unit *Object,
	keepItems uint8,
	deps playerRespawnPacketNativeDeps4EFC30,
) int32 {
	return playerRespawnPacket4EFC30(playerRespawnPacketHooks4EFC30[*Object]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadFrame: deps.loadFrame,
		loadNetCode: func(unit *Object) uint32 {
			return unit.NetCode
		},
		loadWeaponFlags: deps.loadWeaponFlags,
		loadKeepItemsArg: func() uint8 {
			return keepItems
		},
		sendSequence: func(recipient int32, packet [9]byte, related *Object, remove int32) int32 {
			return deps.sendSequence(recipient, packet[:], related, remove)
		},
	})
}

// NetSendPlayerRespawn4EFC30 sends GAME.EXE 004EFC30's exact sequence-enabled
// packet while retaining native-width object pointers in the server layer.
func (s *Server) NetSendPlayerRespawn4EFC30(unit *Object, keepItems uint8) int32 {
	return playerRespawnPacketNative4EFC30(unit, keepItems, playerRespawnPacketNativeDeps4EFC30{
		loadFrame:       s.Frame,
		loadWeaponFlags: s.RespawnWeaponFlags4EF580,
		sendSequence: func(recipient int32, packet []byte, related *Object, remove int32) int32 {
			return int32(s.NetSendPacketXxx1(int(recipient), packet, related, int(remove)))
		},
	})
}
