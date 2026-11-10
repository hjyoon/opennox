package legacy

/*
#include <stdlib.h>
#include <string.h>
#include <wctype.h>
#include "game3_class_config_reference.h"
#include "game3_class_config_options.h"
extern nox_window* dword_5d4594_1308084;
extern nox_window* dword_5d4594_1309720;
static int game3_fixture_config(int reference, int which, char* p) {
 switch(which) {
 case 0: return reference ? game3_reference_sub_4A7A80(p) : sub_4A7A80(p);
 case 1: return reference ? game3_reference_sub_4A7AC0(p) : sub_4A7AC0(p);
 case 2: return reference ? game3_reference_sub_4A7B00(p) : sub_4A7B00(p);
 case 3: return reference ? game3_reference_sub_4A7BC0(p) : sub_4A7BC0(p);
 case 4: return reference ? game3_reference_sub_4A7C00(p) : sub_4A7C00(p);
 case 5: return reference ? game3_reference_sub_4A7D00(p) : sub_4A7D00(p);
 case 6: return reference ? game3_reference_sub_4A7BA0(p) : sub_4A7BA0(p);
 case 7: return reference ? game3_reference_sub_4A7C40(p) : sub_4A7C40(p);
 case 8: return reference ? game3_reference_sub_4A7CE0(p) : sub_4A7CE0(p);
 case 9: return reference ? game3_reference_sub_4A7D50(p) : sub_4A7D50(p);
 case 10: return reference ? game3_reference_sub_4A7C60(p) : sub_4A7C60(p);
 }
 return -99;
}
static int game3_fixture_name(int reference,wchar2_t* p) {return reference?game3_reference_sub_4A6B50(p):sub_4A6B50(p);}
static uintptr_t game3_fixture_match(int reference,nox_window* w,int group,unsigned char* p) {return (uintptr_t)(reference?game3_reference_sub_4A61E0(w,group,p):sub_4A61E0(w,group,p));}
static int game3_fixture_radio(nox_window* w,int code,uintptr_t a,uintptr_t b,int pre) {return pre?nox_xxx_wndRadioButtonProcPre_4A93C0(w,code,a,b):nox_xxx_wndRadioButtonProc_4A84E0(w,code,a,b);}
static size_t game3_fixture_layout(int which) {switch(which){case 0:return sizeof(nox_window);case 1:return offsetof(nox_window,widget_data);case 2:return offsetof(nox_window,draw_data);case 3:return offsetof(nox_window,parent);case 4:return offsetof(nox_window,field_100);case 5:return sizeof(nox_window_data);}return 0;}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/client/gui"
	"unsafe"
)

func game3FixtureConfig(reference bool, which int, text *string) (int, []byte, uintptr) {
	var p unsafe.Pointer
	if text != nil {
		p = C.malloc(C.size_t(len(*text) + 33))
		defer C.free(p)
		C.memset(p, 0xD3, C.size_t(len(*text)+33))
		copy(unsafe.Slice((*byte)(unsafe.Add(p, 8)), len(*text)+1), *text)
		*(*byte)(unsafe.Add(p, 8+len(*text))) = 0
	}
	old := C.int(0)
	if reference {
		old = 1
	}
	var arg *C.char
	if p != nil {
		arg = (*C.char)(unsafe.Add(p, 8))
	}
	r := C.game3_fixture_config(old, C.int(which), arg)
	var after []byte
	if p != nil {
		after = append([]byte(nil), unsafe.Slice((*byte)(p), len(*text)+33)...)
	}
	return int(r), after, uintptr(p)
}
func game3FixtureName(reference bool, words []uint16) (int, []byte) {
	p := C.malloc(C.size_t(2*len(words) + 32))
	defer C.free(p)
	C.memset(p, 0xA7, C.size_t(2*len(words)+32))
	copy(unsafe.Slice((*uint16)(unsafe.Add(p, 8)), len(words)), words)
	ref := C.int(0)
	if reference {
		ref = 1
	}
	r := C.game3_fixture_name(ref, (*C.wchar2_t)(unsafe.Add(p, 8)))
	return int(r), append([]byte(nil), unsafe.Slice((*byte)(p), 2*len(words)+32)...)
}
func game3FixtureMatch(ref bool, w *gui.Window, group int, rgb [3]byte) uintptr {
	p := C.malloc(3)
	defer C.free(p)
	copy(unsafe.Slice((*byte)(p), 3), rgb[:])
	v := C.int(0)
	if ref {
		v = 1
	}
	return uintptr(C.game3_fixture_match(v, (*C.nox_window)(w.C()), C.int(group), (*C.uchar)(p)))
}
func game3FixtureSwapColorRoot(w *gui.Window) *gui.Window {
	old := asWindow(C.dword_5d4594_1308084)
	C.dword_5d4594_1308084 = (*C.nox_window)(w.C())
	return old
}
func game3FixtureSwapOptionsRoot(w *gui.Window) *gui.Window {
	old := asWindow(C.dword_5d4594_1309720)
	C.dword_5d4594_1309720 = (*C.nox_window)(w.C())
	return old
}
func game3FixtureRadio(w *gui.Window, code int, a, b uintptr, pre bool) int {
	v := C.int(0)
	if pre {
		v = 1
	}
	return int(C.game3_fixture_radio((*C.nox_window)(w.C()), C.int(code), C.uintptr_t(a), C.uintptr_t(b), v))
}
func game3FixtureRadioInit(w *gui.Window) int {
	return int(C.nox_xxx_wndRadioButtonSetAllFn_4A87E0((*C.nox_window)(w.C())))
}
func game3FixtureLayout(which int) uintptr { return uintptr(C.game3_fixture_layout(C.int(which))) }
func game3FixtureRadioCallbackWords(w *gui.Window) (uintptr, uintptr) {
	p := (*C.nox_window)(w.C())
	return uintptr(unsafe.Pointer(p.field_93)), uintptr(unsafe.Pointer(p.draw_func))
}
func game3FixtureRadioText(w *gui.Window, words []uint16) []uint16 {
	p := C.malloc(C.size_t(2 * len(words)))
	defer C.free(p)
	copy(unsafe.Slice((*uint16)(p), len(words)), words)
	C.game3_fixture_radio((*C.nox_window)(w.C()), 16385, C.uintptr_t(uintptr(p)), 0, 1)
	cw := (*C.nox_window)(w.C())
	return append([]uint16(nil), unsafe.Slice((*uint16)(unsafe.Pointer(&cw.draw_data.text[0])), 64)...)
}
func game3FixtureConfigAddress() unsafe.Pointer { return unsafe.Pointer(C.sub_4A7EF0()) }
func game3FixtureLookup(text string) int {
	p := C.CString(text)
	defer C.free(unsafe.Pointer(p))
	return int(C.sub_4A7B40(p))
}
func game3FixtureUnknownOptions(w *gui.Window) [3]int {
	return [3]int{int(C.sub_4AAA70()), int(C.sub_4AABE0((*C.nox_window)(w.C()), -902, nil, 0)), int(C.sub_4AB340((*C.nox_window)(w.C()), 21, 1, 0))}
}
