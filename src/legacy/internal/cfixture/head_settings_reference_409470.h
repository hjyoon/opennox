#ifndef HEAD_SETTINGS_PRIVATE_REFERENCE_409470_H
#define HEAD_SETTINGS_PRIVATE_REFERENCE_409470_H
#include "GAME1.h"
#include "GAME3_2.h"
#include "common__binfile.h"
extern uint32_t dword_5d4594_3484;
extern int head_settings_reference_flags_409470(uint32_t);
extern int head_settings_reference_notify_409470(int);
/* These two service names are redirected only AFTER public headers. The
 * original function statements and all other dependencies remain exact. */
#define nox_common_gameFlags_check_40A5C0 head_settings_reference_flags_409470
#define nox_xxx_netNotifyRate_4D7F10 head_settings_reference_notify_409470
static int head_reference_nox_xxx_parseString_409470(FILE* a1, uint8_t* a2) {
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
			v6 = isspace(CharType[0]);
			if (v6) {
				break;
			}
			v4 = 0;
			if (v3 != '/' || v5 != '/') {
				*v2++ = v3;
			} else {
				nox_binfile_skipLine_409520(a1);
				v2 = a2;
				v3 = *(uint32_t*)CharType;
				v4 = 1;
			}
		}
	} while (v4);
	*v2 = 0;
	return 1;
}

static char* head_reference_nox_server_currentMapGetFilename_409B30() { return (char*)getMemAt(0x5D4594, 2598188); }

static char* head_reference_nox_xxx_mapGetMapName_409B40() { return (char*)getMemAt(0x85B3FC, 36); }

static char* head_reference_nox_xxx_gameSetMapPath_409D70(char* a1) {
	char* result;  // eax
	char* v2;      // eax
	signed int v3; // esi

	result = (char*)nox_strcmpi((const char*)getMemAt(0x5D4594, 2598188), a1);
	if (result) {
		strncpy((char*)getMemAt(0x5D4594, 2598188), a1, 0x50u);
		*getMemU8Ptr(0x5D4594, 2598267) = 0;
		v2 = strrchr(a1, 92);
		if (v2) {
			v3 = strlen(v2 + 1) - 4;
			result = strncpy((char*)getMemAt(0x85B3FC, 36), v2 + 1, v3);
		} else {
			v3 = strlen((const char*)getMemAt(0x5D4594, 2598188)) - 4;
			if (v3 < 0) {
				v3 = 0;
			}
			result = strncpy((char*)getMemAt(0x85B3FC, 36), (const char*)getMemAt(0x5D4594, 2598188), v3);
		}
		*getMemU8Ptr(0x85B3FC, 36 + v3) = 0;
		nox_server_gameSettingsUpdated = 1;
	}
	return result;
}

static int head_reference_nox_xxx_getServerSubFlags_409E60() { return dword_5d4594_3484; }

static int head_reference_nox_xxx_servSetPlrLimit_409F80(int a1) {
	int result; // eax

	result = a1;
	if (*getMemU32Ptr(0x5D4594, 3464) != a1) {
		*getMemU32Ptr(0x5D4594, 3464) = a1;
		nox_server_gameSettingsUpdated = 1;
	}
	return result;
}

static int head_reference_nox_xxx_servGetPlrLimit_409FA0() { return *getMemU32Ptr(0x5D4594, 3464); }

static short head_reference_nox_xxx_servGamedataGet_40A020(short a1) { return *getMemU16Ptr(0x5D4594, 3488 + 2 * sub_409A70(a1)); }

static void head_reference_nox_xxx_set3512_40A340(int a1) { *getMemU32Ptr(0x5D4594, 3512) = a1; }

static int head_reference_nox_xxx_get3512_40A350() { return *getMemU32Ptr(0x5D4594, 3512); }

static char* head_reference_nox_xxx_gameSetServername_40A440(char* a1) {
	char* result; // eax
	char v2[16];  // [esp+0h] [ebp-10h]

	result = a1;
	if (a1) {
		strncpy(v2, a1, 0xFu);
		v2[15] = 0;
		result = (char*)nox_strcmpi((const char*)getMemAt(0x5D4594, 1324), v2);
		if (result) {
			result = strncpy((char*)getMemAt(0x5D4594, 1324), v2, 0xFu);
			*getMemU8Ptr(0x5D4594, 1339) = 0;
			nox_server_gameSettingsUpdated = 1;
		}
	} else {
		*getMemU8Ptr(0x5D4594, 1324) = 0;
		nox_server_gameSettingsUpdated = 1;
	}
	return result;
}

static char* head_reference_nox_xxx_serverOptionsGetServername_40A4C0() { return (char*)getMemAt(0x5D4594, 1324); }

static int head_reference_nox_xxx_ruleSetNoRespawn_40A5E0(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 3584) = a1;
	return result;
}

static int head_reference_nox_server_doPlayersAutoRespawn_40A5F0() {
	return !nox_common_gameFlags_check_40A5C0(4096) ? *getMemU32Ptr(0x5D4594, 3584) : 0;
}

static wchar2_t* head_reference_nox_xxx_sysopSetPass_40A610(wchar2_t* a1) { return nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 3540), a1); }

static wchar2_t* head_reference_nox_xxx_sysopGetPass_40A630() { return (wchar2_t*)getMemAt(0x5D4594, 3540); }

static void head_reference_nox_server_gameSettingsUpdated_40A670() { nox_server_gameSettingsUpdated = 1; }

static int head_reference_nox_server_gameDoSwitchMap_40A680() { return nox_server_gameSettingsUpdated; }

static void head_reference_nox_server_gameUnsetMapLoad_40A690() { nox_server_gameSettingsUpdated = 0; }

static int head_reference_nox_xxx_rateGet_40A6C0() { return *getMemU32Ptr(0x587000, 4728); }

static int head_reference_nox_xxx_rateUpdate_40A6D0(int a1) {
	int result; // eax

	result = *getMemU32Ptr(0x587000, 4728);
	if (a1 != *getMemU32Ptr(0x587000, 4728)) {
		result = nox_common_gameFlags_check_40A5C0(0x20000);
		if (result == 1) {
			result = nox_xxx_netNotifyRate_4D7F10(159);
		}
	}
	*getMemU32Ptr(0x587000, 4728) = a1;
	return result;
}
#undef nox_common_gameFlags_check_40A5C0
#undef nox_xxx_netNotifyRate_4D7F10
#endif
