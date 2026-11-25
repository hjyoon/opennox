#include "noxstring_format_abi.h"

extern void nox_format_bounded_wide_go(wchar2_t* out, size_t capacity, nox_format_word* format, nox_format_args* args);

static void format_bounded_wide(wchar2_t* out, size_t capacity, nox_format_word* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	nox_format_bounded_wide_go(out, capacity, format, &args);
	va_end(args.list);
}

void nox_format_wide_strings_abi(wchar2_t* out, size_t capacity, const wchar2_t* format,
								const wchar2_t* first, const wchar2_t* second) {
	format_bounded_wide(out, capacity, format, first, second);
}
