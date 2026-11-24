package legacy

/*
#include "nox_wchar.h"

void nox_format_wide_strings_abi(wchar2_t* out, size_t capacity, const wchar2_t* format,
                               const wchar2_t* first, const wchar2_t* second);
*/
import "C"

func formatWideStrings(out []wchar2_t, format, first, second *wchar2_t) {
	C.nox_format_wide_strings_abi(&out[0], C.size_t(len(out)), format, first, second)
}
