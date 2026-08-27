// Freestanding C11 matrix companion for GAME.EXE 004F3070. Production uses
// the Go restoration; this file independently locks the retained public ABI,
// native pointer width, callback order, live reloads, and signed weight rules.
#include "../inventory_put_4f3070.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct test_player test_player;
typedef struct test_update test_update;
typedef void (*inventory_put_fn)(nox_object_t*, nox_object_t*, int32_t);

struct test_player {
	uintptr_t marker;
	uint8_t index;
	uint32_t protect;
	uint32_t overweight;
};

struct test_update {
	uintptr_t marker;
	test_player* player;
};

struct nox_object_t {
	uintptr_t marker;
	uint32_t flags;
	uint32_t object_class;
	uint8_t weight;
	uint16_t capacity;
	nox_object_t* holder;
	nox_object_t* next;
	nox_object_t* previous;
	nox_object_t* first;
	nox_object_t* owner;
	test_update* update;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4, "report must remain a 32-bit integer");
_Static_assert(
	_Generic(&nox_xxx_inventoryPutImpl_4F3070,
		inventory_put_fn: 1, default: 0),
	"inventory insertion must return void and retain two native pointers plus int32_t");

static test_update* live_update;
static test_player* cached_player;
static test_player* replacement_player;
static nox_object_t* live_weight_head;
static unsigned int sequence;
static int failure_line;

static void set_owner(nox_object_t* owner, nox_object_t* item) {
	sequence = sequence * 10U + 1U;
	item->owner = owner;
	owner->object_class = UINT32_C(0x04);
	owner->update = live_update;
}

static void report_pickup(uint8_t index, nox_object_t* item) {
	sequence = sequence * 10U + 2U;
	if (index != cached_player->index || item == NULL)
		failure_line = __LINE__;
	cached_player->protect = UINT32_C(99);
	live_update->player = replacement_player;
}

static void protect_item(uint32_t protect, nox_object_t* item) {
	sequence = sequence * 10U + 3U;
	if (protect != UINT32_C(99) || item == NULL)
		failure_line = __LINE__;
	item->holder->first = live_weight_head;
	item->holder->capacity = UINT16_C(200);
	item->object_class = UINT32_C(0x40);
}

static void audio_event(int32_t id, nox_object_t* owner, int32_t kind, uint32_t code) {
	sequence = sequence * 10U + 4U;
	if (id != INT32_C(820) || owner == NULL || kind != 0 || code != 0)
		failure_line = __LINE__;
}

void nox_xxx_inventoryPutImpl_4F3070(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t report) {
	if (owner == NULL || item == NULL)
		return;
	if (((uint8_t)owner->flags & UINT8_C(0x20)) != 0)
		return;
	if (((uint8_t)item->flags & UINT8_C(0x20)) != 0)
		return;

	item->previous = NULL;
	item->next = owner->first;
	if (owner->first != NULL)
		owner->first->previous = item;
	owner->first = item;
	item->holder = owner;
	set_owner(owner, item);

	if (((uint8_t)owner->object_class & UINT8_C(0x04)) != 0) {
		test_update* const update = owner->update;
		test_player* const player = update->player;
		if (report != 0)
			report_pickup(player->index, item);
		protect_item(player->protect, item);

		int32_t weight = 0;
		for (nox_object_t* current = owner->first; current != NULL;) {
			uint8_t const item_weight = current->weight;
			current = current->next;
			weight += (int32_t)item_weight;
		}
		player->overweight = weight > (int32_t)owner->capacity ? 1U : 0U;
	}

	if (((uint8_t)item->object_class & UINT8_C(0x40)) != 0)
		audio_event(INT32_C(820), owner, 0, 0);
}

int main(void) {
	test_player initial = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.index = UINT8_C(7),
		.protect = UINT32_C(77),
	};
	test_player replacement = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.index = UINT8_C(9),
	};
	test_update update = {
		.marker = UINTPTR_MAX - (uintptr_t)3,
		.player = &initial,
	};
	nox_object_t stale = {.marker = UINTPTR_MAX - (uintptr_t)4};
	nox_object_t weight2 = {
		.marker = UINTPTR_MAX - (uintptr_t)5,
		.weight = UINT8_C(5),
	};
	nox_object_t weight1 = {
		.marker = UINTPTR_MAX - (uintptr_t)6,
		.weight = UINT8_C(250),
		.next = &weight2,
	};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)7,
		.flags = UINT32_C(0x00002000),
		.object_class = UINT32_C(0x00000400),
		.first = &stale,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)8,
		.flags = UINT32_C(0x00002000),
		.object_class = UINT32_C(0x00004000),
		.previous = &weight2,
	};
	inventory_put_fn const put = nox_xxx_inventoryPutImpl_4F3070;

	live_update = &update;
	cached_player = &initial;
	replacement_player = &replacement;
	live_weight_head = &weight1;
	sequence = 0U;
	failure_line = 0;
	put(&owner, &item, -1);
	assert(failure_line == 0);
	assert(sequence == 1234U);
	assert(item.previous == NULL);
	assert(item.next == &stale);
	assert(stale.previous == &item);
	assert(item.holder == &owner && item.owner == &owner);
	assert(owner.first == &weight1);
	assert(update.player == &replacement);
	assert(initial.overweight == UINT32_C(1));
	assert(replacement.overweight == UINT32_C(0));

	// The original signed comparison treats a wrapped negative sum as not
	// overweight. Lock that arithmetic without constructing millions of nodes.
	assert((int32_t)(UINT32_C(0x7fffffff) + UINT32_C(1)) < 0);
	assert((int32_t)UINT32_C(65535) <= (int32_t)UINT16_MAX);

	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
		assert((uintptr_t)&update > UINT32_MAX);
		assert((uintptr_t)&initial > UINT32_MAX);
	}
	return 0;
}
