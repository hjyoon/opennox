package legacy

/*
#include <stdint.h>

#include "GAME4.h"

extern char* dword_5d4594_1599576;
extern uint32_t dword_5d4594_1599596;

typedef int (*nox_mapgen_lookup_5029A0_fn)(char*);
typedef char* (*nox_mapgen_entry_5029F0_fn)(int);

_Static_assert(_Generic(&sub_5029A0, nox_mapgen_lookup_5029A0_fn: 1, default: 0),
	"mapgen name lookup must return a signed index");
_Static_assert(_Generic(&sub_5029F0, nox_mapgen_entry_5029F0_fn: 1, default: 0),
	"mapgen record access must return a native pointer");
_Static_assert(sizeof(dword_5d4594_1599576) == sizeof(void*),
	"mapgen record storage must preserve native pointer width");

static int nox_test_mapgen_record_lookup_5029A0(uintptr_t* address) {
	char records[4][76] = {"Alpha", "Beta", "Gamma", ""};
	char* old_records = dword_5d4594_1599576;
	uint32_t old_count = dword_5d4594_1599596;
	dword_5d4594_1599576 = &records[0][0];
	dword_5d4594_1599596 = 3;
	*address = (uintptr_t)dword_5d4594_1599576;
	int ok = sub_5029A0("alpha") == 0 &&
		sub_5029A0("BETA") == 1 &&
		sub_5029A0("gamma") == 2 &&
		sub_5029A0("missing") == -1 &&
		sub_5029F0(-1) == 0 &&
		sub_5029F0(0) == &records[0][0] &&
		sub_5029F0(2) == &records[2][0] &&
		sub_5029F0(3) == &records[3][0] &&
		sub_5029F0(4) == 0;
	dword_5d4594_1599596 = UINT32_MAX;
	ok = ok && sub_5029A0("Alpha") == -1;
	dword_5d4594_1599576 = old_records;
	dword_5d4594_1599596 = old_count;
	return ok;
}
*/
import "C"

func mapgenRecordLookupContract5029A0() (uintptr, bool) {
	var address C.uintptr_t
	ok := C.nox_test_mapgen_record_lookup_5029A0(&address) != 0
	return uintptr(address), ok
}
