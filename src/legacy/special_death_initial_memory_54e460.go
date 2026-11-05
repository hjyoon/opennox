package legacy

/*
#include <stdint.h>
#include "defs.h"
_Static_assert(sizeof(((nox_player_update_data_t*)0)->field_29)==4*sizeof(void*),"native marker slots");
_Static_assert(sizeof(((nox_monster_generator_update_data_t*)0)->objects)==12*sizeof(void*),"native generator header");
static void special_initial_slots_54e460(void* data,uintptr_t* out){
 nox_player_update_data_t* d=data;
 for(int i=0;i<4;i++)out[i]=(uintptr_t)d->field_29[i];
}
static void special_initial_generator_54e460(void* data,uintptr_t* pointers,uint32_t* words){
 nox_monster_generator_update_data_t* d=data;
 for(int i=0;i<12;i++)pointers[i]=(uintptr_t)d->objects[i];
 for(int i=0;i<6;i++)words[i]=d->pre_collision[i];
}
*/
import "C"

import "unsafe"

func specialDeathInitialSlotsC54E460(data unsafe.Pointer) (out [4]uintptr) {
	var p [4]C.uintptr_t
	C.special_initial_slots_54e460(data, &p[0])
	for i := range out {
		out[i] = uintptr(p[i])
	}
	return
}

func specialDeathInitialGeneratorC54E460(data unsafe.Pointer) (pointers [12]uintptr, words [6]uint32) {
	var p [12]C.uintptr_t
	var w [6]C.uint32_t
	C.special_initial_generator_54e460(data, &p[0], &w[0])
	for i := range pointers {
		pointers[i] = uintptr(p[i])
	}
	for i := range words {
		words[i] = uint32(w[i])
	}
	return
}
