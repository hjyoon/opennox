package legacy

/*
#include "GAME4_2.h"
#include <stddef.h>
typedef int (*predict_linear_callback_C)(nox_object_t*);
_Static_assert(_Generic(&nox_xxx_netClientPredictLinear_523530,predict_linear_callback_C:1,default:0),"native prediction Object ABI");
static void predict_linear_layout_C(uintptr_t* v) {
 v[0]=sizeof(nox_object_t);v[1]=offsetof(nox_object_t,typ_ind);v[2]=offsetof(nox_object_t,x);v[3]=offsetof(nox_object_t,y);
 v[4]=offsetof(nox_object_t,float_28);v[5]=offsetof(nox_object_t,vel_x);v[6]=offsetof(nox_object_t,vel_y);
 v[7]=offsetof(nox_object_t,direction1);v[8]=offsetof(nox_object_t,data_update);v[9]=sizeof(nox_player_update_data_t);
 v[10]=offsetof(nox_player_update_data_t,player);v[11]=offsetof(nox_playerInfo,playerInd);v[12]=sizeof(void*);
 v[13]=sizeof(((nox_object_t*)0)->typ_ind);v[14]=sizeof(((nox_playerInfo*)0)->playerInd);
 v[15]=offsetof(nox_object_t,obj_class);v[16]=offsetof(nox_object_t,net_code);v[17]=offsetof(nox_object_t,extent);
}
static void predict_linear_context_C(nox_object_t* obj,uint32_t class_bits,uint32_t code,uint32_t extent) {
 obj->obj_class=class_bits;obj->net_code=code;obj->extent=extent;
}
static int predict_linear_call_C(int route,nox_object_t* obj) {
 if(route==1){predict_linear_callback_C f=nox_xxx_netClientPredictLinear_523530;return f(obj);}
 if(route==3){float* v9=(float*)obj;return nox_xxx_netClientPredictLinear_523530((nox_object_t*)v9);}
 return nox_xxx_netClientPredictLinear_523530(obj);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func predictLinearLayoutC() (out [18]uintptr) {
	C.predict_linear_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func predictLinearActualCall(route int, obj *server.Object) int32 {
	if route == 2 {
		return Nox_xxx_netClientPredictLinear_523530(obj)
	}
	return int32(C.predict_linear_call_C(C.int(route), asObjectC(obj)))
}

func predictLinearContextC(obj *server.Object, classBits, code, extent uint32) {
	C.predict_linear_context_C(asObjectC(obj), C.uint32_t(classBits), C.uint32_t(code), C.uint32_t(extent))
}
