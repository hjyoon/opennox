package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include "GAME4.h"

_Static_assert(_Generic(&nox_xxx_playerAddGold_4FA590, uint32_t*(*)(nox_object_t*,int):1, default:0), "native AddGold ABI");
_Static_assert(_Generic(&nox_xxx_playerSubGold_4FA5D0, uint32_t*(*)(nox_object_t*,unsigned int):1, default:0), "native SubGold ABI");
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

// Current native-C floating expressions, separated from the Go implementation.
static int game4_player_follow_float_reference(float dx,float dy,float speed,float cx,float cy,float* fx,float* fy) {
 double v4=dy, v6;
 if(v4*dy+dx*dx>=100.0) {
  *fx += cx*speed;
  v6=cy*speed;
  *fy += v6;
  return 1;
 }
 return 0;
}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func game4GoldC(o *server.Object, op int, n uint32) uintptr {
	return uintptr(C.game4_player_gold_entry(asObjectC(o), C.int(op), C.uint32_t(n)))
}
func game4WaypointC(o *server.Object, op int, x, y float32) int {
	return int(C.game4_player_waypoint_entry(asObjectC(o), C.int(op), C.float(x), C.float(y)))
}
func game4RewardC(o *server.Object, kind int, target *server.Object, flags byte) int {
	return int(C.game4_player_reward_entry(asObjectC(o), C.int(kind), asObjectC(target), C.char(flags)))
}
func game4StringC(o *server.Object, text string, flags byte) {
	p := C.CString(text)
	defer C.free(unsafe.Pointer(p))
	C.game4_player_string_entry(asObjectC(o), p, C.char(flags))
}
func game4GuideC(o, target *server.Object, damage *int32) int32 {
	return int32(C.game4_player_guide_entry(asObjectC(o), asObjectC(target), (*C.int)(unsafe.Pointer(damage))))
}
func game4DamageC(o, target *server.Object, damage *int32) int32 {
	return int32(C.game4_player_damage_entry(asObjectC(o), asObjectC(target), (*C.int)(unsafe.Pointer(damage))))
}
func game4PhonemeC(code uint32, p int8) int32 {
	return int32(C.game4_player_phoneme_entry(C.uint32_t(code), C.int8_t(p)))
}
func game4AnimationC(mask uint32) int32 {
	return int32(C.game4_player_animation_entry(C.int(int32(mask))))
}
func game4FollowFloatC(dx, dy, speed, cx, cy float32, fx, fy *float32) int {
	return int(C.game4_player_follow_float_reference(C.float(dx), C.float(dy), C.float(speed), C.float(cx), C.float(cy), (*C.float)(unsafe.Pointer(fx)), (*C.float)(unsafe.Pointer(fy))))
}
