#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../cast_shock_52c5a0.h"

typedef int (*shock_cast_fn)(int, void*, nox_object_t*, nox_object_t*, void*, int);
_Static_assert(CHAR_BIT == 8 && sizeof(int) == 4, "Shock scalar arguments must remain signed DWORDs");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(_Generic(&nox_xxx_useShock_52C5A0, shock_cast_fn: 1, default: 0),
	"0052C5A0 must retain four native pointers and two signed DWORDs");

struct nox_object_t { uintptr_t marker; };
static void* observed_second;
static nox_object_t* observed_caster;
static nox_object_t* observed_context;
static void* observed_arg;

/* This standalone fixture checks the public declaration/calling convention.
 * The actual linked C entrypoint and Go runtime are tested in legacy tests. */
int nox_xxx_useShock_52C5A0(int id, void* second, nox_object_t* caster,
	nox_object_t* context, void* arg, int power) {
	assert(id == INT_MIN && power == INT_MAX);
	observed_second = second;
	observed_caster = caster;
	observed_context = context;
	observed_arg = arg;
	return INT_MIN;
}

int main(void) {
	nox_object_t caster = {UINTPTR_MAX}, context = {UINTPTR_MAX - 1};
	uintptr_t second = UINTPTR_MAX - 2, arg = UINTPTR_MAX - 3;
	shock_cast_fn const cast = nox_xxx_useShock_52C5A0;
	assert(cast(INT_MIN, &second, &caster, &context, &arg, INT_MAX) == INT_MIN);
	assert(observed_second == &second && observed_caster == &caster);
	assert(observed_context == &context && observed_arg == &arg);
	assert(cast(INT_MIN, NULL, NULL, NULL, NULL, INT_MAX) == INT_MIN);
	assert(observed_second == NULL && observed_caster == NULL);
	assert(observed_context == NULL && observed_arg == NULL);
	return 0;
}
