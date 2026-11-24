package legacy

/*
#include "GAME3_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_xxx_prepareLightningEffects_4BAB30
func nox_xxx_prepareLightningEffects_4BAB30() C.int {
	for _, color := range [...]struct {
		offset  uintptr
		r, g, b int32
	}{
		{1316464, 255, 255, 255},
		{1316488, 128, 128, 255},
		{1316424, 128, 128, 255},
		{1316428, 64, 64, 255},
		{1316516, 200, 200, 255},
		{1316512, 128, 128, 255},
		{1316496, 255, 255, 255},
		{1316468, 255, 255, 0},
		{1316460, 30, 160, 30},
		{1316444, 60, 140, 60},
		{1316504, 40, 225, 40},
		{1316480, 150, 220, 150},
	} {
		*memmap.PtrUint32(0x5D4594, color.offset) = uint32(nox_color_rgb_4344A0(color.r, color.g, color.b))
	}
	var id int32
	for i, name := range [...]string{"BlueSpark", "YellowSpark", "GreenSpark"} {
		id = nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr(name))
		*memmap.PtrUint32(0x5D4594, 1316520+4*uintptr(i)) = uint32(id)
	}
	return C.int(id)
}

//export sub_4BC720
func sub_4BC720(dr *nox_drawable) unsafe.Pointer {
	return asDrawable(dr).PrepareMonsterGeneratorStatus()
}

//export nox_xxx_updDrawMonsterGen_4BC920
func nox_xxx_updDrawMonsterGen_4BC920() C.int { return 1 }

//export nox_xxx_updDrawUndeadKiller_4CCCF0
func nox_xxx_updDrawUndeadKiller_4CCCF0() C.int { return 1 }
