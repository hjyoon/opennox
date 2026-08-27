// Keep the native pointer ABI independent from Win32-only aggregate headers.
#include "../secondary_weapon_53ab90.h"

#include <stdint.h>

struct nox_object_t {
	uintptr_t marker;
};

typedef void (*secondary_weapon_fn)(nox_object_t*, nox_object_t*);

_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&sub_53AB90, secondary_weapon_fn: 1, default: 0),
	"secondary weapon report must preserve both object pointers");

static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static unsigned calls;

void sub_53AB90(nox_object_t* owner, nox_object_t* item) {
	if (owner == expected_owner && item == expected_item)
		++calls;
}

int main(void) {
	nox_object_t owner = {.marker = UINTPTR_MAX};
	nox_object_t item = {.marker = UINTPTR_MAX - 1U};
	secondary_weapon_fn const report = sub_53AB90;
	expected_owner = &owner;
	expected_item = &item;
	report(&owner, &item);
	if (calls != 1U || owner.marker != UINTPTR_MAX ||
		item.marker != UINTPTR_MAX - 1U)
		return __LINE__;
	return 0;
}
