package legacy

/*
#include "respawn_weapon_flags_4ef580.h"
*/
import "C"

//export nox_xxx_getRespawnWeaponFlags_4EF580
func nox_xxx_getRespawnWeaponFlags_4EF580() C.uint8_t {
	return C.uint8_t(GetServer().S().RespawnWeaponFlags4EF580())
}
