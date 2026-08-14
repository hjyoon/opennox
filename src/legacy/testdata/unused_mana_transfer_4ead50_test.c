// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only the records read by the unreferenced 004EAD50 routine.
#define _Static_assert(...)
#include "../common__system__team.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, field_12) == (sizeof(void*) == 4 ? 48 : 52),
	"object team prefix offset");
_Static_assert(offsetof(nox_object_t, field_13) == (sizeof(void*) == 4 ? 52 : 56),
	"object team ID word offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(offsetof(nox_player_update_data_t, mana_cur) == 4,
	"player current-mana offset");
_Static_assert(offsetof(nox_player_update_data_t, mana_prev) == 6,
	"player previous-mana offset");
_Static_assert(offsetof(nox_player_update_data_t, mana_max) == 8,
	"player maximum-mana offset");
_Static_assert(offsetof(nox_player_update_data_t, state) == 88,
	"player state remains stable after maximum mana");
_Static_assert(sizeof(nox_player_update_data_t) == (sizeof(void*) == 4 ? 320 : 400),
	"partial player update-data size");
_Static_assert(sizeof(nox_obelisk_update_data_t) == 4,
	"Obelisk update-data size");
_Static_assert(offsetof(nox_obelisk_update_data_t, mana) == 0,
	"Obelisk mana offset");
_Static_assert(offsetof(nox_team_t, field_57) == 57,
	"team ID offset");
_Static_assert(sizeof(nox_team_t) == (sizeof(void*) == 4 ? 80 : 88),
	"native team size");

static nox_team_t fixture_team;
static uint8_t lookup_id;
static const uint32_t* contained_team;
static uint8_t contained_id;
static nox_object_t* added_target;
static int16_t added_amount;

static uint8_t object_team_id(const nox_object_t* object) {
	return (uint8_t)object->field_13;
}

static nox_team_t* find_team(uint8_t id) {
	lookup_id = id;
	return &fixture_team;
}

static int32_t team_contains(const uint32_t* object_team, uint8_t id) {
	contained_team = object_team;
	contained_id = id;
	return 1;
}

static uint16_t add_player_mana(nox_object_t* target, int16_t amount) {
	added_target = target;
	added_amount = amount;
	return UINT16_MAX;
}

static void unused_mana_transfer_reference(nox_object_t* source, nox_object_t* target) {
	if (target == NULL || ((uint8_t)target->obj_class & UINT8_C(0x04)) == 0) {
		return;
	}
	nox_player_update_data_t* player = target->data_update;
	uint16_t current = player->mana_cur;
	nox_obelisk_update_data_t* obelisk = source->data_update;
	if (current >= player->mana_max || obelisk->mana <= 0) {
		return;
	}
	if (object_team_id(source) != 0) {
		if (object_team_id(target) == 0) {
			return;
		}
		nox_team_t* team = find_team(object_team_id(target));
		if (team == NULL || !team_contains(&source->field_12, team->field_57)) {
			return;
		}
	}
	(void)add_player_mana(target, 1);
	obelisk->mana -= 1;
}

int main(void) {
	nox_obelisk_update_data_t obelisk = {.mana = 3};
	nox_player_update_data_t player = {.mana_cur = 7, .mana_prev = 6, .mana_max = 9};
	nox_object_t source = {.data_update = &obelisk};
	nox_object_t target = {.obj_class = UINT32_C(0x12340004), .data_update = &player};

	unused_mana_transfer_reference(&source, &target);
	if (obelisk.mana != 2 || added_target != &target || added_amount != 1 ||
		player.mana_cur != 7 || player.mana_prev != 6 || player.mana_max != 9) {
		return 1;
	}

	obelisk.mana = 1;
	source.field_13 = UINT32_C(4);
	target.field_13 = UINT32_C(7);
	fixture_team.field_57 = UINT8_C(9);
	unused_mana_transfer_reference(&source, &target);
	if (obelisk.mana != 0 || lookup_id != 7 || contained_team != &source.field_12 ||
		contained_id != 9) {
		return 2;
	}

	obelisk.mana = 5;
	player.mana_cur = player.mana_max;
	source.data_update = NULL;
	added_target = NULL;
	unused_mana_transfer_reference(&source, &target);
	if (obelisk.mana != 5 || added_target != NULL) {
		return 3;
	}
	return 0;
}
