#include "GAME3_3.h"

enum {
	NOX_CROWN_TYPE_CACHE_4E7BE0 = 1567716,
	NOX_GAMEBALL_TYPE_CACHE_4E7C30 = 1567720,
};

static int nox_xxx_unitHasOwnedType_4E7BE0(const nox_object_t* owner, uint32_t* type_cache, char* type_name) {
	uint32_t type_ind = *type_cache;
	if (type_ind == 0) {
		type_ind = (uint32_t)nox_xxx_getNameId_4E3AA0(type_name);
		*type_cache = type_ind;
	}

	for (const nox_object_t* obj = owner->field_129; obj; obj = obj->field_128) {
		if ((uint32_t)obj->typ_ind == type_ind) {
			return 1;
		}
	}
	return 0;
}

//----- (004E7BE0) --------------------------------------------------------
int nox_xxx_unitIsCrown_4E7BE0(const nox_object_t* owner) {
	uint32_t* const type_cache = getMemU32Ptr(UINT32_C(0x5D4594), NOX_CROWN_TYPE_CACHE_4E7BE0);
	return nox_xxx_unitHasOwnedType_4E7BE0(owner, type_cache, "Crown");
}

//----- (004E7C30) --------------------------------------------------------
int nox_xxx_unitIsGameball_4E7C30(const nox_object_t* owner) {
	uint32_t* const type_cache = getMemU32Ptr(UINT32_C(0x5D4594), NOX_GAMEBALL_TYPE_CACHE_4E7C30);
	return nox_xxx_unitHasOwnedType_4E7BE0(owner, type_cache, "GameBall");
}
