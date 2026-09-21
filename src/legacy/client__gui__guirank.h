#ifndef NOX_PORT_CLIENT_GUI_GUIRANK
#define NOX_PORT_CLIENT_GUI_GUIRANK

#include "defs.h"

nox_window* nox_rank_window_at(int offset);
const wchar2_t* nox_rank_class_name(unsigned int class_ind);
int sub_46DC60(nox_window* win, unsigned char color, const wchar2_t* text);
nox_window* nox_xxx_guiDrawRank_46E870();
wchar2_t* sub_46F030();
int sub_46F080(nox_window* win, nox_window_data* draw);
int sub_46F8F0(int page, int empty_rows);
wchar2_t* sub_46FB50(int a1, uint8_t* a2);
char sub_46FC50();
int sub_46FD80();

#endif // NOX_PORT_CLIENT_GUI_GUIRANK
