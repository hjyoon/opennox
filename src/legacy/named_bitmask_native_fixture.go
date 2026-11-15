package legacy

/*
#include "GAME1_1.h"
typedef const char* named_mask_pointer_fixture;
_Static_assert(_Generic(&set_one_bitmask_flag_by_name_4239C0,int (*)(char*,uint32_t*,const char**):1,default:0),"native named-bit setter ABI");
_Static_assert(_Generic(&set_bitmask_flags_from_plus_separated_names_423930,void (*)(const char*,uint32_t*,const char**):1,default:0),"const input/native names/void parser ABI");
_Static_assert(_Generic(&nox_native_parseMonsterFlags_423930,void (*)(const char*,uint32_t*):1,default:0),"explicit packed-owner adapter ABI");
static int named_mask_set_C(char* name,uint32_t* bits,const char** names) { return set_one_bitmask_flag_by_name_4239C0(name,bits,names); }
static void named_mask_parse_C(const char* input,uint32_t* bits,const char** names) { set_bitmask_flags_from_plus_separated_names_423930(input,bits,names); }
static void named_mask_monster_C(const char* input,uint32_t* bits) { nox_native_parseMonsterFlags_423930(input,bits); }
static uint16_t named_mask_monster_low16_C(const char* input,uint32_t initial) { uint32_t bits=initial; nox_native_parseMonsterFlags_423930(input,&bits); return (uint16_t)bits; }
*/
import "C"
import "unsafe"

func namedBitmaskSetC(name, bits, names unsafe.Pointer) int32 {
	return int32(C.named_mask_set_C((*C.char)(name), (*C.uint32_t)(bits), (**C.char)(names)))
}

func namedBitmaskParseC(input, bits, names unsafe.Pointer) {
	C.named_mask_parse_C((*C.char)(input), (*C.uint32_t)(bits), (**C.char)(names))
}

func namedBitmaskMonsterC(input, bits unsafe.Pointer) {
	C.named_mask_monster_C((*C.char)(input), (*C.uint32_t)(bits))
}

func namedBitmaskMonsterLow16C(input unsafe.Pointer, initial uint32) uint16 {
	return uint16(C.named_mask_monster_low16_C((*C.char)(input), C.uint32_t(initial)))
}

func namedBitmaskSizesC() (uintptr, uintptr) {
	return C.sizeof_named_mask_pointer_fixture, C.sizeof_uint32_t
}
