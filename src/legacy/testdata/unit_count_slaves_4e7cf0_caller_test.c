#include "../GAME3_3.h"

#include <assert.h>

#include "../unit_count_slaves_4e7cf0.c"

static int32_t (*const count_slaves_signature_4e7cf0)(
	const nox_object_t*, uint32_t, uint32_t
) = nox_xxx_unitCountSlaves_4E7CF0;

int main(void) {
	const nox_object_t* invalid = (const nox_object_t*)(uintptr_t)1;
	assert(count_slaves_signature_4e7cf0(NULL, UINT32_C(1), UINT32_C(1)) == 0);
	assert(count_slaves_signature_4e7cf0(invalid, 0, UINT32_C(1)) == 0);
	assert(count_slaves_signature_4e7cf0(invalid, UINT32_C(1), 0) == 0);

	nox_object_t owner = {0};
	nox_object_t first = {0};
	nox_object_t second = {0};
	nox_object_t third = {0};
	nox_object_t fourth = {0};

	first.obj_class = UINT32_C(1);
	first.obj_subclass = UINT32_C(0x40000000);
	first.obj_flags = UINT32_C(0x20);
	first.field_128 = &second;
	second.obj_class = UINT32_C(6);
	second.obj_subclass = UINT32_C(1);
	second.field_128 = &third;
	third.obj_class = UINT32_C(2);
	third.obj_subclass = UINT32_C(4);
	third.obj_flags = UINT32_C(0x20);
	third.field_128 = &fourth;
	fourth.obj_class = UINT32_C(0x80000000);
	fourth.obj_subclass = UINT32_C(0x40000000);
	owner.field_129 = &first;

	assert(count_slaves_signature_4e7cf0(&owner, UINT32_C(0x80000002), UINT32_C(0x40000004)) == 2);
	assert(count_slaves_signature_4e7cf0(&owner, UINT32_C(2), UINT32_C(1)) == 1);
	assert(count_slaves_signature_4e7cf0(&owner, UINT32_C(1), UINT32_C(0x40000000)) == 1);
	assert(count_slaves_signature_4e7cf0(&owner, UINT32_C(4), UINT32_C(4)) == 0);
	assert(first.field_128 == &second);
	assert(second.field_128 == &third);
	assert(third.field_128 == &fourth);
	assert(fourth.field_128 == NULL);

	return 0;
}
