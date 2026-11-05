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

// Independently declared complete PlayerUpdate allocation, including the tail
// after the existing native C prefix. No old PE32 offsets are used as native offsets.
typedef struct {
 nox_player_update_data_t prefix;
 uint32_t lives,quest_state[32];
 uint8_t respawn[32],flags_a[32],flags_b[32];
 uint32_t frame137,frame138;
} special_death_player_update_54e460;
_Static_assert(sizeof(special_death_player_update_54e460)==(sizeof(void*)==4?556:656),"complete PlayerUpdate allocation");

static void special_death_old_54e460(void(*fn)(int),nox_object_t* u){fn((int)(uintptr_t)u);}
static void special_death_native_54e460(void(*fn)(nox_object_t*),nox_object_t* u){fn(u);}
#define SPECIAL_DEATH_CALL_54E460(fn,u) _Generic(&(fn),void(*)(int):special_death_old_54e460,void(*)(nox_object_t*):special_death_native_54e460)(&(fn),u)
#define SPECIAL_DEATH_ABI_54E460(fn) _Generic(&(fn),void(*)(int):32,void(*)(nox_object_t*):64)
static void special_death_public_54e460(int entry,nox_object_t* u){
 switch(entry){
 case 0:SPECIAL_DEATH_CALL_54E460(nox_xxx_dieMarker_54E460,u);break;
 case 1:SPECIAL_DEATH_CALL_54E460(nox_xxx_dieBoulder_54E4B0,u);break;
 case 2:SPECIAL_DEATH_CALL_54E460(nox_xxx_dieMonsterGen_54E630,u);break;
 case 3:SPECIAL_DEATH_CALL_54E460(sub_4D61B0,u);break;
 }
}
static int special_death_abi_54e460(int entry){
 switch(entry){
 case 0:return SPECIAL_DEATH_ABI_54E460(nox_xxx_dieMarker_54E460);
 case 1:return SPECIAL_DEATH_ABI_54E460(nox_xxx_dieBoulder_54E4B0);
 case 2:return SPECIAL_DEATH_ABI_54E460(nox_xxx_dieMonsterGen_54E630);
 default:return SPECIAL_DEATH_ABI_54E460(sub_4D61B0);
 }
}
static uintptr_t special_death_layout_54e460(int i){
 switch(i){
 case 0:return sizeof(nox_object_t);
 case 1:return offsetof(nox_object_t,obj_class);
 case 2:return offsetof(nox_object_t,obj_flags);
 case 3:return offsetof(nox_object_t,x);
 case 4:return offsetof(nox_object_t,y);
 case 5:return offsetof(nox_object_t,field_27);
 case 6:return offsetof(nox_object_t,field_29);
 case 7:return offsetof(nox_object_t,owner);
 case 8:return offsetof(nox_object_t,obj_130);
 case 9:return offsetof(nox_object_t,data_update);
 case 10:return sizeof(special_death_player_update_54e460);
 case 11:return offsetof(nox_player_update_data_t,field_29);
 case 12:return offsetof(nox_player_update_data_t,player);
 case 13:return sizeof(nox_monster_generator_update_data_t);
 case 14:return offsetof(nox_monster_generator_update_data_t,pre_collision)+8;
 case 15:return offsetof(nox_monster_generator_update_data_t,pre_collision)+12;
 case 16:return sizeof(nox_playerInfo);
 case 17:return offsetof(nox_playerInfo,field_4668);
 case 18:return offsetof(nox_playerInfo,field_4692);
 case 19:return offsetof(special_death_player_update_54e460,lives);
 case 20:return offsetof(special_death_player_update_54e460,frame137);
 case 21:return offsetof(nox_monster_generator_update_data_t,pre_collision);
 case 22:return offsetof(nox_monster_generator_update_data_t,pre_collision)+4;
 case 23:return offsetof(nox_monster_generator_update_data_t,pre_collision)+16;
 case 24:return offsetof(nox_monster_generator_update_data_t,pre_collision)+20;
 case 25:return offsetof(nox_monster_generator_update_data_t,collision_flags);
 case 26:return sizeof(((nox_monster_generator_update_data_t*)0)->objects);
 case 27:return sizeof(((nox_player_update_data_t*)0)->field_29);
 default:return sizeof(void*);
 }
}
static void special_death_object_54e460(nox_object_t* u,uintptr_t* p,uint32_t* w){
 p[0]=(uintptr_t)u;p[1]=(uintptr_t)u->data_update;p[2]=(uintptr_t)u->owner;p[3]=(uintptr_t)u->obj_130;
 p[4]=(uintptr_t)&u->x;p[5]=(uintptr_t)&u->y;p[6]=(uintptr_t)&u->field_27;p[7]=(uintptr_t)&u->field_29;
 w[0]=u->typ_ind;w[1]=u->obj_class;w[2]=u->obj_flags;memcpy(w+3,&u->x,8);memcpy(w+5,&u->field_27,4);w[6]=u->field_29;
}
static void special_death_player_54e460(nox_object_t* u,uintptr_t* p,uint32_t* w){
 special_death_player_update_54e460* full=u->data_update;
 nox_player_update_data_t* d=&full->prefix;nox_playerInfo* player=d->player;
 p[0]=(uintptr_t)d;p[1]=(uintptr_t)player;
 for(int i=0;i<4;i++)p[2+i]=(uintptr_t)d->field_29[i];
 p[6]=(uintptr_t)&player->field_4668;p[7]=(uintptr_t)&player->field_4692;
 p[8]=(uintptr_t)&full->lives;p[9]=(uintptr_t)&full->frame137;
 w[0]=player->field_4668;w[1]=player->field_4692;w[2]=full->lives;w[3]=full->frame137;
}
static void special_death_generator_54e460(nox_object_t* u,uintptr_t* p,uint32_t* w){
 nox_monster_generator_update_data_t* d=u->data_update;
 p[0]=(uintptr_t)d;p[1]=(uintptr_t)&d->pre_collision[2];p[2]=(uintptr_t)&d->pre_collision[3];p[3]=(uintptr_t)d->objects[0];
 for(int i=0;i<4;i++)w[i]=d->pre_collision[2+i];
}
static nox_object_t* special_death_parent_54e460(nox_object_t* u){return nox_xxx_findParentChainPlayer_4EC580(u);}
static uint32_t special_death_type_54e460(const char* name){return (uint32_t)nox_xxx_getNameId_4E3AA0(name);}
static nox_object_t* special_death_new_54e460(uint32_t id){return nox_xxx_newObjectWithTypeInd_4E3450((int32_t)id);}
static void special_death_audio_54e460(uint32_t id,nox_object_t* u){nox_xxx_aud_501960((int32_t)id,u,0,0);}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func specialDeathOriginalC54E460(entry int, source *server.Object) {
	C.special_death_public_54e460(C.int(entry), asObjectC(source))
}

func specialDeathOriginalABI54E460(entry int) int {
	return int(C.special_death_abi_54e460(C.int(entry)))
}

func specialDeathOriginalLayout54E460() (out [29]uintptr) {
	for i := range out {
		out[i] = uintptr(C.special_death_layout_54e460(C.int(i)))
	}
	return
}

func specialDeathOriginalObject54E460(unit *server.Object) (pointers [8]uintptr, words [7]uint32) {
	var p [8]C.uintptr_t
	var w [7]C.uint32_t
	C.special_death_object_54e460(asObjectC(unit), &p[0], &w[0])
	for i := range p {
		pointers[i] = uintptr(p[i])
	}
	for i := range w {
		words[i] = uint32(w[i])
	}
	return
}

func specialDeathOriginalPlayer54E460(unit *server.Object) (pointers [10]uintptr, words [4]uint32) {
	var p [10]C.uintptr_t
	var w [4]C.uint32_t
	C.special_death_player_54e460(asObjectC(unit), &p[0], &w[0])
	for i := range p {
		pointers[i] = uintptr(p[i])
	}
	for i := range w {
		words[i] = uint32(w[i])
	}
	return
}

func specialDeathOriginalGenerator54E460(unit *server.Object) (pointers [4]uintptr, words [4]uint32) {
	var p [4]C.uintptr_t
	var w [4]C.uint32_t
	C.special_death_generator_54e460(asObjectC(unit), &p[0], &w[0])
	for i := range p {
		pointers[i], words[i] = uintptr(p[i]), uint32(w[i])
	}
	return
}

func specialDeathOriginalParent54E460(unit *server.Object) *server.Object {
	return asObjectS(C.special_death_parent_54e460(asObjectC(unit)))
}

func specialDeathOriginalType54E460(name string) uint32 {
	return uint32(C.special_death_type_54e460(internCStr(name)))
}

func specialDeathOriginalNew54E460(id uint32) *server.Object {
	return asObjectS(C.special_death_new_54e460(C.uint32_t(id)))
}

func specialDeathOriginalAudio54E460(id uint32, unit *server.Object) {
	C.special_death_audio_54e460(C.uint32_t(id), asObjectC(unit))
}
