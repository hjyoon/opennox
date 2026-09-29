#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "common__strman.h"
#include "mapgen_legacy_ptr.h"

#include <string.h>
extern uint32_t dword_5d4594_2487580;
extern uint32_t dword_5d4594_2487672;
extern uint32_t dword_5d4594_2487576;
extern uint32_t dword_5d4594_1550916;

//----- (004D42C0) --------------------------------------------------------
int sub_4D42C0() { return dword_5d4594_1550916; }

//----- (00522AD0) --------------------------------------------------------
static float* nox_mapgenPlaceExitNative_522AD0(uint8_t* room, uint8_t* definition) {
	if (!room || !definition) {
		return NULL;
	}
	float* a1 = (float*)room;
	int v2;        // eax
	double v4;     // st7
	double v5;     // st7
	float* v6;     // ebx
	double v7;     // st7
	float a3;      // [esp+0h] [ebp-34h]
	float a4;      // [esp+4h] [ebp-30h]
	int2 a2a;      // [esp+14h] [ebp-20h]
	float2 a1a;    // [esp+1Ch] [ebp-18h]
	float v13 = 0; // [esp+24h] [ebp-10h]
	float v14 = 0; // [esp+28h] [ebp-Ch]
	float2 v15;    // [esp+2Ch] [ebp-8h]

	v2 = *(uint32_t*)(definition + 60);
	if (*((uint8_t*)a1 + v2 + 216)) {
		return 0;
	}
	switch (v2) {
	case 0:
		a1a.field_0 = a1[7] * 0.5 + a1[9] + 1.0;
		v4 = a1[10] + 10.0;
		a1a.field_4 = v4;
		break;
	case 1:
		a1a.field_0 = a1[7] * 0.5 + a1[9] + 1.0;
		v4 = a1[12] - 10.0;
		a1a.field_4 = v4;
		break;
	case 2:
		v5 = a1[11] - 10.0;
		a1a.field_0 = v5;
		v4 = a1[8] * 0.5 + a1[10] + 1.0;
		a1a.field_4 = v4;
		break;
	case 3:
		v5 = a1[9] + 10.0;
		a1a.field_0 = v5;
		v4 = a1[8] * 0.5 + a1[10] + 1.0;
		a1a.field_4 = v4;
		break;
	default:
		break;
	}
	nox_xxx_mapGenGetObjID_527940((char*)definition);
	v6 = nox_xxx_mapGenPlaceObj_5279B0(&a1a);
	if (v6) {
		nox_xxx_mapGenRoundFloatToPtr_520DF0(&a1a, &a2a);
		switch (*(uint32_t*)(definition + 60)) {
		case 0:
			v13 = 3.0;
			v7 = 2.0;
			--a2a.field_0;
			break;
		case 1:
			v7 = 2.0;
			--a2a.field_0;
			a2a.field_4 -= 2;
			v13 = 3.0;
			break;
		case 2:
			a2a.field_0 -= 2;
			v13 = 2.0;
			v7 = 3.0;
			--a2a.field_4;
			break;
		case 3:
			v13 = 2.0;
			v7 = 3.0;
			--a2a.field_4;
			break;
		default:
			v7 = v14;
			break;
		}
		v15.field_0 = (double)a2a.field_0 * 32.526913;
		v15.field_4 = (double)a2a.field_4 * 32.526913;
		a4 = v7 * 32.526913;
		a3 = v13 * 32.526913;
		nox_mapgenAddOccupiedRectNative_521BC0((uint8_t*)a1, &v15, a3, a4);
	}
	return v6;
}

float* nox_xxx_mapgen_522AD0(float* a1, int a2) {
	return nox_mapgenPlaceExitNative_522AD0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uintptr_t)a1),
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (005259E0) --------------------------------------------------------
int sub_5259E0() { return dword_5d4594_2487576; }

//----- (00527D50) --------------------------------------------------------
void* nox_objectTypeGetXfer(char* id);
static int nox_mapgenSetExitDestinationNative_527D50(nox_object_t* object, const char* destination) {
	if (!object || !destination || !object->collide_data) {
		return 0;
	}
	char* name = nox_xxx_getUnitName_4E39D0(object);
	if (nox_objectTypeGetXfer(name) != nox_xxx_XFerExit_4F4B90) {
		return 0;
	}
	strncpy(((nox_exit_collide_data_t*)object->collide_data)->map_name, destination, 0x50u);
	return 1;
}

int sub_527D50(int a1, char* a2) {
	return nox_mapgenSetExitDestinationNative_527D50(
		(nox_object_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (00522A40) --------------------------------------------------------
static int nox_mapgenMakeExitNative_522A40(uint8_t* theme) {
	if (!theme || !*(uint32_t*)(theme + 472)) {
		return 1;
	}
	for (uint8_t* room = nox_mapgenFarthestRoomNative_5259E0(); room;
		 room = nox_mapgenRoomSortedPrevNative_525C90(room)) {
		if (*(uint32_t*)room != 1) {
			continue;
		}
		uint8_t* definition = theme + 216;
		for (int index = 0; index < *(int32_t*)(theme + 472); ++index, definition += 64) {
			nox_object_t* object = (nox_object_t*)nox_mapgenPlaceExitNative_522AD0(room, definition);
			if (object) {
				nox_mapgenSetExitDestinationNative_527D50(object, (char*)(theme + 476));
				return 1;
			}
		}
	}
	return 0;
}

int nox_xxx_mapGenMakeExit_522A40(int a1) {
	return nox_mapgenMakeExitNative_522A40(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}

//----- (005259D0) --------------------------------------------------------
double sub_5259D0() { return *(float*)&dword_5d4594_2487580; }

//----- (00526A90) --------------------------------------------------------
void sub_526A90() {
	void* records = nox_mapgenLegacyPtrResolve(dword_5d4594_2487672);
	nox_mapgenLegacyPtrForget(records);
	free(records);
	dword_5d4594_2487672 = 0;
}

//----- (005228B0) --------------------------------------------------------
void nox_mapgenFinishPopulateNative_5228B0(uint8_t* theme) {
	if (!theme) {
		return;
	}
	wchar2_t* v1; // eax
	wchar2_t* v2; // eax
	wchar2_t* v3; // eax
	float2 player_start;

	nox_xxx_mapGenSetFlags_5235F0(157);
	if (!nox_mapgenMakeExitNative_522A40(theme)) {
		v1 = nox_strman_loadString_40F1D0("NoExit", 0, "C:\\NoxPost\\src\\Server\\MapGen\\Generate\\populate.c", 848);
		nox_xxx_printToAll_4D9FD0(0, v1);
		v2 = nox_strman_loadString_40F1D0("NoExit", 0, "C:\\NoxPost\\src\\Server\\MapGen\\Generate\\populate.c", 849);
		nox_xxx_printToAll_4D9FD0(0, v2);
		v3 = nox_strman_loadString_40F1D0("NoExit", 0, "C:\\NoxPost\\src\\Server\\MapGen\\Generate\\populate.c", 850);
		nox_xxx_printToAll_4D9FD0(0, v3);
	}
	nox_mapgenMaxRoomDistanceNative_5259D0();
	uint8_t* root = (uint8_t*)nox_mapgenLegacyPtrResolve(dword_5d4594_1550916);
	for (uint8_t* room = root; room; room = nox_mapgenRoomSortedNextNative_525C90(room)) {
		nox_xxx_mapGenSetFlags_5235F0(157);
		if (*(uint32_t*)(room + 372) && !(room[52] & 2)) {
			nox_mapgenPopulateDecorNative_522340(theme, room);
		}
		if (*(uint32_t*)(theme + 60)) {
			for (uint32_t token = *(uint32_t*)(room + 368); token;) {
				uint32_t* rect = (uint32_t*)nox_mapgenLegacyPtrResolve(token);
				if (!rect) {
					break;
				}
				if (*rect) {
					nox_xxx_tileGetDefByName_51D4D0("CrystalBlue");
				} else {
					nox_xxx_tileGetDefByName_51D4D0("CrystalRed");
				}
				sub_5245A0(0, (float*)(rect + 1), (long long)((*(float*)(rect + 3) - *(float*)(rect + 1) + 0.5) * 0.030743772),
						   (long long)((*(float*)(rect + 4) - *(float*)(rect + 2) + 0.5) * 0.030743772));
				token = rect[6];
			}
		}
	}
	if (root) {
		float* root_floats = (float*)root;
		player_start.field_0 = (root_floats[11] + root_floats[9]) * 0.5;
		player_start.field_4 = (root_floats[12] + root_floats[10]) * 0.5;
		nox_xxx_mapGenGetObjID_527940("PlayerStart");
		nox_xxx_mapGenPlaceObj_5279B0(&player_start);
	}
	sub_469B90((int*)(theme + 536));
	sub_526A90();
}

void nox_xxx_mapGenFinishPopulate_5228B0_mapgen_populate(int a1) {
	nox_mapgenFinishPopulateNative_5228B0(
		(uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1));
}
