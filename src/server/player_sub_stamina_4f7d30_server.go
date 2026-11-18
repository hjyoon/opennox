package server

import "github.com/opennox/libs/noxnet/netmsg"

type playerSubStaminaNativeDeps4F7D30 struct {
	reportStamina func(uint8, *Object)
}

type playerSubStaminaReportNativeDeps4D8800 struct {
	sendReliable func(int32, [2]byte, *Object, int32) int32
}

func playerSubStaminaReportNative4D8800(
	playerIndex int32,
	unit *Object,
	deps playerSubStaminaReportNativeDeps4D8800,
) (int32, bool) {
	if unit == nil || uint8(unit.ObjClass)&uint8(playerSubStaminaPlayerClass4F7D30) == 0 || unit.UpdateData == nil {
		return 0, false
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	packet := [2]byte{byte(netmsg.MSG_REPORT_STAMINA), update.Stamina}
	return deps.sendReliable(playerIndex, packet, nil, 1), true
}

func playerSubStaminaNative4F7D30(
	unit *Object,
	amount int32,
	deps playerSubStaminaNativeDeps4F7D30,
) int32 {
	return playerSubStamina4F7D30(unit, playerSubStaminaHooks4F7D30[
		*Object,
		*PlayerUpdateData,
		*Player,
		*MonsterUpdateData,
	]{
		loadClass: func(unit *Object) uint32 {
			return uint32(unit.ObjClass)
		},
		loadPlayerUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadMonsterUpdate: func(unit *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(unit.UpdateData)
		},
		loadAmount: func() int32 {
			return amount
		},
		loadPlayerStamina: func(update *PlayerUpdateData) uint8 {
			return update.Stamina
		},
		storePlayerStamina: func(update *PlayerUpdateData, stamina uint8) {
			update.Stamina = stamina
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		reportStamina: deps.reportStamina,
		loadMonsterStamina: func(update *MonsterUpdateData) uint8 {
			return update.Stamina
		},
		storeMonsterStamina: func(update *MonsterUpdateData, stamina uint8) {
			update.Stamina = stamina
		},
	})
}

func (s *Server) NetReportStamina4D8800(recipient int32, unit *Object) (int32, bool) {
	return playerSubStaminaReportNative4D8800(recipient, unit, playerSubStaminaReportNativeDeps4D8800{
		sendReliable: func(recipient int32, packet [2]byte, related *Object, remove int32) int32 {
			return int32(s.NetSendPacketXxx1(int(recipient), packet[:], related, int(remove)))
		},
	})
}

func playerStaminaReportServer4D8800(s *Server) func(uint8, *Object) {
	return func(recipient uint8, unit *Object) {
		s.NetReportStamina4D8800(int32(recipient), unit)
	}
}

func playerSubStaminaServerDeps4F7D30(s *Server) playerSubStaminaNativeDeps4F7D30 {
	return playerSubStaminaNativeDeps4F7D30{
		reportStamina: playerStaminaReportServer4D8800(s),
	}
}

// PlayerSubStamina4F7D30 binds GAME.EXE 004F7D30 to native-width Object,
// PlayerUpdateData, Player, and MonsterUpdateData pointers.
func (s *Server) PlayerSubStamina4F7D30(unit *Object, amount int32) int32 {
	return playerSubStaminaNative4F7D30(unit, amount, playerSubStaminaServerDeps4F7D30(s))
}
