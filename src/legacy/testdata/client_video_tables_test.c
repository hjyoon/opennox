#include <stdio.h>
#include <stdarg.h>
#include <math.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../defs.h"
#include "../GAME1_2.h"

static uint8_t* table;
static char names[36][32];
static uint32_t tokens[36];
static int32_t trig[320];
static double radians, degrees, scale;
uint64_t qword_581450_9552;
static unsigned calls, formats, cases, policy;
static int compass, order;
static void fail(const char* reason) {
	fprintf(stderr,"case=%u calls=%u table=%p: %s\n",cases,calls,(void*)table,reason);
	exit(1);
}
static void* video_mem(uint32_t blob, uint32_t off) {
	if (blob!=0x5d4594) fail("wrong table blob");
	unsigned first=compass ? 1309516 : 1309840;
	unsigned last=compass ? 1309668 : 1311120;
	if (off<first || off>last) fail("unexpected table access");
	return table+off-first;
}
static double* video_double(uint32_t blob, uint32_t off) {
	if (blob!=0x581450) fail("wrong constant blob");
	if (off==9768) return &degrees;
	if (off==9760) return &radians;
	fail("unexpected trig constant");
	return NULL;
}
int nox_sprintf(char* dst, const char* fmt, ...) {
	if (formats>=36) fail("too many compass formats");
	const char* want=formats<4 ? "Compass%d" : "CompassMainArrow%d";
	va_list ap;
	va_start(ap,fmt);
	int index=va_arg(ap,int);
	va_end(ap);
	if (strcmp(fmt,want) || index!=(int)(formats<4 ? formats+1 : formats-3)) fail("original format/index sequence");
	int result=snprintf(dst,64,fmt,index);
	if (strcmp(dst,names[formats])) fail("original image name differs");
	formats++;
	return result;
}
nox_video_bag_image_t* nox_xxx_gLoadImg_42F970(char* name) {
	if (calls>=36 || strcmp(name,names[calls]) || formats!=calls+1) fail("original load order/name");
	uint32_t token=policy==2 || (policy==1 && (calls&1)) ? 0 : tokens[calls];
	calls++;
	return (nox_video_bag_image_t*)(uintptr_t)token;
}
static double video_sin(double angle) {
	if (calls>=320) fail("too many sine calls");
	double expected=(double)(calls+192)*degrees*radians;
	if (memcmp(&angle,&expected,sizeof(angle))) fail("unchanged double angle expression");
	double result=order ? ((double)((int)calls-160)/scale) : sin(angle);
	calls++;
	return result;
}
#undef getMemAt
#undef getMemU32Ptr
#undef getMemDoublePtr
#define getMemAt video_mem
#define getMemU32Ptr(B,O) ((uint32_t*)video_mem(B,O))
#define getMemDoublePtr video_double
#define sin video_sin
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wint-conversion"
// COMPASS_BODY
// TRIG_BODY
#pragma GCC diagnostic pop
#undef sin

static void* map_near(uintptr_t address) {
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
	compass=strcmp(argv[1],"compass")==0;
	if (!compass && strcmp(argv[1],"trig")) return 1;
	if (fread(names,sizeof(names),1,stdin)!=1 || fread(tokens,sizeof(tokens),1,stdin)!=1 ||
	    fread(trig,sizeof(trig),1,stdin)!=1 || fread(&degrees,8,1,stdin)!=1 ||
	    fread(&radians,8,1,stdin)!=1 || fread(&scale,8,1,stdin)!=1) return 1;
	memcpy(&qword_581450_9552,&scale,8);
	const uintptr_t bases[]={(uintptr_t)0x37ffc0000ull,(uintptr_t)0x3fffc0000ull,(uintptr_t)0x440000000ull};
	const unsigned offsets[]={16,128,144,320,1024};
	const uint8_t patterns[]={0,0x5a,0xff};
	unsigned bytes=compass ? 152 : 1280;
	for (unsigned mapping=0;mapping<3;mapping++) {
		void* ptr=map_near(bases[mapping]);
		if (!ptr || (sizeof(void*)>4 && (uintptr_t)ptr!=bases[mapping])) fail("exact native placement unavailable");
		for (unsigned placement=0;placement<5;placement++) for (unsigned pattern=0;pattern<3;pattern++)
		for (policy=0;policy<(compass ? 3u : 2u);policy++) {
			table=(uint8_t*)ptr+(mapping<2 ? 0x40000-offsets[placement] : 64+offsets[placement]);
			memset(table-32,patterns[pattern],bytes+64);
			calls=formats=0; cases++; order=policy!=0;
			long long result=compass ? nox_xxx_compassGenStrings_4A9C80() : sub_4AEE30();
			if (calls!=(compass ? 36u : 320u)) fail("original fixed loop count incomplete");
			if (result!=(compass ? 1 : order ? 159 : trig[319])) fail("original return/last store");
			if (compass) {
				if (formats!=36) fail("compass format count");
				for (unsigned i=0;i<36;i++) {
					unsigned slot=i<4 ? i+32 : i-4;
					uint32_t want=policy==2 || (policy==1 && (i&1)) ? 0 : tokens[i];
					if (((uint32_t*)table)[slot]!=want) fail("original compass word stores");
				}
				for (unsigned i=144;i<148;i++) if (table[i]!=patterns[pattern]) fail("unrelated text slot changed");
				if (((uint32_t*)table)[37]!=0) fail("original state reset");
			} else for (unsigned i=0;i<320;i++)
				if (((int32_t*)table)[i]!=(order ? (int)i-160 : trig[i])) fail("original trig words or ordered stores");
			for (unsigned i=0;i<32;i++) if (table[-(int)i-1]!=patterns[pattern] || table[bytes+i]!=patterns[pattern]) fail("table guards changed");
		}
		release_map(ptr);
	}
	printf("%s: %u complete native cases; original loop counts, ordered full table, returns, unrelated slots and guards verified\n",argv[1],cases);
	return 0;
}
