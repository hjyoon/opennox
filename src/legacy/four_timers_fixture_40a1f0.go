package legacy

/*
#include "GAME1.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
_Static_assert(sizeof(int)==4 && sizeof(long long)==8, "timer raw C return widths");
typedef int (*timer_fixture_int_arg_40a1f0)(int);
typedef int (*timer_fixture_int_void_40a1f0)(void);
typedef long long (*timer_fixture_i64_void_40a1f0)(void);
typedef long long (*timer_fixture_i64_arg_40a1f0)(int);
_Static_assert(_Generic(&sub_40A1F0,timer_fixture_int_arg_40a1f0:1,default:0), "A1F0 int(int)");
_Static_assert(_Generic(&sub_40A230,timer_fixture_int_void_40a1f0:1,default:0), "A230 int(void)");
_Static_assert(_Generic(&sub_40A250,timer_fixture_i64_void_40a1f0:1,default:0), "A250 long long(void)");
_Static_assert(_Generic(&sub_40A310,timer_fixture_i64_arg_40a1f0:1,default:0), "A310 long long(int)");
static uint64_t timer_fixture_call_40a1f0(uint32_t entry,uint32_t word) {
    int raw;memcpy(&raw,&word,4);
    switch(entry) {
    case 0x40a1f0:return (uint32_t)sub_40A1F0(raw);
    case 0x40a230:return (uint32_t)sub_40A230();
    case 0x40a250:return (uint64_t)sub_40A250();
    case 0x40a310:return (uint64_t)sub_40A310(raw);
    default:abort();
    }
}
static void* timer_fixture_cell_40a1f0(uint32_t base,size_t offset) {return getMemAt(base,offset);}
*/
import "C"

import "unsafe"

func fourTimersCallC40A1F0(entry, argument uint32) uint64 {
	return uint64(C.timer_fixture_call_40a1f0(C.uint32_t(entry), C.uint32_t(argument)))
}

func fourTimersCellC40A1F0(base, offset uint32) unsafe.Pointer {
	return C.timer_fixture_cell_40a1f0(C.uint32_t(base), C.size_t(offset))
}
