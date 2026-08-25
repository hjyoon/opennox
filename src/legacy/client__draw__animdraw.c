#include "client__draw__animdraw.h"
#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME3_1.h"
#include "client__draw__parse__parse.h"
#include "common__random.h"
#include "operators.h"


//----- (004BBD60) --------------------------------------------------------
int nox_thing_animate_draw(unsigned int* a1, struct nox_drawable* dr) {
	int v3;     // esi
	int v6;     // ebx

	nox_animate_draw_data_t* data = dr->field_76;
	if (!data || !data->images || data->frame_count == 0) {
		return 1;
	}
	switch (data->animation_kind) {
	case 0:
		v3 = (gameFrame() - dr->field_79) / ((unsigned int)data->frame_delay + 1);
		if (v3 >= data->frame_count) {
			v3 = data->frame_count - 1;
		}
		break;
	case 1:
		v3 = (gameFrame() - dr->field_79) / ((unsigned int)data->frame_delay + 1);
		if (v3 >= data->frame_count) {
			nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr);
			return 0;
		}
		break;
	case 2:
		if (dr->flags30 & 0x1000000) {
			v3 = (gameFrame() + dr->field_32) / ((unsigned int)data->frame_delay + 1);
			if (v3 >= data->frame_count) {
				v3 %= data->frame_count;
			}
			break;
		}
		if (dr->flags28 & 0x10000000) {
			if (nox_common_gameFlags_check_40A5C0(32)) {
				v3 = (gameFrame() + dr->field_32) / ((unsigned int)data->frame_delay + 1);
				if (v3 >= data->frame_count) {
					v3 %= data->frame_count;
				}
				break;
			}
			if (dr->flags28 & 0x10000000) {
				return 1;
			}
		}
		v3 = 0;
		break;
	case 3:
		v6 = 2 * data->frame_count;
		nox_client_drawEnableAlpha_434560(1);
		v3 = (gameFrame() - dr->field_79) / ((unsigned int)data->frame_delay + 1);
		if (v3 >= v6) {
			nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr);
			return 0;
		}
		nox_client_drawSetAlpha_434580(-56 - 200 * v3 / v6);
		if (v3 >= data->frame_count) {
			v3 %= data->frame_count;
		}
		break;
	case 4:
		v3 = nox_common_randomIntMinMax_415FF0(0, data->frame_count - 1,
											"C:\\NoxPost\\src\\Client\\Draw\\animdraw.c", 24);
		break;
	case 5:
		v3 = dr->field_77;
		break;
	default:
		return 1;
	}
	nox_xxx_drawObject_4C4770_draw(a1, dr, data->images[v3]);
	if (data->animation_kind == 3) {
		nox_client_drawEnableAlpha_434560(0);
	}
	return 1;
}

//----- (0044B390) --------------------------------------------------------
bool nox_things_animate_draw_parse(nox_thing* obj, nox_memfile* f, char* attr_value) {
	nox_animate_draw_data_t* data = calloc(1, sizeof(*data));
	if (!data) {
		return 0;
	}
	data->size = sizeof(*data);
	data->frame_count = nox_memfile_read_u8(f);
	data->frame_delay = nox_memfile_read_u8(f);
	uint8_t kind_len = nox_memfile_read_u8(f);
	nox_memfile_read(attr_value, 1, kind_len, f);
	attr_value[kind_len] = 0;
	data->animation_kind = get_animation_kind_id_44B4C0(attr_value);
	if (data->frame_count != 0) {
		data->images = calloc(data->frame_count, sizeof(*data->images));
		if (!data->images) {
			free(data);
			return 0;
		}
	}
	for (int i = 0; i < data->frame_count; i++) {
		int image_id = nox_memfile_read_i32(f);
		char image_type = 0;
		attr_value[0] = getMemByte(0x5D4594, 830832);
		if (image_id == -1) {
			image_type = nox_memfile_read_u8(f);
			uint8_t name_len = nox_memfile_read_u8(f);
			nox_memfile_read(attr_value, 1, name_len, f);
			attr_value[name_len] = 0;
		}
		data->images[i] = nox_xxx_readImgMB_42FAA0(image_id, image_type, attr_value);
	}
	obj->field_5c = data;
	obj->draw_func = &nox_thing_animate_draw;
	return 1;
}

//----- (0044BE90) --------------------------------------------------------
int sub_44BE90(int a1, nox_memfile* f) {
	int v2;          // esi
	int result;      // eax
	int v4;          // ebx
	int v6;          // ebp
	char v8;         // dl
	int v10;         // esi
	const char* v11; // [esp+4h] [ebp-88h]
	char v12[128];   // [esp+Ch] [ebp-80h]

	v2 = a1;
	result = calloc(*(short*)(a1 + 40), 4);
	*(uint32_t*)(a1 + 4) = result;
	if (result) {
		v4 = 0;
		if (*(uint16_t*)(a1 + 40) > 0) {
			do {
				v6 = nox_memfile_read_i32(f);
				v12[0] = getMemByte(0x5D4594, 830848);
				if (v6 == -1) {
					v8 = nox_memfile_read_u8(f);
					LOBYTE(v11) = v8;
					v10 = nox_memfile_read_u8(f);
					nox_memfile_read(v12, 1, v10, f);
					v12[v10] = 0;
					v2 = a1;
				}
				*(uint32_t*)(*(uint32_t*)(v2 + 4) + 4 * v4++) = nox_xxx_readImgMB_42FAA0(v6, v11, v12);
			} while (v4 < *(short*)(v2 + 40));
		}
		result = 1;
	}
	return result;
}

//----- (0044BD90) --------------------------------------------------------
bool nox_things_animate_state_draw_parse(nox_thing* obj, nox_memfile* f, char* attr_value) {
	const size_t data_sz = 0x94u;
	uint32_t* draw_cb_data = calloc(1u, data_sz);
	draw_cb_data[0] = data_sz;
	while (1) {
		int cmd = nox_memfile_read_u32(f);
		// "END "
		if (cmd == 0x454E4420) {
			break;
		}

		int params = nox_memfile_read_u32(f);
		if (!(params & 0xE)) {
			return 0;
		}

		// TODO: After cleanup: understand significance of these two strings.
		// The first one is always [len=4, data='NULL'], the second is always [len=1, data='']

		unsigned char n;
		n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n);
		n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n);

		unsigned char offset_idx = 0;
		if (params & 2) {
			offset_idx = 0;
		} else if (params & 4) {
			offset_idx = 1;
		} else if (params & 8) {
			offset_idx = 2;
		}

		int v9 = (int)&draw_cb_data[12 * offset_idx + 1];

		if (!nox_xxx_loadVectorAnimated_44B8B0(v9, f)) {
			return 0;
		}

		if (!sub_44BE90(v9, f)) {
			return 0;
		}
	}

	obj->field_54 = 2;
	obj->draw_func = &nox_thing_animate_state_draw;
	obj->field_5c = draw_cb_data;
	return 1;
}
