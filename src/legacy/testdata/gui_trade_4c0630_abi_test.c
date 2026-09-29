#include <limits.h>
#include <stdint.h>

#include "../GAME3_1.h"

typedef int (*trade_event_fn)(nox_window*, int, uintptr_t, uintptr_t);
typedef int (*trade_draw_fn)(nox_window*, nox_window_data*);
typedef int (*trade_tooltip_fn)(nox_window*, nox_window_data*, int);
typedef uint32_t* (*trade_slot_fn)(int2*);

extern nox_window* dword_5d4594_1320940;

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"trade GUI pointers must remain native-width");
_Static_assert(_Generic(&dword_5d4594_1320940, nox_window**: 1, default: 0),
	"the trade root window must remain a native pointer");
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

int main(void) {
	return 0;
}
