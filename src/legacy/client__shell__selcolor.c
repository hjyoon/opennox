#include "client__shell__selcolor.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "client__gui__window.h"
#include "client__shell__optsback.h"
#include "common/fs/nox_fs.h"
#include "common__strman.h"
extern nox_window* dword_5d4594_1308136;
extern nox_window* dword_5d4594_1308104;
extern nox_window* dword_5d4594_1308096;
extern nox_window* dword_5d4594_1308112;
extern nox_window* dword_5d4594_1308148;
extern uint32_t dword_587000_171388;
extern nox_window* dword_5d4594_1308140;
extern nox_window* dword_5d4594_1308144;
extern nox_window* dword_5d4594_1308152;
extern nox_window* dword_5d4594_1308116;
extern nox_window* dword_5d4594_1308100;
extern nox_window* dword_5d4594_1308132;
extern nox_window* dword_5d4594_1308108;
extern nox_window* dword_5d4594_1308120;
extern nox_window* dword_5d4594_1308128;
extern nox_window* dword_5d4594_1308124;
extern nox_gui_animation* nox_wnd_xxx_1308092;
extern nox_window* dword_5d4594_1308088;
extern nox_window* dword_5d4594_1308084;
extern char* dword_5d4594_1307784;

//----- (004A5D00) --------------------------------------------------------
int nox_game_showSelColor_4A5D00() {
	char* v0;
	nox_window* child;

	nox_game_addStateCode_43BDD0(700);
	v0 = nox_xxx_getHostInfoPtr_431770();
	dword_5d4594_1307784 = v0;
	v0[67] = 0;
	dword_5d4594_1308084 = nox_new_window_from_file("SelColor.wnd", sub_4A7330);
	if (dword_5d4594_1308084) {
		nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1308084, sub_4A18E0);
		nox_wnd_xxx_1308092 = nox_gui_makeAnimation_43C5B0(dword_5d4594_1308084, 0, 0, 0, -440, 0, 20, 0, -40);
		if (nox_wnd_xxx_1308092) {
			nox_wnd_xxx_1308092->field_0 = 700;
			nox_wnd_xxx_1308092->field_12 = sub_4A6890;
			nox_wnd_xxx_1308092->fnc_done_out = sub_4A6C90;
			sub_4A5E90();
			for (int i = 720; i <= 729; ++i) {
				child = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1308084, i);
				nox_xxx_wndSetDrawFn_46B340(child, sub_4A6D20);
			}
			for (int i = 761; i <= 792; ++i) {
				child = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1308084, i);
				nox_xxx_wndSetDrawFn_46B340(child, sub_4A6D20);
			}
			if (dword_587000_171388) {
				wchar2_t* name =
					nox_strman_loadString_40F1D0("DefaultName", 0, "C:\\NoxPost\\src\\client\\shell\\SelColor.c", 1138);
				nox_window_call_field_94(dword_5d4594_1308152, 16414, (uintptr_t)name, 0);
			}
			nox_xxx_wndRetNULL_46A8A0();
			dword_5d4594_1308088 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1308084, 760);
			nox_xxx_wndSetProc_46B2C0(dword_5d4594_1308088, sub_4A7330);
			nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1308088, sub_4A7270);
			sub_46B120(dword_5d4594_1308088, 0);
			sub_4BFAD0();
			child = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1308084, 740);
			nox_xxx_wndSetDrawFn_46B340(child, sub_4A6DC0);
			return 1;
		}
	}
	return 0;
}
//----- (004A68C0) --------------------------------------------------------
wchar2_t* sub_4A68C0() {
	wchar2_t* v0;        // esi
	wchar2_t* v1;        // eax
	wchar2_t* v2;        // eax
	unsigned char* v3;  // edx
	char* v4;           // eax
	char* v5;           // ecx
	char* v6;           // eax
	unsigned char* v7;  // eax
	char* v8;           // ecx
	char* v9;           // ecx
	char* v10;          // eax
	char* v11;          // ecx
	unsigned char* v12; // eax
	char* v13;          // ecx
	char* v14;          // eax
	char* v15;          // ecx
	unsigned char* v16; // eax
	char* v17;          // edx
	char* v18;          // eax
	char* v19;          // ecx
	unsigned char* v20; // eax
	wchar2_t* result;    // eax

	v0 = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1308152, 16413, 0, 0);
	if (!*v0) {
		v1 = nox_strman_loadString_40F1D0("DefaultName", 0, "C:\\NoxPost\\src\\client\\shell\\SelColor.c", 225);
		nox_wcscpy(v0, v1);
	}
	nox_wcscpy((wchar2_t*)dword_5d4594_1307784, v0);
	if (!sub_4A6B50((wchar2_t*)dword_5d4594_1307784)) {
		v2 = nox_strman_loadString_40F1D0("DefaultName", 0, "C:\\NoxPost\\src\\client\\shell\\SelColor.c", 232);
		nox_wcscpy((wchar2_t*)dword_5d4594_1307784, v2);
	}
	uint32_t color = nox_selcolor_value(dword_5d4594_1308096);
	v3 = getMemAt(0x5D4594, 1307796 + 3 * ((color >> 16) + 32 * (unsigned short)color));
	v4 = dword_5d4594_1307784 + 71;
	*(uint16_t*)(dword_5d4594_1307784 + 71) = *(uint16_t*)v3;
	*(uint8_t*)(v4 + 2) = v3[2];
	if (dword_5d4594_1308136->flags & 8) {
		color = nox_selcolor_value(dword_5d4594_1308100);
		v7 = getMemAt(0x5D4594, 1307796 + 3 * ((color >> 16) + 32 * (unsigned short)color));
		v8 = dword_5d4594_1307784 + 68;
		*(uint16_t*)(dword_5d4594_1307784 + 68) = *(uint16_t*)v7;
		*(uint8_t*)(v8 + 2) = v7[2];
	} else {
		v5 = dword_5d4594_1307784 + 71;
		v6 = dword_5d4594_1307784 + 68;
		*(uint16_t*)(dword_5d4594_1307784 + 68) = *(uint16_t*)(dword_5d4594_1307784 + 71);
		*(uint8_t*)(v6 + 2) = *(uint8_t*)(v5 + 2);
	}
	if (dword_5d4594_1308140->flags & 8) {
		v11 = dword_5d4594_1307784 + 74;
		color = nox_selcolor_value(dword_5d4594_1308104);
		v12 = getMemAt(0x5D4594, 1307796 + 3 * ((color >> 16) + 32 * (unsigned short)color));
		*(uint16_t*)(dword_5d4594_1307784 + 74) = *(uint16_t*)v12;
		*(uint8_t*)(v11 + 2) = v12[2];
	} else {
		v9 = dword_5d4594_1307784 + 71;
		v10 = dword_5d4594_1307784 + 74;
		*(uint16_t*)(dword_5d4594_1307784 + 74) = *(uint16_t*)(dword_5d4594_1307784 + 71);
		*(uint8_t*)(v10 + 2) = *(uint8_t*)(v9 + 2);
	}
	if (dword_5d4594_1308144->flags & 8) {
		v15 = dword_5d4594_1307784 + 77;
		color = nox_selcolor_value(dword_5d4594_1308108);
		v16 = getMemAt(0x5D4594, 1307796 + 3 * ((color >> 16) + 32 * (unsigned short)color));
		*(uint16_t*)(dword_5d4594_1307784 + 77) = *(uint16_t*)v16;
		*(uint8_t*)(v15 + 2) = v16[2];
	} else {
		v13 = dword_5d4594_1307784 + 71;
		v14 = dword_5d4594_1307784 + 77;
		*(uint16_t*)(dword_5d4594_1307784 + 77) = *(uint16_t*)(dword_5d4594_1307784 + 71);
		*(uint8_t*)(v14 + 2) = *(uint8_t*)(v13 + 2);
	}
	if (dword_5d4594_1308148->flags & 8) {
		v19 = dword_5d4594_1307784 + 80;
		color = nox_selcolor_value(dword_5d4594_1308112);
		v20 = getMemAt(0x5D4594, 1307796 + 3 * ((color >> 16) + 32 * (unsigned short)color));
		*(uint16_t*)(dword_5d4594_1307784 + 80) = *(uint16_t*)v20;
		*(uint8_t*)(v19 + 2) = v20[2];
	} else {
		v17 = dword_5d4594_1307784 + 71;
		v18 = dword_5d4594_1307784 + 80;
		*(uint16_t*)(dword_5d4594_1307784 + 80) = *(uint16_t*)(dword_5d4594_1307784 + 71);
		*(uint8_t*)(v18 + 2) = *(uint8_t*)(v17 + 2);
	}
	dword_5d4594_1307784[83] = nox_selcolor_value(dword_5d4594_1308116) >> 16;
	dword_5d4594_1307784[84] = nox_selcolor_value(dword_5d4594_1308120) >> 16;
	dword_5d4594_1307784[85] = nox_selcolor_value(dword_5d4594_1308124) >> 16;
	dword_5d4594_1307784[86] = nox_selcolor_value(dword_5d4594_1308128) >> 16;
	result = (wchar2_t*)dword_5d4594_1307784;
	dword_5d4594_1307784[87] = nox_selcolor_value(dword_5d4594_1308132) >> 16;
	return result;
}

static void nox_selcolor_copy_palette_rgb(uint8_t dst[3], const nox_window* win) {
	memcpy(dst, getMemAt(0x5D4594, 1307796 + 3 * nox_selcolor_palette_index(win)), 3);
}

static int nox_selcolor_append_path(char* dst, size_t capacity, const char* suffix) {
	size_t used = strlen(dst);
	size_t added = strlen(suffix);
	if (used + added >= capacity) {
		return 0;
	}
	memcpy(dst + used, suffix, added + 1);
	return 1;
}

//----- (004A75C0) --------------------------------------------------------
int sub_4A75C0() {
	nox_savegame_xxx save = {0};
	char filename[16];
	int index = 0;
	wchar2_t* name;
	FILE* file;

	if (nox_common_gameFlags_check_40A5C0(2048)) {
		nox_savegame_rm_4DBE10("WORKING", 0);
	}

	name = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1308152, 16413, 0, 0);
	if (!*name) {
		nox_wcscpy(name, nox_strman_loadString_40F1D0(
							 "DefaultName", 0, "C:\\NoxPost\\src\\client\\shell\\SelColor.c", 605));
	}
	nox_wcscpy(save.player_name, name);
	if (!sub_4A6B50(save.player_name)) {
		nox_wcscpy(save.player_name, nox_strman_loadString_40F1D0(
										  "DefaultName", 0, "C:\\NoxPost\\src\\client\\shell\\SelColor.c", 612));
	}

	save.field_1276 = 1;
	save.player_class = *(uint8_t*)(dword_5d4594_1307784 + 66);
	nox_selcolor_copy_palette_rgb(&save.field_1204[0], dword_5d4594_1308096);
	if (dword_5d4594_1308136->flags & 8) {
		nox_selcolor_copy_palette_rgb(&save.field_1204[3], dword_5d4594_1308100);
	} else {
		memcpy(&save.field_1204[3], &save.field_1204[0], 3);
	}
	if (dword_5d4594_1308140->flags & 8) {
		nox_selcolor_copy_palette_rgb(&save.field_1204[6], dword_5d4594_1308104);
	} else {
		memcpy(&save.field_1204[6], &save.field_1204[0], 3);
	}
	if (dword_5d4594_1308144->flags & 8) {
		nox_selcolor_copy_palette_rgb(&save.field_1204[9], dword_5d4594_1308108);
	} else {
		memcpy(&save.field_1204[9], &save.field_1204[0], 3);
	}
	if (dword_5d4594_1308148->flags & 8) {
		nox_selcolor_copy_palette_rgb(&save.field_1204[12], dword_5d4594_1308112);
	} else {
		memcpy(&save.field_1204[12], &save.field_1204[0], 3);
	}
	save.field_1204[15] = nox_selcolor_value(dword_5d4594_1308116) >> 16;
	save.field_1204[16] = nox_selcolor_value(dword_5d4594_1308120) >> 16;
	save.field_1204[17] = nox_selcolor_value(dword_5d4594_1308124) >> 16;
	save.field_1204[18] = nox_selcolor_value(dword_5d4594_1308128) >> 16;
	save.field_1204[19] = nox_selcolor_value(dword_5d4594_1308132) >> 16;

	if (strlen(nox_fs_root()) >= sizeof(save.path)) {
		return 0;
	}
	strcpy(save.path, nox_fs_root());
	if (!nox_selcolor_append_path(save.path, sizeof(save.path), "\\Save\\")) {
		return 0;
	}
	if (nox_common_gameFlags_check_40A5C0(2048) &&
		!nox_selcolor_append_path(save.path, sizeof(save.path), "WORKING\\")) {
		return 0;
	}
	nox_fs_mkdir(save.path);
	nox_fs_set_workdir(save.path);

	if (nox_common_gameFlags_check_40A5C0(2048)) {
		strcpy(filename, "Player.plr");
	} else {
		for (index = 0; index < 100; ++index) {
			nox_sprintf(filename, "%.6s%2.2d.plr", (char*)save.player_name, index);
			file = nox_fs_open(filename);
			if (!file) {
				break;
			}
			nox_fs_close(file);
		}
	}
	nox_fs_set_workdir(nox_fs_root());
	if (index > 99 || !nox_selcolor_append_path(save.path, sizeof(save.path), filename)) {
		return 0;
	}

	if (nox_common_gameFlags_check_40A5C0(2048)) {
		if (save.player_class == 0) {
			nox_xxx_gameSetMapPath_409D70("War01a.map");
		} else if (save.player_class == 1) {
			nox_xxx_gameSetMapPath_409D70("Wiz01a.map");
		} else if (save.player_class == 2) {
			nox_xxx_gameSetMapPath_409D70("Con01a.map");
		}
	}
	return sub_41CEE0(&save, 1);
}
