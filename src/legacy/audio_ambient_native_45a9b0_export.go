package legacy

/*
#include "defs.h"
#include "audio_ambient_native_45a9b0.h"
void sub_45A9B0(nox_drawable* source, nox_drawable* listener);
_Static_assert(_Generic(&sub_45A9B0, void (*)(nox_drawable*, nox_drawable*): 1, default: 0), "native positional loop ABI");
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client"
	"unsafe"
)

var UpdateDrawableAudioLoop45A9B0 func(source, listener *client.Drawable)

//export nox_client_updateAmbient_native_45A9B0
func nox_client_updateAmbient_native_45A9B0(source, listener unsafe.Pointer) {
	if UpdateDrawableAudioLoop45A9B0 != nil {
		UpdateDrawableAudioLoop45A9B0((*client.Drawable)(source), (*client.Drawable)(listener))
	}
}
