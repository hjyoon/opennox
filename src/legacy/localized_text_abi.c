#include <stdarg.h>
#include "nox_wchar.h"

static void format_bounded_wide(wchar2_t* out, size_t capacity, const wchar2_t* format, ...) {
	va_list args;
	va_start(args, format);
	nox_vsnwprintf(out, capacity - 1, format, args);
	va_end(args);
	out[capacity - 1] = 0;
}

void nox_format_wide_strings_abi(wchar2_t* out, size_t capacity, const wchar2_t* format,
								const wchar2_t* first, const wchar2_t* second) {
	format_bounded_wide(out, capacity, format, first, second);
}
