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
#include "common__binfile.h"
#include "common__crypt.h"
#include "common__net_list.h"
#include "common__random.h"
#include "operators.h"
#include "input.h"
#include "mapgen_decor_records.h"
#include "mapgen_legacy_ptr.h"
#include "mapgen_theme_native.h"
#include "floor_brush_native.h"

#include "client__video__draw_common.h"
#include "common__magic__speltree.h"
#include "server__script__script.h"

#include <time.h>
#include "mapgen_allocation.h"

// Keep each graph allocation with its matching owner. The default service
// remains CRT; high-address allocation is requested by scoped fixtures.
#undef calloc
#undef free
#define calloc nox_mapgen_alloc_memory
#define free nox_mapgen_release_memory

extern uint32_t dword_5d4594_2487656;
extern uint32_t dword_5d4594_3835368;
extern uint32_t dword_5d4594_2487804;
extern uint32_t dword_5d4594_2487632;
extern uint32_t dword_5d4594_2487884;
extern uint32_t dword_5d4594_3835372;
extern uint32_t dword_5d4594_2487628;
extern uint32_t dword_5d4594_2487672;
extern uint32_t dword_5d4594_2487676;
extern uint32_t dword_5d4594_2487652;
extern uint32_t dword_5d4594_2487624;
extern uint32_t dword_5d4594_2487620;
extern uint32_t dword_5d4594_2487708;
extern uint32_t nox_xxx_energyBoltTarget_5d4594_2487880;

extern uint32_t dword_5d4594_2487712;

_Static_assert(sizeof(nox_mapgen_occupied_rect_legacy) == 28,
			   "wrong mapgen occupied-rectangle record size");
_Static_assert(offsetof(nox_mapgen_occupied_rect_legacy, next_token) == 24,
			   "wrong mapgen occupied-rectangle next-token offset");
_Static_assert(sizeof(nox_mapgen_decor_entry_legacy) == 100,
			   "wrong mapgen decor-entry record size");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, choices_token) == 80,
			   "wrong mapgen decor-entry choices-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, foreach_token) == 84,
			   "wrong mapgen decor-entry FOREACH-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, next_token) == 88,
			   "wrong mapgen decor-entry next-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, shuffle_next_token) == 92,
			   "wrong mapgen decor-entry shuffle-next offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, shuffle_prev_token) == 96,
			   "wrong mapgen decor-entry shuffle-prev offset");
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

float get_nox_xxx_warriorMaxHealth_587000_312784();
float get_nox_xxx_wizardMaxHealth_587000_312816();
float get_nox_xxx_conjurerMaxHealth_587000_312800();

float get_nox_xxx_warriorMaxMana_587000_312788();
float get_nox_xxx_wizardMaximumMana_587000_312820();
float get_nox_xxx_conjurerMaxMana_587000_312804();

// 0051DEA0 is implemented by the native-pointer Go callback exported from
// wall.go. The original return value was ignored by the sole foreach caller.

// 0051DED0 is implemented by the native-pointer Go ObjectData writer.

//----- (0051E540) --------------------------------------------------------
// Token reading and comment skipping at 51E540, 51E570 and 51E630
// are implemented in mapgen_token_reader_public.go.

// Theme section parsers and cleanup are implemented in Go.

//----- (00520DF0) --------------------------------------------------------
// GAME.EXE 00520DF0 is implemented in Go.

//----- (00520E60) --------------------------------------------------------
// Native-width implementation: mapgen_room_occupancy_public.go.
//----- (00520EA0) --------------------------------------------------------
// Native-width implementation: mapgen_grid_allocation_public.go.
//----- (00520F80) --------------------------------------------------------
// Native-width implementation: mapgen_grid_allocation_public.go.
//----- (00521100) --------------------------------------------------------
// Native-width implementation: mapgen_room_occupancy_public.go.
//----- (00521180) --------------------------------------------------------
// Native-width implementation: mapgen_room_occupancy_public.go.
//----- (00521200) --------------------------------------------------------
// Native-width implementation: mapgen_room_occupancy_public.go.
//----- (00521290) --------------------------------------------------------
// Native-width implementation: mapgen_grid_allocation_public.go.
//----- (005212B0) --------------------------------------------------------
// Native-width implementation: mapgen_room_displace_public.go.
//----- (005213E0) --------------------------------------------------------
// Native-width implementation: mapgen_buffer_public.go.

//----- (00521400) --------------------------------------------------------
// Native-width implementation: mapgen_buffer_public.go.

//----- (00521710) --------------------------------------------------------
// Native-width implementation: mapgen_room_lifecycle_public.go.

//----- (00521720) --------------------------------------------------------
// Native-width implementation: mapgen_room_lifecycle_public.go.

//----- (00521730) --------------------------------------------------------
// Native-width implementation: mapgen_room_lifecycle_public.go.

//----- (00521760) --------------------------------------------------------
// Native-width implementation: mapgen_room_lifecycle_public.go.

//----- (005217A0) --------------------------------------------------------
// Native-width implementation: mapgen_room_occupancy_public.go.
//----- (00521820) --------------------------------------------------------
// Native-width implementation: mapgen_room_occupancy_public.go.
//----- (00521850) --------------------------------------------------------
// GAME.EXE 00521850 is implemented in Go.

//----- (00521880) --------------------------------------------------------
// GAME.EXE 00521880 is implemented in Go.

//----- (005218B0) --------------------------------------------------------
// Native-width implementation: mapgen_room_link_public.go.

//----- (00521900) --------------------------------------------------------
// Native-width implementation: mapgen_room_link_public.go.

//----- (00521940) --------------------------------------------------------
// Native-width implementation: mapgen_room_allocation_public.go.

//----- (00521990) --------------------------------------------------------
// Native-width implementation: mapgen_room_allocation_public.go.

//----- (00521A10) --------------------------------------------------------
// Native-width implementation: mapgen_room_lifecycle_public.go.

//----- (00521A40) --------------------------------------------------------
// Native-width implementation: mapgen_room_lifecycle_public.go.

//----- (00521A70) --------------------------------------------------------
// Native-width implementation: mapgen_room_link_public.go.

//----- (00521AA0) --------------------------------------------------------
// GAME.EXE 00521AA0 is implemented in Go.

//----- (00521B00) --------------------------------------------------------
// Native-width implementation: mapgen_room_rand_attach_public.go.

//----- (00521B30) --------------------------------------------------------
// Native-width implementation: mapgen_room_rand_attach_public.go.

//----- (00521B60) --------------------------------------------------------
// Native-width implementation: mapgen_room_rand_attach_public.go.

//----- (00521B90) --------------------------------------------------------
// Native-width implementation: mapgen_room_rand_attach_public.go.

//----- (00521BC0) --------------------------------------------------------
// Native-width implementation: mapgen_occupied_rect_public.go.

//----- (00521C10) --------------------------------------------------------
// Native-width implementation: mapgen_occupied_rect_public.go.

//----- (00521C60) --------------------------------------------------------
// Native-width implementation: mapgen_choice_inventory_public.go.

//----- (00521CB0) --------------------------------------------------------
// Native-width implementation: mapgen_place_decor_object_public.go.

//----- (00521EB0) --------------------------------------------------------
// Native-width implementation: mapgen_occupied_rect_public.go.

//----- (00521F10) --------------------------------------------------------
// Native-width implementation: mapgen_occupied_rect_public.go.

//----- (00521FE0) --------------------------------------------------------
// Native-width implementation: mapgen_choice_inventory_public.go.

//----- (005220E0) --------------------------------------------------------
// Native-width implementation: mapgen_spellbook_public.go.

//----- (005221A0) --------------------------------------------------------
nox_object_t* nox_mapgenMakeEnchantedItemNative_5221A0(
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
// Native-width implementation: mapgen_choice_inventory_public.go.

//----- (00522340) --------------------------------------------------------
// Native-width implementation: mapgen_populate_decor_public.go.

//----- (00522370) --------------------------------------------------------
// Native-width implementation: mapgen_populate_decor_public.go.

//----- (005224B0) --------------------------------------------------------
// Native-width implementation: mapgen_populate_decor_public.go.

//----- (005226D0) --------------------------------------------------------
// Native-width implementation: mapgen_room_monster_public.go.

//----- (005227B0) --------------------------------------------------------
// Native-width implementation: mapgen_room_monster_public.go.

//----- (00522810) --------------------------------------------------------
// Native-width implementation: mapgen_room_monster_public.go.

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
	sub_51D0F0((char)0x80);
	for (uint8_t* room = (uint8_t*)nox_xxx_mapGenGetTopRoom_521710(); room;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		nox_xxx_mapGenSetFlags_5235F0((char)0x9C);
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
					sub_522C80(&from);
					sub_522C80(&to);
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
				sub_522C80(&from);
				sub_522C80(&to);
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
		nox_xxx_mapGenSetFlags_5235F0((char)0x9C);
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
// Ported to point_fx_native.go; retained as an inactive source reference.
#if 0
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
#endif

//----- (00523030) --------------------------------------------------------
int nox_xxx_netSendFxAllCli_523030(float2* a1, const void* a2, int a3) {
	return nox_xxx_netSendFxAllCliNative_523030(a1, (void*)a2, a3);
}

//----- (00523150) --------------------------------------------------------
// Ported to point_fx_native.go; retained as inactive source references.
#if 0
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
#endif

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
// Native-width entry: predict_linear_native.go. Preserve the incoming body.
#if 0
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
		nox_netlist_addToMsgListCli_40EBC0(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), 1, (unsigned char*)v14,
										   14);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}
#endif

//----- (005235F0) --------------------------------------------------------
// Progress packets are emitted by mapgen_progress.go.

// GAME.EXE 00523670 uses the Go Shield effect entry.

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
// Native-width implementation: mapgen_room_link_public.go.

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
	float center[2];
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve((uint32_t)a2);
	if (!fill) {
		return 0;
	}

	nox_xxx_tileGetDefByName_51D4D0(fill->tile_name);
	sub_5245A0(a1, a3, *a4, a4[1]);
	sub_544020(fill->object_name);
	center[0] = (double)*a4 * ((double)0x1.04371ep+4f) + *a3;
	center[1] = (double)a4[1] * ((double)0x1.04371ep+4f) + a3[1];
	return sub_543680(&center[0]);
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
	float point[2];
	float center[2];
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
	v9 = (double)v6 * ((double)0x1.04371ep+5f);
	v10 = v5 - 2 * v6;
	v19 = v9;
	v11 = v9 + *a3;
	point[1] = a3[1];
	point[0] = v11;
	if (v7 / 2 >= 0) {
		do {
			sub_5245A0(a1, &point[0], v10, 1);
			v12 = *v4;
			v10 += 2;
			v13 = v10 < *v4;
			point[0] = point[0] - ((double)0x1.04371ep+5f);
			point[1] = point[1] + ((double)0x1.04371ep+5f);
			if (!v13) {
				v10 = v12;
				point[0] = *a3;
			}
			++v24;
		} while (v24 <= v4[1] / 2);
	}
	v14 = v4[1];
	v15 = *v4 - v8;
	v16 = 0;
	point[0] = v19 + *a3;
	point[1] = (double)(v14 - 1) * ((double)0x1.04371ep+5f) + a3[1];
	if (v14 / 2 > 0) {
		do {
			sub_5245A0(a1, &point[0], v15, 1);
			v17 = *v4;
			v15 += 2;
			v13 = v15 < *v4;
			point[0] = point[0] - ((double)0x1.04371ep+5f);
			point[1] = point[1] - ((double)0x1.04371ep+5f);
			if (!v13) {
				v15 = v17;
				point[0] = *a3;
			}
			++v16;
		} while (v16 < v4[1] / 2);
	}
	sub_544020(fill->object_name);
	center[0] = (double)*v4 * ((double)0x1.04371ep+4f) + *a3;
	center[1] = (double)v4[1] * ((double)0x1.04371ep+4f) + a3[1];
	return sub_543680(&center[0]);
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
	float point[2];
	float center[2];
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
		point[0] = *a3 + ((double)0x1.04371ep+5f);
		point[1] = a3[1] + ((double)0x1.04371ep+5f);
		if (v5 < 1) {
			v5 = 1;
		}
		v7 = a4[1] - 2;
		if (v7 < 1) {
			v7 = 1;
		}
		sub_5245A0(a1, &point[0], v5, v7);
		v8 = *a4;
		if ((int)*a4 < 4) {
			if (v8 == 3) {
				v15 = a3[1];
				point[0] = *a3 + ((double)0x1.04371ep+5f);
				point[1] = v15;
				sub_5245A0(a1, &point[0], 1, 1);
				point[1] = point[1] + ((double)0x1.04371ep+6f);
				sub_5245A0(a1, &point[0], 1, 1);
			}
		} else {
			v9 = nox_xxx_mapGenRandFunc_526AC0(1, v8 - 3);
			v10 = v9;
			v11 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - v9 - 2);
			point[1] = a3[1];
			point[0] = (double)v11 * ((double)0x1.04371ep+5f) + *a3;
			sub_5245A0(a1, &point[0], v10, 1);
			v12 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - 3);
			v13 = v12;
			v14 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - v12 - 2);
			v26 = a4[1] - 1;
			point[0] = (double)v14 * ((double)0x1.04371ep+5f) + *a3;
			point[1] = (double)v26 * ((double)0x1.04371ep+5f) + a3[1];
			sub_5245A0(a1, &point[0], v13, 1);
		}
		v16 = v4[1];
		if (v16 < 4) {
			if (v16 == 3) {
				v21 = *a3;
				point[1] = a3[1] + ((double)0x1.04371ep+5f);
				point[0] = v21;
				sub_5245A0(a1, &point[0], 1, 1);
				point[0] = point[0] + ((double)0x1.04371ep+6f);
				sub_5245A0(a1, &point[0], 1, 1);
			}
		} else {
			v17 = nox_xxx_mapGenRandFunc_526AC0(1, v16 - 3);
			v18 = v17;
			v19 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - v17 - 2);
			point[0] = *a3;
			point[1] = (double)v19 * ((double)0x1.04371ep+5f) + a3[1];
			sub_524610(a1, &point[0], v18);
			v20 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - 3);
			v27 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - v20 - 2);
			point[0] = (double)(*v4 - 1) * ((double)0x1.04371ep+5f) + *a3;
			point[1] = (double)v27 * ((double)0x1.04371ep+5f) + a3[1];
			sub_524610(a1, &point[0], v20);
		}
		sub_544020(fill->object_name);
		center[0] = (double)*v4 * ((double)0x1.04371ep+4f) + *a3;
		center[1] = (double)v4[1] * ((double)0x1.04371ep+4f) + a3[1];
		sub_543680(&center[0]);
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
		sub_525330(&v34, v7 - 1);
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
		sub_525330(&v34, v13 - 1);
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
		sub_525370(&v34, v21 - 1);
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
		sub_525370(&v34, v29 - 1);
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
// Native-width implementation: mapgen_wall_paths_public.go.
//----- (00525370) --------------------------------------------------------
// Native-width implementation: mapgen_wall_paths_public.go.
//----- (005253B0) --------------------------------------------------------
int sub_5253B0(float* a1) {
	double v1;        // st7
	double v2;        // st7
	double v3;        // st7
	double v4;        // st7
	int result;       // eax
	unsigned char v6; // [esp+4h] [ebp-18h]
	float2 point;
	void* wall;

	v1 = a1[1] - 32.526913;
	v6 = 0;
	point.field_0 = *a1;
	point.field_4 = v1;
	if (sub_526DD0(&point, &wall)) {
		v6 = 1;
	}
	v2 = a1[1] + 32.526913;
	point.field_0 = *a1;
	point.field_4 = v2;
	if (sub_526DD0(&point, &wall)) {
		v6 |= 2u;
	}
	v3 = *a1 - 32.526913;
	point.field_4 = a1[1];
	point.field_0 = v3;
	if (sub_526DD0(&point, &wall)) {
		v6 |= 4u;
	}
	v4 = *a1 + 32.526913;
	point.field_4 = a1[1];
	point.field_0 = v4;
	if (sub_526DD0(&point, &wall)) {
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
	nox_object_t* object = nox_xxx_mapGenPlaceObj_5279B0(&position);
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
	nox_object_t* object = nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 5);
	}
	position.field_0 += 65.053825f;
	object = nox_xxx_mapGenPlaceObj_5279B0(&position);
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
	nox_object_t* object = nox_xxx_mapGenPlaceObj_5279B0(&position);
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
	nox_object_t* object = nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 7);
	}
	position.field_4 += 65.053825f;
	object = nox_xxx_mapGenPlaceObj_5279B0(&position);
	if (object) {
		nox_mapgenOrientObjNative_527C60(object, 1);
	}
	return 1;
}

int nox_xxx_mapgen_5258E0(int a1, float2* a2, int a3) {
	return nox_mapgenPlaceDoubleDoorVerticalNative_5258E0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2, a3);
}

// Room distance and ranking are implemented in mapgen_room_ranking.go.

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

//----- (00526BC0) --------------------------------------------------------
double sub_526BC0(float a1, float a2) {
	return (double)(unsigned int)nox_platform_rand() * (a2 - a1) * 0.000030518509 + a1;
}

//----- (00526DD0) --------------------------------------------------------
// Native-width implementation: mapgen_wall_paths_public.go.
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
// Native-width implementation: mapgen_wall_paths_public.go.
//----- (00527C60) --------------------------------------------------------
// Native-width implementation: mapgen_room_monster_public.go.

//----- (00527DB0) --------------------------------------------------------
// Native-width implementation: mapgen_spellbook_public.go.

//----- (00527E50) --------------------------------------------------------
#if 0 // Retained PE32 implementation; native public entry is in Go.
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

#endif

//----- (00528030) --------------------------------------------------------
#if 0 // PE32 reference. Native Go uses typed health, update and player records.
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
#endif

//----- (00528190) --------------------------------------------------------
#if 0 // Retained PE32 body; public native entry is implemented in Go.
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

#endif

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
#if 0 // Restored by outbound_object_chat_native.go.
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
				*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(source + 748) + 276) + 2064), 1, (unsigned char*)v10,
				(unsigned char)v10[8] + 11);
		}
	} else {
		result = nox_xxx_getFirstPlayerUnit_4DA7C0();
		for (i = result; result; i = result) {
			nox_netlist_addToMsgListCli_40EBC0(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), 1,
											   (unsigned char*)v10, (unsigned char)v10[8] + 11);
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
		nox_netlist_addToMsgListCli_40EBC0(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), 1, (unsigned char*)v2,
										   3);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}
#endif

extern uintptr_t nox_server_monster_cmd_send_native_528BD0(nox_object_t*, nox_object_t*, char*, short);

uintptr_t nox_xxx_monsterCmdSend_528BD0(nox_object_t* unit, nox_object_t* source, const char* command, short duration) {
	return nox_server_monster_cmd_send_native_528BD0(unit, source, (char*)command, duration);
}

// GAME.EXE 00529BC0 uses the Go spellcaster initialization entry.

// GAME.EXE 00529CE0 uses the Go monster action-data entry.

//----- (0052A420) --------------------------------------------------------
// Implemented in Go.

//----- (0052A440) --------------------------------------------------------
// Implemented in Go.
// 52A440: using guessed type char var_100[256];

// GAME.EXE 0052AAB0 uses the Go monster buff entry.

//----- (0052AD10) --------------------------------------------------------
#if 0 // Native NPC voice: npc_voice_native_exports.go
size_t nox_xxx_readNPCVoiceSet_52AD10(int a1) {
	const char** v1; // eax
	const char** v2; // esi
	size_t result;   // eax
	const char** v4; // eax
	int v5;          // [esp+4h] [ebp-104h]
	char v6[256];    // [esp+8h] [ebp-100h]

	if (nox_crypt_IsReadOnly()) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v5, 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((unsigned char*)v6, (unsigned char)v5);
		v6[(unsigned char)v5] = 0;
		v4 = nox_xxx_getDefaultSoundSet_424350(v6);
		result = nox_xxx_setNPCVoiceSet_424320(a1, (int)v4);
	} else {
		v1 = (const char**)nox_xxx_monsterGetSoundSet_424300(a1);
		v2 = v1;
		if (v1) {
			LOBYTE(v5) = strlen(*v1);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v5, 1u);
			// Write mode reads the immutable sound-set name into the file.
			result = nox_xxx_fileReadWrite_426AC0_file3_fread((unsigned char*)*v2, (unsigned char)v5);
		} else {
			LOBYTE(v5) = 0;
			result = nox_xxx_fileReadWrite_426AC0_file3_fread(&v5, 1u);
		}
	}
	return result;
}
#endif

// 52AD10: using guessed type char var_100[256];

//----- (0052BA70) --------------------------------------------------------
#if 0 // Native NPC equipment: npc_equipment_native_entry.go
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
#endif

//----- (0052BEB0) --------------------------------------------------------
#if 0 // Native Inversion cast: inversion_native_entry.go
int sub_52BEB0(int a1, int a2, int a3, int a4) {
	int v4;   // eax
	float v6; // [esp+0h] [ebp-10h]

	v6 = nox_xxx_gamedataGetFloat_419D40("InversionRange");
	nox_xxx_getMissilesInCircle_518170(a4 + 56, v6, nox_xxx_changeOwner_52BE40, a3);
	v4 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v4, a4, 0, 0);
	return 1;
}
#endif

//----- (0052BF20) --------------------------------------------------------
#if 0 // Native resource spells: resource_spells_native_exports.go
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
#endif

//----- (0052BF50) --------------------------------------------------------
#if 0 // Native resource spells: resource_spells_native_exports.go
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
#endif

//----- (0052BFA0) --------------------------------------------------------
#if 0 // Native Pull cast: pull_push_native_entry.go
int nox_xxx_castPull_52BFA0(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;   // eax
	float v8; // [esp+0h] [ebp-10h]

	v8 = -(nox_xxx_gamedataGetFloat_419D40("PullPowerCoeff") * (double)a6);
	nox_xxx_mapPushUnitsAround_52E040(a4 + 56, 600.0, 10.0, v8, 0, 0, 0);
	v6 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v6, a3, 0, 0);
	return 1;
}
#endif

//----- (0052C000) --------------------------------------------------------
#if 0 // Native Push cast: pull_push_native_entry.go
int nox_xxx_castPush_52C000(int a1, int a2, int a3, int a4, int a5, int a6) {
	int v6;   // eax
	float v8; // [esp+18h] [ebp+18h]

	v8 = nox_xxx_gamedataGetFloat_419D40("PushPowerCoeff") * (double)a6;
	nox_xxx_mapPushUnitsAround_52E040(a4 + 56, 600.0, 10.0, v8, 0, 0, 0);
	v6 = nox_xxx_spellGetAud44_424800(a1, 0);
	nox_xxx_aud_501960(v6, a3, 0, 0);
	return 1;
}
#endif

//----- (0052C060) --------------------------------------------------------
#if 0 // Native Fumble cast: fumble_cast_native_entry.go
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
#endif

//----- (0052C3E0) --------------------------------------------------------
#if 0 // Native Burn cast: burn_cast_native_entry.go
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
#endif

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
#if 0 // Native Poison cast: poison_cast_native_entry.go
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
#endif

//----- (0052C790) --------------------------------------------------------
#if 0 // Native Fireball: fireball_native_exports.go
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
#endif

//----- (0052CA80) --------------------------------------------------------
// Retired: the public Mark selector uses native Go pointers. Keep the
// original four-int implementation only as disabled reference provenance.
#if 0
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
#endif

//----- (0052CBD0) --------------------------------------------------------
// Native-width implementation: server/spell_mark_slots_52cbd0.go, selected
// through legacy/spell_mark_slots_52cbd0.go. Preserve the original body as
// provenance; its four-int entry truncates both caster and marker pointers.
#if 0
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
#endif
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
// Native-width group callback: server/spell_lock_52ce90.go.
// Preserve the original integer-pointer body as disabled provenance.
#if 0
void sub_52CE60(int a1) {
	if (*(uint8_t*)(a1 + 8) & 0x80) {
		*(uint32_t*)(a1 + 508) = *getMemU32Ptr(0x5D4594, 2487716);
		*(uint32_t*)(a1 + 136) = gameFrame() + 60 * gameFPS();
	}
}
#endif

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
// Native-width candidate callback: server/spell_lock_52ce90.go.
// Preserve the original integer-pointer body as disabled provenance.
#if 0
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
#endif

//----- (0052D060) --------------------------------------------------------
// Native-width door-group traversal: server/spell_lock_52ce90.go.
// Preserve the original integer-pointer body as disabled provenance.
#if 0
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
#endif

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
#if 0 // Original PE32 body retained; native entry is in cleansing_flame_native.go.
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
					nox_xxx_netClientPredictLinear_523530((nox_object_t*)v9);
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
#endif

//----- (0052D8A0) --------------------------------------------------------
#if 0 // Original PE32 body retained; native entry is in meteor_cast_native.go.
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
#endif

//----- (0052D9D0) --------------------------------------------------------
#if 0 // Original PE32 body retained; native entry is in meteor_cast_native.go.
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
#endif

//----- (0052DB60) --------------------------------------------------------
// Native-width implementation: server/spell_toxic_cloud_52db60.go, selected
// through legacy/spell_toxic_cloud_52db60.go. Preserve the original body as
// provenance; its five-int entry truncates caster, owner and acceptance pointers.
#if 0
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
#endif

//----- (0052DC80) --------------------------------------------------------
#if 0 // Original PE32 body retained; native entry is in arachnaphobia_native.go.
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
#endif

//----- (0052DD50) --------------------------------------------------------
#if 0 // Native resource spells: resource_spells_native_exports.go
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
#endif

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
#if 0 // Native pointers and field layout: player_marker_movable_native_exports.go:nox_xxx_isObjectMovable_52E020
unsigned int nox_xxx_isObjectMovable_52E020(int a1) {
	unsigned int result; // eax

	if (*(uint32_t*)(a1 + 16) & 0x8068) {
		result = 0;
	} else {
		result = ((unsigned int)~*(uint32_t*)(a1 + 8) >> 22) & 1;
	}
	return result;
}
#endif


//----- (0052E040) --------------------------------------------------------

//----- (0052E0E0) --------------------------------------------------------


//----- (0052F5A0) --------------------------------------------------------

//----- (0052F650) --------------------------------------------------------

//----- (0052F670) --------------------------------------------------------

// GAME.EXE 0052F690 and 0052F710 use Go exports in spell_shield_damage_52f690.go.
