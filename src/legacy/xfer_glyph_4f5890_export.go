package legacy

/*
#include "xfer_glyph_4f5890.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var glyphXferCall4F5890 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
	_ unsafe.Pointer,
) int32 {
	return Nox_xxx_XFerGlyphNative4F5890(cf, object)
}

func glyphXferExportCall4F5890(
	object *server.Object,
	context unsafe.Pointer,
) int32 {
	return int32(C.nox_xxx_XFerGlyph_4F5890(asObjectC(object), context))
}

//export nox_xxx_XFerGlyph_4F5890
func nox_xxx_XFerGlyph_4F5890(
	object *C.nox_object_t,
	context unsafe.Pointer,
) C.int32_t {
	return C.int32_t(glyphXferCall4F5890(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
		context,
	))
}
