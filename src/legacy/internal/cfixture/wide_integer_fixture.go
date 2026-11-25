package cfixture

/*
#include "noxstring.h"
#include <errno.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

struct nox_wide_integer_result {
	int64_t value;
	uintptr_t consumed;
	int error;
};

static void nox_compare_wide_integer(const wchar2_t* text, int base, bool with_end,
                                    struct nox_wide_integer_result* got,
                                    struct nox_wide_integer_result* want) {
	int saved = errno;
	errno = EDOM;
	size_t length = 0;
	while (text[length] != 0) {
		length++;
	}
	char* ascii = calloc(length + 1, 1);
	if (ascii == NULL) {
		abort();
	}
	for (size_t i = 0; i < length; i++) {
		ascii[i] = text[i] < 0x80 ? (char)text[i] : '\x7f';
	}
	char* narrow_end = ascii;
	want->value = strtol(ascii, with_end ? &narrow_end : NULL, base);
	want->consumed = (uintptr_t)narrow_end - (uintptr_t)ascii;
	free(ascii);
	want->error = errno;

	wchar2_t* wide_end = (wchar2_t*)text;
	errno = EDOM;
	got->value = nox_wcstol(text, with_end ? &wide_end : NULL, base);
	got->consumed = ((uintptr_t)wide_end - (uintptr_t)text) / sizeof(wchar2_t);
	got->error = errno;
	errno = saved;
}
*/
import "C"

import "unsafe"

type WideIntegerResult struct {
	Value    int64
	Consumed uintptr
	Errno    int
}

func CompareWideInteger(text unsafe.Pointer, base int, withEnd bool) (actual, reference WideIntegerResult) {
	var got, want C.struct_nox_wide_integer_result
	C.nox_compare_wide_integer((*C.wchar2_t)(text), C.int(base), C.bool(withEnd), &got, &want)
	return WideIntegerResult{int64(got.value), uintptr(got.consumed), int(got.error)},
		WideIntegerResult{int64(want.value), uintptr(want.consumed), int(want.error)}
}
