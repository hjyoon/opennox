#ifndef OPENNOX_CLIENT_RENDER_PALETTE_H
#define OPENNOX_CLIENT_RENDER_PALETTE_H

#include <stdint.h>

typedef struct nox_render_palette {
	uint32_t black;
	uint32_t white;
	uint32_t violet;
	uint32_t red;
	uint32_t green;
	uint32_t cyan;
	uint32_t blue;
	uint32_t orange;
	uint32_t yellow;
} nox_render_palette;

_Static_assert(sizeof(nox_render_palette) == 36, "legacy render color cell layout");
nox_render_palette* nox_render_palette_storage(void);

#endif
