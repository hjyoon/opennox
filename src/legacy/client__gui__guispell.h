#ifndef NOX_PORT_CLIENT_GUI_GUISPELL
#define NOX_PORT_CLIENT_GUI_GUISPELL

#include "defs.h"

typedef struct nox_window_data nox_window_data;
typedef struct nox_window nox_window;

extern void* nox_quickbar_font;

int nox_xxx_spellPutInBox_45DEB0(void* data, int spell, int x, int y);
void nox_client_buildTrap_45E040();
int nox_xxx_quickBarCreate_45E190();
int nox_xxx_quickbarButtonBook_45F3F0();
int sub_45F480(nox_window* win, nox_window_data* draw_data, int a3);
int sub_45F9B0(nox_window* win, nox_window_data* draw_data);
int nox_xxx_quickbarDraw_45FAC0(nox_window* win, nox_window_data* draw_data);
int nox_xxx_quickBarDrawFn_45FBD0(nox_window* win, nox_window_data* draw_data);
int nox_xxx_quickBarWarriorDraw_45FDE0(nox_window* win, nox_window_data* draw_data);
int sub_460070();

uint32_t* nox_xxx_quickbarAddTrap_460EC0(int a1);

#endif // NOX_PORT_CLIENT_GUI_GUISPELL
