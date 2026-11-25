package cfixture

/*
#include "GAME2_1.h"
#include "memmap.h"
extern nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX];
typedef char*(*inventory_lookup_legacy_C)(int);
typedef nox_inventory_cell_t*(*inventory_lookup_native_C)(int,uint32_t*);
_Static_assert(_Generic(&sub_461EF0,inventory_lookup_legacy_C:1,default:0),"legacy inventory search ABI");
_Static_assert(_Generic(&nox_inventory_find_cell_native_461EF0,inventory_lookup_native_C:1,default:0),"typed inventory search ABI");
_Static_assert(sizeof(nox_inventory_cell_t)-offsetof(nox_inventory_cell_t,field_4)==144,"fixed scalar inventory tail");
static inventory_lookup_legacy_C volatile inventory_lookup_legacy_callback_C=sub_461EF0;
static inventory_lookup_native_C volatile inventory_lookup_native_callback_C=nox_inventory_find_cell_native_461EF0;
static void*inventory_lookup_call_C(unsigned route,unsigned entry,int query,uint32_t*out){
 if(route){if(entry){inventory_lookup_native_C f=inventory_lookup_native_callback_C;return f(query,out);}inventory_lookup_legacy_C f=inventory_lookup_legacy_callback_C;return f(query);}
 return entry?(void*)nox_inventory_find_cell_native_461EF0(query,out):(void*)sub_461EF0(query);
}
static void inventory_lookup_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(int);v[2]=sizeof(uint32_t);v[3]=sizeof(uint8_t);v[4]=sizeof(nox_inventory_cell_t);v[5]=offsetof(nox_inventory_cell_t,field_4);v[6]=offsetof(nox_inventory_cell_t,field_140);v[7]=offsetof(nox_inventory_cell_t,field_144);v[8]=sizeof(nox_client_inventory_grid_1050020[0].field_140);v[9]=NOX_INVENTORY_ROW_COUNT;v[10]=NOX_INVENTORY_COL_COUNT;v[11]=NOX_INVENTORY_CELLS_MAX;v[12]=sizeof(nox_client_inventory_grid_1050020);v[13]=sizeof(inventory_lookup_legacy_C);v[14]=sizeof(inventory_lookup_native_C);v[15]=sizeof(nox_client_inventory_grid_1050020[0].field_4);v[16]=sizeof(nox_inventory_cell_t)-offsetof(nox_inventory_cell_t,field_4);
}
static void inventory_lookup_owners_C(uintptr_t*v){v[0]=(uintptr_t)nox_client_inventory_grid_1050020;v[1]=(uintptr_t)getMemAt(0x5d4594,0x6d0000-0x5d4594);v[2]=(uintptr_t)getMemAt(0x5d4594,1049788);v[3]=(uintptr_t)getMemU32Ptr(0x5d4594,1049792);}
static void inventory_lookup_seed_C(const unsigned char*packed,unsigned char*foreign){
 for(unsigned i=0;i<NOX_INVENTORY_CELLS_MAX;i++){
  const unsigned char*source=packed+148*i;nox_inventory_cell_t*cell=&nox_client_inventory_grid_1050020[i];uint32_t pointer;memcpy(&pointer,source,4);
  if(pointer&&(pointer<0x21000000||pointer>=0x21010000))abort();
  cell->field_0=pointer?(nox_drawable*)(foreign+pointer-0x21000000):NULL;
  memcpy(&cell->field_4,source+4,144);
 }
}
static void*inventory_lookup_result_pointer_C(void){return getMemPtr(0x5d4594,1049788);}
static uint32_t inventory_lookup_result_index_C(void){return*getMemU32Ptr(0x5d4594,1049792);}
static unsigned inventory_lookup_parent_drag_calls_C;
static nox_drawable*inventory_lookup_parent_no_drag_C(void){inventory_lookup_parent_drag_calls_C++;return NULL;}
#define sub_467700 inventory_lookup_parent_copy_C
#define nox_client_inventory_get_dragged inventory_lookup_parent_no_drag_C
static int sub_467700(int a1) {
	nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(a1, NULL);
	if (cell) {
		return cell->field_140;
	}
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (dragged && dragged->field_32 == (uint32_t)a1) {
		return 1;
	}
	return 0;
}
#undef sub_467700
#undef nox_client_inventory_get_dragged
static void inventory_lookup_parent_C(int query,uint32_t*v){inventory_lookup_parent_drag_calls_C=0;v[0]=(uint32_t)inventory_lookup_parent_copy_C(query);v[1]=inventory_lookup_parent_drag_calls_C;}
*/
import "C"

import "unsafe"

func InventoryLookupNativeLayoutC() (out [17]uintptr) {
	C.inventory_lookup_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func InventoryLookupNativeOwnersC() (out [4]uintptr) {
	C.inventory_lookup_owners_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func InventoryLookupNativeGridC() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&C.nox_client_inventory_grid_1050020[0])), int(C.sizeof_nox_inventory_cell_t)*int(C.NOX_INVENTORY_CELLS_MAX))
}

func InventoryLookupNativeSeedC(packed, foreign unsafe.Pointer) {
	C.inventory_lookup_seed_C((*C.uchar)(packed), (*C.uchar)(foreign))
}

func InventoryLookupNativeCall(route, entry int, query uint32, output unsafe.Pointer) unsafe.Pointer {
	return C.inventory_lookup_call_C(C.uint(route), C.uint(entry), C.int(query), (*C.uint32_t)(output))
}

func InventoryLookupNativeResultC() (unsafe.Pointer, uint32) {
	return C.inventory_lookup_result_pointer_C(), uint32(C.inventory_lookup_result_index_C())
}

func InventoryLookupNativeParentC(query uint32) (out [2]uint32) {
	C.inventory_lookup_parent_C(C.int(query), (*C.uint32_t)(unsafe.Pointer(&out[0])))
	return
}
