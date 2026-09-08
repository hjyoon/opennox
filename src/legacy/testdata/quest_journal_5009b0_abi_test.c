#include <assert.h>
#include <limits.h>
#include <stdint.h>
#include <string.h>

#include "../quest_journal_500540.h"

typedef uint32_t (*quest_journal_qualify_fn)(char*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint32_t) == 4, "005009B0 returns exactly four bytes");
_Static_assert(sizeof(char*) == sizeof(uintptr_t),
	"005009B0 name pointers must remain native-pointer-sized");
_Static_assert(
	_Generic(&sub_5009B0, quest_journal_qualify_fn: 1, default: 0),
	"005009B0 must preserve its uint32_t mutable-string pointer ABI");

static char* observed_name;
static unsigned int observed_calls;

uint32_t sub_5009B0(char* name) {
	observed_name = name;
	++observed_calls;
	if (strchr(name, ':') != NULL) {
		return (uint32_t)(strlen(name) + 1u);
	}
	return UINT32_C(0);
}

static uint32_t check_call(quest_journal_qualify_fn qualify, char* name) {
	uint32_t const result = qualify(name);
	assert(observed_name == name);
	return result;
}

int main(void) {
	char qualified_name[] = "War01a:Count";
	char unqualified_name[] = "Count";
	quest_journal_qualify_fn const qualify = sub_5009B0;

	assert(check_call(qualify, qualified_name) == sizeof(qualified_name));
	assert(strcmp(observed_name, "War01a:Count") == 0);
	assert(check_call(qualify, unqualified_name) == UINT32_C(0));
	assert(strcmp(observed_name, "Count") == 0);
	assert(observed_calls == 2);
	return 0;
}
