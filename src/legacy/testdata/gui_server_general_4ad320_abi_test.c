#include <limits.h>
#include <stdint.h>

#include "../GAME3.h"
#include "../client__gui__servopts__general.h"

typedef nox_window* (*server_general_load_fn)(nox_window*);
typedef int (*server_general_draw_fn)(nox_window*, nox_window_data*);
typedef int (*server_general_event_fn)(nox_window*, int, nox_window*, uintptr_t);

extern nox_window* dword_5d4594_1309812;

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"server-general pointers must remain native-width");
_Static_assert(_Generic(&dword_5d4594_1309812, nox_window**: 1, default: 0),
	"server-general root storage must remain a native window pointer");
_Static_assert(_Generic(&nox_xxx_gui_4AD320, server_general_load_fn: 1, default: 0),
	"004AD320 must accept and return native window pointers");
_Static_assert(_Generic(&sub_4AD570, server_general_draw_fn: 1, default: 0),
	"004AD570 must preserve the native draw callback ABI");
_Static_assert(
	_Generic(&nox_xxx_windowServerOptionsGeneralProc_4AD5D0,
		server_general_event_fn: 1, default: 0),
	"004AD5D0 must preserve native-width GUI event payloads");

int main(void) {
	return 0;
}
