#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../respawn_weapon_flags_4ef580.h"

typedef uint8_t (*respawn_weapon_flags_fn)(void);

_Static_assert(CHAR_BIT == 8, "packet bytes must remain eight bits");
_Static_assert(sizeof(uint8_t) == 1, "respawn flags must remain one byte");
_Static_assert(
	_Generic(&nox_xxx_getRespawnWeaponFlags_4EF580, respawn_weapon_flags_fn: 1, default: 0),
	"respawn flag detector must take no arguments and return exact uint8_t");

static volatile uint8_t next_result;
static volatile unsigned int observed_calls;

uint8_t nox_xxx_getRespawnWeaponFlags_4EF580(void) {
	++observed_calls;
	return next_result;
}

static void check_result(respawn_weapon_flags_fn detector, uint8_t value) {
	next_result = value;
	assert(detector() == value);
}

int main(void) {
	respawn_weapon_flags_fn const detector = nox_xxx_getRespawnWeaponFlags_4EF580;

	check_result(detector, UINT8_C(0x00));
	check_result(detector, UINT8_C(0x01));
	check_result(detector, UINT8_C(0x7f));
	check_result(detector, UINT8_C(0x80));
	check_result(detector, UINT8_C(0xff));
	assert(observed_calls == 5);
	return 0;
}
