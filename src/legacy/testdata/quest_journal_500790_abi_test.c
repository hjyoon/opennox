#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../quest_journal_500540.h"

typedef void (*quest_journal_delete_entry_fn)(nox_quest_journal_native*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"00500790 links must remain native-pointer-sized");
_Static_assert(offsetof(nox_quest_journal_native, kind) == 132,
	"00500790 kind offset must remain 132");
_Static_assert(offsetof(nox_quest_journal_native, value) == 136,
	"00500790 value offset must remain 136");
#if UINTPTR_MAX == UINT32_MAX
_Static_assert(offsetof(nox_quest_journal_native, next) == 140,
	"00500790 PE32 next offset must remain 140");
_Static_assert(offsetof(nox_quest_journal_native, prev) == 144,
	"00500790 PE32 previous offset must remain 144");
_Static_assert(sizeof(nox_quest_journal_native) == 148,
	"00500790 PE32 entry size must remain 148");
#elif UINTPTR_MAX == UINT64_MAX
_Static_assert(offsetof(nox_quest_journal_native, next) == 144,
	"00500790 64-bit next offset must remain 144");
_Static_assert(offsetof(nox_quest_journal_native, prev) == 152,
	"00500790 64-bit previous offset must remain 152");
_Static_assert(sizeof(nox_quest_journal_native) == 160,
	"00500790 64-bit entry size must remain 160");
#else
#error "unsupported pointer width"
#endif
_Static_assert(
	_Generic(&sub_500790, quest_journal_delete_entry_fn: 1, default: 0),
	"00500790 must preserve its typed native-entry pointer ABI");

static nox_quest_journal_native* observed_entry;
static unsigned int observed_calls;

void sub_500790(nox_quest_journal_native* entry) {
	observed_entry = entry;
	++observed_calls;
}

static void check_call(quest_journal_delete_entry_fn delete_entry, nox_quest_journal_native* entry) {
	delete_entry(entry);
	assert(observed_entry == entry);
}

int main(void) {
	static nox_quest_journal_native first_entry;
	static nox_quest_journal_native second_entry;
	quest_journal_delete_entry_fn const delete_entry = sub_500790;

	check_call(delete_entry, &first_entry);
	check_call(delete_entry, &second_entry);
	assert(observed_calls == 2);
	return 0;
}
