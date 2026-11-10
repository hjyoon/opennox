package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"
import (
	"github.com/opennox/opennox/v1/client"
	"unsafe"
)

//export sub_459DB0
func sub_459DB0(p *C.nox_drawable) C.int {
	return g2ServerBool(p != nil && p.flags28&0x400000 != 0 && p.flags29&8 != 0)
}

//export nox_xxx_cliNextMinimapObj_459EC0
func nox_xxx_cliNextMinimapObj_459EC0(p *C.nox_drawable) *C.nox_drawable {
	if p == nil {
		return nil
	}
	next := p.field_102
	if p == next {
		C.abort()
		return nil
	}
	return next
}

//export sub_45A010
func sub_45A010(p *C.nox_drawable) *C.nox_drawable {
	return (*C.nox_drawable)((*client.Drawable)(unsafe.Pointer(p)).Field104().C())
}

//export nox_drawable_next_45A070
func nox_drawable_next_45A070(p *C.nox_drawable) *C.nox_drawable {
	return (*C.nox_drawable)((*client.Drawable)(unsafe.Pointer(p)).Next().C())
}

//export nox_xxx_spriteSetActiveMB_45A990_drawable
func nox_xxx_spriteSetActiveMB_45A990_drawable(p *C.nox_drawable) *C.nox_drawable {
	(*client.Drawable)(unsafe.Pointer(p)).SetActive()
	return p
}

//export nox_xxx_spriteSetFrameMB_45AB80_drawable
func nox_xxx_spriteSetFrameMB_45AB80_drawable(p *C.nox_drawable, frame C.int) *C.nox_drawable {
	(*client.Drawable)(unsafe.Pointer(p)).SetFrameMB(int(frame))
	return p
}

//export nox_xxx_spriteSetFrameMB_45AB80
func nox_xxx_spriteSetFrameMB_45AB80(p *C.nox_drawable, frame C.int) *C.nox_drawable {
	return nox_xxx_spriteSetFrameMB_45AB80_drawable(p, frame)
}
