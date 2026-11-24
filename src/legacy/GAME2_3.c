#include <math.h>

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5_2.h"
#include "client__drawable__drawable.h"
#include "common__system__team.h"

#include "client__gui__chathelp.h"
#include "client__gui__gadgets__listbox.h"
#include "client__gui__guicon.h"
#include "client__gui__guiinv.h"
#include "client__gui__guiquit.h"
#include "client__gui__guivote.h"
#include "client__gui__window.h"
#include "client__network__cdecode.h"
#include "client__network__deathmsg.h"
#include "client__shell__noxworld.h"

#include "client__draw__fx.h"
#include "client__gui__guibook.h"
#include "client__video__draw_common.h"

#include "common/fs/nox_fs.h"
#include "common__magic__speltree.h"
#include "common__net_list.h"
#include "defs.h"
#include "input_common.h"
#include "operators.h"

extern uint32_t dword_8531A0_2572;
extern uint32_t dword_5d4594_1305788;
extern uint32_t nox_server_sanctuaryHelp_54276;
extern uint32_t dword_5d4594_1197308;
extern uint32_t dword_5d4594_1197328;
extern uint32_t dword_5d4594_1197352;
extern uint32_t dword_5d4594_1197336;
extern uint32_t dword_5d4594_1197356;
extern uint32_t dword_5d4594_1197332;
extern nox_window* dword_5d4594_1193712;
extern uint32_t nox_server_connectionType_3596;
extern nox_window* dword_5d4594_1197316;
extern nox_window* dword_5d4594_1197320;
extern uint32_t nox_wol_servers_sorting_166704;
extern uint32_t dword_5d4594_1197324;
extern nox_window* dword_5d4594_1305680;
extern nox_window* dword_5d4594_1197312;
extern nox_window* dword_5d4594_1303452;
extern nox_window* dword_5d4594_1305684;
extern uint32_t nox_player_netCode_85319C;
extern int nox_win_width;
extern int nox_win_height;

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_black_2650656;

nox_render_data_t* nox_draw_curDrawData_3799572 = 0;

nox_list_item_t nox_gui_wol_servers_list = {0};


void nox_gui_winParentsReset_4A0CF0();
nox_window* nox_gui_winParentsTop_4A14F0();
nox_window* nox_gui_winParentsPop_4A18A0();
void nox_gui_winParentsPush_4A18C0(nox_window* win);

//----- (0048CA70) --------------------------------------------------------
#if 0 // C reference for the observer visibility fixture.
int nox_xxx_showObserverWindow_48CA70(int a1) { return nox_window_set_hidden(dword_5d4594_1193712, a1); }
#endif

// 0x494A60 and 0x494C30 were ported to network_update_stream_494a60.go.
// The PE32 versions stored packet cursors and drawable pointers in int values,
// which truncates both on native-width clients.

//----- (0049A3D0) --------------------------------------------------------
// EquipmentData contains native pointers and is 48 bytes on 64-bit hosts.
// Keep the wire-facing C entry point, but update the Go-owned NPC layout in a
// native-width implementation instead of indexing the original 24-byte PE32
// records and fixed offsets 680/1304/1308.
extern void* nox_xxx_clientEquipNPC_native_49A3D0(uint8_t opcode, int npc_id, uint32_t item_type,
											 uint8_t* modifiers);
char* nox_xxx_clientEquip_49A3D0(uint8_t opcode, int npc_id, uint32_t item_type, const uint8_t modifiers[4]) {
	return (char*)nox_xxx_clientEquipNPC_native_49A3D0(opcode, npc_id, item_type, (uint8_t*)modifiers);
}

//----- (0049C810) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_49C810() { return dword_5d4594_1305680 != 0; }
#endif

//----- (0049CB40) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_49CB40() { return dword_5d4594_1305684 != 0; }
#endif

//----- (0049FDB0) --------------------------------------------------------
void sub_49FDB0(int a1) {
	unsigned char* v1; // ebx
	int j;             // esi
	int v3;            // eax
	unsigned char* v4; // ebx
	int i;             // esi
	int v6;            // eax
	int v7;            // [esp+0h] [ebp-90h]
	char v8[140];      // [esp+4h] [ebp-8Ch]

	if (!dword_5d4594_1305788) {
		if (0) {
			v7 = 0;
			if (*getMemU32Ptr(0x587000, 166016 + 4 * a1) > 0) {
				v4 = getMemAt(0x587000, 166032 + 80 * a1);
				do {
					for (i = 0; i < (char)*v4; ++i) {
						v6 = 8 * (12 * a1 + (char)v4[i + 1]);
						sub_420DA0(*getMemFloatPtr(0x587000, 165360 + v6), *getMemFloatPtr(0x587000, 165364 + v6));
					}
					strcpy(&v8[4], *((const char**)v4 + 3));
					sub_4211D0(v8);
					sub_4214D0();
					v4 += 16;
					++v7;
				} while (v7 < *getMemIntPtr(0x587000, 166016 + 4 * a1));
			}
		} else {
			v1 = getMemAt(0x587000, 165744);
			do {
				for (j = 0; j < (char)*v1; ++j) {
					v3 = 8 * (char)v1[j + 1];
					sub_420DA0(*getMemFloatPtr(0x587000, 165104 + v3), *getMemFloatPtr(0x587000, 165108 + v3));
				}
				strcpy(&v8[4], *((const char**)v1 + 3));
				sub_4211D0(v8);
				sub_4214D0();
				v1 += 16;
			} while ((int)v1 < (int)getMemAt(0x587000, 166016));
		}
		dword_5d4594_1305788 = 1;
	}
}

//----- (0049FF20) --------------------------------------------------------
uint32_t* sub_49FF20() {
	uint32_t* result; // eax

	result = *(uint32_t**)&dword_5d4594_1305788;
	if (dword_5d4594_1305788) {
		result = sub_421B10();
		dword_5d4594_1305788 = 0;
	}
	return result;
}
