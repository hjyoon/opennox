#ifndef NOX_PORT_CLIENT_GUI_GUIMETER
#define NOX_PORT_CLIENT_GUI_GUIMETER

#include "defs.h"

typedef struct nox_client_gui_health_mana_state {
	uint32_t current;
	uint32_t maximum;
	uint32_t primary_color;
	uint32_t secondary_color;
	int poisoned;
	int meter_x;
	int meter_y;
	int root_x;
	int root_y;
	int poison_tube_ready;
} nox_client_gui_health_mana_state;

int sub_4710B0();
int sub_471160(nox_window* a1, int a2, int a3, int a4, int a5);
int nox_xxx_guiHealthManaInit_4714E0();
int nox_client_guiHealthManaState(int meter_index, nox_client_gui_health_mana_state* out);

int nox_xxx_guiHealthManaTubeDraw_471D10(nox_window* win, nox_window_data* draw_data);

#endif // NOX_PORT_CLIENT_GUI_GUIMETER
