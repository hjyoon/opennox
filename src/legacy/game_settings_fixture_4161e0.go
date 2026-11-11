package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include "GAME1.h"
#include "native_fixture_alloc.h"
_Static_assert(sizeof(int)==4 && sizeof(short)==2, "settings scalar ABI");
typedef char* (*settings_index_fn)(int);
typedef void* (*settings_info_fn)(void);
_Static_assert(_Generic(&sub_4165D0,settings_index_fn:1,default:0), "native index full char pointer");
_Static_assert(_Generic(&sub_416640,settings_info_fn:1,default:0), "native settings void pointer");
static int32_t game_settings_fixture_word(int entry,int32_t a,int32_t b){
 switch(entry){
 case 0x4161e0:return sub_4161E0();case 0x4164f0:sub_4164F0();return 0;
 case 0x416580:return sub_416580();case 0x4165f0:return sub_4165F0(a,b);
 case 0x416650:return sub_416650();case 0x416690:sub_416690();return 0;
 case 0x4169c0:return sub_4169C0();
 }abort();
}
static void* game_settings_fixture_pointer(int entry,int32_t a){
 switch(entry){case 0x4165b0:return sub_4165B0();case 0x4165d0:return sub_4165D0(a);
 case 0x416630:return sub_416630();case 0x416640:return sub_416640();case 0x4169f0:return sub_4169F0();}
 abort();
}
*/
import "C"
import "unsafe"

func gameSettingsFixtureWord4161E0(entry uint32, a, b int32) int32 {
	return int32(C.game_settings_fixture_word(C.int(entry), C.int32_t(a), C.int32_t(b)))
}
func gameSettingsFixturePointer4161E0(entry uint32, a int32) unsafe.Pointer {
	return C.game_settings_fixture_pointer(C.int(entry), C.int32_t(a))
}
func gameSettingsFixtureAlloc4161E0(n uintptr) unsafe.Pointer {
	return C.nox_native_fixture_calloc_high(1, C.size_t(n))
}
func gameSettingsFixtureFree4161E0(p unsafe.Pointer) { C.nox_native_fixture_free_high(p) }
