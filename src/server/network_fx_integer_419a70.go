package server

import (
	"github.com/opennox/libs/types"
	"math"
)

// netFXTruncateSignedDWORD419A70 preserves the original startup FISTP
// signed-DWORD conversion before packet fields retain its low WORD/BYTE.
// Invalid operands produce integer-indefinite, whose low WORD/BYTE is zero.
func netFXTruncateSignedDWORD419A70(value float32) int32 {
	if !(value >= -0x1p31 && value < 0x1p31) {
		return math.MinInt32
	}
	return int32(value)
}

// netFXPointSignedDWORD419A70 preserves the original signed-DWORD endpoints
// used by both the ray packet fields and its viewport filter.
func netFXPointSignedDWORD419A70(pos types.Pointf) types.Point {
	return types.Point{
		X: int(netFXTruncateSignedDWORD419A70(pos.X)),
		Y: int(netFXTruncateSignedDWORD419A70(pos.Y)),
	}
}
