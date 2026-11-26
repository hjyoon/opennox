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

_Static_assert(_Generic(&nox_format_wide_number_abi,
	void (*)(wchar2_t*, size_t, const wchar2_t*, int): 1, default: 0), "fixed number ABI");
_Static_assert(_Generic(&nox_format_wide_strings_abi,
	void (*)(wchar2_t*, size_t, const wchar2_t*, const wchar2_t*, const wchar2_t*): 1,
	default: 0), "fixed string ABI");

static void bounded_wide_variadic_reference(wchar2_t* out, size_t capacity, const wchar2_t* format, ...) {
	va_list args;
	va_start(args, format);
	nox_vsnwprintf(out, capacity - 1, format, args);
	va_end(args);
	out[capacity - 1] = 0;
}

static int bounded_wide_fixed_matches(const wchar2_t* format, int value,
                                     const wchar2_t* first, const wchar2_t* second, int strings) {
	for (size_t capacity = 0; capacity <= 32; capacity++) {
		wchar2_t actual[34], expected[34];
		memset(actual, 0xa5, sizeof(actual));
		memset(expected, 0xa5, sizeof(expected));
		if (strings) {
			nox_format_wide_strings_abi(actual + 1, capacity, format, first, second);
			if (capacity) {
				bounded_wide_variadic_reference(expected + 1, capacity, format, first, second);
			}
		} else {
			nox_format_wide_number_abi(actual + 1, capacity, format, value);
			if (capacity) {
				bounded_wide_variadic_reference(expected + 1, capacity, format, value);
			}
		}
		if (memcmp(actual, expected, sizeof(actual))) {
			return 0;
		}
	}
	return 1;
}

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

static double wide_fixed_printf_value(const volatile uint64_t* input) {
	uint64_t bits = *input;
	double value;
	memcpy(&value, &bits, sizeof(value));
	return value;
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
	// Each invocation needs a fresh copy of the input bits. On x87, a
	// floating-point load can quiet a signaling NaN and raise FE_INVALID.
	volatile uint64_t input = bits;
	char expected[32] = {0}, host[32] = {0};
	int precision = digits > 0 ? digits : 5;

	feclearexcept(FE_ALL_EXCEPT);
	wide_fixed_printf_reference(expected, precision, wide_fixed_printf_value(&input));
	int flags = fetestexcept(FE_ALL_EXCEPT);
	size_t length = strlen(expected);

	// Compare CRT exception flags within C. The Go callback may run runtime
	// bookkeeping on this thread before the formatting call.
	feclearexcept(FE_ALL_EXCEPT);
	wide_fixed_printf_host(host, precision, wide_fixed_printf_value(&input));
	int status = (flags != fetestexcept(FE_ALL_EXCEPT) || memcmp(host, expected, sizeof(host))) ? 4 : 0;

	feclearexcept(FE_ALL_EXCEPT);
	actual[length + 1] = 0x4242;
	int used = nox_swprintf(actual, format, wide_fixed_printf_value(&input));
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

func BoundedWideNumberMatches(format *uint16, value int32) bool {
	return C.bounded_wide_fixed_matches((*C.wchar2_t)(unsafe.Pointer(format)), C.int(value), nil, nil, 0) != 0
}

func BoundedWideStringsMatch(format, first, second *uint16) bool {
	return C.bounded_wide_fixed_matches((*C.wchar2_t)(unsafe.Pointer(format)), 0,
		(*C.wchar2_t)(unsafe.Pointer(first)), (*C.wchar2_t)(unsafe.Pointer(second)), 1) != 0
}
