package legacy

/*
#include "GAME2_1.h"
extern uint32_t nox_xxx_minimap_587000_149232;
*/
import "C"

import "github.com/opennox/opennox/v1/client"

var (
	Sub_473670                  func() int
	Nox_client_toggleMap_473610 func() int
)

//export sub_473670
func sub_473670() int32 { return int32(Sub_473670()) }

//export nox_client_toggleMap_473610
func nox_client_toggleMap_473610() C.char { return C.char(Nox_client_toggleMap_473610()) }

func DrawMinimap4Sprite4725C0(drawable *client.Drawable) {
	C.nox_xxx_drawMinimap4Sprite_4725C0((*C.nox_drawable)(drawable.C()))
}

func SetMinimapZoom(zoom uint32) {
	C.nox_xxx_minimap_587000_149232 = C.uint32_t(zoom)
}
