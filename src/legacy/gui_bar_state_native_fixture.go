package legacy

/*
#include "GAME2_1.h"
typedef int(*gui_bar_state_one_C)(int);
typedef int(*gui_bar_state_two_C)(int,int);
typedef int(*gui_bar_state_get_C)(void);
_Static_assert(_Generic(&sub_470C40,gui_bar_state_one_C:1,default:0),"GUI color ABI");
_Static_assert(_Generic(&nox_xxx_cliSetTotalHealth_470C80,gui_bar_state_two_C:1,default:0),"GUI health maximum ABI");
_Static_assert(_Generic(&sub_470CB0,gui_bar_state_one_C:1,default:0),"GUI health current ABI");
_Static_assert(_Generic(&nox_xxx_cliSetManaAndMax_470CE0,gui_bar_state_two_C:1,default:0),"GUI mana maximum ABI");
_Static_assert(_Generic(&nox_xxx_cliSetMana_470D10,gui_bar_state_one_C:1,default:0),"GUI mana current ABI");
_Static_assert(_Generic(&nox_xxx_cliGetMana_470DD0,gui_bar_state_get_C:1,default:0),"GUI mana getter ABI");
static gui_bar_state_one_C volatile gui_bar_state_ones_C[]={sub_470C40,NULL,sub_470CB0,NULL,nox_xxx_cliSetMana_470D10};
static gui_bar_state_two_C volatile gui_bar_state_twos_C[]={NULL,nox_xxx_cliSetTotalHealth_470C80,NULL,nox_xxx_cliSetManaAndMax_470CE0};
static gui_bar_state_get_C volatile gui_bar_state_getter_C=nox_xxx_cliGetMana_470DD0;
static int gui_bar_state_call_C(unsigned route,unsigned i,int a,int b){
 if(route==1){
  if(i==5){gui_bar_state_get_C f=gui_bar_state_getter_C;return f();}
  if(i==1||i==3){gui_bar_state_two_C f=gui_bar_state_twos_C[i];return f(a,b);}
  gui_bar_state_one_C f=gui_bar_state_ones_C[i];return f(a);
 }
 switch(i){
 case 0:return sub_470C40(a);
 case 1:return nox_xxx_cliSetTotalHealth_470C80(a,b);
 case 2:return sub_470CB0(a);
 case 3:return nox_xxx_cliSetManaAndMax_470CE0(a,b);
 case 4:return nox_xxx_cliSetMana_470D10(a);
 case 5:return nox_xxx_cliGetMana_470DD0();
 default:abort();
 }
}
static void* gui_bar_state_player_word_C(void*p,unsigned i){
 nox_playerInfo*player=p;
 return i==12?(void*)&player->info.field_2243:(void*)&player->info.field_2247;
}
static void gui_bar_state_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_playerInfo);v[2]=offsetof(nox_playerInfo,info);v[3]=sizeof(((nox_playerInfo*)0)->info);
 v[4]=offsetof(nox_playerInfo,info.field_2243);v[5]=offsetof(nox_playerInfo,info.field_2247);v[6]=sizeof(uint32_t);
 v[7]=sizeof(nox_window_yyy);v[8]=offsetof(nox_window_yyy,field_1);v[9]=offsetof(nox_window_yyy,field_2);
 v[10]=offsetof(nox_window_yyy,color_1);v[11]=offsetof(nox_window_yyy,color_2);v[12]=sizeof(nox_window_yyy)*7;
 v[13]=sizeof(int);v[14]=sizeof(uintptr_t);v[15]=sizeof(((nox_playerInfo*)0)->info.field_2247);v[16]=sizeof(((nox_playerInfo*)0)->info.field_2243);
}
*/
import "C"

import "unsafe"

func guiBarStateNativeLayoutC() (out [17]uintptr) {
	C.gui_bar_state_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func guiBarStateNativePlayerWordC(player unsafe.Pointer, index int) *uint32 {
	return (*uint32)(C.gui_bar_state_player_word_C(player, C.uint(index)))
}

func guiBarStateNativeCall(route, index int, value, maximum int32) int32 {
	if route != 2 {
		return int32(C.gui_bar_state_call_C(C.uint(route), C.uint(index), C.int(value), C.int(maximum)))
	}
	switch index {
	case 0:
		return int32(sub_470C40(C.int(value)))
	case 1:
		return int32(nox_xxx_cliSetTotalHealth_470C80(C.int(value), C.int(maximum)))
	case 2:
		return int32(sub_470CB0(C.int(value)))
	case 3:
		return int32(nox_xxx_cliSetManaAndMax_470CE0(C.int(value), C.int(maximum)))
	case 4:
		return int32(nox_xxx_cliSetMana_470D10(C.int(value)))
	case 5:
		return int32(nox_xxx_cliGetMana_470DD0())
	}
	panic("invalid GUI bar state entry")
}
