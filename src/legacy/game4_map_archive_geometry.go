package legacy

/*
#include "cgo_const.h"
#include <stdint.h>
#include "GAME4.h"
#include "GAME3_2.h"
#include "GAME1_2.h"
#include "game4_map_archive.h"
typedef const float2 nox_const_float2;

extern uint32_t dword_5d4594_3835396;
*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_mapgenClearPendingScriptIDs_503B30
func nox_mapgenClearPendingScriptIDs_503B30(first *C.nox_object_t) {
	for obj := asObjectS(first); obj != nil; obj = obj.ObjNext {
		obj.ScriptIDVal = 0
	}
}

type mapgenPlaceCorner struct {
	x, y uint32
}

//export sub_4D3C80
func sub_4D3C80(words *C.uint32_t) *C.uint32_t {
	corners := (*[4]mapgenPlaceCorner)(unsafe.Pointer(words))
	a, b, c, d := corners[0], corners[1], corners[2], corners[3]

	// The C entry mixes signed temporaries and unsigned array comparisons.
	top, left, right, bottom := a, b, c, d
	if int32(b.y) < int32(top.y) {
		top = b
	}
	if c.y < top.y {
		top = c
	}
	if d.y < top.y {
		top = d
	}

	if a.x < left.x {
		left = a
	}
	if c.x < left.x {
		left = c
	}
	if int32(d.x) < int32(left.x) {
		left = d
	}

	if a.x > right.x {
		right = a
	}
	if b.x > right.x {
		right = b
	}
	if int32(d.x) > int32(right.x) {
		right = d
	}

	if a.y > bottom.y {
		bottom = a
	}
	if int32(b.y) > int32(bottom.y) {
		bottom = b
	}
	if c.y > bottom.y {
		bottom = c
	}
	*corners = [4]mapgenPlaceCorner{top, left, right, bottom}
	return words
}

//export nox_mapgenBuildPlaceBounds_503B30
func nox_mapgenBuildPlaceBounds_503B30(at, fixedAt *C.nox_const_float2, recordRaw *C.nox_cgo_const_void, cornerBuffer *C.int32_t, bounds *C.int4) {
	record := unsafe.Slice((*byte)(unsafe.Pointer(recordRaw)), game4MapRecordBytes)
	width := math.Float32frombits(binary.LittleEndian.Uint32(record[64:]))
	height := math.Float32frombits(binary.LittleEndian.Uint32(record[68:]))
	corners := unsafe.Slice((*int32)(unsafe.Pointer(cornerBuffer)), 8)
	var corner, fixed C.float2

	corners[2] = x87TruncQwordLow(float64(fixedAt.field_0))
	corners[3] = x87TruncQwordLow(float64(fixedAt.field_4))
	corner.field_0 = C.float(width + float32(at.field_0))
	corner.field_4 = C.float(height + float32(at.field_4))
	C.nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed)
	corners[4] = x87TruncQwordLow(float64(fixed.field_0))
	corners[5] = x87TruncQwordLow(float64(fixed.field_4))
	corner.field_4 = at.field_4
	C.nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed)
	corners[0] = x87TruncQwordLow(float64(fixed.field_0))
	corners[1] = x87TruncQwordLow(float64(fixed.field_4))
	corner.field_0 = at.field_0
	corner.field_4 = C.float(height + float32(at.field_4))
	C.nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed)
	corners[6] = x87TruncQwordLow(float64(fixed.field_0))
	corners[7] = x87TruncQwordLow(float64(fixed.field_4))
	// Order the corners before reading them into the rectangle. They may alias.
	sub_4D3C80((*C.uint32_t)(unsafe.Pointer(cornerBuffer)))
	C.sub_428170(unsafe.Pointer(cornerBuffer), bounds)
}

//export nox_mapgenPlaceOffset_503B30
func nox_mapgenPlaceOffset_503B30(coordinate C.float, origin C.int32_t) C.int32_t {
	return C.int32_t(x87TruncQwordLow(float64(coordinate) - float64(origin)))
}

//export nox_mapgenWallSpan_503B30
func nox_mapgenWallSpan_503B30(count C.uint32_t) C.float {
	return C.float(float32(int32(uint32(count) * 23)))
}

//export nox_mapgenFixCoordsNative_503B30
func nox_mapgenFixCoordsNative_503B30(x, y C.float, fixedX, fixedY *C.float) C.int {
	at := C.float2{field_0: x, field_4: y}
	var fixed C.float2
	if C.nox_xxx_mapGenFixCoords_4D3D90(&at, &fixed) == 0 {
		return 0
	}
	*fixedX = fixed.field_0
	*fixedY = fixed.field_4
	return 1
}

//export nox_mapgenPlaceGeometryNative_503B30
func nox_mapgenPlaceGeometryNative_503B30(x, y, fixedX, fixedY C.float, record *C.nox_cgo_const_void, originX, originY C.int32_t, bounds, delta *C.int32_t) C.int {
	if record == nil || bounds == nil || delta == nil {
		return 0
	}
	at := C.float2{field_0: x, field_4: y}
	fixed := C.float2{field_0: fixedX, field_4: fixedY}
	var corners [8]C.int32_t
	var rectangle C.int4
	nox_mapgenBuildPlaceBounds_503B30(&at, &fixed, record, &corners[0], &rectangle)
	// Match memcpy before the delta stores, including overlapping outputs.
	copy(unsafe.Slice((*byte)(unsafe.Pointer(bounds)), 16), unsafe.Slice((*byte)(unsafe.Pointer(&rectangle)), 16))
	out := unsafe.Slice(delta, 2)
	out[0] = nox_mapgenPlaceOffset_503B30(fixedX, originX)
	out[1] = nox_mapgenPlaceOffset_503B30(fixedY, originY)
	return 1
}

//export sub_503B30
func sub_503B30(at *C.float2) C.int {
	if at == nil {
		return 0
	}
	return C.int(bool2int(mapgenPlaceEntry503B30(float32(at.field_0), float32(at.field_4))))
}

//export sub_503EC0
func sub_503EC0(object *C.nox_object_t, out *C.float2) C.int {
	if game4Mapgen.loaded_index != C.dword_5d4594_3835396 || int32(game4Mapgen.loaded_index) == -1 || game4Mapgen.tiles_placed == 1 {
		return 0
	}
	origin := C.float2{
		field_0: C.float(*memmap.PtrInt32(0x5D4594, 1599508)),
		field_4: C.float(*memmap.PtrInt32(0x5D4594, 1599512)),
	}
	var fixedOrigin, fixedPosition C.float2
	C.sub_4D3E30(&origin, &fixedOrigin)
	// The transform clamps this live pair in place. Native offsetof comes
	// from the actual object type, never the original PE32 +56 literal.
	C.sub_4D3E30((*C.float2)(unsafe.Pointer(&object.x)), &fixedPosition)
	out.field_0 = fixedPosition.field_0 - fixedOrigin.field_0
	out.field_4 = fixedPosition.field_4 - fixedOrigin.field_4
	return 1
}
