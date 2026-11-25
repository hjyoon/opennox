static int nox_gui_server_list_first_visible(const nox_scrollListBox_data* data) {
	if (!data || !data->items || !data->field_11_0) {
		return 0;
	}
	int scroll = data->field_13_1;
	if ((int)data->items[0].field_0 > scroll) {
		return 0;
	}
	for (int i = 1; i < data->field_11_0; ++i) {
		if ((int)data->items[i].field_0 > scroll) {
			return i;
		}
	}
	return 0;
}

int* sub_4535E0(int* a1) {
	int* result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1045452) = *a1;
	return result;
}

int sub_4535F0(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1045456) = a1;
	return result;
}

char* sub_453600() { return (char*)getMemAt(0x5D4594, 1045452); }

int sub_453610() { return *getMemU32Ptr(0x5D4594, 1045456); }

uint8_t* sub_453620(uint8_t* a1, char a2, int a3) {
	uint8_t* result; // eax

	result = a1;
	if (a3) {
		*a1 |= a2;
	} else {
		*a1 &= ~a2;
	}
	return result;
}

uint32_t* sub_453640(uint32_t* a1, int a2, int a3) {
	uint32_t* result; // eax

	result = a1;
	if (a3) {
		*a1 |= a2;
	} else {
		*a1 &= ~a2;
	}
	return result;
}

int sub_453660(int a1) {
	int v1; // ecx
	int v2; // edx
	int v3; // eax

	v1 = a1;
	v2 = 0;
	if (a1 > 0) {
		do {
			v3 = v1 >> 8;
			if (v1 >> 8 > 0) {
				v1 >>= 8;
			}
			++v2;
		} while (v3 > 0);
	}
	return ((unsigned char)v1 & getMemByte(0x5D4594, 1045451 + v2)) != 0;
}

int sub_453690(int a1) { return (a1 & *getMemU32Ptr(0x5D4594, 1045456)) != 0; }

void sub_453F70(const void* a1) { memcpy(getMemAt(0x5D4594, 1045488), a1, 0x14u); }

char* sub_453F90() { return (char*)getMemAt(0x5D4594, 1045488); }

int sub_453FA0(void* a1, int a2, int a3) {
	bool v3;          // sf
	char v4;          // cl
	int result;       // eax
	unsigned char v6; // [esp+8h] [ebp+8h]

	v4 = a2 & 0x1F;
	v3 = (a2 & 0x8000001F & 0x80000000) != 0;
	v6 = a2 / 32;
	if (v3) {
		v4 = ((v4 - 1) | 0xE0) + 1;
	}
	result = 1 << v4;
	if (a3) {
		*((uint32_t*)a1 + v6) |= result;
	} else {
		result = ~result;
		*((uint32_t*)a1 + v6) &= result;
	}
	return result;
}

int sub_454000(const void* a1, int a2) {
	return (((const uint32_t*)a1)[(a2 / 32) & 0xFF] & (1u << (a2 % 32))) != 0;
}

int sub_4559B0(wchar2_t* a1) {
	if (!nox_gui_server_access.wnd_10200) {
		return -1;
	}
	nox_scrollListBox_data* list = nox_gui_server_access.wnd_10200->widget_data;
	if (!list || list->field_11_0 == 0) {
		return -1;
	}
	for (int i = 0; i < list->field_11_0; ++i) {
		const wchar2_t* name =
			(const wchar2_t*)nox_window_call_field_94(nox_gui_server_access.wnd_10200, 16406, i, 0);
		if (name && !nox_wcscmp(name, a1)) {
			return i;
		}
	}
	return -1;
}

int sub_456E40(int a1, int a2) {
	int index = 0;
	for (nox_gui_server_player_entry* entry =
			 (nox_gui_server_player_entry*)nox_common_list_getFirstSafe_425890(&nox_gui_server_players.players);
		 entry; entry = (nox_gui_server_player_entry*)nox_common_list_getNextSafe_4258A0(&entry->list), ++index) {
		if (entry->net_code != (uint32_t)a1) {
			continue;
		}
		if (a2) {
			nox_common_list_remove_425920(&entry->list);
			free(entry);
		}
		return index;
	}
	return -1;
}

int sub_456F10(wchar2_t* a1, int a2) {
	if (!a1) {
		return -1;
	}
	int index = 0;
	for (nox_gui_server_team_entry* entry =
			 (nox_gui_server_team_entry*)nox_common_list_getFirstSafe_425890(&nox_gui_server_players.teams);
		 entry; entry = (nox_gui_server_team_entry*)nox_common_list_getNextSafe_4258A0(&entry->list), ++index) {
		if (_nox_wcsicmp(entry->name, a1)) {
			continue;
		}
		if (a2) {
			nox_common_list_remove_425920(&entry->list);
			free(entry);
		}
		return index;
	}
	return -1;
}

static uint32_t nox_gui_server_players_color_value(int index) {
	switch (index) {
	case 0: return nox_color_black_2650656;
	case 1: return *getMemU32Ptr(0x852978, 4);
	case 2: return *getMemU32Ptr(0x85B3FC, 956);
	case 3: return *getMemU32Ptr(0x5D4594, 2597996);
	case 4: return nox_color_white_2523948;
	case 5: return nox_color_violet_2598268;
	case 6: return *getMemU32Ptr(0x85B3FC, 940);
	case 7: return nox_color_red_2589776;
	case 8: return *getMemU32Ptr(0x85B3FC, 984);
	case 9: return dword_8531A0_2572;
	case 10: return nox_color_green_2614268;
	case 11: return *getMemU32Ptr(0x85B3FC, 944);
	case 12: return nox_color_cyan_2649820;
	case 13: return nox_color_blue_2650684;
	case 14: return nox_color_orange_2614256;
	case 15: return nox_color_yellow_2589772;
	case 16: return *getMemU32Ptr(0x852978, 0);
	default: return *getMemU32Ptr(0x5D4594, 2597996);
	}
}

int sub_4571A0(int a1, int a2) {
	if (!dword_5d4594_1045684) {
		return 0;
	}
	nox_window* players = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10501);
	nox_scrollListBox_data* data = players ? players->widget_data : NULL;
	int index = sub_456E40(a1, 0);
	if (!data || !data->items || index < 0 || index >= data->field_11_0) {
		return 0;
	}
	int color = 3;
	if (a2) {
		for (nox_gui_server_team_entry* entry =
				 (nox_gui_server_team_entry*)nox_common_list_getFirstSafe_425890(&nox_gui_server_players.teams);
			 entry; entry = (nox_gui_server_team_entry*)nox_common_list_getNextSafe_4258A0(&entry->list)) {
			if (entry->team_id == (uint32_t)a2) {
				color = entry->color;
				break;
			}
		}
	}
	data->items[index].field_129 = nox_gui_server_players_color_value(color);
	return 1;
}

void sub_457350(int old_id, int new_id) {
	uint32_t old_value = (uint8_t)old_id;
	for (nox_gui_server_team_entry* entry =
			 (nox_gui_server_team_entry*)nox_common_list_getFirstSafe_425890(&nox_gui_server_players.teams);
		 entry; entry = (nox_gui_server_team_entry*)nox_common_list_getNextSafe_4258A0(&entry->list)) {
		if (entry->team_id == old_value) {
			entry->team_id = (uint8_t)new_id;
		}
	}
}

int sub_459560(int a1) {
	if (!dword_5d4594_1046492) {
		return 0;
	}
	nox_window* selected = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, a1);
	if (!selected || !selected->parent) {
		return 0;
	}
	nox_window* parent = selected->parent;
	nox_window* first = NULL;
	nox_window* second = NULL;
	switch (a1) {
	case 10161:
		first = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10163);
		second = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10162);
		break;
	case 10162:
		first = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10163);
		second = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10161);
		break;
	case 10163:
		first = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10162);
		second = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10161);
		break;
	default:
		return 0;
	}
	if (!first || !second) {
		return 0;
	}
	sub_46B120(first, NULL);
	sub_46B120(first, parent);
	sub_46B120(second, NULL);
	sub_46B120(second, parent);
	sub_46B120(selected, NULL);
	return sub_46B120(selected, parent);
}

char* sub_459870() { return (char*)&nox_gui_server_options_rules; }

int sub_459D50(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1046544) = a1;
	return result;
}

int sub_459D60() { return *getMemU32Ptr(0x5D4594, 1046544); }

int sub_459D80(int a1) {
	int result = 0; // eax

	if (dword_5d4594_1046492) {
		result = nox_window_set_hidden(dword_5d4594_1046492, a1);
	}
	return result;
}

int sub_459DA0() { return dword_5d4594_1046492 != 0; }

int sub_459DB0(nox_drawable* dr) {
	return dr && (dr->flags28 & 0x400000u) != 0 && (dr->flags29 & 8u) != 0;
}

nox_drawable* nox_xxx_cliNextMinimapObj_459EC0(nox_drawable* dr) {
	nox_drawable* next = dr ? dr->field_102 : NULL;
	if (dr && dr == next) {
		printf("nox_xxx_cliNextMinimapObj_459EC0: infinite loop!\n");
		abort();
		return NULL;
	}
	return next;
}

nox_drawable* sub_45A010(nox_drawable* dr) { return dr->field_104; }

nox_drawable* nox_drawable_next_45A070(nox_drawable* a1) {
	return a1 ? a1->field_92 : NULL;
}

nox_drawable* nox_xxx_spriteSetActiveMB_45A990_drawable(nox_drawable* dr) {
	dr->flags30 |= 4u;
	return dr;
}

nox_drawable* nox_xxx_spriteSetFrameMB_45AB80_drawable(nox_drawable* dr, int frame) {
	if (!(dr->flags28 & 2) || !(dr->flags29 & 0x40000) || dr->field_69 != 8) {
		dr->field_78 = dr->field_77;
		dr->field_77 = frame;
	}
	return dr;
}
