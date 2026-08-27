#include "../../legacy/field_guide_loss_eligible_4f2530.h"

#include <stdint.h>

typedef int32_t (*field_guide_loss_eligible_fn)(int32_t);

_Static_assert(
		_Generic(&sub_4F2530, field_guide_loss_eligible_fn: 1, default: 0),
		"sub_4F2530 must keep its exact fixed-width scalar ABI");

int main(void) {
	field_guide_loss_eligible_fn fn = &sub_4F2530;
	return fn == 0;
}
