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
#include "client__draw__drawrays.h"
#include "client__gui__guibook.h"
#include "client__video__draw_common.h"

#include "common/fs/nox_fs.h"
#include "common__magic__speltree.h"
#include "common__net_list.h"
#include "defs.h"
#include "input_common.h"
#include "operators.h"

extern uint32_t dword_8531A0_2572;
extern uintptr_t dword_5d4594_1303508;
extern uint32_t dword_5d4594_1200776;
extern uint32_t dword_5d4594_1203832;
extern uint32_t dword_5d4594_1200796;
extern uint32_t dword_5d4594_1305788;
extern void* nox_alloc_healthChange_1301772;
extern uint32_t nox_server_sanctuaryHelp_54276;
extern uint32_t dword_5d4594_1197308;
extern void* nox_alloc_friendList_1203860;
extern uint32_t dword_5d4594_1305748;
extern uint32_t dword_5d4594_1197328;
extern uint32_t dword_5d4594_1197352;
extern uint32_t dword_5d4594_1197336;
extern uint32_t dword_5d4594_1197356;
extern void* dword_5d4594_1301780;
extern uint32_t dword_5d4594_1203836;
extern uint32_t dword_5d4594_1203840;
extern uint32_t dword_5d4594_1197332;
extern void* nox_alloc_chat_1197364;
extern void* dword_5d4594_1203864;
extern nox_window* dword_5d4594_1193712;
extern uint32_t nox_server_connectionType_3596;
extern void* dword_5d4594_1301776;
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

static nox_chat_bubble* nox_chat_bubble_head_48D850;
static nox_chat_bubble* nox_chat_bubble_tail_48D850;

void nox_gui_winParentsReset_4A0CF0();
nox_window* nox_gui_winParentsTop_4A14F0();
nox_window* nox_gui_winParentsPop_4A18A0();
void nox_gui_winParentsPush_4A18C0(nox_window* win);

//----- (0048C580) --------------------------------------------------------
void sub_48C580(pixel8888* a1, int num) {
	unsigned int* pix = (unsigned int*)a1;
	for (int i = num - 1; i >= 0; i--) {
		unsigned int result = *pix;
		for (unsigned int* it = &pix[i]; it > pix; --it) {
			if (result > *it) {
				result = __sync_lock_test_and_set((volatile signed int*)it, result);
			}
		}
		*pix = result;
		++pix;
	}
}

//----- (0048C690) --------------------------------------------------------
unsigned int sub_48C690(int a1, int a2, int a3, int a4) {
	return sub_48C730((a3 - a1) * (a3 - a1) + (a4 - a2) * (a4 - a2));
}

//----- (0048C6B0) --------------------------------------------------------
unsigned int sub_48C6B0(int a1, int a2) { return sub_48C730(a2 * a2 + a1 * a1); }

//----- (0048C730) --------------------------------------------------------
unsigned sub_48C730(unsigned int a1) {
	int result; // eax

	if (a1 < 0x10000) {
		if (a1 < 0x100) {
			if (a1 < 0x10) {
				if (a1 < 4) {
					result = getMemByte(0x587000, 155956 + 64 * a1) >> 7;
				} else {
					result = getMemByte(0x587000, 155956 + 16 * a1) >> 6;
				}
			} else if (a1 < 0x40) {
				result = getMemByte(0x587000, 155956 + 4 * a1) >> 5;
			} else {
				result = getMemByte(0x587000, 155956 + a1) >> 4;
			}
		} else if (a1 < 0x1000) {
			if (a1 < 0x400) {
				result = getMemByte(0x587000, 155956 + (a1 >> 2)) >> 3;
			} else {
				result = getMemByte(0x587000, 155956 + (a1 >> 4)) >> 2;
			}
		} else if (a1 < 0x4000) {
			result = getMemByte(0x587000, 155956 + (a1 >> 6)) >> 1;
		} else {
			result = getMemByte(0x587000, 155956 + (a1 >> 8));
		}
	} else if (a1 < 0x1000000) {
		if (a1 < 0x100000) {
			if (a1 < 0x40000) {
				result = getMemByte(0x587000, 155956 + (a1 >> 10)) << 1;
			} else {
				result = getMemByte(0x587000, 155956 + (a1 >> 12)) << 2;
			}
		} else if (a1 < 0x400000) {
			result = getMemByte(0x587000, 155956 + (a1 >> 14)) << 3;
		} else {
			result = getMemByte(0x587000, 155956 + (a1 >> 16)) << 4;
		}
	} else if (a1 < 0x10000000) {
		if (a1 < 0x4000000) {
			result = getMemByte(0x587000, 155956 + (a1 >> 18)) << 5;
		} else {
			result = getMemByte(0x587000, 155956 + (a1 >> 20)) << 6;
		}
	} else if (a1 < 0x40000000) {
		result = getMemByte(0x587000, 155956 + (a1 >> 22)) << 7;
	} else {
		result = (unsigned char)getMemByte(0x587000, 155956 + (a1 >> 24)) << 8;
	}
	return result;
}

//----- (0048CA70) --------------------------------------------------------
int nox_xxx_showObserverWindow_48CA70(int a1) { return nox_window_set_hidden(dword_5d4594_1193712, a1); }

//----- (0048CAD0) --------------------------------------------------------
int sub_48CAD0() {
	if (wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1197312)) {
		return 0;
	}
	nox_xxx_wnd_46C6E0(dword_5d4594_1197312);
	nox_window_set_hidden(dword_5d4594_1197312, 1);
	return 1;
}

//----- (0048D000) --------------------------------------------------------
int sub_48D000_initGuiKick() {
	nox_window* v0; // eax

	v0 = nox_new_window_from_file("GuiKick.wnd", nox_xxx_guiKick_48D0A0);
	dword_5d4594_1197312 = v0;
	if (!v0) {
		return 0;
	}
	dword_5d4594_1197316 = nox_xxx_wndGetChildByID_46B0C0(v0, 4320);
	dword_5d4594_1197320 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1197312, 4321);
	nox_window_setPos_46A9B0(dword_5d4594_1197312, (nox_win_width - dword_5d4594_1197312->width) / 2,
							 dword_5d4594_1197312->off_y);
	nox_window_set_hidden(dword_5d4594_1197312, 1);
	dword_5d4594_1197324 = 0;
	dword_5d4594_1197328 = 0;
	dword_5d4594_1197332 = 0;
	dword_5d4594_1197336 = 0;
	return 1;
}

//----- (0048D0A0) --------------------------------------------------------
int nox_xxx_guiKick_48D0A0(int a1, int a2, int* a3, int a4) {
	int v4; // eax

	if (a2 != 16391) {
		return 0;
	}
	v4 = nox_xxx_wndGetID_46B0A0(a3) - 4311;
	if (v4) {
		if (v4 == 1) {
			sub_48CAD0();
			goto LABEL_15;
		}
	} else {
		if (dword_5d4594_1197308 != 4) {
			if (dword_5d4594_1197308 == 2) {
				sub_48D340();
			} else {
				if (dword_5d4594_1197308 && dword_5d4594_1197308 != 1 && dword_5d4594_1197308 != 3) {
					goto LABEL_15;
				}
				sub_48D120();
			}
			sub_48CAD0();
			goto LABEL_15;
		}
		sub_48D410();
	}
LABEL_15:
	nox_xxx_clientPlaySoundSpecial_452D80(921, 100);
	return 1;
}

//----- (0048D120) --------------------------------------------------------
int sub_48D120() {
	int v0;             // ebp
	int* v1;            // eax
	int* v2;            // esi
	int i;              // eax
	const wchar2_t* v4;  // eax
	wchar2_t* v5;        // ebx
	int v6;             // esi
	const wchar2_t* v7;  // edi
	int result;         // eax
	int v9;             // ebp
	wchar2_t* v10;       // ebx
	int v11;            // esi
	const wchar2_t* v12; // edi

	v0 = 0;
	dword_5d4594_1197328 = dword_5d4594_1197324;
	memcpy(getMemAt(0x5D4594, 1195512), getMemAt(0x5D4594, 1193720), 0x700u);
	dword_5d4594_1197324 = 0;
	v1 = (int*)nox_window_call_field_94(dword_5d4594_1197316, 16404, 0, 0);
	v2 = v1;
	for (i = *v1; i != -1; ++v2) {
		v4 = (const wchar2_t*)nox_window_call_field_94(dword_5d4594_1197316, 16406, i, 0);
		if (v4) {
			nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1193720 + 56 * dword_5d4594_1197324), v4);
			++dword_5d4594_1197324;
		}
		i = v2[1];
	}
	if (*(int*)&dword_5d4594_1197328 > 0) {
		v5 = (wchar2_t*)getMemAt(0x5D4594, 1195512);
		do {
			v6 = 0;
			if (*(int*)&dword_5d4594_1197324 <= 0) {
				nox_xxx_voteSend_48D260(v5);
			} else {
				v7 = (const wchar2_t*)getMemAt(0x5D4594, 1193720);
				while (1) {
					if (!nox_wcscmp(v5, v7)) {
						break;
					}
					++v6;
					v7 += 28;
					if (v6 >= *(int*)&dword_5d4594_1197324) {
						nox_xxx_voteSend_48D260(v5);
						break;
					}
				}
			}
			++v0;
			v5 += 28;
		} while (v0 < *(int*)&dword_5d4594_1197328);
	}
	result = dword_5d4594_1197324;
	v9 = 0;
	if (*(int*)&dword_5d4594_1197324 > 0) {
		v10 = (wchar2_t*)getMemAt(0x5D4594, 1193720);
		do {
			v11 = 0;
			if (*(int*)&dword_5d4594_1197328 <= 0) {
				nox_xxx_netSendRenameMb_48D2D0(v10);
			} else {
				v12 = (const wchar2_t*)getMemAt(0x5D4594, 1195512);
				while (1) {
					if (!nox_wcscmp(v10, v12)) {
						break;
					}
					++v11;
					v12 += 28;
					if (v11 >= *(int*)&dword_5d4594_1197328) {
						nox_xxx_netSendRenameMb_48D2D0(v10);
						break;
					}
				}
			}
			result = dword_5d4594_1197324;
			++v9;
			v10 += 28;
		} while (v9 < *(int*)&dword_5d4594_1197324);
	}
	return result;
}

//----- (0048D260) --------------------------------------------------------
char* nox_xxx_voteSend_48D260(wchar2_t* a1) {
	char* result; // eax
	int v2;       // esi
	char v3[52];  // [esp+8h] [ebp-34h]

	result = nox_common_playerInfoGetFirst_416EA0();
	v2 = (int)result;
	if (result) {
		while (nox_wcscmp((const wchar2_t*)(v2 + 4704), a1)) {
			result = nox_common_playerInfoGetNext_416EE0(v2);
			v2 = (int)result;
			if (!result) {
				return result;
			}
		}
		*(uint16_t*)v3 = 750;
		nox_wcscpy((wchar2_t*)&v3[2], a1);
		result = (char*)nox_netlist_addToMsgListCli_40EBC0(31, 0, v3, 52);
	}
	return result;
}

//----- (0048D2D0) --------------------------------------------------------
char* nox_xxx_netSendRenameMb_48D2D0(wchar2_t* a1) {
	char* result; // eax
	int v2;       // esi
	char v3[52];  // [esp+8h] [ebp-34h]

	result = nox_common_playerInfoGetFirst_416EA0();
	v2 = (int)result;
	if (result) {
		while (nox_wcscmp((const wchar2_t*)(v2 + 4704), a1)) {
			result = nox_common_playerInfoGetNext_416EE0(v2);
			v2 = (int)result;
			if (!result) {
				return result;
			}
		}
		*(uint16_t*)v3 = 238;
		nox_wcscpy((wchar2_t*)&v3[2], a1);
		result = (char*)nox_netlist_addToMsgListCli_40EBC0(31, 0, v3, 52);
	}
	return result;
}

//----- (0048D340) --------------------------------------------------------
int sub_48D340() {
	int result; // eax

	if (nox_window_call_field_94(dword_5d4594_1197320, 16404, 0, 0)) {
		result = 0;
		dword_5d4594_1197332 = 0;
		if (dword_5d4594_1197336 == 1) {
			result = nox_xxx_clientVote_48D3E0();
			dword_5d4594_1197336 = dword_5d4594_1197332;
			return result;
		}
	} else {
		result = 1;
		dword_5d4594_1197332 = 1;
		if (!dword_5d4594_1197336) {
			result = sub_48D3B0();
			dword_5d4594_1197336 = dword_5d4594_1197332;
			return result;
		}
	}
	dword_5d4594_1197336 = result;
	return result;
}

//----- (0048D3B0) --------------------------------------------------------
int sub_48D3B0() {
	char v1[2]; // [esp+0h] [ebp-2h]

	v1[0] = -18;
	v1[1] = 4;
	return nox_xxx_netClientSend2_4E53C0(31, v1, 2, 0, 1);
}

//----- (0048D3E0) --------------------------------------------------------
int nox_xxx_clientVote_48D3E0() {
	char v1[2]; // [esp+0h] [ebp-2h]

	v1[0] = -18;
	v1[1] = 5;
	return nox_xxx_netClientSend2_4E53C0(31, v1, 2, 0, 1);
}

//----- (0048D410) --------------------------------------------------------
uint32_t* sub_48D410() {
	uint32_t* result; // eax

	result = (uint32_t*)nox_window_call_field_94(dword_5d4594_1197320, 16404, 0, 0);
	if (!result) {
		return sub_48CB10(2);
	}
	if (result == (uint32_t*)1) {
		result = sub_48CB10(3);
	}
	return result;
}

//----- (0048D450) --------------------------------------------------------
int sub_48D450() {
	int result; // eax

	nox_xxx_wnd_46C6E0(dword_5d4594_1197312);
	nox_xxx_wndClearCaptureMain_46ADE0(dword_5d4594_1197312);
	nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1197312);
	result = 0;
	dword_5d4594_1197312 = 0;
	dword_5d4594_1197316 = 0;
	dword_5d4594_1197320 = 0;
	dword_5d4594_1197324 = 0;
	dword_5d4594_1197328 = 0;
	dword_5d4594_1197332 = 0;
	dword_5d4594_1197336 = 0;
	return result;
}

//----- (0048D4A0) --------------------------------------------------------
int sub_48D4A0() {
	int result; // eax

	result = 0;
	dword_5d4594_1197332 = 0;
	dword_5d4594_1197336 = 0;
	return result;
}

//----- (0048D4B0) --------------------------------------------------------
int sub_48D4B0(int a1) {
	int result; // eax

	*getMemU32Ptr(0x5D4594, 1197304) = a1;
	if (a1 == 1) {
		result = sub_4C3460(0);
	} else {
		result = sub_4C3460(1);
	}
	return result;
}

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

//----- (0048D800) --------------------------------------------------------
int sub_48D800() {
	if (nox_alloc_chat_1197364) {
		nox_free_alloc_class((nox_alloc_class*)nox_alloc_chat_1197364);
	}
	nox_alloc_chat_1197364 = 0;
	nox_chat_bubble_head_48D850 = 0;
	nox_chat_bubble_tail_48D850 = 0;
	*getMemU32Ptr(0x5D4594, 1197368) = 0;
	return 1;
}

//----- (0048D850) --------------------------------------------------------
nox_chat_bubble* nox_xxx_netCode2ChatBubble_48D850(int net_code) {
	for (nox_chat_bubble* bubble = nox_chat_bubble_head_48D850; bubble; bubble = bubble->next) {
		if (bubble->net_code == (uint32_t)net_code) {
			return bubble;
		}
	}
	return 0;
}

//----- (0048D880) --------------------------------------------------------
void nox_xxx_createTextBubble_48D880(void* a1p, wchar2_t* a2) {
	uint8_t* data = (uint8_t*)a1p;
	nox_chat_bubble* bubble = nox_xxx_netCode2ChatBubble_48D850(*(uint16_t*)(data + 1));
	int is_new;
	if (bubble) {
		is_new = 0;
	} else {
		bubble = (nox_chat_bubble*)nox_alloc_class_new_obj_zero((nox_alloc_class*)nox_alloc_chat_1197364);
		if (!bubble) {
			return;
		}
		is_new = 1;
	}
	nox_wcscpy(bubble->text, a2);
	bubble->duration_hint = *(uint8_t*)(data + 8);
	*(uint32_t*)&bubble->drawable_x = *(uint32_t*)(data + 4);
	bubble->net_code = *(uint16_t*)(data + 1);
	uint32_t duration;
	if (*(uint16_t*)(data + 9)) {
		duration = *(uint16_t*)(data + 9);
	} else {
		duration = bubble->duration_hint / 8;
		if (duration >= 8) {
			duration = 8;
		}
		duration = gameFPS() * (duration + 2);
	}
	bubble->expire_frame = gameFrame() + duration;
	if (is_new) {
		bubble->next = 0;
		bubble->prev = nox_chat_bubble_tail_48D850;
		if (nox_chat_bubble_tail_48D850) {
			nox_chat_bubble_tail_48D850->next = bubble;
		} else {
			nox_chat_bubble_head_48D850 = bubble;
		}
		nox_chat_bubble_tail_48D850 = bubble;
	}
}

//----- (0048D990) --------------------------------------------------------
void sub_48D990(nox_draw_viewport_t* a1p) {
	int font_height = nox_xxx_guiFontHeightMB_43F320(0);
	int half_height = font_height / 2;
	sub_437260();
	sub_48DCF0(a1p);
	if (!nox_chat_bubble_head_48D850) {
		sub_437290();
		return;
	}
	for (nox_chat_bubble* bubble = nox_chat_bubble_head_48D850; bubble; bubble = bubble->next) {
		if (bubble->visible) {
			uint32_t text_color = nox_color_white_2523948;
			uint16_t* player_name = 0;
			if (bubble->drawable && bubble->drawable->flags28 & 4) {
				nox_team_t* team = nox_xxx_objGetTeamByNetCode_418C80(bubble->drawable->field_32);
				char* player_info = (char*)nox_common_playerInfoGetByID_417040(bubble->net_code);
				if (player_info) {
					player_name = (uint16_t*)(player_info + 4704);
				}
				if (team) {
					nox_team_t* resolved = nox_xxx_getTeamByID_418AB0(*((unsigned char*)team + 4));
					if (resolved) {
						text_color = nox_xxx_materialGetTeamColor_418D50(resolved);
					}
				}
			}
			int x = bubble->x;
			int y = bubble->y;
			for (int i = 0; i < 2; ++i) {
				int color;
				if (i) {
					color = *getMemU32Ptr(0x85B3FC, 956);
					--x;
					--y;
				} else {
					color = *getMemU32Ptr(0x852978, 4);
				}
				nox_client_drawSetColor_434460(color);
				int left = x - half_height;
				nox_client_drawAddPoint_49F500(x, y - half_height);
				nox_client_drawAddPoint_49F500(x - half_height, y);
				nox_client_drawLineFromPoints_49E4B0();
				int right = bubble->width + x;
				nox_client_drawAddPoint_49F500(x, y - half_height);
				nox_client_drawAddPoint_49F500(right, y - half_height);
				nox_client_drawLineFromPoints_49E4B0();
				int right_outer = half_height + right;
				nox_client_drawAddPoint_49F500(right_outer, y);
				nox_client_drawAddPoint_49F500(right, y - half_height);
				nox_client_drawLineFromPoints_49E4B0();
				int bottom = y + bubble->height;
				nox_client_drawAddPoint_49F500(right_outer, y);
				nox_client_drawAddPoint_49F500(right_outer, bottom);
				nox_client_drawLineFromPoints_49E4B0();
				int bottom_outer = half_height + bottom;
				nox_client_drawAddPoint_49F500(right, bottom_outer);
				nox_client_drawAddPoint_49F500(right_outer, bottom);
				nox_client_drawLineFromPoints_49E4B0();
				if (bubble->draw_tail) {
					int tail_right = half_height + x + bubble->width / 2;
					nox_client_drawAddPoint_49F500(right, bottom_outer);
					nox_client_drawAddPoint_49F500(tail_right, bottom_outer);
					nox_client_drawLineFromPoints_49E4B0();
					int tail_bottom = font_height + bottom_outer;
					int tail_mid = x + bubble->width / 2;
					nox_client_drawAddPoint_49F500(tail_mid, tail_bottom);
					nox_client_drawAddPoint_49F500(tail_right, bottom_outer);
					nox_client_drawLineFromPoints_49E4B0();
					int tail_left = x + bubble->width / 2 - half_height;
					nox_client_drawAddPoint_49F500(tail_mid, tail_bottom);
					nox_client_drawAddPoint_49F500(tail_left, bottom_outer);
					nox_client_drawLineFromPoints_49E4B0();
					nox_client_drawAddPoint_49F500(x, bottom_outer);
					nox_client_drawAddPoint_49F500(tail_left, bottom_outer);
				} else {
					nox_client_drawAddPoint_49F500(x, bottom_outer);
					nox_client_drawAddPoint_49F500(right, bottom_outer);
				}
				nox_client_drawLineFromPoints_49E4B0();
				nox_client_drawAddPoint_49F500(x, bottom_outer);
				nox_client_drawAddPoint_49F500(left, bottom_outer - half_height);
				nox_client_drawLineFromPoints_49E4B0();
				nox_client_drawAddPoint_49F500(left, bottom_outer - half_height - bubble->height);
				nox_client_drawAddPoint_49F500(left, bottom_outer - half_height);
				nox_client_drawLineFromPoints_49E4B0();
			}
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			nox_xxx_drawSetColor_4343E0(*getMemIntPtr(0x852978, 4));
			nox_xxx_drawStringWrapHL_43FD00(0, bubble->text, x, y, 128, 0);
			if (player_name) {
				nox_xxx_drawSetTextColor_434390(text_color);
				nox_xxx_drawStringWrapHL_43FD00(0, player_name, x, y - font_height - 1, 128, 0);
			}
		}
	}
	sub_437290();
}

//----- (0048DCF0) --------------------------------------------------------
static void nox_chat_bubble_unlink_48DCF0(nox_chat_bubble* bubble) {
	if (bubble->prev) {
		bubble->prev->next = bubble->next;
	} else {
		nox_chat_bubble_head_48D850 = bubble->next;
	}
	if (bubble->next) {
		bubble->next->prev = bubble->prev;
	} else {
		nox_chat_bubble_tail_48D850 = bubble->prev;
	}
}

void sub_48DCF0(nox_draw_viewport_t* viewport) {
	int font_height = nox_xxx_guiFontHeightMB_43F320(0);
	int order = 0;
	for (nox_chat_bubble* bubble = nox_chat_bubble_head_48D850; bubble;) {
		nox_chat_bubble* next = bubble->next;
		bubble->order = order++;
		bubble->drawable = 0;
		int code = nox_xxx_netClearHighBit_578B30(bubble->net_code);
		if (nox_xxx_netTestHighBit_578B70(bubble->net_code)) {
			bubble->drawable = nox_xxx_netSpriteByCodeStatic_45A720(code);
		} else {
			bubble->drawable = nox_xxx_netSpriteByCodeDynamic_45A6F0(code);
		}
		if (bubble->drawable) {
			bubble->drawable_x = (uint16_t)bubble->drawable->pos.x;
			bubble->drawable_y = (uint16_t)bubble->drawable->pos.y;
		}
		bubble->x = (int)(viewport->x1 + bubble->drawable_x - viewport->field_4);
		bubble->y = (int)(viewport->y1 + bubble->drawable_y - viewport->field_5);
		nox_xxx_drawGetStringSize_43F840(0, bubble->text, &bubble->width, &bubble->height, 128);
		if (bubble->width > 128) {
			bubble->width = 128;
		}
		bubble->x += bubble->width / -2;
		bubble->y += -64 - bubble->height;
		bubble->visible = 1;
		bubble->draw_tail = 1;
		if (nox_common_gameFlags_check_40A5C0(2048) &&
			(bubble->x < viewport->x1 || bubble->x > viewport->x2 ||
			 bubble->y - 64 < viewport->y1 || bubble->y - 64 > viewport->y2)) {
			bubble->visible = 0;
			bubble->draw_tail = 0;
		}
		if (bubble->visible) {
			int x = (int)viewport->x1 + font_height;
			if (bubble->x >= x) {
				int right = (int)viewport->x2;
				if (font_height + bubble->x + bubble->width <= right) {
					goto x_positioned;
				}
				x = right - bubble->width - font_height;
			}
			bubble->x = x;
			bubble->draw_tail = 0;
		x_positioned:;
			int y = (int)viewport->y1 + 2 * font_height + 2;
			if (bubble->y >= y) {
				int bottom = (int)viewport->y2;
				if (font_height + bubble->height + bubble->y <= bottom) {
					goto y_positioned;
				}
				y = bottom - bubble->height - font_height;
			}
			bubble->y = y;
			bubble->draw_tail = 0;
		y_positioned: {
				int4 rect = {bubble->x, bubble->y, bubble->x + bubble->width, bubble->y + bubble->height};
				sub_48E000(&rect, &bubble->draw_tail);
				bubble->x = rect.field_0;
				bubble->y = rect.field_4;
			}
		}
		if (gameFrame() > bubble->expire_frame) {
			nox_chat_bubble_unlink_48DCF0(bubble);
			nox_alloc_class_free_obj_first((nox_alloc_class*)nox_alloc_chat_1197364, bubble);
		}
		bubble = next;
	}
	for (nox_chat_bubble* bubble = nox_chat_bubble_head_48D850; bubble; bubble = bubble->next) {
		sub_48E240(0, (uint32_t*)bubble);
	}
}
// 48DD5E: variable 'v7' is possibly undefined
// 48DD72: variable 'v9' is possibly undefined

//----- (0048E000) --------------------------------------------------------
bool sub_48E000(int4* a1, uint32_t* a2) {
	int v2;   // ebx
	int v3;   // eax
	int v4;   // eax
	int v5;   // eax
	int v6;   // eax
	int v7;   // eax
	int v8;   // eax
	int v9;   // eax
	int4 v11; // [esp+10h] [ebp-10h]

	if (a1->field_0 < 0 || a1->field_4 < 0 || a1->field_8 > nox_win_width || a1->field_C > nox_win_height) {
		*a2 = 0;
	}
	v11.field_0 = 0;
	v11.field_4 = 0;
	v11.field_8 = 563;
	if (sub_467C80()) {
		v2 = 279;
		v11.field_C = 279;
		v3 = nox_xxx_pointInRect_4281F0((int2*)a1, &v11);
		if (v3 || (v3 = nox_xxx_pointInRect_4281F0((int2*)&a1->field_8, &v11), v3)) {
			a1->field_4 = v2;
			*a2 = 0;
			goto LABEL_13;
		}
	} else {
		v2 = 55;
		v11.field_C = 55;
		v3 = nox_xxx_pointInRect_4281F0((int2*)a1, &v11);
		if (v3) {
			a1->field_4 = v2;
			*a2 = 0;
			goto LABEL_13;
		}
		v3 = nox_xxx_pointInRect_4281F0((int2*)&a1->field_8, &v11);
		if (v3) {
			a1->field_4 = v2;
			*a2 = 0;
			goto LABEL_13;
		}
	}
LABEL_13:
	if (nox_client_getRenderGUI()) {
		v11.field_0 = 0;
		v11.field_C = nox_win_height;
		v11.field_8 = 111;
		v11.field_4 = nox_win_height - 127;
		v4 = nox_xxx_pointInRect_4281F0((int2*)a1, &v11);
		if (v4 || (v5 = nox_xxx_pointInRect_4281F0((int2*)&a1->field_8, &v11), v5)) {
			a1->field_4 += v11.field_4 - a1->field_C;
			*a2 = 0;
		}
		v11.field_4 = nox_win_height - 74;
		v11.field_0 = nox_win_width / 2 - 160;
		v11.field_8 = v11.field_0 + 320;
		v11.field_C = nox_win_height;
		v6 = nox_xxx_pointInRect_4281F0((int2*)a1, &v11);
		if (v6 || (v7 = nox_xxx_pointInRect_4281F0((int2*)&a1->field_8, &v11), v7)) {
			a1->field_4 += v11.field_4 - a1->field_C;
			*a2 = 0;
		}
		v11.field_8 = nox_win_width;
		v11.field_0 = nox_win_width - 91;
		v11.field_C = nox_win_height;
		v11.field_4 = nox_win_height - 201;
		v8 = nox_xxx_pointInRect_4281F0((int2*)a1, &v11);
		if (v8 || (v9 = nox_xxx_pointInRect_4281F0((int2*)&a1->field_8, &v11), v9)) {
			a1->field_4 += v11.field_4 - a1->field_C;
			*a2 = 0;
		}
		v3 = sub_4C3260();
		if (v3) {
			v11.field_0 = nox_win_width - 87;
			v11.field_4 = 0;
			v11.field_8 = nox_win_width;
			v11.field_C = 145;
			v3 = nox_xxx_pointInRect_4281F0((int2*)a1, &v11);
			if (v3 || (v3 = nox_xxx_pointInRect_4281F0((int2*)&a1->field_8, &v11), v3)) {
				a1->field_4 = 145;
				*a2 = 0;
			}
		}
	}
	return v3;
}
// 48E069: variable 'v3' is possibly undefined
// 48E0EF: variable 'v4' is possibly undefined
// 48E104: variable 'v5' is possibly undefined
// 48E156: variable 'v6' is possibly undefined
// 48E16B: variable 'v7' is possibly undefined
// 48E1B4: variable 'v8' is possibly undefined
// 48E1C9: variable 'v9' is possibly undefined

//----- (0048E240) --------------------------------------------------------
char sub_48E240(int a1, uint32_t* a2) {
	int v2 = 0;   // eax
	uint32_t* v3; // esi
	int v4;       // edi
	uint32_t* v5; // ebp
	char v6;      // bl
	int v7;       // edi
	int v8;       // edx
	int v9;       // edi
	int v10;      // edx
	int v12;      // [esp+Ch] [ebp-Ch]
	int v13;      // [esp+10h] [ebp-8h]
	int v14;      // [esp+14h] [ebp-4h]
	char v15;     // [esp+20h] [ebp+8h]

	(void)a1;
	v3 = a2;
	if ((nox_chat_bubble*)a2 != nox_chat_bubble_head_48D850) {
		v4 = 0;
		v5 = (uint32_t*)nox_chat_bubble_head_48D850;
		if (v5) {
			while (1) {
				v2 = v5[165];
				if (v2) {
					v2 = v5[170];
					if (v2 >= a2[170]) {
						goto LABEL_9;
					}
					v2 = sub_48E480(a2, v5);
					if (v2) {
						break;
					}
				}
				v5 = (uint32_t*)((nox_chat_bubble*)v5)->next;
				if (!v5) {
					return v2;
				}
			}
			v4 = 1;
		LABEL_9:
			if (v5 && v4) {
				v6 = 0;
				v15 = 0;
				switch (sub_48E530(v3[162] + v3[168] / 2, v3[163] + v3[169] / 2)) {
				case 17:
					v6 = -96;
					v15 = 8;
					break;
				case 18:
					v6 = 48;
					v15 = -118;
					break;
				case 20:
					v6 = -112;
					v15 = 2;
					break;
				case 33:
					v6 = -64;
					v15 = 44;
					break;
				case 34:
					v6 = -16;
					v15 = 15;
					break;
				case 36:
					v6 = -64;
					v15 = 19;
					break;
				case 65:
					v6 = 96;
					v15 = 4;
					break;
				case 66:
					v6 = 48;
					v15 = 69;
					break;
				case 68:
					v6 = 80;
					v15 = 1;
					break;
				default:
					break;
				}
				v7 = 0;
				while (1) {
					LOBYTE(v2) = 1 << v7;
					LOBYTE(v14) = 1 << v7;
					if ((unsigned char)(1 << v7) & (unsigned char)v6) {
						sub_48E6A0(v14, v3, v5, &v12, &v13);
						v2 = sub_48E5C0(v3, v12, v13);
						if (v2) {
							break;
						}
					}
					if (++v7 >= 8) {
						goto LABEL_27;
					}
				}
				v8 = v13;
				v3[162] = v12;
				v3[163] = v8;
			LABEL_27:
				if (v7 == 8) {
					v9 = 0;
					while (1) {
						LOBYTE(v2) = 1 << v9;
						LOBYTE(v14) = 1 << v9;
						if ((unsigned char)(1 << v9) & (unsigned char)v15) {
							sub_48E6A0(v14, v3, v5, &v12, &v13);
							v2 = sub_48E5C0(v3, v12, v13);
							if (v2) {
								break;
							}
						}
						if (++v9 >= 8) {
							return v2;
						}
					}
					v10 = v13;
					v3[162] = v12;
					v3[163] = v10;
				}
			}
		}
	}
	return v2;
}

//----- (0048E480) --------------------------------------------------------
int sub_48E480(uint32_t* a1, uint32_t* a2) {
	int v2; // eax
	int v3; // esi
	int v4; // edx
	int v5; // ebx
	int v7; // [esp+24h] [ebp-Ch]
	int v8; // [esp+34h] [ebp+4h]

	v2 = nox_xxx_guiFontHeightMB_43F320(0);
	v3 = a1[162];
	v7 = a1[163] - v2;
	v4 = v2 + v3 + a1[168];
	v5 = v2 + a1[163] + a1[169];
	v8 = a2[163];
	return v3 - v2 < v2 + a2[162] + a2[168] && v4 > a2[162] - v2 && v7 < v2 + v8 + a2[169] && v5 > v8 - v2;
}

//----- (0048E530) --------------------------------------------------------
int sub_48E530(int a1, int a2) {
	int v2; // edi
	int v3; // eax

	if (a1 >= nox_win_width / 3) {
		v3 = (a1 >= 2 * nox_win_width / 3) - 1;
		LOBYTE(v3) = v3 & 0xE0;
		v2 = v3 + 64;
	} else {
		v2 = 16;
	}
	if (a2 < nox_win_height / 3) {
		return v2 | 1;
	}
	if (a2 >= 2 * nox_win_height / 3) {
		return v2 | 4;
	}
	return v2 | 2;
}

//----- (0048E5C0) --------------------------------------------------------
int sub_48E5C0(uint32_t* a1, int a2, int a3) {
	int v3;       // edx
	int v4;       // eax
	uint32_t* v6; // edi
	int v7;       // ebx
	int v8;       // ebp
	int v9;       // eax
	int a2a;      // [esp+10h] [ebp-14h]
	int4 a1a;     // [esp+14h] [ebp-10h]

	a2a = 1;
	nox_xxx_guiFontHeightMB_43F320(0);
	a1a.field_0 = a2;
	v3 = a1[168];
	a1a.field_4 = a3;
	v4 = a3 + a1[169];
	a1a.field_8 = a2 + v3;
	a1a.field_C = v4;
	sub_48E000(&a1a, &a2a);
	if (!a2a) {
		return 0;
	}
	v6 = (uint32_t*)nox_chat_bubble_head_48D850;
	if (v6) {
		while (1) {
			if (a1[165]) {
				if (v6[170] >= a1[170]) {
					return 1;
				}
				v7 = a1[162];
				v8 = a1[163];
				a1[162] = a2;
				a1[163] = a3;
				v9 = sub_48E480(a1, v6);
				a2a = v9;
				a1[162] = v7;
				a1[163] = v8;
				if (v9) {
					break;
				}
			}
			v6 = (uint32_t*)((nox_chat_bubble*)v6)->next;
			if (!v6) {
				return 1;
			}
		}
		return 0;
	}
	return 1;
}

//----- (0048E6A0) --------------------------------------------------------
int* sub_48E6A0(char a1, uint32_t* a2, uint32_t* a3, int* a4, int* a5) {
	int* result;  // eax
	int v6;       // ecx
	uint32_t* v7; // ecx
	int v8;       // edx
	char* v9;     // edx
	int v10;      // edx

	result = (int*)(2 * nox_xxx_guiFontHeightMB_43F320(0));
	switch ((unsigned char)a1) {
	case 1:
		*a4 = a3[162] - a2[168] - (uint32_t)result;
		*a5 = a3[163] - a2[169] - (uint32_t)result;
		return result;
	case 2:
		*a4 = (int)result + a3[162] + a3[168];
		v6 = a3[163] - a2[169] - (uint32_t)result;
		result = a5;
		*a5 = v6;
		return result;
	case 4:
		v7 = a3;
		*a4 = a3[162] - a2[168] - (uint32_t)result;
		v9 = (char*)result + v7[163] + v7[169];
		result = a5;
		*a5 = (int)v9;
		break;
	case 8:
		v7 = a3;
		*a4 = (int)result + a3[162] + a3[168];
		v9 = (char*)result + v7[163] + v7[169];
		result = a5;
		*a5 = (int)v9;
		break;
	case 0x10:
		*a4 = a2[162];
		v8 = a3[163] - a2[169] - (uint32_t)result;
		result = a5;
		*a5 = v8;
		return result;
	case 0x20:
		*a4 = a2[162];
		v7 = a3;
		v9 = (char*)result + v7[163] + v7[169];
		result = a5;
		*a5 = (int)v9;
		break;
	case 0x40:
		v10 = a3[162] - a2[168] - (uint32_t)result;
		result = a4;
		*a4 = v10;
		*a5 = a2[163];
		break;
	case 0x80:
		*a4 = (int)result + a3[162] + a3[168];
		result = a5;
		*a5 = a2[163];
		break;
	default:
		return result;
	}
	return result;
}

//----- (0048E8E0) --------------------------------------------------------
void sub_48E8E0(int a1) {
	nox_chat_bubble* bubble = nox_xxx_netCode2ChatBubble_48D850(a1);
	if (bubble) {
		nox_chat_bubble_unlink_48DCF0(bubble);
		nox_alloc_class_free_obj_first((nox_alloc_class*)nox_alloc_chat_1197364, bubble);
	}
}

//----- (0048E940) --------------------------------------------------------
void sub_48E940() {
	nox_alloc_class_free_all((nox_alloc_class*)nox_alloc_chat_1197364);
	nox_chat_bubble_head_48D850 = 0;
	nox_chat_bubble_tail_48D850 = 0;
	*getMemU32Ptr(0x5D4594, 1197368) = 0;
}

//----- (004947E0) --------------------------------------------------------
char* sub_4947E0(int a1) {
	short v1;     // ax
	int v2;       // edi
	char* result; // eax
	int i;        // esi

	if (nox_common_gameFlags_check_40A5C0(1)) {
		v1 = nox_common_gameFlags_getVal_40A5B0();
		v2 = (unsigned short)nox_xxx_servGamedataGet_40A020(v1);
	} else {
		v2 = *((unsigned short*)nox_xxx_cliGamedataGet_416590(0) + 27);
	}
	result = nox_common_playerInfoGetFirst_416EA0();
	for (i = (int)result; result; i = (int)result) {
		if (!(*(uint8_t*)(i + 3680) & 1)) {
			if (i == a1) {
				if (nox_common_gameFlags_check_40A5C0(1024)) {
					if (*(uint32_t*)(i + 2140) >= v2) {
						*(uint32_t*)(i + 2140) = v2 - 1;
					}
				} else {
					*(uint32_t*)(i + 2136) = v2;
				}
			} else if (nox_common_gameFlags_check_40A5C0(1024)) {
				if (*(uint32_t*)(i + 2140) < v2) {
					*(uint32_t*)(i + 2140) = v2;
				}
			} else if (*(uint32_t*)(i + 2136) >= v2) {
				*(uint32_t*)(i + 2136) = v2 - 1;
			}
		}
		result = nox_common_playerInfoGetNext_416EE0(i);
	}
	return result;
}

//----- (004948B0) --------------------------------------------------------
int sub_4948B0(int a1) {
	short v1;   // ax
	int v2;     // edi
	char* i;    // esi
	int result; // eax
	int j;      // ebp
	char* v6;   // eax
	char* v7;   // esi
	int k;      // ebp
	char* v9;   // eax
	char* v10;  // esi

	if (nox_common_gameFlags_check_40A5C0(1)) {
		v1 = nox_common_gameFlags_getVal_40A5B0();
		v2 = (unsigned short)nox_xxx_servGamedataGet_40A020(v1);
	} else {
		v2 = *((unsigned short*)nox_xxx_cliGamedataGet_416590(0) + 27);
	}
	for (i = nox_server_teamFirst_418B10(); i; i = nox_server_teamNext_418B60((int)i)) {
		if (i == (char*)a1) {
			if (!nox_common_gameFlags_check_40A5C0(1024)) {
				*((uint32_t*)i + 13) = v2;
			}
		} else if (!nox_common_gameFlags_check_40A5C0(1024) && *((uint32_t*)i + 13) >= v2) {
			*((uint32_t*)i + 13) = v2 - 1;
		}
	}
	if (nox_common_gameFlags_check_40A5C0(1)) {
		result = nox_xxx_getFirstPlayerUnit_4DA7C0();
		for (j = result; result; j = result) {
			if (!nox_xxx_teamCompare2_419180(j + 48, *(uint8_t*)(a1 + 57))) {
				v6 = nox_common_playerInfoGetByID_417040(*(uint32_t*)(j + 36));
				v7 = v6;
				if (v6) {
					if (!(v6[3680] & 1)) {
						if (nox_common_gameFlags_check_40A5C0(1024)) {
							if (*((uint32_t*)v7 + 535) < v2) {
								*((uint32_t*)v7 + 535) = v2;
							}
						} else if (*((uint32_t*)v7 + 534) >= v2) {
							*((uint32_t*)v7 + 534) = v2 - 1;
						}
					}
				}
			}
			result = nox_xxx_getNextPlayerUnit_4DA7F0(j);
		}
	} else {
		result = nox_xxx_cliGetSpritePlayer_45A000();
		for (k = result; result; k = result) {
			if (!nox_xxx_teamCompare2_419180(k + 24, *(uint8_t*)(a1 + 57))) {
				v9 = nox_common_playerInfoGetByID_417040(*(uint32_t*)(k + 128));
				v10 = v9;
				if (v9) {
					if (!(v9[3680] & 1)) {
						if (nox_common_gameFlags_check_40A5C0(1024)) {
							if (*((uint32_t*)v10 + 535) < v2) {
								*((uint32_t*)v10 + 535) = v2;
							}
						} else if (*((uint32_t*)v10 + 534) >= v2) {
							*((uint32_t*)v10 + 534) = v2 - 1;
						}
					}
				}
			}
			result = sub_45A010(k);
		}
	}
	return result;
}

// 0x494A60 and 0x494C30 were ported to network_update_stream_494a60.go.
// The PE32 versions stored packet cursors and drawable pointers in int values,
// which truncates both on native-width clients.

//----- (00494F00) --------------------------------------------------------
int sub_494F00() {
	int result; // eax
	int v1;     // esi
	int v2;     // eax
	static const char* spark_names[5] = {"Spark", "BlueSpark", "YellowSpark", "CyanSpark", "GreenSpark"};

	*getMemU32Ptr(0x5D4594, 1200772) = nox_xxx_getTTByNameSpriteMB_44CFC0("Spark");
	if (!*getMemU32Ptr(0x5D4594, 1200772)) {
		return 0;
	}
	result = nox_xxx_getTTByNameSpriteMB_44CFC0("BlueSpark");
	dword_5d4594_1200776 = result;
	if (result) {
		result = nox_xxx_getTTByNameSpriteMB_44CFC0("YellowSpark");
		*getMemU32Ptr(0x5D4594, 1200780) = result;
		if (result) {
			result = nox_xxx_getTTByNameSpriteMB_44CFC0("CyanSpark");
			*getMemU32Ptr(0x5D4594, 1200784) = result;
			if (result) {
				result = nox_xxx_getTTByNameSpriteMB_44CFC0("GreenSpark");
				*getMemU32Ptr(0x5D4594, 1200788) = result;
				if (result) {
					result = nox_xxx_getTTByNameSpriteMB_44CFC0("Puff");
					*getMemU32Ptr(0x5D4594, 1200792) = result;
					if (result) {
						v1 = 0;
						while (1) {
							v2 = nox_xxx_getTTByNameSpriteMB_44CFC0((char*)spark_names[v1 / 4]);
							*getMemU32Ptr(0x5D4594, 1200812 + v1) = v2;
							if (!v2) {
								break;
							}
							v1 += 4;
							if (v1 >= 20) {
								dword_5d4594_1200796 = nox_xxx_getTTByNameSpriteMB_44CFC0("VioletSpark");
								return dword_5d4594_1200796 != 0;
							}
						}
						return 0;
					}
				}
			}
		}
	}
	return result;
}

//----- (00494FF0) --------------------------------------------------------
char* sub_494FF0() {
	int v0;            // eax
	unsigned char* v1; // ecx

	v0 = 0;
	v1 = getMemAt(0x5D4594, 1200928);
	while (*(uint32_t*)v1) {
		v1 += 16;
		++v0;
		if ((int)v1 >= (int)getMemAt(0x5D4594, 1201440)) {
			return 0;
		}
	}
	return (char*)getMemAt(0x5D4594, 1200916 + 16 * v0);
}

//----- (00495020) --------------------------------------------------------
char* sub_495020(int a1) {
	int v1;            // eax
	unsigned char* v2; // ecx

	v1 = 0;
	v2 = getMemAt(0x5D4594, 1200916);
	while (!*((uint32_t*)v2 + 3) || *(uint32_t*)v2 != a1) {
		v2 += 16;
		++v1;
		if ((int)v2 >= (int)getMemAt(0x5D4594, 1201428)) {
			return 0;
		}
	}
	return (char*)getMemAt(0x5D4594, 1200916 + 16 * v1);
}

//----- (00495060) --------------------------------------------------------
int sub_495060(int a1, short a2, short a3) {
	unsigned char* v3; // eax
	char* v4;          // eax

	v3 = getMemAt(0x5D4594, 1200916);
	do {
		if (*((uint32_t*)v3 + 3) == 1 && *(uint32_t*)v3 == a1) {
			return 1;
		}
		v3 += 16;
	} while ((int)v3 < (int)getMemAt(0x5D4594, 1201428));
	v4 = sub_494FF0();
	if (v4) {
		*((uint16_t*)v4 + 3) = a2;
		*((uint16_t*)v4 + 4) = a3;
		*((uint32_t*)v4 + 3) = 1;
		v4[4] = 0;
		*(uint32_t*)v4 = a1;
		return 1;
	}
	return 0;
}

//----- (004950C0) --------------------------------------------------------
int sub_4950C0(int a1) {
	char* v1; // eax

	v1 = sub_495020(a1);
	if (!v1) {
		return 0;
	}
	*((uint32_t*)v1 + 3) = 0;
	return 1;
}

//----- (004950F0) --------------------------------------------------------
int sub_4950F0(int a1, char a2) {
	char* v2; // eax

	v2 = sub_495020(a1);
	if (!v2) {
		return 0;
	}
	v2[4] = a2;
	return 1;
}

//----- (00495120) --------------------------------------------------------
int sub_495120(int a1, short a2, short a3) {
	char* v3; // eax

	v3 = sub_495020(a1);
	if (!v3) {
		return 0;
	}
	*((uint16_t*)v3 + 3) = a2;
	*((uint16_t*)v3 + 4) = a3;
	return 1;
}

//----- (00495150) --------------------------------------------------------
int sub_495150(int a1, short a2) {
	char* v2; // eax

	v2 = sub_495020(a1);
	if (!v2) {
		return 0;
	}
	*((uint16_t*)v2 + 3) = a2;
	return 1;
}

//----- (00495180) --------------------------------------------------------
int sub_495180(int a1, uint16_t* a2, uint16_t* a3, uint8_t* a4) {
	char* v4; // eax

	v4 = sub_495020(a1);
	if (!v4) {
		return 0;
	}
	*a2 = *((uint16_t*)v4 + 3);
	*a3 = *((uint16_t*)v4 + 4);
	*a4 = v4[4];
	return 1;
}

//----- (004951C0) --------------------------------------------------------
char* sub_4951C0() {
	char* result; // eax

	result = (char*)getMemAt(0x5D4594, 1200922);
	do {
		*(uint32_t*)(result + 6) = 0;
		*(uint16_t*)result = 0;
		*((uint16_t*)result + 1) = 0;
		*(result - 2) = 0;
		*(uint32_t*)(result - 6) = 0;
		result += 16;
	} while ((int)result < (int)getMemAt(0x5D4594, 1201434));
	return result;
}

//----- (004951F0) --------------------------------------------------------
int nox_xxx_unitSpriteCheckAlly_4951F0(int a1) { return sub_495020(a1) != 0; }

//----- (00495210) --------------------------------------------------------
static uint16_t nox_deathmsg_read_u16(const uint8_t* data) {
	return (uint16_t)data[0] | (uint16_t)data[1] << 8;
}

static void nox_deathmsg_write_u16(uint8_t* data, uint16_t value) {
	data[0] = value;
	data[1] = value >> 8;
}

int sub_495210(uint8_t* data) {
	int v1; // esi

	v1 = dword_5d4594_1203836;
	if ((dword_5d4594_1203836 + 1) % 100 == dword_5d4594_1203840) {
		dword_5d4594_1203840 = (dword_5d4594_1203840 + 1) % 100;
	}
	if (data[10] == 1) {
		uint16_t value = nox_deathmsg_read_u16(data + 8);
		if (value == *getMemU32Ptr(0x5D4594, 1203844)) {
			nox_deathmsg_write_u16(data + 8, *getMemU16Ptr(0x5D4594, 1203856));
			v1 = dword_5d4594_1203836;
			goto LABEL_9;
		}
		if (value == *getMemU32Ptr(0x5D4594, 1203848)) {
			nox_deathmsg_write_u16(data + 8, *getMemU16Ptr(0x5D4594, 1203852));
			v1 = dword_5d4594_1203836;
			goto LABEL_9;
		}
	}
LABEL_9:
	int offset = 24 * v1;
	*getMemU32Ptr(0x5D4594, 1201428 + offset) = nox_deathmsg_read_u16(data + 2);
	*getMemU32Ptr(0x5D4594, 1201432 + offset) = nox_deathmsg_read_u16(data + 4);
	*getMemU32Ptr(0x5D4594, 1201436 + offset) = nox_deathmsg_read_u16(data + 6);
	*getMemU32Ptr(0x5D4594, 1201440 + offset) = nox_deathmsg_read_u16(data + 8);
	*getMemU32Ptr(0x5D4594, 1201444 + offset) = data[10];
	*getMemU32Ptr(0x5D4594, 1201448 + offset) = gameFrame();
	dword_5d4594_1203836 = (v1 + 1) % 100;
	uint16_t words[6];
	for (int i = 0; i < 6; i++) {
		words[i] = nox_deathmsg_read_u16(data + 2 * i);
	}
	return sub_4952E0(words);
}

//----- (00495430) --------------------------------------------------------
int sub_495430() {
	int v0;       // ecx
	int v1;       // esi
	int result;   // eax
	int i;        // edi
	long long v4; // rtt

	dword_5d4594_1203832 = 0;
	nox_client_drawEnableAlpha_434560(0);
	sub_4345F0(0);
	nox_xxx_draw_434600(0);
	v0 = dword_5d4594_1203840;
	v1 = dword_5d4594_1203840;
	result = dword_5d4594_1203836;
	for (i = nox_win_height / 4 / 36; v1 != dword_5d4594_1203836; v1 = (v1 + 1) % 100) {
		if (dword_5d4594_1203832 > i) {
			break;
		}
		if ((unsigned int)(gameFrame() - *getMemU32Ptr(0x5D4594, 1201448 + 24 * v0)) <= 0x5A) {
			sub_495500(getMemIntPtr(0x5D4594, 1201428 + 24 * v1));
			v0 = dword_5d4594_1203840;
			++dword_5d4594_1203832;
		} else {
			v4 = v0 + 1;
			v0 = (v0 + 1) % 100;
			dword_5d4594_1203840 = v4 % 100;
		}
		result = dword_5d4594_1203836;
	}
	return result;
}

//----- (00495500) --------------------------------------------------------
int* sub_495500(int* a1) {
	int v1;          // edi
	int v2;          // eax
	int v3;          // ebp
	bool v4;         // zf
	char* v5;        // eax
	char* v6;        // eax
	char* v7;        // eax
	int v8;          // ebp
	int v11;         // esi
	int v12;         // esi
	int v13;         // ecx
	int v14;         // kr04_4
	int v15;         // esi
	int v16;         // edi
	int v17;         // ebx
	int v18;         // esi
	int* result;     // eax
	int v20;         // esi
	int v21;         // [esp+10h] [ebp-DCh]
	int v22;         // [esp+14h] [ebp-D8h]
	int v23;         // [esp+18h] [ebp-D4h]
	int v24;         // [esp+1Ch] [ebp-D0h]
	int v25;         // [esp+20h] [ebp-CCh]
	int v26;         // [esp+24h] [ebp-C8h]
	int v27;         // [esp+28h] [ebp-C4h]
	wchar2_t v28[32]; // [esp+2Ch] [ebp-C0h]
	wchar2_t v29[32]; // [esp+6Ch] [ebp-80h]
	wchar2_t v30[32]; // [esp+ACh] [ebp-40h]
	nox_thing* t9 = 0;

	v1 = nox_xxx_guiFontPtrByName_43F360("large");
	v2 = *a1;
	v3 = 0;
	v4 = *a1 == 0;
	v24 = v1;
	v29[0] = 0;
	v28[0] = 0;
	v30[0] = 0;
	v23 = 0;
	v25 = 0;
	v27 = 0;
	v26 = 0;
	if (!v4) {
		v5 = nox_common_playerInfoGetByID_417040(v2);
		if (v5) {
			v3 = 1;
			nox_swprintf(v29, (const wchar2_t*)v5 + 2352);
		}
	}
	if (a1[1]) {
		v6 = nox_common_playerInfoGetByID_417040(a1[1]);
		if (v6) {
			if (v3) {
				nox_swprintf(v28, L"+%s", v6 + 4704);
			} else {
				nox_swprintf(v28, (const wchar2_t*)v6 + 2352);
			}
		}
	}
	if (a1[2]) {
		v7 = nox_common_playerInfoGetByID_417040(a1[2]);
		if (v7) {
			nox_swprintf(v30, (const wchar2_t*)v7 + 2352);
		}
	}
	if (a1[4] != 1) {
		if (a1[4] == 2) {
			switch (a1[3]) {
			case 1:
			case 12:
				v8 = nox_xxx_spellIcon_424A90(5);
				goto LABEL_26;
			case 2:
				v8 = nox_xxx_spellGetAbilityIcon_425310(1, 0);
				goto LABEL_26;
			case 4:
				v8 = nox_xxx_spellIcon_424A90(130);
				goto LABEL_26;
			case 5:
				v8 = nox_xxx_spellIcon_424A90(60);
				goto LABEL_26;
			case 9:
			case 17:
				v8 = nox_xxx_spellIcon_424A90(43);
				goto LABEL_26;
			case 15:
				v8 = nox_xxx_spellIcon_424A90(56);
				goto LABEL_26;
			case 16:
				v8 = nox_xxx_spellIcon_424A90(16);
				goto LABEL_26;
			default:
				break;
			}
		}
		v8 = *getMemU32Ptr(0x5D4594, 1203828);
		goto LABEL_28;
	}
	t9 = nox_get_thing(a1[3]);
	if (!t9) {
		v8 = *getMemU32Ptr(0x5D4594, 1203828);
		goto LABEL_28;
	}
	if (t9->pri_class & 0x1001000) {
		sub_4B9650(a1[3]);
	}
	v8 = t9->pretty_image;
LABEL_26:
	if (!v8) {
		v8 = *getMemU32Ptr(0x5D4594, 1203828);
	}
LABEL_28:
	nox_draw_imageMeta_47D5C0(v8, &v27, &v26, &v23, &v25);
	nox_xxx_drawGetStringSize_43F840(v1, v29, &v21, &v22, 0);
	v11 = v21;
	nox_xxx_drawGetStringSize_43F840(v1, v28, &v21, &v22, 0);
	v12 = v21 + v11;
	nox_xxx_drawGetStringSize_43F840(v1, v30, &v21, &v22, 0);
	v13 = v12 + v23 + v21 + 20;
	v14 = nox_win_width - (v12 + v23 + v21 + 10);
	v15 = v14 / 2;
	v16 = 36 * dword_5d4594_1203832;
	nox_client_drawRectFilledAlpha_49CF10(v14 / 2 - 5, 36 * dword_5d4594_1203832, v13, 36);
	v17 = v16 + (36 - v22) / 2;
	if (*a1) {
		nox_xxx_drawSetTextColor_434390(*getMemIntPtr(0x5D4594, 2597996));
		if (*a1 == nox_player_netCode_85319C) {
			nox_xxx_drawSetTextColor_434390(dword_8531A0_2572);
		}
		v15 = nox_xxx_drawString_43F6E0(v24, (short*)v29, v14 / 2, v17);
	}
	if (a1[1]) {
		nox_xxx_drawSetTextColor_434390(*getMemIntPtr(0x5D4594, 2597996));
		if (a1[1] == nox_player_netCode_85319C) {
			nox_xxx_drawSetTextColor_434390(dword_8531A0_2572);
		}
		v15 = nox_xxx_drawString_43F6E0(v24, (short*)v28, v15, v17);
	}
	v18 = v15 + 5;
	if (v8) {
		nox_client_drawImageAt_47D2C0(v8, v18 - v27, v16 + (36 - v25) / 2 - v26);
	}
	result = a1;
	v20 = v18 + v23 + 5;
	if (a1[2]) {
		nox_xxx_drawSetTextColor_434390(*getMemIntPtr(0x5D4594, 2597996));
		if (a1[2] == nox_player_netCode_85319C) {
			nox_xxx_drawSetTextColor_434390(dword_8531A0_2572);
		}
		result = (int*)nox_xxx_drawString_43F6E0(v24, (short*)v30, v20, v17);
	}
	return result;
}

//----- (004958F0) --------------------------------------------------------
int sub_4958F0() {
	int result; // eax

	dword_5d4594_1203840 = 0;
	dword_5d4594_1203836 = 0;
	if (!*getMemU32Ptr(0x5D4594, 1203844)) {
		*getMemU32Ptr(0x5D4594, 1203844) = nox_xxx_getTTByNameSpriteMB_44CFC0("ArcherBolt");
	}
	if (!*getMemU32Ptr(0x5D4594, 1203848)) {
		*getMemU32Ptr(0x5D4594, 1203848) = nox_xxx_getTTByNameSpriteMB_44CFC0("ArcherArrow");
	}
	if (!*getMemU32Ptr(0x5D4594, 1203852)) {
		*getMemU32Ptr(0x5D4594, 1203852) = nox_xxx_getTTByNameSpriteMB_44CFC0("Bow");
	}
	if (!*getMemU32Ptr(0x5D4594, 1203856)) {
		*getMemU32Ptr(0x5D4594, 1203856) = nox_xxx_getTTByNameSpriteMB_44CFC0("CrossBow");
	}
	result = nox_xxx_spellIcon_424A90(15);
	*getMemU32Ptr(0x5D4594, 1203828) = result;
	return result;
}

// FriendListClass is an eight-byte {net_code,next} record in GAME.EXE. Keep
// that exact layout on 32-bit targets, but widen the link on 64-bit targets so
// allocator addresses never pass through a PE32 dword.
typedef struct nox_client_friend_node {
	uint32_t net_code;
	struct nox_client_friend_node* next;
} nox_client_friend_node;

_Static_assert(offsetof(nox_client_friend_node, net_code) == 0, "wrong friend net-code offset");
_Static_assert(offsetof(nox_client_friend_node, next) == (sizeof(void*) == 4 ? 4 : 8),
			   "wrong native friend link offset");
_Static_assert(sizeof(nox_client_friend_node) == (sizeof(void*) == 4 ? 8 : 16),
			   "wrong native friend node size");

//----- (00495980) --------------------------------------------------------
int nox_xxx_allocClassListFriends_495980() {
	nox_alloc_friendList_1203860 = nox_new_alloc_class("FriendListClass", sizeof(nox_client_friend_node), 128);
	return nox_alloc_friendList_1203860 != 0;
}

//----- (004959B0) --------------------------------------------------------
void sub_4959B0() {
	nox_alloc_class_free_all(nox_alloc_friendList_1203860);
	dword_5d4594_1203864 = 0;
}

//----- (004959D0) --------------------------------------------------------
int sub_4959D0() {
	int result; // eax

	nox_free_alloc_class(nox_alloc_friendList_1203860);
	result = 0;
	nox_alloc_friendList_1203860 = 0;
	dword_5d4594_1203864 = 0;
	return result;
}

//----- (004959F0) --------------------------------------------------------
uint32_t* nox_xxx_cliAddObjFriend_4959F0(int a1) {
	nox_client_friend_node* result; // eax

	result = nox_alloc_class_new_obj_zero(nox_alloc_friendList_1203860);
	if (result) {
		result->net_code = a1;
		result->next = dword_5d4594_1203864;
		dword_5d4594_1203864 = result;
	}
	return (uint32_t*)result;
}

//----- (00495A20) --------------------------------------------------------
void sub_495A20(int a1) {
	nox_client_friend_node* v1; // eax
	nox_client_friend_node* v2; // ecx

	v1 = dword_5d4594_1203864;
	v2 = NULL;
	if (v1) {
		while (v1->net_code != (uint32_t)a1) {
			v2 = v1;
			v1 = v1->next;
			if (!v1) {
				return;
			}
		}
		if (v2) {
			v2->next = v1->next;
		} else {
			dword_5d4594_1203864 = v1->next;
		}
		nox_alloc_class_free_obj_first(nox_alloc_friendList_1203860, v1);
	}
}

//----- (00495A80) --------------------------------------------------------
int sub_495A80(int a1) {
	nox_client_friend_node* v1; // eax

	v1 = dword_5d4594_1203864;
	if (!v1) {
		return 0;
	}
	while (v1->net_code != (uint32_t)a1) {
		v1 = v1->next;
		if (!v1) {
			return 0;
		}
	}
	return 1;
}

//----- (00495B50) --------------------------------------------------------
void sub_495B50(nox_drawable_fx* fx) {
	if (!fx) {
		return;
	}
	if (fx->global_next) {
		fx->global_next->global_prev = fx->global_prev;
	}
	if (fx->global_prev) {
		fx->global_prev->global_next = fx->global_next;
	} else {
		setMemPtr(0x5D4594, 1203872, fx->global_next);
	}
	if (fx->next) {
		fx->next->prev = fx->prev;
	}
	if (fx->prev) {
		fx->prev->next = fx->next;
	} else if (fx->owner) {
		fx->owner->field_114 = fx->next;
	}
	fx->owner = NULL;
	fx->next = NULL;
	fx->prev = NULL;
	fx->global_next = NULL;
	fx->global_prev = NULL;
}

//----- (00495BB0) --------------------------------------------------------
void sub_495BB0(nox_drawable* dr, nox_draw_viewport_t* vp) {
	if (!dr || !vp) {
		return;
	}
	for (nox_drawable_fx* fx = dr->field_114; fx;) {
		nox_drawable_fx* next = fx->next;
		if (fx->field_0 == 1) {
			sub_495BF0(dr, fx, vp);
		} else if (fx->field_0 == 2) {
			sub_495D00(dr, fx, vp);
		}
		fx = next;
	}
}

static bool nox_drawable_fx_path_collapsed(const nox_drawable_fx* fx) {
	for (int i = 0; i < fx->count; ++i) {
		// GAME.EXE treats an axis-aligned pair as collapsed. Only a pair
		// whose X and Y both changed keeps the trail alive while idle.
		if (fx->trail[i][0] != fx->trail[i + 1][0] &&
			fx->trail[i][1] != fx->trail[i + 1][1]) {
			return false;
		}
	}
	return true;
}

static bool nox_drawable_fx_owner_moved(const nox_drawable* dr) {
	return (int32_t)dr->pos.x != (int32_t)dr->field_8 ||
		(int32_t)dr->pos.y != (int32_t)dr->field_9;
}

//----- (00495BF0) --------------------------------------------------------
int sub_495BF0(nox_drawable* dr, nox_drawable_fx* fx, nox_draw_viewport_t* vp) {
	if (!dr || !fx || !vp) {
		return 0;
	}
	const int count = fx->count;
	if ((count == 0 || nox_drawable_fx_path_collapsed(fx)) &&
		!nox_drawable_fx_owner_moved(dr)) {
		fx->count = 0;
		return count;
	}

	const intptr_t saved_x = dr->pos.x;
	const intptr_t saved_y = dr->pos.y;
	uint8_t alpha = UINT8_MAX;
	for (int i = 0; i < count; ++i) {
		alpha -= 42;
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(alpha);
		dr->pos.x = fx->trail[i][0];
		dr->pos.y = fx->trail[i][1];
		dr->draw_func((uint32_t*)vp, dr);
	}

	for (int i = count; i > 0; --i) {
		fx->trail[i][0] = fx->trail[i - 1][0];
		fx->trail[i][1] = fx->trail[i - 1][1];
	}
	dr->pos.x = saved_x;
	dr->pos.y = saved_y;
	fx->trail[0][0] = (int32_t)saved_x;
	fx->trail[0][1] = (int32_t)saved_y;
	if (fx->count != 5) {
		++fx->count;
	}
	nox_client_drawEnableAlpha_434560(0);
	return 1;
}

//----- (00495D00) --------------------------------------------------------
int sub_495D00(nox_drawable* dr, nox_drawable_fx* fx, nox_draw_viewport_t* vp) {
	if (!dr || !fx || !vp) {
		return 0;
	}
	const int count = fx->count;
	if ((count == 0 || nox_drawable_fx_path_collapsed(fx)) &&
		!nox_drawable_fx_owner_moved(dr)) {
		fx->count = 0;
		return count;
	}

	const float velocity_x = *getMemFloatPtr(0x587000, 194136 + 64 * dr->field_77) * -12.0f;
	const float velocity_y = *getMemFloatPtr(0x587000, 194140 + 64 * dr->field_77) * -12.0f;
	int previous_x = (int)(dr->pos.x + vp->x1 - vp->field_4) + nox_float2int(velocity_x);
	int previous_y = (int)(dr->pos.y - (int16_t)dr->z - (int16_t)dr->field_26_1 -
		vp->field_5 + vp->y1 - 10) + nox_float2int(velocity_y);
	uint8_t alpha = UINT8_MAX;
	for (int i = 0; i < count; ++i) {
		alpha -= 63;
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(alpha);
		const int current_x = (int)(fx->trail[i][0] + vp->x1 - vp->field_4) +
			nox_float2int(velocity_x);
		const int current_y = (int)(fx->trail[i][1] - (int16_t)dr->z -
			(int16_t)dr->field_26_1 - vp->field_5 + vp->y1 - 10) +
			nox_float2int(velocity_y);
		nox_client_drawSetColor_434460(nox_color_rgb_4344A0(200, 200, 200));
		nox_client_drawAddPoint_49F500(previous_x, previous_y);
		nox_client_drawAddPoint_49F500(current_x, current_y);
		nox_client_drawLineFromPoints_49E4B0();
		previous_x = current_x;
		previous_y = current_y;
	}

	for (int i = count; i > 0; --i) {
		fx->trail[i][0] = fx->trail[i - 1][0];
		fx->trail[i][1] = fx->trail[i - 1][1];
	}
	fx->trail[0][0] = (int32_t)dr->pos.x;
	fx->trail[0][1] = (int32_t)dr->pos.y;
	if (fx->count != 5) {
		++fx->count;
	}
	nox_client_drawEnableAlpha_434560(0);
	return 1;
}

//----- (00495FC0) --------------------------------------------------------
void sub_495FC0(nox_drawable_fx* fx, nox_drawable* dr) {
	if (!fx || !dr) {
		return;
	}
	fx->owner = dr;
	fx->global_next = getMemPtr(0x5D4594, 1203872);
	fx->global_prev = NULL;
	if (fx->global_next) {
		fx->global_next->global_prev = fx;
	}
	setMemPtr(0x5D4594, 1203872, fx);
	fx->next = dr->field_114;
	fx->prev = NULL;
	if (fx->next) {
		fx->next->prev = fx;
	}
	dr->field_114 = fx;
}

//----- (00499360) --------------------------------------------------------
int nox_xxx_loadReflSheild_499360() {
	static const char* const names[9] = {
		"ReflectiveShieldNW", "ReflectiveShieldN", "ReflectiveShieldNE",
		"ReflectiveShieldW", NULL, "ReflectiveShieldE",
		"ReflectiveShieldSW", "ReflectiveShieldS", "ReflectiveShieldSE",
	};

	for (int i = 0; i < 9; ++i) {
		nox_drawable* dr = NULL;
		if (names[i]) {
			int type = nox_xxx_getTTByNameSpriteMB_44CFC0(names[i]);
			dr = nox_new_drawable_for_thing(type);
			if (!dr) {
				return 0;
			}
			dr->flags30 |= 0x1000000u;
		}
		setMemPtr(0x5D4594, 1217468 + 4 * i, dr);
	}
	*getMemU32Ptr(0x5D4594, 1217504) = 0;
	return 1;
}

//----- (00499450) --------------------------------------------------------
int sub_499450() {
	int result = 0; // eax

	for (int i = 0; i < 9; ++i) {
		nox_drawable* dr = getMemPtr(0x5D4594, 1217468 + 4 * i);
		if (dr) {
			result = nox_xxx_spriteDelete_45A4B0(dr);
		}
		setMemPtr(0x5D4594, 1217468 + 4 * i, NULL);
	}
	*getMemU32Ptr(0x5D4594, 1217504) = 0;
	return result;
}

//----- (00499810) --------------------------------------------------------
int nox_xxx_drawShield_499810(nox_draw_viewport_t* vp, nox_drawable* dr) {
	unsigned int dir = dr->field_74_2;
	nox_drawable* shield = getMemPtr(0x5D4594, 1217468 + 4 * dir);
	shield->pos.x = dr->pos.x + *getMemU32Ptr(0x587000, 161776 + 8 * dir);
	shield->pos.y = dr->pos.y + dr->z + *getMemU32Ptr(0x587000, 161780 + 8 * dir);
	shield->draw_func((uint32_t*)vp, shield);
	return 0;
}

//----- (00499880) --------------------------------------------------------
uint32_t* nox_xxx_fxDrawTurnUndead_499880(short* a1) {
	int i;            // ebx
	uint32_t* result; // eax
	uint32_t* v3;     // esi
	int v4;           // eax
	double v5;        // st7

	if (!*getMemU32Ptr(0x5D4594, 1217508)) {
		*getMemU32Ptr(0x5D4594, 1217508) = nox_xxx_getTTByNameSpriteMB_44CFC0("UndeadKiller");
	}
	for (i = 0; i < 256; i += 6) {
		result = (uint32_t*)nox_xxx_spriteLoadAdd_45A360_drawable(*getMemIntPtr(0x5D4594, 1217508), *a1, a1[1]);
		v3 = result;
		if (result) {
			v4 = 8 * (short)i;
			*((uint16_t*)v3 + 254) = i;
			*((float*)v3 + 117) = *getMemFloatPtr(0x587000, 194136 + v4) * 4.0;
			v5 = *getMemFloatPtr(0x587000, 194140 + v4) * 4.0;
			v3[119] = 0;
			*((float*)v3 + 118) = v5;
			v3[79] = gameFrame();
			v3[81] = *a1;
			v3[82] = a1[1];
			v3[115] = nox_xxx_sprite_4CA540;
			nox_xxx_spriteToList_49BC80_drawable(v3);
			nox_xxx_spriteToSightDestroyList_49BAB0_drawable(v3);
		}
	}
	return result;
}

//----- (00499CF0) --------------------------------------------------------
void nox_xxx_bookRewardCli_499CF0(int* a1, int a2, int a3) {
	unsigned int result; // eax
	int v4;              // esi
	int2 a3a;            // [esp+8h] [ebp-8h]

	if (!nox_common_gameFlags_check_40A5C0(2048) ||
		(result = nox_xxx_bookGet_430B40_get_mouse_prev_seq() - *getMemU32Ptr(0x5D4594, 1217504), result >= 2)) {
		*getMemU32Ptr(0x5D4594, 1217504) = nox_xxx_bookGet_430B40_get_mouse_prev_seq();
		if (a1 == (int*)2) {
			v4 = 0;
		} else {
			v4 = (a1 == (int*)3) + 2;
		}
		a3a.field_0 = 5;
		a3a.field_4 = nox_win_height / 3;
		nox_xxx_bookSetForward_45D200(a1, a2, &a3a);
		nox_xxx_draw_499E70(v4, a3a.field_0, a3a.field_4, 271, 166, 1, 1);
		nox_xxx_draw_499E70(v4, a3a.field_0, a3a.field_4, 135, 166, 2, 1);
		nox_xxx_draw_499E70(v4, a3a.field_0, a3a.field_4 + 166, 135, 166, 2, 1);
		nox_xxx_draw_499E70(v4, a3a.field_0 + 271, a3a.field_4, 271, 166, 1, 2);
		nox_xxx_draw_499E70(v4, a3a.field_0 + 135, a3a.field_4, 135, 166, 2, 2);
		nox_xxx_draw_499E70(v4, a3a.field_0 + 135, a3a.field_4 + 166, 135, 166, 2, 2);
		if (a1 != (int*)4 && a3 == 1) {
			nox_xxx_bookFillAll_45D570((int)a1, a2);
		}
	}
}

//----- (00499F60) --------------------------------------------------------
void sub_499F60(int a1, int a2, int a3, short a4, char a5, char a6, char a7, char a8, char a9, int a10) {
	uint32_t* result; // eax
	int v11;          // edx
	int v12;          // ecx
	uint32_t* v13;    // esi
	int v14;          // eax

	if (!*getMemU32Ptr(0x5D4594, 1217512)) {
		*getMemU32Ptr(0x5D4594, 1217512) = nox_xxx_getTTByNameSpriteMB_44CFC0("RedBubbleParticle");
		*getMemU32Ptr(0x5D4594, 1217516) = nox_xxx_getTTByNameSpriteMB_44CFC0("WhiteBubbleParticle");
		*getMemU32Ptr(0x5D4594, 1217520) = nox_xxx_getTTByNameSpriteMB_44CFC0("LightBlueBubbleParticle");
		*getMemU32Ptr(0x5D4594, 1217524) = nox_xxx_getTTByNameSpriteMB_44CFC0("OrangeBubbleParticle");
		*getMemU32Ptr(0x5D4594, 1217528) = nox_xxx_getTTByNameSpriteMB_44CFC0("GreenBubbleParticle");
		*getMemU32Ptr(0x5D4594, 1217532) = nox_xxx_getTTByNameSpriteMB_44CFC0("VioletBubbleParticle");
		*getMemU32Ptr(0x5D4594, 1217536) = nox_xxx_getTTByNameSpriteMB_44CFC0("LightVioletBubbleParticle");
		*getMemU32Ptr(0x5D4594, 1217540) = nox_xxx_getTTByNameSpriteMB_44CFC0("YellowBubbleParticle");
	}
	result = (uint32_t*)nox_xxx_spriteLoadAdd_45A360_drawable(a1, a2, a3);
	v13 = result;
	if (result) {
		BYTE1(v11) = HIBYTE(a4);
		LOBYTE(result) = *((uint8_t*)result + 160);
		LOBYTE(v12) = *((uint8_t*)v13 + 156);
		*((uint16_t*)v13 + 52) = a4;
		LOBYTE(v11) = *((uint8_t*)v13 + 152);
		v13[108] = nox_color_rgb_4344A0(v11, v12, (int)result);
		if (a1 == *getMemU32Ptr(0x5D4594, 1217512)) {
			v14 = nox_color_rgb_4344A0(255, 128, 128);
		} else if (a1 == *getMemU32Ptr(0x5D4594, 1217516)) {
			v14 = nox_color_rgb_4344A0(255, 255, 255);
		} else if (a1 == *getMemU32Ptr(0x5D4594, 1217524)) {
			v14 = nox_color_rgb_4344A0(255, 100, 50);
		} else if (a1 == *getMemU32Ptr(0x5D4594, 1217528)) {
			v14 = nox_color_rgb_4344A0(64, 255, 64);
		} else if (a1 == *getMemU32Ptr(0x5D4594, 1217532)) {
			v14 = nox_color_rgb_4344A0(255, 100, 255);
		} else if (a1 == *getMemU32Ptr(0x5D4594, 1217536)) {
			v14 = nox_color_rgb_4344A0(255, 200, 255);
		} else if (a1 == *getMemU32Ptr(0x5D4594, 1217540)) {
			v14 = nox_color_rgb_4344A0(255, 255, 200);
		} else {
			v14 = nox_color_rgb_4344A0(200, 200, 255);
		}
		v13[109] = v14;
		*((uint8_t*)v13 + 440) = a5;
		*((uint8_t*)v13 + 443) = a6;
		*((uint8_t*)v13 + 442) = a6;
		*((uint8_t*)v13 + 441) = 1;
		*((uint8_t*)v13 + 444) = a8;
		*((uint8_t*)v13 + 445) = a9;
		*((uint8_t*)v13 + 446) = a7;
		nox_xxx_spriteToSightDestroyList_49BAB0_drawable(v13);
		nox_xxx_spriteTransparentDecay_49B950(v13, a10);
		nox_xxx_sprite_45A110_drawable(v13);
	}
}
// 49A025: variable 'v11' is possibly undefined
// 49A025: variable 'v12' is possibly undefined

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

//----- (0049A5F0) --------------------------------------------------------
int nox_xxx_allocArrayHealthChanges_49A5F0() {
	nox_alloc_class* allocator = nox_new_alloc_class("HealthChange", sizeof(nox_health_change), 32);
	nox_alloc_healthChange_1301772 = allocator;
	if (allocator) {
		dword_5d4594_1301780 = nox_xxx_guiFontPtrByName_43F360("numbers");
		return 1;
	}
	return 0;
}

//----- (0049A630) --------------------------------------------------------
void sub_49A630() {
	nox_alloc_class_free_all((nox_alloc_class*)nox_alloc_healthChange_1301772);
	dword_5d4594_1301776 = 0;
}

//----- (0049A650) --------------------------------------------------------
nox_health_change* nox_xxx_cliAddHealthChange_49A650(int drawable_id, short delta) {
	nox_health_change* change =
		(nox_health_change*)nox_alloc_class_new_obj_zero((nox_alloc_class*)nox_alloc_healthChange_1301772);
	if (!change) {
		return 0;
	}
	change->drawable_id = (uint32_t)drawable_id;
	change->delta = delta;
	change->frame = (uint32_t)nox_xxx_bookGet_430B40_get_mouse_prev_seq();
	change->next = (nox_health_change*)dword_5d4594_1301776;
	change->prev = 0;
	if (change->next) {
		change->next->prev = change;
	}
	dword_5d4594_1301776 = change;
	return change->next;
}

//----- (0049A6A0) --------------------------------------------------------
void sub_49A6A0(nox_draw_viewport_t* vp, nox_drawable* dr) {
	uint32_t frame = (uint32_t)nox_xxx_bookGet_430B40_get_mouse_prev_seq();
	uint32_t damage_color = dr == (nox_drawable*)getMemPtr(0x852978, 8)
		? *getMemU32Ptr(0x85B3FC, 940)
		: nox_color_yellow_2589772;
	for (nox_health_change* change = (nox_health_change*)dword_5d4594_1301776; change;) {
		nox_health_change* next = change->next;
		if ((uint32_t)(frame - change->frame) > 30) {
			sub_49A880(change);
		} else if (change->drawable_id == dr->field_32) {
			int x = (int)(vp->x1 + dr->pos.x - vp->field_4);
			int frame_delta = (int32_t)(change->frame - frame);
			int y = (int)(dr->pos.y + vp->y1 + 2 * frame_delta - (int16_t)dr->z -
						  (int)dr->field_25 - vp->field_5);
			wchar2_t text[80];
			int width;
			nox_swprintf(text, L"%d", abs((int)change->delta));
			nox_xxx_drawGetStringSize_43F840(dword_5d4594_1301780, text, &width, 0, 0);
			x += width / -2;
			nox_xxx_drawSetTextColor_434390(nox_color_black_2650656);
			nox_xxx_drawString_43F6E0(dword_5d4594_1301780, text, x - 1, y - 1);
			nox_xxx_drawString_43F6E0(dword_5d4594_1301780, text, x - 1, y + 1);
			nox_xxx_drawString_43F6E0(dword_5d4594_1301780, text, x + 1, y - 1);
			nox_xxx_drawString_43F6E0(dword_5d4594_1301780, text, x + 1, y + 1);
			if (change->delta <= 0) {
				nox_xxx_drawSetTextColor_434390(damage_color);
			} else {
				nox_xxx_drawSetTextColor_434390(dword_8531A0_2572);
			}
			nox_xxx_drawString_43F6E0(dword_5d4594_1301780, text, x, y);
		}
		change = next;
	}
}

//----- (0049A880) --------------------------------------------------------
void sub_49A880(nox_health_change* change) {
	if (change->prev) {
		change->prev->next = change->next;
	} else {
		dword_5d4594_1301776 = change->next;
	}
	if (change->next) {
		change->next->prev = change->prev;
	}
	nox_alloc_class_free_obj_first((nox_alloc_class*)nox_alloc_healthChange_1301772, change);
}

//----- (0049A8C0) --------------------------------------------------------
int sub_49A8C0() {
	nox_free_alloc_class((nox_alloc_class*)nox_alloc_healthChange_1301772);
	nox_alloc_healthChange_1301772 = 0;
	dword_5d4594_1301776 = 0;
	dword_5d4594_1301780 = 0;
	return 0;
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

//----- (0049B3E0) --------------------------------------------------------
int sub_49B3E0() {
	dword_5d4594_1303452 = nox_new_window_from_file("GGOver.wnd", sub_49B420);
	if (dword_5d4594_1303452) {
		nox_window_set_hidden(dword_5d4594_1303452, 1);
		nox_xxx_wnd_46ABB0(dword_5d4594_1303452, 0);
		return 1;
	}
	return 0;
}

//----- (0049B420) --------------------------------------------------------
int sub_49B420(int a1, int a2, int* a3, int a4) {
	int v3; // esi

	if (a2 == 16391) {
		v3 = nox_xxx_wndGetID_46B0A0(a3);
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
		if (v3 == 10701) {
			nox_client_quit_4460C0();
			sub_49B6B0();
		} else if (v3 == 10702) {
			LOWORD(a2) = 1008;
			nox_xxx_netClientSend2_4E53C0(31, &a2, 2, 0, 1);
			sub_49B6B0();
			return 0;
		}
	}
	return 0;
}

//----- (0049B490) --------------------------------------------------------
int sub_49B490() {
	int result; // eax

	result = nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1303452);
	dword_5d4594_1303452 = 0;
	return result;
}

//----- (0049B6B0) --------------------------------------------------------
int sub_49B6B0() {
	nox_window_set_hidden(dword_5d4594_1303452, 1);
	nox_xxx_wnd_46ABB0(dword_5d4594_1303452, 0);
	return nox_xxx_windowFocus_46B500(0);
}

//----- (0049B7A0) --------------------------------------------------------
void nox_xxx_consoleEsc_49B7A0() {
	int v0; // esi

	v0 = 0;
	if (!nox_xxx_guiCursor_477600() && !nox_video_inFadeTransition_44E0D0()) {
		if (sub_460660()) {
			v0 = 1;
		}
		if (!sub_46A6A0() && v0 != 1) {
			if (sub_45D9B0() == 1) {
				sub_45D870();
			} else {
				if (sub_4BFE40()) {
					v0 = 1;
				}
				if (nox_xxx_quickBarClose_4606B0()) {
					v0 = 1;
				}
				if (sub_462740()) {
					v0 = 1;
				}
				if (!sub_44A4E0() && v0 != 1) {
					if (sub_479590() == 2) {
						sub_4795A0(1);
					} else if (sub_479590() == 3) {
						sub_4795A0(1);
					} else if (sub_479590() == 4) {
						sub_4795A0(1);
					} else if (!sub_478040() && !sub_479950()) {
						if (sub_467C10()) {
							v0 = 1;
						}
						if (nox_xxx_bookHideMB_45ACA0(0)) {
							v0 = 1;
						}
						if (nox_gui_console_Hide_4512B0()) {
							v0 = 1;
						}
						if (sub_446780()) {
							v0 = 1;
						}
						if (nox_xxx_guiServerOptionsTryHide_4574D0()) {
							v0 = 1;
						}
						if (sub_48CAD0()) {
							v0 = 1;
						}
						if (sub_4AD9B0(1)) {
							v0 = 1;
						}
						if (sub_4C35B0(1)) {
							v0 = 1;
						}
						if (!sub_46D6F0() && v0 != 1) {
							if (nox_xxx_game_4DCCB0()) {
								sub_445C40();
							} else {
								nox_xxx_clientPlaySoundSpecial_452D80(231, 100);
							}
						}
					}
				}
			}
		}
	}
}
// 49B874: variable 'v1' is possibly undefined
// 49B881: variable 'v2' is possibly undefined

//----- (0049BB80) --------------------------------------------------------
void* sub_49BB80(char a1) {
	void* result; // eax

	*getMemU8Ptr(0x5D4594, 1303504) = a1;
	*getMemU8Ptr(0x5D4594, 1303512) = 0;
	*getMemU32Ptr(0x5D4594, 1303516) = gameFrame();
	result = nox_xxx_spellGetDefArrayPtr_424820();
	dword_5d4594_1303508 = (uintptr_t)result;
	return result;
}

//----- (0049BBB0) --------------------------------------------------------
void sub_49BBB0() { *getMemU8Ptr(0x5D4594, 1303504) = 0; }

//----- (0049BBC0) --------------------------------------------------------
void sub_49BBC0() {
	int v0;           // eax
	unsigned char v1; // [esp+0h] [ebp-4h]

	if (getMemByte(0x5D4594, 1303504)) {
		v1 = nox_xxx_spellPhonemes_424A20(getMemByte(0x5D4594, 1303504), getMemByte(0x5D4594, 1303512));
		if (gameFrame() >= *getMemIntPtr(0x5D4594, 1303516)) {
			v0 = nox_xxx_spellGetPhoneme_4FE1C0(nox_player_netCode_85319C, v1);
			nox_xxx_clientPlaySoundSpecial_452D80(v0, 100);
			nox_client_setPhonemeFrame_476E00(*getMemU32Ptr(0x587000, 163576 + 4 * v1));
			*getMemU32Ptr(0x5D4594, 1303516) = gameFrame() + 3;
			dword_5d4594_1303508 =
				(uintptr_t)nox_xxx_updateSpellRelated_424830((void*)dword_5d4594_1303508, v1);
			++*getMemU8Ptr(0x5D4594, 1303512);
		}
		if (*(uint32_t*)dword_5d4594_1303508 == getMemByte(0x5D4594, 1303504)) {
			sub_49BBB0();
		}
	}
}

//----- (0049C160) --------------------------------------------------------
uint32_t* nox_xxx_clientAddRayEffect_49C160(int a1) {
	uint32_t* result;   // eax
	int v2;             // eax
	int v3;             // esi
	int v4;             // eax
	int v5;             // ebx
	uint32_t* v6;       // eax
	uint32_t* v7;       // edi
	int v8;             // esi
	int v9;             // edi
	int v10;            // kr00_4
	int v11;            // ecx
	int v12;            // edx
	int v13;            // edx
	unsigned char* v14; // ecx

	result = *(uint32_t**)getMemAt(0x5D4594, 1304312);
	if (*getMemIntPtr(0x5D4594, 1304312) < 96) {
		if (!*getMemU32Ptr(0x5D4594, 1304352)) {
			*getMemU32Ptr(0x5D4594, 1304352) = nox_xxx_getTTByNameSpriteMB_44CFC0("DynamicLightning");
			*getMemU32Ptr(0x5D4594, 1304356) = nox_xxx_getTTByNameSpriteMB_44CFC0("DynamicChainLightning");
			*getMemU32Ptr(0x5D4594, 1304360) = nox_xxx_getTTByNameSpriteMB_44CFC0("DynamicEnergyBolt");
			*getMemU32Ptr(0x5D4594, 1304364) = nox_xxx_getTTByNameSpriteMB_44CFC0("OrbRay");
			*getMemU32Ptr(0x5D4594, 1304368) = nox_xxx_getTTByNameSpriteMB_44CFC0("PlasmaRay");
			*getMemU32Ptr(0x5D4594, 1304372) = nox_xxx_getTTByNameSpriteMB_44CFC0("DrainManaRay");
			*getMemU32Ptr(0x5D4594, 1304376) = nox_xxx_getTTByNameSpriteMB_44CFC0("HealRay");
			*getMemU32Ptr(0x5D4594, 1304380) = nox_xxx_getTTByNameSpriteMB_44CFC0("CharmRay");
			*getMemU32Ptr(0x5D4594, 1304384) = nox_xxx_getTTByNameSpriteMB_44CFC0("DrainManaOrb");
			*getMemU32Ptr(0x5D4594, 1304388) = nox_xxx_getTTByNameSpriteMB_44CFC0("HealOrb");
			*getMemU32Ptr(0x5D4594, 1304392) = nox_xxx_getTTByNameSpriteMB_44CFC0("CharmOrb");
			*getMemU32Ptr(0x5D4594, 1304396) = nox_xxx_getTTByNameSpriteMB_44CFC0("HarpoonRope");
		}
		v2 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(a1 + 3));
		v3 = v2;
		v4 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(a1 + 5));
		v5 = v4;
		v6 = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(a1 + 3)) ? nox_xxx_netSpriteByCodeStatic_45A720(v3)
																	   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v3);
		v7 = v6;
		result = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(a1 + 5)) ? nox_xxx_netSpriteByCodeStatic_45A720(v5)
																		   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v5);
		if (v7 && result) {
			v8 = v7[3];
			v9 = v7[4];
			v10 = result[4] - v9;
			v11 = v8 + ((int)result[3] - v8) / 2;
			result = (uint32_t*)(v9 + v10 / 2);
			switch (*(unsigned char*)(a1 + 1)) {
			case 1u:
				v12 = *getMemU32Ptr(0x5D4594, 1304368);
				break;
			case 2u:
				v12 = *getMemU32Ptr(0x5D4594, 1304380);
				break;
			case 3u:
				v12 = *getMemU32Ptr(0x5D4594, 1304356);
				break;
			case 4u:
				v12 = *getMemU32Ptr(0x5D4594, 1304360);
				break;
			case 5u:
				v12 = *getMemU32Ptr(0x5D4594, 1304372);
				break;
			case 6u:
				v12 = *getMemU32Ptr(0x5D4594, 1304376);
				break;
			case 7u:
				v12 = *getMemU32Ptr(0x5D4594, 1304396);
				break;
			case 0x8Cu:
				v12 = *getMemU32Ptr(0x5D4594, 1304352);
				break;
			default:
				return result;
			}
			result = (uint32_t*)nox_xxx_spriteLoadAdd_45A360_drawable(v12, v11, v9 + v10 / 2);
			if (!result) {
				return result;
			}
			*((uint8_t*)result + 432) = 1;
			*(uint32_t*)((char*)result + 437) = *(unsigned short*)(a1 + 3);
			*(uint32_t*)((char*)result + 441) = *(unsigned short*)(a1 + 5);
			v13 = 0;
			*(uint32_t*)((char*)result + 433) = *(unsigned char*)(a1 + 2);
			v14 = getMemAt(0x5D4594, 1303924);
			while (*(uint32_t*)v14) {
				v14 += 4;
				++v13;
				if ((int)v14 >= (int)getMemAt(0x5D4594, 1304308)) {
					return result;
				}
			}
			*getMemU32Ptr(0x5D4594, 1303924 + 4 * v13) = result;
		}
	}
	return result;
}
// 49C248: variable 'v2' is possibly undefined
// 49C256: variable 'v4' is possibly undefined

//----- (0049C450) --------------------------------------------------------
void nox_xxx_clientRemoveRayEffect_49C450(int a1) {
	int v1;  // esi
	int* v2; // ecx

	v1 = 0;
	v2 = getMemIntPtr(0x5D4594, 1303924);
	while (1) {
		int result = *v2;
		if (*v2) {
			if (*(unsigned short*)(a1 + 3) == *(uint32_t*)(result + 437) &&
				*(unsigned short*)(a1 + 5) == *(uint32_t*)(result + 441)) {
				break;
			}
		}
		++v2;
		++v1;
		if ((int)v2 >= (int)getMemAt(0x5D4594, 1304308)) {
			return;
		}
	}
	nox_xxx_spriteDeleteStatic_45A4E0_drawable(*v2);
	*getMemU32Ptr(0x5D4594, 1303924 + 4 * v1) = 0;
}

//----- (0049C4B0) --------------------------------------------------------
void nox_xxx_spriteDeleteSomeList_49C4B0() {
	size_t count = nox_client_transient_ray_count();
	for (size_t i = 0; i < count; ++i) {
		nox_drawable* dr = nox_client_transient_ray_at(i);
		if (dr) {
			nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr);
		}
	}
	nox_client_transient_ray_clear();
	sub_4C5050();
}

//----- (0049C4F0) --------------------------------------------------------
void nox_xxx_sprite_49C4F0() {
#if UINTPTR_MAX == UINT32_MAX
	int* v0 = getMemIntPtr(0x5D4594, 1303924);
	do {
		if (*v0) {
			nox_xxx_spriteDeleteStatic_45A4E0_drawable(*v0);
			*v0 = 0;
		}
		++v0;
	} while ((int)v0 < (int)getMemAt(0x5D4594, 1304308));
#else
	// Duration rays are tracked by Client.fxDurationRays on native-width
	// builds. The legacy PE32 array cannot contain their pointers.
	memset(getMemAt(0x5D4594, 1303924), 0, 96 * sizeof(uint32_t));
#endif
}

//----- (0049C520) --------------------------------------------------------
int sub_49C520(nox_drawable* a1p) {
	if (nox_client_transient_ray_contains(a1p)) {
		return 1;
	}
#if UINTPTR_MAX == UINT32_MAX
	uintptr_t a1 = (uintptr_t)a1p;
	unsigned char* v1; // eax
	int v2;            // eax
	unsigned char* i;  // ecx

	v1 = getMemAt(0x5D4594, 1303924);
	while (a1 != *(uint32_t*)v1) {
		v1 += 4;
		if ((int)v1 >= (int)getMemAt(0x5D4594, 1304308)) {
			v2 = 0;
			if (*getMemIntPtr(0x5D4594, 1304308) <= 0) {
				return 0;
			}
			for (i = getMemAt(0x5D4594, 1303540); a1 != *(uint32_t*)i; i += 4) {
				if (++v2 >= *getMemIntPtr(0x5D4594, 1304308)) {
					return 0;
				}
			}
			return 1;
		}
	}
	return 1;
#else
	// The remaining list is the PE32 duration-ray array. Native-width builds
	// use the Go-owned duration-ray registry instead of truncated slots.
	return 0;
#endif
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
int sub_49C810() { return dword_5d4594_1305680 != 0; }

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
int sub_49CB40() { return dword_5d4594_1305684 != 0; }

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

//----- (0049F6D0) --------------------------------------------------------
int sub_49F6D0(int a1) {
	int result; // eax

	result = nox_draw_curDrawData_3799572->flag_0;
	nox_draw_curDrawData_3799572->flag_0 = a1;
	return result;
}

//----- (0049F6F0) --------------------------------------------------------
int4* nox_client_copyRect_49F6F0(int xLeft, int yTop, int a3, int a4) {
	int4* result; // eax
	nox_rect rcSrc;   // [esp+0h] [ebp-20h]
	nox_rect rc;      // [esp+10h] [ebp-10h]

	noxSetRect(&rc, xLeft, yTop, xLeft + a3, yTop + a4);
	result = nox_xxx_utilRect_49F930((int4*)&rcSrc, (int4*)&rc, (int4*)(&nox_draw_curDrawData_3799572->rect3));
	if (result) {
		noxCopyRect(&nox_draw_curDrawData_3799572->clip, &rcSrc);
		--rcSrc.max_x;
		--rcSrc.max_y;
		result = (int4*)noxCopyRect((&nox_draw_curDrawData_3799572->rect2), &rcSrc);
	}
	return result;
}

//----- (0049F780) --------------------------------------------------------
int4* sub_49F780(int xLeft, int a2) {
	int v2; // esi
	int v3; // eax

	v2 = xLeft;
	if (xLeft < (int)nox_draw_curDrawData_3799572->clip.min_x) {
		v2 = nox_draw_curDrawData_3799572->clip.min_x;
	}
	v3 = a2;
	if (a2 > (int)nox_draw_curDrawData_3799572->clip.max_x) {
		v3 = nox_draw_curDrawData_3799572->clip.max_x;
	}
	return nox_client_copyRect_49F6F0(v2, nox_draw_curDrawData_3799572->clip.min_y, v3 - v2,
									  nox_draw_curDrawData_3799572->clip.max_y - nox_draw_curDrawData_3799572->clip.min_y);
}

//----- (0049F7F0) --------------------------------------------------------
void nox_xxx_wndDraw_49F7F0() {
	if (!dword_5d4594_1305748) {
		*getMemU32Ptr(0x5D4594, 1305772) = nox_draw_curDrawData_3799572->flag_0;
		*getMemU32Ptr(0x5D4594, 1305756) = nox_draw_curDrawData_3799572->clip.min_x;
		*getMemU32Ptr(0x5D4594, 1305760) = nox_draw_curDrawData_3799572->clip.min_y;
		*getMemU32Ptr(0x5D4594, 1305764) = nox_draw_curDrawData_3799572->clip.max_x;
		*getMemU32Ptr(0x5D4594, 1305768) = nox_draw_curDrawData_3799572->clip.max_y;
		*getMemU32Ptr(0x5D4594, 1305732) = nox_draw_curDrawData_3799572->rect2.min_x;
		*getMemU32Ptr(0x5D4594, 1305736) = nox_draw_curDrawData_3799572->rect2.min_y;
		*getMemU32Ptr(0x5D4594, 1305740) = nox_draw_curDrawData_3799572->rect2.max_x;
		*getMemU32Ptr(0x5D4594, 1305744) = nox_draw_curDrawData_3799572->rect2.max_y;
		dword_5d4594_1305748 = 1;
	}
}

//----- (0049F860) --------------------------------------------------------
int sub_49F860() {
	int result; // eax

	result = dword_5d4594_1305748;
	if (dword_5d4594_1305748) {
		nox_draw_curDrawData_3799572->flag_0 = *getMemU32Ptr(0x5D4594, 1305772);
		nox_rect* v1 = &nox_draw_curDrawData_3799572->clip;
		v1->min_x = *getMemU32Ptr(0x5D4594, 1305756);
		v1->min_y = *getMemU32Ptr(0x5D4594, 1305760);
		v1->max_x = *getMemU32Ptr(0x5D4594, 1305764);
		v1->max_y = *getMemU32Ptr(0x5D4594, 1305768);
		nox_rect* v2 = &nox_draw_curDrawData_3799572->rect2;
		v2->min_x = *getMemU32Ptr(0x5D4594, 1305732);
		v2->min_y = *getMemU32Ptr(0x5D4594, 1305736);
		result = *getMemU32Ptr(0x5D4594, 1305740);
		v2->max_x = *getMemU32Ptr(0x5D4594, 1305740);
		v2->max_y = *getMemU32Ptr(0x5D4594, 1305744);
		dword_5d4594_1305748 = 0;
	}
	return result;
}

//----- (0049F930) --------------------------------------------------------
int4* nox_xxx_utilRect_49F930(int4* a1, int4* a2, int4* a3) {
	int v3;       // ecx
	int v4;       // edx
	int v5;       // ebx
	int v6;       // edi
	int4* result; // eax

	v3 = a3->field_0;
	if (a2->field_0 >= a3->field_0) {
		v3 = a2->field_0;
	}
	v4 = a3->field_8;
	if (a2->field_8 <= v4) {
		v4 = a2->field_8;
	}
	if (v3 >= v4) {
		return 0;
	}
	v5 = a3->field_4;
	if (a2->field_4 >= v5) {
		v5 = a2->field_4;
	}
	v6 = a3->field_C;
	if (a2->field_C <= v6) {
		v6 = a2->field_C;
	}
	if (v5 >= v6) {
		return 0;
	}
	result = a1;
	a1->field_0 = v3;
	a1->field_4 = v5;
	a1->field_8 = v4;
	a1->field_C = v6;
	return result;
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
