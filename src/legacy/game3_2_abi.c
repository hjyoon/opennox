#include "GAME3_2.h"
#include "noxstring_format_abi.h"

extern int nox_print_to_all_go_4D9FD0(char tag, nox_format_word* format, nox_format_args* args);
int nox_xxx_printToAll_4D9FD0(char tag, wchar2_t* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	int result = nox_print_to_all_go_4D9FD0(tag, format, &args);
	va_end(args.list);
	return result;
}
