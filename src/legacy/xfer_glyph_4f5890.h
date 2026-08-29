#ifndef NOX_XFER_GLYPH_4F5890_H
#define NOX_XFER_GLYPH_4F5890_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"GlyphXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerGlyph_4F5890(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_GLYPH_4F5890_H
