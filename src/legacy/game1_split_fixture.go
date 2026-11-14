package legacy

/*
#include "GAME1_3.h"
#include <stddef.h>
extern uint32_t dword_5d4594_815748;
extern uint32_t dword_5d4594_815044;
extern uint32_t dword_5d4594_816368, dword_5d4594_816372, dword_5d4594_816376;
extern uint32_t dword_587000_93156;
extern nox_window* dword_5d4594_816412;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_white_2523948, dword_5d4594_825736;
extern int nox_win_width;
static uint32_t* g1s_word(int slot) {
 switch(slot) {
 case 0:return &dword_5d4594_815748;
 case 1:return &dword_5d4594_815044;
 case 2:return &dword_5d4594_816368;
 case 3:return &dword_5d4594_816372;
 case 4:return &dword_5d4594_816376;
 case 5:return &dword_587000_93156;
 case 7:return &nox_color_white_2523948;
 case 8:return &dword_5d4594_825736;
 default:return &nox_color_black_2650656;
 }
}
static int g1s_call(int op,int value,void* data) {
 switch(op) {
 case 0:return sub_43C650();
 case 1:sub_43CEB0();return 0;
 case 2:return sub_43F0E0(data);
 case 3:return sub_43B6D0();
 case 4:return sub_43BDB0();
 case 5:return sub_43DB30(value);
 case 6:return sub_43DB20();
 case 7:return sub_43DA80();
 case 8:sub_43DAD0();return 0;
 case 9:return sub_43DB60();
 case 10:sub_43DBA0();return 0;
 case 11:sub_43DC00();return 0;
 case 12:return sub_43DC10();
 case 13:return sub_43DC30();
 case 14:return sub_43E8C0(value);
 case 15:return nox_xxx_drawMessageLines_445530();
 default:return -12345;
 }
}
static void* g1s_pointer(int op,int value,void* p,void* out) {
 switch(op) {
 case 0:nox_xxx_xferSetBuffer_446520(value,p);return 0;
 case 1:return nox_xxx_xferGetBuffer_446520(value);
 case 2:return nox_xxx_gui_43E1A0(value);
 case 3:return sub_4466F0(p,out);
 case 4:return sub_43BC10(p,(unsigned char)value);
 case 5:return sub_43DB40(value);
 case 6:return sub_445450();
 case 7:sub_4417E0(out,p);return out;
 default:return 0;
 }
}
static size_t g1s_layout(int op) {
 switch(op) {
 case 0:return sizeof(nox_window);
 case 1:return offsetof(nox_window,draw_data.bg_color);
 case 2:return offsetof(nox_window,draw_data.text_color);
 default:return sizeof(void*);
 }
}
*/
import "C"

import "unsafe"

func g1sWord(slot int, value uint32, set bool) uint32 {
	p := C.g1s_word(C.int(slot))
	if set {
		*p = C.uint32_t(value)
	}
	return uint32(*p)
}
func g1sCall(op, value int, p unsafe.Pointer) int32 {
	return int32(C.g1s_call(C.int(op), C.int(value), p))
}
func g1sPointer(op, value int, p, out unsafe.Pointer) unsafe.Pointer {
	return C.g1s_pointer(C.int(op), C.int(value), p, out)
}
func g1sLayout(op int) uintptr { return uintptr(C.g1s_layout(C.int(op))) }
func g1sSwapBackdrop(p unsafe.Pointer) unsafe.Pointer {
	old := C.dword_5d4594_816412
	C.dword_5d4594_816412 = (*nox_window)(p)
	return unsafe.Pointer(old)
}
func g1sBackdrop() unsafe.Pointer { return unsafe.Pointer(C.dword_5d4594_816412) }
func g1sFormat(addr string, port uint16, dst unsafe.Pointer) int32 {
	return int32(C.nox_sprintAddrPort_43BC80(internCStr(addr), C.ushort(port), (*C.char)(dst)))
}

func g1sSwapWidth(value int32) int32 {
	old := C.nox_win_width
	C.nox_win_width = C.int(value)
	return int32(old)
}
