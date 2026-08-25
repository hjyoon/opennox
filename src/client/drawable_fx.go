package client

import "unsafe"

type DrawableFX struct {
	Field0     uint32
	Field4     uint32
	Trail      [6][2]int32
	Count      uint8
	_          [3]byte
	Owner      *Drawable
	Next       *DrawableFX
	Prev       *DrawableFX
	GlobalNext *DrawableFX
	GlobalPrev *DrawableFX
}

func (fx *DrawableFX) C() unsafe.Pointer {
	return unsafe.Pointer(fx)
}
