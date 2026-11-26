#ifndef NOX_PORT_CLIENT_GUI_GUIINV
#define NOX_PORT_CLIENT_GUI_GUIINV

#include <stddef.h>
#include "defs.h"

typedef nox_inventory_cell_t nox_inventory_grid_t[NOX_INVENTORY_CELLS_MAX];
nox_inventory_grid_t* nox_inventory_grid_storage(void);

typedef struct nox_inventory_state_t {
	nox_drawable* equipped[9];
	nox_window* root;
	uint32_t mode;
} nox_inventory_state_t;

_Static_assert(offsetof(nox_inventory_state_t, root) == 9 * sizeof(void*), "inventory root offset");
_Static_assert(offsetof(nox_inventory_state_t, mode) == 10 * sizeof(void*), "inventory mode offset");
nox_inventory_state_t* nox_inventory_state_storage(void);

int nox_xxx_spritePickup_461660(int net_code, int thing_type, const nox_modifier_attrs_t* attrs);
int sub_4617C0(int net_code, int thing_type, const nox_modifier_attrs_t* attrs, int2* position);
void sub_461A80(int a1);
void sub_462040(int a1);
int sub_462740();
uintptr_t sub_4627F0(uint32_t* a1);
int nox_client_inventory_total_weight_463880(void);
void nox_client_makePlayerStatsDlg_463880(int* a1);
int sub_4649B0(nox_drawable* drawable, int column, int row);
int sub_464BD0(nox_window* win, int a2, uintptr_t a3, uintptr_t a4);
void nox_xxx_cliInventorySpriteUpd_465A30();
int sub_466160();
int sub_4661D0();
wchar2_t* sub_466660(int a1, int2* a2);
nox_drawable* nox_inventory_prepare_tooltip_drawable_466660(nox_inventory_cell_t* cell);
int nox_xxx_inventroryOnHovewerSub_4667E0(int a1, int a2, unsigned int a3);
int sub_466E20(uint32_t* a1);
int nox_xxx_inventoryNameSignInit_4671E0();
int sub_467750(int a1, char a2);

#endif // NOX_PORT_CLIENT_GUI_GUIINV
