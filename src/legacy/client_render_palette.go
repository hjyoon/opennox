package legacy

/*
#include "client_render_palette.h"
*/
import "C"

import "github.com/opennox/opennox/v1/legacy/common/alloc"

// The renderer publishes addresses of these color cells for the process lifetime.
var renderPalette, _ = alloc.New(C.nox_render_palette{})

//export nox_render_palette_storage
func nox_render_palette_storage() *C.nox_render_palette {
	return renderPalette
}
