package legacy

/*
#include "GAME1_1.h"
#include "GAME4_2.h"
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func resourceSpellFloatToInt52DD50(value float32) int32 {
	return int32(C.nox_float2int(C.float(value)))
}

//export nox_xxx_castSpellWinkORrestoreHealth_52BF20
func nox_xxx_castSpellWinkORrestoreHealth_52BF20(spellID C.int, second unsafe.Pointer, caster *C.nox_object_t, fourth *C.nox_object_t, arg unsafe.Pointer, level C.int) C.int {
	return C.int(resourceRestoreHealth52BF20((*server.SpellAcceptArg)(arg), resourceSpellDepsForCall52BF20()))
}

//export sub_52BF50
func sub_52BF50(spellID C.int, second unsafe.Pointer, caster *C.nox_object_t, fourth *C.nox_object_t, arg unsafe.Pointer, level C.int) C.int {
	return C.int(resourceRestoreMana52BF50((*server.SpellAcceptArg)(arg), resourceSpellDepsForCall52BF20()))
}

//export sub_52DD50
func sub_52DD50(spellID C.int, second unsafe.Pointer, caster *C.nox_object_t, fourth *C.nox_object_t, arg unsafe.Pointer, level C.int) C.int {
	return C.int(resourceLesserHeal52DD50(int32(spellID), (*server.Object)(second), asObjectS((*nox_object_t)(caster)), (*server.SpellAcceptArg)(arg), resourceSpellDepsForCall52BF20()))
}
