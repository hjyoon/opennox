#define _GNU_SOURCE
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

static uint32_t kind;
static unsigned char* table;
static unsigned char* getMemAt(uintptr_t base, uintptr_t offset) {
	uintptr_t first = kind == 0 ? 4704 : kind == 1 ? 4664 : 191836;
	uintptr_t end = kind == 0 ? 4728 : kind == 1 ? 4704 : 191884;
	if (base != 0x587000 || offset < first || offset > end) {
		fprintf(stderr, "unexpected table service %#lx+%lu\n", (unsigned long)base, (unsigned long)offset);
		exit(2);
	}
	return table + offset - first;
}
#define getMemU32Ptr(base, offset) ((uint32_t*)getMemAt(base, offset))

// SETTINGS_BODY
// RATE_BODY
// MODE_BODY

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
	const char* names[] = {"settings", "rate", "mode"};
	uint32_t wanted = 0;
	while (wanted < 3 && strcmp(names[wanted], argv[1])) ++wanted;
	if (wanted == 3) return 2;
	uint32_t rows = word();
	unsigned long cases = 0, failures = 0;
	unsigned char ordinary[128];
	void* reservation = NULL;
	size_t reserve_size = 0, page_size = 0;
	uintptr_t signed_boundary = 0, wrap_boundary = 0;
	int layouts = 1;
#if UINTPTR_MAX > UINT32_MAX
	/* Reserve only previously free virtual space. Never MAP_FIXED or replace
	 * any process mapping. Only two owned pages per boundary are committed. */
	page_size = (size_t)sysconf(_SC_PAGESIZE);
	reserve_size = (size_t)(UINT64_C(1) << 32) + 4 * page_size;
	reservation = mmap(NULL, reserve_size, PROT_NONE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
	if (reservation == MAP_FAILED) { perror("mmap owned test reservation"); return 2; }
	uintptr_t begin = (uintptr_t)reservation + 2 * page_size;
	signed_boundary = (begin + (UINT64_C(1) << 31) - 1) & ~((UINT64_C(1) << 31) - 1);
	if (!(signed_boundary & (UINT64_C(1) << 31))) signed_boundary += UINT64_C(1) << 31;
	wrap_boundary = (begin + (UINT64_C(1) << 32) - 1) & ~((UINT64_C(1) << 32) - 1);
	if (signed_boundary + page_size > (uintptr_t)reservation + reserve_size || wrap_boundary + page_size > (uintptr_t)reservation + reserve_size) return 2;
	if (mprotect((void*)(signed_boundary - page_size), 2 * page_size, PROT_READ | PROT_WRITE) ||
		mprotect((void*)(wrap_boundary - page_size), 2 * page_size, PROT_READ | PROT_WRITE)) {
		perror("mprotect owned test pages"); return 2;
	}
	layouts = 4;
#else
	(void)reservation; (void)reserve_size; (void)page_size; (void)signed_boundary; (void)wrap_boundary;
#endif
	for (uint32_t row = 0; row < rows; ++row) {
		kind = word();
		uint32_t count = word(), values[12];
		if (kind > 2 || count != (kind == 0 ? 6u : kind == 1 ? 10u : 12u)) return 2;
		for (uint32_t i = 0; i < count; ++i) values[i] = word();
		uint32_t pairs = word();
		for (uint32_t pair = 0; pair < pairs; ++pair) {
			uint32_t input = word(), expected = word();
			if (kind != wanted) continue;
			for (int layout = 0; layout < layouts; ++layout) {
				table = layout == 0 ? ordinary + 16 : layout == 1 ? (unsigned char*)(signed_boundary - 16) : layout == 2 ? (unsigned char*)(wrap_boundary - 16) : (unsigned char*)(signed_boundary + 16);
				memset(table - 8, 0xa5, count * sizeof(uint32_t) + 16);
				memcpy(table, values, count * sizeof(uint32_t));
				unsigned char before[64];
				memcpy(before, table - 8, count * sizeof(uint32_t) + 16);
				uint32_t got = kind == 0 ? (uint32_t)sub_409A70((short)input) : kind == 1 ? (uint32_t)sub_40A710((int)input) : (uint32_t)sub_4D0D50((int)input);
				++cases;
				if (got != expected || memcmp(before, table - 8, count * sizeof(uint32_t) + 16)) {
					if (failures++ < 8) fprintf(stderr, "%s row=%u layout=%d table=%p input=%#x got=%#x GAME.EXE=%#x\n", names[kind], row, layout, (void*)table, input, got, expected);
				}
			}
		}
	}
	if (fgetc(stdin) != EOF) return 2;
#if UINTPTR_MAX > UINT32_MAX
	if (munmap(reservation, reserve_size)) { perror("munmap owned reservation"); return 2; }
#endif
	printf("%s: cases=%lu failures=%lu native_pointer_bytes=%zu layouts=%d\n", names[wanted], cases, failures, sizeof(void*), layouts);
	return failures ? 1 : 0;
}
