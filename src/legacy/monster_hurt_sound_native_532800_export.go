package legacy

/*
#include "monster_hurt_sound_native_532800.h"
_Static_assert(_Generic(&nox_xxx_monsterPlayHurtSound_532800, void (*)(nox_object_t*): 1, default: 0), "native HurtSound void update ABI");
*/
import "C"
import "unsafe"

//export nox_xxx_monsterPlayHurtSound_532800
func nox_xxx_monsterPlayHurtSound_532800(unit *C.nox_object_t) {
	GetServer().S().MonsterHurtSound532800(asObjectS((*nox_object_t)(unit)))
}

func monsterHurtSoundPublicC532800() unsafe.Pointer { return C.nox_xxx_monsterPlayHurtSound_532800 }
