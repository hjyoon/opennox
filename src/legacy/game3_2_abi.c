#include "GAME3_2.h"

nox_list_item_t nox_common_maplist = {0};

extern uintptr_t nox_quest_closing_settings_4D70B0(void);
char* sub_4D70B0(void) { return (char*)nox_quest_closing_settings_4D70B0(); }

extern int nox_broadcast_player_text_4D9FD0(char tag, wchar2_t* text);
int nox_xxx_printToAll_4D9FD0(char tag, wchar2_t* format, ...) {
	wchar2_t text[256];
	va_list args;
	va_start(args, format);
	nox_vsnwprintf(text, sizeof(text) / sizeof(text[0]), format, args);
	va_end(args);
	text[255] = 0;
	return nox_broadcast_player_text_4D9FD0(tag, text);
}

extern uintptr_t nox_gui_refresh_rule_list_4CED40(char* map);
void* sub_4CED40(char* map) { return (void*)nox_gui_refresh_rule_list_4CED40(map); }

extern intptr_t nox_gui_settings_delta_4DF020(void);
char* sub_4DF020(void) { return (char*)nox_gui_settings_delta_4DF020(); }
