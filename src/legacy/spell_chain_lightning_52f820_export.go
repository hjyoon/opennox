package legacy

/*
#include "spell_chain_lightning_52f820.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_onStartLightning_52F820
func nox_xxx_onStartLightning_52F820(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellChainLightningCreate52F820((*server.DurSpell)(record)))
}

//export nox_xxx_onFrameLightning_52F8A0
func nox_xxx_onFrameLightning_52F8A0(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellChainLightningUpdate52F8A0((*server.DurSpell)(record)))
}

//export sub_530100
func sub_530100(record unsafe.Pointer) C.char {
	return C.char(GetServer().SpellChainLightningDestroy530100((*server.DurSpell)(record)))
}
