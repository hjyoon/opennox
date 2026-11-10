package legacy

/*
#include <stdint.h>
#include "GAME4.h"
#include "GAME3_2.h"
#include "GAME1_2.h"
#include "game4_map_archive.h"

extern uint32_t dword_5d4594_1599480;
extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_3835396;
*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// The PE x87 helper truncates to a signed qword and consumes its low dword.
// Invalid FISTP produces INT64_MIN, whose low dword is zero. Keep the upper
// bound exclusive: float64(MaxInt64) itself rounds up to 2^63.
func game4MapTruncQwordLow(value float64) int32 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < -0x1p63 || value >= 0x1p63 {
		return 0
	}
	return int32(uint32(int64(value)))
}

//export nox_mapgenClearPendingScriptIDs_503B30
func nox_mapgenClearPendingScriptIDs_503B30(first *C.nox_object_t) {
	for obj := asObjectS(first); obj != nil; obj = obj.ObjNext {
		obj.ScriptIDVal = 0
	}
}

//export game4MapBuildPlaceBounds
func game4MapBuildPlaceBounds(atRaw, fixedRaw, recordRaw, cornersRaw, boundsRaw unsafe.Pointer) {
	at := (*C.float2)(atRaw)
	fixedAt := (*C.float2)(fixedRaw)
	record := unsafe.Slice((*byte)(recordRaw), game4MapRecordBytes)
	width := math.Float32frombits(binary.LittleEndian.Uint32(record[64:]))
	height := math.Float32frombits(binary.LittleEndian.Uint32(record[68:]))
	corners := unsafe.Slice((*int32)(cornersRaw), 8)
	var corner, fixed C.float2

	corners[2] = game4MapTruncQwordLow(float64(fixedAt.field_0))
	corners[3] = game4MapTruncQwordLow(float64(fixedAt.field_4))
	corner.field_0 = C.float(width + float32(at.field_0))
	corner.field_4 = C.float(height + float32(at.field_4))
	C.nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed)
	corners[4] = game4MapTruncQwordLow(float64(fixed.field_0))
	corners[5] = game4MapTruncQwordLow(float64(fixed.field_4))
	corner.field_4 = at.field_4
	C.nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed)
	corners[0] = game4MapTruncQwordLow(float64(fixed.field_0))
	corners[1] = game4MapTruncQwordLow(float64(fixed.field_4))
	corner.field_0 = at.field_0
	corner.field_4 = C.float(height + float32(at.field_4))
	C.nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed)
	corners[6] = game4MapTruncQwordLow(float64(fixed.field_0))
	corners[7] = game4MapTruncQwordLow(float64(fixed.field_4))
	// These existing external routines retain their actual read/write order,
	// including a caller that aliases the rectangle with the corner array.
	C.sub_4D3C80((*C.uint32_t)(cornersRaw))
	C.sub_428170(cornersRaw, (*C.int4)(boundsRaw))
}

//export nox_mapgenPlaceOffset_503B30
func nox_mapgenPlaceOffset_503B30(coordinate C.float, origin C.int32_t) C.int32_t {
	return C.int32_t(game4MapTruncQwordLow(float64(coordinate) - float64(origin)))
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

//export game4MapPlaceGeometry
func game4MapPlaceGeometry(x, y, fixedX, fixedY C.float, record unsafe.Pointer, originX, originY C.int32_t, bounds, delta *C.int32_t) C.int {
	if record == nil || bounds == nil || delta == nil {
		return 0
	}
	at := C.float2{field_0: x, field_4: y}
	fixed := C.float2{field_0: fixedX, field_4: fixedY}
	var corners [8]int32
	var rectangle C.int4
	game4MapBuildPlaceBounds(unsafe.Pointer(&at), unsafe.Pointer(&fixed), record, unsafe.Pointer(&corners[0]), unsafe.Pointer(&rectangle))
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
	if C.dword_5d4594_1599480 != C.dword_5d4594_3835396 || int32(C.dword_5d4594_1599480) == -1 || C.dword_5d4594_1599476 == 1 {
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
