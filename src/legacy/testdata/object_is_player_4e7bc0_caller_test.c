#include <assert.h>

#include "../object_is_player_4e7bc0.c"

static int (*const object_is_player_signature_4e7bc0)(const nox_object_t*) = sub_4E7BC0;

int main(void) {
	static const uint32_t classes[] = {
		UINT32_C(0),
		UINT32_C(3),
		UINT32_C(4),
		UINT32_C(15),
		UINT32_C(0xfffffffb),
		UINT32_C(0xffffffff),
		UINT32_C(0x80000004),
	};
	nox_object_t obj = {0};

	assert(object_is_player_signature_4e7bc0(NULL) == 0);
	obj.typ_ind = UINT16_C(0x2468);
	obj.obj_subclass = UINT32_C(0xa5a5a5a5);
	obj.obj_flags = UINT32_C(0x5a5a5a5a);
	for (size_t i = 0; i < sizeof(classes) / sizeof(classes[0]); ++i) {
		obj.obj_class = classes[i];
		assert(object_is_player_signature_4e7bc0(&obj) == (int)((classes[i] >> 2) & UINT32_C(1)));
		assert(obj.typ_ind == UINT16_C(0x2468));
		assert(obj.obj_class == classes[i]);
		assert(obj.obj_subclass == UINT32_C(0xa5a5a5a5));
		assert(obj.obj_flags == UINT32_C(0x5a5a5a5a));
	}
	return 0;
}
