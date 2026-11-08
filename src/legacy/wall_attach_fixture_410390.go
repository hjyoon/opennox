package legacy

/*
#include "GAME1.h"
#include "GAME1_1.h"
#include "map_temp_lists_503f40.h"
#include <stddef.h>
#include <stdint.h>
extern uint32_t nox_xxx_polygonNextIdx_587000_60352;
typedef void* (*wall_attach_fixture_fn_410390)(nox_drawable*,int,int);
typedef void* (*wall_attach_fixture_lookup_fn_410390)(int,int);
typedef nox_player_polygon_check_data* (*wall_attach_fixture_polygon_fn_410390)(int2*,int);
typedef int* (*wall_attach_fixture_fallback_fn_410390)(int2*,float,int);
_Static_assert(_Generic(&sub_410390,wall_attach_fixture_fn_410390:1,default:0),"410390 public pointer(drawable*,int,int) ABI");
_Static_assert(_Generic(&nox_xxx_wall_4105E0,wall_attach_fixture_lookup_fn_410390:1,default:0),"4105E0 pointer(int,int) ABI");
_Static_assert(_Generic(&nox_xxx_wallCreateAt_410250,wall_attach_fixture_lookup_fn_410390:1,default:0),"410250 pointer(int,int) ABI");
_Static_assert(_Generic(&nox_xxx_polygonIsPlayerInPolygon_4217B0,wall_attach_fixture_polygon_fn_410390:1,default:0),"4217B0 pointer(int2*,int) ABI");
_Static_assert(_Generic(&sub_421990,wall_attach_fixture_fallback_fn_410390:1,default:0),"421990 pointer(int2*,float,int) ABI");
_Static_assert(sizeof(int)==4 && sizeof(float)==4 && sizeof(int2)==8 && offsetof(int2,field_4)==4,"410390 scalar/local8 ABI");
static void* wall_attach_fixture_call_410390(void* p,int x,int y) {
    wall_attach_fixture_fn_410390 fn=&sub_410390;
    return fn((nox_drawable*)p,x,y);
}
static size_t wall_attach_fixture_offset_410390(int field) {
    switch(field) {
    case 0:return offsetof(nox_drawable,pos.x);
    case 1:return offsetof(nox_drawable,pos.y);
    case 2:return offsetof(nox_map_wall_504290,flags);
    case 3:return offsetof(nox_map_wall_504290,client_data);
    case 4:return offsetof(nox_map_wall_504290,field_8);
    default:return (size_t)-1;
    }
}
static size_t wall_attach_fixture_width_410390(int field) {
    switch(field) {
    case 0:return sizeof(((nox_drawable*)0)->pos.x);
    case 1:return sizeof(((nox_drawable*)0)->pos.y);
    case 2:return sizeof(((nox_map_wall_504290*)0)->flags);
    case 3:return sizeof(((nox_map_wall_504290*)0)->client_data);
    case 4:return 1; // The original accesses only the low BYTE of field_8.
    default:return 0;
    }
}
static size_t wall_attach_fixture_pointer_width_410390(void) {return sizeof(void*);}
static void* wall_attach_fixture_polygon_count_cell_410390(void) {return &nox_xxx_polygonNextIdx_587000_60352;}
*/
import "C"
import "unsafe"

type wallAttachFixtureField410390 struct{ Offset, Width uintptr }

func wallAttachFixtureCall410390(drawable unsafe.Pointer, x, y int32) unsafe.Pointer {
	return C.wall_attach_fixture_call_410390(drawable, C.int(x), C.int(y))
}
func wallAttachFixtureCLayout410390() [5]wallAttachFixtureField410390 {
	var out [5]wallAttachFixtureField410390
	for i := range out {
		out[i] = wallAttachFixtureField410390{uintptr(C.wall_attach_fixture_offset_410390(C.int(i))), uintptr(C.wall_attach_fixture_width_410390(C.int(i)))}
	}
	return out
}
func wallAttachFixturePointerWidth410390() uintptr {
	return uintptr(C.wall_attach_fixture_pointer_width_410390())
}

// Controls save and restore the actual existing C owner, never a shadow.
func wallAttachFixturePolygonCountCell410390() unsafe.Pointer {
	return C.wall_attach_fixture_polygon_count_cell_410390()
}
