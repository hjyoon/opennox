package legacy

/*
#include "audio_event_zone_501c00.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const (
	audioEventZoneUninitializedPolygon501C00 = uint32(0xDEADFACE)
	audioEventZonePolygonCount501C00         = uint32(255)
	audioEventZonePolygonRecordSize501C00    = uintptr(140)
	audioEventZonePolygonRecordsOffset501C00 = uintptr(552228)
)

func audioEventZoneRuntime501C00() server.AudioEventZoneRuntime501C00 {
	return server.AudioEventZoneRuntime501C00{
		PolygonByID: func(id uint32) unsafe.Pointer {
			if id == audioEventZoneUninitializedPolygon501C00 || id >= audioEventZonePolygonCount501C00 {
				return nil
			}
			return memmap.PtrOff(
				0x5D4594,
				audioEventZonePolygonRecordsOffset501C00+audioEventZonePolygonRecordSize501C00*uintptr(id),
			)
		},
		PolygonAtPoint: func(point [2]int32, previous uint32) unsafe.Pointer {
			return unsafe.Pointer(Nox_xxx_polygonIsPlayerInPolygon_4217B0(
				unsafe.Pointer(&point[0]),
				int(int32(previous)),
			))
		},
		PolygonZone: func(polygon unsafe.Pointer) uint8 {
			return uint8((*Nox_player_polygon_check_data)(polygon).Field_0[32] >> 16)
		},
	}
}

var audioEventZoneCall501C00 = func(position *types.Pointf, object *server.Object) uint8 {
	return GetServer().S().AudioEventZone501C00(position, object, audioEventZoneRuntime501C00())
}

func audioEventZoneExportCall501C00(position *types.Pointf, object *server.Object) uint8 {
	return uint8(C.sub_501C00(
		(*C.float)(unsafe.Pointer(position)),
		(*C.nox_object_t)(asObjectC(object)),
	))
}

//export sub_501C00
func sub_501C00(position *C.float, object *C.nox_object_t) C.uint8_t {
	return C.uint8_t(audioEventZoneCall501C00(
		(*types.Pointf)(unsafe.Pointer(position)),
		asObjectS((*nox_object_t)(object)),
	))
}

func Sub_501C00(position types.Pointf, object *server.Object) int {
	return int(audioEventZoneCall501C00(&position, object))
}
