package legacy

/*
#include "GAME1_1.h"
typedef const char nox_const_ability_name;
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func abilityIDBySaveName(name string) server.Ability {
	for index := range server.AbilityMax {
		namePtr := (*C.char)(*memmap.PtrPtr(0x587000, 69736+4*uintptr(index)))
		if namePtr == nil {
			break
		}
		if GoString(namePtr) == name {
			return index
		}
	}
	return server.AbilityInvalid
}

//export nox_xxx_abilityNameToN_424D80
func nox_xxx_abilityNameToN_424D80(name *C.nox_const_ability_name) C.int {
	if name == nil {
		return 0
	}
	return C.int(abilityIDBySaveName(GoString((*C.char)(name))))
}
