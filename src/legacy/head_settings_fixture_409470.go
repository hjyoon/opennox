package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include "GAME1.h"
#include "head_settings_reference_409470.h"
#include "native_fixture_alloc.h"
_Static_assert(sizeof(int)==4 && sizeof(short)==2 && sizeof(wchar2_t)==2,"settings scalar/UTF16 ABI");
typedef int (*head_parse_fn_409470)(FILE*,uint8_t*);
typedef char* (*head_string_set_fn_409470)(char*);
typedef wchar2_t* (*head_wide_set_fn_409470)(wchar2_t*);
typedef int (*head_word_set_fn_409470)(int);
_Static_assert(_Generic(&nox_xxx_parseString_409470,head_parse_fn_409470:1,default:0),"FILE and destination native pointers");
_Static_assert(_Generic(&nox_xxx_gameSetMapPath_409D70,head_string_set_fn_409470:1,default:0),"map public char pointer ABI");
_Static_assert(_Generic(&nox_xxx_gameSetServername_40A440,head_string_set_fn_409470:1,default:0),"servername public char pointer ABI");
_Static_assert(_Generic(&nox_xxx_sysopSetPass_40A610,head_wide_set_fn_409470:1,default:0),"sysop public wchar2 pointer ABI");
_Static_assert(_Generic(&nox_xxx_rateUpdate_40A6D0,head_word_set_fn_409470:1,default:0),"rate remains signed DWORD");
static uint32_t head_fixture_word_409470(int reference,uint32_t entry,int32_t value) {
 switch(entry) {
 case 0x409e60:return (uint32_t)(reference?head_reference_nox_xxx_getServerSubFlags_409E60():nox_xxx_getServerSubFlags_409E60());
 case 0x409f80:return (uint32_t)(reference?head_reference_nox_xxx_servSetPlrLimit_409F80(value):nox_xxx_servSetPlrLimit_409F80(value));
 case 0x409fa0:return (uint32_t)(reference?head_reference_nox_xxx_servGetPlrLimit_409FA0():nox_xxx_servGetPlrLimit_409FA0());
 case 0x40a020:return (uint32_t)(reference?head_reference_nox_xxx_servGamedataGet_40A020((short)value):nox_xxx_servGamedataGet_40A020((short)value));
 case 0x40a340:if(reference){head_reference_nox_xxx_set3512_40A340(value);}else{nox_xxx_set3512_40A340(value);}return 0;
 case 0x40a350:return (uint32_t)(reference?head_reference_nox_xxx_get3512_40A350():nox_xxx_get3512_40A350());
 case 0x40a5e0:return (uint32_t)(reference?head_reference_nox_xxx_ruleSetNoRespawn_40A5E0(value):nox_xxx_ruleSetNoRespawn_40A5E0(value));
 case 0x40a5f0:return (uint32_t)(reference?head_reference_nox_server_doPlayersAutoRespawn_40A5F0():nox_server_doPlayersAutoRespawn_40A5F0());
 case 0x40a670:if(reference){head_reference_nox_server_gameSettingsUpdated_40A670();}else{nox_server_gameSettingsUpdated_40A670();}return 0;
 case 0x40a680:return (uint32_t)(reference?head_reference_nox_server_gameDoSwitchMap_40A680():nox_server_gameDoSwitchMap_40A680());
 case 0x40a690:if(reference){head_reference_nox_server_gameUnsetMapLoad_40A690();}else{nox_server_gameUnsetMapLoad_40A690();}return 0;
 case 0x40a6c0:return (uint32_t)(reference?head_reference_nox_xxx_rateGet_40A6C0():nox_xxx_rateGet_40A6C0());
 case 0x40a6d0:return (uint32_t)(reference?head_reference_nox_xxx_rateUpdate_40A6D0(value):nox_xxx_rateUpdate_40A6D0(value));
 }abort();
}
static void* head_fixture_pointer_409470(int reference,uint32_t entry,void* input) {
 switch(entry) {
 case 0x409b30:return reference?head_reference_nox_server_currentMapGetFilename_409B30():nox_server_currentMapGetFilename_409B30();
 case 0x409b40:return reference?head_reference_nox_xxx_mapGetMapName_409B40():nox_xxx_mapGetMapName_409B40();
 case 0x409d70:return reference?head_reference_nox_xxx_gameSetMapPath_409D70((char*)input):nox_xxx_gameSetMapPath_409D70((char*)input);
 case 0x40a440:return reference?head_reference_nox_xxx_gameSetServername_40A440((char*)input):nox_xxx_gameSetServername_40A440((char*)input);
 case 0x40a4c0:return reference?head_reference_nox_xxx_serverOptionsGetServername_40A4C0():nox_xxx_serverOptionsGetServername_40A4C0();
 case 0x40a610:return reference?head_reference_nox_xxx_sysopSetPass_40A610((wchar2_t*)input):nox_xxx_sysopSetPass_40A610((wchar2_t*)input);
 case 0x40a630:return reference?head_reference_nox_xxx_sysopGetPass_40A630():nox_xxx_sysopGetPass_40A630();
 }abort();
}
static int head_fixture_parse_409470(int reference,void* file,void* out) {
 return reference?head_reference_nox_xxx_parseString_409470((FILE*)file,(uint8_t*)out):nox_xxx_parseString_409470((FILE*)file,(uint8_t*)out);
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func headSettingsFixtureWord409470(reference bool, entry uint32, value int32) uint32 {
	r := C.int(0)
	if reference {
		r = 1
	}
	return uint32(C.head_fixture_word_409470(r, C.uint32_t(entry), C.int32_t(value)))
}
func headSettingsFixturePointer409470(reference bool, entry uint32, input unsafe.Pointer) unsafe.Pointer {
	r := C.int(0)
	if reference {
		r = 1
	}
	out := C.head_fixture_pointer_409470(r, C.uint32_t(entry), input)
	runtime.KeepAlive(input)
	return out
}
func headSettingsFixtureParse409470(reference bool, file, out unsafe.Pointer) int32 {
	r := C.int(0)
	if reference {
		r = 1
	}
	v := int32(C.head_fixture_parse_409470(r, file, out))
	runtime.KeepAlive(file)
	runtime.KeepAlive(out)
	return v
}
func headSettingsFixtureAlloc409470(n uintptr) unsafe.Pointer {
	return C.nox_native_fixture_calloc_high(1, C.size_t(n))
}
func headSettingsFixtureFree409470(p unsafe.Pointer) { C.nox_native_fixture_free_high(p) }
