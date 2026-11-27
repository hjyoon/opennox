package legacy

/*
#include "world_native_storage.h"
#define dword_8531A0_2576 (nox_world_owner_storage()->local_player_word)
#define nox_xxx_aClosewoodengat_587000_133480 (nox_world_owner_storage()->quickbar_data)
#include <stddef.h>
#include <string.h>
#include "GAME2.h"
#include "game2_quickbar_native.h"
static size_t g2qt_layout(int f){switch(f){case 0:return sizeof(nox_quickbar_window_set);case 1:return offsetof(nox_quickbar_window_set,data);case 2:return offsetof(nox_quickbar_window_set,root);case 3:return offsetof(nox_quickbar_window_set,buttons);case 4:return offsetof(nox_quickbar_window_set,nuggets);case 5:return offsetof(nox_quickbar_window_set,nugget_children);case 6:return sizeof(nox_window);case 7:return offsetof(nox_window,field_92);case 8:return sizeof(nox_playerInfo);case 9:return offsetof(nox_playerInfo,info.playerClass);default:return (size_t)-1;}}
static void* g2qt_array(void) {
    return nox_quickbar_window_storage();
}
static void* g2qt_pointer(int field, void* value, int set) {
    nox_quickbar_controls_t* controls = nox_quickbar_controls_storage();
    void** slot;
    switch (field) {
    case 0: slot = &nox_xxx_aClosewoodengat_587000_133480; break;
    case 1: slot = (void**)&dword_8531A0_2576; break;
    case 2: slot = (void**)&controls->trap_construct; break;
    case 3: slot = (void**)&controls->right; break;
    case 4: slot = (void**)&controls->left; break;
    case 5: slot = (void**)&controls->title; break;
    case 6: slot = (void**)&controls->trap_activate; break;
    case 7: slot = (void**)&controls->book_button; break;
    case 9: slot = &controls->drag_source_data; break;
    case 10: slot = &controls->drag_source_row; break;
    default: slot = (void**)&controls->dragged_slot;
    }
    if (set) {
        *slot = value;
    }
    return *slot;
}
static uint32_t g2qt_word(int field, uint32_t value, int set) {
    nox_quickbar_controls_t* controls = nox_quickbar_controls_storage();
    uint32_t* word;
    switch (field) {
    case 0: word = &controls->expanded; break;
    case 1: word = &controls->trap_visible; break;
    case 2: word = &controls->targeting_ability; break;
    default: word = &controls->target_ability_id; break;
    }
    if (set) *word = value;
    return *word;
}
static void g2qt_owner(int index, void* data, nox_window* root,
                       nox_window* button, nox_window* nugget, nox_window* child) {
    nox_quickbar_window_set* set = &nox_quickbar_window_storage()[index];
    set->data = data;
    set->root = root;
    for (int i = 0; i < 5; i++) {
        set->buttons[i] = &button[i];
        set->nuggets[i] = &nugget[i];
        set->nugget_children[i] = &child[i];
    }
}
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
	C.g2qt_owner(C.int(index), data, (*C.nox_window)(root),
		(*C.nox_window)(buttons), (*C.nox_window)(nuggets), (*C.nox_window)(children))
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

func g2qTestBookInsert(kind uintptr, spell, slot int32) uintptr {
	return uintptr(C.nox_xxx_book_45DBE0(C.uintptr_t(kind), C.int(spell), C.int(slot)))
}

func g2qTestEvent(win unsafe.Pointer, event int, pos uint32) int32 {
	return int32(C.g2qt_event(win, C.int(event), C.uint(pos)))
}

func g2qTestPlayerClass(p unsafe.Pointer) byte {
	return byte(game2PlayerClass((*C.nox_playerInfo)(p)))
}
