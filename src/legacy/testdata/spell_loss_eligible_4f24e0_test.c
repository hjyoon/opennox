#include "../../legacy/spell_loss_eligible_4f24e0.h"

#include <stdint.h>

typedef int32_t (*spell_loss_eligible_fn)(int32_t);

_Static_assert(
		_Generic(&sub_4F24E0, spell_loss_eligible_fn: 1, default: 0),
		"sub_4F24E0 must keep its exact fixed-width scalar ABI");

int main(void) {
	spell_loss_eligible_fn fn = &sub_4F24E0;
	return fn == 0;
}
