#include "item_apply_disengage_4f3030.h"

// CGo cannot express a pointer-to-const parameter on an exported Go function,
// so retain the exact public declaration here and forward the same native-width
// item and owner pointers to the Go-owned dispatch contract.
extern void nox_xxx_itemApplyDisengageEffect_4F3030_go(
	nox_object_t* item,
	nox_object_t* owner);

void nox_xxx_itemApplyDisengageEffect_4F3030(
		const nox_object_t* item,
		nox_object_t* owner) {
	nox_xxx_itemApplyDisengageEffect_4F3030_go((nox_object_t*)item, owner);
}

