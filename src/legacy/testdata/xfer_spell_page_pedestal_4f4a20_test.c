#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../award_spell_collide_4ead20.h"
#include "../xfer_spell_page_pedestal_4f4a20.h"

struct nox_object_t {
	uintptr_t marker;
	nox_award_spell_collide_data_t spell_data;
};

typedef int32_t (*spell_page_pedestal_xfer_fn_4F4A20)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4, "transfer result must remain int32");
_Static_assert(sizeof(nox_award_spell_collide_data_t) == 4,
	"spell collide data must remain four bytes");
_Static_assert(
	_Generic(&nox_xxx_XFerSpellPagePedistal_4F4A20,
		spell_page_pedestal_xfer_fn_4F4A20: 1, default: 0),
	"SpellPagePedestalXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerSpellPagePedistal_4F4A20(nox_object_t* object, void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {
		.marker = UINTPTR_MAX,
		.spell_data = {.spell = UINT32_C(0xf1234567)},
	};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 15u);
	spell_page_pedestal_xfer_fn_4F4A20 const transfer =
		nox_xxx_XFerSpellPagePedistal_4F4A20;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_object->spell_data.spell == UINT32_C(0xf1234567));
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
