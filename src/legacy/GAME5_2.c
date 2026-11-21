#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "GAME5_2.h"
#include "client__system__parsecmd.h"
#include "common__binfile.h"
#include "common__random.h"
#include "common__strman.h"

#include "client__gui__guiinv.h"
#include "client__shell__mainmenu.h"
#include "common__system__team.h"

#include "MixPatch.h"
#include "client__video__draw_common.h"
#include "common/fs/nox_fs.h"
#include "common__magic__speltree.h"
#include "operators.h"

extern uint32_t dword_5d4594_2650652;

// The server discovery reply and latency statistics are implemented in Go.

//----- (0056F780) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (0056F7D0) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (0056F820) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (0056F870) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (0056F8C0) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (0056F920) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (0056F9E0) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (0056FA40) --------------------------------------------------------
// The protected value update is implemented in Go.

//----- (00578B00) --------------------------------------------------------
#if 0 // Restored by object_field_native_entry.go.
unsigned int nox_xxx_netGetUnitCodeCli_578B00(nox_drawable* dr) {
	unsigned int result; // eax

	if (!dr) {
		return 0;
	}
	result = dr->field_32;
	if (result >= 0x8000) {
		return 0;
	}
	if (dr->flags28 & 0x20400000) {
		BYTE1(result) |= 0x80u;
	}
	return result;
}
#endif

//----- (00578B30) --------------------------------------------------------
#if 0 // Restored by object_field_native_entry.go.
int nox_xxx_netClearHighBit_578B30(short a1) { return a1 & 0x7FFF; }
#endif

//----- (00578B40) --------------------------------------------------------
// This entry is implemented in object_by_extent_4ed020.go.

//----- (00578B70) --------------------------------------------------------
#if 0 // Restored by object_field_native_entry.go.
unsigned int nox_xxx_netTestHighBit_578B70(unsigned int a1) { return (a1 >> 15) & 1; }
#endif

//----- (00579870) --------------------------------------------------------
// Ported to Go in waypoint.go so the native WpNext field is used on every
// pointer width instead of the PE32 +484 offset and int return ABI.

//----- (005798A0) --------------------------------------------------------
// The duplicate pending-waypoint iterator is exported by waypoint.go too.

//----- (00579E70) --------------------------------------------------------
// This entry is implemented in waypoint.go.

//----- (00579EE0) --------------------------------------------------------
// Native public entry: flag_predicate_native_entry.go.

//----- (0057A160) --------------------------------------------------------
// Preserved original: native Go public entry in game5_public_native.go.
#if 0
int nox_xxx_playerCanTalkMB_57A160(int a1) {
	int result; // eax

	if (a1 && nox_common_gameFlags_check_40A5C0(2)) {
		result = (*(uint32_t*)(a1 + 3680) >> 3) & 1;
	} else {
		result = 0;
	}
	return result;
}
#endif

//----- (0057A1B0) --------------------------------------------------------
// Ported to map_rule_mode_57a1b0.go.
#if 0
char* sub_57A1B0(short a1) {
	for (int i = 0; i < 7; i++) {
		if ((a1 & 0x17F0) == *getMemU32Ptr(0x587000, 312212 + 8 * i)) {
			return (char*)getMemPtr(0x587000, 312208 + 8 * i);
		}
	}
	return 0;
}
#endif

//----- (0057A1E0) --------------------------------------------------------
// Ported to map_rule_defaults_57a1e0.go.
#if 0
char sub_57A1E0(int* a1, const char* a2, int* a3, char a4, short a5) {
	int* v5;           // esi
	int v6;            // ebx
	unsigned char v7;  // dl
	char* v8;          // edi
	int v9;            // ecx
	int v10;           // eax
	unsigned char v12; // cl
	char result;       // al
	char v14[256];     // [esp+14h] [ebp-200h]
	char v15[256];     // [esp+114h] [ebp-100h]

	v5 = a3;
	v6 = a5 & 0x17F0;
	if (a3) {
		sub_57ADF0(a3);
	}
	a1[6] = -1;
	a1[7] = -1;
	a1[8] = -1;
	a1[9] = -1;
	a1[10] = -1;
	a1[11] = -1;
	a1[12] = -1;
	if (a4 & 3) {
		v15[0] = 0;
		strcpy(v14, "maps\\");
		strncat(v14, (const char*)a1, 8u);
		strcat(v14, "\\");
		if (a4 & 2) {
			strcpy(v15, v14);
			if (a2) {
				strcat(v15, a2);
			} else {
				strcat(v15, "user.rul");
			}
			v10 = sub_57A3F0(v15, (char*)a1, a3, v6);
			v5 = a3;
		} else {
			v10 = 0;
		}
		if (a4 & 1 && !v10) {
			strncat(v14, (const char*)a1, 8u);
			strncat(v14, ".rul", 255);
			sub_57A3F0(v14, (char*)a1, v5, v6);
		}
	}
	if (dword_5d4594_2650652 && a4 & 4) {
		sub_57A3F0("internet.rul", (char*)a1, v5, v6);
	}
	result = a5;
	if (a5 & 0x40) {
		result = sub_453FA0(a1 + 6, 132, 0);
	}
	return result;
}
// 57A1E0: using guessed type char var_100[256];
#endif

// The map rule file reader, line parser and token handler are implemented in Go.

// GAME.EXE 0057A950 builds the current map rule path in Go.

//----- (0057A9F0) --------------------------------------------------------
// Ported to map_rule_mode_57a1b0.go.
#if 0
int sub_57A9F0(const char* a1, const char* a2) {
	char FileName[256]; // [esp+Ch] [ebp-100h]

	strcpy(FileName, "maps\\");
	strcat(FileName, a1);
	strcat(FileName, "\\");
	strcat(FileName, a2);
	return nox_fs_remove(FileName);
}
#endif

//----- (0057AAA0) --------------------------------------------------------
// Ported to map_rule_save_57aaa0.go.
#if 0
char* sub_4159B0(int a1);
char* sub_415E40(char* a1);
char sub_57AAA0(const char* a1, char* a2, int* a3) {
	FILE* v3;           // eax
	FILE* v4;           // ebp
	int* i;             // esi
	char* v6;           // eax
	int v7;             // esi
	int v8;             // edi
	char* v9;           // eax
	int v10;            // esi
	int v11;            // edi
	char* v12;          // eax
	int v13;            // ebx
	int v14;            // esi
	char* v15;          // edi
	char* v16;          // eax
	int v18;            // [esp+10h] [ebp-27Ch]
	// Each original 24-byte header was immediately followed by a 36-byte spell
	// list. sub_57A1E0 writes through byte 51, so they form one object.
	struct {
		char header[24];
		char spells[36];
	} v19_v20, v21_v22;
	char v23[256];      // [esp+8Ch] [ebp-200h]
	char FileName[256]; // [esp+18Ch] [ebp-100h]

	LOBYTE(v3) = a2[52];
	if ((char)v3 >= 0) {
		strcpy(FileName, "maps\\");
		strncat(FileName, a2, 8u);
		strncat(FileName, "\\", 10u);
		strcat(FileName, a1);
		v3 = nox_fs_create_text(FileName);
		v4 = v3;
		if (v3) {
			if (dword_5d4594_2650652) {
				strcpy(v21_v22.header, a2);
				strcpy(v19_v20.header, a2);
				sub_57A1E0((int*)&v21_v22, 0, 0, 4, *((uint16_t*)a2 + 26));
				sub_57A1E0((int*)&v19_v20, 0, 0, 3, *((uint16_t*)a2 + 26));
			}
			if (a3) {
				for (nox_list_item_t* link = nox_common_list_getFirstSafe_425890((nox_list_item_t*)a3); link;
					 link = nox_common_list_getNextSafe_4258A0(link)) {
					nox_rule_list_item_t* item = (nox_rule_list_item_t*)link;
					nox_sprintf(v23, "%S\n", item->text);
					nox_fs_fputs(v4, v23);
				}
			}
			v6 = sub_57A1B0(*((uint16_t*)a2 + 26));
			nox_fs_fputs(v4, v6);
			nox_fs_fputs(v4, "\n");
			v7 = 1;
			v8 = 136;
			do {
				if (nox_xxx_spellIsValid_424B50(v7) && !sub_454000(a2 + 24, v7) &&
					nox_xxx_spellFlags_424A70(v7) & 0x7000000 &&
					(!dword_5d4594_2650652 || sub_454000(v21_v22.spells, v7) || !sub_454000(v19_v20.spells, v7))) {
					v9 = nox_xxx_spellNameByN_424870(v7);
					nox_sprintf(v23, "%s %s \"%s\" %s\n", "set", "spell", v9, "off");
					nox_fs_fputs(v4, v23);
				}
				++v7;
				--v8;
			} while (v8);
			v10 = 1;
			v11 = 26;
			do {
				if (!(v10 & *((uint32_t*)a2 + 12))) {
					v12 = sub_415E40((char*)v10);
					if (v12) {
						nox_sprintf(v23, "%s %s \"%s\" %s\n", "set", "Armor", v12, "off");
						nox_fs_fputs(v4, v23);
					}
				}
				v10 *= 2;
				--v11;
			} while (v11);
			v13 = 1;
			v14 = 1;
			v18 = 27;
			v15 = a2 + 44;
			do {
				if (!((unsigned char)v13 & (unsigned char)*v15)) {
					v16 = sub_4159B0((char*)v14);
					if (v16) {
						nox_sprintf(v23, "%s %s \"%s\" %s\n", "set", "weapon", v16, "off");
						nox_fs_fputs(v4, v23);
					}
				}
				if (v13 == 128) {
					v13 = 1;
					++v15;
				} else {
					v13 *= 2;
				}
				v14 *= 2;
				--v18;
			} while (v18);
			LOBYTE(v3) = 0;
			nox_fs_close(v4);
		}
	}
	return (char)v3;
}
#endif

// GAME.EXE 0057ADF0 releases map rule notes in Go.

//----- (0057AE30) --------------------------------------------------------
// Ported to map_rule_mode_57a1b0.go.
#if 0
int sub_57AE30(const char* a1) {
	for (int i = 0; i < 7; i++) {
		if (!strcmp((const char*)getMemPtr(0x587000, 312208 + 8 * i), a1)) {
			return *getMemU32Ptr(0x587000, 312212 + 8 * i);
		}
	}
	return 0;
}
#endif

//----- (0057AEA0) --------------------------------------------------------
// This entry is implemented in spells.go.

//----- (0057AF20) --------------------------------------------------------
// GAME.EXE 0057AF20, 0057B0A0, and 0057B180 are restored by
// pause_fx_57af30.go. These bodies retain the original ABI32 layout only as
// provenance.
#if 0
int nox_xxx_get_57AF20() { return dword_5d4594_2523804; }

//----- (0057B0A0) --------------------------------------------------------
void sub_57B0A0() {
	int result;   // eax
	uint32_t* v1; // ecx

	result = dword_5d4594_2523804;
	if (!result) {
		return;
	}
	v1 = *(uint32_t**)&dword_5d4594_2523780;
	if (dword_5d4594_2523780 && (!*getMemU32Ptr(0x5D4594, 2523772) || *getMemU32Ptr(0x5D4594, 2523772) == 1)) {
		nox_xxx_netSendPointFx_522FF0(154, (float2*)(dword_5d4594_2523780 + 56));
		v1 = *(uint32_t**)&dword_5d4594_2523780;
	}
	if (dword_5d4594_2523776) {
		nox_xxx_delayedDeleteObject_4E5CC0(*(int*)&dword_5d4594_2523776);
		v1 = *(uint32_t**)&dword_5d4594_2523780;
	}
	dword_5d4594_2523776 = 0;
	if (v1) {
		nox_xxx_playerSetState_4FA020(v1, 13);
	}
	dword_5d4594_2523780 = 0;
	if (!sub_45D9B0()) {
		sub_413A00(0);
	}
	dword_5d4594_2523804 = 0;
}

//----- (0057B180) --------------------------------------------------------
long long nox_xxx___Getcvt_57B180() { return *getMemU64Ptr(0x5D4594, 2523788); }
#endif

//----- (0057B3D0) --------------------------------------------------------
int nox_cheat_allowall = 0;

//----- (0057B400) --------------------------------------------------------
// Ported to client_glyph_access_57b400.go.
#if 0
int nox_xxx_client_57B400(nox_drawable* dr) {
	int v1; // eax

	v1 = *getMemU32Ptr(0x5D4594, 2523876);
	if (!*getMemU32Ptr(0x5D4594, 2523876)) {
		v1 = nox_xxx_getTTByNameSpriteMB_44CFC0("Glyph");
		*getMemU32Ptr(0x5D4594, 2523876) = v1;
	}
	if (!dword_8531A0_2576) {
		return 0;
	}
	if (dr->field_27 != v1 || ((nox_playerInfo*)dword_8531A0_2576)->info.playerClass == 1) {
		return 1;
	}
	return 0;
}

//----- (0057B450) --------------------------------------------------------
int sub_57B450(nox_drawable* a1p) {
	int v1;           // eax
	unsigned char v2; // bl

	v1 = *getMemU32Ptr(0x5D4594, 2523880);
	if (!*getMemU32Ptr(0x5D4594, 2523880)) {
		v1 = nox_xxx_getTTByNameSpriteMB_44CFC0("Glyph");
		*getMemU32Ptr(0x5D4594, 2523880) = v1;
	}
	if (!a1p || !getMemPtr(0x852978, 8) || !dword_8531A0_2576 ||
		a1p->field_27 == v1 && ((nox_playerInfo*)dword_8531A0_2576)->info.playerClass != 1) {
		return 0;
	}
	if (nox_cheat_allowall) {
		return 1;
	}
	v2 = 1 << ((nox_playerInfo*)dword_8531A0_2576)->info.playerClass;
	return (v2 & (unsigned char)sub_57B370(a1p->flags28, a1p->flags29, a1p->field_27)) != 0;
}
#endif

//----- (0057B810) --------------------------------------------------------
// Preserved original: native Go public entry in game5_public_native.go.
#if 0
int nox_xxx_collideReflect_57B810(float* a1, int a2) {
	int result; // eax
	double v3;  // st7
	int v4;     // ecx

	result = a2;
	v3 = *(float*)a2;
	if (a1[1] * *a1 <= 0.0) {
		v4 = *(uint32_t*)(a2 + 4);
		*(float*)(a2 + 4) = *(float*)a2;
		*(uint32_t*)a2 = v4;
	} else {
		*(float*)a2 = -*(float*)(a2 + 4);
		*(float*)(a2 + 4) = -v3;
	}
	return result;
}
#endif

//----- (0057B850) --------------------------------------------------------

// The alias table routines at 0057B920, 0057B9A0 and 0057BA10 are in network_alias_57b920.go.

//----- (0057C090) --------------------------------------------------------
// Ported to Go in groups.go so the list follows the native next field rather
// than truncating the group pointer and reading the PE32 +88 offset.

// The ray normal and point projection entries are implemented in Go.

void nullsub_10(uint32_t a1) {}

#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42CC50
int sub_42CC50(void** this) { return sub_42C770(this); }
#endif


int nox_xxx_j_inventoryNameSignInit_467460(void) { return nox_xxx_inventoryNameSignInit_4671E0(); }

int nullsub_8(int a1, int a2) { return 0; }
void nullsub_28(uint32_t a1) {}
void nullsub_30(uint32_t a1) {}
void nullsub_29(void) {}
void nullsub_35(uint32_t a1, uint32_t a2) {}
void nullsub_31(uint32_t a1) {}
void nullsub_9(uint32_t a1) {}
void nox_xxx_j_allocHitArray_511840(void) { nox_xxx_allocHitArray_5486D0(); }
