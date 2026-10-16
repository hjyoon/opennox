#define _GNU_SOURCE
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

static unsigned char *sine, *arc;
static _Alignas(8) unsigned char constants[32];
static unsigned sine_calls, arc_calls;
static unsigned char* getMemAt(uintptr_t base, uintptr_t offset) {
	if (base == 0x85B3FC && offset >= 12260 && offset <= 28644) return sine + (offset - 12260);
	if (base == 0x5D4594 && offset >= 338472 && offset <= 371240) return arc + (offset - 338472);
	if (base == 0x581450 && offset >= 7184 && offset <= 7211) return constants + (offset - 7184);
	fprintf(stderr,"unexpected trig memory %#lx+%lu\n",(unsigned long)base,(unsigned long)offset); exit(2);
}
#define getMemByte(base,offset) (*getMemAt(base,offset))
#define getMemU8Ptr(base,offset) getMemAt(base,offset)
#define getMemU32Ptr(base,offset) ((uint32_t*)getMemAt(base,offset))
#define getMemDoublePtr(base,offset) ((double*)getMemAt(base,offset))
#define getMemFloatPtr(base,offset) ((float*)getMemAt(base,offset))
#define LOBYTE(value) (*((uint8_t*)&(value)))
static double measured_sin(double value) { ++sine_calls; return sin(value); }
static double measured_acos(double value) { ++arc_calls; return acos(value); }
#define sin measured_sin
#define acos measured_acos

// TRIG_BODY

static void exact_read(void* data, size_t n) { if (fread(data,1,n,stdin)!=n) { fputs("short trig oracle input\n",stderr);exit(2); } }
static uint32_t word(void) { uint8_t b[4];exact_read(b,4);return (uint32_t)b[0]|(uint32_t)b[1]<<8|(uint32_t)b[2]<<16|(uint32_t)b[3]<<24; }
static unsigned long cases,failures;
static void check(int mode,int layout,int pattern,int ready,int okay) {
	++cases;
	if (!okay && failures++<12) fprintf(stderr,"trig mode=%d layout=%d sine=%p arc=%p pattern=%#x ready=%#x calls=%u/%u differs\n",mode,layout,(void*)sine,(void*)arc,pattern,ready,sine_calls,arc_calls);
}
int main(int argc,char** argv) {
	if (argc!=2) return 2;
	int mode=atoi(argv[1]);if(mode<0||mode>1)return 2;
	exact_read(constants,28);
	unsigned char table[49152];exact_read(table,sizeof(table));
	uint32_t rows=word();if(rows!=15)return 2;
	unsigned char inputs[15][4];exact_read(inputs,sizeof(inputs));if(fgetc(stdin)!=EOF)return 2;
	_Alignas(8) unsigned char ordinary_sine[16384+32],ordinary_arc[32769+32];
	void* reservation=NULL;size_t size=0;uintptr_t boundary=0,wrap=0;int layouts=1;
#if UINTPTR_MAX > UINT32_MAX
	long page_size=sysconf(_SC_PAGESIZE);if(page_size<=0)return 2;
	size_t page=(size_t)page_size;
	size_t tail=((sizeof(ordinary_arc)+page-1)/page)*page;
	size=(size_t)(UINT64_C(1)<<32)+16*page+2*tail;
	reservation=mmap(NULL,size,PROT_NONE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
	if(reservation==MAP_FAILED){perror("owned trig mmap");return 2;}
	uintptr_t first=(uintptr_t)reservation+4*page;
	boundary=(first+(UINT64_C(1)<<31)-1)&~((UINT64_C(1)<<31)-1);
	if(!(boundary&(UINT64_C(1)<<31)))boundary+=UINT64_C(1)<<31;
	wrap=(first+(UINT64_C(1)<<32)-1)&~((UINT64_C(1)<<32)-1);
	if(boundary+tail>(uintptr_t)reservation+size||wrap+tail>(uintptr_t)reservation+size)return 2;
	if(mprotect((void*)(boundary-page),page+tail,PROT_READ|PROT_WRITE)||mprotect((void*)(wrap-page),page+tail,PROT_READ|PROT_WRITE)){perror("owned trig mprotect");return 2;}
	layouts=6;
#endif
	for(int layout=0;layout<layouts;++layout){
		sine=ordinary_sine+16;arc=ordinary_arc+16;
		if(layout==1)sine=(unsigned char*)(boundary-64);
		if(layout==2)arc=(unsigned char*)(boundary-64);
		if(layout==3)sine=(unsigned char*)(wrap-64);
		if(layout==4)arc=(unsigned char*)(wrap-64);
		if(layout==5)sine=(unsigned char*)(boundary+64);
		for(uint32_t row=0;row<rows;++row){
			int pattern=inputs[row][0],ready=inputs[row][1];
			memset(sine-16,pattern,16384+32);memset(arc-16,pattern,32769+32);
			arc[32768]=(unsigned char)ready;sine_calls=arc_calls=0;
			unsigned char returned=(unsigned char)nox_xxx_initSinCosTables_414C90();
			int okay=returned==inputs[row][2]&&arc[32768]==inputs[row][3];
			okay=okay&&sine_calls==(ready?0u:4096u)&&arc_calls==(ready?0u:8192u);
			for(int i=0;i<16;++i)okay=okay&&sine[-16+i]==pattern&&sine[16384+i]==pattern&&arc[-16+i]==pattern&&arc[32769+i]==pattern;
			if(ready){for(int i=0;i<16384;++i)okay=okay&&sine[i]==pattern;for(int i=0;i<32768;++i)okay=okay&&arc[i]==pattern;}
			if(mode==1&&!ready){
				unsigned sin_bad=0,arc_bad=0;
				for(int i=0;i<4096;++i)if(memcmp(sine+4*i,table+4*i,4))++sin_bad;
				for(int i=0;i<8192;++i)if(memcmp(arc+4*i,table+16384+4*i,4))++arc_bad;
				if((sin_bad||arc_bad)&&failures<12)fprintf(stderr,"trig numeric words layout=%d mismatches sin=%u acos=%u\n",layout,sin_bad,arc_bad);
				okay=okay&&!sin_bad&&!arc_bad;
			}
			check(mode,layout,pattern,ready,okay);
		}
	}
#if UINTPTR_MAX > UINT32_MAX
	if(munmap(reservation,size))return 2;
#else
	(void)reservation;(void)size;(void)boundary;(void)wrap;
#endif
	printf("trig mode=%d cases=%lu failures=%lu pointer_bytes=%zu layouts=%d\n",mode,cases,failures,sizeof(void*),layouts);
	return failures?1:0;
}
