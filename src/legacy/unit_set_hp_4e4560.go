package legacy

/*
#include "GAME3_3.h"
#include "GAME5_2.h"
typedef intptr_t (*unit_set_hp_entry)(nox_object_t* obj, unsigned short amount);
_Static_assert(_Generic(&nox_xxx_unitSetHP_4E4560, unit_set_hp_entry: 1, default: 0), "unit set HP ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitSetHPNative4E4560(obj *server.Object, amount uint16) uintptr {
	return obj.SetHP4E4560(amount, server.UnitSetHPRuntime4E4560{
		CheckHeld: func(token uint32, item *server.Object) {
			updatePlayerProtectionItem(int32(token), item)
		},
		SetPlayerHP: func(token uint32, value uint16) {
			C.nox_xxx_protectPlayerHPMana_56F870(C.int(token), C.ushort(value))
		},
		ProtectHeld: func(token uint32, item *server.Object) int32 {
			return updatePlayerProtectionItem(int32(token), item)
		},
	})
}

//export nox_xxx_unitSetHP_4E4560
func nox_xxx_unitSetHP_4E4560(obj *nox_object_t, amount C.ushort) C.intptr_t {
	return C.intptr_t(unitSetHPNative4E4560(asObjectS(obj), uint16(amount)))
}

func unitSetHPCEntry4E4560(obj *server.Object, amount uint16) uintptr {
	return uintptr(C.nox_xxx_unitSetHP_4E4560(asObjectC(obj), C.ushort(amount)))
}
