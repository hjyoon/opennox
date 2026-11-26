package legacy

/*
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME4_1.h"
#include "client__gui__guiinv.h"
extern unsigned int dword_5d4594_527988;
extern uint32_t dword_5d4594_805820;
extern uint32_t nox_xxx_useAudio_587000_80772;
extern uint32_t dword_5d4594_815052;
extern uint32_t dword_5d4594_1047932;
extern uint32_t dword_5d4594_1062552;
#include "client_ui_words.h"
extern uint32_t dword_5d4594_1098624;
extern uint32_t dword_5d4594_1098628;
extern uint32_t dword_5d4594_1123520;
typedef int(*ui_word_callback_C)(void);
_Static_assert(_Generic(&sub_41E2F0,ui_word_callback_C:1,default:0),"word32 ABI sub_41E2F0");
_Static_assert(_Generic(&nox_client_mousePriKey_430AF0,ui_word_callback_C:1,default:0),"word32 ABI nox_client_mousePriKey_430AF0");
_Static_assert(_Generic(&nox_xxx_cursor_430B00,ui_word_callback_C:1,default:0),"word32 ABI nox_xxx_cursor_430B00");
_Static_assert(_Generic(&sub_43AF30,ui_word_callback_C:1,default:0),"word32 ABI sub_43AF30");
_Static_assert(_Generic(&sub_43AF40,ui_word_callback_C:1,default:0),"word32 ABI sub_43AF40");
_Static_assert(_Generic(&sub_43AF80,ui_word_callback_C:1,default:0),"word32 ABI sub_43AF80");
_Static_assert(_Generic(&sub_4611A0,ui_word_callback_C:1,default:0),"word32 ABI sub_4611A0");
_Static_assert(_Generic(&sub_4674A0,ui_word_callback_C:1,default:0),"word32 ABI sub_4674A0");
_Static_assert(_Generic(&sub_4675B0,ui_word_callback_C:1,default:0),"word32 ABI sub_4675B0");
_Static_assert(_Generic(&sub_46A4A0,ui_word_callback_C:1,default:0),"word32 ABI sub_46A4A0");
_Static_assert(_Generic(&nox_xxx_playerGet_470A90,ui_word_callback_C:1,default:0),"word32 ABI nox_xxx_playerGet_470A90");
_Static_assert(_Generic(&sub_478030,ui_word_callback_C:1,default:0),"word32 ABI sub_478030");
_Static_assert(_Generic(&sub_479590,ui_word_callback_C:1,default:0),"word32 ABI sub_479590");
_Static_assert(_Generic(&sub_47A260,ui_word_callback_C:1,default:0),"word32 ABI sub_47A260");
static ui_word_callback_C ui_word_entries_C[]={sub_41E2F0,nox_client_mousePriKey_430AF0,nox_xxx_cursor_430B00,sub_43AF30,sub_43AF40,sub_43AF80,sub_4611A0,sub_4674A0,sub_4675B0,sub_46A4A0,nox_xxx_playerGet_470A90,sub_478030,sub_479590,sub_47A260};
static uint32_t*ui_word_globals_C[]={&dword_5d4594_527988,&dword_5d4594_805820,&nox_xxx_useAudio_587000_80772,&dword_5d4594_815052,&nox_game_createOrJoin_815048,&dword_5d4594_814548,&dword_5d4594_1047932,&dword_5d4594_1062552,NULL,NULL,NULL,&dword_5d4594_1098624,&dword_5d4594_1098628,&dword_5d4594_1123520};
static uint32_t* ui_word_address_C(unsigned i) {
    switch (i) {
    case 8: return &nox_inventory_state_storage()->mode;
    case 9: return nox_client_chat_active_storage();
    case 10: return nox_client_meter_visible_storage();
    default: return ui_word_globals_C[i];
    }
}
static uint32_t ui_word_get_C(unsigned i){return *ui_word_address_C(i);}
static void ui_word_set_C(unsigned i,uint32_t v){*ui_word_address_C(i)=v;}
static int ui_word_call_C(int route,unsigned i){if(route==1){ui_word_callback_C f=ui_word_entries_C[i];return f();}switch(i){
case 0:return sub_41E2F0();
case 1:return nox_client_mousePriKey_430AF0();
case 2:return nox_xxx_cursor_430B00();
case 3:return sub_43AF30();
case 4:return sub_43AF40();
case 5:return sub_43AF80();
case 6:return sub_4611A0();
case 7:return sub_4674A0();
case 8:return sub_4675B0();
case 9:return sub_46A4A0();
case 10:return nox_xxx_playerGet_470A90();
case 11:return sub_478030();
case 12:return sub_479590();
case 13:return sub_47A260();
default:return 0;}}

#include <stddef.h>
#include <string.h>
static void ui_trade_pointer_store_C(void* field, void* value) { memcpy(field, &value, sizeof(value)); }
typedef struct ui_trade_item_C{nox_object_t*item;uint32_t cost;struct ui_trade_item_C*next,*prev;}ui_trade_item_C;
typedef struct ui_trade_session_C{uint32_t f0,f4;nox_object_t*first,*second;uint32_t f16;ui_trade_item_C*f20;uint32_t f24,f28;ui_trade_item_C*headA,*headB;uint32_t valueA,valueB;nox_object_t*goldA,*goldB;struct ui_trade_session_C*prev,*next;}ui_trade_session_C;
typedef int(*ui_trade_callback_C)(void*,nox_object_t*);
_Static_assert(_Generic(&sub_50FD20,ui_trade_callback_C:1,default:0),"native two-pointer trade ABI");
static int ui_trade_call_C(int route,void*s,nox_object_t*o){if(route==1){ui_trade_callback_C f=sub_50FD20;return f(s,o);}return sub_50FD20(s,o);}
static void ui_trade_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(ui_trade_session_C);v[2]=offsetof(ui_trade_session_C,first);v[3]=offsetof(ui_trade_session_C,headA);v[4]=offsetof(ui_trade_session_C,headB);v[5]=offsetof(ui_trade_session_C,goldA);v[6]=offsetof(ui_trade_session_C,goldB);
 v[7]=sizeof(ui_trade_item_C);v[8]=offsetof(ui_trade_item_C,cost);v[9]=offsetof(ui_trade_item_C,next);v[10]=sizeof(((ui_trade_item_C*)0)->cost);v[11]=sizeof(((ui_trade_item_C*)0)->next);v[12]=offsetof(nox_object_t,init_data);v[13]=sizeof(((nox_object_t*)0)->init_data);
}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func uiStateTradeLayoutC() (out [14]uintptr) {
	C.ui_trade_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

// Seeded foreign fields can contain bits that look like stale Go pointers.
// Populate them in C so Go's deletion barrier does not shade the old canary.
func uiStateTradeStorePointer(field, value unsafe.Pointer) {
	C.ui_trade_pointer_store_C(field, value)
}
func uiStateWordGet(i int) uint32    { return uint32(C.ui_word_get_C(C.uint(i))) }
func uiStateWordSet(i int, v uint32) { C.ui_word_set_C(C.uint(i), C.uint32_t(v)) }
func uiStateWordCall(route, i int) int32 {
	if route != 2 {
		return int32(C.ui_word_call_C(C.int(route), C.uint(i)))
	}
	switch i {
	case 0:
		return int32(sub_41E2F0())
	case 1:
		return int32(nox_client_mousePriKey_430AF0())
	case 2:
		return int32(nox_xxx_cursor_430B00())
	case 3:
		return int32(sub_43AF30())
	case 4:
		return int32(sub_43AF40())
	case 5:
		return int32(sub_43AF80())
	case 6:
		return int32(sub_4611A0())
	case 7:
		return int32(sub_4674A0())
	case 8:
		return int32(sub_4675B0())
	case 9:
		return int32(sub_46A4A0())
	case 10:
		return int32(nox_xxx_playerGet_470A90())
	case 11:
		return int32(sub_478030())
	case 12:
		return int32(sub_479590())
	case 13:
		return int32(sub_47A260())
	}
	panic("invalid getter")
}
func uiStateWordPublic(i int) int64 {
	switch i {
	case 0:
		return int64(Sub_41E2F0())
	case 1:
		return int64(Nox_client_mousePriKey_430AF0())
	case 2:
		return int64(Nox_xxx_cursor_430B00())
	case 3:
		return int64(Sub_43AF30())
	case 4:
		return int64(Sub_43AF40())
	case 5:
		return int64(Sub_43AF80())
	case 6:
		return int64(Sub_4611A0())
	case 7:
		return int64(Nox_client_gold_4674A0())
	case 8:
		return int64(Sub_4675B0())
	case 9:
		return int64(Sub_46A4A0())
	case 10:
		return int64(Nox_xxx_playerGet_470A90())
	case 11:
		return int64(Sub_478030())
	case 12:
		return int64(Sub_479590())
	case 13:
		return int64(Sub_47A260())
	}
	panic("invalid getter")
}
func uiStateTradeCall(route int, s *server.TradeSession, o *server.Object) int32 {
	if route == 2 {
		return int32(game4TradeOfferValue50FD20(s, o))
	}
	return int32(C.ui_trade_call_C(C.int(route), unsafe.Pointer(s), asObjectC(o)))
}
