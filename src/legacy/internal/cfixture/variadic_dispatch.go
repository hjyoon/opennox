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

static void abort_unsupported_wide_format(wchar2_t code) {
	const wchar2_t format[] = {'%', code, 0};
	setvbuf(stdout, NULL, _IONBF, 0);
	nox_swprintf(NULL, format);
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

static void network_log_printf_fixture(char* format) {
	nox_xxx_networkLog_printf_413D30(format, 37, "ready");
}

static int file_printf_fixture(FILE* file, int width) {
	return nox_fs_fprintf(file, "%0*d", width, 7);
}

static int file_mixed_printf_fixture(FILE* file) {
	return nox_fs_fprintf(file, "[%*.*s|%#08x|%+.2f|%cignored]", 5, 2, "abcdef", 0x12u, 2.5, 0);
}

static int narrow_vprintf_preserves_cursor(const char* format, ...) {
	const char expected[] = "-37|2|3|4|5|6|7|spill";
	char bounded[11], full[64];
	unsigned char before[sizeof(va_list)];
	va_list args;
	va_start(args, format);
	(void)va_arg(args, int);
	memcpy(before, &args, sizeof(before));
	memset(bounded, 0x5a, sizeof(bounded));
	memset(full, 0x5a, sizeof(full));

	int bounded_length = nox_vsnprintf(bounded + 1, 9, format, args);
	int unchanged = memcmp(before, &args, sizeof(before)) == 0;
	int full_length = unchanged ? nox_vsprintf(full + 1, format, args) : -1;
	unchanged = unchanged && memcmp(before, &args, sizeof(before)) == 0;
	va_end(args);

	return unchanged && bounded_length == sizeof(expected) - 1 && full_length == bounded_length &&
		bounded[0] == 0x5a && bounded[10] == 0x5a && memcmp(bounded + 1, expected, 9) == 0 &&
		full[0] == 0x5a && full[sizeof(expected) + 1] == 0x5a &&
		memcmp(full + 1, expected, sizeof(expected)) == 0;
}

static int wide_vprintf_preserves_cursor(const wchar2_t* format, ...) {
	const wchar2_t expected[] = {'2', '.', '5', '0', '|', '-', '8', '|', 'w', 'i', 'd', 'e', '|', '%', 0};
	wchar2_t bounded[11], full[64];
	unsigned char before[sizeof(va_list)];
	va_list args;
	va_start(args, format);
	(void)va_arg(args, int);
	memcpy(before, &args, sizeof(before));
	memset(bounded, 0x5a, sizeof(bounded));
	memset(full, 0x5a, sizeof(full));

	int bounded_length = nox_vsnwprintf(bounded + 1, 9, format, args);
	int unchanged = memcmp(before, &args, sizeof(before)) == 0;
	int full_length = unchanged ? nox_vswprintf(full + 1, format, args) : -1;
	unchanged = unchanged && memcmp(before, &args, sizeof(before)) == 0;
	va_end(args);

	return unchanged && bounded_length == sizeof(expected)/sizeof(expected[0]) - 1 && full_length == bounded_length &&
		bounded[0] == 0x5a5a && bounded[10] == 0x5a5a && memcmp(bounded + 1, expected, 9*sizeof(expected[0])) == 0 &&
		full[0] == 0x5a5a && full[sizeof(expected)/sizeof(expected[0]) + 1] == 0x5a5a &&
		memcmp(full + 1, expected, sizeof(expected)) == 0;
}

static int vprintf_preserves_cursor_fixture(void) {
	const wchar2_t format[] = {'%', '.', '2', 'f', '|', '%', 'd', '|', '%', 's', '|', '%', '%', 0};
	const wchar2_t text[] = {'w', 'i', 'd', 'e', 0};
	return narrow_vprintf_preserves_cursor("%d|%d|%d|%d|%d|%d|%d|%s", 101, -37, 2, 3, 4, 5, 6, 7, "spill") &&
		wide_vprintf_preserves_cursor(format, 101, 2.5, -8, text);
}

static int wide_mixed_fixture(wchar2_t* out, const wchar2_t* format, const char* narrow,
			      int first, const wchar2_t* wide, unsigned int second) {
	return nox_swprintf(out, format, narrow, first, wide, second);
}

static int wide_double_fixture(wchar2_t* out, const wchar2_t* format, uint64_t bits, int marker) {
	double value;
	memcpy(&value, &bits, sizeof(value));
	return nox_swprintf(out, format, value, marker);
}

static int variadic_registers_and_stack_fixture(void) {
	const wchar2_t first[] = {'w', 'i', 'd', 'e', 0};
	const wchar2_t tail[] = {'e', 'n', 'd', 0};
	const char narrow_expected[] = "-2147483648|abc|-1|wide|-543210FF|(null)|-9|Q|xyz|2147483647";
	char narrow[160];
	memset(narrow, 0x5a, sizeof(narrow));
	int length = nox_sprintf(narrow + 1, "%d|%s|%u|%S|%X|%s|%d|%c|%s|%d",
		INT32_MIN, "abc", UINT32_MAX, first, UINT32_C(0xabcdef01), (const char*)NULL,
		-9, 'Q', "xyz", INT32_MAX);
	if (length != sizeof(narrow_expected) - 1 || memcmp(narrow + 1, narrow_expected, sizeof(narrow_expected)) ||
		narrow[0] != 0x5a || narrow[sizeof(narrow_expected) + 1] != 0x5a) {
		return 0;
	}

	const char wide_source[] = "%d|%s|%.2f|%S|%u|%.2f|%d|%.2f|%S|%.2f|%X|%.2f|%d|%.2f|%s|%.2f|%S|%.2f|%d|%.2f|%u|%.2f|%s";
	const char wide_expected[] = "-17|wide|0.25|n|-1073741823|-2.50|42|3.25|q|4.50|-543210FF|5.25|-9|6.50|wide|7.25|s|8.50|123|9.25|-1|-10.50|end";
	wchar2_t wide_format[sizeof(wide_source)], wide[200];
	for (size_t i = 0; i < sizeof(wide_source); ++i) {
		wide_format[i] = (unsigned char)wide_source[i];
	}
	memset(wide, 0x5a, sizeof(wide));
	length = nox_swprintf(wide + 1, wide_format,
		-17, first, 0.25, "n", UINT32_C(0xc0000001), -2.5, 42, 3.25, "q", 4.5,
		UINT32_C(0xabcdef01), 5.25, -9, 6.5, first, 7.25, "s", 8.5, 123, 9.25,
		UINT32_MAX, -10.5, tail);
	if (length != sizeof(wide_expected) - 1 || wide[0] != 0x5a5a ||
		wide[sizeof(wide_expected) + 1] != 0x5a5a) {
		return 0;
	}
	for (size_t i = 0; i < sizeof(wide_expected); ++i) {
		if (wide[i + 1] != (unsigned char)wide_expected[i]) {
			return 0;
		}
	}
	return 1;
}

static void wide_fixed_printf_reference(char* output, int digits, ...) {
	char format[32];
	va_list args;
	va_start(args, digits);
	snprintf(format, sizeof(format), "%%.%df", digits);
	snprintf(output, 32, format, va_arg(args, double));
	va_end(args);
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
	nox_format_fixed_bits(input, host, sizeof(host), precision);
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

func VariadicRegistersAndStack() bool {
	return C.variadic_registers_and_stack_fixture() != 0
}

func AbortUnsupportedWideFormat(code uint16) {
	C.abort_unsupported_wide_format(C.wchar2_t(code))
}

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

func NetworkLogPrintf413D30(format *byte) {
	C.network_log_printf_fixture((*C.char)(unsafe.Pointer(format)))
}

func FilePrintf(file unsafe.Pointer, width int) int {
	return int(C.file_printf_fixture((*C.FILE)(file), C.int(width)))
}

func FileMixedPrintf(file unsafe.Pointer) int {
	return int(C.file_mixed_printf_fixture((*C.FILE)(file)))
}

func VPrintfPreservesCursor() bool {
	return C.vprintf_preserves_cursor_fixture() != 0
}

func WideMixedPrintf(out, format *uint16, narrow *byte, first int32, wide *uint16, second uint32) int {
	return int(C.wide_mixed_fixture(
		(*C.wchar2_t)(unsafe.Pointer(out)), (*C.wchar2_t)(unsafe.Pointer(format)),
		(*C.char)(unsafe.Pointer(narrow)), C.int(first), (*C.wchar2_t)(unsafe.Pointer(wide)), C.uint(second),
	))
}

func WideDoublePrintf(out, format *uint16, bits uint64, marker int32) int {
	return int(C.wide_double_fixture(
		(*C.wchar2_t)(unsafe.Pointer(out)), (*C.wchar2_t)(unsafe.Pointer(format)), C.uint64_t(bits), C.int(marker),
	))
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
