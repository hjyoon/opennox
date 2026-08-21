package legacy

/*
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

uint32_t* nox_xxx_protectPlayerHPMana_56F870(int token, uint16_t value);
void nox_xxx_unitClearBuffs_4FF580(nox_object_t* unit);

static inline void nox_playerReset_protectMana_4EFF10(
		uint32_t token, uint16_t value) {
	(void)nox_xxx_protectPlayerHPMana_56F870((int32_t)token, value);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerResetRuntime4EFF10() server.PlayerResetRuntime4EFF10 {
	outer := GetServer()
	s := outer.S()
	return server.PlayerResetRuntime4EFF10{
		AwardBeastScrolls: Nox_xxx_spellAwardAll1_4EFD80,
		AwardSpells:       Nox_xxx_spellAwardAll2_4EFC80,
		CancelAbilities:   Nox_xxx_playerCancelAbils_4FC180,
		ReadValues: func(unit *server.Object, rewardArg int32) {
			_ = playerReadValuesCall4EEDC0(unit, rewardArg)
		},
		AwardWarriorAbilities: Nox_xxx_spellAwardAll3_4EFE10,
		ProtectMana: func(token uint32, value uint16) {
			C.nox_playerReset_protectMana_4EFF10(
				C.uint32_t(token),
				C.uint16_t(value),
			)
		},
		SetHealthMaximum: Nox_xxx_unitHPsetOnMax_4EE6F0,
		SetPlayerState: func(unit *server.Object, state server.PlayerState) {
			_ = Nox_xxx_playerSetState_4FA020(unit, state)
		},
		ClearBuffs: func(unit *server.Object) {
			C.nox_xxx_unitClearBuffs_4FF580(asObjectC(unit))
		},
		CancelSpells:       Nox_xxx_playerCancelSpells_4FEAE0,
		RemovePoison:       Nox_xxx_removePoison_4EE9D0,
		ResetPlayerRuntime: Sub_4F7950,
		ReportTotalHealth: func(playerInd uint8, unit *server.Object) {
			_ = netReportTotalHealthNative4D85C0(s, playerInd, unit)
		},
		ReportTotalMana: func(playerInd uint8, unit *server.Object) {
			netReportTotalManaNative4D88C0(s, playerInd, unit)
		},
	}
}

func playerResetCall4EFF10(unit *server.Object) int32 {
	return GetServer().S().PlayerReset4EFF10(unit, playerResetRuntime4EFF10())
}
