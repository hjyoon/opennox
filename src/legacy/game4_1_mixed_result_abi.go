package legacy

/*
#include "GAME4_1.h"
#include "monster_look_at_5125a0.h"

extern int nox_xxx_pendingOwnAdd_native_516F90(int owner_script_id, int owned_script_id);

// These entry points retain the pointer-returning C ABI, including angle
// and success words. The Go implementations carry those words as integers.
float* nox_xxx_monsterLookAt_5125A0(nox_object_t* obj, int direction) {
	return (float*)nox_server_monster_look_at_5125a0(obj, direction);
}

uint32_t* sub_516F90(int owner_script_id, int owned_script_id) {
	uintptr_t added = nox_xxx_pendingOwnAdd_native_516F90(owner_script_id, owned_script_id) != 0;
	return (uint32_t*)added;
}
*/
import "C"
