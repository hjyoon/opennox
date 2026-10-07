#include <math.h>
#include <stddef.h>

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "GAME5_2.h"
#include "client__gui__guigen.h"
#include "common__binfile.h"
#include "common__crypt.h"
#include "common__net_list.h"
#include "common__random.h"
#include "operators.h"
#include "input.h"
#include "mapgen_legacy_ptr.h"

#include "client__video__draw_common.h"
#include "common__magic__speltree.h"
#include "server__script__script.h"

#include <time.h>

extern uint32_t dword_5d4594_2487656;
extern uint32_t dword_5d4594_3835368;
extern uint32_t dword_5d4594_2487804;
extern uint32_t dword_5d4594_2487632;
extern uint32_t dword_5d4594_3835364;
extern uint32_t dword_5d4594_2487884;
extern uint32_t dword_5d4594_3835372;
extern uint32_t dword_5d4594_3835392;
extern uint32_t dword_5d4594_2487580;
extern uint32_t dword_5d4594_2487568;
extern uint32_t dword_5d4594_2487536;
extern uint32_t dword_5d4594_2487628;
extern uint32_t dword_5d4594_2487584;
extern uint32_t dword_5d4594_2487564;
extern uint32_t dword_5d4594_2487672;
extern uint32_t dword_5d4594_2487676;
extern uint32_t dword_5d4594_3835388;
extern uint32_t dword_5d4594_2487652;
extern uint32_t dword_5d4594_3835348;
extern uint32_t dword_5d4594_2487576;
extern uint32_t dword_5d4594_2487624;
extern uint32_t dword_5d4594_3835352;
extern uint32_t dword_5d4594_2487620;
extern uint32_t dword_5d4594_2487708;
extern uint32_t nox_xxx_energyBoltTarget_5d4594_2487880;
extern uint32_t dword_5d4594_2487532;
extern uint32_t dword_5d4594_2487248;

extern uint32_t dword_5d4594_2487560;
extern uint32_t dword_5d4594_2487540;
extern uint32_t dword_5d4594_2487712;
extern uint32_t dword_5d4594_2487524;
extern uint32_t dword_5d4594_2487556;
extern obj_5D4594_2650668_t** ptr_5D4594_2650668;

static uint8_t** nox_mapgen_occupancy_grid;
static uint8_t* nox_mapgen_room_head;
static uint8_t* nox_mapgen_farthest_room;
static uint8_t* nox_mapgen_sorted_room_head;
static float nox_mapgen_max_room_distance;

static void nox_mapgenSetMaxRoomDistanceNative_5259D0(float value) {
	union {
		float value;
		uint32_t bits;
	} encoded = {value};
	nox_mapgen_max_room_distance = value;
	dword_5d4594_2487580 = encoded.bits;
}

static void nox_mapgenSetFarthestRoomNative_5259E0(uint8_t* room) {
	nox_mapgen_farthest_room = room;
	dword_5d4594_2487576 = nox_mapgenLegacyPtrRegister(room);
}

static void nox_mapgenSetSortedRoomHeadNative_525C90(uint8_t* room) {
	nox_mapgen_sorted_room_head = room;
	dword_5d4594_2487584 = nox_mapgenLegacyPtrRegister(room);
}

static void nox_mapgenResetRoomRankingNative(void) {
	nox_mapgenSetFarthestRoomNative_5259E0(NULL);
	nox_mapgenSetSortedRoomHeadNative_525C90(NULL);
	nox_mapgenSetMaxRoomDistanceNative_5259D0(0.0f);
}

// AreaMap keeps these records in GAME.EXE's fixed PE32 layout. Pointer-like
// fields are registry tokens, not truncated host pointers.
typedef struct nox_mapgen_foreach_legacy {
	uint32_t type_id;
	uint32_t choices_token;
	uint32_t next_token;
} nox_mapgen_foreach_legacy;

typedef struct nox_mapgen_item_set_entry_legacy {
	char lookup_name[60];
	char object_name[60];
	uint32_t modifier_tokens[4];
	uint32_t modifier_counts[4];
	uint32_t next_token;
} nox_mapgen_item_set_entry_legacy;

typedef struct nox_mapgen_occupied_rect_legacy {
	uint32_t transient;
	float min_x;
	float min_y;
	float max_x;
	float max_y;
	int32_t object_index;
	uint32_t next_token;
} nox_mapgen_occupied_rect_legacy;

typedef struct nox_mapgen_decor_fill_legacy {
	uint32_t type;
	char object_name[60];
	char tile_name[60];
	uint32_t next_token;
} nox_mapgen_decor_fill_legacy;

typedef struct nox_mapgen_wall_floor_legacy {
	char wall_name[60];
	char floor_name[60];
	uint32_t fills_token;
	uint32_t next_token;
} nox_mapgen_wall_floor_legacy;

typedef struct nox_mapgen_decor_entry_legacy {
	uint32_t type;
	char name[60];
	uint32_t density_enabled;
	int32_t min_count;
	int32_t max_count;
	float density;
	uint32_t choices_token;
	uint32_t foreach_token;
	uint32_t next_token;
	uint32_t shuffle_prev_token;
	uint32_t shuffle_next_token;
} nox_mapgen_decor_entry_legacy;

typedef struct nox_mapgen_decor_set_legacy {
	int32_t min_count;
	int32_t max_count;
	uint32_t entries_token;
	uint32_t entry_count;
	uint32_t shares_entries;
	uint32_t next_token;
} nox_mapgen_decor_set_legacy;

typedef struct nox_mapgen_decor_definition_legacy {
	char name[60];
	uint32_t room_type;
	uint8_t constraint_mask;
	uint8_t occur_limit;
	uint8_t exhausted;
	uint8_t must_occur;
	uint8_t occurred;
	uint8_t reserved[3];
	uint32_t frequency;
	int32_t min_room_size;
	int32_t max_room_size;
	uint32_t wall_floors_token;
	uint32_t wall_floor_count;
	uint32_t decor_sets_token;
	uint32_t decor_set_count;
	char door_name[60];
	char double_door_name[60];
	uint32_t next_token;
} nox_mapgen_decor_definition_legacy;

enum {
	NOX_MAPGEN_CHOICE_COUNT_INDEX = 513,
	NOX_MAPGEN_CHOICE_NEXT_INDEX = 514,
};

_Static_assert(sizeof(nox_mapgen_foreach_legacy) == 12, "wrong mapgen FOREACH record size");
_Static_assert(sizeof(nox_mapgen_item_set_entry_legacy) == 156, "wrong mapgen item-set record size");
_Static_assert(offsetof(nox_mapgen_item_set_entry_legacy, modifier_tokens) == 120,
			   "wrong mapgen item-set modifier-token offset");
_Static_assert(offsetof(nox_mapgen_item_set_entry_legacy, modifier_counts) == 136,
			   "wrong mapgen item-set modifier-count offset");
_Static_assert(offsetof(nox_mapgen_item_set_entry_legacy, next_token) == 152,
			   "wrong mapgen item-set next-token offset");
_Static_assert(sizeof(nox_mapgen_occupied_rect_legacy) == 28,
			   "wrong mapgen occupied-rectangle record size");
_Static_assert(offsetof(nox_mapgen_occupied_rect_legacy, next_token) == 24,
			   "wrong mapgen occupied-rectangle next-token offset");
_Static_assert(sizeof(nox_mapgen_decor_fill_legacy) == 128,
			   "wrong mapgen decor-fill record size");
_Static_assert(offsetof(nox_mapgen_decor_fill_legacy, next_token) == 124,
			   "wrong mapgen decor-fill next-token offset");
_Static_assert(sizeof(nox_mapgen_wall_floor_legacy) == 128,
			   "wrong mapgen wall/floor record size");
_Static_assert(offsetof(nox_mapgen_wall_floor_legacy, fills_token) == 120,
			   "wrong mapgen wall/floor fill-token offset");
_Static_assert(offsetof(nox_mapgen_wall_floor_legacy, next_token) == 124,
			   "wrong mapgen wall/floor next-token offset");
_Static_assert(sizeof(nox_mapgen_decor_entry_legacy) == 100,
			   "wrong mapgen decor-entry record size");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, choices_token) == 80,
			   "wrong mapgen decor-entry choices-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, foreach_token) == 84,
			   "wrong mapgen decor-entry FOREACH-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, next_token) == 88,
			   "wrong mapgen decor-entry next-token offset");
_Static_assert(sizeof(nox_mapgen_decor_set_legacy) == 24,
			   "wrong mapgen decor-set record size");
_Static_assert(offsetof(nox_mapgen_decor_set_legacy, next_token) == 20,
			   "wrong mapgen decor-set next-token offset");
_Static_assert(sizeof(nox_mapgen_decor_definition_legacy) == 224,
			   "wrong mapgen decor-definition record size");
_Static_assert(offsetof(nox_mapgen_decor_definition_legacy, wall_floors_token) == 84,
			   "wrong mapgen decor-definition wall/floor-token offset");
_Static_assert(offsetof(nox_mapgen_decor_definition_legacy, decor_sets_token) == 92,
			   "wrong mapgen decor-definition decor-set-token offset");
_Static_assert(offsetof(nox_mapgen_decor_definition_legacy, next_token) == 220,
			   "wrong mapgen decor-definition next-token offset");

static nox_mapgen_item_set_entry_legacy* nox_mapgenItemSetEntryResolve(uintptr_t ref) {
	return (nox_mapgen_item_set_entry_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static nox_mapgen_occupied_rect_legacy* nox_mapgenOccupiedRectResolve(uintptr_t ref) {
	return (nox_mapgen_occupied_rect_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static nox_mapgen_decor_fill_legacy* nox_mapgenDecorFillResolve(uintptr_t ref) {
	return (nox_mapgen_decor_fill_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static nox_mapgen_wall_floor_legacy* nox_mapgenWallFloorResolve(uintptr_t ref) {
	return (nox_mapgen_wall_floor_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static nox_mapgen_decor_entry_legacy* nox_mapgenDecorEntryResolve(uintptr_t ref) {
	return (nox_mapgen_decor_entry_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static nox_mapgen_decor_set_legacy* nox_mapgenDecorSetResolve(uintptr_t ref) {
	return (nox_mapgen_decor_set_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static nox_mapgen_decor_definition_legacy* nox_mapgenDecorDefinitionResolve(uintptr_t ref) {
	return (nox_mapgen_decor_definition_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static uint32_t* nox_mapgenChoiceResolve(uintptr_t ref) {
	return (uint32_t*)nox_mapgenLegacyPtrResolve(ref);
}

uint32_t* nox_mapgenChoiceNextNative_520380(uintptr_t choice_ref) {
	uint32_t* choice = nox_mapgenChoiceResolve(choice_ref);
	return choice ? nox_mapgenChoiceResolve(choice[NOX_MAPGEN_CHOICE_NEXT_INDEX]) : NULL;
}

static nox_mapgen_foreach_legacy* nox_mapgenForeachResolve(uintptr_t ref) {
	return (nox_mapgen_foreach_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

static void nox_mapgenFreeChoiceList(uintptr_t ref) {
	uint32_t* choice = nox_mapgenChoiceResolve(ref);
	while (choice) {
		uint32_t* next = nox_mapgenChoiceNextNative_520380((uintptr_t)choice);
		nox_mapgenLegacyPtrForget(choice);
		free(choice);
		choice = next;
	}
}

static void nox_mapgenFreeForeachList(uintptr_t ref) {
	nox_mapgen_foreach_legacy* entry = nox_mapgenForeachResolve(ref);
	while (entry) {
		nox_mapgen_foreach_legacy* next = nox_mapgenForeachResolve(entry->next_token);
		nox_mapgenFreeChoiceList(entry->choices_token);
		nox_mapgenLegacyPtrForget(entry);
		free(entry);
		entry = next;
	}
}

static void nox_mapgenFreeDecorFillList(uintptr_t ref) {
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve(ref);
	while (fill) {
		nox_mapgen_decor_fill_legacy* next = nox_mapgenDecorFillResolve(fill->next_token);
		nox_mapgenLegacyPtrForget(fill);
		free(fill);
		fill = next;
	}
}

static void nox_mapgenFreeWallFloorList(uintptr_t ref) {
	nox_mapgen_wall_floor_legacy* wall_floor = nox_mapgenWallFloorResolve(ref);
	while (wall_floor) {
		nox_mapgen_wall_floor_legacy* next = nox_mapgenWallFloorResolve(wall_floor->next_token);
		nox_mapgenFreeDecorFillList(wall_floor->fills_token);
		nox_mapgenLegacyPtrForget(wall_floor);
		free(wall_floor);
		wall_floor = next;
	}
}

static void nox_mapgenFreeDecorEntryList(uintptr_t ref) {
	nox_mapgen_decor_entry_legacy* entry = nox_mapgenDecorEntryResolve(ref);
	while (entry) {
		nox_mapgen_decor_entry_legacy* next = nox_mapgenDecorEntryResolve(entry->next_token);
		nox_mapgenFreeChoiceList(entry->choices_token);
		nox_mapgenFreeForeachList(entry->foreach_token);
		nox_mapgenLegacyPtrForget(entry);
		free(entry);
		entry = next;
	}
}

static void nox_mapgenFreeDecorSetList(uintptr_t ref) {
	nox_mapgen_decor_set_legacy* set = nox_mapgenDecorSetResolve(ref);
	while (set) {
		nox_mapgen_decor_set_legacy* next = nox_mapgenDecorSetResolve(set->next_token);
		if (!set->shares_entries) {
			nox_mapgenFreeDecorEntryList(set->entries_token);
		}
		nox_mapgenLegacyPtrForget(set);
		free(set);
		set = next;
	}
}

static void nox_mapgenFreeDecorDefinitionList(uintptr_t ref) {
	nox_mapgen_decor_definition_legacy* definition = nox_mapgenDecorDefinitionResolve(ref);
	while (definition) {
		nox_mapgen_decor_definition_legacy* next =
			nox_mapgenDecorDefinitionResolve(definition->next_token);
		nox_mapgenFreeWallFloorList(definition->wall_floors_token);
		nox_mapgenFreeDecorSetList(definition->decor_sets_token);
		nox_mapgenLegacyPtrForget(definition);
		free(definition);
		definition = next;
	}
}

static void nox_mapgenApplyChoiceNative_521FE0(
	uint8_t* theme, nox_object_t* holder, uintptr_t choices_ref);
static nox_object_t* nox_mapgenMakeSpellbookNative_5220E0(uint8_t* theme, const char* spell_name);
static nox_object_t* nox_mapgenMakeEnchantedItemNative_5221A0(
	char* name, uintptr_t list_ref, int count);
static nox_object_t* nox_mapgenAttachInventoryNative_522300(
	nox_object_t* holder, nox_object_t* item);
static int nox_mapgenFinishSpellbookNative_527DB0(nox_object_t* object, char spell_id);
static nox_mapgen_wall_floor_legacy* nox_mapgenReadWallFloorNative_51FE00(
	nox_mapgen_decor_definition_legacy* definition, FILE* file);
static int nox_mapgenReadDecorFillNative_51FEC0(
	nox_mapgen_wall_floor_legacy* wall_floor, int type, FILE* file);
static int nox_mapgenReadDecorSetNative_51FFA0(
	nox_mapgen_decor_definition_legacy* definition, FILE* file);
static int nox_mapgenReadDecorCopyNative_520660(
	uint8_t* theme, nox_mapgen_decor_definition_legacy* destination, FILE* file);

float get_nox_xxx_warriorMaxHealth_587000_312784();
float get_nox_xxx_wizardMaxHealth_587000_312816();
float get_nox_xxx_conjurerMaxHealth_587000_312800();

float get_nox_xxx_warriorMaxMana_587000_312788();
float get_nox_xxx_wizardMaximumMana_587000_312820();
float get_nox_xxx_conjurerMaxMana_587000_312804();

//----- (0051DA70) --------------------------------------------------------
extern nox_tileDef_t nox_tile_defs_arr[176];
int sub_51DA70(int a1, int a2, int a3, int a4, int a5) {
	int v5;  // ebx
	int v6;  // ebp
	int* v7; // esi
	int v8;  // edi
	int v9;  // ecx
	int v10; // eax
	int v11; // eax
	int v12; // edx
	int v13; // ecx
	int v15; // ebx
	int v16; // edi
	int v17; // edx
	int v18; // eax
	int v19; // edx
	int v20; // esi
	int v22; // ebp
	int v23; // edx
	int v24; // edx
	int v25; // esi
	int v26; // ebp
	int v27; // edx
	int v28; // [esp+10h] [ebp-8h]
	int v29; // [esp+14h] [ebp-4h]
	int v30; // [esp+24h] [ebp+Ch]

	v5 = a4;
	v6 = a2;
	v7 = (int*)a3;
	v8 = a1;
	if (!a3) {
		v9 = 255;
		v10 = 0;
		goto LABEL_19;
	}
	if (!*(uint32_t*)(a3 + 8)) {
		v11 = a1;
		v12 = a2;
		v28 = a1;
		v13 = a4 != 1;
		v30 = v13;
		goto LABEL_11;
	}
	if (!*(uint8_t*)(a3 + 20)) {
		v9 = *(uint32_t*)a3;
		v10 = 0;
		*(uint32_t*)(a3 + 12) = a1;
		*(uint32_t*)(a3 + 16) = a2;
		*(uint8_t*)(a3 + 20) = a4;
		goto LABEL_19;
	}
	v11 = a1 - *(uint32_t*)(a3 + 12);
	v12 = a2 - *(uint32_t*)(a3 + 16);
	v28 = a1 - *(uint32_t*)(a3 + 12);
	a2 -= *(uint32_t*)(a3 + 16);
	if (*(unsigned char*)(a3 + 20) != a4) {
		v13 = *(uint8_t*)(a3 + 20) != 1 ? -1 : 1;
		v30 = v13;
		goto LABEL_11;
	}
	v30 = 0;
LABEL_11:
	v9 = *v7;
	v15 = nox_tile_defs_arr[v9].field_52;
	v29 = (v12 + v11) % v15;
	v16 = nox_tile_defs_arr[v9].field_53;
	v17 = (v30 + a2 - v28) % v16;
	v18 = v29;
	if (v29 < 0) {
		v18 = v15 + v29;
	}
	if (v17 < 0) {
		v17 += v16;
	}
	if (*getMemU32Ptr(0x973F18, 35916) == 1) {
		v10 = dword_5d4594_3835348;
	} else {
		v10 = v18 + v17 * v15;
	}
	v8 = a1;
	v5 = a4;
LABEL_19:
	if (v5 == 2) {
		if (dword_5d4594_3835352 == 1) {
			if (v7) {
				v19 = v7[6];
				v20 = v7[7];
			} else {
				v19 = 255;
				v20 = 0;
			}
			return nox_xxx_tile_543C50((uint32_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + 44 * v6 + 24), v9, v10, v19,
									   v20, a5);
		}
		v22 = 44 * v6;
		v23 = v22 + (uint32_t)(ptr_5D4594_2650668[v8]);
		if (*(uint32_t*)(v23 + 24) != v9 || *(uint32_t*)(v23 + 28) != v10) {
			*(uint32_t*)(v23 + 24) = v9;
			*(uint32_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + v22 + 28) = v10;
			if (v9 == 255) {
				nox_xxx_tileFreeTile_422200(&ptr_5D4594_2650668[v8][v6].field_10);
			}
			if (v7) {
				*(uint8_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + v22) |= 2u;
			} else {
				*(uint8_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + v22) =
					*(uint8_t*)(v22 + (uint32_t)(ptr_5D4594_2650668[v8])) & 0xFD;
			}
			return 1;
		}
		return 0;
	}
	if (v5 != 1) {
		return 0;
	}
	if (dword_5d4594_3835352 == 1) {
		if (v7) {
			v24 = v7[6];
			v25 = v7[7];
		} else {
			v24 = 255;
			v25 = 0;
		}
		return nox_xxx_tile_543C50((uint32_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + 44 * v6 + 4), v9, v10, v24, v25,
								   a5);
	}
	v26 = 44 * v6;
	v27 = v26 + (uint32_t)(ptr_5D4594_2650668[v8]);
	if (*(uint32_t*)(v27 + 4) == v9 && *(uint32_t*)(v27 + 8) == v10) {
		return 0;
	}
	*(uint32_t*)(v27 + 4) = v9;
	*(uint32_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + v26 + 8) = v10;
	if (v9 == 255) {
		nox_xxx_tileFreeTile_422200(&ptr_5D4594_2650668[v8][v6].field_5);
	}
	if (v7) {
		*(uint8_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + v26) |= 1u;
	} else {
		*(uint8_t*)((uint32_t)(ptr_5D4594_2650668[v8]) + v26) =
			*(uint8_t*)(v26 + (uint32_t)(ptr_5D4594_2650668[v8])) & 0xFE;
	}
	return 1;
}

//----- (0051DD50) --------------------------------------------------------
void sub_51DD50(int a1, int a2, int a3, int a4) {
	int v4;            // edx
	int v5;            // edi
	unsigned char* v6; // eax
	int v7;            // eax

	if (a1 > 0 && a1 < 127 && a2 > 0 && a2 < 127) {
		if ((v4 = a3, a3 & 2) && (v4 = a3, a4 == *(uint32_t*)((uint32_t)(ptr_5D4594_2650668[a1]) + 44 * a2 + 24)) ||
			v4 & 1 && a4 == *(uint32_t*)((uint32_t)(ptr_5D4594_2650668[a1]) + 44 * a2 + 4)) {
			if ((!(v4 & 1) || a2 != 1) && (!(a3 & 2) || a1 != 1)) {
				v5 = 0;
				if (*(int*)&dword_5d4594_2487248 > 0) {
					v6 = getMemAt(0x973F18, 16204);
					while (1) {
						if (!(*((uint32_t*)v6 - 1) != a1 || *(uint32_t*)v6 != a2 || *((uint32_t*)v6 + 1) != v4)) {
							return;
						}
						++v5;
						v6 += 12;
						if (v5 >= *(int*)&dword_5d4594_2487248) {
							break;
						}
					}
				}
				if (dword_5d4594_2487248 >= 500) {
					*getMemU32Ptr(0x973F18, 22200) = 1;
				} else {
					v7 = 12 * (dword_5d4594_2487248)++;
					*getMemU32Ptr(0x973F18, 16200 + v7) = a1;
					*getMemU32Ptr(0x973F18, 16204 + v7) = a2;
					*getMemU32Ptr(0x973F18, 16208 + v7) = v4;
				}
			}
		}
	}
}

//----- (0051DE30) --------------------------------------------------------
int sub_51DE30(uint32_t* a1, uint32_t* a2, uint32_t* a3) {
	int v3;     // eax
	int result; // eax

	if (*(int*)&dword_5d4594_2487248 <= 0) {
		return 0;
	}
	v3 = dword_5d4594_2487248 - 1;
	dword_5d4594_2487248 = v3;
	*a1 = *getMemU32Ptr(0x973F18, 16200 + 12 * v3);
	*a2 = *getMemU32Ptr(0x973F18, 16204 + 12 * dword_5d4594_2487248);
	result = 1;
	*a3 = *getMemU32Ptr(0x973F18, 16208 + 12 * dword_5d4594_2487248);
	return result;
}

// 0051DEA0 is implemented by the native-pointer Go callback exported from
// wall.go. The original return value was ignored by the sole foreach caller.

// 0051DED0 is implemented by the native-pointer Go ObjectData writer.

//----- (0051E010) --------------------------------------------------------
int nox_xxx_nxzCompressFile_57BDD0(char* a1, char* a2);
int nox_xxx_mapSaveMap_51E010(char* a1, int a2) {
	char* v2;         // edi
	unsigned char v3; // cl
	int result;       // eax
	int v5;           // esi
	int v7;           // [esp+10h] [ebp-804h]
	char v8[1024];    // [esp+14h] [ebp-800h]
	char Mem[1024];   // [esp+414h] [ebp-400h]

	v7 = -86050098;
	strcpy(Mem, a1);
	v8[0] = 0;
	strncat(v8, a1, 1024-1);
	v8[strlen(v8)-4] = 0;
	v2 = &v8[strlen(v8) + 1];
	v3 = getMemByte(0x587000, 253116);
	*(uint32_t*)--v2 = *getMemU32Ptr(0x587000, 253112);
	v2[4] = v3;
	result = nox_xxx_cryptOpen_426910(Mem, 0, 19);
	if (result) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v7, 4u);
		v5 = nox_xxx_cryptFlush_4268E0();
		if (nox_xxx_map_51E140()) {
			sub_4268F0(v5);
			nox_xxx_cryptClose_4269F0();
			if (!a2 || (result = nox_xxx_nxzCompressFile_57BDD0(Mem, (int)v8)) != 0) {
				result = 1;
			}
		} else {
			nox_xxx_cryptClose_4269F0();
			result = 0;
		}
	}
	return result;
}
// 51E0D5: variable 'v6' is possibly undefined

//----- (0051E140) --------------------------------------------------------
void nox_xxx_map_5004F0();
void nox_xxx_mapSetWallInGlobalDir0pr1_5004D0();
int nox_xxx_map_51E140() {
	int result; // eax
	char v2;    // [esp+1h] [ebp-1h]

	*getMemU32Ptr(0x5D4594, 2487252) = 256;
	*getMemU32Ptr(0x5D4594, 2487256) = 256;
	nox_xxx_wallForeachFn_410640(nox_xxx_mapCountWallsMB_51DEA0, 0);
	nox_xxx_fileReadWrite_426AC0_file3_fread(getMemAt(0x5D4594, 2487252), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(getMemAt(0x5D4594, 2487256), 4u);
	nox_xxx_mapWall_426A80(getMemIntPtr(0x5D4594, 2487252));
	nox_xxx_mapSetWallInGlobalDir0pr1_5004D0();
	if (nox_xxx_mapWriteSectionsMB_426E20(0)) {
		nox_xxx_map_5004F0();
		v2 = 0;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v2, 1u);
		result = 1;
	} else {
		nox_xxx_cryptClose_4269F0();
		result = 0;
	}
	return result;
}

//----- (0051E1D0) --------------------------------------------------------
int nox_xxx_mapGenSpellIdByName_51E1D0(const char* a1) {
	unsigned int i; // esi
	char v3[60];    // [esp+8h] [ebp-78h]
	char v4;        // [esp+44h] [ebp-3Ch]

	strcpy(v3, a1);
	for (i = 0; i < strlen(v3); ++i) {
		v3[i] = toupper(v3[i]);
	}
	nox_sprintf(&v4, "SPELL_%s", v3);
	return nox_xxx_spellNameToN_4243F0(&v4);
}
// 51E1D0: using guessed type char var_78[60];

//----- (0051E260) --------------------------------------------------------
int nox_xxx_mapGenReadTheme_51E260(int* a1, int a2) {
	FILE* v2;     // eax
	FILE* v3;     // edi
	int v4;       // eax
	int v5;       // ebx
	char v7[256]; // [esp+18h] [ebp-100h]

	memset(a1, 0, 0x45Cu);
	if (!nox_mapgenLegacyPtrRegister(a1)) {
		return 0;
	}
	nox_sprintf(v7, "mapgen/%s.thm", a2);
	a1[3] = 3;
	a1[5] = 3;
	a1[1] = 5;
	a1[2] = 2;
	a1[4] = 1;
	a1[6] = 40;
	a1[7] = 20;
	a1[8] = 10;
	a1[9] = 5;
	a1[10] = 20;
	a1[16] = 1157234688;
	a1[18] = 100;
	a1[19] = time(0);
	a1[14] = 1;
	a1[22] = 0;
	a1[11] = 0;
	a1[12] = 100;
	a1[13] = 100;
	a1[30] = 0;
	a1[38] = 0;
	a1[46] = 0;
	*getMemU32Ptr(0x5D4594, 2487520) = 0;
	v2 = nox_binfile_open_408CC0(v7, 0);
	v3 = v2;
	if (!v2 || !nox_binfile_cryptSet_408D40((int)v2, 1)) {
		return 0;
	}
	while (nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264))) {
		if (!nox_strcmpi("ALGORITHM_DATA", (const char*)getMemAt(0x5D4594, 2487264))) {
			v4 = nox_xxx_genReadAlgData_51EBB0((int)a1, v3);
			goto LABEL_23;
		}
		if (!nox_strcmpi("EXIT", (const char*)getMemAt(0x5D4594, 2487264))) {
			v4 = nox_xxx_genReadExit_51F800((int)a1, v3);
			goto LABEL_23;
		}
		if (!nox_strcmpi("PREFAB", (const char*)getMemAt(0x5D4594, 2487264))) {
			v4 = nox_xxx_genReadPrefab_520BF0((uint8_t*)a1, v3);
			goto LABEL_23;
		}
		if (!nox_strcmpi("AMBIENT_LIGHT", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (!nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264)) ||
				(a1[134] = atoi((const char*)getMemAt(0x5D4594, 2487264)),
				 !nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264))) ||
				(a1[135] = atoi((const char*)getMemAt(0x5D4594, 2487264)),
				 !nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264)))) {
				v5 = 0;
				goto LABEL_27;
			}
			a1[136] = atoi((const char*)getMemAt(0x5D4594, 2487264));
			continue;
		}
		if (nox_strcmpi("SPELL_SET", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("ARMOR_SET", (const char*)getMemAt(0x5D4594, 2487264))) {
				if (nox_strcmpi("WEAPON_SET", (const char*)getMemAt(0x5D4594, 2487264))) {
					if (nox_strcmpi("DECOR", (const char*)getMemAt(0x5D4594, 2487264))) {
						v5 = 0;
						goto LABEL_27;
					}
					v4 = nox_xxx_genReadDecor_51F9F0(a1, v3);
				} else {
					v4 = nox_xxx_genReadWeaponSet_51F030((uint8_t*)a1, v3);
				}
			} else {
				v4 = nox_xxx_genReadArmorSet_51F640((uint8_t*)a1, v3);
			}
		} else {
			v4 = nox_xxx_genReadSpellSet_51EFB0((int)a1, v3);
		}
	LABEL_23:
		v5 = v4;
		if (!v4) {
			goto LABEL_27;
		}
	}
	v5 = 1;
LABEL_27:
	nox_binfile_close_408D90(v3);
	if (v5 != 1 || nox_xxx_mapgenCheckSettings_520AD0(a1 + 22) && nox_xxx_mapgenCheckSettings_520AD0(a1 + 30) &&
					   (!a1[46] || nox_xxx_mapgenCheckSettings_520AD0(a1 + 46))) {
		return v5;
	}
	return 0;
}

//----- (0051E540) --------------------------------------------------------
int nox_xxx_mapGenReadLine_51E540(FILE* a1, uint8_t* a2) {
	int result; // eax

	result = sub_51E570(a1, a2);
	if (result) {
		result = sub_51E670(a1) != 0;
	}
	return result;
}

//----- (0051E570) --------------------------------------------------------
int sub_51E570(FILE* a1, uint8_t* a2) {
	uint8_t* v2;          // ebx
	int v3;               // ecx
	int v4;               // edi
	int v5;               // ebp
	int v6;               // eax
	uint16_t CharType[2]; // [esp+10h] [ebp-4h]

	v2 = a2;
	v3 = 0;
	*(uint32_t*)CharType = 0;
	v4 = 1;
	do {
		while (1) {
			v5 = v3;
			nox_binfile_fread_408E40((char*)CharType, 1, 1, a1);
			if (nox_binfile_lastErr_409370(a1) == -1) {
				return 0;
			}
			v3 = *(uint32_t*)CharType;
			if (*(uint32_t*)CharType == 10) {
				++*getMemU32Ptr(0x5D4594, 2487520);
			}
			v6 = isspace(CharType[0]);
			if (v6) {
				break;
			}
			v4 = 0;
			if (v3 != 47 || v5 != 47) {
				*v2++ = v3;
			} else {
				sub_51E630(a1);
				v2 = a2;
				v3 = *(uint32_t*)CharType;
				v4 = 1;
			}
		}
	} while (v4);
	*v2 = 0;
	return 1;
}

//----- (0051E630) --------------------------------------------------------
int sub_51E630(FILE* a1) {
	FILE* v1;   // esi
	int result; // eax

	v1 = a1;
	while (1) {
		LOBYTE(a1) = 0;
		nox_binfile_fread_408E40((char*)&a1, 1, 1, v1);
		result = nox_binfile_lastErr_409370(v1);
		if (result == -1) {
			break;
		}
		if ((uint8_t)a1 == 10) {
			++*getMemU32Ptr(0x5D4594, 2487520);
			return result;
		}
	}
	return result;
}

//----- (0051E670) --------------------------------------------------------
int sub_51E670(FILE* a1) {
	int result; // eax
	int v2;     // [esp+4h] [ebp-4h]

	if (nox_strcmpi("IF", (const char*)getMemAt(0x5D4594, 2487264))) {
		if (nox_strcmpi("ELSE", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("ENDIF", (const char*)getMemAt(0x5D4594, 2487264))) {
				return 1;
			}
		} else {
			result = sub_51E720(a1);
			if (!result) {
				return result;
			}
		}
		return nox_xxx_mapGenReadLine_51E540(a1, getMemAt(0x5D4594, 2487264)) != 0;
	}
	result = sub_51E800((int)a1, &v2);
	if (result) {
		if (v2) {
			return nox_xxx_mapGenReadLine_51E540(a1, getMemAt(0x5D4594, 2487264)) != 0;
		}
		result = sub_51E780(a1);
		if (result) {
			return nox_xxx_mapGenReadLine_51E540(a1, getMemAt(0x5D4594, 2487264)) != 0;
		}
	}
	return result;
}

//----- (0051E720) --------------------------------------------------------
int sub_51E720(FILE* a1) {
	do {
		while (1) {
			if (!sub_51E570(a1, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			if (nox_strcmpi("IF", (const char*)getMemAt(0x5D4594, 2487264))) {
				break;
			}
			if (!sub_51E720(a1)) {
				return 0;
			}
		}
	} while (nox_strcmpi("ENDIF", (const char*)getMemAt(0x5D4594, 2487264)));
	return 1;
}

//----- (0051E780) --------------------------------------------------------
int sub_51E780(FILE* a1) {
	do {
		while (1) {
			if (!sub_51E570(a1, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			if (nox_strcmpi("IF", (const char*)getMemAt(0x5D4594, 2487264))) {
				break;
			}
			if (!sub_51E720(a1)) {
				return 0;
			}
		}
	} while (nox_strcmpi("ENDIF", (const char*)getMemAt(0x5D4594, 2487264)) &&
			 nox_strcmpi("ELSE", (const char*)getMemAt(0x5D4594, 2487264)));
	return 1;
}

//----- (0051E800) --------------------------------------------------------
int sub_51E800(int a1, uint32_t* a2) {
	FILE* v2;     // edi
	uint32_t* v3; // esi
	char* v4;     // eax
	int result;   // eax
	char* v6;     // eax
	char* v7;     // eax
	int v8;       // eax
	int v9;       // esi
	int v10;      // eax
	int v11;      // esi

	v2 = (FILE*)a1;
	if (!nox_xxx_mapGenReadLine_51E540((FILE*)a1, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	if (!nox_strcmpi("WIZARD", (const char*)getMemAt(0x5D4594, 2487264))) {
		v3 = a2;
		*a2 = 0;
		v4 = nox_common_playerInfoGetFirst_416EA0();
		if (v4) {
			while (v4[2251] != 1) {
				v4 = nox_common_playerInfoGetNext_416EE0((int)v4);
				if (!v4) {
					return 1;
				}
			}
			*v3 = 1;
			return 1;
		}
		return 1;
	}
	if (!nox_strcmpi("CONJURER", (const char*)getMemAt(0x5D4594, 2487264))) {
		v3 = a2;
		*a2 = 0;
		v6 = nox_common_playerInfoGetFirst_416EA0();
		if (v6) {
			while (v6[2251] != 2) {
				v6 = nox_common_playerInfoGetNext_416EE0((int)v6);
				if (!v6) {
					return 1;
				}
			}
			*v3 = 1;
			return 1;
		}
		return 1;
	}
	if (!nox_strcmpi("WARRIOR", (const char*)getMemAt(0x5D4594, 2487264))) {
		v3 = a2;
		*a2 = 0;
		v7 = nox_common_playerInfoGetFirst_416EA0();
		if (v7) {
			while (v7[2251]) {
				v7 = nox_common_playerInfoGetNext_416EE0((int)v7);
				if (!v7) {
					return 1;
				}
			}
			*v3 = 1;
			return 1;
		}
		return 1;
	}
	if (!nox_strcmpi("EXPERIENCE_LEVEL", (const char*)getMemAt(0x5D4594, 2487264))) {
		if (sub_51EAF0((int)v2, &a1) && nox_xxx_mapGenReadLine_51E540(v2, getMemAt(0x5D4594, 2487264))) {
			v8 = atoi((const char*)getMemAt(0x5D4594, 2487264));
			switch (a1) {
			case 0:
				*a2 = v8 > 3;
				result = 1;
				break;
			case 1:
				*a2 = v8 < 3;
				result = 1;
				break;
			case 2:
				*a2 = v8 == 3;
				result = 1;
				break;
			case 3:
				*a2 = v8 != 3;
				result = 1;
				break;
			default:
				return 1;
			}
			return result;
		}
		return 0;
	}
	if (nox_strcmpi("NUMPLAYERS", (const char*)getMemAt(0x5D4594, 2487264))) {
		if (strchr((const char*)getMemAt(0x5D4594, 2487264), 37)) {
			v11 = atoi((const char*)getMemAt(0x5D4594, 2487264));
			*a2 = nox_xxx_mapGenRandFunc_526AC0(1, 100) <= v11;
			return 1;
		}
		return 0;
	}
	v9 = nox_common_playerInfoCount_416F40();
	if (!sub_51EAF0((int)v2, &a1) || !nox_xxx_mapGenReadLine_51E540(v2, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	v10 = atoi((const char*)getMemAt(0x5D4594, 2487264));
	switch (a1) {
	case 0:
		*a2 = v9 < v10;
		result = 1;
		break;
	case 1:
		*a2 = v9 > v10;
		result = 1;
		break;
	case 2:
		*a2 = v9 == v10;
		result = 1;
		break;
	case 3:
		*a2 = v9 != v10;
		result = 1;
		break;
	default:
		return 1;
	}
	return result;
}

//----- (0051EAF0) --------------------------------------------------------
int sub_51EAF0(int a1, uint32_t* a2) {
	int result; // eax

	if (!nox_xxx_mapGenReadLine_51E540((FILE*)a1, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	if (!nox_strcmpi("<", (const char*)getMemAt(0x5D4594, 2487264))) {
		*a2 = 0;
		return 1;
	}
	if (!nox_strcmpi(">", (const char*)getMemAt(0x5D4594, 2487264))) {
		result = 1;
		*a2 = 1;
		return result;
	}
	if (!nox_strcmpi("==", (const char*)getMemAt(0x5D4594, 2487264))) {
		*a2 = 2;
		return 1;
	}
	if (nox_strcmpi("!=", (const char*)getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	result = 1;
	*a2 = 3;
	return result;
}

//----- (0051EBB0) --------------------------------------------------------
int nox_xxx_genReadAlgData_51EBB0(int a1, FILE* a2) {
	int v2;            // edi
	const char** v3;   // eax
	unsigned char* v4; // esi
	int v5;            // ecx
	char v7;           // [esp+10h] [ebp-3Ch]

	while (nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		if (!nox_strcmpi("END", (const char*)getMemAt(0x5D4594, 2487264))) {
			return 1;
		}
		if (!nox_xxx_mapGenReadLine_51E540(a2, &v7)) {
			break;
		}
		if (nox_strcmpi("midHallLength", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("hallLengthVariance", (const char*)getMemAt(0x5D4594, 2487264))) {
				if (nox_strcmpi("midHallWidth", (const char*)getMemAt(0x5D4594, 2487264))) {
					if (nox_strcmpi("hallWidthVariance", (const char*)getMemAt(0x5D4594, 2487264))) {
						if (nox_strcmpi("hallLimit", (const char*)getMemAt(0x5D4594, 2487264))) {
							if (nox_strcmpi("hallBranchRate", (const char*)getMemAt(0x5D4594, 2487264))) {
								if (nox_strcmpi("hallRoomRate", (const char*)getMemAt(0x5D4594, 2487264))) {
									if (nox_strcmpi("midRoomSize", (const char*)getMemAt(0x5D4594, 2487264))) {
										if (nox_strcmpi("roomVariance", (const char*)getMemAt(0x5D4594, 2487264))) {
											if (nox_strcmpi("irregularRoomRate",
														 (const char*)getMemAt(0x5D4594, 2487264))) {
												if (nox_strcmpi("mapSize", (const char*)getMemAt(0x5D4594, 2487264))) {
													if (nox_strcmpi("recurs"
																 "ionLim"
																 "it",
																 (const char*)getMemAt(0x5D4594, 2487264))) {
														if (nox_strcmpi("seed",
																	 (const char*)getMemAt(0x5D4594, 2487264))) {
															if (nox_strcmpi("emptyRoomRate",
																		 (const char*)getMemAt(0x5D4594, 2487264))) {
																if (nox_strcmpi("mergeRate", (const char*)getMemAt(
																							  0x5D4594, 2487264))) {
																	if (nox_strcmpi("useDoors", (const char*)getMemAt(
																								 0x5D4594, 2487264))) {
																		if (nox_strcmpi("debug", (const char*)getMemAt(
																								  0x5D4594, 2487264))) {
																			if (nox_strcmpi("adjacentPortalRate",
																						 (const char*)getMemAt(
																							 0x5D4594, 2487264))) {
																				if (!nox_strcmpi("skeleton",
																							  (const char*)getMemAt(
																								  0x5D4594, 2487264))) {
																					v2 = 0;
																					if (*getMemU32Ptr(0x587000,
																									  253244)) {
																						v3 = (const char**)getMemAt(
																							0x587000, 253244);
																						v4 = getMemAt(0x587000, 253244);
																						do {
																							if (!nox_strcmpi(*v3, &v7)) {
																								break;
																							}
																							v5 = *((uint32_t*)v4 + 1);
																							v4 += 4;
																							++v2;
																							v3 = (const char**)v4;
																						} while (v5);
																					}
																					*(uint32_t*)a1 =
																						*getMemU32Ptr(0x587000,
																									  253244 +
																										  4 * v2) != 0
																							? v2
																							: 0;
																				}
																			} else {
																				*(uint32_t*)(a1 + 52) = atoi(&v7);
																			}
																		} else if (nox_strcmpi(&v7, "true")) {
																			*(uint32_t*)(a1 + 60) = 0;
																		} else {
																			*(uint32_t*)(a1 + 60) = 1;
																		}
																	} else if (nox_strcmpi(&v7, "true")) {
																		*(uint32_t*)(a1 + 56) = 0;
																	} else {
																		*(uint32_t*)(a1 + 56) = 1;
																	}
																} else {
																	*(uint32_t*)(a1 + 48) = atoi(&v7);
																}
															} else {
																*(uint32_t*)(a1 + 44) = atoi(&v7);
															}
														} else {
															*(uint32_t*)(a1 + 76) = atoi(&v7);
														}
													} else {
														*(uint32_t*)(a1 + 72) = atoi(&v7);
													}
												} else {
													*(float*)(a1 + 64) = atof(&v7);
												}
											} else {
												*(uint32_t*)(a1 + 40) = atoi(&v7);
											}
										} else {
											*(uint32_t*)(a1 + 36) = atoi(&v7);
										}
									} else {
										*(uint32_t*)(a1 + 32) = atoi(&v7);
									}
								} else {
									*(uint32_t*)(a1 + 28) = atoi(&v7);
								}
							} else {
								*(uint32_t*)(a1 + 24) = atoi(&v7);
							}
						} else {
							*(uint32_t*)(a1 + 20) = atoi(&v7);
						}
					} else {
						*(uint32_t*)(a1 + 16) = atoi(&v7);
					}
				} else {
					*(uint32_t*)(a1 + 12) = atoi(&v7);
				}
			} else {
				*(uint32_t*)(a1 + 8) = atoi(&v7);
			}
		} else {
			*(uint32_t*)(a1 + 4) = atoi(&v7);
		}
	}
	return 0;
}

//----- (0051EFB0) --------------------------------------------------------
int nox_xxx_genReadSpellSet_51EFB0(int a1, FILE* a2) {
	int v2; // eax

	while (1) {
		if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
			return 0;
		}
		if (!nox_strcmpi("END", (const char*)getMemAt(0x5D4594, 2487264)) || *(uint32_t*)(a1 + 1096) >= 137) {
			break;
		}
		v2 = nox_xxx_mapGenSpellIdByName_51E1D0((const char*)getMemAt(0x5D4594, 2487264));
		if (v2) {
			*(uint32_t*)(a1 + 4 * (*(uint32_t*)(a1 + 1096))++ + 548) = v2;
		}
	}
	return 1;
}

//----- (0051F030) --------------------------------------------------------
enum {
	NOX_MAPGEN_ITEM_SET_MODIFIER_SLOTS = 4,
	NOX_MAPGEN_ITEM_SET_MAX_MODIFIERS = 256,
	NOX_MAPGEN_ITEM_SET_NAME_SIZE = 60,
};

static int nox_mapgenCopyItemSetName_51F030(char destination[NOX_MAPGEN_ITEM_SET_NAME_SIZE], const char* source) {
	if (!source) {
		return 0;
	}
	size_t length = strlen(source);
	if (length >= NOX_MAPGEN_ITEM_SET_NAME_SIZE) {
		return 0;
	}
	memcpy(destination, source, length + 1);
	return 1;
}

static int nox_mapgenFindItemSetName_51F230(const char names[][NOX_MAPGEN_ITEM_SET_NAME_SIZE], uint32_t count,
											const char* name) {
	for (uint32_t i = 0; i < count; ++i) {
		if (!nox_strcmpi(names[i], name)) {
			return (int)i;
		}
	}
	return -1;
}

static void nox_mapgenFreeItemSetModifiers_51F1F0(nox_mapgen_item_set_entry_legacy* entry) {
	if (!entry) {
		return;
	}
	for (int i = 0; i < NOX_MAPGEN_ITEM_SET_MODIFIER_SLOTS; ++i) {
		void* modifiers = nox_mapgenLegacyPtrResolve(entry->modifier_tokens[i]);
		if (modifiers) {
			nox_mapgenLegacyPtrForget(modifiers);
			free(modifiers);
		}
		entry->modifier_tokens[i] = 0;
		entry->modifier_counts[i] = 0;
	}
}

//----- (0051F1F0) --------------------------------------------------------
void nox_xxx_mapGenFreeStr_51F1F0(void* lpMem) {
	nox_mapgen_item_set_entry_legacy* entry =
		(nox_mapgen_item_set_entry_legacy*)nox_mapgenLegacyPtrResolve((uintptr_t)lpMem);
	if (!entry) {
		return;
	}
	nox_mapgenFreeItemSetModifiers_51F1F0(entry);
	nox_mapgenLegacyPtrForget(entry);
	free(entry);
}

static void nox_mapgenFreeItemSetTemplate_51F030(void) {
	uint32_t token = dword_5d4594_2487524;
	dword_5d4594_2487524 = 0;
	if (token) {
		nox_xxx_mapGenFreeStr_51F1F0(nox_mapgenItemSetEntryResolve(token));
	}
}

static int nox_mapgenItemSetModifierSlot_51F230(const char* name) {
	if (!nox_strcmpi("EFFECTIVENESS", name) || !nox_strcmpi("QUALITY", name)) {
		return 0;
	}
	if (!nox_strcmpi("MATERIAL", name)) {
		return 1;
	}
	if (!nox_strcmpi("PRIMARY_ENCHANTMENT", name)) {
		return 2;
	}
	if (!nox_strcmpi("SECONDARY_ENCHANTMENT", name)) {
		return 3;
	}
	return -1;
}

//----- (0051F230) --------------------------------------------------------
int sub_51F230(uint8_t* a1, FILE* a2) {
	nox_mapgen_item_set_entry_legacy* entry = (nox_mapgen_item_set_entry_legacy*)a1;
	char parsed[NOX_MAPGEN_ITEM_SET_MODIFIER_SLOTS][NOX_MAPGEN_ITEM_SET_MAX_MODIFIERS][NOX_MAPGEN_ITEM_SET_NAME_SIZE] =
		{0};
	uint32_t parsed_counts[NOX_MAPGEN_ITEM_SET_MODIFIER_SLOTS] = {0};

	if (!entry || !a2) {
		return 0;
	}
	while (nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		const char* line = (const char*)getMemAt(0x5D4594, 2487264);
		if (!nox_strcmpi("END", line)) {
			nox_mapgen_item_set_entry_legacy* item_template = nox_mapgenItemSetEntryResolve(dword_5d4594_2487524);
			for (int slot = 0; slot < NOX_MAPGEN_ITEM_SET_MODIFIER_SLOTS; ++slot) {
				char merged[NOX_MAPGEN_ITEM_SET_MAX_MODIFIERS][NOX_MAPGEN_ITEM_SET_NAME_SIZE] = {0};
				uint32_t merged_count = 0;
				if (item_template) {
					merged_count = item_template->modifier_counts[slot];
					char(*inherited)[NOX_MAPGEN_ITEM_SET_NAME_SIZE] =
						(char(*)[NOX_MAPGEN_ITEM_SET_NAME_SIZE])nox_mapgenLegacyPtrResolve(
							item_template->modifier_tokens[slot]);
					if (merged_count > NOX_MAPGEN_ITEM_SET_MAX_MODIFIERS || (merged_count && !inherited)) {
						nox_mapgenFreeItemSetModifiers_51F1F0(entry);
						return 0;
					}
					for (uint32_t i = 0; i < merged_count; ++i) {
						memcpy(merged[i], inherited[i], NOX_MAPGEN_ITEM_SET_NAME_SIZE);
						merged[i][NOX_MAPGEN_ITEM_SET_NAME_SIZE - 1] = '\0';
					}
				}

				for (uint32_t i = 0; i < parsed_counts[slot]; ++i) {
					const char* change = parsed[slot][i];
					if (change[0] == '-') {
						int found = nox_mapgenFindItemSetName_51F230(merged, merged_count, change + 1);
						if (found >= 0) {
							uint32_t tail = merged_count - (uint32_t)found - 1;
							if (tail) {
								memmove(merged[found], merged[found + 1], tail * NOX_MAPGEN_ITEM_SET_NAME_SIZE);
							}
							--merged_count;
						}
						continue;
					}
					if (nox_mapgenFindItemSetName_51F230(merged, merged_count, change) >= 0) {
						continue;
					}
					if (merged_count >= NOX_MAPGEN_ITEM_SET_MAX_MODIFIERS) {
						nox_mapgenFreeItemSetModifiers_51F1F0(entry);
						return 0;
					}
					memcpy(merged[merged_count], change, NOX_MAPGEN_ITEM_SET_NAME_SIZE);
					++merged_count;
				}

				if (merged_count) {
					char(*modifiers)[NOX_MAPGEN_ITEM_SET_NAME_SIZE] =
						calloc(merged_count, NOX_MAPGEN_ITEM_SET_NAME_SIZE);
					if (!modifiers) {
						nox_mapgenFreeItemSetModifiers_51F1F0(entry);
						return 0;
					}
					uint32_t token = nox_mapgenLegacyPtrRegister(modifiers);
					if (!token) {
						free(modifiers);
						nox_mapgenFreeItemSetModifiers_51F1F0(entry);
						return 0;
					}
					memcpy(modifiers, merged, merged_count * NOX_MAPGEN_ITEM_SET_NAME_SIZE);
					entry->modifier_tokens[slot] = token;
				}
				entry->modifier_counts[slot] = merged_count;
			}
			return 1;
		}

		int slot = nox_mapgenItemSetModifierSlot_51F230(line);
		if (slot < 0) {
			return 0;
		}
		int section_done = 0;
		while (nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
			line = (const char*)getMemAt(0x5D4594, 2487264);
			if (!nox_strcmpi("END", line)) {
				section_done = 1;
				break;
			}
			if (parsed_counts[slot] >= NOX_MAPGEN_ITEM_SET_MAX_MODIFIERS ||
				!nox_mapgenCopyItemSetName_51F030(parsed[slot][parsed_counts[slot]], line)) {
				return 0;
			}
			++parsed_counts[slot];
		}
		if (!section_done) {
			return 0;
		}
	}
	return 0;
}

static int nox_mapgenAppendItemSetEntry_51F030(uint8_t* theme, size_t head_offset, size_t count_offset,
											   nox_mapgen_item_set_entry_legacy* entry, uint32_t entry_token) {
	uint32_t* head_token = (uint32_t*)(theme + head_offset);
	if (!*head_token) {
		*head_token = entry_token;
	} else {
		nox_mapgen_item_set_entry_legacy* tail = nox_mapgenItemSetEntryResolve(*head_token);
		if (!tail) {
			return 0;
		}
		while (tail->next_token) {
			tail = nox_mapgenItemSetEntryResolve(tail->next_token);
			if (!tail) {
				return 0;
			}
		}
		tail->next_token = entry_token;
	}
	entry->next_token = 0;
	++*(uint32_t*)(theme + count_offset);
	return 1;
}

static int nox_mapgenReadItemSet_51F030(uint8_t* theme, FILE* file, const char* record_name, size_t head_offset,
										size_t count_offset) {
	if (!theme || !file) {
		return 0;
	}
	nox_mapgenFreeItemSetTemplate_51F030();
	while (nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		const char* line = (const char*)getMemAt(0x5D4594, 2487264);
		if (!nox_strcmpi("END", line)) {
			nox_mapgenFreeItemSetTemplate_51F030();
			return 1;
		}
		if (nox_strcmpi(record_name, line)) {
			continue;
		}

		nox_mapgen_item_set_entry_legacy* entry = calloc(1, sizeof(*entry));
		if (!entry) {
			nox_mapgenFreeItemSetTemplate_51F030();
			return 0;
		}
		uint32_t entry_token = nox_mapgenLegacyPtrRegister(entry);
		if (!entry_token) {
			free(entry);
			nox_mapgenFreeItemSetTemplate_51F030();
			return 0;
		}
		if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264)) ||
			!nox_mapgenCopyItemSetName_51F030(entry->object_name, (const char*)getMemAt(0x5D4594, 2487264))) {
			nox_xxx_mapGenFreeStr_51F1F0(entry);
			nox_mapgenFreeItemSetTemplate_51F030();
			return 0;
		}
		int is_template = !nox_strcmpi("TEMPLATE", entry->object_name);
		if (!is_template &&
			(!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264)) ||
			 !nox_mapgenCopyItemSetName_51F030(entry->lookup_name, (const char*)getMemAt(0x5D4594, 2487264)))) {
			nox_xxx_mapGenFreeStr_51F1F0(entry);
			nox_mapgenFreeItemSetTemplate_51F030();
			return 0;
		}
		if (!sub_51F230((uint8_t*)entry, file)) {
			nox_xxx_mapGenFreeStr_51F1F0(entry);
			nox_mapgenFreeItemSetTemplate_51F030();
			return 0;
		}

		if (is_template) {
			nox_mapgenFreeItemSetTemplate_51F030();
			dword_5d4594_2487524 = entry_token;
		} else if (!nox_mapgenAppendItemSetEntry_51F030(theme, head_offset, count_offset, entry, entry_token)) {
			nox_xxx_mapGenFreeStr_51F1F0(entry);
			nox_mapgenFreeItemSetTemplate_51F030();
			return 0;
		}
	}
	nox_mapgenFreeItemSetTemplate_51F030();
	return 0;
}

int nox_xxx_genReadWeaponSet_51F030(uint8_t* a1, FILE* a2) {
	return nox_mapgenReadItemSet_51F030(a1, a2, "WEAPON", 1100, 1104);
}

//----- (0051F640) --------------------------------------------------------
int nox_xxx_genReadArmorSet_51F640(uint8_t* a1, FILE* a2) {
	return nox_mapgenReadItemSet_51F030(a1, a2, "ARMOR", 1108, 1112);
}

//----- (0051F800) --------------------------------------------------------
int nox_xxx_genReadExit_51F800(int a1, FILE* a2) {
	while (1) {
		if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
			return 0;
		}
		if (!nox_strcmpi("END", (const char*)getMemAt(0x5D4594, 2487264))) {
			break;
		}
		if (nox_strcmpi("OBJECT", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("LINKDATA", (const char*)getMemAt(0x5D4594, 2487264)) ||
				!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			strcpy((char*)(a1 + 476), (const char*)getMemAt(0x5D4594, 2487264));
		} else {
			if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			strcpy((char*)(a1 + 216 + (*(uint32_t*)(a1 + 472) << 6)), (const char*)getMemAt(0x5D4594, 2487264));
			if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			if (nox_strcmpi("NORTH", (const char*)getMemAt(0x5D4594, 2487264))) {
				if (nox_strcmpi("SOUTH", (const char*)getMemAt(0x5D4594, 2487264))) {
					if (nox_strcmpi("EAST", (const char*)getMemAt(0x5D4594, 2487264))) {
						if (!nox_strcmpi("WEST", (const char*)getMemAt(0x5D4594, 2487264))) {
							*(uint32_t*)((*(uint32_t*)(a1 + 472) << 6) + a1 + 216 + 60) = 3;
						}
						++*(uint32_t*)(a1 + 472);
					} else {
						*(uint32_t*)(((*(uint32_t*)(a1 + 472))++ << 6) + a1 + 216 + 60) = 2;
					}
				} else {
					*(uint32_t*)(((*(uint32_t*)(a1 + 472))++ << 6) + a1 + 216 + 60) = 1;
				}
			} else {
				*(uint32_t*)(((*(uint32_t*)(a1 + 472))++ << 6) + a1 + 216 + 60) = 0;
			}
		}
	}
	return 1;
}

//----- (0051F9F0) --------------------------------------------------------
int nox_xxx_genReadDecor_51F9F0(uint32_t* a1, FILE* a2) {
	if (!a1 || !a2) {
		return 0;
	}
	nox_mapgen_decor_definition_legacy* definition = calloc(1u, sizeof(*definition));
	if (!definition) {
		return 0;
	}
	uint32_t definition_token = nox_mapgenLegacyPtrRegister(definition);
	if (!definition_token) {
		free(definition);
		return 0;
	}
	definition->frequency = 1000;
	definition->max_room_size = 999999;
	nox_mapgen_wall_floor_legacy* current_wall_floor = NULL;

	if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264)) ||
		!nox_xxx_mapGenReadLine_51E540(a2, (uint8_t*)definition->name)) {
		goto fail;
	}
	const char* room_type = (const char*)getMemAt(0x5D4594, 2487264);
	if (!nox_strcmpi("ROOM", room_type)) {
		definition->room_type = 0;
	} else if (!nox_strcmpi("HALL", room_type)) {
		definition->room_type = 1;
	} else if (!nox_strcmpi("TEMPLATE", room_type)) {
		definition->room_type = 2;
	} else if (!nox_strcmpi("BACKDROP", room_type)) {
		definition->room_type = 3;
	} else {
		goto fail;
	}

	for (;;) {
		if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
			goto fail;
		}
		const char* line = (const char*)getMemAt(0x5D4594, 2487264);
		if (!nox_strcmpi("END", line)) {
			static const uint8_t head_index[] = {22, 30, 38, 46};
			uint8_t index = head_index[definition->room_type];
			definition->next_token = a1[index];
			a1[index] = definition_token;
			++a1[index + 1];
			return 1;
		}
		if (!nox_strcmpi("WALL_FLOOR", line)) {
			current_wall_floor = nox_mapgenReadWallFloorNative_51FE00(definition, a2);
			if (!current_wall_floor) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("OCCUR_CONSTRAINT", line)) {
			if (!nox_xxx_genDecorReadOccurConstraint_520810((int)definition_token, a2)) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("OCCUR_LIMIT", line)) {
			if (!nox_xxx_genDecorReadOccurLimit_5208D0((int)definition_token, a2)) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("MUST_OCCUR", line)) {
			definition->must_occur = 1;
			continue;
		}
		if (!nox_strcmpi("FREQUENCY", line)) {
			if (!nox_xxx_genDecorReadFrequency_520910((int)definition_token, a2)) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("DOOR", line)) {
			if (!nox_xxx_genDecorReadDoor_520A90((int)definition_token, a2)) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("DOUBLE_DOOR", line)) {
			if (!nox_xxx_genDecorReadDoubleDoor_520AB0((int)definition_token, a2)) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("ROOM_SIZE_CONSTRAINT", line)) {
			if (!nox_xxx_genDecorReadRoomSizeCon_5209F0((int)definition_token, a2)) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("COPY", line)) {
			if (!nox_mapgenReadDecorCopyNative_520660((uint8_t*)a1, definition, a2)) {
				goto fail;
			}
			continue;
		}
		if (!nox_strcmpi("DECOR_SET", line)) {
			if (!nox_mapgenReadDecorSetNative_51FFA0(definition, a2)) {
				goto fail;
			}
			continue;
		}

		int fill_type = 0;
		const char* fill_name = (const char*)getMemPtr(0x587000, 253200);
		while (fill_name && nox_strcmpi(fill_name, line)) {
			++fill_type;
			fill_name = (const char*)getMemPtr(0x587000, 253200 + 4 * fill_type);
		}
		if (!fill_name || !current_wall_floor ||
			!nox_mapgenReadDecorFillNative_51FEC0(current_wall_floor, fill_type, a2)) {
			goto fail;
		}
	}

fail:
	nox_mapgenFreeDecorDefinitionList(definition_token);
	return 0;
}

//----- (0051FE00) --------------------------------------------------------
static nox_mapgen_wall_floor_legacy* nox_mapgenReadWallFloorNative_51FE00(
	nox_mapgen_decor_definition_legacy* definition, FILE* file) {
	if (!definition || !file) {
		return NULL;
	}
	nox_mapgen_wall_floor_legacy* wall_floor = calloc(1u, sizeof(*wall_floor));
	if (!wall_floor) {
		return NULL;
	}
	uint32_t token = nox_mapgenLegacyPtrRegister(wall_floor);
	if (!token) {
		free(wall_floor);
		return NULL;
	}
	if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		goto fail;
	}
	strcpy(wall_floor->wall_name, (const char*)getMemAt(0x5D4594, 2487264));
	if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		goto fail;
	}
	strcpy(wall_floor->floor_name, (const char*)getMemAt(0x5D4594, 2487264));
	wall_floor->next_token = definition->wall_floors_token;
	definition->wall_floors_token = token;
	++definition->wall_floor_count;
	return wall_floor;

fail:
	nox_mapgenLegacyPtrForget(wall_floor);
	free(wall_floor);
	return NULL;
}

char* nox_xxx_genDecorReadWallFloor_51FE00(int a1, FILE* a2) {
	return (char*)nox_mapgenReadWallFloorNative_51FE00(
		nox_mapgenDecorDefinitionResolve((uint32_t)a1), a2);
}

//----- (0051FEC0) --------------------------------------------------------
static int nox_mapgenReadDecorFillNative_51FEC0(
	nox_mapgen_wall_floor_legacy* wall_floor, int type, FILE* file) {
	if (!wall_floor || !file) {
		return 0;
	}
	nox_mapgen_decor_fill_legacy* fill = calloc(1u, sizeof(*fill));
	if (!fill) {
		return 0;
	}
	uint32_t token = nox_mapgenLegacyPtrRegister(fill);
	if (!token) {
		free(fill);
		return 0;
	}
	fill->type = (uint32_t)type;
	if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		goto fail;
	}
	strcpy(fill->object_name, (const char*)getMemAt(0x5D4594, 2487264));
	if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		goto fail;
	}
	strcpy(fill->tile_name, (const char*)getMemAt(0x5D4594, 2487264));

	nox_mapgen_decor_fill_legacy* tail = nox_mapgenDecorFillResolve(wall_floor->fills_token);
	if (!tail) {
		wall_floor->fills_token = token;
		return 1;
	}
	while (tail->next_token) {
		tail = nox_mapgenDecorFillResolve(tail->next_token);
		if (!tail) {
			goto fail;
		}
	}
	tail->next_token = token;
	return 1;

fail:
	nox_mapgenLegacyPtrForget(fill);
	free(fill);
	return 0;
}

int sub_51FEC0(int a1, int a2, FILE* a3) {
	return nox_mapgenReadDecorFillNative_51FEC0(
		nox_mapgenWallFloorResolve((uint32_t)a1), a2, a3);
}

//----- (0051FFA0) --------------------------------------------------------
static int nox_mapgenReadDecorSetNative_51FFA0(
	nox_mapgen_decor_definition_legacy* definition, FILE* file) {
	if (!definition || !file) {
		return 0;
	}
	nox_mapgen_decor_set_legacy* set = calloc(1u, sizeof(*set));
	if (!set) {
		return 0;
	}
	uint32_t set_token = nox_mapgenLegacyPtrRegister(set);
	if (!set_token) {
		free(set);
		return 0;
	}
	if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		goto fail;
	}
	set->min_count = atoi((const char*)getMemAt(0x5D4594, 2487264));
	if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		goto fail;
	}
	set->max_count = atoi((const char*)getMemAt(0x5D4594, 2487264));

	nox_mapgen_decor_entry_legacy* current = NULL;
	for (;;) {
		if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
			goto fail;
		}
		const char* line = (const char*)getMemAt(0x5D4594, 2487264);
		if (!nox_strcmpi("END", line)) {
			nox_mapgen_decor_set_legacy* tail =
				nox_mapgenDecorSetResolve(definition->decor_sets_token);
			if (!tail) {
				definition->decor_sets_token = set_token;
			} else {
				while (tail->next_token) {
					tail = nox_mapgenDecorSetResolve(tail->next_token);
					if (!tail) {
						goto fail;
					}
				}
				tail->next_token = set_token;
			}
			++definition->decor_set_count;
			return 1;
		}
		if (!nox_strcmpi("FOREACH", line)) {
			if (!current || current->type != 1) {
				goto fail;
			}
			nox_mapgen_foreach_legacy* foreach =
				(nox_mapgen_foreach_legacy*)nox_xxx_gen_5205B0(file);
			if (!foreach) {
				goto fail;
			}
			foreach->next_token = current->foreach_token;
			current->foreach_token = nox_mapgenLegacyPtrRegister(foreach);
			continue;
		}
		if (!nox_strcmpi("CONTAINS", line)) {
			if (!current || current->type != 0) {
				goto fail;
			}
			uint32_t* choices = nox_xxx_gen_520380(file);
			if (!choices) {
				goto fail;
			}
			current->choices_token = nox_mapgenLegacyPtrRegister(choices);
			continue;
		}

		nox_mapgen_decor_entry_legacy* entry = calloc(1u, sizeof(*entry));
		if (!entry) {
			goto fail;
		}
		uint32_t entry_token = nox_mapgenLegacyPtrRegister(entry);
		if (!entry_token) {
			free(entry);
			goto fail;
		}
		const char* type_name =
			(const char*)getMemPtr(0x587000, 253216 + 4 * entry->type);
		while (type_name) {
			if (!nox_strcmpi(type_name, line)) {
				break;
			}
			++entry->type;
			type_name =
				(const char*)getMemPtr(0x587000, 253216 + 4 * entry->type);
		}
		if (!type_name) {
			nox_mapgenLegacyPtrForget(entry);
			free(entry);
			goto fail;
		}

		if (entry->type == 0 || entry->type == 1) {
			if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
				goto fail_entry;
			}
			strcpy(entry->name, (const char*)getMemAt(0x5D4594, 2487264));
			if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
				goto fail_entry;
			}
			const char* amount = (const char*)getMemAt(0x5D4594, 2487264);
			if (nox_strcmpi("DENSITY", amount)) {
				entry->min_count = atoi(amount);
				if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
					goto fail_entry;
				}
				entry->max_count = atoi((const char*)getMemAt(0x5D4594, 2487264));
			} else {
				entry->density_enabled = 1;
				if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
					goto fail_entry;
				}
				entry->density = atof((const char*)getMemAt(0x5D4594, 2487264));
				if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
					goto fail_entry;
				}
				const char* minimum = (const char*)getMemAt(0x5D4594, 2487264);
				entry->min_count = !strcmp("*", minimum) ? 0 : atoi(minimum);
				if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
					goto fail_entry;
				}
				const char* maximum = (const char*)getMemAt(0x5D4594, 2487264);
				entry->max_count = !strcmp("*", maximum) ? 999999 : atoi(maximum);
			}
		} else if (entry->type == 3 || entry->type == 4 || entry->type == 5) {
			if (!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
				goto fail_entry;
			}
			strcpy(entry->name, (const char*)getMemAt(0x5D4594, 2487264));
		}

		entry->next_token = set->entries_token;
		set->entries_token = entry_token;
		++set->entry_count;
		current = entry;
		continue;

	fail_entry:
		nox_mapgenLegacyPtrForget(entry);
		free(entry);
		goto fail;
	}

fail:
	nox_mapgenFreeDecorEntryList(set->entries_token);
	nox_mapgenLegacyPtrForget(set);
	free(set);
	return 0;
}

int nox_xxx_genDecorReadDecorSet_51FFA0(int a1, FILE* a2) {
	return nox_mapgenReadDecorSetNative_51FFA0(
		nox_mapgenDecorDefinitionResolve((uint32_t)a1), a2);
}

//----- (00520380) --------------------------------------------------------
uint32_t* nox_xxx_gen_520380(FILE* a1) {
	uint32_t* head = NULL;
	uint32_t* tail = NULL;
	int wildcard_count = 0;
	int probability_left = 100;

	for (;;) {
		uint32_t* choice = calloc(1u, 0x80Cu);
		if (!choice) {
			goto fail;
		}
		uint32_t choice_token = nox_mapgenLegacyPtrRegister(choice);
		if (!choice_token) {
			free(choice);
			goto fail;
		}
		if (!nox_xxx_mapGenReadLine_51E540(a1, getMemAt(0x5D4594, 2487264))) {
			nox_mapgenLegacyPtrForget(choice);
			free(choice);
			goto fail;
		}
		if (!strcmp("*", (const char*)getMemAt(0x5D4594, 2487264))) {
			choice[0] = UINT32_MAX;
			++wildcard_count;
		} else {
			choice[0] = (uint32_t)atoi((const char*)getMemAt(0x5D4594, 2487264));
			probability_left -= (int)choice[0];
		}

		for (;;) {
			if (!nox_xxx_mapGenReadLine_51E540(a1, getMemAt(0x5D4594, 2487264))) {
				nox_mapgenLegacyPtrForget(choice);
				free(choice);
				goto fail;
			}
			const char* line = (const char*)getMemAt(0x5D4594, 2487264);
			int is_end = !nox_strcmpi("END", line);
			if (is_end || !nox_strcmpi("OR", line)) {
				if (tail) {
					tail[NOX_MAPGEN_CHOICE_NEXT_INDEX] = choice_token;
				} else {
					head = choice;
				}
				tail = choice;
				if (!is_end) {
					break;
				}
				if (wildcard_count) {
					for (uint32_t* it = head; it;
						 it = nox_mapgenChoiceNextNative_520380((uintptr_t)it)) {
						if ((int32_t)it[0] < 0) {
							it[0] = (uint32_t)(probability_left / wildcard_count + 1);
						}
					}
				}
				return head;
			}

			int object_type = 0;
			const char* type_name =
				(const char*)getMemPtr(0x587000, 253216 + 4 * object_type);
			while (type_name) {
				if (!nox_strcmpi(type_name, line)) {
					break;
				}
				++object_type;
				type_name =
					(const char*)getMemPtr(0x587000, 253216 + 4 * object_type);
			}
			if (!type_name) {
				nox_mapgenLegacyPtrForget(choice);
				free(choice);
				goto fail;
			}
			uint32_t count = choice[NOX_MAPGEN_CHOICE_COUNT_INDEX];
			choice[16 * count + 1] = (uint32_t)object_type;
			if (!nox_xxx_mapGenReadLine_51E540(a1, (uint8_t*)&choice[16 * count + 2])) {
				nox_mapgenLegacyPtrForget(choice);
				free(choice);
				goto fail;
			}
			choice[NOX_MAPGEN_CHOICE_COUNT_INDEX] = count < 31 ? count + 1 : 31;
		}
	}

fail:
	nox_mapgenFreeChoiceList((uintptr_t)head);
	return NULL;
}

//----- (005205B0) --------------------------------------------------------
uint32_t* nox_xxx_gen_5205B0(FILE* a1) {
	if (!nox_xxx_mapGenReadLine_51E540(a1, getMemAt(0x5D4594, 2487264))) {
		return NULL;
	}
	nox_mapgen_foreach_legacy* entry = calloc(1u, sizeof(*entry));
	if (!entry) {
		return NULL;
	}
	if (!nox_mapgenLegacyPtrRegister(entry)) {
		free(entry);
		return NULL;
	}
	entry->type_id = nox_xxx_getNameId_4E3AA0((char*)getMemAt(0x5D4594, 2487264));
	if (!nox_xxx_mapGenReadLine_51E540(a1, getMemAt(0x5D4594, 2487264)) ||
		nox_strcmpi("CONTAINS", (const char*)getMemAt(0x5D4594, 2487264))) {
		nox_mapgenLegacyPtrForget(entry);
		free(entry);
		return NULL;
	}
	uint32_t* choices = nox_xxx_gen_520380(a1);
	if (!choices) {
		nox_mapgenLegacyPtrForget(entry);
		free(entry);
		return NULL;
	}
	entry->choices_token = nox_mapgenLegacyPtrRegister(choices);
	return (uint32_t*)entry;
}

//----- (00520660) --------------------------------------------------------
static uint32_t nox_mapgenCloneDecorFillList_520660(uintptr_t source_ref, int* ok) {
	uint32_t head_token = 0;
	nox_mapgen_decor_fill_legacy* tail = NULL;
	*ok = 1;
	for (nox_mapgen_decor_fill_legacy* source = nox_mapgenDecorFillResolve(source_ref);
		 source; source = nox_mapgenDecorFillResolve(source->next_token)) {
		nox_mapgen_decor_fill_legacy* clone = calloc(1u, sizeof(*clone));
		if (!clone) {
			*ok = 0;
			break;
		}
		memcpy(clone, source, sizeof(*clone));
		clone->next_token = 0;
		uint32_t clone_token = nox_mapgenLegacyPtrRegister(clone);
		if (!clone_token) {
			free(clone);
			*ok = 0;
			break;
		}
		if (tail) {
			tail->next_token = clone_token;
		} else {
			head_token = clone_token;
		}
		tail = clone;
	}
	if (!*ok) {
		nox_mapgenFreeDecorFillList(head_token);
		return 0;
	}
	return head_token;
}

static int nox_mapgenReadDecorCopyNative_520660(
	uint8_t* theme, nox_mapgen_decor_definition_legacy* destination, FILE* file) {
	if (!theme || !destination || !file ||
		!nox_xxx_mapGenReadLine_51E540(file, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	uint32_t source_token;
	switch (destination->room_type) {
	case 0:
		source_token = *(uint32_t*)(theme + 88);
		break;
	case 1:
		source_token = *(uint32_t*)(theme + 120);
		break;
	case 2:
		source_token = *(uint32_t*)(theme + 152);
		break;
	default:
		source_token = nox_mapgenLegacyPtrRegister(destination);
		break;
	}
	const char* source_name = (const char*)getMemAt(0x5D4594, 2487264);
	nox_mapgen_decor_definition_legacy* source = nox_mapgenDecorDefinitionResolve(source_token);
	while (source && strcmp(source->name, source_name)) {
		source = nox_mapgenDecorDefinitionResolve(source->next_token);
	}
	if (!source) {
		source = nox_mapgenDecorDefinitionResolve(*(uint32_t*)(theme + 152));
		while (source && strcmp(source->name, source_name)) {
			source = nox_mapgenDecorDefinitionResolve(source->next_token);
		}
	}
	if (!source) {
		return 0;
	}

	for (nox_mapgen_wall_floor_legacy* source_wall =
			 nox_mapgenWallFloorResolve(source->wall_floors_token);
		 source_wall; source_wall = nox_mapgenWallFloorResolve(source_wall->next_token)) {
		nox_mapgen_wall_floor_legacy* clone = calloc(1u, sizeof(*clone));
		if (!clone) {
			return 0;
		}
		memcpy(clone, source_wall, sizeof(*clone));
		int ok = 0;
		clone->fills_token = nox_mapgenCloneDecorFillList_520660(source_wall->fills_token, &ok);
		clone->next_token = destination->wall_floors_token;
		if (!ok) {
			free(clone);
			return 0;
		}
		uint32_t clone_token = nox_mapgenLegacyPtrRegister(clone);
		if (!clone_token) {
			nox_mapgenFreeDecorFillList(clone->fills_token);
			free(clone);
			return 0;
		}
		destination->wall_floors_token = clone_token;
		++destination->wall_floor_count;
	}

	nox_mapgen_decor_set_legacy* tail =
		nox_mapgenDecorSetResolve(destination->decor_sets_token);
	while (tail && tail->next_token) {
		tail = nox_mapgenDecorSetResolve(tail->next_token);
	}
	for (nox_mapgen_decor_set_legacy* source_set =
			 nox_mapgenDecorSetResolve(source->decor_sets_token);
		 source_set; source_set = nox_mapgenDecorSetResolve(source_set->next_token)) {
		nox_mapgen_decor_set_legacy* clone = calloc(1u, sizeof(*clone));
		if (!clone) {
			return 0;
		}
		memcpy(clone, source_set, sizeof(*clone));
		clone->shares_entries = 1;
		clone->next_token = 0;
		uint32_t clone_token = nox_mapgenLegacyPtrRegister(clone);
		if (!clone_token) {
			free(clone);
			return 0;
		}
		if (tail) {
			tail->next_token = clone_token;
		} else {
			destination->decor_sets_token = clone_token;
		}
		tail = clone;
		++destination->decor_set_count;
	}
	return 1;
}

int nox_xxx_genDecorReadCopy_520660(uint32_t* a1, const char* a2, FILE* a3) {
	return nox_mapgenReadDecorCopyNative_520660(
		(uint8_t*)a1, nox_mapgenDecorDefinitionResolve((uintptr_t)a2), a3);
}

//----- (00520810) --------------------------------------------------------
int nox_xxx_genDecorReadOccurConstraint_520810(int a1, FILE* a2) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve((uint32_t)a1);
	if (!definition) {
		return 0;
	}
	definition->constraint_mask = 0;
	if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	if (nox_strcmpi("NONE", (const char*)getMemAt(0x5D4594, 2487264))) {
		char* name = strtok((char*)getMemAt(0x5D4594, 2487264), "+");
		if (name) {
			while (1) {
				int index = 0;
				const char* constraint = (const char*)getMemPtr(0x587000, 253144);
				while (constraint && nox_strcmpi(constraint, name)) {
					++index;
					constraint = (const char*)getMemPtr(0x587000, 253144 + 4 * index);
				}
				if (!constraint) {
					break;
				}
				definition->constraint_mask |= 1 << index;
				name = strtok(0, "+");
				if (!name) {
					return 1;
				}
			}
			return 0;
		}
	}
	return 1;
}

//----- (005208D0) --------------------------------------------------------
int nox_xxx_genDecorReadOccurLimit_5208D0(int a1, FILE* a2) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve((uint32_t)a1);
	if (!definition || !nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	definition->occur_limit = atoi((const char*)getMemAt(0x5D4594, 2487264));
	return 1;
}

//----- (00520910) --------------------------------------------------------
int nox_xxx_genDecorReadFrequency_520910(int a1, FILE* a2) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve((uint32_t)a1);
	if (!definition || !nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	const char* frequency = (const char*)getMemAt(0x5D4594, 2487264);
	if (!nox_strcmpi("COMMON", frequency)) {
		definition->frequency = 1000;
	} else if (!nox_strcmpi("UNCOMMON", frequency)) {
		definition->frequency = 500;
	} else if (!nox_strcmpi("RARE", frequency)) {
		definition->frequency = 100;
	} else if (!nox_strcmpi("VERY_RARE", frequency)) {
		definition->frequency = 10;
	} else if (!nox_strcmpi("HARDLY_EVER", frequency)) {
		definition->frequency = 1;
	}
	return 1;
}

//----- (005209F0) --------------------------------------------------------
int nox_xxx_genDecorReadRoomSizeCon_5209F0(int a1, FILE* a2) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve((uint32_t)a1);
	if (!definition || !nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	const char* minimum = (const char*)getMemAt(0x5D4594, 2487264);
	definition->min_room_size = nox_strcmpi("*", minimum) ? atoi(minimum) : 0;
	if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		return 0;
	}
	const char* maximum = (const char*)getMemAt(0x5D4594, 2487264);
	definition->max_room_size = nox_strcmpi("*", maximum) ? atoi(maximum) : 999999;
	return 1;
}

//----- (00520A90) --------------------------------------------------------
int nox_xxx_genDecorReadDoor_520A90(int a1, FILE* a2) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve((uint32_t)a1);
	return definition && nox_xxx_mapGenReadLine_51E540(a2, (uint8_t*)definition->door_name) != 0;
}

//----- (00520AB0) --------------------------------------------------------
int nox_xxx_genDecorReadDoubleDoor_520AB0(int a1, FILE* a2) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve((uint32_t)a1);
	return definition && nox_xxx_mapGenReadLine_51E540(a2, (uint8_t*)definition->double_door_name) != 0;
}

//----- (00520AD0) --------------------------------------------------------
int nox_xxx_mapgenCheckSettings_520AD0(int* a1) {
	if (!a1) {
		return 0;
	}
	a1[2] = 0;
	a1[3] = 0;
	a1[4] = 0;
	a1[5] = 0;
	a1[6] = 0;
	a1[7] = 0;
	for (nox_mapgen_decor_definition_legacy* definition =
			 nox_mapgenDecorDefinitionResolve((uint32_t)a1[0]);
		 definition; definition = nox_mapgenDecorDefinitionResolve(definition->next_token)) {
		if (!definition->constraint_mask) {
			a1[2] += definition->frequency;
			a1[3] += definition->frequency;
			a1[4] += definition->frequency;
			a1[5] += definition->frequency;
			a1[6] += definition->frequency;
			a1[7] += definition->frequency;
		}
		if (definition->constraint_mask & 1) {
			a1[2] += definition->frequency;
		}
		if (definition->constraint_mask & 2) {
			a1[3] += definition->frequency;
		}
		if (definition->constraint_mask & 4) {
			a1[4] += definition->frequency;
		}
		if (definition->constraint_mask & 8) {
			a1[5] += definition->frequency;
		}
		if (definition->constraint_mask & 0x10) {
			a1[6] += definition->frequency;
		}
		if (definition->constraint_mask & 0x20) {
			a1[7] += definition->frequency;
		}
	}
	if (!a1[2]) {
		return 0;
	}
	if (!a1[3]) {
		return 0;
	}
	if (!a1[4]) {
		return 0;
	}
	if (!a1[5]) {
		return 0;
	}
	if (a1[6]) {
		return a1[7] != 0;
	}
	return 0;
}

//----- (00520BF0) --------------------------------------------------------
int nox_xxx_genReadPrefab_520BF0(uint8_t* a1, FILE* a2) {
	char* v2;     // esi
	int v3;       // ebx
	char* v4;     // eax
	uint32_t* v5; // eax

	v2 = 0;
	v3 = 0;
	while (nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		if (!nox_strcmpi("END", (const char*)getMemAt(0x5D4594, 2487264))) {
			return 1;
		}
		if (nox_strcmpi("AREAMAP", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("MUST_OCCUR", (const char*)getMemAt(0x5D4594, 2487264))) {
				if (!nox_strcmpi("FOREACH", (const char*)getMemAt(0x5D4594, 2487264))) {
					if (!v2) {
						return 0;
					}
					v5 = nox_xxx_gen_5205B0(a2);
					if (!v5) {
						return 0;
					}
					v5[2] = *((uint32_t*)v2 + 38);
					*((uint32_t*)v2 + 38) = nox_mapgenLegacyPtrRegister(v5);
				}
			} else {
				v3 = 1;
			}
		} else {
			v4 = sub_520CE0(a1, a2);
			v2 = v4;
			if (!v4) {
				return 0;
			}
			if (v3) {
				*((uint32_t*)v4 + 18) = 1;
				v3 = 0;
			}
		}
	}
	return 0;
}

//----- (00520CE0) --------------------------------------------------------
char* sub_520CE0(uint8_t* a1, FILE* a2) {
	char* result; // eax
	int v3;       // ecx

	result = (char*)nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264));
	if (result) {
		result = (char*)calloc(1u, 0xA0u);
		if (result) {
			uint32_t token = nox_mapgenLegacyPtrRegister(result);
			if (!token) {
				free(result);
				return NULL;
			}
			strcpy(result, (const char*)getMemAt(0x5D4594, 2487264));
			*((uint32_t*)result + 39) = *(uint32_t*)(a1 + 80);
			v3 = *(uint32_t*)(a1 + 84) + 1;
			*(uint32_t*)(a1 + 80) = token;
			*(uint32_t*)(a1 + 84) = v3;
		}
	}
	return result;
}

//----- (00520D50) --------------------------------------------------------
uint32_t* sub_520D50(uint32_t* a1) {
	if (!a1) {
		return NULL;
	}
	// COPY records share decor entries with earlier definitions. Lists are
	// prepended while parsing, so freeing each list from the head releases the
	// shared copies before their owners; templates are released last because
	// room, hall, and backdrop definitions may refer to them.
	static const uint16_t definition_indices[] = {22, 30, 46, 38};
	for (size_t i = 0; i < sizeof(definition_indices) / sizeof(definition_indices[0]); ++i) {
		uint16_t index = definition_indices[i];
		nox_mapgenFreeDecorDefinitionList(a1[index]);
		a1[index] = 0;
		a1[index + 1] = 0;
	}

	nox_mapgen_item_set_entry_legacy* item = nox_mapgenItemSetEntryResolve(a1[275]);
	while (item) {
		nox_mapgen_item_set_entry_legacy* next = nox_mapgenItemSetEntryResolve(item->next_token);
		nox_xxx_mapGenFreeStr_51F1F0(item);
		item = next;
	}
	a1[275] = 0;
	a1[276] = 0;
	item = nox_mapgenItemSetEntryResolve(a1[277]);
	while (item) {
		nox_mapgen_item_set_entry_legacy* next = nox_mapgenItemSetEntryResolve(item->next_token);
		nox_xxx_mapGenFreeStr_51F1F0(item);
		item = next;
	}
	a1[277] = 0;
	a1[278] = 0;
	uint32_t* prefab = (uint32_t*)nox_mapgenLegacyPtrResolve(a1[20]);
	while (prefab) {
		uint32_t* next = (uint32_t*)nox_mapgenLegacyPtrResolve(prefab[39]);
		nox_mapgenFreeForeachList(prefab[38]);
		nox_mapgenLegacyPtrForget(prefab);
		free(prefab);
		prefab = next;
	}
	a1[20] = 0;
	a1[21] = 0;
	return NULL;
}

//----- (00520DF0) --------------------------------------------------------
long long nox_xxx_mapGenRoundFloatToPtr_520DF0(float2* a1, uint32_t* a2) {
	double v2;        // st7
	double v3;        // st7
	long long result; // rax

	if (a1->field_0 >= 0.0) {
		v2 = 0.5;
	} else {
		v2 = -0.5;
	}
	*a2 = (long long)(a1->field_0 * 0.030743772 + v2);
	if (a1->field_4 >= 0.0) {
		v3 = 0.5;
	} else {
		v3 = -0.5;
	}
	result = (long long)(a1->field_4 * 0.030743772 + v3);
	a2[1] = result;
	return result;
}

//----- (00520E60) --------------------------------------------------------
void* sub_520E60(int2* a1) {
	int v1;     // eax
	int v2;     // ecx
	void* result;

	v1 = dword_5d4594_2487536 + a1->field_0;
	v2 = dword_5d4594_2487536 + a1->field_4;
	if (v1 < 0 || v1 >= *(int*)&dword_5d4594_2487540 || v2 < 0 || v2 >= *(int*)&dword_5d4594_2487540) {
		result = NULL;
	} else {
		result = nox_mapgen_occupancy_grid[v1] + 20 * v2;
	}
	return result;
}

//----- (00520EA0) --------------------------------------------------------
int sub_520EA0(uint8_t* a1) {
	int v2; // edx
	int v3; // esi
	int v4; // esi
	int v5; // ecx
	int v6; // eax
	int v7; // ebx

	dword_5d4594_2487540 = 2 * *(uint32_t*)(a1 + 68) + 1;
	dword_5d4594_2487536 = *(uint32_t*)(a1 + 68);
	nox_mapgen_occupancy_grid = (uint8_t**)calloc(dword_5d4594_2487540, sizeof(*nox_mapgen_occupancy_grid));
	if (!nox_mapgen_occupancy_grid) {
		return 0;
	}
	dword_5d4594_2487532 = nox_mapgenLegacyPtrRegister(nox_mapgen_occupancy_grid);
	if (!dword_5d4594_2487532) {
		free(nox_mapgen_occupancy_grid);
		nox_mapgen_occupancy_grid = NULL;
		return 0;
	}
	v2 = dword_5d4594_2487540;
	v3 = 0;
	if (dword_5d4594_2487540 > 0) {
		do {
			nox_mapgen_occupancy_grid[v3] = (uint8_t*)calloc(v2, 20);
			if (!nox_mapgen_occupancy_grid[v3]) {
				for (int i = 0; i < v3; ++i) {
					free(nox_mapgen_occupancy_grid[i]);
				}
				nox_mapgenLegacyPtrForget(nox_mapgen_occupancy_grid);
				free(nox_mapgen_occupancy_grid);
				nox_mapgen_occupancy_grid = NULL;
				dword_5d4594_2487532 = 0;
				return 0;
			}
			v2 = dword_5d4594_2487540;
		} while (++v3 < *(int*)&dword_5d4594_2487540);
	}
	v4 = 0;
	if (v2 > 0) {
		v5 = 0;
		do {
			v6 = 0;
			if (v2 > 0) {
				do {
					*(uint32_t*)(nox_mapgen_occupancy_grid[v6] + v5 + 4) = v6 - dword_5d4594_2487536;
					v7 = v6++;
					*(uint32_t*)(nox_mapgen_occupancy_grid[v7] + v5 + 8) = v4 - dword_5d4594_2487536;
					*(uint32_t*)(nox_mapgen_occupancy_grid[v7] + v5 + 16) = 0;
					v2 = dword_5d4594_2487540;
				} while (v6 < *(int*)&dword_5d4594_2487540);
			}
			++v4;
			v5 += 20;
		} while (v4 < v2);
	}
	return 1;
}

//----- (00520F80) --------------------------------------------------------
void sub_520F80() {
	int i; // esi

	if (!nox_mapgen_occupancy_grid) {
		dword_5d4594_2487532 = 0;
		return;
	}
	for (i = 0; i < *(int*)&dword_5d4594_2487540; ++i) {
		free(nox_mapgen_occupancy_grid[i]);
	}
	nox_mapgenLegacyPtrForget(nox_mapgen_occupancy_grid);
	free(nox_mapgen_occupancy_grid);
	nox_mapgen_occupancy_grid = NULL;
	dword_5d4594_2487532 = 0;
}

//----- (00521100) --------------------------------------------------------
int nox_mapgenRoomOccupyNative_521100(uint8_t* a1) {
	int result; // eax
	int v2;     // ebx
	int v3;     // edi
	uint8_t* v4;
	int v5;     // eax
	int2 a1a;   // [esp+8h] [ebp-10h]
	int2 a2;    // [esp+10h] [ebp-8h]

	nox_xxx_mapGenRoundFloatToPtr_520DF0((float2*)(a1 + 20), &a2);
	uint32_t room_token = nox_mapgenLegacyPtrRegister(a1);
	if (!room_token) {
		return 0;
	}
	result = *(uint32_t*)(a1 + 16);
	v2 = 0;
	for (a1a.field_4 = a2.field_4; v2 < result; ++a1a.field_4) {
		v3 = 0;
		a1a.field_0 = a2.field_0;
		if (*(uint32_t*)(a1 + 12) > 0) {
			do {
				v4 = sub_520E60(&a1a);
				if (v4) {
					*(uint32_t*)(v4 + 16) = room_token;
				}
				v5 = *(uint32_t*)(a1 + 12);
				++v3;
				++a1a.field_0;
			} while (v3 < v5);
		}
		result = *(uint32_t*)(a1 + 16);
		++v2;
	}
	return result;
}

int sub_521100(int a1) {
	return nox_mapgenRoomOccupyNative_521100((uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (00521180) --------------------------------------------------------
int nox_mapgenRoomVacateNative_521180(uint8_t* a1) {
	int result; // eax
	int v2;     // ebx
	int v3;     // esi
	uint8_t* v4;
	int v5;     // eax
	int2 a1a;   // [esp+Ch] [ebp-10h]
	int2 a2;    // [esp+14h] [ebp-8h]

	nox_xxx_mapGenRoundFloatToPtr_520DF0((float2*)(a1 + 20), &a2);
	result = *(uint32_t*)(a1 + 16);
	v2 = 0;
	for (a1a.field_4 = a2.field_4; v2 < result; ++a1a.field_4) {
		v3 = 0;
		a1a.field_0 = a2.field_0;
		if (*(uint32_t*)(a1 + 12) > 0) {
			do {
				v4 = sub_520E60(&a1a);
				if (v4) {
					*(uint32_t*)(v4 + 16) = 0;
				}
				v5 = *(uint32_t*)(a1 + 12);
				++v3;
				++a1a.field_0;
			} while (v3 < v5);
		}
		result = *(uint32_t*)(a1 + 16);
		++v2;
	}
	return result;
}

int sub_521180(int a1) {
	return nox_mapgenRoomVacateNative_521180((uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (00521200) --------------------------------------------------------
void* nox_mapgenRoomAtNative_521200(uint8_t* a1) {
	int v1;   // eax
	int v2;   // ebx
	int v3;   // esi
	uint8_t* v4;
	int v5;   // eax
	int v6;   // eax
	int2 a1a; // [esp+Ch] [ebp-10h]
	int2 a2;  // [esp+14h] [ebp-8h]

	nox_xxx_mapGenRoundFloatToPtr_520DF0((float2*)(a1 + 20), &a2);
	v1 = *(uint32_t*)(a1 + 16);
	v2 = 0;
	a1a.field_4 = a2.field_4;
	if (v1 <= 0) {
		return 0;
	}
	while (1) {
		v3 = 0;
		a1a.field_0 = a2.field_0;
		if (*(uint32_t*)(a1 + 12) > 0) {
			while (1) {
				v4 = sub_520E60(&a1a);
				if (v4) {
					if (*(uint32_t*)(v4 + 16)) {
						return nox_mapgenLegacyPtrResolve(*(uint32_t*)(v4 + 16));
					}
				}
				v5 = *(uint32_t*)(a1 + 12);
				++v3;
				++a1a.field_0;
				if (v3 >= v5) {
					goto LABEL_6;
				}
			}
		}
	LABEL_6:
		v6 = *(uint32_t*)(a1 + 16);
		++v2;
		++a1a.field_4;
		if (v2 >= v6) {
			return 0;
		}
	}
}

int sub_521200(int a1) {
	void* room = nox_mapgenRoomAtNative_521200((uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
	return (int)nox_mapgenLegacyPtrRegister(room);
}

//----- (00521290) --------------------------------------------------------
void* nox_mapgenRoomAtCellNative_521290(int2* cell) {
	uint8_t* entry = sub_520E60(cell);
	return entry ? nox_mapgenLegacyPtrResolve(*(uint32_t*)(entry + 16)) : NULL;
}

int sub_521290(int2* a1) {
	uint8_t* v1;
	int result; // eax

	v1 = sub_520E60(a1);
	if (v1) {
		result = *(uint32_t*)(v1 + 16);
	} else {
		result = 0;
	}
	return result;
}

//----- (005212B0) --------------------------------------------------------
int sub_5212B0(uint8_t* a1, uint32_t* a2) {
	uint32_t* v2; // eax
	int v3;       // ebp
	int v4;       // ecx
	int v5;       // edx
	int v6;       // edi
	int v7;       // ecx
	int v8;       // edi
	int v9;       // edx
	int v10;      // eax
	int v11;      // ecx
	int v12;      // ebx
	int v13;      // ecx
	int v14;      // ecx
	int v15;      // eax
	int* v16;     // edx
	float2 a2a;   // [esp+10h] [ebp-18h]
	int v19[4];   // [esp+18h] [ebp-10h]

	v2 = a2;
	v3 = 0;
	while (1) {
		v4 = *(uint32_t*)(a1 + 4);
		v5 = v2[1];
		v6 = v2[3] - v4;
		v7 = *(uint32_t*)(a1 + 12) + v4 - v5;
		v8 = v5 + v6;
		v9 = v2[2];
		v10 = v2[4];
		v19[3] = v7;
		v11 = *(uint32_t*)(a1 + 8);
		v12 = -1;
		v19[0] = v9 + v10 - v11;
		v13 = *(uint32_t*)(a1 + 16) + v11 - v9;
		v19[2] = v8;
		v19[1] = v13;
		v14 = 99999;
		v15 = 0;
		v16 = v19;
		do {
			if (*v16 < v14) {
				v14 = *v16;
				v12 = v15;
			}
			++v15;
			++v16;
		} while (v15 < 4);
		switch (v12) {
		case 0:
			a2a.field_0 = *(float*)(a1 + 20);
			a2a.field_4 = (double)v19[0] * 32.526913 + *(float*)(a1 + 24);
			break;
		case 1:
			a2a.field_0 = *(float*)(a1 + 20);
			a2a.field_4 = *(float*)(a1 + 24) - (double)v19[1] * 32.526913;
			break;
		case 2:
			a2a.field_4 = *(float*)(a1 + 24);
			a2a.field_0 = (double)v19[2] * 32.526913 + *(float*)(a1 + 20);
			break;
		case 3:
			a2a.field_4 = *(float*)(a1 + 24);
			a2a.field_0 = *(float*)(a1 + 20) - (double)v19[3] * 32.526913;
			break;
		default:
			break;
		}
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)a1, &a2a);
		v2 = (uint32_t*)nox_mapgenRoomAtNative_521200(a1);
		if (!v2) {
			return 1;
		}
		if (++v3 >= 25) {
			return 0;
		}
	}
}

//----- (005213E0) --------------------------------------------------------
int nox_xxx_mapgenAllocBuffer_5213E0() {
	dword_5d4594_2487556 = calloc(1, 0x2000u);
	return dword_5d4594_2487556 != 0;
}

//----- (00521400) --------------------------------------------------------
void nox_xxx_mapgenFreeBuffer_521400() { free(*(void**)&dword_5d4594_2487556); }

//----- (00521710) --------------------------------------------------------
void* nox_xxx_mapGenGetTopRoom_521710() { return nox_mapgen_room_head; }

void* nox_mapgenRoomNextNative_521720(const void* a1) {
	if (!a1) {
		return NULL;
	}
	return nox_mapgenLegacyPtrResolve(*(const uint32_t*)((const uint8_t*)a1 + 56));
}

//----- (00521720) --------------------------------------------------------
int sub_521720(int a1) {
	int result; // eax

	if (a1) {
		result = (int)nox_mapgenLegacyPtrRegister(nox_mapgenRoomNextNative_521720(
			(void*)nox_mapgenLegacyPtrResolve((uint32_t)a1)));
	} else {
		result = 0;
	}
	return result;
}

//----- (00521730) --------------------------------------------------------
int nox_xxx_mapGenAddNewRoom_521730(uint32_t* a1) {
	uint32_t token = nox_mapgenLegacyPtrRegister(a1);
	uint32_t head_token = nox_mapgenLegacyPtrRegister(nox_mapgen_room_head);
	if (!token) {
		return 0;
	}
	a1[15] = 0;
	a1[14] = head_token;
	if (nox_mapgen_room_head) {
		*(uint32_t*)(nox_mapgen_room_head + 60) = token;
	}
	nox_mapgen_room_head = (uint8_t*)a1;
	dword_5d4594_2487560 = token;
	return nox_mapgenRoomOccupyNative_521100((uint8_t*)a1);
}

//----- (00521760) --------------------------------------------------------
int nox_mapgenRemoveRoomNative_521760(uint8_t* room) {
	if (!room) {
		return 0;
	}
	uint8_t* v1;
	uint8_t* v2;

	v1 = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room + 60));
	if (v1) {
		*(uint32_t*)(v1 + 56) = *(uint32_t*)(room + 56);
	} else {
		nox_mapgen_room_head = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room + 56));
		dword_5d4594_2487560 = nox_mapgenLegacyPtrRegister(nox_mapgen_room_head);
	}
	v2 = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room + 56));
	if (v2) {
		*(uint32_t*)(v2 + 60) = *(uint32_t*)(room + 60);
	}
	return nox_mapgenRoomVacateNative_521180(room);
}

int sub_521760(int a1) {
	return nox_mapgenRemoveRoomNative_521760(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (005217A0) --------------------------------------------------------
int sub_5217A0(int a1, int a2) {
	return nox_mapgenRoomWithinThemeNative_5217A0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

int nox_mapgenRoomWithinThemeNative_5217A0(uint8_t* a1, uint8_t* a2) {
	int v2;         // edi
	int v3;         // ebx
	long double v4; // st7
	int v5;         // edx
	long double v6; // st6
	int v7;         // ecx

	v2 = 0;
	v3 = *(uint32_t*)(a2 + 16);
	v4 = *(float*)(a2 + 24);
	if (v3 <= 0) {
		return 1;
	}
	v5 = *(uint32_t*)(a2 + 12);
	while (1) {
		v6 = *(float*)(a2 + 20);
		v7 = 0;
		if (v5 > 0) {
			while (*(float*)(a1 + 64) >= fabs(v6) && *(float*)(a1 + 64) >= fabs(v4)) {
				v6 = v6 + 32.526913;
				if (++v7 >= v5) {
					goto LABEL_7;
				}
			}
			return 0;
		}
	LABEL_7:
		++v2;
		v4 = v4 + 32.526913;
		if (v2 >= v3) {
			return 1;
		}
	}
}

//----- (00521820) --------------------------------------------------------
int nox_mapgenRoomFitsNative_521820(uint8_t* theme, uint8_t* room) {
	return nox_mapgenRoomWithinThemeNative_5217A0(theme, room) &&
		nox_mapgenRoomAtNative_521200(room) == NULL;
}

int sub_521820(int a1, int a2) {
	return nox_mapgenRoomFitsNative_521820(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00521850) --------------------------------------------------------
int nox_xxx_mapGenUpdateRoomRect_521850(uint8_t* a1) {
	double v2;  // st7
	int v3;     // edx

	v2 = *(float*)(a1 + 28) + *(float*)(a1 + 20);
	v3 = *(uint32_t*)(a1 + 24);
	*(uint32_t*)(a1 + 36) = *(uint32_t*)(a1 + 20);
	*(uint32_t*)(a1 + 40) = v3;
	*(float*)(a1 + 44) = v2;
	*(float*)(a1 + 48) = *(float*)(a1 + 32) + *(float*)(a1 + 24);
	return (int)(uint32_t)(uintptr_t)a1;
}

//----- (00521880) --------------------------------------------------------
int nox_xxx_mapGenSetRoomPos_521880(uint32_t* a1, float2* a2) {
	*(float2*)(a1 + 5) = *a2;
	nox_xxx_mapGenRoundFloatToPtr_520DF0(a2, a1 + 1);
	return nox_xxx_mapGenUpdateRoomRect_521850((uint8_t*)a1);
}

int nox_mapgenRoomHasTypeOneNeighborNative_5218B0(uint8_t* room, int direction) {
	if (!room || direction < 0 || direction >= 4) {
		return 0;
	}
	int count = room[216 + direction];
	for (int i = 0; i < count; ++i) {
		uint32_t token = *(uint32_t*)(room + 88 + 32 * direction + 4 * i);
		uint32_t* neighbor = (uint32_t*)nox_mapgenLegacyPtrResolve(token);
		if (neighbor && neighbor[0] == 1) {
			return 1;
		}
	}
	return 0;
}

//----- (005218B0) --------------------------------------------------------
int sub_5218B0(int a1, int a2) {
	return nox_mapgenRoomHasTypeOneNeighborNative_5218B0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

int nox_mapgenRoomLinkNative_521900(uint8_t* room, uint8_t* neighbor, int direction) {
	if (!room || !neighbor || direction < 0 || direction >= 4 || room[216 + direction] >= 8u) {
		return 0;
	}
	uint32_t token = nox_mapgenLegacyPtrRegister(neighbor);
	if (!token) {
		return 0;
	}
	uint8_t index = room[216 + direction]++;
	*(uint32_t*)(room + 88 + 32 * direction + 4 * index) = token;
	return 1;
}

//----- (00521900) --------------------------------------------------------
int sub_521900(int a1, int a2, int a3) {
	return nox_mapgenRoomLinkNative_521900(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2), a3);
}

//----- (00521940) --------------------------------------------------------
float* nox_xxx_mapGenMakeRoomStruct_521940(int a1, int a2) {
	float* result; // eax

	result = (float*)calloc(1u, 0x178u);
	if (result) {
		if (!nox_mapgenLegacyPtrRegister(result)) {
			free(result);
			return NULL;
		}
		*(uint32_t*)result = 1;
		*((uint32_t*)result + 3) = a1;
		*((uint32_t*)result + 4) = a2;
		result[7] = (double)a1 * 32.526913;
		result[8] = (double)a2 * 32.526913;
	}
	return result;
}

//----- (00521990) --------------------------------------------------------
float* nox_mapgenPrepareRoomNative_521990(uint8_t* theme) {
	int v2;        // eax
	int v3;        // esi
	int v4;        // edi
	int v5;        // ebx
	signed int v7; // [esp-4h] [ebp-14h]
	int v8;        // [esp+14h] [ebp+4h]

	if (!theme) {
		return NULL;
	}
	v2 = *(uint32_t*)(theme + 32);
	v7 = *(uint32_t*)(theme + 36);
	v3 = v2 - v7;
	v4 = nox_xxx_mapGenRandFunc2_526B00(v2, v7);
	v8 = v4;
	if (v4 < 5) {
		v4 = 5;
		v8 = 5;
	}
	v5 = nox_xxx_mapGenRandFunc2_526B00(v4, (long long)((double)v8 * 0.30000001));
	if (v5 < 5) {
		v5 = 5;
	}
	if (v5 < v3) {
		v5 = v3;
	}
	if (*(uint32_t*)(theme + 56)) {
		nox_xxx_mapGenRandFunc_526AC0(0, 2);
	}
	return nox_xxx_mapGenMakeRoomStruct_521940(v4, v5);
}

float* nox_xxx_mapGenPrepareRoom_521990(int a1) {
	return nox_mapgenPrepareRoomNative_521990(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (00521A10) --------------------------------------------------------
void sub_521A10(void* lpMem) {
	uint32_t token = *((uint32_t*)lpMem + 92);
	while (token) {
		nox_mapgen_occupied_rect_legacy* rect = nox_mapgenOccupiedRectResolve(token);
		if (!rect) {
			break;
		}
		token = rect->next_token;
		nox_mapgenLegacyPtrForget(rect);
		free(rect);
	}
	*((uint32_t*)lpMem + 92) = 0;
	nox_mapgenLegacyPtrForget(lpMem);
	free(lpMem);
}

//----- (00521A40) --------------------------------------------------------
uint32_t* nox_xxx_mapGenFreeTopRoom_521A40() {
	uint32_t* result; // eax
	uint32_t* v1;     // esi

	result = (uint32_t*)nox_mapgen_room_head;
	if (nox_mapgen_room_head) {
		do {
			v1 = (uint32_t*)nox_mapgenLegacyPtrResolve(result[14]);
			sub_521A10(result);
			result = v1;
		} while (v1);
		nox_mapgen_room_head = NULL;
		dword_5d4594_2487560 = 0;
	} else {
		nox_mapgen_room_head = NULL;
		dword_5d4594_2487560 = 0;
	}
	nox_mapgenResetRoomRankingNative();
	return result;
}

int nox_mapgenRoomLinkBothNative_521A70(uint8_t* room, uint8_t* neighbor, int direction) {
	nox_mapgenRoomLinkNative_521900(room, neighbor, direction);
	return nox_mapgenRoomLinkNative_521900(neighbor, room, sub_523960(direction));
}

//----- (00521A70) --------------------------------------------------------
int sub_521A70(int a1, int a2, int a3) {
	return nox_mapgenRoomLinkBothNative_521A70(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2), a3);
}

//----- (00521AA0) --------------------------------------------------------
int sub_521AA0(uint32_t* a1, int a2) {
	int v2;     // eax

	if (*a1 == 1) {
		return 0;
	}
	switch (*a1) {
	case 2:
		v2 = 1;
		break;
	case 3:
		v2 = 0;
		break;
	case 4:
		v2 = 3;
		break;
	case 5:
		v2 = 2;
		break;
	default:
		v2 = (int)a1;
		break;
	}
	if (a2 != v2) {
		return 0;
	} else {
		return 1;
	}
}

//----- (00521B00) --------------------------------------------------------
double nox_mapgenRoomRandPosXNative_521B00(const uint8_t* room, const uint8_t* child) {
	return (double)nox_xxx_mapGenRandFunc_526AC0(
			   0, *(const uint32_t*)(room + 12) - *(const uint32_t*)(child + 12)) *
			32.526913 +
		*(const float*)(room + 20);
}

double sub_521B00(int a1, int a2) {
	return nox_mapgenRoomRandPosXNative_521B00(
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00521B30) --------------------------------------------------------
double nox_mapgenRoomRandPosYNative_521B30(const uint8_t* room, const uint8_t* child) {
	return (double)nox_xxx_mapGenRandFunc_526AC0(
			   0, *(const uint32_t*)(room + 16) - *(const uint32_t*)(child + 16)) *
			32.526913 +
		*(const float*)(room + 24);
}

double sub_521B30(int a1, int a2) {
	return nox_mapgenRoomRandPosYNative_521B30(
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00521B60) --------------------------------------------------------
double nox_mapgenRoomAttachPosXNative_521B60(const uint8_t* room, const uint8_t* parent) {
	return *(const float*)(parent + 20) -
		(double)nox_xxx_mapGenRandFunc_526AC0(
			0, *(const uint32_t*)(room + 12) - *(const uint32_t*)(parent + 12)) *
			32.526913;
}

double sub_521B60(int a1, int a2) {
	return nox_mapgenRoomAttachPosXNative_521B60(
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00521B90) --------------------------------------------------------
double nox_mapgenRoomAttachPosYNative_521B90(const uint8_t* room, const uint8_t* parent) {
	return *(const float*)(parent + 24) -
		(double)nox_xxx_mapGenRandFunc_526AC0(
			0, *(const uint32_t*)(room + 16) - *(const uint32_t*)(parent + 16)) *
			32.526913;
}

double sub_521B90(int a1, int a2) {
	return nox_mapgenRoomAttachPosYNative_521B90(
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(const uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00521BC0) --------------------------------------------------------
float* nox_mapgenAddOccupiedRectNative_521BC0(uint8_t* room, float2* pos, float width, float height) {
	if (!room || !pos) {
		return NULL;
	}
	nox_mapgen_occupied_rect_legacy* rect = calloc(1u, sizeof(*rect));
	if (!rect) {
		return NULL;
	}
	uint32_t token = nox_mapgenLegacyPtrRegister(rect);
	if (!token) {
		free(rect);
		return NULL;
	}
	rect->min_x = pos->field_0;
	rect->min_y = pos->field_4;
	rect->max_x = width + pos->field_0;
	rect->max_y = height + pos->field_4;
	rect->next_token = *(uint32_t*)(room + 368);
	*(uint32_t*)(room + 368) = token;
	return (float*)rect;
}

float* sub_521BC0(int a1, float2* a2, float a3, float a4) {
	return nox_mapgenAddOccupiedRectNative_521BC0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2, a3, a4);
}

//----- (00521C10) --------------------------------------------------------
uint32_t* nox_mapgenClearTransientOccupiedRectsNative_521C10(uint8_t* room) {
	if (!room) {
		return NULL;
	}
	nox_mapgen_occupied_rect_legacy* previous = NULL;
	uint32_t token = *(uint32_t*)(room + 368);
	while (token) {
		nox_mapgen_occupied_rect_legacy* rect = nox_mapgenOccupiedRectResolve(token);
		if (!rect) {
			break;
		}
		uint32_t next_token = rect->next_token;
		if (rect->transient == 1) {
			if (previous) {
				previous->next_token = next_token;
			} else {
				*(uint32_t*)(room + 368) = next_token;
			}
			nox_mapgenLegacyPtrForget(rect);
			free(rect);
		} else {
			previous = rect;
		}
		token = next_token;
	}
	return NULL;
}

uint32_t* sub_521C10(int a1) {
	return nox_mapgenClearTransientOccupiedRectsNative_521C10(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (00521C60) --------------------------------------------------------
uint32_t* nox_mapgenFindForeachChoicesNative_521C60(uintptr_t foreach_ref, uint16_t type_id) {
	for (nox_mapgen_foreach_legacy* entry = nox_mapgenForeachResolve(foreach_ref); entry;
		 entry = nox_mapgenForeachResolve(entry->next_token)) {
		if (entry->type_id == type_id) {
			return nox_mapgenChoiceResolve(entry->choices_token);
		}
	}
	return NULL;
}

void nox_mapgenApplyForeachNative_521C60(uint8_t* theme, uintptr_t foreach_ref) {
	for (nox_object_t* object = sub_504980(); object; object = sub_5049C0(object)) {
		uint32_t* choices = nox_mapgenFindForeachChoicesNative_521C60(foreach_ref, object->typ_ind);
		if (choices) {
			nox_mapgenApplyChoiceNative_521FE0(theme, object, (uintptr_t)choices);
		}
	}
}

int nox_xxx_mapgen_521C60(int a1, int a2) {
	nox_mapgenApplyForeachNative_521C60(
		(uint8_t*)(uintptr_t)(uint32_t)a1, (uintptr_t)(uint32_t)a2);
	return 0;
}

//----- (00521CB0) --------------------------------------------------------
int nox_mapgenPlaceDecorObjectNative_521CB0(
	uint8_t* theme, uint8_t* room, uint8_t* decor_entry, int object_index) {
	if (!theme || !room || !decor_entry) {
		return 0;
	}
	nox_mapgen_occupied_rect_legacy* rect = calloc(1u, sizeof(*rect));
	if (!rect) {
		return 0;
	}
	uint32_t rect_token = nox_mapgenLegacyPtrRegister(rect);
	if (!rect_token) {
		free(rect);
		return 0;
	}

	float width = sub_502E70(object_index);
	float height = sub_502EA0(object_index);
	int width_cells = (long long)(width * 0.030743772);
	int height_cells = (long long)(height * 0.030743772);
	int horizontal_range = *(uint32_t*)(room + 12) - width_cells;
	int vertical_range = *(uint32_t*)(room + 16) - height_cells;
	if (horizontal_range >= 0 && vertical_range >= 0) {
		char* object_record = sub_526AA0(object_index);
		if (object_record) {
			float2 position;
			position.field_0 = (double)nox_xxx_mapGenRandFunc_526AC0(0, horizontal_range) * 32.526913 + *(float*)(room + 36);
			position.field_4 = (double)nox_xxx_mapGenRandFunc_526AC0(0, vertical_range) * 32.526913 + *(float*)(room + 40);
			uint32_t flags = *((uint32_t*)object_record + 15);
			if (flags & 1) {
				position.field_4 = *(float*)(room + 40);
			} else if (flags & 2) {
				position.field_4 = *(float*)(room + 48) - height;
			}
			if (flags & 4) {
				position.field_0 = *(float*)(room + 44) - width;
			} else if (flags & 8) {
				position.field_0 = *(float*)(room + 36);
			}
			if (object_record[60] & 0x10) {
				position.field_0 = (double)(int)(horizontal_range / 2) * 32.526913 + *(float*)(room + 36);
				position.field_4 = (double)(vertical_range / 2) * 32.526913 + *(float*)(room + 40);
			}

			rect->transient = 1;
			rect->min_x = position.field_0;
			rect->min_y = position.field_4;
			rect->max_x = position.field_0 + width;
			rect->max_y = position.field_4 + height;
			rect->object_index = object_index;
			if (sub_521EB0((float*)room, (float*)rect) &&
				!nox_mapgenOccupiedRectsIntersectNative_521F10(room, (float*)rect)) {
				sub_502D70(object_index);
				nox_mapgenApplyForeachNative_521C60(theme, *(uint32_t*)(decor_entry + 84));
				sub_503B30(&position);
				rect->next_token = *(uint32_t*)(room + 368);
				*(uint32_t*)(room + 368) = rect_token;
				return 1;
			}
		}
	}
	nox_mapgenLegacyPtrForget(rect);
	free(rect);
	return 0;
}

int sub_521CB0(int a1, int a2, int a3, int a4) {
	return nox_mapgenPlaceDecorObjectNative_521CB0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a3), a4);
}

//----- (00521EB0) --------------------------------------------------------
int sub_521EB0(float* a1, float* a2) {
	return a2[1] + 0.5 >= a1[9] && a2[3] - 0.5 <= a1[11] && a2[2] + 0.5 >= a1[10] && a2[4] - 0.5 <= a1[12];
}

//----- (00521F10) --------------------------------------------------------
int nox_mapgenOccupiedRectsIntersectNative_521F10(uint8_t* room, float* a2) {
	float v4; // [esp+0h] [ebp-20h]
	float v5; // [esp+4h] [ebp-1Ch]
	float v6; // [esp+8h] [ebp-18h]
	float v7; // [esp+Ch] [ebp-14h]
	float v8; // [esp+14h] [ebp-Ch]
	float v9; // [esp+1Ch] [ebp-4h]

	if (!room || !a2) {
		return 0;
	}
	uint32_t token = *(uint32_t*)(room + 368);
	while (token) {
		nox_mapgen_occupied_rect_legacy* rect = nox_mapgenOccupiedRectResolve(token);
		if (!rect) {
			return 0;
		}
		if ((float*)rect != a2) {
			v4 = a2[1] + 0.5;
			if (v4 < rect->max_x - 0.5) {
				v6 = a2[3] - 0.5;
				if (v6 > rect->min_x + 0.5) {
					v9 = rect->max_y - 0.5;
					v5 = a2[2] + 0.5;
					if (v5 < (double)v9) {
						v8 = rect->min_y + 0.5;
						v7 = a2[4] - 0.5;
						if (v7 > (double)v8) {
							return 1;
						}
					}
				}
			}
		}
		token = rect->next_token;
	}
	return 0;
}

int sub_521F10(int a1, float* a2) {
	return nox_mapgenOccupiedRectsIntersectNative_521F10(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (00521FE0) --------------------------------------------------------
static void nox_mapgenApplyChoiceNative_521FE0(
	uint8_t* theme, nox_object_t* holder, uintptr_t choices_ref) {
	uint32_t* choice = nox_mapgenChoiceResolve(choices_ref);
	if (!choice || !theme || !holder) {
		return;
	}
	int roll = nox_xxx_mapGenRandFunc_526AC0(1, 100);
	while (choice) {
		roll -= (int)choice[0];
		if (roll <= 0) {
			break;
		}
		choice = nox_mapgenChoiceNextNative_520380((uintptr_t)choice);
	}
	if (!choice) {
		return;
	}

	char* entry = (char*)(choice + 2);
	for (uint32_t i = 0; i < choice[NOX_MAPGEN_CHOICE_COUNT_INDEX]; ++i, entry += 64) {
		nox_object_t* item = NULL;
		switch (*((uint32_t*)entry - 1)) {
		case 0:
			item = nox_xxx_newObjectByTypeID_4E3810(entry);
			break;
		case 3:
			item = nox_mapgenMakeEnchantedItemNative_5221A0(
				entry, *(uint32_t*)(theme + 1100), *(uint32_t*)(theme + 1104));
			break;
		case 4:
			item = nox_mapgenMakeEnchantedItemNative_5221A0(
				entry, *(uint32_t*)(theme + 1108), *(uint32_t*)(theme + 1112));
			break;
		case 5:
			item = nox_mapgenMakeSpellbookNative_5220E0(theme, entry);
			break;
		default:
			break;
		}
		if (item) {
			nox_mapgenAttachInventoryNative_522300(holder, item);
		}
	}
}

void nox_xxx_mapgen_521FE0(int a1, int a2, uint32_t* a3) {
	nox_mapgenApplyChoiceNative_521FE0(
		(uint8_t*)(uintptr_t)(uint32_t)a1,
		(nox_object_t*)(uintptr_t)(uint32_t)a2,
		(uintptr_t)a3);
}

//----- (005220E0) --------------------------------------------------------
static nox_object_t* nox_mapgenMakeSpellbookNative_5220E0(uint8_t* theme, const char* spell_name) {
	uint32_t spell_count = *(uint32_t*)(theme + 1096);
	if (!spell_count) {
		return NULL;
	}
	nox_object_t* object = nox_xxx_newObjectByTypeID_4E3810("SpellBook");
	if (!object) {
		return NULL;
	}
	char spell_id;
	if (!strcmp("*", spell_name)) {
		spell_id = *(uint8_t*)(theme + 4 * nox_xxx_mapGenRandFunc_526AC0(0, spell_count - 1) + 548);
	} else {
		spell_id = nox_xxx_mapGenSpellIdByName_51E1D0(spell_name);
		if (!spell_id) {
			free(object);
			return NULL;
		}
	}
	nox_mapgenFinishSpellbookNative_527DB0(object, spell_id);
	return object;
}

uint32_t* nox_xxx_mapGenMakeSpellbook_5220E0(int a1, const char* a2) {
	return (uint32_t*)nox_mapgenMakeSpellbookNative_5220E0(
		(uint8_t*)(uintptr_t)(uint32_t)a1, a2);
}

//----- (005221A0) --------------------------------------------------------
static nox_object_t* nox_mapgenMakeEnchantedItemNative_5221A0(
	char* name, uintptr_t list_ref, int count) {
	char* definition = (char*)nox_mapgenLegacyPtrResolve(list_ref);
	char* object_name = name;
	if (!strcmp("*", name)) {
		if (!definition || count <= 0) {
			return NULL;
		}
		int selected = nox_xxx_mapGenRandFunc_526AC0(0, count - 1);
		for (int i = 0; definition && i < selected; ++i) {
			definition = (char*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(definition + 152));
		}
		if (!definition) {
			return NULL;
		}
		object_name = definition + 60;
	} else {
		for (; definition; definition = (char*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(definition + 152))) {
			if (!nox_strcmpi(definition, name)) {
				object_name = definition + 60;
				break;
			}
		}
	}

	nox_object_t* object = nox_xxx_newObjectByTypeID_4E3810(object_name);
	if (object && definition) {
		nox_modifier_attrs_t attrs;
		memset(attrs.modifiers, 0, sizeof(attrs.modifiers));
		for (int index = 0; index < 4; ++index) {
			uint32_t modifier_count = *(uint32_t*)(definition + 136 + 4 * index);
			if (!modifier_count ||
				nox_xxx_mapGenRandFunc_526AC0(1, 100) > *getMemU32Ptr(0x587000, 254688 + 4 * index)) {
				continue;
			}
			char* modifiers = (char*)nox_mapgenLegacyPtrResolve(
				*(uint32_t*)(definition + 120 + 4 * index));
			if (!modifiers) {
				continue;
			}
			const char* modifier = modifiers +
				60 * nox_xxx_mapGenRandFunc_526AC0(0, modifier_count - 1);
			if (nox_strcmpi("none", modifier)) {
				int modifier_id = nox_xxx_modifGetIdByName_413290(modifier);
				attrs.modifiers[index] = nox_xxx_modifGetDescById_413330(modifier_id);
			}
		}
		if (!attrs.modifiers[2]) {
			attrs.modifiers[3] = NULL;
		}
		nox_xxx_modifSetItemAttrs_4E4990(object, &attrs);
	}
	return object;
}

uint32_t* nox_xxx_mapGenMakeEnchantedItem_5221A0(char* a1, char* a2, int a3) {
	return (uint32_t*)nox_mapgenMakeEnchantedItemNative_5221A0(a1, (uintptr_t)a2, a3);
}

//----- (00522300) --------------------------------------------------------
static nox_object_t* nox_mapgenAttachInventoryNative_522300(
	nox_object_t* holder, nox_object_t* item) {
	item->field_125 = NULL;
	item->inv_next_item = holder->inv_first_item;
	if (holder->inv_first_item) {
		holder->inv_first_item->field_125 = item;
	}
	holder->inv_first_item = item;
	item->inv_holder = holder;
	return item;
}

uint32_t* sub_522300(int a1, uint32_t* a2) {
	return (uint32_t*)nox_mapgenAttachInventoryNative_522300(
		(nox_object_t*)(uintptr_t)(uint32_t)a1, (nox_object_t*)a2);
}

static int nox_mapgenFindFreePointNative_5226D0(uint8_t* room, float scale, float2* point) {
	if (!room || !point) {
		return 0;
	}
	float center_x = (*(float*)(room + 36) + *(float*)(room + 44)) * 0.5f;
	float center_y = (*(float*)(room + 40) + *(float*)(room + 48)) * 0.5f;
	float half_width = (*(float*)(room + 44) - *(float*)(room + 36)) * 0.5f;
	float half_height = (*(float*)(room + 48) - *(float*)(room + 40)) * 0.5f;
	for (int attempt = 0; attempt < 10; ++attempt) {
		point->field_0 = sub_526BC0(-half_width, half_width) * scale + center_x;
		point->field_4 = sub_526BC0(-half_height, half_height) * scale + center_y;
		if (!nox_mapgenPointOccupiedNative_5227B0(room, (float*)point)) {
			return 1;
		}
	}
	return 0;
}

static nox_object_t* nox_mapgenMakeMonsterInRoomNative_522810(uint8_t* room, const char* name) {
	if (!room || !name) {
		return NULL;
	}
	float center_x = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5f;
	float center_y = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5f;
	float2 position;
	if (!nox_mapgenFindFreePointNative_5226D0(room, 0.94999999f, &position)) {
		return NULL;
	}
	nox_xxx_mapGenGetObjID_527940((char*)name);
	nox_object_t* object = (nox_object_t*)nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object && ((uint8_t)((float*)object)[2] & 2)) {
		float2 direction = {center_x - position.field_0, center_y - position.field_4};
		nox_mapgenOrientObjNative_527C60(
			object, nox_xxx_math_509EA0(nox_xxx_math_509ED0(&direction)));
	}
	return object;
}

static void* nox_mapgenPopulateDecorEntryNative_5224B0(
	uint8_t* theme, uint8_t* room, nox_mapgen_decor_entry_legacy* entry) {
	if (!theme || !room || !entry) {
		return NULL;
	}
	int count;
	if (entry->density_enabled) {
		count = (long long)((double)(*(int32_t*)(room + 12) * *(int32_t*)(room + 16)) * entry->density);
		if (count < entry->min_count) {
			count = entry->min_count;
		} else if (count > entry->max_count) {
			count = entry->max_count;
		}
	} else {
		count = nox_xxx_mapGenRandFunc_526AC0(entry->min_count, entry->max_count);
	}

	void* result = NULL;
	switch (entry->type) {
	case 0:
		for (int i = 0; i < count; ++i) {
			nox_object_t* object = nox_mapgenMakeMonsterInRoomNative_522810(room, entry->name);
			if (object) {
				nox_mapgenApplyChoiceNative_521FE0(theme, object, entry->choices_token);
				result = object;
			}
		}
		break;
	case 1: {
		int object_index = sub_5268F0(entry->name);
		if (object_index >= 0) {
			for (int i = 0; i < count; ++i) {
				for (int attempt = 0; attempt < 3; ++attempt) {
					if (nox_mapgenPlaceDecorObjectNative_521CB0(
							theme, room, (uint8_t*)entry, object_index)) {
						break;
					}
				}
			}
		}
		break;
	}
	case 2:
		result = nox_mapgenClearTransientOccupiedRectsNative_521C10(room);
		break;
	case 3:
	case 4:
	case 5: {
		float2 position;
		if (!nox_mapgenFindFreePointNative_5226D0(room, 0.89999998f, &position)) {
			break;
		}
		nox_object_t* object = NULL;
		if (entry->type == 3) {
			object = nox_mapgenMakeEnchantedItemNative_5221A0(
				entry->name, *(uint32_t*)(theme + 1100), *(uint32_t*)(theme + 1104));
		} else if (entry->type == 4) {
			object = nox_mapgenMakeEnchantedItemNative_5221A0(
				entry->name, *(uint32_t*)(theme + 1108), *(uint32_t*)(theme + 1112));
		} else {
			object = nox_mapgenMakeSpellbookNative_5220E0(theme, entry->name);
		}
		if (object) {
			result = nox_xxx_mapGenMoveObject_527A10((float*)object, (float*)&position);
		}
		break;
	}
	default:
		break;
	}
	return result;
}

static void nox_mapgenPopulateDecorSetNative_522370(
	uint8_t* theme, uint8_t* room, nox_mapgen_decor_set_legacy* set) {
	if (!theme || !room || !set || !set->entries_token) {
		return;
	}
	int count = nox_xxx_mapGenRandFunc_526AC0(set->min_count, set->max_count);
	if (count > (int)set->entry_count) {
		count = set->entry_count;
	}

	uint32_t head_token = set->entries_token;
	uint32_t tail_token = head_token;
	nox_mapgen_decor_entry_legacy* first = nox_mapgenDecorEntryResolve(head_token);
	if (!first) {
		return;
	}
	first->shuffle_next_token = 0;
	first->shuffle_prev_token = 0;
	uint32_t current_token = first->next_token;
	for (uint32_t processed = 1; current_token && processed < set->entry_count; ++processed) {
		nox_mapgen_decor_entry_legacy* current = nox_mapgenDecorEntryResolve(current_token);
		if (!current) {
			break;
		}
		uint32_t next_token = current->next_token;
		nox_mapgen_decor_entry_legacy* head = nox_mapgenDecorEntryResolve(head_token);
		nox_mapgen_decor_entry_legacy* tail = nox_mapgenDecorEntryResolve(tail_token);
		if (!head || !tail) {
			break;
		}
		if (nox_xxx_mapGenRandFunc_526AC0(1, 100) >= 50) {
			if (nox_xxx_mapGenRandFunc_526AC0(1, 100) >= 50) {
				tail->shuffle_next_token = current_token;
				current->shuffle_prev_token = tail_token;
				current->shuffle_next_token = 0;
				tail_token = current_token;
			} else {
				uint32_t previous_token = tail->shuffle_prev_token;
				current->shuffle_next_token = tail_token;
				current->shuffle_prev_token = previous_token;
				if (previous_token) {
					nox_mapgen_decor_entry_legacy* previous = nox_mapgenDecorEntryResolve(previous_token);
					if (previous) {
						previous->shuffle_next_token = current_token;
					}
				} else {
					head_token = current_token;
				}
				tail->shuffle_prev_token = current_token;
			}
		} else if (nox_xxx_mapGenRandFunc_526AC0(1, 100) >= 50) {
			uint32_t following_token = head->shuffle_next_token;
			current->shuffle_prev_token = head_token;
			current->shuffle_next_token = following_token;
			if (following_token) {
				nox_mapgen_decor_entry_legacy* following = nox_mapgenDecorEntryResolve(following_token);
				if (following) {
					following->shuffle_prev_token = current_token;
				}
			} else {
				tail_token = current_token;
			}
			head->shuffle_next_token = current_token;
		} else {
			current->shuffle_next_token = head_token;
			current->shuffle_prev_token = 0;
			head->shuffle_prev_token = current_token;
			head_token = current_token;
		}
		current_token = next_token;
	}

	for (uint32_t token = head_token; token && count-- > 0;) {
		nox_mapgen_decor_entry_legacy* entry = nox_mapgenDecorEntryResolve(token);
		if (!entry) {
			break;
		}
		nox_mapgenPopulateDecorEntryNative_5224B0(theme, room, entry);
		token = entry->shuffle_next_token;
	}
}

void nox_mapgenPopulateDecorNative_522340(uint8_t* theme, uint8_t* room) {
	if (!theme || !room) {
		return;
	}
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve(*(uint32_t*)(room + 372));
	if (!definition) {
		return;
	}
	for (uint32_t token = definition->decor_sets_token; token;) {
		nox_mapgen_decor_set_legacy* set = nox_mapgenDecorSetResolve(token);
		if (!set) {
			break;
		}
		nox_mapgenPopulateDecorSetNative_522370(theme, room, set);
		token = set->next_token;
	}
}

//----- (00522340) --------------------------------------------------------
void nox_xxx_mapgen_522340(int a1, int a2) {
	nox_mapgenPopulateDecorNative_522340(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00522370) --------------------------------------------------------
void nox_xxx_mapgen_522370(int a1, int a2, int* a3) {
	nox_mapgenPopulateDecorSetNative_522370(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2),
		nox_mapgenDecorSetResolve((uintptr_t)a3));
}

//----- (005224B0) --------------------------------------------------------
float* nox_xxx_mapgen_5224B0(int a1, int a2, int a3) {
	return (float*)nox_mapgenPopulateDecorEntryNative_5224B0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2),
		nox_mapgenDecorEntryResolve((uint32_t)a3));
}

//----- (005226D0) --------------------------------------------------------
int sub_5226D0(int a1, float a2, int a3) {
	return nox_mapgenFindFreePointNative_5226D0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2,
		(float2*)nox_mapgenLegacyPtrResolve((uint32_t)a3));
}

//----- (005227B0) --------------------------------------------------------
int nox_mapgenPointOccupiedNative_5227B0(uint8_t* room, float* point) {
	if (!room || !point) {
		return 0;
	}
	for (uint32_t token = *(uint32_t*)(room + 368); token;) {
		nox_mapgen_occupied_rect_legacy* rect = nox_mapgenOccupiedRectResolve(token);
		if (!rect) {
			return 0;
		}
		if (*point >= (double)rect->min_x && *point <= (double)rect->max_x &&
			point[1] >= (double)rect->min_y && point[1] <= (double)rect->max_y) {
			return 1;
		}
		token = rect->next_token;
	}
	return 0;
}

int sub_5227B0(int a1, float* a2) {
	return nox_mapgenPointOccupiedNative_5227B0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (00522810) --------------------------------------------------------
int nox_xxx_mapGenMakeMonsterInRoom_522810(float* a1, char* a2) {
	nox_object_t* object = nox_mapgenMakeMonsterInRoomNative_522810(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uintptr_t)a1), a2);
	return (int)nox_mapgenLegacyPtrRegister(object);
}

//----- (00522C80) --------------------------------------------------------
float* sub_522C80(float* a1) {
	float* result; // eax

	result = sub_51D1A0((float2*)a1);
	if (!result) {
		result = (float*)sub_51D120(a1);
	}
	return result;
}

//----- (00522CA0) --------------------------------------------------------
char nox_mapgenAddRoomPointNative_522CA0(uint8_t* room, const float2* point) {
	if (!room || !point) {
		return 0;
	}
	uint8_t count = room[352];
	if (count >= 16u) {
		return (char)count;
	}
	for (uint8_t i = 0; i < count; ++i) {
		float2* current = (float2*)(room + 224 + 8 * i);
		if (sub_524660(current->field_0, point->field_0) &&
			sub_524660(current->field_4, point->field_4)) {
			return 1;
		}
	}
	*(float2*)(room + 224 + 8 * count) = *point;
	room[352] = count + 1;
	return (char)room[352];
}

char sub_522CA0(int a1, float* a2) {
	return nox_mapgenAddRoomPointNative_522CA0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), (const float2*)a2);
}

//----- (00522D30) --------------------------------------------------------
float* nox_mapgenConnectRoomLinesNative_522D30(uint8_t* theme) {
	sub_51D0F0(128);
	for (uint8_t* room = (uint8_t*)nox_xxx_mapGenGetTopRoom_521710(); room;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		nox_xxx_mapGenSetFlags_5235F0(156);
		if (*(uint32_t*)room == 1) {
			for (int direction = 0; direction < 4; ++direction) {
				int count = room[216 + direction];
				if (count > 8) {
					count = 8;
				}
				for (int index = 0; index < count; ++index) {
					uint32_t token = *(uint32_t*)(room + 88 + 32 * direction + 4 * index);
					uint8_t* neighbor = (uint8_t*)nox_mapgenLegacyPtrResolve(token);
					if (!neighbor || !nox_xxx_mapGenCheckRoomType_5238F0((int*)neighbor) ||
						sub_523920(*(uint32_t*)neighbor) != direction) {
						continue;
					}
					float2 from;
					float2 to;
					nox_mapgenConnectionPointFarNative_523CB0(neighbor, &from);
					nox_mapgenAddRoomPointNative_522CA0(room, &from);
					nox_mapgenConnectionPointCenterNative_523D30(neighbor, &to);
					sub_522C80(&from.field_0);
					sub_522C80(&to.field_0);
					sub_51D3F0(&from, &to);
					if (!theme || !*(uint32_t*)(theme + 60)) {
						sub_51D3F0(&to, &from);
					}
				}
			}
			continue;
		}

		float2 from;
		nox_mapgenConnectionPointCenterNative_523D30(room, &from);
		for (int direction = 0; direction < 4; ++direction) {
			if (sub_521AA0((uint32_t*)room, direction)) {
				continue;
			}
			int count = room[216 + direction];
			if (count > 8) {
				count = 8;
			}
			for (int index = 0; index < count; ++index) {
				uint32_t token = *(uint32_t*)(room + 88 + 32 * direction + 4 * index);
				uint8_t* neighbor = (uint8_t*)nox_mapgenLegacyPtrResolve(token);
				if (!neighbor) {
					continue;
				}
				float2 to;
				if (nox_xxx_mapGenCheckRoomType_5238F0((int*)neighbor)) {
					nox_mapgenConnectionPointCenterNative_523D30(neighbor, &to);
				} else {
					nox_mapgenConnectionPointNearNative_523C30(room, &to);
					nox_mapgenAddRoomPointNative_522CA0(neighbor, &to);
				}
				sub_522C80(&from.field_0);
				sub_522C80(&to.field_0);
				sub_51D3F0(&from, &to);
				if (!theme || !*(uint32_t*)(theme + 60)) {
					sub_51D3F0(&to, &from);
				}
			}
		}
	}
	return NULL;
}

float* sub_522D30(int a1) {
	return nox_mapgenConnectRoomLinesNative_522D30(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (00522F40) --------------------------------------------------------
unsigned char* nox_mapgenConnectRoomPointsNative_522F40(uint8_t* theme) {
	if (!theme || *(uint32_t*)(theme + 60)) {
		return theme;
	}
	for (uint8_t* room = (uint8_t*)nox_xxx_mapGenGetTopRoom_521710(); room;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		if (*(uint32_t*)room != 1) {
			continue;
		}
		nox_xxx_mapGenSetFlags_5235F0(156);
		int count = room[352];
		if (count > 16) {
			count = 16;
		}
		float2* points = (float2*)(room + 224);
		for (int first = 0; first < count; ++first) {
			for (int second = 0; second < count; ++second) {
				if (first != second) {
					sub_51D3F0(&points[first], &points[second]);
				}
			}
		}
	}
	return NULL;
}

unsigned char* nox_xxx_mapGenTryNextRoom_522F40(uint32_t* a1) {
	return nox_mapgenConnectRoomPointsNative_522F40(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uintptr_t)a1));
}

//----- (00522FF0) --------------------------------------------------------
int nox_xxx_netSendPointFx_522FF0(char a1, float2* a2) {
	short v2;    // ax
	float v3;    // edx
	char a2a[5]; // [esp+4h] [ebp-8h]

	a2a[0] = a1;
	v2 = nox_float2int(a2->field_0);
	v3 = a2->field_4;
	*(uint16_t*)&a2a[1] = v2;
	*(uint16_t*)&a2a[3] = nox_float2int(v3);
	return nox_xxx_netSendFxAllCli_523030(a2, a2a, 5);
}

//----- (00523030) --------------------------------------------------------
int nox_xxx_netSendFxAllCli_523030(float2* a1, const void* a2, int a3) {
	return nox_xxx_netSendFxAllCliNative_523030(a1, (void*)a2, a3);
}

//----- (00523150) --------------------------------------------------------
int sub_523150(char a1, char a2, float* a3) {
	char v4[6]; // [esp+4h] [ebp-8h]

	v4[0] = a1;
	v4[1] = a2;
	*(uint16_t*)&v4[2] = nox_float2int(*a3);
	*(uint16_t*)&v4[4] = nox_float2int(a3[1]);
	return nox_xxx_netSendFxAllCli_523030((float2*)a3, v4, 6);
}

//----- (005231B0) --------------------------------------------------------
int nox_xxx_netSparkExplosionFx_5231B0(float* a1, char a2) {
	short v2;   // ax
	float v3;   // ecx
	char v5[6]; // [esp+4h] [ebp-8h]

	v5[0] = -109;
	v2 = nox_float2int(*a1);
	v3 = a1[1];
	*(uint16_t*)&v5[1] = v2;
	*(uint16_t*)&v5[3] = nox_float2int(v3);
	v5[5] = a2;
	return nox_xxx_netSendFxAllCli_523030((float2*)a1, v5, 6);
}

//----- (00523200) --------------------------------------------------------
void nox_xxx_sendGeneratorBreakFX_523200(float* a1, char a2) {
	short v2;   // ax
	float v3;   // ecx
	char v5[7]; // [esp+4h] [ebp-8h]

	v5[0] = -16;
	v5[1] = 25;
	v2 = nox_float2int(*a1);
	v3 = a1[1];
	*(uint16_t*)&v5[2] = v2;
	*(uint16_t*)&v5[4] = nox_float2int(v3);
	v5[6] = a2;
	nox_xxx_netSendFxAllCli_523030((float2*)a1, v5, 7);
}

//----- (00523270) --------------------------------------------------------
int nox_xxx_netSendVampFx_523270(char a1, short* a2, short a3) {
	short v3;          // dx
	unsigned short v4; // cx
	unsigned short v5; // ax
	float2 a1a;        // [esp+0h] [ebp-14h]
	char a2a[11];      // [esp+8h] [ebp-Ch]

	a2a[0] = a1;
	v3 = a2[2];
	*(uint16_t*)&a2a[1] = *a2;
	v4 = a2[4];
	v5 = a2[6];
	*(uint16_t*)&a2a[3] = v3;
	*(uint16_t*)&a2a[5] = v4;
	*(uint16_t*)&a2a[7] = v5;
	a1a.field_0 = (double)v4;
	*(uint16_t*)&a2a[9] = a3;
	a1a.field_4 = (double)v5;
	return nox_xxx_netSendFxAllCli_523030(&a1a, a2a, 11);
}

//----- (00523530) --------------------------------------------------------
int nox_xxx_netClientPredictLinear_523530(int a1) {
	short v1;      // ax
	double v2;     // st7
	long long v3;  // rax
	double v4;     // st7
	long long v5;  // rax
	double v6;     // st7
	short v7;      // cx
	long long v8;  // rax
	double v9;     // st7
	long long v10; // rax
	double v11;    // st7
	int result;    // eax
	int i;         // esi
	char v14[14];  // [esp+4h] [ebp-10h]

	v14[0] = -75;
	v1 = nox_xxx_netGetUnitCodeServ_578AC0((uint32_t*)a1);
	v2 = *(float*)(a1 + 56);
	*(uint16_t*)&v14[1] = v1;
	*(uint16_t*)&v14[3] = *(uint16_t*)(a1 + 4);
	v3 = (long long)v2;
	v4 = *(float*)(a1 + 60);
	*(uint16_t*)&v14[5] = v3;
	v5 = (long long)v4;
	v6 = *(float*)(a1 + 112) * 16.0;
	v7 = *(uint16_t*)(a1 + 124);
	*(uint16_t*)&v14[7] = v5;
	*(uint16_t*)&v14[9] = v7;
	v8 = (long long)v6;
	v9 = *(float*)(a1 + 80) * 16.0;
	v14[11] = v8;
	v10 = (long long)v9;
	v11 = *(float*)(a1 + 84) * 16.0;
	v14[12] = v10;
	v14[13] = (long long)v11;
	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		nox_netlist_addToMsgListCli_40EBC0(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), 1, v14,
										   14);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}

//----- (005235F0) --------------------------------------------------------
void nox_xxx_mapGenSetFlags_5235F0(char a1) {
	int v3; // [esp+0h] [ebp-4h]

	if (!nox_common_gameFlags_check_40A5C0(0x200000)) {
		dword_5d4594_2487564 = nox_platform_get_ticks();
		if (dword_5d4594_2487568 > *(int*)&dword_5d4594_2487564) {
			dword_5d4594_2487568 = 0;
		}
		nox_input_pollEvents_4453A0();
		if (dword_5d4594_2487564 - dword_5d4594_2487568 > *getMemIntPtr(0x587000, 254948)) {
			*(uint16_t*)((char*)&v3 + 1) = *getMemU16Ptr(0x5D4594, 2487572);
			LOBYTE(v3) = a1;
			++*getMemU32Ptr(0x5D4594, 2487572);
			nox_xxx_mapGenClientText_4A9D00((unsigned char*)&v3);
			dword_5d4594_2487568 = dword_5d4594_2487564;
		}
	}
}

//----- (00523670) --------------------------------------------------------
int nox_xxx_netSendShieldFx_523670(int a1, float* a2) {
	char v2;    // al
	int v3;     // eax
	char v5[4]; // [esp+4h] [ebp-Ch]
	float2 v6;  // [esp+8h] [ebp-8h]

	v5[0] = -128;
	*(uint16_t*)&v5[1] = nox_xxx_netGetUnitCodeServ_578AC0((uint32_t*)a1);
	if (a2) {
		v6.field_0 = *(float*)(a1 + 56) - *a2;
		v6.field_4 = *(float*)(a1 + 60) - a2[1];
		v3 = nox_xxx_math_509ED0(&v6);
		v2 = nox_xxx_math_509EA0(v3);
	} else {
		v2 = nox_xxx_math_509EA0(*(short*)(a1 + 124));
	}
	v5[3] = v2;
	return nox_xxx_netSendFxAllCli_523030((float2*)(a1 + 56), v5, 4);
}

//----- (005236F0) --------------------------------------------------------
int nox_xxx_sendSummonStartFX_5236F0(short a1, float* a2, char a3, short a4, short a5) {
	double v5;    // st7
	long long v6; // rax
	double v7;    // st7
	char v9[12];  // [esp+4h] [ebp-Ch]

	v9[0] = 126;
	*(uint16_t*)&v9[5] = a1;
	*(uint16_t*)&v9[7] = a4;
	v5 = *a2;
	v9[9] = a3;
	v6 = (long long)v5;
	v7 = a2[1];
	*(uint16_t*)&v9[1] = v6;
	*(uint16_t*)&v9[3] = (long long)v7;
	*(uint16_t*)&v9[10] = a5;
	return nox_xxx_netSendPacket0_4E5420(255, v9, 12, 0, 1);
}

//----- (00523760) --------------------------------------------------------
int nox_xxx_sendSummonCancelFX_523760(short a1) {
	char v3[3]; // [esp+0h] [ebp-4h]
	v3[0] = 127;
	*(uint16_t*)&v3[1] = a1;
	return nox_xxx_netSendPacket0_4E5420(255, v3, 3, 0, 1);
}

//----- (00523830) --------------------------------------------------------
void nox_xxx_sendGeneratorSpawnFX_523830(int4* a1, short a2) {
	double v2;    // st7
	short v3;     // cx
	short v4;     // dx
	double v5;    // st7
	short v6;     // cx
	float2 a1a;   // [esp+0h] [ebp-14h]
	char a2a[12]; // [esp+8h] [ebp-Ch]

	a2a[0] = -16;
	a2a[1] = 16;
	v2 = (double)a1->field_8;
	v3 = a1->field_0;
	*(uint16_t*)&a2a[4] = a1->field_4;
	v4 = a1->field_C;
	a1a.field_0 = v2;
	v5 = (double)a1->field_C;
	*(uint16_t*)&a2a[2] = v3;
	v6 = a1->field_8;
	*(uint16_t*)&a2a[8] = v4;
	*(uint16_t*)&a2a[6] = v6;
	a1a.field_4 = v5;
	*(uint16_t*)&a2a[10] = a2;
	nox_xxx_netSendFxAllCli_523030(&a1a, a2a, 12);
}

//----- (005238A0) --------------------------------------------------------
void nox_xxx_sendArrowTrapFX_5238A0(float* a1, char a2) {
	short v2;   // ax
	float v3;   // ecx
	char v5[6]; // [esp+4h] [ebp-8h]

	v5[0] = -95;
	v2 = nox_float2int16(*a1);
	v3 = a1[1];
	*(uint16_t*)&v5[1] = v2;
	*(uint16_t*)&v5[3] = nox_float2int16(v3);
	v5[5] = a2;
	nox_xxx_netSendFxAllCli_523030((float2*)a1, v5, 6);
}

//----- (005238F0) --------------------------------------------------------
int nox_xxx_mapGenCheckRoomType_5238F0(int* a1) {
	int v1; // eax

	v1 = *a1;
	return *a1 == 2 || v1 == 3 || v1 == 4 || v1 == 5;
}

//----- (00523920) --------------------------------------------------------
int sub_523920(int a1) {
	switch (a1) {
	case 2:
		return 0;
	case 3:
		return 1;
	case 4:
		return 2;
	}
	return a1 != 5 ? 0 : 3;
}

//----- (00523960) --------------------------------------------------------
int sub_523960(int a1) { return *getMemU32Ptr(0x587000, 254952 + 4 * a1); }

//----- (00523970) --------------------------------------------------------
int sub_523970(int a1) {
	int result; // eax

	switch (a1) {
	case 2:
		result = 3;
		break;
	case 3:
		result = 2;
		break;
	case 5:
		result = 1;
		break;
	default:
		result = 0;
		break;
	}
	return result;
}

//----- (005239B0) --------------------------------------------------------
int sub_5239B0(int a1) {
	int v1; // eax

	v1 = sub_523970(a1);
	return sub_523960(v1);
}

//----- (00523A10) --------------------------------------------------------
int nox_mapgenAdjustHallNative_523A10(uint8_t* a1, const float* a2) {
	double v2;  // st7
	int v3;     // eax
	double v4;  // st6
	int v5;     // eax
	float v7;   // ecx
	int v8;     // eax
	double v9;  // st7
	int v10;    // eax
	float v11;  // ecx
	int v12;    // eax
	double v13; // st7
	int v14;    // eax
	double v15; // st7
	int v16;    // eax
	double v17; // st6
	int v18;    // eax
	float2 v19;

	if (!a1 || !a2) {
		return 0;
	}
	switch (*(uint32_t*)a1) {
	case 2:
		if (*(float*)(a1 + 36) < (double)a2[9] || *(float*)(a1 + 44) > (double)a2[11]) {
			return 0;
		}
		v2 = *(float*)(a1 + 24);
		v19.field_0 = *(float*)(a1 + 20);
		v3 = *(uint32_t*)(a1 + 16);
		v19.field_4 = v2;
		if (v3 > 1) {
			do {
				v2 = v2 + 32.526913;
				v4 = *(float*)(a1 + 32) - 32.526913;
				v5 = *(uint32_t*)(a1 + 16) - 1;
				*(uint32_t*)(a1 + 16) = v5;
				*(float*)(a1 + 32) = v4;
			} while (v2 + 0.1 <= a2[12] && v5 > 1);
			v19.field_4 = v2;
		}
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)a1, &v19);
		return nox_mapgenRoomAtNative_521200(a1) == NULL;
	case 3:
		if (*(float*)(a1 + 36) < (double)a2[9] || *(float*)(a1 + 44) > (double)a2[11]) {
			return 0;
		}
		v7 = *(float*)(a1 + 24);
		v19.field_0 = *(float*)(a1 + 20);
		v8 = *(uint32_t*)(a1 + 16);
		v19.field_4 = v7;
		if (v8 <= 1) {
			nox_xxx_mapGenSetRoomPos_521880((uint32_t*)a1, &v19);
			return nox_mapgenRoomAtNative_521200(a1) == NULL;
		}
		break;
	case 4:
		if (*(float*)(a1 + 40) < (double)a2[10] || *(float*)(a1 + 48) > (double)a2[12]) {
			return 0;
		}
		v11 = *(float*)(a1 + 24);
		v19.field_0 = *(float*)(a1 + 20);
		v12 = *(uint32_t*)(a1 + 12);
		v19.field_4 = v11;
		if (v12 > 1) {
			do {
				v13 = *(float*)(a1 + 28) - 32.526913;
				v14 = *(uint32_t*)(a1 + 12) - 1;
				*(uint32_t*)(a1 + 12) = v14;
				*(float*)(a1 + 28) = v13;
			} while (v13 + v19.field_0 - 0.1 >= a2[9] && v14 > 1);
		}
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)a1, &v19);
		return nox_mapgenRoomAtNative_521200(a1) == NULL;
	case 5:
		if (*(float*)(a1 + 40) < (double)a2[10] || *(float*)(a1 + 48) > (double)a2[12]) {
			return 0;
		}
		v15 = *(float*)(a1 + 20);
		v19.field_4 = *(float*)(a1 + 24);
		v16 = *(uint32_t*)(a1 + 12);
		v19.field_0 = v15;
		if (v16 > 1) {
			do {
				v15 = v15 + 32.526913;
				v17 = *(float*)(a1 + 28) - 32.526913;
				v18 = *(uint32_t*)(a1 + 12) - 1;
				*(uint32_t*)(a1 + 12) = v18;
				*(float*)(a1 + 28) = v17;
			} while (v15 + 0.1 <= a2[11] && v18 > 1);
			v19.field_0 = v15;
		}
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)a1, &v19);
		return nox_mapgenRoomAtNative_521200(a1) == NULL;
	default:
		return nox_mapgenRoomAtNative_521200(a1) == NULL;
	}
	while (1) {
		v9 = *(float*)(a1 + 32) - 32.526913;
		v10 = *(uint32_t*)(a1 + 16) - 1;
		*(uint32_t*)(a1 + 16) = v10;
		*(float*)(a1 + 32) = v9;
		if (v9 + v19.field_4 - 0.1 < a2[10]) {
			break;
		}
		if (v10 <= 1) {
			nox_xxx_mapGenSetRoomPos_521880((uint32_t*)a1, &v19);
			return nox_mapgenRoomAtNative_521200(a1) == NULL;
		}
	}
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)a1, &v19);
	return nox_mapgenRoomAtNative_521200(a1) == NULL;
}

int sub_523A10(int a1, float* a2) {
	return nox_mapgenAdjustHallNative_523A10(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (00523C30) --------------------------------------------------------
int nox_mapgenConnectionPointNearNative_523C30(uint8_t* room, float2* point) {
	int result = (int)nox_mapgenLegacyPtrRegister(room);
	if (!room || !point) {
		return result;
	}
	switch (*(uint32_t*)room) {
	case 2:
		point->field_0 = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5;
		result = *(uint32_t*)(room + 40);
		point->field_4 = *(float*)(room + 40);
		break;
	case 3:
		point->field_0 = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5;
		point->field_4 = *(float*)(room + 48);
		break;
	case 4:
		result = *(uint32_t*)(room + 44);
		point->field_0 = *(float*)(room + 44);
		point->field_4 = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5;
		break;
	case 5:
		result = *(uint32_t*)(room + 36);
		point->field_0 = *(float*)(room + 36);
		point->field_4 = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5;
		break;
	}
	return result;
}

int sub_523C30(int a1, int a2) {
	return nox_mapgenConnectionPointNearNative_523C30(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(float2*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00523CB0) --------------------------------------------------------
int nox_mapgenConnectionPointFarNative_523CB0(uint8_t* room, float2* point) {
	int result = (int)nox_mapgenLegacyPtrRegister(room);
	if (!room || !point) {
		return result;
	}
	switch (*(uint32_t*)room) {
	case 2:
		point->field_0 = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5;
		result = *(uint32_t*)(room + 48);
		point->field_4 = *(float*)(room + 48);
		break;
	case 3:
		point->field_0 = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5;
		point->field_4 = *(float*)(room + 40);
		break;
	case 4:
		result = *(uint32_t*)(room + 36);
		point->field_0 = *(float*)(room + 36);
		point->field_4 = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5;
		break;
	case 5:
		result = *(uint32_t*)(room + 44);
		point->field_0 = *(float*)(room + 44);
		point->field_4 = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5;
		break;
	}
	return result;
}

int sub_523CB0(int a1, int a2) {
	return nox_mapgenConnectionPointFarNative_523CB0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(float2*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00523D30) --------------------------------------------------------
float* nox_mapgenConnectionPointCenterNative_523D30(uint8_t* room, float2* point) {
	if (!room || !point) {
		return (float*)room;
	}
	int type = *(uint32_t*)room;
	int first_offset = type == 2 || type == 3 ? 88 : 152;
	int second_offset = type == 2 || type == 3 ? 120 : 184;
	uint8_t* first = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room + first_offset));
	uint8_t* second = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room + second_offset));
	if (first && second && *(uint32_t*)first == 1 && *(uint32_t*)second == 1) {
		point->field_0 = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5;
		point->field_4 = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5;
		return (float*)room;
	}
	switch (type) {
	case 2:
		point->field_0 = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5;
		point->field_4 = *(float*)(room + 28) * 0.5 + *(float*)(room + 40);
		break;
	case 3:
		point->field_0 = (*(float*)(room + 44) + *(float*)(room + 36)) * 0.5;
		point->field_4 = *(float*)(room + 48) - *(float*)(room + 28) * 0.5;
		break;
	case 4:
		point->field_0 = *(float*)(room + 44) - *(float*)(room + 32) * 0.5;
		point->field_4 = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5;
		break;
	case 5:
		point->field_0 = *(float*)(room + 32) * 0.5 + *(float*)(room + 36);
		point->field_4 = (*(float*)(room + 48) + *(float*)(room + 40)) * 0.5;
		break;
	}
	return (float*)room;
}

float* sub_523D30(float* a1, float* a2) {
	return nox_mapgenConnectionPointCenterNative_523D30(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uintptr_t)a1),
		(float2*)nox_mapgenLegacyPtrResolve((uintptr_t)a2));
}

//----- (00523E30) --------------------------------------------------------
float* nox_mapgenMakeHallStructNative_523E30(int a1, int a2, int a3) {
	float* result; // eax

	result = (float*)calloc(1u, 0x178u);
	if (result) {
		if (!nox_mapgenLegacyPtrRegister(result)) {
			free(result);
			return NULL;
		}
		*(uint32_t*)result = a1;
		switch (a1) {
		case 2:
		case 3:
			*((uint32_t*)result + 3) = a2;
			if (a2 < 1) {
				*((uint32_t*)result + 3) = 1;
			}
			*((uint32_t*)result + 4) = a3;
			break;
		case 4:
		case 5:
			*((uint32_t*)result + 3) = a3;
			*((uint32_t*)result + 4) = a2;
			if (a2 < 1) {
				*((uint32_t*)result + 4) = 1;
			}
			break;
		default:
			break;
		}
		result[7] = (double)*((int*)result + 3) * 32.526913;
		result[8] = (double)*((int*)result + 4) * 32.526913;
	}
	return result;
}

float* sub_523E30(int a1, int a2, int a3) {
	return nox_mapgenMakeHallStructNative_523E30(a1, a2, a3);
}

//----- (00523EC0) --------------------------------------------------------
float* nox_mapgenMakeHallNative_523EC0(uint8_t* theme, int type, int size) {
	if (!theme) {
		return NULL;
	}
	int depth = nox_xxx_mapGenRandFunc2_526B00(*(uint32_t*)(theme + 4), *(uint32_t*)(theme + 8));
	return nox_mapgenMakeHallStructNative_523E30(type, size, depth);
}

float* nox_xxx_mapGenMakeHall_523EC0(int a1, int a2, int a3) {
	return nox_mapgenMakeHallNative_523EC0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2, a3);
}

//----- (00524070) --------------------------------------------------------
static uint32_t nox_mapgenSelectDecorNative_524090(uint8_t* room, int* settings);

int nox_mapgenSelectBackdropNative_524070(uint8_t* theme, uint8_t* room) {
	if (!theme || !room) {
		return 0;
	}
	uint32_t token = nox_mapgenSelectDecorNative_524090(room, (int*)(theme + 184));
	*(uint32_t*)(room + 372) = token;
	return token != 0;
}

int sub_524070(int a1, int a2) {
	return nox_mapgenSelectBackdropNative_524070(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00524090) --------------------------------------------------------
static int nox_mapgenDecorMatchesConstraint_5241C0(
	const nox_mapgen_decor_definition_legacy* definition, const uint8_t* room) {
	return definition && room &&
		((*(const uint32_t*)(room + 364) & definition->constraint_mask) != 0 ||
		 !definition->constraint_mask);
}

static int nox_mapgenDecorFillsRoomNative_5241F0(
	const nox_mapgen_decor_definition_legacy* definition, const uint8_t* room) {
	if (!definition || !room) {
		return 0;
	}
	int size = *(const int32_t*)(room + 12);
	if (size <= *(const int32_t*)(room + 16)) {
		size = *(const int32_t*)(room + 16);
	}
	return size >= definition->min_room_size && size <= definition->max_room_size;
}

static void nox_mapgenDecorCheckLimitNative_524220(int* settings, uint32_t definition_token) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve(definition_token);
	if (!settings || !definition || !definition->occur_limit) {
		return;
	}
	if (--definition->occur_limit) {
		return;
	}
	if (definition->constraint_mask & 2) {
		settings[3] -= definition->frequency;
	}
	if (definition->constraint_mask & 4) {
		settings[4] -= definition->frequency;
	}
	if (definition->constraint_mask & 8) {
		settings[5] -= definition->frequency;
	}
	if (definition->constraint_mask & 0x10) {
		settings[6] -= definition->frequency;
	}

	uint32_t previous_token = 0;
	uint32_t token = (uint32_t)settings[0];
	while (token && token != definition_token) {
		nox_mapgen_decor_definition_legacy* item = nox_mapgenDecorDefinitionResolve(token);
		if (!item) {
			return;
		}
		previous_token = token;
		token = item->next_token;
	}
	if (!token) {
		return;
	}
	if (previous_token) {
		nox_mapgen_decor_definition_legacy* previous =
			nox_mapgenDecorDefinitionResolve(previous_token);
		if (!previous) {
			return;
		}
		previous->next_token = definition->next_token;
	} else {
		settings[0] = (int)definition->next_token;
	}
	definition->next_token = 0;
	definition->exhausted = 1;

	uint32_t tail_token = (uint32_t)settings[0];
	if (!tail_token) {
		settings[0] = (int)definition_token;
		return;
	}
	nox_mapgen_decor_definition_legacy* tail = nox_mapgenDecorDefinitionResolve(tail_token);
	while (tail && tail->next_token) {
		tail = nox_mapgenDecorDefinitionResolve(tail->next_token);
	}
	if (tail) {
		tail->next_token = definition_token;
	}
}

static uint32_t nox_mapgenSelectDecorNative_524090(uint8_t* room, int* settings) {
	if (!room || !settings) {
		return 0;
	}
	for (int attempt = 0; attempt <= 100; ++attempt) {
		int total = 0;
		switch (*(uint32_t*)(room + 364)) {
		case 1:
			total = settings[2];
			break;
		case 2:
			total = settings[3];
			break;
		case 4:
			total = settings[4];
			break;
		case 8:
			total = settings[5];
			break;
		case 0x10:
			total = settings[6];
			break;
		case 0x20:
			total = settings[7];
			break;
		default:
			break;
		}
		if (total <= 0) {
			nullsub_28((int)(uint32_t)nox_mapgenLegacyPtrRegister(room));
			return 0;
		}
		int roll = nox_xxx_mapGenRandFunc_526AC0(0, total);
		uint32_t selected_token = 0;
		for (uint32_t token = (uint32_t)settings[0]; token;) {
			nox_mapgen_decor_definition_legacy* definition =
				nox_mapgenDecorDefinitionResolve(token);
			if (!definition) {
				break;
			}
			if (nox_mapgenDecorMatchesConstraint_5241C0(definition, room)) {
				roll -= definition->frequency;
				if (roll <= 0) {
					selected_token = token;
					break;
				}
			}
			token = definition->next_token;
		}
		if (!selected_token) {
			continue;
		}
		nox_mapgen_decor_definition_legacy* selected =
			nox_mapgenDecorDefinitionResolve(selected_token);
		if (nox_mapgenDecorFillsRoomNative_5241F0(selected, room) && !selected->exhausted) {
			nox_mapgenDecorCheckLimitNative_524220(settings, selected_token);
			return selected_token;
		}
	}
	nullsub_28((int)(uint32_t)nox_mapgenLegacyPtrRegister(room));
	return 0;
}

int sub_524090(int a1, int* a2) {
	return (int)nox_mapgenSelectDecorNative_524090(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}
// 524300: using guessed type void  nullsub_28(uint32_t);

//----- (005241C0) --------------------------------------------------------
int nox_xxx_mapGenDecorChkConstaint_5241C0(int a1, int a2) {
	return nox_mapgenDecorMatchesConstraint_5241C0(
		nox_mapgenDecorDefinitionResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (005241F0) --------------------------------------------------------
int nox_xxx_mapGenChkDecorFillsRoom_5241F0(int a1, int a2) {
	return nox_mapgenDecorFillsRoomNative_5241F0(
		nox_mapgenDecorDefinitionResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00524220) --------------------------------------------------------
char nox_xxx_mapGenDecorChkLimit_524220(int* a1, int a2) {
	nox_mapgenDecorCheckLimitNative_524220(a1, (uint32_t)a2);
	return 0;
}

//----- (00524310) --------------------------------------------------------
static int nox_mapgenAssignRequiredDecorNative_524310(
	int* settings, int halls, uint8_t* first_room) {
	for (uint32_t token = (uint32_t)settings[0]; token;) {
		nox_mapgen_decor_definition_legacy* definition =
			nox_mapgenDecorDefinitionResolve(token);
		if (!definition) {
			return 0;
		}
		uint32_t next_token = definition->next_token;
		if (definition->must_occur && !definition->occurred) {
			uint8_t* room = first_room;
			while (room &&
				(*(uint32_t*)(room + 372) ||
				 nox_xxx_mapGenCheckRoomType_5238F0((int*)room) != halls ||
				 !nox_mapgenDecorMatchesConstraint_5241C0(definition, room) ||
				 !nox_mapgenDecorFillsRoomNative_5241F0(definition, room) ||
				 definition->exhausted)) {
				room = (uint8_t*)nox_mapgenRoomNextNative_521720(room);
			}
			if (!room) {
				return 0;
			}
			nox_mapgenDecorCheckLimitNative_524220(settings, token);
			*(uint32_t*)(room + 372) = token;
			definition->occurred = 1;
		}
		token = next_token;
	}
	return 1;
}

int nox_mapgenMakeRoomsNative_524310(uint8_t* theme) {
	if (!theme) {
		return 0;
	}
	int* rooms = (int*)(theme + 88);
	int* halls = (int*)(theme + 120);
	uint8_t* first_room = (uint8_t*)nox_xxx_mapGenGetTopRoom_521710();
	if (!nox_mapgenAssignRequiredDecorNative_524310(rooms, 0, first_room) ||
		!nox_mapgenAssignRequiredDecorNative_524310(halls, 1, first_room)) {
		return 0;
	}
	for (uint8_t* room = first_room; room;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		if (*(uint32_t*)(room + 372)) {
			continue;
		}
		int* settings = nox_xxx_mapGenCheckRoomType_5238F0((int*)room) ? halls : rooms;
		uint32_t token = nox_mapgenSelectDecorNative_524090(room, settings);
		*(uint32_t*)(room + 372) = token;
		if (!token) {
			return 0;
		}
	}
	return 1;
}

int nox_xxx_mapGenMakeRooms_524310(int a1) {
	return nox_mapgenMakeRoomsNative_524310(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (005244D0) --------------------------------------------------------
int sub_5244D0(int a1) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve((uint32_t)a1);
	if (!definition || !definition->wall_floor_count) {
		return 0;
	}
	int selected = nox_xxx_mapGenRandFunc_526AC0(0, definition->wall_floor_count - 1);
	uint32_t token = definition->wall_floors_token;
	while (token && selected-- > 0) {
		nox_mapgen_wall_floor_legacy* wall_floor = nox_mapgenWallFloorResolve(token);
		token = wall_floor ? wall_floor->next_token : 0;
	}
	return (int)token;
}

//----- (00524500) --------------------------------------------------------
int sub_524500(float2* a1, int a2) {
	float v2;   // edx
	int result; // eax
	int v4;     // esi
	float2 v5;  // [esp+0h] [ebp-8h]

	v2 = a1->field_4;
	v5.field_0 = a1->field_0;
	v5.field_4 = v2;
	sub_526D50(0);
	result = a2;
	if (a2 > 0) {
		v4 = a2;
		do {
			result = sub_526E60(&v5);
			--v4;
			v5.field_0 = v5.field_0 + 32.526913;
		} while (v4);
	}
	return result;
}

//----- (00524550) --------------------------------------------------------
int sub_524550(int* a1, int a2) {
	int v2;     // edx
	int result; // eax
	int v4;     // esi
	int v5;     // [esp+4h] [ebp-8h]
	float v6;   // [esp+8h] [ebp-4h]

	v2 = a1[1];
	v5 = *a1;
	v6 = *(float*)&v2;
	result = sub_526D50(1);
	v4 = a2;
	if (a2 > 0) {
		do {
			result = sub_526E60((float*)&v5);
			--v4;
			v6 = v6 + 32.526913;
		} while (v4);
	}
	return result;
}

//----- (005245A0) --------------------------------------------------------
float* sub_5245A0(int a1, float* a2, int a3, int a4) {
	float* result; // eax
	int v5;        // ebp
	int v6;        // esi
	float v7;      // [esp+4h] [ebp-8h]
	float v8;      // [esp+8h] [ebp-4h]

	result = (float*)a4;
	v8 = a2[1] + 16.263456;
	if (a4 > 0) {
		v5 = a4;
		do {
			v7 = *a2 + 16.263456;
			if (a3 > 0) {
				v6 = a3;
				do {
					result = sub_51D5E0(&v7);
					--v6;
					v7 = v7 + 32.526913;
				} while (v6);
			}
			--v5;
			v8 = v8 + 32.526913;
		} while (v5);
	}
	return result;
}

//----- (00524610) --------------------------------------------------------
float* sub_524610(int a1, float* a2, int a3) {
	float* result; // eax
	int v4;        // esi
	float v5;      // [esp+4h] [ebp-8h]
	float v6;      // [esp+8h] [ebp-4h]

	result = a2;
	v4 = a3;
	v5 = *a2 + 16.263456;
	v6 = a2[1] + 16.263456;
	if (a3 > 0) {
		do {
			result = sub_51D5E0(&v5);
			--v4;
			v6 = v6 + 32.526913;
		} while (v4);
	}
	return result;
}

//----- (00524660) --------------------------------------------------------
int sub_524660(float a1, float a2) { return fabs(a1 - a2) < *getMemDoublePtr(0x581450, 10432); }

static uint8_t* nox_mapgenRoomNeighborNative_524680(uint8_t* room, int offset) {
	return room ? (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room + offset)) : NULL;
}

int nox_mapgenGenerateWallFloorNative_524680(
	uint8_t* theme, uint8_t* room, nox_mapgen_wall_floor_legacy* wall_floor) {
	if (!theme || !room || !wall_floor) {
		return 0;
	}
	nox_xxx_tileGetDefByName_51D4D0(wall_floor->floor_name);
	sub_5245A0(0, (float*)(room + 20), *(int32_t*)(room + 12), *(int32_t*)(room + 16));

	int dimensions[2] = {0, 0};
	float2 origin = {0.0f, 0.0f};
#define width dimensions[0]
#define height dimensions[1]
#define origin_x origin.field_0
#define origin_y origin.field_4
	switch (*(uint32_t*)room) {
	case 1:
		width = *(int32_t*)(room + 12) - 2;
		height = *(int32_t*)(room + 16) - 2;
		origin_x = *(float*)(room + 20) + 32.526913f;
		origin_y = *(float*)(room + 24) + 32.526913f;
		break;
	case 2: {
		uint8_t* north = nox_mapgenRoomNeighborNative_524680(room, 88);
		uint8_t* south = nox_mapgenRoomNeighborNative_524680(room, 120);
		width = *(int32_t*)(room + 12) - 2;
		origin_x = *(float*)(room + 20) + 32.526913f;
		if (north && *(uint32_t*)north != 1) {
			origin_y = *(float*)(room + 24);
			height = *(int32_t*)(room + 16);
		} else {
			origin_y = *(float*)(room + 24) + 32.526913f;
			height = *(int32_t*)(room + 16) - 1;
		}
		if (south) {
			height += *(uint32_t*)south == 1 ? -1 : 1;
		}
		break;
	}
	case 3: {
		uint8_t* south = nox_mapgenRoomNeighborNative_524680(room, 120);
		uint8_t* north = nox_mapgenRoomNeighborNative_524680(room, 88);
		width = *(int32_t*)(room + 12) - 2;
		origin_x = *(float*)(room + 20) + 32.526913f;
		origin_y = *(float*)(room + 24);
		height = south && *(uint32_t*)south != 1
			? *(int32_t*)(room + 16) : *(int32_t*)(room + 16) - 1;
		if (north) {
			if (*(uint32_t*)north == 1) {
				--height;
				origin_y += 32.526913f;
			} else {
				++height;
				origin_y -= 32.526913f;
			}
		}
		break;
	}
	case 4: {
		uint8_t* east = nox_mapgenRoomNeighborNative_524680(room, 152);
		uint8_t* west = nox_mapgenRoomNeighborNative_524680(room, 184);
		height = *(int32_t*)(room + 16) - 2;
		origin_x = *(float*)(room + 20);
		origin_y = *(float*)(room + 24) + 32.526913f;
		width = east && *(uint32_t*)east != 1
			? *(int32_t*)(room + 12) : *(int32_t*)(room + 12) - 1;
		if (west) {
			if (*(uint32_t*)west != 1) {
				++width;
				origin_x -= 32.526913f;
			} else {
				--width;
				origin_x += 32.526913f;
			}
		}
		break;
	}
	case 5: {
		uint8_t* west = nox_mapgenRoomNeighborNative_524680(room, 184);
		uint8_t* east = nox_mapgenRoomNeighborNative_524680(room, 152);
		height = *(int32_t*)(room + 16) - 2;
		origin_y = *(float*)(room + 24) + 32.526913f;
		if (west && *(uint32_t*)west != 1) {
			width = *(int32_t*)(room + 12);
			origin_x = *(float*)(room + 20);
		} else {
			width = *(int32_t*)(room + 12) - 1;
			origin_x = *(float*)(room + 20) + 32.526913f;
		}
		if (east) {
			width += *(uint32_t*)east == 1 ? -1 : 1;
		}
		break;
	}
	default:
		break;
	}

	for (uint32_t token = wall_floor->fills_token; token;) {
		nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve(token);
		if (!fill || width <= 0 || height <= 0) {
			break;
		}
		if (fill->type == 1) {
			sub_5249C0(0, (int)token, (float*)&origin, dimensions);
		} else if (fill->type == 2) {
			sub_524B50(0, (int)token, (float*)&origin, dimensions);
		} else {
			sub_524950(0, (int)token, (float*)&origin, dimensions);
		}
		width -= 2;
		height -= 2;
		origin_x += 32.526913f;
		origin_y += 32.526913f;
		token = fill->next_token;
	}
#undef origin_y
#undef origin_x
#undef height
#undef width
	return dimensions[0];
}

//----- (00524680) --------------------------------------------------------
int nox_xxx_gen_524680(int a1, int a2, int a3) {
	return nox_mapgenGenerateWallFloorNative_524680(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2),
		nox_mapgenWallFloorResolve((uint32_t)a3));
}

//----- (00524950) --------------------------------------------------------
int sub_524950(int a1, int a2, float* a3, int* a4) {
	float v5; // [esp+Ch] [ebp-8h]
	float v6; // [esp+10h] [ebp-4h]
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve((uint32_t)a2);
	if (!fill) {
		return 0;
	}

	nox_xxx_tileGetDefByName_51D4D0(fill->tile_name);
	sub_5245A0(a1, a3, *a4, a4[1]);
	sub_544020(fill->object_name);
	v5 = (double)*a4 * 16.263456 + *a3;
	v6 = (double)a4[1] * 16.263456 + a3[1];
	return sub_543680(&v5);
}

//----- (005249C0) --------------------------------------------------------
int sub_5249C0(int a1, int a2, float* a3, int* a4) {
	int* v4;    // esi
	int v5;     // ecx
	int v6;     // edx
	int v7;     // eax
	int v8;     // ebp
	double v9;  // st7
	int v10;    // edi
	double v11; // st7
	int v12;    // eax
	bool v13;   // cc
	int v14;    // eax
	int v15;    // edi
	int v16;    // ebp
	int v17;    // eax
	float v19;  // [esp+10h] [ebp-14h]
	float v20;  // [esp+14h] [ebp-10h]
	float v21;  // [esp+18h] [ebp-Ch]
	float v22;  // [esp+1Ch] [ebp-8h]
	float v23;  // [esp+20h] [ebp-4h]
	int v24;    // [esp+34h] [ebp+10h]
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve((uint32_t)a2);
	if (!fill) {
		return 0;
	}

	nox_xxx_tileGetDefByName_51D4D0(fill->tile_name);
	v4 = a4;
	v5 = *a4;
	v6 = *a4 / 3;
	v7 = a4[1];
	v8 = 2 * v6;
	v24 = 0;
	v9 = (double)v6 * 32.526913;
	v10 = v5 - 2 * v6;
	v19 = v9;
	v11 = v9 + *a3;
	v21 = a3[1];
	v20 = v11;
	if (v7 / 2 >= 0) {
		do {
			sub_5245A0(a1, &v20, v10, 1);
			v12 = *v4;
			v10 += 2;
			v13 = v10 < *v4;
			v20 = v20 - 32.526913;
			v21 = v21 + 32.526913;
			if (!v13) {
				v10 = v12;
				v20 = *a3;
			}
			++v24;
		} while (v24 <= v4[1] / 2);
	}
	v14 = v4[1];
	v15 = *v4 - v8;
	v16 = 0;
	v20 = v19 + *a3;
	v21 = (double)(v14 - 1) * 32.526913 + a3[1];
	if (v14 / 2 > 0) {
		do {
			sub_5245A0(a1, &v20, v15, 1);
			v17 = *v4;
			v15 += 2;
			v13 = v15 < *v4;
			v20 = v20 - 32.526913;
			v21 = v21 - 32.526913;
			if (!v13) {
				v15 = v17;
				v20 = *a3;
			}
			++v16;
		} while (v16 < v4[1] / 2);
	}
	sub_544020(fill->object_name);
	v22 = (double)*v4 * 16.263456 + *a3;
	v23 = (double)v4[1] * 16.263456 + a3[1];
	return sub_543680(&v22);
}

//----- (00524B50) --------------------------------------------------------
void sub_524B50(int a1, int a2, float* a3, int* a4) {
	int* v4;        // esi
	int v5;         // eax
	int v7;         // ecx
	int v8;         // eax
	signed int v9;  // eax
	int v10;        // ebp
	signed int v11; // eax
	signed int v12; // eax
	int v13;        // ebp
	signed int v14; // eax
	float v15;      // eax
	int v16;        // eax
	signed int v17; // eax
	int v18;        // ebp
	signed int v19; // eax
	int v20;        // ebp
	float v21;      // ecx
	float v22;      // [esp+4h] [ebp-10h]
	float v23;      // [esp+8h] [ebp-Ch]
	float v24;      // [esp+Ch] [ebp-8h]
	float v25;      // [esp+10h] [ebp-4h]
	int v26;        // [esp+24h] [ebp+10h]
	signed int v27; // [esp+24h] [ebp+10h]
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve((uint32_t)a2);
	if (!fill) {
		return;
	}

	v4 = a4;
	if (*a4 >= 3 && a4[1] >= 3) {
		nox_xxx_tileGetDefByName_51D4D0(fill->tile_name);
		v5 = *a4 - 2;
		v22 = *a3 + 32.526913;
		v23 = a3[1] + 32.526913;
		if (v5 < 1) {
			v5 = 1;
		}
		v7 = a4[1] - 2;
		if (v7 < 1) {
			v7 = 1;
		}
		sub_5245A0(a1, &v22, v5, v7);
		v8 = *a4;
		if ((int)*a4 < 4) {
			if (v8 == 3) {
				v15 = a3[1];
				v22 = *a3 + 32.526913;
				v23 = v15;
				sub_5245A0(a1, &v22, 1, 1);
				v23 = v23 + 65.053825;
				sub_5245A0(a1, &v22, 1, 1);
			}
		} else {
			v9 = nox_xxx_mapGenRandFunc_526AC0(1, v8 - 3);
			v10 = v9;
			v11 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - v9 - 2);
			v23 = a3[1];
			v22 = (double)v11 * 32.526913 + *a3;
			sub_5245A0(a1, &v22, v10, 1);
			v12 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - 3);
			v13 = v12;
			v14 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - v12 - 2);
			v26 = a4[1] - 1;
			v22 = (double)v14 * 32.526913 + *a3;
			v23 = (double)v26 * 32.526913 + a3[1];
			sub_5245A0(a1, &v22, v13, 1);
		}
		v16 = v4[1];
		if (v16 < 4) {
			if (v16 == 3) {
				v21 = *a3;
				v23 = a3[1] + 32.526913;
				v22 = v21;
				sub_5245A0(a1, &v22, 1, 1);
				v22 = v22 + 65.053825;
				sub_5245A0(a1, &v22, 1, 1);
			}
		} else {
			v17 = nox_xxx_mapGenRandFunc_526AC0(1, v16 - 3);
			v18 = v17;
			v19 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - v17 - 2);
			v22 = *a3;
			v23 = (double)v19 * 32.526913 + a3[1];
			sub_524610(a1, &v22, v18);
			v20 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - 3);
			v27 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - v20 - 2);
			v22 = (double)(*v4 - 1) * 32.526913 + *a3;
			v23 = (double)v27 * 32.526913 + a3[1];
			sub_524610(a1, &v22, v20);
		}
		sub_544020(fill->object_name);
		v24 = (double)*v4 * 16.263456 + *a3;
		v25 = (double)v4[1] * 16.263456 + a3[1];
		sub_543680(&v24);
	}
}

//----- (00524E00) --------------------------------------------------------
int sub_526CA0(char* a1);
static void nox_mapgenGenerateDoorsNative_525510(uint8_t* theme, uint8_t* room);

void nox_mapgenGenerateRoomNative_524E00(uint8_t* theme, uint8_t* room) {
	if (!theme || !room || (*(uint8_t*)(room + 52) & 2)) {
		return;
	}
	uint32_t wall_floor_token = (uint32_t)sub_5244D0(*(uint32_t*)(room + 372));
	nox_mapgen_wall_floor_legacy* wall_floor = nox_mapgenWallFloorResolve(wall_floor_token);
	if (!wall_floor) {
		return;
	}
	nox_mapgenGenerateWallFloorNative_524680(theme, room, wall_floor);
	sub_526CA0(wall_floor->wall_name);

	float2 point = {*(float*)(room + 36), *(float*)(room + 40)};
	sub_526D50(8);
	sub_526E60(&point);
	point.field_0 += 32.526913f;
	sub_524500(&point, *(int32_t*)(room + 12) - 1);

	point.field_0 = *(float*)(room + 44);
	point.field_4 = *(float*)(room + 40);
	sub_526D50(9);
	sub_526E60(&point);

	point.field_0 = *(float*)(room + 36);
	point.field_4 = *(float*)(room + 48);
	sub_526D50(7);
	sub_526E60(&point);
	point.field_0 += 32.526913f;
	sub_524500(&point, *(int32_t*)(room + 12) - 1);

	point.field_0 = *(float*)(room + 44);
	point.field_4 = *(float*)(room + 48);
	sub_526D50(10);
	sub_526E60(&point);

	point.field_0 = *(float*)(room + 36);
	point.field_4 = *(float*)(room + 40) + 32.526913f;
	sub_524550((int*)&point, *(int32_t*)(room + 16) - 1);
	point.field_0 = *(float*)(room + 44);
	point.field_4 = *(float*)(room + 40) + 32.526913f;
	sub_524550((int*)&point, *(int32_t*)(room + 16) - 1);

	sub_526C80(0);
	for (int direction = 0; direction < 4; ++direction) {
		for (int index = 0; index < room[216 + direction]; ++index) {
			uint32_t token = *(uint32_t*)(room + 88 + 32 * direction + 4 * index);
			uint8_t* neighbor = (uint8_t*)nox_mapgenLegacyPtrResolve(token);
			if (neighbor) {
				nox_mapgenReserveAdjacentNative_524FB0(room, neighbor, direction);
			}
		}
	}
	sub_526C80(1);
	nox_mapgenGenerateDoorsNative_525510(theme, room);
}

void nox_xxx_gen_524E00(int a1, int a2) {
	nox_mapgenGenerateRoomNative_524E00(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00524FB0) --------------------------------------------------------
int nox_mapgenReserveAdjacentNative_524FB0(uint8_t* room, uint8_t* neighbor, int direction) {
	int result; // eax
	double v4;  // st7
	int v5;     // ecx
	int v6;     // eax
	int v7;     // ebx
	float v8;   // edx
	double v9;  // st7
	float v10;  // ecx
	int v11;    // eax
	int v12;    // ecx
	int v13;    // ebx
	double v14; // st7
	float v15;  // eax
	double v16; // st7
	double v17; // st6
	double v18; // st7
	int v19;    // ecx
	int v20;    // eax
	int v21;    // ebx
	double v22; // st7
	double v23; // st6
	double v24; // st7
	double v25; // st6
	double v26; // st7
	int v27;    // ecx
	int v28;    // eax
	int v29;    // ebx
	double v30; // st7
	double v31; // st6
	double v32; // st7
	float2 a2a; // [esp+Ch] [ebp-10h]
	float2 v34; // [esp+14h] [ebp-8h]

	if (!room || !neighbor) {
		return 0;
	}
	result = direction;
	switch (direction) {
	case 0:
		if (*(float*)(neighbor + 36) >= (double)*(float*)(room + 36)) {
			v4 = *(float*)(neighbor + 36);
		} else {
			v4 = *(float*)(room + 36);
		}
		v5 = *(uint32_t*)(room + 12);
		a2a.field_4 = *(float*)(room + 40);
		v6 = *(uint32_t*)(neighbor + 12);
		a2a.field_0 = v4;
		if (v5 >= v6) {
			v7 = v6;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &a2a, *(float*)(neighbor + 28), 32.526913);
		} else {
			v7 = v5;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &a2a, *(float*)(room + 28), 32.526913);
		}
		v34.field_0 = a2a.field_0 + 32.526913;
		v34.field_4 = a2a.field_4;
		sub_525330((int*)&v34, v7 - 1);
		sub_5253B0(&a2a.field_0);
		v8 = *(float*)(room + 40);
		if (*(float*)(neighbor + 44) <= (double)*(float*)(room + 44)) {
			a2a.field_0 = *(float*)(neighbor + 44);
		} else {
			a2a.field_0 = *(float*)(room + 44);
		}
		a2a.field_4 = v8;
		result = sub_5253B0(&a2a.field_0);
		break;
	case 1:
		if (*(float*)(neighbor + 36) >= (double)*(float*)(room + 36)) {
			v9 = *(float*)(neighbor + 36);
		} else {
			v9 = *(float*)(room + 36);
		}
		a2a.field_0 = v9;
		v10 = *(float*)(room + 48);
		v11 = *(uint32_t*)(neighbor + 12);
		v34.field_0 = v9;
		a2a.field_4 = v10;
		v12 = *(uint32_t*)(room + 12);
		v34.field_4 = a2a.field_4 - 32.526913;
		if (v12 >= v11) {
			v13 = v11;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &v34, *(float*)(neighbor + 28), 32.526913);
		} else {
			v13 = v12;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &v34, *(float*)(room + 28), 32.526913);
		}
		v34.field_0 = a2a.field_0 + 32.526913;
		v34.field_4 = a2a.field_4;
		sub_525330((int*)&v34, v13 - 1);
		sub_5253B0(&a2a.field_0);
		if (*(float*)(neighbor + 44) <= (double)*(float*)(room + 44)) {
			v14 = *(float*)(neighbor + 44);
		} else {
			v14 = *(float*)(room + 44);
		}
		v15 = *(float*)(room + 48);
		a2a.field_0 = v14;
		a2a.field_4 = v15;
		result = sub_5253B0(&a2a.field_0);
		break;
	case 2:
		v16 = *(float*)(room + 40);
		v17 = *(float*)(neighbor + 40);
		a2a.field_0 = *(float*)(room + 44);
		if (v17 >= v16) {
			v18 = *(float*)(neighbor + 40);
		} else {
			v18 = *(float*)(room + 40);
		}
		a2a.field_4 = v18;
		v19 = *(uint32_t*)(room + 16);
		v20 = *(uint32_t*)(neighbor + 16);
		v34.field_0 = a2a.field_0 - 32.526913;
		v34.field_4 = v18;
		if (v19 >= v20) {
			v21 = v20;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &v34, 32.526913, *(float*)(neighbor + 32));
		} else {
			v21 = v19;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &v34, 32.526913, *(float*)(room + 32));
		}
		v34.field_4 = a2a.field_4 + 32.526913;
		v34.field_0 = a2a.field_0;
		sub_525370((int*)&v34, v21 - 1);
		sub_5253B0(&a2a.field_0);
		v22 = *(float*)(room + 48);
		v23 = *(float*)(neighbor + 48);
		a2a.field_0 = *(float*)(room + 44);
		if (v23 <= v22) {
			a2a.field_4 = *(float*)(neighbor + 48);
		} else {
			a2a.field_4 = *(float*)(room + 48);
		}
		result = sub_5253B0(&a2a.field_0);
		break;
	case 3:
		v24 = *(float*)(room + 40);
		v25 = *(float*)(neighbor + 40);
		a2a.field_0 = *(float*)(room + 36);
		if (v25 >= v24) {
			v26 = *(float*)(neighbor + 40);
		} else {
			v26 = *(float*)(room + 40);
		}
		v27 = *(uint32_t*)(room + 16);
		v28 = *(uint32_t*)(neighbor + 16);
		a2a.field_4 = v26;
		if (v27 >= v28) {
			v29 = v28;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &a2a, 32.526913, *(float*)(neighbor + 32));
		} else {
			v29 = v27;
			nox_mapgenAddOccupiedRectNative_521BC0(
				room, &a2a, 32.526913, *(float*)(room + 32));
		}
		v34.field_4 = a2a.field_4 + 32.526913;
		v34.field_0 = a2a.field_0;
		sub_525370((int*)&v34, v29 - 1);
		sub_5253B0(&a2a.field_0);
		v30 = *(float*)(room + 48);
		v31 = *(float*)(neighbor + 48);
		a2a.field_0 = *(float*)(room + 36);
		if (v31 <= v30) {
			v32 = *(float*)(neighbor + 48);
		} else {
			v32 = *(float*)(room + 48);
		}
		a2a.field_4 = v32;
		result = sub_5253B0(&a2a.field_0);
		break;
	default:
		return result;
	}
	return result;
}

int sub_524FB0(int a1, int a2, int a3) {
	return nox_mapgenReserveAdjacentNative_524FB0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2), a3);
}

//----- (00525330) --------------------------------------------------------
float2* sub_525330(float2* a1, int a2) {
	float2* result; // eax
	int v3;         // esi
	float v4;       // edx
	float2 a1a;     // [esp+4h] [ebp-8h]

	result = a1;
	v3 = a2;
	v4 = a1->field_4;
	a1a.field_0 = a1->field_0;
	a1a.field_4 = v4;
	if (a2 > 0) {
		do {
			result = (float2*)sub_527030(&a1a);
			--v3;
			a1a.field_0 = a1a.field_0 + 32.526913;
		} while (v3);
	}
	return result;
}

//----- (00525370) --------------------------------------------------------
float2* sub_525370(float2* a1, int a2) {
	float2* result; // eax
	int v3;         // esi
	float v4;       // edx
	float2 a1a;     // [esp+4h] [ebp-8h]

	result = a1;
	v3 = a2;
	v4 = a1->field_4;
	a1a.field_0 = a1->field_0;
	a1a.field_4 = v4;
	if (a2 > 0) {
		do {
			result = (float2*)sub_527030(&a1a);
			--v3;
			a1a.field_4 = a1a.field_4 + 32.526913;
		} while (v3);
	}
	return result;
}

//----- (005253B0) --------------------------------------------------------
int sub_5253B0(float* a1) {
	double v1;        // st7
	double v2;        // st7
	double v3;        // st7
	double v4;        // st7
	int result;       // eax
	unsigned char v6; // [esp+4h] [ebp-18h]
	float v7;         // [esp+8h] [ebp-14h]
	float v8;         // [esp+Ch] [ebp-10h]
	int v9;           // [esp+10h] [ebp-Ch]

	v1 = a1[1] - 32.526913;
	v6 = 0;
	v7 = *a1;
	v8 = v1;
	if (sub_526DD0(&v7, &v9)) {
		v6 = 1;
	}
	v2 = a1[1] + 32.526913;
	v7 = *a1;
	v8 = v2;
	if (sub_526DD0(&v7, &v9)) {
		v6 |= 2u;
	}
	v3 = *a1 - 32.526913;
	v8 = a1[1];
	v7 = v3;
	if (sub_526DD0(&v7, &v9)) {
		v6 |= 4u;
	}
	v4 = *a1 + 32.526913;
	v8 = a1[1];
	v7 = v4;
	if (sub_526DD0(&v7, &v9)) {
		v6 |= 8u;
	}
	result = v6 - 3;
	switch (v6) {
	case 3u:
		sub_526D50(1);
		return sub_526E60(a1);
	case 5u:
		sub_526D50(10);
		return sub_526E60(a1);
	case 6u:
		sub_526D50(9);
		return sub_526E60(a1);
	case 7u:
		sub_526D50(6);
		return sub_526E60(a1);
	case 9u:
		sub_526D50(7);
		return sub_526E60(a1);
	case 0xAu:
		sub_526D50(8);
		return sub_526E60(a1);
	case 0xBu:
		sub_526D50(4);
		return sub_526E60(a1);
	case 0xCu:
		sub_526D50(0);
		return sub_526E60(a1);
	case 0xDu:
		sub_526D50(3);
		return sub_526E60(a1);
	case 0xEu:
		sub_526D50(5);
		return sub_526E60(a1);
	case 0xFu:
		sub_526D50(2);
		return sub_526E60(a1);
	default:
		return result;
	}
}

//----- (00525510) --------------------------------------------------------
static int nox_mapgenPlaceDoorHorizontalNative_525690(uint8_t* room, float2* point, int span);
static int nox_mapgenPlaceDoubleDoorHorizontalNative_525740(uint8_t* room, float2* point, int span);
static int nox_mapgenPlaceDoorVerticalNative_525830(uint8_t* room, float2* point, int span);
static int nox_mapgenPlaceDoubleDoorVerticalNative_5258E0(uint8_t* room, float2* point, int span);

static void nox_mapgenGenerateDoorBetweenNative_525570(
	uint8_t* theme, uint8_t* room, uint8_t* neighbor, int direction) {
	(void)theme;
	if (!room || !neighbor) {
		return;
	}
	float2 point;
	switch (direction) {
	case 0:
	case 1: {
		point.field_0 = *(float*)(neighbor + 36) >= *(float*)(room + 36)
			? *(float*)(neighbor + 36) : *(float*)(room + 36);
		point.field_4 = direction ? *(float*)(room + 48) : *(float*)(room + 40);
		int span = *(int32_t*)(neighbor + 12);
		if (*(int32_t*)(room + 12) < span) {
			span = *(int32_t*)(room + 12);
		}
		if (span >= 2 &&
			(span == 2 || !nox_mapgenPlaceDoubleDoorHorizontalNative_525740(room, &point, span))) {
			nox_mapgenPlaceDoorHorizontalNative_525690(room, &point, span);
		}
		break;
	}
	case 2:
	case 3: {
		point.field_0 = direction == 3 ? *(float*)(room + 36) : *(float*)(room + 44);
		point.field_4 = *(float*)(neighbor + 40) >= *(float*)(room + 40)
			? *(float*)(neighbor + 40) : *(float*)(room + 40);
		int span = *(int32_t*)(neighbor + 16);
		if (*(int32_t*)(room + 16) < span) {
			span = *(int32_t*)(room + 16);
		}
		if (span >= 2 &&
			(span == 2 || !nox_mapgenPlaceDoubleDoorVerticalNative_5258E0(room, &point, span))) {
			nox_mapgenPlaceDoorVerticalNative_525830(room, &point, span);
		}
		break;
	}
	default:
		break;
	}
}

static void nox_mapgenGenerateDoorsNative_525510(uint8_t* theme, uint8_t* room) {
	if (!theme || !room) {
		return;
	}
	for (int direction = 0; direction < 4; ++direction) {
		for (int index = 0; index < room[216 + direction]; ++index) {
			uint32_t token = *(uint32_t*)(room + 88 + 32 * direction + 4 * index);
			uint8_t* neighbor = (uint8_t*)nox_mapgenLegacyPtrResolve(token);
			if (neighbor) {
				nox_mapgenGenerateDoorBetweenNative_525570(theme, room, neighbor, direction);
			}
		}
	}
}

void nox_xxx_mapgen_525510(int a1, int a2) {
	nox_mapgenGenerateDoorsNative_525510(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (00525570) --------------------------------------------------------
void nox_xxx_mapgen_525570(int a1, int a2, int a3, int a4) {
	nox_mapgenGenerateDoorBetweenNative_525570(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a3), a4);
}

//----- (00525690) --------------------------------------------------------
static int nox_mapgenPlaceDoorHorizontalNative_525690(uint8_t* room, float2* point, int span) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve(*(uint32_t*)(room + 372));
	if (!definition || !definition->door_name[0]) {
		return 0;
	}
	sub_524500(point, span + 1);
	float2 position = {
		(double)(span / 2) * 32.526913 + point->field_0,
		point->field_4,
	};
	sub_527030(&position);
	position.field_0 -= 16.263456f;
	nox_xxx_mapGenGetObjID_527940(definition->door_name);
	nox_object_t* object = (nox_object_t*)nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 5);
	}
	return 1;
}

int nox_xxx_mapgen_525690(int a1, float2* a2, int a3) {
	return nox_mapgenPlaceDoorHorizontalNative_525690(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2, a3);
}

//----- (00525740) --------------------------------------------------------
static int nox_mapgenPlaceDoubleDoorHorizontalNative_525740(uint8_t* room, float2* point, int span) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve(*(uint32_t*)(room + 372));
	if (!definition || !definition->double_door_name[0]) {
		return 0;
	}
	sub_524500(point, span + 1);
	float2 position = {
		(double)(span / 2) * 32.526913 + point->field_0,
		point->field_4,
	};
	sub_527030(&position);
	position.field_0 += 32.526913f;
	sub_527030(&position);
	position.field_0 -= 48.790367f;
	nox_xxx_mapGenGetObjID_527940(definition->double_door_name);
	nox_object_t* object = (nox_object_t*)nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 5);
	}
	position.field_0 += 65.053825f;
	object = (nox_object_t*)nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 3);
	}
	return 1;
}

int nox_xxx_mapgen_525740(int a1, float2* a2, int a3) {
	return nox_mapgenPlaceDoubleDoorHorizontalNative_525740(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2, a3);
}

//----- (00525830) --------------------------------------------------------
static int nox_mapgenPlaceDoorVerticalNative_525830(uint8_t* room, float2* point, int span) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve(*(uint32_t*)(room + 372));
	if (!definition || !definition->door_name[0]) {
		return 0;
	}
	sub_524550((int*)point, span + 1);
	float2 position = {
		point->field_0,
		(double)(span / 2) * 32.526913 + point->field_4,
	};
	sub_527030(&position);
	position.field_4 -= 16.263456f;
	nox_xxx_mapGenGetObjID_527940(definition->door_name);
	nox_object_t* object = (nox_object_t*)nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 7);
	}
	return 1;
}

int nox_xxx_mapgen_525830(int a1, float2* a2, int a3) {
	return nox_mapgenPlaceDoorVerticalNative_525830(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2, a3);
}

//----- (005258E0) --------------------------------------------------------
static int nox_mapgenPlaceDoubleDoorVerticalNative_5258E0(uint8_t* room, float2* point, int span) {
	nox_mapgen_decor_definition_legacy* definition =
		nox_mapgenDecorDefinitionResolve(*(uint32_t*)(room + 372));
	if (!definition || !definition->double_door_name[0]) {
		return 0;
	}
	sub_524550((int*)point, span + 1);
	float2 position = {
		point->field_0,
		(double)(span / 2) * 32.526913 + point->field_4,
	};
	sub_527030(&position);
	position.field_4 += 32.526913f;
	sub_527030(&position);
	position.field_4 -= 48.790367f;
	nox_xxx_mapGenGetObjID_527940(definition->double_door_name);
	nox_object_t* object = (nox_object_t*)nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 7);
	}
	position.field_4 += 65.053825f;
	object = (nox_object_t*)nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 1);
	}
	return 1;
}

int nox_xxx_mapgen_5258E0(int a1, float2* a2, int a3) {
	return nox_mapgenPlaceDoubleDoorVerticalNative_5258E0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2, a3);
}

//----- (005259F0) --------------------------------------------------------
uint8_t* nox_mapgenRoomSortedPrevNative_525C90(const uint8_t* room) {
	return room ? (uint8_t*)nox_mapgenLegacyPtrResolve(*(const uint32_t*)(room + 68)) : NULL;
}

uint8_t* nox_mapgenRoomSortedNextNative_525C90(const uint8_t* room) {
	return room ? (uint8_t*)nox_mapgenLegacyPtrResolve(*(const uint32_t*)(room + 64)) : NULL;
}

static void nox_mapgenSetRoomSortedNextNative_525C90(uint8_t* room, uint8_t* next) {
	if (room) {
		*(uint32_t*)(room + 64) = nox_mapgenLegacyPtrRegister(next);
	}
}

static void nox_mapgenSetRoomSortedPrevNative_525C90(uint8_t* room, uint8_t* prev) {
	if (room) {
		*(uint32_t*)(room + 68) = nox_mapgenLegacyPtrRegister(prev);
	}
}

uint8_t* nox_mapgenFarthestRoomNative_5259E0(void) { return nox_mapgen_farthest_room; }

float nox_mapgenMaxRoomDistanceNative_5259D0(void) { return nox_mapgen_max_room_distance; }

uint8_t* nox_mapgenSortedRoomHeadNative_525C90(void) { return nox_mapgen_sorted_room_head; }

void nox_mapgenComputeRoomDistancesNative_5259F0(uint8_t* room, uint8_t* parent, float distance) {
	if (!room) {
		return;
	}
	nox_xxx_mapGenSetFlags_5235F0(156);
	double edge_distance;
	if (parent) {
		switch (*(uint32_t*)parent) {
		case 1:
			edge_distance = sqrt(
				*(float*)(parent + 28) * *(float*)(parent + 28) +
				*(float*)(parent + 32) * *(float*)(parent + 32));
			break;
		case 2:
		case 3:
			edge_distance = *(float*)(parent + 32);
			break;
		case 4:
		case 5:
			edge_distance = *(float*)(parent + 28);
			break;
		default:
			edge_distance = distance;
			break;
		}
	} else {
		edge_distance = 0.0;
	}
	float candidate = edge_distance + distance;
	if (room[220] == 1 && !(candidate < (double)*(float*)(room + 356))) {
		return;
	}
	*(float*)(room + 356) = candidate;
	room[220] = 1;
	for (int direction = 0; direction < 4; ++direction) {
		uint8_t count = room[216 + direction];
		if (count > 8) {
			count = 8;
		}
		for (uint8_t index = 0; index < count; ++index) {
			uint32_t token = *(uint32_t*)(room + 88 + 32 * direction + 4 * index);
			uint8_t* neighbor = (uint8_t*)nox_mapgenLegacyPtrResolve(token);
			if (neighbor) {
				nox_mapgenComputeRoomDistancesNative_5259F0(neighbor, room, candidate);
			}
		}
	}
}

void sub_5259F0(int a1, int a2, float a3) {
	nox_mapgenComputeRoomDistancesNative_5259F0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2), a3);
}

//----- (00525AF0) --------------------------------------------------------
uint8_t* nox_mapgenClassifyRoomsNative_525AF0(uint8_t* root) {
	if (!root) {
		nox_mapgenResetRoomRankingNative();
		return NULL;
	}
	nox_mapgenSetMaxRoomDistanceNative_5259D0(0.0f);
	nox_mapgenSetFarthestRoomNative_5259E0(root);
	nox_mapgenFindFarthestRoomNative_525BF0(root);
	if (!nox_mapgen_farthest_room) {
		return NULL;
	}
	*(uint32_t*)(nox_mapgen_farthest_room + 52) |= 1u;

	int room_count = 0;
	for (uint8_t* room = nox_mapgen_room_head; room;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		++room_count;
		*(float*)(room + 360) = *(float*)(room + 356) / nox_mapgen_max_room_distance;
	}
	nox_mapgenSortRoomsNative_525C90();

	int rank = 0;
	for (uint8_t* room = root; room; room = nox_mapgenRoomSortedNextNative_525C90(room)) {
		int percentile = room_count > 1 ? rank / (room_count - 1) : 100;
		if (room == root) {
			*(uint32_t*)(room + 364) = 1;
		} else if (percentile >= 90) {
			*(uint32_t*)(room + 364) = 16;
		} else if (percentile >= 60) {
			*(uint32_t*)(room + 364) = 8;
		} else if (percentile >= 30) {
			*(uint32_t*)(room + 364) = 4;
		} else {
			*(uint32_t*)(room + 364) = 2;
		}
		rank += 100;
	}
	*(uint32_t*)(nox_mapgen_farthest_room + 364) = 32;
	return nox_mapgen_farthest_room;
}

float* sub_525AF0(int a1) {
	return (float*)nox_mapgenClassifyRoomsNative_525AF0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (00525BF0) --------------------------------------------------------
void nox_mapgenFindFarthestRoomNative_525BF0(uint8_t* room) {
	if (!room || room[220] == 2) {
		return;
	}
	float distance = *(float*)(room + 356);
	if (nox_mapgen_max_room_distance < (double)distance && *(uint32_t*)room == 1) {
		nox_mapgenSetFarthestRoomNative_5259E0(room);
		nox_mapgenSetMaxRoomDistanceNative_5259D0(distance);
	}
	room[220] = 2;
	for (int direction = 0; direction < 4; ++direction) {
		uint8_t count = room[216 + direction];
		if (count > 8) {
			count = 8;
		}
		for (uint8_t index = 0; index < count; ++index) {
			uint32_t token = *(uint32_t*)(room + 88 + 32 * direction + 4 * index);
			uint8_t* neighbor = (uint8_t*)nox_mapgenLegacyPtrResolve(token);
			if (neighbor) {
				nox_mapgenFindFarthestRoomNative_525BF0(neighbor);
			}
		}
	}
}

void sub_525BF0(int a1) {
	nox_mapgenFindFarthestRoomNative_525BF0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (00525C90) --------------------------------------------------------
uint8_t* nox_mapgenSortRoomsNative_525C90(void) {
	nox_mapgenSetSortedRoomHeadNative_525C90(NULL);
	for (uint8_t* room = nox_mapgen_room_head; room;) {
		uint8_t* master_next = (uint8_t*)nox_mapgenRoomNextNative_521720(room);
		if (!nox_mapgen_sorted_room_head) {
			nox_mapgenSetSortedRoomHeadNative_525C90(room);
			nox_mapgenSetRoomSortedNextNative_525C90(room, NULL);
			nox_mapgenSetRoomSortedPrevNative_525C90(room, NULL);
			room = master_next;
			continue;
		}

		uint8_t* current = nox_mapgen_sorted_room_head;
		uint8_t* previous = NULL;
		while (current && !(*(float*)(room + 356) < (double)*(float*)(current + 356))) {
			previous = current;
			current = nox_mapgenRoomSortedNextNative_525C90(current);
		}
		if (current) {
			uint8_t* current_prev = nox_mapgenRoomSortedPrevNative_525C90(current);
			if (current_prev) {
				nox_mapgenSetRoomSortedNextNative_525C90(current_prev, room);
			} else {
				nox_mapgenSetSortedRoomHeadNative_525C90(room);
			}
			nox_mapgenSetRoomSortedNextNative_525C90(room, current);
			nox_mapgenSetRoomSortedPrevNative_525C90(room, current_prev);
			nox_mapgenSetRoomSortedPrevNative_525C90(current, room);
		} else {
			nox_mapgenSetRoomSortedPrevNative_525C90(room, previous);
			nox_mapgenSetRoomSortedNextNative_525C90(room, NULL);
			nox_mapgenSetRoomSortedNextNative_525C90(previous, room);
		}
		room = master_next;
	}
	return nox_mapgen_sorted_room_head;
}

float* sub_525C90() {
	nox_mapgenSortRoomsNative_525C90();
	return NULL;
}

//----- (00525D20) --------------------------------------------------------
int nox_mapgenInitialPrefabsWithCallback_525D20(uint8_t* a1, nox_mapgen_prefab_room_cb_525d20 callback) {
	int v1;      // edi
	uint32_t* i; // esi
	uint32_t* j; // esi

	if (!a1 || !callback) {
		return 0;
	}
	v1 = 0;
	if (!dword_5d4594_2487656) {
		dword_5d4594_2487656 = nox_xxx_getNameId_4E3AA0("ExitNorthMarker");
		*getMemU32Ptr(0x5D4594, 2487660) = nox_xxx_getNameId_4E3AA0("ExitSouthMarker");
		*getMemU32Ptr(0x5D4594, 2487664) = nox_xxx_getNameId_4E3AA0("ExitEastMarker");
		*getMemU32Ptr(0x5D4594, 2487668) = nox_xxx_getNameId_4E3AA0("ExitWestMarker");
	}
	sub_525DF0(a1);
	for (i = (uint32_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(a1 + 80)); i;
		 i = (uint32_t*)nox_mapgenLegacyPtrResolve(i[39])) {
		if (i[18]) {
			if (v1 == 5 || !callback(a1, (uint8_t*)i)) {
				return 0;
			}
			i[19] = 1;
			++v1;
		}
	}
	for (j = (uint32_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(a1 + 80)); j;
		 j = (uint32_t*)nox_mapgenLegacyPtrResolve(j[39])) {
		if (!j[18]) {
			if (v1 == 5) {
				return 1;
			}
			if (!callback(a1, (uint8_t*)j)) {
				return 0;
			}
			j[19] = 1;
			++v1;
		}
	}
	return 1;
}

int nox_xxx_mapGen_InPrefab1_525D20(uint8_t* a1) {
	return nox_mapgenInitialPrefabsWithCallback_525D20(a1, nox_xxx_mapGenPrefabMkRoom_526100);
}

//----- (00525DF0) --------------------------------------------------------
void sub_525DF0(uint8_t* a1) {
	double v1;         // st7
	int v2;            // eax
	int v3;            // ecx
	unsigned char* v4; // edx
	int v5;            // esi
	signed int v6;     // eax
	int v7;            // ecx
	double v8;         // st7
	signed int v9;     // eax
	double v10;        // st7
	double v11;        // st7
	double v12;        // st7
	double v13;        // st7
	float v14;         // [esp+8h] [ebp-4h]
	float v15;         // [esp+8h] [ebp-4h]
	float v16;         // [esp+10h] [ebp+4h]

	v1 = *(float*)(a1 + 64) * 0.5;
	*getMemU32Ptr(0x5D4594, 2487608) = 0;
	v2 = *(uint32_t*)(a1 + 84);
	v16 = v1;
	if (v2 >= 5) {
		v2 = 5;
		dword_5d4594_2487652 = 5;
	} else {
		dword_5d4594_2487652 = v2;
		if (v2 == 1) {
			*getMemU32Ptr(0x5D4594, 2487588) = 0;
			goto LABEL_10;
		}
		if (v2 == 2) {
			if (nox_xxx_mapGenRandFunc_526AC0(0, 100) >= 50) {
				v2 = dword_5d4594_2487652;
				*getMemU32Ptr(0x5D4594, 2487588) = 1;
				*getMemU32Ptr(0x5D4594, 2487592) = 0;
			} else {
				v2 = dword_5d4594_2487652;
				*getMemU32Ptr(0x5D4594, 2487588) = 0;
				*getMemU32Ptr(0x5D4594, 2487592) = 1;
			}
			goto LABEL_10;
		}
	}
	v3 = 0;
	if (v2 > 0) {
		v4 = getMemAt(0x5D4594, 2487588);
		do {
			*(uint32_t*)v4 = v3++;
			v4 += 4;
		} while (v3 < v2);
	}
	v5 = 0;
	if (v2 > 0) {
		do {
			v6 = nox_xxx_mapGenRandFunc_526AC0(0, v2 - 2);
			v7 = *getMemU32Ptr(0x5D4594, 2487588 + 4 * v6);
			*getMemU32Ptr(0x5D4594, 2487588 + 4 * v6) = *getMemU32Ptr(0x5D4594, 2487592 + 4 * v6);
			*getMemU32Ptr(0x5D4594, 2487592 + 4 * v6) = v7;
			v2 = dword_5d4594_2487652;
			++v5;
		} while (v5 < *(int*)&dword_5d4594_2487652);
	}
LABEL_10:
	switch (v2) {
	case 1:
		v14 = -v16;
		*getMemFloatPtr(0x5D4594, 2487612) = sub_526BC0(v14, v16);
		*getMemFloatPtr(0x5D4594, 2487616) = sub_526BC0(v14, v16);
		break;
	case 2:
		v8 = -v16;
		v15 = v8;
		if (nox_xxx_mapGenRandFunc_526AC0(0, 100) >= 50) {
			*getMemFloatPtr(0x5D4594, 2487612) = v8;
			*getMemFloatPtr(0x5D4594, 2487616) = sub_526BC0(v15, v16);
			*(float*)&dword_5d4594_2487620 = v16;
			*(float*)&dword_5d4594_2487624 = sub_526BC0(v15, v16);
		} else {
			*getMemFloatPtr(0x5D4594, 2487612) = sub_526BC0(v15, v16);
			*getMemFloatPtr(0x5D4594, 2487616) = v15;
			*(float*)&dword_5d4594_2487620 = sub_526BC0(v15, v16);
			*(float*)&dword_5d4594_2487624 = v16;
		}
		break;
	case 3:
		v9 = nox_xxx_mapGenRandFunc_526AC0(0, 100);
		if (v9 >= 25) {
			if (v9 >= 50) {
				v12 = -v16;
				if (v9 >= 75) {
					*getMemFloatPtr(0x5D4594, 2487612) = v12;
					*getMemFloatPtr(0x5D4594, 2487616) = v12;
					*(float*)&dword_5d4594_2487620 = v12;
					*(float*)&dword_5d4594_2487624 = v16;
					*(float*)&dword_5d4594_2487628 = v16;
				} else {
					*getMemFloatPtr(0x5D4594, 2487616) = v12;
					*(float*)&dword_5d4594_2487628 = v12;
					*getMemFloatPtr(0x5D4594, 2487612) = v16;
					*(float*)&dword_5d4594_2487620 = v16;
					*(float*)&dword_5d4594_2487624 = v16;
				}
				dword_5d4594_2487632 = 0;
			} else {
				v11 = -v16;
				*getMemFloatPtr(0x5D4594, 2487612) = v11;
				*(float*)&dword_5d4594_2487632 = v11;
				*getMemFloatPtr(0x5D4594, 2487616) = v16;
				*(float*)&dword_5d4594_2487620 = v16;
				*(float*)&dword_5d4594_2487624 = v16;
				dword_5d4594_2487628 = 0;
			}
		} else {
			v10 = -v16;
			*getMemFloatPtr(0x5D4594, 2487612) = v10;
			*getMemFloatPtr(0x5D4594, 2487616) = v10;
			*(float*)&dword_5d4594_2487624 = v10;
			*(float*)&dword_5d4594_2487620 = v16;
			dword_5d4594_2487628 = 0;
			*(float*)&dword_5d4594_2487632 = v16;
		}
		break;
	case 4:
	case 5:
		*getMemU32Ptr(0x5D4594, 2487644) = 0;
		v13 = -v16;
		*getMemFloatPtr(0x5D4594, 2487612) = v13;
		*getMemFloatPtr(0x5D4594, 2487616) = v13;
		*(float*)&dword_5d4594_2487624 = v13;
		*(float*)&dword_5d4594_2487628 = v13;
		*(float*)&dword_5d4594_2487620 = v16;
		*(float*)&dword_5d4594_2487632 = v16;
		*getMemFloatPtr(0x5D4594, 2487636) = v16;
		*getMemFloatPtr(0x5D4594, 2487640) = v16;
		*getMemU32Ptr(0x5D4594, 2487648) = 0;
		break;
	default:
		return;
	}
}

//----- (00526100) --------------------------------------------------------
int nox_xxx_mapGenPrefabMkRoom_526100(uint8_t* a1, uint8_t* a2) {
	int result;   // eax
	float* v3;    // eax
	int v4;       // edi
	uint32_t* v5; // eax
	int v6;       // edx
	int v7;       // ecx
	uint32_t* v8; // eax
	float2 v9;    // [esp+Ch] [ebp-8h]

	result = sub_5262F0(a1, a2);
	if (result) {
		v3 = nox_xxx_mapGenMakeRoomStruct_521940(
			4 * *(uint32_t*)(a2 + 144) + (unsigned long long)(long long)(*(float*)(a2 + 60) * 0.030743772 + 0.5),
			4 * *(uint32_t*)(a2 + 144) + (unsigned long long)(long long)(*(float*)(a2 + 64) * 0.030743772 + 0.5));
		if (!v3) {
			return 0;
		}
		*(uint32_t*)(a2 + 148) = nox_mapgenLegacyPtrRegister(v3);
		sub_526260((uint8_t*)v3, &v9.field_0);
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)v3, &v9);
		v4 = 1;
		v5 = (uint32_t*)nox_mapgenRoomAtNative_521200((uint8_t*)v3);
		if (v5 && !sub_5212B0((uint8_t*)v3, v5)) {
			v4 = 0;
		}
		if (nox_mapgenRoomWithinThemeNative_5217A0(a1, (uint8_t*)v3) && v4) {
			nox_xxx_mapGenAddNewRoom_521730((uint32_t*)v3);
			v6 = 4;
			*(uint32_t*)((uint8_t*)v3 + 52) |= 2u;
			v7 = 2 * *(uint32_t*)(a2 + 144);
			v8 = (uint32_t*)(a2 + 80);
			do {
				if (v8[3]) {
					*v8 += v7 + *(uint32_t*)((uint8_t*)v3 + 4);
					v8[1] += v7 + *(uint32_t*)((uint8_t*)v3 + 8);
				}
				v8 += 4;
				--v6;
			} while (v6);
			result = 1;
		} else {
			sub_521A10(v3);
			*(uint32_t*)(a2 + 148) = 0;
			result = 0;
		}
	}
	return result;
}

//----- (00526260) --------------------------------------------------------
long long sub_526260(uint8_t* a1, float* a2) {
	int v2;           // ecx
	long long result; // rax

	v2 = *getMemU32Ptr(0x5D4594, 2487588 + 4 * *getMemU32Ptr(0x5D4594, 2487608));
	*a2 = *getMemFloatPtr(0x5D4594, 2487612 + 8 * v2);
	a2[1] = *getMemFloatPtr(0x5D4594, 2487616 + 8 * v2);
	++*getMemU32Ptr(0x5D4594, 2487608);
	*a2 = *a2 - *(float*)(a1 + 28) * 0.5;
	a2[1] = a2[1] - *(float*)(a1 + 32) * 0.5;
	*a2 = (double)(int)(long long)(*a2 * 0.030743772) * 32.526913;
	result = (long long)(a2[1] * 0.030743772);
	a2[1] = (double)(int)result * 32.526913;
	return result;
}

//----- (005262F0) --------------------------------------------------------
int sub_5262F0(uint8_t* theme, uint8_t* prefab) {
	return nox_mapgenAnalyzePrefabNative_5262F0(theme, prefab);
}

uint8_t* nox_mapgenClosestRoomsNative_526550(uint8_t* prefab, int direction) {
	uint8_t* candidates[6] = {0};
	uint64_t distances[6] = {0};
	int count = 0;
	if (!prefab || direction < 0 || direction >= 4) {
		return NULL;
	}

	const int marker_x = *(int32_t*)(prefab + 80 + 16 * direction);
	const int marker_y = *(int32_t*)(prefab + 84 + 16 * direction);
	for (uint8_t* room = (uint8_t*)nox_xxx_mapGenGetTopRoom_521710(); room;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		if (nox_xxx_mapGenCheckRoomType_5238F0((int*)room)) {
			continue;
		}
		const int64_t dx = (int64_t)*(int32_t*)(room + 4) + *(int32_t*)(room + 12) / 2 - marker_x;
		const int64_t dy = (int64_t)*(int32_t*)(room + 8) + *(int32_t*)(room + 16) / 2 - marker_y;
		if ((direction == 0 && dy >= 0) || (direction == 1 && dy <= 0) ||
			(direction == 2 && dx <= 0) || (direction == 3 && dx >= 0)) {
			continue;
		}
		const uint64_t distance = (uint64_t)(dx * dx + dy * dy);
		if (count < 6) {
			candidates[count] = room;
			distances[count] = distance;
			++count;
			continue;
		}
		int farthest = 0;
		for (int i = 1; i < count; ++i) {
			if (distances[i] > distances[farthest]) {
				farthest = i;
			}
		}
		if (distance < distances[farthest]) {
			candidates[farthest] = room;
			distances[farthest] = distance;
		}
	}

	for (int i = 1; i < count; ++i) {
		uint8_t* room = candidates[i];
		uint64_t distance = distances[i];
		int j = i;
		while (j > 0 && distances[j - 1] > distance) {
			candidates[j] = candidates[j - 1];
			distances[j] = distances[j - 1];
			--j;
		}
		candidates[j] = room;
		distances[j] = distance;
	}
	for (int i = 0; i < count; ++i) {
		*(uint32_t*)(candidates[i] + 72) =
			i + 1 < count ? nox_mapgenLegacyPtrRegister(candidates[i + 1]) : 0;
	}
	return count ? candidates[0] : NULL;
}

//----- (00526550) --------------------------------------------------------
int sub_526550(int a1, int a2) {
	uint8_t* first = nox_mapgenClosestRoomsNative_526550(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
	return (int)nox_mapgenLegacyPtrRegister(first);
}

int nox_mapgenConnectPrefabsWithCallback_5266F0(uint8_t* theme, nox_mapgen_prefab_connect_cb_5266f0 callback) {
	if (!theme || !callback) {
		return 0;
	}
	for (uint8_t* prefab = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(theme + 80)); prefab;
		 prefab = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(prefab + 156))) {
		if (!*(uint32_t*)(prefab + 76)) {
			continue;
		}
		uint32_t old_room_token = *(uint32_t*)(prefab + 148);
		uint8_t* old_room = (uint8_t*)nox_mapgenLegacyPtrResolve(old_room_token);
		if (!old_room) {
			return 0;
		}
		float2 position = {
			.field_0 = *(float*)(old_room + 20) + (double)*(int32_t*)(prefab + 144) * 65.053825,
			.field_4 = *(float*)(old_room + 24) + (double)*(int32_t*)(prefab + 144) * 65.053825,
		};
		sub_521760((int)old_room_token);
		sub_521A10(old_room);

		uint8_t* room = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(
			(long long)(*(float*)(prefab + 60) * 0.030743772 + 0.5),
			(long long)(*(float*)(prefab + 64) * 0.030743772 + 0.5));
		uint32_t room_token = nox_mapgenLegacyPtrRegister(room);
		if (!room || !room_token) {
			return 0;
		}
		*(uint32_t*)(prefab + 148) = room_token;
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)room, &position);
		nox_xxx_mapGenAddNewRoom_521730((uint32_t*)room);
		*(uint32_t*)(room + 52) |= 2u;

		for (int direction = 0; direction < 4; ++direction) {
			if (!*(uint32_t*)(prefab + 92 + 16 * direction)) {
				continue;
			}
			uint8_t* candidate = nox_mapgenClosestRoomsNative_526550(prefab, direction);
			while (candidate && !callback(prefab, direction, candidate)) {
				candidate = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(candidate + 72));
			}
			if (!candidate) {
				return 0;
			}
		}
	}
	return 1;
}

static int nox_mapgenConnectPrefabNative_5266F0(uint8_t* prefab, int direction, uint8_t* candidate) {
	return sub_54B2D0(prefab, direction, candidate);
}

//----- (005266F0) --------------------------------------------------------
int nox_xxx_mapGen_InPrefab2_5266F0(uint8_t* theme) {
	return nox_mapgenConnectPrefabsWithCallback_5266F0(theme, nox_mapgenConnectPrefabNative_5266F0);
}

//----- (00526830) --------------------------------------------------------
int nox_mapgenPlacePrefabsWithCallback_526830(
	uint8_t* theme, nox_mapgen_prefab_finalize_cb_526830 callback) {
	if (!theme || !callback) {
		return 0;
	}
	for (uint8_t* prefab = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(theme + 80)); prefab;
		 prefab = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(prefab + 156))) {
		if (!*(uint32_t*)(prefab + 76)) {
			continue;
		}
		uint8_t* foreach_entries = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(prefab + 152));
		uint8_t* room = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(prefab + 148));
		callback(theme, prefab, foreach_entries, room);
	}
	return 1;
}

static void nox_mapgenFinalizePrefabNative_526830(
	uint8_t* theme, uint8_t* prefab, uint8_t* foreach_entries, uint8_t* room) {
	sub_502D70(*(int32_t*)(prefab + 68));
	for (nox_object_t* object = sub_504980(); object;) {
		nox_object_t* next = sub_5049C0(object);
		uint16_t type_id = object->typ_ind;
		if (type_id == dword_5d4594_2487656 || type_id == *getMemU32Ptr(0x5D4594, 2487660) ||
			type_id == *getMemU32Ptr(0x5D4594, 2487664) || type_id == *getMemU32Ptr(0x5D4594, 2487668)) {
			sub_504A10(object);
		}
		object = next;
	}
	nox_mapgenApplyForeachNative_521C60(theme, (uintptr_t)foreach_entries);
	if (room) {
		sub_503B30((float2*)(room + 20));
	}
}

int nox_xxx_mapGenPlacePrefabs_526830(uint8_t* theme) {
	return nox_mapgenPlacePrefabsWithCallback_526830(theme, nox_mapgenFinalizePrefabNative_526830);
}

//----- (005268F0) --------------------------------------------------------
int sub_5268F0(const char* a1) {
	int v1;        // ebp
	const char* i; // edi

	v1 = 0;
	const char* records = (const char*)nox_mapgenLegacyPtrResolve(dword_5d4594_2487672);
	if (*(int*)&dword_5d4594_2487676 <= 0 || !records) {
		return -1;
	}
	for (i = records; strcmp(i, a1); i += 64) {
		if (++v1 >= *(int*)&dword_5d4594_2487676) {
			return -1;
		}
	}
	return v1;
}

//----- (00526950) --------------------------------------------------------
char* sub_526950() {
	int v0;           // eax
	char* result;     // eax
	int v2;           // ebp
	char* v3;         // ebx
	char* v4;         // esi
	int v5;           // edi
	unsigned char v6; // al
	int v7;           // [esp+4h] [ebp-104h]
	char v8[256];     // [esp+8h] [ebp-100h]

	v0 = sub_502A20();
	dword_5d4594_2487676 = v0;
	result = (char*)calloc(v0, 0x40u);
	v2 = 0;
	dword_5d4594_2487672 = nox_mapgenLegacyPtrRegister(result);
	if (result && !dword_5d4594_2487672) {
		free(result);
		result = NULL;
	}
	if (result) {
		v7 = 0;
		if (dword_5d4594_2487676 > 0) {
			while (1) {
				v3 = &result[v2];
				strcpy(v8, (const char*)sub_5029F0(v7));
				strcpy(result + v2, v8);
				strtok(v8, "-");
				strtok(0, "-");
				v4 = strtok(0, "-");
				if (v4) {
					v5 = 0;
					if (getMemByte(0x587000, 255032)) {
						do {
							if (strchr(v4, (char)getMemByte(0x587000, 255032 + v5))) {
								*((uint32_t*)v3 + 15) |= 1 << v5;
							}
							v6 = getMemByte(0x587000, 255033 + v5++);
						} while (v6);
					}
				}
				result = (char*)(v7 + 1);
				v2 += 64;
				if (++v7 >= *(int*)&dword_5d4594_2487676) {
					break;
				}
				result = (char*)nox_mapgenLegacyPtrResolve(dword_5d4594_2487672);
			}
		}
	} else {
		dword_5d4594_2487676 = 0;
	}
	return result;
}

//----- (00526AA0) --------------------------------------------------------
char* sub_526AA0(int a1) {
	char* records = (char*)nox_mapgenLegacyPtrResolve(dword_5d4594_2487672);
	return records ? records + (a1 << 6) : NULL;
}

//----- (00526AB0) --------------------------------------------------------
void nox_xxx_mapGenSetRngSeed_526AB0(unsigned int a1) { nox_platform_srand(a1); }

//----- (00526AC0) --------------------------------------------------------
signed int nox_xxx_mapGenRandFunc_526AC0(int a1, signed int a2) {
	signed int result; // eax

	result = a1 + (a2 - a1 + 1) * nox_platform_rand() / 0x7FFFu;
	if (result > a2) {
		result = a2;
	}
	return result;
}

//----- (00526B00) --------------------------------------------------------
int nox_xxx_mapGenRandFunc2_526B00(int a1, signed int a2) {
	int v3;        // edi
	signed int v4; // ebx
	int v5;        // ecx

	if (!a2) {
		return a1;
	}
	v3 = 10000 / a2;
	do {
		do {
			v4 = nox_xxx_mapGenRandFunc_526AC0(-a2, a2);
			v5 = nox_xxx_mapGenRandFunc_526AC0(0, 10000 / a2);
		} while (v5 >= v3 * (a2 - v4) / a2);
	} while (v5 >= v3 * (v4 + a2) / a2);
	return v4 + a1;
}

//----- (00526BC0) --------------------------------------------------------
double sub_526BC0(float a1, float a2) {
	return (double)(unsigned int)nox_platform_rand() * (a2 - a1) * 0.000030518509 + a1;
}

//----- (00526C40) --------------------------------------------------------
int sub_526C40(int a1) {
	if (a1 != 1 && a1) {
		return 0;
	}
	dword_5d4594_3835364 = a1;
	return 1;
}

//----- (00526C80) --------------------------------------------------------
int sub_526C80(int a1) {
	if (a1 != 1 && a1) {
		return 0;
	}
	dword_5d4594_3835368 = a1;
	return 1;
}

//----- (00526D50) --------------------------------------------------------
int sub_526D50(int a1) {
	int result; // eax

	if (a1 < 0 || a1 >= 15) {
		*getMemU32Ptr(0x973F18, 35952) = 0;
		result = 0;
	} else {
		*getMemU32Ptr(0x973F18, 35952) = a1;
		result = 1;
	}
	return result;
}

//----- (00526DD0) --------------------------------------------------------
int sub_526DD0(float2* a1, int* a2) {
	int v2;       // esi
	long long v3; // rax
	int v4;       // eax
	float2 a2a;   // [esp+8h] [ebp-8h]

	if (!nox_xxx_mapGenFixCoords_4D3D90(a1, &a2a)) {
		return 0;
	}
	if (!a2) {
		return 0;
	}
	v2 = (long long)(a2a.field_0 * 0.043478262);
	v3 = (long long)(a2a.field_4 * 0.043478262);
	if (((uint8_t)v3 + (uint8_t)v2) & 1) {
		return 0;
	}
	if (v2 < 0) {
		return 0;
	}
	if (v2 >= 256) {
		return 0;
	}
	if ((int)v3 < 0) {
		return 0;
	}
	if ((int)v3 >= 256) {
		return 0;
	}
	v4 = nox_server_getWallAtGrid_410580(v2, v3);
	if (!v4) {
		return 0;
	}
	*a2 = v4;
	return 1;
}

//----- (00526E60) --------------------------------------------------------
int sub_526E60(float* a1) {
	int v2;            // edi
	long long v3;      // rax
	int v4;            // ebp
	unsigned char* v5; // eax
	unsigned char* v6; // esi
	unsigned char* v7; // eax
	unsigned char* v8; // esi
	int v9;            // edx
	unsigned char v10; // al
	short v11;         // dx
	int v12;           // [esp-Ch] [ebp-20h]
	int v13;           // [esp-Ch] [ebp-20h]
	int v14;           // [esp-8h] [ebp-1Ch]
	int v15;           // [esp-8h] [ebp-1Ch]
	float2 v16;        // [esp+Ch] [ebp-8h]

	if (*getMemU32Ptr(0x973F18, 35948) == 255) {
		return 1;
	}
	if (nox_xxx_mapGenFixCoords_4D3D90((float2*)a1, &v16)) {
		v2 = (long long)(v16.field_0 * 0.043478262);
		v3 = (long long)(v16.field_4 * 0.043478262);
		v4 = v3;
		if (!(((uint8_t)v2 + (uint8_t)v3) & 1) && v2 >= 0 && v2 < 256 && (int)v3 >= 0 && (int)v3 < 256) {
			v5 = (unsigned char*)nox_server_getWallAtGrid_410580(v2, v3);
			v6 = v5;
			if (v5) {
				v5[1] = getMemByte(0x973F18, 35948);
				if (dword_5d4594_3835368 == 1) {
					v10 = nox_xxx_wall_42A6C0(getMemByte(0x973F18, 35952), *v5);
				} else {
					v10 = getMemByte(0x973F18, 35952);
				}
				*v6 = v10;
				if (dword_5d4594_3835372) {
					v11 = v6[5] % (short)nox_xxx_map_410E00(v6[1]);
				} else {
					LOBYTE(v11) = getMemByte(0x973F18, 35956);
				}
				v15 = *v6;
				v13 = v6[1];
				v6[2] = v11;
				if (v6[2] >= nox_xxx_mapWallMaxVariation_410DD0(v13, v15, 0)) {
					v6[2] = 0;
				}
				v6[4] = -128;
				return 1;
			}
			v7 = (unsigned char*)nox_xxx_wallCreateAt_410250(v2, v4);
			v8 = v7;
			if (v7) {
				v7[1] = getMemByte(0x973F18, 35948);
				*v7 = getMemByte(0x973F18, 35952);
				if (dword_5d4594_3835372) {
					v9 = v2 % nox_xxx_map_410E00(v7[1]);
				} else {
					LOBYTE(v9) = getMemByte(0x973F18, 35956);
				}
				v14 = *v8;
				v12 = v8[1];
				v8[2] = v9;
				v8[4] = -128;
				if (v8[2] >= nox_xxx_mapWallMaxVariation_410DD0(v12, v14, 0)) {
					v8[2] = 0;
					return 1;
				}
				return 1;
			}
		}
	}
	return 0;
}

//----- (00527030) --------------------------------------------------------
int sub_527030(float2* a1) {
	int v1;       // edi
	long long v2; // rax
	int v3;       // ebx
	uint8_t* v4;  // eax
	uint8_t* v5;  // esi
	float2 a2;    // [esp+Ch] [ebp-8h]

	if (!nox_xxx_mapGenFixCoords_4D3D90(a1, &a2)) {
		return 0;
	}
	v1 = (long long)(a2.field_0 * 0.043478262);
	v2 = (long long)(a2.field_4 * 0.043478262);
	v3 = v2;
	if (((uint8_t)v2 + (uint8_t)v1) & 1) {
		return 0;
	}
	if (v1 < 0) {
		return 0;
	}
	if (v1 >= 256) {
		return 0;
	}
	if ((int)v2 < 0) {
		return 0;
	}
	if ((int)v2 >= 256) {
		return 0;
	}
	v4 = nox_server_getWallAtGrid_410580(v1, v2);
	v5 = v4;
	if (!v4) {
		return 0;
	}
	void* wall_data = nox_server_wallData(v4);
	if (wall_data) {
		sub_4107A0(wall_data);
		nox_server_wallSetData(v5, NULL);
	}
	nox_xxx_mapDelWallAtPt_410430(v1, v3);
	return 1;
}

//----- (00527380) --------------------------------------------------------
int sub_527380(float* a1) {
	double v1;    // st7
	int v2;       // esi
	long long v3; // rax
	int v4;       // esi
	int v5;       // esi
	int v6;       // eax
	int v8;       // [esp+8h] [ebp-8h]
	int v9;       // [esp+Ch] [ebp-4h]

	v1 = a1[1] * 0.043478262;
	v2 = (long long)(*a1 * 0.043478262);
	v8 = (long long)(*a1 * 0.043478262);
	v3 = (long long)v1;
	v9 = (long long)v1;
	if (v2 > 0 && v2 < 255 && (int)v3 > 0 && (int)v3 < 255) {
		sub_527450(&v8);
		LODWORD(v3) = v9;
		v2 = v8;
	}
	v4 = v2 - 1;
	v8 = v4;
	if (v4 > 0 && v4 < 255 && (int)v3 > 0 && (int)v3 < 255) {
		sub_527450(&v8);
		LODWORD(v3) = v9;
		v4 = v8;
	}
	v5 = v4 + 1;
	v6 = v3 - 1;
	v8 = v5;
	v9 = v6;
	if (v5 > 0 && v5 < 255 && v6 > 0 && v6 < 255) {
		sub_527450(&v8);
	}
	return 1;
}

//----- (00527450) --------------------------------------------------------
int sub_527450(uint32_t* a1) {
	int v1;             // ebp
	int v2;             // ebp
	int v3;             // eax
	int result;         // eax
	int v5;             // esi
	int v6;             // ebx
	int v7;             // edi
	uint8_t* v8;        // eax
	uint8_t* v9;        // esi
	int v10;            // edi
	int v11;            // ecx
	long long v12;      // rax
	int v13;            // kr00_4
	int v14;            // ebx
	int v15;            // ebp
	int v16;            // edi
	int v17;            // ecx
	unsigned char* v18; // eax
	unsigned char* v19; // esi
	int v20;            // ecx
	unsigned char v21;  // dl
	unsigned char v22;  // al
	unsigned char* v23; // eax
	unsigned char v24;  // cl
	int v25;            // eax
	int v26;            // ecx
	int v27;            // eax
	unsigned char* v28; // eax
	uint8_t* v29;       // esi
	int v30;            // ecx
	unsigned char v31;  // dl
	unsigned char v32;  // al
	unsigned char v33;  // al
	int v34;            // [esp+10h] [ebp-10h]
	int v35;            // [esp+14h] [ebp-Ch]
	int v36;            // [esp+18h] [ebp-8h]
	int v37;            // [esp+1Ch] [ebp-4h]
	int v38;            // [esp+24h] [ebp+4h]
	int v39;            // [esp+24h] [ebp+4h]

	v1 = *a1 - 1;
	if (v1 > 0) {
		v2 = v1 & 0xFFFE;
		v36 = v2;
	} else {
		v36 = 0;
		v2 = 0;
	}
	v3 = a1[1] - 1;
	if (v3 > 0) {
		result = v3 & 0xFFFE;
		v38 = result;
	} else {
		v38 = 0;
		result = 0;
	}
	v5 = result + 4;
	v6 = result;
	v37 = result + 4;
	while (v6 <= v5) {
		v7 = v2;
		while (v7 <= v2 + 4) {
			v8 = nox_server_getWallAtGrid_410580(v7, v6);
			v9 = v8;
			if (v8 && !(v8[4] & 0x8C)) {
				void* wall_data = nox_server_wallData(v8);
				if (wall_data) {
					sub_4107A0(wall_data);
					nox_server_wallSetData(v9, NULL);
				}
				nox_xxx_mapDelWallAtPt_410430(v9[5], v9[6]);
			}
			++v7;
		}
		v5 = v37;
		result = v38;
		++v6;
	}
	v10 = result;
	v35 = result;
	if (result > v5) {
		return result;
	}
	v11 = v2 + 4;
	do {
		v34 = v2;
		v12 = v10;
		v13 = v12;
		result = v12 - HIDWORD(v12);
		v14 = v13 / 2;
		if (v2 > v11) {
			goto LABEL_72;
		}
		do {
			v15 = 0;
			v16 = v34 / 2;
			if (v34 / 2 >= 128) {
				goto LABEL_70;
			}
			if (v14 < 128) {
				v17 = 44 * v14 + (uint32_t)(ptr_5D4594_2650668[v16]);
				if (*(uint32_t*)(v17 + 24) != 255) {
					v15 = 8;
				}
				if (*(uint32_t*)(v17 + 4) != 255) {
					v15 |= 2u;
				}
				if (v14 > 0 && *(uint32_t*)(v17 - 20) != 255) {
					v15 |= 4u;
				}
				if (v16 > 0 && *(uint32_t*)((uint32_t)(ptr_5D4594_2650668[v16 - 1]) + 44 * v14 + 4) != 255) {
					v15 |= 1u;
				}
				if (*getMemU32Ptr(0x587000, 255052 + 4 * v15) != 255) {
					v18 = (unsigned char*)nox_server_getWallAtGrid_410580(v34, v35);
					v19 = v18;
					if (v18) {
						if (v18[4] & 0x8C) {
							goto LABEL_45;
						}
						if (*v18 == *getMemU32Ptr(0x587000, 255052 + 4 * v15)) {
							goto LABEL_45;
						}
						v20 = v18[1];
						if (v20 == *getMemU32Ptr(0x973F18, 35948)) {
							goto LABEL_45;
						}
						v21 = getMemByte(0x587000, 255052 + 4 * v15);
						v18[2] = 0;
						*v18 = v21;
						if (v21) {
							if (v21 != 1) {
								goto LABEL_45;
							}
						}
						v22 = nox_xxx_map_410E00(v20);
					} else {
						v23 = (unsigned char*)nox_xxx_wallCreateAt_410250(v34, v35);
						v19 = v23;
						if (!v23) {
							goto LABEL_45;
						}
						v24 = getMemByte(0x587000, 255052 + 4 * v15);
						v23[2] = 0;
						*v23 = v24;
						v23[1] = getMemByte(0x973F18, 35948);
						if (v24) {
							if (v24 != 1) {
								goto LABEL_45;
							}
						}
						v22 = nox_xxx_map_410E00(v23[1]);
					}
					v19[2] = v19[5] % (short)v22;
				}
			}
		LABEL_45:
			if (v14 < 128) {
				v39 = 0;
				if (v14 > 0 && *(uint32_t*)((uint32_t)(ptr_5D4594_2650668[v16]) + 44 * v14 - 20) != 255) {
					v39 = 2;
				}
				if (v16 > 0) {
					v25 = 44 * v14;
					v26 = ptr_5D4594_2650668[v16 - 1];
					if (*(uint32_t*)(v26 + 44 * v14 + 4) != 255) {
						v39 |= 8u;
					}
					if (v14 > 0) {
						if (*(uint32_t*)(v26 + v25 - 40) != 255) {
							v39 |= 4u;
						}
						if (*(uint32_t*)(v26 + v25 - 20) != 255) {
							v27 = v39;
							LOBYTE(v27) = v39 | 1;
							v39 = v27;
						}
					}
				}
				if (*getMemU32Ptr(0x587000, 255052 + 4 * v39) != 255) {
					v28 = (unsigned char*)nox_server_getWallAtGrid_410580(v34 - 1, v35 - 1);
					v29 = v28;
					if (v28) {
						if (!(v28[4] & 0x8C) && *v28 != *getMemU32Ptr(0x587000, 255052 + 4 * v39)) {
							v30 = v28[1];
							if (v30 != *getMemU32Ptr(0x973F18, 35948)) {
								v31 = getMemByte(0x587000, 255052 + 4 * v39);
								v28[2] = 0;
								*v28 = v31;
								if (!v31 || v31 == 1) {
									v32 = nox_xxx_map_410E00(v30);
									v29[2] = (unsigned char)v29[5] % (short)v32;
									goto LABEL_70;
								}
							}
						}
					} else {
						v29 = nox_xxx_wallCreateAt_410250(v34 - 1, v35 - 1);
						if (v29) {
							v33 = getMemByte(0x587000, 255052 + 4 * v39);
							v29[2] = 0;
							*v29 = v33;
							v29[1] = getMemByte(0x973F18, 35948);
							if (!v33 || v33 == 1) {
								v32 = nox_xxx_map_410E00((unsigned char)v29[1]);
								v29[2] = (unsigned char)v29[5] % (short)v32;
								goto LABEL_70;
							}
						}
					}
				}
			}
		LABEL_70:
			result = v34 + 2;
			v11 = v36 + 4;
			v34 = result;
		} while (result <= v36 + 4);
		v5 = v37;
		v10 = v35;
		v2 = v36;
	LABEL_72:
		v10 += 2;
		v35 = v10;
	} while (v10 <= v5);
	return result;
}

//----- (00527940) --------------------------------------------------------
int nox_xxx_mapGenGetObjID_527940(char* a1) {
	int result; // eax

	if (nox_strcmpi(a1, "NONE")) {
		if (a1) {
			dword_5d4594_3835388 = nox_xxx_getNameId_4E3AA0(a1);
			result = 1;
		} else {
			result = 0;
		}
	} else {
		dword_5d4594_3835388 = 0;
		result = 1;
	}
	return result;
}

//----- (005279B0) --------------------------------------------------------
float* nox_xxx_mapGenPlaceObj_5279B0(float2* a1) {
	float* result; // eax
	float2 a2;     // [esp+4h] [ebp-8h]

	if (!dword_5d4594_3835388) {
		return 0;
	}
	result = (float*)nox_xxx_mapGenFixCoords_4D3D90(a1, &a2);
	if (result) {
		result = (float*)nox_xxx_newObjectWithTypeInd_4E3450(*(int*)&dword_5d4594_3835388);
		if (result) {
			result = nox_xxx_mapGenMoveObject_527A10(result, &a1->field_0);
		}
	}
	return result;
}

//----- (00527A10) --------------------------------------------------------
void* nox_objectTypeGetXfer(char* id);
float* nox_xxx_mapGenMoveObject_527A10(float* a1, float2* a2) {
	float v3;  // ecx
	float v4;  // eax
	int v5;    // ecx
	char* v6;  // eax
	float2 v7; // [esp+4h] [ebp-8h]

	if (!nox_xxx_mapGenFixCoords_4D3D90(a2, &v7)) {
		return a1;
	}
	if (!a1) {
		return 0;
	}
	v3 = v7.field_4;
	a1[10] = *(float*)&dword_5d4594_3835392;
	v4 = v7.field_0;
	++dword_5d4594_3835392;
	a1[15] = v3;
	v5 = *((uint32_t*)a1 + 4) | 0x1000000;
	a1[14] = v4;
	*((uint32_t*)a1 + 4) = v5;
	v6 = (char*)nox_xxx_getUnitName_4E39D0((int)a1);
	if (nox_objectTypeGetXfer(v6) == nox_xxx_XFerDoor_4F4CB0) {
		a1[14] = (double)(int)(23 * (unsigned long long)(long long)(v7.field_0 * 0.043478262 + 0.5));
		a1[15] = (double)(int)(23 * (unsigned long long)(long long)(v7.field_4 * 0.043478262 + 0.5));
	}
	nox_xxx_createAt_4DAA50((int)a1, 0, a1[14], a1[15]);
	nox_xxx_unitClearPendingMB_4DB030();
	return a1;
}

//----- (00527C60) --------------------------------------------------------
int nox_mapgenOrientObjNative_527C60(nox_object_t* object, int direction) {
	if (!object) {
		return 0;
	}
	if ((object->obj_class & 2) == 2) {
		if (!object->data_update) {
			return 0;
		}
		int direction4 = sub_4D3FF0(direction);
		int angle = nox_xxx_mathDirection4ToAngle_509E90(direction4);
		*(uint32_t*)((uint8_t*)object->data_update + 376) = angle;
		object->direction1 = angle;
		return 1;
	}
	if (!(object->obj_class & 0x80) || !object->data_update) {
		return 0;
	}
	uint32_t* update = (uint32_t*)object->data_update;
	uint32_t frame;
	switch (direction) {
	case 1:
		frame = 0;
		break;
	case 3:
		frame = 24;
		break;
	case 5:
		frame = 8;
		break;
	case 7:
		frame = 16;
		break;
	default:
		return 0;
	}
	update[1] = frame;
	update[2] = frame;
	update[3] = frame;
	return 1;
}

int nox_xxx_mapGenOrientObj_527C60(int a1, int a2) {
	return nox_mapgenOrientObjNative_527C60(
		(nox_object_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (00527DB0) --------------------------------------------------------
void* nox_objectTypeGetXfer(char* id);
static int nox_mapgenFinishSpellbookNative_527DB0(nox_object_t* object, char spell_id) {
	if (!object) {
		return 0;
	}
	char* name = nox_xxx_getUnitName_4E39D0(object);
	if (nox_objectTypeGetXfer(name) != nox_xxx_XFerSpellReward_4F5F30 || !object->use_data) {
		return 0;
	}
	*(uint8_t*)object->use_data = (uint8_t)spell_id;
	return 1;
}

int nox_xxx_mapGenFinishSpellbook_527DB0(int a1, char a2) {
	return nox_mapgenFinishSpellbookNative_527DB0(
		(nox_object_t*)(uintptr_t)(uint32_t)a1, a2);
}

//----- (00527E50) --------------------------------------------------------
int nox_xxx_netUpdateObjectSpecial_527E50(nox_object_t* a1p, nox_object_t* a2p) {
	int a1 = a1p;
	uint32_t* a2 = a2p;
	unsigned int v2; // edi
	int v3;          // eax

	v2 = *(unsigned char*)(*(uint32_t*)(*(uint32_t*)(a1 + 748) + 276) + 2064);
	if (a2 && v2 < 0x20) {
		v3 = a2[v2 + 140];
		if (!(v3 & 0xFFF0000)) {
			return 0;
		}
		if ((v3 & 0x10000) == 0x10000) {
			nox_xxx_netReportAnimFrame_4D81F0(v2, a2);
			a2[v2 + 140] &= 0xFFFEFFFF;
		}
		if ((a2[v2 + 140] & 0x20000) == 0x20000) {
			if (!nox_xxx_unitIsEnemyTo_5330C0(a1, (int)a2)) {
				nox_xxx_netReportUnitCurrentHP_4D8620(v2, a2);
			}
			a2[v2 + 140] &= 0xFFFDFFFF;
		}
		if ((a2[v2 + 140] & 0x40000) == 0x40000) {
			nox_xxx_netReportObjHidden_4D8FD0(v2, a2);
			a2[v2 + 140] &= 0xFFFBFFFF;
		}
		if ((a2[v2 + 140] & 0x80000) == 0x80000) {
			nox_xxx_netReportXStatus_4D8230(v2, a2);
			a2[v2 + 140] &= 0xFFF7FFFF;
		}
		if ((a2[v2 + 140] & 0x400000) == 0x400000) {
			nox_xxx_netReportUnitHeight_4D9020(v2, (int)a2);
			a2[v2 + 140] &= 0xFFBFFFFF;
		}
		if ((0x800000 & a2[v2 + 140]) == 0x800000) {
			nox_xxx_netReportEnchant_4D8F90(v2, a2);
			a2[v2 + 140] &= 0xFF7FFFFF;
		}
		if ((a2[v2 + 140] & 0x2000000) == 0x2000000) {
			nox_xxx_netReportTeamBase_4D92D0(v2, (int)a2);
			a2[v2 + 140] &= 0xFDFFFFFF;
		}
		if ((a2[v2 + 140] & 0x4000000) == 0x4000000) {
			nox_xxx_netSendReportNPC_4D93A0(v2, (int)a2);
			a2[v2 + 140] &= 0xFBFFFFFF;
		}
	}
	return 1;
}

//----- (00528030) --------------------------------------------------------
short sub_528030(int a1) {
	int v1;             // ebp
	int v2;             // esi
	unsigned short* v3; // edi
	int v4;             // ebx
	int v5;             // edi
	int v6;             // edx
	unsigned int v7;    // ebx
	int v8;             // eax
	short result;       // ax
	int v10;            // edx
	int v11;            // [esp+18h] [ebp+4h]

	v1 = a1;
	v2 = *(uint32_t*)(a1 + 748);
	v3 = *(unsigned short**)(a1 + 556);
	v4 = *(uint32_t*)(v2 + 276);
	v11 = *(unsigned char*)(v4 + 2064);
	if (*(uint16_t*)(v2 + 10) == *v3) {
		v5 = gameFrame();
		v7 = gameFPS();
	} else if (abs(*v3 - *(unsigned short*)(v2 + 10)) >= v3[2] / 10 ||
		(v5 = gameFrame(), v6 = *(uint32_t*)(v4 + 2176), v7 = gameFPS(),
		 (unsigned int)(gameFrame() - v6) > (int)gameFPS() >> 2)) {
		nox_xxx_netSendPlrHealthToTeam_4D86E0(v11);
		v8 = *(uint32_t*)(v2 + 276);
		*(uint16_t*)(v2 + 10) = **(uint16_t**)(v1 + 556);
		*(uint32_t*)(v8 + 2176) = gameFrame();
		v5 = gameFrame();
		v7 = gameFPS();
	}
	result = *(uint16_t*)(v2 + 4);
	if (*(uint16_t*)(v2 + 6) != result) {
		result = 0;
		if (abs(*(unsigned short*)(v2 + 4) - *(unsigned short*)(v2 + 6)) >= *(unsigned short*)(v2 + 8) / 10 ||
			v5 - *(uint32_t*)(*(uint32_t*)(v2 + 276) + 2180) > v7 >> 2) {
			nox_xxx_netReportMana_4D8930(v11, v1);
			v10 = *(uint32_t*)(v2 + 276);
			*(uint16_t*)(v2 + 6) = *(uint16_t*)(v2 + 4);
			result = (unsigned short)gameFrame();
			*(uint32_t*)(v10 + 2180) = gameFrame();
		}
	}
	return result;
}

//----- (00528190) --------------------------------------------------------
int nox_xxx_checkIsKillable_528190(nox_object_t* a1p) {
	int a1 = a1p;
	uint16_t* v1; // eax
	bool v2;      // zf

	v1 = *(uint16_t**)(a1 + 556);
	if (!v1) {
		return 0;
	}
	if (*v1) {
		v2 = v1[2] == 0;
		if (v1[2]) {
			return 1;
		}
	} else {
		v2 = v1[2] == 0;
	}
	if (v2) {
		return 1;
	} else {
		return 0;
	}
}

//----- (005281D0) --------------------------------------------------------
int nox_xxx_frameCounterSetCopyToNextFrame_5281D0() {
	int result; // eax

	result = gameFrame() + 1;
	*getMemU32Ptr(0x5D4594, 2487684) = gameFrame() + 1;
	return result;
}

//----- (005281E0) --------------------------------------------------------
int nox_xxx_frameCounterSetCopy_5281E0() {
	int result; // eax

	result = gameFrame();
	*getMemU32Ptr(0x5D4594, 2487684) = gameFrame();
	return result;
}

//----- (005281F0) --------------------------------------------------------
int nox_xxx_unitCanSee_536FB0(nox_object_t* a1, nox_object_t* a2, int a3);
#if 0
// Original PE32 oracle implementation. It is intentionally retained beside
// the native bridge: offsets 1132..1192 and 1216 are object-pointer slots and
// therefore cannot be dereferenced as uint32_t on 64-bit targets.
void nox_xxx_unitUpdateSightMB_5281F0(nox_object_t* a1p) {
	uint32_t a1 = a1p;
	uint32_t v1;   // edi
	int v2;     // eax
	int v3;     // ebp
	double v4;  // st7
	int v5;     // esi
	int* v6;    // ebx
	double v7;  // st7
	double v8;  // st6
	double v9;  // st7
	double v10; // st6
	int v11;    // eax
	int v12;    // eax
	int v13;    // esi
	int v14;    // eax
	int v15;    // eax
	int v16;    // esi
	int v17;    // [esp+10h] [ebp-10h]
	float v18;  // [esp+10h] [ebp-10h]
	int v19;    // [esp+14h] [ebp-Ch]
	float v20;  // [esp+18h] [ebp-8h]
	float v21;  // [esp+24h] [ebp+4h]

	v1 = a1;
	v17 = 0;
	v2 = *(uint32_t*)(a1 + 16);
	v3 = *(uint32_t*)(a1 + 748);
	if ((v2 & 0x8000) != 0 && !nox_xxx_unitIsZombie_534A40(a1)) {
		return;
	}
	if (nox_common_gameFlags_check_40A5C0(4096)) {
		v4 = 640.0;
	} else {
		v4 = 250.0;
	}
	if (v4 >= *(float*)(v3 + 1312)) {
		v21 = v4;
	} else {
		v21 = *(float*)(v3 + 1312);
	}
	if (gameFrame() - *(uint32_t*)(v3 + 1212) <= (unsigned int)(2 * gameFPS())) {
		v19 = 0;
	} else {
		v19 = 1;
		*(uint32_t*)(v3 + 1212) = gameFrame();
	}
	v5 = 0;
	if (*(uint8_t*)(v3 + 1129)) {
		v6 = (int*)(v3 + 1132);
		do {
			if (*(uint32_t*)(*v6 + 16) & 0x8020 || !nox_xxx_unitCanSee_536FB0(v1, *v6, 0) ||
				(v7 = *(float*)(v1 + 56) - *(float*)(*v6 + 56),
				 v8 = *(float*)(v1 + 60) - *(float*)(*v6 + 60), v20 = (v21 + 30.0) * (v21 + 30.0),
				 v8 * v8 + v7 * v7 > v20) ||
				(v9 = *(float*)(v1 + 56) - *(float*)(v1 + 72),
				 v10 = *(float*)(v1 + 60) - *(float*)(v1 + 76), v10 * v10 + v9 * v9 > 1000.0) ||
				v19 && !nox_xxx_unitCanInteractWith_5370E0(v1, *v6, 0)) {
				nox_xxx_aiLostSight_528560(v1, v5--);
				v17 = 1;
				--v6;
			}
			++v5;
			++v6;
		} while (v5 < *(unsigned char*)(v3 + 1129));
	}
	v11 = *(uint32_t*)(v3 + 1196);
	if (v11 && nox_xxx_testUnitBuffs_4FF350(v11, 28)) {
		v17 = 1;
	}
	if ((!*(uint32_t*)(v3 + 1196) ||
		 gameFrame() - *(uint32_t*)(v3 + 1204) > (unsigned int)(2 * gameFPS())) &&
		(*(uint32_t*)(v3 + 1208) <= gameFrame() ||
		 gameFrame() == *getMemU32Ptr(0x5D4594, 2487684))) {
		nox_xxx_unitsGetInCircle_517F90((float2*)(v1 + 56), v21, nox_xxx_monsterUpdateSeenEnemies_5286D0, v1);
		*(uint32_t*)(v3 + 1204) = gameFrame();
		*(uint32_t*)(v3 + 1212) = gameFrame();
		v17 = 1;
	}
	if (v17) {
		v12 = *(uint32_t*)(v3 + 1196);
		if (v12) {
			v13 = *(uint32_t*)(v12 + 36);
		} else {
			v13 = 0;
		}
		sub_528610(v1);
		v14 = *(uint32_t*)(v3 + 1196);
		if (v14 && v13 && v13 != *(uint32_t*)(v14 + 36)) {
			*(uint32_t*)(v3 + 1200) = v13;
		}
	}
	if (*(uint32_t*)(v3 + 1204) == gameFrame()) {
		v15 = *(uint32_t*)(v3 + 1440);
		if (v15 & 0x400 || nox_common_gameFlags_check_40A5C0(0x2000) || *(uint32_t*)(v3 + 1196)) {
			*(uint32_t*)(v3 + 1208) = gameFrame() + nox_common_randomInt_415FA0(5, 10);
		} else {
			v16 = 5 * gameFPS();
			v18 = sub_5336D0(v1);
			*(float*)(v3 + 524) = v18;
			if (v18 < 0.0) {
				*(uint32_t*)(v3 + 1208) = v16 + gameFrame();
			} else if (v18 > (double)v21) {
				*(uint32_t*)(v3 + 1208) = (unsigned long long)(long long)((v18 - v21) * (double)v16 / (1000.0 - v21)) +
										  10 + gameFrame();
			} else {
				*(uint32_t*)(v3 + 1208) = nox_common_randomInt_415FA0(5, 10) + gameFrame();
			}
		}
	}
}

//----- (00528560) --------------------------------------------------------
int nox_xxx_aiLostSight_528560(int a1, int a2) {
	int v2;     // esi
	int v3;     // eax
	int* v4;    // edi
	int v5;     // eax
	int v6;     // eax
	int v7;     // edx
	int v8;     // ecx
	int result; // eax
	int v10;    // [esp-4h] [ebp-14h]

	v2 = *(uint32_t*)(a1 + 748);
	v3 = *(uint32_t*)(v2 + 4 * a2 + 1132);
	v4 = (int*)(v2 + 4 * a2 + 1132);
	v10 = *(uint32_t*)(v3 + 36);
	v5 = nox_xxx_getUnitName_4E39D0(v3);
	nox_ai_debug_printf_5341A0("%d: Lost sight of %s(#%d)\n", gameFrame(), v5, v10);
	nox_xxx_scriptCallByEventBlock_502490((int*)(v2 + 1296), *v4, a1, 15);
	v6 = *(uint32_t*)(v2 + 1196);
	if (*v4 == v6) {
		v7 = *(uint32_t*)(v6 + 36);
		*(uint32_t*)(v2 + 1196) = 0;
		*(uint32_t*)(v2 + 1200) = v7;
	}
	v8 = a2;
	LOBYTE(result) = *(uint8_t*)(v2 + 1129) - 1;
	*(uint8_t*)(v2 + 1129) = result;
	result = (unsigned char)result;
	if (a2 < (unsigned char)result) {
		result = v2 + 4 * a2 + 1132;
		do {
			++v8;
			*(uint32_t*)result = *(uint32_t*)(result + 4);
			result += 4;
		} while (v8 < *(unsigned char*)(v2 + 1129));
	}
	return result;
}

//----- (00528610) --------------------------------------------------------
void sub_528610(int a1) {
	int v1;    // ebx
	int v2;    // esi
	int v3;    // ebp
	char v4;   // al
	int* i;    // edi
	double v6; // st7
	double v7; // st6
	double v8; // st5
	float v9;  // [esp+14h] [ebp+4h]

	v1 = a1;
	v2 = *(uint32_t*)(a1 + 748);
	v3 = 0;
	v9 = 100000000.0;
	v4 = *(uint8_t*)(v2 + 1129);
	*(uint32_t*)(v2 + 1196) = 0;
	if (v4) {
		for (i = (int*)(v2 + 1132); *i != *(uint32_t*)(v2 + 1216); ++i) {
			if (nox_xxx_unitIsEnemyTo_5330C0(v1, *i) && nox_xxx_checkIsKillable_528190(*i)) {
				v6 = *(float*)(*i + 56) - *(float*)(v1 + 56);
				v7 = *(float*)(*i + 60) - *(float*)(v1 + 60);
				v8 = v7 * v7 + v6 * v6;
				if (v8 < v9) {
					v9 = v8;
					*(uint32_t*)(v2 + 1196) = *i;
				}
			}
			if (++v3 >= *(unsigned char*)(v2 + 1129)) {
				return;
			}
		}
		*(uint32_t*)(v2 + 1196) = *(uint32_t*)(v2 + 1216);
	}
}

//----- (005286D0) --------------------------------------------------------
void nox_xxx_monsterUpdateSeenEnemies_5286D0(int a1, int a2) {
	int v2;    // esi
	int v3;    // ebx
	int v4;    // eax
	int v5;    // eax
	double v6; // st7
	double v7; // st6
	float* v8; // eax
	float v9;  // [esp+10h] [ebp-4h]
	float v10; // [esp+1Ch] [ebp+8h]

	v2 = a2;
	v3 = *(uint32_t*)(a2 + 748);
	if (a2 != a1) {
		if (*(uint8_t*)(a1 + 8) & 6) {
			if (!(*(uint32_t*)(a1 + 16) & 0x8020)) {
				v4 = *(uint32_t*)(v3 + 1440);
				if ((v4 & 0x400 || nox_xxx_unitIsEnemyTo_5330C0(a2, a1)) && !sub_528950(a2, a1)) {
					v5 = *(uint32_t*)(v3 + 1440);
					if (v5 & 0x100 ||
						(v6 = *(float*)(a1 + 56) - *(float*)(a2 + 56), v7 = *(float*)(a1 + 60) - *(float*)(a2 + 60),
						 v9 = v7, v8 = getMemFloatPtr(0x587000, 194136 + 8 * *(short*)(a2 + 124)),
						 v10 = sqrt(v7 * v9 + v6 * v6) + 0.001, v9 / v10 * v8[1] + v6 / v10 * *v8 >= 0.5)) {
						if (nox_xxx_unitCanInteractWith_5370E0(v2, a1, 0)) {
							nox_xxx_monsterVisionSeeEnemy_5287B0(v2, a1);
						}
					}
				}
			}
		}
	}
}

//----- (005287B0) --------------------------------------------------------
void nox_xxx_monsterVisionSeeEnemy_5287B0(int a1, int a2) {
	int v2;     // esi
	double v3;  // st7
	int v4;     // edi
	int v5;     // ebx
	int v6;     // ebp
	int v7;     // edx
	double v8;  // st6
	double v9;  // st5
	double v10; // st4
	int v11;    // ebx
	double v12;
	double v13;
	unsigned int v15; // eax
	int v16;          // eax
	int v17;          // eax
	float v18;        // [esp+14h] [ebp+4h]

	v2 = a1;
	v3 = 0.0;
	v4 = *(uint32_t*)(a1 + 748);
	v5 = 16;
	v6 = 0;
	if (*(uint8_t*)(v4 + 1129) == 16) {
		v7 = v4 + 1132;
		do {
			v8 = *(float*)(*(uint32_t*)v7 + 56) - *(float*)(v2 + 56);
			v9 = *(float*)(*(uint32_t*)v7 + 60) - *(float*)(v2 + 60);
			v10 = v9 * v9 + v8 * v8;
			if (v10 > v3) {
				v18 = v10;
				v3 = v18;
				v6 = *(uint32_t*)v7;
			}
			v7 += 4;
			--v5;
		} while (v5);
		v11 = a2;
		v12 = *(float*)(a2 + 56) - *(float*)(v2 + 56);
		v13 = *(float*)(a2 + 60) - *(float*)(v2 + 60);
		if (v3 <= v13 * v13 + v12 * v12) {
			return;
		}
		sub_528910(v2, v6);
	} else {
		v11 = a2;
	}
	*(uint32_t*)(v4 + 4 * *(unsigned char*)(v4 + 1129) + 1132) = v11;
	v15 = *(uint32_t*)(v4 + 536);
	++*(uint8_t*)(v4 + 1129);
	if (gameFrame() > v15) {
		if (nox_xxx_unitIsEnemyTo_5330C0(v2, v11)) {
			if (!nox_xxx_unitIsZombie_534A40(v2) || (v16 = *(uint32_t*)(v2 + 16), (v16 & 0x8000) == 0)) {
				v17 = nox_xxx_monsterGetSoundSet_424300(v2);
				if (v17) {
					nox_xxx_aud_501960(*(uint32_t*)(v17 + 68), v2, 0, 0);
				}
				*(uint32_t*)(v4 + 536) =
					gameFrame() + nox_common_randomInt_415FA0(2 * gameFPS(), 4 * gameFPS());
			}
		}
	}
	nox_xxx_scriptCallByEventBlock_502490((int*)(v4 + 1232), v11, v2, 14);
}

//----- (00528910) --------------------------------------------------------
int sub_528910(int a1, int a2) {
	int result;  // eax
	int v3;      // edx
	int v4;      // ecx
	uint32_t* i; // edx

	result = 0;
	v3 = *(uint32_t*)(a1 + 748);
	v4 = *(unsigned char*)(v3 + 1129);
	if (v4 > 0) {
		for (i = (uint32_t*)(v3 + 1132); *i != a2; ++i) {
			if (++result >= v4) {
				return result;
			}
		}
		result = nox_xxx_aiLostSight_528560(a1, result);
	}
	return result;
}

//----- (00528950) --------------------------------------------------------
int sub_528950(int a1, int a2) {
	int v2;      // ecx
	int v3;      // eax
	int v4;      // edx
	uint32_t* i; // ecx

	v2 = *(uint32_t*)(a1 + 748);
	v3 = 0;
	v4 = *(unsigned char*)(v2 + 1129);
	if (v4 <= 0) {
		return 0;
	}
	for (i = (uint32_t*)(v2 + 1132); *i != a2; ++i) {
		if (++v3 >= v4) {
			return 0;
		}
	}
	return 1;
}

//----- (00528990) --------------------------------------------------------
nox_object_t* nox_xxx_getFirstUpdatableObject_4DA8A0();
nox_object_t* nox_xxx_getNextUpdatableObject_4DA8B0(nox_object_t* obj);
int sub_528990(nox_object_t* a1) {
	int result; // eax
	int i;      // esi

	result = nox_xxx_getFirstUpdatableObject_4DA8A0();
	for (i = result; result; i = result) {
		if (*(uint8_t*)(i + 8) & 2) {
			if (!(*(uint8_t*)(i + 16) & 0x20)) {
				sub_528910(i, a1);
				sub_528610(i);
			}
		}
		result = nox_xxx_getNextUpdatableObject_4DA8B0(i);
	}
	return result;
}
#endif

nox_object_t* nox_xxx_getFirstUpdatableObject_4DA8A0(void);
nox_object_t* nox_xxx_getNextUpdatableObject_4DA8B0(nox_object_t* obj);

void nox_xxx_unitUpdateSightMB_5281F0(nox_object_t* monster) {
	nox_server_unit_update_sight_native(monster);
}

int nox_xxx_aiLostSight_528560(nox_object_t* monster, int index) {
	return nox_server_ai_lost_sight_native(monster, index);
}

void sub_528610(nox_object_t* monster) {
	nox_server_monster_select_enemy_native(monster);
}

void nox_xxx_monsterUpdateSeenEnemies_5286D0(nox_object_t* candidate, nox_object_t* monster) {
	nox_server_monster_update_seen_native(candidate, monster);
}

void nox_xxx_monsterVisionSeeEnemy_5287B0(nox_object_t* monster, nox_object_t* target) {
	nox_server_monster_see_enemy_native(monster, target);
}

int sub_528910(nox_object_t* monster, nox_object_t* target) {
	return nox_server_monster_remove_seen_native(monster, target);
}

int sub_528950(nox_object_t* monster, nox_object_t* target) {
	return nox_server_monster_has_seen_native(monster, target);
}

int sub_528990(nox_object_t* target) {
	nox_object_t* monster = nox_xxx_getFirstUpdatableObject_4DA8A0();
	while (monster) {
		if ((monster->obj_class & 2) && !(monster->obj_flags & 0x20)) {
			sub_528910(monster, target);
			sub_528610(monster);
		}
		monster = nox_xxx_getNextUpdatableObject_4DA8B0(monster);
	}
	return 0;
}

//----- (005289D0) --------------------------------------------------------
void nox_xxx_netReportDestroyObject_5289D0(nox_object_t* object) {
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		if ((UINT32_C(1) << player->playerInd) & object->field_37) {
			uint8_t packet[3];
			uint16_t code = (uint16_t)nox_xxx_netGetUnitCodeServ_578AC0(object);
			packet[0] = ((uint8_t)object->field_5 >> 6) | 0x31;
			packet[1] = (uint8_t)code;
			packet[2] = (uint8_t)(code >> 8);
			nox_xxx_netSendPacket0_4E5420(player->playerInd, packet, sizeof(packet), 0, 1);
		}
		if (object->obj_class & 6) {
			nox_xxx_netFriendAddRemove_4D97A0(player->playerInd, object, 0);
		}
	}
}

//----- (00528A60) --------------------------------------------------------
int nox_xxx_netObjectOutOfSight_528A60(int a1, uint32_t* a2) {
	char v4[3]; // [esp+0h] [ebp-4h]
	v4[0] = 50;
	*(uint16_t*)&v4[1] = nox_xxx_netGetUnitCodeServ_578AC0(a2);
	return nox_xxx_netSendPacket0_4E5420(a1, v4, 3, 0, 1);
}

//----- (00528A90) --------------------------------------------------------
int nox_xxx_netObjectInShadows_528A90(int a1, uint32_t* a2) {
	char v4[3]; // [esp+0h] [ebp-4h]
	v4[0] = 51;
	*(uint16_t*)&v4[1] = nox_xxx_netGetUnitCodeServ_578AC0(a2);
	return nox_xxx_netSendPacket0_4E5420(a1, v4, 3, 0, 1);
}

//----- (00528BD0) --------------------------------------------------------
int nox_xxx_monsterCmdSend_528BD0(int unit, int source, const char* command, short a4) {
	short v4;      // ax
	double v5;     // st7
	long long v6;  // rax
	double v7;     // st7
	int result;    // eax
	int i;         // esi
	char v10[520]; // [esp+Ch] [ebp-208h]

	v10[0] = -88; // MSG_TEXT_MESSAGE
	v10[3] = 8;
	v4 = nox_xxx_netGetUnitCodeServ_578AC0((uint32_t*)unit);
	v5 = *(float*)(unit + 56);
	*(uint16_t*)&v10[1] = v4;
	v6 = (long long)v5;
	v7 = *(float*)(unit + 60);
	*(uint16_t*)&v10[4] = v6;
	*(uint16_t*)&v10[6] = (long long)v7;
	*(uint16_t*)&v10[9] = a4;
	v10[8] = strlen(command) + 1;
	result = source;
	strcpy(&v10[11], command);
	if (source) {
		if (*(uint8_t*)(source + 8) & 4) { // if source is player / local player ?
			result = nox_netlist_addToMsgListCli_40EBC0(
				*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(source + 748) + 276) + 2064), 1, v10,
				(unsigned char)v10[8] + 11);
		}
	} else {
		result = nox_xxx_getFirstPlayerUnit_4DA7C0();
		for (i = result; result; i = result) {
			nox_netlist_addToMsgListCli_40EBC0(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), 1,
											   v10, (unsigned char)v10[8] + 11);
			result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
		}
	}
	return result;
}

//----- (00528D60) --------------------------------------------------------
int nox_xxx_destroyEveryChatMB_528D60() {
	int result; // eax
	int i;      // esi
	char v2[3]; // [esp+4h] [ebp-4h]

	v2[0] = -54;
	*(uint16_t*)&v2[1] = -8531;
	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		nox_netlist_addToMsgListCli_40EBC0(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), 1, v2,
										   3);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}

//----- (00528DB0) --------------------------------------------------------
#if UINTPTR_MAX == UINT32_MAX
int nox_xxx_XFerMonster_528DB0(nox_object_t* a1p) {
	int a1 = a1p;
	int v1;            // esi
	int result;        // eax
	char* v3;          // edi
	char* v4;          // eax
	char* v5;          // eax
	char* v6;          // eax
	char* v7;          // eax
	char* v8;          // eax
	char* v9;          // eax
	char* v10;         // eax
	char* v11;         // eax
	char* v12;         // eax
	char* v13;         // edi
	int v14;           // eax
	short v15;         // ax
	int v16;           // edx
	uint32_t* v17;     // edi
	uint32_t* v18;     // eax
	int v19;           // ecx
	int v20;           // ebx
	char* v21;         // ebp
	int i;             // edi
	int v23;           // eax
	int* v24;          // ebp
	char* v25;         // ebx
	char* v26;         // ebx
	int v27;           // ebp
	unsigned char v28; // bl
	uint8_t* v29;      // edi
	int v30;           // eax
	int v31;           // eax
	int v32;           // edi
	int v33;           // ebx
	int j;             // eax
	int v35;           // esi
	int v36;           // ecx
	uint32_t* v37;     // eax
	uint32_t* v38;     // eax
	int v39;           // edi
	int* v40;          // ecx
	uint32_t* v41;     // edx
	int v42;           // esi
	int v43;           // eax
	int v44;           // [esp+10h] [ebp-128h]
	int v45;           // [esp+14h] [ebp-124h]
	int v46;           // [esp+18h] [ebp-120h]
	uint8_t* v47;      // [esp+1Ch] [ebp-11Ch]
	int v48;           // [esp+20h] [ebp-118h]
	int v49;           // [esp+24h] [ebp-114h]
	int v50;           // [esp+28h] [ebp-110h]
	int v51;           // [esp+2Ch] [ebp-10Ch]
	uint32_t v52[2];   // [esp+30h] [ebp-108h]
	char v53[256];     // [esp+38h] [ebp-100h]

	v1 = *(uint32_t*)(a1 + 748);
	v51 = *(uint32_t*)(a1 + 136);
	if (!*getMemU32Ptr(0x5D4594, 2487692)) {
		*getMemU32Ptr(0x5D4594, 2487692) = nox_xxx_getNameId_4E3AA0("Glyph");
	}
	v45 = 64;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 2u);
	if ((short)v45 > 64) {
		return 0;
	}
	result = nox_xxx_mapReadWriteObjData_4F4530((int*)a1, (short)v45);
	if (!result) {
		return result;
	}
	if (!nox_crypt_IsReadOnly()) {
		nox_xxx_xferIndexedDirection_509E20(*(short*)(a1 + 124), (int2*)v52);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(v52, 8u);
	if ((short)v45 >= 3) {
		v3 = *(char**)(a1 + 756);
		if (v3) {
			v4 = v3 + 640;
		} else {
			v4 = 0;
		}
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1232), v4);
		if (v3) {
			v5 = v3 + 896;
		} else {
			v5 = 0;
		}
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1264), v5);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1220), 2u);
		if (v3) {
			v6 = v3 + 768;
		} else {
			v6 = 0;
		}
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1224), v6);
		if ((short)v45 >= 31) {
			v7 = v3 ? v3 + 1024 : 0;
			nox_xxx_xferReadScriptHandler_4F5580(
				(nox_script_callback_t*)(uintptr_t)(v1 + 1240), v7);
			v8 = v3 ? v3 + 1152 : 0;
			nox_xxx_xferReadScriptHandler_4F5580(
				(nox_script_callback_t*)(uintptr_t)(v1 + 1248), v8);
			v9 = v3 ? v3 + 1280 : 0;
			nox_xxx_xferReadScriptHandler_4F5580(
				(nox_script_callback_t*)(uintptr_t)(v1 + 1256), v9);
			v10 = v3 ? v3 + 1408 : 0;
			nox_xxx_xferReadScriptHandler_4F5580(
				(nox_script_callback_t*)(uintptr_t)(v1 + 1272), v10);
			v11 = v3 ? v3 + 1536 : 0;
			nox_xxx_xferReadScriptHandler_4F5580(
				(nox_script_callback_t*)(uintptr_t)(v1 + 1280), v11);
			v12 = v3 ? v3 + 1664 : 0;
			nox_xxx_xferReadScriptHandler_4F5580(
				(nox_script_callback_t*)(uintptr_t)(v1 + 1288), v12);
			if ((short)v45 >= 52) {
				if (v3) {
					v13 = v3 + 1792;
				} else {
					v13 = 0;
				}
				nox_xxx_xferReadScriptHandler_4F5580(
					(nox_script_callback_t*)(uintptr_t)(v1 + 1296), v13);
			}
		}
	} else {
		sub_4F5540((nox_script_callback_t*)(uintptr_t)(v1 + 1232));
		sub_4F5540((nox_script_callback_t*)(uintptr_t)(v1 + 1264));
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1220), 2u);
		sub_4F5540((nox_script_callback_t*)(uintptr_t)(v1 + 1224));
	}
	v14 = nox_crypt_IsReadOnly();
	if (nox_crypt_IsReadOnly() != 1 ||
		(v15 = nox_xxx_xferDirectionToAngle_509E00(v52), *(uint16_t*)(a1 + 126) = v15, *(uint16_t*)(a1 + 124) = v15,
		 v14 = nox_crypt_IsReadOnly(), nox_crypt_IsReadOnly() != 1)) {
		if (!v14) {
			v46 = 0;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v46, 4u);
		}
	} else if ((short)v45 >= 11) {
		v46 = 0;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v46, 4u);
	}
	if ((short)v45 >= 31) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1332), 1u);
		if ((short)v45 < 51) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v46, 2u);
			*(uint32_t*)(v1 + 1440) = (unsigned short)v46;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1440), 4u);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1352), 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1336), 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1344), 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1312), 4u);
		if ((short)v45 < 33) {
			nox_xxx_cryptSeekCur_40E0A0(2);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1304), 4u);
		*(uint32_t*)(v1 + 1308) = *(uint32_t*)(v1 + 1304);
		if ((short)v45 < 34) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1360), 4u);
		}
		LOBYTE(v48) = strlen((const char*)(v1 + 1364));
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v48, 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1364), (unsigned char)v48);
		*(uint8_t*)((unsigned char)v48 + v1 + 1364) = 0;
		if ((short)v45 >= 34) {
			if (nox_crypt_IsReadOnly()) {
				memset((void*)(v1 + 1488), 0, 0x224u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 4u);
				for (i = 0; i < v44; ++i) {
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v48, 1u);
					nox_xxx_fileReadWrite_426AC0_file3_fread(v53, (unsigned char)v48);
					v53[(unsigned char)v48] = 0;
					v23 = nox_xxx_spellNameToN_4243F0(v53);
					nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 4 * v23 + 1488), 4u);
				}
			} else {
				v16 = 0;
				v17 = (uint32_t*)(v1 + 1488);
				v44 = 0;
				v18 = (uint32_t*)(v1 + 1488);
				v19 = 137;
				do {
					if (*v18) {
						++v16;
					}
					++v18;
					--v19;
				} while (v19);
				v44 = v16;
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 4u);
				v20 = 0;
				v47 = (uint8_t*)(v1 + 1488);
				do {
					if (*v17) {
						v21 = nox_xxx_spellNameByN_424870(v20);
						LOBYTE(v46) = strlen(v21);
						nox_xxx_fileReadWrite_426AC0_file3_fread(&v46, 1u);
						nox_xxx_fileReadWrite_426AC0_file3_fread(v21, (unsigned char)v46);
						nox_xxx_fileReadWrite_426AC0_file3_fread(v47, 4u);
					}
					++v20;
					v17 = v47 + 4;
					v47 += 4;
				} while (v20 < 137);
			}
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1488), 0x224u);
		}
		if ((short)v45 < 46) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1448) = (unsigned char)v44;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1450) = (unsigned char)v44;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1448), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1450), 2u);
		}
		if ((short)v45 <= 32) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 4u);
		}
		if ((short)v45 < 46) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1456) = (unsigned char)v44;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1458) = (unsigned char)v44;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1456), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1458), 2u);
		}
		if ((short)v45 <= 32) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 4u);
		}
		if ((short)v45 < 46) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1464) = (unsigned char)v44;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1466) = (unsigned char)v44;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1464), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1466), 2u);
		}
		if ((short)v45 <= 32) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 4u);
		}
		if ((short)v45 < 46) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1472) = (unsigned char)v44;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1474) = (unsigned char)v44;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1472), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1474), 2u);
		}
		if ((short)v45 <= 32) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 4u);
		}
		if ((short)v45 < 46) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1480) = (unsigned char)v44;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			*(uint16_t*)(v1 + 1482) = (unsigned char)v44;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1480), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1482), 2u);
		}
		if ((short)v45 > 32 || (nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 4u), (short)v45 >= 32)) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1316), 4u);
		}
		if ((short)v45 >= 33) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2040), 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1320), 4u);
			if ((short)v45 < 42) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 2u);
				if (!(uint16_t)v44) {
					*(uint8_t*)(v1 + 1445) = 1;
				}
			}
			if ((short)v45 < 53) {
				nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2044), 4u);
				nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2048), 4u);
				nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2052), 4u);
			} else {
				v24 = (int*)(v1 + 2044);
				v46 = 3;
				do {
					if (nox_crypt_IsReadOnly() == 1) {
						nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
						nox_xxx_fileReadWrite_426AC0_file3_fread(v53, (unsigned char)v44);
						v53[(unsigned char)v44] = 0;
						*v24 = nox_xxx_spellNameToN_4243F0(v53);
					} else {
						v25 = nox_xxx_spellNameByN_424870(*v24);
						LOBYTE(v44) = strlen(v25);
						nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
						nox_xxx_fileReadWrite_426AC0_file3_fread(v25, (unsigned char)v44);
					}
					++v24;
					--v46;
				} while (v46);
			}
		}
		if ((short)v45 >= 34) {
			if (nox_crypt_IsReadOnly()) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v53, (unsigned char)v44);
				v53[(unsigned char)v44] = 0;
				*(uint32_t*)(v1 + 1360) = nox_xxx_actionNByNameMB_5345F0(v53);
			} else {
				v26 = sub_5345B0(*(uint32_t*)(v1 + 1360));
				LOBYTE(v44) = strlen(v26);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v26, (unsigned char)v44);
			}
		}
	}
	if ((short)v45 >= 41) {
		result = nox_xxx_XFer_ActionData_529CE0(a1);
		if (!result) {
			return result;
		}
	}
	if ((short)v45 >= 42) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1445), 1u);
	}
	if ((short)v45 >= 43 && *(uint8_t*)(a1 + 12) & 8) {
		LOBYTE(v44) = 0;
		v27 = *(uint32_t*)(a1 + 692);
		if ((short)v45 >= 50) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v27 + 1716), 4u);
		}
		if ((short)v45 >= 61) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v27 + 1720), 4u);
		}
		if ((short)v45 >= 48) {
			LOBYTE(v47) = strlen((const char*)(v27 + 1684));
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v47, 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v27 + 1684), (unsigned char)v47);
			*(uint8_t*)((unsigned char)v47 + v27 + 1684) = 0;
		}
		if (nox_crypt_IsReadOnly() == 1) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
			if (!v27) {
				goto LABEL_137;
			}
			*(uint8_t*)v27 = v44;
		} else {
			if (v27) {
				LOBYTE(v44) = *(uint8_t*)v27;
			}
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 1u);
		}
		if (v27) {
			LOBYTE(v44) = 0;
			if (*(uint8_t*)v27) {
				do {
					if (nox_crypt_IsReadOnly() == 1) {
						nox_xxx_XFer_ReadShopItem_52A840(v27 + 28 * (unsigned char)v44 + 4, (short)v45);
					} else {
						nox_xxx_XFer_WriteShopItem_52A5F0(v27 + 28 * (unsigned char)v44 + 4);
					}
					LOBYTE(v44) = v44 + 1;
				} while ((unsigned char)v44 < *(uint8_t*)v27);
			}
		}
	}
LABEL_137:
	if ((short)v45 >= 44) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)v1, 4u);
	}
	if ((short)v45 >= 45) {
		v50 = *(uint32_t*)(a1 + 12) & 0x180;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v50, 4u);
		*(uint32_t*)(a1 + 12) |= v50;
	}
	if ((short)v45 >= 49) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(*(uint8_t**)(a1 + 556), 2u);
	}
	if ((short)v45 >= 51) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1348), 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1340), 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1444), 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2036), 1u);
	}
	if (*(uint8_t*)(a1 + 12) & 0x20 && (short)v45 >= 54) {
		v28 = 0;
		v29 = (uint8_t*)(v1 + 2076);
		LOBYTE(v46) = 0;
		do {
			nox_xxx_fileReadWrite_426AC0_file3_fread(v29, 3u);
			if (nox_crypt_IsReadOnly() == 1) {
				nox_xxx_setNPCColor_4E4A90(a1p, v46, (const nox_color3_t*)v29);
			}
			++v28;
			v29 += 3;
			LOBYTE(v46) = v28;
		} while (v28 < 6u);
	}
	if ((short)v45 >= 55 && *(uint8_t*)(a1 + 12) & 0x20) {
		nox_xxx_readNPCVoiceSet_52AD10(a1);
	}
	if (!((short)v45 < 62 || (result = nox_xxx_XFer_ReadMonsterBuffs_52AAB0((uint32_t*)a1)) != 0)) {
		return result;
	}
	if ((short)v45 >= 63 && *(uint32_t*)(a1 + 12) & 0x80000) {
		nox_xxx_readNPCVoiceSet_52AD10(a1);
	}
	if ((short)v45 >= 64) {
		LOBYTE(v47) = *(uint8_t*)(a1 + 540);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v47, 1u);
		if (nox_crypt_IsReadOnly() != 1) {
			goto LABEL_171;
		}
		if (!(uint8_t)v47) {
			goto LABEL_164;
		}
		nox_xxx_setSomePoisonData_4EEA90(a1, (unsigned char)v47);
	}
	if (nox_crypt_IsReadOnly() != 1) {
		goto LABEL_171;
	}
LABEL_164:
	if (!*(uint8_t*)(v1 + 1445)) {
		v30 = *(uint32_t*)(a1 + 16);
		if ((v30 & 0x8000) != 0 && !nox_xxx_unitIsZombie_534A40(a1)) {
			*(uint32_t*)(a1 + 16) |= 0x40u;
		}
		goto LABEL_171;
	}
	if (nox_common_gameFlags_check_40A5C0(1)) {
		*(uint16_t*)(*(uint32_t*)(a1 + 556) + 4) = 0;
		**(uint16_t**)(a1 + 556) = 0;
	}
	if (nox_crypt_IsReadOnly() == 1) {
		v30 = *(uint32_t*)(a1 + 16);
		if ((v30 & 0x8000) != 0 && !nox_xxx_unitIsZombie_534A40(a1)) {
			*(uint32_t*)(a1 + 16) |= 0x40u;
		}
	}
LABEL_171:
	if (*(uint32_t*)(a1 + 136)) {
		if (nox_crypt_IsReadOnly() != 1) {
			*(uint32_t*)(a1 + 136) = v51;
			return 1;
		}
		result = nox_xxx_xfer_4F3E30(v45, a1, *(uint32_t*)(a1 + 136));
		if (!result) {
			return result;
		}
	}
	if (nox_crypt_IsReadOnly() == 1) {
		if (nox_common_gameFlags_check_40A5C0(0x200000) || !nox_xxx_gameIsSwitchToSolo_4DB240()) {
			nox_xxx_monsterOnSpawnSpellcaster_529BC0(a1);
		}
		if (nox_crypt_IsReadOnly() == 1) {
			if (*(uint8_t*)(a1 + 8) & 2) {
				v31 = *(uint32_t*)(a1 + 12);
				if (v31 & 0x2000) {
					v32 = 0;
					v33 = 0;
					for (j = nox_xxx_inventoryGetFirst_4E7980(a1); j;
						 j = nox_xxx_inventoryGetNext_4E7990(j)) {
						if (*(unsigned short*)(j + 4) == *getMemU32Ptr(0x5D4594, 2487692)) {
							v32 = 1;
						}
					}
					v35 = v1 + 2044;
					v36 = 3;
					v37 = (uint32_t*)v35;
					do {
						if (*v37) {
							++v33;
						}
						++v37;
						--v36;
					} while (v36);
					if (!v32 && v33) {
						v38 = nox_xxx_newObjectByTypeID_4E3810("Glyph");
						v46 = (int)v38;
						if (v38) {
							v39 = v38[173];
							if (v33 > 0) {
								v40 = (int*)v35;
								v41 = (uint32_t*)v38[173];
								v42 = v33;
								do {
									v43 = *v40;
									++v40;
									*v41 = v43;
									++v41;
									--v42;
								} while (v42);
								v38 = (uint32_t*)v46;
							}
							*(uint8_t*)(v39 + 20) = v33;
							*(uint32_t*)(v39 + 24) = 0;
							*(uint32_t*)(v39 + 28) = *(uint32_t*)(a1 + 56);
							*(uint32_t*)(v39 + 32) = *(uint32_t*)(a1 + 60);
						}
						nox_xxx_inventoryPutImpl_4F3070(a1, (int)v38, 1);
					}
				}
			}
		}
	}
	*(uint32_t*)(a1 + 136) = v51;
	return 1;
}
// 528DB0: using guessed type char var_100[256];
#else
extern int32_t nox_xxx_XFerMonster_native_528DB0(nox_object_t* obj);
int nox_xxx_XFerMonster_528DB0(nox_object_t* obj) { return nox_xxx_XFerMonster_native_528DB0(obj); }
#endif

//----- (00529BC0) --------------------------------------------------------
void nox_xxx_monsterOnSpawnSpellcaster_529BC0(int a1) {
	int v1;       // esi
	uint32_t* v2; // eax
	int i;        // ecx
	int v4;       // edx
	int v5;       // edi
	int v6;       // edi
	int v7;       // ecx

	if (a1) {
		v1 = *(uint32_t*)(a1 + 748);
		v2 = nox_xxx_monsterDefByTT_517560(*(unsigned short*)(a1 + 4));
		if (v2) {
			if (!*(uint8_t*)(v1 + 1445)) {
				**(uint16_t**)(a1 + 556) = *((uint16_t*)v2 + 34);
				*(uint16_t*)(*(uint32_t*)(a1 + 556) + 4) = *((uint16_t*)v2 + 34);
				*(uint16_t*)(*(uint32_t*)(a1 + 556) + 2) = *((uint16_t*)v2 + 34);
			}
			if (*(uint8_t*)(v1 + 1444) == 1) {
				*(uint32_t*)(v1 + 1440) = v2[23];
			} else {
				for (i = 0; i < 22; ++i) {
					v4 = 1 << i;
					if (!((1 << i) & 0x19C40)) {
						v5 = *(uint32_t*)(v1 + 1440);
						if (v5 & v4 && !(v4 & v2[23])) {
							*(uint32_t*)(v1 + 1440) = v5 & ~v4;
						}
						v6 = *(uint32_t*)(v1 + 1440);
						if (!(v6 & v4) && v4 & v2[23]) {
							*(uint32_t*)(v1 + 1440) = v4 | v6;
						}
					}
				}
			}
			v7 = *(uint32_t*)(v1 + 1440);
			if (!(v7 & 0x20)) {
				BYTE1(v7) &= 0xE7u;
				*(uint32_t*)(v1 + 1440) = v7;
			}
			if (*(uint8_t*)(v1 + 1340) == 1) {
				*(uint32_t*)(v1 + 1336) = v2[20];
			}
			if (*(uint8_t*)(v1 + 1348) == 1) {
				*(uint32_t*)(v1 + 1344) = v2[21];
			}
			if (*(uint8_t*)(v1 + 2036) == 1) {
				nox_xxx_monsterAutoSpells_54C0C0(a1);
			}
		}
	}
}

//----- (00529CE0) --------------------------------------------------------
nox_waypoint_t* sub_579C80(unsigned int a1);
int nox_xxx_XFer_ActionData_529CE0(int a1) {
	int v1;             // ebp
	int v3;             // ebx
	bool v4;            // cc
	uint8_t** v5;       // esi
	int v6;             // esi
	bool v7;            // zf
	uint8_t* v8;        // esi
	int v9;             // ecx
	int v10;            // eax
	int v11;            // eax
	int v12;            // eax
	int v13;            // eax
	int v14;            // ebx
	unsigned char* v15; // edi
	int* v16;           // ebp
	int v17;            // eax
	int v18;            // [esp+4h] [ebp-118h]
	char v19;           // [esp+Bh] [ebp-111h]
	int v20;            // [esp+Ch] [ebp-110h]
	int v21;            // [esp+10h] [ebp-10Ch]
	int v22;            // [esp+14h] [ebp-108h]
	int v23;            // [esp+18h] [ebp-104h]
	char v24[256];      // [esp+1Ch] [ebp-100h]

	v1 = *(uint32_t*)(a1 + 748);
	v21 = 4;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v21, 2u);
	if ((short)v21 > 4) {
		return 0;
	}
	v19 = 1;
	if ((short)v21 < 2) {
		goto LABEL_67;
	}
	v19 = 0;
	if (nox_common_gameFlags_check_40A5C0(1) && !nox_common_gameFlags_check_40A5C0(0x400000)) {
		v19 = 1;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v19, 1u);
	if (!(v19 || (uint16_t)v21 == 1)) {
		return 1;
	}
LABEL_67:
	v23 = gameFrame();
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v23, 4u);
	v3 = gameFrame() - v23;
	v18 = 0;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 8), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 12), 8 * *(uint32_t*)(v1 + 8));
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 268), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 272), 8u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 280), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 284), 1u);
	if (nox_crypt_IsReadOnly() == 1) {
		nox_xxx_AssignIfGreater_52A420((int*)(v1 + 280), v3);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 296), 4u);
	v4 = *(int*)(v1 + 296) <= 0;
	v18 = 0;
	if (!v4) {
		v5 = (uint8_t**)(v1 + 300);
		do {
			if (nox_crypt_IsReadOnly()) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v22, 4u);
				*v5 = sub_579C80(v22);
			} else {
				nox_xxx_fileReadWrite_426AC0_file3_fread(*v5, 4u);
			}
			++v5;
			v4 = ++v18 < *(int*)(v1 + 296);
		} while (v4);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 364), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 368), 8u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 376), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 380), 8u);
	strcpy(v24, nox_xxx_getSndName_40AF80(*(uint32_t*)(v1 + 388)));
	LOBYTE(v20) = strlen(v24);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v20, 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(v24, (unsigned char)v20);
	v24[(unsigned char)v20] = 0;
	*(uint32_t*)(v1 + 388) = nox_xxx_utilFindSound_40AF50(v24);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 396), 8u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 404), 4u);
	if (nox_crypt_IsReadOnly() == 1) {
		nox_xxx_AssignIfGreater_52A420((int*)(v1 + 404), v3);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 481), 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 482), 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 483), 1u);
	if ((short)v21 < 3) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 496), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 500), 8u);
	if (nox_crypt_IsReadOnly() == 1) {
		nox_xxx_AssignIfGreater_52A420((int*)(v1 + 496), v3);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 536), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 540), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 536), v3);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 540), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 544), 1u);
	v6 = 0;
	if ((signed char)*(uint8_t*)(v1 + 544) >= 0) {
		v18 = v1 + 552;
		do {
			sub_52A440(a1, v18, v3);
			++v6;
			v18 += 24;
		} while (v6 <= *(char*)(v1 + 544));
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1129), 1u);
	v7 = *(uint8_t*)(v1 + 1129) == 0;
	v18 = 0;
	if (!v7) {
		v8 = (uint8_t*)(v1 + 1132);
		do {
			if (nox_crypt_IsReadOnly()) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(v8, 4u);
			} else {
				if (!*getMemU32Ptr(0x5D4594, 2487688)) {
					*getMemU32Ptr(0x5D4594, 2487688) = nox_xxx_getNameId_4E3AA0("NewPlayer");
				}
				nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(*(uint32_t*)v8 + 44), 4u);
			}
			v9 = *(unsigned char*)(v1 + 1129);
			v8 += 4;
			++v18;
		} while (v18 < v9);
	}
	if (nox_crypt_IsReadOnly()) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1196), 4u);
	} else {
		v10 = *(uint32_t*)(v1 + 1196);
		if (v10) {
			v22 = *(uint32_t*)(v10 + 44);
		} else {
			v22 = 0;
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v22, 4u);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1204), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 1204), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2096), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2100), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2104), 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2105), 1u);
	LOBYTE(v20) = strlen((const char*)(v1 + 2106));
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v20, 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2106), (unsigned char)v20);
	*(uint8_t*)((unsigned char)v20 + v1 + 2106) = 0;
	if ((short)v21 < 4) {
		return 1;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 4), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 288), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 292), 4u);
	v11 = nox_server_getObjectFromNetCode_4ECCB0(*(uint32_t*)(v1 + 392));
	if (v11) {
		v18 = *(uint32_t*)(v11 + 44);
	} else {
		v18 = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
	if (nox_crypt_IsReadOnly() == 1) {
		*(uint32_t*)(v1 + 392) = v18;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 492), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 508), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 508), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 512), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 512), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 516), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 516), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 520), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 520), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 528), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 528), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 532), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 532), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 524), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 548), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 548), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1128), 1u);
	v12 = nox_server_getObjectFromNetCode_4ECCB0(*(uint32_t*)(v1 + 1200));
	if (v12) {
		v18 = *(uint32_t*)(v12 + 44);
	} else {
		v18 = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
	if (nox_crypt_IsReadOnly() == 1) {
		*(uint32_t*)(v1 + 1200) = v18;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1208), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 1208), v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1212), 4u);
	nox_xxx_AssignIfGreater_52A420((int*)(v1 + 1212), v3);
	v13 = *(uint32_t*)(v1 + 1216);
	v14 = 0;
	if (v13) {
		v18 = *(uint32_t*)(v13 + 44);
	} else {
		v18 = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
	if (nox_crypt_IsReadOnly() == 1) {
		*(uint32_t*)(v1 + 1216) = v18;
	}
	v15 = (unsigned char*)(v1 + 2172);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2172), 1u);
	if (*(uint8_t*)(v1 + 2172)) {
		v16 = (int*)(v1 + 2140);
		do {
			v17 = nox_server_getObjectFromNetCode_4ECCB0(*v16);
			if (v17) {
				v18 = *(uint32_t*)(v17 + 44);
			} else {
				v18 = 0;
			}
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
			if (nox_crypt_IsReadOnly() == 1) {
				*v16 = v18;
			}
			++v14;
			++v16;
		} while (v14 < *v15);
	}
	return 1;
}
// 529CE0: using guessed type char var_100[256];

//----- (0052A420) --------------------------------------------------------
int nox_xxx_AssignIfGreater_52A420(int* a1, int a2) {
	int result; // eax

	result = a2 + *a1;
	if (result >= 1) {
		*a1 = result;
	} else {
		*a1 = 1;
	}
	return result;
}

//----- (0052A440) --------------------------------------------------------
int sub_52A440(int a1, int a2, int a3) {
	int v3;        // eax
	int v4;        // edi
	int* v5;       // esi
	int result;    // eax
	size_t v7;     // [esp-4h] [ebp-120h]
	int v8;        // [esp+10h] [ebp-10Ch]
	int v9;        // [esp+14h] [ebp-108h]
	int v10;       // [esp+18h] [ebp-104h]
	char v11[256]; // [esp+1Ch] [ebp-100h]

	strcpy(v11, sub_534650(*(uint32_t*)a2));
	LOBYTE(v10) = strlen(v11);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v10, 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(v11, (unsigned char)v10);
	v11[(unsigned char)v10] = 0;
	v3 = nox_xxx_actionByName_534670(v11);
	*(uint32_t*)a2 = v3;
	LOBYTE(v8) = getMemByte(0x587000, 255604 + 16 * v3);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v8, 1u);
	v4 = 0;
	if ((unsigned char)v8 > 0u) {
		v5 = (int*)(a2 + 4);
		while (1) {
			result = *getMemU32Ptr(0x587000, 255608 + 4 * (v4 + 4 * *(uint32_t*)a2));
			switch (result) {
			case 0:
				v7 = 8;
				nox_xxx_fileReadWrite_426AC0_file3_fread(v5, v7);
				break;
			case 1:
				if (nox_crypt_IsReadOnly()) {
					v7 = 4;
					nox_xxx_fileReadWrite_426AC0_file3_fread(v5, v7);
					break;
				}
				if (*v5) {
					nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(*v5 + 44), 4u);
				} else {
					v9 = 0;
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v9, 4u);
				}
				break;
			case 2:
				if (nox_crypt_IsReadOnly()) {
					v7 = 4;
					nox_xxx_fileReadWrite_426AC0_file3_fread(v5, v7);
					break;
				}
				if (*v5) {
					nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)*v5, 4u);
				} else {
					v9 = 0;
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v9, 4u);
				}
				break;
			case 3:
			case 4:
			case 6:
				v7 = 4;
				nox_xxx_fileReadWrite_426AC0_file3_fread(v5, v7);
				break;
			case 5:
				nox_xxx_fileReadWrite_426AC0_file3_fread(v5, 4u);
				if (nox_crypt_IsReadOnly() == 1) {
					nox_xxx_AssignIfGreater_52A420(v5, a3);
				}
				break;
			case 7:
				v7 = 1;
				nox_xxx_fileReadWrite_426AC0_file3_fread(v5, v7);
				break;
			default:
				return result;
			}
			++v4;
			v5 += 2;
			if (v4 >= (unsigned char)v8) {
				return nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(a2 + 20), 4u);
			}
		}
	}
	return nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(a2 + 20), 4u);
}
// 52A440: using guessed type char var_100[256];

// 64-bit monster and NPC transfers use the native Go implementations below.
// Keep this ABI32 helper only for the original 32-bit transfer bodies: its
// local argument record stores object pointers in four-byte slots.
#if UINTPTR_MAX == UINT32_MAX
//----- (0052AAB0) --------------------------------------------------------
int nox_xxx_XFer_ReadMonsterBuffs_52AAB0(uint32_t* a1) {
	int v1;   // ebp
	char* v2; // ebx
	void* v3; // eax
	int v5;   // ebx
	int v6;   // edi
	int v7;   // ecx
	int v8;   // eax
	void* v9; // eax
	int v10;  // [esp-4h] [ebp-138h]
	int v11;  // [esp+10h] [ebp-124h]
	int v12;  // [esp+14h] [ebp-120h]
	int v13;  // [esp+18h] [ebp-11Ch]
	int v14;  // [esp+1Ch] [ebp-118h]
	int v15;  // [esp+20h] [ebp-114h]
	int v16;  // [esp+24h] [ebp-110h]
	int v17[3];
	char v20[256]; // [esp+34h] [ebp-100h]

	v13 = 2;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v13, 2u);
	if ((short)v13 > 2 || (short)v13 <= 0) {
		return 0;
	}
	LOBYTE(v16) = sub_424CB0((int)a1);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v16, 1u);
	if (nox_crypt_IsReadOnly()) {
		v5 = 0;
		if (!(uint8_t)v16) {
			return 1;
		}
		while (1) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v11, 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(v20, (unsigned char)v11);
			v20[(unsigned char)v11] = 0;
			v6 = nox_xxx_enchantByName_424880(v20);
			if (v6 == -1) {
				break;
			}
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v15, 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 4u);
			v7 = a1[15];
			v17[1] = a1[14];
			v10 = (unsigned char)v15;
			v17[0] = a1;
			v17[2] = v7;
			v8 = nox_xxx_getEnchantSpell_424920(v6);
			nox_xxx_spellAccept_4FD400(v8, (int)a1, a1, (int)a1, v17, v10);
			*((uint16_t*)a1 + v6 + 172) = v14;
			if (v6 == 26 && (short)v13 >= 2) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v12, 4u);
				v9 = sub_4FF2D0(51, (nox_object_t*)a1);
				if (v9) {
					*(uint32_t*)((uint8_t*)v9 + 72) = v12;
				}
			}
			if (++v5 >= (unsigned char)v16) {
				return 1;
			}
		}
		return 0;
	}
	v1 = sub_424D00();
	if (v1 == -1) {
		return 1;
	}
	do {
		if (nox_xxx_testUnitBuffs_4FF350((nox_object_t*)a1, v1)) {
			v2 = nox_xxx_getEnchantName_4248F0(v1);
			LOBYTE(v11) = strlen(v2);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v11, 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(v2, (unsigned char)v11);
			LOBYTE(v15) = nox_xxx_buffGetPower_4FF570((nox_object_t*)a1, v1);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v15, 1u);
			v14 = nox_xxx_unitGetBuffTimer_4FF550((nox_object_t*)a1, v1);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 4u);
			if (v1 == 26) {
				v3 = sub_4FF2D0(51, (nox_object_t*)a1);
				if (v3) {
					v12 = *(uint32_t*)((uint8_t*)v3 + 72);
				} else {
					v12 = 100;
				}
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v12, 4u);
			}
		}
		v1 = sub_424D20(v1);
	} while (v1 != -1);
	return 1;
}
// 52AAB0: using guessed type char var_100[256];
#endif

//----- (0052AD10) --------------------------------------------------------
size_t nox_xxx_readNPCVoiceSet_52AD10(int a1) {
	const char** v1; // eax
	const char** v2; // esi
	size_t result;   // eax
	const char** v4; // eax
	int v5;          // [esp+4h] [ebp-104h]
	char v6[256];    // [esp+8h] [ebp-100h]

	if (nox_crypt_IsReadOnly()) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v5, 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(v6, (unsigned char)v5);
		v6[(unsigned char)v5] = 0;
		v4 = nox_xxx_getDefaultSoundSet_424350(v6);
		result = nox_xxx_setNPCVoiceSet_424320(a1, (int)v4);
	} else {
		v1 = (const char**)nox_xxx_monsterGetSoundSet_424300(a1);
		v2 = v1;
		if (v1) {
			LOBYTE(v5) = strlen(*v1);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v5, 1u);
			result = nox_xxx_fileReadWrite_426AC0_file3_fread(*v2, (unsigned char)v5);
		} else {
			LOBYTE(v5) = 0;
			result = nox_xxx_fileReadWrite_426AC0_file3_fread(&v5, 1u);
		}
	}
	return result;
}
// 52AD10: using guessed type char var_100[256];

//----- (0052ADE0) --------------------------------------------------------
#if UINTPTR_MAX == UINT32_MAX
int nox_xxx_XFerNPC_52ADE0(nox_object_t* a1p) {
	int a1 = a1p;
	int v1;            // esi
	char* v2;          // edi
	int result;        // eax
	uint32_t* v4;      // eax
	char* v5;          // eax
	char* v6;          // eax
	char* v7;          // eax
	char* v8;          // eax
	char* v9;          // eax
	char* v10;         // eax
	char* v11;         // eax
	char* v12;         // eax
	char* v13;         // eax
	char* v14;         // edi
	int v15;           // eax
	short v16;         // ax
	unsigned char v17; // bl
	uint8_t* v18;      // edi
	int v19;           // ebx
	unsigned char i;   // bl
	int v21;           // edx
	uint16_t* v22;     // eax
	uint16_t* v23;     // eax
	int v24;           // edx
	uint8_t* v25;      // ebx
	uint32_t* v26;     // eax
	int v27;           // ecx
	char* v28;         // ebp
	bool v29;          // zf
	unsigned char j;   // bl
	int v31;           // eax
	int* v32;          // ebp
	char* v33;         // ebx
	char* v34;         // ebx
	int v35;           // edi
	int v36;           // eax
	int v37;           // eax
	int v38;           // eax
	int v39;           // eax
	uint32_t* k;       // esi
	int v41;           // eax
	int v42;           // eax
	int v43;           // [esp+10h] [ebp-238h]
	int v44;           // [esp+14h] [ebp-234h]
	int v45;           // [esp+18h] [ebp-230h]
	unsigned char v46; // [esp+1Fh] [ebp-229h]
	int v47;           // [esp+20h] [ebp-228h]
	int v48;           // [esp+24h] [ebp-224h]
	int v49;           // [esp+28h] [ebp-220h]
	int v50;           // [esp+2Ch] [ebp-21Ch]
	int v51;           // [esp+30h] [ebp-218h]
	int v52;           // [esp+34h] [ebp-214h]
	char v53[3];       // [esp+38h] [ebp-210h]
	int v54;           // [esp+3Ch] [ebp-20Ch]
	uint32_t v55[2];   // [esp+40h] [ebp-208h]
	char v56[256];     // [esp+48h] [ebp-200h]
	char v57[256];     // [esp+148h] [ebp-100h]

	v1 = a1p->data_update;
	v2 = *(char**)(a1 + 756);
	v54 = *(uint32_t*)(a1 + 136);
	v44 = 62;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 2u);
	if ((short)v44 > 62) {
		return 0;
	}
	if (nox_crypt_IsReadOnly() == 1) {
		v4 = nox_xxx_monsterDefByTT_517560(*(unsigned short*)(a1 + 4));
		*(uint32_t*)(v1 + 484) = v4;
		if (v4) {
			*(uint32_t*)(v1 + 1440) = v4[23];
		}
	}
	result = nox_xxx_mapReadWriteObjData_4F4530((int*)a1, (short)v44);
	if (!result) {
		return 0;
	}
	if (!nox_crypt_IsReadOnly()) {
		nox_xxx_xferIndexedDirection_509E20(*(short*)(a1 + 124), (int2*)v55);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(v55, 8u);
	if (v2) {
		v5 = v2 + 640;
	} else {
		v5 = 0;
	}
	nox_xxx_xferReadScriptHandler_4F5580(
		(nox_script_callback_t*)(uintptr_t)(v1 + 1232), v5);
	if (v2) {
		v6 = v2 + 896;
	} else {
		v6 = 0;
	}
	nox_xxx_xferReadScriptHandler_4F5580(
		(nox_script_callback_t*)(uintptr_t)(v1 + 1264), v6);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1220), 2u);
	if (v2) {
		v7 = v2 + 768;
	} else {
		v7 = 0;
	}
	nox_xxx_xferReadScriptHandler_4F5580(
		(nox_script_callback_t*)(uintptr_t)(v1 + 1224), v7);
	if ((short)v44 >= 32) {
		v8 = v2 ? v2 + 1024 : 0;
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1240), v8);
		v9 = v2 ? v2 + 1152 : 0;
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1248), v9);
		v10 = v2 ? v2 + 1280 : 0;
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1256), v10);
		v11 = v2 ? v2 + 1408 : 0;
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1272), v11);
		v12 = v2 ? v2 + 1536 : 0;
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1280), v12);
		v13 = v2 ? v2 + 1664 : 0;
		nox_xxx_xferReadScriptHandler_4F5580(
			(nox_script_callback_t*)(uintptr_t)(v1 + 1288), v13);
		if ((short)v44 >= 50) {
			if (v2) {
				v14 = v2 + 1792;
			} else {
				v14 = 0;
			}
			nox_xxx_xferReadScriptHandler_4F5580(
				(nox_script_callback_t*)(uintptr_t)(v1 + 1296), v14);
		}
	}
	v47 = 0; // FIXME: set to direction? was uninitialized
	v15 = nox_crypt_IsReadOnly();
	if (nox_crypt_IsReadOnly() != 1 ||
		(v16 = nox_xxx_xferDirectionToAngle_509E00(v55), *(uint16_t*)(a1 + 126) = v16, *(uint16_t*)(a1 + 124) = v16,
		 v15 = nox_crypt_IsReadOnly(), nox_crypt_IsReadOnly() != 1)) {
		if (!v15) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v47, 4u);
		}
	} else {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v47, 4u);
	}
	if (nox_crypt_IsReadOnly() == 1) {
		v17 = 0;
		LOBYTE(v48) = 0;
		do {
			nox_xxx_fileReadWrite_426AC0_file3_fread(v53, 3u);
			nox_xxx_setNPCColor_4E4A90(a1p, v48, (const nox_color3_t*)v53);
			LOBYTE(v48) = ++v17;
		} while (v17 < 6u);
	} else {
		v18 = (uint8_t*)(v1 + 2076);
		v19 = 6;
		do {
			nox_xxx_fileReadWrite_426AC0_file3_fread(v18, 3u);
			v18 += 3;
			--v19;
		} while (v19);
	}
	if (nox_crypt_IsReadOnly() == 1 && (uint16_t)v44 == 31) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 2u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v46, 1u);
		for (i = 0; i < v46; ++i) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v50, 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(v56, (unsigned char)v50);
			v56[(unsigned char)v50] = 0;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
		}
	}
	if ((short)v44 >= 32) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1332), 1u);
		if ((short)v44 < 49) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v47, 2u);
			*(uint32_t*)(v1 + 1440) = (unsigned short)v47;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1440), 4u);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1352), 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1336), 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1344), 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1312), 4u);
		v43 = 0;
		v22 = *(uint16_t**)(a1 + 556);
		if (v22) {
			LOWORD(v21) = *v22;
			v43 = v21;
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 2u);
		v23 = *(uint16_t**)(a1 + 556);
		if (v23) {
			*v23 = v43;
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1304), 4u);
		*(uint32_t*)(v1 + 1308) = *(uint32_t*)(v1 + 1304);
		if ((short)v44 < 35) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1360), 4u);
		}
		LOBYTE(v49) = strlen((const char*)(v1 + 1364));
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1364), (unsigned char)v49);
		*(uint8_t*)((unsigned char)v49 + v1 + 1364) = 0;
		if ((short)v44 >= 34) {
			if (nox_crypt_IsReadOnly()) {
				memset((void*)(v1 + 1488), 0, 0x224u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 4u);
				for (j = 0; j < v43; LOBYTE(v48) = j) {
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v49, 1u);
					nox_xxx_fileReadWrite_426AC0_file3_fread(v56, (unsigned char)v49);
					v56[(unsigned char)v49] = 0;
					v31 = nox_xxx_spellNameToN_4243F0(v56);
					nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 4 * v31 + 1488), 4u);
					++j;
				}
			} else {
				v24 = 0;
				v25 = (uint8_t*)(v1 + 1488);
				v43 = 0;
				v26 = (uint32_t*)(v1 + 1488);
				v27 = 137;
				do {
					if (*v26) {
						++v24;
					}
					++v26;
					--v27;
				} while (v27);
				v43 = v24;
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 4u);
				v48 = 0;
				v47 = 137;
				do {
					if (*(uint32_t*)v25) {
						v28 = nox_xxx_spellNameByN_424870(v48);
						LOBYTE(v52) = strlen(v28);
						nox_xxx_fileReadWrite_426AC0_file3_fread(&v52, 1u);
						nox_xxx_fileReadWrite_426AC0_file3_fread(v28, (unsigned char)v52);
						nox_xxx_fileReadWrite_426AC0_file3_fread(v25, 4u);
					}
					v25 += 4;
					v29 = v47 == 1;
					++v48;
					--v47;
				} while (!v29);
			}
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1488), 0x224u);
		}
		if ((short)v44 < 47) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1448) = (unsigned char)v43;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1450) = (unsigned char)v43;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1448), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1450), 2u);
		}
		if ((short)v44 < 34) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 4u);
		}
		if ((short)v44 < 47) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1456) = (unsigned char)v43;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1458) = (unsigned char)v43;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1456), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1458), 2u);
		}
		if ((short)v44 < 34) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 4u);
		}
		if ((short)v44 < 47) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1464) = (unsigned char)v43;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1466) = (unsigned char)v43;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1464), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1466), 2u);
		}
		if ((short)v44 < 34) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 4u);
		}
		if ((short)v44 < 47) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1472) = (unsigned char)v43;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1474) = (unsigned char)v43;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1472), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1474), 2u);
		}
		if ((short)v44 < 34) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 4u);
		}
		if ((short)v44 < 47) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1480) = (unsigned char)v43;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
			*(uint16_t*)(v1 + 1482) = (unsigned char)v43;
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1480), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1482), 2u);
		}
		if ((short)v44 < 34) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 4u);
		}
		if ((short)v44 >= 33) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1316), 4u);
		}
		if ((short)v44 >= 34) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 2040), 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1324), 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1328), 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1320), 4u);
			if ((short)v44 < 42) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 2u);
				if (!(uint16_t)v43) {
					*(uint8_t*)(v1 + 1445) = 1;
				}
			}
			v32 = (int*)(v1 + 2044);
			v47 = 3;
			do {
				if (nox_crypt_IsReadOnly() == 1) {
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
					nox_xxx_fileReadWrite_426AC0_file3_fread(v56, (unsigned char)v43);
					v56[(unsigned char)v43] = 0;
					*v32 = nox_xxx_spellNameToN_4243F0(v56);
				} else {
					v33 = nox_xxx_spellNameByN_424870(*v32);
					LOBYTE(v43) = strlen(v33);
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
					nox_xxx_fileReadWrite_426AC0_file3_fread(v33, (unsigned char)v43);
				}
				++v32;
				--v47;
			} while (v47);
		}
		if ((short)v44 >= 35) {
			if (nox_crypt_IsReadOnly()) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v57, (unsigned char)v43);
				v57[(unsigned char)v43] = 0;
				*(uint32_t*)(v1 + 1360) = nox_xxx_actionNByNameMB_5345F0(v57);
			} else {
				v34 = sub_5345B0(*(uint32_t*)(v1 + 1360));
				LOBYTE(v43) = strlen(v34);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v43, 1u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v34, (unsigned char)v43);
			}
		}
	}
	if ((short)v44 < 41) {
		v35 = a1;
	} else {
		v35 = a1;
		result = nox_xxx_XFer_ActionData_529CE0(a1);
		if (!result) {
			return result;
		}
	}
	if ((short)v44 >= 42) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 1445), 1u);
	}
	if ((short)v44 >= 44) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)v1, 4u);
	}
	if ((short)v44 >= 45) {
		v36 = *(uint32_t*)(v35 + 556);
		v45 = 0;
		if (v36) {
			v45 = *(unsigned short*)(v36 + 4);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 4u);
		v37 = *(uint32_t*)(v35 + 556);
		if (v37) {
			*(uint16_t*)(v37 + 4) = v45;
		}
	}
	if ((short)v44 >= 46) {
		v51 = *(uint32_t*)(v35 + 12) & 0x180;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v51, 4u);
		*(uint32_t*)(v35 + 12) |= v51;
	}
	if ((short)v44 >= 48) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(*(uint8_t**)(v35 + 556), 2u);
	}
	if ((short)v44 >= 51) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v35 + 28), 4u);
	}
	if ((short)v44 >= 52) {
		nox_xxx_readNPCVoiceSet_52AD10(v35);
	}
	if (!((short)v44 < 61 || (result = nox_xxx_XFer_ReadMonsterBuffs_52AAB0((uint32_t*)v35)) != 0)) {
		return result;
	}
	if ((short)v44 < 62) {
		if (nox_crypt_IsReadOnly() == 1) {
			goto LABEL_149;
		}
		goto LABEL_156;
	}
	LOBYTE(v45) = *(uint8_t*)(v35 + 540);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v45, 1u);
	if (nox_crypt_IsReadOnly() != 1) {
		goto LABEL_156;
	}
	if ((uint8_t)v45) {
		nox_xxx_setSomePoisonData_4EEA90(v35, (unsigned char)v45);
		if (nox_crypt_IsReadOnly() == 1) {
			goto LABEL_149;
		}
		goto LABEL_156;
	}
LABEL_149:
	if (!*(uint8_t*)(v1 + 1445)) {
		goto LABEL_170;
	}
	if (nox_common_gameFlags_check_40A5C0(1)) {
		v38 = *(uint32_t*)(v35 + 556);
		if (v38) {
			*(uint16_t*)(v38 + 4) = 0;
			**(uint16_t**)(v35 + 556) = 0;
		}
	}
	if (nox_crypt_IsReadOnly() != 1) {
		goto LABEL_156;
	}
LABEL_170:
	v39 = *(uint32_t*)(v35 + 16);
	if ((v39 & 0x8000) != 0) {
		LOBYTE(v39) = v39 | 0x40;
		*(uint32_t*)(v35 + 16) = v39;
	}
LABEL_156:
	if (!*(uint32_t*)(v35 + 136) || nox_crypt_IsReadOnly() != 1 ||
		(result = nox_xxx_xfer_4F3E30(v44, v35, *(uint32_t*)(v35 + 136))) != 0) {
		sub_52BA70(v35);
		if (nox_crypt_IsReadOnly() == 1) {
			if (nox_common_gameFlags_check_40A5C0(1)) {
				for (k = *(uint32_t**)(v35 + 504); k; k = (uint32_t*)k[124]) {
					v41 = k[4];
					if (v41 & 0x100) {
						v42 = k[2];
						k[4] &= 0xFFFFFEFF;
						if (v42 & 0x1001000) {
							nox_xxx_NPCEquipWeapon_53A2C0(v35, (int)k);
						} else {
							nox_xxx_NPCEquipArmor_53E520(v35, k);
						}
					}
				}
			}
		}
		*(uint32_t*)(v35 + 136) = v54;
		return 1;
	}
	return result;
}
// 52B1CD: variable 'v21' is possibly undefined
// 52ADE0: using guessed type char var_200[256];
// 52ADE0: using guessed type char var_100[256];
#else
extern int32_t nox_xxx_XFerNPC_native_52ADE0(nox_object_t* obj);
int nox_xxx_XFerNPC_52ADE0(nox_object_t* obj) { return nox_xxx_XFerNPC_native_52ADE0(obj); }
#endif

//----- (0052BA70) --------------------------------------------------------
int sub_52BA70(int a1) {
	int result; // eax
	int v2;     // edi
	int v3;     // esi
	int v4;     // ecx
	int v5;     // edx

	result = a1;
	v2 = 0;
	v3 = 0;
	if (a1) {
		for (result = *(uint32_t*)(a1 + 504); result; result = *(uint32_t*)(result + 496)) {
			v4 = *(uint32_t*)(result + 16);
			if (v4 & 0x100) {
				v5 = *(uint32_t*)(result + 8);
				if (v5 & 0x1001000 && *(uint32_t*)(result + 12) & 0x7FFE40C) {
					if (v3 == 1) {
						BYTE1(v4) &= 0xFEu;
						*(uint32_t*)(result + 16) = v4;
					} else {
						v2 = 1;
					}
				} else if (v5 & 0x2000000 && *(uint8_t*)(result + 12) & 2) {
					if (v2 == 1) {
						BYTE1(v4) &= 0xFEu;
						*(uint32_t*)(result + 16) = v4;
					} else {
						v3 = 1;
					}
				}
			}
		}
	}
	return result;
}

//----- (0052BAF0) --------------------------------------------------------
int sub_52BAF0(int a1) {
	int v1;       // eax
	int result;   // eax
	uint32_t* v3; // ebx
	int v4;       // edi
	int v5;       // ebp
	int* v6;      // esi
	int v7;       // [esp+0h] [ebp-4h]
	int v8;       // [esp+8h] [ebp+4h]

	v1 = a1;
	v8 = 0;
	result = *(uint32_t*)(v1 + 748);
	v7 = result;
	if (*(uint8_t*)(result + 544) & 0x80) {
		return result;
	}
	v3 = (uint32_t*)(result + 552);
	do {
		v4 = 0;
		v5 = *getMemU32Ptr(0x587000, 255604 + 16 * *v3);
		if (v5 <= 0) {
			goto LABEL_14;
		}
		v6 = v3 + 1;
		do {
			if (*getMemU32Ptr(0x587000, 255608 + 4 * (v4 + 4 * *v3)) == 1) {
				if (*v6) {
					*v6 = sub_4ECF10(*v6);
					goto LABEL_12;
				}
			} else {
				if (*getMemU32Ptr(0x587000, 255608 + 4 * (v4 + 4 * *v3)) != 2) {
					goto LABEL_12;
				}
				if (*v6) {
					*v6 = nox_server_getWaypointById_579C40(*v6);
					goto LABEL_12;
				}
			}
			*v6 = 0;
		LABEL_12:
			++v4;
			v6 += 2;
		} while (v4 < v5);
		result = v7;
	LABEL_14:
		v3 += 6;
		++v8;
	} while (v8 <= *(char*)(result + 544));
	return result;
}

//----- (0052BEB0) --------------------------------------------------------
int sub_52BEB0(int a1, int a2, int a3, int a4) {
	int v4;   // eax
	float v6; // [esp+0h] [ebp-10h]

	v6 = nox_xxx_gamedataGetFloat_419D40("InversionRange");
	nox_xxx_getMissilesInCircle_518170(a4 + 56, v6, nox_xxx_changeOwner_52BE40, a3);
	v4 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v4, a4, 0, 0);
	return 1;
}

//----- (0052BF20) --------------------------------------------------------
int nox_xxx_castSpellWinkORrestoreHealth_52BF20(int a1, int a2, int a3, int a4, int* a5) {
	int result; // eax

	result = *a5;
	if (*a5) {
		nox_xxx_unitHPsetOnMax_4EE6F0((nox_object_t*)(uintptr_t)(uint32_t)*a5);
		nox_xxx_aud_501960(754, *a5, 0, 0);
		result = 1;
	}
	return result;
}

//----- (0052BF50) --------------------------------------------------------
int sub_52BF50(int a1, int a2, int a3, int a4, int* a5) {
	int result; // eax

	result = *a5;
	if (*a5) {
		if (*(uint8_t*)(result + 8) & 4) {
			nox_xxx_playerManaAdd_4EEB80(result, *(uint16_t*)(*(uint32_t*)(result + 748) + 8));
			nox_xxx_aud_501960(755, *a5, 0, 0);
		}
		result = 1;
	}
	return result;
}

//----- (0052BFA0) --------------------------------------------------------
int nox_xxx_castPull_52BFA0(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;   // eax
	float v8; // [esp+0h] [ebp-10h]

	v8 = -(nox_xxx_gamedataGetFloat_419D40("PullPowerCoeff") * (double)a6);
	nox_xxx_mapPushUnitsAround_52E040(a4 + 56, 600.0, 10.0, v8, 0, 0, 0);
	v6 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v6, a3, 0, 0);
	return 1;
}

//----- (0052C000) --------------------------------------------------------
int nox_xxx_castPush_52C000(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;   // eax
	float v8; // [esp+18h] [ebp+18h]

	v8 = nox_xxx_gamedataGetFloat_419D40("PushPowerCoeff") * (double)a6;
	nox_xxx_mapPushUnitsAround_52E040(a4 + 56, 600.0, 10.0, v8, 0, 0, 0);
	v6 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v6, a3, 0, 0);
	return 1;
}

//----- (0052C060) --------------------------------------------------------
int nox_xxx_castFumble_52C060(int a1, int a2, int a3, int a4, int* a5) {
	int result; // eax
	int v6;     // ecx
	int v7;     // ecx
	int v8;     // eax
	int v9;     // ecx
	int v10;    // esi
	int v11;    // ecx
	int v12;    // eax
	int v13;    // esi
	int v14;    // eax
	int v15;    // [esp-10h] [ebp-14h]

	result = *a5;
	if (*a5) {
		if (*(uint32_t*)(result + 8) & 4 ||
			(v6 = *(uint32_t*)(result + 8) & 2) != 0 && *(uint8_t*)(result + 12) & 0x10) {
			v8 = *(uint32_t*)(result + 504);
			if (v8) {
				do {
					v9 = *(uint32_t*)(v8 + 16);
					v10 = *(uint32_t*)(v8 + 496);
					if (v9 & 0x100) {
						v11 = *(uint32_t*)(v8 + 8);
						if (v11 & 0x1001000 || v11 & 0x2000000 && *(uint8_t*)(v8 + 12) & 2) {
						nox_xxx_invForceDropItem_4ED930(
							(nox_object_t*)(uintptr_t)(uint32_t)*a5,
							(nox_object_t*)(uintptr_t)(uint32_t)v8);
						}
					}
					v8 = v10;
				} while (v10);
			}
			v12 = *getMemU32Ptr(0x5D4594, 2487728);
			if (!*getMemU32Ptr(0x5D4594, 2487728)) {
				v12 = nox_xxx_getNameId_4E3AA0("GameBall");
				*getMemU32Ptr(0x5D4594, 2487728) = v12;
			}
			v13 = *(uint32_t*)(*a5 + 516);
			if (v13) {
				while (*(unsigned short*)(v13 + 4) != v12) {
					v13 = *(uint32_t*)(v13 + 512);
					if (!v13) {
						goto LABEL_22;
					}
				}
				nox_xxx_objectApplyForce_52DF80(*a5 + 56, v13, 100.0);
				nox_xxx_unitClearOwner_4EC300(v13);
				nox_xxx_aud_501960(926, *a5, 0, 0);
			}
		} else if (!v6 || (v7 = *(uint32_t*)(result + 12), !(v7 & 0x2000))) {
			nox_xxx_dropAllItems_4EDA40((nox_object_t*)(uintptr_t)result);
			nox_xxx_objectApplyForce_52DF80(a4 + 56, *a5, 50.0);
		}
	LABEL_22:
		v15 = *a5;
		v14 = nox_xxx_spellGetAud44_424800(a1, 1);
		nox_xxx_aud_501960(v14, v15, 0, 0);
		result = 1;
	}
	return result;
}

//----- (0052C3E0) --------------------------------------------------------
int nox_xxx_castBurn_52C3E0(int a1, int a2, int a3, int a4, int a5) {
	int v5;     // eax
	int v6;     // ebp
	float* v7;  // ebx
	float v8;   // edx
	float v9;   // ecx
	float v10;  // edx
	int v11;    // edx
	int v13;    // eax
	float* v14; // esi
	int v15;    // eax
	int v16;    // eax
	float v17;  // [esp+0h] [ebp-24h]
	float4 v18; // [esp+14h] [ebp-10h]

	v5 = dword_5d4594_2487712;
	if (!dword_5d4594_2487712) {
		v5 = nox_xxx_getNameId_4E3AA0("Glyph");
		dword_5d4594_2487712 = v5;
	}
	v6 = a5;
	if (!a5 || !a4) {
		return 0;
	}
	if (*(unsigned short*)(a4 + 4) != v5) {
		v8 = *(float*)(a4 + 56);
		v9 = *(float*)(a5 + 4);
		v18.field_4 = *(float*)(a4 + 60);
		v18.field_0 = v8;
		v10 = *(float*)(a5 + 8);
		v18.field_8 = v9;
		v18.field_C = v10;
		if ((unsigned char)nox_xxx_traceRay_5374B0(&v18)) {
			v7 = (float*)(v6 + 4);
			goto LABEL_12;
		}
		if (*(uint8_t*)(a4 + 8) & 4) {
			v11 = *(uint32_t*)(a4 + 748);
			a5 = 2;
			nox_xxx_netInformTextMsg_4DA0F0(*(unsigned char*)(*(uint32_t*)(v11 + 276) + 2064), 0, &a5);
		}
		return 0;
	}
	v7 = (float*)(a4 + 56);
LABEL_12:
	v13 = *getMemU32Ptr(0x5D4594, 2487732);
	if (!*getMemU32Ptr(0x5D4594, 2487732)) {
		v13 = nox_xxx_getNameId_4E3AA0("MediumFlame");
		*getMemU32Ptr(0x5D4594, 2487732) = v13;
	}
	v14 = (float*)nox_xxx_newObjectWithTypeInd_4E3450(v13);
	if (v14) {
		nox_xxx_createAt_4DAA50((int)v14, a4, *v7, v7[1]);
		v17 = nox_xxx_gamedataGetFloat_419D40("BurnDuration");
		v15 = nox_float2int(v17);
		nox_xxx_unitSetDecayTime_511660(v14, v15);
		nox_xxx_netSparkExplosionFx_5231B0(v14 + 14, 64);
	}
	v16 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_audCreate_501A30(v16, (float2*)(v6 + 4), 0, 0);
	return 1;
}

//----- (0052C5A0) --------------------------------------------------------
// Native-width implementation: server/cast_shock_52c5a0.go and the typed
// legacy/cast_shock_52c5a0_export.go entry. Retain this PE32 body only as
// provenance; its integer context and first-dword target truncate pointers.
#if 0
int nox_xxx_useShock_52C5A0(int a1, int a2, int a3, int a4, int* a5, int a6) {
	int result; // eax
	int v7;     // eax
	int v8;     // eax
	short v9;   // ax
	float v10;  // [esp+0h] [ebp-8h]
	float v11;  // [esp+0h] [ebp-8h]

	if (!*a5) {
		return 0;
	}
	v7 = dword_5d4594_2487712;
	if (!dword_5d4594_2487712) {
		v7 = nox_xxx_getNameId_4E3AA0("Glyph");
		dword_5d4594_2487712 = v7;
	}
	if (a4 && *(unsigned short*)(a4 + 4) == v7) {
		v10 = nox_xxx_gamedataGetFloatTable_419D70("ShockTrapDamage", a6 - 1);
		v8 = nox_float2int(v10);
		(*(void (**)(int, int, int, int, int))(*a5 + 716))(*a5, a3, a3, v8, 9);
		result = 1;
	} else {
		v11 = nox_xxx_gamedataGetFloat_419D40("ShockEnchantDuration");
		v9 = nox_float2int(v11);
		nox_xxx_buffApplyTo_4FF380((nox_object_t*)(uintptr_t)(uint32_t)*a5, 22, v9, a6);
		result = 1;
	}
	return result;
}
#endif

//----- (0052C720) --------------------------------------------------------
int nox_xxx_castPoison_52C720(int a1, int a2, int a3, int a4, int* a5, int a6) {
	int result; // eax

	result = *a5;
	if (*a5) {
		nox_xxx_activatePoison_4EE7E0(result, a6, a6);
		sub_4E7540((nox_object_t*)(uintptr_t)a3, (nox_object_t*)(uintptr_t)*a5);
		result = 1;
	}
	return result;
}

//----- (0052C790) --------------------------------------------------------
int nox_xxx_castFireball_52C790(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;     // ebx
	float* v7;  // esi
	int v8;     // edi
	double v9;  // st7
	int v10;    // eax
	float v11;  // ecx
	double v12; // st6
	double v13; // st7
	double v14; // st7
	double v15; // st7
	short v16;  // t0
	int v17;    // eax
	float4 v19; // [esp+8h] [ebp-10h]
	float v20;  // [esp+28h] [ebp+10h]
	float v21;  // [esp+30h] [ebp+18h]
	float v22;  // [esp+30h] [ebp+18h]

	v6 = a6;
	// This table is part of the original PE32 image, so its entries are packed
	// four bytes apart. On a 64-bit host, dereferencing it as char** reads two
	// adjacent PE32 slots as one native pointer. mem_getPtrValue keeps the
	// packed layout while returning the native-width side-table value.
	v7 = (float*)nox_xxx_newObjectByTypeID_4E3810((char*)getMemPtr(0x587000, 258864 + 4 * a6));
	if (!v7) {
		return 1;
	}
	v8 = a4;
	v9 = *(float*)(a4 + 176) + *(float*)(a4 + 176);
	v10 = 8 * *(short*)(a4 + 124);
	v21 = *getMemFloatPtr(0x587000, 194136 + v10);
	v20 = *getMemFloatPtr(0x587000, 194140 + v10);
	v11 = *(float*)(v8 + 60);
	v19.field_0 = *(float*)(v8 + 56);
	v12 = v9 * v21 + *(float*)(v8 + 56);
	v19.field_4 = v11;
	v19.field_8 = v12;
	v13 = v9 * v20 + *(float*)(v8 + 60);
	v19.field_8 = v19.field_8 + *(float*)(v8 + 80);
	v19.field_C = v13 + *(float*)(v8 + 84);
	if (!nox_xxx_mapTraceRay_535250(&v19, 0, 0, 5)) {
		v19.field_8 = v19.field_0;
		v19.field_C = v19.field_4;
	}
	nox_xxx_createAt_4DAA50((int)v7, v8, v19.field_8, v19.field_C);
	v14 = nox_xxx_gamedataGetFloatTable_419D70("FireballSpeedCoeff", v6 - 1) * v7[136];
	v7[136] = v14;
	v22 = v14 * v21;
	v7[20] = v22;
	v15 = v14 * v20;
	v7[21] = v15;
	v7[20] = v22 + *(float*)(v8 + 80);
	v7[21] = v15 + *(float*)(v8 + 84);
	v16 = *(uint16_t*)(v8 + 124);
	*((uint16_t*)v7 + 62) = *(uint16_t*)(v8 + 124);
	*((uint16_t*)v7 + 63) = v16;
	v17 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v17, v8, 0, 0);
	return 1;
}

//----- (0052CA80) --------------------------------------------------------
int sub_52CA80(int a1, int a2, int a3, int a4) {
	int v4;             // eax
	int v5;             // eax
	float2* v6;         // ebp
	int v7;             // edi
	int v8;             // esi
	int v9;             // ebx
	uint32_t* v10;      // eax
	unsigned int v11;   // edx
	int i;              // eax
	int v13;            // eax
	uint32_t* v14;      // eax
	int v15;            // eax
	const char* v17[4]; // [esp+0h] [ebp-10h]

	v4 = dword_5d4594_2487712;
	if (!dword_5d4594_2487712) {
		v4 = nox_xxx_getNameId_4E3AA0("Glyph");
		dword_5d4594_2487712 = v4;
	}
	if (a4 && *(unsigned short*)(a4 + 4) == v4) {
		v5 = a3;
		v6 = (float2*)(a4 + 56);
	} else {
		v5 = a3;
		v6 = (float2*)(a3 + 56);
	}
	if (*(uint8_t*)(v5 + 8) & 4) {
		v7 = *(uint32_t*)(v5 + 748);
		v8 = 0;
		v9 = v7 + 116;
		v10 = (uint32_t*)(v7 + 116);
		do {
			if (!*v10) {
				break;
			}
			++v8;
			++v10;
		} while (v8 < 4);
		if (v8 == 4) {
			v11 = gameFrame();
			v8 = a3;
			for (i = 0; i < 4; ++i) {
				if (*(uint32_t*)(*(uint32_t*)v9 + 136) < v11) {
					v11 = *(uint32_t*)(*(uint32_t*)v9 + 136);
					v8 = i;
				}
				v9 += 4;
			}
		}
		v13 = *(uint32_t*)(v7 + 4 * v8 + 116);
		if (v13) {
			nox_xxx_unitMove_4E7010(v13, v6);
			*(uint32_t*)(*(uint32_t*)(v7 + 4 * v8 + 116) + 136) = gameFrame();
		} else {
			v17[0] = "TeleportGlyph1";
			v17[1] = "TeleportGlyph2";
			v17[2] = "TeleportGlyph3";
			v17[3] = "TeleportGlyph4";
			v14 = nox_xxx_newObjectByTypeID_4E3810(v17[v8]);
			*(uint32_t*)(v7 + 4 * v8 + 116) = v14;
			if (!v14) {
				v15 = nox_xxx_spellGetAud44_424800(a1, 0);
				nox_xxx_aud_501960(v15, a3, 0, 0);
				return 1;
			}
			nox_xxx_createAt_4DAA50((int)v14, a3, v6->field_0, v6->field_4);
		}
		*(uint8_t*)(v8 + v7 + 156) = 3;
		v15 = nox_xxx_spellGetAud44_424800(a1, 0);
		nox_xxx_aud_501960(v15, a3, 0, 0);
		return 1;
	}
	return 1;
}

//----- (0052CBD0) --------------------------------------------------------
int sub_52CBD0(int a1, int a2, int a3, int a4) {
	int v4;       // eax
	int v5;       // ebx
	int v6;       // edi
	int v7;       // eax
	uint32_t* v8; // eax
	float v9;     // ecx
	float v10;    // edx
	int v11;      // eax
	int v14;      // [esp+9Ch] [ebp-2Ch]
	int v15;      // [esp+A0h] [ebp-28h]
	int v16;      // [esp+A4h] [ebp-24h]
	const char* v17[4];

	v4 = dword_5d4594_2487712;
	if (!dword_5d4594_2487712) {
		v4 = nox_xxx_getNameId_4E3AA0("Glyph");
		dword_5d4594_2487712 = v4;
	}
	if (!a4 || (v5 = a4 + 56, *(unsigned short*)(a4 + 4) != v4)) {
		v5 = a3 + 56;
	}
	if (*(uint8_t*)(a3 + 8) & 4) {
		v6 = *(uint32_t*)(a3 + 748);
		v7 = *(uint32_t*)(v6 + 4 * a1 - 68);
		if (v7) {
			nox_xxx_unitMove_4E7010(v7, (float2*)v5);
			*(uint32_t*)(*(uint32_t*)(v6 + 4 * a1 - 68) + 136) = gameFrame();
		} else {
			v17[0] = "TeleportGlyph1";
			v17[1] = "TeleportGlyph2";
			v17[2] = "TeleportGlyph3";
			v17[3] = "TeleportGlyph4";
			v8 = nox_xxx_newObjectByTypeID_4E3810(v17[a1 - 46]);
			*(uint32_t*)(v6 + 4 * a1 - 68) = v8;
			if (!v8) {
				v16 = 0;
				v15 = 0;
				v14 = a3;
				v11 = nox_xxx_spellGetAud44_424800(a1, 0);
				nox_xxx_aud_501960(v11, v14, v15, v16);
				return 1;
			}
			v9 = *(float*)(v5 + 4);
			v10 = *(float*)v5;
			v16 = 0;
			nox_xxx_createAt_4DAA50((int)v8, a3, v10, v9);
		}
		*(uint8_t*)(a1 + v6 + 110) = 3;
		v16 = 0;
		v15 = 0;
		v14 = a3;
		v11 = nox_xxx_spellGetAud44_424800(a1, 0);
		nox_xxx_aud_501960(v11, v14, v15, v16);
		return 1;
	}
	return 1;
}
// 52CBD0: using guessed type int var_C8[39];

//----- (0052CCD0) --------------------------------------------------------
// Native-width implementation: server/spell_trigger_glyph_52ccd0.go, selected
// through legacy/spell_trigger_glyph_52ccd0.go. Keep the original body as
// provenance; the three-int entry truncates world objects and the caster.
#if 0
int sub_52CCD0(int a1, int a2, int a3) {
	int v3;        // ebx
	double v4;     // st7
	double v5;     // st6
	double v6;     // st5
	int v7;        // eax
	float v9;      // [esp+Ch] [ebp-8h]
	struct tm* Tm; // [esp+10h] [ebp-4h]

	Tm = 0;
	v9 = 100000000.0;
	v3 = nox_server_getFirstObject_4DA790();
	if (!v3) {
		return 0;
	}
	do {
		if (nox_xxx_unitHasThatParent_4EC4F0(v3, a3) && !strcmp((const char*)nox_xxx_getUnitName_4E39D0(v3), "Glyph")) {
			v4 = *(float*)(a3 + 56) - *(float*)(v3 + 56);
			v5 = *(float*)(a3 + 60) - *(float*)(v3 + 60);
			v6 = v5 * v5 + v4 * v4;
			if (v6 < v9) {
				v9 = v6;
				Tm = (struct tm*)v3;
			}
		}
		v3 = nox_server_getNextObject_4DA7A0(v3);
	} while (v3);
	if (!Tm) {
		return 0;
	}
	v7 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v7, a3, 0, 0);
	nox_xxx_dieGlyph_54DF30(Tm);
	return 1;
}
#endif

//----- (0052CDB0) --------------------------------------------------------
// Native-width implementation: server/spell_cure_poison_52cdb0.go, shared
// by the public legacy selector and the game spell dispatcher. Keep the
// original body as provenance; its int* target load truncates native objects.
#if 0
int nox_xxx_castCurePoison_52CDB0(int a1, int a2, int a3, int a4, int* a5, int a6) {
	int result; // eax
	int v7;     // eax
	short v8;   // ax
	int v9;     // [esp-Ch] [ebp-10h]

	result = *a5;
	if (*a5) {
		if (*(uint8_t*)(result + 540)) {
			if (*(unsigned char*)(result + 540) > a6) {
				nox_xxx_updatePoison_4EE8F0(result, a6);
				nox_xxx_netPriMsgToPlayer_4DA2C0(*a5, "ExecSpel.c:PoisonCure", 0);
			} else {
				nox_xxx_removePoison_4EE9D0(*a5);
				nox_xxx_netPriMsgToPlayer_4DA2C0(*a5, "ExecSpel.c:PoisonClean", 0);
			}
			v9 = *a5;
			v7 = nox_xxx_spellGetAud44_424800(a1, 1);
			nox_xxx_aud_501960(v7, v9, 0, 0);
			return 1;
		}
		if (result != a2) {
			v9 = *a5;
			v7 = nox_xxx_spellGetAud44_424800(a1, 1);
			nox_xxx_aud_501960(v7, v9, 0, 0);
			return 1;
		}
		v8 = nox_xxx_spellManaCost_4249A0(a1, 1);
		sub_4FD030((nox_object_t*)(uintptr_t)(uint32_t)*a5, v8);
		result = 1;
	}
	return result;
}
#endif

//----- (0052CE60) --------------------------------------------------------
void sub_52CE60(int a1) {
	if (*(uint8_t*)(a1 + 8) & 0x80) {
		*(uint32_t*)(a1 + 508) = *getMemU32Ptr(0x5D4594, 2487716);
		*(uint32_t*)(a1 + 136) = gameFrame() + 60 * gameFPS();
	}
}

//----- (0052CE90) --------------------------------------------------------
// Native-width entry: legacy/spell_lock_52ce90.go and server/spell_lock_52ce90.go.
// Keep the original integer-pointer body as disabled provenance.
#if 0
int nox_xxx_castLock_52CE90(int a1, int a2, int a3, int a4) {
	int v5;    // ecx
	int v7;    // eax
	int v8;    // [esp-Ch] [ebp-20h]
	float4 v9; // [esp+4h] [ebp-10h]

	v9.field_0 = *(float*)(a3 + 56) - 150.0;
	v9.field_4 = *(float*)(a3 + 60) - 150.0;
	v9.field_8 = *(float*)(a3 + 56) + 150.0;
	v9.field_C = *(float*)(a3 + 60) + 150.0;
	dword_5d4594_2487708 = 0;
	*getMemU32Ptr(0x5D4594, 2487704) = 1287568416;
	nox_xxx_getUnitsInRect_517C10(&v9, sub_52CF90, a4);
	if (!dword_5d4594_2487708) {
		return 0;
	}
	v5 = *(uint32_t*)(dword_5d4594_2487708 + 508);
	if (v5 && v5 != a3) {
		nox_xxx_netPriMsgToPlayer_4DA2C0(a3, "ExecSpel.c:DoorAlreadyLocked", 0);
		return 0;
	}
	*(uint32_t*)(dword_5d4594_2487708 + 508) = a3;
	*(uint32_t*)(dword_5d4594_2487708 + 136) = gameFrame() + 60 * gameFPS();
	sub_52D060(*(int*)&dword_5d4594_2487708, a3);
	v8 = dword_5d4594_2487708;
	v7 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v7, v8, 0, 0);
	return 1;
}
#endif

//----- (0052CF90) --------------------------------------------------------
void sub_52CF90(int a1, int a2) {
	int v2;    // esi
	double v3; // st7
	double v4; // st6
	int v5;    // eax
	float v6;  // ecx
	float4 v7; // [esp+4h] [ebp-10h]
	float v8;  // [esp+18h] [ebp+4h]

	v2 = a1;
	if (*(uint8_t*)(a1 + 8) & 0x80) {
		v3 = *(float*)(a2 + 56) - *(float*)(a1 + 56);
		v4 = *(float*)(a2 + 60) - *(float*)(a1 + 60);
		v8 = v4 * v4 + v3 * v3;
		if (v8 <= 22500.0 && v8 < (double)*getMemFloatPtr(0x5D4594, 2487704)) {
			v5 = *(uint32_t*)(v2 + 748);
			v6 = *(float*)(a2 + 60);
			v7.field_0 = *(float*)(a2 + 56);
			v7.field_4 = v6;
			v7.field_8 = (double)*getMemIntPtr(0x587000, 196184 + 8 * *(uint32_t*)(v5 + 12)) * 0.5 + *(float*)(v2 + 56);
			v7.field_C = (double)*getMemIntPtr(0x587000, 196188 + 8 * *(uint32_t*)(v5 + 12)) * 0.5 + *(float*)(v2 + 60);
			if (nox_xxx_mapTraceRay_535250(&v7, 0, 0, 0)) {
				dword_5d4594_2487708 = v2;
				*getMemFloatPtr(0x5D4594, 2487704) = v4 * v4 + v3 * v3;
			}
		}
	}
}

//----- (0052D060) --------------------------------------------------------
void sub_52D060(int a1, int a2) {
	int v2;    // eax
	int v3;    // eax
	float4 v5; // [esp+0h] [ebp-10h]

	v2 = *(uint32_t*)(a1 + 748);
	v5.field_0 = (double)(int)(23 * *(uint32_t*)(v2 + 16)) - 34.0;
	v5.field_4 = (double)(int)(23 * *(uint32_t*)(v2 + 20)) - 34.0;
	v5.field_8 = (double)(int)(23 * *(uint32_t*)(v2 + 16)) + 34.0;
	v3 = *(uint32_t*)(v2 + 20);
	*getMemU32Ptr(0x5D4594, 2487716) = a2;
	v5.field_C = (double)(23 * v3) + 34.0;
	nox_xxx_getUnitsInRect_517C10(&v5, sub_52CE60, 0);
	*getMemU32Ptr(0x5D4594, 2487716) = 0;
}

//----- (0052D330) --------------------------------------------------------
// Native-width implementation: server/spell_telekinesis_52d330.go, selected
// through legacy/spell_telekinesis_52d330.go. Retain the body as provenance
// without compiling an entry whose integer target and argument truncate pointers.
#if 0
int nox_xxx_castTelekinesis_52D330(int a1, int a2, int a3, int a4, int* a5, char a6) {
	int result;   // eax
	uint32_t* v7; // eax
	int v8;       // eax
	int v9;       // [esp-Ch] [ebp-10h]

	result = *a5;
	if (*a5) {
		if (*(uint8_t*)(result + 8) & 4) {
			v7 = nox_xxx_newObjectByTypeID_4E3810("TelekinesisHand");
			if (v7) {
				nox_xxx_createAt_4DAA50((int)v7, *a5, *(float*)(*a5 + 56), *(float*)(*a5 + 60));
				nox_xxx_buffApplyTo_4FF380(
					(nox_object_t*)(uintptr_t)(uint32_t)*a5, 24, 20 * (uint16_t)gameFPS(), a6);
				nox_xxx_spellCancelDurSpell_4FEB10(24, *a5);
				nox_xxx_spellCancelDurSpell_4FEB10(43, *a5);
				v9 = *a5;
				v8 = nox_xxx_spellGetAud44_424800(a1, 1);
				nox_xxx_aud_501960(v8, v9, 0, 0);
			}
		}
		result = 1;
	}
	return result;
}
#endif

//----- (0052D3C0) --------------------------------------------------------
// Native-width implementation: server/spell_fist_52d3c0.go, selected through
// legacy/spell_fist_52d3c0.go. Retain the original body only as provenance;
// do not compile an entry whose integer object arguments truncate pointers.
#if 0
int nox_xxx_castFist_52D3C0(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;            // edx
	unsigned char* v7; // eax
	float v8;          // ecx
	float v9;          // edx
	float v10;         // eax
	int v11;           // eax
	int v12;           // eax
	int result;        // eax
	uint32_t* v14;     // eax
	int v15;           // esi
	uint32_t* v16;     // ebp
	double v17;        // st7
	unsigned int v18;  // ecx
	int v19;           // eax
	float v20;         // [esp+0h] [ebp-24h]
	float4 v21;        // [esp+14h] [ebp-10h]

	if (!*getMemU32Ptr(0x5D4594, 2487736)) {
		*getMemU32Ptr(0x5D4594, 2487740) = nox_xxx_getNameId_4E3AA0("SmallFist");
		*getMemU32Ptr(0x5D4594, 2487744) = nox_xxx_getNameId_4E3AA0("MediumFist");
		*getMemU32Ptr(0x5D4594, 2487748) = nox_xxx_getNameId_4E3AA0("LargeFist");
		*getMemU32Ptr(0x5D4594, 2487752) = nox_xxx_getNameId_4E3AA0("LargeFist");
		*getMemU32Ptr(0x5D4594, 2487756) = nox_xxx_getNameId_4E3AA0("LargeFist");
		*getMemU32Ptr(0x5D4594, 2487736) = 1;
	}
	v6 = *(uint32_t*)(a3 + 516);
	if (v6) {
		while (2) {
			v7 = getMemAt(0x5D4594, 2487740);
			do {
				if (*(unsigned short*)(v6 + 4) == *(uint32_t*)v7) {
					nox_xxx_netPriMsgToPlayer_4DA2C0(a3, "ExecSpel.c:TooManyFists", 0);
					return 0;
				}
				v7 += 4;
			} while ((int)v7 < (int)getMemAt(0x5D4594, 2487760));
			v6 = *(uint32_t*)(v6 + 512);
			if (v6) {
				continue;
			}
			break;
		}
	}
	v8 = *(float*)(a4 + 60);
	v9 = *(float*)(a5 + 4);
	v21.field_0 = *(float*)(a4 + 56);
	v10 = *(float*)(a5 + 8);
	v21.field_4 = v8;
	v21.field_8 = v9;
	v21.field_C = v10;
	v11 = nox_common_gameFlags_check_40A5C0(2048);
	if (nox_xxx_mapTraceRay_535250(&v21, 0, 0, !v11 ? 73 : 9)) {
		v14 = nox_xxx_newObjectWithTypeInd_4E3450(*getMemU32Ptr(0x5D4594, 2487736 + 4 * a6));
		v15 = (int)v14;
		if (v14) {
			v16 = (uint32_t*)v14[187];
			v20 = nox_xxx_gamedataGetFloatTable_419D70("FistOfVengeanceDamage", a6 - 1);
			*v16 = nox_float2int(v20);
			nox_xxx_createAt_4DAA50(v15, a4, *(float*)(a5 + 4), *(float*)(a5 + 8));
			*(uint32_t*)(v15 + 20) |= 0x20u;
			nox_xxx_unitRaise_4E46F0(v15, 255.0);
			v17 = nox_xxx_gamedataGetFloat_419D40("FistSpeed");
			v18 = 0x800000 | *(uint32_t*)(v15 + 16);
			*(float*)(v15 + 108) = -v17;
			*(uint32_t*)(v15 + 16) = v18;
			*(uint32_t*)(v15 + 116) = 1091567616;
			v19 = nox_xxx_spellGetAud44_424800(a1, 0);
			nox_xxx_aud_501960(v19, a4, 0, 0);
		}
		result = 1;
	} else {
		if (*(uint8_t*)(a4 + 8) & 4) {
			v12 = *(uint32_t*)(a4 + 748);
			a6 = 2;
			nox_xxx_netInformTextMsg_4DA0F0(*(unsigned char*)(*(uint32_t*)(v12 + 276) + 2064), 0, &a6);
		}
		result = 0;
	}
	return result;
}
#endif

//----- (0052D5C0) --------------------------------------------------------
int nox_xxx_spellCastCleansingFlame_52D5C0(int a1, nox_object_t* a2p, nox_object_t* a3p, nox_object_t* a4p, void* a5p, int a6) {
	int a2 = a2p;
	int a3 = a3p;
	int a4 = a4p;
	int a5 = a5p;
	int v6;            // edx
	unsigned char* v7; // eax
	int v8;            // ecx
	float* v9;         // esi
	short v10;         // ax
	double v11;        // st7
	double v12;        // st7
	int v13;           // eax
	float v14;         // edx
	double v15;        // st6
	int v17;           // edx
	int v18;           // eax
	int v19;           // [esp+0h] [ebp-14h]
	float4 v20;        // [esp+4h] [ebp-10h]
	float v21;         // [esp+1Ch] [ebp+8h]
	float v22;         // [esp+20h] [ebp+Ch]

	if (!*getMemU32Ptr(0x5D4594, 2487760)) {
		*getMemU32Ptr(0x5D4594, 2487760) = nox_xxx_getNameId_4E3AA0("SmallFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487764) = nox_xxx_getNameId_4E3AA0("SmallFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487768) = nox_xxx_getNameId_4E3AA0("MediumFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487772) = nox_xxx_getNameId_4E3AA0("FlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487776) = nox_xxx_getNameId_4E3AA0("LargeFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487780) = nox_xxx_getNameId_4E3AA0("SmallBlueFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487784) = nox_xxx_getNameId_4E3AA0("SmallBlueFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487788) = nox_xxx_getNameId_4E3AA0("MediumBlueFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487792) = nox_xxx_getNameId_4E3AA0("BlueFlameCleanse");
		*getMemU32Ptr(0x5D4594, 2487796) = nox_xxx_getNameId_4E3AA0("LargeBlueFlameCleanse");
	}
	if (a2) {
		v6 = *(uint32_t*)(a2 + 516);
		if (v6) {
			while (2) {
				v7 = getMemAt(0x5D4594, 2487760);
				do {
					if (*(unsigned short*)(v6 + 4) == *(uint32_t*)v7) {
						nox_xxx_netPriMsgToPlayer_4DA2C0(a3, "plyrspel.c:TooManySpells", 0);
						return 0;
					}
					v7 += 4;
				} while ((int)v7 < (int)getMemAt(0x5D4594, 2487800));
				v6 = *(uint32_t*)(v6 + 512);
				if (v6) {
					continue;
				}
				break;
			}
		}
	}
	if (!nox_common_gameFlags_check_40A5C0(2048)) {
		a6 = 4;
	}
	v19 = 48;
	do {
		v8 = a6 - nox_common_randomInt_415FA0(0, 1);
		if (v8 >= 1) {
			v9 = (float*)nox_xxx_newObjectWithTypeInd_4E3450(
				*getMemU32Ptr(0x5D4594, 2487756 + 4 * ((a1 != 10 ? 5 : 0) + v8)));
			if (v9) {
				v10 = nox_common_randomInt_415FA0(0, 255);
				v11 = v9[44];
				*((uint16_t*)v9 + 62) = v10;
				v12 = v11 + *(float*)(a4 + 176) + 4.0;
				v13 = 8 * v10;
				v21 = *getMemFloatPtr(0x587000, 194136 + v13);
				v22 = *getMemFloatPtr(0x587000, 194140 + v13);
				v14 = *(float*)(a4 + 60);
				v15 = v12 * v21 + *(float*)(a4 + 56);
				v20.field_0 = *(float*)(a4 + 56);
				v20.field_4 = v14;
				v20.field_8 = v15;
				v20.field_C = v12 * v22 + *(float*)(a4 + 60);
				if (nox_xxx_mapTraceRay_535250(&v20, 0, 0, 65)) {
					nox_xxx_createAt_4DAA50((int)v9, a4, v20.field_8, v20.field_C);
					*((uint16_t*)v9 + 63) = *((uint16_t*)v9 + 62);
					v9[20] = v21 * 4.0;
					v9[21] = v22 * 4.0;
					*((uint32_t*)v9 + 34) =
						gameFrame() + nox_common_randomInt_415FA0(3 * gameFPS(), 6 * gameFPS());
					v9[39] = *(float*)(a4 + 56);
					v9[40] = *(float*)(a4 + 60);
					*((uint32_t*)v9 + 186) = nox_xxx_updateFlameCleanse_53D510;
					nox_xxx_unitAddToUpdatable_4DA8D0((int)v9);
					v17 = *((uint32_t*)v9 + 2) | 0x40000000;
					v9[28] = 0.0;
					*((uint32_t*)v9 + 2) = v17;
					nox_xxx_netClientPredictLinear_523530((int)v9);
				} else {
					nox_xxx_delayedDeleteObject_4E5CC0((int)v9);
				}
			}
		}
		--v19;
	} while (v19);
	v18 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v18, a4, 0, 0);
	return 1;
}

//----- (0052D8A0) --------------------------------------------------------
int nox_xxx_castMeteorShower_52D8A0(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;        // esi
	int v7;        // edi
	float v8;      // edx
	float v9;      // ecx
	float v10;     // eax
	int v11;       // eax
	int v12;       // eax
	int result;    // eax
	uint32_t* v14; // eax
	int v15;       // edi
	uint32_t* v16; // ebx
	int v17;       // eax
	float v18;     // [esp+0h] [ebp-20h]
	float4 v19;    // [esp+10h] [ebp-10h]

	if (!*getMemU32Ptr(0x5D4594, 2487800)) {
		*getMemU32Ptr(0x5D4594, 2487800) = nox_xxx_getNameId_4E3AA0("MeteorShower");
	}
	v6 = a5;
	v7 = a4;
	v8 = *(float*)(a5 + 4);
	v9 = *(float*)(a4 + 60);
	v19.field_0 = *(float*)(a4 + 56);
	v10 = *(float*)(a5 + 8);
	v19.field_4 = v9;
	v19.field_8 = v8;
	v19.field_C = v10;
	v11 = nox_common_gameFlags_check_40A5C0(2048);
	if (nox_xxx_mapTraceRay_535250(&v19, 0, 0, !v11 ? 73 : 9)) {
		v14 = nox_xxx_newObjectWithTypeInd_4E3450(*getMemIntPtr(0x5D4594, 2487800));
		v15 = (int)v14;
		if (v14) {
			v16 = (uint32_t*)v14[187];
			v18 = nox_xxx_gamedataGetFloatTable_419D70("MeteorDamage", a6 - 1);
			*v16 = nox_float2int(v18);
			nox_xxx_createAt_4DAA50(v15, a3, *(float*)(v6 + 4), *(float*)(v6 + 8));
			v17 = nox_xxx_spellGetAud44_424800(a1, 0);
			nox_xxx_aud_501960(v17, a3, 0, 0);
		}
		result = 1;
	} else {
		if (*(uint8_t*)(v7 + 8) & 4) {
			v12 = *(uint32_t*)(v7 + 748);
			a4 = 2;
			nox_xxx_netInformTextMsg_4DA0F0(*(unsigned char*)(*(uint32_t*)(v12 + 276) + 2064), 0, &a4);
		}
		result = 0;
	}
	return result;
}

//----- (0052D9D0) --------------------------------------------------------
int nox_xxx_castMeteor_52D9D0(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;        // eax
	int v7;        // ecx
	int v8;        // esi
	int v9;        // ebp
	float v10;     // ecx
	float v11;     // edx
	float2* v12;   // edi
	float v13;     // eax
	int v14;       // eax
	int v15;       // eax
	int result;    // eax
	uint32_t* v17; // eax
	int v18;       // esi
	int* v19;      // ebx
	int v20;       // eax
	int v21;       // ecx
	int v22;       // eax
	float v23;     // [esp+0h] [ebp-24h]
	float4 v24;    // [esp+14h] [ebp-10h]

	v6 = dword_5d4594_2487804;
	if (!dword_5d4594_2487804) {
		v6 = nox_xxx_getNameId_4E3AA0("Meteor");
		dword_5d4594_2487804 = v6;
	}
	v7 = *(uint32_t*)(a3 + 516);
	if (v7) {
		while (*(unsigned short*)(v7 + 4) != v6) {
			v7 = *(uint32_t*)(v7 + 512);
			if (!v7) {
				goto LABEL_6;
			}
		}
		nox_xxx_netPriMsgToPlayer_4DA2C0(a3, "ExecSpel.c:TooManyMeteors", 0);
		return 0;
	}
LABEL_6:
	v8 = a4;
	v9 = a5;
	v10 = *(float*)(a4 + 60);
	v11 = *(float*)(a5 + 4);
	v12 = (float2*)(a5 + 4);
	v24.field_0 = *(float*)(a4 + 56);
	v24.field_4 = v10;
	v13 = *(float*)(a5 + 8);
	v24.field_8 = v11;
	v24.field_C = v13;
	v14 = nox_common_gameFlags_check_40A5C0(2048);
	if (nox_xxx_mapTraceRay_535250(&v24, 0, 0, !v14 ? 73 : 9)) {
		v17 = nox_xxx_newObjectWithTypeInd_4E3450(*(int*)&dword_5d4594_2487804);
		v18 = (int)v17;
		if (v17) {
			v19 = (int*)v17[187];
			v23 = nox_xxx_gamedataGetFloatTable_419D70("MeteorDamage", a6 - 1);
			v20 = nox_float2int(v23);
			v21 = a3;
			*v19 = v20;
			nox_xxx_createAt_4DAA50(v18, v21, v12->field_0, *(float*)(v9 + 8));
			*(uint32_t*)(v18 + 20) |= 0x20u;
			nox_xxx_unitRaise_4E46F0(v18, 255.0);
			*(float*)(v18 + 108) = -nox_xxx_gamedataGetFloat_419D40("MeteorSpeed");
			v22 = nox_xxx_spellGetAud44_424800(a1, 0);
			nox_xxx_audCreate_501A30(v22, v12, 0, 0);
		}
		result = 1;
	} else {
		if (*(uint8_t*)(v8 + 8) & 4) {
			v15 = *(uint32_t*)(v8 + 748);
			a3 = 2;
			nox_xxx_netInformTextMsg_4DA0F0(*(unsigned char*)(*(uint32_t*)(v15 + 276) + 2064), 0, &a3);
		}
		result = 0;
	}
	return result;
}

//----- (0052DB60) --------------------------------------------------------
int nox_xxx_castToxicCloud_52DB60(int a1, int a2, int a3, int a4, int a5) {
	int v5;        // esi
	int v6;        // edi
	float v7;      // eax
	float v8;      // edx
	float v9;      // eax
	int v10;       // eax
	int result;    // eax
	uint32_t* v12; // esi
	uint32_t* v13; // ebx
	int v14;       // eax
	float v15;     // [esp+0h] [ebp-28h]
	float4 v16;    // [esp+18h] [ebp-10h]

	if (!*getMemU32Ptr(0x5D4594, 2487808)) {
		*getMemU32Ptr(0x5D4594, 2487808) = nox_xxx_getNameId_4E3AA0("ToxicCloud");
	}
	v5 = a4;
	v6 = a5;
	v7 = *(float*)(a4 + 56);
	v8 = *(float*)(a5 + 4);
	v16.field_4 = *(float*)(a4 + 60);
	v16.field_0 = v7;
	v9 = *(float*)(a5 + 8);
	v16.field_8 = v8;
	v16.field_C = v9;
	if ((unsigned char)nox_xxx_traceRay_5374B0(&v16)) {
		v12 = nox_xxx_newObjectWithTypeInd_4E3450(*getMemIntPtr(0x5D4594, 2487808));
		if (v12) {
			v13 = (uint32_t*)v12[187];
			nox_xxx_createAt_4DAA50((int)v12, a3, *(float*)(v6 + 4), *(float*)(v6 + 8));
			v15 = nox_xxx_gamedataGetFloat_419D40("ToxicCloudLifetime") * (double)(int)gameFPS();
			*v13 = nox_float2int(v15);
		}
		v14 = nox_xxx_spellGetAud44_424800(a1, 0);
		nox_xxx_aud_501960(v14, (int)v12, 0, 0);
		result = 1;
	} else {
		if (*(uint8_t*)(v5 + 8) & 4) {
			v10 = *(uint32_t*)(v5 + 748);
			a4 = 2;
			nox_xxx_netInformTextMsg_4DA0F0(*(unsigned char*)(*(uint32_t*)(v10 + 276) + 2064), 0, &a4);
		}
		result = 0;
	}
	return result;
}

//----- (0052DC80) --------------------------------------------------------
int nox_xxx_spellArachna_52DC80(int a1, int a2, int a3, int a4, int a5) {
	int v5;        // esi
	int v6;        // edi
	float v7;      // eax
	float v8;      // edx
	float v9;      // eax
	int v10;       // eax
	int result;    // eax
	uint32_t* v12; // eax
	float4 v13;    // [esp+0h] [ebp-10h]

	if (!*getMemU32Ptr(0x5D4594, 2487812)) {
		*getMemU32Ptr(0x5D4594, 2487812) = nox_xxx_getNameId_4E3AA0("ArachnaphobiaFocus");
	}
	v5 = a4;
	v6 = a5;
	v7 = *(float*)(a4 + 56);
	v8 = *(float*)(a5 + 4);
	v13.field_4 = *(float*)(a4 + 60);
	v13.field_0 = v7;
	v9 = *(float*)(a5 + 8);
	v13.field_8 = v8;
	v13.field_C = v9;
	if ((unsigned char)nox_xxx_traceRay_5374B0(&v13)) {
		v12 = nox_xxx_newObjectWithTypeInd_4E3450(*getMemIntPtr(0x5D4594, 2487812));
		if (v12) {
			nox_xxx_createAt_4DAA50((int)v12, a3, *(float*)(v6 + 4), *(float*)(v6 + 8));
		}
		result = 1;
	} else {
		if (*(uint8_t*)(v5 + 8) & 4) {
			v10 = *(uint32_t*)(v5 + 748);
			a4 = 2;
			nox_xxx_netInformTextMsg_4DA0F0(*(unsigned char*)(*(uint32_t*)(v10 + 276) + 2064), 0, &a4);
		}
		result = 0;
	}
	return result;
}

//----- (0052DD50) --------------------------------------------------------
int sub_52DD50(int a1, int a2, int a3, int a4, void* a5) {
	void* v5;   // edi
	int v6;     // esi
	short v8;   // bx
	short v9;   // ax
	char v10;   // al
	double v11; // st7
	int v12;    // eax
	int v13;    // eax
	int v14;    // [esp-Ch] [ebp-14h]
	float v15;  // [esp+1Ch] [ebp+14h]

	v5 = a5;
	v6 = *(uint32_t*)a5;
	if (!*(uint32_t*)a5) {
		return 0;
	}
	v8 = nox_xxx_unitGetHP_4EE780(*(uint32_t*)a5);
	if (v8 == nox_xxx_unitGetMaxHP_4EE7A0(v6) && a2 == *(uint32_t*)a5) {
		v9 = nox_xxx_spellManaCost_4249A0(a1, 1);
		sub_4FD030((nox_object_t*)(uintptr_t)(uint32_t)a3, v9);
		return 1;
	}
	v15 = nox_xxx_gamedataGetFloat_419D40("LesserHealAmount");
	if (a3 && *(uint8_t*)(a3 + 8) & 4) {
		v10 = *(uint8_t*)(*(uint32_t*)(*(uint32_t*)(a3 + 748) + 276) + 2251);
		switch (v10) {
		case 0:
			v11 = get_nox_xxx_warriorMaxHealth_587000_312784();
			v15 = v11 * v15;
			break;
		case 2:
			v11 = get_nox_xxx_conjurerMaxHealth_587000_312800();
			v15 = v11 * v15;
			break;
		case 1:
			v11 = get_nox_xxx_wizardMaxHealth_587000_312816();
			v15 = v11 * v15;
			break;
		}
	}
	v12 = nox_float2int(v15);
	nox_xxx_unitAdjustHP_4EE460(*(uint32_t*)v5, v12);
	v14 = *(uint32_t*)v5;
	v13 = nox_xxx_spellGetAud44_424800(a1, 1);
	nox_xxx_aud_501960(v13, v14, 0, 0);
	return 1;
}

//----- (0052DE40) --------------------------------------------------------
#if 0
int nox_xxx_castEquake_52DE40(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;    // eax
	int v7;    // eax
	float v9;  // [esp+0h] [ebp-18h]
	float v10; // [esp+8h] [ebp-10h]

	*getMemU32Ptr(0x5D4594, 2487700) = a6;
	v9 = nox_xxx_gamedataGetFloat_419D40("EarthquakeRange");
	nox_xxx_unitsGetInCircle_517F90((float2*)(a4 + 56), v9, nox_xxx_equakeDamage_52DEC0, a4);
	v6 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v6, a4, 0, 0);
	v10 = nox_xxx_gamedataGetFloatTable_419D70("EarthquakeJiggle", a6 - 1);
	v7 = nox_float2int(v10);
	nox_xxx_earthquakeSend_4D9110((float*)(a4 + 56), v7);
	return 1;
}
#endif

//----- (0052DEC0) --------------------------------------------------------
#if 0
short nox_xxx_equakeDamage_52DEC0(int a1, int a2) {
	int v2;    // edi
	int v3;    // eax
	int v4;    // ebx
	double v5; // st7
	float v7;  // [esp+14h] [ebp+8h]
	float v8;  // [esp+14h] [ebp+8h]

	v2 = a2;
	v3 = nox_xxx_findParentChainPlayer_4EC580(a2);
	v4 = v3;
	if (a1 != a2) {
		v3 = *(uint32_t*)(a1 + 16);
		if (!(v3 & 0x4000)) {
			LOWORD(v3) = nox_xxx_unitGetHP_4EE780(a1);
			if ((uint16_t)v3) {
				v7 = nox_xxx_calcDistance_4E6C00((nox_object_t*)(uintptr_t)a1, (nox_object_t*)(uintptr_t)a2);
				v8 = 1.0 - v7 / nox_xxx_gamedataGetFloat_419D40("EarthquakeRange");
				v5 =
					nox_xxx_gamedataGetFloatTable_419D70("EarthquakeDamage", *getMemU32Ptr(0x5D4594, 2487700) - 1) * v8;
				LOWORD(v3) = (*(int (**)(int, int, int, uint32_t, int))(a1 + 716))(a1, v4, v2, (long long)v5, 11);
			}
		}
	}
	return v3;
}
#endif

//----- (0052E020) --------------------------------------------------------
unsigned int nox_xxx_isObjectMovable_52E020(int a1) {
	unsigned int result; // eax

	if (*(uint32_t*)(a1 + 16) & 0x8068) {
		result = 0;
	} else {
		result = ((unsigned int)~*(uint32_t*)(a1 + 8) >> 22) & 1;
	}
	return result;
}

//----- (0052E040) --------------------------------------------------------
#if UINTPTR_MAX > UINT32_MAX
extern void nox_xxx_mapPushUnitsAround_native_52E040(
	void* origin, float outer_radius, float inner_radius, float force,
	nox_object_t* source, int callback, int callback_arg);
#endif

void nox_xxx_mapPushUnitsAround_52E040(void* a1p, float a2, float a3p, float a4, nox_object_t* a5p, int a6, int a7) {
#if UINTPTR_MAX > UINT32_MAX
	// The original callback record below stores the origin, source and each
	// candidate object in PE32-sized integer slots. Route every wide-host C
	// caller through the native-width implementation instead.
	nox_xxx_mapPushUnitsAround_native_52E040(a1p, a2, a3p, a4, a5p, a6, a7);
#else
	int a1 = a1p;
	int a3 = *(int*)(&a3p);
	int a5 = a5p;
	double v7;  // st7
	double v8;  // st7
	float4 a1a; // [esp+0h] [ebp-2Ch]
	int a3a[7]; // [esp+10h] [ebp-1Ch]

	a3a[0] = a1;
	v7 = *(float*)a1 - a2;
	a3a[5] = a6;
	a3a[6] = a7;
	a1a.field_0 = v7;
	a1a.field_4 = *(float*)(a1 + 4) - a2;
	a1a.field_8 = a2 + *(float*)a1;
	v8 = a2 + *(float*)(a1 + 4);
	a3a[4] = a5;
	a3a[2] = a3;
	a1a.field_C = v8;
	*(float*)&a3a[3] = a4 * 10.0;
	if (a2 >= (double)a3p) {
		*(float*)&a3a[1] = a2;
	} else {
		a3a[1] = a3;
	}
	nox_xxx_getUnitsInRectAdv_517ED0(&a1a, nox_xxx_unitPushAroundFn_52E0E0, (int)a3a);
#endif
}

//----- (0052E0E0) --------------------------------------------------------
void nox_xxx_unitPushAroundFn_52E0E0(int a1, int** a2) {
	int v2;         // esi
	int** v3;       // edi
	int* v4;        // eax
	float v5;       // edx
	float v6;       // eax
	float v7;       // ecx
	double v8;      // st7
	long double v9; // st7
	double v10;     // st7
	int* v11;       // eax
	bool v12;       // zf
	float v13;      // [esp+8h] [ebp-18h]
	float v14;      // [esp+Ch] [ebp-14h]
	float4 v15;     // [esp+10h] [ebp-10h]
	float v16;      // [esp+24h] [ebp+4h]
	float v17;      // [esp+28h] [ebp+8h]
	float v18;      // [esp+28h] [ebp+8h]

	v2 = a1;
	if (nox_xxx_isObjectMovable_52E020(a1)) {
		v3 = a2;
		v4 = *a2;
		LODWORD(v15.field_0) = **a2;
		v5 = *((float*)v4 + 1);
		v6 = *(float*)(a1 + 56);
		v7 = *(float*)(a1 + 60);
		v15.field_4 = v5;
		v15.field_8 = v6;
		v15.field_C = v7;
		if (nox_xxx_mapTraceRay_535250(&v15, 0, 0, 0)) {
			v13 = *(float*)(a1 + 56) - *(float*)*a2;
			v8 = *(float*)(a1 + 60) - *((float*)*a2 + 1);
			v14 = v8;
			v9 = sqrt(v8 * v14 + v13 * v13) + 0.1;
			v16 = v9;
			if (v9 <= *((float*)a2 + 1)) {
				if (v16 > (double)*((float*)a2 + 2)) {
					v17 =
						(1.0 - (v16 - *((float*)a2 + 2)) / (*((float*)a2 + 1) - *((float*)a2 + 2))) * *((float*)a2 + 3);
				} else {
					v17 = *((float*)a2 + 3);
				}
				v10 = nox_xxx_objectGetMass_4E4A70((const nox_object_t*)v2);
				v11 = v3[5];
				v18 = v17 / v10;
				if (v11) {
					((void (*)(int, uint32_t, int*))v11)(v2, LODWORD(v16), v3[6]);
				}
				v12 = (*(uint8_t*)(v2 + 8) & 1) == 0;
				*(float*)(v2 + 88) = v18 * v13 / v16 + *(float*)(v2 + 88);
				*(float*)(v2 + 92) = v18 * v14 / v16 + *(float*)(v2 + 92);
				if (v12) {
					nox_xxx_unitHasCollideOrUpdateFn_537610(v2);
				}
			}
		}
	}
}

//----- (0052E210) --------------------------------------------------------
int nox_xxx_spellDrainMana_52E210(float a1) {
	int v1;            // esi
	int v2;            // eax
	int v3;            // eax
	float v4;          // edx
	int v5;            // eax
	int v7;            // edi
	unsigned short v8; // bx
	int v9;            // eax
	int v10;           // ecx
	double v11;        // st7
	double v12;        // st7
	int v13;           // eax
	int v14;           // eax
	uint32_t* v15;     // ecx
	double v16;        // st7
	uint32_t* v17;     // eax
	int v18;           // eax
	float2 v19;        // [esp+8h] [ebp-8h]
	float v21;         // [esp+14h] [ebp+4h]
	float v22;         // [esp+14h] [ebp+4h]

	v1 = LODWORD(a1);
	v2 = *(uint32_t*)(LODWORD(a1) + 16);
	if (v2) {
		if (nox_xxx_testUnitBuffs_4FF350(v2, 8)) {
			return 1;
		}
	} else if (!*(uint32_t*)(LODWORD(a1) + 20)) {
		return 1;
	}
	if (*(uint32_t*)(LODWORD(a1) + 20)) {
		v3 = *(uint32_t*)(LODWORD(a1) + 16);
		if (v3) {
			v19.field_0 = *(float*)(v3 + 56);
			v4 = *(float*)(v3 + 60);
		} else {
			v4 = *(float*)(LODWORD(a1) + 32);
			v19.field_0 = *(float*)(LODWORD(a1) + 28);
		}
		v19.field_4 = v4;
		v5 = sub_52E610((int*)&v19, v3);
		if (v5) {
			nox_xxx_playerManaSub_4EEBF0(v5, 50);
			return 1;
		}
		return 1;
	}
	v7 = *(uint32_t*)(LODWORD(a1) + 16);
	if (*(uint8_t*)(v7 + 8) & 4) {
		v8 = nox_xxx_unitGetOldMana_4EEC80(v7);
		if (v8 >= (unsigned short)nox_xxx_playerGetMaxMana_4EECB0(v7)) {
			return 1;
		}
	}
	v9 = *(uint32_t*)(LODWORD(a1) + 16);
	if (*(uint8_t*)(v9 + 8) & 2) {
		if (sub_4FEA70((nox_object_t*)(uintptr_t)(uint32_t)v9, (float2*)(LODWORD(a1) + 28))) {
			return 1;
		}
	}
	v10 = *(uint32_t*)(LODWORD(a1) + 16);
	v11 = *(float*)(LODWORD(a1) + 28) - *(float*)(v10 + 56);
	if (v11 < 0.0) {
		v11 = -v11;
	}
	v19.field_0 = v11;
	v12 = *(float*)(LODWORD(a1) + 32) - *(float*)(v10 + 60);
	if (v12 < 0.0) {
		v12 = -v12;
	}
	v19.field_4 = v12;
	if (sub_4E6BD0((nox_object_t*)(uintptr_t)v10) || v19.field_0 >= 5.0 || v19.field_4 >= 5.0) {
		return 1;
	}
	v13 = sub_52E610((int*)(*(uint32_t*)(LODWORD(a1) + 16) + 56), *(uint32_t*)(LODWORD(a1) + 16));
	*(uint32_t*)(LODWORD(a1) + 48) = v13;
	if (!v13) {
		if (*(uint32_t*)(LODWORD(a1) + 36)) {
			nox_xxx_netStopRaySpell_4FEF90(SLODWORD(a1), *(uint32_t**)(LODWORD(a1) + 36));
		}
		return 1;
	}
	v21 = *(float*)(LODWORD(a1) + 72);
	v22 = nox_xxx_gamedataGetFloatTable_419D70("ManaDrainCoeff", *(uint32_t*)(v1 + 8) - 1) + v21;
	*(float*)&v14 = COERCE_FLOAT(nox_float2int(v22));
	v15 = *(uint32_t**)(v1 + 48);
	v19.field_0 = *(float*)&v14;
	v16 = (double)v14;
	v17 = *(uint32_t**)(v1 + 36);
	*(float*)(v1 + 72) = v22 - v16;
	if (v15 != v17) {
		if (v17) {
			nox_xxx_netStopRaySpell_4FEF90(v1, v17);
		}
		nox_xxx_netStartDurationRaySpell_4FF130(v1);
	}
	v18 = nox_float2int(v22);
	if (sub_52E450(*(uint32_t*)(v1 + 16), *(uint32_t*)(v1 + 48), v18) &&
		!(gameFrame() % (gameFPS() >> 1))) {
		nox_xxx_aud_501960(230, *(uint32_t*)(v1 + 16), 0, 0);
		nox_xxx_aud_501960(229, *(uint32_t*)(v1 + 48), 0, 0);
	}
	*(uint32_t*)(v1 + 36) = *(uint32_t*)(v1 + 48);
	return 0;
}

//----- (0052E450) --------------------------------------------------------
int sub_52E450(int a1, int a2, int a3) {
	int v3;    // esi
	int v4;    // eax
	int* v5;   // ebp
	int v6;    // esi
	int v7;    // eax
	char v8;   // al
	double v9; // st7
	float v11; // [esp+0h] [ebp-14h]

	if (*(uint8_t*)(a1 + 8) & 4) {
		v3 = *(uint32_t*)(a1 + 748);
		if ((unsigned short)nox_xxx_unitGetOldMana_4EEC80(a1) >= *(uint16_t*)(v3 + 8)) {
			return 0;
		}
	}
	v4 = *(uint32_t*)(a2 + 8);
	if (v4 & 0x400000 && *(uint8_t*)(a2 + 12) & 0x18) {
		v5 = *(int**)(a2 + 748);
		if (nox_xxx_servObjectHasTeam_419130(a2 + 48) && !nox_xxx_servCompareTeams_419150(a2 + 48, a1 + 48)) {
			return 0;
		}
		v6 = a3;
		if (*v5 <= a3) {
			v6 = *v5;
			a3 = *v5;
			if (!nox_common_gameFlags_check_40A5C0(4096)) {
				*v5 = 0;
			}
		} else if (!nox_common_gameFlags_check_40A5C0(4096)) {
			*v5 -= a3;
		}
		if (!v6) {
			return 0;
		}
		if (!nox_common_gameFlags_check_40A5C0(4096)) {
			nox_xxx_unitNeedSync_4E44F0(a2);
		}
		if (v6 > 0) {
			goto LABEL_24;
		}
		return 0;
	}
	if (v4 & 2) {
		if (*(uint8_t*)(*(uint32_t*)(a2 + 748) + 1440) & 0x20) {
			LOWORD(v6) = 1;
			a3 = 1;
			goto LABEL_24;
		}
		v6 = a3;
		if (v6 > 0) {
			goto LABEL_24;
		}
		return 0;
	}
	if (!(v4 & 4)) {
		v6 = a3;
		if (v6 > 0) {
			goto LABEL_24;
		}
		return 0;
	}
	v6 = a3;
	v7 = *(unsigned short*)(*(uint32_t*)(a2 + 748) + 4);
	if ((unsigned short)v7 <= a3) {
		v6 = *(unsigned short*)(*(uint32_t*)(a2 + 748) + 4);
		a3 = *(unsigned short*)(*(uint32_t*)(a2 + 748) + 4);
		nox_xxx_playerManaSub_4EEBF0(a2, v7);
	} else {
		nox_xxx_playerManaSub_4EEBF0(a2, a3);
	}
	if (v6 <= 0) {
		return 0;
	}
LABEL_24:
	if (!nox_common_gameFlags_check_40A5C0(4096) || !(*(uint8_t*)(a1 + 8) & 4)) {
		nox_xxx_playerManaAdd_4EEB80(a1, v6);
		return 1;
	}
	v8 = *(uint8_t*)(*(uint32_t*)(*(uint32_t*)(a1 + 748) + 276) + 2251);
	if (v8) {
		if (v8 == 1) {
			v9 = (double)a3 * get_nox_xxx_wizardMaximumMana_587000_312820();
		} else {
			if (v8 != 2) {
				nox_xxx_playerManaAdd_4EEB80(a1, v6);
				return 1;
			}
			v9 = (double)a3 * get_nox_xxx_conjurerMaxMana_587000_312804();
		}
	} else {
		v9 = (double)a3 * get_nox_xxx_warriorMaxMana_587000_312788();
	}
	v11 = v9;
	LOWORD(v6) = nox_float2int(v11);
	nox_xxx_playerManaAdd_4EEB80(a1, v6);
	return 1;
}

//----- (0052E610) --------------------------------------------------------
int sub_52E610(int* a1, int a2) {
	float v3; // [esp+0h] [ebp-10h]

	*getMemU32Ptr(0x5D4594, 2487828) = 0;
	*getMemU32Ptr(0x5D4594, 2487876) = 1287568416;
	*getMemU32Ptr(0x5D4594, 2487836) = *a1;
	*getMemU32Ptr(0x5D4594, 2487840) = a1[1];
	v3 = nox_xxx_gamedataGetFloat_419D40("ManaDrainRange");
	nox_xxx_unitsGetInCircle_517F90((float2*)a1, v3, sub_52E660, a2);
	return *getMemU32Ptr(0x5D4594, 2487828);
}

//----- (0052E660) --------------------------------------------------------
void sub_52E660(int a1, int a2) {
	int v2;     // eax
	int v3;     // ebx
	double v4;  // st7
	int v5;     // ebx
	double v6;  // st6
	double v7;  // st5
	double v8;  // st5
	float v9;   // eax
	float v10;  // ecx
	float4 v11; // [esp+Ch] [ebp-10h]
	float v12;  // [esp+20h] [ebp+4h]

	if (a1 == a2 || !sub_52E7C0(a1) || *(uint32_t*)(a1 + 16) & 0x8020) {
		return;
	}
	v2 = *(uint32_t*)(a1 + 8);
	if (!(v2 & 2)) {
		if (v2 & 0x400000 && *(uint8_t*)(a1 + 12) & 0x18) {
			v4 = 1.0;
		} else {
			if (!(v2 & 4)) {
				return;
			}
			v5 = *(uint32_t*)(a1 + 748);
			if (a2) {
				if (!nox_xxx_unitIsEnemyTo_5330C0(a2, a1)) {
					return;
				}
			}
			if (nox_xxx_unitsHaveSameTeam_4EC520(a1, a2) || *(uint8_t*)(*(uint32_t*)(v5 + 276) + 3680) & 1) {
				return;
			}
			v4 = 0.5;
		}
		goto LABEL_18;
	}
	v3 = *(uint32_t*)(a1 + 748);
	if (!((!a2 || nox_xxx_unitIsEnemyTo_5330C0(a2, a1)) && *(uint8_t*)(v3 + 1440) & 0x20)) {
		return;
	}
	v4 = 1.0;
LABEL_18:
	v6 = *getMemFloatPtr(0x5D4594, 2487836) - *(float*)(a1 + 56);
	v7 = *getMemFloatPtr(0x5D4594, 2487840) - *(float*)(a1 + 60);
	v8 = v4 * (v7 * v7 + v6 * v6);
	if (v8 < *getMemFloatPtr(0x5D4594, 2487876)) {
		v9 = *(float*)(a1 + 56);
		v11.field_4 = *getMemFloatPtr(0x5D4594, 2487840);
		v11.field_0 = *getMemFloatPtr(0x5D4594, 2487836);
		v10 = *(float*)(a1 + 60);
		v11.field_8 = v9;
		v11.field_C = v10;
		if (nox_xxx_mapTraceRay_535250(&v11, 0, 0, 5)) {
			*getMemU32Ptr(0x5D4594, 2487828) = a1;
			v12 = v8;
			*getMemFloatPtr(0x5D4594, 2487876) = v12;
		}
	}
}

//----- (0052E7C0) --------------------------------------------------------
int sub_52E7C0(int a1) {
	int v1; // eax

	v1 = *(uint32_t*)(a1 + 8);
	if (v1 & 0x400000 && *(uint8_t*)(a1 + 12) & 0x18) {
		if (**(uint32_t**)(a1 + 748) > 0) {
			return 1;
		}
	} else if (v1 & 2) {
		if (*(uint8_t*)(*(uint32_t*)(a1 + 748) + 1440) & 0x20) {
			return 1;
		}
	} else if (v1 & 4 && nox_xxx_unitGetOldMana_4EEC80(a1)) {
		return 1;
	}
	return 0;
}

//----- (0052E820) --------------------------------------------------------
int nox_xxx_spellEnergyBoltStop_52E820(int a1) {
	if (*(uint32_t*)(a1 + 16)) {
		nox_xxx_spellCancelDurSpell_4FEB10(43, *(uint32_t*)(a1 + 16));
	}
	nox_xxx_netSendPointFx_522FF0(130, (float2*)(a1 + 28));
	return 0;
}

//----- (0052E850) --------------------------------------------------------
int nox_xxx_spellEnergyBoltTick_52E850(float a1) {
	int v1;                                              // esi
	int v2;                                              // eax
	int result;                                          // eax
	int v5;                                              // eax
	void (**v6)(uint32_t, uint32_t, uint32_t, int, int); // edi
	int v7;                                              // eax
	int v8;                                              // eax
	int v9;                                              // eax
	int v10;                                             // eax
	int v11;                                             // eax
	int v12;                                             // ecx
	int v13;                                             // edi
	int v14;                                             // eax
	float v16;                                           // edi
	int v17;                                             // eax
	uint32_t* v18;                                       // ecx
	double v19;                                          // st7
	uint32_t* v20;                                       // eax
	int v21;                                             // eax
	int v22;                                             // eax
	int v23;                                             // eax
	int v24;                                             // eax
	int v25;                                             // ecx
	int v26;                                             // [esp-4h] [ebp-20h]
	float v27;                                           // [esp+0h] [ebp-1Ch]
	float v28;                                           // [esp+4h] [ebp-18h]
	float v31;                                           // [esp+20h] [ebp+4h]
	float v32;                                           // [esp+20h] [ebp+4h]
	float v33;                                           // [esp+20h] [ebp+4h]
	int v34;                                             // [esp+20h] [ebp+4h]

	v1 = LODWORD(a1);
	v2 = *(uint32_t*)(LODWORD(a1) + 16);
	if (v2) {
		if (nox_xxx_testUnitBuffs_4FF350(v2, 8)) {
			return 1;
		}
	} else if (!*(uint32_t*)(LODWORD(a1) + 20)) {
		return 1;
	}
	v31 = nox_xxx_gamedataGetFloat_419D40("LightningRange");
	if (!*(uint32_t*)(v1 + 20)) {
		v9 = *(uint32_t*)(v1 + 16);
		if (v9 && *(uint8_t*)(v9 + 8) & 2 &&
			sub_4FEA70((nox_object_t*)(uintptr_t)(uint32_t)v9, (float2*)(v1 + 28))) {
			return 1;
		}
		if ((unsigned int)(gameFrame() - *(uint32_t*)(v1 + 60)) > 2 &&
			sub_4E6BD0((nox_object_t*)(uintptr_t)*(uint32_t*)(v1 + 16))) {
			return 1;
		}
		v10 = *(uint32_t*)(v1 + 48);
		if (v10) {
			if (!(*(uint32_t*)(v10 + 16) & 0x8020) &&
				nox_server_testTwoPointsAndDirection_4E6E50((float2*)(*(uint32_t*)(v1 + 16) + 56),
															*(short*)(*(uint32_t*)(v1 + 16) + 124),
															(float2*)(v10 + 56)) &
					1 &&
				nox_xxx_calcDistance_4E6C00((nox_object_t*)(uintptr_t)*(uint32_t*)(v1 + 48),
										(nox_object_t*)(uintptr_t)*(uint32_t*)(v1 + 16)) <= v31 &&
				nox_xxx_unitCanInteractWith_5370E0(*(uint32_t*)(v1 + 16), *(uint32_t*)(v1 + 48), 0)) {
				goto LABEL_31;
			}
			*(uint32_t*)(v1 + 48) = 0;
		}
		v11 = *(uint32_t*)(v1 + 16);
		if (*(uint8_t*)(v11 + 8) & 4) {
			v12 = *(uint32_t*)(v11 + 748);
			v13 = *(uint32_t*)(v12 + 288);
			if (v13) {
				if (nox_xxx_unitIsEnemyTo_5330C0(v11, *(uint32_t*)(v12 + 288)) &&
					nox_xxx_calcDistance_4E6C00((nox_object_t*)(uintptr_t)*(uint32_t*)(v1 + 16),
										 (nox_object_t*)(uintptr_t)v13) <= v31) {
					*(uint32_t*)(v1 + 48) = v13;
				}
			}
		}
		if (*(uint32_t*)(v1 + 48)) {
			goto LABEL_32;
		}
		*getMemU32Ptr(0x5D4594, 2487832) = 0;
		nox_xxx_energyBoltTarget_5d4594_2487880 = 0;
		v14 = *(uint32_t*)(v1 + 16);
		*getMemFloatPtr(0x5D4594, 2487868) = *(float*)(v14 + 56);
		*getMemFloatPtr(0x5D4594, 2487872) = *(float*)(v14 + 60);
		*(float*)&dword_5d4594_2487884 = v31 * v31;
		nox_xxx_unitsGetInCircle_517F90((float2*)(*(uint32_t*)(v1 + 16) + 56), v31,
										nox_xxx_spellEnergyBoltSetTarget_52EC60, *(uint32_t*)(v1 + 16));
		*(uint32_t*)(v1 + 48) = nox_xxx_energyBoltTarget_5d4594_2487880;
	LABEL_31:
		if (!*(uint32_t*)(v1 + 48)) {
			if (*(uint32_t*)(v1 + 36)) {
				nox_xxx_netStopRaySpell_4FEF90(v1, *(uint32_t**)(v1 + 36));
				*(uint32_t*)(v1 + 36) = 0;
			}
			return 0;
		}
	LABEL_32:
		v32 = *(float*)(v1 + 72);
		v33 = nox_xxx_gamedataGetFloatTable_419D70("EnergyBoltDamage", *(uint32_t*)(v1 + 8) - 1) + v32;
		v16 = v33;
		v17 = nox_float2int(v33);
		v18 = *(uint32_t**)(v1 + 48);
		v19 = (double)v17;
		v20 = *(uint32_t**)(v1 + 36);
		*(float*)(v1 + 72) = v33 - v19;
		if (v18 != v20) {
			if (v20) {
				nox_xxx_netStopRaySpell_4FEF90(v1, v20);
			}
			nox_xxx_netStartDurationRaySpell_4FF130(v1);
		}
		v34 = *(uint32_t*)(v1 + 48);
		v21 = nox_float2int(v16);
		(*(void (**)(uint32_t, uint32_t, uint32_t, int, int))(v34 + 716))(*(uint32_t*)(v1 + 48), *(uint32_t*)(v1 + 16),
																		  0, v21, 17);
		v22 = *(uint32_t*)(v1 + 48);
		if (*(uint32_t*)(v22 + 16) & 0x8020) {
			nox_xxx_netSendPointFx_522FF0(130, (float2*)(v22 + 56));
		}
		v23 = *(uint32_t*)(v1 + 16);
		*(uint32_t*)(v1 + 36) = *(uint32_t*)(v1 + 48);
		if (*(uint8_t*)(v23 + 8) & 4) {
			nox_xxx_playerSetState_4FA020((uint32_t*)v23, 10);
		}
		if (!(gameFrame() % (gameFPS() / 3u))) {
			nox_xxx_aud_501960(32, *(uint32_t*)(v1 + 16), 0, 0);
			nox_xxx_aud_501960(32, *(uint32_t*)(v1 + 48), 0, 0);
		}
		v28 = nox_xxx_gamedataGetFloat_419D40("LightningSearchTime");
		*(uint32_t*)(v1 + 68) = gameFrame() + nox_float2int(v28);
		v24 = *(uint32_t*)(v1 + 16);
		if (*(uint8_t*)(v24 + 8) & 4) {
			nox_xxx_playerSetState_4FA020((uint32_t*)v24, 10);
			nox_xxx_playerManaSub_4EEBF0(*(uint32_t*)(v1 + 16), 1);
			if (!nox_xxx_unitGetOldMana_4EEC80(*(uint32_t*)(v1 + 16))) {
				return 1;
			}
		}
		v25 = *(uint32_t*)(*(uint32_t*)(v1 + 48) + 16);
		if ((v25 & 0x8000) != 0) {
			result = 0;
			*(uint32_t*)(v1 + 68) = gameFrame() + 1;
			return result;
		}
		return 0;
	}
	float2 v29;
	v29.field_0 = *(float*)(v1 + 28);
	v29.field_4 = *(float*)(v1 + 32);
	*getMemFloatPtr(0x5D4594, 2487868) = v29.field_0;
	*getMemFloatPtr(0x5D4594, 2487872) = v29.field_4;
	nox_xxx_energyBoltTarget_5d4594_2487880 = 0;
	*(float*)&dword_5d4594_2487884 = v31 * v31;
	*getMemU32Ptr(0x5D4594, 2487832) = 1;
	v5 = *(uint32_t*)(v1 + 16);
	nox_xxx_unitsGetInCircle_517F90(&v29, v31, nox_xxx_spellEnergyBoltSetTarget_52EC60, v5);
	if (nox_xxx_energyBoltTarget_5d4594_2487880) {
		v6 = (void (**)(uint32_t, uint32_t, uint32_t, int, int))(nox_xxx_energyBoltTarget_5d4594_2487880 + 716);
		v27 = nox_xxx_gamedataGetFloat_419D40("EnergyBoltGlyphDamage");
		v7 = nox_float2int(v27);
		(*v6)(nox_xxx_energyBoltTarget_5d4594_2487880, *(uint32_t*)(v1 + 12), 0, v7, 17);
		v26 = nox_xxx_energyBoltTarget_5d4594_2487880;
		v8 = nox_xxx_spellGetAud44_424800(24, 0);
		nox_xxx_aud_501960(v8, v26, 0, 0);
		nox_xxx_netSendPointFx_522FF0(130, (float2*)(nox_xxx_energyBoltTarget_5d4594_2487880 + 56));
	}
	return 1;
}

//----- (0052EC60) --------------------------------------------------------
void nox_xxx_spellEnergyBoltSetTarget_52EC60(int target, int source) {
	int v2;    // eax
	int v3;    // eax
	double v4; // st7
	double v5; // st6
	double v6; // st5

	v2 = *(uint32_t*)(target + 8);
	if (v2 & 0x20006) {
		if (!(*(uint32_t*)(target + 16) & 0x8020) && target != source) {
			if (!(v2 & 2) || (v3 = *(uint32_t*)(target + 12), (v3 & 0x8000) == 0)) {
				if (!source || nox_xxx_unitIsEnemyTo_5330C0(source, target) &&
								   (*getMemU32Ptr(0x5D4594, 2487832) ||
									nox_server_testTwoPointsAndDirection_4E6E50(
										(float2*)(source + 56), *(short*)(source + 124), (float2*)(target + 56)) &
											1 &&
										nox_xxx_unitCanInteractWith_5370E0(source, target, 0))) {
					v4 = *(float*)(target + 56) - *getMemFloatPtr(0x5D4594, 2487868);
					v5 = *(float*)(target + 60) - *getMemFloatPtr(0x5D4594, 2487872);
					v6 = v5 * v5 + v4 * v4;
					if (v6 < *(float*)&dword_5d4594_2487884) {
						*(float*)&dword_5d4594_2487884 = v6;
						nox_xxx_energyBoltTarget_5d4594_2487880 = target;
					}
				}
			}
		}
	}
}

//----- (0052ED40) --------------------------------------------------------
int nox_xxx_firewalkTick_52ED40(float* a1) {
	float* v1;      // edi
	int v2;         // eax
	int result;     // eax
	int v4;         // edx
	float* v5;      // eax
	int v6;         // ecx
	double v7;      // st7
	double v8;      // st6
	long double v9; // st5
	int v10;        // ebx
	float v11;      // edx
	int v12;        // ebp
	int v13;        // eax
	uint32_t* v14;  // eax
	uint32_t* v15;  // esi
	int v16;        // ecx
	float v17;      // [esp+4h] [ebp-Ch]
	float2 v18;     // [esp+8h] [ebp-8h]
	float v19;      // [esp+14h] [ebp+4h]

	v1 = a1;
	v2 = *((uint32_t*)a1 + 12);
	if (!v2) {
		return 1;
	}
	if (*(uint32_t*)(v2 + 16) & 0x8020) {
		return 1;
	}
	if (!*getMemU32Ptr(0x5D4594, 2487888)) {
		*getMemU32Ptr(0x5D4594, 2487888) = nox_xxx_getNameId_4E3AA0("SmallFlame");
		*getMemU32Ptr(0x5D4594, 2487892) = nox_xxx_getNameId_4E3AA0("MediumFlame");
		*getMemU32Ptr(0x5D4594, 2487896) = nox_xxx_getNameId_4E3AA0("Flame");
	}
	if (*((uint32_t*)a1 + 15) == *((uint32_t*)a1 + 16)) {
		v4 = *(uint32_t*)(*((uint32_t*)a1 + 12) + 56);
		*((uint32_t*)a1 + 18) = v4;
		a1[19] = *(float*)(*((uint32_t*)a1 + 12) + 60);
		*((uint32_t*)a1 + 20) = v4;
		a1[21] = a1[19];
		++*((uint32_t*)a1 + 16);
		result = 0;
	} else {
		v5 = (float*)*((uint32_t*)a1 + 12);
		v6 = *((uint32_t*)a1 + 2);
		v7 = v5[14] - a1[18];
		v8 = v5[15] - a1[19];
		v9 = sqrt(v8 * v8 + v7 * v7);
		if (v6 >= 2) {
			v10 = (v6 >= 4) + 1;
		} else {
			v10 = 0;
		}
		if (v9 - v5[44] > 15.0) {
			v11 = a1[19];
			v18.field_0 = a1[18];
			v18.field_4 = v11;
			v12 = 2;
			do {
				v13 = nox_common_randomInt_415FA0(0, v10);
				v14 = nox_xxx_newObjectWithTypeInd_4E3450(*getMemU32Ptr(0x5D4594, 2487888 + 4 * v13));
				v15 = v14;
				if (v14) {
					nox_xxx_createAt_4DAA50((int)v14, 0, v18.field_0, v18.field_4);
					nox_xxx_audCreate_501A30(46, &v18, 0, 0);
					nox_xxx_unitSetDecayTime_511660(v15, 25 * gameFPS());
				}
				v19 = v18.field_0 - v1[20];
				v17 = v18.field_4 - v1[21];
				if (v19 != 0.0 && v17 != 0.0) {
					v18.field_0 = v18.field_0 - v19 * 0.5;
					v18.field_4 = v18.field_4 - v17 * 0.5;
				}
				--v12;
			} while (v12);
			v1[20] = v1[18];
			v1[21] = v1[19];
			v16 = *((uint32_t*)v1 + 12);
			v1[18] = *(float*)(v16 + 56);
			v1[19] = *(float*)(v16 + 60);
		}
		result = 0;
	}
	return result;
}

//----- (0052EF30) --------------------------------------------------------
int sub_52EF30(int a1) {
	int v1;       // eax
	int v2;       // eax
	int v3;       // eax
	float v4;     // ebx
	float v5;     // ebp
	int v6;       // eax
	uint32_t* v7; // eax
	uint32_t* v8; // edi

	if (*(uint32_t*)(a1 + 20)) {
		*(uint16_t*)(a1 + 72) = *(uint16_t*)(*(uint32_t*)(a1 + 24) + 124);
		v3 = *(uint32_t*)(a1 + 24);
		v4 = *(float*)(v3 + 56);
		v5 = *(float*)(v3 + 60);
	} else {
		v1 = *(uint32_t*)(a1 + 16);
		if (*(uint8_t*)(v1 + 8) & 4) {
			v2 = *(uint32_t*)(*(uint32_t*)(v1 + 748) + 104);
			if (v2) {
				if (*(uint32_t*)(v2 + 12) & 0x200000 && *(uint8_t*)(*(uint32_t*)(v2 + 736) + 96) & 4) {
					*(uint8_t*)(a1 + 88) |= 2u;
				}
			}
		}
		v6 = *(uint32_t*)(a1 + 16);
		v4 = *(float*)(v6 + 56);
		v5 = *(float*)(v6 + 60);
		*(uint16_t*)(a1 + 72) = *(uint16_t*)(v6 + 124);
	}
	v7 = nox_xxx_newObjectByTypeID_4E3810("ForceOfNatureCharge");
	v8 = v7;
	if (v7) {
		nox_xxx_createAt_4DAA50((int)v7, 0, v4, v5);
		*(uint32_t*)(a1 + 76) = v8;
	}
	return 0;
}

//----- (0052EFD0) --------------------------------------------------------
int sub_52EFD0(int a1) {
	int v1;     // esi
	int v2;     // eax
	int result; // eax
	float* v4;  // edi
	float v5;   // edx
	float v6;   // ebp
	short v7;   // bx
	double v8;  // st7
	int v9;     // ecx
	int v10;    // eax
	int v11;    // esi
	float v12;  // [esp+4h] [ebp-1Ch]
	float v13;  // [esp+8h] [ebp-18h]
	float v14;  // [esp+Ch] [ebp-14h]
	float4 v15; // [esp+10h] [ebp-10h]
	float v16;  // [esp+24h] [ebp+4h]

	v1 = a1;
	v2 = *(uint32_t*)(a1 + 16);
	if (v2 && nox_xxx_testUnitBuffs_4FF350(v2, 8)) {
		return 1;
	}
	if (*(uint32_t*)(a1 + 68) - 7 == gameFrame() && *(uint32_t*)(a1 + 76)) {
		nox_xxx_delayedDeleteObject_4E5CC0(*(uint32_t*)(a1 + 76));
		*(uint32_t*)(a1 + 76) = 0;
	}
	if (*(uint32_t*)(a1 + 68) - 1 == gameFrame()) {
		v4 = (float*)nox_xxx_newObjectByTypeID_4E3810("DeathBall");
		if (v4) {
			if (*(uint32_t*)(a1 + 20)) {
				v5 = *(float*)(a1 + 28);
				v6 = *(float*)(a1 + 32);
				v7 = *(uint16_t*)(a1 + 72);
				v13 = *(float*)(a1 + 28);
				v8 = 0.0;
				v14 = *(float*)(a1 + 32);
			} else {
				v9 = *(uint32_t*)(a1 + 16);
				v5 = *(float*)(v9 + 56);
				v6 = *(float*)(v9 + 60);
				v7 = *(uint16_t*)(v9 + 124);
				v13 = *(float*)(v9 + 56);
				v14 = *(float*)(v9 + 60);
				switch (*(uint32_t*)(v9 + 172)) {
				case 1:
					v8 = 4.0;
					break;
				case 2:
					v8 = *(float*)(v9 + 176) + 4.0;
					break;
				case 3:
					if (*(float*)(v9 + 184) <= (double)*(float*)(v9 + 188)) {
						v8 = *(float*)(v9 + 188) + 4.0;
					} else {
						v8 = *(float*)(v9 + 184) + 4.0;
					}
					break;
				default:
					v8 = 24.0;
					break;
				}
			}
			v10 = 8 * v7;
			v15.field_4 = v6;
			v16 = *getMemFloatPtr(0x587000, 194136 + v10);
			v12 = *getMemFloatPtr(0x587000, 194140 + v10);
			v15.field_0 = v5;
			v15.field_8 = v16 * v8 + v13;
			v15.field_C = v12 * v8 + v14;
			if (!nox_xxx_mapTraceRay_535250(&v15, 0, 0, 5)) {
				v15.field_8 = v15.field_0;
				v15.field_C = v15.field_4;
			}
			nox_xxx_createAt_4DAA50((int)v4, *(uint32_t*)(v1 + 16), v15.field_8, v15.field_C);
			if (*(uint32_t*)(v1 + 20) == 1) {
				v4[20] = 0.0;
				v4[21] = 0.0;
			} else {
				v4[20] = v16 * v4[136];
				v4[21] = v12 * v4[136];
			}
			*((uint16_t*)v4 + 62) = v7;
			*((uint16_t*)v4 + 63) = v7;
			nox_xxx_aud_501960(38, *(uint32_t*)(v1 + 16), 0, 0);
		}
		result = 1;
	} else {
		if (!*(uint32_t*)(a1 + 20)) {
			v11 = *(uint32_t*)(a1 + 16);
			if (v11) {
				if (*(uint8_t*)(v11 + 8) & 4) {
					nox_xxx_playerSetState_4FA020((uint32_t*)v11, 10);
				}
			}
		}
		result = 0;
	}
	return result;
}

//----- (0052F1D0) --------------------------------------------------------
int sub_52F1D0(int a1) {
	int result; // eax

	if (*(uint32_t*)(a1 + 76)) {
		nox_xxx_delayedDeleteObject_4E5CC0(*(uint32_t*)(a1 + 76));
	}
	result = *(uint32_t*)(a1 + 16);
	if (result) {
		if (*(uint8_t*)(result + 8) & 4) {
			result = *(uint32_t*)(*(uint32_t*)(result + 748) + 104);
			if (result) {
				if (*(uint32_t*)(result + 12) & 0x200000) {
					result = *(uint32_t*)(result + 736);
					*(uint32_t*)(result + 96) &= 0xFFFFFFFB;
				}
			}
		}
	}
	return result;
}

//----- (0052F220) --------------------------------------------------------
int sub_52F220(int* a1) {
	int v1;     // ecx
	int result; // eax
	int v3;     // eax
	int v4;     // eax
	int v5;     // eax
	int v6;     // eax
	int v7;     // [esp-4h] [ebp-8h]

	v1 = a1[5];
	if (!a1[4]) {
		if (!v1) {
			return 1;
		}
		v3 = nox_xxx_spellFlags_424A70(a1[1]);
		v4 = nox_xxx_spellFlySearchTarget_540610((float2*)(a1 + 13), 0, v3, 400.0, 1, 0);
		if (v4) {
			nox_xxx_unitAdjustHP_4EE460(v4, 20);
		}
		return 1;
	}
	if (v1) {
		v3 = nox_xxx_spellFlags_424A70(a1[1]);
		v4 = nox_xxx_spellFlySearchTarget_540610((float2*)(a1 + 13), 0, v3, 400.0, 1, 0);
		if (v4) {
			nox_xxx_unitAdjustHP_4EE460(v4, 20);
		}
		return 1;
	}
	v7 = a1[4];
	v5 = nox_xxx_spellFlags_424A70(a1[1]);
	v6 = nox_xxx_spellFlySearchTarget_540610((float2*)(a1 + 13), a1[4], v5, 400.0, 1, v7);
	a1[12] = v6;
	if (v6) {
		result = nox_xxx_unitIsEnemyTo_5330C0(a1[4], v6);
		if (!result) {
			nox_xxx_netStartDurationRaySpell_4FF130(a1);
			result = 0;
		}
	} else {
		nox_xxx_netPriMsgToPlayer_4DA2C0(a1[4], "ExecDur.c:GreaterHealNoTarget", 0);
		result = 1;
	}
	return result;
}

//----- (0052F2E0) --------------------------------------------------------
int sub_52F2E0(float a1) {
	float v1;  // esi
	int v2;    // eax
	int v4;    // eax
	int v5;    // eax
	short v6;  // di
	int v7;    // eax
	char v8;   // al
	double v9; // st7
	int v10;   // eax
	float v11; // [esp+10h] [ebp+4h]

	v1 = a1;
	v2 = *(uint32_t*)(LODWORD(a1) + 48);
	if (!v2) {
		return 1;
	}
	if (*(uint32_t*)(v2 + 16) & 0x8020) {
		return 1;
	}
	v4 = *(uint32_t*)(LODWORD(a1) + 16);
	if (v4 && nox_xxx_testUnitBuffs_4FF350(v4, 8)) {
		return 1;
	}
	if (!nox_xxx_unitCanInteractWith_5370E0(*(uint32_t*)(LODWORD(a1) + 16), *(uint32_t*)(LODWORD(a1) + 48), 0)) {
		return 1;
	}
	if (!nox_xxx_unitGetOldMana_4EEC80(*(uint32_t*)(LODWORD(a1) + 16))) {
		return 1;
	}
	v5 = *(uint32_t*)(LODWORD(a1) + 16);
	if (*(uint8_t*)(v5 + 8) & 2 &&
		sub_4FEA70((nox_object_t*)(uintptr_t)(uint32_t)v5, (float2*)(LODWORD(a1) + 28))) {
		return 1;
	}
	if (sub_4E6BD0((nox_object_t*)(uintptr_t)*(uint32_t*)(LODWORD(a1) + 16))) {
		return 1;
	}
	v6 = nox_xxx_unitGetMaxHP_4EE7A0(*(uint32_t*)(LODWORD(a1) + 48));
	if (v6 == nox_xxx_unitGetHP_4EE780(*(uint32_t*)(LODWORD(a1) + 48))) {
		return 1;
	}
	v7 = *(uint32_t*)(LODWORD(a1) + 16);
	v11 = *(float*)(LODWORD(a1) + 72) + *getMemFloatPtr(0x587000, 260360 + 4 * *(uint32_t*)(LODWORD(a1) + 8));
	if (v7 && *(uint8_t*)(v7 + 8) & 4) {
		v8 = *(uint8_t*)(*(uint32_t*)(*(uint32_t*)(v7 + 748) + 276) + 2251);
		switch (v8) {
		case 0:
			v9 = get_nox_xxx_warriorMaxHealth_587000_312784();
			v11 = v9 * v11;
			break;
		case 2:
			v9 = get_nox_xxx_conjurerMaxHealth_587000_312800();
			v11 = v9 * v11;
			break;
		case 1:
			v9 = get_nox_xxx_wizardMaxHealth_587000_312816();
			v11 = v9 * v11;
			break;
		}
	}
	*(float*)(LODWORD(v1) + 72) = v11 - (double)nox_float2int(v11);
	v10 = nox_float2int(v11);
	nox_xxx_unitAdjustHP_4EE460(*(uint32_t*)(LODWORD(v1) + 48), v10);
	nox_xxx_playerManaSub_4EEBF0(*(uint32_t*)(LODWORD(v1) + 16), 1);
	return 0;
}

//----- (0052F460) --------------------------------------------------------
int sub_52F460(float a1) {
	float v1;   // esi
	int v2;     // eax
	int result; // eax
	int v4;     // eax
	int v5;     // eax
	short v6;   // di
	short v7;   // ax
	float v8;   // [esp+10h] [ebp+4h]
	float v9;   // [esp+10h] [ebp+4h]

	v1 = a1;
	v2 = *(uint32_t*)(LODWORD(a1) + 48);
	if (!v2) {
		return 1;
	}
	if (*(uint32_t*)(v2 + 16) & 0x8020) {
		return 1;
	}
	if (*(uint32_t*)(LODWORD(a1) + 20)) {
		nox_xxx_playerManaAdd_4EEB80(v2, 20);
		nox_xxx_unitDamageClear_4EE5E0(*(uint32_t*)(LODWORD(a1) + 48), 20);
		result = 1;
	} else {
		v4 = *(uint32_t*)(LODWORD(a1) + 16);
		if (v4 && nox_xxx_testUnitBuffs_4FF350(v4, 8)) {
			result = 1;
		} else {
			v5 = *(uint32_t*)(LODWORD(a1) + 48);
			if (*(uint8_t*)(v5 + 8) & 2 &&
				sub_4FEA70((nox_object_t*)(uintptr_t)(uint32_t)v5, (float2*)(LODWORD(a1) + 28))) {
				result = 1;
			} else {
				v6 = nox_xxx_playerGetMaxMana_4EECB0(*(uint32_t*)(LODWORD(a1) + 48));
				if (v6 == nox_xxx_unitGetOldMana_4EEC80(*(uint32_t*)(LODWORD(a1) + 48))) {
					result = 1;
				} else if ((unsigned short)nox_xxx_unitGetHP_4EE780(*(uint32_t*)(LODWORD(a1) + 16)) > 1u) {
					if (nox_xxx_unitGetHP_4EE780(*(uint32_t*)(LODWORD(a1) + 16))) {
						v8 = *(float*)(LODWORD(a1) + 72);
						v9 = nox_xxx_gamedataGetFloatTable_419D70("ChannelLifeCoeff",
																  *(uint32_t*)(LODWORD(v1) + 8) - 1) +
							 v8;
						*(float*)(LODWORD(v1) + 72) = v9 - (double)nox_float2int(v9);
						v7 = nox_float2int(v9);
						nox_xxx_playerManaAdd_4EEB80(*(uint32_t*)(LODWORD(v1) + 48), v7);
						nox_xxx_unitDamageClear_4EE5E0(*(uint32_t*)(LODWORD(v1) + 16), 1);
					}
					result = 0;
				} else {
					result = 1;
				}
			}
		}
	}
	return result;
}

extern int nox_xxx_castShield1_native_52F5A0(void* spell);
extern int sub_52F650_native(void* spell);
extern int sub_52F670_native(void* spell);

//----- (0052F5A0) --------------------------------------------------------
int nox_xxx_castShield1_52F5A0(void* a1) { return nox_xxx_castShield1_native_52F5A0(a1); }

//----- (0052F650) --------------------------------------------------------
int sub_52F650(void* a1) { return sub_52F650_native(a1); }

//----- (0052F670) --------------------------------------------------------
int sub_52F670(void* a1) { return sub_52F670_native(a1); }

//----- (0052F690) --------------------------------------------------------
void nox_xxx_unitShield_52F690(int a1, int a2) {
	uint32_t* v2; // eax
	int v3;       // ecx

	v2 = (uint32_t*)nox_xxx_spellCastedFirst_4FE930();
	if (v2) {
		while (v2[12] != a1 || v2[1] != 51) {
			v2 = (uint32_t*)nox_xxx_spellCastedNext_4FE940(v2);
			if (!v2) {
				nox_xxx_spellBuffOff_4FF5B0((nox_object_t*)(uintptr_t)(uint32_t)a1, 26);
				return;
			}
		}
		v3 = v2[12];
		if (v3) {
			if (*(uint32_t*)(v3 + 16) & 0x8020) {
				nox_xxx_spellCancelSpellDo_4FE9D0(v2);
			} else if ((int)v2[18] - a2 > 0) {
				v2[18] -= a2;
			} else {
				nox_xxx_spellCancelSpellDo_4FE9D0(v2);
			}
		} else {
			nox_xxx_spellCancelSpellDo_4FE9D0(v2);
		}
	} else {
		nox_xxx_spellBuffOff_4FF5B0((nox_object_t*)(uintptr_t)(uint32_t)a1, 26);
	}
}

//----- (0052F710) --------------------------------------------------------
void nox_xxx_unitShieldReduceDamage_52F710(int a1, int* a2, int a3, int a4) {
	int v4; // edi
	int v5; // eax

	if (a1 && a2) {
		nox_xxx_aud_501960(131, a1, 0, 0);
		if (a4) {
			nox_xxx_netSendShieldFx_523670(a1, (float*)(a4 + 56));
		} else {
			nox_xxx_netSendShieldFx_523670(a1, 0);
		}
		v4 = *a2;
		if ((unsigned short)nox_xxx_unitGetHP_4EE780(a1) > v4) {
			*a2 = v4 / 2;
			if (!(v4 / 2)) {
				*a2 = 1;
			}
			nox_xxx_unitShield_52F690(a1, *a2);
		} else {
			if (*(uint8_t*)(a1 + 8) & 4 && (v5 = *(uint32_t*)(a1 + 748), !*(uint8_t*)(*(uint32_t*)(v5 + 276) + 2251)) &&
				*(uint8_t*)(v5 + 88) == 16 && a4 &&
				nox_server_testTwoPointsAndDirection_4E6E50((float2*)(a1 + 56), *(short*)(a1 + 124),
															(float2*)(a4 + 72)) &
					1) {
				*a2 = 1;
			} else {
				nox_xxx_unitSetHP_4E4560(a1, 2u);
				*a2 = 0;
			}
			nox_xxx_unitShield_52F690(a1, 999999);
			*(uint32_t*)(a1 + 520) = a4;
			*(uint32_t*)(a1 + 524) = a3;
			*(uint32_t*)(a1 + 536) = gameFrame();
		}
	}
}

//----- (0052F820) --------------------------------------------------------
int nox_xxx_onStartLightning_52F820(int a1) {
	int v1; // eax
	int v2; // eax

	*(uint32_t*)(a1 + 72) = 0;
	*(uint32_t*)(a1 + 76) = 0;
	if (*(uint32_t*)(a1 + 16)) {
		nox_xxx_spellCancelDurSpell_4FEB10(24, *(uint32_t*)(a1 + 16));
	}
	if (!*(uint32_t*)(a1 + 20)) {
		v1 = *(uint32_t*)(a1 + 16);
		if (*(uint8_t*)(v1 + 8) & 4) {
			v2 = *(uint32_t*)(*(uint32_t*)(v1 + 748) + 104);
			if (v2) {
				if (*(uint32_t*)(v2 + 12) & 0x40000 && *(uint8_t*)(*(uint32_t*)(v2 + 736) + 96) & 4) {
					*(uint32_t*)(a1 + 72) = v2;
				}
			}
		}
	}
	if (*(uint32_t*)(a1 + 72)) {
		*(uint8_t*)(a1 + 88) |= 2u;
	}
	nox_xxx_netSendPointFx_522FF0(129, (float2*)(a1 + 28));
	return 0;
}
