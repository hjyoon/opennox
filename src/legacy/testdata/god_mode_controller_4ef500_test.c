#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../god_mode_controller_4ef500.h"

typedef void (*god_mode_controller_fn)(int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "god-mode input must remain exact int32");
_Static_assert(
	_Generic(&nox_xxx_set_god_4EF500, god_mode_controller_fn: 1, default: 0),
	"god-mode controller must receive one exact int32 value");

static int32_t observed_value;
static unsigned int observed_calls;

void nox_xxx_set_god_4EF500(int32_t value) {
	observed_value = value;
	++observed_calls;
}

static void check_call(god_mode_controller_fn controller, int32_t value) {
	controller(value);
	assert(observed_value == value);
}

int main(void) {
	god_mode_controller_fn const controller = nox_xxx_set_god_4EF500;

	check_call(controller, INT32_MIN);
	check_call(controller, INT32_C(-1));
	check_call(controller, INT32_C(0));
	check_call(controller, INT32_C(1));
	check_call(controller, INT32_C(2));
	check_call(controller, INT32_MAX);
	assert(observed_calls == 6);
	return 0;
}
