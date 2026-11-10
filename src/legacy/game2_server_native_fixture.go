package legacy

/*
#include "game2_server_bridge.h"
typedef nox_drawable* (*game2_server_native_frame_sig)(nox_drawable*,int);
_Static_assert(_Generic(&nox_xxx_spriteSetFrameMB_45AB80,game2_server_native_frame_sig:1,default:0),"legacy frame input and return must be full native pointers");
_Static_assert(_Generic(&nox_xxx_spriteSetFrameMB_45AB80_drawable,game2_server_native_frame_sig:1,default:0),"typed frame input and return must be full native pointers");
*/
import "C"
import "unsafe"

// Only install with the new native algorithms, after the before-C control.
func game2ServerFixtureLegacyFrame(p unsafe.Pointer, frame int32) unsafe.Pointer {
	return unsafe.Pointer(nox_xxx_spriteSetFrameMB_45AB80((*C.nox_drawable)(p), C.int(frame)))
}
