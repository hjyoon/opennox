#ifndef NOX_PORT_CLIENT_SHELL_SELCOLOR
#define NOX_PORT_CLIENT_SHELL_SELCOLOR

#include "defs.h"

// Modifier records are Go-native structures. These accessors keep legacy C
// code independent of pointer-width-dependent Go padding.
uint32_t nox_modifier_getColorRGB(void* modifier, int index);
int32_t nox_modifier_getEffectiveness(void* modifier);
int32_t nox_modifier_getMaterial(void* modifier);
int32_t nox_modifier_getPriEnchant(void* modifier);
int32_t nox_modifier_getColorSlot(void* modifier, int index);
uint32_t nox_modifier_effect_getColorRGB(void* modifier);

int nox_game_showSelColor_4A5D00();
int nox_client_selColorSetEquipmentPalette(unsigned char index);
wchar2_t* sub_4A68C0();
int sub_4A75C0();

#endif // NOX_PORT_CLIENT_SHELL_SELCOLOR
