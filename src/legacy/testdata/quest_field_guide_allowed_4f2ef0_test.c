// Freestanding C11 matrix companion for GAME.EXE 004F2EF0. This is not an
// active implementation: production uses server.QuestFieldGuideAllowed4F2EF0.
// It independently fixes the scalar conversions, row and family layout,
// native pointer-table width, sentinel order, and normalized result.
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct quest_field_guide_reward_row_4f2ef0 {
	uint8_t weight;
	uint32_t guide_id;
	uint32_t slots;
} quest_field_guide_reward_row_4f2ef0;

typedef uint32_t quest_field_guide_family_4f2ef0[6];

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "guide input and result must remain exact dwords");
_Static_assert(sizeof(uint32_t) == 4, "guide data must remain exact dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(quest_field_guide_reward_row_4f2ef0) == 12, "reward row size");
_Static_assert(offsetof(quest_field_guide_reward_row_4f2ef0, weight) == 0, "weight offset");
_Static_assert(offsetof(quest_field_guide_reward_row_4f2ef0, guide_id) == 4, "guide-ID offset");
_Static_assert(offsetof(quest_field_guide_reward_row_4f2ef0, slots) == 8, "slot-mask offset");
_Static_assert(sizeof(quest_field_guide_family_4f2ef0) == 24, "guide-family width");

int32_t quest_field_guide_allowed_fixture_4f2ef0(
	int32_t guide_id,
	quest_field_guide_reward_row_4f2ef0 const* rows,
	uint32_t const* const* families
) {
	size_t row_index = 0;
	uint32_t current_guide_id = rows[row_index].guide_id;
	uint32_t const target = (uint32_t)guide_id;
	int32_t allowed = INT32_C(0);

	if (current_guide_id != UINT32_C(0)) {
		for (;;) {
			if (current_guide_id == target && rows[row_index].slots != UINT32_C(0)) {
				allowed = INT32_C(1);
				break;
			}
			++row_index;
			current_guide_id = rows[row_index].guide_id;
			if (current_guide_id == UINT32_C(0))
				break;
		}
	}

	for (size_t family_index = 0; families[family_index] != NULL; ++family_index) {
		uint32_t const* family = families[family_index];
		if (family[0] == UINT32_C(0))
			continue;
		for (size_t member_index = 1;; ++member_index) {
			uint32_t const member = family[member_index];
			if (member == target) {
				allowed = INT32_C(1);
				break;
			}
			if (member == UINT32_C(0))
				break;
		}
	}
	return allowed != INT32_C(0) ? INT32_C(1) : INT32_C(0);
}

typedef int32_t (*quest_field_guide_allowed_signature_4f2ef0)(
	int32_t,
	quest_field_guide_reward_row_4f2ef0 const*,
	uint32_t const* const*
);

_Static_assert(
	_Generic(
		&quest_field_guide_allowed_fixture_4f2ef0,
		quest_field_guide_allowed_signature_4f2ef0: 1,
		default: 0
	),
	"Quest field-guide admission must keep exact int32 scalars and native pointers"
);

int main(void) {
	quest_field_guide_reward_row_4f2ef0 const rows[] = {
		{UINT8_C(0), UINT32_C(7), UINT32_C(0)},
		{UINT8_C(0), UINT32_C(8), UINT32_C(2)},
		{UINT8_C(0), UINT32_C(7), UINT32_C(4)},
		{UINT8_C(0), UINT32_C(0), UINT32_C(31)},
	};
	quest_field_guide_reward_row_4f2ef0 const first_sentinel[] = {
		{UINT8_C(0), UINT32_C(0), UINT32_C(31)},
	};
	quest_field_guide_reward_row_4f2ef0 const raw_bits[] = {
		{UINT8_C(0), UINT32_MAX, UINT32_C(1)},
		{UINT8_C(0), UINT32_C(0), UINT32_C(31)},
	};
	quest_field_guide_family_4f2ef0 const family_24 = {
		UINT32_C(24), UINT32_C(7), UINT32_C(8), UINT32_C(25), UINT32_C(26), UINT32_C(0),
	};
	quest_field_guide_family_4f2ef0 const empty_header = {
		UINT32_C(0), UINT32_C(7), UINT32_C(0), UINT32_C(0), UINT32_C(0), UINT32_C(0),
	};
	uint32_t const* const families[] = {family_24, NULL};
	uint32_t const* const empty_families[] = {empty_header, NULL};
	uint32_t const* const no_families[] = {NULL};

	_Static_assert(sizeof(families) == 2 * sizeof(void*), "native family pointer table");
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(7), rows, families) != INT32_C(1))
		return __LINE__;
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(8), rows, families) != INT32_C(1))
		return __LINE__;
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(6), rows, families) != INT32_C(0))
		return __LINE__;
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(24), first_sentinel, families) != INT32_C(0))
		return __LINE__;
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(25), first_sentinel, families) != INT32_C(1))
		return __LINE__;
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(0), first_sentinel, families) != INT32_C(1))
		return __LINE__;
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(7), first_sentinel, empty_families) != INT32_C(0))
		return __LINE__;
	if (quest_field_guide_allowed_fixture_4f2ef0(INT32_C(-1), raw_bits, no_families) != INT32_C(1))
		return __LINE__;
	return 0;
}
