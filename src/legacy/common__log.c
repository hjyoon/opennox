#include "common__log.h"
#include "noxstring_format_abi.h"

//----- (00413D30) --------------------------------------------------------
void nox_network_log_vprintf_413D30(nox_format_byte* format, nox_format_args* args);
void nox_xxx_networkLog_printf_413D30(char* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	nox_network_log_vprintf_413D30(format, &args);
	va_end(args.list);
}
