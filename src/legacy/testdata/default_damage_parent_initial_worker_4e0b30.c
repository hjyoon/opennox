// Standalone initial-field reader. No Go runtime or parent/helper algorithm.
#define _DARWIN_C_SOURCE 1
#include <inttypes.h>
#include <mach/arm/thread_status.h>
#include <setjmp.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/ucontext.h>
#include <unistd.h>
#include <zlib.h>
#include "../default_damage_parent_full_initial_4e0b30.h"
#if !defined(__APPLE__) || (!defined(__arm64__) && !defined(__aarch64__))
#error "This preserved worker is for native local Darwin ARM64"
#endif

static sigjmp_buf fault_jump;
static volatile sig_atomic_t armed, actual_signal, actual_code;
static volatile uintptr_t actual_fault_address, actual_fault_pc;
static void fault_handler(int signal_number,siginfo_t* info,void* context) {
 if(!armed)_exit(200+signal_number);
 const ucontext_t* uc=(const ucontext_t*)context;
 actual_signal=signal_number;actual_code=info?info->si_code:0;
 actual_fault_address=info?(uintptr_t)info->si_addr:0;
 actual_fault_pc=uc&&uc->uc_mcontext?(uintptr_t)arm_thread_state64_get_pc(uc->uc_mcontext->__ss):0;
 siglongjmp(fault_jump,1);
}

// One actual fixed-width load per function. Root must preserve/check the
// actual compiled disassembly before launch; each fault PC is also recorded.
__attribute__((noinline,used)) uint64_t ddi_initial_load1_4e0b30(void* p) {
 uint64_t v;__asm__ volatile("ldrb %w0,[%1]":"=r"(v):"r"(p):"memory");return v;
}
__attribute__((noinline,used)) uint64_t ddi_initial_load2_4e0b30(void* p) {
 uint64_t v;__asm__ volatile("ldrh %w0,[%1]":"=r"(v):"r"(p):"memory");return v;
}
__attribute__((noinline,used)) uint64_t ddi_initial_load4_4e0b30(void* p) {
 uint64_t v;__asm__ volatile("ldr %w0,[%1]":"=r"(v):"r"(p):"memory");return v;
}
__attribute__((noinline,used)) uint64_t ddi_initial_load8_4e0b30(void* p) {
 uint64_t v;__asm__ volatile("ldr %0,[%1]":"=r"(v):"r"(p):"memory");return v;
}
static void die(const char* message) {fprintf(stderr,"INITIAL_WORKER_ERROR %s\n",message);exit(78);}
static void input(gzFile f,void* data,size_t n) {
 if(n>(1u<<22)||gzread(f,data,(unsigned int)n)!=(int)n)die("truncated/oversized input");
}
static uint32_t in32(gzFile f) {uint32_t n;input(f,&n,4);return n;}
static uint64_t in64(gzFile f) {uint64_t n;input(f,&n,8);return n;}
static void output(const void* data,size_t n) {if(fwrite(data,1,n,stdout)!=n)die("output write");}
static void out32(uint32_t n) {output(&n,4);}
static void out64(uint64_t n) {output(&n,8);}
typedef struct {unsigned char* map;uint32_t base_offset;unsigned char* before;} allocation;

static uintptr_t checked_offset(uint32_t id,uint64_t go_offset) {
 if(id>=27)die("field id");
 uintptr_t c=ddi_c_offset_4e0b30(id);
 if(c!=UINTPTR_MAX&&c!=go_offset)die("actual named C/Go offsetof mismatch");
 return c==UINTPTR_MAX?(uintptr_t)go_offset:c;
}
static uint64_t read_value(void* address,uint32_t width) {
 switch(width) {
 case 1:return ddi_initial_load1_4e0b30(address);
 case 2:return ddi_initial_load2_4e0b30(address);
 case 4:return ddi_initial_load4_4e0b30(address);
 case 8:return ddi_initial_load8_4e0b30(address);
 default:die("native load width");return 0;
 }
}

int main(int argc,char** argv) {
 if(argc!=3||strcmp(argv[1],"--input-gzip"))die("arguments");
 gzFile plan=gzopen(argv[2],"rb");if(!plan)die("gzopen");
 char magic[8];input(plan,magic,8);if(memcmp(magic,"DDIPLN1\n",8))die("plan magic");
 uint32_t version=in32(plan),cases=in32(plan),pointer=in32(plan),page=in32(plan);
 if(version!=1||cases!=1038||pointer!=sizeof(void*)||pointer!=8||page!=(uint32_t)sysconf(_SC_PAGESIZE))die("native profile");
 struct sigaction sa;memset(&sa,0,sizeof(sa));sa.sa_sigaction=fault_handler;sa.sa_flags=SA_SIGINFO;sigemptyset(&sa.sa_mask);
 if(sigaction(SIGBUS,&sa,0)||sigaction(SIGSEGV,&sa,0))die("sigaction");
 output("DDIOUT1\n",8);out32(version);out32(cases);out32(pointer);out32(page);out32((uint32_t)getpid());out64(sizeof(nox_object_t));
 for(uint32_t i=0;i<27;i++)out64(ddi_c_offset_4e0b30(i));
 out64(ddi_callback_4e0b30(2));out64(ddi_callback_4e0b30(3));
 out64((uintptr_t)&ddi_initial_load1_4e0b30);out64((uintptr_t)&ddi_initial_load2_4e0b30);
 out64((uintptr_t)&ddi_initial_load4_4e0b30);out64((uintptr_t)&ddi_initial_load8_4e0b30);
 uint32_t total=0,all_canaries=1;
 for(uint32_t ordinal=0;ordinal<cases;ordinal++) {
  uint32_t index=in32(plan),count=in32(plan),fixups=in32(plan),probes=in32(plan);
  if(index!=ordinal||count<14||count>128||fixups>128||probes!=153)die("case bounds");
  allocation* a=calloc(count,sizeof(*a));if(!a)die("calloc");
  for(uint32_t i=0;i<count;i++) {
   uint32_t id=in32(plan),base=in32(plan),bytes=in32(plan);
   if(id!=i||base>page||bytes!=page)die("allocation bounds");
   void* p=mmap(0,2*(size_t)page,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
   if(p==MAP_FAILED||(uintptr_t)p<=UINT32_MAX)die("actual C high mmap");
   if(mprotect((unsigned char*)p+page,page,PROT_NONE))die("guard page");
   a[i].map=p;a[i].base_offset=base;a[i].before=malloc(page);if(!a[i].before)die("canary allocation");
   input(plan,p,page);
  }
  for(uint32_t i=0;i<fixups;i++) {
   uint32_t dest=in32(plan),field=in32(plan);uint64_t go_offset=in64(plan);
   uint32_t kind=in32(plan),target=in32(plan);uint64_t offset=in64(plan);
   if(dest>=count||!ddi_pointer_field_4e0b30(field))die("fixup field");
   uintptr_t native=checked_offset(field,go_offset),value=0;
   if(native>page||sizeof(value)>page-native||(uint64_t)a[dest].base_offset+native+sizeof(value)>page)die("fixup crosses guard");
   if(kind==1) {if(target>=count||offset>page)die("fixup target");value=(uintptr_t)a[target].map+offset;}
   else if(kind==2||kind==3)value=ddi_callback_4e0b30(kind);
   else if(kind!=0)die("fixup kind");
   memcpy(a[dest].map+a[dest].base_offset+native,&value,sizeof(value));
  }
  for(uint32_t i=0;i<count;i++)memcpy(a[i].before,a[i].map,page);
  out32(index);out32(count);out32(probes);
  for(uint32_t i=0;i<count;i++)out64((uintptr_t)a[i].map);
  for(uint32_t i=0;i<probes;i++) {
   uint32_t probe=in32(plan),region=in32(plan),field=in32(plan),width=in32(plan);uint64_t go_offset=in64(plan);
   if(probe!=i||region>=14||(width!=1&&width!=2&&width!=4&&width!=8))die("probe bounds");
   uintptr_t offset=checked_offset(field,go_offset);
   if(offset>page||width>page-offset)die("field outside guarded allocation span");
   if(ddi_pointer_field_4e0b30(field)&&width!=sizeof(uintptr_t))die("narrowed pointer probe");
   void* address=a[region].map+a[region].base_offset+offset;
   actual_signal=actual_code=0;actual_fault_address=actual_fault_pc=0;
   volatile uint64_t value=0;uint32_t status=0;
   if(sigsetjmp(fault_jump,1)==0) {armed=1;value=read_value(address,width);armed=0;}
   else {armed=0;status=1;}
   out32(probe);out32(status);out64(value);out32((uint32_t)actual_signal);out32((uint32_t)actual_code);
   out64(actual_fault_address);out64(actual_fault_pc);total++;
  }
  out32(count);
  for(uint32_t i=0;i<count;i++) {
   uint32_t equal=memcmp(a[i].before,a[i].map,page)==0;out32(equal);
   if(!equal) {all_canaries=0;out32(page);output(a[i].before,page);output(a[i].map,page);}
   free(a[i].before);if(munmap(a[i].map,2*(size_t)page))die("munmap");
  }
  free(a);
 }
 if(gzread(plan,magic,1)!=0||gzclose(plan)!=Z_OK)die("trailing input or CRC");
 output("DDIEND1\n",8);out32(total);out32(all_canaries);fflush(stdout);
 return total==158814&&all_canaries?0:3;
}
