//go:build darwin && arm64

package legacy

/*
#include "GAME1.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stddef.h>
#include <mach/mach.h>
#include <mach/mach_vm.h>
typedef int (*setting_flags_fixture_fn_409e40)(int);
_Static_assert(sizeof(int)==4,"raw int ABI is DWORD");
_Static_assert(_Generic(&sub_409E40,setting_flags_fixture_fn_409e40:1,default:0),"E40 int(int)");
_Static_assert(_Generic(&sub_409E70,setting_flags_fixture_fn_409e40:1,default:0),"E70 int(int)");
_Static_assert(_Generic(&sub_409EC0,setting_flags_fixture_fn_409e40:1,default:0),"EC0 int(int)");
_Static_assert(_Generic(&sub_409EF0,setting_flags_fixture_fn_409e40:1,default:0),"EF0 int(int)");
_Static_assert(_Generic(&sub_409F40,setting_flags_fixture_fn_409e40:1,default:0),"F40 int(int)");
static int setting_flags_fixture_call_409e40(uint32_t entry,int word,int*known) {
 setting_flags_fixture_fn_409e40 fn=0;*known=1;
 switch(entry){case 0x409e40:fn=&sub_409E40;break;case 0x409e70:fn=&sub_409E70;break;
 case 0x409ec0:fn=&sub_409EC0;break;case 0x409ef0:fn=&sub_409EF0;break;
 case 0x409f40:fn=&sub_409F40;break;default:*known=0;return 0;}
 return fn(word);
}
typedef struct setting_flags_fixture_game_409e40 {
 uint8_t prefix[16];uint32_t value;uint8_t suffix[16];
} setting_flags_fixture_game_409e40;
_Static_assert(offsetof(setting_flags_fixture_game_409e40,value)==16,"declared private game witness field");
_Static_assert(sizeof(setting_flags_fixture_game_409e40)==36,"private game witness extent");
static void* setting_flags_fixture_game_new_409e40(void){
 setting_flags_fixture_game_409e40*p=malloc(sizeof(*p));if(!p)return 0;
 for(unsigned i=0;i<16;i++){p->prefix[i]=(uint8_t)(0xa0+i);p->suffix[i]=(uint8_t)(0xe0-i);}p->value=0;return p;
}
static void* setting_flags_fixture_game_cell_409e40(void*p){return &((setting_flags_fixture_game_409e40*)p)->value;}
// C globals are separate symbols. Their neighbours are never guessed to be a
// struct. The actual readable VM interval bounds this optional small window.
static void* setting_flags_fixture_window_409e40(void*cell,size_t*length,size_t*center){
 mach_vm_address_t region=(mach_vm_address_t)(uintptr_t)cell;mach_vm_size_t size=0;
 vm_region_basic_info_data_64_t info;mach_msg_type_number_t count=VM_REGION_BASIC_INFO_COUNT_64;
 mach_port_t object=MACH_PORT_NULL;
 kern_return_t code=mach_vm_region(mach_task_self(),&region,&size,VM_REGION_BASIC_INFO_64,(vm_region_info_t)&info,&count,&object);
 if(object!=MACH_PORT_NULL)mach_port_deallocate(mach_task_self(),object);
 uintptr_t address=(uintptr_t)cell;
 if(code!=KERN_SUCCESS||!(info.protection&VM_PROT_READ)||!(info.protection&VM_PROT_WRITE)||region>address||size<4||address-region>size-4)return 0;
 uintptr_t before=address-region<16?address-region:16;
 uintptr_t available=region+size-(address+4),after=available<16?available:16;
 *center=(size_t)before;*length=(size_t)(before+4+after);return (void*)(address-before);
}
static void setting_flags_fixture_copy_409e40(void*dst,const void*src,size_t length){memcpy(dst,src,length);}
static void setting_flags_fixture_free_409e40(void*p){free(p);}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

func settingFlagsFixtureCall409E40(entry uint32, word int32) int32 {
	var known C.int
	result := C.setting_flags_fixture_call_409e40(C.uint32_t(entry), C.int(word), &known)
	if known != 1 {
		panic("unknown five-flags typed C entry")
	}
	return int32(result)
}
func settingFlagsFixtureGameNew409E40() (allocation, cell unsafe.Pointer) {
	allocation = C.setting_flags_fixture_game_new_409e40()
	if allocation == nil {
		panic("private C game input allocation")
	}
	cell = C.setting_flags_fixture_game_cell_409e40(allocation)
	return
}
func settingFlagsFixtureFree409E40(p unsafe.Pointer) { C.setting_flags_fixture_free_409e40(p) }
func settingFlagsFixtureCopy409E40(destination unsafe.Pointer, source []byte) {
	if len(source) == 0 {
		return
	}
	C.setting_flags_fixture_copy_409e40(destination, unsafe.Pointer(&source[0]), C.size_t(len(source)))
	runtime.KeepAlive(source)
	runtime.KeepAlive(destination)
}
func settingFlagsFixtureSnapshot409E40(source unsafe.Pointer, length int) []byte {
	if length <= 0 || length > 36 {
		panic("bounded mapped C window")
	}
	out := make([]byte, length)
	C.setting_flags_fixture_copy_409e40(unsafe.Pointer(&out[0]), source, C.size_t(length))
	runtime.KeepAlive(source)
	return out
}
func settingFlagsFixtureWindow409E40(cell unsafe.Pointer) (base unsafe.Pointer, length, center int, err error) {
	var n, c C.size_t
	base = C.setting_flags_fixture_window_409e40(cell, &n, &c)
	if base == nil {
		return nil, 0, 0, fmt.Errorf("C-global readable writable VM extent")
	}
	return base, int(n), int(c), nil
}
