package legacy

/*
#include "audio_ambient_native_45a9b0.h"
_Static_assert(_Generic(&sub_45A9B0, void (*)(nox_drawable*, nox_drawable*): 1, default: 0), "native positional loop ABI");
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client"
	"unsafe"
)

var UpdateDrawableAudioLoop45A9B0 func(source, listener *client.Drawable)

//export sub_45A9B0
func sub_45A9B0(source, listener *C.nox_drawable) {
	if UpdateDrawableAudioLoop45A9B0 != nil {
		UpdateDrawableAudioLoop45A9B0((*client.Drawable)(unsafe.Pointer(source)), (*client.Drawable)(unsafe.Pointer(listener)))
	}
}
