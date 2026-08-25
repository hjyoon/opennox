package legacy

/*
#include "inventory_detach_4ed0c0.h"
*/
import "C"

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func inventoryDetachRuntime4ED0C0() server.InventoryDetachRuntime4ED0C0 {
	return server.InventoryDetachRuntime4ED0C0{
		GameFlag: func(flag uint32) uint32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		NetReportDequip: func(index uint8, item *server.Object) {
			C.nox_xxx_netReportDequip_4D84C0(C.int(index), asObjectC(item))
		},
		DequipArmor: func(owner, item *server.Object, mode, report int32) {
			C.sub_53E430(
				asObjectC(owner),
				asObjectC(item),
				C.int(mode),
				C.int(report),
			)
		},
		DequipWeapon: func(owner, item *server.Object, mode, report int32) {
			C.nox_xxx_playerDequipWeapon_53A140(
				asObjectC(owner),
				asObjectC(item),
				C.int(mode),
				C.int(report),
			)
		},
		NetReportDrop: func(index uint8, item *server.Object) {
			C.nox_xxx_netReportDrop_4D8B50(C.int(index), asObjectC(item))
		},
		ProtectItem: func(value uint32, item *server.Object) {
			C.nox_xxx_protect_56FC50(C.int32_t(value), asObjectC(item))
		},
		NPCSetItemEquip: func(owner, item *server.Object, equipped int32) {
			owner.SetNPCItemEquipFlags(
				item,
				equipped == 1,
				objectNPCWeaponEquipFlags,
				objectNPCArmorEquipFlags,
			)
		},
	}
}

func inventoryDetach4ED0C0(owner, item *server.Object) {
	s := GetServer().S()
	s.DetachInventory4ED0C0(owner, item, inventoryDetachRuntime4ED0C0())
}

//export sub_4ED0C0
func sub_4ED0C0(owner, item *C.nox_object_t) {
	inventoryDetach4ED0C0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	)
}
