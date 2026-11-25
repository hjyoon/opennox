package cfixture

/*
#include "GAME3_2.h"
#include "GAME4_3.h"
#include "client__gui__guicon.h"
#include "common__log.h"
#include "common/fs/nox_fs.h"
#include "noxstring.h"
#include "noxstring_format_abi.h"
#include <fenv.h>
#include <stdio.h>
#include <string.h>

static intptr_t line_message_fixture(nox_object_t* unit, wchar2_t* format, int points) {
	const wchar2_t suffix[] = {'!', 0};
	return nox_xxx_netSendLineMessage_4D9EB0(unit, format, "xp", points, suffix);
}

static int broadcast_message_fixture(char tag, wchar2_t* format, int points) {
	const wchar2_t suffix[] = {0x17a, 0};
	return nox_xxx_printToAll_4D9FD0(tag, format, "xp", points, suffix);
}

static int console_printf_fixture(unsigned char color, wchar2_t* format, int points) {
	const wchar2_t suffix[] = {0x17a, 0};
	return nox_gui_console_Printf_450C00(color, format, "xp", points, suffix);
}

static void ai_debug_printf_fixture(char* format, int frame) {
	nox_ai_debug_printf_5341A0(format, frame, "monster");
}

static void network_log_printf_fixture(void) {
	char format[] = "queue %04d %s";
	nox_xxx_networkLog_printf_413D30(format, 37, "ready");
}

static int file_printf_fixture(FILE* file, int width) {
	return nox_fs_fprintf(file, "%0*d", width, 7);
}

static void wide_fixed_printf_reference(char* output, int digits, ...) {
	char format[32];
	va_list args;
	va_start(args, digits);
	snprintf(format, sizeof(format), "%%.%df", digits);
	snprintf(output, 32, format, va_arg(args, double));
	va_end(args);
}

static void wide_fixed_printf_host(char* output, int digits, ...) {
	nox_format_args args;
	va_start(args.list, digits);
	nox_format_arg_fixed(&args, output, 32, digits);
	va_end(args.list);
}

static int wide_fixed_printf_matches(uint64_t bits, int digits, int rounding) {
	const int modes[] = {FE_TONEAREST, FE_TOWARDZERO, FE_DOWNWARD, FE_UPWARD};
	fenv_t saved;
	if (fegetenv(&saved) || fesetround(modes[rounding])) {
		return -1;
	}

	char narrow[32];
	wchar2_t format[32], actual[33] = {0};
	if (digits < 0) {
		strcpy(narrow, "%f");
	} else {
		snprintf(narrow, sizeof(narrow), "%%.%df", digits);
	}
	for (size_t i = 0; i <= strlen(narrow); ++i) {
		format[i] = (unsigned char)narrow[i];
	}
	double value;
	memcpy(&value, &bits, sizeof(value));
	char expected[32] = {0}, host[32] = {0};
	int precision = digits > 0 ? digits : 5;

	feclearexcept(FE_ALL_EXCEPT);
	wide_fixed_printf_reference(expected, precision, value);
	int flags = fetestexcept(FE_ALL_EXCEPT);
	size_t length = strlen(expected);

	// Compare CRT exception flags within C. The Go callback may run runtime
	// bookkeeping on this thread before the formatting call.
	feclearexcept(FE_ALL_EXCEPT);
	wide_fixed_printf_host(host, precision, value);
	int status = (flags != fetestexcept(FE_ALL_EXCEPT) || memcmp(host, expected, sizeof(host))) ? 4 : 0;

	feclearexcept(FE_ALL_EXCEPT);
	actual[length + 1] = 0x4242;
	int used = nox_swprintf(actual, format, value);
	if (used != (int)length) status |= 2;
	if (actual[length + 1] != 0x4242) status |= 8;
	if (fegetround() != modes[rounding]) status |= 16;
	for (size_t i = 0; i <= length; ++i) {
		if (actual[i] != (unsigned char)expected[i]) status |= 32;
	}
	fesetenv(&saved);
	return status ? status : 1;
}
*/
import "C"

import "unsafe"

func LineMessage4D9EB0(unit unsafe.Pointer, format *uint16, points int) uintptr {
	return uintptr(C.line_message_fixture((*C.nox_object_t)(unit), (*C.wchar2_t)(unsafe.Pointer(format)), C.int(points)))
}

func BroadcastMessage4D9FD0(tag byte, format *uint16, points int) int {
	return int(C.broadcast_message_fixture(C.char(tag), (*C.wchar2_t)(unsafe.Pointer(format)), C.int(points)))
}

func ConsolePrintf450C00(color byte, format *uint16, points int) int {
	return int(C.console_printf_fixture(C.uchar(color), (*C.wchar2_t)(unsafe.Pointer(format)), C.int(points)))
}

func AIDebugPrintf5341A0(format *byte, frame int) {
	C.ai_debug_printf_fixture((*C.char)(unsafe.Pointer(format)), C.int(frame))
}

func NetworkLogPrintf413D30() {
	C.network_log_printf_fixture()
}

func FilePrintf(file unsafe.Pointer, width int) int {
	return int(C.file_printf_fixture((*C.FILE)(file), C.int(width)))
}

func WideFixedPrintfMatches(bits uint64, digits, rounding int) int {
	return int(C.wide_fixed_printf_matches(C.uint64_t(bits), C.int(digits), C.int(rounding)))
}
