package legacy

/*
#include <stdint.h>
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_spellCastByPlayer_4FEEF0 func()
)

//export nox_xxx_spellCastedCaster_native
func nox_xxx_spellCastedCaster_native(a1 unsafe.Pointer) *nox_object_t {
	if a1 == nil {
		return nil
	}
	return asObjectC((*server.DurSpell)(a1).Caster16)
}

//export nox_xxx_spellCastedSpell_native
func nox_xxx_spellCastedSpell_native(a1 unsafe.Pointer) int32 {
	if a1 == nil {
		return 0
	}
	return int32((*server.DurSpell)(a1).Spell)
}

//export nox_xxx_spellCastByPlayer_4FEEF0
func nox_xxx_spellCastByPlayer_4FEEF0() { Nox_xxx_spellCastByPlayer_4FEEF0() }

//export sub_4FF310
func sub_4FF310(a1 *nox_object_t) {
	GetServer().S().Spells.Dur.CancelOffensiveFor(asObjectS(a1))
}
