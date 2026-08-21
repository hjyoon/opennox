package legacy

/*
#include "GAME3_3.h"

void nox_xxx_unitClearBuffs_4FF580(nox_object_t* unit);
int sub_4CFE00(void);
*/
import "C"

import (
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func netReportTotalManaNative4D88C0(s *server.Server, playerInd uint8, unit *server.Object) {
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	if uint8(unit.ObjClass)&0x04 == 0 || update != nil && update.Player.PlayerClass() == 0 {
		return
	}
	var packet [7]byte
	packet[0] = byte(netmsg.MSG_REPORT_TOTAL_MANA)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(unit)))
	binary.LittleEndian.PutUint16(packet[3:], update.ManaCur)
	binary.LittleEndian.PutUint16(packet[5:], update.ManaMax)
	sendImportantPacketWrapperC(int(playerInd), packet[:], nil, 1, importantPacketReplaceExisting)
}

func playerMakeDefItemsRuntime4EF7D0() server.PlayerMakeDefItemsRuntime4EF7D0 {
	outer := GetServer()
	s := outer.S()
	return server.PlayerMakeDefItemsRuntime4EF7D0{
		RemovePoison:     Nox_xxx_removePoison_4EE9D0,
		SetHealthMaximum: Nox_xxx_unitHPsetOnMax_4EE6F0,
		RefreshMana: func(unit *server.Object) {
			_ = Nox_xxx_playerManaRefresh_4EECF0(unit)
		},
		CancelAbilities: Nox_xxx_playerCancelAbils_4FC180,
		SetPlayerState: func(unit *server.Object, state server.PlayerState) {
			_ = Nox_xxx_playerSetState_4FA020(unit, state)
		},
		ClearBuffs: func(unit *server.Object) {
			C.nox_xxx_unitClearBuffs_4FF580(asObjectC(unit))
		},
		ResetPlayerRuntime: Sub_4F7950,
		ReportTotalHealth: func(playerInd uint8, unit *server.Object) {
			_ = netReportTotalHealthNative4D85C0(s, playerInd, unit)
		},
		ReportTotalMana: func(playerInd uint8, unit *server.Object) {
			netReportTotalManaNative4D88C0(s, playerInd, unit)
		},
		SendRespawn: func(unit *server.Object, keepItems uint8) uint8 {
			return uint8(s.NetSendPlayerRespawn4EFC30(unit, keepItems))
		},
		DelayedDelete: outer.DelayedDelete,
		RespawnItem: func(unit *server.Object, typeID string, attrs *server.ModifierInitData, a4, a5 int32) *server.Object {
			if attrs == nil {
				return playerRespawnItemCall4EF750(unit, typeID, nil, a4, a5)
			}
			cattrs, free := alloc.New(server.ModifierInitData{})
			defer free()
			*cattrs = *attrs
			return playerRespawnItemCall4EF750(unit, typeID, cattrs, a4, a5)
		},
		QuestDefaultsReady: func() int32 {
			return int32(C.sub_4CFE00())
		},
	}
}

func playerMakeDefItemsCall4EF7D0(unit *server.Object, restoreStats, keepItems int32) uint8 {
	return GetServer().S().PlayerMakeDefItems4EF7D0(
		unit,
		restoreStats,
		keepItems,
		playerMakeDefItemsRuntime4EF7D0(),
	)
}
