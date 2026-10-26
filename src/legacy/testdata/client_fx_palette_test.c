#include <stdio.h>
#include <limits.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../defs.h"
#include "../client__video__draw_common.h"
#include "../operators.h"

static uint8_t* table;
static uint32_t expected_calls[256][4], expected_words[256], expected_return;
static unsigned calls, cases;
static int check_colors;
static void fail(const char* reason) {
	fprintf(stderr,"case %u call %u table=%p: %s\n",cases,calls,(void*)table,reason);
	exit(1);
}
static void* palette_mem(uint32_t blob, uint32_t off) {
	if (blob!=0x5d4594 || off<1312500 || off>1313524) fail("unexpected palette memmap access");
	return table+off-1312500;
}
uint32_t nox_color_rgb_4344A0(int r, int g, int b) {
	if (calls>=256) fail("too many RGB calls");
	if (check_colors && ((uint8_t)r!=expected_calls[calls][0] || (uint8_t)g!=expected_calls[calls][1] || (uint8_t)b!=expected_calls[calls][2]))
		fail("RGB bytes differ from unchanged original execution");
	uint32_t value=check_colors ? expected_calls[calls][3] : calls;
	calls++;
	return value;
}
#undef getMemAt
#define getMemAt palette_mem
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wuninitialized"
// PALETTE_BODY
#pragma GCC diagnostic pop

static void* map_near(uintptr_t address) {
#if UINTPTR_MAX <= UINT32_MAX
	/* Keep every data/guard case on native 32-bit allocations. Exact high
	 * address and 4-GiB boundary placements remain mandatory on 64-bit. */
	address=0;
#endif
#ifdef _WIN32
	return VirtualAlloc((void*)address,0x80000,MEM_RESERVE|MEM_COMMIT,PAGE_READWRITE);
#else
	void* p=mmap((void*)address,0x80000,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
	return p==MAP_FAILED ? NULL : p;
#endif
}
static void release_map(void* ptr) {
#ifdef _WIN32
	VirtualFree(ptr,0,MEM_RELEASE);
#else
	munmap(ptr,0x80000);
#endif
}
int main(int argc, char** argv) {
	if (argc!=2) return 1;
	check_colors=strcmp(argv[1],"colors")==0;
	if (!check_colors && strcmp(argv[1],"bounds")) return 1;
	if (fread(expected_calls,sizeof(expected_calls),1,stdin)!=1 || fread(expected_words,sizeof(expected_words),1,stdin)!=1 || fread(&expected_return,sizeof(expected_return),1,stdin)!=1) return 1;
	const uintptr_t bases[]={(uintptr_t)0x37ffc0000ull,(uintptr_t)0x3fffc0000ull,(uintptr_t)0x440000000ull};
	const unsigned offsets[]={16,272,400,528,784};
	const uint8_t patterns[]={0,0x5a,0xff};
	for (unsigned mapping=0;mapping<3;mapping++) {
		void* ptr=map_near(bases[mapping]);
		if (!ptr || (sizeof(void*)>4 && (uintptr_t)ptr!=bases[mapping])) fail("native high-address placement unavailable");
		for (unsigned placement=0;placement<5;placement++) for (unsigned pattern=0;pattern<3;pattern++) {
			table=(uint8_t*)ptr+(mapping<2 ? 0x40000-offsets[placement] : 64+offsets[placement]);
			if (sizeof(void*)>4 && (uintptr_t)table<=UINT32_MAX) fail("native address required");
			memset(table-32,patterns[pattern],1024+64);
			calls=0; cases++;
			uint32_t returned=(uint32_t)sub_4B5CD0();
			if (calls!=256) fail("all four original loops must issue 256 RGB calls");
			if (returned!=(check_colors ? expected_return : 255)) fail("original final RGB return");
			for (unsigned i=0;i<256;i++) {
				uint32_t want=check_colors ? expected_words[i] : expected_words[i]&255;
				if (((uint32_t*)table)[i]!=want) fail("whole 256-word palette/order differs from original");
			}
			for (unsigned i=0;i<32;i++) if (table[-(int)i-1]!=patterns[pattern] || table[1024+i]!=patterns[pattern]) fail("palette guard changed");
		}
		release_map(ptr);
	}
	printf("%s: %u complete native palette cases; original 256 RGB calls/stores/return, four loop boundaries and guards verified\n",argv[1],cases);
	return 0;
}
