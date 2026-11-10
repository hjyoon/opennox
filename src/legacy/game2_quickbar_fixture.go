package legacy

/*
#include <stddef.h>
#include <string.h>
#include "GAME2.h"
#include "game2_quickbar_native.h"
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_5d4594_1049496,dword_5d4594_1049484,dword_5d4594_1047932,dword_5d4594_1047936;
extern void*dword_5d4594_1049692,*dword_5d4594_1049696;
extern nox_window *dword_5d4594_1049500,*dword_5d4594_1049504,*dword_5d4594_1049508,*dword_5d4594_1049512,*dword_5d4594_1049520,*dword_5d4594_1049524,*dword_5d4594_1049532;
static size_t g2qt_layout(int f){switch(f){case 0:return sizeof(nox_quickbar_window_set);case 1:return offsetof(nox_quickbar_window_set,data);case 2:return offsetof(nox_quickbar_window_set,root);case 3:return offsetof(nox_quickbar_window_set,buttons);case 4:return offsetof(nox_quickbar_window_set,nuggets);case 5:return offsetof(nox_quickbar_window_set,nugget_children);case 6:return sizeof(nox_window);case 7:return offsetof(nox_window,field_92);case 8:return sizeof(nox_playerInfo);case 9:return offsetof(nox_playerInfo,info.playerClass);default:return (size_t)-1;}}
static void* g2qt_array(void){return nox_quickbar_window_sets;}
static void* g2qt_pointer(int field,void* value,int set){
 void**p;switch(field){case 0:p=&nox_xxx_aClosewoodengat_587000_133480;break;case 1:p=(void**)&dword_8531A0_2576;break;case 2:p=(void**)&dword_5d4594_1049500;break;case 3:p=(void**)&dword_5d4594_1049504;break;case 4:p=(void**)&dword_5d4594_1049508;break;case 5:p=(void**)&dword_5d4594_1049512;break;case 6:p=(void**)&dword_5d4594_1049520;break;case 7:p=(void**)&dword_5d4594_1049524;break;case 9:p=&dword_5d4594_1049692;break;case 10:p=&dword_5d4594_1049696;break;default:p=(void**)&dword_5d4594_1049532;}
 if(set)*p=value;return *p;
}
static uint32_t g2qt_word(int field,uint32_t value,int set){uint32_t*p;switch(field){case 0:p=&dword_5d4594_1049496;break;case 1:p=&dword_5d4594_1049484;break;case 2:p=&dword_5d4594_1047932;break;default:p=&dword_5d4594_1047936;}if(set)*p=value;return*p;}
static void g2qt_owner(int index,void*data,void*root,void*button,void*nugget,void*child){nox_quickbar_window_set*s=&nox_quickbar_window_sets[index];s->data=data;s->root=root;for(int i=0;i<5;i++){s->buttons[i]=(nox_window*)((char*)button+i*sizeof(nox_window));s->nuggets[i]=(nox_window*)((char*)nugget+i*sizeof(nox_window));s->nugget_children[i]=(nox_window*)((char*)child+i*sizeof(nox_window));}}
static void* g2qt_lookup(int kind,void*data,int index){switch(kind){case 0:return nox_quickbar_root(data);case 1:return nox_quickbar_button(data,index);case 2:return nox_quickbar_nugget(data,index);case 3:return nox_quickbar_nugget_child(data,index);default:return nox_quickbar_data_for_window(data);}}
static void g2qt_set(int child,void*data,int index,void*win){if(child)nox_quickbar_set_nugget_child(data,index,win);else nox_quickbar_set_nugget(data,index,win);}
static int g2qt_call(int op,int arg,void*data){switch(op){case 0:nox_quickbar_windows_clear(data);return 0;case 1:nox_quickbar_windows_destroy_all();return 0;case 2:return nox_xxx_clientUpdateButtonRow_45E110(arg);case 3:return nox_xxx_buttonsGetSelectedRow_45E180();case 4:return nox_xxx_quickBarClose_4606B0();case 5:sub_460920();return 0;case 6:return nox_client_trapSetSelect_4604B0(arg);case 7:nox_client_trapSetNext_4603A0();return 0;case 8:nox_client_trapSetPrev_4603F0();return 0;case 9:return sub_4604E0();case 10:return nox_xxx_quickbarButtonBookDraw_45EF30();case 11:return sub_45EF40();default:return sub_460B90(arg);}}
static int g2qt_event(void*win,int event,unsigned pos){return nox_xxx_quickBarWnd_45EF50(win,event,pos);}
static void g2qt_swap(void*data,int a,int b){nox_xxx_clientSwapQuickbarKeys_45F300(data,a,b);}
*/
import "C"
import "unsafe"

func g2qTestLayout(f int) uintptr  { return uintptr(C.g2qt_layout(C.int(f))) }
func g2qTestArray() unsafe.Pointer { return C.g2qt_array() }
func g2qTestPointer(field int, p unsafe.Pointer, set bool) unsafe.Pointer {
	s := C.int(0)
	if set {
		s = 1
	}
	return C.g2qt_pointer(C.int(field), p, s)
}
func g2qTestWord(field int, v uint32, set bool) uint32 {
	s := C.int(0)
	if set {
		s = 1
	}
	return uint32(C.g2qt_word(C.int(field), C.uint32_t(v), s))
}
func g2qTestOwner(index int, data, root, buttons, nuggets, children unsafe.Pointer) {
	C.g2qt_owner(C.int(index), data, root, buttons, nuggets, children)
}
func g2qTestLookup(kind int, data unsafe.Pointer, index int) unsafe.Pointer {
	return C.g2qt_lookup(C.int(kind), data, C.int(index))
}
func g2qTestSet(child bool, data unsafe.Pointer, index int, win unsafe.Pointer) {
	c := C.int(0)
	if child {
		c = 1
	}
	C.g2qt_set(c, data, C.int(index), win)
}
func g2qTestCall(op, arg int, data unsafe.Pointer) int32 {
	return int32(C.g2qt_call(C.int(op), C.int(arg), data))
}
func g2qTestSwap(data unsafe.Pointer, a, b int) { C.g2qt_swap(data, C.int(a), C.int(b)) }

func g2qTestEvent(win unsafe.Pointer, event int, pos uint32) int32 {
	return int32(C.g2qt_event(win, C.int(event), C.uint(pos)))
}

func g2qTestPlayerClass(p unsafe.Pointer) byte {
	return byte(game2PlayerClass((*C.nox_playerInfo)(p)))
}
