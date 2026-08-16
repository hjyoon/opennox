package legacy

/*
#include "glyph_drop_4ed500.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_GlyphDrop_4ED500
func nox_GlyphDrop_4ED500(
	owner, glyph *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(glyphDropCall4ED500(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(glyph)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
