package cfixture

/*
#include "GAME1_1.h"
#include <stdint.h>
#include <stddef.h>
#include "client_native_storage.h"
#define dword_5d4594_588120 (nox_client_audio_storage()->soundset_reference_head)
typedef struct { uint32_t name_token, sounds[18], next_token, reserved; } soundset_packed_fixture;
_Static_assert(sizeof(soundset_packed_fixture)==84 && offsetof(soundset_packed_fixture,sounds)==4 && offsetof(soundset_packed_fixture,next_token)==76 && offsetof(soundset_packed_fixture,reserved)==80, "packed numeric record remains PE32-sized");
_Static_assert(_Generic(&sub_4240F0, int (*)(void*, const char*, int): 1, default: 0), "native data pointer and const field ABI");
_Static_assert(_Generic(&nox_xxx_parseSoundSetBin_424170, int (*)(char*): 1, default: 0), "public native parser ABI");
static int soundset_field_C(void* data, const char* name, int value) { return sub_4240F0(data,name,value); }
static int soundset_parser_C(char* path) { return nox_xxx_parseSoundSetBin_424170(path); }
static uint32_t* soundset_old_head_C(void) { return &dword_5d4594_588120; }
*/
import "C"
import "unsafe"

func SoundSetFieldC(data, name unsafe.Pointer, value uint32) int32 {
	return int32(C.soundset_field_C(data, (*C.char)(name), C.int(int32(value))))
}

func SoundSetParseC(path unsafe.Pointer) int32 {
	return int32(C.soundset_parser_C((*C.char)(path)))
}

func SoundSetOldHeadC() *uint32 {
	return (*uint32)(unsafe.Pointer(C.soundset_old_head_C()))
}

func SoundSetPackedSizeC() uintptr { return C.sizeof_soundset_packed_fixture }
