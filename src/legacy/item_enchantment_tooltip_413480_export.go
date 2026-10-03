package legacy

/*
#include "common__object__modifier.h"
#include "common__strman.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var itemEnchantmentTooltipLoad413480 = func(key *byte, source string, line int32) unsafe.Pointer {
	return unsafe.Pointer(C.nox_strman_loadString_40F1D0(
		(*C.char)(unsafe.Pointer(key)), nil, internCStr(source), C.int(line),
	))
}

func itemEnchantmentTooltipNativeHooks413480() itemEnchantmentTooltipHooks413480 {
	return itemEnchantmentTooltipHooks413480{
		loadFlag: func(i int) byte { return memmap.Uint8(0x587000, 27332+20*uintptr(i)) },
		loadKey: func(i int) *byte {
			return (*byte)(*memmap.PtrPtr(0x587000, 27344+20*uintptr(i)))
		},
		loadText: itemEnchantmentTooltipLoad413480,
	}
}

var itemEnchantmentTooltipCall413480 = func(flag byte) unsafe.Pointer {
	return itemEnchantmentTooltip413480(flag, itemEnchantmentTooltipNativeHooks413480())
}

//export nox_item_enchantment_tooltip_native_413480
func nox_item_enchantment_tooltip_native_413480(flag C.uchar) *wchar2_t {
	return (*wchar2_t)(itemEnchantmentTooltipCall413480(byte(flag)))
}

func itemEnchantmentTooltipCEntry413480(flag byte) unsafe.Pointer {
	return unsafe.Pointer(C.sub_413480(C.char(flag)))
}
