#include <stdio.h>
#define _Static_assert(...)
#include "../defs.h"
#undef _Static_assert

static uint32_t pe_record[64];
static uint32_t dword_5d4594_1047520;
static uint32_t dword_5d4594_1046652;
static uint32_t dword_5d4594_1046636;
static uint32_t dword_5d4594_1046640;
static uint32_t dword_5d4594_1047524;
static uint32_t dword_5d4594_1046852;
static void* nox_xxx_aClosewoodengat_587000_133480;
static unsigned int spell_flags;
static unsigned int random_calls;
static unsigned int particle_calls;
static unsigned int insert_calls;
static unsigned int flag_calls;
static unsigned int finish_calls;
static unsigned int case_count;
static int random_values[3];
static int failure;

#define CHECK(test) do { if (!(test)) { \
	fprintf(stderr, "line %d: %s (case=%u mode=%u trail=%u spell=%u flags=%#x row=%u slot=%u)\n", \
		__LINE__, #test, case_count, pe_record[12], pe_record[13], dword_5d4594_1047524, \
		spell_flags, ((unsigned char*)nox_xxx_aClosewoodengat_587000_133480)[200], dword_5d4594_1046852); return 1; \
} } while (0)

void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base != 0x5D4594 || off < 1046628 || off >= 1046628 + sizeof(pe_record)) abort();
	return (unsigned char*)pe_record + off - 1046628;
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base, off); }

// PRODUCTION_SELECTED_ROW

static void nox_xxx_spellKeyPackSetSpell_45DC40(void* base, int spell, int slot) {
	if (base != nox_xxx_aClosewoodengat_587000_133480 || spell != (int)dword_5d4594_1047524 ||
		slot != (int)dword_5d4594_1046852 || particle_calls != 50 || random_calls != 150 || finish_calls) failure = 1;
	insert_calls++;
	unsigned char* row = nox_quickbar_selected_row(base);
	if (row && slot >= 0 && slot < 5) memcpy(row + 8 * slot, &spell, 4);
}
static bool nox_xxx_spellHasFlags_424A50(int spell, int flags) {
	if (insert_calls != 1 || finish_calls || spell != (int)dword_5d4594_1047524 || flags != 0x600) failure = 1;
	flag_calls++;
	return (spell_flags & (unsigned int)flags) != 0;
}

// The existing receiver casts its boolean result to the legacy void* return
// type. Keep that unrelated production expression unchanged in this fixture.
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wint-to-pointer-cast"
// PRODUCTION_BOOK_INSERT
#pragma clang diagnostic pop

static int nox_float2int(float value) { return (int)value; }
static int nox_common_randomIntMinMax_415FF0(int min, int max, const char* file, int line) {
	const int lines[] = {2483, 2479, 2478};
	unsigned int index = random_calls % 3;
	if (line != lines[index] || strcmp(file, "C:\\NoxPost\\src\\Client\\Gui\\guibook.c") ||
		min != (index ? 0 : 3) || max != (index ? 30 : 4) || insert_calls || finish_calls) failure = 1;
	int value = min + (int)(random_calls % (unsigned int)(max - min + 1));
	random_values[index] = value;
	random_calls++;
	return value;
}
static void nox_client_newScreenParticle_431540(int kind, int x, int y, int dx, int dy,
	int a6, int a7, int a8, int a9, int a10) {
	if (random_calls != 3 * (particle_calls + 1) ||
		kind != (dword_5d4594_1046652 == 1 ? 3 : 0) ||
		x != 100 + (int)particle_calls * 10 + random_values[2] ||
		y != 300 + (int)particle_calls * 14 - random_values[1] ||
		dx || dy || a6 != 1 || a7 != random_values[0] || a8 || a9 || a10 != 1 || insert_calls || finish_calls) failure = 1;
	particle_calls++;
}
static void sub_45D810(void) {
	if (insert_calls != 1 || particle_calls != 50 || random_calls != 150) failure = 1;
	finish_calls++;
	dword_5d4594_1047520 = 0;
}

// PRODUCTION_FINISH_BODY

static int check_finish(unsigned int mode, unsigned int trail, unsigned int flags,
	unsigned int row, int slot, unsigned int spell, unsigned int active, unsigned int kind) {
	case_count++;
	unsigned char* bar = nox_xxx_aClosewoodengat_587000_133480;
	memset(bar, 0xA5, 256);
	bar[200] = (unsigned char)row;
	memset(pe_record, 0x5A, sizeof(pe_record));
	*mem_getU32Ptr(0x5D4594, 1046668) = 600;
	*mem_getU32Ptr(0x5D4594, 1046672) = 1000;
	*mem_getU32Ptr(0x5D4594, 1046676) = mode;
	*mem_getU32Ptr(0x5D4594, 1046680) = trail;
	const float initial_x = 100, initial_y = 300;
	memcpy(&dword_5d4594_1046636, &initial_x, 4);
	memcpy(&dword_5d4594_1046640, &initial_y, 4);
	dword_5d4594_1047524 = spell;
	dword_5d4594_1046852 = (uint32_t)slot;
	dword_5d4594_1047520 = active;
	dword_5d4594_1046652 = kind;
	spell_flags = flags;
	random_calls = particle_calls = insert_calls = flag_calls = finish_calls = 0;
	failure = 0;
	unsigned char expected[256];
	memcpy(expected, bar, sizeof(expected));
	unsigned int valid = row < 5 && slot >= 0 && slot < 5;
	if (active && valid) {
		memcpy(expected + 40 * row + 8 * slot, &spell, 4);
		if (mode == 2) expected[40 * row + 8 * slot + 4] = spell && (flags & 0x600) ? 1 : 0;
	}
	uint32_t old_record[64];
	memcpy(old_record, pe_record, sizeof(old_record));
	sub_45D870();
	CHECK(!failure);
	CHECK(random_calls == (active ? 150U : 0U) && particle_calls == (active ? 50U : 0U));
	CHECK(insert_calls == (active ? 1U : 0U) && finish_calls == insert_calls);
	CHECK(flag_calls == (active && valid && mode == 2 && spell ? 1U : 0U));
	CHECK(!memcmp(bar, expected, sizeof(expected)));
	CHECK(!memcmp(pe_record, old_record, sizeof(old_record)));
	CHECK(dword_5d4594_1047520 == 0);
	CHECK(dword_5d4594_1047524 == spell && dword_5d4594_1046852 == (uint32_t)slot && dword_5d4594_1046652 == kind);
	return 0;
}
int main(void) {
	nox_xxx_aClosewoodengat_587000_133480 = calloc(1, 256);
	if (!nox_xxx_aClosewoodengat_587000_133480) return 2;
	if (sizeof(void*) == 8) CHECK((uintptr_t)nox_xxx_aClosewoodengat_587000_133480 > UINT32_MAX);
	// A self-default spell must remain self-directed after the animation writes
	// its separate trail count into the DWORD immediately following the mode.
	CHECK(check_finish(2, 20, 0x200, 0, 0, 1, 1, 0) == 0);
	const unsigned int modes[] = {2, 3, 4, 0, 0xFFFFFFFF};
	const unsigned int trails[] = {0, 1, 20, 0xFFFFFFFF};
	const unsigned int flags[] = {0, 0x200, 0x400, 0x600, 0x1000, 0x2000, 0x200400, 0x800000};
	for (unsigned int m = 0; m < sizeof(modes) / sizeof(modes[0]); m++) {
		for (unsigned int n = 0; n < sizeof(trails) / sizeof(trails[0]); n++) {
			for (unsigned int f = 0; f < sizeof(flags) / sizeof(flags[0]); f++) {
				for (unsigned int row = 0; row < 5; row++) {
					for (int slot = 0; slot < 5; slot++) {
						for (unsigned int kind = 0; kind < 3; kind++) {
							CHECK(check_finish(modes[m], trails[n], flags[f], row, slot, 136, 1, kind) == 0);
						}
					}
				}
			}
		}
	}
	for (unsigned int spell = 0; spell < 137; spell++) CHECK(check_finish(2, 20, 0x600, 4, 4, spell, 1, 0) == 0);
	const int slots[] = {-1, 0, 4, 5, 255};
	const unsigned int rows[] = {0, 4, 5, 255};
	for (unsigned int r = 0; r < sizeof(rows) / sizeof(rows[0]); r++) {
		for (unsigned int s = 0; s < sizeof(slots) / sizeof(slots[0]); s++) {
			CHECK(check_finish(2, 20, 0x600, rows[r], slots[s], 12, 1, 0) == 0);
			CHECK(check_finish(2, 20, 0x600, rows[r], slots[s], 12, 0, 0) == 0);
		}
	}
	free(nox_xxx_aClosewoodengat_587000_133480);
	printf("book reward finish: %u mode/trail/default-self/slot cases passed\n", case_count);
	return 0;
}
