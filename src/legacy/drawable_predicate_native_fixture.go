package legacy

/*
#include "GAME2_1.h"
#include "client__draw__animdraw.h"
int nox_thing_static_draw(uint32_t*,nox_drawable*);
int nox_thing_static_random_draw(uint32_t*,nox_drawable*);
typedef int(*drawable_predicate_C)(nox_drawable*);
typedef int(*drawable_predicate_draw_C)(uint32_t*,nox_drawable*);
_Static_assert(_Generic(&nox_xxx_sprite_4756E0_drawable,drawable_predicate_C:1,default:0),"back-wall predicate ABI");
_Static_assert(_Generic(&nox_xxx_sprite_475740_drawable,drawable_predicate_C:1,default:0),"front predicate ABI");
_Static_assert(_Generic(&nox_xxx_sprite_4757A0_drawable,drawable_predicate_C:1,default:0),"overlay predicate ABI");
_Static_assert(_Generic(&sub_4757D0_drawable,drawable_predicate_C:1,default:0),"world predicate ABI");
static drawable_predicate_C volatile drawable_predicate_callbacks_C[]={nox_xxx_sprite_4756E0_drawable,nox_xxx_sprite_475740_drawable,nox_xxx_sprite_4757A0_drawable,sub_4757D0_drawable};
static int drawable_predicate_call_C(unsigned route,unsigned index,nox_drawable*dr){
 if(route){if(index>=4)abort();drawable_predicate_C f=drawable_predicate_callbacks_C[index];return f(dr);}
 switch(index){case 0:return nox_xxx_sprite_4756E0_drawable(dr);case 1:return nox_xxx_sprite_475740_drawable(dr);case 2:return nox_xxx_sprite_4757A0_drawable(dr);case 3:return sub_4757D0_drawable(dr);default:abort();}
}
static void*drawable_predicate_role_C(unsigned role){
 switch(role){case 0:return NULL;case 1:return(void*)nox_thing_static_draw;case 2:return(void*)nox_thing_static_random_draw;case 3:return(void*)nox_thing_animate_draw;case 4:return(void*)nox_thing_animate_state_draw;default:abort();}
}
static void drawable_predicate_seed_C(void*p,size_t placement,const uint32_t*q){
 unsigned char*raw=p;unsigned char*owner=raw+0x1000+placement;nox_drawable*dr=(nox_drawable*)owner;
 memcpy(owner+112,q+3,4);memcpy(owner+120,q+4,4);memcpy(owner+300,q+5,4);
 dr->flags28=q[0];dr->flags30=q[1];dr->draw_func=(drawable_predicate_draw_C)drawable_predicate_role_C(q[2]);
}
static void drawable_predicate_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_drawable);v[2]=offsetof(nox_drawable,flags28);v[3]=offsetof(nox_drawable,flags30);v[4]=offsetof(nox_drawable,draw_func);
 v[5]=sizeof(((nox_drawable*)0)->flags28);v[6]=sizeof(((nox_drawable*)0)->flags30);v[7]=sizeof(((nox_drawable*)0)->draw_func);v[8]=offsetof(nox_drawable,field_127);v[9]=sizeof(((nox_drawable*)0)->field_127);
 v[10]=sizeof(int);v[11]=sizeof(drawable_predicate_C);v[12]=sizeof(drawable_predicate_draw_C);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func drawablePredicateNativeLayoutC() (out [13]uintptr) {
	C.drawable_predicate_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func drawablePredicateNativeRoleC(role int) unsafe.Pointer {
	return C.drawable_predicate_role_C(C.uint(role))
}

func drawablePredicateNativeSeedC(raw unsafe.Pointer, placement int, input *[6]uint32) {
	C.drawable_predicate_seed_C(raw, C.size_t(placement), (*C.uint32_t)(unsafe.Pointer(&input[0])))
}

func drawablePredicateNativeCall(route, index int, owner unsafe.Pointer) int32 {
	if route != 2 {
		return int32(C.drawable_predicate_call_C(C.uint(route), C.uint(index), (*C.nox_drawable)(owner)))
	}
	dr := (*client.Drawable)(owner)
	var value bool
	switch index {
	case 0:
		value = client.DrawableBackWallQueue(dr, drawablePredicateStatic(dr))
	case 1:
		value = client.DrawableFrontQueue(dr, drawablePredicateStatic(dr))
	case 2:
		value = client.DrawableOverlayQueue(dr)
	case 3:
		value = client.DrawableWorldQueue(dr)
	default:
		panic("invalid Drawable queue predicate")
	}
	if value {
		return 1
	}
	return 0
}
