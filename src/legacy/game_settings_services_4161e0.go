package legacy

/*
#include <stdint.h>
#include <stddef.h>
#include <limits.h>
#include <string.h>
#include "GAME1.h"
#include "GAME2.h"
#include "GAME1_3.h"
extern uint32_t dword_5d4594_371692;
static uint32_t* game_settings_dirty_4161e0(void){return &dword_5d4594_371692;}
static uint32_t game_settings_word_4161e0(int service,int32_t arg){
 switch(service){
 case 1:return (uint32_t)nox_common_playerInfoCount_416F40();
 case 2:return (uint32_t)nox_xxx_servGetPlrLimit_409FA0();
 case 3:return nox_common_getEngineFlag((nox_engine_flag)arg)?1u:0u;
 case 4:return (uint32_t)sub_43BE50_get_video_mode_id();
 case 5:return nox_common_gameFlags_getVal_40A5B0();
 case 6:return (uint32_t)(int32_t)nox_xxx_servGamedataGet_40A020((short)arg);
 case 7:return (uint32_t)sub_40A180((short)arg);
 case 8:return nox_common_gameFlags_check_40A5C0((unsigned int)arg)?1u:0u;
 case 9:return (uint32_t)sub_453710();
 }
 return 0;
}
static char* game_settings_string_4161e0(int service){return service==12?nox_xxx_serverOptionsGetServername_40A4C0():nox_xxx_mapGetMapName_409B40();}
static void game_settings_fill_4161e0(int service,void* out){
 if(service==10){uint32_t words[5];sub_454040(words);memcpy(out,words,20);}
 else {uint32_t word;sub_4536B0(&word);memcpy(out,&word,4);}
}
*/
import "C"
import "unsafe"

const gameSettingsPlainCharSigned4161E0 = C.CHAR_MIN < 0

func gameSettingsActualDirtyCell4161E0() unsafe.Pointer {
	return unsafe.Pointer(C.game_settings_dirty_4161e0())
}
func gameSettingsDefaultWord4161E0(service int, arg int32) uint32 {
	return uint32(C.game_settings_word_4161e0(C.int(service), C.int32_t(arg)))
}
func gameSettingsDefaultString4161E0(service int) unsafe.Pointer {
	return unsafe.Pointer(C.game_settings_string_4161e0(C.int(service)))
}
func gameSettingsDefaultFill4161E0(service int, out []byte) {
	C.game_settings_fill_4161e0(C.int(service), unsafe.Pointer(&out[0]))
}
func gameSettingsCopy4161E0(dst, src unsafe.Pointer, n uintptr) { C.memcpy(dst, src, C.size_t(n)) }
