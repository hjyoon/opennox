#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../client__gui__servopts__access.h"

typedef nox_window* (*server_access_load_fn)(nox_window*);
typedef int (*server_access_draw_fn)(nox_window*, nox_window_data*);
typedef int (*server_access_event_fn)(nox_window*, int, uintptr_t, uintptr_t);
typedef int (*server_access_layout_fn)(void);
typedef int* (*server_access_populate_fn)(void);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"server-access pointers must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_guiServerAccessLoad_4541D0,
		server_access_load_fn: 1, default: 0),
	"004541D0 must accept and return native window pointers");
_Static_assert(
	_Generic(&sub_454A90, server_access_draw_fn: 1, default: 0),
	"00454A90 must preserve the native draw callback ABI");
_Static_assert(
	_Generic(&nox_xxx_windowAccessProc_454BA0,
		server_access_event_fn: 1, default: 0),
	"00454BA0 must preserve native-width GUI event payloads");
_Static_assert(sizeof(nox_gui_server_access.root) == sizeof(void*),
	"server-access root must remain native-width");
_Static_assert(sizeof(nox_gui_server_access_state) == 21 * sizeof(void*),
	"server-access state must contain only native-width window pointers");
_Static_assert(sizeof(nox_gui_server_access.wnd_10200) == sizeof(void*),
	"server-access player list must remain native-width");
_Static_assert(sizeof(nox_gui_server_access.active_list) == sizeof(void*),
	"server-access active list must remain native-width");
_Static_assert(_Generic(&sub_454640, server_access_layout_fn: 1, default: 0),
	"00454640 must preserve its no-argument layout ABI");
_Static_assert(_Generic(&sub_454740, server_access_populate_fn: 1, default: 0),
	"00454740 must return a native-width pointer");

int main(void) {
	return 0;
}
