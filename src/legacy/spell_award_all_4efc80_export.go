package legacy

/*
#include "spell_award_all_4efc80.h"
*/
import "C"

//export nox_xxx_spellAwardAll2_4EFC80
func nox_xxx_spellAwardAll2_4EFC80(player *C.nox_playerInfo) {
	Nox_xxx_spellAwardAll2_4EFC80(asPlayerS(player))
}
