package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func monsterRoamWaypointRuntime545B00() server.MonsterRoamWaypointRuntime545B00 {
	if runtime := monsterRoamWaypointRuntimeOverride545B00; runtime != nil {
		return *runtime
	}
	return server.MonsterRoamWaypointRuntime545B00{
		Random: func(min, max int) int {
			return GetServer().S().Rand.Logic.IntClamp(min, max)
		},
		UnitName: func(unit *server.Object) string {
			return GetServer().S().Types.ByInd(int(unit.TypeInd)).ID()
		},
		Frame: func() uint32 { return GetServer().S().Frame() },
		Debug: func(format string, frame uint32, name string, code int32) {
			if noxflags.HasEngine(noxflags.EngineShowAI) {
				ai.Log.Printf(format, int32(frame), name, code)
			}
		},
		Pop: func(unit *server.Object) int { return unit.MonsterPopAction() },
	}
}

//export sub_545B00
func sub_545B00(data unsafe.Pointer, waypoint *C.nox_waypoint_t) {
	(*server.MonsterUpdateData)(data).RoamRecordWaypoint545B00((*server.Waypoint)(unsafe.Pointer(waypoint)))
}

//export sub_545B60
func sub_545B60(data unsafe.Pointer, mask C.uchar) *C.nox_waypoint_t {
	return (*C.nox_waypoint_t)(unsafe.Pointer((*server.MonsterUpdateData)(data).RoamPreviousWaypoint545B60(byte(mask))))
}

//export sub_547EE0
func sub_547EE0(waypoint *C.nox_waypoint_t, mask C.uchar) C.int {
	return C.int(bool2int((*server.Waypoint)(unsafe.Pointer(waypoint)).RoamValid547EE0(byte(mask))))
}

//export sub_545C60
func sub_545C60(data unsafe.Pointer, current *C.nox_waypoint_t, mask C.uchar) *C.nox_waypoint_t {
	next := (*server.MonsterUpdateData)(data).RoamNextWaypoint545C60((*server.Waypoint)(unsafe.Pointer(current)), byte(mask), monsterRoamWaypointRuntime545B00())
	return (*C.nox_waypoint_t)(unsafe.Pointer(next))
}

//export nox_xxx_monsterRoamDeadEnd_545BB0
func nox_xxx_monsterRoamDeadEnd_545BB0(unit *C.nox_object_t, current *C.nox_waypoint_t) C.int {
	return C.int(bool2int((*server.Object)(unsafe.Pointer(unit)).RoamDeadEnd545BB0((*server.Waypoint)(unsafe.Pointer(current)), monsterRoamWaypointRuntime545B00())))
}
