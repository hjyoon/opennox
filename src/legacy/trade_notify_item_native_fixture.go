package legacy

/*
#include "GAME4_1.h"
#include <stddef.h>
#include <stdint.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
typedef int(*trade_notify_callback_C)(nox_object_t*);
typedef int(*trade_notify_item_callback_C)(nox_object_t*,nox_object_t*,void*,nox_object_t*,int);
_Static_assert(_Generic(&nox_xxx_sendEndTradeMsg_50F560,trade_notify_callback_C:1,default:0),"native end notification ABI");
_Static_assert(_Generic(&sub_50F6E0,trade_notify_callback_C:1,default:0),"native continue notification ABI");
_Static_assert(_Generic(&sub_50FAE0,trade_notify_item_callback_C:1,default:0),"native offered-item pointer ABI");
_Static_assert(sizeof(int)==4,"original cost/return int32 ABI");
static int trade_notify_item_call_C(int route,int entry,nox_object_t*recipient,nox_object_t*offering,void*unused,nox_object_t*item,int cost){
 if(entry==0){if(route==1){trade_notify_callback_C f=nox_xxx_sendEndTradeMsg_50F560;return f(recipient);}return nox_xxx_sendEndTradeMsg_50F560(recipient);}
 if(entry==1){if(route==1){trade_notify_callback_C f=sub_50F6E0;return f(recipient);}return sub_50F6E0(recipient);}
 if(route==1){trade_notify_item_callback_C f=sub_50FAE0;return f(recipient,offering,unused,item,cost);}return sub_50FAE0(recipient,offering,unused,item,cost);
}
static void trade_notify_item_index_C(void*mod,uint32_t index){((obj_412ae0_t*)mod)->field_1=index;}
static void trade_notify_item_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=offsetof(nox_object_t,typ_ind);v[3]=sizeof(((nox_object_t*)0)->typ_ind);v[4]=offsetof(nox_object_t,net_code);v[5]=sizeof(((nox_object_t*)0)->net_code);v[6]=offsetof(nox_object_t,obj_class);v[7]=sizeof(((nox_object_t*)0)->obj_class);v[8]=offsetof(nox_object_t,data_update);v[9]=sizeof(((nox_object_t*)0)->data_update);
 v[10]=sizeof(nox_player_update_data_t);v[11]=offsetof(nox_player_update_data_t,player);v[12]=sizeof(((nox_player_update_data_t*)0)->player);v[13]=sizeof(nox_playerInfo);v[14]=offsetof(nox_playerInfo,playerInd);v[15]=sizeof(((nox_playerInfo*)0)->playerInd);v[16]=offsetof(nox_object_t,init_data);v[17]=sizeof(((nox_object_t*)0)->init_data);
 v[18]=sizeof(nox_modifier_attrs_t);v[19]=offsetof(nox_modifier_attrs_t,modifiers);v[20]=sizeof(((nox_modifier_attrs_t*)0)->modifiers[0]);v[21]=offsetof(nox_modifier_attrs_t,field_16);v[22]=offsetof(obj_412ae0_t,field_1);v[23]=sizeof(((obj_412ae0_t*)0)->field_1);v[24]=sizeof(trade_notify_callback_C);v[25]=sizeof(trade_notify_item_callback_C);v[26]=sizeof(int);v[27]=sizeof(uint32_t);v[28]=sizeof(((nox_modifier_attrs_t*)0)->modifiers);v[29]=offsetof(obj_412ae0_t,field_0);v[30]=sizeof(((obj_412ae0_t*)0)->field_0);
}
// Reserve address space only, then commit two64KiB ends fourGiB apart.
// Every touched page belongs to this reservation; no fixed mapping is used.
static void*trade_notify_item_sparse_pair_C(void){
#if UINTPTR_MAX > UINT32_MAX
 const size_t gap=(size_t)UINT64_C(0x100000000),span=gap+65536;
#ifdef _WIN32
 void*p=VirtualAlloc(NULL,span,MEM_RESERVE|MEM_TOP_DOWN,PAGE_NOACCESS);
 if(!p)return NULL;
 if((uintptr_t)p<=UINT32_MAX||!VirtualAlloc(p,65536,MEM_COMMIT,PAGE_READWRITE)||!VirtualAlloc((uint8_t*)p+gap,65536,MEM_COMMIT,PAGE_READWRITE)){VirtualFree(p,0,MEM_RELEASE);return NULL;}
#else
 void*p=mmap(NULL,span,PROT_NONE,MAP_PRIVATE|MAP_ANON,-1,0);
 if(p==MAP_FAILED)return NULL;
 if((uintptr_t)p<=UINT32_MAX||mprotect(p,65536,PROT_READ|PROT_WRITE)||mprotect((uint8_t*)p+gap,65536,PROT_READ|PROT_WRITE)){munmap(p,span);return NULL;}
#endif
 return p;
#else
 return NULL;
#endif
}
static void*trade_notify_item_sparse_far_C(void*p){
#if UINTPTR_MAX > UINT32_MAX
 return(uint8_t*)p+(size_t)UINT64_C(0x100000000);
#else
 return NULL;
#endif
}
static int trade_notify_item_sparse_free_C(void*p){
#if UINTPTR_MAX > UINT32_MAX
#ifdef _WIN32
 return VirtualFree(p,0,MEM_RELEASE)!=0;
#else
 return munmap(p,(size_t)UINT64_C(0x100000000)+65536)==0;
#endif
#else
 return 0;
#endif
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func tradeNotifyItemNativeLayoutC() (out [31]uintptr) {
	C.trade_notify_item_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func tradeNotifyItemNativeSetIndex(mod *server.ModifierEff, index uint32) {
	C.trade_notify_item_index_C(unsafe.Pointer(mod), C.uint32_t(index))
}

func tradeNotifyItemNativeCall(route, entry int, recipient, offering *server.Object, unused unsafe.Pointer, item *server.Object, cost uint32) int32 {
	if route != 2 {
		return int32(C.trade_notify_item_call_C(C.int(route), C.int(entry), asObjectC(recipient), asObjectC(offering), unused, asObjectC(item), C.int(cost)))
	}
	switch entry {
	case 0:
		return tradeNotifyEnd50F560(recipient)
	case 1:
		return tradeNotifyContinue50F6E0(recipient)
	default:
		return tradeNotifyItem50FAE0(recipient, offering, unused, item, cost)
	}
}

func tradeNotifyItemNativeOwnerPair() (near, far []byte, free func()) {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		all, release := alloc.MakeHigh([]byte{}, 2*65536)
		return all[:65536], all[65536:], release
	}
	p := C.trade_notify_item_sparse_pair_C()
	if p == nil {
		panic("cannot reserve sparse native owner pair")
	}
	return unsafe.Slice((*byte)(p), 65536), unsafe.Slice((*byte)(C.trade_notify_item_sparse_far_C(p)), 65536), func() {
		if C.trade_notify_item_sparse_free_C(p) == 0 {
			panic("cannot release sparse native owner pair")
		}
	}
}
