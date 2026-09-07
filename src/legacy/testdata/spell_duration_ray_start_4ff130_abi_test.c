#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_ray_start_4ff130.h"

typedef void (*duration_ray_start_fn)(void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF130 record must remain pointer-sized");
_Static_assert(
	_Generic(&nox_xxx_netStartDurationRaySpell_4FF130,
		duration_ray_start_fn: 1,
		default: 0),
	"004FF130 must receive one native pointer and return void");

static void* observed_record;
static unsigned int observed_calls;

void nox_xxx_netStartDurationRaySpell_4FF130(void* record) {
	observed_record = record;
	++observed_calls;
}

static void check_call(duration_ray_start_fn start, void* record) {
	start(record);
	assert(observed_record == record);
}

int main(void) {
	static unsigned char first_record;
	static unsigned char second_record;
	duration_ray_start_fn const start = nox_xxx_netStartDurationRaySpell_4FF130;

	check_call(start, NULL);
	check_call(start, &first_record);
	check_call(start, &second_record);
	assert(observed_calls == 3);
	return 0;
}
