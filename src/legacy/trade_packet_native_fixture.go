package legacy

/*
#include "GAME4_1.h"
#include <stddef.h>
typedef struct trade_packet_item_C{nox_object_t*item;uint32_t cost;struct trade_packet_item_C*next,*prev;}trade_packet_item_C;
typedef struct trade_packet_session_C{uint32_t f0,f4;nox_object_t*first,*second;uint32_t f16;trade_packet_item_C*f20;uint32_t f24,f28;trade_packet_item_C*headA,*headB;uint32_t valueA,valueB;nox_object_t*goldA,*goldB;struct trade_packet_session_C*prev,*next;}trade_packet_session_C;
typedef int(*trade_packet_session_callback_C)(nox_object_t*,void*);
typedef int(*trade_packet_remove_callback_C)(nox_object_t*,void*,nox_object_t*);
_Static_assert(_Generic(&sub_50F720,trade_packet_session_callback_C:1,default:0),"native status pointer ABI");
_Static_assert(_Generic(&nox_xxx_tradeP2PUpdStuff_50FA00,trade_packet_session_callback_C:1,default:0),"native balance pointer ABI");
_Static_assert(_Generic(&sub_50FF90,trade_packet_remove_callback_C:1,default:0),"native remove pointer ABI");
_Static_assert(sizeof(int)==4,"original network service return32 ABI");
static int trade_packet_call_C(int route,int entry,nox_object_t*recipient,void*session,nox_object_t*item){
 if(entry==0){if(route==1){trade_packet_session_callback_C f=sub_50F720;return f(recipient,session);}return sub_50F720(recipient,session);}
 if(entry==1){if(route==1){trade_packet_session_callback_C f=nox_xxx_tradeP2PUpdStuff_50FA00;return f(recipient,session);}return nox_xxx_tradeP2PUpdStuff_50FA00(recipient,session);}
 if(route==1){trade_packet_remove_callback_C f=sub_50FF90;return f(recipient,session,item);}return sub_50FF90(recipient,session,item);
}
static void trade_packet_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(trade_packet_session_C);v[2]=offsetof(trade_packet_session_C,first);v[3]=offsetof(trade_packet_session_C,second);v[4]=offsetof(trade_packet_session_C,f24);v[5]=offsetof(trade_packet_session_C,f28);v[6]=offsetof(trade_packet_session_C,valueA);v[7]=offsetof(trade_packet_session_C,valueB);v[8]=offsetof(trade_packet_session_C,goldA);v[9]=offsetof(trade_packet_session_C,goldB);
 v[10]=sizeof(nox_object_t);v[11]=offsetof(nox_object_t,obj_class);v[12]=offsetof(nox_object_t,data_update);v[13]=sizeof(((nox_object_t*)0)->data_update);v[14]=offsetof(nox_object_t,init_data);v[15]=sizeof(((nox_object_t*)0)->init_data);v[16]=offsetof(nox_object_t,net_code);v[17]=sizeof(((nox_object_t*)0)->net_code);
 v[18]=sizeof(nox_player_update_data_t);v[19]=offsetof(nox_player_update_data_t,player);v[20]=sizeof(((nox_player_update_data_t*)0)->player);v[21]=sizeof(nox_playerInfo);v[22]=offsetof(nox_playerInfo,playerInd);v[23]=sizeof(((nox_playerInfo*)0)->playerInd);v[24]=sizeof(int);v[25]=sizeof(trade_packet_session_callback_C);v[26]=sizeof(trade_packet_remove_callback_C);v[27]=sizeof(((trade_packet_session_C*)0)->f24);v[28]=sizeof(((trade_packet_session_C*)0)->valueA);v[29]=sizeof(((trade_packet_session_C*)0)->first);v[30]=sizeof(uint32_t);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func tradePacketNativeLayoutC() (out [31]uintptr) {
	C.trade_packet_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func tradePacketNativeCall(route, entry int, recipient *server.Object, session unsafe.Pointer, item *server.Object) int32 {
	if route != 2 {
		return int32(C.trade_packet_call_C(C.int(route), C.int(entry), asObjectC(recipient), session, asObjectC(item)))
	}
	switch entry {
	case 0:
		return tradePacketStatus50F720(recipient, (*server.TradeSession)(session))
	case 1:
		return tradePacketBalance50FA00(recipient, (*server.TradeSession)(session))
	default:
		return tradePacketRemove50FF90(recipient, session, item)
	}
}
