#include "vote_5066d0.h"

#include <limits.h>

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "player_respawn_4f7ef0.h"

extern uint32_t nox_server_kickQuestPlayerMinVotes_229992;
extern uint32_t nox_server_resetQuestMinVotes_229988;
extern void* nox_alloc_vote_1599652;
extern uint32_t dword_5d4594_1599656;

static nox_vote_5066D0* vote_head_5066D0;

static nox_playerInfo* vote_player_5066D0(nox_object_t* unit) {
	if (unit == NULL || !(unit->obj_class & 4) || unit->data_update == NULL) {
		return NULL;
	}
	nox_player_update_data_t* update = (nox_player_update_data_t*)unit->data_update;
	return update->player;
}

static nox_object_team_t* vote_object_team_5066D0(nox_object_t* object) {
	return (nox_object_team_t*)&object->field_12;
}

static nox_playerInfo* vote_find_player_5066D0(const wchar2_t* name) {
	if (name == NULL) {
		return NULL;
	}
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player != NULL;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		if (player->active == 1 && nox_wcscmp(player->name_final, name) == 0) {
			return player;
		}
	}
	return NULL;
}

static nox_vote_5066D0* vote_find_5066D0(uint32_t type, nox_object_t* target) {
	for (nox_vote_5066D0* vote = vote_head_5066D0; vote != NULL; vote = vote->next) {
		if (vote->type == type && (target == NULL || vote->target == target)) {
			return vote;
		}
	}
	return NULL;
}

static char vote_add_voter_5066D0(nox_vote_5066D0* vote, uint32_t mask) {
	uint32_t result = vote->voters;
	if (!(result & mask)) {
		vote->voters |= mask;
		vote->count++;
		result = vote->count;
	}
	return (char)result;
}

static void vote_remove_voter_5066D0(nox_vote_5066D0* vote, uint32_t mask) {
	if (vote == NULL || !(vote->voters & mask)) {
		return;
	}
	vote->voters &= ~mask;
	vote->count--;
	if (vote->count == 0) {
		sub_5067B0(vote);
	}
}

int nox_xxx_allocVoteArray_5066D0(void) {
	nox_alloc_class* allocator = nox_new_alloc_class("VoteClass", (int)sizeof(nox_vote_5066D0), 64);
	nox_alloc_vote_1599652 = allocator;
	if (allocator == NULL) {
		return 0;
	}
	vote_head_5066D0 = NULL;
	// The original PE32 global cannot represent a native pointer. Keep it
	// cleared for old diagnostic readers while all live users use the head above.
	dword_5d4594_1599656 = 0;
	return 1;
}

void sub_506700(void) {
	if (nox_alloc_vote_1599652 != NULL) {
		nox_alloc_class_free_all((nox_alloc_class*)nox_alloc_vote_1599652);
	}
	vote_head_5066D0 = NULL;
	dword_5d4594_1599656 = 0;
}

int sub_506720(void) {
	if (nox_alloc_vote_1599652 != NULL) {
		nox_free_alloc_class((nox_alloc_class*)nox_alloc_vote_1599652);
	}
	nox_alloc_vote_1599652 = NULL;
	vote_head_5066D0 = NULL;
	dword_5d4594_1599656 = 0;
	return 0;
}

int sub_506740(nox_object_t* player_unit) {
	nox_playerInfo* player = vote_player_5066D0(player_unit);
	if (player == NULL || player->playerInd >= 32) {
		return 0;
	}
	const uint32_t player_mask = UINT32_C(1) << player->playerInd;
	for (nox_vote_5066D0* vote = vote_head_5066D0; vote != NULL;) {
		nox_vote_5066D0* next = vote->next;
		if (vote->voters & player_mask) {
			vote->voters &= ~player_mask;
			vote->count--;
		}
		// GAME.EXE also removes an already-empty record even when this
		// player's bit was not set.
		if (vote->count == 0) {
			sub_5067B0(vote);
		}
		vote = next;
	}
	return 0;
}

void sub_5067B0(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return;
	}
	if (vote->type == 2) {
		for (int player_index = 0; player_index < 32; player_index++) {
			if (vote->voters & (UINT32_C(1) << player_index)) {
				nox_xxx_netSendVote_506840(player_index);
			}
		}
	}
	sub_506810(vote);
	if (nox_alloc_vote_1599652 != NULL) {
		nox_alloc_class_free_obj_first((nox_alloc_class*)nox_alloc_vote_1599652, vote);
	}
	if (vote_head_5066D0 == NULL) {
		sub_507190(255, 0);
	}
}

nox_vote_5066D0* sub_506810(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return NULL;
	}
	nox_vote_5066D0* result = vote;
	if (vote->next != NULL) {
		vote->next->previous = vote->previous;
	}
	if (vote->previous != NULL) {
		result = vote->next;
		vote->previous->next = vote->next;
	} else {
		vote_head_5066D0 = vote->next;
	}
	return result;
}

int nox_xxx_netSendVote_506840(int player_index) {
	const uint8_t packet[2] = {0xEE, 7};
	return nox_xxx_netSendPacket1_4E5390(player_index, packet, sizeof(packet), NULL, 1);
}

char sub_506870(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	if (player_unit == NULL || !(player_unit->obj_class & 4)) {
		return 0;
	}
	switch (type) {
	case 0:
	case 1:
		return sub_5068E0(type, player_unit, player_name);
	case 2:
		return sub_506B00(type, player_unit) != NULL;
	case 3:
		return sub_506B80(type, player_unit, player_name) != NULL;
	default:
		return 0;
	}
}

char sub_5068E0(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	const uint32_t minimum_votes = *getMemU32Ptr(0x587000, 229980);
	if (minimum_votes == 0 || minimum_votes > 32 || player_name == NULL) {
		return (char)minimum_votes;
	}
	nox_playerInfo* voter = vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32) {
		return 0;
	}
	nox_playerInfo* target_player = vote_find_player_5066D0(player_name);
	if (target_player == NULL || target_player->playerInd == 31 || target_player->playerUnit == NULL ||
		target_player->playerUnit == player_unit) {
		return 0;
	}
	nox_object_t* target = target_player->playerUnit;
	const int team_mode = nox_xxx_CheckGameplayFlags_417DA0(4);
	if (team_mode && !nox_xxx_servCompareTeams_419150(
				 vote_object_team_5066D0(player_unit), vote_object_team_5066D0(target))) {
		return 0;
	}
	nox_vote_5066D0* vote = vote_find_5066D0((uint32_t)type, target);
	if (vote == NULL) {
		vote = sub_506A20(type, player_unit);
		if (vote == NULL) {
			return 0;
		}
		vote->target = target;
		vote->team_mode = team_mode ? 1u : 0u;
	}
	return vote_add_voter_5066D0(vote, UINT32_C(1) << voter->playerInd);
}

nox_vote_5066D0* nox_vote_record_alloc_5066D0(void) {
	if (nox_alloc_vote_1599652 == NULL) {
		return NULL;
	}
	return (nox_vote_5066D0*)nox_alloc_class_new_obj_zero((nox_alloc_class*)nox_alloc_vote_1599652);
}

nox_vote_5066D0* sub_506A20(int type, nox_object_t* player_unit) {
	if (player_unit == NULL || !(player_unit->obj_class & 4)) {
		return NULL;
	}
	const int was_empty = vote_head_5066D0 == NULL;
	nox_vote_5066D0* vote = nox_vote_record_alloc_5066D0();
	if (vote == NULL) {
		return NULL;
	}
	vote->type = (uint32_t)type;
	vote->frame = gameFrame();
	vote->team = vote_object_team_5066D0(player_unit);
	switch (type) {
	case 0:
	case 1:
		vote->threshold = getMemByte(0x587000, 229980);
		break;
	case 2:
	case 3:
		vote->threshold = 6;
		break;
	default:
		vote->threshold = getMemByte(0x587000, 229984);
		break;
	}
	nox_xxx_voteAddMB_506AD0(vote);
	if (was_empty) {
		sub_507190(255, 1);
	}
	return vote;
}

nox_vote_5066D0* nox_xxx_voteAddMB_506AD0(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return NULL;
	}
	vote->previous = NULL;
	vote->next = vote_head_5066D0;
	if (vote_head_5066D0 != NULL) {
		vote_head_5066D0->previous = vote;
	}
	vote_head_5066D0 = vote;
	return vote;
}

nox_vote_5066D0* sub_506B00(int type, nox_object_t* player_unit) {
	if (nox_server_resetQuestMinVotes_229988 == 0) {
		return NULL;
	}
	nox_playerInfo* voter = vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32 || voter->field_4792 == 0) {
		return NULL;
	}
	nox_vote_5066D0* vote = vote_find_5066D0((uint32_t)type, NULL);
	if (vote == NULL) {
		vote = sub_506A20(type, player_unit);
		if (vote == NULL) {
			return NULL;
		}
		vote->team_mode = 0;
	}
	vote_add_voter_5066D0(vote, UINT32_C(1) << voter->playerInd);
	return vote;
}

nox_vote_5066D0* sub_506B80(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	if (nox_server_kickQuestPlayerMinVotes_229992 == 0 || player_name == NULL) {
		return NULL;
	}
	nox_playerInfo* voter = vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32 || voter->field_4792 == 0) {
		return NULL;
	}
	nox_playerInfo* target_player = vote_find_player_5066D0(player_name);
	if (target_player == NULL || target_player->playerInd == 31 || target_player->field_4792 == 0 ||
		target_player->playerUnit == NULL || target_player->playerUnit == player_unit) {
		return NULL;
	}
	nox_object_t* target = target_player->playerUnit;
	nox_vote_5066D0* vote = vote_find_5066D0((uint32_t)type, target);
	if (vote == NULL) {
		vote = sub_506A20(type, player_unit);
		if (vote == NULL) {
			return NULL;
		}
		vote->target = target;
	}
	vote_add_voter_5066D0(vote, UINT32_C(1) << voter->playerInd);
	return vote;
}

void sub_506C90(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	if (player_unit == NULL || !(player_unit->obj_class & 4)) {
		return;
	}
	switch (type) {
	case 0:
	case 1:
		sub_506D00(player_unit, player_name);
		break;
	case 2:
		sub_506DE0(player_unit);
		break;
	case 3:
		sub_506E50(player_unit, player_name);
		break;
	default:
		break;
	}
}

void sub_506D00(nox_object_t* player_unit, wchar2_t* player_name) {
	nox_playerInfo* voter = vote_player_5066D0(player_unit);
	nox_playerInfo* target_player = vote_find_player_5066D0(player_name);
	if (voter == NULL || voter->playerInd >= 32 || target_player == NULL ||
		target_player->playerInd == 31 || target_player->playerUnit == NULL) {
		return;
	}
	const uint32_t mask = UINT32_C(1) << voter->playerInd;
	for (nox_vote_5066D0* vote = vote_head_5066D0; vote != NULL; vote = vote->next) {
		// GAME.EXE routes both remove opcodes 0 and 1 through the type-0 lookup.
		if (vote->type == 0 && vote->target == target_player->playerUnit && (vote->voters & mask)) {
			vote_remove_voter_5066D0(vote, mask);
			return;
		}
	}
}

void sub_506DE0(nox_object_t* player_unit) {
	nox_playerInfo* voter = vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32) {
		return;
	}
	const uint32_t mask = UINT32_C(1) << voter->playerInd;
	for (nox_vote_5066D0* vote = vote_head_5066D0; vote != NULL; vote = vote->next) {
		if (vote->type == 2) {
			vote_remove_voter_5066D0(vote, mask);
			return;
		}
	}
}

void sub_506E50(nox_object_t* player_unit, wchar2_t* player_name) {
	nox_playerInfo* voter = vote_player_5066D0(player_unit);
	nox_playerInfo* target_player = vote_find_player_5066D0(player_name);
	if (voter == NULL || voter->playerInd >= 32 || target_player == NULL ||
		target_player->playerInd == 31 || target_player->playerUnit == NULL) {
		return;
	}
	const uint32_t mask = UINT32_C(1) << voter->playerInd;
	for (nox_vote_5066D0* vote = vote_head_5066D0; vote != NULL; vote = vote->next) {
		if (vote->type == 3 && vote->target == target_player->playerUnit && (vote->voters & mask)) {
			vote_remove_voter_5066D0(vote, mask);
			return;
		}
	}
}

void nox_xxx_voteUptate_506F30(void) {
	for (nox_vote_5066D0* vote = vote_head_5066D0; vote != NULL;) {
		nox_vote_5066D0* next = vote->next;
		switch (vote->type) {
		case 0:
		case 1:
			sub_506F80(vote);
			break;
		case 2:
			sub_507090(vote);
			break;
		case 3:
			sub_507100(vote);
			break;
		default:
			break;
		}
		vote = next;
	}
}

void sub_506F80(nox_vote_5066D0* vote) {
	if (vote == NULL || vote->target == NULL || (vote->target->obj_flags & 0x20)) {
		sub_5067B0(vote);
		return;
	}
	vote->team = vote_object_team_5066D0(vote->target);
	if (sub_507000(vote) != 1) {
		return;
	}
	nox_playerInfo* target_player = vote_player_5066D0(vote->target);
	if (target_player != NULL) {
		nox_xxx_playerCallDisconnect_4DEAB0(target_player->playerInd, 4);
		sub_416770(15, target_player->name_final, target_player->serial);
	}
	sub_5067B0(vote);
}

int sub_507000(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return 0;
	}
	if (vote->count >= vote->threshold) {
		return 1;
	}
	uint32_t player_count = 0;
	for (nox_object_t* player = nox_xxx_getFirstPlayerUnit_4DA7C0(); player != NULL;
		 player = nox_xxx_getNextPlayerUnit_4DA7F0(player)) {
		if (vote->team_mode != 1 ||
			nox_xxx_servCompareTeams_419150(vote->team, vote_object_team_5066D0(player))) {
			player_count++;
		}
	}
	return vote->count >= (uint32_t)(player_count - 1) && vote->count >= 2;
}

static uint32_t vote_quest_player_count_5066D0(void) {
	uint32_t count = 0;
	for (nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0(); unit != NULL;
		 unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit)) {
		nox_playerInfo* player = vote_player_5066D0(unit);
		if (player == NULL) {
			continue;
		}
		if ((!nox_common_gameFlags_check_40A5C0(1) ||
			 !nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) ||
			 player->playerInd != 31) &&
			player->field_4792 == 1) {
			count++;
		}
	}
	return count;
}

void sub_507090(nox_vote_5066D0* vote) {
	if (vote == NULL || vote->count < vote_quest_player_count_5066D0()) {
		return;
	}
	for (nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0(); unit != NULL;
		 unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit)) {
		nox_playerInfo* player = vote_player_5066D0(unit);
		if (player != NULL && player->field_4792 == 1) {
			nox_xxx_playerRespawn_4F7EF0(player->playerUnit);
		}
	}
	nox_game_setQuestStage_4E3CD0(0);
	nox_xxx_mapLoad_4D2450(nox_xxx_getQuestMapFile_4D0F60());
	sub_5067B0(vote);
}

void sub_507100(nox_vote_5066D0* vote) {
	if (vote == NULL || vote->target == NULL || (vote->target->obj_flags & 0x20)) {
		sub_5067B0(vote);
		return;
	}
	nox_playerInfo* target_player = vote_player_5066D0(vote->target);
	if (target_player == NULL || target_player->field_4792 == 0) {
		sub_5067B0(vote);
		return;
	}
	if (vote->count < vote->threshold) {
		const uint32_t player_count = vote_quest_player_count_5066D0();
		if (player_count <= 1) {
			sub_5067B0(vote);
			return;
		}
		if (!(vote->count >= player_count - 1 && vote->count >= 2)) {
			return;
		}
	}
	sub_4DCFB0(vote->target);
	sub_416770(15, target_player->name_final, target_player->serial);
	sub_5067B0(vote);
}

int sub_507190(int recipient, char active) {
	const uint8_t packet[3] = {0xEE, 6, (uint8_t)active};
	return nox_xxx_netSendPacket1_4E5390(recipient, packet, sizeof(packet), NULL, 1);
}

int sub_5071C0(void) {
	return vote_head_5066D0 != NULL;
}

nox_vote_5066D0* nox_vote_head_5066D0(void) {
	return vote_head_5066D0;
}

void* nox_vote_allocator_5066D0(void) {
	return nox_alloc_vote_1599652;
}

size_t nox_vote_record_size_5066D0(void) {
	return sizeof(nox_vote_5066D0);
}

nox_vote_5066D0* nox_vote_next_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? NULL : vote->next;
}

nox_vote_5066D0* nox_vote_previous_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? NULL : vote->previous;
}

nox_object_team_t* nox_vote_team_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? NULL : vote->team;
}

uint32_t nox_vote_frame_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->frame;
}

uint32_t nox_vote_voters_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->voters;
}

uint8_t nox_vote_count_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->count;
}

uint8_t nox_vote_threshold_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->threshold;
}

void nox_vote_set_voters_5066D0(nox_vote_5066D0* vote, uint32_t voters, uint8_t count) {
	if (vote != NULL) {
		vote->voters = voters;
		vote->count = count;
	}
}
