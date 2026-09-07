#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../quest_journal_500540.h"

typedef nox_quest_journal_native* (*quest_journal_set_fn)(char*, int32_t);

_Static_assert(offsetof(nox_quest_journal_native, kind) == 132,
	"00500540 kind must remain at byte 132");
_Static_assert(offsetof(nox_quest_journal_native, value) == 136,
	"00500540 value must remain at byte 136");
_Static_assert(
	_Generic(&nox_xxx_journalQuestSet_500540, quest_journal_set_fn: 1, default: 0),
	"00500540 must preserve its native-pointer, signed-dword ABI");
_Static_assert(
	_Generic(&nox_xxx_journalQuestSetBool_5006B0, quest_journal_set_fn: 1, default: 0),
	"005006B0 must preserve its native-pointer, signed-dword ABI");

#if UINTPTR_MAX == UINT32_MAX
_Static_assert(offsetof(nox_quest_journal_native, next) == 140,
	"PE32 next link must remain at byte 140");
_Static_assert(offsetof(nox_quest_journal_native, prev) == 144,
	"PE32 previous link must remain at byte 144");
_Static_assert(sizeof(nox_quest_journal_native) == 148,
	"PE32 quest-journal node must remain 148 bytes");
#elif UINTPTR_MAX == UINT64_MAX
_Static_assert(offsetof(nox_quest_journal_native, next) == 144,
	"native 64-bit next link must remain at byte 144");
_Static_assert(offsetof(nox_quest_journal_native, prev) == 152,
	"native 64-bit previous link must remain at byte 152");
_Static_assert(sizeof(nox_quest_journal_native) == 160,
	"native 64-bit quest-journal node must remain 160 bytes");
#else
#error unsupported pointer width
#endif

static nox_quest_journal_native entries[4];
static nox_quest_journal_native* head;
static unsigned int used;

static nox_quest_journal_native* set_new(uint32_t kind, int32_t value) {
	nox_quest_journal_native* const old_head = head;
	nox_quest_journal_native* const entry = &entries[used++];
	entry->kind = kind;
	entry->value = (uint32_t)value;
	entry->next = old_head;
	entry->prev = NULL;
	if (old_head) {
		old_head->prev = entry;
	}
	head = entry;
	return old_head;
}

nox_quest_journal_native* nox_xxx_journalQuestSet_500540(char* name, int32_t value) {
	(void)name;
	return set_new(0, value);
}

nox_quest_journal_native* nox_xxx_journalQuestSetBool_5006B0(char* name, int32_t value) {
	(void)name;
	return set_new(1, value);
}

int main(void) {
	char name[] = "War01a:Value";
	quest_journal_set_fn const set_number = nox_xxx_journalQuestSet_500540;
	quest_journal_set_fn const set_boolean = nox_xxx_journalQuestSetBool_5006B0;

	assert(set_number(name, INT32_MIN) == NULL);
	assert(head == &entries[0]);
	assert(head->kind == 0);
	assert(head->value == UINT32_C(0x80000000));

	assert(set_boolean(name, INT32_MAX) == &entries[0]);
	assert(head == &entries[1]);
	assert(head->kind == 1);
	assert(head->value == UINT32_C(0x7fffffff));
	assert(head->next == &entries[0]);
	assert(entries[0].prev == head);
	return 0;
}
