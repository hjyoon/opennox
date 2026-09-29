#include <stdint.h>

#include "../GAME1_2.h"
#include "../GAME2_2.h"
#include "../client__gui__window.h"

typedef nox_window* (*filter_create_fn)(nox_window*);
typedef int (*filter_event_fn)(nox_window*, int, uintptr_t, uintptr_t);
typedef int (*window_draw_fn)(nox_window*, nox_window_data*);

extern nox_window* nox_wol_wnd_world_814980;
extern nox_window* dword_5d4594_814984;
extern nox_window* dword_5d4594_814988;
extern nox_window* dword_5d4594_814992;
extern nox_window* dword_5d4594_814996;
extern nox_window* dword_5d4594_815000;
extern nox_window* dword_5d4594_815008;
extern nox_window* dword_5d4594_1193380;
extern nox_window* dword_5d4594_1193384;

#define ASSERT_NATIVE_WINDOW_GLOBAL(name) \
	_Static_assert(_Generic(&(name), nox_window**: 1, default: 0), \
		#name " must remain a native window pointer")

_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"WOL GUI pointers and event payloads must remain native-width");
ASSERT_NATIVE_WINDOW_GLOBAL(nox_wol_wnd_world_814980);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_814984);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_814988);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_814992);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_814996);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_815000);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_815008);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1193380);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1193384);
_Static_assert(_Generic(&sub_489B80, filter_create_fn: 1, default: 0),
	"00489B80 must preserve its parent and result window pointers");
_Static_assert(_Generic(&nox_xxx_windowMplayFilterProc_489E70,
	filter_event_fn: 1, default: 0),
	"00489E70 must preserve native-width GUI event payloads");
_Static_assert(_Generic(&sub_438C80, window_draw_fn: 1, default: 0),
	"00438C80 must use the native window draw callback ABI");

int main(void) {
	return 0;
}
