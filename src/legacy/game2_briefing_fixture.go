package legacy

/*
#include <stddef.h>
#include "GAME2.h"
#include "quest_win_screen_450770.h"
extern uint32_t dword_5d4594_831220;
extern uint32_t dword_587000_122848,dword_5d4594_831092,dword_5d4594_831260,dword_5d4594_831244;
extern uintptr_t dword_5d4594_832484;
extern nox_window*dword_5d4594_831236,*nox_wnd_briefing_831232;
static size_t g2bt_layout(int f){switch(f){case 0:return sizeof(nox_thing);case 1:return offsetof(nox_thing,pri_class);case 2:return offsetof(nox_thing,sub_class);case 3:return sizeof(nox_quest_stats_row_450770);case 4:return offsetof(nox_quest_stats_row_450770,score);default:return (size_t)-1;}}
static uint32_t g2bt_word(int f,uint32_t v,int set){uint32_t*p;switch(f){case 0:p=&dword_587000_122848;break;case 1:p=&dword_5d4594_831092;break;case 2:p=&dword_5d4594_831220;break;case 3:p=&dword_5d4594_831260;break;default:p=&dword_5d4594_831244;}if(set)*p=v;return*p;}
static void* g2bt_pointer(int f,void*p,int set){void**q;switch(f){case 0:q=(void**)&dword_5d4594_831236;break;case 1:q=(void**)&nox_wnd_briefing_831232;break;default:q=(void**)&dword_5d4594_832484;}if(set)*q=p;return*q;}
static int g2bt_call(int op,int v,void*p){switch(op){case 0:return sub_44D040(v);case 1:return sub_44D060(v);case 2:return sub_44D090(v);case 3:sub_44D960();return 0;case 4:return sub_44D970();case 5:return sub_44D990();case 6:return (unsigned char)sub_450760((char)v);case 7:return sub_450750();case 8:nox_gui_setQuestStage_450B00(v);return 0;case 9:return nox_gui_getQuestStage_450B10();case 10:return sub_4505E0();case 11:return sub_44E8D0(p,p);default:return nox_xxx_wndProc_44E6E0(p,v,(uintptr_t)p,(uintptr_t)p);}}
static double g2bt_scroll(void){return sub_44E8B0();}
static int g2bt_compare(void*a,void*b){return sub_450960(a,b);}
static void* g2bt_image(void*p){return sub_450AD0(p);}
static void* g2bt_text(void*p){return sub_450AF0(p);}
*/
import "C"
import "unsafe"

func g2bTestLayout(f int) uintptr { return uintptr(C.g2bt_layout(C.int(f))) }
func g2bTestWord(f int, v uint32, set bool) uint32 {
	s := C.int(0)
	if set {
		s = 1
	}
	return uint32(C.g2bt_word(C.int(f), C.uint32_t(v), s))
}
func g2bTestPointer(f int, p unsafe.Pointer, set bool) unsafe.Pointer {
	s := C.int(0)
	if set {
		s = 1
	}
	return C.g2bt_pointer(C.int(f), p, s)
}
func g2bTestCall(op, v int, p unsafe.Pointer) int32 {
	return int32(C.g2bt_call(C.int(op), C.int(v), p))
}
func g2bTestScroll() float64                       { return float64(C.g2bt_scroll()) }
func g2bTestCompare(a, b unsafe.Pointer) int32     { return int32(C.g2bt_compare(a, b)) }
func g2bTestImage(p unsafe.Pointer) unsafe.Pointer { return C.g2bt_image(p) }
func g2bTestText(p unsafe.Pointer) unsafe.Pointer  { return C.g2bt_text(p) }
