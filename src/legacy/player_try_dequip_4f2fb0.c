#include "player_try_dequip_4f2fb0.h"

// CGo cannot express a pointer-to-const parameter on an exported Go function,
// so keep the original public declaration in this narrow adapter and pass the
// same native-width pointer to the Go-owned decision contract.
extern int32_t nox_xxx_playerTryDequip_4F2FB0_go(
	nox_object_t* owner,
	nox_object_t* item);

int32_t nox_xxx_playerTryDequip_4F2FB0(
		nox_object_t* owner,
		const nox_object_t* item) {
	return nox_xxx_playerTryDequip_4F2FB0_go(owner, (nox_object_t*)item);
}
