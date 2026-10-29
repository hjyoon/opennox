package legacy

/*
#include "monster_terminal_native_5320e0.h"
_Static_assert(_Generic(&nox_xxx_monsterShieldBlockStop_5320E0, char (*)(nox_object_t*): 1, default: 0), "native shield finish ABI");
_Static_assert(_Generic(&sub_532110, char (*)(nox_object_t*): 1, default: 0), "native weapon block ABI");
_Static_assert(_Generic(&sub_544D60, char (*)(nox_object_t*): 1, default: 0), "native dying update ABI");
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_monsterShieldBlockStop_5320E0
func nox_xxx_monsterShieldBlockStop_5320E0(unit *C.nox_object_t) C.char {
	return C.char(server.MonsterActionTerminal5320E0(asObjectS((*nox_object_t)(unit))))
}

//export sub_532110
func sub_532110(unit *C.nox_object_t) C.char {
	return C.char(server.MonsterActionTerminal5320E0(asObjectS((*nox_object_t)(unit))))
}

//export sub_544D60
func sub_544D60(unit *C.nox_object_t) C.char {
	return C.char(server.MonsterActionTerminal5320E0(asObjectS((*nox_object_t)(unit))))
}

// The helper only exposes the original public addresses to native ABI tests.
func monsterTerminalPublicCEntry5320E0(mode string) unsafe.Pointer {
	switch mode {
	case "shield-finish":
		return C.nox_xxx_monsterShieldBlockStop_5320E0
	case "weapon-block":
		return C.sub_532110
	case "dying":
		return C.sub_544D60
	}
	return nil
}
