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

//----- (00554040) --------------------------------------------------------
unsigned int nox_server_makeServerInfoPacket_554040(const char* inBuf, int inSz, char* out) {
	char buf[72];

	char* v3 = sub_416640();
	char* game = nox_xxx_cliGamedataGet_416590(0);
	if (!sub_43AF30() || sub_4D6F30()) {
		return 0;
	}
	char playerLimit = nox_xxx_servGetPlrLimit_409FA0();
	char playerCount = nox_common_playerInfoCount_416F40();
	if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
		--playerCount;
		--playerLimit;
	}
	char* srvName = nox_xxx_serverOptionsGetServername_40A4C0();
	buf[0] = 0;
	buf[1] = 0;
	buf[2] = 13;
	buf[3] = playerCount;
	buf[4] = playerLimit;
	buf[5] = v3[101] & 0xF;
	buf[6] = ((unsigned char)v3[101]) >> 4;
	*(uint32_t*)&buf[7] = *((uint32_t*)game + 11);
	strcpy(&buf[10], nox_xxx_mapGetMapName_409B40());
	buf[19] = v3[102] | sub_43BE50_get_video_mode_id();
	buf[20] = v3[100];
	buf[21] = v3[100] & 0x10;
	*(uint32_t*)&buf[24] = *((uint32_t*)v3 + 12);
	unsigned int gameFlags = nox_common_gameFlags_getVal_40A5B0();
	if (nox_xxx_isQuest_4D6F50()) {
		gameFlags = (gameFlags & 0xFFFFFF7Fu) | 0x1000u;
		*(uint16_t*)&buf[68] = nox_game_getQuestStage_4E3CC0();
	}
	*(uint32_t*)&buf[28] = gameFlags;
	*(uint32_t*)&buf[32] = *((uint32_t*)game + 12);
	*(uint16_t*)&buf[36] = *(uint16_t*)(v3 + 105);
	*(uint16_t*)&buf[38] = *(uint16_t*)(v3 + 107);
	*(uint32_t*)&buf[40] = *((uint32_t*)v3 + 11);
	*(uint32_t*)&buf[44] = *(uint32_t*)(&inBuf[8]); // timestamp of the packet
	memcpy(&buf[48], game + 24, 20);
	memcpy(&out[0], buf, 72);
	strcpy(&out[72], srvName);
	return 72 + strlen(srvName) + 1;
}

unsigned int sub_554290() {
	unsigned int v0;     // edi
	int v1;              // ebx
	nox_playerInfo* v2;  // esi
	unsigned int v3;     // eax

	v0 = -1;
	v1 = 0;
	v2 = nox_common_playerInfoGetFirst_416EA0();
	if (!v2) {
		return 0;
	}
	do {
		uint8_t* raw = (uint8_t*)v2;
		if (raw[2064] != 31 && sub_554240(raw[2064]) > 0) {
			v3 = sub_554240(raw[2064]);
			if (v3 < v0) {
				v0 = v3;
			}
			++v1;
		}
		v2 = nox_common_playerInfoGetNext_416EE0(v2);
	} while (v2);
	if (v1) {
		return v0;
	} else {
		return 0;
	}
}

int sub_554300() {
	int v0;     // ebx
	int v1;     // edi
	nox_playerInfo* v2; // esi

	v0 = 0;
	v1 = 0;
	v2 = nox_common_playerInfoGetFirst_416EA0();
	if (!v2) {
		return 0;
	}
	do {
		uint8_t* raw = (uint8_t*)v2;
		if (raw[2064] != 31 && (int)sub_554240(raw[2064]) > 0) {
			v0 += sub_554240(raw[2064]);
			++v1;
		}
		v2 = nox_common_playerInfoGetNext_416EE0(v2);
	} while (v2);
	if (v1) {
		return v0 / v1;
	} else {
		return 0;
	}
}

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
int nox_xxx_packetDynamicUnitCode_578B40(int a1) {
	int result; // eax
	nox_object_t* v2; // eax

	result = a1;
	if ((a1 & 0x8000) == 0x8000) {
		BYTE1(result) &= 0x7Fu;
		v2 = nox_xxx_netGetUnitByExtent_4ED020(result);
		if (v2) {
			result = v2->net_code;
		} else {
			result = 0;
		}
	}
	return result;
}

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
nox_waypoint_t* sub_579E70() {
	nox_waypoint_t* result = calloc(1u, sizeof(*result));
	if (result) {
		result->flags |= 0x1000000u;
	}
	return result;
}

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
int nox_xxx_playerCheckSpellClass_57AEA0(int a1, int a2) {
	int v2;     // eax
	int result; // eax
	int v4;     // ecx

	v2 = nox_xxx_spellFlags_424A70(a2);
	if (a1 == 1) {
		v4 = 0x2000000;
	} else {
		if (a1 != 2) {
			return 9;
		}
		v4 = 0x4000000;
	}
	if (v2 & 0x1000000 || v4 & v2) {
		result = 0;
	} else {
		result = 9;
	}
	return result;
}

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

//----- (0057C790) --------------------------------------------------------
void sub_57C790(float4* a1, float2* a2, float2* a3, float a4) {
	double v4; // st5
	double v5; // st7
	double v6; // st6
	double v7; // st7
	float v8;  // [esp+0h] [ebp-10h]
	float v9;  // [esp+4h] [ebp-Ch]
	float v10; // [esp+8h] [ebp-8h]
	float v11; // [esp+14h] [ebp+4h]
	float v12; // [esp+18h] [ebp+8h]
	float v13; // [esp+20h] [ebp+10h]

	v8 = a1->field_8 - a1->field_0;
	v4 = a1->field_C - a1->field_4;
	v9 = v4;
	v5 = v4 * (a2->field_4 - a1->field_4) + v8 * (a2->field_0 - a1->field_0);
	v6 = a4 * a4;
	v10 = v5 * v8 / v6;
	a3->field_0 = v10 + a1->field_0;
	a3->field_4 = v5 * v9 / v6 + a1->field_4;
	if (a1->field_0 >= (double)a1->field_8) {
		v12 = a1->field_0;
		v13 = a1->field_8;
	} else {
		v13 = a1->field_0;
		v12 = a1->field_8;
	}
	if (a1->field_4 >= (double)a1->field_C) {
		v7 = a1->field_C;
		v11 = a1->field_4;
	} else {
		v7 = a1->field_4;
		v11 = a1->field_C;
	}
	if (a3->field_0 >= (double)v13) {
		if (a3->field_0 > (double)v12) {
			a3->field_0 = v12;
		}
	} else {
		a3->field_0 = v13;
	}
	if (v7 <= a3->field_4) {
		if (a3->field_4 > (double)v11) {
			a3->field_4 = v11;
		}
	} else {
		a3->field_4 = v7;
	}
}

//----- (0057C8A0) --------------------------------------------------------
int nox_xxx_mathPointOnTheLine_57C8A0(float4* a1, float2* a2, float2* a3) {
	float4* v3; // ecx
	float2* v4; // edx
	double v5;  // st7
	double v6;  // st6
	double v7;  // st6
	double v8;  // st7
	float v10;  // [esp+0h] [ebp-Ch]
	float v11;  // [esp+0h] [ebp-Ch]
	float v12;  // [esp+8h] [ebp-4h]
	float v13;  // [esp+8h] [ebp-4h]
	float v14;  // [esp+10h] [ebp+4h]
	float v15;  // [esp+10h] [ebp+4h]
	float v16;  // [esp+14h] [ebp+8h]
	float v17;  // [esp+18h] [ebp+Ch]

	v3 = a1;
	v4 = a3;
	v5 = a1->field_8 - a1->field_0;
	v6 = a1->field_C - a1->field_4;
	v12 = v6;
	v10 = v6 * v12 + v5 * v5;
	v7 = (a2->field_4 - a1->field_4) * v12 + (a2->field_0 - a1->field_0) * v5;
	v14 = v7;
	v13 = v14 * v12 / v10;
	a3->field_0 = v5 * v7 / v10 + v3->field_0;
	v15 = v13 + v3->field_4;
	a3->field_4 = v15;
	if (v3->field_0 >= (double)v3->field_8) {
		v8 = v3->field_8;
		v16 = v3->field_0;
	} else {
		v8 = v3->field_0;
		v16 = v3->field_8;
	}
	if (v3->field_4 >= (double)v3->field_C) {
		v11 = v3->field_4;
		v17 = v3->field_C;
	} else {
		v17 = v3->field_4;
		v11 = v3->field_C;
	}
	return v8 <= v4->field_0 && v4->field_0 <= (double)v16 && v15 >= (double)v17 && v15 <= (double)v11;
}

//----- (0057CDB0) --------------------------------------------------------
int sub_57F2A0(float2* a1, int a2, int a3);
char sub_57F1D0(float2* a1);
int sub_57CDB0(int2* a1, float* a2, float2* a3) {
	int2* v3;   // esi
	char v4;    // bl
	int result; // eax
	float2* v6; // eax
	float2* v7; // eax
	float2* v8; // eax
	char v9;    // [esp+10h] [ebp+4h]

	v3 = a1;
	v9 = sub_57F2A0((float2*)a2, a1->field_0, a1->field_4);
	v4 = sub_57F1D0((float2*)a2 + 1);
	switch (sub_57B500(v3->field_0, v3->field_4, 64)) {
	case 0:
		if (v9 != 1 && v9) {
			v8 = a3;
			v8->field_0 = 0.70709997;
			v8->field_4 = 0.70709997;
			return 1;
		}
		a3->field_0 = -0.70709997;
		a3->field_4 = -0.70709997;
		return 1;
	case 1:
		if (v9 == 1 || v9 == 2) {
			a3->field_0 = 0.70709997;
			a3->field_4 = -0.70709997;
			return 1;
		}
		a3->field_0 = -0.70709997;
		a3->field_4 = 0.70709997;
		return 1;
	case 2:
		switch (v9) {
		case 0:
			v8 = a3;
			a3->field_0 = -0.70709997;
			if (!(v4 & 2)) {
				a3->field_4 = -0.70709997;
				return 1;
			}
			v8->field_4 = 0.70709997;
			return 1;
		case 1:
			v6 = a3;
			if (!(v4 & 1)) {
				v6->field_0 = 0.70709997;
				v6->field_4 = -0.70709997;
				return 1;
			}
			a3->field_0 = -0.70709997;
			a3->field_4 = -0.70709997;
			result = 1;
			break;
		case 2:
			v8 = a3;
			a3->field_0 = 0.70709997;
			if (v4 & 1) {
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_4 = -0.70709997;
			return 1;
		case 3:
			v8 = a3;
			if (!(v4 & 4)) {
				v8->field_0 = 0.70709997;
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_0 = -0.70709997;
			a3->field_4 = 0.70709997;
			return 1;
		default:
			return 1;
		}
		return result;
	case 3:
		if (!v9) {
			v8 = a3;
			a3->field_0 = -0.70709997;
			if (!(v4 & 2)) {
				a3->field_4 = -0.70709997;
				return 1;
			}
			v8->field_4 = 0.70709997;
			return 1;
		}
		if (v9 != 1) {
			v8 = a3;
			v8->field_0 = 0.70709997;
			v8->field_4 = 0.70709997;
			return 1;
		}
		v7 = a3;
		if (!(v4 & 1)) {
			v7->field_0 = 0.70709997;
			v7->field_4 = -0.70709997;
			return 1;
		}
		v7->field_0 = -0.70709997;
		v7->field_4 = -0.70709997;
		return 1;
	case 4:
		if (v9 == 1) {
			v7 = a3;
			if (!(v4 & 1)) {
				v7->field_0 = 0.70709997;
				v7->field_4 = -0.70709997;
				return 1;
			}
			v7->field_0 = -0.70709997;
			v7->field_4 = -0.70709997;
			return 1;
		}
		if (v9 == 2) {
			v8 = a3;
			a3->field_0 = 0.70709997;
			if (v4 & 1) {
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_4 = -0.70709997;
			return 1;
		}
		a3->field_0 = -0.70709997;
		a3->field_4 = 0.70709997;
		return 1;
	case 5:
		if (v9 == 2) {
			v8 = a3;
			a3->field_0 = 0.70709997;
			if (v4 & 1) {
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_4 = -0.70709997;
			return 1;
		}
		if (v9 == 3) {
			v8 = a3;
			if (!(v4 & 4)) {
				v8->field_0 = 0.70709997;
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_0 = -0.70709997;
			a3->field_4 = 0.70709997;
			return 1;
		}
		a3->field_0 = -0.70709997;
		a3->field_4 = -0.70709997;
		return 1;
	case 6:
		if (!v9) {
			v8 = a3;
			a3->field_0 = -0.70709997;
			if (!(v4 & 2)) {
				a3->field_4 = -0.70709997;
				return 1;
			}
			v8->field_4 = 0.70709997;
			return 1;
		}
		if (v9 == 3) {
			v8 = a3;
			if (!(v4 & 4)) {
				v8->field_0 = 0.70709997;
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_0 = -0.70709997;
			a3->field_4 = 0.70709997;
			return 1;
		}
		v6 = a3;
		v6->field_0 = 0.70709997;
		v6->field_4 = -0.70709997;
		return 1;
	case 7:
		if (v9 == 1) {
			v7 = a3;
			if (!(v4 & 1)) {
				v7->field_0 = 0.70709997;
				v7->field_4 = -0.70709997;
				return 1;
			}
			v7->field_0 = -0.70709997;
			v7->field_4 = -0.70709997;
			return 1;
		}
		v8 = a3;
		if (v4 & 1) {
			v8->field_0 = 0.70709997;
			v8->field_4 = 0.70709997;
			return 1;
		}
		a3->field_0 = -0.70709997;
		a3->field_4 = 0.70709997;
		return 1;
	case 8:
		if (v9 == 2) {
			v8 = a3;
			a3->field_0 = 0.70709997;
			if (v4 & 1) {
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_4 = -0.70709997;
			return 1;
		} else {
			v8 = a3;
			a3->field_0 = -0.70709997;
			if (!(v4 & 1)) {
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_4 = -0.70709997;
			return 1;
		}
	case 9:
		if (v9 == 3) {
			v8 = a3;
			if (!(v4 & 4)) {
				v8->field_0 = 0.70709997;
				v8->field_4 = 0.70709997;
				return 1;
			}
			a3->field_0 = -0.70709997;
			a3->field_4 = 0.70709997;
			return 1;
		} else {
			v7 = a3;
			if (v4 & 4) {
				v7->field_0 = 0.70709997;
				v7->field_4 = -0.70709997;
				return 1;
			} else {
				v7->field_0 = -0.70709997;
				v7->field_4 = -0.70709997;
				return 1;
			}
		}
	case 0xA:
		if (v9) {
			if (v4 & 2) {
				a3->field_0 = 0.70709997;
				a3->field_4 = -0.70709997;
				return 1;
			}
			v8 = a3;
			v8->field_0 = 0.70709997;
			v8->field_4 = 0.70709997;
			return 1;
		} else {
			v8 = a3;
			a3->field_0 = -0.70709997;
			if (!(v4 & 2)) {
				a3->field_4 = -0.70709997;
				return 1;
			}
			v8->field_4 = 0.70709997;
			return 1;
		}
	default:
		return 0;
	}
}

//----- (0057F1D0) --------------------------------------------------------
char sub_57F1D0(float2* a1) {
	char v1;          // bl
	int v2;           // edi
	double v3;        // st7
	double v4;        // st6
	double v5;        // st7
	unsigned char v7; // [esp+Ch] [ebp-4h]

	v1 = 0;
	v2 = nox_float2int(a1->field_0);
	v7 = nox_float2int(a1->field_4) % 23;
	v3 = (double)(unsigned char)(v2 % 23);
	if (v3 >= 11.5) {
		v4 = (double)v7;
		if (v4 >= 11.5) {
			v1 = 4;
		}
		if (v4 <= 11.5) {
			v1 |= 1u;
		}
	}
	if (v3 <= 11.5) {
		v5 = (double)v7;
		if (v5 >= 11.5) {
			v1 |= 8u;
		}
		if (v5 <= 11.5) {
			v1 |= 2u;
		}
	}
	return v1;
}

//----- (0057F2A0) --------------------------------------------------------
int sub_57F2A0(float2* a1, int a2, int a3) {
	int v3;     // esi
	int v4;     // eax
	int result; // eax
	int v6;     // eax
	float v7;   // [esp+0h] [ebp-Ch]
	float v8;   // [esp+0h] [ebp-Ch]

	v7 = a1->field_0 - (double)(23 * a2);
	v3 = nox_float2int(v7);
	v8 = a1->field_4 - (double)(23 * a3);
	v4 = nox_float2int(v8);
	if (v3 <= v4) {
		LOBYTE(v4) = 22 - v3 <= v4;
		v6 = v4 - 1;
		LOBYTE(v6) = v6 & 0xFD;
		result = v6 + 3;
	} else {
		LOBYTE(v4) = 22 - v3 <= v4;
		result = v4 + 1;
	}
	return result;
}

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
