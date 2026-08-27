#include "../../legacy/ability_loss_eligible_4f2570.h"

#include <stdint.h>

typedef int32_t (*ability_loss_eligible_fn)(int32_t);

_Static_assert(
		_Generic(&sub_4F2570, ability_loss_eligible_fn: 1, default: 0),
		"sub_4F2570 must keep its exact fixed-width scalar ABI");

int main(void) {
	ability_loss_eligible_fn fn = &sub_4F2570;
	return fn == 0;
}
