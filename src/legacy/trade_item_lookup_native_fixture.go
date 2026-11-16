package legacy

/*
#include "GAME4_1.h"
#include <stddef.h>
typedef struct lookup_native_node_C{nox_object_t*item;uint32_t cost;struct lookup_native_node_C*next,*prev;}lookup_native_node_C;
typedef void*(*lookup_native_callback_C)(void*,int);
_Static_assert(_Generic(&sub_50FFE0,lookup_native_callback_C:1,default:0),"native pointer argument and pointer return with int32 key");
_Static_assert(sizeof(int)==4,"original key32 ABI");
static void*lookup_native_call_C(int route,void*head,int code){if(route==1){lookup_native_callback_C f=sub_50FFE0;return f(head,code);}return sub_50FFE0(head,code);}
static void lookup_native_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(lookup_native_node_C);v[2]=offsetof(lookup_native_node_C,item);v[3]=offsetof(lookup_native_node_C,cost);v[4]=offsetof(lookup_native_node_C,next);v[5]=offsetof(lookup_native_node_C,prev);
 v[6]=sizeof(((lookup_native_node_C*)0)->item);v[7]=sizeof(((lookup_native_node_C*)0)->cost);v[8]=sizeof(((lookup_native_node_C*)0)->next);v[9]=offsetof(nox_object_t,net_code);v[10]=sizeof(((nox_object_t*)0)->net_code);v[11]=sizeof(lookup_native_callback_C);v[12]=sizeof(int);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func tradeItemLookupNativeLayoutC() (out [13]uintptr) {
	C.lookup_native_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func tradeItemLookupNativeCall(route int, head *server.TradeItem, code uint32) *server.TradeItem {
	if route == 2 {
		return game4TradeFindItem50FFE0(head, code)
	}
	return (*server.TradeItem)(C.lookup_native_call_C(C.int(route), unsafe.Pointer(head), C.int(code)))
}
