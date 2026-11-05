package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5.h"
typedef struct {
 uint32_t scan,fire;uint8_t state,pad9[3];uint32_t projectile;
 char name[32];uint8_t enabled,pad49[3];
} skull_original_data_54f9a0;
typedef struct {int32_t damage,field4;} skull_original_collide_54f9a0;
_Static_assert(sizeof(skull_original_data_54f9a0)==52,"independent Skull update width");
_Static_assert(sizeof(skull_original_collide_54f9a0)==8,"independent projectile collide width");
static int skull_old_int_54f9a0(int(*f)(int),nox_object_t* u){return f((int)(uintptr_t)u);}
static int skull_native_int_54f9a0(int(*f)(nox_object_t*),nox_object_t* u){return f(u);}
static void skull_old_void_54f9a0(void(*f)(int),nox_object_t* u){f((int)(uintptr_t)u);}
static void skull_native_void_54f9a0(void(*f)(nox_object_t*),nox_object_t* u){f(u);}
static void skull_old_num_54f9a0(void(*f)(int,int),nox_object_t* u,uint32_t n){f((int)(uintptr_t)u,(int32_t)n);}
static void skull_native_num_54f9a0(void(*f)(nox_object_t*,int),nox_object_t* u,uint32_t n){f(u,(int32_t)n);}
static void skull_old_pair_54f9a0(void(*f)(int,int),nox_object_t* a,nox_object_t* b){f((int)(uintptr_t)a,(int)(uintptr_t)b);}
static void skull_native_pair_54f9a0(void(*f)(nox_object_t*,nox_object_t*),nox_object_t* a,nox_object_t* b){f(a,b);}
static uint32_t skull_public_54f9a0(int entry,nox_object_t* source,nox_object_t* target,uint32_t id){
 switch(entry){
 case 0:return (uint32_t)_Generic(&nox_xxx_updateShootingTrap_54F9A0,int(*)(int):skull_old_int_54f9a0,int(*)(nox_object_t*):skull_native_int_54f9a0)(&nox_xxx_updateShootingTrap_54F9A0,source);
 case 1:_Generic(&nox_xxx_createArrowTrapProjectile_54FA80,void(*)(int,int):skull_old_num_54f9a0,void(*)(nox_object_t*,int):skull_native_num_54f9a0)(&nox_xxx_createArrowTrapProjectile_54FA80,source,id);break;
 case 2:_Generic(&sub_54FBB0,void(*)(int):skull_old_void_54f9a0,void(*)(nox_object_t*):skull_native_void_54f9a0)(&sub_54FBB0,source);break;
 case 3:return (uint32_t)_Generic(&sub_54FBF0,int(*)(int):skull_old_int_54f9a0,int(*)(nox_object_t*):skull_native_int_54f9a0)(&sub_54FBF0,source);
 case 4:_Generic(&nox_xxx_unitIsAttackReachable_54FC50,void(*)(int,int):skull_old_pair_54f9a0,void(*)(nox_object_t*,nox_object_t*):skull_native_pair_54f9a0)(&nox_xxx_unitIsAttackReachable_54FC50,target,source);break;
 }
 return 0;
}
static int skull_abi_54f9a0(int e){
 switch(e){
 case 0:return _Generic(&nox_xxx_updateShootingTrap_54F9A0,int(*)(int):32,int(*)(nox_object_t*):64);
 case 1:return _Generic(&nox_xxx_createArrowTrapProjectile_54FA80,void(*)(int,int):32,void(*)(nox_object_t*,int):64);
 case 2:return _Generic(&sub_54FBB0,void(*)(int):32,void(*)(nox_object_t*):64);
 case 3:return _Generic(&sub_54FBF0,int(*)(int):32,int(*)(nox_object_t*):64);
 default:return _Generic(&nox_xxx_unitIsAttackReachable_54FC50,void(*)(int,int):32,void(*)(nox_object_t*,nox_object_t*):64);
 }
}
static uintptr_t skull_layout_54f9a0(int i){
 switch(i){
 case 0:return sizeof(nox_object_t);case 1:return offsetof(nox_object_t,typ_ind);
 case 2:return offsetof(nox_object_t,obj_class);case 3:return offsetof(nox_object_t,obj_flags);
 case 4:return offsetof(nox_object_t,x);case 5:return offsetof(nox_object_t,y);
 case 6:return offsetof(nox_object_t,direction1);case 7:return offsetof(nox_object_t,direction2);
 case 8:return offsetof(nox_object_t,shape)+offsetof(nox_shape,circle_r);
 case 9:return offsetof(nox_object_t,speed_cur);case 10:return offsetof(nox_object_t,data_update);
 case 11:return offsetof(nox_object_t,collide_data);case 12:return sizeof(skull_original_data_54f9a0);
 case 13:return offsetof(skull_original_data_54f9a0,enabled);case 14:return offsetof(skull_original_data_54f9a0,projectile);
 case 15:return offsetof(skull_original_data_54f9a0,state);case 16:return sizeof(skull_original_collide_54f9a0);
 case 17:return offsetof(skull_original_collide_54f9a0,damage);default:return offsetof(skull_original_data_54f9a0,name);
 }
}
static void skull_control_54f9a0(nox_object_t* u,uintptr_t* p,uint32_t* w){
 p[0]=(uintptr_t)u;p[1]=(uintptr_t)u->data_update;p[2]=(uintptr_t)u->collide_data;
 p[3]=(uintptr_t)&u->x;p[4]=(uintptr_t)&u->y;p[5]=(uintptr_t)&u->shape.circle_r;p[6]=(uintptr_t)&u->speed_cur;
 w[0]=u->typ_ind;w[1]=u->obj_class;w[2]=u->obj_flags;memcpy(w+3,&u->x,8);
 w[5]=u->direction1;w[6]=u->direction2;memcpy(w+7,&u->shape.circle_r,4);memcpy(w+8,&u->speed_cur,4);
 if(u->data_update){skull_original_data_54f9a0* d=u->data_update;w[9]=d->scan;w[10]=d->fire;w[11]=d->state;w[12]=d->projectile;w[13]=d->enabled;}
 if(u->collide_data){skull_original_collide_54f9a0* d=u->collide_data;w[14]=(uint32_t)d->damage;w[15]=(uint32_t)d->field4;}
}
static uint32_t skull_type_54f9a0(const char* name){return (uint32_t)nox_xxx_getNameId_4E3AA0(name);}
static nox_object_t* skull_new_54f9a0(uint32_t id){return nox_xxx_newObjectWithTypeInd_4E3450((int32_t)id);}
static void skull_audio_54f9a0(nox_object_t* u){nox_xxx_aud_501960(889,u,0,0);}
static int skull_direction_54f9a0(nox_object_t* source,nox_object_t* target){return nox_server_testTwoPointsAndDirection_4E6E50((float2*)&source->x,(int16_t)source->direction1,(float2*)&target->x);}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func skullOriginalC54F9A0(entry int, source, target *server.Object, projectile uint32) uint32 {
	return uint32(C.skull_public_54f9a0(C.int(entry), asObjectC(source), asObjectC(target), C.uint32_t(projectile)))
}

// The marker distinguishes the signed32 and typed-pointer public prototypes
// on both host widths; it is not a measurement of sizeof(void*).
func skullOriginalABI54F9A0(entry int) int {
	return int(C.skull_abi_54f9a0(C.int(entry)))
}

func skullOriginalLayout54F9A0() (out [19]uintptr) {
	for i := range out {
		out[i] = uintptr(C.skull_layout_54f9a0(C.int(i)))
	}
	return
}

func skullOriginalControl54F9A0(source *server.Object) (pointers [7]uintptr, words [16]uint32) {
	var p [7]C.uintptr_t
	var w [16]C.uint32_t
	C.skull_control_54f9a0(asObjectC(source), &p[0], &w[0])
	for i := range p {
		pointers[i] = uintptr(p[i])
	}
	for i := range w {
		words[i] = uint32(w[i])
	}
	return
}

func skullOriginalType54F9A0(name string) uint32 {
	return uint32(C.skull_type_54f9a0(internCStr(name)))
}

func skullOriginalNew54F9A0(id uint32) *server.Object {
	return asObjectS(C.skull_new_54f9a0(C.uint32_t(id)))
}

func skullOriginalAudioControl54F9A0(source *server.Object) {
	C.skull_audio_54f9a0(asObjectC(source))
}

func skullOriginalDirectionControl54F9A0(source, target *server.Object) uint32 {
	return uint32(C.skull_direction_54f9a0(asObjectC(source), asObjectC(target)))
}
