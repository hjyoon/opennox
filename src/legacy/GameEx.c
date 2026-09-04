#define DG_DYNARR_IMPLEMENTATION
#include "GameEx.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5_2.h"
#include "client__gui__guimsg.h"
#include "client__gui__window.h"
#include "common/fs/nox_fs.h"
#include "memmap.h"

#include "client__shell__noxworld.h"
#include "operators.h"


//-------------------------------------------------------------------------
// Data declarations

unsigned int gameex_flags = 0x1E;

int nox_CharToOemW(const wchar2_t* pSrc, char* pDst) { return nox_sprintf(pDst, "%S", pSrc); }

//----- (10001C20) --------------------------------------------------------
char getPlayerClassFromObjPtr(int a1) { return *(uint8_t*)(*(uint32_t*)(*(uint32_t*)(a1 + 748) + 276) + 2251); }

//----- (10001D40) --------------------------------------------------------
char playerInfoStructParser_0(void* a1p) {
	char* a1 = a1p;
	char* v1;  // esi
	char pDst; // [esp+10h] [ebp-18h]

	if (a1 == (char*)-2)
		return 0;
	v1 = nox_common_playerInfoGetFirst_416EA0();
	if (!v1)
		return 0;
	while (1) {
		nox_CharToOemW((const wchar2_t*)v1 + 2352, &pDst);
		if (!strcmp(&pDst, a1 + 2))
			break;
		v1 = nox_common_playerInfoGetNext_416EE0((int)v1);
		if (!v1)
			return 0;
	}
	a1[1] = *((uint8_t*)nox_xxx_objGetTeamByNetCode_418C80(*((uint32_t*)v1 + 515)) + 4);
	*a1 = v1[2251];
	return 1;
}

//----- (10001E10) --------------------------------------------------------
char playerInfoStructParser_1(void* a1p, int* a3) {
	int a1 = a1p;
	char* v3;     // eax
	char* v4;     // eax
	uint32_t* v6; // eax
	char pDst;    // [esp+Ch] [ebp-18h]

	if (a1 == -2)
		return 0;
	v3 = nox_common_playerInfoGetFirst_416EA0();
	int a2 = v3;
	if (!v3)
		return 0;
	while (1) {
		nox_CharToOemW((const wchar2_t*)(a2 + 4704), &pDst);
		if (!strcmp(&pDst, (const char*)(a1 + 2)))
			break;
		v4 = nox_common_playerInfoGetNext_416EE0(a2);
		a2 = v4;
		if (!v4)
			return 0;
	}
	v6 = nox_xxx_objGetTeamByNetCode_418C80(*(uint32_t*)(a2 + 2060));
	*a3 = (int)v6;
	*(uint8_t*)(a1 + 1) = *((uint8_t*)v6 + 4);
	*(uint8_t*)a1 = *(uint8_t*)(a2 + 2251);
	return 1;
}

//----- (10002030) --------------------------------------------------------
char playerDropATrap(int playerObj) {
	int v2; // eax
	int i;  // esi
	// int v5; // [esp+10h] [ebp-14h]
	char v7; // [esp+18h] [ebp-Ch]
	char v8; // [esp+1Fh] [ebp-5h]
	float pos[2] = {0};

	v7 = 17;
	if (!playerObj)
		return 0;
	v8 = 0;
	v2 = *(uint32_t*)(*(uint32_t*)(playerObj + 0x2EC) + 0x114);
	pos[0] = *(float*)(v2 + 0xE30);
	pos[1] = *(float*)(v2 + 0xE34);
	if (!(*(uint8_t*)(*(uint32_t*)(*(uint32_t*)(playerObj + 0x2EC) + 0x114) + 0xE60) &
		  3) // check playerGameStatus/isObs
		&& *(uint8_t*)(*(uint32_t*)(playerObj + 0x2EC) + 0x58) != 1) {
		for (i = *(uint32_t*)(playerObj + 0x1F8); i; i = *(uint32_t*)(i + 0x1F0)) {
			if (*(uint8_t*)(i + 0xA) == v7) // check if something from *(byte*)(unit+0xA)=17
			{
				nox_xxx_drop_4ED810(
					(nox_object_t*)(uintptr_t)(uint32_t)playerObj,
					(nox_object_t*)(uintptr_t)(uint32_t)i,
					(float2*)pos); // drop this item
				return 1;
			}
		}
	}
	return v8;
}

void OnLibraryNotice_420(uint32_t arg1, uint32_t arg2, uint32_t arg3, uint32_t arg4) {
	int v23 = arg1;
	int v19 = arg2;
	uint32_t* v16 = getPlayerClassFromObjPtr(arg1);
	if (*(uint8_t*)(v19 + 0xA) != 17) {
		nox_xxx_inventoryServPlace_4F36F0(v23, v19, 1, 1);
		return;
	}
	char v17 = *(uint8_t*)(v19 + 4);
	if (v17 != 0x6A) {
		if ((v17 == 0x6B || v17 == 0x6D) && (uint8_t)v16) {
			goto ifIsWarrior;
		}
		nox_xxx_inventoryServPlace_4F36F0(v23, v19, 1, 1);
		return;
	}
	if ((uint8_t)v16 == 1) {
		nox_xxx_inventoryServPlace_4F36F0(v23, v19, 1, 1);
		return;
	}
ifIsWarrior:
	nox_xxx_netPriMsgToPlayer_4DA2C0(v23, (const char*)getMemAt(0x587000, 215732),
									 0); // 0x5BBAB4 = pickup.c:ObjectEquipClassFail
	nox_xxx_aud_501960(925, v23, 2, *(uint32_t*)(v23 + 36));
}

//----- (10004330) --------------------------------------------------------
int getFlagValueFromFlagIndex(signed int a1) {
	signed int v1;   // edx
	unsigned int v2; // eax
	signed int v3;   // ecx
	int result;      // eax

	v1 = 2;
	v2 = a1;
	if (a1 < 0)
		v2 = -a1;
	v3 = 1;
	while (1) {
		if (v2 & 1)
			v3 *= v1;
		v2 >>= 1;
		if (!v2)
			break;
		v1 *= v1;
	}
	if (a1 >= 0)
		result = v3;
	else
		result = 1 / v3;
	return result;
}
