#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../GAME3_1.h"

typedef int (*summon_draw_fn)(nox_window*, nox_window_data*);
typedef int (*summon_event_fn)(nox_window*, int, uintptr_t, uintptr_t);
typedef int (*summon_tooltip_fn)(nox_window*, nox_window_data*, uintptr_t);
typedef nox_gui_summon_record* (*summon_grid_lookup_fn)(const int2*);
typedef void (*summon_order_fn)(nox_gui_summon_record*, int);
typedef int* (*summon_grid_clear_fn)(void);
typedef wchar2_t* (*summon_name_fn)(nox_window*, const int2*);
typedef nox_gui_summon_record* (*summon_record_first_fn)(void);
typedef nox_gui_summon_record* (*summon_record_next_fn)(nox_gui_summon_record*);
typedef int (*summon_record_predicate_fn)(const nox_gui_summon_record*);
typedef void (*summon_create_fn)(int, int, int);
typedef int (*summon_record_place_fn)(nox_gui_summon_record*);
typedef int (*summon_grid_write_fn)(const int32_t*, int, nox_gui_summon_record*);
typedef int (*summon_grid_available_fn)(const int32_t*, int);
typedef void (*summon_remove_fn)(int, void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"summon-control pointers must remain native-width");
_Static_assert(sizeof(nox_gui_summon_record) == 32,
	"summon records must preserve their fixed PE32 payload layout");
_Static_assert(offsetof(nox_gui_summon_record, net_code) == 0,
	"summon net codes must remain at offset zero");
_Static_assert(offsetof(nox_gui_summon_record, thing_id) == 4,
	"summon thing IDs must remain at offset four");
_Static_assert(offsetof(nox_gui_summon_record, active) == 8,
	"summon active flags must remain at offset eight");
_Static_assert(offsetof(nox_gui_summon_record, grid_x) == 12 &&
	offsetof(nox_gui_summon_record, grid_y) == 16,
	"summon grid coordinates must preserve their PE32 offsets");
_Static_assert(offsetof(nox_gui_summon_record, size) == 20,
	"summon sizes must remain at offset twenty");
_Static_assert(offsetof(nox_gui_summon_record, slot) == 24 &&
	offsetof(nox_gui_summon_record, highlight) == 28,
	"summon slot and highlight fields must preserve their PE32 offsets");
_Static_assert(sizeof(dword_5d4594_1321204) == sizeof(void*),
	"the selected summon must remain a native pointer");

_Static_assert(_Generic(&nox_xxx_guiDrawSummonBox_4C1FE0,
	summon_draw_fn: 1, default: 0),
	"004C1FE0 must preserve the native draw callback ABI");
_Static_assert(_Generic(&sub_4C24A0, summon_draw_fn: 1, default: 0),
	"004C24A0 must preserve the native draw callback ABI");
_Static_assert(_Generic(&sub_4C26F0, summon_draw_fn: 1, default: 0),
	"004C26F0 must preserve the native draw callback ABI");
_Static_assert(_Generic(&sub_4C27F0, summon_draw_fn: 1, default: 0),
	"004C27F0 must preserve the native draw callback ABI");
_Static_assert(_Generic(&nox_xxx_wndSummonBigButtonProc_4C24B0,
	summon_event_fn: 1, default: 0),
	"004C24B0 must preserve native-width event payloads");
_Static_assert(_Generic(&nox_xxx_clientOrderCreature_4C2A60,
	summon_event_fn: 1, default: 0),
	"004C2A60 must preserve native-width event payloads");
_Static_assert(_Generic(&nox_xxx_wndSummonProc_4C2B10,
	summon_event_fn: 1, default: 0),
	"004C2B10 must preserve native-width event payloads");
_Static_assert(_Generic(&sub_4C2BD0, summon_event_fn: 1, default: 0),
	"004C2BD0 must preserve the native event callback ABI");
_Static_assert(_Generic(&sub_4C2BE0, summon_event_fn: 1, default: 0),
	"004C2BE0 must preserve the native event callback ABI");
_Static_assert(_Generic(&sub_4C2C20, summon_tooltip_fn: 1, default: 0),
	"004C2C20 must preserve native-width tooltip payloads");
_Static_assert(_Generic(&sub_4C2CE0, summon_tooltip_fn: 1, default: 0),
	"004C2CE0 must preserve native-width tooltip payloads");

_Static_assert(_Generic(&nox_xxx_wndSummonGet_4C2410,
	summon_grid_lookup_fn: 1, default: 0),
	"004C2410 must return a native summon-record pointer");
_Static_assert(_Generic(&nox_client_orderCreature,
	summon_order_fn: 1, default: 0),
	"creature orders must accept a native summon-record pointer");
_Static_assert(_Generic(&sub_4C2BF0, summon_grid_clear_fn: 1, default: 0),
	"004C2BF0 must retain its fixed-width grid result");
_Static_assert(_Generic(&sub_4C2C60, summon_name_fn: 1, default: 0),
	"004C2C60 must accept a native window pointer");
_Static_assert(_Generic(&sub_4C2D60, summon_record_first_fn: 1, default: 0),
	"004C2D60 must return a native summon-record pointer");
_Static_assert(_Generic(&sub_4C2D90, summon_record_next_fn: 1, default: 0),
	"004C2D90 must traverse native summon-record pointers");
_Static_assert(_Generic(&sub_4C2DD0,
	summon_record_predicate_fn: 1, default: 0),
	"004C2DD0 must inspect a native summon-record pointer");
_Static_assert(_Generic(&nox_xxx_cliSummonCreat_4C2E50,
	summon_create_fn: 1, default: 0),
	"004C2E50 must not narrow a summon-record pointer into its result");
_Static_assert(_Generic(&sub_4C2F20, summon_record_first_fn: 1, default: 0),
	"004C2F20 must return a native summon-record pointer");
_Static_assert(_Generic(&sub_4C2F70, summon_record_first_fn: 1, default: 0),
	"004C2F70 must preserve its native summon-record result");
_Static_assert(_Generic(&sub_4C2FD0,
	summon_record_place_fn: 1, default: 0),
	"004C2FD0 must accept a native summon-record pointer");
_Static_assert(_Generic(&sub_4C3030, summon_grid_write_fn: 1, default: 0),
	"004C3030 must store native summon-record pointers in the grid");
_Static_assert(_Generic(&sub_4C30C0,
	summon_grid_available_fn: 1, default: 0),
	"004C30C0 must preserve its fixed-width grid coordinates");
_Static_assert(_Generic(&nox_xxx_cliSummonOnDieOrBanish_4C3140,
	summon_remove_fn: 1, default: 0),
	"004C3140 must accept a native sprite payload pointer");
_Static_assert(_Generic(&sub_4C3210,
	summon_record_next_fn: 1, default: 0),
	"004C3210 must accept and return native summon-record pointers");

int main(void) {
	return 0;
}
