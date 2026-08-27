// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../item_apply_engage_4f2ff0.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct modifier_effect modifier_effect;
typedef void (*engage_callback)(
	void*, nox_object_t*, const nox_object_t*);
typedef void (*item_apply_engage_fn)(nox_object_t*, nox_object_t*);

struct modifier_effect {
	uintptr_t marker;
	engage_callback engage;
};

typedef struct modifier_data {
	modifier_effect* modifiers[4];
} modifier_data;

struct nox_object_t {
	uintptr_t marker;
	modifier_data* init_data;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(engage_callback) == sizeof(void*), "engage callbacks must use native pointer width");
_Static_assert(
	_Generic(&nox_xxx_itemApplyEngageEffect_4F2FF0,
		item_apply_engage_fn: 1, default: 0),
	"item engage dispatch must return void and receive two native object pointers");

static nox_object_t* expected_item;
static nox_object_t* expected_owner;
static modifier_data* expected_data;
static modifier_effect* replacement_modifier;
static unsigned int call_sequence;
static int failure_line;

static void first_engage(
		void* effect,
		nox_object_t* owner,
		const nox_object_t* item) {
	call_sequence = call_sequence * 10U + 1U;
	if (effect != expected_data->modifiers[2] ||
		owner != expected_owner || item != expected_item)
		failure_line = __LINE__;
	// InitData itself is cached, while slot three remains a live load.
	expected_item->init_data = NULL;
	expected_data->modifiers[3] = replacement_modifier;
}

static void replacement_engage(
		void* effect,
		nox_object_t* owner,
		const nox_object_t* item) {
	call_sequence = call_sequence * 10U + 2U;
	if (effect != replacement_modifier ||
		owner != expected_owner || item != expected_item)
		failure_line = __LINE__;
}

// Freestanding semantic companion for GAME.EXE 004F2FF0. Production uses the
// Go restoration; this definition lets every target frontend independently
// check the exact public declaration, pointer width, and callback order.
void nox_xxx_itemApplyEngageEffect_4F2FF0(
		nox_object_t* item,
		nox_object_t* owner) {
	modifier_data* const data = item->init_data;
	for (size_t slot = 2; slot < 4; ++slot) {
		modifier_effect* const modifier = data->modifiers[slot];
		if (modifier == NULL)
			continue;
		engage_callback const engage = modifier->engage;
		if (engage != NULL)
			engage(modifier, owner, item);
	}
}

int main(void) {
	modifier_effect first = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.engage = first_engage,
	};
	modifier_effect skipped = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.engage = NULL,
	};
	modifier_effect replacement = {
		.marker = UINTPTR_MAX - (uintptr_t)3,
		.engage = replacement_engage,
	};
	modifier_data data = {
		.modifiers = {NULL, NULL, &first, &skipped},
	};
	modifier_data other_data = {0};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)4,
		.init_data = &data,
	};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)5,
		.init_data = &other_data,
	};
	item_apply_engage_fn const apply = nox_xxx_itemApplyEngageEffect_4F2FF0;

	expected_item = &item;
	expected_owner = &owner;
	expected_data = &data;
	replacement_modifier = &replacement;
	call_sequence = 0U;
	failure_line = 0;
	apply(&item, &owner);
	assert(failure_line == 0);
	assert(call_sequence == 12U);
	assert(item.init_data == NULL);
	assert(data.modifiers[3] == &replacement);
	assert(first.marker == UINTPTR_MAX - (uintptr_t)1);
	assert(replacement.marker == UINTPTR_MAX - (uintptr_t)3);

	// Nil modifier and nil callback slots are skipped without invocation.
	item.init_data = &data;
	data.modifiers[2] = NULL;
	data.modifiers[3] = &skipped;
	call_sequence = 0U;
	apply(&item, NULL);
	assert(call_sequence == 0U);

	if (sizeof(void*) == 8) {
		assert((uintptr_t)&item > UINT32_MAX);
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&data > UINT32_MAX);
		assert((uintptr_t)&first > UINT32_MAX);
	}
	return 0;
}
