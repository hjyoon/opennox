#ifndef NOX_PORT_CLIENT_GUI_GUIBRIEF
#define NOX_PORT_CLIENT_GUI_GUIBRIEF

#include <stddef.h>
#include "defs.h"
#include "quest_briefing_450980.h"
#include "quest_briefing_450a30.h"
#include "quest_win_screen_450770.h"

typedef struct {
	nox_window* outer;
	nox_window* panel;
	void* preview_font;
	char* dialog_path;
	uint32_t mode;
	uint32_t ready_for_dialog_end;
	uint32_t active;
	uint32_t dialog_enabled;
} nox_briefing_state_t;
_Static_assert(offsetof(nox_briefing_state_t, mode) == 4 * sizeof(void*), "briefing state word offset");
_Static_assert(sizeof(nox_briefing_state_t) == 4 * sizeof(void*) + 4 * sizeof(uint32_t), "briefing state size");
nox_briefing_state_t* nox_briefing_storage(void);

wchar2_t* sub_44E410();
int sub_44E8E0(nox_window* win, nox_window_data* draw);
int sub_44F0F0(nox_window* win, nox_window_data* draw);
int sub_44F300(nox_window* win, nox_window_data* draw);

#endif // NOX_PORT_CLIENT_GUI_GUIBRIEF
