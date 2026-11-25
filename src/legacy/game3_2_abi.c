#include "GAME3_2.h"
#include "noxstring_format_abi.h"

nox_list_item_t nox_common_maplist = {0};

extern uintptr_t nox_quest_closing_settings_4D70B0(void);
char* sub_4D70B0(void) { return (char*)nox_quest_closing_settings_4D70B0(); }

extern int nox_print_to_all_go_4D9FD0(char tag, nox_format_word* format, nox_format_args* args);
int nox_xxx_printToAll_4D9FD0(char tag, wchar2_t* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	int result = nox_print_to_all_go_4D9FD0(tag, format, &args);
	va_end(args.list);
	return result;
}

extern uintptr_t nox_gui_refresh_rule_list_4CED40(char* map);
void* sub_4CED40(char* map) { return (void*)nox_gui_refresh_rule_list_4CED40(map); }

extern intptr_t nox_gui_settings_delta_4DF020(void);
char* sub_4DF020(void) { return (char*)nox_gui_settings_delta_4DF020(); }
