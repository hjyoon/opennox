#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../item_defend_effects_4e1320.h"

typedef int (*item_defend_effects_fn)(nox_object_t*, nox_object_t*,
	nox_object_t*, int32_t*, int32_t);

_Static_assert(CHAR_BIT == 8 && sizeof(int) == 4 && sizeof(int32_t) == 4,
	"004E1320 preserves signed DWORD damage/type and return values");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(_Generic(&nox_xxx_itemApplyDefendEffect2_4E1320,
	item_defend_effects_fn: 1, default: 0),
	"004E1320 must receive three native object pointers and a DWORD address");

struct nox_object_t { uintptr_t marker; };
static nox_object_t* observed[3];
static int32_t* observed_damage;
static int32_t observed_type;

// ABI fixture only; this does not execute the production GAME3_2 body.
int nox_xxx_itemApplyDefendEffect2_4E1320(nox_object_t* target,
	nox_object_t* source, nox_object_t* weapon, int32_t* damage, int32_t typ) {
	observed[0] = target;
	observed[1] = source;
	observed[2] = weapon;
	observed_damage = damage;
	observed_type = typ;
	return -17;
}

int main(void) {
	nox_object_t target = {UINTPTR_MAX}, source = {UINTPTR_MAX - 1}, weapon = {UINTPTR_MAX - 2};
	int32_t damage = INT32_MIN;
	item_defend_effects_fn const apply = nox_xxx_itemApplyDefendEffect2_4E1320;
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&target > UINT32_MAX && (uintptr_t)&source > UINT32_MAX &&
			(uintptr_t)&weapon > UINT32_MAX && (uintptr_t)&damage > UINT32_MAX);
	}
	assert(apply(&target, &source, &weapon, &damage, INT32_MIN + 1) == -17);
	assert(observed[0] == &target && observed[1] == &source && observed[2] == &weapon);
	assert(observed[0]->marker == UINTPTR_MAX && observed[1]->marker == UINTPTR_MAX - 1 && observed[2]->marker == UINTPTR_MAX - 2);
	assert(observed_damage == &damage && damage == INT32_MIN && observed_type == INT32_MIN + 1);
	assert(apply(NULL, NULL, NULL, NULL, INT32_MAX) == -17);
	assert(observed[0] == NULL && observed[1] == NULL && observed[2] == NULL && observed_damage == NULL && observed_type == INT32_MAX);
	return 0;
}
