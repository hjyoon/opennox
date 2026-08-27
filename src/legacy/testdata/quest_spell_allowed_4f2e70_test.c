// Freestanding C11 matrix companion for GAME.EXE 004F2E70. This is not an
// active implementation: production uses server.QuestSpellAllowed4F2E70.
// It independently fixes the scalar conversions, row layout, sentinel scan,
// and fallback domains for every supported compiler frontend.
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct quest_spell_reward_row_4f2e70 {
	uint8_t weight;
	uint32_t spell_id;
	uint32_t slots;
} quest_spell_reward_row_4f2e70;

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell input and result must remain exact dwords");
_Static_assert(sizeof(uint32_t) == 4, "reward row fields must remain exact dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(quest_spell_reward_row_4f2e70) == 12, "reward row size");
_Static_assert(offsetof(quest_spell_reward_row_4f2e70, weight) == 0, "weight offset");
_Static_assert(offsetof(quest_spell_reward_row_4f2e70, spell_id) == 4, "spell-ID offset");
_Static_assert(offsetof(quest_spell_reward_row_4f2e70, slots) == 8, "slot-mask offset");

static int explicit_spell_4f2e70(int32_t spell_id) {
	switch (spell_id) {
	case INT32_C(46):
	case INT32_C(47):
	case INT32_C(48):
	case INT32_C(49):
	case INT32_C(122):
	case INT32_C(123):
	case INT32_C(124):
	case INT32_C(125):
		return 1;
	default:
		return 0;
	}
}

int32_t quest_spell_allowed_fixture_4f2e70(
	int32_t spell_id,
	quest_spell_reward_row_4f2e70 const* rows
) {
	size_t index = 0;
	uint32_t current_spell_id = rows[index].spell_id;
	int32_t allowed = INT32_C(0);

	if (current_spell_id != UINT32_C(0)) {
		uint32_t const target = (uint32_t)spell_id;
		for (;;) {
			if (current_spell_id == target && rows[index].slots != UINT32_C(0)) {
				allowed = INT32_C(1);
				break;
			}
			++index;
			current_spell_id = rows[index].spell_id;
			if (current_spell_id == UINT32_C(0))
				break;
		}
	}
	if (explicit_spell_4f2e70(spell_id))
		allowed = INT32_C(1);
	if (spell_id >= INT32_C(75) && spell_id <= INT32_C(114))
		return INT32_C(1);
	return allowed;
}

int main(void) {
	quest_spell_reward_row_4f2e70 const rows[] = {
		{UINT8_C(0), UINT32_C(7), UINT32_C(0)},
		{UINT8_C(0), UINT32_C(8), UINT32_C(2)},
		{UINT8_C(0), UINT32_C(7), UINT32_C(4)},
		{UINT8_C(0), UINT32_C(0), UINT32_C(31)},
	};
	quest_spell_reward_row_4f2e70 raw_bits[] = {
		{UINT8_C(0), UINT32_MAX, UINT32_C(1)},
		{UINT8_C(0), UINT32_C(0), UINT32_C(31)},
	};
	quest_spell_reward_row_4f2e70 const first_sentinel[] = {
		{UINT8_C(0), UINT32_C(0), UINT32_C(31)},
	};

	if (quest_spell_allowed_fixture_4f2e70(INT32_C(7), rows) != INT32_C(1))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(8), rows) != INT32_C(1))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(6), rows) != INT32_C(0))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(46), first_sentinel) != INT32_C(1))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(75), first_sentinel) != INT32_C(1))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(114), first_sentinel) != INT32_C(1))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(115), first_sentinel) != INT32_C(0))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(122), first_sentinel) != INT32_C(1))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(-1), raw_bits) != INT32_C(1))
		return __LINE__;
	raw_bits[0].slots = UINT32_C(0);
	if (quest_spell_allowed_fixture_4f2e70(INT32_C(-1), raw_bits) != INT32_C(0))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_MIN, first_sentinel) != INT32_C(0))
		return __LINE__;
	if (quest_spell_allowed_fixture_4f2e70(INT32_MAX, first_sentinel) != INT32_C(0))
		return __LINE__;
	return 0;
}
