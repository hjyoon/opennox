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
extern void* nox_alloc_pixelSpan_1301844;
extern nox_window* dword_5d4594_1197320;
extern uint32_t nox_wol_servers_sorting_166704;
extern uint32_t dword_5d4594_1197324;
extern nox_window* dword_5d4594_1305680;
extern uint32_t dword_5d4594_1301848;
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

//----- (0048D4F0) --------------------------------------------------------
int sub_48D4F0(unsigned short a1, unsigned short a2) {
	unsigned short v2; // cx

	v2 = 10000;
	if (a1 - 10000 < 0) {
		if (a2 >= 0xFFFF - (unsigned short)(10000 - a1)) {
			return 1;
		}
		v2 = a1;
	}
	return a2 < a1 && a2 >= a1 - v2;
}

// MSG_SEQ_IMPORTANT reorder records use the third 32-bit list word as the
// sequence number in GAME.EXE. On a native 64-bit list that word is part of a
// pointer, so store the sequence beside the widened list header. These layout
// checks retain the original PE32 offsets while giving AMD64/ARM64 full-width
// links and payload pointers.
typedef struct nox_net_reorder_entry_t {
	nox_list_item_t list;
#if UINTPTR_MAX > UINT32_MAX
	uint32_t sequence;
	uint32_t reserved_1c;
#else
	uint32_t reserved_0c;
#endif
	uint64_t received_at_ms;
	uint16_t size;
	uint8_t reserved_before_data[6];
	uint8_t data[];
} nox_net_reorder_entry_t;

_Static_assert(offsetof(nox_net_reorder_entry_t, received_at_ms) == (sizeof(void*) == 4 ? 16 : 32),
	"wrong native offset of sequenced-packet timestamp");
_Static_assert(offsetof(nox_net_reorder_entry_t, size) == (sizeof(void*) == 4 ? 24 : 40),
	"wrong native offset of sequenced-packet size");
_Static_assert(offsetof(nox_net_reorder_entry_t, data) == (sizeof(void*) == 4 ? 32 : 48),
	"wrong native offset of sequenced-packet payload");

static nox_list_item_t nox_net_reorder_list_48D560;
static nox_net_reorder_entry_t* nox_net_reorder_current_1197352;
static nox_net_reorder_entry_t* nox_net_reorder_pending_1197356;
static bool nox_net_reorder_initialized;

static void nox_net_reorder_ensure_48D560(void) {
	if (nox_net_reorder_initialized) {
		return;
	}
	nox_common_list_clear_425760(&nox_net_reorder_list_48D560);
	nox_net_reorder_initialized = true;
}

static uint32_t nox_net_reorder_sequence_48D560(const nox_net_reorder_entry_t* entry) {
#if UINTPTR_MAX > UINT32_MAX
	return entry->sequence;
#else
	return (uint32_t)(uintptr_t)entry->list.field_2;
#endif
}

static void nox_net_reorder_set_sequence_48D560(nox_net_reorder_entry_t* entry, uint32_t sequence) {
#if UINTPTR_MAX > UINT32_MAX
	entry->sequence = sequence;
#else
	entry->list.field_2 = (nox_list_item_t*)(uintptr_t)sequence;
#endif
}

static nox_net_reorder_entry_t* nox_net_reorder_first_48D560(void) {
	nox_net_reorder_ensure_48D560();
	return (nox_net_reorder_entry_t*)nox_common_list_getFirstSafe_425890(&nox_net_reorder_list_48D560);
}

static nox_net_reorder_entry_t* nox_net_reorder_next_48D560(nox_net_reorder_entry_t* entry) {
	return (nox_net_reorder_entry_t*)nox_common_list_getNextSafe_4258A0(&entry->list);
}

static int nox_net_reorder_insert_48D5A0(nox_net_reorder_entry_t* entry) {
	int index = 0;
	uint32_t sequence = nox_net_reorder_sequence_48D560(entry);
	for (nox_net_reorder_entry_t* it = nox_net_reorder_first_48D560(); it;
		 it = nox_net_reorder_next_48D560(it), ++index) {
		if (sequence <= nox_net_reorder_sequence_48D560(it)) {
			nox_common_list_append_4258E0(&it->list, &entry->list);
			return index;
		}
	}
	nox_common_list_append_4258E0(&nox_net_reorder_list_48D560, &entry->list);
	return index;
}

//----- (0048D560) --------------------------------------------------------
int sub_48D560(unsigned short a1) {
	nox_net_reorder_entry_t* v1; // eax

	v1 = nox_net_reorder_first_48D560();
	if (!v1) {
		return 0;
	}
	while (nox_net_reorder_sequence_48D560(v1) != a1) {
		v1 = nox_net_reorder_next_48D560(v1);
		if (!v1) {
			return 0;
		}
	}
	return 1;
}

//----- (0048D5A0) --------------------------------------------------------
int sub_48D5A0(const uint8_t* data) {
	uint16_t sequence;
	memcpy(&sequence, data + 1, sizeof(sequence));
	if (sub_48D4F0(*getMemU16Ptr(0x5D4594, 1197360), sequence)) {
		return 1;
	}
	if (sub_48D560(sequence)) {
		return 1;
	}
	uint8_t size = data[3];
	nox_net_reorder_entry_t* entry = calloc(offsetof(nox_net_reorder_entry_t, data) + size, 1);
	if (!entry) {
		return 0;
	}
	sub_425770(&entry->list);
	nox_net_reorder_set_sequence_48D560(entry, sequence);
	entry->size = size;
	entry->received_at_ms = nox_platform_get_ticks();
	memcpy(entry->data, data + 4, size);
	if (*getMemU16Ptr(0x5D4594, 1197360) == sequence) {
		nox_net_reorder_current_1197352 = entry;
	}
	return nox_net_reorder_insert_48D5A0(entry);
}

//----- (0048D660) --------------------------------------------------------
int sub_48D660() {
	uint64_t v0 = (uintptr_t)nox_net_reorder_current_1197352; // rax
	nox_net_reorder_entry_t* v1; // esi
	nox_net_reorder_entry_t* v2; // edi

	if (!nox_net_reorder_current_1197352) {
		if (nox_net_reorder_pending_1197356) {
			v0 = nox_platform_get_ticks() - nox_net_reorder_pending_1197356->received_at_ms;
			if (v0 > 0x7530) {
				*getMemU16Ptr(0x5D4594, 1197360) = nox_net_reorder_sequence_48D560(nox_net_reorder_pending_1197356);
				nox_net_reorder_current_1197352 = nox_net_reorder_pending_1197356;
				nox_net_reorder_pending_1197356 = nox_net_reorder_next_48D560(nox_net_reorder_pending_1197356);
				v0 = (uintptr_t)nox_net_reorder_pending_1197356;
			}
		}
	}
	v1 = nox_net_reorder_current_1197352;
	if (v1) {
		do {
			v0 = nox_net_reorder_sequence_48D560(v1);
			if ((uint32_t)v0 != *getMemU16Ptr(0x5D4594, 1197360)) {
				break;
			}
			v2 = nox_net_reorder_next_48D560(v1);
			if (!v2) {
				v2 = nox_net_reorder_first_48D560();
				if (v2 == v1) {
					v2 = 0;
				}
			}
			nox_xxx_netOnPacketRecvCli_48EA70(31, v1->data, v1->size);
			++*getMemU16Ptr(0x5D4594, 1197360);
			nox_common_list_remove_425920(&v1->list);
			free(v1);
			v1 = v2;
		} while (v2);
	}
	nox_net_reorder_pending_1197356 = v1;
	nox_net_reorder_current_1197352 = 0;
	return (int)v0;
}

//----- (0048D740) --------------------------------------------------------
int sub_48D740() {
	int result; // eax

	nox_net_reorder_ensure_48D560();
	nox_common_list_clear_425760(&nox_net_reorder_list_48D560);
	result = 0;
	nox_net_reorder_current_1197352 = 0;
	nox_net_reorder_pending_1197356 = 0;
	*getMemU16Ptr(0x5D4594, 1197360) = 0;
	return result;
}

//----- (0048D760) --------------------------------------------------------
void sub_48D760() {
	nox_net_reorder_entry_t* v0; // esi
	nox_net_reorder_entry_t* v1; // edi

	v0 = nox_net_reorder_first_48D560();
	if (v0) {
		do {
			v1 = nox_net_reorder_next_48D560(v0);
			nox_common_list_remove_425920(&v0->list);
			free(v0);
			v0 = v1;
		} while (v1);
	}
	nox_common_list_clear_425760(&nox_net_reorder_list_48D560);
	nox_net_reorder_current_1197352 = 0;
	nox_net_reorder_pending_1197356 = 0;
	*getMemU16Ptr(0x5D4594, 1197360) = 0;
}

//----- (0048D7B0) --------------------------------------------------------
int* sub_48D7B0() {
	nox_net_reorder_entry_t* result; // eax

	for (result = nox_net_reorder_first_48D560(); result; result = nox_net_reorder_next_48D560(result)) {
		;
	}
	return (int*)result;
}



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

//----- (0049AEA0) --------------------------------------------------------
int sub_49AEA0() {
	if (nox_alloc_pixelSpan_1301844) {
		nox_free_alloc_class(*(void**)&nox_alloc_pixelSpan_1301844);
		nox_alloc_pixelSpan_1301844 = 0;
	}
	if (dword_5d4594_1301848) {
		free(*(void**)&dword_5d4594_1301848);
		dword_5d4594_1301848 = 0;
	}
	return 1;
}

//----- (0049C760) --------------------------------------------------------
int nox_xxx_wnd_49C760(int a1, int a2, int* a3, int a4) {
	int v3; // esi

	if (a2 == 16391) {
		v3 = nox_xxx_wndGetID_46B0A0(a3);
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
		if (v3 == 4103) {
			sub_49C7A0();
		}
	}
	return 1;
}

//----- (0049C7A0) --------------------------------------------------------
int sub_49C7A0() {
	int result = 0; // eax

	if (dword_5d4594_1305680) {
		nox_server_sanctuaryHelp_54276 =
			((unsigned int)~(
				 nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1305680, 4104)->draw_data.field_0) >>
			 2) &
			1;
		nox_xxx_wnd_46C6E0(dword_5d4594_1305680);
		nox_xxx_wndClearCaptureMain_46ADE0(dword_5d4594_1305680);
		nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1305680);
		dword_5d4594_1305680 = 0;
		nox_xxx_windowFocus_46B500(0);
		result = nox_common_gameFlags_check_40A5C0(1);
		if (result) {
			result = sub_459D80(0);
		}
	}
	return result;
}

//----- (0049C810) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_49C810() { return dword_5d4594_1305680 != 0; }
#endif

//----- (0049CA60) --------------------------------------------------------
int sub_49CA60(nox_window* win, int event, nox_window* event_win, uintptr_t event_arg) {
	(void)win;
	(void)event_arg;
	if (event == 16391 && event_win) {
		int id = nox_xxx_wndGetID_46B0A0(event_win);
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
		if (id == 10353 && dword_5d4594_1305684) {
			nox_xxx_wnd_46C6E0(dword_5d4594_1305684);
			nox_xxx_wndClearCaptureMain_46ADE0(dword_5d4594_1305684);
			if (nox_common_gameFlags_check_40A5C0(128) && nox_server_sanctuaryHelp_54276) {
				nox_xxx_cliShowHelpGui_49C560();
			} else {
				sub_459D80(0);
			}
			nox_window* list = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1305684, 10352);
			if (list) {
				int selected = (int)nox_window_call_field_94(list, 16404, 0, 0);
				nox_server_connectionType_3596 = selected + 1;
				nox_xxx_rateUpdate_40A6D0(sub_40A710(selected + 1));
			}
			nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1305684);
			dword_5d4594_1305684 = 0;
			nox_xxx_windowFocus_46B500(0);
		}
	}
	return 1;
}

//----- (0049CB40) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_49CB40() { return dword_5d4594_1305684 != 0; }
#endif

// Most probably borders with alpha
//----- (0049CD30) --------------------------------------------------------
void sub_49CD30(int xLeft, int yTop, int a3, int a4, int a5, int a6) {
	nox_rect rc; // [esp+10h] [ebp-10h]

	if (a3) {
		if (a4) {
			if (!nox_draw_curDrawData_3799572->flag_0 ||
				(noxSetRect(&rc, xLeft, yTop, xLeft + a3, yTop + a4),
				 nox_xxx_utilRect_49F930((int4*)&rc, (int4*)&rc, (int4*)(&nox_draw_curDrawData_3799572->clip)))) {
				nox_draw_set54RGB32_434040(a5);
				sub_434080(a6);
				nox_client_drawAddPoint_49F500(xLeft, yTop);
				nox_client_drawAddPoint_49F500(xLeft + a3 - 1, yTop);
				sub_49E4F0(64);
				nox_client_drawAddPoint_49F500(xLeft + a3, yTop);
				nox_client_drawAddPoint_49F500(xLeft + a3, yTop + a4 - 1);
				sub_49E4F0(64);
				nox_client_drawAddPoint_49F500(xLeft + a3, yTop + a4);
				nox_client_drawAddPoint_49F500(xLeft, yTop + a4);
				sub_49E4F0(64);
				nox_client_drawAddPoint_49F500(xLeft, yTop + a4 - 1);
				nox_client_drawAddPoint_49F500(xLeft, yTop + 1);
				sub_49E4F0(64);
			}
		}
	}
}

//----- (0049D1C0) --------------------------------------------------------
int sub_49D1C0(void* a1, int a2, int a3) {
	sub_49E3C0(a1, a2, a3);
	return 0;
}

//----- (0049E3C0) --------------------------------------------------------
void sub_49E3C0(uint32_t* a1, int a2, unsigned int a3) {
	uint32_t* v3;  // edi
	signed int v4; // ecx

	if ((int)a3 > 0) {
		v3 = a1;
		v4 = a3 >> 2;
		if (a3 >> 2) {
			do {
				*v3 = a2;
				++v3;
			} while (v4-- > 1);
		}
		if ((int)(a3 & 3) >= 2) {
			*(uint16_t*)v3 = a2;
		}
	}
}

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

//----- (0049FFA0) --------------------------------------------------------
nox_window* nox_wol_server_row_get(const nox_gui_server_ent_t* server) {
	const nox_gui_server_node_t* node = nox_wol_server_node_from_record_const(server);
	return node ? node->row : NULL;
}

void nox_wol_server_row_set(nox_gui_server_ent_t* server, nox_window* row) {
	nox_gui_server_node_t* node = nox_wol_server_node_from_record(server);
	if (node) {
		node->row = row;
	}
}

nox_list_item_t* sub_49FFA0(int destroy_rows) {
	if (!*getMemU32Ptr(0x5D4594, 1305808) || !nox_gui_wol_servers_list.field_0 ||
		!nox_gui_wol_servers_list.field_1) {
		nox_common_list_clear_425760(&nox_gui_wol_servers_list);
	}

	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&nox_gui_wol_servers_list); item;) {
		nox_list_item_t* next = nox_common_list_getNextSafe_4258A0(item);
		nox_gui_server_node_t* node = nox_wol_server_node_from_list(item);
		nox_common_list_remove_425920(item);
		if (destroy_rows && node->row) {
			nox_xxx_windowDestroyMB_46C4E0(node->row);
		}
		free(node);
		item = next;
	}
	*getMemU32Ptr(0x5D4594, 1305808) = 1;
	return NULL;
}

//----- (004A0020) --------------------------------------------------------
nox_list_item_t* sub_4A0020(void) { return &nox_gui_wol_servers_list; }

static int nox_wol_server_sort_key(nox_gui_server_ent_t* server) {
	switch (nox_wol_servers_sorting_166704) {
	case 2:
		return server->sort_key = server->players;
	case 3:
		return server->sort_key = 32 - server->players;
	case 6:
		return server->sort_key = server->ping;
	case 7:
		return server->sort_key = 1000 - server->ping;
	case 8:
		return server->sort_key = server->status & 0x30;
	case 9:
		return server->sort_key = 48 - (server->status & 0x30);
	default:
		return 0;
	}
}

static bool nox_wol_server_sorts_before(const nox_gui_server_ent_t* server,
										const nox_gui_server_ent_t* current) {
	switch (nox_wol_servers_sorting_166704) {
	case 0:
		return nox_strcmpi(server->server_name, current->server_name) <= 0;
	case 1:
		return nox_strcmpi(server->server_name, current->server_name) >= 0;
	case 4:
		return nox_wcscmp(nox_gui_wol_gameModeString_43BCB0(server->flags),
						 nox_gui_wol_gameModeString_43BCB0(current->flags)) <= 0;
	case 5:
		return nox_wcscmp(nox_gui_wol_gameModeString_43BCB0(server->flags),
						 nox_gui_wol_gameModeString_43BCB0(current->flags)) >= 0;
	case 2:
	case 3:
	case 6:
	case 7:
	case 8:
	case 9:
		return server->sort_key <= current->sort_key;
	default:
		return false;
	}
}

static int nox_wol_server_insert_node(nox_gui_server_node_t* node) {
	nox_wol_server_sort_key(&node->server);
	int index = 0;
	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&nox_gui_wol_servers_list); item;
		 item = nox_common_list_getNextSafe_4258A0(item), ++index) {
		nox_gui_server_node_t* current = nox_wol_server_node_from_list(item);
		if (nox_wol_server_sorts_before(&node->server, &current->server)) {
			nox_common_list_append_4258E0(item, &node->list);
			return index;
		}
	}
	nox_common_list_append_4258E0(&nox_gui_wol_servers_list, &node->list);
	return index;
}

//----- (004A0030) --------------------------------------------------------
int nox_wol_servers_addResult_4A0030(nox_gui_server_ent_t* srv) {
	if (!srv) {
		return -1;
	}
	if (!nox_gui_wol_servers_list.field_0 || !nox_gui_wol_servers_list.field_1) {
		nox_common_list_clear_425760(&nox_gui_wol_servers_list);
	}
	nox_gui_server_node_t* node = calloc(1, sizeof(*node));
	if (!node) {
		return -1;
	}
	memcpy(&node->server, srv, sizeof(node->server));
	return nox_wol_server_insert_node(node);
}

//----- (004A0290) --------------------------------------------------------
void nox_wol_servers_sortBtnHandler_4A0290(int id) {
	switch (id - 10047) {
	case 0:
		nox_wol_servers_sorting_166704 = (nox_wol_servers_sorting_166704 == 0) + 0;
		break;
	case 1:
		nox_wol_servers_sorting_166704 = (nox_wol_servers_sorting_166704 == 2) + 2;
		break;
	case 2:
		nox_wol_servers_sorting_166704 = (nox_wol_servers_sorting_166704 == 4) + 4;
		break;
	case 3:
		nox_wol_servers_sorting_166704 = (nox_wol_servers_sorting_166704 == 6) + 6;
		break;
	case 4:
		nox_wol_servers_sorting_166704 = (nox_wol_servers_sorting_166704 == 8) + 8;
		break;
	}
}

//----- (004A0360) --------------------------------------------------------
nox_list_item_t* sub_4A0360(void) {
	nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&nox_gui_wol_servers_list);
	while (item) {
		nox_gui_server_node_t* node = nox_wol_server_node_from_list(item);
		nox_gui_wol_newServerLine_43B7C0(&node->server);
		item = nox_common_list_getNextSafe_4258A0(item);
	}
	return NULL;
}

//----- (004A0390) --------------------------------------------------------
nox_list_item_t* sub_4A0390(void) {
	nox_list_item_t unsorted;
	nox_common_list_clear_425760(&unsorted);
	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&nox_gui_wol_servers_list); item;) {
		nox_list_item_t* next = nox_common_list_getNextSafe_4258A0(item);
		nox_common_list_remove_425920(item);
		nox_common_list_append_4258E0(&unsorted, item);
		item = next;
	}
	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&unsorted); item;) {
		nox_list_item_t* next = nox_common_list_getNextSafe_4258A0(item);
		nox_common_list_remove_425920(item);
		nox_wol_server_insert_node(nox_wol_server_node_from_list(item));
		item = next;
	}
	return sub_4A0360();
}

//----- (004A0410) --------------------------------------------------------
int sub_4A0410(const char* a1, short a2) {
	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&nox_gui_wol_servers_list); item;
		 item = nox_common_list_getNextSafe_4258A0(item)) {
		const nox_gui_server_ent_t* server = &nox_wol_server_node_from_list_const(item)->server;
		if (!strcmp(a1, server->addr) && a2 == server->port) {
			return 0;
		}
	}
	return 1;
}

//----- (004A0490) --------------------------------------------------------
nox_gui_server_ent_t* sub_4A0490(int a1) {
	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&nox_gui_wol_servers_list); item;
		 item = nox_common_list_getNextSafe_4258A0(item)) {
		nox_gui_server_node_t* node = nox_wol_server_node_from_list(item);
		if (node->server.field_9 == (uint32_t)a1) {
			return &node->server;
		}
	}
	return NULL;
}

//----- (004A04C0) --------------------------------------------------------
nox_gui_server_ent_t* sub_4A04C0(int a1) {
	int index = 0;
	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(&nox_gui_wol_servers_list); item;
		 item = nox_common_list_getNextSafe_4258A0(item), ++index) {
		if (index == a1) {
			return &nox_wol_server_node_from_list(item)->server;
		}
	}
	return NULL;
}
