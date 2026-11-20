package legacy

import "github.com/opennox/opennox/v1/common/memmap"

func trigInitNative414C90() byte {
	flag := memmap.PtrUint8(0x5D4594, 371240)
	if *flag != 0 {
		return *flag
	}
	*flag = 1
	step := float64(*memmap.PtrFloat32(0x581450, 7208))
	sineScale := *memmap.PtrFloat64(0x581450, 7200)
	for i := 0; i < 4096; i++ {
		v := trigSin414C90(float64(i) * step)
		*memmap.PtrUint32(0x85B3FC, 12260+uintptr(i)*4) = uint32(int64(mulFloat64TowardZero(v, sineScale)))
	}
	step = float64(*memmap.PtrFloat32(0x581450, 7196))
	subtract := float64(*memmap.PtrFloat32(0x581450, 7192))
	scale := *memmap.PtrFloat64(0x581450, 7184)
	var last int64
	for i := 0; i < 8192; i++ {
		product := float64(float64(i) * step)
		v := trigAcos414C90(product - subtract)
		last = int64(mulFloat64TowardZero(v, scale))
		*memmap.PtrUint32(0x5D4594, 338472+uintptr(i)*4) = uint32(last)
	}
	return byte(last)
}
