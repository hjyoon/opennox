#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_item_cancel_4feb60.h"

typedef void (*spell_duration_item_cancel_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FEB60 object arguments must remain pointer-sized");
_Static_assert(
	_Generic(&sub_4FEB60, spell_duration_item_cancel_fn: 1, default: 0),
	"004FEB60 must preserve its two native-pointer void ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_owner;
static nox_object_t* observed_item;
static unsigned int observed_calls;

void sub_4FEB60(nox_object_t* owner, nox_object_t* item) {
	observed_owner = owner;
	observed_item = item;
	++observed_calls;
}

static void check_call(spell_duration_item_cancel_fn cancel,
	nox_object_t* owner, nox_object_t* item) {
	cancel(owner, item);
	assert(observed_owner == owner);
	assert(observed_item == item);
}

int main(void) {
	nox_object_t first = {.marker = UINTPTR_MAX};
	nox_object_t second = {.marker = (uintptr_t)UINT32_C(0x12345678)};
	spell_duration_item_cancel_fn const cancel = sub_4FEB60;

	check_call(cancel, NULL, NULL);
	check_call(cancel, &first, NULL);
	check_call(cancel, NULL, &second);
	check_call(cancel, &first, &second);
	check_call(cancel, &second, &first);
	assert(observed_calls == 5);
	assert(first.marker == UINTPTR_MAX);
	assert(second.marker == (uintptr_t)UINT32_C(0x12345678));
	return 0;
}
