package legacy

/*
#include "GAME3_2.h"

typedef int (*item_damage_entry)(nox_object_t*, nox_object_t*, nox_object_t*, int, int);
_Static_assert(_Generic(&sub_4E14A0, item_damage_entry: 1, default: 0), "ball damage ABI");
_Static_assert(_Generic(&sub_4E14B0, item_damage_entry: 1, default: 0), "weapon damage ABI");
_Static_assert(_Generic(&nox_xxx_damageArmor_4E1500, item_damage_entry: 1, default: 0), "armor damage ABI");

static int call_item_damage_entry(int kind, nox_object_t* target, nox_object_t* source, nox_object_t* weapon) {
	item_damage_entry volatile entries[] = {sub_4E14A0, sub_4E14B0, nox_xxx_damageArmor_4E1500};
	return entries[kind](target, source, weapon, 17, 2);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func itemDamagePublicEntries() [3]unsafe.Pointer {
	return [3]unsafe.Pointer{C.sub_4E14A0, C.sub_4E14B0, C.nox_xxx_damageArmor_4E1500}
}

func callItemDamagePublicEntry(kind int, target, source, weapon *server.Object) int {
	return int(C.call_item_damage_entry(C.int(kind), asObjectC(target), asObjectC(source), asObjectC(weapon)))
}
