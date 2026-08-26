#include "GAME2_1.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "client__draw__debugdraw.h"
#include "client__draw__staticdraw.h"
#include "client__drawable__drawable.h"
#include "client__gui__window.h"
#include "common__net_list.h"
#include "common__object__modifier.h"
#include "common__system__team.h"
#include "operators.h"

#include "client__gui__gamewin__gamewin.h"
#include "client__gui__gui_ctf.h"
#include "client__gui__guicon.h"
#include "client__gui__guiinput.h"
#include "client__gui__guiinv.h"
#include "client__gui__guijourn.h"
#include "client__gui__guimeter.h"
#include "client__gui__guiquit.h"
#include "client__gui__guirank.h"
#include "client__gui__guisave.h"
#include "client__gui__guispell.h"
#include "client__gui__guisumn.h"
#include "client__gui__guitrade.h"
#include "client__gui__tooltip.h"
#include "client__shell__selcolor.h"
#include "client__video__draw_common.h"

#include "client__system__ctrlevnt.h"
#include "common/fs/nox_fs.h"
#include "common__magic__speltree.h"
#include "input_common.h"

extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_8531A0_2572;
extern uint32_t dword_5d4594_1062552;
extern uint32_t dword_5d4594_1049844;
extern uint32_t dword_5d4594_1050008;
extern uint32_t dword_5d4594_1096272;
extern nox_window* dword_5d4594_1049516;
extern uint32_t dword_5d4594_1062520;
extern uint32_t dword_5d4594_1096276;
extern uint32_t dword_5d4594_1049976;
extern uint32_t dword_5d4594_1090284;
extern uint32_t dword_5d4594_1062484;
extern uint32_t dword_5d4594_1049992;
extern uint32_t dword_5d4594_1062556;
extern uint32_t dword_5d4594_1096264;
extern uint32_t dword_5d4594_1062564;
extern uint32_t dword_5d4594_1090280;
extern uint32_t dword_5d4594_1062560;
extern uint32_t dword_5d4594_1096280;
extern uint32_t dword_587000_145672;
extern uint32_t dword_5d4594_1064868;
extern uint32_t dword_5d4594_1049996;
extern uintptr_t dword_5d4594_1062492;
extern uint32_t dword_5d4594_1062496;
extern uint32_t dword_5d4594_1096260;
extern uint32_t nox_client_gui_flag_1556112;
extern uint32_t dword_5d4594_1096284;
extern uint32_t dword_587000_145664;
extern void* dword_5d4594_1096288;
extern uint32_t dword_5d4594_1047936;
extern uint32_t dword_5d4594_1062488;
extern uint32_t dword_5d4594_1062468;
extern nox_window* dword_5d4594_1064860;
extern uint32_t dword_5d4594_1096252;
extern void* dword_5d4594_1064864;
extern uint32_t dword_587000_136184;
extern void nox_xxx_playerInitColors_461460_go(nox_playerInfo* pl);
extern nox_window* dword_5d4594_1049532;
extern uint32_t dword_5d4594_1049484;
extern uint32_t dword_5d4594_1062516;
extern uint32_t nox_client_translucentFrontWalls_805844;
extern uint64_t qword_581450_9512;
extern uint64_t qword_581450_9544;
extern nox_window* dword_5d4594_1090276;
extern uint32_t dword_5d4594_1062476;
extern uint32_t dword_5d4594_1047932;
extern nox_window* dword_5d4594_1049512;
extern uint32_t nox_client_highResFloors_154952;
extern uint32_t dword_5d4594_1062528;
extern uint32_t dword_5d4594_1062524;
extern nox_window* dword_5d4594_1064856;
extern uint32_t dword_5d4594_1049856;
extern nox_window* dword_5d4594_1049520;
extern uint32_t nox_client_highResFrontWalls_80820;
extern uint32_t dword_5d4594_1049800_inventory_click_row_index;
extern uint32_t nox_xxx_minimap_587000_149232;
extern uint32_t dword_5d4594_1062456;
extern uint32_t dword_5d4594_1063636;
extern uint32_t dword_5d4594_1049796_inventory_click_column_index;
extern nox_window* dword_5d4594_1049508;
extern nox_window* dword_5d4594_1090048;
extern nox_window* dword_5d4594_1049500;
extern uint32_t dword_5d4594_1062512;
extern uint32_t dword_5d4594_1049864;
extern uint32_t dword_5d4594_1062508;
extern nox_window* dword_5d4594_1049504;
extern uint32_t dword_5d4594_1090120;
extern nox_drawable* dword_5d4594_1063116;
extern uintptr_t dword_5d4594_1062480;
extern uint32_t nox_player_netCode_85319C;
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern int nox_win_width;
extern int nox_win_height;
extern uintptr_t array_5D4594_1049872[9];

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_black_2650656;

nox_window* nox_win_unk5 = 0;
nox_window* dword_5d4594_1062452 = 0;
nox_window* nox_inventory_window = 0;
nox_window* nox_inventory_current_weapon_window = 0;
nox_window* nox_inventory_identify_window = 0;
nox_window* nox_inventory_scroll_window = 0;
nox_window* nox_inventory_journal_button = 0;
nox_window* nox_inventory_stats_button = 0;
nox_window* nox_inventory_scroll_up_button = 0;
nox_window* nox_inventory_scroll_down_button = 0;
nox_window* nox_inventory_close_button = 0;
static nox_window* nox_inventory_overlay_window = 0;
void* nox_inventory_font = 0;

typedef struct {
	int min;
	int max;
	float scale;
	int value;
} nox_slider_data_t;

enum {
	NOX_INV_IMG_BASE,
	NOX_INV_IMG_IDENTIFY_BASE,
	NOX_INV_IMG_TRAY1,
	NOX_INV_IMG_TRAY2,
	NOX_INV_IMG_TRAY3,
	NOX_INV_IMG_TRAY_SPECIAL,
	NOX_INV_IMG_TRAY_IDENTIFY_LIT,
	NOX_INV_IMG_TRAY_MAP_LIT,
	NOX_INV_IMG_UP,
	NOX_INV_IMG_UP_LIT,
	NOX_INV_IMG_DOWN,
	NOX_INV_IMG_DOWN_LIT,
	NOX_INV_IMG_SLIDER,
	NOX_INV_IMG_SLIDER_LIT,
	NOX_INV_IMG_EQUIP_RING,
	NOX_INV_IMG_QUICK_ITEM_RING,
	NOX_INV_IMG_CLOSE_LIT,
	NOX_INV_IMG_JOURNAL_LIT,
	NOX_INV_IMG_INVENTORY,
	NOX_INV_IMG_INVENTORY_LIT,
	NOX_INV_IMG_DOLL_LIT,
	NOX_INV_IMG_STATS,
	NOX_INV_IMG_STATS_LIT,
	NOX_INV_IMG_FIST,
	NOX_INV_IMG_SHARED_KEY_MODE,
	NOX_INV_IMG_COUNT,
};
static nox_video_bag_image_t* nox_inventory_images[NOX_INV_IMG_COUNT];
static nox_things_imageRef_t* nox_inventory_extra_lives_anim = 0;

nox_window_yyy nox_windows_arr_1093036[7] = {0};
nox_drawable nox_gui_bottle_drawables[3] = {0};

obj_5D4594_2650668_t** ptr_5D4594_2650668 = 0;
const int ptr_5D4594_2650668_cap = 128;

nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX] = {0};
static nox_drawable* nox_client_inventory_dragged_1049848 = NULL;

nox_drawable* nox_client_inventory_get_dragged(void) { return nox_client_inventory_dragged_1049848; }

void nox_client_inventory_set_dragged(nox_drawable* drawable) {
	nox_client_inventory_dragged_1049848 = drawable;
#if UINTPTR_MAX == UINT32_MAX
	// Preserve the original GAME.EXE slot as a 32-bit oracle mirror. A wider
	// build must never put a live pointer in this fixed-width memory map.
	*getMemU32Ptr(0x5D4594, 1049848) = (uint32_t)(uintptr_t)drawable;
#endif
}

int nox_client_inventory_animation_state(void) { return getMemByte(0x5D4594, 1049868); }

int nox_client_inventory_animation_offset(void) { return (int32_t)dword_587000_136184; }

int nox_client_inventory_has_dragged(void) { return nox_client_inventory_dragged_1049848 != NULL; }

//----- (00460D40) --------------------------------------------------------
int sub_460D40() { return dword_5d4594_1049508 != 0; }

//----- (00460D50) --------------------------------------------------------
int sub_460D50() {
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1049500);
	dword_5d4594_1049500 = 0;
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1049504);
	dword_5d4594_1049504 = 0;
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1049520);
	dword_5d4594_1049520 = 0;
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1049508);
	dword_5d4594_1049508 = 0;
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1049512);
	dword_5d4594_1049512 = 0;
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1049516);
	dword_5d4594_1049516 = 0;

	// These Window pointers occupied 32-bit slots embedded in the original
	// quick-bar data. Native builds keep them in a pointer-width sidecar.
	nox_quickbar_windows_destroy_all();
	for (int row = 0; row < 5; ++row) {
		size_t base = 1048196 + 256 * (size_t)row;
		*getMemU32Ptr(0x5D4594, base + 208) = 0;
		for (int slot = 0; slot < 5; ++slot) {
			*getMemU32Ptr(0x5D4594, base + 212 + 4 * (size_t)slot) = 0;
			*getMemU32Ptr(0x5D4594, base + 232 + 4 * (size_t)slot) = 0;
		}
	}
	*getMemU32Ptr(0x5D4594, 1048148) = 0;
	for (size_t off = 1048152; off < 1048164; off += 4) {
		*getMemU32Ptr(0x5D4594, off) = 0;
	}
	dword_5d4594_1049532 = 0;
	*getMemU32Ptr(0x5D4594, 1047928) = 0;
	dword_5d4594_1047932 = 0;
	return 0;
}

//----- (00460E60) --------------------------------------------------------
int nox_xxx_cliPrepareGameplay1_460E60() {
	int result; // eax

	if (sub_460D40()) {
		sub_460D50();
	}
	result = nox_xxx_quickBarCreate_45E190();
	if (result) {
		sub_460EA0(nox_client_getRenderGUI());
		result = 1;
	}
	return result;
}

//----- (00460EA0) --------------------------------------------------------
int sub_460EA0(int a1) { return sub_460B90(a1); }

//----- (00460EB0) --------------------------------------------------------
void sub_460EB0(int a1, char a2) {
	if (a1 < 0 || a1 >= 140) {
		return;
	}
	*getMemU8Ptr(0x5D4594, 1049544 + a1) = a2;
}

//----- (00461010) --------------------------------------------------------
void sub_461010() {
	if (!dword_5d4594_1049484) {
		return;
	}
	nox_window_set_hidden(*getMemIntPtr(0x5D4594, 1048148), 1);
	nox_window_set_hidden(*(int*)&dword_5d4594_1049512, 0);
	sub_46AE10(dword_5d4594_1049500, 0);
	nox_xxx_clientPlaySoundSpecial_452D80(797, 100);
	dword_5d4594_1049484 = 0;
}

//----- (00461060) --------------------------------------------------------
void sub_461060() {
	if (dword_5d4594_1049484 == 1) {
		sub_461010();
		return;
	}
	if (*getMemU32Ptr(0x5D4594, 1049476) == 1) {
		nox_xxx_quickBarClose_4606B0();
	}
	nox_window_set_hidden(*getMemIntPtr(0x5D4594, 1048148), 0);
	nox_window_set_hidden(*(int*)&dword_5d4594_1049512, 1);
	sub_46AE10(dword_5d4594_1049500, 1);
	nox_xxx_clientPlaySoundSpecial_452D80(796, 100);
	dword_5d4594_1049484 = 1;
}

//----- (00461090) --------------------------------------------------------
char* sub_461090(int a1, int a2) {
	int v2;       // edx
	char* result; // eax

	v2 = gameFrame();
	result = (char*)getMemAt(0x5D4594, 1047764 + 24*1 + 20);
	do {
		if (*((uint32_t*)result - 5) == a1) {
			*(uint32_t*)result = a2 == 0 ? v2 : 0;
			*((uint32_t*)result - 3) = a2;
		}
		result += 24;
	} while ((int)result < (int)getMemAt(0x5D4594, 1047928));
	return result;
}

//----- (004610D0) --------------------------------------------------------
char* sub_4610D0(unsigned char a1) {
	int* v1;      // esi
	char* result; // eax

	if (a1 != 6) {
		return sub_461090(*getMemU32Ptr(0x5D4594, 1047764 + 24*a1), 1);
	}
	v1 = getMemIntPtr(0x5D4594, 1047764 + 24*1);
	do {
		result = sub_461090(*v1, 1);
		v1 += 6;
	} while ((int)v1 < (int)getMemAt(0x5D4594, 1047908));
	return result;
}

//----- (00461120) --------------------------------------------------------
char* sub_461120(int a1, int a2) {
	int v2;       // edx
	char* result; // eax

	v2 = 1 << a1;
	result = (char*)getMemAt(0x5D4594, 1047764 + 24*1 + 12);
	do {
		if (*((uint32_t*)result - 3) == a1) {
			if (a2) {
				*(uint32_t*)result |= v2;
			} else {
				*(uint32_t*)result &= ~v2;
			}
		}
		result += 24;
	} while ((int)result < (int)getMemAt(0x5D4594, 1047920));
	return result;
}

//----- (00461160) --------------------------------------------------------
int sub_461160(int a1) {
	int v1;            // edx
	unsigned char* v2; // eax

	v1 = 1;
	v2 = getMemAt(0x5D4594, 1047764 + 24*1);
	while (*(uint32_t*)v2 != a1) {
		v2 += 24;
		++v1;
		if ((int)v2 >= (int)getMemAt(0x5D4594, 1047908)) {
			return 0;
		}
	}
	return ((1 << a1) & *getMemU32Ptr(0x5D4594, 1047764 + 24*v1 + 12)) != 0;
}

//----- (004611A0) --------------------------------------------------------
int sub_4611A0() { return dword_5d4594_1047932; }

//----- (004611B0) --------------------------------------------------------
int sub_4611B0() {
	int result; // eax

	result = dword_5d4594_1047936;
	if (dword_5d4594_1047936) {
		result = nox_xxx_clientSendAbil_45DAF0(*(int*)&dword_5d4594_1047936);
		dword_5d4594_1047936 = 0;
		dword_5d4594_1047932 = 0;
	}
	return result;
}

//----- (004611E0) --------------------------------------------------------
void nox_xxx_netAbilityRewardCli_4611E0(int a1, int a2, char* a3) {
	unsigned char* v3; // esi

	if (a1 >= 1 && a1 < 6) {
		v3 = getMemAt(0x5D4594, 1047764 + 24*1 + 16);
		do {
			if (*((uint32_t*)v3 - 4) == a1 && *(uint32_t*)v3 != a2) {
				if (nox_common_gameFlags_check_40A5C0(2) && dword_8531A0_2576) {
					*(uint32_t*)(dword_8531A0_2576 + 4 * a1 + 3696) = a2;
				}
				*(uint32_t*)v3 = a2;
				if (a2) {
					nox_xxx_abilityReward_45D290(a1, a3, (int)a3);
				}
			}
			v3 += 24;
		} while ((int)v3 < (int)getMemAt(0x5D4594, 1047924));
	}
}

//----- (00461250) --------------------------------------------------------
int nox_xxx_buttonFindFirstEmptySlot_461250() {
	int v0;       // ecx
	int v1;       // esi
	uint32_t* v2; // eax

	v0 = *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200);
	do {
		v1 = 0;
		v2 = (uint32_t*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 40 * v0);
		do {
			if (!*v2) {
				nox_xxx_clientUpdateButtonRow_45E110(v0);
				return v1;
			}
			++v1;
			v2 += 2;
		} while (v1 < 5);
		if (++v0 >= 5) {
			v0 = 0;
		}
	} while (v0 != *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200));
	return -1;
}

//----- (004612A0) --------------------------------------------------------
int sub_4612A0() {
	int result;  // eax
	uint32_t* i; // ecx

	result = 0;
	for (i = (uint32_t*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 +
						 40 * *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200));
		 *i; i += 2) {
		if (++result >= 5) {
			return -1;
		}
	}
	return result;
}

//----- (004612D0) --------------------------------------------------------
int nox_xxx_buttonHaveSpellInBarMB_4612D0(int a1) {
	int v1;       // edx
	int v2;       // eax
	uint32_t* v3; // ecx

	v1 = *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200);
	do {
		v2 = 0;
		v3 = (uint32_t*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 40 * v1);
		do {
			if (*v3 == a1) {
				return 1;
			}
			++v2;
			v3 += 2;
		} while (v2 < 5);
		if (++v1 >= 5) {
			v1 = 0;
		}
	} while (v1 != *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200));
	return 0;
}

//----- (00461320) --------------------------------------------------------
void nox_xxx_buttonSetImgMB_461320(int a1, uint32_t* a2) {
	if (a2) {
		if (a1 >= 0 && a1 < 5) {
			nox_client_wndGetPosition_46AA60(
				*(uint32_t**)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 4 * a1 + 212), a2, a2 + 1);
		}
	}
}

//----- (00461360) --------------------------------------------------------
int sub_461360(int a1) {
	int v1;     // edx
	int v2;     // ecx
	int v3;     // ebx
	int v4;     // esi
	int result; // eax

	v1 = nox_xxx_aClosewoodengat_587000_133480;
	v2 = *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200);
	v3 = *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200);
	do {
		v4 = 5;
		result = 40 * v2;
		do {
			if (*(uint32_t*)(result + v1) == a1) {
				*(uint32_t*)(result + v1) = 0;
				v1 = nox_xxx_aClosewoodengat_587000_133480;
			}
			result += 8;
			--v4;
		} while (v4);
		if (++v2 >= 5) {
			v2 = 0;
		}
	} while (v2 != v3);
	return result;
}

//----- (00461400) --------------------------------------------------------
int sub_461400() {
	int i;      // esi
	int result; // eax
	int v2;     // ecx
	uint8_t* data = nox_xxx_aClosewoodengat_587000_133480;

	for (i = 0; i < 40; i += 8) {
		result = i;
		v2 = 5;
		do {
			*(uint32_t*)(data + result) = *getMemU32Ptr(0x5D4594, 1047564 + result);
			*(uint8_t*)(data + result + 4) =
				getMemByte(0x5D4594, 1047568 + result);
			result += 40;
			--v2;
		} while (v2);
	}
	return result;
}

//----- (00461440) --------------------------------------------------------
int sub_461440(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1049688) = a1;
	return result;
}

//----- (00461450) --------------------------------------------------------
int sub_461450() { return *getMemU32Ptr(0x5D4594, 1049688); }

//----- (00461460) --------------------------------------------------------
void nox_xxx_playerInitColors_461460(nox_playerInfo* pl) {
	nox_xxx_playerInitColors_461460_go(pl);
}

//----- (00461520) --------------------------------------------------------
char* sub_461520() {
	nox_playerInfo* result; // eax
	nox_playerInfo* i;      // esi

	result = nox_common_playerInfoGetFirst_416EA0();
	for (i = result; result; i = result) {
		nox_xxx_playerInitColors_461460(i);
		result = nox_common_playerInfoGetNext_416EE0(i);
	}
	return (char*)result;
}

//----- (00461550) --------------------------------------------------------
int nox_xxx_clientSetAltWeapon_461550(nox_inventory_cell_t* cell) {
	nox_inventory_cell_t* previous = (nox_inventory_cell_t*)dword_5d4594_1062480;
	if (previous) {
		dword_5d4594_1062484 = previous->field_4;
	} else {
		dword_5d4594_1062484 = 0;
	}
	dword_5d4594_1062480 = (uintptr_t)cell;
	sub_4619F0();
	if (!cell) {
		return nox_xxx_clientReportSecondaryWeapon_4BF010(0);
	}
	cell->field_0->field_32 = cell->field_4;
	cell->field_136 = 1;
	return nox_xxx_clientReportSecondaryWeapon_4BF010(cell->field_0);
}

//----- (004615C0) --------------------------------------------------------
nox_drawable* sub_4615C0() {
	uint32_t bow_type = *getMemU32Ptr(0x5D4594, 1063640);
	if (!bow_type) {
		bow_type = nox_xxx_getTTByNameSpriteMB_44CFC0("Bow");
		*getMemU32Ptr(0x5D4594, 1063640) = bow_type;
	}
	nox_drawable* ranged = (nox_drawable*)array_5D4594_1049872[8];
	for (nox_drawable* drawable = ranged; drawable; drawable = drawable->field_92) {
		if (drawable->field_27 == bow_type) {
			return ranged;
		}
	}
	return (nox_drawable*)array_5D4594_1049872[7];
}

//----- (00461600) --------------------------------------------------------
nox_drawable* sub_461600(int thing_type) {
	for (size_t slot = 0; slot < 9; slot++) {
		for (nox_drawable* drawable = (nox_drawable*)array_5D4594_1049872[slot]; drawable;
			 drawable = drawable->field_92) {
			if (drawable->field_27 == (uint32_t)thing_type) {
				return drawable;
			}
		}
	}
	return NULL;
}

//----- (00461630) --------------------------------------------------------
int nox_xxx_send2ServInvenFail_461630(short a1) {
	char v3[3]; // [esp+0h] [ebp-4h]
	v3[0] = -15;
	*(uint16_t*)&v3[1] = a1;
	return nox_xxx_netClientSend2_4E53C0(31, v3, 3, 0, 0);
}

//----- (00461930) --------------------------------------------------------
int sub_461930() {
	for (size_t slot = 0; slot < 9; slot++) {
		for (nox_drawable* drawable = (nox_drawable*)array_5D4594_1049872[slot]; drawable;
			 drawable = drawable->field_92) {
			if (drawable->flags28 & 0x1001000) {
				return 1;
			}
		}
	}
	return 0;
}

//----- (00461970) --------------------------------------------------------
nox_inventory_cell_t* sub_461970(int net_code, int thing_type) {
	if (!(nox_get_thing(thing_type)->pri_class & 0x4000000)) {
		for (int row = 0; row < NOX_INVENTORY_ROW_COUNT - 1; row++) {
			for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
				nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
					row + NOX_INVENTORY_ROW_COUNT * column];
				if (cell->field_140 && cell->field_0 && cell->field_0->field_27 == thing_type &&
					cell->field_140 < 0x20u) {
					cell->data_4[cell->field_140 - 1] = (uint32_t)net_code;
					cell->field_140++;
					return cell;
				}
			}
		}
	}
	return 0;
}

//----- (004619F0) --------------------------------------------------------
char* sub_4619F0() {
	char* v0;     // edi
	char* result; // eax
	int v2;       // esi
	int v3;       // ecx

	v0 = (char*)&(nox_client_inventory_grid_1050020[0].field_140);
	do {
		result = v0;
		v2 = 4;
		do {
			v3 = 0;
			if ((unsigned char)*result > 0) {
				do {
					*((uint32_t*)result - 1) = 0;
					++v3;
				} while (v3 < (unsigned char)*result);
			}
			result += NOX_INVENTORY_ROW_COUNT * sizeof(nox_inventory_cell_t);
			--v2;
		} while (v2);
		v0 += sizeof(nox_inventory_cell_t);
	} while ((int)v0 <= (int)&(nox_client_inventory_grid_1050020[NOX_INVENTORY_ROW_COUNT - 1].field_140));
	return result;
}

//----- (00461B50) --------------------------------------------------------
void sub_461B50(void) {
	nox_inventory_cell_t* ordered[(NOX_INVENTORY_ROW_COUNT - 1) * NOX_INVENTORY_COL_COUNT];
	size_t ordered_count = 0;
	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT - 1; row++) {
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			ordered[ordered_count++] =
				&nox_client_inventory_grid_1050020[row + NOX_INVENTORY_ROW_COUNT * column];
		}
	}

	for (size_t target_index = 0; target_index < ordered_count;) {
		nox_inventory_cell_t* target = ordered[target_index];
		if (target->field_140) {
			target_index++;
			continue;
		}

		size_t source_index = target_index + 1;
		while (source_index < ordered_count && !ordered[source_index]->field_140) {
			source_index++;
		}
		if (source_index == ordered_count) {
			for (size_t index = target_index; index < ordered_count; index++) {
				ordered[index]->field_132 = 0;
				ordered[index]->field_136 = 0;
			}
			return;
		}

		nox_inventory_cell_t* source = ordered[source_index];
		nox_drawable* source_drawable = source->field_0;
		if (source_drawable && !(source_drawable->flags28 & 0x4000000)) {
			for (size_t merge_index = 0; merge_index < ordered_count && source->field_140; merge_index++) {
				nox_inventory_cell_t* destination = ordered[merge_index];
				if (destination == source || !destination->field_140 || destination->field_140 == 32 ||
					!destination->field_0 || destination->field_0->field_27 != source_drawable->field_27) {
					continue;
				}
				uint32_t* source_codes = &source->field_4;
				uint32_t* destination_codes = &destination->field_4;
				while (source->field_140 && destination->field_140 < 32) {
					source->field_140--;
					destination_codes[destination->field_140] = source_codes[source->field_140];
					destination->field_140++;
				}
			}
			if (!source->field_140) {
				nox_xxx_spriteDelete_45A4B0((uint64_t*)source->field_0);
				source->field_0 = NULL;
				continue;
			}
		}

		*target = *source;
		if (target->field_136) {
			dword_5d4594_1062480 = (uintptr_t)target;
		}
		source->field_140 = 0;
		source->field_0 = NULL;
		source->field_132 = 0;
		target_index++;
	}
}

//----- (00461E60) --------------------------------------------------------
nox_inventory_cell_t* sub_461E60(nox_inventory_cell_t* cell, uint32_t stack_index) {
	if (!cell || stack_index >= cell->field_140) {
		return cell;
	}
	uint32_t* codes = &cell->field_4;
	for (uint32_t index = stack_index; index + 1 < cell->field_140; index++) {
		codes[index] = codes[index + 1];
	}
	cell->field_140--;
	if (!cell->field_140) {
		nox_xxx_spriteDelete_45A4B0((uint64_t*)cell->field_0);
		cell->field_0 = NULL;
	}
	if (cell->field_136) {
		nox_xxx_clientSetAltWeapon_461550(NULL);
		cell->field_136 = 0;
	}
	return cell;
}

//----- (00461EF0) --------------------------------------------------------
nox_inventory_cell_t* nox_inventory_find_cell_native_461EF0(int net_code, uint32_t* stack_index) {
	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT; row++) {
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
				row + NOX_INVENTORY_ROW_COUNT * column];
			uint32_t* codes = &cell->field_4;
			for (uint32_t index = 0; index < cell->field_140; index++) {
				if (codes[index] == (uint32_t)net_code) {
					if (stack_index) {
						*stack_index = index;
					}
					return cell;
				}
			}
		}
	}
	return NULL;
}

char* sub_461EF0(int a1) {
	int row_idx = 0;
	do {
		int col_idx = 0;
		do {
			const nox_inventory_cell_t* p_item = &nox_client_inventory_grid_1050020[row_idx + NOX_INVENTORY_ROW_COUNT * col_idx];
			const int field140_val = p_item->field_140;
			if (field140_val > 0) {
				const uint32_t* p_maybe_stack_items = &p_item->field_4;
				for (int maybe_stack_idx = 0; maybe_stack_idx < field140_val; ++maybe_stack_idx) {
					if (*(p_maybe_stack_items + maybe_stack_idx) == a1) {
						*getMemU32Ptr(0x5D4594, 1049792) = maybe_stack_idx;
						const char* result = (char*)getMemAt(0x5D4594, 1049788);
						*getMemU32Ptr(0x5D4594, 1049788) = p_item;
						return result;
					}
				}
			}
			++col_idx;
		} while (col_idx < NOX_INVENTORY_COL_COUNT);
		++row_idx;
	} while (row_idx <= NOX_INVENTORY_ROW_COUNT);

	return 0;
}

//----- (00461F90) --------------------------------------------------------
nox_drawable* sub_461F90(int net_code) {
	for (size_t slot = 0; slot < 9; slot++) {
		for (nox_drawable* drawable = (nox_drawable*)array_5D4594_1049872[slot]; drawable;
			 drawable = drawable->field_92) {
			if (drawable->field_32 != (uint32_t)net_code) {
				continue;
			}
			nox_drawable* previous = drawable->field_93;
			nox_drawable* next = drawable->field_92;
			if (previous) {
				previous->field_92 = next;
			} else {
				array_5D4594_1049872[slot] = (uintptr_t)next;
			}
			if (next) {
				next->field_93 = previous;
			}
			if ((drawable->flags28 & 0x1000) || nox_xxx_ammoCheck_415880(drawable->field_27) == 2 ||
				nox_xxx_ammoCheck_415880(drawable->field_27) == 128) {
				sub_470D70();
			}
			drawable->field_92 = NULL;
			drawable->field_93 = NULL;
			return drawable;
		}
	}
	return NULL;
}

//----- (004622E0) --------------------------------------------------------
int sub_4622E0(nox_drawable* drawable) {
	uint32_t primary = drawable->flags28;
	uint32_t secondary = drawable->flags29;
	int result;

	if (primary & 0x1000000 && secondary & 2) {
		return 0;
	}
	if (!(primary & 0x2000000) || (result = 1, !(secondary & 1))) {
		if (primary & 0x2000000) {
			if (secondary & 0x144) {
				return 2;
			}
			if (secondary & 0x90) {
				return 3;
			}
			if (secondary & 0x20) {
				return 4;
			}
			if (secondary & 2) {
				return 8;
			}
			if (secondary & 8) {
				return 5;
			}
		}
		if (primary & 0x1000000) {
			if (secondary & 4) {
				return 8;
			}
			return 7;
		}
		if (primary & 0x1000) {
			return 7;
		}
		result = 9;
	}
	return result;
}

//----- (004623B0) --------------------------------------------------------
int nox_xxx_clientEquip_4623B0(nox_drawable* drawable) {
	char v3[3]; // [esp+0h] [ebp-4h]
	v3[0] = 117;
	*(uint16_t*)&v3[1] = nox_xxx_netGetUnitCodeCli_578B00(drawable);
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, v3, 3);
}

//----- (004623E0) --------------------------------------------------------
nox_drawable* sub_4623E0(nox_drawable* drawable, int slot) {
	if (slot < 0 || slot >= 9) {
		return NULL;
	}
	nox_drawable* insertion_point = NULL;
	if (drawable->flags28 & 0x2000000) {
		uint32_t secondary = drawable->flags29;
		if (secondary & 0x140) {
			insertion_point = (nox_drawable*)array_5D4594_1049872[slot];
			if (insertion_point) {
				while (insertion_point->field_92) {
					insertion_point = insertion_point->field_92;
				}
				if ((secondary & 0x40) && (insertion_point->flags28 & 0x2000000) &&
					(insertion_point->flags29 & 0x100)) {
					insertion_point = insertion_point->field_93;
				}
			}
		} else if (secondary & 0x10) {
			insertion_point = (nox_drawable*)array_5D4594_1049872[slot];
			if (insertion_point) {
				while (insertion_point->field_92) {
					insertion_point = insertion_point->field_92;
				}
			}
		}
	}
	if (insertion_point) {
		nox_drawable* next = insertion_point->field_92;
		drawable->field_92 = next;
		if (next) {
			next->field_93 = drawable;
		}
		insertion_point->field_92 = drawable;
		drawable->field_93 = insertion_point;
		return insertion_point;
	}
	drawable->field_93 = NULL;
	nox_drawable* old_head = (nox_drawable*)array_5D4594_1049872[slot];
	drawable->field_92 = old_head;
	if (old_head) {
		old_head->field_93 = drawable;
	}
	array_5D4594_1049872[slot] = (uintptr_t)drawable;
	return old_head;
}

//----- (004624D0) --------------------------------------------------------
int sub_4624D0(int net_code) {
	nox_drawable* equipped = sub_461F90(net_code);
	if (!equipped) {
		return 0;
	}
	nox_inventory_cell_t* source = nox_inventory_find_cell_native_461EF0(net_code, NULL);
	if (!source) {
		return nox_xxx_spriteDelete_45A4B0((uint64_t*)equipped);
	}
	source->field_132 = 0;
	if ((nox_drawable*)dword_5d4594_1062492 != equipped) {
		nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
		nox_drawable* alternate_drawable = alternate ? alternate->field_0 : NULL;
		if ((nox_xxx_ammoCheck_415880(equipped->field_27) & 0xC) && alternate_drawable &&
			nox_xxx_ammoCheck_415880(alternate_drawable->field_27) == 2) {
			alternate->field_136 = 0;
			nox_xxx_clientSetAltWeapon_461550(NULL);
		}
		return nox_xxx_spriteDelete_45A4B0((uint64_t*)equipped);
	}
	dword_5d4594_1062492 = 0;
	nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
	if (alternate && alternate->field_0) {
		dword_5d4594_1062496 = equipped->field_32;
		alternate->field_0->field_32 = alternate->field_4;
		nox_xxx_clientEquip_4623B0(alternate->field_0);
	} else {
		nox_xxx_clientSetAltWeapon_461550(source);
		source->field_136 = 1;
	}
	return nox_xxx_spriteDelete_45A4B0((uint64_t*)equipped);
}

//----- (004625D0) --------------------------------------------------------
int sub_4625D0(nox_window* win, nox_window_data* draw) {
	(void)draw;
	int x;
	int y;
	int width;
	int height;

	if (dword_5d4594_1049864 == 5) {
		return 1;
	}
	nox_client_wndGetPosition_46AA60(win, &x, &y);
	nox_window_get_size(win, &width, &height);
	if (y + height > 0) {
		nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
		nox_inventory_cell_t* cell = (nox_inventory_cell_t*)dword_5d4594_1062480;
		nox_drawable* drawable = cell ? cell->field_0 : NULL;
		if (drawable) {
			drawable->pos.x = x + width / 2;
			drawable->pos.y = y + height / 2;
			if (drawable->draw_func) {
				drawable->draw_func((uint32_t*)getMemAt(0x5D4594, 1049732), drawable);
			}
		}
		wchar2_t* key = (wchar2_t*)sub_42E8E0(35, 1);
		if (key) {
			nox_xxx_drawString_43F6E0(nox_inventory_font, key, x + 22, y + 41);
		}
	}
	return 1;
}

//----- (004626C0) --------------------------------------------------------
double sub_4626C0(const nox_drawable* drawable) {
	if (!drawable || !(drawable->flags28 & 0x13001000)) {
		return 0.0;
	}
	for (int i = 2; i < 4; i++) {
		void* modifier = drawable->item_modifiers[i];
		if (modifier && nox_modifier_effect_getPreHitFunc(modifier) == (void*)nox_xxx_lightngEffect_4E06F0) {
			return nox_modifier_effect_getPreHitFloat(modifier);
		}
	}
	return 0.0;
}

//----- (00462700) --------------------------------------------------------
double sub_462700(const nox_drawable* drawable) {
	if (!drawable || !(drawable->flags28 & 0x13001000)) {
		return 0.0;
	}
	for (int i = 2; i < 4; i++) {
		void* modifier = drawable->item_modifiers[i];
		if (modifier && nox_modifier_effect_getPreHitFunc(modifier) == (void*)nox_xxx_fireEffect_4E0550) {
			return nox_modifier_effect_getPreHitFloat(modifier);
		}
	}
	return 0.0;
}

// 462878: variable 'v3' is possibly undefined
// 46288E: variable 'v4' is possibly undefined
// 4628A4: variable 'v5' is possibly undefined
// 4628CE: variable 'v6' is possibly undefined
// 4628E8: variable 'v7' is possibly undefined

//----- (00463370) --------------------------------------------------------
int sub_463370(nox_window* win, nox_point* pos, uint32_t* out) {
	unsigned int x;
	unsigned int y;
	nox_client_wndGetPosition_46AA60(win, &x, &y);
	out[0] = pos->x - x;
	out[1] = pos->y - y;
	return out[1];
}

//----- (004633B0) --------------------------------------------------------
void sub_4633B0(const nox_drawable* drawable, float* current, float* maximum) {
	*current = drawable->field_73_1;
	*maximum = drawable->field_73_2;
	if (drawable->flags28 & 0x13001000) {
		void* modifier = drawable->item_modifiers[1];
		if (modifier && nox_modifier_effect_getDefendFunc(modifier) == (void*)sub_4E0380) {
			float multiplier = nox_modifier_effect_getDefendFloat(modifier);
			*current = nox_float2int((double)*current * multiplier);
			*maximum = nox_float2int((double)*maximum * multiplier);
		}
	}
}

//----- (00463420) --------------------------------------------------------
int sub_463420(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1050012) = a1;
	return result;
}

//----- (00463430) --------------------------------------------------------
int nox_xxx_inventoryDrawAllMB_463430(nox_window* win, nox_window_data* draw) {
	(void)draw;
	int v1;          // et1
	int v2;          // ebp
	int v3;          // esi
	int v4;          // edi
	int v5;          // eax
	void* v6;        // eax, nox_video_bag_image_t*
	int v7;          // ebp
	int v8;          // eax
	int v9;          // edi
	int v10;         // esi
	int v11;         // esi
	int v12;         // edi
	int v13;         // et1
	int v14;         // et1
	int2 v16;        // [esp+10h] [ebp-28h]
	wchar2_t v17[16]; // [esp+18h] [ebp-20h]

	v1 = dword_587000_136184;
	nox_window_setPos_46A9B0(win->parent, 0, v1);
	nox_client_wndGetPosition_46AA60(win, &v16, &v16.field_4);
	nox_xxx_guiFontHeightMB_43F320(0);
	v2 = 0;
	v3 = v16.field_0 + 10;
	v4 = v16.field_4 + 234;
	do {
		if ((1 << v2) & *getMemU32Ptr(0x5D4594, 1062540) && v2 != 31 && v2 != 30) {
			v5 = nox_xxx_getEnchantSpell_424920(v2);
			v6 = nox_xxx_spellIcon_424A90(v5);
			nox_client_drawImageAt_47D2C0(v6, v3, v4);
			v3 += 35;
		}
		++v2;
	} while (v2 < 32);
	v7 = 0;
	do {
		if (getMemByte(0x5D4594, 1062536) & (unsigned char)(1 << v7)) {
			v8 = sub_413420(1 << v7);
			nox_client_drawImageAt_47D2C0(v8, v3, v4);
			v3 += 35;
		}
		++v7;
	} while (v7 < 6);
	if (nox_common_gameFlags_check_40A5C0(4096) && nox_inventory_extra_lives_anim) {
		v4 += 5;
		v3 += 6;
		nox_image_ref_anim_t* anim = nox_inventory_extra_lives_anim->field_24;
		if (anim && anim->images_sz) {
			nox_client_drawImageAt_47D2C0(anim->images[gameFrame() % anim->images_sz], v3 - 58, v4 - 53);
		}
		nox_swprintf(v17, L"X %d", *getMemU32Ptr(0x5D4594, 1050012));
		nox_xxx_drawSetTextColor_434390(*getMemIntPtr(0x852978, 0));
		nox_xxx_drawString_43F6E0(nox_inventory_font, v17, v3 + 20, v4 + 9);
		*getMemU32Ptr(0x5D4594, 1049812) = v3 - 30;
		*getMemU32Ptr(0x5D4594, 1049816) = v4 - 20;
		*getMemU32Ptr(0x5D4594, 1049820) = v3 + 30;
		*getMemU32Ptr(0x5D4594, 1049824) = v4 + 20;
	}
	if (nox_common_gameFlags_check_40A5C0(4096) && sub_4BFD30()) {
		v9 = v4 + 5;
		v10 = v3 + 66;
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_SHARED_KEY_MODE], v10 - 64, v9 - 58);
		*getMemU32Ptr(0x5D4594, 1049828) = v10 - 30;
		*getMemU32Ptr(0x5D4594, 1049832) = v9 - 20;
		*getMemU32Ptr(0x5D4594, 1049836) = v10 + 30;
		*getMemU32Ptr(0x5D4594, 1049840) = v9 + 20;
	}
	if (getMemByte(0x5D4594, 1049868)) {
		v11 = v16.field_4 + 13;
		v12 = v16.field_0 + 254;
		if (v16.field_4 + 163 > 0) {
			nox_xxx_wndDraw_49F7F0();
			nox_client_copyRect_49F6F0(v12, v11, 260, 150);
			if (getMemByte(0x5D4594, 1049869)) {
				if (getMemByte(0x5D4594, 1049869) == 1) {
					nox_xxx_guiDrawJournal_469D40(v12, v11, *(int*)&dword_5d4594_1062512);
				}
			} else {
				nox_xxx_guiDrawInventoryTray_4643B0(v12, v11);
			}
			sub_49F860();
		}
		if (dword_5d4594_1049864 == 5) {
			sub_4627F0(&v16);
			nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_IDENTIFY_BASE], v16.field_0, v16.field_4);
		} else {
			if (getMemByte(0x5D4594, 1049870)) {
				if (getMemByte(0x5D4594, 1049870) == 1) {
					nox_client_makePlayerStatsDlg_463880(&v16.field_0);
					nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_IDENTIFY_BASE], v16.field_0, v16.field_4);
				}
			} else {
				sub_4BF7E0(&v16);
				nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_BASE], v16.field_0, v16.field_4);
			}
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			nox_xxx_drawStringWrap_43FAF0(0, getMemAt(0x5D4594, 1062588), v16.field_0 + 13, v16.field_4 + 17, 196, 0);
		}
	}
	if (getMemByte(0x5D4594, 1049868) == 1) {
		v14 = dword_587000_136184;
		dword_587000_136184 = v14 + 64;
		if (v14 + 64 > 0) {
			dword_587000_136184 = 0;
			*getMemU8Ptr(0x5D4594, 1049868) = 2;
		}
	} else if (getMemByte(0x5D4594, 1049868) == 3) {
		v13 = dword_587000_136184;
		dword_587000_136184 = v13 - 32;
		if (v13 - 32 <= -225) {
			dword_587000_136184 = -225;
			*getMemU8Ptr(0x5D4594, 1049868) = 0;
			if (getMemByte(0x5D4594, 1049869)) {
				if (getMemByte(0x5D4594, 1049869) == 1) {
					dword_5d4594_1062520 = dword_5d4594_1062512;
				}
			} else {
				dword_5d4594_1062516 = dword_5d4594_1062512;
			}
			*getMemU8Ptr(0x5D4594, 1049869) = 0;
			dword_5d4594_1062512 = dword_5d4594_1062516;
			nox_slider_data_t* slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16395, 0, 850);
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - dword_5d4594_1062512, 0);
			nox_xxx_wndSetIcon_46AE60(nox_inventory_journal_button, 0);
			sub_46AEC0(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_JOURNAL_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9105);
			*getMemU8Ptr(0x5D4594, 1049870) = 0;
			nox_xxx_wndSetIcon_46AE60(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS]);
			sub_46AEC0(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9107);
			nox_window_set_hidden(nox_inventory_current_weapon_window, 0);
		}
	}
	if (sub_467C80() && nox_xxx_playerAnimCheck_4372B0()) {
		sub_467C10();
	}
	return 1;
}

//----- (004643B0) --------------------------------------------------------
int nox_xxx_guiDrawInventoryTray_4643B0(int a1, int a2) {
	int text_width;
	wchar2_t text[16];

	nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY_SPECIAL], a1, a2);
	nox_itow(*(int*)&dword_5d4594_1062552, text, 10);
	nox_xxx_drawSetTextColor_434390(nox_color_yellow_2589772);
	nox_xxx_drawGetStringSize_43F840(nox_inventory_font, text, &text_width, 0, 0);
	nox_xxx_drawString_43F6E0(nox_inventory_font, text, a1 - text_width + 43, a2 + 36);
	if (dword_5d4594_1049864 == 5) {
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY_IDENTIFY_LIT], a1, a2 + 50);
	}
	if (sub_473670()) {
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY_MAP_LIT], a1, a2 + 100);
	}

	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT; row++) {
		int y = a2 - dword_5d4594_1062512 + 50 * row;
		if (y <= a2 - 50) {
			continue;
		}
		if (y > a2 + 150) {
			break;
		}
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY1 + row % 3], a1 + 60, y);
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			int x = a1 + 60 + 50 * column;
			nox_inventory_cell_t* cell =
				&nox_client_inventory_grid_1050020[row + NOX_INVENTORY_ROW_COUNT * column];
			nox_drawable* drawable = cell->field_140 ? cell->field_0 : NULL;
			if (!drawable) {
				continue;
			}

			bool depleted_ammo = false;
			nox_client_drawEnableAlpha_434560(1);
			nox_client_drawSetAlpha_434580(0x40u);
			double amount = (double)drawable->field_73_1;
			double capacity = (double)drawable->field_73_2;
			uint32_t color = 0x80000000;
			if (amount < capacity * *getMemDoublePtr(0x581450, 9608)) {
				color = *getMemU32Ptr(0x85B3FC, 940);
			} else if (amount < capacity * *(double*)&qword_581450_9544) {
				color = nox_color_yellow_2589772;
			}
			if (color != 0x80000000) {
				nox_client_drawSetColor_434460(color);
				nox_client_drawRectFilledOpaque_49CE30(x, y, 50, 50);
			}

			nox_client_drawSetAlpha_434580(0x80u);
			if (cell->field_132) {
				nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_EQUIP_RING], x, y);
			} else if (cell->field_136) {
				nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_QUICK_ITEM_RING], x, y);
			} else if (cell->field_4 == dword_5d4594_1062488) {
				nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
				nox_drawable* alternate_drawable = alternate ? alternate->field_0 : NULL;
				if (alternate_drawable && (alternate_drawable->flags28 & 0x1000000) &&
					(alternate_drawable->flags29 & 0xC)) {
					nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_QUICK_ITEM_RING], x, y);
				}
			}
			nox_client_drawEnableAlpha_434560(0);

			drawable->pos.x = x + 25;
			drawable->pos.y = y + 25;
			if (drawable->draw_func) {
				drawable->draw_func((uint32_t*)getMemAt(0x5D4594, 1049732), drawable);
			}
			if (dword_5d4594_1049864 == 6) {
				if ((drawable->flags28 & 0x13001000) && (drawable->flags28 & 0x1000) &&
					drawable->item_field_112_2 && drawable->item_field_112_0 < drawable->item_field_112_2) {
					depleted_ammo = true;
				}
				if ((drawable->field_73_1 == drawable->field_73_2 || !drawable->field_73_2) && !depleted_ammo) {
					nox_client_drawRectFilledAlpha_49CF10(x, y, 50, 50);
				}
			}
			if (cell->field_140 > 1) {
				nox_swprintf(text, L"%d", cell->field_140);
				nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
				nox_xxx_drawString_43F6E0(nox_inventory_font, text, x + 6, y + 6);
			}
			if ((drawable->flags28 & 0x13001000) && drawable->item_field_112_0 >= 0) {
				nox_swprintf(text, L"%d", drawable->item_field_112_0);
				nox_xxx_drawSetTextColor_434390(nox_color_blue_2650684);
				nox_xxx_drawGetStringSize_43F840(nox_inventory_font, text, &text_width, 0, 0);
				nox_xxx_drawString_43F6E0(nox_inventory_font, text, x - text_width + 44, y + 6);
			}
		}
	}
	return a2 + 150;
}

//----- (00464770) --------------------------------------------------------
int sub_464770(int a1, int a2, unsigned int a3) {
	(void)a1;
	nox_drawable* dragged = NULL;

	if (dword_5d4594_1049864 == 6) {
		return 1;
	}
	switch (a2) {
	case 5:
	case 8:
		return 1;
	case 6:
		dragged = nox_client_inventory_get_dragged();
		if (!dragged) {
			goto LABEL_25;
		}
		if (!nox_xxx_wndPointInWnd_46AAB0((unsigned int*)nox_inventory_window, (unsigned short)a3, a3 >> 16)) {
			goto LABEL_22;
		}
		if (dword_5d4594_1049856) {
			if (dragged->flags28 & 0x1001000) {
				if (dword_5d4594_1062480) {
					nox_client_invAlterWeapon_4672C0();
				} else {
					dword_5d4594_1062492 = (uintptr_t)dragged;
					nox_xxx_clientDequip_464B70(dragged);
				}
			}
		} else if (!(dragged->flags28 & 0x1001000) ||
				   nox_client_inventory_grid_1050020[dword_5d4594_1049800_inventory_click_row_index +
									   NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049796_inventory_click_column_index]
					   .field_132) {
			sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
					   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
		} else {
			if (nox_xxx_ammoCheck_415880(dragged->field_27) == 2) {
				nox_drawable* ammo4 = sub_461600(sub_415840(4));
				nox_drawable* ammo8 = sub_461600(sub_415840(8));
				if (!ammo4 && !ammo8) {
					sub_4649B0(dragged,
							   *(int*)&dword_5d4594_1049796_inventory_click_column_index,
							   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
					nox_xxx_cursorResetDraggedItem_4776A0();
					if (!dword_5d4594_1049856) {
						nox_xxx_spriteDelete_45A4B0(dragged);
					}
					nox_client_inventory_set_dragged(NULL);
					dword_5d4594_1049856 = 0;
					return 1;
				}
			}
			nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
			if (alternate) {
				alternate->field_136 = 0;
			}
			sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
					   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
			alternate = &nox_client_inventory_grid_1050020[dword_5d4594_1049800_inventory_click_row_index +
														 NOX_INVENTORY_ROW_COUNT *
															 dword_5d4594_1049796_inventory_click_column_index];
			nox_xxx_clientSetAltWeapon_461550(alternate);
			alternate->field_136 = 1;
		}
	LABEL_22:
		nox_xxx_cursorResetDraggedItem_4776A0();
		if (!dword_5d4594_1049856) {
			nox_xxx_spriteDelete_45A4B0(dragged);
		}
		nox_client_inventory_set_dragged(NULL);
		dword_5d4594_1049856 = 0;
	LABEL_25:
		nox_xxx_wndClearCaptureMain_46ADE0(nox_inventory_window);
		return 1;
	case 7:
		if (dword_5d4594_1062480) {
			nox_client_invAlterWeapon_4672C0();
		}
		return 1;
	default:
		return 0;
	}
}

//----- (00464B40) --------------------------------------------------------
int sub_464B40(int a1, int a2) { return a1 >= 0 && a1 < 4 && a2 >= 0 && a2 < 21; }

//----- (00464B70) --------------------------------------------------------
int nox_xxx_clientDequip_464B70(nox_drawable* drawable) {
	char v3[3]; // [esp+0h] [ebp-4h]
	v3[0] = 118;
	*(uint16_t*)&v3[1] = nox_xxx_netGetUnitCodeCli_578B00(drawable);
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, v3, 3);
}

//----- (00464BA0) --------------------------------------------------------
int nox_xxx_XorEaxEaxSub_464BA0() { return 0; }

//----- (00464BB0) --------------------------------------------------------
int nox_xxx_inventoryWndProc_464BB0(int a1, int a2) { return a2 != 8 && a2 != 12 && a2 != 16; }

// 464C71: variable 'v4' is possibly undefined
// 464C8B: variable 'v5' is possibly undefined
// 464CD6: variable 'v6' is possibly undefined
// 464CF0: variable 'v7' is possibly undefined
// 464D4C: variable 'v8' is possibly undefined
// 464E24: variable 'v14' is possibly undefined
// 464E64: variable 'v15' is possibly undefined
// 464F2B: variable 'v21' is possibly undefined
// 464F52: variable 'v22' is possibly undefined
// 464F6C: variable 'v23' is possibly undefined
// 465009: variable 'v24' is possibly undefined
// 465020: variable 'v25' is possibly undefined
// 465066: variable 'v26' is possibly undefined
// 465200: variable 'v30' is possibly undefined
// 46521A: variable 'v31' is possibly undefined
// 465281: variable 'v34' is possibly undefined
// 46535F: variable 'v36' is possibly undefined
// 4653B8: variable 'v37' is possibly undefined

//----- (004657B0) --------------------------------------------------------
int nox_xxx_trade_4657B0(short a1) {
	char v2[4]; // [esp+0h] [ebp-4h]

	v2[0] = -55;
	v2[1] = 30;
	*(uint16_t*)&v2[2] = a1;
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, v2, 4);
}

//----- (004657E0) --------------------------------------------------------
char nox_xxx_clientTradeMB_4657E0(uint32_t* a1) {
	int v1; // eax

	v1 = nox_xxx_pointInRect_4281F0((int2*)a1, (int4*)getMemAt(0x587000, 136352));
	if (v1) {
		int i = (a1[1] + dword_5d4594_1062512 - 13) / 50 + NOX_INVENTORY_ROW_COUNT * ((*a1 - 314) / 50);
		uint8_t n = nox_client_inventory_grid_1050020[i].field_140;
		LOBYTE(v1) = n;
		if ((uint8_t)v1) {
			LOBYTE(v1) = nox_xxx_clientTrade_465870(*(uint32_t*)((uint8_t*)&nox_client_inventory_grid_1050020[i] + 4 * n));
		}
	}
	return v1;
}
// 4657F5: variable 'v1' is possibly undefined

//----- (00465870) --------------------------------------------------------
int nox_xxx_clientTrade_465870(short a1) {
	char v2[4]; // [esp+0h] [ebp-4h]

	v2[0] = -55;
	v2[1] = 28;
	*(uint16_t*)&v2[2] = a1;
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, v2, 4);
}

//----- (004658A0) --------------------------------------------------------
void sub_4658A0(nox_window* win, int2* a2) {
	(void)win;
	if (getMemByte(0x5D4594, 1049868) == 2) {
		if (nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136336))) {
			int slot = sub_465990(a2);
			nox_drawable* dragged = slot >= 0 ? (nox_drawable*)array_5D4594_1049872[slot] : NULL;
			nox_client_inventory_set_dragged(dragged);
			dword_5d4594_1049856 = dragged != NULL;
		} else {
			dword_5d4594_1049856 = 0;
			if (nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136368))) {
				if ((a2->field_4 - 13) / 50 == 2) {
					nox_client_toggleMap_473610();
				}
			} else if (nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136352))) {
				dword_5d4594_1049796_inventory_click_column_index = (a2->field_0 - 314) / 50;
				dword_5d4594_1049800_inventory_click_row_index = (a2->field_4 + *(int*)&dword_5d4594_1062512 - 13) / 50;
				if (sub_464B40(*(int*)&dword_5d4594_1049796_inventory_click_column_index,
							   *(int*)&dword_5d4594_1049800_inventory_click_row_index)) {
					nox_xxx_cliInventorySpriteUpd_465A30();
				}
			}
		}
	}
}
// 4658C2: variable 'v2' is possibly undefined
// 4658FF: variable 'v3' is possibly undefined
// 465934: variable 'v5' is possibly undefined

//----- (00465990) --------------------------------------------------------
int sub_465990(uint32_t* a1) {
	int v1;  // eax
	int v2;  // esi
	int v3;  // eax
	int v5;  // eax
	int2 v6; // [esp+4h] [ebp-8h]

	v1 = a1[1] - 15;
	v6.field_0 = *a1 - 11;
	v6.field_4 = v1;
	v2 = 0;
	while (1) {
		v3 = nox_xxx_pointInRect_4281F0(&v6, (int4*)getMemAt(0x587000, 136192 + 16 * v2));
		if (!v3) {
			goto LABEL_6;
		}
		if (v2 == 6) {
			break;
		}
		if (v2) {
			return v2;
		}
		if (array_5D4594_1049872[0]) {
			return 0;
		}
	LABEL_6:
		if (++v2 >= 9) {
			return -1;
		}
	}
	v5 = array_5D4594_1049872[8];
	if (!array_5D4594_1049872[8]) {
		return 5;
	}
	while (!(*(uint32_t*)(v5 + 112) & 0x2000000) || !(*(uint8_t*)(v5 + 116) & 2)) {
		v5 = *(uint32_t*)(v5 + 368);
		if (!v5) {
			return 5;
		}
	}
	return 8;
}
// 4659C7: variable 'v3' is possibly undefined

//----- (00465BE0) --------------------------------------------------------
int nox_xxx_clientDrop_465BE0(int2* a1) {
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (!dragged) {
		return 0;
	}
	short v2; // dx
	char v3[7]; // [esp+0h] [ebp-8h]

	v3[0] = 114;
	*(uint16_t*)&v3[1] = nox_xxx_netGetUnitCodeCli_578B00(dragged);
	v2 = a1->field_4;
	*(uint16_t*)&v3[3] = a1->field_0;
	*(uint16_t*)&v3[5] = v2;
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, v3, 7);
}

//----- (00465C30) --------------------------------------------------------
int nox_xxx_clientKeyEquip_465C30(int a1, int a2) {
	dword_5d4594_1049796_inventory_click_column_index = a1;
	dword_5d4594_1049800_inventory_click_row_index = a2;
	nox_xxx_cliInventorySpriteUpd_465A30();
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (!dragged) {
		return 0;
	}
	nox_xxx_clientEquip_4623B0(dragged);
	return sub_4649B0(dragged, a1, a2);
}

//----- (00465C70) --------------------------------------------------------
void nox_xxx_clientUse_465C70(nox_drawable* drawable) {
	if (drawable) {
		uint8_t message[3];
		message[0] = 116;
		*(uint16_t*)&message[1] = nox_xxx_netGetUnitCodeCli_578B00(drawable);
		nox_netlist_addToMsgListCli_40EBC0(31, 0, message, sizeof(message));
	}
}

//----- (00465CA0) --------------------------------------------------------
int sub_465CA0() {
	nox_window_set_hidden(nox_inventory_identify_window, 0);
	dword_5d4594_1049864 = 5;
	nox_client_setCursorType_477610(6);
	return nox_xxx_wndSetCaptureMain_46ADC0(nox_inventory_window);
}

//----- (00465CD0) --------------------------------------------------------
void sub_465CD0(uint32_t* a1, int a2, int a3, int a4) {
	int2 v7;  // [esp+8h] [ebp-8h]

	(void)a3;
	if (a4 > 0) {
		sub_473970((int2*)a1, &v7);
		nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(a2, NULL);
		if (cell && cell->field_0) {
			int count = a4 < cell->field_140 ? a4 : cell->field_140;
			for (int i = 0; i < count; i++) {
				nox_drawable* dragged = cell->field_0;
				nox_client_inventory_set_dragged(dragged);
				dragged->field_32 = (&cell->field_4)[i];
				if (!sub_4C12C0()) {
					nox_xxx_clientDrop_465BE0(&v7);
				}
				nox_client_inventory_set_dragged(NULL);
			}
		}
	}
}

//----- (00465D50) --------------------------------------------------------
int sub_465D50_draw(nox_window* win) {
	int x;
	int y;

	if (!win || !win->parent) {
		return 1;
	}
	nox_client_wndGetPosition_46AA60(win->parent, &x, &y);
	nox_drawable* drawable = (nox_drawable*)(uintptr_t)sub_4615C0();
	if (drawable) {
		drawable->pos.x = x + 51;
		drawable->pos.y = y + 81;
		if (drawable->draw_func) {
			drawable->draw_func(getMemAt(0x5D4594, 1049732), drawable);
		}
	} else if (!dword_5d4594_1062496 && !dword_5d4594_1062492) {
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_FIST], x + 21, y + 50);
	}
	return 1;
}

//----- (00465DE0) --------------------------------------------------------
int sub_465DE0(int a1) {
	dword_5d4594_1049844 = a1;
	return nox_xxx_j_inventoryNameSignInit_467460();
}
// 467460: using guessed type int nox_xxx_j_inventoryNameSignInit_467460(void);


//----- (00465E00) --------------------------------------------------------
nox_window* nox_xxx_wndCreateInventoryMB_465E00() {
	nox_window* v0; // eax
	int result;     // eax
	nox_window* v5; // eax
	int v6;       // eax
	int v7;       // eax

	nox_xxx_inventoryLoadImages_467050();
	nox_xxx_inventoryNameSignInit_4671E0();
	nox_inventory_font = nox_xxx_guiFontPtrByName_43F360("small");
	*getMemU32Ptr(0x5D4594, 1049732) = 0;
	*getMemU32Ptr(0x5D4594, 1049736) = 0;
	*getMemU32Ptr(0x5D4594, 1049740) = nox_win_width;
	*getMemU32Ptr(0x5D4594, 1049744) = nox_win_height;
	*getMemU32Ptr(0x5D4594, 1049764) = nox_win_width;
	*getMemU32Ptr(0x5D4594, 1049768) = nox_win_height;
	*getMemU32Ptr(0x5D4594, 1049748) = 0;
	*getMemU32Ptr(0x5D4594, 1049752) = 0;
	dword_5d4594_1062452 = nox_window_new(0, 552, 0, 0, 563, 264, 0);
	if (!dword_5d4594_1062452) {
		return 0;
	}
	nox_window_set_all_funcs(dword_5d4594_1062452, 0, nox_xxx_movEax1Sub_4661C0, 0);
	v0 = nox_window_new(dword_5d4594_1062452, 8, 0, 224, nox_win_width, 40, 0);
	nox_window_set_all_funcs(v0, nox_xxx_XorEaxEaxSub_464BA0, nox_xxx_movEax1Sub_4661C0,
							 nox_xxx_inventroryOnHovewerSub_4667E0);
	nox_inventory_window = nox_window_new(dword_5d4594_1062452, 40, 0, 0, 563, 224, sub_466220);
	if (!nox_inventory_window) {
		return 0;
	}
	nox_window_set_all_funcs(nox_inventory_window, sub_464BD0, nox_xxx_inventoryDrawAllMB_463430, sub_466620);
	nox_inventory_window->draw_data.style |= 0x100u;
	nox_inventory_overlay_window = nox_window_new(dword_5d4594_1062452, 40, 0, 0, 1, 1, 0);
	nox_window_set_all_funcs(nox_inventory_overlay_window, sub_464BD0, nox_xxx_movEax1Sub_4661C0, 0);
	nox_inventory_current_weapon_window = nox_window_new(nox_inventory_window, 40, 173, 174, 50, 50, 0);
	nox_window_set_all_funcs(nox_inventory_current_weapon_window, sub_464770, sub_4625D0, sub_4661D0);
	nox_inventory_current_weapon_window->draw_data.style |= 0x100u;
	result = sub_466950(nox_inventory_window);
	if (!result) {
		return 0;
	}

	result = sub_466C40(nox_inventory_window);
	if (!result) {
		return 0;
	}

	result = sub_466ED0(nox_inventory_window);
	if (!result) {
		return 0;
	}

	nox_win_unk5 = nox_window_new(0, 0x408 | NOX_WIN_LAYER_BACK, -1, nox_win_height - 127, 111, 127, 0);
	if (!nox_win_unk5) {
		return 0;
	}

	nox_window_set_all_funcs(nox_win_unk5, nox_xxx_inventoryWndProc_464BB0, nox_xxx_inventoryDrawProc_466580, 0);
	nox_xxx_wndSetIcon_46AE60(nox_win_unk5, nox_xxx_gLoadImg_42F970("CurrentWeapon"));
	nox_xxx_wndSetIconLit_46AEA0(nox_win_unk5, nox_xxx_gLoadImg_42F970("CurrentWeaponLit"));
	nox_xxx_wndSetOffsetMB_46AE40(nox_win_unk5, -1, 0);
	nox_win_init_cur_weapon(nox_win_unk5, 24, 51, 53, 53);
	sub_471160(nox_win_unk5, 79, 40, 20, 127);
	sub_470D70();
	v5 = nox_window_new(nox_win_unk5, 8, 5, 11, 28, 29, 0);
	nox_window_set_all_funcs(v5, sub_466550, nox_xxx_movEax1Sub_4661C0, sub_466160);
	memset(nox_client_inventory_grid_1050020, 0, sizeof(nox_inventory_cell_t) * NOX_INVENTORY_CELLS_MAX);
	if (!dword_5d4594_1062560) {
		dword_5d4594_1062560 = nox_xxx_getTTByNameSpriteMB_44CFC0("Gold");
		*getMemU32Ptr(0x5D4594, 1049728) = nox_xxx_getTTByNameSpriteMB_44CFC0("QuestGoldPile");
		*getMemU32Ptr(0x5D4594, 1049724) = nox_xxx_getTTByNameSpriteMB_44CFC0("QuestGoldChest");
	}
	nox_client_inventory_grid_1050020[1 * NOX_INVENTORY_ROW_COUNT - 1].field_0 =
		nox_new_drawable_for_thing(*(int*)&dword_5d4594_1062560);
	if (nox_client_inventory_grid_1050020[NOX_INVENTORY_ROW_COUNT - 1].field_0) {
		nox_client_inventory_grid_1050020[NOX_INVENTORY_ROW_COUNT - 1].field_140 = 1;
	}
	v6 = dword_5d4594_1062564;
	if (!dword_5d4594_1062564) {
		v6 = nox_xxx_getTTByNameSpriteMB_44CFC0("Identify");
		dword_5d4594_1062564 = v6;
	}
	nox_client_inventory_grid_1050020[2 * NOX_INVENTORY_ROW_COUNT - 1].field_0 = nox_new_drawable_for_thing(v6);
	if (nox_client_inventory_grid_1050020[2 * NOX_INVENTORY_ROW_COUNT - 1].field_0) {
		nox_client_inventory_grid_1050020[2 * NOX_INVENTORY_ROW_COUNT - 1].field_140 = 1;
	}
	v7 = dword_5d4594_1062556;
	if (!dword_5d4594_1062556) {
		v7 = nox_xxx_getTTByNameSpriteMB_44CFC0("AutoMap");
		dword_5d4594_1062556 = v7;
	}
	nox_client_inventory_grid_1050020[3 * NOX_INVENTORY_ROW_COUNT - 1].field_0 = nox_new_drawable_for_thing(v7);
	if (nox_client_inventory_grid_1050020[3 * NOX_INVENTORY_ROW_COUNT - 1].field_0) {
		nox_client_inventory_grid_1050020[3 * NOX_INVENTORY_ROW_COUNT - 1].field_140 = 1;
	}

	return nox_inventory_window;
}

//----- (004661C0) --------------------------------------------------------
int nox_xxx_movEax1Sub_4661C0(nox_window* win, void* data) {
	(void)win;
	(void)data;
	return 1;
}

//----- (00466220) --------------------------------------------------------
int sub_466220(nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	(void)win;
	int result; // eax
	int v5;     // ecx
	int v6;     // ecx
	int v7;     // ecx
	int v8;     // eax
	nox_slider_data_t* slider;

	if (a2 == 16391) {
		switch (nox_xxx_wndGetID_46B0A0((nox_window*)a3)) {
		case 9102:
			if (*(int*)&dword_5d4594_1062512 - 25 >= 0) {
				v5 = *(int*)&dword_5d4594_1062512 - 25 - (*(int*)&dword_5d4594_1062512 - 25) % 50;
			} else {
				v5 = 0;
			}
			dword_5d4594_1062512 = v5;
			slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - v5, 0);
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
			result = 0;
			break;
		case 9103:
			v6 = dword_5d4594_1062512 + 50;
			dword_5d4594_1062512 = v6;
			slider = nox_inventory_scroll_window->widget_data;
			if (v6 <= slider->max) {
				v7 = v6 - v6 % 50;
			} else {
				v7 = slider->max;
			}
			dword_5d4594_1062512 = v7;
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - v7, 0);
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
			result = 0;
			break;
		case 9105:
			v8 = sub_469FA0() - 150;
			if (dword_5d4594_1049864 == 5) {
				return 0;
			}
			if (v8 < 0) {
				v8 = 0;
			}
			*getMemU8Ptr(0x5D4594, 1049869) = 1;
			dword_5d4594_1062516 = dword_5d4594_1062512;
			dword_5d4594_1062512 = dword_5d4594_1062520;
			slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16395, 0, v8);
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - dword_5d4594_1062512, 0);
			nox_xxx_wndSetIcon_46AE60(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_INVENTORY]);
			sub_46AEC0(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_INVENTORY_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9106);
			result = 0;
			break;
		case 9106:
			*getMemU8Ptr(0x5D4594, 1049869) = 0;
			dword_5d4594_1062520 = dword_5d4594_1062512;
			dword_5d4594_1062512 = dword_5d4594_1062516;
			slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16395, 0, 850);
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - dword_5d4594_1062512, 0);
			nox_xxx_wndSetIcon_46AE60(nox_inventory_journal_button, 0);
			sub_46AEC0(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_JOURNAL_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9105);
			result = 0;
			break;
		case 9107:
			if (dword_5d4594_1049864 == 5) {
				return 0;
			}
			*getMemU8Ptr(0x5D4594, 1049870) = 1;
			nox_xxx_wndSetIcon_46AE60(nox_inventory_stats_button, 0);
			sub_46AEC0(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_DOLL_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9108);
			nox_window_set_hidden(nox_inventory_current_weapon_window, 1);
			result = 0;
			break;
		case 9108:
			if (dword_5d4594_1049864 != 5) {
				*getMemU8Ptr(0x5D4594, 1049870) = 0;
				nox_xxx_wndSetIcon_46AE60(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS]);
				sub_46AEC0(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS_LIT]);
				nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9107);
				nox_window_set_hidden(nox_inventory_current_weapon_window, 0);
			}
			return 0;
		case 9111:
			sub_467C10();
			result = 0;
			break;
		default:
			return 0;
		}
	} else if (a2 == 16393) {
		result = 0;
		slider = nox_inventory_scroll_window->widget_data;
		dword_5d4594_1062512 = slider->max - a4;
	} else {
		return 0;
	}
	return result;
}

//----- (00466550) --------------------------------------------------------
int sub_466550(nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	(void)win;
	(void)a3;
	(void)a4;
	if (a2 >= 5) {
		if (a2 <= 6) {
			return 1;
		}
		if (a2 == 7) {
			nox_client_toggleInventory_467C60();
			return 1;
		}
	}
	return 0;
}

//----- (00466580) --------------------------------------------------------
int nox_xxx_inventoryDrawProc_466580(nox_window* win, nox_window_data* draw) {
	unsigned int x;
	unsigned int y;
	nox_client_wndGetPosition_46AA60(win, &x, &y);
	if (getMemByte(0x5D4594, 1049868)) {
		nox_client_drawImageAt_47D2C0(draw->hl_image, x, y);
	} else {
		nox_client_drawImageAt_47D2C0(draw->bg_image, x, y);
	}
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	wchar2_t* key = (wchar2_t*)sub_42E8E0(35, 1);
	if (key) {
		nox_xxx_drawString_43F6E0(nox_inventory_font, key, x + 19, y + 102);
	}
	return 1;
}

//----- (00466620) --------------------------------------------------------
int sub_466620(int a1, int a2, unsigned int a3) {
	wchar2_t* v3; // eax
	int2 a2a;    // [esp+0h] [ebp-8h]

	a2a.field_4 = a3 >> 16;
	a2a.field_0 = (unsigned short)a3;
	v3 = sub_466660(a1, &a2a);
	nox_xxx_cursorSetTooltip_4776B0(v3);
	return 1;
}

// 466676: variable 'v2' is possibly undefined
// 4666FF: variable 'v6' is possibly undefined
// 466737: variable 'v10' is possibly undefined

// 4668C0: variable 'v11' is possibly undefined
// 466900: variable 'v13' is possibly undefined

//----- (00466950) --------------------------------------------------------
int sub_466950(nox_window* parent) {
	nox_window_data draw = {0};
	draw.style = 1;
	draw.win = parent;
	draw.en_image = nox_inventory_images[NOX_INV_IMG_UP];
	draw.hl_image = nox_inventory_images[NOX_INV_IMG_UP_LIT];
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_UP_LIT];
	draw.text_color = 0x80000000;
	nox_inventory_scroll_up_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 522, 2, 20, 25, &draw);
	if (!nox_inventory_scroll_up_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_scroll_up_button, 9102);
	draw.en_image = nox_inventory_images[NOX_INV_IMG_DOWN];
	draw.hl_image = nox_inventory_images[NOX_INV_IMG_DOWN_LIT];
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_DOWN_LIT];
	nox_inventory_scroll_down_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 522, 148, 20, 25, &draw);
	if (!nox_inventory_scroll_down_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_scroll_down_button, 9103);
	memset(&draw, 0, sizeof(draw));
	draw.style = 8;
	draw.win = parent;
	draw.bg_color = 0x80000000;
	draw.en_color = 0x80000000;
	draw.hl_color = 0x80000000;
	draw.dis_color = 0x80000000;
	draw.sel_color = 0x80000000;
	nox_slider_data_t values = {.min = 0, .max = 850};
	nox_inventory_scroll_window = nox_gui_newSlider_4B4EE0(parent, 1033, 524, 42, 16, 91, &draw, (float*)&values);
	if (!nox_inventory_scroll_window) {
		return 0;
	}
	nox_xxx_wndSetWindowProc_46B300(nox_inventory_scroll_window, sub_466BF0);
	nox_xxx_wndSetWindowProc_46B300(nox_inventory_scroll_window->field_100, sub_466BA0);
	nox_inventory_scroll_window->field_100->width = 16;
	nox_inventory_scroll_window->field_100->height = 16;
	nox_xxx_wndSetOffsetMB_46AE40(nox_inventory_scroll_window->field_100, 0, -15);
	sub_4B5700(nox_inventory_scroll_window, 0, 0, nox_inventory_images[NOX_INV_IMG_SLIDER],
			   nox_inventory_images[NOX_INV_IMG_SLIDER_LIT], nox_inventory_images[NOX_INV_IMG_SLIDER_LIT]);
	return 1;
}

//----- (00466BA0) --------------------------------------------------------
int nox_xxx_wndButtonProc_4A7F50(nox_window* win, int ev, int a3, int a4);
int sub_466BA0(nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	int result; // eax

	if (nox_client_inventory_get_dragged()) {
		result = sub_464BD0(win, a2, a3, a4);
	} else {
		result = nox_xxx_wndButtonProc_4A7F50(win, a2, a3, a4);
	}
	return result;
}

//----- (00466BF0) --------------------------------------------------------
int sub_466BF0(nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	int result; // eax

	if (nox_client_inventory_get_dragged()) {
		result = sub_464BD0(win, a2, a3, a4);
	} else {
		result = nox_xxx_wndScrollBoxDraw_4B4BA0((int)(uintptr_t)win, a2, a3, a4);
	}
	return result;
}

//----- (00466C40) --------------------------------------------------------
int sub_466C40(nox_window* parent) {
	nox_window_data draw = {0};
	draw.style = 1;
	draw.win = parent;
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_JOURNAL_LIT];
	draw.img_px = -243;
	draw.img_py = -170;
	nox_inventory_journal_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 243, 170, 34, 34, &draw);
	if (!nox_inventory_journal_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9105);
	nox_gui_winSetFunc96_46B070(nox_inventory_journal_button, sub_466E20);
	memset(&draw, 0, sizeof(draw));
	draw.style = 1;
	draw.win = parent;
	draw.bg_image = nox_inventory_images[NOX_INV_IMG_STATS];
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_STATS_LIT];
	draw.img_px = -5;
	draw.img_py = -186;
	nox_inventory_stats_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 5, 186, 34, 34, &draw);
	if (!nox_inventory_stats_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9107);
	nox_gui_winSetFunc96_46B070(nox_inventory_stats_button, sub_466E20);
	memset(&draw, 0, sizeof(draw));
	draw.style = 1;
	draw.win = parent;
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_CLOSE_LIT];
	draw.img_px = -547;
	draw.img_py = -2;
	nox_inventory_close_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 547, 2, 16, 16, &draw);
	if (!nox_inventory_close_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_close_button, 9111);
	nox_gui_winSetFunc96_46B070(nox_inventory_close_button, sub_466E20);
	return 1;
}

//----- (00466ED0) --------------------------------------------------------
int sub_466ED0(nox_window* parent) {
	nox_inventory_identify_window = nox_new_window_from_file("identify.wnd", 0);
	if (nox_inventory_identify_window) {
		sub_46AB20(nox_inventory_identify_window, 200, 200);
		sub_46B120(nox_inventory_identify_window, parent);
		nox_window_setPos_46A9B0(nox_inventory_identify_window, 51, 15);
		nox_window* child = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9155);
		nox_xxx_wndSetDrawFn_46B340(child, sub_466F50);
		return 1;
	}
	return 0;
}

//----- (00466F50) --------------------------------------------------------
int sub_466F50(nox_window* win, nox_window_data* draw) {
	nox_drawable* drawable = dword_5d4594_1063116;
	if (!drawable) {
		return 1;
	}
	if (drawable->flags28 & 0x13001000) {
		void* def = drawable->flags28 & 0x11001000
			? nox_xxx_getProjectileClassById_413250(drawable->field_27)
			: nox_xxx_equipClothFindDefByTT_413270(drawable->field_27);
		if (def) {
			for (int i = 1; i < 7; i++) {
				uint32_t color = nox_modifier_getColorRGB(def, i);
				nox_draw_setMaterial_4340A0(i, color & 0xFF, (color >> 8) & 0xFF, (color >> 16) & 0xFF);
			}
			for (int i = 0; i < 4; i++) {
				void* modifier = drawable->item_modifiers[i];
				if (modifier) {
					uint32_t color = nox_modifier_effect_getColorRGB(modifier);
					nox_draw_setMaterial_4340A0(nox_modifier_getColorSlot(def, i), color & 0xFF,
						(color >> 8) & 0xFF, (color >> 16) & 0xFF);
				}
			}
		}
	}
	unsigned int x;
	unsigned int y;
	nox_client_wndGetPosition_46AA60(win, &x, &y);
	nox_client_drawImageAt_47D2C0(draw->bg_image, (int)draw->img_px + (int)x, (int)draw->img_py + (int)y);
	return 1;
}

//----- (00467050) --------------------------------------------------------
nox_things_imageRef_t* nox_xxx_inventoryLoadImages_467050() {
	static const char* const names[NOX_INV_IMG_COUNT] = {
		"InventoryBase", "InventoryIdentifyBase", "InventoryTray1", "InventoryTray2", "InventoryTray3",
		"InventoryTraySpecial", "InventoryTrayIdentifyLit", "InventoryTrayMapLit", "InventoryUpButton",
		"InventoryUpButtonLit", "InventoryDownButton", "InventoryDownButtonLit", "InventorySliderButton",
		"InventorySliderButtonLit", "InventoryEquipRing", "InventoryQuickItemRing", "InventoryCloseButtonLit",
		"InventoryJournalButtonLit", "InventoryInventoryButton", "InventoryInventoryButtonLit",
		"InventoryDollButtonLit", "InventoryStatsButton", "InventoryStatsButtonLit", "GUIFist", "SharedKeyMode",
	};
	for (int i = 0; i < NOX_INV_IMG_COUNT; i++) {
		nox_inventory_images[i] = nox_xxx_gLoadImg_42F970(names[i]);
	}
	nox_inventory_extra_lives_anim = nox_xxx_gLoadAnim_42FA20("ExtraLives");
	return nox_inventory_extra_lives_anim;
}

//----- (004672C0) --------------------------------------------------------
void nox_client_invAlterWeapon_4672C0() { // Switch onto secondary weapon
	nox_drawable* local_player = getMemPtr(0x852978, 8);
	if (!local_player) {
		return;
	}
	if (nox_xxx_guiCursor_477600()) {
		return;
	}
	if (sub_461160(1)) {
		return;
	}
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	if (!player) {
		return;
	}
	if (local_player->field_69 == 34) {
		return;
	}
	if (nox_xxx_pointInRect_4281F0((int2*)getMemAt(0x5D4594, 1062572),
									 (int4*)getMemAt(0x587000, 136336)) == 1) {
		nox_xxx_cursorSetDraggedItem_477690(NULL);
	}
	nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
	if (alternate && alternate->field_0) {
		if (nox_xxx_ammoCheck_415880(alternate->field_0->field_27) == 2) {
			nox_drawable* equipped = sub_461600(sub_415840(2));
			if (!equipped) {
				return;
			}
			dword_5d4594_1062492 = (uintptr_t)equipped;
			nox_xxx_clientDequip_464B70(equipped);
			nox_xxx_clientPlaySoundSpecial_452D80(895, 100);
			return;
		}
	}
	for (int bit = 1; bit < 27; ++bit) {
		int mask = 1 << bit;
		if (mask != 2 && (mask & player->field_4)) {
			nox_drawable* equipped = sub_461600(sub_415840(mask));
			if (equipped) {
				dword_5d4594_1062492 = (uintptr_t)equipped;
				nox_xxx_clientDequip_464B70(equipped);
				nox_xxx_clientPlaySoundSpecial_452D80(895, 100);
				return;
			}
		}
	}
	if (alternate && alternate->field_0) {
		alternate->field_0->field_32 = alternate->field_4;
		nox_xxx_clientEquip_4623B0(alternate->field_0);
		nox_xxx_clientPlaySoundSpecial_452D80(895, 100);
	}
}
// 467323: variable 'v2' is possibly undefined

//----- (004673F0) --------------------------------------------------------
int sub_4673F0(int a1, int a2) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1062580) = a1;
	*getMemU32Ptr(0x5D4594, 1062584) = a2;
	return result;
}

//----- (00467410) --------------------------------------------------------
int sub_467410(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1062540) = a1;
	return result;
}

//----- (00467420) --------------------------------------------------------
char sub_467420(char a1) {
	char result; // al

	result = a1;
	*getMemU8Ptr(0x5D4594, 1062536) = a1;
	return result;
}

//----- (00467430) --------------------------------------------------------
unsigned char sub_467430() { return getMemByte(0x5D4594, 1062536); }

//----- (00467440) --------------------------------------------------------
int sub_467440(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1062544) = a1;
	return result;
}

//----- (00467450) --------------------------------------------------------
int sub_467450(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1062548) = a1;
	return result;
}

//----- (00467470) --------------------------------------------------------
int sub_467470(int a1, float a2) {
	int result; // eax

	result = (unsigned char)a1;
	*getMemFloatPtr(0x5D4594, 1063100 + 4 * (unsigned char)a1) = a2;
	return result;
}

//----- (00467490) --------------------------------------------------------
int sub_467490(int a1) {
	int result; // eax

	result = a1;
	dword_5d4594_1062552 = a1;
	return result;
}

//----- (004674A0) --------------------------------------------------------
int sub_4674A0() { return dword_5d4594_1062552; }

//----- (004674B0) --------------------------------------------------------
void nox_window_set_visible_unk5(int visible) {
	if (visible) {
		nox_window_set_hidden(nox_win_unk5, 0);
	} else {
		nox_window_set_hidden(nox_win_unk5, 1);
	}
}

//----- (004674E0) --------------------------------------------------------
void nox_xxx_cliUseCurePoison_4674E0(int a1) {
	if (nox_xxx_guiCursor_477600() != 1) {
		if (!nox_xxx_checkGameFlagPause_413A50()) {
			nox_inventory_cell_t* cell = nox_xxx_cliInventoryFirstItemByTT_467520(a1);
			if (cell && cell->field_0) {
				cell->field_0->field_32 = cell->field_4;
				nox_xxx_clientUse_465C70(cell->field_0);
			}
		}
	}
}

//----- (00467520) --------------------------------------------------------
nox_inventory_cell_t* nox_xxx_cliInventoryFirstItemByTT_467520(int thing_type) {
	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT; row++) {
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
				row + NOX_INVENTORY_ROW_COUNT * column];
			if (cell->field_140 && cell->field_0 && cell->field_0->field_27 == (uint32_t)thing_type) {
				return cell;
			}
		}
	}
	return NULL;
}

//----- (00467590) --------------------------------------------------------
int sub_467590() {
	int result; // eax

	if (dword_8531A0_2576) {
		result = *(char*)(dword_8531A0_2576 + 3684);
	} else {
		result = 1;
	}
	return result;
}

//----- (004675B0) --------------------------------------------------------
int sub_4675B0() { return dword_5d4594_1049864; }

//----- (004675E0) --------------------------------------------------------
short sub_4675E0(int net_code, short field_73_1, short field_73_2) {
	nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(net_code, NULL);
	if (cell && cell->field_0) {
		cell->field_0->field_73_1 = (uint16_t)field_73_1;
		cell->field_0->field_73_2 = (uint16_t)field_73_2;
		return field_73_2;
	}
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (dragged && dragged->field_32 == (uint32_t)net_code) {
		dragged->field_73_1 = (uint16_t)field_73_1;
		dragged->field_73_2 = (uint16_t)field_73_2;
		return field_73_2;
	}
	return 0;
}

//----- (00467650) --------------------------------------------------------
int sub_467650() {
	int result; // eax

	sub_462740();
	dword_5d4594_1049864 = 6;
	nox_client_setCursorType_477610(8);
	result = sub_467C80();
	if (!result) {
		result = sub_467BB0();
	}
	return result;
}

//----- (00467680) --------------------------------------------------------
void sub_467680() {
	if (dword_5d4594_1049864 == 6) {
		dword_5d4594_1049864 = 0;
	}
}

//----- (004676A0) --------------------------------------------------------
nox_window* nox_xxx_wndGetHandle_4676A0() { return dword_5d4594_1062452; }

//----- (004676D0) --------------------------------------------------------
nox_drawable* sub_4676D0(int net_code) {
	nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(net_code, NULL);
	if (cell) {
		return cell->field_0;
	}
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (!dragged || dragged->field_32 != (uint32_t)net_code) {
		return NULL;
	}
	return dragged;
}

//----- (00467700) --------------------------------------------------------
int sub_467700(int a1) {
	nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(a1, NULL);
	if (cell) {
		return cell->field_140;
	}
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (dragged && dragged->field_32 == (uint32_t)a1) {
		return 1;
	}
	return 0;
}

//----- (00467740) --------------------------------------------------------
int sub_467740(int a1) {
	dword_5d4594_1062488 = a1;
	return a1;
}

//----- (00467810) --------------------------------------------------------
int sub_467810(int a1, int a2) {
	if (a1 < 0 || a2 < 0 || a1 >= 4 || a2 >= 20) {
		return 0;
	}
	return nox_client_inventory_grid_1050020[a2 + NOX_INVENTORY_ROW_COUNT * a1].field_140;
}

//----- (00467850) --------------------------------------------------------
int sub_467850(int a1) {
	nox_inventory_cell_t* cell = nox_xxx_cliInventoryFirstItemByTT_467520(a1);
	if (cell) {
		return cell->field_140;
	}
	return 0;
}

//----- (00467870) --------------------------------------------------------
char* sub_467870(int a1, int a2) {
	if (a1 < 0 || a2 < 0 || a1 >= 4 || a2 >= 20) {
		return 0;
	}
	return &(nox_client_inventory_grid_1050020[a2 + NOX_INVENTORY_ROW_COUNT * a1].field_4);
}

//----- (004678B0) --------------------------------------------------------
int sub_4678B0() {
	nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
	return alternate ? alternate->field_4 : 0;
}

//----- (004678C0) --------------------------------------------------------
int sub_4678C0() { return dword_5d4594_1062488; }

//----- (004678D0) --------------------------------------------------------
nox_drawable* sub_4678D0() {
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	if (!player) {
		return NULL;
	}
	nox_drawable* equipped = NULL;
	for (int bit = 1; bit < 27; bit++) {
		int mask = 1 << bit;
		if (mask != 2 && (mask & player->field_4)) {
			equipped = sub_461600(sub_415840(mask));
			if (equipped) {
				break;
			}
		}
	}
	if (!equipped) {
		return NULL;
	}
	nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(equipped->field_32, NULL);
	return cell ? cell->field_0 : NULL;
}

//----- (00467930) --------------------------------------------------------
nox_drawable* sub_467930(int net_code, int field_73_1, int field_73_2) {
	if (!net_code) {
		return NULL;
	}
	nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(net_code, NULL);
	if (!cell || !cell->field_0) {
		return NULL;
	}
	nox_drawable* drawable = cell->field_0;
	drawable->item_field_112_0 = (int16_t)field_73_1;
	drawable->item_field_112_2 = (int16_t)field_73_2;
	if (cell->field_132 == 1) {
		sub_470D90(field_73_1, field_73_2);
	}
	return drawable;
}

//----- (00467980) --------------------------------------------------------
int sub_467980() {
	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT; row++) {
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
				row + NOX_INVENTORY_ROW_COUNT * column];
			if (cell->field_0) {
				nox_xxx_spriteDelete_45A4B0(cell->field_0);
				cell->field_0 = NULL;
			}
			cell->field_140 = 0;
			cell->field_132 = 0;
			cell->field_136 = 0;
		}
	}
	nox_client_inventory_set_dragged(NULL);
	dword_5d4594_1049856 = 0;
	sub_462740();
	dword_5d4594_1049864 = 0;
	nox_xxx_clientSetAltWeapon_461550(0);
	dword_5d4594_1062488 = 0;
	memset(array_5D4594_1049872, 0, sizeof(array_5D4594_1049872)); // equipped weapon array
	dword_5d4594_1062492 = 0;
	dword_5d4594_1062496 = 0;
	*getMemU8Ptr(0x5D4594, 1062536) = 0;
	*getMemU32Ptr(0x5D4594, 1062540) = 0;
	*getMemU32Ptr(0x5D4594, 1062544) = 0;
	*getMemU32Ptr(0x5D4594, 1062548) = 0;
	dword_5d4594_1062552 = 0;
	sub_472310();
	dword_587000_136184 = -225;
	*getMemU8Ptr(0x5D4594, 1049868) = 0;
	*getMemU8Ptr(0x5D4594, 1049869) = 0;
	dword_5d4594_1062516 = 0;
	dword_5d4594_1062520 = 0;
	dword_5d4594_1062512 = 0;
	nox_slider_data_t* slider = nox_inventory_scroll_window->widget_data;
	nox_window_call_field_94(nox_inventory_scroll_window, 16395, 0, 850);
	nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - dword_5d4594_1062512, 0);
	nox_xxx_wndSetIcon_46AE60(nox_inventory_journal_button, 0);
	sub_46AEC0(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_JOURNAL_LIT]);
	nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9105);
	*getMemU8Ptr(0x5D4594, 1049870) = 0;
	nox_xxx_wndSetIcon_46AE60(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS]);
	sub_46AEC0(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS_LIT]);
	nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9107);
	return nox_window_set_hidden(nox_inventory_current_weapon_window, 0);
}

//----- (00467B00) --------------------------------------------------------
int sub_467B00(int a1, int a2) {
	int v2;            // ebx
	unsigned char* v3; // ebp
	int i;             // esi
	int v5;            // edi
	int v6;            // eax
	int v8;            // [esp+10h] [ebp-8h]
	unsigned char* v9; // [esp+14h] [ebp-4h]

	v2 = 0;
	v8 = 0;
	v9 = nox_client_inventory_grid_1050020;
	do {
		v3 = v9;
		for (i = 0; i < 4; ++i) {
			v5 = sub_467810(i, v2);
			if (!v5) {
				++v8;
				v3 += NOX_INVENTORY_ROW_COUNT * sizeof(nox_inventory_cell_t);
				continue;
			}
			if (*(uint32_t*)(*(uint32_t*)v3 + 108) == a1) {
				v6 = 31;
				if (*(uint8_t*)(*(uint32_t*)v3 + 112) & 0x10) {
					v6 = nox_common_gameFlags_check_40A5C0(6144) ? 9 : 3;
				}
				if (!(*(uint32_t*)(*(uint32_t*)v3 + 112) & 0x4000000) && a2 + v5 <= v6) {
					++v8;
				}
			}
			v3 += NOX_INVENTORY_ROW_COUNT * sizeof(nox_inventory_cell_t);
		}
		++v2;
		v9 += sizeof(nox_inventory_cell_t);
	} while ((int)v9 < (int)&nox_client_inventory_grid_1050020[NOX_INVENTORY_ROW_COUNT - 1]);
	return v8;
}

//----- (00467BB0) --------------------------------------------------------
int sub_467BB0() {
	int result; // eax

	result = nox_gui_xxx_check_446360();
	if (!result) {
		result = sub_4AE3D0();
		if (!result) {
			result = nox_xxx_guiCursor_477600();
			if (!result) {
				result = nox_xxx_playerAnimCheck_4372B0();
				if (!result) {
					result = nox_xxx_get_57AF20();
					if (!result) {
						if (getMemByte(0x5D4594, 1049868) == 3 || !getMemByte(0x5D4594, 1049868)) {
							*getMemU8Ptr(0x5D4594, 1049868) = 1;
							nox_xxx_clientPlaySoundSpecial_452D80(789, 100);
						}
						result = dword_5d4594_1062516;
						dword_5d4594_1062512 = dword_5d4594_1062516;
					}
				}
			}
		}
	}
	return result;
}

//----- (00467C10) --------------------------------------------------------
int sub_467C10() {
	if (dword_5d4594_1049864 == 6) {
		return 1;
	}
	if (!sub_467C80()) {
		return 0;
	}
	*getMemU8Ptr(0x5D4594, 1049868) = 3;
	nox_xxx_clientPlaySoundSpecial_452D80(790, 100);
	if (dword_5d4594_1049864 == 5) {
		sub_462740();
	}
	sub_467CD0();
	return 1;
}

//----- (00467C60) --------------------------------------------------------
int nox_client_toggleInventory_467C60() {
	int result; // eax

	if (sub_467C80()) {
		result = sub_467C10();
	} else {
		result = sub_467BB0();
	}
	return result;
}

//----- (00467C80) --------------------------------------------------------
int sub_467C80() { return getMemByte(0x5D4594, 1049868) == 1 || getMemByte(0x5D4594, 1049868) == 2; }

//----- (00467CA0) --------------------------------------------------------
int sub_467CA0() {
	int result; // eax

	result = sub_467C80();
	if (!result) {
		dword_5d4594_1062516 = 0;
		result = nox_inventory_scroll_window != 0;
		if (nox_inventory_scroll_window) {
			nox_slider_data_t* slider = nox_inventory_scroll_window->widget_data;
			result = nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max, 0);
		}
	}
	return result;
}

//----- (00467CD0) --------------------------------------------------------
int sub_467CD0() {
	int handled = 0;
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (dragged) {
		if (!dword_5d4594_1049856 &&
			!sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
						*(int*)&dword_5d4594_1049800_inventory_click_row_index)) {
			nox_modifier_attrs_t attrs = {
				.modifiers = {dragged->item_modifiers[0], dragged->item_modifiers[1],
							  dragged->item_modifiers[2], dragged->item_modifiers[3]},
				.field_16 = (uint16_t)dragged->item_field_112_0 |
							 ((uint32_t)(uint16_t)dragged->item_field_112_2 << 16),
			};
			nox_xxx_spritePickup_461660(dragged->field_32, dragged->field_27, &attrs);
			nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(dragged->field_32, NULL);
			if (cell) {
				cell->field_132 = 0;
				for (size_t slot = 0; slot < 9; slot++) {
					for (nox_drawable* equipped = (nox_drawable*)array_5D4594_1049872[slot]; equipped;
						 equipped = equipped->field_92) {
						if (equipped->field_32 == dragged->field_32) {
							cell->field_132 = 1;
							if (cell->field_136) {
								nox_xxx_clientSetAltWeapon_461550(NULL);
								cell->field_136 = 0;
							}
							break;
						}
					}
				}
			}
		}
		nox_client_inventory_set_dragged(NULL);
		dword_5d4594_1049856 = 0;
		nox_xxx_cursorResetDraggedItem_4776A0();
		handled = 1;
	}
	nox_window* capture = nox_xxx_wndGetCaptureMain_46AE00();
	if (nox_window_is_child(nox_xxx_wndGetHandle_4676A0(), capture) == 1) {
		nox_xxx_wndClearCaptureMain_46ADE0(capture);
	}
	return handled;
}

//----- (00469B90) --------------------------------------------------------
int sub_469B90(int* a1) {
	int result; // eax

	*getMemU32Ptr(0x587000, 142296) = *a1;
	*getMemU32Ptr(0x587000, 142300) = a1[1];
	result = a1[2];
	*getMemU32Ptr(0x587000, 142304) = a1[2];
	return result;
}

//----- (00469BB0) --------------------------------------------------------
char* nox_xxx_getAmbientColor_469BB0() { return (char*)getMemAt(0x587000, 142296); }

//----- (00469FA0) --------------------------------------------------------
int sub_469FA0() { return *getMemU32Ptr(0x5D4594, 1064848); }

//----- (0046A430) --------------------------------------------------------
void nox_client_chatStart_46A430(int a1) {
	if (!nox_common_gameFlags_check_40A5C0(2048)) {
		if (!dword_5d4594_1064868) {
			*(uint16_t*)dword_5d4594_1064864 = 0;
			*(uint16_t*)((unsigned char*)dword_5d4594_1064864 + 1052) = 0;
			nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1064856);
			sub_46C690(dword_5d4594_1064856);
			nox_xxx_windowFocus_46B500(dword_5d4594_1064860);
			dword_5d4594_1064868 = 1;
			*getMemU32Ptr(0x5D4594, 1064872) = a1;
		}
	}
}

//----- (0046A4A0) --------------------------------------------------------
int sub_46A4A0() { return dword_5d4594_1064868; }

//----- (0046A4B0) --------------------------------------------------------
size_t nox_xxx_cmdSayDo_46A4B0(wchar2_t* a1, int a2) {
	uint32_t* v2;      // ebp
	size_t v3;         // edi
	size_t result;     // eax
	const wchar2_t* v5; // edi
	char v6;           // al
	int v7;            // eax
	char v8[520];      // [esp+Ch] [ebp-208h]

	v2 = nox_xxx_netSpriteByCodeDynamic_45A6F0(nox_player_netCode_85319C);
	v3 = nox_wcsspn(a1, L" ");
	result = nox_wcslen(a1);
	if (v3 != result) {
		v5 = &a1[v3];
		v8[0] = -88; // MSG_TEXT_MESSAGE
		*(uint16_t*)&v8[9] = 0;
		*(uint16_t*)&v8[1] = nox_player_netCode_85319C;
		v8[3] = 0;
		if (nox_xxx_cliCanTalkMB_4100F0((short*)a1)) {
			v6 = v8[3] | 2;
		} else {
			v6 = v8[3] | 4;
		}
		v8[3] = v6;
		if (a2) {
			v8[3] |= 1u;
		}
		v8[8] = nox_wcslen(v5) + 1;
		if (v8[3] & 4) {
			nox_wcscpy((wchar2_t*)&v8[11], v5);
			v7 = 2;
		} else {
			nox_sprintf(&v8[11], "%S", v5);
			v7 = 1;
		}
		if (v2) {
			*(uint16_t*)&v8[4] = *((uint16_t*)v2 + 6);
			*(uint16_t*)&v8[6] = *((uint16_t*)v2 + 8);
		} else {
			*(uint16_t*)&v8[6] = -1;
			*(uint16_t*)&v8[4] = -1;
		}
		result = nox_netlist_addToMsgListCli_40EBC0(31, 0, v8, v7 * (unsigned char)v8[8] + 11);
	}
	return result;
}

//----- (0046A5D0) --------------------------------------------------------
int sub_46A5D0(nox_window* win, nox_window_data* draw) {
	int v2;  // ecx
	bool v3; // sf
	int v5;  // [esp+4h] [ebp-8h]
	int v6;  // [esp+8h] [ebp-4h]

	v5 = 0;
	v6 = 0;
	nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1064856);
	nox_xxx_windowFocus_46B500(dword_5d4594_1064860);
	nox_xxx_drawGetStringSize_43F840(0, dword_5d4594_1064864, &v5, 0, 0);
	nox_xxx_drawGetStringSize_43F840(0, (unsigned char*)dword_5d4594_1064864 + 512, &v6, 0, 0);
	v3 = v5 + v6 - 90 < 0;
	v5 += v6 + 10;
	v2 = v5;
	if (v5 < 100) {
		v2 = 100;
		v5 = v2;
	} else if (v5 > 320) {
		v2 = 320;
		v5 = v2;
	}
	nox_window_setPos_46A9B0(dword_5d4594_1064856, (nox_win_width - v2) / 2,
		(int)dword_5d4594_1064856->off_y);
	sub_46AB20(win, v5, 20);
	if (sizeof(void*) == 4) {
		return nox_xxx_wndEditDrawNoImage_488160((int)(uintptr_t)win, (int)(uintptr_t)draw);
	}
	return 1;
}

//----- (0046A6A0) --------------------------------------------------------
int sub_46A6A0() {
	if (wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1064856)) {
		return 0;
	}
	if (nox_xxx_wndGetFocus_46B4F0() == dword_5d4594_1064860) {
		nox_xxx_windowFocus_46B500(0);
	}
	nox_xxx_wnd_46C6E0(dword_5d4594_1064856);
	nox_window_set_hidden(dword_5d4594_1064856, 1);
	dword_5d4594_1064856->flags &= ~8u;
	dword_5d4594_1064860->flags &= ~8u;
	set_dword_5d4594_3799468(1);
	dword_5d4594_1064868 = 0;
	return 1;
}

//----- (0046A730) --------------------------------------------------------
nox_window* sub_46A730() {
	nox_window* result;

	*getMemU32Ptr(0x5D4594, 1064876) = nox_win_width / 2;
	*getMemU32Ptr(0x5D4594, 1064880) = 2 * nox_win_height / 3;
	result = nox_new_window_from_file("GuiChat.wnd", sub_46A820);
	dword_5d4594_1064856 = result;
	if (result) {
		nox_window_setPos_46A9B0(result, *getMemIntPtr(0x5D4594, 1064876), *getMemIntPtr(0x5D4594, 1064880));
		result = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1064856, 9201);
		dword_5d4594_1064860 = result;
		if (result) {
			// The native entry-field constructor already installs a width-safe
			// renderer. Keep it instead of the PE32 draw callback.
			nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1064860, sub_46A7E0);
			result = dword_5d4594_1064856;
			dword_5d4594_1064864 = dword_5d4594_1064860->widget_data;
		}
	}
	return result;
}

//----- (0046A7E0) --------------------------------------------------------
int sub_46A7E0(nox_window* win, int a2, int a3, int a4) {
	if (a2 != 21 || a3 != 1) {
		return nox_xxx_wndEditProc_487D70(win, a2, a3, a4);
	}
	if (a4 == 2) {
		nox_xxx_consoleEsc_49B7A0();
	}
	return 1;
}

//----- (0046A820) --------------------------------------------------------
int sub_46A820(nox_window* win, int a2, int a3, int a4) {
	if (a2 == 16415) {
		if (*(uint16_t*)((unsigned char*)dword_5d4594_1064864 + 1052)) {
			nox_xxx_cmdSayDo_46A4B0(dword_5d4594_1064864, *getMemIntPtr(0x5D4594, 1064872));
		}
		sub_46A6A0();
	}
	return 0;
}

//----- (0046A860) --------------------------------------------------------
int sub_46A860() {
	int result = 0;

	if (dword_5d4594_1064856) {
		result = nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1064856);
		dword_5d4594_1064856 = 0;
	}
	dword_5d4594_1064860 = 0;
	dword_5d4594_1064864 = 0;
	dword_5d4594_1064868 = 0;
	*getMemU32Ptr(0x5D4594, 1064872) = 0;
	return result;
}

//----- (0046A8A0) --------------------------------------------------------
int nox_xxx_wndRetNULL_46A8A0() { return 0; }

//----- (0046A8B0) --------------------------------------------------------
int nox_xxx_wndRetNULL_0_46A8B0() { return 0; }

//----- (0046AE10) --------------------------------------------------------
int sub_46AE10(nox_window* win, int enabled) {
	if (!win) {
		return 0;
	}
	if (enabled) {
		win->draw_data.field_0 |= 2;
	} else {
		win->draw_data.field_0 &= ~2u;
	}
	return 1;
}

//----- (0046AE40) --------------------------------------------------------
int nox_xxx_wndSetOffsetMB_46AE40(nox_window* win, int x, int y) {
	if (win) {
		win->draw_data.img_px = x;
		win->draw_data.img_py = y;
	}
	return (intptr_t)win;
}

//----- (0046AE60) --------------------------------------------------------
int nox_xxx_wndSetIcon_46AE60(nox_window* win, nox_video_bag_image_t* image) {
	if (!win) {
		return -2;
	}
	win->draw_data.bg_image = image;
	return 0;
}

//----- (0046AEA0) --------------------------------------------------------
int nox_xxx_wndSetIconLit_46AEA0(nox_window* win, nox_video_bag_image_t* image) {
	if (!win) {
		return -2;
	}
	win->draw_data.hl_image = image;
	return 0;
}

//----- (0046AEC0) --------------------------------------------------------
int sub_46AEC0(nox_window* win, nox_video_bag_image_t* image) {
	if (!win) {
		return -2;
	}
	win->draw_data.sel_image = image;
	return 0;
}

//----- (0046AEE0) --------------------------------------------------------
int sub_46AEE0(int a1, int a2) {
	nox_window_call_field_94(a1, 16385, a2, 0);
	return 0;
}

//----- (0046AF00) --------------------------------------------------------
wchar2_t* sub_46AF00(nox_window* win) {
	int v1; // ecx

	if (!win) {
		return NULL;
	}
	v1 = win->draw_data.style;
	if (v1 & 0x800) {
		return (wchar2_t*)nox_window_call_field_94(win, 16386, 0, 0);
	}
	if ((v1 & 0x80u) != 0) {
		return (wchar2_t*)nox_window_call_field_94(win, 16413, 0, 0);
	} else {
		return NULL;
	}
}

//----- (0046AF40) --------------------------------------------------------
void* sub_46AF40(nox_window* win) {
	if (!win) {
		return NULL;
	}
	return win->draw_data.font;
}

//----- (0046AF80) --------------------------------------------------------
int nox_gui_windowCopyDrawData_46AF80(nox_window* win, const void* p) {
	if (!win) {
		return -2;
	}
	if (!p) {
		return -3;
	}
	memcpy(&win->draw_data, p, sizeof(nox_window_data));
	return 0;
}

//----- (0046B630) --------------------------------------------------------
nox_window* sub_46B630(nox_window* a1p, int a2, int a3) {
	nox_window* result; // eax
	nox_window* i;      // esi
	nox_window* v5;     // ecx
	intptr_t v6;        // edx
	intptr_t j;         // edi

	result = a1p;
	if (a1p) {
	LABEL_2:
		for (i = result->field_100; i; i = i->prev) {
			v5 = i->parent;
			v6 = i->off_x;
			for (j = i->off_y; v5; v5 = v5->parent) {
				v6 += v5->off_x;
				j += v5->off_y;
			}
			if (a2 >= v6 && a2 <= v6 + i->width && a3 >= j && a3 <= j + i->height &&
				!(i->flags & NOX_WIN_HIDDEN)) {
				result = i;
				goto LABEL_2;
			}
		}
	}
	return result;
}

//----- (0046C2A0) --------------------------------------------------------
int nox_xxx_wnd_46C2A0(nox_window* a1p) {
	int a1 = a1p;
	int v2; // eax

	if (!a1) {
		return 1;
	}
	if (*(uint8_t*)(a1 + 4) & 0x10) {
		return 1;
	}
	v2 = *(uint32_t*)(a1 + 396);
	if (v2) {
		while (!(*(uint8_t*)(v2 + 4) & 0x10)) {
			v2 = *(uint32_t*)(v2 + 396);
			if (!v2) {
				return 0;
			}
		}
		return 1;
	}
	return 0;
}

//----- (0046DB80) --------------------------------------------------------
int sub_46DB80() {
	int i;      // esi
	int result; // eax

	for (i = 0; i < 8; i += 4) {
		nox_window_call_field_94(*getMemU32Ptr(0x5D4594, 1090060 + i), 16399, 1, 0);
		nox_window_call_field_94(*getMemU32Ptr(0x5D4594, 1090068 + i), 16399, 1, 0);
		nox_window_call_field_94(*getMemU32Ptr(0x5D4594, 1090076 + i), 16399, 1, 0);
		nox_window_call_field_94(*getMemU32Ptr(0x5D4594, 1090084 + i), 16399, 1, 0);
		result = nox_window_call_field_94(*getMemU32Ptr(0x5D4594, 1090092 + i), 16399, 1, 0);
	}
	return result;
}

//----- (0046DC00) --------------------------------------------------------
int sub_46DC00(int a1, unsigned char a2, int a3) {
	nox_window_call_field_94(a1, 16397, a3, a2);
	return 1;
}

//----- (0046DC30) --------------------------------------------------------
int sub_46DC30(int a1, unsigned char a2, wchar2_t* a3, ...) {
	va_list va; // [esp+10h] [ebp+10h]

	va_start(va, a3);
	nox_vswprintf((wchar2_t*)getMemAt(0x5D4594, 1089000), a3, va);
	return sub_46DC00(a1, a2, (int)getMemAt(0x5D4594, 1089000));
}

//----- (0046DCC0) --------------------------------------------------------
char* sub_46DCC0() {
	char* result;      // eax
	signed int v1;     // ebx
	unsigned int v2;   // esi
	unsigned int v3;   // edi
	int v4;            // ebp
	char* k;           // esi
	int v6;            // eax
	char* l;           // eax
	int v8;            // ecx
	int v9;            // ebx
	int v10;           // esi
	char* v11;         // edi
	int m;             // ebp
	unsigned char v13; // dl
	int v14;           // eax
	int v15;           // ecx
	int v16;           // edi
	uint32_t* v17;     // eax
	uint32_t* v18;     // edi
	char* v19;         // eax
	int v20;           // edx
	unsigned char v21; // al
	int v22;           // ecx
	unsigned int v23;  // esi
	char* i;           // eax
	int v25;           // ecx
	unsigned int v26;  // ebp
	int v27;           // esi
	char* v28;         // edi
	unsigned int j;    // ebx
	unsigned char v30; // dl
	int v31;           // eax
	int v32;           // ecx
	int v33;           // ecx
	int v34;           // edi
	uint32_t* v35;     // eax
	uint32_t* v36;     // edi
	char* v37;         // eax
	int v38;           // edx
	unsigned char v39; // al
	int v40;           // ecx
	unsigned int v41;  // [esp+0h] [ebp-8h]
	wchar2_t* v42;      // [esp+4h] [ebp-4h]
	wchar2_t* v43;      // [esp+4h] [ebp-4h]

	if (dword_5d4594_1090120 == 5) {
		v23 = nox_common_playerInfoCount_416F40();
		v43 = (wchar2_t*)v23;
		*getMemU8Ptr(0x5D4594, 1090117) = 0;
		*getMemU8Ptr(0x5D4594, 1090118) = 0;
		if (nox_common_gameFlags_check_40A5C0(1) &&
			nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
			v43 = (wchar2_t*)--v23;
		}
		for (i = nox_common_playerInfoGetFirst_416EA0(); i; i = nox_common_playerInfoGetNext_416EE0((int)i)) {
			v25 = *((uint32_t*)i + 920);
			if (!(v25 & 1) || v25 & 0x20) {
				if (!*((uint32_t*)i + 527)) {
					*((uint32_t*)i + 527) = 0x8000000;
				}
			} else {
				*((uint32_t*)i + 527) |= 0x80000000;
			}
		}
		v26 = 0;
		if (getMemByte(0x5D4594, 1090117) < v23) {
			v27 = (int)v43;
			do {
				v28 = nox_common_playerInfoGetFirst_416EA0();
				for (j = -1; v28; v28 = nox_common_playerInfoGetNext_416EE0((int)v28)) {
					if ((v28[2064] != 31 || !nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) &&
						*((int*)v28 + 527) >= v26 && !sub_46E1E0(*((uint32_t*)v28 + 515)) && *((int*)v28 + 527) < j) {
						j = *((uint32_t*)v28 + 527);
						v27 = (int)v28;
					}
				}
				v30 = getMemByte(0x5D4594, 1090117);
				v31 = 80 * getMemByte(0x5D4594, 1090117);
				*getMemU32Ptr(0x5D4594, 1084192 + v31) = *(uint32_t*)(v27 + 2060);
				v32 = *(uint32_t*)(v27 + 3680);
				if (!(v32 & 1) || v32 & 0x20) {
					v33 = *(uint32_t*)(v27 + 2108);
					if (v33 == 0x8000000) {
						*getMemU32Ptr(0x5D4594, 1084196 + v31) = 0;
					} else {
						*getMemU32Ptr(0x5D4594, 1084196 + v31) = v33;
						++*getMemU8Ptr(0x5D4594, 1090118);
					}
				} else {
					*getMemU32Ptr(0x5D4594, 1084196 + v31) = *(uint32_t*)(v27 + 2108) + 0x80000000;
				}
				*getMemU32Ptr(0x5D4594, 1084200 + v31) = *(unsigned short*)(v27 + 2148);
				*getMemU32Ptr(0x5D4594, 1084208 + v31) = *(uint32_t*)(v27 + 3680);
				v34 = 80 * v30;
				*getMemU32Ptr(0x5D4594, 1084204 + v34) = sub_46E080(v27);
				nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1084132 + v34), (const wchar2_t*)(v27 + 4704));
				sub_46E170((wchar2_t*)getMemAt(0x5D4594, 1084132 + 80 * getMemByte(0x5D4594, 1090117)));
				*getMemU8Ptr(0x5D4594, 1084188 + 80 * getMemByte(0x5D4594, 1090117)) = *(uint8_t*)(v27 + 2251);
				v35 = nox_xxx_objGetTeamByNetCode_418C80(*(uint32_t*)(v27 + 2060));
				v36 = v35;
				if (v35 && nox_xxx_servObjectHasTeam_419130((int)v35)) {
					v37 = nox_xxx_getTeamByID_418AB0(*((unsigned char*)v36 + 4));
					if (v37) {
						v38 = (unsigned char)v37[57];
						v39 = getMemByte(0x5D4594, 1090117);
						*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = v38;
					} else {
						v39 = getMemByte(0x5D4594, 1090117);
						*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = -1;
					}
				} else {
					v39 = getMemByte(0x5D4594, 1090117);
					*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = -1;
				}
				*getMemU8Ptr(0x5D4594, 1090117) = v39 + 1;
				v26 = j;
			} while ((unsigned char)(v39 + 1) < (unsigned int)v43);
		}
		for (result = nox_common_playerInfoGetFirst_416EA0(); result;
			 result = nox_common_playerInfoGetNext_416EE0((int)result)) {
			v40 = *((uint32_t*)result + 920);
			if (!(v40 & 1) || v40 & 0x20) {
				if (*((uint32_t*)result + 527) == 0x8000000) {
					*((uint32_t*)result + 527) = 0;
				}
			} else {
				*((uint32_t*)result + 527) &= 0x7FFFFFFFu;
			}
		}
	} else if (nox_common_gameFlags_check_40A5C0(1024)) {
		v1 = 0x80000000;
		v41 = nox_common_playerInfoCount_416F40();
		v2 = nox_xxx_getTeamCounter_417DD0();
		*getMemU8Ptr(0x5D4594, 1090116) = 0;
		v42 = (wchar2_t*)v2;
		*getMemU8Ptr(0x5D4594, 1090117) = 0;
		*getMemU8Ptr(0x5D4594, 1090118) = 0;
		if (nox_common_gameFlags_check_40A5C0(1) &&
			nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
			--v41;
		}
		if (getMemByte(0x5D4594, 1090116) < v2) {
			v3 = v2;
			do {
				v4 = 0x7FFFFFFF;
				for (k = nox_server_teamFirst_418B10(); k; k = nox_server_teamNext_418B60((int)k)) {
					if (*((int*)k + 13) >= v1 && !sub_46E130((unsigned char)k[57]) && *((int*)k + 13) < v4) {
						v4 = *((uint32_t*)k + 13);
						v3 = (unsigned int)k;
					}
				}
				v6 = 56 * getMemByte(0x5D4594, 1090116);
				*getMemU32Ptr(0x5D4594, 1087252 + v6) = *(uint32_t*)(v3 + 52);
				*getMemU32Ptr(0x5D4594, 1087248 + v6) = *(unsigned char*)(v3 + 57);
				*getMemU8Ptr(0x5D4594, 1087256 + v6) = *(uint8_t*)(v3 + 56);
				nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1087204 + v6), (const wchar2_t*)v3);
				sub_46E170((wchar2_t*)getMemAt(0x5D4594, 1087204 + 56 * getMemByte(0x5D4594, 1090116)));
				++*getMemU8Ptr(0x5D4594, 1090116);
				v1 = v4;
			} while (getMemByte(0x5D4594, 1090116) < (unsigned int)v42);
		}
		for (l = nox_common_playerInfoGetFirst_416EA0(); l; l = nox_common_playerInfoGetNext_416EE0((int)l)) {
			v8 = *((uint32_t*)l + 920);
			if (v8 & 1 && !(v8 & 0x20)) {
				*((uint32_t*)l + 535) += 0xFFFF;
			}
		}
		v9 = -1;
		if (getMemByte(0x5D4594, 1090117) < v41) {
			v10 = (int)v42;
			do {
				v11 = nox_common_playerInfoGetFirst_416EA0();
				for (m = 0x7FFFFFFF; v11; v11 = nox_common_playerInfoGetNext_416EE0((int)v11)) {
					if ((v11[2064] != 31 || !nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) &&
						*((int*)v11 + 535) >= v9 && !sub_46E1E0(*((uint32_t*)v11 + 515)) && *((int*)v11 + 535) < m) {
						m = *((uint32_t*)v11 + 535);
						v10 = (int)v11;
					}
				}
				v13 = getMemByte(0x5D4594, 1090117);
				v14 = 80 * getMemByte(0x5D4594, 1090117);
				*getMemU32Ptr(0x5D4594, 1084192 + v14) = *(uint32_t*)(v10 + 2060);
				v15 = *(uint32_t*)(v10 + 3680);
				if (!(v15 & 1) || v15 & 0x20) {
					*getMemU32Ptr(0x5D4594, 1084196 + v14) = *(uint32_t*)(v10 + 2140);
					++*getMemU8Ptr(0x5D4594, 1090118);
				} else {
					*getMemU32Ptr(0x5D4594, 1084196 + v14) = *(uint32_t*)(v10 + 2140) - 0xFFFF;
				}
				*getMemU32Ptr(0x5D4594, 1084200 + v14) = *(unsigned short*)(v10 + 2148);
				*getMemU32Ptr(0x5D4594, 1084208 + v14) = *(uint32_t*)(v10 + 3680);
				v16 = 80 * v13;
				*getMemU32Ptr(0x5D4594, 1084204 + v16) = sub_46E080(v10);
				nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1084132 + v16), (const wchar2_t*)(v10 + 4704));
				sub_46E170((wchar2_t*)getMemAt(0x5D4594, 1084132 + 80 * getMemByte(0x5D4594, 1090117)));
				*getMemU8Ptr(0x5D4594, 1084188 + 80 * getMemByte(0x5D4594, 1090117)) = *(uint8_t*)(v10 + 2251);
				v17 = nox_xxx_objGetTeamByNetCode_418C80(*(uint32_t*)(v10 + 2060));
				v18 = v17;
				if (v17 && nox_xxx_servObjectHasTeam_419130((int)v17)) {
					v19 = nox_xxx_getTeamByID_418AB0(*((unsigned char*)v18 + 4));
					if (v19) {
						v20 = (unsigned char)v19[57];
						v21 = getMemByte(0x5D4594, 1090117);
						*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = v20;
					} else {
						v21 = getMemByte(0x5D4594, 1090117);
						*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = -1;
					}
				} else {
					v21 = getMemByte(0x5D4594, 1090117);
					*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = -1;
				}
				*getMemU8Ptr(0x5D4594, 1090117) = v21 + 1;
				v9 = m;
			} while ((unsigned char)(v21 + 1) < v41);
		}
		for (result = nox_common_playerInfoGetFirst_416EA0(); result;
			 result = nox_common_playerInfoGetNext_416EE0((int)result)) {
			v22 = *((uint32_t*)result + 920);
			if (v22 & 1) {
				if (!(v22 & 0x20)) {
					*((uint32_t*)result + 535) -= 0xFFFF;
				}
			}
		}
	} else {
		result = sub_46E4E0();
	}
	return result;
}

//----- (0046E080) --------------------------------------------------------
int sub_46E080(int a1) {
	int v1;       // eax
	uint32_t* v3; // eax

	if (nox_common_gameFlags_check_40A5C0(32)) {
		v1 = *(uint32_t*)(a1 + 2060);
		if (v1 == *getMemU16Ptr(0x5D4594, 1090128)) {
			return 2;
		}
		if (v1 == *getMemU16Ptr(0x5D4594, 1090130)) {
			return 3;
		}
	} else if (nox_common_gameFlags_check_40A5C0(64)) {
		if (*(uint32_t*)(a1 + 2060) == *getMemU16Ptr(0x5D4594, 1090132)) {
			return 4;
		}
	} else if (nox_common_gameFlags_check_40A5C0(16)) {
		v3 = nox_xxx_netSpriteByCodeDynamic_45A6F0(*(uint32_t*)(a1 + 2060));
		if (v3) {
			if (nox_client_drawable_testBuff_4356C0((int)v3, 30)) {
				return 1;
			}
		}
	}
	return 0;
}

//----- (0046E130) --------------------------------------------------------
int sub_46E130(int a1) {
	int v1;           // eax
	unsigned char* i; // ecx

	v1 = 0;
	if (!getMemByte(0x5D4594, 1090116)) {
		return 0;
	}
	for (i = getMemAt(0x5D4594, 1087248); *(uint32_t*)i != a1; i += 56) {
		if (++v1 >= (unsigned int)getMemByte(0x5D4594, 1090116)) {
			return 0;
		}
	}
	return 1;
}

//----- (0046E170) --------------------------------------------------------
unsigned short* sub_46E170(wchar2_t* a1) {
	if (!a1 || !*getMemIntPtr(0x5D4594, 1084036)) {
		return 0;
	}
	unsigned short* v1; // esi
	size_t v2;          // edi

	v1 = a1;
	v2 = nox_wcslen(a1);
	int a1a = 0;
	nox_xxx_drawGetStringSize_43F840(0, v1, &a1a, 0, 0);
	if (a1a == 0) {
		return 0;
	}
	if ((a1a + 5) > *getMemIntPtr(0x5D4594, 1084036)) {
		unsigned short* v3 = &v1[v2];
		do {
			*v3 = 0;
			--v3;
			nox_xxx_drawGetStringSize_43F840(0, v1, (int*)&a1a, 0, 0);
		} while ((a1a + 5) > *getMemIntPtr(0x5D4594, 1084036));
	}
	return v1;
}

//----- (0046E1E0) --------------------------------------------------------
int sub_46E1E0(int a1) {
	int v1;           // eax
	unsigned char* i; // ecx

	v1 = 0;
	if (!getMemByte(0x5D4594, 1090117)) {
		return 0;
	}
	for (i = getMemAt(0x5D4594, 1084192); *(uint32_t*)i != a1; i += 80) {
		if (++v1 >= (unsigned int)getMemByte(0x5D4594, 1090117)) {
			return 0;
		}
	}
	return 1;
}

//----- (0046E4E0) --------------------------------------------------------
char* sub_46E4E0() {
	int v0;            // ebx
	unsigned int v1;   // esi
	unsigned int v2;   // edi
	signed int v3;     // ebp
	char* i;           // esi
	int v5;            // eax
	char* j;           // eax
	int v7;            // ecx
	int v8;            // ebx
	int v9;            // esi
	char* v10;         // edi
	signed int k;      // ebp
	unsigned char v12; // dl
	int v13;           // eax
	int v14;           // ecx
	int v15;           // edi
	uint32_t* v16;     // eax
	uint32_t* v17;     // edi
	char* v18;         // eax
	int v19;           // edx
	unsigned char v20; // al
	char* result;      // eax
	int v22;           // ecx
	unsigned int v23;  // [esp+10h] [ebp-8h]
	wchar2_t* v24;      // [esp+14h] [ebp-4h]

	v0 = 0x7FFFFFFF;
	v1 = nox_xxx_getTeamCounter_417DD0();
	v24 = (wchar2_t*)v1;
	v23 = nox_common_playerInfoCount_416F40();
	*getMemU8Ptr(0x5D4594, 1090116) = 0;
	*getMemU8Ptr(0x5D4594, 1090117) = 0;
	*getMemU8Ptr(0x5D4594, 1090118) = 0;
	if (nox_common_gameFlags_check_40A5C0(1) && nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
		--v23;
	}
	if (getMemByte(0x5D4594, 1090116) < v1) {
		v2 = v1;
		do {
			v3 = 0x80000000;
			for (i = nox_server_teamFirst_418B10(); i; i = nox_server_teamNext_418B60((int)i)) {
				if (*((int*)i + 13) <= v0 && !sub_46E130((unsigned char)i[57]) && *((int*)i + 13) > v3) {
					v3 = *((uint32_t*)i + 13);
					v2 = (unsigned int)i;
				}
			}
			v5 = 56 * getMemByte(0x5D4594, 1090116);
			*getMemU32Ptr(0x5D4594, 1087252 + v5) = *(uint32_t*)(v2 + 52);
			*getMemU32Ptr(0x5D4594, 1087248 + v5) = *(unsigned char*)(v2 + 57);
			*getMemU8Ptr(0x5D4594, 1087256 + v5) = *(uint8_t*)(v2 + 56);
			nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1087204 + v5), (const wchar2_t*)v2);
			sub_46E170((wchar2_t*)getMemAt(0x5D4594, 1087204 + 56 * getMemByte(0x5D4594, 1090116)));
			++*getMemU8Ptr(0x5D4594, 1090116);
			v0 = v3;
		} while (getMemByte(0x5D4594, 1090116) < (unsigned int)v24);
	}
	for (j = nox_common_playerInfoGetFirst_416EA0(); j; j = nox_common_playerInfoGetNext_416EE0((int)j)) {
		v7 = *((uint32_t*)j + 920);
		if (v7 & 1 && !(v7 & 0x20)) {
			*((uint32_t*)j + 534) -= 0xFFFF;
		}
	}
	v8 = 0x7FFFFFFF;
	if (getMemByte(0x5D4594, 1090117) < v23) {
		v9 = (int)v24;
		do {
			v10 = nox_common_playerInfoGetFirst_416EA0();
			for (k = 0x80000000; v10; v10 = nox_common_playerInfoGetNext_416EE0((int)v10)) {
				if ((v10[2064] != 31 || !nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) &&
					*((int*)v10 + 534) <= v8 && !sub_46E1E0(*((uint32_t*)v10 + 515)) && *((int*)v10 + 534) > k) {
					k = *((uint32_t*)v10 + 534);
					v9 = (int)v10;
				}
			}
			v12 = getMemByte(0x5D4594, 1090117);
			v13 = 80 * getMemByte(0x5D4594, 1090117);
			*getMemU32Ptr(0x5D4594, 1084192 + v13) = *(uint32_t*)(v9 + 2060);
			v14 = *(uint32_t*)(v9 + 3680);
			if (!(v14 & 1) || v14 & 0x20) {
				*getMemU32Ptr(0x5D4594, 1084196 + v13) = *(uint32_t*)(v9 + 2136);
				++*getMemU8Ptr(0x5D4594, 1090118);
			} else {
				*getMemU32Ptr(0x5D4594, 1084196 + v13) = *(uint32_t*)(v9 + 2136) + 0xFFFF;
			}
			*getMemU32Ptr(0x5D4594, 1084200 + v13) = *(unsigned short*)(v9 + 2148);
			v15 = 80 * v12;
			*getMemU32Ptr(0x5D4594, 1084204 + v15) = sub_46E080(v9);
			*getMemU32Ptr(0x5D4594, 1084208 + v15) = *(uint32_t*)(v9 + 3680);
			nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1084132 + v15), (const wchar2_t*)(v9 + 4704));
			sub_46E170((wchar2_t*)getMemAt(0x5D4594, 1084132 + 80 * getMemByte(0x5D4594, 1090117)));
			*getMemU8Ptr(0x5D4594, 1084188 + 80 * getMemByte(0x5D4594, 1090117)) = *(uint8_t*)(v9 + 2251);
			v16 = nox_xxx_objGetTeamByNetCode_418C80(*(uint32_t*)(v9 + 2060));
			v17 = v16;
			if (v16 && nox_xxx_servObjectHasTeam_419130((int)v16)) {
				v18 = nox_xxx_getTeamByID_418AB0(*((unsigned char*)v17 + 4));
				if (v18) {
					v19 = (unsigned char)v18[57];
					v20 = getMemByte(0x5D4594, 1090117);
					*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = v19;
				} else {
					v20 = getMemByte(0x5D4594, 1090117);
					*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = -1;
				}
			} else {
				v20 = getMemByte(0x5D4594, 1090117);
				*getMemU32Ptr(0x5D4594, 1084184 + 80 * getMemByte(0x5D4594, 1090117)) = -1;
			}
			*getMemU8Ptr(0x5D4594, 1090117) = v20 + 1;
			v8 = k;
		} while ((unsigned char)(v20 + 1) < v23);
	}
	for (result = nox_common_playerInfoGetFirst_416EA0(); result;
		 result = nox_common_playerInfoGetNext_416EE0((int)result)) {
		v22 = *((uint32_t*)result + 920);
		if (v22 & 1) {
			if (!(v22 & 0x20)) {
				*((uint32_t*)result + 534) += 0xFFFF;
			}
		}
	}
	return result;
}

//----- (0046F060) --------------------------------------------------------
int sub_46F060() { return 0; }

//----- (0046F070) --------------------------------------------------------
int nox_xxx_Proc_46F070() { return 0; }

//----- (0046FAE0) --------------------------------------------------------
void sub_46FAE0() {
	int yTop; // [esp+0h] [ebp-8h]
	int v1;   // [esp+4h] [ebp-4h]

	nox_client_wndGetPosition_46AA60(*(uint32_t**)getMemAt(0x5D4594, 1090060 + 4 * *getMemU32Ptr(0x5D4594, 1088996)),
									 &v1, &yTop);
	yTop +=
		dword_587000_145672 *
			*(short*)(*(uint32_t*)(*getMemU32Ptr(0x5D4594, 1090060 + 4 * *getMemU32Ptr(0x5D4594, 1088996)) + 32) + 2) +
		*(short*)(*(uint32_t*)(*getMemU32Ptr(0x5D4594, 1090060 + 4 * *getMemU32Ptr(0x5D4594, 1088996)) + 32) + 2) / 2;
	nox_client_drawSetColor_434460(nox_color_yellow_2589772);
	nox_xxx_drawPointMB_499B70(v1 + 1, yTop, 3);
}

//----- (0046FE60) --------------------------------------------------------
unsigned char sub_46FE60(int a1) {
	unsigned char result; // al
	unsigned char v2;     // [esp+8h] [ebp-4h]

	result = 0;
	v2 = 0;
	if (!getMemByte(0x5D4594, 1090116)) {
		return 0;
	}
	while (*getMemU32Ptr(0x5D4594, 1087248 + 56 * v2) != a1) {
		v2 = ++result;
		if (result >= getMemByte(0x5D4594, 1090116)) {
			return 0;
		}
	}
	return result;
}

//----- (0046FEB0) --------------------------------------------------------
unsigned char sub_46FEB0(unsigned char a1) {
	return getMemByte(0x587000, 145584 + 8 * (getMemByte(0x5D4594, 1087256 + 56 * a1) % 10));
}

//----- (0046FEE0) --------------------------------------------------------
char sub_46FEE0() {
	int v0;            // eax
	char v1;           // bl
	unsigned char* i;  // edx
	int v4;            // esi
	int v5;            // ecx
	unsigned char* v6; // eax
	int v7;            // ecx
	unsigned char* v8; // eax

	v0 = 0;
	v1 = 1;
	if (!getMemByte(0x5D4594, 1090117)) {
		return 0;
	}
	for (i = getMemAt(0x5D4594, 1084192); *(uint32_t*)i != nox_player_netCode_85319C; i += 80) {
		if (++v0 >= (unsigned int)getMemByte(0x5D4594, 1090117)) {
			return 0;
		}
	}
	v4 = *getMemU32Ptr(0x5D4594, 1084196 + 80 * v0);
	if (!nox_common_gameFlags_check_40A5C0(1024)) {
		v7 = getMemByte(0x5D4594, 1090117);
		if (getMemByte(0x5D4594, 1090117)) {
			v8 = getMemAt(0x5D4594, 1084196);
			do {
				if (*(uint32_t*)v8 > v4) {
					++v1;
				}
				v8 += 80;
				--v7;
			} while (v7);
		}
		return v1;
	}
	v5 = getMemByte(0x5D4594, 1090117);
	if (!getMemByte(0x5D4594, 1090117)) {
		return v1;
	}
	v6 = getMemAt(0x5D4594, 1084196);
	do {
		if (*(uint32_t*)v6 < v4) {
			++v1;
		}
		v6 += 80;
		--v5;
	} while (v5);
	return v1;
}

//----- (0046FF70) --------------------------------------------------------
char sub_46FF70(int a1) {
	char v1;           // bl
	int v2;            // ecx
	unsigned char* v3; // eax
	int v5;            // ecx
	unsigned char* v6; // eax

	v1 = 1;
	if (!nox_common_gameFlags_check_40A5C0(1024)) {
		v5 = getMemByte(0x5D4594, 1090116);
		if (getMemByte(0x5D4594, 1090116)) {
			v6 = getMemAt(0x5D4594, 1087252);
			do {
				if (*(uint32_t*)v6 > a1) {
					++v1;
				}
				v6 += 56;
				--v5;
			} while (v5);
		}
		return v1;
	}
	v2 = getMemByte(0x5D4594, 1090116);
	if (!getMemByte(0x5D4594, 1090116)) {
		return v1;
	}
	v3 = getMemAt(0x5D4594, 1087252);
	do {
		if (*(uint32_t*)v3 < a1) {
			++v1;
		}
		v3 += 56;
		--v2;
	} while (v2);
	return v1;
}

//----- (0046FFD0) --------------------------------------------------------
unsigned char sub_46FFD0() {
	unsigned char result; // al
	int v1;               // ebx
	unsigned char* v2;    // ebp
	char* v3;             // eax
	unsigned char* v4;    // esi
	unsigned char v5;     // al
	unsigned char v6;     // di
	int v7;               // edx
	int v8;               // eax
	int v9;               // eax
	int v10;              // eax
	wchar2_t* v11;         // eax
	int v12;              // eax
	int v13;              // edx
	int v14;              // eax
	unsigned char v15;    // cl
	float v16;            // [esp+0h] [ebp-4Ch]
	int v17;              // [esp+14h] [ebp-38h]
	int v18;              // [esp+18h] [ebp-34h]
	int v19;              // [esp+1Ch] [ebp-30h]
	int v20;              // [esp+20h] [ebp-2Ch]
	int v21;              // [esp+24h] [ebp-28h]
	int v22;              // [esp+28h] [ebp-24h]
	int v23;              // [esp+2Ch] [ebp-20h]
	int v24;              // [esp+30h] [ebp-1Ch]
	float v25;            // [esp+34h] [ebp-18h]
	int v26;              // [esp+38h] [ebp-14h]
	int v27;              // [esp+3Ch] [ebp-10h]
	int v28;              // [esp+40h] [ebp-Ch]
	int v29;              // [esp+44h] [ebp-8h]
	int v30;              // [esp+48h] [ebp-4h]

	sub_46DB80();
	sub_46F8F0(0, 0);
	result = getMemByte(0x5D4594, 1090117);
	v23 = 0;
	if (getMemByte(0x5D4594, 1090117)) {
		v1 = 1;
		v2 = getMemAt(0x5D4594, 1084184);
		do {
			v3 = nox_common_playerInfoGetByID_417040(*((uint32_t*)v2 + 2));
			v4 = (unsigned char*)v3;
			if (v3 && *((uint32_t*)v3 + 1198)) {
				if (v2[24] & 1) {
					LOBYTE(v21) = 9;
				} else if (*(int*)v2 == -1) {
					LOBYTE(v21) = 4;
				} else {
					v5 = sub_46FE60(*(uint32_t*)v2);
					LOBYTE(v21) = sub_46FEB0(v5);
				}
				if (*((uint32_t*)v2 + 2) == nox_player_netCode_85319C) {
					dword_587000_145672 = *(short*)(*(uint32_t*)(*getMemU32Ptr(0x5D4594, 1090060) + 32) + 46);
					*getMemU32Ptr(0x5D4594, 1088996) = 0;
				}
				v6 = v21;
				sub_46DC30(*getMemIntPtr(0x5D4594, 1090060), v21, (wchar2_t*)getMemAt(0x587000, 147828), v2 - 52);
				sub_46DC30(*getMemIntPtr(0x5D4594, 1090076), v6, (wchar2_t*)getMemAt(0x587000, 147836), v4[4816]);
				v7 = *((uint32_t*)v2 + 2);
				v8 = v4[2282];
				LOBYTE(v22) = 4;
				if (v7 == nox_player_netCode_85319C) {
					v9 = sub_470CD0();
					v24 = v9;
					v25 = (double)v9;
					v10 = sub_470CC0();
					v24 = v10;
					v16 = (double)v10 / v25 * 100.0;
					v8 = nox_float2int(v16);
				}
				if (v8 > 25) {
					if (v8 <= 50) {
						LOBYTE(v22) = 15;
					}
				} else {
					LOBYTE(v22) = 6;
				}
				sub_46DC30(*getMemIntPtr(0x5D4594, 1090084), v22, (wchar2_t*)getMemAt(0x587000, 147844), v8);
				sub_46DC30(*getMemIntPtr(0x5D4594, 1090092), v6, (wchar2_t*)getMemAt(0x587000, 147856),
						   *getMemU32Ptr(0x5D4594, 1084056 + 4 * v2[4]));
				v11 = sub_46FB50(*((uint32_t*)v2 + 5), &v26);
				sub_46DC60(*getMemIntPtr(0x5D4594, 1090068), v26, (int)v11);
				if (v4[4824]) {
					nox_client_wndGetPosition_46AA60(*(uint32_t**)getMemAt(0x5D4594, 1090068), &v19, &v18);
					nox_window_get_size(*getMemIntPtr(0x5D4594, 1090068), &v28, &v27);
					v12 = *(uint32_t*)(*getMemU32Ptr(0x5D4594, 1090068) + 32);
					v19 += 5;
					v18 = v18 + *(short*)(v12 + 2) / 2 + *(short*)(v12 + 2) * v1 - 1;
					nox_client_drawSetColor_434460(nox_color_white_2523948);
					nox_video_drawCircleColored_4C3270(v19, v18, 2, nox_color_white_2523948);
					nox_client_drawAddPoint_49F500(v19 + 2, v18);
					nox_client_drawAddPoint_49F500(v19 + 9, v18);
					nox_client_drawLineFromPoints_49E4B0();
					nox_client_drawAddPoint_49F500(v19 + 9, v18);
					nox_client_drawAddPoint_49F500(v19 + 9, v18 + 3);
					nox_client_drawLineFromPoints_49E4B0();
					nox_client_drawAddPoint_49F500(v19 + 7, v18);
					nox_client_drawAddPoint_49F500(v19 + 7, v18 + 2);
					nox_client_drawLineFromPoints_49E4B0();
				}
				if (v4[4825]) {
					nox_client_wndGetPosition_46AA60(*(uint32_t**)getMemAt(0x5D4594, 1090068), &v17, &v20);
					nox_window_get_size(*getMemIntPtr(0x5D4594, 1090068), &v30, &v29);
					v13 = v17 + 5;
					v14 = *(uint32_t*)(*getMemU32Ptr(0x5D4594, 1090068) + 32);
					v15 = v4[4824];
					v17 += 5;
					if (v15 == 1) {
						v17 = v13 + 15;
					}
					v20 = v20 + *(short*)(v14 + 2) / 2 + *(short*)(v14 + 2) * v1 - 1;
					nox_client_drawSetColor_434460(nox_color_yellow_2589772);
					nox_video_drawCircleColored_4C3270(v17, v20, 2, nox_color_yellow_2589772);
					nox_client_drawAddPoint_49F500(v17 + 2, v20);
					nox_client_drawAddPoint_49F500(v17 + 9, v20);
					nox_client_drawLineFromPoints_49E4B0();
					nox_client_drawAddPoint_49F500(v17 + 9, v20);
					nox_client_drawAddPoint_49F500(v17 + 9, v20 + 3);
					nox_client_drawLineFromPoints_49E4B0();
					nox_client_drawAddPoint_49F500(v17 + 7, v20);
					nox_client_drawAddPoint_49F500(v17 + 7, v20 + 2);
					nox_client_drawLineFromPoints_49E4B0();
				}
				++v1;
			}
			result = v23 + 1;
			v2 += 80;
			++v23;
		} while (v23 < getMemByte(0x5D4594, 1090117));
		dword_587000_145664 = 1;
	} else {
		dword_587000_145664 = 1;
	}
	return result;
}

//----- (00470580) --------------------------------------------------------
int sub_470580() {
	return dword_5d4594_1090120 && wndIsShown_nox_xxx_wndIsShown_46ACC0(*(int*)&dword_5d4594_1090048) != 1;
}

//----- (004705B0) --------------------------------------------------------
void sub_4705B0() {
	if (dword_5d4594_1090048) {
		if (wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1090048)) {
			nox_window_set_hidden(dword_5d4594_1090048, 0);
		}
		dword_5d4594_1090120 = 0;
		sub_4703F0();
	}
}

//----- (004705F0) --------------------------------------------------------
char sub_4705F0(char a1, char a2, short a3) {
	char result; // al

	result = a2;
	if (a2 == 1) {
		result = a1;
		if (a1 != 2 && a1) {
			if (a1 == 1) {
				result = a3;
				*getMemU16Ptr(0x5D4594, 1090128) = a3;
			}
		} else {
			*getMemU16Ptr(0x5D4594, 1090128) = 0;
		}
	} else if (a2 == 2) {
		result = a1;
		if (a1 != 2 && a1) {
			if (a1 == 1) {
				*getMemU16Ptr(0x5D4594, 1090130) = a3;
			}
		} else {
			*getMemU16Ptr(0x5D4594, 1090130) = 0;
		}
	}
	return result;
}

//----- (00470650) --------------------------------------------------------
char sub_470650(char a1, short a2) {
	char result; // al

	result = a1;
	if (a1 && a1 != 1) {
		if (a1 == 4 || a1 == 2) {
			result = a2;
			*getMemU16Ptr(0x5D4594, 1090132) = a2;
		}
	} else {
		*getMemU16Ptr(0x5D4594, 1090132) = 0;
	}
	return result;
}

//----- (00470680) --------------------------------------------------------
int sub_470680() {
	int result; // eax

	result = 0;
	*getMemU16Ptr(0x5D4594, 1090128) = 0;
	*getMemU16Ptr(0x5D4594, 1090130) = 0;
	*getMemU16Ptr(0x5D4594, 1090132) = 0;
	return result;
}

//----- (004706A0) --------------------------------------------------------
int sub_4706A0() { return dword_5d4594_1090048 && dword_5d4594_1090120; }

//----- (00470A90) --------------------------------------------------------
int nox_xxx_playerGet_470A90() { return dword_5d4594_1096252; }

//----- (00470AA0) --------------------------------------------------------
void nox_xxx_cliShowHideTubes_470AA0(int a1) {
	dword_5d4594_1096252 = a1;
	if (*getMemU32Ptr(0x5D4594, 1093176)) {
		if (a1) {
			nox_window_set_hidden(nox_windows_arr_1093036[2].win, 0);
			nox_window_set_hidden(nox_windows_arr_1093036[3].win, 0);
		} else {
			nox_window_set_hidden(nox_windows_arr_1093036[2].win, 1);
			nox_window_set_hidden(nox_windows_arr_1093036[3].win, 1);
		}
	}
}

//----- (00470B00) --------------------------------------------------------
unsigned char* nox_xxx_guiHealthManaColorInit_470B00() {
	unsigned char* result; // eax

	dword_5d4594_1090284 = nox_color_rgb_4344A0(255, 0, 0);
	dword_5d4594_1090280 = nox_color_rgb_4344A0(100, 0, 0);
	*getMemU32Ptr(0x5D4594, 1091964) = nox_color_rgb_4344A0(0, 255, 0);
	*getMemU32Ptr(0x5D4594, 1092992) = nox_color_rgb_4344A0(0, 100, 0);

	nox_windows_arr_1093036[0].color_1 = dword_5d4594_1090284;
	nox_windows_arr_1093036[0].color_2 = dword_5d4594_1090280;

	nox_windows_arr_1093036[1].color_1 = nox_color_rgb_4344A0(0, 0, 255);
	nox_windows_arr_1093036[1].color_2 = nox_color_rgb_4344A0(0, 0, 100);

	nox_windows_arr_1093036[4].color_1 = nox_color_rgb_4344A0(240, 0, 240);
	nox_windows_arr_1093036[4].color_2 = nox_color_rgb_4344A0(50, 0, 50);

	nox_windows_arr_1093036[5].color_1 = nox_color_rgb_4344A0(255, 0, 255);
	nox_windows_arr_1093036[5].color_2 = nox_color_rgb_4344A0(50, 0, 50);

	nox_windows_arr_1093036[6].color_1 = nox_color_rgb_4344A0(255, 0, 255);
	nox_windows_arr_1093036[6].color_2 = nox_color_rgb_4344A0(50, 0, 50);

	result = getMemAt(0x5D4594, 1094732);
	do {
		*((uint32_t*)result - 384) = 0;
		*(uint32_t*)result = 0;
		result += 24;
	} while ((int)result < (int)getMemAt(0x5D4594, 1096268));
	return result;
}

//----- (00470C40) --------------------------------------------------------
int sub_470C40(int a1) {
	int result; // eax

	dword_5d4594_1096264 = a1;
	if (a1) {
		result = *getMemU32Ptr(0x5D4594, 1091964);
		nox_windows_arr_1093036[0].color_1 = *getMemU32Ptr(0x5D4594, 1091964);
		nox_windows_arr_1093036[0].color_2 = *getMemU32Ptr(0x5D4594, 1092992);
	} else {
		result = dword_5d4594_1090280;
		nox_windows_arr_1093036[0].color_1 = dword_5d4594_1090284;
		nox_windows_arr_1093036[0].color_2 = dword_5d4594_1090280;
	}
	return result;
}

//----- (00470C80) --------------------------------------------------------
int nox_xxx_cliSetTotalHealth_470C80(int a1, int a2) {
	int result; // eax

	if (dword_8531A0_2576) {
		*(uint32_t*)(dword_8531A0_2576 + 2247) = a2;
	}
	result = a1;
	nox_windows_arr_1093036[0].field_2 = a2;
	nox_windows_arr_1093036[0].field_1 = a1;
	dword_5d4594_1096260 = 32;
	return result;
}

//----- (00470CB0) --------------------------------------------------------
int sub_470CB0(int a1) {
	int result; // eax

	result = a1;
	nox_windows_arr_1093036[0].field_1 = a1;
	return result;
}

//----- (00470CC0) --------------------------------------------------------
int sub_470CC0() { return nox_windows_arr_1093036[0].field_1; }

//----- (00470CD0) --------------------------------------------------------
int sub_470CD0() { return nox_windows_arr_1093036[0].field_2; }

//----- (00470CE0) --------------------------------------------------------
int nox_xxx_cliSetManaAndMax_470CE0(int a1, int a2) {
	int result; // eax

	if (dword_8531A0_2576) {
		*(uint32_t*)(dword_8531A0_2576 + 2243) = a2;
	}
	result = a1;
	nox_windows_arr_1093036[1].field_2 = a2;
	nox_windows_arr_1093036[1].field_1 = a1;
	dword_5d4594_1096260 = 32;
	return result;
}

//----- (00470D10) --------------------------------------------------------
int nox_xxx_cliSetMana_470D10(int a1) {
	int result; // eax

	result = a1;
	nox_windows_arr_1093036[1].field_1 = a1;
	return result;
}

//----- (00470D20) --------------------------------------------------------
int sub_470D20(int a1, int a2) {
	int result; // eax

	result = a1;
	nox_windows_arr_1093036[4].field_1 = a1;
	nox_windows_arr_1093036[4].field_2 = a2;
	if (a1 != a2) {
		result = nox_xxx_setKeybTimeout_4160D0(17);
	}
	return result;
}

//----- (00470D70) --------------------------------------------------------
void sub_470D70() {
	nox_window_set_hidden(nox_windows_arr_1093036[5].win, 1);
	nox_window_set_hidden(nox_windows_arr_1093036[6].win, 1);
}

//----- (00470D90) --------------------------------------------------------
int sub_470D90(int a1, int a2) {
	int result; // eax

	nox_window_set_hidden(nox_windows_arr_1093036[5].win, 0);
	nox_window_set_hidden(nox_windows_arr_1093036[6].win, 0);
	result = a1;
	nox_windows_arr_1093036[5].field_1 = a1;
	nox_windows_arr_1093036[5].field_2 = a2;
	nox_windows_arr_1093036[6].field_1 = a1;
	nox_windows_arr_1093036[6].field_2 = a2;
	return result;
}

//----- (00470DD0) --------------------------------------------------------
int nox_xxx_cliGetMana_470DD0() { return nox_windows_arr_1093036[1].field_1; }

//----- (00470DE0) --------------------------------------------------------
int sub_470DE0() {
	int result; // eax
	int v1;     // ebp
	int v2;     // edx
	int v3;     // esi

	result = nox_player_netCode_85319C;
	if (nox_player_netCode_85319C) {
		v1 = nox_windows_arr_1093036[0].field_1;
		if (nox_windows_arr_1093036[0].field_1 >= 1) {
			result = -858993458 * nox_windows_arr_1093036[0].field_2;
			v2 = 2 * nox_windows_arr_1093036[0].field_2 / 5;
			v3 = v2;
			if (nox_windows_arr_1093036[0].field_1 < v2) {
				*getMemU32Ptr(0x5D4594, 1091960) =
					gameFPS() / 3u + nox_windows_arr_1093036[0].field_1 * ((unsigned int)(3 * gameFPS()) >> 2) / v2;
				result = nox_xxx_checkKeybTimeout_4160F0(4u, *getMemU32Ptr(0x5D4594, 1091960) - 1);
				if (result) {
					nox_xxx_clientPlaySoundSpecial_452D80(896, 66 * (v3 - v1) / v3 + 33);
					result = nox_xxx_setKeybTimeout_4160D0(4);
				}
			}
		}
	}
	return result;
}

//----- (00470E90) --------------------------------------------------------
int sub_470E90(nox_window* win, int event) {
	(void)win;
	switch (event) {
	case 5:
		nox_client_invAlterWeapon_4672C0();
		return 1;
	case 8:
	case 12:
	case 16:
		return 0;
	default:
		return 1;
	}
}

//----- (00470EE0) --------------------------------------------------------
void nox_win_init_cur_weapon(nox_window* a1, int a2, int a3, int w, int h) {
	nox_windows_arr_1093036[4].win = nox_window_new(a1, 0x408, a2, a3, w, h, 0);
	nox_window_set_all_funcs(nox_windows_arr_1093036[4].win, sub_470E90, sub_470F40_draw, sub_4710B0);
	nox_windows_arr_1093036[4].win->widget_data = 4;
}

//----- (00470F40) --------------------------------------------------------
int sub_470F40_draw(nox_window* win, nox_window_data* draw_data) {
	nox_window_yyy* v3; // edi
	int v4;            // ebx
	int v5;            // esi
	int v6;            // ecx
	nox_drawable* v7;  // eax
	double v8;         // st7
	double v9;         // st6
	int v12;           // [esp+10h] [ebp-1Ch]
	int v14;           // [esp+18h] [ebp-14h]
	int v16;           // [esp+20h] [ebp-Ch]
	int v17;           // [esp+24h] [ebp-8h]
	int v18;           // [esp+30h] [ebp+4h]

	v18 = 1;
	(void)draw_data;
	v3 = &nox_windows_arr_1093036[(intptr_t)win->widget_data];
	nox_client_wndGetPosition_46AA60(win, &v14, &v16);

	int w;
	int h;
	nox_window_get_size(win, &w, &h);
	v4 = w / 2;
	v17 = w / 2 + v14;
	v5 = h / 2 + v16;

	v6 = v3->field_2;
	if (v6) {
		v12 = (v3->field_1 << 8) / v6;
	} else {
		v18 = 0;
	}
	v3->color_2 = v3->color_1;
	if (!v18) {
		sub_465D50_draw(win);
		return 1;
	}
	if (v12 >= 256) {
		v7 = sub_4678D0();
		if (v7) {
			v8 = (double)v7->field_73_1;
			v9 = (double)v7->field_73_2;
			if (v8 < v9 * *getMemDoublePtr(0x581450, 9608)) {
				v3->color_2 = *getMemU32Ptr(0x85B3FC, 940);
				v12 = 1;
			} else if (v8 < v9 * *(double*)&qword_581450_9544) {
				v3->color_2 = nox_color_yellow_2589772;
				v12 = 1;
			} else {
				v18 = 0;
				v12 = 1;
			}
		} else {
			v18 = 0;
			v12 = 1;
		}
	}
	if (v18) {
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(0x40u);
		sub_4AE6F0(v17, v5, v4, v12, v3->color_2);
		nox_client_drawEnableAlpha_434560(0);
	}
	sub_465D50_draw(win);
	return 1;
}
// 470FE2: variable 'v12' is possibly undefined

//----- (00471250) --------------------------------------------------------
int sub_471250(nox_window* win, nox_window_data* draw_data) {
	nox_window_yyy* v1; // esi
	int v2;             // edi
	unsigned char* v3;  // esi
	int result;         // eax
	int v5;             // eax
	int v6;             // ecx
	int v7;             // ebx
	int v8;             // ebp
	int v9;             // ecx
	int v10;            // esi
	int v11;            // edi
	double v12;         // st7
	int v13;            // eax
	unsigned char* v14; // esi
	int v15;            // [esp+10h] [ebp-1Ch]
	int v16;            // [esp+14h] [ebp-18h]
	int v17;            // [esp+18h] [ebp-14h]
	int v18;            // [esp+1Ch] [ebp-10h]
	int v19;            // [esp+20h] [ebp-Ch]
	nox_window_yyy* v20; // [esp+24h] [ebp-8h]
	float v21;          // [esp+28h] [ebp-4h]
	float v22;          // [esp+30h] [ebp+4h]

	(void)draw_data;
	v20 = &nox_windows_arr_1093036[(intptr_t)win->widget_data];
	v1 = v20;
	nox_client_wndGetPosition_46AA60(win, &v18, &v17);
	v2 = 1;
	if (v1->field_2 >= 1) {
		v15 = 1;
		v5 = nox_xxx_bookGet_430B40_get_mouse_prev_seq();
		v6 = v1->field_2;
		v19 = v5;
		if (v6 > 30) {
			v15 = 0;
			v2 = 0;
		}
		v7 = (v2 + 61) / v6 - v2;
		v8 = 61 - v7;
		v22 = 0.001;
		v16 = 0;
		v21 = (double)(v2 - v6 * ((v2 + 61) / v6) + 61) / (double)v6;
		if (v6 > 0) {
			while (1) {
				if (v16 >= v1->field_1) {
					v9 = v1->color_2;
				} else {
					v9 = v1->color_1;
				}
				v10 = v8;
				v8 -= v7 + v2;
				v11 = v7;
				if (v7 <= 0) {
					v11 = 1;
				}
				v12 = v22 + v21;
				v22 = v12;
				if (v12 >= *(double*)&qword_581450_9512) {
					--v8;
					--v10;
					++v11;
					v22 = v22 - *(double*)&qword_581450_9512;
				}
				nox_client_drawSetColor_434460(v9);
				nox_client_drawEnableAlpha_434560(1);
				if (v10 < 0) {
					v13 = -v10;
					v10 = 0;
					v11 -= v13;
				}
				if (v11 > 0) {
					v14 = getMemAt(0x587000, 147905 + 8 * v10);
					do {
						if ((int)v14 >= (int)getMemAt(0x587000, 148393)) {
							break;
						}
						if (*(uint32_t*)(v14 + 3) != v19) {
							nox_client_drawRectFilledOpaque_49CE30(v18 + *(v14 - 1), v17 + *v14, v14[1], 1);
							*(uint32_t*)(v14 + 3) = v19;
						}
						v14 += 8;
						--v11;
					} while (v11 > 0);
				}
				nox_client_drawEnableAlpha_434560(0);
				if (++v16 >= v20->field_2) {
					break;
				}
				v2 = v15;
				v1 = v20;
			}
		}
		result = 1;
	} else {
		nox_client_drawSetColor_434460(v1->color_1);
		nox_client_drawEnableAlpha_434560(1);
		v3 = getMemAt(0x587000, 147905);
		do {
			nox_client_drawRectFilledOpaque_49CE30(v18 + *(v3 - 1), v17 + *v3, v3[1], 1);
			v3 += 8;
		} while ((int)v3 < (int)getMemAt(0x587000, 148393));
		nox_client_drawEnableAlpha_434560(0);
		result = 1;
	}
	return result;
}

//----- (00471450) --------------------------------------------------------
int sub_471450(nox_window* win, nox_window_data* draw_data) {
	int v3;                 // [esp+4h] [ebp-10h]
	int v4;                 // [esp+8h] [ebp-Ch]
	int y;
	wchar2_t WideCharStr[4]; // [esp+Ch] [ebp-8h]

	(void)draw_data;
	nox_itow(nox_windows_arr_1093036[(intptr_t)win->widget_data].field_1, WideCharStr, 10);
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	nox_xxx_drawGetStringSize_43F840(dword_5d4594_1096288, WideCharStr, &v3, 0, 0);
	nox_client_wndGetPosition_46AA60(win, &v4, &y);
	nox_xxx_drawString_43F6E0(dword_5d4594_1096288, WideCharStr, v4 - v3 / 2 + 8, y + 1);
	return 1;
}

//----- (00471A80) --------------------------------------------------------
int nox_xxx_guiBottleSlotDrawFn_471A80(nox_window* win, nox_window_data* draw_data) {
	int v1;        // esi
	int v2;        // esi
	int v4;        // eax
	short* v5;     // esi
	int v6;        // eax
	int v8;        // [esp+4h] [ebp-14h]
	int y;
	wchar2_t v9[8]; // [esp+8h] [ebp-10h]
	nox_drawable* drawable;

	(void)draw_data;
	v1 = (intptr_t)win->widget_data;
	nox_client_wndGetPosition_46AA60(win, &v8, &y);
	v2 = 536 * v1;
	if (*getMemU16Ptr(0x5D4594, 1090312 + v2)) {
		drawable = &nox_gui_bottle_drawables[v1];
		if (drawable->draw_func) {
			drawable->pos.x = v8 + 14;
			drawable->pos.y = y + 15;
			drawable->draw_func(getMemAt(0x5D4594, 1091908), drawable);
		}
		nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
		nox_swprintf(v9, L"%d", *getMemU16Ptr(0x5D4594, 1090312 + v2));
		v4 = nox_xxx_guiFontHeightMB_43F320(dword_5d4594_1096288);
		nox_xxx_drawString_43F6E0(dword_5d4594_1096288, v9, v8 - 2, y - v4 + 10);
	}
	v5 = getMemI16Ptr(0x5D4594, 1090300 + v2);
	if (v5) {
		v6 = nox_xxx_guiFontHeightMB_43F320(dword_5d4594_1096288);
		nox_xxx_drawString_43F6E0(dword_5d4594_1096288, (wchar2_t*)v5, v8 - 2, y - v6 + 33);
	}
	return 1;
}

//----- (00471B90) --------------------------------------------------------
int nox_xxx_guiBottleSlotProc_471B90(nox_window* win, int event) {
	int slot = (intptr_t)win->widget_data;
	switch (event) {
	case 5:
		if (*getMemU32Ptr(0x5D4594, 1090308 + 536 * slot)) {
			nox_xxx_cliUseCurePoison_4674E0(*getMemU32Ptr(0x5D4594, 1090308 + 536 * slot));
		}
		return 1;
	case 8:
	case 12:
	case 16:
		return 0;
	default:
		return 1;
	}
}

//----- (00471C00) --------------------------------------------------------
extern uint32_t nox_gameDisableMapDraw_5d4594_2650672;
int nox_xxx_drawHealthManaBar_471C00(nox_window* win, nox_window_data* draw_data) {
	int v1;            // esi
	nox_window_yyy* v2; // ebp
	int v3;            // edi
	int v4;            // esi
	int v5;            // ebx
	int v7;            // [esp+Ch] [ebp+4h]

	(void)draw_data;
	v1 = (intptr_t)win->widget_data;
	v7 = v1;
	v2 = &nox_windows_arr_1093036[v1];
	if (nox_xxx_clientIsObserver_4372E0() || nox_gameDisableMapDraw_5d4594_2650672 ||
		nox_common_gameFlags_check_40A5C0(9437184)) {
		return 1;
	}
	if (v1) {
		v3 = nox_win_width / 2 + 21;
	} else {
		v3 = nox_win_width / 2 + 15;
	}
	v4 = nox_win_height / 2 - 48;
	v5 = v2->field_2 ? 48 * v2->field_1 / v2->field_2 : 0;
	nox_client_drawSetColor_434460(nox_color_black_2650656);
	nox_client_drawRectFilledOpaque_49CE30(v3, v4, 2, 48);
	nox_client_drawSetColor_434460(v2->color_1);
	nox_client_drawRectFilledOpaque_49CE30(v3, v4 - v5 + 48, 2, v5);
	if (v7) {
		nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 944));
	} else if (dword_5d4594_1096264) {
		nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 984));
	} else {
		nox_client_drawSetColor_434460(nox_color_violet_2598268);
	}
	nox_client_drawBorderLines_49CC70(v3 - 1, v4 - 1, 4, 50);
	return 1;
}

//----- (00472080) --------------------------------------------------------
int sub_472080() {
	int result; // eax

	result = nox_windows_arr_1093036[4].field_1;
	if (nox_windows_arr_1093036[4].field_1 != nox_windows_arr_1093036[4].field_2) {
		result = sub_416120(0x11u);
		if (result) {
			result = 0x64u / (int)gameFPS();
			nox_windows_arr_1093036[4].field_1 += 0x64u / (int)gameFPS();
		}
	}
	return result;
}

//----- (004720C0) --------------------------------------------------------
int sub_4720C0(int xLeft, int a2) {
	nox_client_drawPixel_49EFA0(xLeft + 1, a2);
	nox_client_drawRectFilledOpaque_49CE30(xLeft, a2 + 1, 3, 1);
	nox_client_drawPixel_49EFA0(xLeft + 1, a2 + 2);
	return 0;
}

//----- (00472100) --------------------------------------------------------
int nox_xxx_guiHealthManaTubeProc_472100(nox_window* win, int event) {
	int v3;     // [esp-4h] [ebp-4h]
	(void)win;

	switch (event) {
	case 7:
		v3 = dword_5d4594_1096252 == 1;
		dword_5d4594_1096252 = 1 - dword_5d4594_1096252;
		nox_window_set_hidden(nox_windows_arr_1093036[2].win, v3);
		if (getMemByte(0x85B3FC, 12254) != 0) {
			nox_window_set_hidden(nox_windows_arr_1093036[3].win, dword_5d4594_1096252 == 0);
		}
		nox_xxx_clientPlaySoundSpecial_452D80(901, 100);
		return 1;
	case 8:
	case 12:
	case 16:
		return 0;
	default:
		return 1;
	}
}

//----- (004721A0) --------------------------------------------------------
int sub_4721A0(int a1) {
	if (a1) {
		return nox_window_set_hidden(dword_5d4594_1090276, 0);
	} else {
		return nox_window_set_hidden(dword_5d4594_1090276, 1);
	}
}

//----- (004721D0) --------------------------------------------------------
int nox_xxx_cliPrepareGameplay2_4721D0() {
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1090276);
	nox_xxx_windowDestroyMB_46C4E0(nox_windows_arr_1093036[2].win);
	if (nox_windows_arr_1093036[3].win) {
		nox_xxx_windowDestroyMB_46C4E0(nox_windows_arr_1093036[3].win);
	}
	nox_xxx_guiHealthManaInit_4714E0();
	sub_472310();
	return sub_4721A0(nox_client_getRenderGUI());
}

//----- (00472220) --------------------------------------------------------
void nox_client_quickHealthPotion_472220() {
	if (!nox_xxx_guiCursor_477600()) {
		if (*getMemU32Ptr(0x5D4594, 1090308)) {
			nox_xxx_cliUseCurePoison_4674E0(*getMemIntPtr(0x5D4594, 1090308));
		}
	}
}

//----- (00472240) --------------------------------------------------------
void nox_client_quickManaPotion_472240() {
	if (!nox_xxx_guiCursor_477600()) {
		if (*getMemU32Ptr(0x5D4594, 1090844)) {
			nox_xxx_cliUseCurePoison_4674E0(*getMemIntPtr(0x5D4594, 1090844));
		}
	}
}

//----- (00472260) --------------------------------------------------------
void nox_client_quickCurePoisonPotion_472260() {
	if (!nox_xxx_guiCursor_477600()) {
		if (*getMemU32Ptr(0x5D4594, 1091380)) {
			nox_xxx_cliUseCurePoison_4674E0(*getMemIntPtr(0x5D4594, 1091380));
		}
	}
}

//----- (00472280) --------------------------------------------------------
wchar2_t* sub_472280() {
	wchar2_t* result; // eax
	char* v1;        // eax
	char* v2;        // eax
	char* v3;        // eax

	result = *(wchar2_t**)(&dword_8531A0_2576);
	if (dword_8531A0_2576) {
		v1 = sub_42E8E0(38, 1);
		nox_wcsncpy((wchar2_t*)getMemAt(0x5D4594, 1091372), (const wchar2_t*)v1, 3u);
		*getMemU16Ptr(0x5D4594, 1091378) = 0;
		v2 = sub_42E8E0(36, 1);
		nox_wcsncpy((wchar2_t*)getMemAt(0x5D4594, 1090300), (const wchar2_t*)v2, 3u);
		result = *(wchar2_t**)(&dword_8531A0_2576);
		*getMemU16Ptr(0x5D4594, 1090306) = 0;
		if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
			v3 = sub_42E8E0(37, 1);
			result = nox_wcsncpy((wchar2_t*)getMemAt(0x5D4594, 1090836), (const wchar2_t*)v3, 3u);
			*getMemU16Ptr(0x5D4594, 1090842) = 0;
		}
	}
	return result;
}

//----- (00472310) --------------------------------------------------------
static void nox_gui_link_bottle_drawable(int slot, int thing_id) {
	nox_drawable* drawable = &nox_gui_bottle_drawables[slot];
	memset(drawable, 0, sizeof(*drawable));
	if (!thing_id) {
		return;
	}
	nox_thing* thing = nox_get_thing(thing_id);
	if (!thing) {
		return;
	}
	nox_drawable_link_thing(drawable, thing->field_1c);
	drawable->flags30 |= 0x40000000u;
}

unsigned char* sub_472310() {
	int cure_id = dword_5d4594_1096276;
	int mana_id = dword_5d4594_1096272;
	int health_id = *getMemU32Ptr(0x5D4594, 1096268);

	*getMemU16Ptr(0x5D4594, 1091384) = sub_467850(cure_id);
	*getMemU16Ptr(0x5D4594, 1090848) = sub_467850(mana_id);
	*getMemU16Ptr(0x5D4594, 1090312) = sub_467850(health_id);
	if (!*getMemU16Ptr(0x5D4594, 1090312)) {
		health_id = dword_5d4594_1096284;
		*getMemU16Ptr(0x5D4594, 1090312) = sub_467850(health_id);
	}
	if (!*getMemU16Ptr(0x5D4594, 1090312)) {
		health_id = dword_5d4594_1096280;
		*getMemU16Ptr(0x5D4594, 1090312) = sub_467850(health_id);
	}
	if (!*getMemU16Ptr(0x5D4594, 1090312)) {
		health_id = 0;
	}

	*getMemU32Ptr(0x5D4594, 1090308) = health_id;
	*getMemU32Ptr(0x5D4594, 1090844) = mana_id;
	*getMemU32Ptr(0x5D4594, 1091380) = cure_id;
	nox_gui_link_bottle_drawable(0, health_id);
	nox_gui_link_bottle_drawable(1, mana_id);
	nox_gui_link_bottle_drawable(2, cure_id);
	return health_id ? (unsigned char*)&nox_gui_bottle_drawables[0] : NULL;
}

//----- (004724E0) --------------------------------------------------------
void nox_client_mapZoomIn_4724E0() {
	nox_xxx_minimap_587000_149232 -= 10;
	if (*(int*)&nox_xxx_minimap_587000_149232 < 500) {
		nox_xxx_minimap_587000_149232 = 500;
	}
}

//----- (00472500) --------------------------------------------------------
void nox_client_mapZoomOut_472500() {
	nox_xxx_minimap_587000_149232 += 10;
	if (nox_xxx_minimap_587000_149232 > 4000) {
		nox_xxx_minimap_587000_149232 = 4000;
	}
}

//----- (00472520) --------------------------------------------------------
int nox_xxx_cliSetMinimapZoom_472520(int a1) {
	int result; // eax

	result = a1;
	nox_xxx_minimap_587000_149232 = a1;
	return result;
}

//----- (00472540) --------------------------------------------------------
int sub_472540(nox_drawable* dr) {
	int v1;     // edx
	int v2;     // eax
	int result; // eax
	int2 a1a;   // [esp+0h] [ebp-8h]

	if (dr == getMemPtr(0x852978, 8)) {
		nox_xxx_getSomeCoods_435670(&a1a);
	} else {
		v1 = (int)dr->pos.y;
		a1a.field_0 = (int)dr->pos.x;
		a1a.field_4 = v1;
	}
	v2 = nox_xxx_polygonGetIdxA_421790(&a1a, *getMemIntPtr(0x5D4594, 1096312));
	if (v2) {
		*getMemU32Ptr(0x5D4594, 1096312) = v2;
	} else {
		v2 = *getMemU32Ptr(0x5D4594, 1096312);
	}
	if (v2) {
		result = (unsigned char)nox_xxx_polygonGetByIdx_4214A0(v2)[130];
	} else {
		result = 1;
	}
	return result;
}

//----- (004725C0) --------------------------------------------------------
void nox_xxx_drawMinimap4Sprite_4725C0(nox_drawable* dr) {
	nox_drawable* local = getMemPtr(0x852978, 8);
	if (local && !nox_client_drawable_testBuff_4356C0(local, 2)) {
		sub_437260();
		*getMemU32Ptr(0x5D4594, 1096316) = sub_472540(dr);
		nox_xxx_cliDrawMinimap_472600(dr, *getMemIntPtr(0x5D4594, 1096316));
		sub_437290();
	}
}

//----- (00472600) --------------------------------------------------------
void* sub_4106A0(int a1);
void* nox_server_wallNextByY_4106B0(void* wall);
int sub_50CB00();
void* sub_50CB10();
int nox_xxx_cliDrawMinimap_472600(nox_drawable* a1, int a2) {
	char* v2;                           // ebp
	int v3;                             // esi
	int v4;                             // kr08_4
	int v5;                             // ebx
	int v6;                             // et1
	int v7;                             // esi
	int v8;                             // ebx
	int v9;                             // ebp
	int v10;                            // eax
	uint8_t* v11;                       // edi
	char v12;                           // al
	int v13;                            // eax
	int v14;                            // esi
	int v15;                            // ebp
	unsigned char* v16;                 // esi
	uint8_t* v17;                       // eax
	uint8_t* v18;                       // eax
	uint8_t* v19;                       // eax
	int v20;                            // et1
	int v21;                            // ecx
	int v22;                            // ebx
	int v23;                            // ebp
	int v24;                            // esi
	int v25;                            // ecx
	int v26;                            // et1
	char v27;                           // al
	char v28;                           // dl
	int v29;                            // edi
	char* v30;                          // esi
	float* v31;                         // esi
	int v32;                            // et1
	double v33;                         // st7
	int v34;                            // et1
	double v35;                         // st7
	float* j;                           // esi
	int v37;                            // et1
	double v38;                         // st7
	nox_object_team_t* v39;             // eax
	nox_drawable* k;                    // esi
	nox_player_polygon_check_data* v41; // eax
	int v42;                            // et1
	int v43;                            // eax
	int v44;                            // eax
	nox_drawable* v45;                  // edi
	nox_object_team_t* v46;             // eax
	nox_team_t* v47;                    // eax
	int v48;                           // eax
	nox_object_team_t* v49;             // eax
	nox_object_team_t* v50;             // edi
	nox_team_t* v51;                    // eax
	int v52;                           // eax
	int v53;                            // eax
	int v54;                            // eax
	nox_team_t* v55;                    // eax
	nox_playerInfo* v56;                // eax
	nox_object_team_t* v57;             // eax
	int v58;                           // eax
	nox_drawable* l;                    // esi
	int v60;                            // eax
	int v61;                            // edx
	nox_object_team_t* v62;             // edi
	nox_player_polygon_check_data* v63; // eax
	int v64;                            // et1
	nox_team_t* v65;                    // eax
	int v66;                           // eax
	int v68;                            // [esp-10h] [ebp-70h]
	int v69;                            // [esp+10h] [ebp-50h]
	nox_object_team_t* v70;             // [esp+10h] [ebp-50h]
	int i;                              // [esp+14h] [ebp-4Ch]
	int v72;                            // [esp+14h] [ebp-4Ch]
	int v73;                            // [esp+14h] [ebp-4Ch]
	int v74;                            // [esp+18h] [ebp-48h]
	nox_playerInfo* v75;                // [esp+18h] [ebp-48h]
	int2 v76;                           // [esp+20h] [ebp-40h]
	int v77;                            // [esp+28h] [ebp-38h]
	int v78;                            // [esp+2Ch] [ebp-34h]
	int v79;                            // [esp+30h] [ebp-30h]
	int v80;                            // [esp+34h] [ebp-2Ch]
	int v81;                            // [esp+38h] [ebp-28h]
	int2 xLeft;                         // [esp+40h] [ebp-20h]
	int yTop;                           // [esp+4Ch] [ebp-14h]
	int2 v84;                           // [esp+50h] [ebp-10h]
	int v85;                            // [esp+5Ch] [ebp-4h]
	int2 minimap_pos;

	v2 = nox_draw_getViewport_437250();
	if (!getMemByte(0x5D4594, 1096300)) {
		*getMemU8Ptr(0x5D4594, 1096300) = nox_xxx_wallTileByName_410D60("InvisibleWallSet");
		*getMemU8Ptr(0x5D4594, 1096301) = nox_xxx_wallTileByName_410D60("InvisibleBlockingWallSet");
	}
	nox_client_drawEnableAlpha_434560(0);
	nox_xxx_wndDraw_49F7F0();
	v3 = nox_win_width / 6;
	v4 = nox_win_height - nox_win_width / 6;
	yTop = v4 / 2;
	nox_client_copyRect_49F6F0(0, 0, nox_win_width, nox_win_height);
	v5 = *(uint32_t*)v2;
	if (*(uint32_t*)v2 <= 0) {
		nox_client_drawRectFilledAlpha_49CF10(0, v4 / 2, v3, v3);
	} else {
		nox_client_drawSetColor_434460(nox_color_black_2650656);
		if (v5 >= v3) {
			nox_client_drawRectFilledOpaque_49CE30(0, v4 / 2, v3, v3);
		} else {
			nox_client_drawRectFilledOpaque_49CE30(0, v4 / 2, v5, v3);
			nox_client_drawRectFilledAlpha_49CF10(v5, v4 / 2, v3 - v5, v3);
		}
	}
	nox_client_drawEnableAlpha_434560(1);
	nox_client_drawSetColor_434460(nox_color_black_2650656);
	nox_client_drawSetAlpha_434580(0x5Au);
	nox_client_drawRectLines_473510(-1, yTop - 1, v3 + 2, v3 + 2);
	nox_client_drawSetAlpha_434580(0x3Cu);
	nox_client_drawRectLines_473510(-2, yTop - 2, v3 + 4, v3 + 4);
	nox_client_drawSetAlpha_434580(0x28u);
	nox_client_drawRectLines_473510(-3, yTop - 3, v3 + 6, v3 + 6);
	nox_client_drawEnableAlpha_434560(0);
	nox_client_copyRect_49F6F0(0, yTop, v3, v3);
	v6 = nox_xxx_minimap_587000_149232;
	v7 = v3 * v6 / 100;
	nox_xxx_getSomeCoods_435670(&v84);
	v8 = v84.field_0 - v7 / 2;
	v9 = v84.field_4 - v7 / 2;
	xLeft.field_0 = v84.field_0 - v7 / 2;
	xLeft.field_4 = v9;
	v81 = v8 / 23;
	v77 = (v8 + v7) / 23;
	v78 = (v9 + v7) / 23;
	v74 = 23 * (v8 / 23);
	v80 = 23 * (v9 / 23);
	v10 = v9 / 23;
	for (i = v9 / 23; i <= v78; ++i) {
		v11 = sub_4106A0(v10);
		while (v11) {
			v12 = *(uint8_t*)(v11 + 1);
			if (v12 == getMemByte(0x5D4594, 1096300)) {
				goto LABEL_37;
			}
			if (v12 == getMemByte(0x5D4594, 1096301)) {
				goto LABEL_37;
			}
			if (*(uint8_t*)(v11 + 8) && *(unsigned char*)(v11 + 8) != a2) {
				goto LABEL_37;
			}
			v13 = *(unsigned char*)(v11 + 5);
			if (v13 < v81) {
				goto LABEL_37;
			}
			if (v13 > v77) {
				break;
			}
			v14 = v74 + 23 * (v13 - v81);
			if (*(uint8_t*)(v11 + 4) & 0x10) {
				v15 = *(uint32_t*)(v11 + 32);
				if (!v15) {
					goto LABEL_37;
				}
				v69 = 0;
				v16 = getMemAt(0x587000, 149244);
				while (1) {
							v17 = nox_server_getWallAtGrid_410580(*((uint32_t*)v16 - 1) + *(unsigned char*)(v11 + 5),
														  *(uint32_t*)v16 + *(unsigned char*)(v11 + 6));
							if (v17) {
								if (*(uint32_t*)(v17 + 12)) {
							if (v69 < 4) {
								v20 = nox_xxx_minimap_587000_149232;
								v21 = 8 * *(unsigned char*)(v15 + 299);
								v22 = 100 * (*(int*)(v15 + 12) - xLeft.field_0) / v20;
								v85 = 100 * (*(int*)(v15 + 16) - xLeft.field_4) / v20;
								v23 = 100 * *getMemIntPtr(0x587000, 196184 + v21) / v20;
								v24 = 100 * *getMemIntPtr(0x587000, 196188 + v21) / v20;
								nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 940));
								nox_client_drawAddPoint_49F500(v22, yTop + v85);
								nox_xxx_rasterPointRel_49F570(v23, v24);
								nox_client_drawLineFromPoints_49E4B0();
								v8 = xLeft.field_0;
							}
							goto LABEL_37;
						}
					} else {
						v18 = nox_server_getWallAtGrid_410580(*((uint32_t*)v16 - 1) + *(unsigned char*)(v11 + 5),
															  *(unsigned char*)(v11 + 6));
						if (v18 && *(uint32_t*)(v18 + 12) ||
							(v19 = nox_server_getWallAtGrid_410580(
								 *(unsigned char*)(v11 + 5), *(uint32_t*)v16 + *(unsigned char*)(v11 + 6))) != 0 &&
								*(uint32_t*)(v19 + 12)) {
							if (v69 < 4) {
								v20 = nox_xxx_minimap_587000_149232;
								v21 = 8 * *(unsigned char*)(v15 + 299);
								v22 = 100 * (*(int*)(v15 + 12) - xLeft.field_0) / v20;
								v85 = 100 * (*(int*)(v15 + 16) - xLeft.field_4) / v20;
								v23 = 100 * *getMemIntPtr(0x587000, 196184 + v21) / v20;
								v24 = 100 * *getMemIntPtr(0x587000, 196188 + v21) / v20;
								nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 940));
								nox_client_drawAddPoint_49F500(v22, yTop + v85);
								nox_xxx_rasterPointRel_49F570(v23, v24);
								nox_client_drawLineFromPoints_49E4B0();
								v8 = xLeft.field_0;
							}
							goto LABEL_37;
						}
					}
					v16 += 8;
					++v69;
					if ((int)v16 >= (int)getMemAt(0x587000, 149276)) {
						goto LABEL_37;
					}
				}
			}
			if (nox_common_gameFlags_check_40A5C0(0x10000) || *(uint32_t*)(v11 + 12)) {
				v26 = nox_xxx_minimap_587000_149232;
				v25 = v26;
				v76.field_0 = 100 * (v14 - v8) / v26;
				v76.field_4 = yTop + 100 * (v80 - v9) / v26;
				v27 = *(uint8_t*)(v11 + 4);
				if (!(v27 & 4) || (v28 = *(uint8_t*)(*(uint32_t*)(v11 + 28) + 21), v28 != 3) && v28 != 2) {
					if (!(v27 & 0x20)) {
						sub_4730D0(&v76, *(uint8_t*)v11, 2300 / v25);
					}
				}
			}
		LABEL_37:
			v11 = nox_server_wallNextByY_4106B0(v11);
			v9 = xLeft.field_4;
		}
		v10 = i + 1;
		v80 += 23;
	}
	if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_SHOW_AI)) {
		v29 = sub_50CB00();
		v30 = (char*)sub_50CB10();
		if (v29 >= 2) {
			nox_client_drawSetColor_434460(dword_8531A0_2572);
			if (v29 - 1 > 0) {
				v31 = (float*)(v30 + 8);
				v72 = v29 - 1;
				do {
					v32 = nox_xxx_minimap_587000_149232;
					v33 = *(v31 - 1);
					xLeft.field_0 = (int)(100 * ((unsigned long long)(long long)*(v31 - 2) - v8)) / v32;
					nox_client_drawAddPoint_49F500(xLeft.field_0,
												   yTop + (int)(100 * ((unsigned long long)(long long)v33 - v9)) / v32);
					v34 = nox_xxx_minimap_587000_149232;
					v35 = v31[1];
					xLeft.field_0 = (int)(100 * ((unsigned long long)(long long)*v31 - v8)) / v34;
					nox_client_drawAddPoint_49F500(xLeft.field_0,
												   yTop + (int)(100 * ((unsigned long long)(long long)v35 - v9)) / v34);
					nox_client_drawLineFromPoints_49E4B0();
					v31 += 2;
					--v72;
				} while (v72);
			}
		}
		for (j = (float*)nox_xxx_minimapFirstMonster_50AAE0(); j; j = (float*)nox_xxx_minimapNextMonster_50AB10()) {
			nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 940));
			v37 = nox_xxx_minimap_587000_149232;
			v38 = j[1];
			xLeft.field_0 = (int)(100 * ((unsigned long long)(long long)*j - v8)) / v37;
			nox_xxx_minimapDrawPoint_473570(xLeft.field_0,
											yTop + (int)(100 * ((unsigned long long)(long long)v38 - v9)) / v37);
		}
	}
	v73 = 0;
	if (!*getMemU32Ptr(0x5D4594, 1096304)) {
		*getMemU32Ptr(0x5D4594, 1096304) = nox_xxx_getTTByNameSpriteMB_44CFC0("Crown");
		*getMemU32Ptr(0x5D4594, 1096308) = nox_xxx_getTTByNameSpriteMB_44CFC0("GameBall");
	}
	v39 = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
	v70 = v39;
	if (v39 && nox_xxx_servObjectHasTeam_419130(v39)) {
		v73 = 1;
	}
	for (k = nox_xxx_cliFirstMinimapObj_459EB0(); k; k = nox_xxx_cliNextMinimapObj_459EC0(k)) {
		minimap_pos.field_0 = (int)k->pos.x;
		minimap_pos.field_4 = (int)k->pos.y;
		v41 = nox_xxx_polygonIsPlayerInPolygon_4217B0(&minimap_pos, 0);
		if (v41) {
			if (BYTE2(v41->field_0[32]) != a2) {
				continue;
			}
		} else if (a2 != 1) {
			continue;
		}
		v42 = nox_xxx_minimap_587000_149232;
		xLeft.field_0 = 100 * (minimap_pos.field_0 - v8) / v42;
		xLeft.field_4 = yTop + 100 * (minimap_pos.field_4 - v9) / v42;
		if (!(k->flags28 & 0x400000) || (v43 = nox_color_blue_2650684, !(k->flags29 & 8))) {
			v43 = *getMemU32Ptr(0x85B3FC, 940);
		}
		nox_client_drawSetColor_434460(v43);
		v44 = k->field_27;
		if (v44 == *getMemIntPtr(0x5D4594, 1096304)) {
			if (nox_server_teamFirst_418B10() || (v45 = nox_xxx_cliGetSpritePlayer_45A000()) == 0) {
				// nop
			} else {
				while (!nox_client_drawable_testBuff_4356C0(v45, 30)) {
					v45 = sub_45A010(v45);
					if (!v45) {
						goto LABEL_64;
					}
				}
				continue;
			}
		LABEL_64:
			nox_client_drawSetColor_434460(dword_8531A0_2572);
			v46 = nox_xxx_objGetTeamByNetCode_418C80(k->field_32);
			if (v46) {
				v47 = nox_xxx_getTeamByID_418AB0(v46->id);
				if (v47) {
					v48 = nox_xxx_materialGetTeamColor_418D50(v47);
					nox_client_drawSetColor_434460(v48);
				}
			}
			sub_473420(&xLeft);
			continue;
		}
		if (v44 == *getMemIntPtr(0x5D4594, 1096308)) {
			v49 = nox_xxx_objGetTeamByNetCode_418C80(k->field_32);
			v50 = v49;
			if (v49 && nox_xxx_servObjectHasTeam_419130(v49)) {
				v51 = nox_xxx_getTeamByID_418AB0(v50->id);
				if (v51) {
					v52 = nox_xxx_materialGetTeamColor_418D50(v51);
					nox_client_drawSetColor_434460(v52);
				}
			} else {
				nox_client_drawSetColor_434460(nox_color_white_2523948);
			}
			nox_video_drawCircleRad3_4734F0(&xLeft.field_0);
			continue;
		}
		v53 = k->flags28;
		if (v53 & 0x10000000) {
			if (k->flags30 & 0x1000000) {
				nox_client_drawSetColor_434460(nox_color_white_2523948);
				v54 = sub_4B94E0(k);
				v55 = nox_xxx_getTeamByID_418AB0(v54);
				if (v55) {
					v58 = nox_xxx_materialGetTeamColor_418D50(v55);
					nox_client_drawSetColor_434460(v58);
					sub_4733B0(&xLeft);
					continue;
				}
				sub_4733B0(&xLeft);
				continue;
			}
		} else {
			if (!(v53 & 4)) {
				nox_xxx_minimapDrawPoint_473570(xLeft.field_0, xLeft.field_4);
				continue;
			}
			if (!nox_common_gameFlags_check_40A5C0(32)) {
				if (k == getMemPtr(0x852978, 8)) {
					sub_4735C0(xLeft.field_0, xLeft.field_4);
				} else {
					nox_xxx_minimapDrawPoint_473570(xLeft.field_0, xLeft.field_4);
				}
				continue;
			}
			v56 = nox_common_playerInfoGetByID_417040(k->field_32);
			if (v56) {
				v81 = v56->field_4 & 1;
				if (v81) {
					nox_client_drawSetColor_434460(nox_color_white_2523948);
					v57 = nox_xxx_objGetTeamByNetCode_418C80(k->field_32);
					if (v57) {
						v55 = v57->id == 1 ? nox_xxx_getTeamByID_418AB0(2) : nox_xxx_getTeamByID_418AB0(1);
						if (v55) {
							v58 = nox_xxx_materialGetTeamColor_418D50(v55);
							nox_client_drawSetColor_434460(v58);
							sub_4733B0(&xLeft);
							continue;
						}
					}
					sub_4733B0(&xLeft);
					continue;
				}
			}
		}
	}
	v79 = dword_8531A0_2572;
	for (l = nox_xxx_cliGetSpritePlayer_45A000(); l; l = sub_45A010(l)) {
		v60 = nox_client_drawable_testBuff_4356C0(l, 30);
		v61 = l->field_32;
		v77 = v60;
		v62 = nox_xxx_objGetTeamByNetCode_418C80(v61);
		v68 = l->field_32;
		v75 = nox_common_playerInfoGetByID_417040(v68);
		if (v70 && v62 && v73) {
			v76.field_0 = nox_xxx_servCompareTeams_419150(v70, v62);
			if (v76.field_0) {
				goto LABEL_103;
			}
		} else {
			v76.field_0 = 0;
		}
		if (!(v77 || l == getMemPtr(0x852978, 8) ||
			  dword_8531A0_2576 && (((nox_playerInfo*)dword_8531A0_2576)->field_3680 & 1))) {
			continue;
		}
	LABEL_103:
		minimap_pos.field_0 = (int)l->pos.x;
		minimap_pos.field_4 = (int)l->pos.y;
		v63 = nox_xxx_polygonIsPlayerInPolygon_4217B0(&minimap_pos, 0);
		if ((!v63 || BYTE2(v63->field_0[32]) == a2) && v75 && (v75->field_3680 & 1) != 1) {
			v64 = nox_xxx_minimap_587000_149232;
			xLeft.field_0 = 100 * (minimap_pos.field_0 - v8) / v64;
			xLeft.field_4 = yTop + 100 * (minimap_pos.field_4 - v9) / v64;
			if (l == getMemPtr(0x852978, 8) || v76.field_0) {
				nox_client_drawSetColor_434460(v79);
			} else {
				nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 940));
			}
			if (v77) {
				if (v62) {
					v65 = nox_xxx_getTeamByID_418AB0(v62->id);
					if (v65) {
						v66 = nox_xxx_materialGetTeamColor_418D50(v65);
						nox_client_drawSetColor_434460(v66);
					}
				}
				sub_473420(&xLeft);
			} else {
				nox_xxx_minimapDrawPoint_473570(xLeft.field_0, xLeft.field_4);
			}
		}
	}
	return sub_49F860();
}

//----- (004730D0) --------------------------------------------------------
int sub_4730D0(int2* a1, unsigned char a2, int a3) {
	int result = 0; // eax
	int v4;     // ebx
	int v5;     // edi
	int2* v6;   // ebp

	if (nox_xxx_minimap_587000_149232 <= 2000) {
		v4 = *getMemU32Ptr(0x85B3FC, 956);
		result = a2;
		v5 = a3 / 2;
		switch (a2) {
		case 0u:
			result =
				sub_473380(a1->field_0, a3 + a1->field_4, a1->field_0 + a3, a1->field_4, *getMemIntPtr(0x85B3FC, 956));
			break;
		case 1u:
			result =
				sub_473380(a1->field_0, a1->field_4, a1->field_0 + a3, a1->field_4 + a3, *getMemIntPtr(0x85B3FC, 956));
			break;
		case 2u:
			sub_473380(a1->field_0, a3 + a1->field_4, a1->field_0 + a3, a1->field_4, *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(a1->field_0, a1->field_4, a1->field_0 + a3, a1->field_4 + a3, v4);
			break;
		case 3u:
			sub_473380(a1->field_0, a1->field_4, a1->field_0 + v5, a1->field_4 + v5, *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(a1->field_0, a3 + a1->field_4, a3 + a1->field_0, a1->field_4, v4);
			break;
		case 4u:
			sub_473380(a1->field_0, a1->field_4, a1->field_0 + a3, a1->field_4 + a3, *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(v5 + a1->field_0, v5 + a1->field_4, a3 + a1->field_0, a1->field_4, v4);
			break;
		case 5u:
			sub_473380(a1->field_0, a3 + a1->field_4, a1->field_0 + a3, a1->field_4, *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(v5 + a1->field_0, v5 + a1->field_4, a3 + a1->field_0, a1->field_4 + a3, v4);
			break;
		case 6u:
			v6 = a1;
			sub_473380(a1->field_0, a1->field_4, a1->field_0 + a3, a1->field_4 + a3, *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(v6->field_0, a3 + v6->field_4, v5 + v6->field_0, v6->field_4 + v5, v4);
			break;
		case 7u:
			sub_473380(a1->field_0, a1->field_4, a1->field_0 + v5, a1->field_4 + v5, *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(v5 + a1->field_0, v5 + a1->field_4, a3 + a1->field_0, a1->field_4, v4);
			break;
		case 8u:
			sub_473380(v5 + a1->field_0, v5 + a1->field_4, a1->field_0 + a3, a1->field_4 + a3,
					   *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(v5 + a1->field_0, v5 + a1->field_4, a3 + a1->field_0, a1->field_4, v4);
			break;
		case 9u:
			sub_473380(v5 + a1->field_0, v5 + a1->field_4, a1->field_0 + a3, a1->field_4 + a3,
					   *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(v5 + a1->field_0, v5 + a1->field_4, a1->field_0, a1->field_4 + a3, v4);
			break;
		case 0xAu:
			v6 = a1;
			sub_473380(a1->field_0, a1->field_4, a1->field_0 + v5, a1->field_4 + v5, *getMemIntPtr(0x85B3FC, 956));
			result = sub_473380(v6->field_0, a3 + v6->field_4, v5 + v6->field_0, v6->field_4 + v5, v4);
			break;
		default:
			return result;
		}
	} else {
		nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 956));
		nox_client_drawPixel_49EFA0(a1->field_0, a1->field_4);
	}
	return result;
}

//----- (00473380) --------------------------------------------------------
int sub_473380(int a1, int a2, int a3, int a4, int a5) {
	nox_client_drawSetColor_434460(a5);
	nox_client_drawAddPoint_49F500(a1, a2);
	nox_client_drawAddPoint_49F500(a3, a4);
	return nox_client_drawLineFromPoints_49E4B0();
}

//----- (004733B0) --------------------------------------------------------
int sub_4733B0(uint32_t* a1) {
	int v1; // esi
	int v2; // edi

	v1 = a1[1] + 4;
	v2 = *a1 - 2;
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 -= 8;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	v2 += 4;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 += 4;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawAddPoint_49F500(v2 - 4, v1);
	return nox_client_drawLineFromPoints_49E4B0();
}

//----- (00473420) --------------------------------------------------------
int sub_473420(uint32_t* a1) {
	int v1; // edi
	int v2; // esi

	v1 = a1[1] + 6;
	v2 = *a1 - 4;
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 -= 12;
	v2 -= 2;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 += 6;
	v2 += 4;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 -= 6;
	v2 += 2;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 += 6;
	v2 += 2;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 -= 6;
	v2 += 4;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	v1 += 12;
	v2 -= 2;
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v2, v1);
	nox_client_drawAddPoint_49F500(v2 - 8, v1);
	return nox_client_drawLineFromPoints_49E4B0();
}

//----- (004734F0) --------------------------------------------------------
void nox_video_drawCircle_4B0B90(int a1, int a2, int a3);
void nox_video_drawCircleRad3_4734F0(int* a1) { nox_video_drawCircle_4B0B90(a1[0], a1[1], 3); }

//----- (00473510) --------------------------------------------------------
int nox_client_drawRectLines_473510(int a1, int a2, int a3, int a4) {
	int v4; // esi
	int v5; // edi

	nox_client_drawAddPoint_49F500(a1, a2);
	v4 = a1 + a3 - 1;
	nox_client_drawAddPoint_49F500(v4, a2);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v4, a2);
	v5 = a4 - 1 + a2;
	nox_client_drawAddPoint_49F500(v4, v5);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v4, v5);
	nox_client_drawAddPoint_49F500(a1, v5);
	return nox_client_drawLineFromPoints_49E4B0();
}

//----- (00473570) --------------------------------------------------------
void nox_xxx_minimapDrawPoint_473570(int xLeft, int yTop) {
	if (nox_xxx_minimap_587000_149232 > 1200) {
		nox_xxx_drawPointMB_499B70(xLeft, yTop, (nox_xxx_minimap_587000_149232 < 1750) + 2);
	} else {
		nox_xxx_drawPointMB_499B70(xLeft, yTop, 4);
	}
}

//----- (004735C0) --------------------------------------------------------
void sub_4735C0(int xLeft, int yTop) {
	if (nox_xxx_minimap_587000_149232 > 1200) {
		nox_xxx_drawPointMB_499B70(xLeft, yTop, (nox_xxx_minimap_587000_149232 < 1750) + 4);
	} else {
		nox_xxx_drawPointMB_499B70(xLeft, yTop, 6);
	}
}

//----- (004738E0) --------------------------------------------------------
int nox_xxx_drawMinimapAndLines_4738E0() {
	int result;   // eax
	nox_drawable* v1; // eax

	result = 1;
	if (nox_client_gui_flag_1556112 != 1) {
		if (getMemByte(0x5D4594, 1096424) & 1) {
			v1 = nox_xxx_netSpriteByCodeDynamic_45A6F0(nox_player_netCode_85319C);
			nox_xxx_drawMinimap4Sprite_4725C0(v1);
		}
		result = nox_xxx_drawMessageLines_445530();
	}
	return result;
}

//----- (00473920) --------------------------------------------------------
void nox_xxx____setargv_11_473920() { *getMemU32Ptr(0x5D4594, 1096520) = 1; }

//----- (00473930) --------------------------------------------------------
char* sub_473930() {
	char* result; // eax

	*getMemU32Ptr(0x5D4594, 1096456) = nox_xxx_gLoadAnim_42FA20("ConfusedBirdies");
	result = nox_xxx_gLoadAnim_42FA20("SphericalShieldAnim");
	*getMemU32Ptr(0x5D4594, 1096460) = result;
	return result;
}

//----- (00473960) --------------------------------------------------------
int sub_473960() {
	int result; // eax

	result = 0;
	*getMemU32Ptr(0x5D4594, 1096456) = 0;
	*getMemU32Ptr(0x5D4594, 1096460) = 0;
	return result;
}

//----- (004739E0) --------------------------------------------------------
int sub_4739E0(uint32_t* a1, int2* a2, int2* a3) {
	int result; // eax

	a3->field_0 = a2->field_0 + *a1 - a1[4];
	result = a2->field_4;
	a3->field_4 = result + a1[1] - a1[5];
	return result;
}

//----- (00473A10) --------------------------------------------------------
int sub_473A10(uint32_t* a1, int2* a2, uint32_t* a3) {
	int result; // eax

	*a3 = a2->field_0 + a1[4] - *a1;
	result = a2->field_4;
	a3[1] = result + a1[5] - a1[1];
	return result;
}

//----- (00473C10) --------------------------------------------------------
uint32_t nox_xxx_wallFlags(int i);
void nox_xxx_drawWalls_473C10(nox_draw_viewport_t* vp, void* data) {
	uint32_t* a1 = vp;
	unsigned char* a2 = data;
	unsigned char* v3; // esi
	unsigned char v4;  // dl
	int v5;            // ecx
	int v6;            // ebx
	int v7;            // ebp
	int v8;            // eax
	int v9;            // edi
	int v10;           // eax
	int v11;           // ecx
	int v12;           // eax
	int v13;           // ecx
	int v14;           // eax
	int v15;           // edx
	int v16;           // eax
	int v17;           // ecx
	int v18;           // eax
	int v19;           // ecx
	bool v20;          // zf
	int v21;           // edx
	int v22;           // ebx
	int v23;           // ecx
	int v24;           // edx
	int v25;           // eax
	unsigned char v26; // al
	unsigned char v27; // al
	char v28;          // cl
	int v29;           // eax
	int v30;           // eax
	int v31;           // eax
	int* v32;          // edi
	int v33;           // eax
	int v34;           // eax
	int v35;           // eax
	int v36;           // edx
	int v37;           // eax
	int v38;           // eax
	int v39;           // eax
	int v40;           // edx
	int v41;           // eax
	int v42;           // eax
	int v43;           // eax
	int v44;           // edx
	int v45x;         // eax
	int v45y;         // eax
	int v46;           // ebx
	int v47;           // ebp
	int v48;           // eax
	int v49;           // ecx
	int v50;           // edx
	int v51;           // eax
	int v52;           // eax
	int v53;           // eax
	uint8_t* v54;      // edi
	int v55x;         // eax
	int v55y;         // eax
	int v56;           // ebx
	int v57;           // ebp
	int v58;           // eax
	int v59;           // edx
	int v60;           // ecx
	int v61;           // eax
	int v63;           // [esp-18h] [ebp-80h]
	int v64;           // [esp-14h] [ebp-7Ch]
	int v65;           // [esp-10h] [ebp-78h]
	int v66;           // [esp-Ch] [ebp-74h]
	int v67;           // [esp-8h] [ebp-70h]
	int v68;           // [esp-4h] [ebp-6Ch]
	int v69;           // [esp-4h] [ebp-6Ch]
	int a3;            // [esp+10h] [ebp-58h]
	int a4;            // [esp+14h] [ebp-54h]
	int v72;           // [esp+18h] [ebp-50h]
	int v73;           // [esp+1Ch] [ebp-4Ch]
	int v74;           // [esp+20h] [ebp-48h]
	int v75;           // [esp+24h] [ebp-44h]
	int v76;           // [esp+28h] [ebp-40h]
	int2 v77;          // [esp+2Ch] [ebp-3Ch]
	int2 a2a;          // [esp+34h] [ebp-34h]
	int2 a1a;          // [esp+3Ch] [ebp-2Ch]
	int2 v80;          // [esp+44h] [ebp-24h]
	int2 v81;          // [esp+4Ch] [ebp-1Ch]
	int v82;           // [esp+54h] [ebp-14h]
	int v83[3];        // [esp+5Ch] [ebp-Ch]
	int v84;           // [esp+70h] [ebp+8h]

	v3 = a2;
	a4 = nox_win_width;
	v72 = 0;
	a3 = 0;
	if (!a2) {
		return;
	}
	v4 = a2[4];
	if (!(v4 & 1)) {
		return;
	}
	v5 = a2[6];
	v6 = *a1 + 23 * a2[5] - a1[4];
	v82 = *a1 + 23 * a2[5] - a1[4];
	v7 = a1[1] + 23 * v5 - a1[5];
	v74 = *getMemU32Ptr(0x587000, 149364 + 4 * a2[3]);
	v8 = v74;
	if (v74 == -1) {
		v8 = *a2;
		v74 = *a2;
	}
	v84 = v8;
	if (v8) {
		if (v8 == 1 && v4 & 0x40) {
			v84 = 12;
		}
	} else if (v4 & 0x40) {
		v84 = 11;
	}
	if (*getMemU32Ptr(0x587000, 80808)) {
		v9 = 16 * v74;
		v10 = *getMemU32Ptr(0x587000, 85440 + 16 * v74);
		v11 = *getMemU32Ptr(0x587000, 85448 + 16 * v74);
		a1a.field_4 = v7 + *getMemU32Ptr(0x587000, 85444 + 16 * v74);
		v12 = v6 + v10;
		a2a.field_4 = v7 + *getMemU32Ptr(0x587000, 85452 + 16 * v74);
		v13 = v6 + v11;
		a1a.field_0 = v12;
		a2a.field_0 = v13;
		if (v74 == 7 || v74 == 9) {
			if (sub_4C42A0(&a1a, &a2a, &a3, &a4)) {
				v22 = 1;
			} else {
				v22 = 0;
				a3 = a2a.field_0;
			}
			v23 = *getMemU32Ptr(0x587000, 85508 + v9);
			a1a.field_0 = v82 + *getMemU32Ptr(0x587000, 85504 + v9);
			v24 = v82 + *getMemU32Ptr(0x587000, 85512 + v9);
			v25 = *getMemU32Ptr(0x587000, 85516 + v9);
			a1a.field_4 = v7 + v23;
			a2a.field_0 = v24;
			a2a.field_4 = v7 + v25;
			if (sub_4C42A0(&a1a, &a2a, &a3, &a4)) {
				v19 = a3;
			} else {
				if (!v22) {
					return;
				}
				if (a4 > a1a.field_0) {
					a4 = a1a.field_0;
				}
				v19 = a3;
			}
		} else {
			if (v74 != 8 && v74 != 10) {
				if (!sub_4C42A0(&a1a, &a2a, &a3, &a4)) {
					return;
				}
				v19 = a3;
			} else {
				v76 = v13;
				v75 = v12;
				if (sub_4C42A0(&a1a, &a2a, &v75, &v76)) {
					v73 = v76 - v75 >= 3;
				} else {
					v73 = 0;
				}
				v14 = *getMemU32Ptr(0x587000, 85504 + 16 * v74);
				v15 = *getMemU32Ptr(0x587000, 85516 + 16 * v74);
				v80.field_4 = v7 + *getMemU32Ptr(0x587000, 85508 + 16 * v74);
				v16 = v6 + v14;
				v17 = v6 + *getMemU32Ptr(0x587000, 85512 + 16 * v74);
				v80.field_0 = v16;
				a3 = v16;
				v81.field_0 = v17;
				a4 = v17;
				v81.field_4 = v7 + v15;
				v18 = sub_4C42A0(&v80, &v81, &a3, &a4);
				v19 = a3;
				v20 = v18 == 0;
				if (v20) {
					v21 = 0;
				} else {
					v21 = a4 - a3 >= 3;
				}
				if (v73) {
					if (v21) {
						if (a3 > v75) {
							v19 = v75;
							a3 = v75;
						}
						if (v19 <= v80.field_0) {
							v19 = 0;
							a3 = 0;
						}
						if (a4 < v76) {
							a4 = v76;
						}
						if (a4 >= v81.field_0) {
							a4 = nox_win_width;
						}
					} else {
						v19 = v75;
						a3 = v75;
						a4 = v76;
						if (v74 != 8) {
							v84 = 1;
							if (v19 == v80.field_0) {
								v19 = 0;
								a3 = 0;
							}
						} else {
							v84 = 0;
							if (v76 == v81.field_0) {
								a4 = nox_win_width;
							}
						}
					}
				} else {
					if (!v21) {
						return;
					}
					v84 = (v74 != 8) + 13;
					if (a4 == v81.field_0) {
						a4 = nox_win_width;
					}
					if (v19 == v80.field_0) {
						v19 = 0;
						a3 = 0;
					}
				}
			}
		}
		if (v19 >= a4) {
			v26 = v3[4];
			v3[3] = 0;
			v3[4] = v26 & 0xFC;
			return;
		}
	}
	v27 = v3[4];
	v28 = v3[4] & 2;
	if (!v28) {
		v29 = (v3[4] >> 2) & 2;
		goto LABEL_64;
	}
	if (*getMemU32Ptr(0x5D4594, 805848) && nox_client_translucentFrontWalls_805844) {
		if (!nox_client_highResFrontWalls_80820 && nox_client_highResFloors_154952) {
			v72 |= 4u;
			goto LABEL_61;
		}
		v72 = 8;
	}
	if (!nox_client_highResFrontWalls_80820) {
		v72 |= 4u;
	}
LABEL_61:
	v29 = (v27 & 8 | 4u) >> 2;
LABEL_64:
	v73 = v29;
	if (v28 && nox_client_translucentFrontWalls_805844 && !(nox_xxx_wallFlags(v3[1]) & 4)) {
		v30 = v72;
		LOBYTE(v30) = v72 | 2;
		v72 = v30;
	} else {
		v72 |= 1u;
	}
	if (*getMemU32Ptr(0x587000, 80816)) {
		switch (v74) {
		case 0:
		case 3:
			v31 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * (v31 + 1);
			v32 = sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v33 = v32[2];
				v32 = v83;
				v83[2] = v33;
			}
			v77.field_0 += 23;
			v77.field_4 -= 23;
			v34 = sub_469920(&v77);
			break;
		case 1:
		case 4:
			v35 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * v35;
			v32 = sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v36 = v32[2];
				v32 = v83;
				v83[2] = v36;
			}
			v77.field_0 += 23;
			v77.field_4 += 23;
			v34 = sub_469920(&v77);
			break;
		case 7:
			v37 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * v37;
			v32 = sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v38 = v32[2];
				v32 = v83;
				v83[2] = v38;
			}
			v77.field_0 += 23;
			v34 = sub_469920(&v77);
			break;
		case 8:
			v39 = v3[6];
			v77.field_0 = 23 * v3[5] + 11;
			v77.field_4 = 23 * v39 + 11;
			v32 = sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v40 = v32[2];
				v32 = v83;
				v83[2] = v40;
			}
			v77.field_0 -= 34;
			v77.field_4 -= 34;
			v34 = sub_469920(&v77);
			break;
		case 10:
			v41 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * (v41 + 1);
			v32 = sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v42 = v32[2];
				v32 = v83;
				v83[2] = v42;
			}
			v77.field_0 += 11;
			v77.field_4 -= 11;
			v34 = sub_469920(&v77);
			break;
		default:
			v43 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * (v43 + 1);
			v32 = sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v44 = v32[2];
				v32 = v83;
				v83[2] = v44;
			}
			v77.field_0 += 23;
			v34 = sub_469920(&v77);
			break;
		}
		v74 = v34;
		nox_xxx_getWallDrawOffset_46A3F0(v3[1], v84, v3[2], v73, &v45x, &v45y);
		v46 = v82 + v45x - 51;
		v47 = -73 - v45y + v7;
		sub_4345F0(1);
		v48 = *((uint8_t*)v32 + 8);
		LOBYTE(v49) = *((uint8_t*)v32 + 4);
		LOBYTE(v50) = *(uint8_t*)v32;
		nox_draw_setColorMultAndIntensityRGB_433CD0(v50, v49, v48);
		if (!(v72 & 2)) {
			v69 = v72;
			v66 = a4;
			v65 = a3;
			v64 = nox_win_height;
			v63 = v74;
			v52 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
			nox_xxx_edgeDraw_480EF0(v52, v46, v47, v32, v63, v64, v65, v66, 0, v69);
			goto LABEL_106;
		}
		if (!sub_47D380(a3, a4)) {
			goto LABEL_106;
		}
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(0x80u);
		sub_47D400(nox_client_highResFrontWalls_80820 == 0, a1[5]);
		v68 = v47;
		v67 = v46;
		v51 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
	} else {
		v53 = v3[6];
		v77.field_0 = 23 * v3[5] + 11;
		v77.field_4 = 23 * v53 + 11;
		v54 = sub_469920(&v77);
		nox_xxx_getWallDrawOffset_46A3F0(v3[1], v84, v3[2], v73, &v55x, &v55y);
		v56 = v82 + v55x - 50;
		v57 = -72 - v55y + v7;
		sub_4345F0(1);
		LOBYTE(v59) = v54[8];
		v58 = v54[4];
		LOBYTE(v60) = *v54;
		nox_draw_setColorMultAndIntensityRGB_433CD0(v60, v58, v59);
		if (!(v72 & 2)) {
			if (sub_47D380(a3, a4)) {
				sub_47D400(nox_client_highResFrontWalls_80820 == 0, a1[5]);
				v61 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
				nox_client_drawImageAt_47D2C0(v61, v56, v57);
				sub_47D400(0, 0);
			}
			goto LABEL_106;
		}
		if (!sub_47D380(a3, a4)) {
			goto LABEL_106;
		}
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(0x80u);
		sub_47D400(nox_client_highResFrontWalls_80820 == 0, a1[5]);
		v68 = v57;
		v67 = v56;
		v51 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
	}
	nox_client_drawImageAt_47D2C0(v51, v67, v68);
	sub_47D400(0, 0);
	nox_client_drawEnableAlpha_434560(0);
LABEL_106:
	sub_4345F0(0);
	v3[3] = 0;
	v3[4] &= 0xFC;
	*((uint32_t*)v3 + 3) = 1;
	return;
}
// 474366: variable 'v50' is possibly undefined
// 474366: variable 'v49' is possibly undefined
// 4744A3: variable 'v60' is possibly undefined
// 4744A3: variable 'v59' is possibly undefined

//----- (00474B40) --------------------------------------------------------
int sub_474B40(nox_drawable* dr) {
	int a1 = dr;
	uint32_t* v1; // edi
	uint32_t* v2; // eax
	int v3;       // eax

	v1 = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
	if (v1) {
		v2 = nox_xxx_objGetTeamByNetCode_418C80(*(uint32_t*)(a1 + 128));
		if (v2) {
			if (nox_player_netCode_85319C == *(uint32_t*)(a1 + 128) ||
				nox_xxx_servCompareTeams_419150((int)v1, (int)v2)) {
				return 1;
			}
		}
	}
	v3 = *getMemU32Ptr(0x852978, 8);
	if (a1 == *getMemU32Ptr(0x852978, 8)) {
		return 1;
	}
	if (*getMemU32Ptr(0x852978, 8)) {
		if (!nox_client_drawable_testBuff_4356C0(*getMemIntPtr(0x852978, 8), 21)) {
			v3 = *getMemU32Ptr(0x852978, 8);
			goto LABEL_9;
		}
		return 1;
	}
LABEL_9:
	if (*(uint8_t*)(a1 + 112) & 4) {
		if (a1 != v3) {
			nox_common_playerInfoGetByID_417040(*(uint32_t*)(a1 + 128));
		}
	}
	return 0;
}

//----- (004756E0) --------------------------------------------------------
int nox_xxx_sprite_4756E0_drawable(nox_drawable* dr) {
	uint32_t* a1 = dr;
	int result;           // eax
	int (*v2)(int*, int); // esi
	int v3;               // edx
	int v4;               // ecx

	result = 0;
	v2 = (int (*)(int*, int))a1[75];
	if (v2) {
		v3 = a1[30];
		v4 = a1[28];
		if (!(v3 & 0x1000) && v3 & 1 && (v2 == nox_thing_static_draw || v2 == nox_thing_static_random_draw) &&
			!(v4 & 0x80800000) && (v3 & 0x48 || v4 & 0x400000) && !(v3 & 0x800)) {
			result = 1;
		}
	}
	return result;
}

//----- (00475740) --------------------------------------------------------
int nox_xxx_sprite_475740_drawable(nox_drawable* dr) {
	uint32_t* a1 = dr;
	int result;           // eax
	int (*v2)(int*, int); // edx
	int v3;               // ebx
	int v4;               // ecx

	result = 0;
	v2 = (int (*)(int*, int))a1[75];
	if (v2) {
		v3 = a1[30];
		v4 = a1[28];
		if (!(v3 & 0x1000)) {
			if (v3 & 1) {
				result = 1;
				if ((v2 == nox_thing_static_draw || v2 == nox_thing_static_random_draw) && !(v4 & 0x80800000) &&
					!(v3 & 0x800) && (v3 & 0x48 || v4 & 0x400000)) {
					result = 0;
				}
			}
		}
	}
	return result;
}

//----- (004757A0) --------------------------------------------------------
int nox_xxx_sprite_4757A0_drawable(nox_drawable* dr) {
	int a1 = dr;
	int result; // eax
	int v2;     // ecx

	result = 0;
	if (*(uint32_t*)(a1 + 300)) {
		v2 = *(uint32_t*)(a1 + 120);
		if (!(v2 & 0x1000)) {
			if (v2 & 0x4000) {
				if (v2 & 0x40) {
					result = 1;
				}
			}
		}
	}
	return result;
}

//----- (004757D0) --------------------------------------------------------
int sub_4757D0_drawable(nox_drawable* dr) {
	uint32_t* a1 = dr;
	int result; // eax
	int v2;     // ecx

	result = 0;
	if (a1[75]) {
		v2 = a1[30];
		if (!(v2 & 1) && (!(a1[28] & 0x2000) || v2 & 0x1000000) && !(v2 & 0x1000)) {
			result = 1;
		}
	}
	return result;
}

int nox_xxx_drawAllMB_475810_draw_B(nox_draw_viewport_t* vp) {
	int v10 = 1;
	int v11;
	float2 v38;
	if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_FLOOR_RENDERING) ||
		(v38.field_0 = (double)vp->field_6, v38.field_4 = (double)vp->field_7,
		 v11 = nox_xxx_tileNFromPoint_411160(&v38), v11 == 255) ||
		v11 == -1) {
		v10 = 0;
	}
	return v10;
}
