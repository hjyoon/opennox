#ifndef NOX_PORT_CLIENT_GUI_GUISUMN
#define NOX_PORT_CLIENT_GUI_GUISUMN

#include "defs.h"

typedef struct nox_gui_summon_record {
	uint32_t net_code;
	uint32_t thing_id;
	uint32_t active;
	int32_t grid_x;
	int32_t grid_y;
	uint8_t size;
	uint8_t reserved_21[3];
	uint32_t slot;
	uint32_t highlight;
} nox_gui_summon_record;

_Static_assert(sizeof(nox_gui_summon_record) == 32, "wrong summon-record size");
_Static_assert(offsetof(nox_gui_summon_record, net_code) == 0, "wrong summon-record net-code offset");
_Static_assert(offsetof(nox_gui_summon_record, thing_id) == 4, "wrong summon-record thing-id offset");
_Static_assert(offsetof(nox_gui_summon_record, active) == 8, "wrong summon-record active offset");
_Static_assert(offsetof(nox_gui_summon_record, grid_x) == 12, "wrong summon-record grid-x offset");
_Static_assert(offsetof(nox_gui_summon_record, grid_y) == 16, "wrong summon-record grid-y offset");
_Static_assert(offsetof(nox_gui_summon_record, size) == 20, "wrong summon-record size offset");
_Static_assert(offsetof(nox_gui_summon_record, slot) == 24, "wrong summon-record slot offset");
_Static_assert(offsetof(nox_gui_summon_record, highlight) == 28, "wrong summon-record highlight offset");

extern nox_gui_summon_record* dword_5d4594_1321204;

int nox_xxx_guiSummonCreatureLoad_4C1D80(void);
void nox_xxx_wndSummonCreateList_4C2560(int2* a1);
int sub_4C27F0(nox_window* win, nox_window_data* draw);
int sub_4C2CE0(nox_window* win, nox_window_data* draw, uintptr_t packed_position);

#endif // NOX_PORT_CLIENT_GUI_GUISUMN
