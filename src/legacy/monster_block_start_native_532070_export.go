package legacy

/*
#include "monster_block_start_native_532070.h"
_Static_assert(_Generic(&nox_xxx_monsterShieldBlockStart_532070, char (*)(nox_object_t*): 1, default: 0), "native shield block start ABI");
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_monsterShieldBlockStart_532070
func nox_xxx_monsterShieldBlockStart_532070(unit *C.nox_object_t) C.char {
	return C.char(GetServer().S().MonsterActionBlockAttackRaw532070(asObjectS((*nox_object_t)(unit)), server.MonsterActionBlockRuntime532070{
		TestShield: Nox_xxx_monsterTestBlockShield_533E70,
	}))
}

func monsterBlockStartPublicC532070() unsafe.Pointer {
	return C.nox_xxx_monsterShieldBlockStart_532070
}
