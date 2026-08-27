// Freestanding C11 matrix companion for GAME.EXE 004F36F0. Production uses
// the Go restoration; this independently locks the four-argument ABI, native
// pointer width, exact callback result, guard order, and live post-pickup state.
#include "../inventory_serv_place_4f36f0.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*inventory_place_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);
typedef int32_t (*pickup_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);
typedef void (*collide_fn)(nox_object_t*);

struct nox_object_t {
	uint16_t type_ind;
	uint16_t carry_capacity;
	uint32_t object_class;
	uint32_t object_flags;
	pickup_fn pickup;
	collide_fn collide;
	int32_t script_pickup_func;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4 && sizeof(uint32_t) == 4,
	"callback scalars, flags, and result must remain exact 32-bit values");
_Static_assert(
	_Generic(&nox_xxx_inventoryServPlace_4F36F0,
		inventory_place_fn: 1, default: 0),
	"inventory placement must retain two native pointers, two int32_t values, and int32_t result");

#define FLAG_DESTROYED UINT32_C(0x20)
#define FLAG_NO_COLLIDE UINT32_C(0x40)
#define FLAG_DEAD UINT32_C(0x8000)
#define CLASS_UNITS UINT32_C(0x06)

static unsigned int trace[32];
static size_t trace_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t allowed_result;
static int32_t pickup_result;
static int32_t default_result;

static void event(unsigned int value) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = value;
}

static void expect_trace(const unsigned int* values, size_t count) {
	size_t i;
	assert(trace_count == count);
	for (i = 0U; i < count; i++)
		assert(trace[i] == values[i]);
}

static void reset_case(nox_object_t* owner, nox_object_t* item) {
	trace_count = 0U;
	expected_owner = owner;
	expected_item = item;
	expected_arg3 = INT32_C(0);
	expected_arg4 = INT32_C(0);
	allowed_result = INT32_C(1);
	pickup_result = INT32_C(1);
	default_result = INT32_C(1);
}

static int32_t item_type_allowed(uint16_t type_ind) {
	event(1U);
	assert(expected_item != NULL);
	assert(type_ind == expected_item->type_ind);
	return allowed_result;
}

static int32_t custom_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	event(2U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	item->object_flags = UINT32_C(0xA5A500C3);
	return pickup_result;
}

static int32_t fallback_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	event(3U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return default_result;
}

static void refresh_collide(nox_object_t* item) {
	event(4U);
	assert(item == expected_item);
	assert(item->object_flags == UINT32_C(0xA5A50083));
	item->script_pickup_func = INT32_C(0x10203040);
}

static void pickup_script(nox_object_t* owner, nox_object_t* item) {
	event(5U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(item->script_pickup_func == INT32_C(0x10203040));
	item->script_pickup_func = INT32_MIN;
}

int32_t nox_xxx_inventoryServPlace_4F36F0(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	pickup_fn pickup;
	int32_t result;
	uint32_t flags;

	if (owner == NULL)
		return INT32_C(0);
	if (item == NULL)
		return INT32_C(0);
	if (owner->carry_capacity == UINT16_C(0))
		return INT32_C(0);
	if (((uint8_t)item->object_flags & (uint8_t)FLAG_DESTROYED) != 0U)
		return INT32_C(0);
	if ((owner->object_flags & FLAG_DEAD) != 0U)
		return INT32_C(0);
	if (item_type_allowed(item->type_ind) == INT32_C(0))
		return INT32_C(0);
	if (((uint8_t)owner->object_class & (uint8_t)CLASS_UNITS) == 0U)
		return INT32_C(0);

	pickup = item->pickup;
	result = pickup != NULL
		? pickup(owner, item, arg3, arg4)
		: fallback_pickup(owner, item, arg3, arg4);
	if (result == INT32_C(0))
		return result;

	flags = item->object_flags;
	if ((flags & FLAG_NO_COLLIDE) != 0U) {
		item->object_flags = flags & ~FLAG_NO_COLLIDE;
		if (item->collide != NULL)
			item->collide(item);
	}
	if (item->script_pickup_func != INT32_C(-1)) {
		pickup_script(owner, item);
		item->script_pickup_func = INT32_C(-1);
	}
	return result;
}

static nox_object_t valid_owner(void) {
	nox_object_t owner = {0};
	owner.carry_capacity = UINT16_C(1);
	owner.object_class = UINT32_C(0x80000004);
	return owner;
}

static nox_object_t valid_item(void) {
	nox_object_t item = {0};
	item.type_ind = UINT16_C(0xF123);
	item.object_flags = UINT32_C(1);
	item.script_pickup_func = INT32_C(-1);
	return item;
}

static void test_exact_custom_result_and_live_state(void) {
	nox_object_t owner = valid_owner();
	nox_object_t item = valid_item();
	static const unsigned int want[] = {1U, 2U, 4U, 5U};

	reset_case(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	pickup_result = INT32_MIN;
	item.pickup = custom_pickup;
	item.collide = refresh_collide;
	assert(nox_xxx_inventoryServPlace_4F36F0(
		&owner, &item, expected_arg3, expected_arg4) == INT32_MIN);
	expect_trace(want, sizeof(want) / sizeof(want[0]));
	assert(item.object_flags == UINT32_C(0xA5A50083));
	assert(item.script_pickup_func == INT32_C(-1));
}

static void test_default_and_zero_short_circuit(void) {
	nox_object_t owner = valid_owner();
	nox_object_t item = valid_item();
	static const unsigned int want_default[] = {1U, 3U};
	static const unsigned int want_zero[] = {1U, 2U};

	reset_case(&owner, &item);
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	default_result = INT32_MAX;
	assert(nox_xxx_inventoryServPlace_4F36F0(
		&owner, &item, expected_arg3, expected_arg4) == INT32_MAX);
	expect_trace(want_default, sizeof(want_default) / sizeof(want_default[0]));

	reset_case(&owner, &item);
	expected_arg3 = INT32_C(3);
	expected_arg4 = INT32_C(4);
	pickup_result = INT32_C(0);
	item.pickup = custom_pickup;
	item.collide = refresh_collide;
	item.script_pickup_func = INT32_C(7);
	assert(nox_xxx_inventoryServPlace_4F36F0(
		&owner, &item, expected_arg3, expected_arg4) == INT32_C(0));
	expect_trace(want_zero, sizeof(want_zero) / sizeof(want_zero[0]));
	assert(item.object_flags == UINT32_C(0xA5A500C3));
	assert(item.script_pickup_func == INT32_C(7));
}

static void test_all_entry_guards(void) {
	nox_object_t owner = valid_owner();
	nox_object_t item = valid_item();

	reset_case(&owner, &item);
	assert(nox_xxx_inventoryServPlace_4F36F0(NULL, &item, 1, 1) == 0);
	assert(nox_xxx_inventoryServPlace_4F36F0(&owner, NULL, 1, 1) == 0);
	owner.carry_capacity = 0;
	assert(nox_xxx_inventoryServPlace_4F36F0(&owner, &item, 1, 1) == 0);
	owner = valid_owner();
	item.object_flags = FLAG_DESTROYED;
	assert(nox_xxx_inventoryServPlace_4F36F0(&owner, &item, 1, 1) == 0);
	item = valid_item();
	owner.object_flags = FLAG_DEAD;
	assert(nox_xxx_inventoryServPlace_4F36F0(&owner, &item, 1, 1) == 0);
	owner = valid_owner();
	allowed_result = 0;
	assert(nox_xxx_inventoryServPlace_4F36F0(&owner, &item, 1, 1) == 0);
	assert(trace_count == 1U);
	allowed_result = 1;
	owner.object_class = UINT32_C(0xFFFFFFF8);
	assert(nox_xxx_inventoryServPlace_4F36F0(&owner, &item, 1, 1) == 0);
	assert(trace_count == 2U);
}

int main(void) {
	nox_object_t owner = valid_owner();
	nox_object_t item = valid_item();
	if (sizeof(void*) == 8U) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
	}
	test_exact_custom_result_and_live_state();
	test_default_and_zero_short_circuit();
	test_all_entry_guards();
	return 0;
}
