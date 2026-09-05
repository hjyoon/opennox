package legacy

/*
#include "spell_book_insert_4fe340.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func spellBookInsertLegacy4FE340(
	unit *server.Object,
	sequence *int32,
	count, delay, targetMode int32,
) int32 {
	return GetServer().SpellBookInsert4FE340(unit, sequence, count, delay, targetMode)
}

func spellBookInsertExportCall4FE340(
	unit *server.Object,
	sequence *int32,
	count, delay, targetMode int32,
) int32 {
	return int32(C.nox_xxx_spellByBookInsert_4FE340(
		asObjectC(unit),
		(*C.int32_t)(unsafe.Pointer(sequence)),
		C.int32_t(count),
		C.int32_t(delay),
		C.int32_t(targetMode),
	))
}

//export nox_xxx_spellByBookInsert_4FE340
func nox_xxx_spellByBookInsert_4FE340(
	unit *C.nox_object_t,
	sequence *C.int32_t,
	count, delay, targetMode C.int32_t,
) C.int32_t {
	return C.int32_t(spellBookInsertLegacy4FE340(
		asObjectS((*nox_object_t)(unit)),
		(*int32)(unsafe.Pointer(sequence)),
		int32(count),
		int32(delay),
		int32(targetMode),
	))
}
