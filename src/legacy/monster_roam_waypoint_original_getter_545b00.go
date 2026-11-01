package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
typedef struct {
 uint32_t prefix[75];
 nox_waypoint_t* waypoints[16];
 uint32_t index;
} waypoint_history_layout_545b00;
_Static_assert(_Generic(&sub_545B00,void (*)(void*,nox_waypoint_t*):1,default:0),"native Record ABI");
_Static_assert(_Generic(&sub_545B60,nox_waypoint_t* (*)(void*,unsigned char):1,default:0),"native Previous ABI");
_Static_assert(_Generic(&sub_547EE0,int (*)(nox_waypoint_t*,unsigned char):1,default:0),"native Valid ABI");
_Static_assert(_Generic(&sub_545C60,nox_waypoint_t* (*)(void*,nox_waypoint_t*,unsigned char):1,default:0),"native Next ABI");
_Static_assert(_Generic(&nox_xxx_monsterRoamDeadEnd_545BB0,int (*)(nox_object_t*,nox_waypoint_t*):1,default:0),"native DeadEnd ABI");
static uintptr_t waypoint_original_public_545b00(int root,void* data,nox_waypoint_t* point,nox_object_t* unit,unsigned char mask) {
 switch(root) {
 case 0:sub_545B00(data,point);return 0;
 case 1:return (uintptr_t)sub_545B60(data,mask);
 case 2:return sub_547EE0(point,mask);
 case 3:return (uintptr_t)sub_545C60(data,point,mask);
 default:return nox_xxx_monsterRoamDeadEnd_545BB0(unit,point);
 }
}
static uintptr_t waypoint_original_layout_545b00(int field) {
 switch(field) {
 case 0:return offsetof(waypoint_history_layout_545b00,waypoints);
 case 1:return offsetof(waypoint_history_layout_545b00,index);
 case 2:return sizeof(nox_waypoint_t);
 case 3:return offsetof(nox_waypoint_t,flags);
 case 4:return offsetof(nox_waypoint_t,flags_2);
 default:return offsetof(nox_waypoint_t,points_cnt);
 }
}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var monsterRoamWaypointRuntimeOverride545B00 *server.MonsterRoamWaypointRuntime545B00

func monsterRoamWaypointOriginalC545B00(root string, data *server.MonsterUpdateData, point *server.Waypoint, unit *server.Object, mask byte) uintptr {
	entries := map[string]int{"Record": 0, "Previous": 1, "Valid": 2, "Next": 3, "DeadEnd": 4}
	entry, ok := entries[root]
	if !ok {
		panic("unknown waypoint original root")
	}
	return uintptr(C.waypoint_original_public_545b00(C.int(entry), unsafe.Pointer(data), (*C.nox_waypoint_t)(unsafe.Pointer(point)), (*C.nox_object_t)(unsafe.Pointer(unit)), C.uchar(mask)))
}
func monsterRoamWaypointOriginalLayout545B00() [6]uintptr {
	var out [6]uintptr
	for i := range out {
		out[i] = uintptr(C.waypoint_original_layout_545b00(C.int(i)))
	}
	return out
}
