#ifndef NOX_XFER_SPELL_PAGE_PEDESTAL_4F4A20_H
#define NOX_XFER_SPELL_PAGE_PEDESTAL_4F4A20_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"SpellPagePedestalXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerSpellPagePedistal_4F4A20(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_SPELL_PAGE_PEDESTAL_4F4A20_H
