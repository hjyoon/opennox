package legacy

/*
#include "GAME2_1.h"
extern uintptr_t array_5D4594_1049872[9];
extern uint32_t dword_5d4594_1049864;
extern nox_window* dword_5d4594_1062452;
// Address bridges only; the existing owners remain shared with C callers.
static void* inventory_state_slot_C(unsigned index) {
 switch(index) {
 case 0:return &array_5D4594_1049872[0];
 case 1:return &dword_5d4594_1049864;
 case 2:return &dword_5d4594_1062452;
 default:return NULL;
 }
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
)

func inventoryStateNativeSlot(index int) unsafe.Pointer {
	return C.inventory_state_slot_C(C.uint(index))
}

func inventoryStateNativeHeads() *[9]*client.Drawable {
	return (*[9]*client.Drawable)(inventoryStateNativeSlot(0))
}

func inventoryStateFindType461600(thingType uint32) *client.Drawable {
	for _, head := range inventoryStateNativeHeads() {
		for drawable := head; drawable != nil; drawable = drawable.NextPtr {
			if drawable.TypeIDVal == thingType {
				return drawable
			}
		}
	}
	return nil
}

func inventoryStateHasClass461930() int32 {
	for _, head := range inventoryStateNativeHeads() {
		for drawable := head; drawable != nil; drawable = drawable.NextPtr {
			if uint32(drawable.ObjClass)&0x01001000 != 0 {
				return 1
			}
		}
	}
	return 0
}

func inventoryStateBounds464B40(column, row int32) int32 {
	if column >= 0 && column < 4 && row >= 0 && row < 21 {
		return 1
	}
	return 0
}

func inventoryStateEquippedZone465990(point ntype.Point32) int32 {
	point.X -= 11
	point.Y -= 15
	heads := inventoryStateNativeHeads()
	zones := memmap.PtrT[[9][4]int32](0x587000, 136192)
	for slot := range zones {
		if guiGeometryPointInRect4281F0(&point, &zones[slot]) == 0 {
			continue
		}
		switch {
		case slot == 0 && heads[0] == nil:
			continue
		case slot == 6:
			for item := heads[8]; item != nil; item = item.NextPtr {
				if item.Class().Has(object.ClassArmor) && object.ArmorClass(item.SubClass()).Has(object.ArmorShield) {
					return 8
				}
			}
			return 5
		default:
			return int32(slot)
		}
	}
	return -1
}

func inventoryStateReset467680() {
	state := (*uint32)(inventoryStateNativeSlot(1))
	if *state == 6 {
		*state = 0
	}
}

func inventoryStateSet467740(value int32) int32 {
	*globalStateNativeWord(4) = uint32(value)
	return value
}
