package legacy

/*
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME4_1.h"
typedef int(*global_state_value_callback_C)(void);
typedef void(*global_state_void_callback_C)(void);
_Static_assert(_Generic(&nox_xxx_clientIsObserver_4372E0,global_state_value_callback_C:1,default:0),"global entry ABI nox_xxx_clientIsObserver_4372E0");
_Static_assert(_Generic(&sub_467590,global_state_value_callback_C:1,default:0),"global entry ABI sub_467590");
_Static_assert(_Generic(&sub_4678B0,global_state_value_callback_C:1,default:0),"global entry ABI sub_4678B0");
_Static_assert(_Generic(&sub_460D40,global_state_value_callback_C:1,default:0),"global entry ABI sub_460D40");
_Static_assert(_Generic(&sub_4678C0,global_state_value_callback_C:1,default:0),"global entry ABI sub_4678C0");
_Static_assert(_Generic(&sub_4BFD30,global_state_value_callback_C:1,default:0),"global entry ABI sub_4BFD30");
_Static_assert(_Generic(&sub_4C12C0,global_state_value_callback_C:1,default:0),"global entry ABI sub_4C12C0");
_Static_assert(_Generic(&sub_479870,global_state_value_callback_C:1,default:0),"global entry ABI sub_479870");
_Static_assert(_Generic(&sub_49C810,global_state_value_callback_C:1,default:0),"global entry ABI sub_49C810");
_Static_assert(_Generic(&sub_49CB40,global_state_value_callback_C:1,default:0),"global entry ABI sub_49CB40");
_Static_assert(_Generic(&sub_4C5050,global_state_void_callback_C:1,default:0),"global entry ABI sub_4C5050");
_Static_assert(_Generic(&sub_481410,global_state_void_callback_C:1,default:0),"global entry ABI sub_481410");
_Static_assert(_Generic(&sub_51D0E0,global_state_void_callback_C:1,default:0),"global entry ABI sub_51D0E0");
_Static_assert(_Generic(&nox_client_mapZoomIn_4724E0,global_state_void_callback_C:1,default:0),"global entry ABI nox_client_mapZoomIn_4724E0");
_Static_assert(_Generic(&nox_client_mapZoomOut_472500,global_state_void_callback_C:1,default:0),"global entry ABI nox_client_mapZoomOut_472500");
_Static_assert(_Generic(&sub_470CC0,global_state_value_callback_C:1,default:0),"global entry ABI sub_470CC0");
_Static_assert(_Generic(&sub_470CD0,global_state_value_callback_C:1,default:0),"global entry ABI sub_470CD0");
static global_state_value_callback_C volatile global_state_values_C[]={nox_xxx_clientIsObserver_4372E0,sub_467590,sub_4678B0,sub_460D40,sub_4678C0,sub_4BFD30,sub_4C12C0,sub_479870,sub_49C810,sub_49CB40,NULL,NULL,NULL,NULL,NULL,sub_470CC0,sub_470CD0};
static global_state_void_callback_C volatile global_state_voids_C[]={NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,sub_4C5050,sub_481410,sub_51D0E0,nox_client_mapZoomIn_4724E0,nox_client_mapZoomOut_472500,NULL,NULL};
static int global_state_call_C(int route,unsigned i){
if(route==1){if(i>=10&&i<=14){global_state_void_callback_C f=global_state_voids_C[i];f();return 0;}global_state_value_callback_C f=global_state_values_C[i];return f();}
switch(i){
case 0: return nox_xxx_clientIsObserver_4372E0();
case 1: return sub_467590();
case 2: return sub_4678B0();
case 3: return sub_460D40();
case 4: return sub_4678C0();
case 5: return sub_4BFD30();
case 6: return sub_4C12C0();
case 7: return sub_479870();
case 8: return sub_49C810();
case 9: return sub_49CB40();
case 10: sub_4C5050();return 0;
case 11: sub_481410();return 0;
case 12: sub_51D0E0();return 0;
case 13: nox_client_mapZoomIn_4724E0();return 0;
case 14: nox_client_mapZoomOut_472500();return 0;
case 15: return sub_470CC0();
case 16: return sub_470CD0();
default:return 0;}}
static void global_state_seed_cell_C(void*p,uint32_t word){((nox_inventory_cell_t*)p)->field_4=word;}
static void global_state_layout_C(uintptr_t*v){
v[0]=sizeof(void*);v[1]=sizeof(nox_playerInfo);v[2]=offsetof(nox_playerInfo,active);v[3]=sizeof(((nox_playerInfo*)0)->active);v[4]=sizeof(((nox_playerInfo*)0)->active_padding);
v[5]=offsetof(nox_playerInfo,field_3680);v[6]=sizeof(((nox_playerInfo*)0)->field_3680);v[7]=offsetof(nox_playerInfo,field_3684);v[8]=sizeof(((nox_playerInfo*)0)->field_3684);
v[9]=sizeof(nox_inventory_cell_t);v[10]=offsetof(nox_inventory_cell_t,field_4);v[11]=sizeof(((nox_inventory_cell_t*)0)->field_4);
v[12]=sizeof(nox_window_yyy);v[13]=offsetof(nox_window_yyy,field_1);v[14]=offsetof(nox_window_yyy,field_2);v[15]=sizeof(((nox_window_yyy*)0)->field_1);v[16]=sizeof(int);
}
*/
import "C"

import "unsafe"

func globalStateNativeLayoutC() (out [17]uintptr) {
	C.global_state_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}
func globalStateNativeWindows() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(clientMeterRecords)), int(unsafe.Sizeof(*clientMeterRecords)))
}
func globalStateNativeSeedCell(p unsafe.Pointer, word uint32) {
	C.global_state_seed_cell_C(p, C.uint32_t(word))
}
func globalStateNativeCall(route, index int) int32 {
	if route != 2 {
		return int32(C.global_state_call_C(C.int(route), C.uint(index)))
	}
	switch index {
	case 0:
		return int32(nox_xxx_clientIsObserver_4372E0())
	case 1:
		return int32(sub_467590())
	case 2:
		return int32(sub_4678B0())
	case 3:
		return int32(sub_460D40())
	case 4:
		return int32(sub_4678C0())
	case 5:
		return int32(sub_4BFD30())
	case 6:
		return int32(sub_4C12C0())
	case 7:
		return int32(sub_479870())
	case 8:
		return int32(sub_49C810())
	case 9:
		return int32(sub_49CB40())
	case 10:
		sub_4C5050()
		return 0
	case 11:
		sub_481410()
		return 0
	case 12:
		sub_51D0E0()
		return 0
	case 13:
		nox_client_mapZoomIn_4724E0()
		return 0
	case 14:
		nox_client_mapZoomOut_472500()
		return 0
	case 15:
		return int32(sub_470CC0())
	case 16:
		return int32(sub_470CD0())
	}
	panic("invalid global entry")
}
