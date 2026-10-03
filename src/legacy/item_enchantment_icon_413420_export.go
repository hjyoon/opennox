package legacy

/*
#include "GAME1.h"
#include "GAME1_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var itemEnchantmentIconLoad413420 = func(name *byte) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_gLoadImg_42F970((*C.char)(unsafe.Pointer(name))))
}

func itemEnchantmentIconNativeHooks413420() itemEnchantmentIconHooks413420 {
	return itemEnchantmentIconHooks413420{
		loadLoaded:  func() uint32 { return memmap.Uint32(0x5D4594, 251624) },
		storeLoaded: func(v uint32) { *memmap.PtrUint32(0x5D4594, 251624) = v },
		loadName: func(i int) *byte {
			return (*byte)(*memmap.PtrPtr(0x587000, 27336+20*uintptr(i)))
		},
		loadImage: itemEnchantmentIconLoad413420,
		storeImage: func(i int, image unsafe.Pointer) {
			*memmap.PtrPtr(0x587000, 27340+20*uintptr(i)) = image
		},
		loadFlag: func(i int) byte { return memmap.Uint8(0x587000, 27332+20*uintptr(i)) },
		loadCachedImage: func(i int) unsafe.Pointer {
			return *memmap.PtrPtr(0x587000, 27340+20*uintptr(i))
		},
	}
}

var itemEnchantmentIconCall413420 = func(flag byte) unsafe.Pointer {
	return itemEnchantmentIcon413420(flag, itemEnchantmentIconNativeHooks413420())
}

//export nox_item_enchantment_icon_native_413420
func nox_item_enchantment_icon_native_413420(flag C.uchar) *nox_video_bag_image_t {
	return (*nox_video_bag_image_t)(itemEnchantmentIconCall413420(byte(flag)))
}

func itemEnchantmentIconCEntry413420(flag byte) unsafe.Pointer {
	return unsafe.Pointer(C.sub_413420(C.char(flag)))
}
