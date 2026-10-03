#include <stdio.h>

// Match the existing standalone layout probes: unrelated Win32-only aggregate
// assertions are suppressed, but the fields used here are checked explicitly.
#define _Static_assert(...)
#include "../defs.h"
#undef _Static_assert

_Static_assert(offsetof(nox_object_t, obj_subclass) ==
	(sizeof(void*) == 4 ? 12 : 16), "native subclass offset");
_Static_assert(offsetof(nox_object_t, field_12) ==
	(sizeof(void*) == 4 ? 48 : 52), "unrelated DWORD twelve offset");
_Static_assert(offsetof(nox_player_update_data_t, state) == 88,
	"cached player state offset");

static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static nox_player_update_data_t* cached_update;
static nox_player_update_data_t* replacement_update;
static unsigned int sequence;
static int failure;
static int change_subclass;
static int change_state;
static int replace_update;
static int state_calls;

#define CHECK(test) do { if (!(test)) { \
	fprintf(stderr, "line %d: %s (state=%u subclass=%#x field12=%#x sequence=%u)\n", \
		__LINE__, #test, cached_update->state, expected_item->obj_subclass, \
		expected_item->field_12, sequence); return 1; } } while (0)

static int sub_53E3A0(nox_object_t* owner, nox_object_t* item) {
	if (owner != expected_owner || item != expected_item) failure = 1;
	sequence = sequence * 10U + 7U;
	return -7;
}

static int nox_xxx_unitArmorInventoryEquipFlags_415C70(nox_object_t* item) {
	if (item != expected_item || (item->obj_flags & UINT32_C(0x10000100))) failure = 1;
	sequence = sequence * 10U + 1U;
	return UINT32_C(0x2000000);
}

static int nox_xxx_netReportDequip_4D8590(int index, const nox_object_t* item) {
	if (index != 31 || item != expected_item ||
		cached_update->player->field_0 != UINT32_C(0x1000005)) failure = 1;
	sequence = sequence * 10U + 2U;
	return -1;
}

static int nox_xxx_netReportDequip_4D84C0(int index, const nox_object_t* item) {
	if (index != 255 || item != expected_item) failure = 1;
	sequence = sequence * 10U + 3U;
	return -1;
}

static int nox_xxx_recalculateArmorVal_53E300(nox_object_t* owner) {
	if (owner != expected_owner) failure = 1;
	sequence = sequence * 10U + 4U;
	return -1;
}

static void nox_xxx_itemApplyDisengageEffect_4F3030(nox_object_t* item, nox_object_t* owner) {
	if (owner != expected_owner || item != expected_item) failure = 1;
	sequence = sequence * 10U + 5U;
	if (change_subclass >= 0) item->obj_subclass = (unsigned int)change_subclass;
	if (change_state >= 0) cached_update->state = (unsigned char)change_state;
	if (replace_update) owner->data_update = replacement_update;
}

static int nox_xxx_playerSetState_4FA020(nox_object_t* owner, int state) {
	if (owner != expected_owner || state != 13) failure = 1;
	sequence = sequence * 10U + 6U;
	state_calls++;
	// The decision uses cached_update, while the called state setter receives the
	// original owner and therefore sees its live update binding.
	((nox_player_update_data_t*)owner->data_update)->state = (unsigned char)state;
	return 0;
}

// PRODUCTION_BODY_53E430

static void prepare(unsigned int state, unsigned int subclass, unsigned int field12) {
	memset(expected_owner, 0, sizeof(*expected_owner));
	memset(expected_item, 0, sizeof(*expected_item));
	memset(replacement_update, 0, sizeof(*replacement_update));
	expected_owner->obj_class = 4;
	expected_owner->inv_first_item = expected_item;
	expected_owner->data_update = cached_update;
	expected_item->obj_class = UINT32_C(0x2000000);
	expected_item->obj_subclass = subclass;
	expected_item->field_12 = field12;
	expected_item->obj_flags = UINT32_C(0x10000110);
	cached_update->state = (unsigned char)state;
	cached_update->player->field_0 = UINT32_C(0x3000005);
	cached_update->player->playerInd = 31;
	sequence = 0;
	failure = 0;
	change_subclass = -1;
	change_state = -1;
	replace_update = 0;
	state_calls = 0;
}

int main(void) {
	expected_owner = calloc(1, sizeof(*expected_owner));
	expected_item = calloc(1, sizeof(*expected_item));
	cached_update = calloc(1, sizeof(*cached_update));
	replacement_update = calloc(1, sizeof(*replacement_update));
	nox_playerInfo* player = calloc(1, sizeof(*player));
	nox_object_t* predecessor = calloc(1, sizeof(*predecessor));
	if (!expected_owner || !expected_item || !cached_update || !replacement_update || !player || !predecessor)
		return 2;
	cached_update->player = player;
	if (sizeof(void*) == 8) {
		CHECK((uintptr_t)expected_owner > UINT32_MAX && (uintptr_t)expected_item > UINT32_MAX &&
			(uintptr_t)cached_update > UINT32_MAX && (uintptr_t)player > UINT32_MAX &&
			(uintptr_t)predecessor > UINT32_MAX);
	}
	const unsigned int subclasses[] = {0, 2, 0x100, 0x102};
	for (unsigned int state = 0; state < 256; state++) {
		for (unsigned int sc = 0; sc < 4; sc++) {
			for (unsigned int trap = 0; trap < 2; trap++) {
				for (unsigned int reports = 0; reports < 4; reports++) {
					prepare(state, subclasses[sc], trap * 2);
					predecessor->inv_next_item = expected_item;
					expected_owner->inv_first_item = predecessor;
					int reset = (subclasses[sc] & 2) && state >= 15 && state <= 17;
					unsigned int want_sequence = 1;
					if (reports & 1) want_sequence = want_sequence * 10U + 2U;
					if (reports & 2) want_sequence = want_sequence * 10U + 3U;
					want_sequence = want_sequence * 100U + 45U;
					if (reset) want_sequence = want_sequence * 10U + 6U;
					CHECK(sub_53E430(expected_owner, expected_item, reports & 1, reports & 2) == 1);
					CHECK(!failure && sequence == want_sequence && state_calls == reset);
					CHECK(cached_update->state == (reset ? 13 : state));
					CHECK(expected_item->obj_flags == UINT32_C(0x10));
					CHECK(player->field_0 == UINT32_C(0x1000005));
					CHECK(expected_item->obj_subclass == subclasses[sc] && expected_item->field_12 == trap * 2);
				}
			}
		}
	}
	// Both the subclass and cached state must be read after disengage.
	prepare(14, 0, 0);
	change_subclass = 2;
	change_state = 16;
	CHECK(sub_53E430(expected_owner, expected_item, 0, 0) == 1);
	CHECK(!failure && sequence == 1456 && cached_update->state == 13);
	prepare(16, 2, 2);
	change_subclass = 0;
	CHECK(sub_53E430(expected_owner, expected_item, 0, 0) == 1);
	CHECK(!failure && sequence == 145 && cached_update->state == 16);
	prepare(16, 2, 0);
	change_state = 18;
	CHECK(sub_53E430(expected_owner, expected_item, 0, 0) == 1);
	CHECK(!failure && sequence == 145 && cached_update->state == 18);
	prepare(16, 2, 0);
	replace_update = 1;
	CHECK(sub_53E430(expected_owner, expected_item, -1, 2) == 1);
	CHECK(!failure && sequence == 123456 && state_calls == 1);
	CHECK(cached_update->state == 16 && replacement_update->state == 13);
	// Eligibility failures must leave equipment and reports untouched.
	for (unsigned int gate = 0; gate < 6; gate++) {
		prepare(16, 2, 0);
		if (gate == 0) expected_item->obj_class = 0;
		if (gate == 1) expected_item->obj_flags &= ~UINT32_C(0x100);
		if (gate == 2) expected_owner->obj_class = 0;
		if (gate == 3) expected_owner->inv_first_item = NULL;
		if (gate == 4) expected_owner->data_update = NULL;
		if (gate == 5) cached_update->player = NULL;
		unsigned int old_flags = expected_item->obj_flags;
		CHECK(sub_53E430(expected_owner, expected_item, 1, 1) == 0);
		CHECK(!failure && sequence == 0 && state_calls == 0 && expected_item->obj_flags == old_flags);
		CHECK(cached_update->state == 16 && player->field_0 == UINT32_C(0x3000005));
		cached_update->player = player;
	}
	prepare(16, 2, 0);
	expected_owner->obj_class = 2 | 4;
	CHECK(sub_53E430(expected_owner, expected_item, 1, 1) == -7);
	CHECK(!failure && sequence == 7 && state_calls == 0);
	free(predecessor);
	free(player);
	free(replacement_update);
	free(cached_update);
	free(expected_item);
	free(expected_owner);
	return 0;
}
