#include "client__draw__canidraw.h"
#include "GAME1_2.h"
#include "GAME3_1.h"
#include "client__draw__parse__parse.h"
#include "common__random.h"
#include "operators.h"

//----- (004BC930) --------------------------------------------------------
int nox_thing_cond_animate_draw(unsigned int* a1, struct nox_drawable* dr) {
	nox_cond_animate_draw_data_t* data = dr->field_76;
	int state = (dr->flags30 & 0x1000000) ? 0 : 1;
	if (!data || !data->images[state] || data->frame_count[state] == 0) {
		return 0;
	}
	int frame;
	if (data->animation_kind[state] != 2) {
		if (data->animation_kind[state] != 4) {
			if (data->animation_kind[state] != 5) {
				return 0;
			}
			frame = dr->field_77;
		} else {
			frame = nox_common_randomIntMinMax_415FF0(
				0, data->frame_count[state] - 1, "C:\\NoxPost\\src\\client\\Draw\\CAniDraw.c", 57);
		}
	} else {
		frame = (gameFrame() + dr->field_32) / (unsigned int)(data->frame_delay[state] + 1);
		if (frame >= data->frame_count[state]) {
			frame %= data->frame_count[state];
		}
	}
	if (frame < 0 || frame >= data->frame_count[state]) {
		return 0;
	}
	nox_xxx_drawObject_4C4770_draw(a1, dr, data->images[state][frame]);
	return 1;
}

//----- (0044B560) --------------------------------------------------------
bool nox_things_cond_animate_draw_parse(nox_thing* obj, nox_memfile* f, char* attr_value) {
	nox_cond_animate_draw_data_t* data = calloc(1, sizeof(*data));
	if (!data) {
		return 0;
	}
	data->size = sizeof(*data);
	uint8_t state_count = nox_memfile_read_u8(f);
	if (state_count > 5) {
		free(data);
		return 0;
	}
	for (int state = 0; state < state_count; state++) {
		data->frame_count[state] = nox_memfile_read_u8(f);
		data->frame_delay[state] = nox_memfile_read_u8(f);
		uint8_t kind_len = nox_memfile_read_u8(f);
		nox_memfile_read(attr_value, 1, kind_len, f);
		attr_value[kind_len] = 0;
		data->animation_kind[state] = get_animation_kind_id_44B4C0(attr_value);
		if (data->frame_count[state] != 0) {
			data->images[state] = calloc(data->frame_count[state], sizeof(*data->images[state]));
			if (!data->images[state]) {
				for (int prev = 0; prev < state; prev++) {
					free(data->images[prev]);
				}
				free(data);
				return 0;
			}
		}
		for (int frame = 0; frame < data->frame_count[state]; frame++) {
			int image_id = nox_memfile_read_i32(f);
			char image_type = 0;
			attr_value[0] = getMemByte(0x5D4594, 830836);
			if (image_id == -1) {
				image_type = nox_memfile_read_u8(f);
				uint8_t name_len = nox_memfile_read_u8(f);
				nox_memfile_read(attr_value, 1, name_len, f);
				attr_value[name_len] = 0;
			}
			data->images[state][frame] = nox_xxx_readImgMB_42FAA0(image_id, image_type, attr_value);
		}
	}
	obj->field_5c = data;
	obj->draw_func = &nox_thing_cond_animate_draw;
	obj->field_60 = 0;
	return 1;
}
