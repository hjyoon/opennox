#include "mapgen_allocation.h"
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "common__strman.h"
#include "mapgen_legacy_ptr.h"
#include "server__mapgen__generate__populate.h"

extern uint32_t dword_5d4594_2487580;
extern uint32_t dword_5d4594_2487672;
extern uint32_t dword_5d4594_2487576;
extern uint32_t dword_5d4594_1550916;

//----- (004D42C0) --------------------------------------------------------
int sub_4D42C0() { return dword_5d4594_1550916; }

//----- (005259E0) --------------------------------------------------------
int sub_5259E0() { return dword_5d4594_2487576; }

//----- (005259D0) --------------------------------------------------------
double sub_5259D0() { return *(float*)&dword_5d4594_2487580; }

//----- (00526A90) --------------------------------------------------------
void sub_526A90() {
	void* records = nox_mapgenLegacyPtrResolve(dword_5d4594_2487672);
	nox_mapgenLegacyPtrForget(records);
	nox_mapgen_release_memory(records);
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

	nox_xxx_mapGenSetFlags_5235F0((char)0x9D);
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
		nox_xxx_mapGenSetFlags_5235F0((char)0x9D);
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
