package legacy

/*
#include "GAME4_3.h"

typedef int (*pixie_cast_fn)(int, nox_object_t*, nox_object_t*, nox_object_t*, nox_spell_accept_arg_t*, int);
_Static_assert(_Generic(&nox_xxx_castPixies_540440, pixie_cast_fn: 1, default: 0), "Pixie cast ABI");

static int pixie_cast_native_fixture(int id, nox_object_t* source, nox_object_t* owner, nox_object_t* caster, nox_spell_accept_arg_t* args, int level) {
	pixie_cast_fn volatile cast = nox_xxx_castPixies_540440;
	return cast(id, source, owner, caster, args, level);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func pixieCastStockFixture540440(id int32, source, owner, caster *server.Object, args *server.SpellAcceptArg, level int32) int {
	return int(C.pixie_cast_native_fixture(C.int(id), asObjectC(source), asObjectC(owner), asObjectC(caster),
		(*C.nox_spell_accept_arg_t)(unsafe.Pointer(args)), C.int(level)))
}
