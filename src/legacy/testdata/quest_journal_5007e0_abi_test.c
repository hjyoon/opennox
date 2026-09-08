#include <assert.h>
#include <limits.h>
#include <stdint.h>
#include <string.h>

#include "../quest_journal_500540.h"

typedef void (*quest_journal_delete_pattern_fn)(char*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(char*) == sizeof(uintptr_t),
	"005007E0 pattern pointers must remain native-pointer-sized");
_Static_assert(
	_Generic(&sub_5007E0, quest_journal_delete_pattern_fn: 1, default: 0),
	"005007E0 must preserve its void mutable-string pointer ABI");

static char* observed_pattern;
static unsigned int observed_calls;

void sub_5007E0(char* pattern) {
	observed_pattern = pattern;
	++observed_calls;
}

static void check_call(quest_journal_delete_pattern_fn delete_pattern, char* pattern) {
	delete_pattern(pattern);
	assert(observed_pattern == pattern);
}

int main(void) {
	char first_pattern[] = "War01a:*";
	char second_pattern[] = "*:Count";
	quest_journal_delete_pattern_fn const delete_pattern = sub_5007E0;

	check_call(delete_pattern, first_pattern);
	assert(strcmp(observed_pattern, "War01a:*") == 0);
	check_call(delete_pattern, second_pattern);
	assert(strcmp(observed_pattern, "*:Count") == 0);
	assert(observed_calls == 2);
	return 0;
}
