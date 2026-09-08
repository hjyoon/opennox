#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../quest_journal_500540.h"

typedef int32_t (*quest_journal_read_fn)(void);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "00500B70 returns exactly four bytes");
_Static_assert(sizeof(sub_500B70()) == 4,
	"00500B70 call expressions must remain fixed-width");
_Static_assert(
	_Generic(&sub_500B70, quest_journal_read_fn: 1, default: 0),
	"00500B70 must preserve its int32_t no-argument ABI");

static int32_t next_result;
static unsigned int observed_calls;

int32_t sub_500B70(void) {
	++observed_calls;
	return next_result;
}

static int32_t check_call(quest_journal_read_fn read_journal) {
	return read_journal();
}

int main(void) {
	quest_journal_read_fn const read_journal = sub_500B70;

	next_result = INT32_C(1);
	assert(check_call(read_journal) == INT32_C(1));
	next_result = INT32_MIN;
	assert(check_call(read_journal) == INT32_MIN);
	next_result = INT32_MAX;
	assert(check_call(read_journal) == INT32_MAX);
	assert(observed_calls == 3);
	return 0;
}
