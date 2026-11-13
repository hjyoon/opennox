package server

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
)

// These entries share the already restored numeric kernels with the ordinary
// collision path. The four-float bounds retain their in-place ABI layout.
func (s *Server) BoxWallAt550580(grid image.Point, obj *Object, coefficient func() float32, addCollision func(*Object, types.Pointf)) bool {
	return boxWallCollisionAt550580(grid, obj, boxWallCollisionHooks5504B0{
		spring: coefficient, wallType: s.Sub_57B500,
		wallDef: func(kind byte) boxWallDef5504B0 {
			off := uintptr(kind) * 24
			return boxWallDef5504B0{
				firstEnabled: memmap.Uint8(0x587000, 292520+off), firstMask: memmap.Uint8(0x587000, 292521+off),
				firstOffset: memmap.Float32(0x587000, 292524+off), firstLength: memmap.Float32(0x587000, 292528+off),
				secondEnabled: memmap.Uint8(0x587000, 292532+off), secondMask: memmap.Uint8(0x587000, 292533+off),
				secondOffset: memmap.Float32(0x587000, 292536+off), secondLength: memmap.Float32(0x587000, 292540+off),
			}
		}, addCollision: addCollision,
	})
}

func BoxWallSegment550760(first bool, obj *Object, current, previous *types.Pointf, bounds *[4]float32, line types.Pointf, length float32, coefficient func() float32, addCollision func(*Object, types.Pointf)) bool {
	b := (*boxWallBounds5504B0)(unsafe.Pointer(bounds))
	if first {
		return boxWallCollisionFirst550A10(obj, current, previous, b, line, length, coefficient, addCollision)
	}
	return boxWallCollisionSecond550760(obj, current, previous, b, line, length, coefficient, addCollision)
}

func BoxWallQuadrant550CB0(point, wall types.Pointf) byte { return boxWallQuadrant550CB0(point, wall) }

func game5DoorSegment(raw [4]float32) doorBoxSegment551250 {
	return doorBoxSegment551250{start: types.Ptf(raw[0], raw[1]), end: types.Ptf(raw[2], raw[3])}
}

func game5DoorRect(raw [4]float32) doorBoxRect551250 {
	return doorBoxRect551250{minX: raw[0], minY: raw[1], maxX: raw[2], maxY: raw[3]}
}

func DoorBoxAxis551780(vertical bool, segment [4]float32, edge1, edge2, fixed float32, allowNegative bool) (types.Pointf, bool) {
	return doorBoxAxisIntersection551780(game5DoorSegment(segment), edge1, edge2, fixed, allowNegative, vertical)
}

func DoorBoxIntersections5516A0(segment, rect [4]float32, limit int, allowNegative bool) []types.Pointf {
	return doorBoxSegmentRectIntersections5516A0(game5DoorSegment(segment), game5DoorRect(rect), limit, allowNegative)
}

func DoorBoxMidpoint551960(segment, segmentBounds, boxBounds [4]float32) (types.Pointf, bool) {
	return doorBoxCollisionMidpoint551960(game5DoorSegment(segment), game5DoorRect(segmentBounds), game5DoorRect(boxBounds))
}

func DoorBoxPoint551A90(point types.Pointf, rect [4]float32) bool {
	return doorBoxPointInRect551A90(point, game5DoorRect(rect))
}
