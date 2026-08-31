#include "client__gui__guimsg.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5_2.h"
#include "common__strman.h"
#include "operators.h"

//----- (004D8060) --------------------------------------------------------
int nox_xxx_netAbilityReport_4D8060(nox_object_t* unit, int ability, int rewarded) {
	int result = (int)(uintptr_t)unit;
	if (unit->obj_class & UINT32_C(4)) {
		nox_player_update_data_t* update = unit->data_update;
		nox_playerInfo* player = update->player;
		uint8_t packet[3] = {
			UINT8_C(0xCD),
			(uint8_t)ability,
			(uint8_t)player->spell_lvl[ability],
		};
		if (rewarded) {
			packet[2] |= UINT8_C(0x80);
		}
		result = nox_xxx_netSendPacket1_4E5390(player->playerInd, packet, sizeof(packet), NULL, 1);
	}
	return result;
}

// GAME.EXE 004FB9C0 is restored by ability_reward_4fb9c0_export.go. Keep the
// ABI32 transcription as oracle provenance; active callers use the
// native-pointer CGo export declared in server__ability__ability.h.
#if 0
//----- (004FB9C0) --------------------------------------------------------
int nox_xxx_abilityRewardServ_4FB9C0_ability(nox_object_t* a1, int a2, int a3) {
	int result;    // eax
	wchar2_t* v10; // eax

	if (!(a1->obj_class & UINT32_C(4))) {
		return 0;
	}
	if (a2 <= 0 || a2 >= 6) {
		v10 = nox_strman_loadString_40F1D0("AwardAbilityError", 0, "C:\\NoxPost\\src\\Server\\Ability\\Ability.c", 108);
		nox_xxx_netSendLineMessage_4D9EB0(a1, v10);
		return 0;
	}
	nox_player_update_data_t* update = a1->data_update;
	nox_playerInfo* player = update->player;
	uint32_t* ability_level = &player->spell_lvl[a2];
	if (*ability_level) {
		nox_xxx_netPriMsgToPlayer_4DA2C0(a1, "use.c:HadAbility", 0);
		result = 0;
	} else {
		*ability_level = 5;
		ability_level = &((nox_player_update_data_t*)a1->data_update)->player->spell_lvl[a2];
		if (*ability_level > 5) {
			*ability_level = 5;
		}
		player = ((nox_player_update_data_t*)a1->data_update)->player;
		nox_xxx_playerAwardSpellProtectionCRC_56FCE0(player->prot_4636, a2, player->spell_lvl[a2]);
		nox_xxx_netAbilityReport_4D8060(a1, a2, a3);
		if (nox_common_gameFlags_check_40A5C0(4096)) {
			nox_xxx_netSendRewardNotify_4FAD50(a1, 2, a1, a2);
			if (!sub_419E60(a1)) {
				for (nox_object_t* i = nox_xxx_getFirstPlayerUnit_4DA7C0(); i; i = nox_xxx_getNextPlayerUnit_4DA7F0(i)) {
					if (i != a1) {
						nox_xxx_netSendRewardNotify_4FAD50(i, 2, a1, a2);
					}
				}
			}
		}
		result = 1;
	}
	return result;
}
#endif
