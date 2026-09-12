package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <string.h>

#include "GAME4.h"
#include "memmap.h"

extern void* dword_5d4594_1599588;
extern void* dword_5d4594_1599592;
extern uint32_t dword_5d4594_1599596;
extern uint32_t dword_5d4594_3835396;
extern FILE* nox_file_8;

typedef int (*nox_mapgen_count_502A20_fn)(void);
typedef int (*nox_mapgen_save_name_502A30_fn)(char*);
typedef int (*nox_mapgen_set_name_502A50_fn)(char*);
typedef char* (*nox_mapgen_get_name_502A90_fn)(void);

_Static_assert(_Generic(&sub_502A20, nox_mapgen_count_502A20_fn: 1, default: 0),
	"mapgen record count must be a signed dword");
_Static_assert(_Generic(&sub_502A30, nox_mapgen_save_name_502A30_fn: 1, default: 0),
	"mapgen save-by-name must take a C string");
_Static_assert(_Generic(&sub_502A50, nox_mapgen_set_name_502A50_fn: 1, default: 0),
	"the first mapgen name setter must take a C string");
_Static_assert(_Generic(&sub_502AB0, nox_mapgen_set_name_502A50_fn: 1, default: 0),
	"the second mapgen name setter must take a C string");
_Static_assert(_Generic(&sub_502A90, nox_mapgen_get_name_502A90_fn: 1, default: 0),
	"the first mapgen name getter must return a native pointer");
_Static_assert(_Generic(&sub_502AF0, nox_mapgen_get_name_502A90_fn: 1, default: 0),
	"the second mapgen name getter must return a native pointer");

static int nox_test_mapgen_buffers_502A20(uintptr_t* first_address, uintptr_t* second_address) {
	char first[2048] = {0};
	char second[2048] = {0};
	char long_name[2049];
	void* old_first = dword_5d4594_1599588;
	void* old_second = dword_5d4594_1599592;
	uint32_t old_count = dword_5d4594_1599596;
	uint32_t old_selected = dword_5d4594_3835396;
	FILE* old_file = nox_file_8;
	uint8_t* first_default = getMemU8Ptr(0x5D4594, 1599608);
	uint8_t* second_default = getMemU8Ptr(0x5D4594, 1599612);
	uint8_t old_first_default = *first_default;
	uint8_t old_second_default = *second_default;
	int ok = 1;

	dword_5d4594_1599588 = first;
	dword_5d4594_1599592 = second;
	dword_5d4594_1599596 = 0;
	dword_5d4594_3835396 = 123;
	nox_file_8 = NULL;
	*first_default = 'A';
	*second_default = 'B';
	*first_address = (uintptr_t)first;
	*second_address = (uintptr_t)second;

	ok = ok && sub_502A20() == 0 && sub_502A30("absent") == 0 &&
		dword_5d4594_3835396 == 123;
	dword_5d4594_1599596 = UINT32_MAX;
	ok = ok && sub_502A20() == -1;
	dword_5d4594_1599596 = 0;
	ok = ok && sub_502A90() == NULL && sub_502AF0() == NULL;
	ok = ok && sub_502A50("FirstMap") == 1 && sub_502AB0("SecondMap") == 1;
	ok = ok && sub_502A90() == first && sub_502AF0() == second &&
		strcmp(first, "FirstMap") == 0 && strcmp(second, "SecondMap") == 0;
	ok = ok && sub_502A50(NULL) == 0 && sub_502AB0(NULL) == 0 &&
		first[0] == 'A' && second[0] == 'B' &&
		first[1] == 'i' && second[1] == 'e';

	memset(long_name, 'L', 2048);
	long_name[2048] = 0;
	first[2047] = 'Z';
	second[2047] = 'Y';
	ok = ok && sub_502A50(long_name) == 1 && sub_502AB0(long_name) == 1 &&
		memcmp(first, long_name, 2047) == 0 && memcmp(second, long_name, 2047) == 0 &&
		first[2047] == 'Z' && second[2047] == 'Y';

	*first_default = old_first_default;
	*second_default = old_second_default;
	nox_file_8 = old_file;
	dword_5d4594_1599588 = old_first;
	dword_5d4594_1599592 = old_second;
	dword_5d4594_1599596 = old_count;
	dword_5d4594_3835396 = old_selected;
	return ok;
}
*/
import "C"

func mapgenBufferContract502A20() (uintptr, uintptr, bool) {
	var first, second C.uintptr_t
	ok := C.nox_test_mapgen_buffers_502A20(&first, &second) != 0
	return uintptr(first), uintptr(second), ok
}
