#include <stddef.h>
#include <stdint.h>

#include "../client__gui__guiquit.h"
#include "../client__gui__window.h"

typedef int (*quit_menu_proc_fn)(nox_window*, int, nox_window*, uintptr_t);

_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"quit-menu GUI pointers must remain native-width");
_Static_assert(_Generic(&nox_xxx_menuGameOnButton_445840, quit_menu_proc_fn: 1, default: 0),
	"00445840 must preserve native-width window and event pointers");
_Static_assert(offsetof(nox_window, draw_data) == (sizeof(void*) == 4 ? 36 : 64),
	"quit-menu button state must use the native window layout");
_Static_assert(offsetof(nox_window_data, field_0) == 0,
	"quit-menu button state must address the draw-data flags field");

int main(void) {
	return 0;
}
