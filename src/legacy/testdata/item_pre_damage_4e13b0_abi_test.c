#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../item_pre_damage_4e13b0.h"

typedef int (*item_pre_damage_fn)(nox_object_t*, nox_object_t*,
	nox_object_t*, int32_t*);

_Static_assert(CHAR_BIT == 8 && sizeof(int) == 4 && sizeof(int32_t) == 4,
	"004E13B0 preserves the signed DWORD damage and return values");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(_Generic(&nox_xxx_itemApplyPreDamageEffect_4E13B0,
	item_pre_damage_fn: 1, default: 0),
	"004E13B0 must receive three native object pointers and a DWORD address");

struct nox_object_t { uintptr_t marker; };
static nox_object_t* observed[3];
static int32_t* observed_damage;

// ABI fixture only; this does not execute the production GAME3_2 body.
int nox_xxx_itemApplyPreDamageEffect_4E13B0(nox_object_t* target,
	nox_object_t* source, nox_object_t* weapon, int32_t* damage) {
	observed[0] = target;
	observed[1] = source;
	observed[2] = weapon;
	observed_damage = damage;
	return -17;
}

int main(void) {
	nox_object_t target = {UINTPTR_MAX}, source = {UINTPTR_MAX - 1}, weapon = {UINTPTR_MAX - 2};
	int32_t damage = INT32_MIN;
	item_pre_damage_fn const apply = nox_xxx_itemApplyPreDamageEffect_4E13B0;
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&target > UINT32_MAX && (uintptr_t)&source > UINT32_MAX &&
			(uintptr_t)&weapon > UINT32_MAX && (uintptr_t)&damage > UINT32_MAX);
	}
	assert(apply(&target, &source, &weapon, &damage) == -17);
	assert(observed[0] == &target && observed[1] == &source && observed[2] == &weapon);
	assert(observed[0]->marker == UINTPTR_MAX && observed[1]->marker == UINTPTR_MAX - 1 && observed[2]->marker == UINTPTR_MAX - 2);
	assert(observed_damage == &damage && damage == INT32_MIN);
	assert(apply(NULL, NULL, NULL, NULL) == -17);
	assert(observed[0] == NULL && observed[1] == NULL && observed[2] == NULL && observed_damage == NULL);
	return 0;
}
