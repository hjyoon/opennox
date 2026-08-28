package legacy

/*
#include <stdint.h>

#include "GAME3_2.h"
#include "GAME4_3.h"
#include "pickup_ammo_4f3b00.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupAmmoCall4F3B00(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupAmmo_4F3B00(owner, item, arg3, arg4)
}

func pickupAmmoExportCall4F3B00(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_pickupAmmo_4F3B00(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

// Nox_xxx_pickupWeapon_53A720 keeps the existing WeaponPickup implementation
// behind an exact native-pointer/fixed-width Go callback.
func Nox_xxx_pickupWeapon_53A720(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.sub_53A720(
		asObjectC(owner),
		asObjectC(item),
		C.int(arg3),
		C.int(arg4),
	))
}

// Nox_xxx_netReportCharges_4D82B0 forwards the final two charge bytes for the
// native inventory object. The original network helper's return is ignored by
// AmmoPickup.
func Nox_xxx_netReportCharges_4D82B0(
	playerInd uint8,
	item *server.Object,
	charge1, charge0 uint8,
) {
	C.nox_xxx_netReportCharges_4D82B0(
		C.int(playerInd),
		asObjectC(item),
		C.char(charge1),
		C.char(charge0),
	)
}

// Nox_xxx_pickupWeaponAudio_53A6C0 retains the material/class-specific pickup
// audio that AmmoPickup emits after scheduling deletion of the merged item.
func Nox_xxx_pickupWeaponAudio_53A6C0(owner, item *server.Object) {
	C.sub_53A6C0(asObjectC(owner), asObjectC(item))
}

//export nox_xxx_pickupAmmo_4F3B00
func nox_xxx_pickupAmmo_4F3B00(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(pickupAmmoCall4F3B00(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
