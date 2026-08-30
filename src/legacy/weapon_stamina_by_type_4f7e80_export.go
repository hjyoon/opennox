package legacy

/*
#include "weapon_stamina_by_type_4f7e80.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func weaponStaminaByTypeExportCall4F7E80(flags uint32) int32 {
	return int32(C.nox_xxx_weaponGetStaminaByType_4F7E80(C.uint32_t(flags)))
}

//export nox_xxx_weaponGetStaminaByType_4F7E80
func nox_xxx_weaponGetStaminaByType_4F7E80(flags C.uint32_t) C.int32_t {
	return C.int32_t(server.WeaponStaminaByType4F7E80(uint32(flags)))
}
