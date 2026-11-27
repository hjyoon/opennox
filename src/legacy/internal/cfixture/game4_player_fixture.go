package cfixture

/*
#include <stdint.h>
#include <stdlib.h>
#include "GAME4.h"

_Static_assert(_Generic(&nox_xxx_playerAddGold_4FA590, int32_t(*)(nox_object_t*,int32_t):1, default:0), "native AddGold ABI");
_Static_assert(_Generic(&nox_xxx_playerSubGold_4FA5D0, int32_t(*)(nox_object_t*,uint32_t):1, default:0), "native SubGold ABI");
_Static_assert(_Generic(&sub_4FB050, int(*)(nox_object_t*,nox_object_t*,int*):1, default:0), "native guide damage ABI");
static uintptr_t game4_player_gold_entry(nox_object_t* unit, int op, uint32_t value) {
 switch(op) {
 case 0: return (uintptr_t)nox_xxx_playerAddGold_4FA590(unit,(int32_t)value);
 case 1: return (uintptr_t)nox_xxx_playerSubGold_4FA5D0(unit,value);
 case 2: nox_object_setGold_4FA620(unit,(int32_t)value); return 0;
 case 3: return (uint32_t)nox_xxx_playerGetGold_4FA6B0(unit);
 default: return (uint32_t)nox_object_getGold_4FA6D0(unit);
 }
}
static int game4_player_waypoint_entry(nox_object_t* unit,int op,float x,float y) {
 switch(op) {
 case 0: sub_4F7950(unit);return 0;
 case 1: nox_xxx_playerSetCustomWP_4F79A0(unit,x,y);return 0;
 case 2: return sub_4F9A80(unit);
 case 3: return sub_4F9AB0(unit);
 case 4: return nox_xxx_playerCanMove_4F9BC0(unit);
 default:return nox_xxx_playerCanAttack_4F9C40(unit);
 }
}
static int game4_player_reward_entry(nox_object_t* unit,int kind,nox_object_t* target,char flags) { return nox_xxx_netSendRewardNotify_4FAD50(unit,kind,target,flags); }
static void game4_player_string_entry(nox_object_t* unit,const char* text,char flags) { sub_4FADD0(unit,text,flags); }
static int game4_player_guide_entry(nox_object_t* unit,nox_object_t* target,int* damage) { return damage?sub_4FB050(unit,target,damage):sub_4FB000(unit,target); }
static int game4_player_damage_entry(nox_object_t* unit,nox_object_t* target,int* damage) { return sub_4FB050(unit,target,damage); }
static int game4_player_phoneme_entry(uint32_t code,int8_t phoneme) { return nox_xxx_spellGetPhoneme_4FE1C0(code,phoneme); }
static int game4_player_animation_entry(int mask) {return sub_4FA280(mask);}

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func Game4GoldC(o *server.Object, op int, n uint32) uintptr {
	return uintptr(C.game4_player_gold_entry((*C.nox_object_t)(o.CObj()), C.int(op), C.uint32_t(n)))
}
func Game4WaypointC(o *server.Object, op int, x, y float32) int {
	return int(C.game4_player_waypoint_entry((*C.nox_object_t)(o.CObj()), C.int(op), C.float(x), C.float(y)))
}
func Game4RewardC(o *server.Object, kind int, target *server.Object, flags byte) int {
	return int(C.game4_player_reward_entry((*C.nox_object_t)(o.CObj()), C.int(kind), (*C.nox_object_t)(target.CObj()), C.char(flags)))
}
func Game4StringC(o *server.Object, text string, flags byte) {
	p := C.CString(text)
	defer C.free(unsafe.Pointer(p))
	C.game4_player_string_entry((*C.nox_object_t)(o.CObj()), p, C.char(flags))
}
func Game4GuideC(o, target *server.Object, damage *int32) int32 {
	return int32(C.game4_player_guide_entry((*C.nox_object_t)(o.CObj()), (*C.nox_object_t)(target.CObj()), (*C.int)(unsafe.Pointer(damage))))
}
func Game4DamageC(o, target *server.Object, damage *int32) int32 {
	return int32(C.game4_player_damage_entry((*C.nox_object_t)(o.CObj()), (*C.nox_object_t)(target.CObj()), (*C.int)(unsafe.Pointer(damage))))
}
func Game4PhonemeC(code uint32, p int8) int32 {
	return int32(C.game4_player_phoneme_entry(C.uint32_t(code), C.int8_t(p)))
}
func Game4AnimationC(mask uint32) int32 {
	return int32(C.game4_player_animation_entry(C.int(int32(mask))))
}
