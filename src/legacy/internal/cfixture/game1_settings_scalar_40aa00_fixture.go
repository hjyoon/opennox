package cfixture

/*
#include "world_native_storage.h"
#define dword_5d4594_10984 (nox_world_word_storage()->settings_scalar)
#include "GAME1.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
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
    size_t capacity;
    uint8_t* blob = nox_legacy_memory_blob(0x5D4594, &capacity);
    if (!blob || offset > capacity || width > capacity - offset) abort();
    return blob + offset;
}
static void* game1_settings_global_40aa00(void) { return &dword_5d4594_10984; }
static size_t game1_settings_global_width_40aa00(void) { return sizeof(dword_5d4594_10984); }
static size_t game1_settings_global_align_40aa00(void) { return _Alignof(uint32_t); }
*/
import "C"

import "unsafe"

func Game1SettingsCallC40AA00(entry, word uint32) uint32 {
	return uint32(C.game1_settings_call_40aa00(C.uint32_t(entry), C.uint32_t(word)))
}

func Game1SettingsCellC40AA00(offset uint32) unsafe.Pointer {
	return C.game1_settings_cell_40aa00(C.size_t(offset))
}

func Game1SettingsBlobC40AA00(offset, width int) []byte {
	return unsafe.Slice((*byte)(C.game1_settings_raw_blob_40aa00(C.size_t(offset), C.size_t(width))), width)
}

func Game1SettingsGlobalC40AA00() (unsafe.Pointer, uintptr, uintptr) {
	return C.game1_settings_global_40aa00(), uintptr(C.game1_settings_global_width_40aa00()), uintptr(C.game1_settings_global_align_40aa00())
}
