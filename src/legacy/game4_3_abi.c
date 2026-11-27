//go:build !386 && !amd64 && !arm && !arm64

#include "GAME4_3.h"
#include "noxstring_format_abi.h"

//----- (005341A0) --------------------------------------------------------
void nox_ai_debug_vprintf_5341A0(nox_format_byte* format, nox_format_args* args);
void nox_ai_debug_printf_5341A0(char* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	nox_ai_debug_vprintf_5341A0(format, &args);
	va_end(args.list);
}
