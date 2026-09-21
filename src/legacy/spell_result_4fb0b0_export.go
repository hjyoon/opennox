package legacy

/*
#include "server__magic__plyrspel.h"
*/
import "C"

func spellResultExportCall4FB0B0(status uint32) {
	C.nox_xxx_abilGetError_4FB0B0_magic_plyrspel(C.uint32_t(status))
}

//export nox_xxx_abilGetError_4FB0B0_magic_plyrspel
func nox_xxx_abilGetError_4FB0B0_magic_plyrspel(status C.uint32_t) {
	GetClient().SpellResult4FB0B0(uint32(status))
}
