package legacy

/*
#include "GAME5_2.h"
*/
import "C"

import (
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

func clientGlyphType57B400(offset uintptr) uint32 {
	cache := memmap.PtrUint32(0x5D4594, offset)
	if *cache == 0 {
		*cache = uint32(GetClient().Cli().Things.IndByID("Glyph"))
	}
	return *cache
}

//export nox_xxx_client_57B400
func nox_xxx_client_57B400(drawable *C.nox_drawable) C.int {
	glyph := clientGlyphType57B400(2523876)
	p := Get_dword_8531A0_2576()
	if p == nil {
		return 0
	}
	return C.int(bool2int(asDrawable(drawable).TypeIDVal != glyph || p.PlayerClass() == player.Wizard))
}

//export sub_57B450
func sub_57B450(drawable *C.nox_drawable) C.int {
	return C.int(clientDrawableCanPickup57B450(asDrawable(drawable)))
}

func clientDrawableCanPickup57B450(d *client.Drawable) int {
	glyph := clientGlyphType57B400(2523880)
	if d == nil || *memmap.PtrPtr(0x852978, 8) == nil {
		return 0
	}
	p := Get_dword_8531A0_2576()
	if p == nil || (d.TypeIDVal == glyph && p.PlayerClass() != player.Wizard) {
		return 0
	}
	if cheatEquipAll.Load() {
		return 1
	}
	// The original x86 byte shift uses the low five bits of the class.
	classBit := byte(1 << (p.PlayerClass() & 31))
	allowed := Sub_57B370(d.ObjClass, d.ObjSubClass, int(int32(d.TypeIDVal)))
	return bool2int(classBit&allowed != 0)
}
