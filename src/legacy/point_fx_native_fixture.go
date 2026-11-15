package legacy

/*
#include "GAME4_2.h"
#include <stddef.h>
typedef int (*point_fx_point_C)(char,float2*);
typedef int (*point_fx_subtype_C)(char,char,float*);
typedef int (*point_fx_spark_C)(float*,char);
typedef void (*point_fx_break_C)(float*,char);
_Static_assert(_Generic(&nox_xxx_netSendPointFx_522FF0,point_fx_point_C:1,default:0),"point FX native ABI");
_Static_assert(_Generic(&sub_523150,point_fx_subtype_C:1,default:0),"subtype FX native ABI");
_Static_assert(_Generic(&nox_xxx_netSparkExplosionFx_5231B0,point_fx_spark_C:1,default:0),"spark FX native ABI");
_Static_assert(_Generic(&nox_xxx_sendGeneratorBreakFX_523200,point_fx_break_C:1,default:0),"break FX native ABI");
static void point_fx_layout_C(uintptr_t* out) {
 out[0]=sizeof(float2);out[1]=offsetof(float2,field_0);out[2]=offsetof(float2,field_4);
 out[3]=sizeof(float);out[4]=sizeof(char);out[5]=sizeof(int);
}
static int point_fx_call_C(int route,int root,char op,char subtype,float2* point) {
 if(route) {
  switch(root) {
   case 0:{point_fx_point_C f=nox_xxx_netSendPointFx_522FF0;return f(op,point);}
   case 1:{point_fx_subtype_C f=sub_523150;return f(op,subtype,(float*)point);}
   case 2:{point_fx_spark_C f=nox_xxx_netSparkExplosionFx_5231B0;return f((float*)point,subtype);}
   case 3:{point_fx_break_C f=nox_xxx_sendGeneratorBreakFX_523200;f((float*)point,subtype);return 0;}
  }
 } else {
  switch(root) {
   case 0:return nox_xxx_netSendPointFx_522FF0(op,point);
   case 1:return sub_523150(op,subtype,(float*)point);
   case 2:return nox_xxx_netSparkExplosionFx_5231B0((float*)point,subtype);
   case 3:nox_xxx_sendGeneratorBreakFX_523200((float*)point,subtype);return 0;
  }
 }
 return 0;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func pointFXLayoutC() (out [6]uintptr) {
	C.point_fx_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func pointFXActualCall(route, root int, op, subtype byte, point *types.Pointf) int32 {
	if route != 2 {
		return int32(C.point_fx_call_C(C.int(route), C.int(root), C.char(op), C.char(subtype), (*C.float2)(unsafe.Pointer(point))))
	}
	rt := pointFXRuntimeForCall522FF0()
	switch root {
	case 0:
		return server.NetSendPointFXNative522FF0(op, point, rt)
	case 1:
		return server.NetSendPointSubtypeFXNative523150(op, subtype, point, rt)
	case 2:
		return server.NetSparkExplosionFXNative5231B0(point, subtype, rt)
	case 3:
		server.SendGeneratorBreakFXNative523200(point, subtype, rt)
		return 0 // Observer only: the public entry has no defined return value.
	default:
		panic("unknown point FX root")
	}
}
