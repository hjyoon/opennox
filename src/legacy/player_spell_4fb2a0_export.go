package legacy

/*
#include "server__magic__plyrspel.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerSpellExportCall4FB2A0(unit *server.Object) {
	C.nox_xxx_playerSpell_4FB2A0_magic_plyrspel(asObjectC(unit))
}
