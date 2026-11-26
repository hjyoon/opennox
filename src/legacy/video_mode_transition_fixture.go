//go:build !server

package legacy

/*
#include "defs.h"
int nox_getBackbufferPitch(void);
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

// Test access to the live render-data slot and the C pitch entry.
var (
	VideoModeRenderDataC = (**noxrender.RenderData)(unsafe.Pointer(&currentDrawData))
	VideoModePitchCEntry = C.nox_getBackbufferPitch
)
