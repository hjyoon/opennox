#ifndef NOX_PORT_CLIENT_DRAW_PARSE_PARSE
#define NOX_PORT_CLIENT_DRAW_PARSE_PARSE

#include "defs.h"

typedef struct nox_static_draw_data_t {
	uint32_t size;
	nox_video_bag_image_t* image;
} nox_static_draw_data_t;

typedef struct nox_static_random_draw_data_t {
	uint32_t size;
	nox_video_bag_image_t** images;
	uint8_t count;
} nox_static_random_draw_data_t;

typedef struct nox_animate_draw_data_t {
	uint32_t size;
	nox_video_bag_image_t** images;
	uint8_t frame_count;
	uint8_t frame_delay;
	uint32_t animation_kind;
} nox_animate_draw_data_t;

typedef struct nox_cond_animate_draw_data_t {
	uint32_t size;
	nox_video_bag_image_t** images[5];
	uint8_t frame_count[5];
	uint8_t frame_delay[5];
	uint32_t animation_kind[5];
} nox_cond_animate_draw_data_t;

nox_static_random_draw_data_t* nox_xxx_spriteLoadStaticRandomData_44C000(char* attr_value, nox_memfile* f);
int nox_xxx_loadVectorAnimated_44B8B0(int a1, nox_memfile* f);
int nox_xxx_loadVectorAnimated_44BC50(int a1, nox_memfile* f);
int get_animation_kind_id_44B4C0(const char* a1);

#endif // NOX_PORT_CLIENT_DRAW_PARSE_PARSE
