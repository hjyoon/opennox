package legacy

/*
#include "GAME4_3.h"

typedef char* (*nox_action_name_getter_t)(int);
typedef int (*nox_action_name_lookup_t)(const char*);
_Static_assert(_Generic(&sub_5345B0, nox_action_name_getter_t: 1, default: 0), "selected action name pointer ABI");
_Static_assert(_Generic(&sub_534650, nox_action_name_getter_t: 1, default: 0), "action name pointer ABI");
_Static_assert(_Generic(&nox_xxx_actionNByNameMB_5345F0, nox_action_name_lookup_t: 1, default: 0), "const short lookup ABI");
_Static_assert(_Generic(&nox_xxx_actionByName_534670, nox_action_name_lookup_t: 1, default: 0), "const full lookup ABI");

static char* nox_action_names_fixture_get(int selected, int value) {
	return selected ? sub_5345B0(value) : sub_534650(value);
}
static int nox_action_names_fixture_lookup(int short_table, const char* value) {
	return short_table ? nox_xxx_actionNByNameMB_5345F0(value) : nox_xxx_actionByName_534670(value);
}
*/
import "C"

import "unsafe"

func monsterActionNamesGetC(selected bool, value uint32) unsafe.Pointer {
	var cselected C.int
	if selected {
		cselected = 1
	}
	return unsafe.Pointer(C.nox_action_names_fixture_get(cselected, C.int(value)))
}

func monsterActionNamesLookupC(short bool, value unsafe.Pointer) int32 {
	var cshort C.int
	if short {
		cshort = 1
	}
	return int32(C.nox_action_names_fixture_lookup(cshort, (*C.char)(value)))
}
