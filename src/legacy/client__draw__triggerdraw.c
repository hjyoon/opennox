#include "client__draw__triggerdraw.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2_3.h"
#include <math.h>

extern uint32_t nox_color_black_2650656;

static int32_t nox_trigger_float2int(float value) {
	// Original x87 nearest-even/integer-indefinite, without changing the
	// native floating-point mode or the legacy truncating global helper.
	if (!isfinite(value) || value >= 2147483648.0f || value < -2147483648.0f) {
		return INT32_MIN;
	}
	float magnitude = fabsf(value);
	float whole = floorf(magnitude);
	float fraction = magnitude - whole;
	if (fraction > 0.5f || (fraction == 0.5f && ((uint32_t)whole & 1))) {
		whole += 1.0f;
	}
	return (int32_t)(value < 0 ? -(double)whole : (double)whole);
}

static int32_t nox_trigger_coord(intptr_t pos, int32_t view_off, float shape) {
	// Numeric ADD endpoints wrap as dwords; native pointers never narrow.
	return (int32_t)((uint32_t)pos + (uint32_t)view_off + (uint32_t)nox_trigger_float2int(shape));
}

//----- (004BB9D0) --------------------------------------------------------
int nox_thing_trigger_draw(nox_draw_viewport_t* a1, nox_drawable* dr) {
	int32_t v2 = (int32_t)((uint32_t)a1->x1 - (uint32_t)a1->field_4);
	int32_t v3 = (int32_t)((uint32_t)a1->y1 - (uint32_t)a1->field_5);

	int v4 = nox_trigger_coord(dr->pos.x, v2, dr->shape.box_left_top);
	int v12 = nox_trigger_coord(dr->pos.y, v3, dr->shape.box_left_bottom);
	int v7 = nox_trigger_coord(dr->pos.x, v2, dr->shape.box_right_bottom_2);
	int v8 = nox_trigger_coord(dr->pos.y, v3, dr->shape.box_right_top_2);
	int v9 = nox_trigger_coord(dr->pos.x, v2, dr->shape.box_left_bottom_2);
	int v10 = nox_trigger_coord(dr->pos.y, v3, dr->shape.box_left_top_2);
	int v11 = nox_trigger_coord(dr->pos.x, v2, dr->shape.box_right_top);
	int v5 = nox_trigger_coord(dr->pos.y, v3, dr->shape.box_right_bottom);

	nox_client_drawSetColor_434460(nox_color_black_2650656);
	nox_client_drawEnableAlpha_434560(1);
	nox_client_drawAddPoint_49F500(v4, v12);
	nox_client_drawAddPoint_49F500(v11, v5);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v7, v8);
	nox_client_drawAddPoint_49F500(v11, v5);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v4, v12);
	nox_client_drawAddPoint_49F500(v9, v10);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawAddPoint_49F500(v7, v8);
	nox_client_drawAddPoint_49F500(v9, v10);
	nox_client_drawLineFromPoints_49E4B0();
	nox_client_drawEnableAlpha_434560(0);

	return 1;
}
