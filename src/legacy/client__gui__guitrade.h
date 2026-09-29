#ifndef NOX_PORT_CLIENT_GUI_GUITRADE
#define NOX_PORT_CLIENT_GUI_GUITRADE

#include <stddef.h>

#include "defs.h"

enum { NOX_GUI_TRADE_SLOT_COUNT = 4, NOX_GUI_TRADE_SLOT_ITEM_CAPACITY = 32 };

typedef struct nox_gui_trade_slot {
	nox_drawable* drawable;
	uint32_t count;
	uint32_t item_ids[NOX_GUI_TRADE_SLOT_ITEM_CAPACITY];
	uint32_t total_cost;
} nox_gui_trade_slot;

_Static_assert(offsetof(nox_gui_trade_slot, drawable) == 0,
	"trade drawable must remain the first slot field");
_Static_assert(offsetof(nox_gui_trade_slot, count) == sizeof(void*),
	"trade slot count must follow the native drawable pointer");
_Static_assert(offsetof(nox_gui_trade_slot, item_ids) == sizeof(void*) + sizeof(uint32_t),
	"trade item IDs must follow the slot count");
_Static_assert(offsetof(nox_gui_trade_slot, total_cost) == sizeof(void*) + 33 * sizeof(uint32_t),
	"trade total must follow all item IDs");
_Static_assert(sizeof(nox_gui_trade_slot) == (sizeof(void*) == 4 ? 140 : 144),
	"trade slot must use one native pointer and fixed-width scalar fields");

extern nox_gui_trade_slot* dword_5d4594_1320932;
extern nox_gui_trade_slot* dword_5d4594_1320936;
extern nox_drawable* dword_5d4594_1320968;
extern nox_gui_trade_slot* dword_5d4594_1320972;

nox_gui_trade_slot* nox_gui_trade_slot_at(int remote, int index);
void nox_gui_trade_slots_reset(void);
int sub_4C09D0();
int sub_4C15D0(const uint8_t* data);

#endif // NOX_PORT_CLIENT_GUI_GUITRADE
