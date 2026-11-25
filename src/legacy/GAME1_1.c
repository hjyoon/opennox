#include <math.h>
#include <time.h>

// For htonl
#ifdef _WIN32
#include <winsock.h>
#else
#include <arpa/inet.h>
#endif

#include "GAME1.h"
#include "GAME1_1.h"
#include "private_player_records_native.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5_2.h"
#include "common__system__team.h"
#include "server__ability__ability.h"
#include "server__magic__plyrgide.h"
#include "server__magic__plyrspel.h"
#include "server__system__server.h"

#include "client__gui__guiinv.h"
#include "client__gui__guijourn.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__servopts__playrlst.h"
#include "common__binfile.h"
#include "common__crypt.h"
#include "common__log.h"
#include "common__random.h"
#include "operators.h"

#include "client__gui__guibook.h"

#include "client__video__draw_common.h"
#include "common/fs/nox_fs.h"
#include "common__magic__speltree.h"
#include "server__script__builtin.h"
#include "server__script__script.h"

extern uint32_t dword_5d4594_739392;
extern uint32_t dword_5d4594_527988;
extern uint32_t dword_5d4594_600116;
extern uint32_t dword_5d4594_528264;
extern uint32_t dword_5d4594_599496;
extern uint32_t dword_5d4594_528260;
extern uint32_t dword_587000_66116;
extern void* dword_5d4594_588068;
extern uint32_t dword_5d4594_608316;
extern uint32_t dword_5d4594_588120;
extern uint32_t dword_5d4594_528252;
extern uint32_t dword_587000_60044;
extern uint32_t nox_xxx_polygonNextAngle_587000_60356;
extern uint32_t dword_5d4594_528256;
extern uint32_t nox_xxx_polygonNextIdx_587000_60352;
extern uint32_t nox_player_netCode_85319C;
extern obj_5D4594_2650668_t** ptr_5D4594_2650668;
extern int ptr_5D4594_2650668_cap;

//----- (00418C60) --------------------------------------------------------
extern uintptr_t nox_server_team_first_member_native(nox_team_t* team);
nox_object_team_t* nox_xxx_teamCheckSmth_418C60(nox_team_t* team) {
	return (nox_object_team_t*)nox_server_team_first_member_native(team);
}

//----- (00418C70) --------------------------------------------------------
extern uintptr_t nox_server_team_next_member_native(nox_object_team_t* value);
nox_object_team_t* sub_418C70(nox_object_team_t* value) {
	return (nox_object_team_t*)nox_server_team_next_member_native(value);
}

//----- (00419130) --------------------------------------------------------
#if 0 // Restored by team_predicate_native_entry.go.
int nox_xxx_servObjectHasTeam_419130(nox_object_team_t* value) {
	return value && value->id != 0;
}
#endif

//----- (00419150) --------------------------------------------------------
#if 0 // Restored by team_predicate_native_entry.go.
int nox_xxx_servCompareTeams_419150(nox_object_team_t* first, nox_object_team_t* second) {
	return first && second && first->id != 0 && second->id != 0 && first->id == second->id;
}
#endif

//----- (00419180) --------------------------------------------------------

#if 0 // Preserved original for the team report oracle.
//----- (004198A0) --------------------------------------------------------
void sub_4198A0(nox_object_team_t* a1, int a2, int a3) {
	nox_object_t* v3; // eax
	int v4;      // ecx
	char v5[10]; // [esp+8h] [ebp-Ch]

	if (a1) {
		v3 = nox_server_getObjectFromNetCode_4ECCB0(a3);
		if (v3) {
			v5[0] = -60;
			v4 = a1->id;
			v5[1] = 1;
			*(uint16_t*)&v5[6] = a3;
			*(uint32_t*)&v5[2] = v4;
			*(uint16_t*)&v5[8] = v3->typ_ind;
			nox_xxx_netSendPacket1_4E5390(a2, v5, 10, 0, 1);
		}
	}
}
#endif

//----- (0041A590) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width player links.
int sub_41A590(void* a1p, void* a2p) {
	int a1 = a1p;
	int a2 = a2p;
	int v2;    // esi
	int v3;    // edi
	int v4;    // ebx
	size_t v6; // eax
	int v7;    // eax
	int v8;    // ebx
	char* v9;  // eax
	int v10;   // eax
	int v11;   // esi
	float v12; // [esp+0h] [ebp-20h]
	int v13;   // [esp+14h] [ebp-Ch]
	int v14;   // [esp+18h] [ebp-8h]
	int v15;   // [esp+1Ch] [ebp-4h]

	v2 = a2;
	v3 = 0;
	v4 = 0;
	v15 = 0;
	if (!a2) {
		return 0;
	}
	if (a1) {
		v3 = *(uint32_t*)(a1 + 748);
		if (v3) {
			v15 = *(uint32_t*)(v3 + 276);
			v4 = v15;
		}
	}
	a2 = 5;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&a2, 2u);
	if ((short)a2 > 5) {
		return 0;
	}
	if ((short)a2 >= 5) {
		v13 = nox_common_gameFlags_check_40A5C0(4096) ? 4 : 2 - nox_common_gameFlags_check_40A5C0(2048);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v13, 4u);
		if (v4) {
			if (*(uint8_t*)(v4 + 2064) != 31) {
				if (nox_common_gameFlags_check_40A5C0(4096)) {
					if (v13 != 4) {
						nox_xxx_playerCallDisconnect_4DEAB0(*(unsigned char*)(v4 + 2064), 1);
						return 0;
					}
				} else if (v13 != 4) {
					goto LABEL_20;
				}
				if (!nox_common_gameFlags_check_40A5C0(4096)) {
					nox_xxx_playerCallDisconnect_4DEAB0(*(unsigned char*)(v4 + 2064), 1);
					return 0;
				}
			}
		}
	}
LABEL_20:
	LOBYTE(v14) = nox_wcslen((const wchar2_t*)v2);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 1u);
	if ((unsigned char)v14 >= 0x19u) {
		return 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)v2, 2 * (unsigned char)v14);
	*(uint16_t*)(v2 + 2 * (unsigned char)v14) = 0;
	if (v3) {
		nox_wcscpy((wchar2_t*)(*(uint32_t*)(v3 + 276) + 4704), (const wchar2_t*)v2);
	}
	if (nox_crypt_IsReadOnly() == 1 && v3) {
		v6 = nox_wcslen((const wchar2_t*)v2);
		v7 = nox_xxx_protectionStringCRCLen_56FAE0((int*)v2, 2 * v6);
		nox_xxx_playerResetProtectionCRC_56F7D0(*(uint32_t*)(*(uint32_t*)(v3 + 276) + 4628), v7);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 50), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 54), 4u);
	if (nox_crypt_IsReadOnly() == 1 && v3) {
		sub_56F780(*(uint32_t*)(*(uint32_t*)(v3 + 276) + 4624), *(uint32_t*)(v2 + 54));
		sub_56F780(*(uint32_t*)(*(uint32_t*)(v3 + 276) + 4620), *(uint32_t*)(v2 + 50));
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 58), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 62), 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 66), 1u);
	if (nox_crypt_IsReadOnly() == 1 && v3) {
		sub_56F820(*(uint32_t*)(*(uint32_t*)(v3 + 276) + 4616), *(uint8_t*)(v2 + 66));
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 67), 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 68), 3u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 71), 3u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 74), 3u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 77), 3u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 80), 3u);
	if ((short)a2 >= 2) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 83), 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 84), 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 85), 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 86), 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 87), 1u);
	}
	v8 = a1;
	if (nox_crypt_IsReadOnly() == 1) {
		if (a1) {
			v9 = nox_common_playerInfoGetByID_417040(*(uint32_t*)(a1 + 36));
			if (v9) {
				nox_xxx_playerInitColors_461460((nox_playerInfo*)v9);
			}
		}
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 88), 1u);
	if ((short)a2 >= 3) {
		v13 = 0;
		if (v3) {
			v13 = *(uint32_t*)(v3 + 320);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v13, 4u);
		if (nox_crypt_IsReadOnly() == 1 && v3) {
			*(uint32_t*)(v3 + 320) = v13;
		}
		v12 = nox_xxx_gamedataGetFloat_419D40("MaxExtraLives");
		v10 = nox_float2int(v12);
		if (v3 && *(uint32_t*)(v3 + 320) > v10) {
			return 0;
		}
		if ((uint16_t)a2 == 3) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
		}
	}
	if (nox_crypt_IsReadOnly() == 1) {
		sub_4D6000(v8);
	}
	if ((short)a2 >= 4) {
		v11 = v15;
		a1 = 0;
		if (v15) {
			a1 = *(uint32_t*)(v15 + 4696);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
		if (nox_crypt_IsReadOnly() == 1) {
			if (!v11 || (*(uint32_t*)(v11 + 4696) = a1, nox_crypt_IsReadOnly() == 1)) {
				if (v3) {
					sub_4D7450(*(unsigned char*)(*(uint32_t*)(v3 + 276) + 2064),
							   *(uint32_t*)(*(uint32_t*)(v3 + 276) + 4696));
				}
			}
		}
	}
	return 1;
}
#endif



//----- (0041AA30) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width player links.
int sub_41AA30(void* a1p, void* a2p) {
	int a1 = a1p;
	int a2 = a2p;
	int v1;       // esi
	int v2;       // ebp
	char* result; // eax
	int v4;       // eax
	int v5;       // eax
	int v6;       // [esp+10h] [ebp-Ch]
	int v7;       // [esp+14h] [ebp-8h]
	int v8;       // [esp+18h] [ebp-4h]

	v1 = a1;
	v2 = *(uint32_t*)(a1 + 748);
	result = nox_common_playerInfoGetByID_417040(*(uint32_t*)(a1 + 36));
	LOBYTE(a1) = 1;
	if (result) {
		v7 = 2;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v7, 2u);
		if ((short)v7 > 2) {
			return 0;
		}
		if (!nox_common_gameFlags_check_40A5C0(2048)) {
			LOBYTE(a1) = 0;
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 1u);
		if ((uint8_t)a1) {
			result = (char*)nox_common_gameFlags_check_40A5C0(2048);
			if (!result) {
				return result;
			}
			LOWORD(v4) = nox_xxx_unitGetMaxHP_4EE7A0(v1);
			v6 = v4;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v6, 2u);
			if (nox_crypt_IsReadOnly() == 1) {
				nox_xxx_unitSetMaxHP_4EE7C0(v1, v6);
				nox_xxx_unitSetHP_4E4560(v1, v6);
			}
			LOWORD(v5) = nox_xxx_playerGetMaxMana_4EECB0(v1);
			v6 = v5;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v6, 2u);
			if (nox_crypt_IsReadOnly() == 1) {
				nox_xxx_playerSetMaxMana_4EECD0((nox_object_t*)(uintptr_t)(uint32_t)v1, (short)v6);
				nox_xxx_playerManaRefresh_4EECF0((nox_object_t*)(uintptr_t)(uint32_t)v1);
			}
			*getMemU32Ptr(0x5D4594, 527696) = **(unsigned short**)(v1 + 556);
			nox_xxx_fileReadWrite_426AC0_file3_fread(getMemAt(0x5D4594, 527696), 2u);
			*getMemU32Ptr(0x5D4594, 527700) = *(unsigned short*)(v2 + 4);
			nox_xxx_fileReadWrite_426AC0_file3_fread(getMemAt(0x5D4594, 527700), 2u);
			LOBYTE(v8) = *(uint8_t*)(v1 + 540);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v8, 1u);
			if (nox_crypt_IsReadOnly() == 1) {
				nox_xxx_setSomePoisonData_4EEA90(v1, (unsigned char)v8);
			}
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 541), 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 542), 2u);
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 28), 4u);
			if (nox_crypt_IsReadOnly() == 1) {
				sub_56F8C0(*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4604), *(float*)(v1 + 28));
				sub_4D81A0(v1);
			}
			if ((short)v7 >= 2) {
				nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v1 + 124), 2u);
				if (nox_crypt_IsReadOnly() == 1) {
					*(uint16_t*)(v1 + 126) = *(uint16_t*)(v1 + 124);
				}
			}
		}
		result = (char*)1;
	}
	return result;
}
#endif

// 41AAD8: variable 'v4' is possibly undefined
// 41AB12: variable 'v5' is possibly undefined

//----- (0041AC30) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width inventory links.
int sub_41AC30(void* a1p, void* a2p) {
	uint32_t* a1 = a1p;
	int a2 = a2p;
	uint32_t* v1;  // ebp
	int result;    // eax
	int v3;        // eax
	int v4;        // esi
	int v5;        // edi
	int v6;        // edi
	int v7;        // esi
	int v8;        // ebp
	char* v9;      // eax
	int v10;       // ebx
	int* v11;      // ebp
	int v12;       // eax
	int v13;       // esi
	uint8_t* v14;  // eax
	int v15;       // esi
	uint8_t* v16;  // eax
	int v17;       // eax
	int m;         // esi
	int v19;       // eax
	int v20;       // eax
	int v21;       // eax
	int v22;       // esi
	int v23;       // ebx
	uint32_t* v24; // eax
	uint32_t* v25; // esi
	int v26;       // eax
	int v27;       // edi
	int i;         // esi
	uint32_t* v29; // eax
	uint32_t* v30; // eax
	uint32_t* j;   // eax
	uint8_t* v32;  // esi
	size_t v33;    // [esp-4h] [ebp-13Ch]
	size_t v34;    // [esp-4h] [ebp-13Ch]
	char l;        // [esp+12h] [ebp-126h]
	char v36;      // [esp+13h] [ebp-125h]
	int v37;       // [esp+14h] [ebp-124h]
	int k;         // [esp+18h] [ebp-120h]
	int v39;       // [esp+1Ch] [ebp-11Ch]
	int v40;       // [esp+20h] [ebp-118h]
	int v41;       // [esp+24h] [ebp-114h]
	int v42;       // [esp+28h] [ebp-110h]
	int v43;       // [esp+2Ch] [ebp-10Ch]
	int v44;       // [esp+30h] [ebp-108h]
	int v45;       // [esp+34h] [ebp-104h]
	char v46[256]; // [esp+38h] [ebp-100h]

	v1 = a1;
	v36 = 1;
	v40 = a1[187];
	if (!*getMemU32Ptr(0x5D4594, 527704)) {
		*getMemU32Ptr(0x5D4594, 527704) = nox_xxx_getNameId_4E3AA0("Glyph");
	}
	if (nox_crypt_IsReadOnly() == 1) {
		sub_4EF140((nox_object_t*)a1);
	}
	v42 = 3;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v42, 2u);
	if ((short)v42 > 3) {
		return 0;
	}
	if (nox_common_gameFlags_check_40A5C0(0x2000) && !nox_common_gameFlags_check_40A5C0(4096)) {
		v36 = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v36, 1u);
	if (!v36) {
		goto LABEL_115;
	}
	if (!nox_common_gameFlags_check_40A5C0(2048)) {
		result = nox_common_gameFlags_check_40A5C0(4096);
		if (!result) {
			return 0;
		}
	}
	v44 = *(uint32_t*)(*(uint32_t*)(v40 + 276) + 2164);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v44, 4u);
	if (nox_crypt_IsReadOnly() == 1) {
		v3 = nox_xxx_playerGetGold_4FA6B0((int)a1);
		nox_xxx_playerSubGold_4FA5D0((int)a1, v3);
		nox_xxx_playerAddGold_4FA590((int)a1, v44);
	}
	if (nox_crypt_IsReadOnly()) {
		v21 = a1[126];
		if (v21) {
			do {
				v22 = *(uint32_t*)(v21 + 496);
				nox_xxx_delayedDeleteObject_4E5CC0(v21);
				v21 = v22;
			} while (v22);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&k, 4u);
		if (nox_common_gameFlags_check_40A5C0(4096) && k > 2560) {
			return 0;
		}
		v23 = 0;
		if (k > 0) {
			do {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v39, 1u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v46, (unsigned char)v39);
				v46[(unsigned char)v39] = 0;
				v24 = nox_xxx_newObjectByTypeID_4E3810(v46);
				v25 = v24;
				if (!v24) {
					return 0;
				}
				if (!((int (*)(uint32_t*, uint32_t))v24[176])(v24, 0)) {
					return 0;
				}
				v25[14] = 1161297920;
				v25[15] = 1161297920;
				if (nox_common_gameFlags_check_40A5C0(4096)) {
					if (!sub_4F2590((int)v25)) {
						return 0;
					}
				}
				nox_xxx_servMapLoadPlaceObj_4F3F50(
					(nox_object_t*)v25, (nox_object_t*)a1, NULL);
				nox_xxx_unitsNewAddToList_4DAC00();
				if (!nox_xxx_inventoryServPlace_4F36F0((int)a1, (int)v25, 1, 1)) {
					if (!nox_common_gameFlags_check_40A5C0(4096)) {
						return 0;
					}
					nox_xxx_delayedDeleteObject_4E5CC0((int)v25);
				}
				v26 = v25[4];
				if (!(v26 & 0x20) && v26 & 0x100) {
					nox_xxx_playerTryDequip_4F2FB0(a1, (nox_object_t*)v25);
				}
			} while (++v23 < k);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v41, 1u);
		v27 = 0;
		if ((uint8_t)v41) {
			do {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v37, 4u);
				for (i = a1[126]; i; i = *(uint32_t*)(i + 496)) {
					if (*(uint32_t*)(i + 44) == v37) {
						nox_xxx_playerTryEquip_4F2F70(a1, i);
					}
				}
				++v27;
			} while (v27 < (unsigned char)v41);
		}
		if (nox_common_gameFlags_check_40A5C0(2048)) {
			sub_467750(0, 0);
			sub_467740(0);
		}
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v37, 4u);
		if (v37) {
			v29 = (uint32_t*)a1[126];
			if (v29) {
				while (1) {
					if (v29[11] == v37) {
						nox_xxx_netSendSecondaryWeapon_4D9670(*(unsigned char*)(*(uint32_t*)(v40 + 276) + 2064), v29, 0);
						break;
					}
					v29 = (uint32_t*)v29[124];
					if (!v29) {
						break;
					}
				}
			}
		}
		if ((short)v42 >= 2) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v37, 4u);
			if (v37) {
				v30 = (uint32_t*)a1[126];
				if (v30) {
					while (1) {
						if (v30[11] == v37) {
							nox_xxx_netMsgLastQuiver_4D96B0(*(unsigned char*)(*(uint32_t*)(v40 + 276) + 2064), v30);
							break;
						}
						v30 = (uint32_t*)v30[124];
						if (!v30) {
							break;
						}
					}
				}
			}
		}
		if (nox_common_gameFlags_check_40A5C0(4096)) {
			for (j = (uint32_t*)nox_xxx_inventoryGetFirst_4E7980((int)a1); j;
				 j = (uint32_t*)nox_xxx_inventoryGetNext_4E7990((int)j)) {
				j[11] = nox_server_NextObjectScriptID();
				j[10] = j[9];
			}
		}
		goto LABEL_109;
	}
	v4 = a1[126];
	v5 = 0;
	for (k = 0; v4; v4 = *(uint32_t*)(v4 + 496)) {
		if (!nox_common_gameFlags_check_40A5C0(4096) || sub_41B3E0(v4)) {
			++k;
		}
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&k, 4u);
	if (nox_common_gameFlags_check_40A5C0(2048)) {
		v5 = sub_41B3B0();
	}
	if (k != v5 || !nox_common_gameFlags_check_40A5C0(2048)) {
		v15 = a1[126];
		if (v15) {
			while (1) {
				if (!nox_common_gameFlags_check_40A5C0(4096) || sub_41B3E0(v15)) {
					LOBYTE(v39) = strlen((const char*)nox_xxx_getUnitName_4E39D0(v15));
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v39, 1u);
					v34 = (unsigned char)v39;
					v16 = (uint8_t*)nox_xxx_getUnitName_4E39D0(v15);
					nox_xxx_fileReadWrite_426AC0_file3_fread(v16, v34);
					if (!(*(int (**)(int, uint32_t))(v15 + 704))(v15, 0)) {
						return 0;
					}
				}
				v15 = *(uint32_t*)(v15 + 496);
				if (!v15) {
					v1 = a1;
					break;
				}
			}
		}
	} else {
		v6 = 0;
		v41 = 0;
		while (1) {
			v7 = 0;
			v43 = 0;
			do {
				v8 = sub_467810(v7, v6);
				v45 = v8;
				if (v8) {
					v9 = sub_467870(v7, v6);
					v10 = 0;
					if (v8 > 0) {
						v11 = (int*)v9;
						while (1) {
							v12 = nox_xxx_equipedItemByCode_4F7920((int)a1, *v11);
							v13 = v12;
							if (!v12) {
								return 0;
							}
							LOBYTE(v39) = strlen((const char*)nox_xxx_getUnitName_4E39D0(v12));
							nox_xxx_fileReadWrite_426AC0_file3_fread(&v39, 1u);
							v33 = (unsigned char)v39;
							v14 = (uint8_t*)nox_xxx_getUnitName_4E39D0(v13);
							nox_xxx_fileReadWrite_426AC0_file3_fread(v14, v33);
							if (!(*(int (**)(int, uint32_t))(v13 + 704))(v13, 0)) {
								return 0;
							}
							++v10;
							++v11;
							if (v10 >= v45) {
								v6 = v41;
								v7 = v43;
								break;
							}
						}
					}
				}
				++v7;
				v43 = v7;
			} while (v7 < 4);
			++v6;
			v41 = v6;
			if (v6 >= 20) {
				break;
			}
		}
		v1 = a1;
	}
	v17 = v1[126];
	for (l = 0; v17; v17 = *(uint32_t*)(v17 + 496)) {
		if (*(uint32_t*)(v17 + 16) & 0x100) {
			++l;
		}
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&l, 1u);
	for (m = v1[126]; m; m = *(uint32_t*)(m + 496)) {
		if (*(uint32_t*)(m + 16) & 0x100) {
			nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(m + 44), 4u);
		}
	}
	v19 = sub_4678B0();
	v37 = v19;
	if (!v19) {
		if (!m) {
			v37 = 0;
		}
	} else {
		m = v1[126];
		if (m) {
			while (1) {
				if (*(uint32_t*)(m + 36) == v19) {
					v37 = *(uint32_t*)(m + 44);
					if (!m) {
						v37 = 0;
					}
					break;
				}
				m = *(uint32_t*)(m + 496);
				if (!m) {
					v37 = 0;
					break;
				}
			}
		} else {
			v37 = 0;
		}
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v37, 4u);
	v20 = sub_4678C0();
	v37 = v20;
	if (v20) {
		m = v1[126];
		if (!m) {
			v37 = 0;
			goto LABEL_67;
		}
		while (*(uint32_t*)(m + 36) != v20) {
			m = *(uint32_t*)(m + 496);
			if (!m) {
				v37 = 0;
				goto LABEL_67;
			}
		}
		v37 = *(uint32_t*)(m + 44);
	}
	if (!m) {
		v37 = 0;
	}
LABEL_67:
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v37, 4u);
LABEL_109:
	v32 = (uint8_t*)(v40 + 244);
	if ((short)v42 < 3) {
		*(uint8_t*)(v40 + 244) = 0;
	} else {
		nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v40 + 244), 1u);
	}
	if (nox_crypt_IsReadOnly() == 1 && nox_common_gameFlags_check_40A5C0(4096)) {
		*v32 = 0;
	}
LABEL_115:
	if (nox_common_gameFlags_check_40A5C0(4096) && !sub_4F2C30((int)v1)) {
		return 0;
	}
	nox_xxx_netMsgInventoryLoaded_4D96E0(*(unsigned char*)(*(uint32_t*)(v40 + 276) + 2064));
	return 1;
}
#endif

// 41AC30: using guessed type char var_100[256];

//----- (0041B420) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width player links.
int nox_xxx_guiFieldbook_41B420(void* a1p, void* a2p) {
	int a1 = a1p;
	int a2 = a2p;
	char* v1;      // esi
	int result;    // eax
	uint32_t* v3;  // ebp
	uint32_t* v4;  // eax
	int v5;        // ecx
	int i;         // esi
	char* v7;      // eax
	int v8;        // esi
	int v9;        // eax
	int v10;       // eax
	size_t v11;    // [esp-4h] [ebp-124h]
	char v12;      // [esp+13h] [ebp-10Dh]
	int v13;       // [esp+14h] [ebp-10Ch]
	int v14;       // [esp+18h] [ebp-108h]
	int v15;       // [esp+1Ch] [ebp-104h]
	char v16[256]; // [esp+20h] [ebp-100h]

	v1 = nox_common_playerInfoGetByID_417040(*(uint32_t*)(a1 + 36));
	v12 = 1;
	if (!v1) {
		return 0;
	}
	v15 = 1;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v15, 2u);
	if ((short)v15 > 1) {
		return 0;
	}
	if (nox_common_gameFlags_check_40A5C0(0x2000) && !nox_common_gameFlags_check_40A5C0(4096)) {
		v12 = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v12, 1u);
	if (!v12) {
		return 1;
	}
	if (!nox_common_gameFlags_check_40A5C0(2048)) {
		result = nox_common_gameFlags_check_40A5C0(4096);
		if (!result) {
			return result;
		}
	}
	if (nox_crypt_IsReadOnly()) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v13, 1u);
		if ((unsigned char)v13 > 0x29u) {
			return 0;
		}
		v8 = 0;
		if ((unsigned char)v13 > 0u) {
			while (1) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 1u);
				if ((int)(unsigned char)v14 >= 256) {
					break;
				}
				nox_xxx_fileReadWrite_426AC0_file3_fread(v16, (unsigned char)v14);
				v16[(unsigned char)v14] = 0;
				if (nox_common_gameFlags_check_40A5C0(4096)) {
					v9 = nox_xxx_guide_427010(v16);
					if (!sub_4F2EF0(v9)) {
						break;
					}
				}
				v10 = nox_xxx_guide_427010(v16);
				nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide(a1, v10, 0);
				if (++v8 >= (unsigned char)v13) {
					return 1;
				}
			}
			return 0;
		}
		return 1;
	}
	v3 = v1 + 4248;
	LOBYTE(v13) = 0;
	v4 = v1 + 4248;
	v5 = 40;
	do {
		if (*v4) {
			LOBYTE(v13) = v13 + 1;
		}
		++v4;
		--v5;
	} while (v5);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v13, 1u);
	for (i = 1; i < 41; ++i) {
		if (*v3) {
			LOBYTE(v14) = strlen(nox_xxx_guideNameByN_427230(i));
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 1u);
			v11 = (unsigned char)v14;
			v7 = nox_xxx_guideNameByN_427230(i);
			nox_xxx_fileReadWrite_426AC0_file3_fread(v7, v11);
		}
		++v3;
	}
	return 1;
}
#endif

// 41B420: using guessed type char var_100[256];

//----- (0041B660) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width player links.
int nox_xxx_guiSpellbook_41B660(void* a1p, void* a2p) {
	int a1 = a1p;
	int a2 = a2p;
	int result;    // eax
	int v2;        // esi
	uint8_t* v3;   // ebp
	uint32_t* v4;  // eax
	int v5;        // ecx
	int v6;        // ecx
	char v7;       // al
	int v8;        // esi
	char* v9;      // eax
	char* v10;     // eax
	int v11;       // ebp
	bool v12;      // cc
	int v13;       // eax
	int v14;       // eax
	int v15;       // eax
	size_t v16;    // [esp-4h] [ebp-128h]
	size_t v17;    // [esp-4h] [ebp-128h]
	int v18;       // [esp-4h] [ebp-128h]
	char v19;      // [esp+13h] [ebp-111h]
	int v20;       // [esp+14h] [ebp-110h]
	int v21;       // [esp+18h] [ebp-10Ch]
	int v22;       // [esp+1Ch] [ebp-108h]
	int v23;       // [esp+20h] [ebp-104h]
	char v24[256]; // [esp+24h] [ebp-100h]

	result = nox_common_playerInfoGetByID_417040(*(uint32_t*)(a1 + 36));
	v2 = result;
	v19 = 1;
	if (!result) {
		return result;
	}
	v22 = 3;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v22, 2u);
	if ((short)v22 > 3) {
		return 0;
	}
	if (nox_common_gameFlags_check_40A5C0(0x2000) && !nox_common_gameFlags_check_40A5C0(4096)) {
		v19 = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v19, 1u);
	if (!v19) {
		return 1;
	}
	if (!nox_common_gameFlags_check_40A5C0(2048)) {
		result = nox_common_gameFlags_check_40A5C0(4096);
		if (!result) {
			return result;
		}
	}
	if (nox_crypt_IsReadOnly()) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v20, 1u);
		if ((unsigned char)v20 > 0x89u) {
			return 0;
		}
		v11 = 0;
		if ((uint8_t)v20) {
			while (1) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v21, 1u);
				if ((int)(unsigned char)v21 >= 256) {
					break;
				}
				nox_xxx_fileReadWrite_426AC0_file3_fread(v24, (unsigned char)v21);
				v12 = (short)v22 < 2;
				v23 = 3;
				v24[(unsigned char)v21] = 0;
				if (!v12) {
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v23, 4u);
				}
				if (nox_common_gameFlags_check_40A5C0(4096) && v23 > 3 && *(uint8_t*)(v2 + 2251)) {
					break;
				}
				if (nox_common_gameFlags_check_40A5C0(4096)) {
					if (*(uint8_t*)(v2 + 2251)) {
						v13 = nox_xxx_spellNameToN_4243F0(v24);
						if (!nox_xxx_spell_4F2E70(v13)) {
							break;
						}
					}
				}
				if ((short)v22 < 3 || *(uint8_t*)(v2 + 2251)) {
					v18 = v23;
					v15 = nox_xxx_spellNameToN_4243F0(v24);
					nox_xxx_spellGrantToPlayer_4FB550(a1, v15, 0, 0, v18);
				} else {
					v14 = nox_xxx_abilityNameToN_424D80(v24);
					nox_xxx_abilityRewardServ_4FB9C0_ability(a1, v14, 0);
				}
				if (++v11 >= (unsigned char)v20) {
					return 1;
				}
			}
			return 0;
		}
		return 1;
	}
	LOBYTE(v20) = 0;
	v3 = (uint8_t*)(v2 + 3700);
	v4 = (uint32_t*)(v2 + 3700);
	if (*(uint8_t*)(v2 + 2251)) {
		v6 = 136;
		do {
			if (*v4) {
				LOBYTE(v20) = v20 + 1;
			}
			++v4;
			--v6;
		} while (v6);
	} else {
		v5 = 5;
		do {
			if (*v4) {
				LOBYTE(v20) = v20 + 1;
			}
			++v4;
			--v5;
		} while (v5);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v20, 1u);
	v7 = *(uint8_t*)(v2 + 2251);
	v8 = 1;
	if (v7) {
		do {
			if (*(uint32_t*)v3) {
				LOBYTE(v21) = strlen(nox_xxx_spellNameByN_424870(v8));
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v21, 1u);
				v17 = (unsigned char)v21;
				v10 = nox_xxx_spellNameByN_424870(v8);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v10, v17);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v3, 4u);
			}
			++v8;
			v3 += 4;
		} while (v8 < 137);
		result = 1;
	} else {
		do {
			if (*(uint32_t*)v3) {
				LOBYTE(v21) = strlen(nox_xxx_abilityGetName_425250(v8));
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v21, 1u);
				v16 = (unsigned char)v21;
				v9 = nox_xxx_abilityGetName_425250(v8);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v9, v16);
				nox_xxx_fileReadWrite_426AC0_file3_fread(v3, 4u);
			}
			++v8;
			v3 += 4;
		} while (v8 < 6);
		result = 1;
	}
	return result;
}
#endif

// 41B660: using guessed type char var_100[256];

//----- (0041B9C0) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width player links.
void nox_xxx_spellCastByPlayer_4FEEF0();
int nox_xxx_guiEnchantment_41B9C0(void* a1p, void* a2p) {
	uint32_t* a1 = a1p;
	int a2 = a2p;
	int result;    // eax
	int i;         // ebp
	char* v3;      // ebx
	void* v4;      // eax
	int v5;        // ebp
	int v6;        // edi
	int v7;        // ecx
	int v8;        // eax
	short v9;      // ax
	void* v10;     // eax
	int j;         // edi
	int k;         // edi
	int v13;       // [esp-8h] [ebp-140h]
	char v14;      // [esp+Dh] [ebp-12Bh]
	char v15;      // [esp+Eh] [ebp-12Ah]
	char v16;      // [esp+Fh] [ebp-129h]
	int v17;       // [esp+10h] [ebp-128h]
	int v18;       // [esp+14h] [ebp-124h]
	int v19;       // [esp+18h] [ebp-120h]
	int v20;       // [esp+1Ch] [ebp-11Ch]
	int v21;       // [esp+20h] [ebp-118h]
	int v22;       // [esp+24h] [ebp-114h]
	int v23;       // [esp+28h] [ebp-110h]
	int v24[3];    // [esp+2Ch] [ebp-10Ch]
	char v27[256]; // [esp+38h] [ebp-100h]

	v19 = a1[187];
	v16 = 1;
	v18 = 5;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 2u);
	if ((short)v18 > 5) {
		return 0;
	}
	if (nox_crypt_IsReadOnly() == 1 && nox_common_gameFlags_check_40A5C0(2048)) {
		nox_xxx_spellCastByPlayer_4FEEF0();
		sub_4FE8A0(0);
	}
	if (nox_common_gameFlags_check_40A5C0(0x2000)) {
		v16 = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v16, 1u);
	if (v16) {
		result = nox_common_gameFlags_check_40A5C0(2048);
		if (!result) {
			return result;
		}
		if (nox_crypt_IsReadOnly()) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v22, 1u);
			v5 = 0;
			if ((uint8_t)v22) {
				do {
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v20, 1u);
					nox_xxx_fileReadWrite_426AC0_file3_fread(v27, (unsigned char)v20);
					v27[(unsigned char)v20] = 0;
					v6 = nox_xxx_enchantByName_424880(v27);
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v17, 2u);
					LOBYTE(v21) = 2;
					if ((short)v18 >= 2) {
						nox_xxx_fileReadWrite_426AC0_file3_fread(&v21, 1u);
					}
					v7 = a1[15];
					v24[1] = a1[14];
					v13 = (unsigned char)v21;
					v24[0] = a1;
					v24[2] = v7;
					v8 = nox_xxx_getEnchantSpell_424920(v6);
					nox_xxx_spellAccept_4FD400(v8, (int)a1, a1, (int)a1, (int*)&v24, v13);
					v9 = v17;
					if (!(uint16_t)v17) {
						v9 = (uint16_t)gameFPS();
						v17 = gameFPS();
					}
					*((uint16_t*)a1 + v6 + 172) = v9;
					if (v6 == 26 && (short)v18 >= 3) {
						nox_xxx_fileReadWrite_426AC0_file3_fread(&v23, 4u);
						v10 = sub_4FF2D0(51, (nox_object_t*)a1);
						if (v10) {
							*(uint32_t*)((uint8_t*)v10 + 72) = v23;
						}
					}
					++v5;
				} while (v5 < (unsigned char)v22);
			}
		} else {
			LOBYTE(v22) = sub_424CB0((int)a1);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v22, 1u);
			for (i = sub_424D00(); i != -1; i = sub_424D20(i)) {
				if (nox_xxx_testUnitBuffs_4FF350((nox_object_t*)a1, i)) {
					v3 = nox_xxx_getEnchantName_4248F0(i);
					LOBYTE(v20) = strlen(v3);
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v20, 1u);
					nox_xxx_fileReadWrite_426AC0_file3_fread(v3, (unsigned char)v20);
					nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)a1 + 2 * i + 344, 2u);
					LOBYTE(v21) = nox_xxx_buffGetPower_4FF570((nox_object_t*)a1, i);
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v21, 1u);
					if (i == 26) {
						v4 = sub_4FF2D0(51, (nox_object_t*)a1);
						if (v4) {
							v17 = *(uint32_t*)((uint8_t*)v4 + 72);
						} else {
							v17 = 100;
						}
						nox_xxx_fileReadWrite_426AC0_file3_fread(&v17, 4u);
					}
				}
			}
		}
		if ((short)v18 >= 5 && !*(uint8_t*)(*(uint32_t*)(v19 + 276) + 2251)) {
			v14 = nox_common_playerIsAbilityActive_4FC250(a1, 1);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 1u);
			if (nox_crypt_IsReadOnly() == 1 && v14 == 1) {
				sub_4FC670(1);
			}
			v15 = nox_common_playerIsAbilityActive_4FC250(a1, 4);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v15, 1u);
			v23 = sub_4FC030(a1, 4);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v23, 4u);
			if (nox_crypt_IsReadOnly() == 1 && v15 == 1) {
				nox_xxx_playerExecuteAbil_4FBB70((int)a1, 4);
				sub_4FC070(a1, 4, v23);
			}
			for (j = 2 - (v14 != 1); j < 6; ++j) {
				v17 = sub_4FBE60(a1, j);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v17, 4u);
				if (nox_crypt_IsReadOnly() == 1) {
					sub_4FBEA0(a1, j, v17);
					if (v17) {
						nox_xxx_netAbilRepotState_4D8100((int)a1, j, 0);
					}
				}
			}
		}
	}
	if ((uint16_t)v18 == 4 && !*(uint8_t*)(*(uint32_t*)(v19 + 276) + 2251)) {
		v14 = nox_common_playerIsAbilityActive_4FC250(a1, 1);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 1u);
		if (nox_crypt_IsReadOnly() == 1 && v14 == 1) {
			sub_4FC670(1);
		}
		v15 = nox_common_playerIsAbilityActive_4FC250(a1, 4);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v15, 1u);
		v19 = sub_4FC030(a1, 4);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v19, 4u);
		if (nox_crypt_IsReadOnly() == 1 && v15 == 1) {
			nox_xxx_playerExecuteAbil_4FBB70((int)a1, 4);
			sub_4FC070(a1, 4, v19);
		}
		for (k = 2 - (v14 != 1); k < 6; ++k) {
			v17 = sub_4FBE60(a1, k);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v17, 4u);
			if (nox_crypt_IsReadOnly() == 1) {
				sub_4FBEA0(a1, k, v17);
				if (v17) {
					nox_xxx_netAbilRepotState_4D8100((int)a1, k, 0);
				}
			}
		}
	}
	return 1;
}
#endif

// 41B9C0: using guessed type char var_100[256];

//----- (0041BEC0) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width journal links.
int sub_41BEC0(void* a1p, void* a2p) {
	int a1 = a1p;
	int a2 = a2p;
	int v1;       // edi
	int v2;       // esi
	int result;   // eax
	int i;        // eax
	int v5;       // esi
	int v6;       // esi
	int j;        // eax
	int v8;       // [esp+Ch] [ebp-50h]
	int v9;       // [esp+10h] [ebp-4Ch]
	int v10;      // [esp+14h] [ebp-48h]
	int v11;      // [esp+18h] [ebp-44h]
	char v12[64]; // [esp+1Ch] [ebp-40h]

	v1 = a1;
	v2 = *(uint32_t*)(*(uint32_t*)(a1 + 748) + 276);
	LOBYTE(a1) = 1;
	v10 = 1;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v10, 2u);
	if ((short)v10 > 1) {
		return 0;
	}
	if (nox_common_gameFlags_check_40A5C0(0x2000)) {
		LOBYTE(a1) = 0;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 1u);
	if (!(uint8_t)a1) {
		return 1;
	}
	result = nox_common_gameFlags_check_40A5C0(2048);
	if (!result) {
		return result;
	}
	v8 = 0;
	for (i = *(uint32_t*)(v2 + 3644); i; i = *(uint32_t*)(i + 64)) {
		++v8;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v8, 2u);
	if (nox_crypt_IsReadOnly() != 1) {
		if ((unsigned short)v8 > 0u) {
			v6 = *(uint32_t*)(v2 + 3644);
			for (j = *(uint32_t*)(v6 + 64); j; j = *(uint32_t*)(j + 64)) {
				v6 = j;
			}
			for (; v6; v6 = *(uint32_t*)(v6 + 68)) {
				LOBYTE(v9) = strlen((const char*)v6);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v9, 1u);
				nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)v6, (unsigned char)v9);
				nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v6 + 72), 2u);
			}
		}
		return 1;
	}
	sub_4277B0(v1, 0xFFFFu);
	v5 = 0;
	if ((unsigned short)v8 <= 0u) {
		return 1;
	}
	do {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v9, 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(v12, (unsigned char)v9);
		v12[(unsigned char)v9] = 0;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v11, 2u);
		nox_xxx_comJournalEntryAdd_427500(v1, v12, v11);
		++v5;
	} while (v5 < (unsigned short)v8);
	return 1;
}
#endif

// 41BEC0: using guessed type char var_40[64];

//----- (0041C080) --------------------------------------------------------
#if 0 // Restored by player_save_native.go with native-width player links.
int sub_5000B0(nox_object_t* a1);
int sub_41C080(void* a1p, void* a2p) {
	int a1 = a1p;
	int a2 = a2p;
	uint32_t* v1;    // ebp
	int v2;          // ebx
	int result;      // eax
	unsigned int v4; // [esp+8h] [ebp-Ch]
	int v5;          // [esp+Ch] [ebp-8h]
	char v6[4];      // [esp+10h] [ebp-4h]

	v1 = (uint32_t*)a1;
	v2 = *(uint32_t*)(*(uint32_t*)(a1 + 748) + 276);
	if (nox_common_gameFlags_check_40A5C0(0x2000)) {
		return 1;
	}
	a1 = 5;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 2u);
	if ((short)a1 > 5) {
		return 0;
	}
	if ((short)a1 >= 5) {
		if (nox_crypt_IsReadOnly()) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(v6, 4u);
		} else {
			int oid = nox_server_LastObjectScriptID();
			nox_xxx_fileReadWrite_426AC0_file3_fread(&oid, 4u);
			nox_server_SetLastObjectScriptID(oid);
		}
	}
	if (!nox_crypt_IsReadOnly()) {
		strcpy((char*)(v2 + 4760), nox_xxx_mapGetMapName_409B40());
	}
	v4 = strlen((const char*)(v2 + 4760));
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v4, 2u);
	nox_xxx_fileReadWrite_426AC0_file3_fread((uint8_t*)(v2 + 4760), 2 * (unsigned short)v4);
	*(uint8_t*)((unsigned short)v4 + v2 + 4760) = 0;
	if ((short)a1 < 2 || (nox_crypt_IsReadOnly() ? (result = sub_500B70()) : (result = sub_500A60()), result)) {
		if ((short)a1 < 3 || (result = sub_5000B0(v1)) != 0) {
			if ((short)a1 >= 4) {
				LOBYTE(v5) = sub_450750();
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v5, 1u);
				sub_450760(v5);
				return 1;
			}
			sub_450760(0);
			return 1;
		}
	}
	return result;
}
#endif


//----- (0041C280) --------------------------------------------------------
// Related to placing traps??
// 41CF42: variable 'v5' is possibly undefined
// 41CF51: variable 'v7' is possibly undefined

//----- (0041E2F0) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_41E2F0() { return dword_5d4594_527988; }
#endif

//----- (0041FA40) --------------------------------------------------------
#if 0 // Restored by global_record_native_entry.go.
char* sub_41FA40() { return (char*)getMemAt(0x5D4594, 534756); }
#endif

//----- (0041FA50) --------------------------------------------------------
#if 0 // Restored by global_record_native_entry.go.
void sub_41FA50(const char* a1) {
	if (a1) {
		strcpy((char*)getMemAt(0x5D4594, 534756), a1);
	}
}
#endif

//----- (0041FBE0) --------------------------------------------------------
#if 0 // Restored by global_record_native_entry.go.
int sub_41FBE0(uint32_t* a1, uint32_t* a2) {
	int result; // eax

	if (*(int*)&dword_587000_60044 == -1) {
		return 0;
	}
	*a1 = getMemAt(0x5D4594, 531660 + 24 * dword_587000_60044);
	result = 1;
	*a2 = getMemAt(0x5D4594, 531670 + 24 * dword_587000_60044);
	return result;
}
#endif

//----- (004200E0) --------------------------------------------------------
#if 0 // Restored by global_record_native_entry.go.
int sub_4200E0() { return *getMemU32Ptr(0x587000, 60072); }
#endif

//----- (00420100) --------------------------------------------------------
#if 0 // Restored by global_record_native_entry.go.
int sub_420100() { return *getMemU32Ptr(0x587000, 60072) >> 8; }
#endif

//----- (00420C40) --------------------------------------------------------


//----- (00421660) --------------------------------------------------------
// 4217F5: variable 'v2' is possibly undefined
// 421840: variable 'v6' is possibly undefined

// 421F64: variable 'v3' is possibly undefined
// 421FBE: variable 'v7' is possibly undefined

//----- (00422140) --------------------------------------------------------
#if 0 // Native pointers and field layout: player_marker_movable_native_exports.go:sub_422140
int sub_422140(int a1) {
	int result; // eax

	result = a1;
	if (a1) {
		*(uint32_t*)(a1 + 3660) = -559023410;
		*(uint32_t*)(a1 + 3664) = -559023410;
	}
	return result;
}
#endif


//----- (00423930) --------------------------------------------------------
#if 0 // Native named bitmasks: named_bitmask_native_exports.go
void set_bitmask_flags_from_plus_separated_names_423930(const char* input, uint32_t* bitmask,
														const char** allowed_names) {
	char input_copy[256];
	strcpy(input_copy, input);
	if (strncmp(input_copy, "NULL", 4u) == 0) {
		return;
	}

	char* cur_value = strtok(input_copy, "+");
	while (cur_value) {
		set_one_bitmask_flag_by_name_4239C0(cur_value, bitmask, allowed_names);
		cur_value = strtok(0, "+");
	}
}
#endif


//----- (004239C0) --------------------------------------------------------
#if 0 // Native named bitmasks: named_bitmask_native_exports.go
int set_one_bitmask_flag_by_name_4239C0(char* name, uint32_t* bitmask, const char** allowed_names) {
	for (char i = 0; allowed_names[i]; ++i) {
		if (!nox_strcmpi(allowed_names[i], name)) {
			*bitmask |= 1 << i;
			return 1;
		}
	}

	return 0;
}
#endif


//----- (004240F0) --------------------------------------------------------
#if 0 // Native sound-set parsing: soundset_parser_native_exports.go
int sub_4240F0(int a1, const char* a2, int a3) {
	const char* v3;    // esi
	int v4;            // ebp
	unsigned char* v5; // edi

	v3 = *(const char**)getMemAt(0x587000, 64704);
	v4 = 0;
	if (!*getMemU32Ptr(0x587000, 64704)) {
		return 0;
	}
	v5 = getMemAt(0x587000, 64704);
	while (strcmp(a2, v3)) {
		v3 = (const char*)*((uint32_t*)v5 + 2);
		v5 += 8;
		++v4;
		if (!v3) {
			return 0;
		}
	}
	*(uint32_t*)(*getMemU32Ptr(0x587000, 64708 + 8 * v4) + a1) = a3;
	return 1;
}
#endif


//----- (00424170) --------------------------------------------------------
#if 0 // Native sound-set parsing: soundset_parser_native_exports.go
int nox_xxx_parseSoundSetBin_424170(char* a1) {
	FILE* v2;     // ebp
	uint32_t* v3; // ebx
	char* v4;     // edx
	int v5;       // eax
	char v6[256]; // [esp+10h] [ebp-200h]
	char v7[256]; // [esp+110h] [ebp-100h]

	v2 = nox_binfile_open_408CC0(a1, 0);
	if (!v2) {
		return 0;
	}
	if (!nox_binfile_cryptSet_408D40(v2, 5)) {
		return 0;
	}
	while (nox_xxx_parseString_409470(v2, (uint8_t*)v6)) {
		v3 = calloc(1u, 0x54u);
		v3[19] = dword_5d4594_588120;
		v3[20] = 0;
		dword_5d4594_588120 = v3;
		v4 = (char*)calloc(strlen(v6) + 1, 1);
		*v3 = v4;
		strcpy(v4, v6);
		while (nox_xxx_parseString_409470(v2, (uint8_t*)v6) && strcmp(v6, "END") && nox_xxx_parseString_409470(v2, (uint8_t*)v7)) {
			v5 = nox_xxx_utilFindSound_40AF50(v7);
			if (!sub_4240F0((int)v3, v6, v5)) {
				return 0;
			}
		}
	}
	nox_binfile_close_408D90(v2);
	return 1;
}
#endif


//----- (00424320) --------------------------------------------------------
#if 0 // Native NPC voice: npc_voice_native_exports.go
int nox_xxx_setNPCVoiceSet_424320(int a1, int a2) {
	int result; // eax

	result = a1;
	if (a1) {
		if (*(uint8_t*)(a1 + 8) & 2) {
			*(uint32_t*)(*(uint32_t*)(a1 + 748) + 488) = a2;
			result = 1;
		} else {
			result = 0;
		}
	}
	return result;
}
#endif


//----- (00424350) --------------------------------------------------------
#if 0 // Native NPC voice: npc_voice_native_exports.go
const char** nox_xxx_getDefaultSoundSet_424350(const char* a1) {
	int v1; // edi

	if (!a1) {
		return 0;
	}
	v1 = dword_5d4594_588120;
	if (!dword_5d4594_588120) {
		return 0;
	}
	while (strcmp(*(const char**)v1, a1)) {
		v1 = *(uint32_t*)(v1 + 76);
		if (!v1) {
			return 0;
		}
	}
	return (const char**)v1;
}
#endif


//----- (00424CB0) --------------------------------------------------------
#if 0 // Native buff count/iteration: buff_iteration_native_exports.go
char sub_424CB0(int a1) {
	char v1;           // bl
	int v2;            // esi
	unsigned char* v3; // edi

	v1 = 0;
	v2 = 0;
	if (*(int*)&dword_587000_66116 <= 0) {
		return 0;
	}
	v3 = getMemAt(0x587000, 66000);
	do {
		if (nox_xxx_testUnitBuffs_4FF350(a1, *(uint32_t*)v3)) {
			++v1;
		}
		++v2;
		v3 += 4;
	} while (v2 < *(int*)&dword_587000_66116);
	return v1;
}
#endif


//----- (00424D00) --------------------------------------------------------
#if 0 // Native buff count/iteration: buff_iteration_native_exports.go
int sub_424D00() {
	int result; // eax

	if (*(int*)&dword_587000_66116 <= 0) {
		result = -1;
	} else {
		result = *getMemU32Ptr(0x587000, 66000);
	}
	return result;
}
#endif


//----- (00424D20) --------------------------------------------------------
#if 0 // Native buff count/iteration: buff_iteration_native_exports.go
int sub_424D20(int a1) {
	int v1;           // eax
	unsigned char* i; // edx

	v1 = 0;
	if (*(int*)&dword_587000_66116 <= 0) {
		return -1;
	}
	for (i = getMemAt(0x587000, 66000); *(int*)i != a1 || v1 >= *(int*)&dword_587000_66116 - 1; i += 4) {
		if (++v1 >= *(int*)&dword_587000_66116) {
			return -1;
		}
	}
	return *getMemU32Ptr(0x587000, 66004 + 4 * v1);
}
#endif


// 424FE7: variable 'v34' is possibly undefined
// 424F00: using guessed type char var_100[256];

//----- (004254A0) --------------------------------------------------------
#if 0 // Native bit cursor: bitstream_exports_4254a0.go:sub_4254A0
int sub_4254A0(int a1, uint8_t* a2) {
	*(uint32_t*)a1 = a2;
	*(uint8_t*)(a1 + 4) = 0;
	return *a2 & 1;
}
#endif


//----- (004254C0) --------------------------------------------------------
#if 0 // Native bit cursor: bitstream_exports_4254a0.go:sub_4254C0
bool sub_4254C0(unsigned char** a1) {
	char v1;           // cl
	unsigned char* v2; // ecx

	v1 = *((uint8_t*)a1 + 4) + 1;
	*((uint8_t*)a1 + 4) = v1;
	if (v1 == 8) {
		v2 = *a1;
		*((uint8_t*)a1 + 4) = 0;
		*a1 = v2 + 1;
	}
	return ((1 << *((uint8_t*)a1 + 4)) & **a1) > 0;
}
#endif


//----- (00425500) --------------------------------------------------------
#if 0 // Native bit cursor: bitstream_exports_4254a0.go:sub_425500
uint8_t* sub_425500(int a1, uint8_t* a2, char a3) {
	uint8_t* result; // eax

	result = a2;
	*(uint32_t*)a1 = a2;
	*(uint8_t*)(a1 + 4) = 0;
	*a2 = a3;
	return result;
}
#endif


//----- (00425520) --------------------------------------------------------
#if 0 // Native bit cursor: bitstream_exports_4254a0.go:sub_425520
char sub_425520(int a1, char a2) {
	char v2;     // cl
	uint8_t* v3; // ecx
	char result; // al

	v2 = *(uint8_t*)(a1 + 4) + 1;
	*(uint8_t*)(a1 + 4) = v2;
	if (v2 == 8) {
		v3 = *(uint8_t**)a1;
		*(uint8_t*)(a1 + 4) = 0;
		*(uint32_t*)a1 = ++v3;
		*v3 = 0;
	}
	result = a2 << *(uint8_t*)(a1 + 4);
	**(uint8_t**)a1 |= result;
	return result;
}
#endif


//----- (00425550) --------------------------------------------------------
#if 0 // Native bit cursor: bitstream_exports_4254a0.go:sub_425550
int sub_425550(uint8_t* a1, uint8_t* a2, int a3) {
	int v3;     // edi
	int v4;     // esi
	char v5;    // al
	int v6;     // ebx
	char v8[8]; // [esp+8h] [ebp-10h]
	char v9[8]; // [esp+10h] [ebp-8h]
	bool v10;   // [esp+1Ch] [ebp+4h]

	v3 = 1;
	v4 = 0;
	v5 = sub_4254A0((int)v9, a1);
	sub_425500((int)v8, a2, v5);
	if (a3 == 1) {
		return 1;
	}
	v6 = a3 - 1;
	do {
		if (!(++v4 % 7u)) {
			sub_425520((int)v8, 1);
			++v3;
		}
		v10 = sub_4254C0((unsigned char**)v9);
		sub_425520((int)v8, v10);
		--v6;
	} while (v6);
	return v3;
}
#endif


//----- (00425760) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
void nox_common_list_clear_425760(nox_list_item_t* list) {
	list->field_0 = list;
	list->field_1 = list;
	list->field_2 = list;
}
#endif


//----- (00425770) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
void* sub_425770(void* a1p) {
	nox_list_item_t* a1 = a1p;

	a1->field_0 = a1;
	a1->field_1 = a1;
	a1->field_2 = NULL;
	return a1;
}
#endif


//----- (00425790) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
int sub_425790(int* a1, uint32_t* a2) {
	int v2;     // esi
	int v3;     // ebx
	int* v4;    // eax

	v2 = 0;
	v3 = a2[2];
	v4 = (int*)nox_common_list_getNext_425940(a1);
	if (!v4) {
		nox_common_list_append_4258E0((int)a1, a2);
		return v2;
	}
	while (v3 > v4[2]) {
		++v2;
		v4 = nox_common_list_getNext_425940(v4);
		if (!v4) {
			nox_common_list_append_4258E0((int)a1, a2);
			return v2;
		}
	}
	nox_common_list_append_4258E0((int)v4, a2);
	return v2;
}
#endif


//----- (004257F0) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
void sub_4257F0(int* a1, uint32_t* a2) {
	int v2;  // esi
	int* v3; // eax

	v2 = a2[2];
	v3 = (int*)nox_common_list_getNext_425940(a1);
	if (!v3) {
		nox_common_list_append_4258E0((int)a1, a2);
		return;
	}
	while (v3[2] > v2) {
		v3 = (int*)nox_common_list_getNext_425940(v3);
		if (!v3) {
			nox_common_list_append_4258E0((int)a1, a2);
			return;
		}
	}
	nox_common_list_append_4258E0((int)v3, a2);
}
#endif


//----- (00425890) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
nox_list_item_t* nox_common_list_getFirstSafe_425890(nox_list_item_t* list) {
	return nox_common_list_getNextSafe_4258A0(list);
}
#endif


//----- (004258A0) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
nox_list_item_t* nox_common_list_getNextSafe_4258A0(nox_list_item_t* list) {
	if (!list) {
		return 0;
	}
	return nox_common_list_getNext_425940(list);
}
#endif


//----- (004258C0) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
uint32_t* sub_4258C0(uint32_t** a1, int a2) {
	uint32_t* result; // eax
	int v3;           // edx
	int v4;           // esi

	result = *a1;
	if ((void*)*a1 == (void*)a1) {
		return 0;
	}
	v3 = a2;
	while (1) {
		v4 = v3--;
		if (!v4) {
			break;
		}
		result = (uint32_t*)*result;
		if ((void*)result == (void*)a1) {
			return 0;
		}
	}
	return result;
}
#endif


//----- (004258E0) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
void nox_common_list_append_4258E0(nox_list_item_t* list, nox_list_item_t* cur) {
	if (!list || !cur) {
		abort();
	}
	nox_list_item_t* it = list->field_1;

	// FIXME: in some cases 'it' is null, which suggests that some lists are not initialized properly
	//        auto-initializing them however leads to more serious issues like double-free

	//	if (!it && !list->field_0 && !list->field_2) {
	//		nox_common_list_clear_425760(list);
	//		it = list->field_1;
	//	}

	cur->field_0 = list;
	cur->field_1 = it;

	list->field_1 = cur;
	if (it) { // see above note
		it->field_0 = cur;
	}
}
#endif


//----- (00425900) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
uint32_t* sub_425900(uint32_t* a1, uint32_t* a2) {
	nox_list_item_t* list = (nox_list_item_t*)a1;
	nox_list_item_t* cur = (nox_list_item_t*)a2;
	cur->field_1 = list;
	cur->field_0 = list->field_0;
	list->field_0 = cur;
	cur->field_0->field_1 = cur;
	return (uint32_t*)cur;
}
#endif


//----- (00425920) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
void nox_common_list_remove_425920(void* a1p) {
	nox_list_item_t* cur = a1p;
	if (!cur || !cur->field_0 || !cur->field_1) {
		return;
	}
	cur->field_1->field_0 = cur->field_0;
	cur->field_0->field_1 = cur->field_1;
	cur->field_0 = cur;
	cur->field_1 = cur;
}
#endif


//----- (00425940) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
nox_list_item_t* nox_common_list_getNext_425940(nox_list_item_t* list) {
	nox_list_item_t* it = list->field_0;
	if (it && it == it->field_2) {
		return 0;
	}
	return it;
}
#endif


//----- (00425960) --------------------------------------------------------
#if 0 // Native intrusive list: intrusive_list_native_exports.go
int sub_425960(int a1) {
	return *(uint32_t*)(*(uint32_t*)(a1 + 4) + 8) != *(uint32_t*)(a1 + 4) ? *(uint32_t*)(a1 + 4) : 0;
}
#endif


//----- (004259C0) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
void sub_4259C0() {
	if (!*getMemU32Ptr(0x5D4594, 599472)) {
		nox_common_list_clear_425760(getMemAt(0x5D4594, 599460));
		*getMemU32Ptr(0x5D4594, 599472) = 1;
	}
}
#endif

//----- (004259F0) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
int* sub_4259F0() {
	int* result; // eax
	int* v1;     // ebx
	int* v2;     // ebp
	int* v3;     // esi
	int* v4;     // edi

	result = sub_425A50();
	v1 = result;
	if (result) {
		do {
			v2 = sub_425A60(v1);
			v3 = sub_425BC0(v1);
			if (v3) {
				do {
					v4 = sub_425BE0(v3);
					nox_common_list_remove_425920((uint32_t**)v3);
					free(v3);
					v3 = v4;
				} while (v4);
			}
			nox_common_list_remove_425920((uint32_t**)v1);
			free(v1);
			v1 = v2;
		} while (v2);
	}
	return result;
}
#endif

//----- (00425A50) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
int* sub_425A50() { return nox_common_list_getFirstSafe_425890(getMemIntPtr(0x5D4594, 599460)); }
#endif

//----- (00425A60) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
int* sub_425A60(int* a1) { return nox_common_list_getNextSafe_4258A0(a1); }
#endif

//----- (00425A70) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
int* sub_425A70(int a1) {
	int* result; // eax

	result = sub_425A50();
	if (!result) {
		return 0;
	}
	while (result[8] != a1) {
		result = sub_425A60(result);
		if (!result) {
			return 0;
		}
	}
	return result;
}
#endif

//----- (00425AA0) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
int sub_425AA0(int a1) {
	int* v1; // eax

	v1 = sub_425A50();
	if (!v1) {
		return 0;
	}
	while (v1[8] != a1) {
		v1 = sub_425A60(v1);
		if (!v1) {
			return 0;
		}
	}
	return 1;
}
#endif

//----- (00425AD0) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
wchar2_t* sub_425AD0(int a1, wchar2_t* a2) {
	wchar2_t* v2; // esi

	v2 = 0;
	if (!sub_425AA0(a1)) {
		v2 = (wchar2_t*)calloc(1, 0x34u);
		*((uint32_t*)v2 + 8) = a1;
		nox_wcscpy(v2 + 6, a2);
		*((uint32_t*)v2 + 9) = 0;
		sub_425770(v2);
		nox_common_list_clear_425760((uint32_t*)v2 + 10);
		nox_common_list_append_4258E0((int)getMemAt(0x5D4594, 599460), v2);
	}
	return v2;
}
#endif

//----- (00425B30) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
void sub_425B30(void* a1, int a2) {
	uint32_t* v2; // eax
	uint32_t* v3; // esi

	v2 = calloc(1, 0x10u);
	v3 = v2;
	v2[3] = a2;
	sub_425770(v2);
	nox_common_list_append_4258E0((int)a1 + 40, v3);
}
#endif

//----- (00425B60) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
char* sub_425B60(void* lpMem, int a2) {
	int* v2;      // esi
	char* result; // eax

	v2 = sub_425BC0(lpMem);
	if (v2) {
		while (v2[3] != a2) {
			v2 = sub_425BE0(v2);
			if (!v2) {
				goto LABEL_6;
			}
		}
		nox_common_list_remove_425920((uint32_t**)v2);
		free(v2);
	}
LABEL_6:
	result = (char*)lpMem + 40;
	if (*((void**)lpMem + 11) == (char*)lpMem + 40) {
		nox_common_list_remove_425920((uint32_t**)lpMem);
		free(lpMem);
	}
	return result;
}
#endif

//----- (00425BC0) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
int* sub_425BC0(void* a1) {
	// A record's embedded children header is a required member, not an optional list pointer.
	return (int*)nox_common_list_getNext_425940((nox_list_item_t*)((char*)a1 + 40));
}
#endif

//----- (00425BE0) --------------------------------------------------------
#if 0 // Native private player records: private_player_records_native_exports.go
int* sub_425BE0(int* a1) { return nox_common_list_getNextSafe_4258A0(a1); }
#endif

//----- (00425CA0) --------------------------------------------------------
#if 0 // Replaced by server.PlayerKillStats425CA0.
char* sub_425CA0(int a1, int a2) {
	char* result; // eax
	int v3;       // ebx
	int v4;       // ebp
	int v5;       // ebp
	int v6;       // eax
	int v7;       // eax
	int v8;       // esi
	int v9;       // eax
	int v10;      // eax
	int v11;      // eax
	int v12;      // [esp+8h] [ebp+4h]
	int v13;      // [esp+Ch] [ebp+8h]

	result = (char*)nox_common_gameFlags_check_40A5C0(0x2000);
	if (result) {
		result = (char*)nox_common_gameFlags_check_40A5C0(4096);
		if (!result) {
			if (a1) {
				v3 = a2;
				if (*(int*)(a1 + 4648) == -1) {
					v4 = (dword_5d4594_608316)++;
					v13 = v4;
					v5 = 32 * v4;
					strcpy((char*)getMemAt(0x5D4594, 600124 + v5), (const char*)(a1 + 2096));
					if (*(uint8_t*)(v3 + 2064) == 31) {
						v6 = nox_xxx_net_getIP_554200(0);
					} else {
						v6 = nox_xxx_net_getIP_554200(*(unsigned char*)(a1 + 2064) + 1);
					}
					*getMemU32Ptr(0x5D4594, 600136 + v5) = htonl(v6);
					*getMemU32Ptr(0x5D4594, 600140 + v5) = *(uint32_t*)(a1 + 2068);
					*getMemU8Ptr(0x5D4594, 600144 + v5) = *(uint8_t*)(a1 + 2251);
					*(uint32_t*)(a1 + 4648) = v13;
				} else {
					v13 = *(uint32_t*)(a1 + 4648);
				}
				if (v3) {
					if (*(int*)(v3 + 4648) == -1) {
						v12 = (dword_5d4594_608316)++;
						strcpy((char*)getMemAt(0x5D4594, 600124 + 32 * v13), (const char*)(v3 + 2096));
						if (*(uint8_t*)(v3 + 2064) == 31) {
							v7 = nox_xxx_net_getIP_554200(0);
							v8 = v12;
							*getMemU32Ptr(0x5D4594, 600136 + 32 * v13) = htonl(v7);
						} else {
							v9 = nox_xxx_net_getIP_554200(*(unsigned char*)(v3 + 2064) + 1);
							v8 = v12;
							*getMemU32Ptr(0x5D4594, 600136 + 32 * v12) = htonl(v9);
						}
						v10 = 32 * v8;
						*getMemU32Ptr(0x5D4594, 600140 + v10) = *(uint32_t*)(v3 + 2068);
						*getMemU8Ptr(0x5D4594, 600144 + v10) = *(uint8_t*)(v3 + 2251);
						*(uint32_t*)(v3 + 4648) = v8;
					} else {
						v12 = *(uint32_t*)(v3 + 4648);
					}
				} else {
					LOBYTE(v12) = -1;
				}
				v11 = dword_5d4594_739392;
				*getMemU8Ptr(0x5D4594, 608320 + 2 * dword_5d4594_739392) = v13;
				*getMemU8Ptr(0x5D4594, 608321 + 2 * v11) = v12;
				dword_5d4594_739392 = v11 + 1;
				result = *(char**)&dword_5d4594_608316;
				if (dword_5d4594_608316 >= 0xFFu) {
					result = nox_xxx_net_4263C0();
				}
			}
		}
	}
	return result;
}
#endif

//----- (00427490) --------------------------------------------------------
#if 0 // Restored by journal_native_427490.go with native-width list ownership.
nox_playerInfo_journal* nox_xxx_journalEntryAdd_427490(nox_playerInfo* a1p, char* a2, short a3) {
	int a1 = a1p;
	uint8_t* result; // eax
	uint8_t* v4;     // esi
	int v5;          // eax

	result = calloc(1, 0x4Cu);
	v4 = result;
	if (result) {
		memset(result, 0, 0x4Cu);
		strncpy(result, a2, 0x3Fu);
		v4[63] = 0;
		*((uint16_t*)v4 + 36) = a3;
		*((uint32_t*)v4 + 17) = 0;
		*((uint32_t*)v4 + 16) = *(uint32_t*)(a1 + 3644);
		v5 = *(uint32_t*)(a1 + 3644);
		if (v5) {
			*(uint32_t*)(v5 + 68) = v4;
		}
		*(uint32_t*)(a1 + 3644) = v4;
		result = v4;
	}
	return result;
}

//----- (00427500) --------------------------------------------------------
void nox_xxx_comJournalEntryAdd_427500(nox_object_t* a1p, char* a2, short a3) {
	int a1 = a1p;
	int v3;      // esi
	uint8_t* v4; // eax

	v3 = *(uint32_t*)(*(uint32_t*)(a1 + 748) + 276);
	v4 = nox_xxx_journalEntryAdd_427490(v3, a2, a3);
	if (v4) {
		if (*(uint8_t*)(v3 + 2064) == 31) {
			nox_xxx_cliBuildJournalString_469BC0();
		} else {
			nox_xxx_netSendJournalAdd_4D9440(*(unsigned char*)(v3 + 2064), (int)v4);
		}
	}
}

//----- (00427590) --------------------------------------------------------
int nox_xxx_journalEntryRemove_427590(nox_playerInfo* a1p, const char* a2) {
	int a1 = a1p;
	int v2; // edi
	int v4; // eax
	int v5; // eax

	v2 = *(uint32_t*)(a1 + 3644);
	if (!v2) {
		return 0;
	}
	while (strcmp((const char*)v2, a2)) {
		v2 = *(uint32_t*)(v2 + 64);
		if (!v2) {
			return 0;
		}
	}
	v4 = *(uint32_t*)(v2 + 68);
	if (v4) {
		*(uint32_t*)(v4 + 64) = *(uint32_t*)(v2 + 64);
	}
	v5 = *(uint32_t*)(v2 + 64);
	if (v5) {
		*(uint32_t*)(v5 + 68) = *(uint32_t*)(v2 + 68);
	}
	if (v2 == *(uint32_t*)(a1 + 3644)) {
		*(uint32_t*)(a1 + 3644) = *(uint32_t*)(v2 + 64);
	}
	free((void*)v2);
	return 1;
}

//----- (004276B0) --------------------------------------------------------
int nox_xxx_journalUpdateEntry_4276B0(nox_playerInfo* a1p, const char* a2, short a3) {
	int a1 = a1p;
	int v3;     // edi
	int result; // eax

	v3 = *(uint32_t*)(a1 + 3644);
	if (!v3) {
		return 0;
	}
	while (strcmp((const char*)v3, a2)) {
		v3 = *(uint32_t*)(v3 + 64);
		if (!v3) {
			return 0;
		}
	}
	result = v3;
	*(uint16_t*)(v3 + 72) = a3;
	return result;
}

//----- (004277B0) --------------------------------------------------------
// This removes player's journal with matching bitmask
int sub_4277B0(nox_object_t* a1p, unsigned short a2) {
	nox_player_update_data_t* update = a1p->data_update;
	nox_playerInfo* player = update->player;
	nox_playerInfo_journal* entry = player->field_3644;

	while (entry) {
		nox_playerInfo_journal* next = entry->next;
		if (a2 & entry->field_3) {
			if (entry->prev) {
				entry->prev->next = entry->next;
			}
			if (entry->next) {
				entry->next->prev = entry->prev;
			}
			if (entry == player->field_3644) {
				player->field_3644 = entry->next;
			}
			free(entry);
		}
		entry = next;
	}
	return 0;
}
#endif

// POINT_SEGMENT_PROJECTION_ORIGINAL427DF0_BEGIN
// The original uses precision53/ToZero, with explicit binary32 stores.
// Int32 geometry bounds all finite intermediates and FMA residuals away
// from binary64 overflow/underflow. Volatile stores prevent contraction.
// POINT_SEGMENT_PROJECTION_ORIGINAL427DF0_END
