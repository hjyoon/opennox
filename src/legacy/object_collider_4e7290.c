#include "GAME3_3.h"

#include <string.h>

nox_object_t* nox_xxx_objectUnkUpdateCoords_4E7290(nox_object_t* obj) {
	switch (obj->shape.kind) {
	case NOX_SHAPE_CENTER: {
		// GAME.EXE uses integer MOV instructions for this case. memcpy keeps
		// every float bit intact without violating C's aliasing rules.
		uint32_t x;
		uint32_t y;
		memcpy(&x, &obj->x, sizeof(x));
		memcpy(&obj->collide_x1, &x, sizeof(x));
		memcpy(&y, &obj->y, sizeof(y));
		memcpy(&obj->collide_x2, &x, sizeof(x));
		memcpy(&obj->collide_y1, &y, sizeof(y));
		memcpy(&obj->collide_y2, &y, sizeof(y));
		break;
	}
	case NOX_SHAPE_CIRCLE:
		obj->collide_x1 = obj->x - obj->shape.circle_r;
		obj->collide_y1 = obj->y - obj->shape.circle_r;
		obj->collide_x2 = obj->shape.circle_r + obj->x;
		obj->collide_y2 = obj->shape.circle_r + obj->y;
		break;
	case NOX_SHAPE_BOX:
		obj->collide_x1 = obj->shape.box_left_bottom_2 + obj->x;
		obj->collide_y1 = obj->shape.box_left_bottom + obj->y;
		obj->collide_x2 = obj->shape.box_right_top + obj->x;
		obj->collide_y2 = obj->shape.box_right_top_2 + obj->y;
		break;
	default:
		break;
	}
	return obj;
}

nox_object_t* sub_4E7350(nox_object_t* obj) {
	switch (obj->shape.kind) {
	case NOX_SHAPE_CENTER: {
		uint32_t x;
		uint32_t y;
		memcpy(&x, &obj->new_x, sizeof(x));
		memcpy(&obj->collide_x1, &x, sizeof(x));
		memcpy(&y, &obj->new_y, sizeof(y));
		memcpy(&obj->collide_x2, &x, sizeof(x));
		memcpy(&obj->collide_y1, &y, sizeof(y));
		memcpy(&obj->collide_y2, &y, sizeof(y));
		break;
	}
	case NOX_SHAPE_CIRCLE:
		obj->collide_x1 = obj->new_x - obj->shape.circle_r;
		obj->collide_y1 = obj->new_y - obj->shape.circle_r;
		obj->collide_x2 = obj->shape.circle_r + obj->new_x;
		obj->collide_y2 = obj->shape.circle_r + obj->new_y;
		break;
	case NOX_SHAPE_BOX:
		obj->collide_x1 = obj->shape.box_left_bottom_2 + obj->new_x;
		obj->collide_y1 = obj->shape.box_left_bottom + obj->new_y;
		obj->collide_x2 = obj->shape.box_right_top + obj->new_x;
		obj->collide_y2 = obj->shape.box_right_top_2 + obj->new_y;
		break;
	default:
		break;
	}
	return obj;
}
