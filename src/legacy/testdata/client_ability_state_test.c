#include <stdio.h>
#include <limits.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../defs.h"
#include "../GAME2_1.h"

static uint8_t* table;
static uint32_t frame;
static unsigned cases;
// GAME.EXE 00461090..0046119C walks records 1..5 at 006D4268,
// each 24 scalar bytes. ID +0, availability +8, active DWORD +12,
// reward +16 and cooldown frame +20. Record zero is not part of the scan.
void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base!=0x5D4594 || off<1047764 || off>1047928) abort();
	return table+off-1047764;
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base,off); }
int32_t* mem_getI32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base,off); }
uint32_t gameFrame(void) { return frame; }
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wsign-compare"
// COMPLETE_SUBSYSTEM
#pragma GCC diagnostic pop

static uint32_t read32(const uint8_t* p) { return (uint32_t)p[0] | (uint32_t)p[1]<<8 | (uint32_t)p[2]<<16 | (uint32_t)p[3]<<24; }
static void write32(uint8_t* p, uint32_t x) { p[0]=x; p[1]=x>>8; p[2]=x>>16; p[3]=x>>24; }
static void reference_state(uint8_t* data, uint32_t key, uint32_t state) {
	for (unsigned row=1;row<=5;row++) if (read32(data+24*row)==key) {
		write32(data+24*row+8,state); write32(data+24*row+20,state==0 ? frame : 0);
	}
}
static void reference_active(uint8_t* data, unsigned key, int active) {
	for (unsigned row=1;row<=5;row++) if (read32(data+24*row)==key) {
		uint32_t old=read32(data+24*row+12), bit=1u<<key;
		write32(data+24*row+12,active ? old|bit : old&~bit);
	}
}
static int reference_lookup(const uint8_t* data, unsigned key) {
	for (unsigned row=1;row<=5;row++) if (read32(data+24*row)==key)
		return (read32(data+24*row+12)&(1u<<key))!=0;
	return 0;
}
#define CHECK(x) do { if (!(x)) { fprintf(stderr,"case %u line %d: %s table=%p\n",cases,__LINE__,#x,(void*)table); return 1; } } while (0)
static int intact(const uint8_t* want) {
	if (memcmp(table,want,192)) return 0;
	for (int i=1;i<=32;i++) if (table[-i]!=0xa5 || table[191+i]!=0xa5) return 0;
	return 1;
}
static int run_case(const char* mode, unsigned layout, unsigned key, unsigned value, unsigned clock) {
	cases++;
	const uint32_t frames[]={0,0x7fffffffu,0x80000000u,UINT32_MAX};
	const uint32_t states[]={0,1,255,UINT32_MAX};
	frame=frames[clock];
	memset(table-32,0xa5,256); memset(table,0x5a,192);
	write32(table,0xabcdef00u); // selector zero: a missing ID, not record zero.
	for (unsigned row=1;row<=5;row++) {
		uint32_t id=layout<5 ? (row+layout-1)%5+1 : layout<10 ? layout-4 : 0;
		write32(table+24*row,id);
		write32(table+24*row+8,0x80000000u+row);
		const uint32_t masks[]={0,UINT32_MAX,1u<<id,0xffffff80u};
		write32(table+24*row+12,masks[value]);
		write32(table+24*row+16,0x12345600u+row);
		write32(table+24*row+20,0xfedcba00u+row);
	}
	uint8_t want[192]; memcpy(want,table,192);
	if (strcmp(mode,"state")==0) {
		reference_state(want,key,states[value]);
		CHECK(sub_461090((int)key,(int32_t)states[value])==(char*)table+164);
		CHECK(intact(want));
	} else if (strcmp(mode,"reset")==0) {
		if (key==6) {
			for (unsigned row=1;row<=5;row++) reference_state(want,read32(want+24*row),1);
		} else reference_state(want,read32(want+24*key),1);
		CHECK(sub_4610D0((unsigned char)key)==(char*)table+164);
		CHECK(intact(want));
	} else if (strcmp(mode,"active")==0) {
		int active=clock==0 ? 0 : clock==1 ? 1 : -1;
		reference_active(want,key,active);
		CHECK(sub_461120((int)key,active)==(char*)table+156);
		CHECK(intact(want));
		CHECK(sub_461160((int)key)==reference_lookup(want,key));
		CHECK(intact(want));
	} else if (strcmp(mode,"lookup")==0) {
		CHECK(sub_461160((int)key)==reference_lookup(want,key));
		CHECK(intact(want));
	} else abort();
	return 0;
}
static void* map_near(uintptr_t address) {
#if UINTPTR_MAX <= UINT32_MAX
	/* Keep every data/guard case on native 32-bit allocations. Exact high
	 * address and 4-GiB boundary placements remain mandatory on 64-bit. */
	address=0;
#endif
#ifdef _WIN32
	return VirtualAlloc((void*)address,0x80000,MEM_RESERVE|MEM_COMMIT,PAGE_READWRITE);
#else
	void* ptr=mmap((void*)address,0x80000,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
	return ptr==MAP_FAILED ? NULL : ptr;
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
	CHECK(argc==2);
	const uintptr_t bases[]={(uintptr_t)0x37ffc0000ull,(uintptr_t)0x3fffc0000ull,(uintptr_t)0x440000000ull};
	for (unsigned map=0;map<3;map++) {
		// Hints only; never overwrite another mapping with MAP_FIXED.
		void* ptr=map_near(bases[map]); CHECK(ptr);
		table=(uint8_t*)ptr+(map<2 ? 0x40000-72 : 64);
		if (sizeof(void*)>4) {
			CHECK((uintptr_t)ptr==bases[map] && (uintptr_t)table>UINT32_MAX);
			if (map==0) CHECK((uintptr_t)table<0x380000000ull && (uintptr_t)(table+192)>0x380000000ull);
			if (map==1) CHECK((uintptr_t)table<0x400000000ull && (uintptr_t)(table+192)>0x400000000ull);
		}
		unsigned clocks=strcmp(argv[1],"lookup")==0 ? 1 : strcmp(argv[1],"active")==0 ? 3 : 4;
		for (unsigned layout=0;layout<11;layout++) for (unsigned key=0;key<=6;key++)
			for (unsigned value=0;value<4;value++) for (unsigned clock=0;clock<clocks;clock++)
				if (run_case(argv[1],layout,key,value,clock)) return 1;
		release_map(ptr);
	}
	printf("%s: %u complete native ability-state cases; five records, original returns, duplicates, frame/mask bytes and guards checked\n",argv[1],cases);
	return 0;
}
