#include <stdio.h>
#include <limits.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../defs.h"
#include "../GAME2_3.h"
#if defined(__APPLE__) && UINTPTR_MAX > UINT32_MAX
#include "native_pages.h"
#endif

static uint8_t* table;
static unsigned cases;
// GAME.EXE 00494FF0..004951F0 uses 32 sixteen-byte scalar records at
// 006F98A8..006F9AA8. Active DWORD is +12, alternate BYTE +4 and HP words
// +6/+8; duplicate insertion requires active==1, but lookup accepts any
// nonzero active DWORD. Preserve padding +5/+10/+11 and full 32-bit IDs.
void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base!=0x5D4594 || off<1200916 || off>1201440) abort();
	return table+off-1200916;
}
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wsign-compare"
// COMPLETE_SUBSYSTEM
#pragma GCC diagnostic pop

static uint16_t read16(const uint8_t* p) { return (uint16_t)p[0] | (uint16_t)p[1]<<8; }
static uint32_t read32(const uint8_t* p) { return (uint32_t)p[0] | (uint32_t)p[1]<<8 | (uint32_t)p[2]<<16 | (uint32_t)p[3]<<24; }
static void write16(uint8_t* p, uint16_t x) { p[0]=x; p[1]=x>>8; }
static void write32(uint8_t* p, uint32_t x) { p[0]=x; p[1]=x>>8; p[2]=x>>16; p[3]=x>>24; }
static int reference_find(const uint8_t* data, uint32_t key) {
	for (int i=0;i<32;i++) if (read32(data+16*i+12)!=0 && read32(data+16*i)==key) return i;
	return -1;
}
static int reference_free(const uint8_t* data) {
	for (int i=0;i<32;i++) if (read32(data+16*i+12)==0) return i;
	return -1;
}
static int reference_add(uint8_t* data, uint32_t key, uint16_t cur, uint16_t max) {
	for (int i=0;i<32;i++) if (read32(data+16*i+12)==1 && read32(data+16*i)==key) return 1;
	int slot=reference_free(data); if (slot<0) return 0;
	uint8_t* row=data+16*slot;
	write32(row,key); row[4]=0; write16(row+6,cur); write16(row+8,max); write32(row+12,1);
	return 1;
}
static uint32_t key_for(int i, int status) {
	return (status==0 ? 0x200u : status==1 ? 0x80000200u : 0xffffffe0u)+(uint32_t)i;
}
#define CHECK(x) do { if (!(x)) { fprintf(stderr,"case %u line %d: %s table=%p\n",cases,__LINE__,#x,(void*)table); return 1; } } while (0)
static int intact(const uint8_t* want) {
	if (memcmp(table,want,512)) return 0;
	for (int i=1;i<=32;i++) if (table[-i]!=0xa5 || table[511+i]!=0xa5) return 0;
	return 1;
}
static int run_case(const char* mode, int hole, int status, int target) {
	cases++;
	memset(table-32,0xa5,576); memset(table,0x5a,512);
	const uint32_t active[]={1,2,UINT32_MAX};
	for (int i=0;i<32;i++) {
		uint8_t* row=table+16*i;
		write32(row,key_for(i,status)); row[4]=(uint8_t)(0x80+i);
		write16(row+6,(uint16_t)(0x8000+i)); write16(row+8,(uint16_t)(UINT16_MAX-i));
		write32(row+12,i==hole ? 0 : active[status]);
	}
	uint8_t want[512]; memcpy(want,table,512);
	uint32_t key=target==32 ? 0x12345678u : key_for(target,status);
	const uint16_t values[]={0,0x8001,0xffff};
	uint16_t cur=values[status], max=values[(status+1)%3];
	int found=reference_find(want,key);
	if (strcmp(mode,"free")==0) {
		int free_slot=reference_free(want);
		char* got=sub_494FF0();
		CHECK(got==(free_slot<0 ? NULL : (char*)table+16*free_slot));
		CHECK(intact(want));
	} else if (strcmp(mode,"lookup")==0) {
		CHECK(sub_495020((int32_t)key)==(found<0 ? NULL : (char*)table+16*found));
		CHECK(nox_xxx_unitSpriteCheckAlly_4951F0((int32_t)key)==(found>=0));
		CHECK(intact(want));
		uint16_t actual_cur=0xbeef, actual_max=0xcafe; uint8_t alt=0x7d;
		CHECK(sub_495180((int32_t)key,&actual_cur,&actual_max,&alt)==(found>=0));
		CHECK(actual_cur==(found<0 ? 0xbeef : read16(want+16*found+6)));
		CHECK(actual_max==(found<0 ? 0xcafe : read16(want+16*found+8)));
		CHECK(alt==(found<0 ? 0x7d : want[16*found+4]));
		CHECK(intact(want));
		CHECK(sub_4950F0((int32_t)key,(char)0xfe)==(found>=0));
		if (found>=0) want[16*found+4]=0xfe;
		CHECK(intact(want));
		CHECK(sub_495120((int32_t)key,(short)cur,(short)max)==(found>=0));
		if (found>=0) { write16(want+16*found+6,cur); write16(want+16*found+8,max); }
		CHECK(intact(want));
		CHECK(sub_495150((int32_t)key,(short)0xffff)==(found>=0));
		if (found>=0) write16(want+16*found+6,0xffff);
		CHECK(intact(want));
		CHECK(sub_4950C0((int32_t)key)==(found>=0));
		if (found>=0) write32(want+16*found+12,0);
		CHECK(intact(want));
		actual_cur=0xbeef; actual_max=0xcafe; alt=0x7d;
		CHECK(sub_495180((int32_t)key,&actual_cur,&actual_max,&alt)==0);
		CHECK(actual_cur==0xbeef && actual_max==0xcafe && alt==0x7d && intact(want));
	} else if (strcmp(mode,"insert")==0) {
		int added=reference_add(want,key,cur,max);
		CHECK(sub_495060((int32_t)key,(short)cur,(short)max)==added);
		CHECK(intact(want));
		// Re-insertion must retain the original value fields, even when a
		// preceding active=2 duplicate exists before the active=1 record.
		CHECK(sub_495060((int32_t)key,(short)~cur,(short)~max)==added);
		CHECK(intact(want));
	} else if (strcmp(mode,"reset")==0) {
		for (int i=0;i<32;i++) {
			uint8_t* row=want+16*i;
			write32(row,0); row[4]=0; write16(row+6,0); write16(row+8,0); write32(row+12,0);
		}
		CHECK(sub_4951C0()==(char*)table+518);
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
#elif defined(__APPLE__) && UINTPTR_MAX > UINT32_MAX
	/* main owns the full reservation; commit only this table's pages. */
	return native_fixture_commit((void*)address,0x80000) ? NULL : (void*)address;
#else
	void* ptr=mmap((void*)address,0x80000,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
	return ptr==MAP_FAILED ? NULL : ptr;
#endif
}
static void release_map(void* ptr) {
#ifdef _WIN32
	VirtualFree(ptr,0,MEM_RELEASE);
#elif defined(__APPLE__) && UINTPTR_MAX > UINT32_MAX
	(void)ptr; /* Release the complete owned reservation after all cases. */
#else
	munmap(ptr,0x80000);
#endif
}
int main(int argc, char** argv) {
	CHECK(argc==2);
#if defined(__APPLE__) && UINTPTR_MAX > UINT32_MAX
	/* ASLR may occupy an absolute hint. Own the range while preserving the
	 * original low DWORDs and both signed-bit and 4-GiB table crossings. */
	size_t reserve_size=(size_t)(UINT64_C(8)<<30)+0x100000;
	void* reservation=native_fixture_reserve(reserve_size); CHECK(reservation);
	uintptr_t first=(uintptr_t)reservation+0x80000;
	uintptr_t sign=(first+(UINT64_C(1)<<31)-1)&~((UINT64_C(1)<<31)-1);
	if (!(sign&(UINT64_C(1)<<31))) sign+=UINT64_C(1)<<31;
	uintptr_t wrap=(first+(UINT64_C(1)<<32)-1)&~((UINT64_C(1)<<32)-1);
	uintptr_t bases[]={sign-0x40000,wrap-0x40000,wrap+0x40000000};
	for (unsigned i=0;i<3;i++) {
		CHECK(bases[i]>UINT32_MAX && bases[i]>=(uintptr_t)reservation);
		CHECK(bases[i]+0x80000<=(uintptr_t)reservation+reserve_size);
	}
	CHECK((uint32_t)bases[0]==UINT32_C(0x7ffc0000));
	CHECK((uint32_t)bases[1]==UINT32_C(0xfffc0000));
	CHECK((uint32_t)bases[2]==UINT32_C(0x40000000));
#else
	const uintptr_t bases[]={(uintptr_t)0x37ffc0000ull,(uintptr_t)0x3fffc0000ull,(uintptr_t)0x440000000ull};
#endif
	for (unsigned map=0;map<3;map++) {
		// A hint only: do not overwrite any existing mapping with MAP_FIXED.
		void* ptr=map_near(bases[map]); CHECK(ptr);
		table=(uint8_t*)ptr+(map<2 ? 0x40000-256 : 64);
		if (sizeof(void*)>4) {
			CHECK((uintptr_t)ptr==bases[map] && (uintptr_t)table>UINT32_MAX);
#if defined(__APPLE__) && UINTPTR_MAX > UINT32_MAX
			if (map==0) CHECK((uintptr_t)table<sign && (uintptr_t)(table+512)>sign);
			if (map==1) CHECK((uintptr_t)table<wrap && (uintptr_t)(table+512)>wrap);
#else
			if (map==0) CHECK((uintptr_t)table<0x380000000ull && (uintptr_t)(table+512)>0x380000000ull);
			if (map==1) CHECK((uintptr_t)table<0x400000000ull && (uintptr_t)(table+512)>0x400000000ull);
#endif
		}
		for (int hole=0;hole<=32;hole++) for (int status=0;status<3;status++) for (int target=0;target<=32;target++)
			if (run_case(argv[1],hole,status,target)) return 1;
		release_map(ptr);
	}
#if defined(__APPLE__) && UINTPTR_MAX > UINT32_MAX
	CHECK(native_fixture_release(reservation,reserve_size)==0);
#endif
	printf("%s: %u complete native ally-health cases; 32 slots, full IDs, lifecycle, padding and guards checked\n",argv[1],cases);
	return 0;
}
