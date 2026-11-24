package legacy

/*
#include "GAME3_2.h"

extern uint32_t dword_5d4594_1556144;
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func questObserverReadyPlayer4D7150(unit *server.Object) *server.Player {
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	if update == nil || update.Player == nil {
		return nil
	}
	player := update.Player
	if player.Field3680&0x11 != 0x11 || player.Field4792 != 1 ||
		update.QuestExit != nil || update.QuestWarpGate != nil {
		return nil
	}
	return player
}

//export sub_4D7150
func sub_4D7150() {
	if C.dword_5d4594_1556144 == 0 || gameFrame() <= uint32(C.dword_5d4594_1556144) {
		return
	}
	s := GetServer().S()
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.NextPlayerUnit4DA7F0(unit) {
		if player := questObserverReadyPlayer4D7150(unit); player != nil {
			Sub_4DF3C0(player)
			Nox_xxx_playerLeaveObserver_0_4E6AA0(player)
			Nox_xxx_playerCameraUnlock_4E6040(unit)
		}
	}
}
