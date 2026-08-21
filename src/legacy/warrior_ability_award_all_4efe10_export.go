package legacy

/*
#include "warrior_ability_award_all_4efe10.h"
*/
import "C"

//export nox_xxx_spellAwardAll3_4EFE10
func nox_xxx_spellAwardAll3_4EFE10(player *C.nox_playerInfo) {
	Nox_xxx_spellAwardAll3_4EFE10(asPlayerS(player))
}
