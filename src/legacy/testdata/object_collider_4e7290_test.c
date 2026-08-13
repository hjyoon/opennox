#include "../GAME3_3.h"

#include <assert.h>
#include <stdint.h>
#include <string.h>

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
	return 0;
}

