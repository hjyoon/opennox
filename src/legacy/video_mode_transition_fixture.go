//go:build !server

package legacy

/*
#include "defs.h"
extern nox_render_data_t* nox_draw_curDrawData_3799572;
int nox_getBackbufferPitch(void);
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

// Declaration-only test access to the real C globals and entry. These aliases
// neither replace a video callback nor install a test row/pixel buffer.
var (
	VideoModeRenderDataC = (**noxrender.RenderData)(unsafe.Pointer(&C.nox_draw_curDrawData_3799572))
	VideoModePitchCEntry = C.nox_getBackbufferPitch
)
