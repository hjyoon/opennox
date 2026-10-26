#define _GNU_SOURCE
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "native_pages.h"

_Static_assert(sizeof(short) == 2 && sizeof(int) == 4, "fixed original scalar widths");
static unsigned char* table;
static unsigned char* getMemAt(uintptr_t base, uintptr_t offset) {
	if (base != 0x587000 || (offset != 87532 && offset != 87564)) {
		fputs("unexpected region table service\n", stderr);
		exit(2);
	}
	return table + offset - 87528;
}

// PRODUCTION_BODY

static uint32_t word(void) {
	unsigned char p[4];
	if (fread(p, 1, 4, stdin) != 4) {
		fputs("short immutable original input\n", stderr);
		exit(2);
	}
	return (uint32_t)p[0] | (uint32_t)p[1] << 8 | (uint32_t)p[2] << 16 | (uint32_t)p[3] << 24;
}

int main(int argc, char** argv) {
	if (argc != 2) return 2;
	int layout = atoi(argv[1]);
	static _Alignas(16) unsigned char ordinary[96];
	table = ordinary + 16;
	void* reservation = NULL;
	size_t reserve_size = 0;
#if UINTPTR_MAX > UINT32_MAX
	if (layout < 0 || layout > 6) return 2;
	if (layout) {
		/* Reserve only unused address space; never MAP_FIXED or replace
		 * a process mapping. Two owned pages per boundary become writable. */
		size_t page_size = (size_t)native_fixture_page_size();
		reserve_size = (size_t)(UINT64_C(1) << 32) + 4 * page_size;
		reservation = native_fixture_reserve(reserve_size);
		if (reservation == NULL) { perror("mmap owned test reservation"); return 2; }
		uintptr_t begin = (uintptr_t)reservation + 2 * page_size;
		uintptr_t signed_boundary = (begin + (UINT64_C(1) << 31) - 1) & ~((UINT64_C(1) << 31) - 1);
		if (!(signed_boundary & (UINT64_C(1) << 31))) signed_boundary += UINT64_C(1) << 31;
		uintptr_t wrap_boundary = (begin + (UINT64_C(1) << 32) - 1) & ~((UINT64_C(1) << 32) - 1);
		if (signed_boundary + page_size > (uintptr_t)reservation + reserve_size || wrap_boundary + page_size > (uintptr_t)reservation + reserve_size) return 2;
		if (native_fixture_commit((void*)(signed_boundary - page_size), 2 * page_size) ||
			native_fixture_commit((void*)(wrap_boundary - page_size), 2 * page_size)) {
			perror("mprotect owned test pages"); return 2;
		}
		table = layout <= 4 ? (unsigned char*)(signed_boundary - 40 + 8 * layout) :
			layout == 5 ? (unsigned char*)(wrap_boundary - 16) : (unsigned char*)(signed_boundary + 16);
	}
	if ((uintptr_t)table <= UINT32_MAX) { fputs("missing high native table address\n", stderr); return 2; }
#else
	if (layout) return 2;
	(void)reservation; (void)reserve_size;
#endif
	uint32_t rows = word();
	unsigned long cases = 0, failures = 0;
	for (uint32_t row = 0; row < rows; ++row) {
		int16_t values[16];
		for (unsigned int i = 0; i < 16; ++i) values[i] = (int16_t)word();
		memset(table - 8, 0xa5, 48);
		memcpy(table, values, 32);
		unsigned char before[48];
		memcpy(before, table - 8, sizeof(before));
		uint32_t count = word();
		for (uint32_t pair = 0; pair < count; ++pair) {
			int32_t x = (int32_t)word(), y = (int32_t)word(), expected = (int32_t)word();
			int32_t got = sub_437860(x, y);
			++cases;
			if (got != expected || memcmp(before, table - 8, sizeof(before))) {
				if (failures++ < 8) fprintf(stderr, "row=%u layout=%d table=%p x=%d y=%d got=%d GAME.EXE=%d\n", row, layout, (void*)table, x, y, got, expected);
			}
		}
	}
	if (fgetc(stdin) != EOF) return 2;
	printf("region: cases=%lu failures=%lu native_pointer_bytes=%zu layout=%d table=%p\n", cases, failures, sizeof(void*), layout, (void*)table);
#if UINTPTR_MAX > UINT32_MAX
	if (layout && native_fixture_release(reservation, reserve_size)) { perror("munmap owned reservation"); return 2; }
#endif
	return failures ? 1 : 0;
}
