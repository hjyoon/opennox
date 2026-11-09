package legacy

/*
#include "defs.h"
#include <stdio.h>
*/
import "C"
import "unsafe"

//export nox_xxx_parseString_409470
func nox_xxx_parseString_409470(file *C.FILE, out *C.uint8_t) C.int {
	return C.int(headSettingsParse409470(unsafe.Pointer(file), unsafe.Pointer(out)))
}

//export nox_server_currentMapGetFilename_409B30
func nox_server_currentMapGetFilename_409B30() *C.char { return (*C.char)(headSettingsPath409470()) }

//export nox_xxx_mapGetMapName_409B40
func nox_xxx_mapGetMapName_409B40() *C.char { return (*C.char)(headSettingsMap409470()) }

//export nox_xxx_gameSetMapPath_409D70
func nox_xxx_gameSetMapPath_409D70(p *C.char) *C.char {
	return (*C.char)(headSettingsSetPath409470(unsafe.Pointer(p)))
}

//export nox_xxx_getServerSubFlags_409E60
func nox_xxx_getServerSubFlags_409E60() C.int { return C.int(headSettingsSubflags409470()) }

//export nox_xxx_servSetPlrLimit_409F80
func nox_xxx_servSetPlrLimit_409F80(v C.int) C.int {
	return C.int(headSettingsSetLimit409470(int32(v)))
}

//export nox_xxx_servGetPlrLimit_409FA0
func nox_xxx_servGetPlrLimit_409FA0() C.int { return C.int(headSettingsGetLimit409470()) }

//export nox_xxx_servGamedataGet_40A020
func nox_xxx_servGamedataGet_40A020(v C.short) C.short {
	return C.short(headSettingsGamedata409470(int16(v)))
}

//export nox_xxx_set3512_40A340
func nox_xxx_set3512_40A340(v C.int) { headSettingsSet3512409470(int32(v)) }

//export nox_xxx_get3512_40A350
func nox_xxx_get3512_40A350() C.int { return C.int(headSettingsGet3512409470()) }

//export nox_xxx_gameSetServername_40A440
func nox_xxx_gameSetServername_40A440(p *C.char) *C.char {
	return (*C.char)(headSettingsSetName409470(unsafe.Pointer(p)))
}

//export nox_xxx_serverOptionsGetServername_40A4C0
func nox_xxx_serverOptionsGetServername_40A4C0() *C.char { return (*C.char)(headSettingsName409470()) }

//export nox_xxx_ruleSetNoRespawn_40A5E0
func nox_xxx_ruleSetNoRespawn_40A5E0(v C.int) C.int {
	return C.int(headSettingsSetRespawn409470(int32(v)))
}

//export nox_server_doPlayersAutoRespawn_40A5F0
func nox_server_doPlayersAutoRespawn_40A5F0() C.int { return C.int(headSettingsRespawn409470()) }

//export nox_xxx_sysopSetPass_40A610
func nox_xxx_sysopSetPass_40A610(p *C.wchar2_t) *C.wchar2_t {
	return (*C.wchar2_t)(headSettingsSetPass409470(unsafe.Pointer(p)))
}

//export nox_xxx_sysopGetPass_40A630
func nox_xxx_sysopGetPass_40A630() *C.wchar2_t { return (*C.wchar2_t)(headSettingsPass409470()) }

//export nox_server_gameSettingsUpdated_40A670
func nox_server_gameSettingsUpdated_40A670() { headSettingsDirty409470(1) }

//export nox_server_gameDoSwitchMap_40A680
func nox_server_gameDoSwitchMap_40A680() C.int { return C.int(headSettingsDirtyGet409470()) }

//export nox_server_gameUnsetMapLoad_40A690
func nox_server_gameUnsetMapLoad_40A690() { headSettingsDirty409470(0) }

//export nox_xxx_rateGet_40A6C0
func nox_xxx_rateGet_40A6C0() C.int { return C.int(headSettingsRate409470()) }

//export nox_xxx_rateUpdate_40A6D0
func nox_xxx_rateUpdate_40A6D0(v C.int) C.int { return C.int(headSettingsRateUpdate409470(int32(v))) }
