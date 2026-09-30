#include <stdint.h>
#include "../GAME2.h"

typedef void (*spell_mask_fn)(uint32_t*);
_Static_assert(_Generic(&sub_454040, spell_mask_fn: 1, default: 0),
	"00454040 must retain the native mask pointer without an integer result");
_Static_assert(sizeof(uint32_t[5]) == 20,
	"spell settings masks must remain exactly five packed dwords");

int main(void) { return 0; }
