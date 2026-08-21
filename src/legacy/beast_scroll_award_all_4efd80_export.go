package legacy

/*
#include "beast_scroll_award_all_4efd80.h"
*/
import "C"

//export nox_xxx_spellAwardAll1_4EFD80
func nox_xxx_spellAwardAll1_4EFD80(player *C.nox_playerInfo) {
	Nox_xxx_spellAwardAll1_4EFD80(asPlayerS(player))
}
