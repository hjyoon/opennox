// Compile-only native-width contract for FieldGuide transfer restored from
// GAME.EXE 004F6390.
#define _Static_assert(...)
#include "../GAME4.h"
#undef _Static_assert

#include <stddef.h>

typedef int (*field_guide_xfer_fn)(nox_object_t*);

_Static_assert(sizeof(nox_field_guide_use_data_t) == 64,
	"FieldGuide use-data size");
_Static_assert(offsetof(nox_field_guide_use_data_t, creature) == 0,
	"FieldGuide creature-name offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object field-34 offset");
_Static_assert(offsetof(nox_object_t, use_data) == (sizeof(void*) == 4 ? 736 : 848),
	"object use-data offset");
_Static_assert(
	_Generic(&nox_xxx_XFerFieldGuide_4F6390,
		field_guide_xfer_fn: 1, default: 0),
	"FieldGuide xfer signature");

int main(void) {
	return 0;
}
