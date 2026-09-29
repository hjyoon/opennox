#include <limits.h>
#include <stdint.h>

#include "../GAME3_1.h"

typedef int (*trade_event_fn)(nox_window*, int, uintptr_t, uintptr_t);
typedef int (*trade_draw_fn)(nox_window*, nox_window_data*);
typedef int (*trade_tooltip_fn)(nox_window*, nox_window_data*, int);
typedef nox_gui_trade_slot* (*trade_slot_fn)(int2*);
typedef int (*trade_packet_fn)(const uint8_t*);
typedef nox_gui_trade_slot* (*trade_add_item_fn)(const uint8_t*);

extern nox_window* dword_5d4594_1320940;

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"trade GUI pointers must remain native-width");
_Static_assert(_Generic(&dword_5d4594_1320940, nox_window**: 1, default: 0),
	"the trade root window must remain a native pointer");
_Static_assert(_Generic(&dword_5d4594_1320932, nox_gui_trade_slot**: 1, default: 0),
	"the local recent trade slot must remain a native pointer");
_Static_assert(_Generic(&dword_5d4594_1320936, nox_gui_trade_slot**: 1, default: 0),
	"the remote recent trade slot must remain a native pointer");
_Static_assert(_Generic(&dword_5d4594_1320968, nox_drawable**: 1, default: 0),
	"the dragged trade item must remain a native pointer");
_Static_assert(_Generic(&dword_5d4594_1320972, nox_gui_trade_slot**: 1, default: 0),
	"the dragged item's source slot must remain a native pointer");
_Static_assert(sizeof(nox_gui_trade_slot) == (sizeof(void*) == 4 ? 140 : 144),
	"trade slots must widen only their drawable pointer");
_Static_assert(offsetof(nox_gui_trade_slot, total_cost) == sizeof(void*) + 33 * sizeof(uint32_t),
	"trade slot scalar fields must retain fixed-width ordering");
_Static_assert(_Generic(&sub_4C0630, trade_event_fn: 1, default: 0),
	"004C0630 must preserve native-width GUI event payloads");
_Static_assert(_Generic(&sub_4C0C90, trade_event_fn: 1, default: 0),
	"004C0C90 must preserve native-width GUI event payloads");
_Static_assert(_Generic(&sub_4C0D00, trade_draw_fn: 1, default: 0),
	"004C0D00 must preserve the native draw callback ABI");
_Static_assert(_Generic(&sub_4C1120, trade_tooltip_fn: 1, default: 0),
	"004C1120 must preserve the native tooltip callback ABI");
_Static_assert(_Generic(&sub_4C0910, trade_slot_fn: 1, default: 0),
	"004C0910 must return the complete trade-slot pointer");
_Static_assert(_Generic(&sub_4C11E0, trade_slot_fn: 1, default: 0),
	"004C11E0 must return the complete trade-slot pointer");
_Static_assert(_Generic(&nox_xxx_netP2PStartTrade_4C1320, trade_packet_fn: 1, default: 0),
	"004C1320 must consume a packet pointer without narrowing it");
_Static_assert(_Generic(&sub_4C15D0, trade_packet_fn: 1, default: 0),
	"004C15D0 must consume a packet pointer without narrowing it");
_Static_assert(_Generic(&nox_xxx_tradeClientAddItem_4C1790, trade_add_item_fn: 1, default: 0),
	"004C1790 must preserve packet and slot pointer widths");
_Static_assert(_Generic(&sub_4C1B50, trade_packet_fn: 1, default: 0),
	"004C1B50 must consume a packet pointer without narrowing it");
_Static_assert(_Generic(&sub_4C1BC0, trade_packet_fn: 1, default: 0),
	"004C1BC0 must consume a packet pointer without narrowing it");

int main(void) {
	return 0;
}
