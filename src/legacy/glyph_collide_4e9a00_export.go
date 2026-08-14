package legacy

/*
#include "glyph_collide_4e9a00.h"
*/
import "C"

import "unsafe"

//export nox_xxx_collideGlyph_4E9A00
func nox_xxx_collideGlyph_4E9A00(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	Nox_xxx_collideGlyph_4E9A00(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}

//export sub_4E9A30
func sub_4E9A30(source, target *C.nox_object_t) C.int {
	return C.int(GetServer().S().GlyphCollideAllowed4E9A30(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
	))
}
