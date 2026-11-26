package legacy

/*
#include "GAME2_1.h"

typedef const nox_draw_viewport_t wall_const_viewport_t;
typedef const unsigned char wall_const_byte_t;
*/
import "C"

import (
	"image"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func clientWallScreenPosition(view *C.nox_draw_viewport_t, x, y byte) (int32, int32) {
	return int32(view.x1) + 23*int32(x) - int32(view.field_4),
		int32(view.y1) + 23*int32(y) - int32(view.field_5)
}

//export nox_client_wall_screen_position_473C10
func nox_client_wall_screen_position_473C10(view *C.wall_const_viewport_t, data *C.wall_const_byte_t, x, y *C.int) {
	wall := unsafe.Slice((*byte)(unsafe.Pointer(data)), 7)
	left, top := clientWallScreenPosition((*C.nox_draw_viewport_t)(view), wall[5], wall[6])
	*x, *y = C.int(left), C.int(top)
}

//export nox_xxx_drawWalls_473C10
func nox_xxx_drawWalls_473C10(view *C.nox_draw_viewport_t, data unsafe.Pointer) {
	wall := (*server.Wall)(data)
	if wall == nil || wall.Flags4&1 == 0 {
		return
	}
	x, y := clientWallScreenPosition(view, wall.X5, wall.Y6)
	grid := image.Pt(23*int(wall.X5), 23*int(wall.Y6))
	shape := memmap.Int32(0x587000, 149364+4*uintptr(wall.Field3))
	if shape == -1 {
		shape = int32(wall.Dir0)
	}
	spriteShape := shape
	if wall.Flags4&0x40 != 0 && (shape == 0 || shape == 1) {
		spriteShape += 11
	}
	left, right := int32(0), int32(clientDisplayStorage.width)
	if memmap.Uint32(0x587000, 80808) != 0 {
		var visible bool
		left, right, spriteShape, visible = clientWallClipSpan(shape, spriteShape, x, y, right)
		if !visible {
			return
		}
		if left >= right {
			wall.Field3 = 0
			wall.Flags4 &^= 3
			return
		}
	}

	front := wall.Flags4&2 != 0
	translucent := Get_nox_client_translucentFrontWalls_805844() != 0
	highFront := Get_nox_client_highResFrontWalls_80820() != 0
	variant := int32(wall.Flags4>>2) & 2
	var mode C.int
	if front {
		variant |= 1
		if memmap.Uint32(0x5D4594, 805848) != 0 && translucent &&
			(highFront || Get_nox_client_highResFloors_154952() == 0) {
			mode |= 8
		}
		if !highFront {
			mode |= 4
		}
	}
	tile, variation := int32(wall.Tile1), int32(wall.Field2)
	seeThrough := front && translucent && nox_xxx_wallFlags(tile)&4 == 0
	if seeThrough {
		mode |= 2
	} else {
		mode |= 1
	}

	gradient := memmap.Uint32(0x587000, 80816) != 0
	var first, last []uint32
	var firstColor [3]uint32
	var inset int32 = 50
	if gradient {
		from, to := clientWallLightEndpoints(grid, shape)
		copy(firstColor[:], GetClient().Sub469920(from))
		first = firstColor[:]
		last = GetClient().Sub469920(to)
		inset = 51
	} else {
		first = GetClient().Sub469920(grid.Add(image.Pt(11, 11)))
	}
	var offsetX, offsetY C.int
	nox_xxx_getWallDrawOffset_46A3F0(tile, spriteShape, variation, variant, &offsetX, &offsetY)
	x += int32(offsetX) - inset
	y -= int32(offsetY) + inset + 22
	sub_4345F0(1)
	nox_draw_setColorMultAndIntensityRGB_433CD0(C.uchar(first[0]), C.uchar(first[1]), C.uchar(first[2]))
	if gradient && !seeThrough {
		sprite := (*nox_video_bag_image_t)(nox_xxx_getWallSprite_46A3B0(tile, spriteShape, variation, variant))
		start := (*C.int)(unsafe.Pointer(&first[0]))
		end := (*C.int)(unsafe.Pointer(&last[0]))
		nox_xxx_edgeDraw_480EF0(sprite, C.int(x), C.int(y), start, end, clientDisplayStorage.height, C.int(left), C.int(right), 0, mode)
	} else if sub_47D380(C.int(left), C.int(right)) != 0 {
		if seeThrough {
			nox_client_drawEnableAlpha_434560(1)
			nox_client_drawSetAlpha_434580(0x80)
		}
		sub_47D400(int32(bool2int(Get_nox_client_highResFrontWalls_80820() == 0)), C.char(view.field_5))
		sprite := (*nox_video_bag_image_t)(nox_xxx_getWallSprite_46A3B0(tile, spriteShape, variation, variant))
		nox_client_drawImageAt_47D2C0(sprite, x, y)
		sub_47D400(0, 0)
		if seeThrough {
			nox_client_drawEnableAlpha_434560(0)
		}
	}
	sub_4345F0(0)
	wall.Field3 = 0
	wall.Flags4 &^= 3
	wall.Field12 = 1
}

func clientWallClipEndpoints(offset uintptr, x, y int32) (image.Point, image.Point) {
	p := memmap.PtrT[[4]int32](0x587000, offset)
	return image.Pt(int(x+p[0]), int(y+p[1])), image.Pt(int(x+p[2]), int(y+p[3]))
}

func clientWallClipSegment(from, to image.Point, left, right int32) (int32, int32, bool) {
	l, r := int(left), int(right)
	result := GetClient().Sub4C42A0(from, to, &l, &r)
	return int32(l), int32(r), result != 0
}

func clientWallClipSpan(shape, spriteShape, x, y, screenWidth int32) (left, right, sprite int32, visible bool) {
	offset := 16 * uintptr(shape)
	from, to := clientWallClipEndpoints(85440+offset, x, y)
	switch shape {
	case 7, 9:
		left, right, visible = clientWallClipSegment(from, to, 0, screenWidth)
		if !visible {
			left = int32(to.X)
		}
		secondFrom, secondTo := clientWallClipEndpoints(85504+offset, x, y)
		var secondVisible bool
		left, right, secondVisible = clientWallClipSegment(secondFrom, secondTo, left, right)
		if !secondVisible {
			if !visible {
				return left, right, spriteShape, false
			}
			right = min(right, int32(secondFrom.X))
		}
		return left, right, spriteShape, true
	case 8, 10:
		firstLeft, firstRight, firstOK := clientWallClipSegment(from, to, int32(from.X), int32(to.X))
		firstOK = firstOK && firstRight-firstLeft >= 3
		secondFrom, secondTo := clientWallClipEndpoints(85504+offset, x, y)
		startX, endX := int32(secondFrom.X), int32(secondTo.X)
		left, right, visible = clientWallClipSegment(secondFrom, secondTo, startX, endX)
		visible = visible && right-left >= 3
		switch {
		case firstOK && visible:
			left = min(left, firstLeft)
			right = max(right, firstRight)
			if left <= startX {
				left = 0
			}
			if right >= endX {
				right = screenWidth
			}
		case firstOK:
			left, right = firstLeft, firstRight
			if shape == 8 {
				spriteShape = 0
				if right == endX {
					right = screenWidth
				}
			} else {
				spriteShape = 1
				if left == startX {
					left = 0
				}
			}
		case visible:
			spriteShape = 13
			if shape != 8 {
				spriteShape = 14
			}
			if left == startX {
				left = 0
			}
			if right == endX {
				right = screenWidth
			}
		}
		return left, right, spriteShape, firstOK || visible
	default:
		left, right, visible = clientWallClipSegment(from, to, 0, screenWidth)
		return left, right, spriteShape, visible
	}
}

func clientWallLightEndpoints(grid image.Point, shape int32) (from, to image.Point) {
	switch shape {
	case 0, 3:
		from = grid.Add(image.Pt(0, 23))
		to = grid.Add(image.Pt(23, 0))
	case 1, 4:
		from = grid
		to = grid.Add(image.Pt(23, 23))
	case 7:
		from = grid
		to = grid.Add(image.Pt(23, 0))
	case 8:
		from = grid.Add(image.Pt(11, 11))
		to = grid.Sub(image.Pt(23, 23))
	case 10:
		from = grid.Add(image.Pt(0, 23))
		to = grid.Add(image.Pt(11, 12))
	default:
		from = grid.Add(image.Pt(0, 23))
		to = grid.Add(image.Pt(23, 23))
	}
	return
}

//export sub_474B40
func sub_474B40(drawable *C.nox_drawable) C.int {
	code := ClientPlayerNetCode()
	if local := nox_xxx_objGetTeamByNetCode_418C80(C.int(code)); local != nil {
		observed := nox_xxx_objGetTeamByNetCode_418C80(C.int(drawable.field_32))
		if observed != nil && (uint32(code) == uint32(drawable.field_32) || nox_xxx_servCompareTeams_419150(local, observed) != 0) {
			return 1
		}
	}
	local := (*C.nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if drawable == local || local != nil && bool(nox_client_drawable_testBuff_4356C0(local, 21)) {
		return 1
	}
	if drawable.flags28&4 != 0 {
		nox_common_playerInfoGetByID_417040(int32(drawable.field_32))
	}
	return 0
}

//export nox_xxx_drawAllMB_475810_draw_B
func nox_xxx_drawAllMB_475810_draw_B(view *C.nox_draw_viewport_t) C.int {
	if noxflags.HasEngine(noxflags.EngineNoFloorRendering) {
		return 0
	}
	position := C.float2{field_0: C.float(float64(view.field_6)), field_4: C.float(float64(view.field_7))}
	tile := nox_xxx_tileNFromPoint_411160(&position)
	return C.int(bool2int(tile != 255 && tile != -1))
}
