package legacy

/*
#include "GAME1.h"
#include "defs.h"
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
typedef int (*sp_int0_4134d0)(void);
typedef void (*sp_void0_4134d0)(void);
typedef char* (*sp_string0_4134d0)(void);
typedef void (*sp_void1_4134d0)(int);
typedef int (*sp_compare_4134d0)(const void*,const void*);
_Static_assert(sizeof(int)==4,"startup/pause DWORD ABI");
_Static_assert(_Generic(&sub_4134D0,sp_int0_4134d0:1,default:0),"startup ABI");
_Static_assert(_Generic(&sub_4137E0,sp_void0_4134d0:1,default:0),"music ABI");
_Static_assert(_Generic(&sub_413890,sp_string0_4134d0:1,default:0),"path ABI");
_Static_assert(_Generic(&sub_4138E0,sp_void1_4134d0:1,default:0),"gain callback ABI");
_Static_assert(_Generic(&sub_413900,sp_void1_4134d0:1,default:0),"loss callback ABI");
_Static_assert(_Generic(&sub_413920,sp_int0_4134d0:1,default:0),"register ABI");
_Static_assert(_Generic(&sub_4139B0,sp_int0_4134d0:1,default:0),"query ABI");
_Static_assert(_Generic(&sub_413F60,sp_compare_4134d0:1,default:0),"const qsort ABI");
static int32_t sp_fixture_call_4134d0(uint32_t entry,int32_t arg,void* a,void* b) {
    switch(entry) {
    case 0x4134d0:return sub_4134D0();
    case 0x4137e0:sub_4137E0();return 0;
    case 0x4138e0:sub_4138E0(arg);return 0;
    case 0x413900:sub_413900(arg);return 0;
    case 0x413920:return sub_413920();
    case 0x4139b0:return sub_4139B0();
    case 0x413f60:return sub_413F60(a,b);
    default:abort();
    }
}
static char* sp_fixture_drive_4134d0(void) { return sub_413890(); }
static void sp_fixture_qsort_4134d0(void* base,size_t count) { qsort(base,count,84,sub_413F60); }
static void sp_fixture_callback_4134d0(uint32_t kind,int32_t arg) {
    void (*fn)(int)=kind==1?sub_413900:sub_4138E0;fn(arg);
}
extern obj_5D4594_754088_t* ptr_5D4594_754088;
extern obj_5D4594_754088_t* ptr_5D4594_754092;
extern int ptr_5D4594_754088_cnt,ptr_5D4594_754092_cnt;
static struct {obj_5D4594_754088_t *a,*b;int na,nb;} sp_focus_saved_4134d0;
static void sp_fixture_focus_begin_4134d0(void) {
    sp_focus_saved_4134d0.a=ptr_5D4594_754088;sp_focus_saved_4134d0.b=ptr_5D4594_754092;
    sp_focus_saved_4134d0.na=ptr_5D4594_754088_cnt;sp_focus_saved_4134d0.nb=ptr_5D4594_754092_cnt;
    ptr_5D4594_754088=0;ptr_5D4594_754092=0;ptr_5D4594_754088_cnt=0;ptr_5D4594_754092_cnt=0;
}
static int sp_fixture_focus_count_4134d0(int kind) {return kind==1?ptr_5D4594_754088_cnt:ptr_5D4594_754092_cnt;}
static int sp_fixture_focus_record_4134d0(int kind,int index) {
    obj_5D4594_754088_t* p=kind==1?ptr_5D4594_754088:ptr_5D4594_754092;
    return p[index].fnc==(kind==1?sub_413900:sub_4138E0) && p[index].field_4==0 && strcmp(p[index].name,"Pause")==0;
}
static void sp_fixture_focus_end_4134d0(void) {
    free(ptr_5D4594_754088);free(ptr_5D4594_754092);
    ptr_5D4594_754088=sp_focus_saved_4134d0.a;ptr_5D4594_754092=sp_focus_saved_4134d0.b;
    ptr_5D4594_754088_cnt=sp_focus_saved_4134d0.na;ptr_5D4594_754092_cnt=sp_focus_saved_4134d0.nb;
}
static void* sp_fixture_focus_owner_4134d0(int kind) {return kind==1?(void*)ptr_5D4594_754088:(void*)ptr_5D4594_754092;}
static void sp_fixture_focus_invoke_4134d0(int kind,int index,int arg) {
    obj_5D4594_754088_t* p=kind==1?ptr_5D4594_754088:ptr_5D4594_754092;p[index].fnc(arg);
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func startupPauseFixtureCall4134D0(entry uint32, arg int32, a, b unsafe.Pointer) int32 {
	v := int32(C.sp_fixture_call_4134d0(C.uint32_t(entry), C.int32_t(arg), a, b))
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	return v
}
func startupPauseFixtureDrive4134D0() unsafe.Pointer {
	return unsafe.Pointer(C.sp_fixture_drive_4134d0())
}
func startupPauseFixtureQsort4134D0(base unsafe.Pointer, count int) {
	C.sp_fixture_qsort_4134d0(base, C.size_t(count))
	runtime.KeepAlive(base)
}
func startupPauseFixtureCallback4134D0(kind uint32, arg int32) {
	C.sp_fixture_callback_4134d0(C.uint32_t(kind), C.int32_t(arg))
}
func startupPauseFixtureFocusCount4134D0(kind int) int {
	return int(C.sp_fixture_focus_count_4134d0(C.int(kind)))
}
func startupPauseFixtureFocusRecord4134D0(kind, index int) bool {
	return C.sp_fixture_focus_record_4134d0(C.int(kind), C.int(index)) != 0
}
func startupPauseFixtureFocusBegin4134D0() { C.sp_fixture_focus_begin_4134d0() }
func startupPauseFixtureFocusEnd4134D0()   { C.sp_fixture_focus_end_4134d0() }
func startupPauseFixtureFocusOwner4134D0(kind int) unsafe.Pointer {
	return C.sp_fixture_focus_owner_4134d0(C.int(kind))
}
func startupPauseFixtureFocusInvoke4134D0(kind, index int, arg int32) {
	C.sp_fixture_focus_invoke_4134d0(C.int(kind), C.int(index), C.int(arg))
}
