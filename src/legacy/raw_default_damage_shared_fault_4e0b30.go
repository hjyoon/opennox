//go:build darwin && arm64

package legacy

/*
#include <stdint.h>
#include <stddef.h>
#include <signal.h>
#include <pthread.h>
#include <string.h>
#include <unistd.h>
#include <sys/mman.h>
#include <sys/ucontext.h>
#include <mach/arm/thread_status.h>
typedef struct rdd_shared_fault_4e0b30 {
 uint64_t magic,version,status,signal_number,signal_code,address,pc;
 uint64_t expected_operand,expected_width,site,operation,loader,callback;
 uint64_t progress,expected_pc,expected_sp,actual_sp,branch_site,actual_lr,expected_lr;
 uint64_t owner_thread,actual_thread,armed,signal_matches_operation,unqualified_reason;
 uint64_t caller_address,caller_entry_bits,caller_sample_bits,caller_sample_valid,caller_scope_active,caller_scope_thread,caller_samples;
 uint64_t caller_observed_bits,caller_observed_progress,caller_observed_site,caller_observed_count;
 uint64_t validation_failed,validation_position,root_dispatched,root_entry,root_backend;
} rdd_shared_fault_4e0b30;
_Static_assert(sizeof(rdd_shared_fault_4e0b30)==41*8,"shared signal record extent");
_Static_assert(offsetof(rdd_shared_fault_4e0b30,expected_pc)==112,"STRH expected PC offset");
_Static_assert(offsetof(rdd_shared_fault_4e0b30,expected_sp)==120,"STRH expected SP offset");
_Static_assert(offsetof(rdd_shared_fault_4e0b30,branch_site)==136,"BLR site offset");
_Static_assert(offsetof(rdd_shared_fault_4e0b30,expected_lr)==152,"BLR LR offset");
static volatile rdd_shared_fault_4e0b30* rdd_fault_active_4e0b30;
static struct sigaction rdd_old_bus_4e0b30,rdd_old_segv_4e0b30;
static int32_t* volatile rdd_caller_cell_4e0b30;
static volatile rdd_shared_fault_4e0b30* rdd_caller_header_4e0b30;
// Match Go1.27.3 runtime/tls_arm64.h + asm_arm64.s exactly. No TLS accessor,
// pthread call, lazy allocation or library lock is touched by the handler.
__attribute__((always_inline)) static inline uint64_t rdd_thread_cookie_4e0b30(void){
 uint64_t value;__asm__ volatile("mrs %0,tpidrro_el0":"=r"(value));return value&~UINT64_C(7);
}
static int rdd_hardware_code_4e0b30(int number,int code){
 return number==SIGSEGV?(code==SEGV_MAPERR||code==SEGV_ACCERR):
        number==SIGBUS?(code==BUS_ADRALN||code==BUS_ADRERR||code==BUS_OBJERR):0;
}
// A dedicated child terminates after capture. There is no Go callback,
// longjmp, heap allocation, formatting, target access or mmap in this handler.
static void rdd_fault_signal_4e0b30(int number,siginfo_t*info,void*context){
 volatile rdd_shared_fault_4e0b30*r=rdd_fault_active_4e0b30;
 const ucontext_t*u=(const ucontext_t*)context;
 uint64_t thread=rdd_thread_cookie_4e0b30();int exit_code=171;
 if(r){
  r->signal_number=(uint64_t)number;r->signal_code=(uint64_t)(uint32_t)(info?info->si_code:0);
  r->address=(uint64_t)(uintptr_t)(info?info->si_addr:0);
  r->pc=(u&&u->uc_mcontext)?(uint64_t)arm_thread_state64_get_pc(u->uc_mcontext->__ss):0;
  r->actual_sp=(u&&u->uc_mcontext)?(uint64_t)arm_thread_state64_get_sp(u->uc_mcontext->__ss):0;
  r->actual_lr=(u&&u->uc_mcontext)?(uint64_t)arm_thread_state64_get_lr(u->uc_mcontext->__ss):0;
  r->actual_thread=thread;r->status=4;r->unqualified_reason=1;
  if(thread&&thread==r->owner_thread){
   r->unqualified_reason=2;
   if(info&&rdd_hardware_code_4e0b30(number,info->si_code)){
    int operand_ok=r->operation==2?(r->address==r->callback&&r->pc==r->callback):
       ((r->operation==1||r->operation==3)&&r->expected_width>0&&r->expected_width<=8&&
        r->address>=r->expected_operand&&r->address-r->expected_operand<r->expected_width);
    r->unqualified_reason=3;
    if(operand_ok){
     // This is the actual Go scalar, explicitly pinned by the active caller.
     // No arbitrary graph memory is sampled from the signal handler.
     if(rdd_caller_header_4e0b30==r&&rdd_caller_cell_4e0b30&&r->caller_scope_active&&r->caller_scope_thread==thread){
      r->caller_sample_bits=(uint32_t)*(volatile int32_t*)rdd_caller_cell_4e0b30;r->caller_sample_valid=1;r->caller_samples++;
     }
     // Generic native progress has not supplied an independently qualified PC
     // yet. Retain raw state, but never label it as a qualified primitive fault.
     if(!r->armed){r->status=5;r->unqualified_reason=0;exit_code=172;}
     else{
      int pc_ok=r->pc==r->expected_pc&&r->actual_sp==r->expected_sp;
      if(r->operation==2)pc_ok=pc_ok&&r->actual_lr==r->expected_lr;
      r->unqualified_reason=4;
      if(pc_ok){r->status=3;r->signal_matches_operation=1;r->unqualified_reason=0;exit_code=170;}
     }
    }
   }
  }
 }
 static const char marker[]="RAW_DEFAULT_DAMAGE_SHARED_SIGNAL_CAPTURE\n";
 (void)write(STDERR_FILENO,marker,sizeof(marker)-1);_exit(exit_code);
}
static void* rdd_shared_map_4e0b30(int fd,uint64_t offset,size_t page){
 if(!page||offset%page)return 0;
 void*reserved=mmap(0,2*page,PROT_NONE,MAP_PRIVATE|MAP_ANON,-1,0);if(reserved==MAP_FAILED)return 0;
 void*p=mmap(reserved,page,PROT_READ|PROT_WRITE,MAP_SHARED|MAP_FIXED,fd,(off_t)offset);
 if(p==MAP_FAILED){munmap(reserved,2*page);return 0;}return p;
}
static int rdd_shared_sync_4e0b30(void*p,size_t page){return msync(p,page,MS_SYNC);}
static void rdd_shared_unmap_4e0b30(void*p,size_t page){munmap(p,2*page);}
static int rdd_shared_begin_4e0b30(void*header){
 if(rdd_fault_active_4e0b30)return -1;
 struct sigaction action;memset(&action,0,sizeof(action));action.sa_sigaction=rdd_fault_signal_4e0b30;action.sa_flags=SA_SIGINFO|SA_ONSTACK;
 sigemptyset(&action.sa_mask);sigaddset(&action.sa_mask,SIGBUS);sigaddset(&action.sa_mask,SIGSEGV);
 volatile rdd_shared_fault_4e0b30*r=header;r->owner_thread=rdd_thread_cookie_4e0b30();if(!r->owner_thread)return -1;
 rdd_fault_active_4e0b30=r;
 if(sigaction(SIGBUS,&action,&rdd_old_bus_4e0b30)!=0){rdd_fault_active_4e0b30=0;return -1;}
 if(sigaction(SIGSEGV,&action,&rdd_old_segv_4e0b30)!=0){sigaction(SIGBUS,&rdd_old_bus_4e0b30,0);rdd_fault_active_4e0b30=0;return -1;}
 r->status=1;return 0;
}
static int rdd_shared_end_4e0b30(void){
 volatile rdd_shared_fault_4e0b30*r=rdd_fault_active_4e0b30;
 if(!r||r->owner_thread!=rdd_thread_cookie_4e0b30())return -1;
 int uncleared=rdd_caller_cell_4e0b30!=0||rdd_caller_header_4e0b30!=0;
 rdd_caller_cell_4e0b30=0;rdd_caller_header_4e0b30=0;r->caller_scope_active=0;
 int a=sigaction(SIGBUS,&rdd_old_bus_4e0b30,0),b=sigaction(SIGSEGV,&rdd_old_segv_4e0b30,0);
 r->armed=0;r->status=2;rdd_fault_active_4e0b30=0;return a||b||uncleared?-1:0;
}
static void rdd_shared_root_dispatched_4e0b30(void*header,uint64_t entry,uint64_t backend){volatile rdd_shared_fault_4e0b30*r=header;r->root_entry=entry;r->root_backend=backend;r->root_dispatched=1;}
static void rdd_shared_validation_failed_4e0b30(void*header,uint64_t position){volatile rdd_shared_fault_4e0b30*r=header;r->validation_failed=1;r->validation_position=position;}
static int rdd_shared_caller_begin_4e0b30(void*header,int32_t*actual){
 volatile rdd_shared_fault_4e0b30*r=header;
 if(!actual||rdd_fault_active_4e0b30!=r||rdd_caller_cell_4e0b30||r->owner_thread!=rdd_thread_cookie_4e0b30())return -1;
 r->caller_address=(uintptr_t)actual;r->caller_entry_bits=(uint32_t)*actual;r->caller_sample_valid=0;
 r->caller_observed_bits=(uint32_t)*actual;r->caller_observed_site=0x4e0b30;r->caller_observed_count=1;
 r->caller_scope_thread=rdd_thread_cookie_4e0b30();r->caller_scope_active=1;rdd_caller_header_4e0b30=r;rdd_caller_cell_4e0b30=actual;return 0;
}
static int rdd_shared_caller_observe_4e0b30(void*header,uint64_t progress,uint32_t site,int32_t*actual,uint32_t value){
 volatile rdd_shared_fault_4e0b30*r=header;
 if(rdd_caller_header_4e0b30!=r||rdd_caller_cell_4e0b30!=actual||r->caller_scope_thread!=rdd_thread_cookie_4e0b30()||(uint32_t)*actual!=value)return -1;
 r->caller_observed_bits=value;r->caller_observed_progress=progress;r->caller_observed_site=site;r->caller_observed_count++;return 0;
}
static int rdd_shared_caller_end_4e0b30(void*header){
 volatile rdd_shared_fault_4e0b30*r=header;
 if(rdd_caller_header_4e0b30!=r||!rdd_caller_cell_4e0b30||r->caller_scope_thread!=rdd_thread_cookie_4e0b30())return -1;
 r->caller_sample_bits=(uint32_t)*rdd_caller_cell_4e0b30;r->caller_sample_valid=2;r->caller_samples++;
 rdd_caller_cell_4e0b30=0;rdd_caller_header_4e0b30=0;r->caller_scope_active=0;return 0;
}
static void rdd_shared_progress_4e0b30(void*header,uint64_t progress,uint32_t site,void*address,uint32_t width,uint32_t operation,void*callback){
 volatile rdd_shared_fault_4e0b30*r=header;r->armed=0;r->signal_matches_operation=0;
 r->progress=progress;r->site=site;r->expected_operand=(uintptr_t)address;r->expected_width=width;r->operation=operation;r->callback=(uintptr_t)callback;
 r->expected_pc=0;r->expected_sp=0;r->branch_site=0;r->expected_lr=0;
}
__attribute__((noinline)) static void rdd_shared_strh_4e0b30(void*header,void*address,uint32_t value){
 __asm__ volatile("adr x9,1f\nstr x9,[%[header],#112]\nmov x10,sp\nstr x10,[%[header],#120]\n1: strh %w[value],[%[address]]"
   ::[header]"r"(header),[address]"r"(address),[value]"r"(value):"x9","x10","memory");
}
__attribute__((noinline)) static void rdd_shared_invoke_4e0b30(void*header,uintptr_t callback){
 __asm__ volatile("adr x9,1f\nstr x9,[%[header],#136]\nadd x10,x9,#4\nstr x10,[%[header],#152]\nmov x10,sp\nstr x10,[%[header],#120]\n1: blr %[callback]"
   ::[header]"r"(header),[callback]"r"(callback):"x9","x10","x30","memory");
}
static void rdd_shared_caller_mutate_then_fault_4e0b30(void*header,int32_t*actual,uint32_t value){
 volatile rdd_shared_fault_4e0b30*r=header;
 // Deliberately no returned/after observer can execute after this assignment.
 *actual=(int32_t)value;r->armed=1;r->loader=(uintptr_t)&rdd_shared_invoke_4e0b30;r->expected_pc=r->callback;
 rdd_shared_invoke_4e0b30(header,r->callback);
}
static void rdd_shared_control_4e0b30(void*header,void*address,uint32_t value,int mode){
 volatile rdd_shared_fault_4e0b30*r=header;r->armed=1;
 if(mode==2){r->armed=0;(void)raise(SIGBUS);return;}
 if(mode==1){r->loader=(uintptr_t)&rdd_shared_invoke_4e0b30;r->expected_pc=r->callback;rdd_shared_invoke_4e0b30(header,r->callback);}
 else{r->loader=(uintptr_t)&rdd_shared_strh_4e0b30;rdd_shared_strh_4e0b30(header,address,value);}
 r->armed=0;
}
typedef struct rdd_other_thread_input_4e0b30 {void*header;} rdd_other_thread_input_4e0b30;
static void* rdd_other_thread_fault_4e0b30(void*raw){
 rdd_other_thread_input_4e0b30*input=raw;rdd_shared_control_4e0b30(input->header,0,0,1);return 0;
}
static int rdd_shared_other_thread_control_4e0b30(void*header){
 pthread_t thread;rdd_other_thread_input_4e0b30 input={header};
 if(pthread_create(&thread,0,rdd_other_thread_fault_4e0b30,&input))return -1;
 return pthread_join(thread,0)?-1:0;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

const rawDefaultDamageSharedMagic4E0B30 = uint64(0x3146444452574152)
const rawDefaultDamageSharedVersion4E0B30 = uint64(2)
const rawDefaultDamageSharedHeaderWords4E0B30 = 41
const rawDefaultDamageSharedWrite4E0B30 = 1
const rawDefaultDamageSharedCallback4E0B30 = 2
const rawDefaultDamageSharedRead4E0B30 = 3

func rawDefaultDamageSharedMap4E0B30(fd int, offset uint64, page int) (unsafe.Pointer, error) {
	p := C.rdd_shared_map_4e0b30(C.int(fd), C.uint64_t(offset), C.size_t(page))
	if p == nil {
		return nil, fmt.Errorf("shared native guarded mmap")
	}
	return p, nil
}
func rawDefaultDamageSharedUnmap4E0B30(p unsafe.Pointer, page int) {
	C.rdd_shared_unmap_4e0b30(p, C.size_t(page))
}
func rawDefaultDamageSharedSync4E0B30(p unsafe.Pointer, page int) error {
	if C.rdd_shared_sync_4e0b30(p, C.size_t(page)) != 0 {
		return fmt.Errorf("shared AFTER msync")
	}
	return nil
}
func rawDefaultDamageSharedBegin4E0B30(header unsafe.Pointer) error {
	runtime.LockOSThread()
	if C.rdd_shared_begin_4e0b30(header) != 0 {
		runtime.UnlockOSThread()
		return fmt.Errorf("dedicated child signal handler install")
	}
	return nil
}
func rawDefaultDamageSharedEnd4E0B30() error {
	defer runtime.UnlockOSThread()
	if C.rdd_shared_end_4e0b30() != 0 {
		return fmt.Errorf("restore prior dedicated child signal handlers")
	}
	return nil
}
func rawDefaultDamageSharedProgress4E0B30(header unsafe.Pointer, position uint64, site uint32, address unsafe.Pointer, width, operation int, callback unsafe.Pointer) {
	C.rdd_shared_progress_4e0b30(header, C.uint64_t(position), C.uint32_t(site), address, C.uint32_t(width), C.uint32_t(operation), callback)
	runtime.KeepAlive(address)
	runtime.KeepAlive(callback)
	runtime.KeepAlive(header)
}
func rawDefaultDamageSharedOperand4E0B30(header unsafe.Pointer, site uint32, address unsafe.Pointer, width, operation int) {
	rawDefaultDamageSharedProgress4E0B30(header, 0, site, address, width, operation, nil)
}
func rawDefaultDamageSharedControl4E0B30(header, address unsafe.Pointer, value uint32, callback bool) {
	mode := 0
	if callback {
		mode = 1
	}
	C.rdd_shared_control_4e0b30(header, address, C.uint32_t(value), C.int(mode))
	runtime.KeepAlive(address)
	runtime.KeepAlive(header)
}
func rawDefaultDamageSharedUnrelatedSignalControl4E0B30(header unsafe.Pointer) {
	C.rdd_shared_control_4e0b30(header, nil, 0, 2)
}

// The caller must Pin(actual), retain its Pinner/lifetime through the real root
// call, clear this typed C pointer here, and only then Unpin. No uintptr bypass.
func rawDefaultDamageSharedCallerBegin4E0B30(header unsafe.Pointer, actual *int32) error {
	if actual == nil {
		return fmt.Errorf("actual pinned caller scalar required")
	}
	if C.rdd_shared_caller_begin_4e0b30(header, (*C.int32_t)(unsafe.Pointer(actual))) != 0 {
		return fmt.Errorf("actual caller scalar scope begin")
	}
	runtime.KeepAlive(actual)
	runtime.KeepAlive(header)
	return nil
}
func rawDefaultDamageSharedCallerEnd4E0B30(header unsafe.Pointer) error {
	if C.rdd_shared_caller_end_4e0b30(header) != 0 {
		return fmt.Errorf("actual caller scalar scope end before Unpin")
	}
	runtime.KeepAlive(header)
	return nil
}
func rawDefaultDamageSharedCallerMutationControl4E0B30(header unsafe.Pointer, actual *int32, value uint32) {
	C.rdd_shared_caller_mutate_then_fault_4e0b30(header, (*C.int32_t)(unsafe.Pointer(actual)), C.uint32_t(value))
	runtime.KeepAlive(actual)
	runtime.KeepAlive(header)
}

func rawDefaultDamageSharedCallerObserve4E0B30(header unsafe.Pointer, position uint64, site uint32, actual *int32, value uint32) error {
	if C.rdd_shared_caller_observe_4e0b30(header, C.uint64_t(position), C.uint32_t(site), (*C.int32_t)(unsafe.Pointer(actual)), C.uint32_t(value)) != 0 {
		return fmt.Errorf("actual observed caller scalar identity/value")
	}
	runtime.KeepAlive(actual)
	runtime.KeepAlive(header)
	return nil
}
func rawDefaultDamageSharedOtherThreadControl4E0B30(header unsafe.Pointer) error {
	if C.rdd_shared_other_thread_control_4e0b30(header) != 0 {
		return fmt.Errorf("C-owned other-thread control")
	}
	runtime.KeepAlive(header)
	return nil
}

func rawDefaultDamageSharedValidationFailed4E0B30(header unsafe.Pointer, position uint64) {
	C.rdd_shared_validation_failed_4e0b30(header, C.uint64_t(position))
	runtime.KeepAlive(header)
}

func rawDefaultDamageSharedRootDispatched4E0B30(header unsafe.Pointer, entry uintptr, backend uint64) {
	C.rdd_shared_root_dispatched_4e0b30(header, C.uint64_t(entry), C.uint64_t(backend))
	runtime.KeepAlive(header)
}
