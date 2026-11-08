package legacy

/*
#include "GAME1.h"
#include <stddef.h>
#include <stdint.h>
#include <string.h>
typedef int (*tile_list_fn_350)(nox_tile_list_node_t*,int*,int);
typedef int (*tile_list_fn_3a0)(int*,int);
typedef int (*tile_list_fn_490)(int,int);
_Static_assert(sizeof(int)==4,"tile list DWORD ABI");
_Static_assert(_Generic(&sub_411350,tile_list_fn_350:1,default:0),"350 ABI");
_Static_assert(_Generic(&sub_4113A0,tile_list_fn_3a0:1,default:0),"3A0 ABI");
_Static_assert(_Generic(&sub_411490,tile_list_fn_490:1,default:0),"490 ABI");
static int tile_list_fixture_list_411350(void* head,void* xy,int32_t value) {
    tile_list_fn_350 fn=&sub_411350;return fn((nox_tile_list_node_t*)head,(int*)xy,value);
}
static int tile_list_fixture_point_411350(void* xy,int32_t selector) {
    tile_list_fn_3a0 fn=&sub_4113A0;return fn((int*)xy,selector);
}
static int tile_list_fixture_edge_411350(int32_t edge,int32_t variation) {
    tile_list_fn_490 fn=&sub_411490;return fn(edge,variation);
}
static size_t tile_list_fixture_layout_411350(int index) {
    switch(index) {
    case 0:return offsetof(nox_tile_list_node_t,field_0);
    case 1:return offsetof(nox_tile_list_node_t,field_1);
    case 2:return offsetof(nox_tile_list_node_t,field_2);
    case 3:return offsetof(nox_tile_list_node_t,field_3);
    case 4:return offsetof(nox_tile_list_node_t,next);
    case 5:return sizeof(nox_tile_list_node_t);
    case 6:return sizeof(void*);
    default:return (size_t)-1;
    }
}
static void tile_list_fixture_store_pointer_411350(void* slot,void* pointer) {
    memcpy(slot,&pointer,sizeof(pointer));
}
static void* tile_list_fixture_load_pointer_411350(const void* slot) {
    void* pointer;memcpy(&pointer,slot,sizeof(pointer));return pointer;
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func tileListFixtureList411350(head, xy unsafe.Pointer, initial int32) int32 {
	value := int32(C.tile_list_fixture_list_411350(head, xy, C.int32_t(initial)))
	runtime.KeepAlive(head)
	runtime.KeepAlive(xy)
	return value
}
func tileListFixturePoint411350(xy unsafe.Pointer, selector int32) int32 {
	value := int32(C.tile_list_fixture_point_411350(xy, C.int32_t(selector)))
	runtime.KeepAlive(xy)
	return value
}
func tileListFixtureEdge411350(edge, variation int32) int32 {
	return int32(C.tile_list_fixture_edge_411350(C.int32_t(edge), C.int32_t(variation)))
}
func tileListFixtureLayout411350() [7]uintptr {
	var values [7]uintptr
	for i := range values {
		values[i] = uintptr(C.tile_list_fixture_layout_411350(C.int(i)))
	}
	return values
}
func tileListFixtureStorePointer411350(slot, pointer unsafe.Pointer) {
	C.tile_list_fixture_store_pointer_411350(slot, pointer)
	runtime.KeepAlive(slot)
	runtime.KeepAlive(pointer)
}
func tileListFixtureLoadPointer411350(slot unsafe.Pointer) unsafe.Pointer {
	pointer := C.tile_list_fixture_load_pointer_411350(slot)
	runtime.KeepAlive(slot)
	return pointer
}
