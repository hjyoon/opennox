package server

const playerGoldReportClassLow4D9900 = uint8(0x04)

type playerGoldReportNativeDeps4D996D struct {
	sendReliable func(int32, [ShopGoldReportPacketSize4D8870]byte, *Object, int32) int32
}

func goldReportNative4D8870(
	recipient int32,
	unit *Object,
	deps playerGoldReportNativeDeps4D996D,
) {
	if uint8(unit.ObjClass)&playerGoldReportClassLow4D9900 == 0 {
		return
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	packet := BuildShopGoldReportPacket4D8870(player.GoldVal)
	deps.sendReliable(recipient, packet, nil, 1)
}

func playerGoldReportNative4D996D(
	unit *Object,
	update *PlayerUpdateData,
	deps playerGoldReportNativeDeps4D996D,
) {
	playerGoldReport4D996D(unit, update, playerGoldReportHooks4D996D[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadReportedGold: func(player *Player) uint32 {
			return player.Field2168
		},
		loadGold: func(player *Player) uint32 {
			return player.GoldVal
		},
		loadPlayerIndex: func(player *Player) int32 {
			return int32(player.PlayerInd)
		},
		reportGold: func(recipient int32, unit *Object) {
			goldReportNative4D8870(recipient, unit, deps)
		},
		storeReported: func(player *Player, value uint32) {
			player.Field2168 = value
		},
	})
}

// PlayerGoldReportSync4D9900 restores the gold synchronization slice of
// GAME.EXE's player-report routine. The outer nil and Player-class gates stay
// at 004D9900; the exact changed-gold block starts at 004D996D.
func (s *Server) PlayerGoldReportSync4D9900(unit *Object) {
	if unit == nil || uint8(unit.ObjClass)&playerGoldReportClassLow4D9900 == 0 {
		return
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	playerGoldReportNative4D996D(unit, update, playerGoldReportNativeDeps4D996D{
		sendReliable: func(recipient int32, packet [ShopGoldReportPacketSize4D8870]byte, related *Object, remove int32) int32 {
			return int32(s.NetSendPacketXxx0(int(recipient), packet[:], related, int(remove)))
		},
	})
}
