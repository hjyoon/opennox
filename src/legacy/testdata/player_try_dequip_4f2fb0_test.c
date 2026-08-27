// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../player_try_dequip_4f2fb0.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

struct nox_object_t {
	uint32_t marker;
};

typedef int32_t (*player_try_dequip_fn)(
	nox_object_t*, const nox_object_t*);
typedef int32_t (*dequip_lower_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "dequip result and flags must be exact dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_playerTryDequip_4F2FB0,
		player_try_dequip_fn: 1, default: 0),
	"player dequip must retain a const item pointer and exact int32 result");

static nox_object_t* expected_owner;
static const nox_object_t* expected_item;
static int32_t weapon_result;
static int32_t armor_result;
static unsigned int call_sequence;
static int failure_line;

static int32_t try_weapon(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t flag1,
		int32_t flag2) {
	call_sequence = call_sequence * 10U + 1U;
	if (owner != expected_owner || item != expected_item ||
		flag1 != INT32_C(1) || flag2 != INT32_C(1))
		failure_line = __LINE__;
	return weapon_result;
}

static int32_t try_armor(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t flag1,
		int32_t flag2) {
	call_sequence = call_sequence * 10U + 2U;
	if (owner != expected_owner || item != expected_item ||
		flag1 != INT32_C(1) || flag2 != INT32_C(1))
		failure_line = __LINE__;
	return armor_result;
}

// Freestanding semantic companion for the Go export used by the production
// const adapter. It independently checks the original callback order while the
// linked adapter checks that the public const pointer is forwarded unchanged.
int32_t nox_xxx_playerTryDequip_4F2FB0_go(
		nox_object_t* owner,
		nox_object_t* item) {
	dequip_lower_fn const weapon = try_weapon;
	dequip_lower_fn const armor = try_armor;
	if (weapon(owner, item, INT32_C(1), INT32_C(1)) != INT32_C(0))
		return INT32_C(1);
	if (armor(owner, item, INT32_C(1), INT32_C(1)) != INT32_C(0))
		return INT32_C(1);
	return INT32_C(0);
}

static int check_case(
		nox_object_t* owner,
		const nox_object_t* item,
		int32_t next_weapon_result,
		int32_t next_armor_result,
		int32_t expected_result,
		unsigned int expected_sequence) {
	player_try_dequip_fn const try_dequip = nox_xxx_playerTryDequip_4F2FB0;
	expected_owner = owner;
	expected_item = item;
	weapon_result = next_weapon_result;
	armor_result = next_armor_result;
	call_sequence = 0U;
	failure_line = 0;
	if (try_dequip(owner, item) != expected_result)
		return __LINE__;
	if (failure_line != 0)
		return failure_line;
	if (call_sequence != expected_sequence)
		return __LINE__;
	return 0;
}

int main(void) {
	nox_object_t owner = {.marker = UINT32_C(0xAAAAAAAA)};
	const nox_object_t item = {.marker = UINT32_C(0x55555555)};
	int line;

	line = check_case(&owner, &item, INT32_MIN, INT32_MAX, INT32_C(1), 1U);
	if (line != 0)
		return line;
	line = check_case(&owner, &item, INT32_MAX, INT32_MIN, INT32_C(1), 1U);
	if (line != 0)
		return line;
	line = check_case(&owner, &item, INT32_C(0), INT32_MIN, INT32_C(1), 12U);
	if (line != 0)
		return line;
	line = check_case(&owner, &item, INT32_C(0), INT32_MAX, INT32_C(1), 12U);
	if (line != 0)
		return line;
	line = check_case(&owner, &item, INT32_C(0), INT32_C(0), INT32_C(0), 12U);
	if (line != 0)
		return line;
	line = check_case(NULL, NULL, INT32_C(0), INT32_C(0), INT32_C(0), 12U);
	if (line != 0)
		return line;
	if (owner.marker != UINT32_C(0xAAAAAAAA) ||
		item.marker != UINT32_C(0x55555555))
		return __LINE__;
	return 0;
}
