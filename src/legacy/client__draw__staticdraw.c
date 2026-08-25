#include "client__draw__staticdraw.h"
#include "GAME1_2.h"
#include "GAME3_1.h"
#include "client__draw__parse__parse.h"
#include "operators.h"

//----- (004BCC20) --------------------------------------------------------
int nox_thing_static_draw(uint32_t* a1, nox_drawable* dr) {
	nox_static_draw_data_t* data = dr->field_76;
	if (data && (!(dr->flags28 & 0x40000) || dr->flags30 & 0x1000000)) {
		nox_xxx_drawObject_4C4770_draw(a1, dr, data->image);
	}
	return 1;
}

//----- (004BCC60) --------------------------------------------------------
int nox_thing_static_random_draw(uint32_t* a1, nox_drawable* dr) {
	nox_static_random_draw_data_t* data = dr->field_76;
	if (data && data->images && dr->field_77 < data->count) {
		nox_xxx_drawObject_4C4770_draw(a1, dr, data->images[dr->field_77]);
	}
	return 1;
}

//----- (0044C160) --------------------------------------------------------
bool nox_things_static_draw_parse(nox_thing* obj, nox_memfile* f, char* attr_value) {
	nox_static_draw_data_t* data = calloc(1, sizeof(*data));
	if (!data) {
		return 0;
	}

	data->size = sizeof(*data);
	int image_id = nox_memfile_read_i32(f);
	char image_type = 0;
	attr_value[0] = getMemByte(0x5D4594, 830856);
	if (image_id == -1) {
		image_type = nox_memfile_read_u8(f);
		uint8_t name_len = nox_memfile_read_u8(f);
		nox_memfile_read(attr_value, 1, name_len, f);
		attr_value[name_len] = 0;
	}
	data->image = nox_xxx_readImgMB_42FAA0(image_id, image_type, attr_value);
	obj->draw_func = &nox_thing_static_draw;
	obj->field_5c = data;
	return 1;
}

//----- (0044BFD0) --------------------------------------------------------
bool nox_things_static_random_draw_parse(nox_thing* obj, nox_memfile* f, char* attr_value) {
	obj->draw_func = nox_thing_static_random_draw;
	obj->field_5c = nox_xxx_spriteLoadStaticRandomData_44C000(attr_value, f);

	return obj->field_5c != 0;
}
