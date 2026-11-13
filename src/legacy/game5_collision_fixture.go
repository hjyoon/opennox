package legacy

/*
#include "GAME5.h"
extern nox_object_t* dword_5d4594_2491592;
static uintptr_t game5_hit_objects[8][2];
static float2 game5_hit_normals[8];
static int game5_hit_calls;
static int game5_hit_mutate;
static void game5_hit_callback(nox_object_t* first,nox_object_t* second,float2* normal) {
 int i=game5_hit_calls++;
 if(i<8){game5_hit_objects[i][0]=(uintptr_t)first;game5_hit_objects[i][1]=(uintptr_t)second;game5_hit_normals[i]=*normal;}
 if(game5_hit_mutate && i==0){normal->field_0=7;normal->field_4=-9;}
}
static void* game5_hit_callback_address(void){return (void*)game5_hit_callback;}
static void game5_hit_reset(int mutate){game5_hit_calls=0;game5_hit_mutate=mutate;}
static int game5_hit_count(void){return game5_hit_calls;}
static uintptr_t game5_hit_object(int i,int j){return game5_hit_objects[i][j];}
static float2 game5_hit_normal(int i){return game5_hit_normals[i];}
static void game5_hit_add(nox_object_t* first,nox_object_t* second,uintptr_t token,float2* normal){
 nox_xxx_collSysAddCollision_548630(first,second?(uintptr_t)second:token,normal);
}
static void game5_door_enqueue(void* data){sub_548830((uintptr_t)data);}
*/
import "C"
import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func game5HitInit(mutate bool) {
	C.nox_xxx_allocHitArray_5486D0()
	C.game5_hit_reset(C.int(bool2int(mutate)))
}
func game5HitCallback() unsafe.Pointer { return C.game5_hit_callback_address() }
func game5HitAdd(first, second *server.Object, token uintptr, normal *types.Pointf) {
	C.game5_hit_add(asObjectC(first), asObjectC(second), C.uintptr_t(token), (*C.float2)(unsafe.Pointer(normal)))
}
func game5HitFlush()     { C.nox_xxx_collide_548740() }
func game5HitCount() int { return int(C.game5_hit_count()) }
func game5HitRecord(i int) (uintptr, uintptr, types.Pointf) {
	p := C.game5_hit_normal(C.int(i))
	return uintptr(C.game5_hit_object(C.int(i), 0)), uintptr(C.game5_hit_object(C.int(i), 1)), types.Ptf(float32(p.field_0), float32(p.field_4))
}
func game5DoorQueue(data *server.DoorUpdateData)     { C.game5_door_enqueue(unsafe.Pointer(data)) }
func game5DoorDelta(obj *server.Object, delta int16) { C.sub_548860(asObjectC(obj), C.short(delta)) }
func game5DoorFlush()                                { C.sub_548B60() }
func game5IntersectionC(segment, rect *[4]float32, out *types.Pointf, limit int32) int {
	return int(C.sub_5516A0((*C.float4)(unsafe.Pointer(segment)), (*C.float4)(unsafe.Pointer(rect)), (*C.float2)(unsafe.Pointer(out)), C.int(limit), 0))
}
func game5AxisC(vertical bool, segment *[4]float32, first, second, fixed float32, out *types.Pointf) bool {
	if vertical {
		return C.sub_551870((*C.float4)(unsafe.Pointer(segment)), C.float(fixed), C.float(first), C.float(second), (*C.float2)(unsafe.Pointer(out)), 0) != 0
	}
	return C.sub_551780((*C.float4)(unsafe.Pointer(segment)), C.float(first), C.float(second), C.float(fixed), (*C.float2)(unsafe.Pointer(out)), 0) != 0
}
func game5CollisionPredicateC(first, second *server.Object, projectile bool) bool {
	if projectile {
		return C.sub_54E730(asObjectC(first), asObjectC(second)) != 0
	}
	return C.sub_548360(asObjectC(first), asObjectC(second)) != 0
}
func game5CursorContainsForTest(obj *server.Object, cursor *types.Pointf) bool {
	return game5CursorContains(obj, cursor)
}

func game5SecretWallTouchC(grid [2]int32, obj *server.Object) {
	C.sub_548100((*C.int2)(unsafe.Pointer(&grid)), asObjectC(obj))
}
func game5CursorFindC(obj *server.Object) *server.Object {
	return asObjectS(C.nox_xxx_findObjectAtCursor_54AF40(asObjectC(obj)))
}
func game5CursorScanC(obj *server.Object, cursor *types.Pointf) {
	C.nox_xxx_playerCursorScanFn_54AFB0(asObjectC(obj), (*C.float2)(unsafe.Pointer(cursor)))
}

func game5CursorClearFixture() {
	game5Cursor.source = nil
	game5Cursor.result = nil
	C.dword_5d4594_2491592 = nil
}
