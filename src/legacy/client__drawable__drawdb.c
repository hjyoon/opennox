#include "defs.h"

// Preserve the address return for call-scoped Go drawables.
void nox_drawable_set_light_color_go(nox_drawable* dr, int r, int g, int b);

nox_drawable* nox_drawable_change_light_color_native(nox_drawable* dr, int r, int g, int b) {
	nox_drawable_set_light_color_go(dr, r, g, b);
	return dr;
}

void nox_drawable_set_packed_light_color_go(uint32_t* light, int r, int g, int b);

uint32_t* nox_xxx_spriteChangeLightColor_484BE0(uint32_t* light, int r, int g, int b) {
	nox_drawable_set_packed_light_color_go(light, r, g, b);
	return light;
}
