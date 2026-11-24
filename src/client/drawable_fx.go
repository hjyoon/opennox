package client

import (
	"image"
	"math"
	"unsafe"
)

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

func (d *Drawable) SetTurnUndeadEffect(direction byte, pos image.Point, cosine, sine float32, frame uint32, update unsafe.Pointer) {
	d.Field_127 = d.Field_127&0xffff0000 | uint32(direction)
	d.Field_117 = math.Float32bits(cosine * 4)
	d.Field_118 = math.Float32bits(sine * 4)
	d.Field_119 = 0
	d.AnimStart = frame
	d.Field_81, d.Field_82 = uint32(pos.X), uint32(pos.Y)
	d.Field_115 = update
}
