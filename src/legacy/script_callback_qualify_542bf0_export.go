package legacy

/*
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

var scriptCallbackQualifyCall542BF0 = scriptCallbackQualifyNative542BF0

func scriptCallbackQualifyExportCall542BF0(a1, a2, a3 int32) {
	C.sub_542BF0(C.int32_t(a1), C.int32_t(a2), C.int32_t(a3))
}

func scriptCallbackQualifyNameExportCall5435C0(name string, a2, a3, a4 int32) (string, uintptr) {
	cname, freeName := alloc.CString(name)
	defer freeName()
	result := C.sub_5435C0(
		(*C.char)(unsafe.Pointer(cname)),
		C.int32_t(a2), C.int32_t(a3), C.int32_t(a4),
	)
	return alloc.GoString((*byte)(unsafe.Pointer(result))), uintptr(unsafe.Pointer(cname))
}

func scriptObjectQualifyNameExportCall543620(name string, value int32) (string, uintptr) {
	cname, freeName := alloc.CString(name)
	defer freeName()
	result := C.sub_543620((*C.char)(unsafe.Pointer(cname)), C.int32_t(value))
	return alloc.GoString((*byte)(unsafe.Pointer(result))), uintptr(unsafe.Pointer(cname))
}

func scriptCallbackObjectKindNative542BF0(object *server.Object) scriptCallbackObjectKind542BF0 {
	typ := GetServer().S().Types.ByInd(int(object.TypeInd))
	if typ == nil {
		panic("sub_542BF0: object type is nil")
	}
	switch typ.XferFunc() {
	case unsafe.Pointer(C.nox_xxx_unitTriggerXfer_4F4E50):
		return scriptCallbackObjectTrigger542BF0
	case unsafe.Pointer(C.nox_xxx_XFerMonster_528DB0):
		return scriptCallbackObjectMonster542BF0
	case unsafe.Pointer(C.nox_xxx_XFerHole_4F51D0):
		return scriptCallbackObjectHole542BF0
	case unsafe.Pointer(C.nox_xxx_XFerMonsterGen_4F7130):
		return scriptCallbackObjectGenerator542BF0
	default:
		return scriptCallbackObjectUnknown542BF0
	}
}

func scriptCallbackObjectFromNode542BF0(node *C.nox_map_object_list_node_5048A0) *server.Object {
	return asObjectS((*nox_object_t)(unsafe.Pointer(C.nox_map_object_list_node_object_5048A0(node))))
}

func scriptCallbackStoreObjectID542BF0(object *server.Object, value string) {
	ptr := alloc.Realloc(object.IDPtr, uintptr(len(value)+1))
	object.IDPtr = ptr
	dst := unsafe.Slice((*byte)(ptr), len(value)+1)
	copy(dst, value)
	dst[len(value)] = 0
}

func scriptCallbackQualifyNative542BF0(a1, a2, a3 int32) {
	srv := GetServer().S()
	scriptCallbackQualify542BF0(a1, a2, a3, scriptCallbackQualifyHooks542BF0[
		*C.nox_map_object_list_node_5048A0,
		*server.Waypoint,
	]{
		firstObject: func() *C.nox_map_object_list_node_5048A0 {
			return C.sub_5049D0()
		},
		nextObject: func(node *C.nox_map_object_list_node_5048A0) *C.nox_map_object_list_node_5048A0 {
			return C.sub_5049E0(node)
		},
		loadObjectFlags: func(node *C.nox_map_object_list_node_5048A0) uint32 {
			return uint32(scriptCallbackObjectFromNode542BF0(node).ObjFlags)
		},
		storeObjectFlags: func(node *C.nox_map_object_list_node_5048A0, flags uint32) {
			scriptCallbackObjectFromNode542BF0(node).ObjFlags = object.Flags(flags)
		},
		loadObjectID: func(node *C.nox_map_object_list_node_5048A0) (string, bool) {
			object := scriptCallbackObjectFromNode542BF0(node)
			if object.IDPtr == nil {
				return "", false
			}
			return object.ID(), true
		},
		storeObjectID: func(node *C.nox_map_object_list_node_5048A0, value string) {
			scriptCallbackStoreObjectID542BF0(scriptCallbackObjectFromNode542BF0(node), value)
		},
		loadCallbackName: func(node *C.nox_map_object_list_node_5048A0, event int) (string, bool) {
			return srv.NoxScriptVM.Nox_script_objCallbackName_508CB0(
				scriptCallbackObjectFromNode542BF0(node), event,
			)
		},
		objectKind: func(node *C.nox_map_object_list_node_5048A0) scriptCallbackObjectKind542BF0 {
			return scriptCallbackObjectKindNative542BF0(scriptCallbackObjectFromNode542BF0(node))
		},
		storeCallbackName: func(node *C.nox_map_object_list_node_5048A0, event int, name string) {
			scriptCallbackSetCall509120(scriptCallbackObjectFromNode542BF0(node), event, name)
		},
		firstWaypoint: func() *server.Waypoint {
			return srv.WPs.First()
		},
		nextWaypoint: func(waypoint *server.Waypoint) *server.Waypoint {
			return waypoint.Next()
		},
		loadWaypointFlags: func(waypoint *server.Waypoint) uint32 {
			return waypoint.Flags
		},
		storeWaypointFlags: func(waypoint *server.Waypoint, flags uint32) {
			waypoint.Flags = flags
		},
		loadWaypointName: func(waypoint *server.Waypoint) string {
			return waypoint.Name()
		},
		storeWaypointName: func(waypoint *server.Waypoint, name string) {
			waypoint.SetName(name)
		},
	})
}

//export sub_542BF0
func sub_542BF0(a1, a2, a3 C.int32_t) *C.char {
	scriptCallbackQualifyCall542BF0(int32(a1), int32(a2), int32(a3))
	return nil
}
