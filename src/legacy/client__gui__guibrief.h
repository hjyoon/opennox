#ifndef NOX_PORT_CLIENT_GUI_GUIBRIEF
#define NOX_PORT_CLIENT_GUI_GUIBRIEF

#include "defs.h"
#include "quest_briefing_450980.h"
#include "quest_briefing_450a30.h"
#include "quest_win_screen_450770.h"

typedef struct {
	nox_window* outer;
	nox_window* panel;
	void* preview_font;
} nox_briefing_state_t;
_Static_assert(sizeof(nox_briefing_state_t) == 3 * sizeof(void*), "briefing state size");
nox_briefing_state_t* nox_briefing_storage(void);

wchar2_t* sub_44E410();
int sub_44E8E0(nox_window* win, nox_window_data* draw);
int sub_44F0F0(nox_window* win, nox_window_data* draw);
int sub_44F300(nox_window* win, nox_window_data* draw);

#endif // NOX_PORT_CLIENT_GUI_GUIBRIEF
