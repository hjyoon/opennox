#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#undef _Static_assert
_Static_assert(sizeof(void*) == 8, "native 64-bit callback probe");
_Static_assert(sizeof(int) == 4, "original integer width");
_Static_assert(sizeof(wchar2_t) == 2, "original wide character width");
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
#include "pickup_default_4f31e0.h"
#include "decay_511660.h"

static nox_object_t owner, item, other;
static nox_player_update_data_t old_update, new_update;
static __typeof__(*old_update.player) old_player, new_player;
static uint8_t use_data[128];
static int pickup_return, replace_update, states[2], flags[2], equip_return;
static unsigned state_index, flags_index, event_count;

static void argument(uintptr_t value) {
	if (value == (uintptr_t)&owner) printf("\"owner\"");
	else if (value == (uintptr_t)&item) printf("\"item\"");
	else if (value == (uintptr_t)&other) printf("\"other-weapon\"");
	else printf("%llu", (unsigned long long)value);
}

static void event(const char* service, uintptr_t* args, unsigned count, int result) {
	assert(event_count < 64);
	printf("%s{\"service\":\"%s\",\"args\":[", event_count++ ? "," : "", service);
	for (unsigned i = 0; i < count; i++) {
		if (i) printf(",");
		argument(args[i]);
	}
	printf("],\"result\":%d}", result);
}

bool nox_common_gameFlags_check_40A5C0(unsigned int value) {
	int result = (value & 2048) != 0;
	uintptr_t args[] = {value}; event("flags", args, 1, result); return result;
}
int sub_409F40(int value) { (void)value; assert(!"unrepresented game-mode service"); return 0; }
double nox_xxx_gamedataGetFloat_419D40(char* key) { (void)key; assert(!"unrepresented limit branch"); return 0; }
int nox_float2int(float value) { (void)value; assert(!"unrepresented limit conversion"); return 0; }
int32_t nox_xxx_inventoryCountObjects_4E7D30(nox_object_t* unit, int32_t type) { (void)unit; (void)type; assert(!"unrepresented limit count"); return 0; }
int32_t sub_4E7EC0(const nox_object_t* unit, const nox_object_t* object) { (void)unit; (void)object; assert(!"unrepresented duplicate service"); return 0; }
int nox_xxx_playerClassCanUseItem_57B3D0(nox_object_t* object, int8_t cls) { (void)object; (void)cls; assert(!"unrepresented class service"); return 0; }
void nox_xxx_netPriMsgToPlayer_4DA2C0(nox_object_t* unit, const char* text, char flags) { (void)unit; (void)text; (void)flags; assert(!"unrepresented rejection message"); }
void nox_xxx_aud_501960(int sound, nox_object_t* unit, int a3, int a4) { (void)sound; (void)unit; (void)a3; (void)a4; assert(!"unrepresented rejection audio"); }
int32_t nox_xxx_pickupDefault_4F31E0(nox_object_t* unit, nox_object_t* object, int report, int equip_report) {
	assert(unit == &owner && object == &item && report == 17 && equip_report == 23);
	if (replace_update) unit->data_update = &new_update;
	uintptr_t args[] = {(uintptr_t)unit, (uintptr_t)object, (uintptr_t)report, (uintptr_t)equip_report};
	event("pickup", args, 4, pickup_return); return pickup_return;
}
int sub_419E60(nox_object_t* unit) {
	assert(unit == &owner); int result = states[state_index < 2 ? state_index : 1]; state_index++;
	uintptr_t args[] = {(uintptr_t)unit}; event("state", args, 1, result); return result;
}
int nox_xxx_weaponInventoryEquipFlags_415820(nox_object_t* object) {
	assert(object == &item); int result = flags[flags_index < 2 ? flags_index : 1]; flags_index++;
	uintptr_t args[] = {(uintptr_t)object}; event("equipFlags", args, 1, result); return result;
}
int nox_xxx_playerEquipWeapon_53A420(nox_object_t* unit, nox_object_t* object, int report, int play_sound) {
	assert(unit == &owner && object == &item && report == 23 && play_sound == 0);
	if (equip_return) ((nox_player_update_data_t*)unit->data_update)->equipped_weapon = object;
	uintptr_t args[] = {(uintptr_t)unit, (uintptr_t)object, (uintptr_t)report, (uintptr_t)play_sound};
	event("equip", args, 4, equip_return); return equip_return;
}
int nox_xxx_netReportCharges_4D82B0(int index, nox_object_t* object, char current, char maximum) {
	assert(object == &item && (uint8_t)current == 134 && (uint8_t)maximum == 243);
	uintptr_t args[] = {(uintptr_t)index, (uintptr_t)object, (uint8_t)current, (uint8_t)maximum};
	event("charges", args, 4, 0); return 0;
}
void sub_53A6C0(nox_object_t* unit, nox_object_t* object) {
	assert(unit == &owner && object == &item); uintptr_t args[] = {(uintptr_t)unit, (uintptr_t)object}; event("finalize", args, 2, 0);
}
void nox_xxx_decay_5116F0(nox_object_t* object) {
	assert(object == &item); uintptr_t args[] = {(uintptr_t)object}; event("decay", args, 1, 0);
}

// PRODUCTION_BODY_53A720

int main(int argc, char** argv) {
	assert(argc == 14);
	for (unsigned i = 0; i < 7; i++) {
		uintptr_t pointers[] = {(uintptr_t)&owner, (uintptr_t)&item, (uintptr_t)&other, (uintptr_t)&old_update, (uintptr_t)&new_update, (uintptr_t)&old_player, (uintptr_t)&new_player};
		assert(pointers[i] > UINT32_MAX);
	}
	owner.obj_class = strtoul(argv[1], NULL, 0);
	item.obj_class = strtoul(argv[2], NULL, 0); item.obj_subclass = strtoul(argv[3], NULL, 0); item.typ_ind = 0x3456;
	pickup_return = atoi(argv[4]); old_update.equipped_weapon = atoi(argv[5]) ? &other : NULL; new_update.equipped_weapon = atoi(argv[6]) ? &other : NULL;
	replace_update = atoi(argv[7]); states[0] = atoi(argv[8]); states[1] = atoi(argv[9]); flags[0] = atoi(argv[10]); flags[1] = atoi(argv[11]); equip_return = atoi(argv[12]);
	old_player.field_4 = new_player.field_4 = strtoul(argv[13], NULL, 0);
	old_player.playerInd = 3; new_player.playerInd = 7; old_update.player = &old_player; new_update.player = &new_player;
	owner.data_update = &old_update; item.use_data = use_data; use_data[108] = 134; use_data[109] = 243;
	nox_object_t original_item = item, original_owner = owner;
	printf("{\"pointer_bytes\":8,\"high_native_pointers\":true,\"events\":[");
	int result = sub_53A720(&owner, &item, 17, 23);
	original_owner.data_update = owner.data_update;
	assert(memcmp(&original_owner, &owner, sizeof(owner)) == 0 && memcmp(&original_item, &item, sizeof(item)) == 0 && use_data[108] == 134 && use_data[109] == 243);
	nox_player_update_data_t* live = owner.data_update;
	printf("],\"return_value\":%d,\"live_update\":\"%s\",\"live_equipped\":", result, live == &new_update ? "new" : "old");
	argument((uintptr_t)live->equipped_weapon);
	printf(",\"nonoutput_canaries_unchanged\":true}\n");
	return 0;
}
