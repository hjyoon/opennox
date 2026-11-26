package legacy

/*
#include <stdlib.h>
#include "GAME1_1.h"
#include "GAME2_1.h"

extern uint32_t nox_xxx_polygonNextIdx_587000_60352;
*/
import "C"

import (
	"math"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	polygonDraftCount    = 588072
	polygonDraftVertices = 534820
	polygonDraftCapacity = (mapAngleTableBase - polygonDraftVertices) / 4
)

//export sub_420E80
func sub_420E80(x, y, limit C.float) *C.uint32_t {
	var best *Nox_polygon_angle_data
	for index := uint32(1); index < uint32(legacyWords.nox_xxx_polygonNextAngle_587000_60356); index++ {
		angle := mapAngleByIndex(index)
		if angle.Active == 0 {
			continue
		}
		dx := float32(angle.X - float32(x))
		dy := float64(float32(angle.Y - float32(y)))
		distance := dy*dy + float64(float32(dx*dx))
		if distance < float64(limit) {
			limit = C.float(distance)
			best = angle
		}
	}
	return (*C.uint32_t)(unsafe.Pointer(best))
}

//export sub_420DA0
func sub_420DA0(x, y C.float) *C.uint {
	count := memmap.PtrUint16(0x5D4594, polygonDraftCount)
	if *count >= polygonDraftCapacity {
		return nil
	}
	if noxflags.HasGame(noxflags.GameFlag22) {
		if existing := sub_420E80(x, y, 900); existing != nil {
			if sub_421B40(existing) == 0 {
				*memmap.PtrUint32(0x5D4594, polygonDraftVertices+4*uintptr(*count)) = uint32(*existing)
				*count++
			}
			return nil
		}
		point := [2]int32{server.X87TruncateSignedDWORD419A70(float32(x)), server.X87TruncateSignedDWORD419A70(float32(y))}
		if findMapPolygonAt(&point, 0, false) != nil {
			return nil
		}
	}
	index := sub_420D10()
	if index >= mapPolygonAngleCapacity {
		legacyWords.nox_xxx_polygonNextAngle_587000_60356 = C.uint32_t(index)
		return nil
	}
	bitsX := C.int(int32(math.Float32bits(float32(x))))
	bitsY := C.int(int32(math.Float32bits(float32(y))))
	angle := nox_xxx_polygonSetAngle_420D40(bitsX, bitsY, C.uint(index), 0)
	*memmap.PtrUint32(0x5D4594, polygonDraftVertices+4*uintptr(*count)) = uint32(*angle)
	*count++
	return angle
}

//export sub_421160
func sub_421160(raw unsafe.Pointer) C.int {
	polygon := (*mapPolygonRecord)(raw)
	name := memmap.PtrT[[100]byte](0x587000, 60364)
	alloc.StrCopy(polygon.Name[:], alloc.GoStringS(name[:]))
	polygon.Color = *memmap.PtrT[[3]byte](0x587000, 60464)
	polygon.AudioZone = memmap.Uint8(0x587000, 60490)
	polygon.Flags, polygon.Visited = 0, 0
	return 0
}

//export sub_4211D0
func sub_4211D0(raw unsafe.Pointer) C.int {
	polygon := (*mapPolygonRecord)(raw)
	name := memmap.PtrT[[100]byte](0x587000, 60364)
	alloc.StrCopy(name[:], alloc.GoStringS(polygon.Name[:]))
	*memmap.PtrT[[3]byte](0x587000, 60464) = polygon.Color
	*memmap.PtrUint8(0x587000, 60490) = polygon.AudioZone
	polygon.Flags, polygon.Visited = 0, 0
	return 0
}

//export sub_421230
func sub_421230() *C.uchar {
	index := uint32(sub_421130())
	if index >= mapPolygonCapacity {
		C.nox_xxx_polygonNextIdx_587000_60352 = C.uint32_t(index)
		return nil
	}
	polygon := mapPolygonByIndex(index)
	raw := unsafe.Pointer(polygon)
	nox_xxx_polygonSetDataNative(raw, nil)
	nox_xxx_polygonSetVertexIndicesNative(raw, nil)
	if noxflags.HasGame(noxflags.GameFlag22) {
		data := C.calloc(1, 256)
		if data == nil {
			return nil
		}
		nox_xxx_polygonSetDataNative(raw, data)
	}
	polygon.Index = index
	digits := strconv.AppendUint(polygon.Name[:0], uint64(index), 10)
	polygon.Name[len(digits)] = 0
	sub_421160(raw)
	polygon.Active = 1
	return (*C.uchar)(raw)
}

//export sub_4214D0
func sub_4214D0() {
	count := memmap.PtrUint16(0x5D4594, polygonDraftCount)
	if *count < 3 || *count > polygonDraftCapacity {
		return
	}
	vertices := memmap.PtrT[[polygonDraftCapacity]uint32](0x5D4594, polygonDraftVertices)
	bounds, err := mapPolygonBounds(vertices[:*count])
	if err != nil {
		return
	}
	data := C.calloc(C.size_t(*count), 4)
	if data == nil {
		return
	}
	raw := unsafe.Pointer(sub_421230())
	if raw == nil {
		C.free(data)
		return
	}
	nox_xxx_polygonSetVertexIndicesNative(raw, (*C.uint32_t)(data))
	copy(unsafe.Slice((*uint32)(data), *count), vertices[:*count])
	polygon := (*mapPolygonRecord)(raw)
	polygon.VertexCount, polygon.Bounds = *count, bounds
	*count = 0
}

//export nox_xxx_polygonDrawColor_421B80
func nox_xxx_polygonDrawColor_421B80() {
	drawable := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if drawable == nil {
		return
	}
	var now, previous C.int2
	nox_xxx_getSomeCoods_435670(&now)
	sub_435690((*C.uint32_t)(unsafe.Pointer(&previous)))
	player := asPlayerS(nox_common_playerInfoGetByID_417040(int32(drawable.field_32)))
	if player == nil {
		return
	}
	if C.nox_xxx_polygonNextIdx_587000_60352 > 1 {
		id := player.LocalPolygonID()
		if id != 0xdeadface && now == previous {
			return
		}
		point := [2]int32{int32(now.field_0), int32(now.field_4)}
		if polygon := findMapPolygonAt(&point, id, false); polygon != nil {
			if id != polygon.Index {
				player.SetAudioZone(polygon.AudioZone)
				player.SetLocalPolygonID(polygon.Index)
				sub_434990(int32(polygon.Color[0]), int32(polygon.Color[1]), int32(polygon.Color[2]))
			}
			return
		}
	}
	player.SetAudioZone(1)
	player.SetLocalPolygonID(0)
	sub_4349C0((*C.uint)(unsafe.Pointer(ambientMapColor())))
}
