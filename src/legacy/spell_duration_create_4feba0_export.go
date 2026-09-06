package legacy

/*
#include "spell_duration_create_4feba0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellDurationBased_4FEBA0
func nox_xxx_spellDurationBased_4FEBA0(
	spellID C.int32_t,
	second, third, fourth *C.nox_object_t,
	arg *C.nox_spell_accept_arg_t,
	level C.int32_t,
	create, update, destroy unsafe.Pointer,
	duration C.int32_t,
) C.int32_t {
	return C.int32_t(GetServer().SpellDurationCreate4FEBA0(
		int32(spellID),
		asObjectS((*nox_object_t)(second)),
		asObjectS((*nox_object_t)(third)),
		asObjectS((*nox_object_t)(fourth)),
		(*server.SpellAcceptArg)(unsafe.Pointer(arg)),
		int32(level),
		create,
		update,
		destroy,
		int32(duration),
	))
}

func spellDurationCreateExportCall4FEBA0(
	spellID int32,
	second, third, fourth *server.Object,
	arg *server.SpellAcceptArg,
	level int32,
	create, update, destroy unsafe.Pointer,
	duration int32,
) int32 {
	return int32(C.nox_xxx_spellDurationBased_4FEBA0(
		C.int32_t(spellID),
		(*C.nox_object_t)(unsafe.Pointer(second)),
		(*C.nox_object_t)(unsafe.Pointer(third)),
		(*C.nox_object_t)(unsafe.Pointer(fourth)),
		(*C.nox_spell_accept_arg_t)(unsafe.Pointer(arg)),
		C.int32_t(level),
		create,
		update,
		destroy,
		C.int32_t(duration),
	))
}

func spellDurationCreateArgCSize4FEBA0() uintptr {
	return uintptr(C.sizeof_nox_spell_accept_arg_t)
}
