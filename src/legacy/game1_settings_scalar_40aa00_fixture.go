package legacy

/*
#include "GAME1.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
extern uint32_t dword_5d4594_10984;
extern unsigned char byte_5D4594[2598284];
_Static_assert(sizeof(int)==4 && sizeof(uint32_t)==4, "settings scalar DWORD ABI");
typedef int (*game1_settings_get_fn)(void);
typedef int (*game1_settings_set_fn)(int);
typedef void (*game1_settings_void_fn)(void);
_Static_assert(_Generic(&sub_40AA00,game1_settings_get_fn:1,default:0), "AA00 int(void)");
_Static_assert(_Generic(&sub_40AA20,game1_settings_get_fn:1,default:0), "AA20 int(void)");
_Static_assert(_Generic(&sub_40AA30,game1_settings_set_fn:1,default:0), "AA30 int(int)");
_Static_assert(_Generic(&sub_40AA40,game1_settings_get_fn:1,default:0), "AA40 int(void)");
_Static_assert(_Generic(&sub_40AA60,game1_settings_set_fn:1,default:0), "AA60 int(int)");
_Static_assert(_Generic(&sub_40E090,game1_settings_void_fn:1,default:0), "E090 void(void)");
static uint32_t game1_settings_call_40aa00(uint32_t entry,uint32_t word) {
    int raw; memcpy(&raw,&word,4);
    switch(entry) {
    case 0x40aa00:return (uint32_t)sub_40AA00();
    case 0x40aa20:return (uint32_t)sub_40AA20();
    case 0x40aa30:return (uint32_t)sub_40AA30(raw);
    case 0x40aa40:return (uint32_t)sub_40AA40();
    case 0x40aa60:return (uint32_t)sub_40AA60(raw);
    case 0x40e090:sub_40E090();return 0; // transport placeholder, no void return claim
    default:abort();
    }
}
static void* game1_settings_cell_40aa00(size_t offset) { return getMemAt(0x5d4594,offset); }
static void* game1_settings_raw_blob_40aa00(size_t offset,size_t width) {
    if(offset>sizeof(byte_5D4594)||width>sizeof(byte_5D4594)-offset) abort();
    return byte_5D4594+offset;
}
static void* game1_settings_global_40aa00(void) { return &dword_5d4594_10984; }
static size_t game1_settings_global_width_40aa00(void) { return sizeof(dword_5d4594_10984); }
static size_t game1_settings_global_align_40aa00(void) { return _Alignof(uint32_t); }
*/
import "C"

import "unsafe"

func game1SettingsCallC40AA00(entry, word uint32) uint32 {
	return uint32(C.game1_settings_call_40aa00(C.uint32_t(entry), C.uint32_t(word)))
}

func game1SettingsCellC40AA00(offset uint32) unsafe.Pointer {
	return C.game1_settings_cell_40aa00(C.size_t(offset))
}

func game1SettingsBlobC40AA00(offset, width int) []byte {
	return unsafe.Slice((*byte)(C.game1_settings_raw_blob_40aa00(C.size_t(offset), C.size_t(width))), width)
}

func game1SettingsGlobalC40AA00() (unsafe.Pointer, uintptr, uintptr) {
	return C.game1_settings_global_40aa00(), uintptr(C.game1_settings_global_width_40aa00()), uintptr(C.game1_settings_global_align_40aa00())
}
