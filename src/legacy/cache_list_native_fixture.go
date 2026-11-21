package legacy

/*
#include "GAME3_3.h"
static nox_server_netCodeCacheStruct cache_list_fixture;
typedef nox_dblLinkedList*(*cache_list_entry_C)(nox_dblLinkedList**,nox_dblLinkedList*);
_Static_assert(_Generic(&sub_4ECDE0,cache_list_entry_C:1,default:0),"native prepend ABI");
_Static_assert(_Generic(&sub_4ECE10,cache_list_entry_C:1,default:0),"native remove ABI");
static cache_list_entry_C volatile cache_list_prepend_callback_C=sub_4ECDE0;
static cache_list_entry_C volatile cache_list_remove_callback_C=sub_4ECE10;
static void*cache_list_call_C(unsigned route,unsigned entry,void*list,void*node){
 if(route){cache_list_entry_C f=entry?cache_list_remove_callback_C:cache_list_prepend_callback_C;return f(list,node);}
 return entry?sub_4ECE10(list,node):sub_4ECDE0(list,node);
}
static void cache_list_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(uint32_t);v[2]=sizeof(nox_dblLinkedList);
 v[3]=offsetof(nox_dblLinkedList,value);v[4]=offsetof(nox_dblLinkedList,next);v[5]=offsetof(nox_dblLinkedList,prev);
 v[6]=sizeof(nox_server_netCodeCacheStruct);v[7]=offsetof(nox_server_netCodeCacheStruct,firstFreeObject);v[8]=offsetof(nox_server_netCodeCacheStruct,lastFreeObject);
 v[9]=offsetof(nox_server_netCodeCacheStruct,objArray);v[10]=offsetof(nox_server_netCodeCacheStruct,firstUsedObject);v[11]=offsetof(nox_server_netCodeCacheStruct,lastUsedObject);
 v[12]=sizeof(cache_list_fixture.objArray)/sizeof(cache_list_fixture.objArray[0]);v[13]=sizeof(cache_list_entry_C);
}
static void*cache_list_cache_owner_C(void){return &cache_list_fixture;}
static void*cache_list_read_pointer_C(void*owner,unsigned cell){return ((void**)owner)[cell];}
*/
import "C"

import "unsafe"

func cacheListNativeLayoutC() (out [14]uintptr) {
	C.cache_list_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func cacheListNativeCacheC() []byte {
	return unsafe.Slice((*byte)(C.cache_list_cache_owner_C()), int(C.sizeof_nox_server_netCodeCacheStruct))
}

func cacheListNativeReadC(owner unsafe.Pointer, cell int) unsafe.Pointer {
	return C.cache_list_read_pointer_C(owner, C.uint(cell))
}

func cacheListNativeCall(route, entry int, list, node unsafe.Pointer) unsafe.Pointer {
	if route == 2 {
		if entry == 0 {
			return unsafe.Pointer(sub_4ECDE0((**C.nox_dblLinkedList)(list), (*C.nox_dblLinkedList)(node)))
		}
		return unsafe.Pointer(sub_4ECE10((**C.nox_dblLinkedList)(list), (*C.nox_dblLinkedList)(node)))
	}
	return C.cache_list_call_C(C.uint(route), C.uint(entry), list, node)
}
