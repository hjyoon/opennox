#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../script_callback_init_4f5540.h"

struct nox_script_callback_t {
	uint32_t flags;
	int32_t func;
};

typedef int32_t (*script_callback_init_fn_4F5540)(nox_script_callback_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(nox_script_callback_t) == 8,
	"script callback must remain eight bytes");
_Static_assert(offsetof(nox_script_callback_t, flags) == 0,
	"script flags must remain at byte zero");
_Static_assert(offsetof(nox_script_callback_t, func) == 4,
	"script function must remain at byte four");
_Static_assert(
	_Generic(&sub_4F5540, script_callback_init_fn_4F5540: 1, default: 0),
	"004F5540 must preserve a native callback pointer and exact int32 result");

static nox_script_callback_t* observed_handler;

int32_t sub_4F5540(nox_script_callback_t* handler) {
	observed_handler = handler;
	if (handler == NULL) {
		return INT32_MIN;
	}
	handler->func = -1;
	return INT32_MAX;
}

int main(void) {
	nox_script_callback_t handler = {
		.flags = UINT32_C(0xa1b2c3d4),
		.func = INT32_C(17),
	};
	script_callback_init_fn_4F5540 const initialize = sub_4F5540;

	assert(initialize(&handler) == INT32_MAX);
	assert(observed_handler == &handler);
	assert(handler.flags == UINT32_C(0xa1b2c3d4));
	assert(handler.func == -1);
	assert(initialize(NULL) == INT32_MIN);
	assert(observed_handler == NULL);
	return 0;
}
