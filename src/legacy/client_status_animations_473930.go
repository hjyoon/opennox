package legacy

/*
#include "GAME2_1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

//export sub_473930
func sub_473930() *C.char {
	// GAME.EXE 00473930 publishes the first reference before loading the
	// second and returns the second reference. Packed PE32 globals keep their
	// four-byte layout; native builds store the full addresses in side slots.
	*memmap.PtrPtr(0x5D4594, 1096456) = Nox_xxx_gLoadAnim("ConfusedBirdies").C()
	shield := Nox_xxx_gLoadAnim("SphericalShieldAnim").C()
	*memmap.PtrPtr(0x5D4594, 1096460) = shield
	return (*C.char)(shield)
}
