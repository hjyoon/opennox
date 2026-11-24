package legacy

/*
#include <float.h>
#include "GAME4_2.h"

enum { nox_mapgen_float_eval_method = FLT_EVAL_METHOD };
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_5244D0
func sub_5244D0(token C.int) C.int {
	def := mapgenResolveDecorDefinition(C.uint32_t(token))
	if def == nil || def.wall_floor_count == 0 {
		return 0
	}
	index := nox_xxx_mapGenRandFunc_526AC0(0, C.int(def.wall_floor_count-1))
	ref := def.wall_floors_token
	for ; ref != 0 && index > 0; index-- {
		surface := mapgenResolveWallFloor(ref)
		if surface == nil {
			return 0
		}
		ref = surface.next_token
	}
	return C.int(ref)
}

//export sub_524660
func sub_524660(a, b C.float) C.int {
	difference := float64(float32(a - b))
	if C.nox_mapgen_float_eval_method > 0 {
		difference = float64(a) - float64(b)
	}
	return C.int(bool2int(math.Abs(difference) < memmap.Float64(0x581450, 10432)))
}

func mapgenHallFloorLength(origin *C.float, length C.int, start, end *mapgenRoomLinkRecord, fromLow bool) C.int {
	if start == nil || start.kind == 1 {
		length--
		if fromLow {
			*origin += mapgenCellSize
		}
	}
	if end == nil {
		return length
	}
	if end.kind == 1 {
		length--
		if !fromLow {
			*origin += mapgenCellSize
		}
	} else {
		length++
		if !fromLow {
			*origin -= mapgenCellSize
		}
	}
	return length
}

func mapgenRoomFloorInterior(room *mapgenRoomGeometry) (C.float2, [2]C.int) {
	links := (*mapgenRoomLinkRecord)(unsafe.Pointer(room))
	neighbor := func(direction int) *mapgenRoomLinkRecord {
		return mapgenResolveRoomLinkRecord(links.links[direction][0])
	}
	x := C.float(math.Float32frombits(uint32(room.position[0])))
	y := C.float(math.Float32frombits(uint32(room.position[1])))
	width := C.int(room.cellDimensions[0])
	height := C.int(room.cellDimensions[1])
	switch room.kind {
	case 1:
		x += mapgenCellSize
		y += mapgenCellSize
		width, height = width-2, height-2
	case 2:
		x += mapgenCellSize
		width -= 2
		height = mapgenHallFloorLength(&y, height, neighbor(0), neighbor(1), true)
	case 3:
		x += mapgenCellSize
		width -= 2
		height = mapgenHallFloorLength(&y, height, neighbor(1), neighbor(0), false)
	case 4:
		y += mapgenCellSize
		height -= 2
		width = mapgenHallFloorLength(&x, width, neighbor(2), neighbor(3), false)
	case 5:
		y += mapgenCellSize
		height -= 2
		width = mapgenHallFloorLength(&x, width, neighbor(3), neighbor(2), true)
	default:
		return C.float2{}, [2]C.int{}
	}
	return C.float2{field_0: x, field_4: y}, [2]C.int{width, height}
}

//export nox_mapgenGenerateWallFloorNative_524680
func nox_mapgenGenerateWallFloorNative_524680(theme, raw *C.uint8_t, surface *C.nox_mapgen_wall_floor_legacy) C.int {
	if theme == nil || raw == nil || surface == nil {
		return 0
	}
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	nox_xxx_tileGetDefByName_51D4D0(&surface.floor_name[0])
	sub_5245A0(0, (*C.float)(unsafe.Pointer(&room.position[0])),
		C.int(room.cellDimensions[0]), C.int(room.cellDimensions[1]))

	origin, dimensions := mapgenRoomFloorInterior(room)
	for token := surface.fills_token; token != 0 && dimensions[0] > 0 && dimensions[1] > 0; {
		fill := mapgenResolveDecorFill(token)
		if fill == nil {
			break
		}
		start := (*C.float)(unsafe.Pointer(&origin))
		switch fill._type {
		case 1:
			C.sub_5249C0(0, C.int(token), start, &dimensions[0])
		case 2:
			C.sub_524B50(0, C.int(token), start, &dimensions[0])
		default:
			C.sub_524950(0, C.int(token), start, &dimensions[0])
		}
		dimensions[0] -= 2
		dimensions[1] -= 2
		origin.field_0 += mapgenCellSize
		origin.field_4 += mapgenCellSize
		token = fill.next_token
	}
	return dimensions[0]
}

//export nox_xxx_gen_524680
func nox_xxx_gen_524680(theme, room, surface C.int) C.int {
	return nox_mapgenGenerateWallFloorNative_524680(mapgenThemeByToken(theme), mapgenThemeByToken(room),
		mapgenResolveWallFloor(C.uint32_t(surface)))
}
