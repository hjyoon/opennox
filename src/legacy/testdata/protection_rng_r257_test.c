#define _GNU_SOURCE
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "native_pages.h"

static unsigned char* state;
static double constants[5];
static uint32_t dword_5d4594_2516372, dword_5d4594_2516380;
static unsigned char* getMemAt(uintptr_t base, uintptr_t offset) {
	if (base == 0x5D4594 && offset >= 2516372 && offset <= 2516428) return state + (offset - 2516372) - 16;
	if (base == 0x581450 && offset >= 11344 && offset <= 11376 && !(offset & 7)) return (unsigned char*)constants + offset - 11344;
	fprintf(stderr, "unexpected protection RNG memory %#lx+%lu\n", (unsigned long)base, (unsigned long)offset);
	exit(2);
}
#define getMemU32Ptr(base, offset) ((uint32_t*)getMemAt(base, offset))
#define getMemU64Ptr(base, offset) ((uint64_t*)getMemAt(base, offset))
#define getMemDoublePtr(base, offset) ((double*)getMemAt(base, offset))

// RNG_BODY

static uint32_t word(void) {
	unsigned char p[4];
	if (fread(p, 1, 4, stdin) != 4) { fputs("short original input\n", stderr); exit(2); }
	return (uint32_t)p[0] | (uint32_t)p[1] << 8 | (uint32_t)p[2] << 16 | (uint32_t)p[3] << 24;
}
static uint64_t wide(void) { uint64_t low = word(); return low | (uint64_t)word() << 32; }
static void read_state(uint64_t* values) { for (int i = 0; i < 5; ++i) values[i] = wide(); }
static void read_bounds(uint32_t* values) { for (int i = 0; i < 3; ++i) values[i] = word(); }
static int same_bounds(const uint32_t* want) {
	return want[0] == dword_5d4594_2516372 && want[1] == *getMemU32Ptr(0x5D4594, 2516376) && want[2] == dword_5d4594_2516380;
}
static unsigned long failures, cases;
static int wanted;
static void check(int kind, uint32_t seed, int layout, int step, int okay) {
	if (kind != wanted) return;
	++cases;
	if (!okay && failures++ < 8) fprintf(stderr, "protection kind=%d seed=%#x layout=%d state=%p step=%d differs from GAME.EXE\n", kind, seed, layout, (void*)state, step);
}
int main(int argc, char** argv) {
	if (argc != 2) return 2;
	wanted = atoi(argv[1]);
	if (wanted < 0 || wanted > 2) return 2;
	for (int i = 0; i < 5; ++i) { uint64_t bits = wide(); memcpy(&constants[i], &bits, 8); }
	uint32_t rows = word();
	unsigned char ordinary[128];
	void* reservation = NULL;
	size_t size = 0;
	uintptr_t boundary = 0, wrap = 0;
	int layouts = 1;
#if UINTPTR_MAX > UINT32_MAX
	size_t page = (size_t)native_fixture_page_size();
	size = (size_t)(UINT64_C(1) << 32) + 4 * page;
	reservation = native_fixture_reserve(size);
	if (reservation == NULL) { perror("owned mmap"); return 2; }
	uintptr_t first = (uintptr_t)reservation + 2 * page;
	boundary = (first + (UINT64_C(1) << 31) - 1) & ~((UINT64_C(1) << 31) - 1);
	if (!(boundary & (UINT64_C(1) << 31))) boundary += UINT64_C(1) << 31;
	wrap = (first + (UINT64_C(1) << 32) - 1) & ~((UINT64_C(1) << 32) - 1);
	if (boundary + page > (uintptr_t)reservation + size || wrap + page > (uintptr_t)reservation + size) return 2;
	if (native_fixture_commit((void*)(boundary - page), 2 * page) || native_fixture_commit((void*)(wrap - page), 2 * page)) { perror("owned mprotect"); return 2; }
	layouts = 4;
#endif
	for (uint32_t row = 0; row < rows; ++row) {
		uint32_t seed = word(), seed_bounds[3];
		uint64_t seeded[5]; read_state(seeded); read_bounds(seed_bounds);
		uint32_t draws = word();
		if (draws != 128) return 2;
		uint64_t draw_values[128], draw_states[128][5];
		for (uint32_t i = 0; i < draws; ++i) { draw_values[i] = wide(); read_state(draw_states[i]); }
		uint32_t ranges = word(), lows[40], highs[40], values[40], range_bounds[40][3];
		uint64_t range_states[40][5];
		if (ranges != 40) return 2;
		for (uint32_t i = 0; i < ranges; ++i) { lows[i] = word(); highs[i] = word(); values[i] = word(); read_state(range_states[i]); read_bounds(range_bounds[i]); }
		for (int layout = 0; layout < layouts; ++layout) {
			state = layout == 0 ? ordinary + 32 : layout == 1 ? (unsigned char*)(boundary - 16) : layout == 2 ? (unsigned char*)(wrap - 16) : (unsigned char*)(boundary + 16);
			memset(state - 24, 0xa5, 72);
			dword_5d4594_2516372 = dword_5d4594_2516380 = 0xa5a5a5a5;
			sub_56FF00((int)seed);
			check(0, seed, layout, -1, !memcmp(state, seeded, 40) && same_bounds(seed_bounds));
			for (uint32_t i = 0; i < draws; ++i) {
				double value = nox_xxx_unkDoubleSmth_56FE30(); uint64_t bits; memcpy(&bits, &value, 8);
				check(1, seed, layout, (int)i, bits == draw_values[i] && !memcmp(state, draw_states[i], 40));
			}
			for (uint32_t i = 0; i < ranges; ++i) {
				uint32_t value = (uint32_t)sub_56FF80((int)lows[i], (int)highs[i]);
				check(2, seed, layout, (int)i, value == values[i] && !memcmp(state, range_states[i], 40) && same_bounds(range_bounds[i]));
			}
			for (int i = 0; i < 8; ++i) if (state[-24+i] != 0xa5 || state[40+i] != 0xa5) return 2;
		}
	}
	if (fgetc(stdin) != EOF) return 2;
#if UINTPTR_MAX > UINT32_MAX
	if (native_fixture_release(reservation, size)) return 2;
#else
	(void)reservation; (void)size; (void)boundary; (void)wrap;
#endif
	printf("protection kind=%d cases=%lu failures=%lu pointer_bytes=%zu layouts=%d\n", wanted, cases, failures, sizeof(void*), layouts);
	return failures ? 1 : 0;
}
