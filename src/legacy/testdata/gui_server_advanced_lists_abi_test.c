#include <limits.h>
#include <stdint.h>

#include "../GAME1.h"
#include "../GAME2.h"
#include "../GAME3_1.h"
#include "../GAME3_2.h"
#include "../client__gui__servopts__objlst.h"
#include "../client__gui__servopts__spelllst.h"

typedef nox_window* (*object_list_load_fn)(nox_window*, int);
typedef nox_window* (*spell_list_load_fn)(nox_window*);
typedef int (*list_draw_fn)(nox_window*, nox_window_data*);
typedef int (*list_event_fn)(nox_window*, int, nox_window*, uintptr_t);
typedef int (*advanced_load_fn)(int*);
typedef int (*rules_event_fn)(nox_window*, unsigned int, nox_window*, uintptr_t);
typedef nox_window* (*rules_load_fn)(nox_window*, char*);
typedef void* (*rules_scan_fn)(char*);
typedef int (*title_to_mask_fn)(wchar2_t*);
typedef wchar2_t* (*mask_to_title_fn)(int);

extern nox_window* dword_5d4594_1045464;
extern nox_window* dword_5d4594_1045468;
extern nox_window* dword_5d4594_1045480;
extern nox_window* dword_5d4594_1045484;
extern nox_window* dword_5d4594_1045508;
extern nox_window* dword_5d4594_1316708;
extern nox_window* dword_5d4594_1316712;
extern nox_window* dword_5d4594_1523024;
extern nox_window* dword_5d4594_1523028;
extern nox_window* dword_5d4594_1523032;
extern nox_window* dword_5d4594_1523036;
extern nox_window* dword_5d4594_1523040;
extern nox_window* dword_5d4594_1523044;
extern nox_window* dword_5d4594_1523048;

#define ASSERT_NATIVE_WINDOW_GLOBAL(name) \
	_Static_assert(_Generic(&(name), nox_window**: 1, default: 0), \
		#name " must store a native window pointer")

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"GUI callback payloads must remain native-width");

ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1045464);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1045468);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1045480);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1045484);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1045508);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1316708);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1316712);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1523024);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1523028);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1523032);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1523036);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1523040);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1523044);
ASSERT_NATIVE_WINDOW_GLOBAL(dword_5d4594_1523048);

_Static_assert(_Generic(&nox_xxx_guiObjlistLoad_4530C0,
	object_list_load_fn: 1, default: 0),
	"004530C0 must accept and return native window pointers");
_Static_assert(_Generic(&sub_453350, list_draw_fn: 1, default: 0),
	"00453350 must preserve the native object-list draw ABI");
_Static_assert(_Generic(&sub_4533D0, list_event_fn: 1, default: 0),
	"004533D0 must preserve native object-list event payloads");
_Static_assert(_Generic(&nox_xxx_guiSpelllistLoad_453850,
	spell_list_load_fn: 1, default: 0),
	"00453850 must accept and return native window pointers");
_Static_assert(_Generic(&sub_453B80, list_draw_fn: 1, default: 0),
	"00453B80 must preserve the native spell-list draw ABI");
_Static_assert(_Generic(&sub_453C00, list_event_fn: 1, default: 0),
	"00453C00 must preserve native spell-list event payloads");

_Static_assert(_Generic(&nox_xxx_loadAdvancedWnd_4BDC10,
	advanced_load_fn: 1, default: 0),
	"004BDC10 must preserve the advanced-options settings pointer");
_Static_assert(_Generic(&sub_4BDDA0, list_event_fn: 1, default: 0),
	"004BDDA0 must preserve native advanced-window event payloads");
_Static_assert(_Generic(&nox_xxx_windowAdvancedServProc_4BDDB0,
	list_event_fn: 1, default: 0),
	"004BDDB0 must preserve native advanced-options event payloads");

_Static_assert(_Generic(&sub_4CEBA0, rules_load_fn: 1, default: 0),
	"004CEBA0 must accept and return native window pointers");
_Static_assert(_Generic(&sub_4CED40, rules_scan_fn: 1, default: 0),
	"004CED40 must preserve the rule-file scan result");
_Static_assert(_Generic(&sub_4CEED0, list_draw_fn: 1, default: 0),
	"004CEED0 must preserve the native rule-list draw ABI");
_Static_assert(_Generic(&sub_4CF060, rules_event_fn: 1, default: 0),
	"004CF060 must preserve native rule-list event payloads");

_Static_assert(_Generic(&sub_415960, title_to_mask_fn: 1, default: 0),
	"00415960 must accept native weapon-title pointers");
_Static_assert(_Generic(&sub_4159F0, mask_to_title_fn: 1, default: 0),
	"004159F0 must return native weapon-title pointers");
_Static_assert(_Generic(&sub_415DA0, title_to_mask_fn: 1, default: 0),
	"00415DA0 must accept native armor-title pointers");
_Static_assert(_Generic(&sub_415E80, mask_to_title_fn: 1, default: 0),
	"00415E80 must return native armor-title pointers");

int main(void) {
	return 0;
}
