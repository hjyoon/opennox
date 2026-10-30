package legacy

/*
#include "monster_random_walk_native_545020.h"
_Static_assert(_Generic(&nox_xxx_mobActionRandomWalk_545020, void (*)(nox_object_t*): 1, default: 0), "native random walk void update ABI");
_Static_assert(_Generic(&sub_545090, int (*)(nox_object_t*): 1, default: 0), "native direction public int ABI");
_Static_assert(_Generic(&nox_xxx_mobActionConfuse_545140, void (*)(nox_object_t*): 1, default: 0), "native confused void update ABI");
*/
import "C"
import "unsafe"

//export nox_xxx_mobActionRandomWalk_545020
func nox_xxx_mobActionRandomWalk_545020(unit *C.nox_object_t) {
	GetServer().S().MonsterRandomWalkMoveNative545020(asObjectS((*nox_object_t)(unit)), Nox_xxx_tileNFromPoint_411160)
}

//export sub_545090
func sub_545090(unit *C.nox_object_t) C.int {
	return C.int(GetServer().S().MonsterRandomWalkDirectionNative545090(asObjectS((*nox_object_t)(unit)), Nox_xxx_tileNFromPoint_411160))
}

//export nox_xxx_mobActionConfuse_545140
func nox_xxx_mobActionConfuse_545140(unit *C.nox_object_t) {
	GetServer().S().MonsterActionConfused545140(asObjectS((*nox_object_t)(unit)), Nox_xxx_tileNFromPoint_411160)
}

func monsterRandomWalkPublicC545020() unsafe.Pointer      { return C.nox_xxx_mobActionRandomWalk_545020 }
func monsterRandomDirectionPublicC545090() unsafe.Pointer { return C.sub_545090 }
func monsterConfusedPublicC545140() unsafe.Pointer        { return C.nox_xxx_mobActionConfuse_545140 }
