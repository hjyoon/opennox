#include "../GAME3_3.h"

#include <assert.h>
#include <stdint.h>
#include <string.h>

#ifdef NOX_COLLIDER_TEST_NATIVE_LAYOUT
// The wider legacy header still contains unrelated Win32-only assertions.
// A native harness may suppress those globally, then re-enable the exact
// shape/object layout consumed by 004E7290 here.
#undef _Static_assert
_Static_assert(offsetof(nox_shape, kind) == 0, "shape kind");
_Static_assert(offsetof(nox_shape, circle_r) == 4, "circle radius");
_Static_assert(offsetof(nox_shape, box_left_bottom) == 24, "box min y");
_Static_assert(offsetof(nox_shape, box_left_bottom_2) == 28, "box min x");
_Static_assert(offsetof(nox_shape, box_right_top) == 36, "box max x");
_Static_assert(offsetof(nox_shape, box_right_top_2) == 48, "box max y");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60), "object x");
_Static_assert(offsetof(nox_object_t, y) == (sizeof(void*) == 4 ? 60 : 64), "object y");
_Static_assert(offsetof(nox_object_t, shape) == (sizeof(void*) == 4 ? 172 : 176), "object shape");
_Static_assert(offsetof(nox_object_t, collide_x1) == (sizeof(void*) == 4 ? 232 : 236), "object min x");
_Static_assert(offsetof(nox_object_t, collide_y1) == (sizeof(void*) == 4 ? 236 : 240), "object min y");
_Static_assert(offsetof(nox_object_t, collide_x2) == (sizeof(void*) == 4 ? 240 : 244), "object max x");
_Static_assert(offsetof(nox_object_t, collide_y2) == (sizeof(void*) == 4 ? 244 : 248), "object max y");
#endif

static uint32_t float_bits(float value) {
	uint32_t bits;
	memcpy(&bits, &value, sizeof(bits));
	return bits;
}

static void set_float_bits(float* out, uint32_t bits) { memcpy(out, &bits, sizeof(bits)); }

int main(void) {
	nox_object_t obj = {0};

	obj.shape.kind = NOX_SHAPE_CENTER;
	set_float_bits(&obj.x, UINT32_C(0x7fa12345));
	set_float_bits(&obj.y, UINT32_C(0x80000000));
	assert(nox_xxx_objectUnkUpdateCoords_4E7290(&obj) == &obj);
	assert(float_bits(obj.collide_x1) == UINT32_C(0x7fa12345));
	assert(float_bits(obj.collide_y1) == UINT32_C(0x80000000));
	assert(float_bits(obj.collide_x2) == UINT32_C(0x7fa12345));
	assert(float_bits(obj.collide_y2) == UINT32_C(0x80000000));

	obj.shape.kind = NOX_SHAPE_CIRCLE;
	obj.x = 12.5f;
	obj.y = -3.25f;
	obj.shape.circle_r = 2.5f;
	nox_xxx_objectUnkUpdateCoords_4E7290(&obj);
	assert(obj.collide_x1 == 10.0f);
	assert(obj.collide_y1 == -5.75f);
	assert(obj.collide_x2 == 15.0f);
	assert(obj.collide_y2 == -0.75f);

	obj.shape.kind = NOX_SHAPE_BOX;
	obj.x = 10.0f;
	obj.y = 20.0f;
	obj.shape.box_left_bottom_2 = -4.0f;
	obj.shape.box_left_bottom = -7.0f;
	obj.shape.box_right_top = 6.0f;
	obj.shape.box_right_top_2 = 9.0f;
	nox_xxx_objectUnkUpdateCoords_4E7290(&obj);
	assert(obj.collide_x1 == 6.0f);
	assert(obj.collide_y1 == 13.0f);
	assert(obj.collide_x2 == 16.0f);
	assert(obj.collide_y2 == 29.0f);

	obj.shape.kind = (nox_shape_kind)99;
	obj.collide_x1 = 1.0f;
	obj.collide_y1 = 2.0f;
	obj.collide_x2 = 3.0f;
	obj.collide_y2 = 4.0f;
	assert(nox_xxx_objectUnkUpdateCoords_4E7290(&obj) == &obj);
	assert(obj.collide_x1 == 1.0f);
	assert(obj.collide_y1 == 2.0f);
	assert(obj.collide_x2 == 3.0f);
	assert(obj.collide_y2 == 4.0f);

	memset(&obj, 0, sizeof(obj));
	obj.shape.kind = NOX_SHAPE_CENTER;
	obj.x = 100.0f;
	obj.y = 200.0f;
	set_float_bits(&obj.new_x, UINT32_C(0x7fa54321));
	set_float_bits(&obj.new_y, UINT32_C(0x80000000));
	assert(sub_4E7350(&obj) == &obj);
	assert(float_bits(obj.collide_x1) == UINT32_C(0x7fa54321));
	assert(float_bits(obj.collide_y1) == UINT32_C(0x80000000));
	assert(float_bits(obj.collide_x2) == UINT32_C(0x7fa54321));
	assert(float_bits(obj.collide_y2) == UINT32_C(0x80000000));

	obj.shape.kind = NOX_SHAPE_CIRCLE;
	obj.x = 100.0f;
	obj.y = 200.0f;
	obj.new_x = 12.5f;
	obj.new_y = -3.25f;
	obj.shape.circle_r = 2.5f;
	sub_4E7350(&obj);
	assert(obj.collide_x1 == 10.0f);
	assert(obj.collide_y1 == -5.75f);
	assert(obj.collide_x2 == 15.0f);
	assert(obj.collide_y2 == -0.75f);

	obj.shape.kind = NOX_SHAPE_BOX;
	obj.new_x = 10.0f;
	obj.new_y = 20.0f;
	obj.shape.box_left_bottom_2 = -4.0f;
	obj.shape.box_left_bottom = -7.0f;
	obj.shape.box_right_top = 6.0f;
	obj.shape.box_right_top_2 = 9.0f;
	sub_4E7350(&obj);
	assert(obj.collide_x1 == 6.0f);
	assert(obj.collide_y1 == 13.0f);
	assert(obj.collide_x2 == 16.0f);
	assert(obj.collide_y2 == 29.0f);

	obj.shape.kind = (nox_shape_kind)99;
	obj.collide_x1 = 1.0f;
	obj.collide_y1 = 2.0f;
	obj.collide_x2 = 3.0f;
	obj.collide_y2 = 4.0f;
	assert(sub_4E7350(&obj) == &obj);
	assert(obj.collide_x1 == 1.0f);
	assert(obj.collide_y1 == 2.0f);
	assert(obj.collide_x2 == 3.0f);
	assert(obj.collide_y2 == 4.0f);
	return 0;
}
