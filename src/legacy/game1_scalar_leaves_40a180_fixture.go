package legacy

/*
#include "GAME1.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
_Static_assert(sizeof(int)==4 && sizeof(short)==2 && sizeof(unsigned char)==1, "six scalar public ABI widths");
typedef unsigned char (*game1_leaf_byte_fn)(short);
typedef int (*game1_leaf_get_fn)(void);
typedef int (*game1_leaf_int_fn)(int);
typedef unsigned int (*game1_leaf_uint_fn)(unsigned int);
_Static_assert(_Generic(&sub_40A180,game1_leaf_byte_fn:1,default:0), "A180 uchar(short)");
_Static_assert(_Generic(&sub_40A220,game1_leaf_get_fn:1,default:0), "A220 int(void)");
_Static_assert(_Generic(&sub_40A3C0,game1_leaf_uint_fn:1,default:0), "A3C0 uint(uint)");
_Static_assert(_Generic(&sub_40A6A0,game1_leaf_int_fn:1,default:0), "A6A0 int(int)");
_Static_assert(_Generic(&sub_40A6B0,game1_leaf_get_fn:1,default:0), "A6B0 int(void)");
_Static_assert(_Generic(&sub_40A710,game1_leaf_int_fn:1,default:0), "A710 int(int)");
static uint32_t game1_leaf_call_40a180(unsigned entry,uint32_t word) {
    int raw; uint16_t low=(uint16_t)word; short small;
    memcpy(&raw,&word,4); memcpy(&small,&low,2);
    switch(entry) {
    case 0x40a180:return sub_40A180(small);
    case 0x40a220:return (uint32_t)sub_40A220();
    case 0x40a3c0:return sub_40A3C0(word);
    case 0x40a6a0:return (uint32_t)sub_40A6A0(raw);
    case 0x40a6b0:return (uint32_t)sub_40A6B0();
    case 0x40a710:return (uint32_t)sub_40A710(raw);
    default:abort();
    }
}
static void* game1_leaf_cell_40a180(uint32_t base,size_t offset) {return getMemAt(base,offset);}
*/
import "C"

import "unsafe"

func game1LeafCallC40A180(entry, word uint32) uint32 {
	return uint32(C.game1_leaf_call_40a180(C.uint(entry), C.uint32_t(word)))
}

func game1LeafCellC40A180(base, offset uint32) unsafe.Pointer {
	return C.game1_leaf_cell_40a180(C.uint32_t(base), C.size_t(offset))
}
