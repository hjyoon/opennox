package legacy

/*
#include <stdint.h>
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_spellCastByPlayer_4FEEF0 func()
)

//export nox_xxx_spellCastedFirst_4FE930
func nox_xxx_spellCastedFirst_4FE930() unsafe.Pointer { return GetServer().S().Spells.Dur.List.C() }

//export nox_xxx_spellCastedNext_4FE940
func nox_xxx_spellCastedNext_4FE940(a1 unsafe.Pointer) unsafe.Pointer {
	return ((*server.DurSpell)(a1)).Next.C()
}

//export nox_xxx_spellCastedCaster_native
func nox_xxx_spellCastedCaster_native(a1 unsafe.Pointer) *nox_object_t {
	if a1 == nil {
		return nil
	}
	return asObjectC((*server.DurSpell)(a1).Caster16)
}

//export nox_xxx_spellCastedTarget_native
func nox_xxx_spellCastedTarget_native(a1 unsafe.Pointer) *nox_object_t {
	if a1 == nil {
		return nil
	}
	return asObjectC((*server.DurSpell)(a1).Target48)
}

//export nox_xxx_spellCastedSpell_native
func nox_xxx_spellCastedSpell_native(a1 unsafe.Pointer) int32 {
	if a1 == nil {
		return 0
	}
	return int32((*server.DurSpell)(a1).Spell)
}

//export sub_4FE8A0
func sub_4FE8A0(a1_cgo int32) { a1 := int(a1_cgo); GetServer().S().Spells.Dur.Sub4FE8A0(a1) }

//export nox_xxx_spellCastByPlayer_4FEEF0
func nox_xxx_spellCastByPlayer_4FEEF0() { Nox_xxx_spellCastByPlayer_4FEEF0() }

//export nox_xxx_spellCancelDurSpell_4FEB10
func nox_xxx_spellCancelDurSpell_4FEB10(a1_cgo int32, a2 *nox_object_t) {
	a1 := int(a1_cgo)
	GetServer().S().Spells.Dur.CancelFor(spell.ID(a1), asObjectS(a2))
}

//export sub_4FE980
func sub_4FE980(a1 unsafe.Pointer) { GetServer().S().Spells.Dur.FreeRecursive((*server.DurSpell)(a1)) }

//export sub_4FF310
func sub_4FF310(a1 *nox_object_t) {
	GetServer().S().Spells.Dur.CancelOffensiveFor(asObjectS(a1))
}
