#ifndef NOX_PORT_CLIENT_SHELL_SELCOLOR
#define NOX_PORT_CLIENT_SHELL_SELCOLOR

#include "defs.h"

static inline uint32_t nox_selcolor_value(const nox_window* win) {
	return win ? (uint32_t)(uintptr_t)win->widget_data : 0;
}

static inline void nox_selcolor_set_value(nox_window* win, uint32_t value) {
	if (win) {
		win->widget_data = (void*)(uintptr_t)value;
	}
}

static inline unsigned int nox_selcolor_palette_index(const nox_window* win) {
	uint32_t value = nox_selcolor_value(win);
	return (uint16_t)(value >> 16) + 32 * (uint16_t)value;
}

// Modifier records are Go-native structures. These accessors keep legacy C
// code independent of pointer-width-dependent Go padding.
uint32_t nox_modifier_getColorRGB(void* modifier, int index);
int32_t nox_modifier_getEffectiveness(void* modifier);
int32_t nox_modifier_getMaterial(void* modifier);
int32_t nox_modifier_getPriEnchant(void* modifier);

int nox_game_showSelColor_4A5D00();
wchar2_t* sub_4A68C0();
int sub_4A75C0();

#endif // NOX_PORT_CLIENT_SHELL_SELCOLOR
