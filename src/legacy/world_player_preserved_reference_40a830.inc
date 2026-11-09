int nox_xxx_countNonEliminatedPlayersInTeam_40A830(nox_team_t* a1p) {
	int count = 0;
	for (nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0(); unit;
		 unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit)) {
		if (nox_xxx_teamCompare2_419180((nox_object_team_t*)&unit->field_12, a1p->field_57)) {
			nox_playerInfo* player = ((nox_player_update_data_t*)unit->data_update)->player;
			if (!(player->field_3680 & 1) &&
				(player->playerInd != 31 ||
				 !nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING))) {
				++count;
			}
		}
	}
	return count;
}

int nox_xxx_gamePlayIsAnyPlayers_40A8A0() {
	int count = 0;
	if (!nox_xxx_CheckGameplayFlags_417DA0(4)) {
		for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
			 player = nox_common_playerInfoGetNext_416EE0(player)) {
			if (player->playerUnit) {
				uint32_t flags = player->field_3680;
				if (!(flags & 1) || flags & 0x20) {
					++count;
				}
			}
		}
		return count > 1;
	}
	for (nox_team_t* team = nox_server_teamFirst_418B10(); team; team = nox_server_teamNext_418B60(team)) {
		nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0();
		if (unit) {
			while (unit) {
				if (nox_xxx_teamCompare2_419180((nox_object_team_t*)&unit->field_12, team->field_57)) {
					nox_playerInfo* player = ((nox_player_update_data_t*)unit->data_update)->player;
					uint32_t flags = player->field_3680;
					if (!(flags & 1) || flags & 0x20) {
						break;
					}
				}
				unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit);
			}
			if (unit) {
				++count;
			}
		}
	}
	return count > 1;
}

int nox_xxx_cliCanTalkMB_4100F0(short* a1) {
	short v1;  // ax
	short* v2; // ecx

	v1 = *a1;
	v2 = a1 + 1;
	if (!*a1) {
		return 1;
	}
	while (!HIBYTE(v1)) {
		v1 = *v2;
		++v2;
		if (!v1) {
			return 1;
		}
	}
	return 0;
}

void* nox_xxx_doorAttachWall_410360(nox_object_t* obj, int x, int y) {
	void* wall = nox_xxx_wallCreateAt_410250(x, y);
	return nox_server_wallAttachDoor(wall, obj);
}

nox_secret_wall_t* nox_xxx_wallSecretBlock_410760(nox_secret_wall_t* a1) {
	nox_secret_wall_t* result; // eax

	result = a1;
	a1->next = (nox_secret_wall_t*)dword_5d4594_251560;
	dword_5d4594_251560 = a1;
	return result;
}

nox_secret_wall_t* nox_xxx_wallSecretGetFirstWall_410780() {
	return (nox_secret_wall_t*)dword_5d4594_251560;
}

nox_secret_wall_t* nox_xxx_wallSecretNext_410790(nox_secret_wall_t* a1) {
	return a1 ? a1->next : NULL;
}

int nox_xxx_tileAlloc_410F60_init() {
	ptr_5D4594_2650668 = calloc(ptr_5D4594_2650668_cap, sizeof(void*));
	if (!ptr_5D4594_2650668) {
		return 0;
	}

	for (int i = 0; i < ptr_5D4594_2650668_cap; i++) {
		ptr_5D4594_2650668[i] = (obj_5D4594_2650668_t*)calloc(ptr_5D4594_2650668_cap, sizeof(obj_5D4594_2650668_t));
		if (!ptr_5D4594_2650668[i]) {
			return 0;
		}
	}
	return 1;
}

void nox_xxx_tileFree_410FC0_free() {
	for (int i = 0; i < ptr_5D4594_2650668_cap; i++) {
		obj_5D4594_2650668_t* ptr = ptr_5D4594_2650668[i];
		if (ptr) {
			free(ptr);
		}
	}
}

int nox_xxx_tileNFromPoint_411160(float2* a1) {
	float v12 = (a1->field_0 + 11.5) * 0.021739131;
	float v13 = (a1->field_4 + 11.5) * 0.021739131;

	int i = nox_float2int(v12);
	int j = nox_float2int(v13);

	float v14 = a1->field_0 + 11.5;
	float v15 = a1->field_4 + 11.5;

	int v4 = nox_float2int(v14) % 46;
	int v5 = nox_float2int(v15) % 46;

	if (i - 1 <= 0 || i >= 127 || j - 1 <= 0 || j >= 127) {
		return -1;
	}
	// The client and server share this legacy tile query during an in-process
	// host transition. Treat an uninitialized row as outside the tile map;
	// GAME.EXE never observes a missing row after its fixed-address allocation.
	if (!ptr_5D4594_2650668 || !ptr_5D4594_2650668[i] || !ptr_5D4594_2650668[i - 1]) {
		return -1;
	}

	int result = 0;
	int v16[2] = {0};
	if (v4 <= v5) {
		if (46 - v4 <= v5) {
			obj_5D4594_2650668_t* obj = &ptr_5D4594_2650668[i][j];
			result = obj->field_6;
			if (obj->field_10) {
				v16[0] = v4;
				v16[1] = v5 - 23;
				result = sub_411350(obj->field_10, v16, result);
			}
		} else {
			obj_5D4594_2650668_t* obj = &ptr_5D4594_2650668[i - 1][j];
			result = obj->field_1;
			if (obj->field_5) {
				v16[1] = v5;
				v16[0] = v4 + 23;
				result = sub_411350(obj->field_5, v16, result);
			}
		}
	} else if (46 - v4 <= v5) {
		obj_5D4594_2650668_t* obj = &ptr_5D4594_2650668[i][j];
		result = obj->field_1;
		if (obj->field_5) {
			v16[1] = v5;
			v16[0] = v4 - 23;
			result = sub_411350(obj->field_5, v16, result);
		}
	} else {
		obj_5D4594_2650668_t* obj = &ptr_5D4594_2650668[i][j - 1];
		result = obj->field_6;
		if (obj->field_10) {
			v16[0] = v4;
			v16[1] = v5 + 23;
			result = sub_411350(obj->field_10, v16, result);
		}
	}
	return result;
}

int nox_xxx_mapTileAllowTeleport_411A90(float2* a1) {
	if (*getMemIntPtr(0x587000, 26520) == -1) {
		for (int i = 0; i < 176; i++) {
			nox_tileDef_t* p = &nox_tile_defs_arr[i];
			if (!strcmp(&p->name[0], "WaterNoTeleport")) {
				*getMemU32Ptr(0x587000, 26516) = i;
			} else if (!strcmp(&p->name[0], "WaterDeepNoTeleport")) {
				*getMemU32Ptr(0x587000, 26520) = i;
			} else if (!strcmp(&p->name[0], "WaterShallowNoTeleport")) {
				*getMemU32Ptr(0x587000, 26524) = i;
			} else if (!strcmp(&p->name[0], "WaterSwampDeepNoTeleport")) {
				*getMemU32Ptr(0x587000, 26528) = i;
			} else if (!strcmp(&p->name[0], "WaterSwampShallowNoTeleport")) {
				*getMemU32Ptr(0x587000, 26532) = i;
			}
		}
	}
	int v3 = nox_xxx_tileNFromPoint_411160(a1);
	return v3 == *getMemU32Ptr(0x587000, 26516) || v3 == *getMemU32Ptr(0x587000, 26520) ||
		   v3 == *getMemU32Ptr(0x587000, 26524) || v3 == *getMemU32Ptr(0x587000, 26528) ||
		   v3 == *getMemU32Ptr(0x587000, 26532);
}

obj_412ae0_t* nox_xxx_modifNext_4133C0(obj_412ae0_t* a1) { return a1->field_34; }

int nox_xxx_checkGameFlagPause_413A50() { return nox_common_gameFlags_check_40A5C0(0x40000); }

double nox_xxx_itemApplyDefendEffect_415C00(nox_object_t* item) {
	if (!item || !(item->obj_class & 0x2000000)) {
		return 0.0;
	}

	void* def = nox_xxx_equipClothFindDefByTT_413270(item->typ_ind);
	if (!def) {
		return 0.0;
	}
	float value = nox_modifier_getArmor(def);
	nox_modifier_attrs_t* attrs = item->init_data;
	void* effect = attrs ? attrs->modifiers[0] : NULL;
	void* callback = nox_modifier_effect_getDefendFunc(effect);
	if (callback) {
		((void (*)(void*, nox_object_t*, uintptr_t, nox_object_t*, uintptr_t, float*))callback)(
			effect, item, 0, item, 0, &value);
	}
	return value;
}

char* nox_xxx_cliGamedataGet_416590(int a1) { return (char*)getMemAt(0x5D4594, 371380 + 58*a1); }

int nox_xxx_cliSetSettingsAcquired_4169D0(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 371704) = a1;
	return result;
}

char* nox_xxx_playerForceSendLessons_416E50(int a1) {
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		player->lessons = 0;
		player->field_2140 = 0;
		if (a1 && player->playerUnit) {
			nox_xxx_netReportLesson_4D8EF0(player->playerUnit);
		}
	}
	return 0;
}

nox_playerInfo* nox_xxx_playerByName_4170D0(wchar2_t* a1) {
	if (!a1) {
		return 0;
	}
	nox_playerInfo* v1 = nox_common_playerInfoGetFirst_416EA0();
	if (!v1) {
		return 0;
	}
	while (_nox_wcsicmp(v1->name_final, a1)) {
		v1 = nox_common_playerInfoGetNext_416EE0(v1);
		if (!v1) {
			return 0;
		}
	}
	return v1;
}

char* nox_xxx_netUnmarkMinimapSpec_417470(nox_object_t* a1, int a2) {
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		nox_xxx_netUnmarkMinimapObj_417300(player->playerInd, a1, a2);
	}
	return 0;
}

char* nox_xxx_netMarkMinimapForAll_4174B0(nox_object_t* a1, int a2) {
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		nox_xxx_netMarkMinimapObject_417190(player->playerInd, a1, a2);
	}
	return 0;
}

char nox_xxx_playerUnsetStatus_417530(nox_playerInfo* a1p, int a2) {
	int v2;   // eax
	short v3; // ax

	a1p->field_3680 &= ~a2;
	v2 = nox_common_gameFlags_check_40A5C0(1);
	if (v2) {
		if (a2 & 1) {
			v2 = nox_common_gameFlags_check_40A5C0(128);
			if (!v2) {
				v2 = nox_xxx_gamePlayIsAnyPlayers_40A8A0();
				if (v2) {
					v2 = sub_40A220();
					if (!v2) {
						v3 = nox_common_gameFlags_getVal_40A5B0();
						LOBYTE(v2) = sub_40A180(v3);
						if ((uint8_t)v2) {
							sub_40A250();
							LOBYTE(v2) = sub_40A1F0(1);
						}
					}
				}
			}
		}
		if (a2 & 0x423) {
			LOBYTE(v2) = nox_xxx_netReportPlayerStatus_417630(a1p);
		}
	}
	return v2;
}

char* nox_xxx_sendAllClientStatus_4175C0(nox_playerInfo* a1) {
	int v3;       // [esp-18h] [ebp-24h]
	char v4[7];   // [esp+4h] [ebp-8h]

	for (nox_playerInfo* i = nox_common_playerInfoGetFirst_416EA0(); i; i = nox_common_playerInfoGetNext_416EE0(i)) {
		v4[0] = 106;
		*(uint16_t*)&v4[1] = i->netCode;
		v3 = a1->playerInd;
		*(uint32_t*)&v4[3] = i->field_3680 & 0x423;
		nox_xxx_netSendPacket1_4E5390(v3, v4, 7, 0, 0);
	}
	return NULL;
}

int nox_xxx_netReportPlayerStatus_417630(nox_playerInfo* pl) {
	short v1;   // cx
	int v2;     // edx
	char v4[7]; // [esp+0h] [ebp-8h]

	v1 = pl->netCode;
	v2 = pl->field_3680 & 0x423;
	v4[0] = 106;
	*(uint16_t*)&v4[1] = v1;
	*(uint32_t*)&v4[3] = v2;
	return nox_xxx_netSendPacket1_4E5390(255, v4, 7, 0, 0);
}
