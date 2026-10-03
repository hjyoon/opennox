package legacy

import "unsafe"

const itemEnchantmentTooltipSource413480 = `C:\NoxPost\src\common\Object\Modifier.c`

type itemEnchantmentTooltipHooks413480 struct {
	loadFlag func(int) byte
	loadKey  func(int) *byte
	loadText func(*byte, string, int32) unsafe.Pointer
}

// itemEnchantmentTooltip413480 reads the live flag bytes in original table
// order. Only a complete byte match reads the key and calls the string loader.
func itemEnchantmentTooltip413480(flag byte, h itemEnchantmentTooltipHooks413480) unsafe.Pointer {
	for i := 0; i < 6; i++ {
		if h.loadFlag(i) == flag {
			return h.loadText(h.loadKey(i), itemEnchantmentTooltipSource413480, 2087)
		}
	}
	return nil
}
