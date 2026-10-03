package legacy

import "unsafe"

type itemEnchantmentIconHooks413420 struct {
	loadLoaded      func() uint32
	storeLoaded     func(uint32)
	loadName        func(int) *byte
	loadImage       func(*byte) unsafe.Pointer
	storeImage      func(int, unsafe.Pointer)
	loadFlag        func(int) byte
	loadCachedImage func(int) unsafe.Pointer
}

// itemEnchantmentIcon413420 preserves 00413420..0041347B: load all six
// images in table order before any lookup, mark the cache ready even when a
// loader returns nil, and match the complete flag byte rather than a bit set.
func itemEnchantmentIcon413420(flag byte, h itemEnchantmentIconHooks413420) unsafe.Pointer {
	if h.loadLoaded() == 0 {
		for i := 0; i < 6; i++ {
			h.storeImage(i, h.loadImage(h.loadName(i)))
		}
		h.storeLoaded(1)
	}
	for i := 0; i < 6; i++ {
		if h.loadFlag(i) == flag {
			return h.loadCachedImage(i)
		}
	}
	return nil
}
