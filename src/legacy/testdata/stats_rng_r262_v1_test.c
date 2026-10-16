#define _GNU_SOURCE
#include <fenv.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

enum { STATE_SIZE=308 };
static unsigned char* state;
static _Alignas(8) unsigned char coefficient[8];
static unsigned char* getMemAt(uintptr_t base,uintptr_t offset) {
 if(base==0x5D4594 && offset>=741368 && offset<741368+STATE_SIZE)return state+(offset-741368);
 if(base==0x581450 && offset==8368)return coefficient;
 fprintf(stderr,"unexpected stats RNG memory %#lx+%lu\n",(unsigned long)base,(unsigned long)offset);exit(2);
}
#define getMemU32Ptr(base,offset) ((uint32_t*)getMemAt(base,offset))
#define getMemDoublePtr(base,offset) ((double*)getMemAt(base,offset))
#define dword_5d4594_741648 (*(int32_t*)getMemAt(0x5D4594,741648))
#define dword_5d4594_741652 (*(int32_t*)getMemAt(0x5D4594,741652))

// STATS_RNG_BODY

static void exact_read(void* p,size_t n) { if(fread(p,1,n,stdin)!=n){fputs("short stats RNG input\n",stderr);exit(2);} }
static uint32_t word(void) { uint8_t p[4];exact_read(p,4);return (uint32_t)p[0]|(uint32_t)p[1]<<8|(uint32_t)p[2]<<16|(uint32_t)p[3]<<24; }
static void exact_write(const void* p,size_t n) { if(fwrite(p,1,n,stdout)!=n)exit(2); }
int main(int argc,char** argv) {
 if(argc!=2)return 2;
 int layout=atoi(argv[1]);if(layout<0||layout>5)return 2;
 _Alignas(8) unsigned char ordinary[STATE_SIZE];
 _Alignas(8) unsigned char seed_guard[36];
 void* allocation=NULL;size_t size=0;
 state=ordinary;
#if UINTPTR_MAX > UINT32_MAX
 long page_size=sysconf(_SC_PAGESIZE);if(page_size<=0)return 2;
 size_t page=(size_t)page_size;
 size_t tail=((STATE_SIZE+page-1)/page)*page;
 size=(size_t)(UINT64_C(1)<<32)+16*page+2*tail;
 allocation=mmap(NULL,size,PROT_NONE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
 if(allocation==MAP_FAILED){perror("owned stats RNG mmap");return 2;}
 uintptr_t first=(uintptr_t)allocation+4*page;
 uintptr_t boundary=(first+(UINT64_C(1)<<31)-1)&~((UINT64_C(1)<<31)-1);
 if(!(boundary&(UINT64_C(1)<<31)))boundary+=UINT64_C(1)<<31;
 uintptr_t wrap=(first+(UINT64_C(1)<<32)-1)&~((UINT64_C(1)<<32)-1);
 if(boundary+tail>(uintptr_t)allocation+size||wrap+tail>(uintptr_t)allocation+size)return 2;
 if(mprotect((void*)(boundary-page),page+tail,PROT_READ|PROT_WRITE)||mprotect((void*)(wrap-page),page+tail,PROT_READ|PROT_WRITE)){perror("owned stats RNG mprotect");return 2;}
 if(layout==1)state=(unsigned char*)(boundary-84);
 if(layout==2)state=(unsigned char*)(wrap-84);
 if(layout==3)state=(unsigned char*)(boundary+64);
 if(layout==4)state=(unsigned char*)(wrap+64);
 if(layout==5){state=malloc(STATE_SIZE);if(!state)return 2;}
 if((uintptr_t)state<=UINT32_MAX || (uintptr_t)(seed_guard+16)<=UINT32_MAX)return 2;
#else
 if(layout!=0)return 2;
#endif
 int rounding=fegetround();if(rounding!=FE_TONEAREST)return 2;
 exact_read(coefficient,8);
 uint32_t rows=word();if(rows!=162)return 2;
 unsigned long draws=0;
 for(uint32_t row=0;row<rows;++row) {
  uint32_t pattern=word(),seed=word(),count=word();if(pattern>255||count!=128)return 2;
  exact_read(state,STATE_SIZE);
  memset(seed_guard,(int)pattern,sizeof(seed_guard));memcpy(seed_guard+16,&seed,4);
  for(uint32_t draw=0;draw<count;++draw) {
   double value=sub_42AAA0((int*)(seed_guard+16));
   for(int i=0;i<16;++i)if(seed_guard[i]!=pattern||seed_guard[i+20]!=pattern)return 2;
   exact_write(&value,8);exact_write(seed_guard+16,4);exact_write(state,STATE_SIZE);
   ++draws;
  }
 }
 if(fgetc(stdin)!=EOF || fegetround()!=rounding)return 2;
#if UINTPTR_MAX > UINT32_MAX
 if(layout==5)free(state);
 if(munmap(allocation,size))return 2;
#else
 (void)allocation;(void)size;
#endif
 fprintf(stderr,"stats RNG layout=%d rows=%u draws=%lu pointer_bytes=%zu rounding_unchanged=true\n",layout,rows,draws,sizeof(void*));
 return 0;
}
