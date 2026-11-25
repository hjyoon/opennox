package legacy

/*
#include "GAME2_1.h"
extern uint32_t nox_xxx_minimap_587000_149232;
extern unsigned int nox_player_netCode_85319C;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func minimapZoom() uint32 {
	return uint32(C.nox_xxx_minimap_587000_149232)
}

//export nox_xxx_cliSetMinimapZoom_472520
func nox_xxx_cliSetMinimapZoom_472520(zoom C.int) C.int {
	SetMinimapZoom(uint32(zoom))
	return zoom
}

//export sub_472540
func sub_472540(drawable *C.nox_drawable) C.int {
	var position C.int2
	local := (*C.nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if drawable == local {
		nox_xxx_getSomeCoods_435670(&position)
	} else if drawable != nil {
		position.field_0, position.field_4 = C.int(drawable.pos.x), C.int(drawable.pos.y)
	}
	previous := memmap.PtrInt32(0x5D4594, 1096312)
	if index := nox_xxx_polygonGetIdxA_421790(&position, C.int(*previous)); index != 0 {
		*previous = int32(index)
	}
	if *previous != 0 {
		if polygon := mapPolygonByIndex(uint32(*previous)); polygon != nil {
			return C.int(polygon.AudioZone)
		}
	}
	return 1
}

//export nox_xxx_drawMinimap4Sprite_4725C0
func nox_xxx_drawMinimap4Sprite_4725C0(drawable *C.nox_drawable) {
	local := (*C.nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if drawable == nil || local == nil || bool(nox_client_drawable_testBuff_4356C0(local, 2)) {
		return
	}
	sub_437260()
	layer := sub_472540(drawable)
	*memmap.PtrInt32(0x5D4594, 1096316) = int32(layer)
	nox_xxx_cliDrawMinimap_472600(drawable, layer)
	sub_437290()
}

const (
	minimapTopLeft = iota
	minimapTopRight
	minimapBottomLeft
	minimapBottomRight
	minimapCenter
)

var minimapWallSegments = [...][][2]uint8{
	{{minimapBottomLeft, minimapTopRight}},
	{{minimapTopLeft, minimapBottomRight}},
	{{minimapBottomLeft, minimapTopRight}, {minimapTopLeft, minimapBottomRight}},
	{{minimapTopLeft, minimapCenter}, {minimapBottomLeft, minimapTopRight}},
	{{minimapTopLeft, minimapBottomRight}, {minimapCenter, minimapTopRight}},
	{{minimapBottomLeft, minimapTopRight}, {minimapCenter, minimapBottomRight}},
	{{minimapTopLeft, minimapBottomRight}, {minimapBottomLeft, minimapCenter}},
	{{minimapTopLeft, minimapCenter}, {minimapCenter, minimapTopRight}},
	{{minimapCenter, minimapBottomRight}, {minimapCenter, minimapTopRight}},
	{{minimapCenter, minimapBottomRight}, {minimapCenter, minimapBottomLeft}},
	{{minimapTopLeft, minimapCenter}, {minimapBottomLeft, minimapCenter}},
}

func drawMinimapLine(x0, y0, x1, y1 int32) C.int {
	nox_client_drawAddPoint_49F500(x0, y0)
	nox_client_drawAddPoint_49F500(x1, y1)
	return C.int(nox_client_drawLineFromPoints_49E4B0())
}

//export sub_4730D0
func sub_4730D0(at *C.int2, shape C.uchar, edge C.int) C.int {
	color := memmap.Int32(0x85B3FC, 956)
	if minimapZoom() > 2000 {
		nox_client_drawSetColor_434460(color)
		nox_client_drawPixel_49EFA0(int32(at.field_0), int32(at.field_4))
		return 0
	}
	result := C.int(shape)
	if int(shape) >= len(minimapWallSegments) {
		return result
	}
	x, y, step := int32(at.field_0), int32(at.field_4), int32(edge)
	half := step / 2
	corners := [...][2]int32{{x, y}, {x + step, y}, {x, y + step}, {x + step, y + step}, {x + half, y + half}}
	for _, segment := range minimapWallSegments[shape] {
		a, b := corners[segment[0]], corners[segment[1]]
		result = sub_473380(C.int(a[0]), C.int(a[1]), C.int(b[0]), C.int(b[1]), C.int(color))
	}
	return result
}

//export sub_473380
func sub_473380(x0, y0, x1, y1, color C.int) C.int {
	nox_client_drawSetColor_434460(int32(color))
	return drawMinimapLine(int32(x0), int32(y0), int32(x1), int32(y1))
}

func drawMinimapOutline(at *[2]int32, points [][2]int32) C.int {
	x, y := at[0], at[1]
	var result C.int
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		result = drawMinimapLine(x+a[0], y+a[1], x+b[0], y+b[1])
	}
	return result
}

//export sub_4733B0
func sub_4733B0(at *C.uint32_t) C.int {
	points := [...][2]int32{{-2, 4}, {-2, -4}, {2, -4}, {2, 0}, {-2, 0}}
	return drawMinimapOutline((*[2]int32)(unsafe.Pointer(at)), points[:])
}

//export sub_473420
func sub_473420(at *C.uint32_t) C.int {
	points := [...][2]int32{{-4, 6}, {-6, -6}, {-2, 0}, {0, -6}, {2, 0}, {6, -6}, {4, 6}, {-4, 6}}
	return drawMinimapOutline((*[2]int32)(unsafe.Pointer(at)), points[:])
}

//export nox_video_drawCircleRad3_4734F0
func nox_video_drawCircleRad3_4734F0(at *C.int) {
	point := (*[2]int32)(unsafe.Pointer(at))
	nox_video_drawCircle_4B0B90(point[0], point[1], 3)
}

//export nox_client_drawRectLines_473510
func nox_client_drawRectLines_473510(x, y, width, height C.int) C.int {
	left, top := int32(x), int32(y)
	right, bottom := int32(x+width-1), int32(y+height-1)
	drawMinimapLine(left, top, right, top)
	drawMinimapLine(right, top, right, bottom)
	return drawMinimapLine(right, bottom, left, bottom)
}

func minimapDotSize() int32 {
	switch zoom := minimapZoom(); {
	case zoom <= 1200:
		return 4
	case zoom < 1750:
		return 3
	default:
		return 2
	}
}

//export nox_xxx_minimapDrawPoint_473570
func nox_xxx_minimapDrawPoint_473570(x, y C.int) {
	nox_xxx_drawPointMB_499B70(int32(x), int32(y), minimapDotSize())
}

//export sub_4735C0
func sub_4735C0(x, y C.int) {
	nox_xxx_drawPointMB_499B70(int32(x), int32(y), minimapDotSize()+2)
}

//export nox_xxx_drawMinimapAndLines_4738E0
func nox_xxx_drawMinimapAndLines_4738E0() C.int {
	if Get_nox_client_gui_flag_1556112() == 1 {
		return 1
	}
	if memmap.Uint8(0x5D4594, 1096424)&1 != 0 {
		drawable := nox_xxx_netSpriteByCodeDynamic_45A6F0(int32(C.nox_player_netCode_85319C))
		nox_xxx_drawMinimap4Sprite_4725C0(drawable)
	}
	return nox_xxx_drawMessageLines_445530()
}

//export nox_xxx____setargv_11_473920
func nox_xxx____setargv_11_473920() {
	*memmap.PtrUint32(0x5D4594, 1096520) = 1
}
