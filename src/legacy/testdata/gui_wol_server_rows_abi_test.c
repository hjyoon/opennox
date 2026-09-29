#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../GAME1_2.h"
#include "../GAME2_3.h"
#include "../GAME3.h"
#include "../client__shell__noxworld.h"

typedef nox_window* (*server_row_get_fn)(const nox_gui_server_ent_t*);
typedef void (*server_row_set_fn)(nox_gui_server_ent_t*, nox_window*);
typedef nox_list_item_t* (*server_list_fn)(void);
typedef nox_gui_server_ent_t* (*server_lookup_fn)(int);
typedef int (*server_near_count_fn)(uint32_t*, nox_list_item_t*);
typedef nox_window* (*server_near_open_fn)(nox_window*, uint32_t*, nox_list_item_t*);
typedef void (*server_info_fn)(const nox_gui_server_ent_t*);
typedef int (*server_icon_fn)(nox_gui_server_ent_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(nox_gui_server_ent_t) == 169,
	"WOL server wire payload must remain exactly 169 bytes");
_Static_assert(offsetof(nox_gui_server_node_t, server) >= sizeof(nox_list_item_t),
	"native list links must not overlap the server payload");
_Static_assert(sizeof(((nox_gui_server_node_t*)0)->row) == sizeof(uintptr_t),
	"server-row windows must remain native-width");
_Static_assert(_Generic(&nox_wol_server_row_get, server_row_get_fn: 1, default: 0),
	"server row lookup must preserve native window pointers");
_Static_assert(_Generic(&nox_wol_server_row_set, server_row_set_fn: 1, default: 0),
	"server row storage must preserve native window pointers");
_Static_assert(_Generic(&sub_4A0020, server_list_fn: 1, default: 0),
	"WOL server list access must use native list links");
_Static_assert(_Generic(&sub_4A0490, server_lookup_fn: 1, default: 0),
	"map-row lookup must return a typed server payload");
_Static_assert(_Generic(&sub_4A04C0, server_lookup_fn: 1, default: 0),
	"list-row lookup must return a typed server payload");
_Static_assert(_Generic(&sub_4A28C0, server_lookup_fn: 1, default: 0),
	"nearby-row lookup must return a typed server payload");
_Static_assert(_Generic(&sub_4A25C0, server_near_count_fn: 1, default: 0),
	"nearby counting must traverse native list links");
_Static_assert(_Generic(&sub_4A2610, server_near_open_fn: 1, default: 0),
	"nearby popup creation must traverse native list links");
_Static_assert(_Generic(&nox_client_gui_serverInfoBlock_4394D0, server_info_fn: 1, default: 0),
	"server details must consume a typed server payload");
_Static_assert(_Generic(&sub_437320, server_icon_fn: 1, default: 0),
	"server icon refresh must consume a typed server payload");

int main(void) {
	nox_gui_server_node_t node = {0};
#if UINTPTR_MAX > UINT32_MAX
	uintptr_t marker = (uintptr_t)(UINT64_C(1) << 40) | UINT64_C(0x12345678);
#else
	uintptr_t marker = UINT32_C(0x12345678);
#endif
	nox_window* row = (nox_window*)marker;

	node.row = row;
	node.server.field_7 = (int)UINT32_C(0xa5a5a5a5);
	if (nox_wol_server_node_from_list(&node.list) != &node ||
		nox_wol_server_node_from_record(&node.server) != &node || node.row != row ||
		(uint32_t)node.server.field_7 != UINT32_C(0xa5a5a5a5)) {
		return 1;
	}
	return 0;
}
