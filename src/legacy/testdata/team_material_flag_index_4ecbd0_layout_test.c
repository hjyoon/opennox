#include "../GAME3_3.h"

#ifdef NOX_TEAM_MATERIAL_FLAG_INDEX_4ECBD0_NATIVE_LAYOUT
// Reassert the complete native-width field path consumed by 004ECBD0/004ECC00
// and pin the retained C caller to the generated CGo signature.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, init_data) == (sizeof(void*) == 4 ? 692 : 760),
	"object init-data offset");
_Static_assert(offsetof(nox_modifier_attrs_t, modifiers) == 0,
	"modifier array offset");
_Static_assert(sizeof(((nox_modifier_attrs_t*)0)->modifiers[0]) == sizeof(void*),
	"modifier pointer width");
_Static_assert(sizeof(nox_modifier_attrs_t) == (sizeof(void*) == 4 ? 20 : 40),
	"modifier attrs size");
_Static_assert(offsetof(obj_412ae0_t, field_0) == 0,
	"modifier name offset");
_Static_assert(sizeof(obj_412ae0_t) == (sizeof(void*) == 4 ? 144 : 208),
	"modifier definition size");

static int (*const team_material_object_index_signature_4ecbd0)(nox_object_t*) =
	sub_4ECBD0;

static const char* team_material_second_modifier_name_4ecbd0(
	const nox_object_t* object
) {
	const nox_modifier_attrs_t* attrs = object->init_data;
	const obj_412ae0_t* material = attrs->modifiers[1];
	return material->field_0;
}
#endif
