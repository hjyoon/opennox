package legacy

/*
#include <string.h>
#ifndef _WIN32
#include <sys/mman.h>
#include <unistd.h>
#endif
#include "game3_widgets_native.h"
// Defined reference service: invoke the actual native-width callback directly.
// The global main-client legacy-address recovery service is not an algorithm
// under comparison and cannot own these private C test windows.
static uintptr_t game3_test_reference_event(nox_window* w,int e,uintptr_t a,uintptr_t b) {
 if(!w||!w->field_94)return 0;
 return ((uintptr_t(*)(nox_window*,int,uintptr_t,uintptr_t))(void*)w->field_94)(w,e,a,b);
}
#undef nox_window_call_field_94
#define nox_window_call_field_94(w,e,a,b) game3_test_reference_event(w,e,a,b)
#include "game3_widgets_reference.h"

_Static_assert(sizeof(nox_scrollListBox_data)==(sizeof(void*)==8?80:56), "list data native shape");
_Static_assert(sizeof(((nox_window*)0)->widget_data)==sizeof(void*), "widget full pointer");
_Static_assert(sizeof(((nox_scrollListBox_data*)0)->field_12)==sizeof(void*), "selection full pointer");
typedef struct {int count,event;uintptr_t a,b;} game3_test_trace;
static game3_test_trace* game3_trace_now;
static uintptr_t game3_test_capture(nox_window* w,int e,uintptr_t a,uintptr_t b) {
 (void)w;game3_trace_now->count++;game3_trace_now->event=e;game3_trace_now->a=a;game3_trace_now->b=b;return 0x12345678;
}
static uint32_t* game3_test_position(int ref,int32_t x,int32_t y,uint32_t out[2]) {
 return ref?game3_ref_sub_4A2830(x,y,out):sub_4A2830(x,y,out);
}
static int game3_test_position_owned(int ref,int32_t x,int32_t y,uint32_t out[2],uintptr_t* address) {
 struct block {uint64_t before[2];uint32_t xy[2];uint64_t after[2];}*p=calloc(1,sizeof(*p));
 if(!p)return 0;
 p->before[0]=p->before[1]=p->after[0]=p->after[1]=0xfeedcafe01234567ULL;
 uint32_t* returned=game3_test_position(ref,x,y,p->xy);
 int intact=returned==p->xy&&p->before[0]==0xfeedcafe01234567ULL&&p->before[1]==p->before[0]&&p->after[0]==p->before[0]&&p->after[1]==p->before[0];
 *address=(uintptr_t)p->xy;memcpy(out,p->xy,sizeof(p->xy));free(p);return intact;
}
static int game3_test_distance(int ref,uint32_t x,uint32_t y,int16_t sx,int16_t sy,double radius,uintptr_t* addr) {
 nox_gui_server_ent_t* p=calloc(1,sizeof(*p));uint32_t xy[2]={x,y};p->field_11_0=sx;p->field_11_2=sy;*addr=(uintptr_t)p;
 double* r=getMemDoublePtr(0x581450,9720);double old=*r;*r=radius;
 int v=ref?game3_ref_sub_4A2560(xy,p):sub_4A2560(xy,p);*r=old;free(p);return v;
}
static void game3_test_slider(int ref,int vertical,int pre,int ev,uintptr_t a,uintptr_t b,
 int lo,int hi,uint32_t scale,int value,int64_t out[20]) {
 struct block {uint64_t before[2];nox_window w,thumb,parent,owner;nox_slider_data_4B4860 d;uint64_t after[2];}*p=calloc(1,sizeof(*p));
 p->before[0]=p->before[1]=p->after[0]=p->after[1]=0xfeedcafe01234567ULL;
 p->w.flags=8;p->w.width=101;p->w.height=91;p->w.off_x=13;p->w.off_y=17;
 p->w.end_x=114;p->w.end_y=108;p->w.field_100=&p->thumb;p->w.parent=&p->parent;p->w.draw_data.win=&p->owner;p->w.id=0x12345678;
 p->thumb.flags=8;p->thumb.width=10;p->thumb.height=10;p->w.widget_data=&p->d;
 p->w.field_94=(void*)game3_test_capture;p->parent.field_94=(void*)game3_test_capture;p->owner.field_94=(void*)game3_test_capture;
 p->d.min=lo;p->d.max=hi;memcpy(&p->d.scale,&scale,4);p->d.value=value;
 game3_test_trace trace={0};game3_trace_now=&trace;int v;
 if(pre) {if(vertical)v=ref?game3_ref_nox_xxx_wndScrollBoxProc_4B5320(&p->w,ev,a,b):nox_xxx_wndScrollBoxProc_4B5320(&p->w,ev,a,b);
 else v=ref?game3_ref_sub_4B5010(&p->w,ev,a,b):sub_4B5010(&p->w,ev,a,b);}
 else {if(vertical)v=ref?game3_ref_nox_xxx_wndScrollBoxDraw_4B4BA0(&p->w,ev,a,b):nox_xxx_wndScrollBoxDraw_4B4BA0(&p->w,ev,a,b);
 else v=ref?game3_ref_sub_4B4860(&p->w,ev,a,b):sub_4B4860(&p->w,ev,a,b);}
 memcpy(&scale,&p->d.scale,4);
 int64_t values[20]={v,p->d.min,p->d.max,scale,p->d.value,p->thumb.off_x,p->thumb.off_y,p->thumb.end_x,p->thumb.end_y,
 p->w.draw_data.field_0,trace.count,trace.event,(int64_t)trace.b,trace.a==(uintptr_t)&p->w,
 p->w.widget_data==&p->d,p->w.field_100==&p->thumb,(uintptr_t)&p->w>UINT32_MAX,
 p->before[0]==0xfeedcafe01234567ULL&&p->before[1]==p->before[0]&&p->after[0]==p->before[0]&&p->after[1]==p->before[0],
 (int64_t)(trace.a==(uintptr_t)&p->w?0:trace.a),0};
 memcpy(out,values,sizeof(values));game3_trace_now=NULL;free(p);
}
static void game3_test_list(int ref,int op,int scroll,int delta,int update,uint64_t out[16]) {
 struct block{uint64_t before;nox_window w,slider,thumb;nox_scrollListBox_data d;nox_scrollListBox_item items[4];nox_slider_data_4B4860 values;uint64_t after;}*p=calloc(1,sizeof(*p));
 p->before=p->after=0xfeedbeef12345678ULL;p->w.flags=8;p->w.widget_data=&p->d;
 p->d.items=p->items;p->d.count=4;p->d.field_11_0=4;p->d.field_13_0=21;p->d.field_13_1=scroll;p->d.field_9=&p->slider;
 for(int i=0;i<4;i++){p->items[i].field_130=0x76543200u+(i+1)*7;p->items[i].field_0=(i+1)*8;}
 p->d.field_10=32;p->slider.height=100;p->slider.field_100=&p->thumb;p->thumb.height=10;p->slider.widget_data=&p->values;p->slider.field_94=(void*)game3_test_capture;
 game3_test_trace trace={0};game3_trace_now=&trace;int result;
 if(op==0)result=ref?game3_ref_nox_xxx_wndListBoxFirstVisible(&p->w):nox_xxx_wndListBoxFirstVisible(&p->w);
 else if(op==1)result=ref?game3_ref_nox_xxx_wndListBox_4A2D10(&p->w,delta,update):nox_xxx_wndListBox_4A2D10(&p->w,delta,update);
 else if(op==2)result=ref?game3_ref_nox_xxx_wndListBox_4A3A70(&p->w):nox_xxx_wndListBox_4A3A70(&p->w);
 else result=ref?game3_ref_sub_4A4800(&p->d):sub_4A4800(&p->d);
 uint32_t scale;memcpy(&scale,&p->values.scale,4);
 uint64_t values[16]={(uint32_t)result,p->d.field_13_1,p->d.field_10,(uint32_t)p->values.max,scale,trace.count,trace.event,trace.a,
 p->items[0].field_0,p->items[1].field_0,p->items[2].field_0,p->items[3].field_0,
 p->items[0].field_130,p->before==p->after,(uintptr_t)&p->items[0]>UINT32_MAX,p->w.widget_data==&p->d&&p->d.field_9==&p->slider};
 memcpy(out,values,sizeof(values));game3_trace_now=NULL;free(p);
}
static uint32_t game3_test_palette(int i){return game3_ref_nox_xxx_wndListBoxColor(i);}
static int game3_test_selection(uintptr_t* address) {
 struct {uint64_t before;nox_scrollListBox_data d;uint32_t selection[6];uint64_t after;}*p=calloc(1,sizeof(*p));
 p->before=p->after=0x123456789abcdef0ULL;p->d.count=4;p->d.field_12=p->selection;
 for(int i=0;i<6;i++)p->selection[i]=100+i;*address=(uintptr_t)p;
 short* result=sub_4A3090((short*)&p->d,1);
 int ok=result==(short*)&p->d&&p->d.field_12==p->selection&&p->selection[0]==100&&p->selection[1]==102&&p->selection[2]==103&&p->selection[3]==UINT32_MAX&&p->selection[4]==104&&p->selection[5]==105&&p->before==p->after;free(p);return ok;
}
static int game3_test_destroy(int vertical) {
 nox_window* w=calloc(1,sizeof(*w));w->flags=8;w->widget_data=calloc(1,16);
 int v=vertical?nox_xxx_wndScrollBoxProc_4B5320(w,2,0,0):sub_4B5010(w,2,0,0);
 int ok=v==0&&w->widget_data==NULL;free(w);return ok;
}
static int game3_test_server_count(int ref,uintptr_t* address) {
 nox_list_item_t head; nox_common_list_clear_425760(&head);
 nox_gui_server_node_t* nodes=calloc(3,sizeof(*nodes));*address=(uintptr_t)nodes;
 nodes[0].server.field_11_0=3;nodes[0].server.field_11_2=4;
 nodes[1].server.field_11_0=6;nodes[1].server.field_11_2=8;
 nodes[2].server.field_11_0=-3;nodes[2].server.field_11_2=-4;
 for(int i=0;i<3;i++)nox_common_list_append_4258E0(&head,&nodes[i].list);
 uint32_t xy[2]={0,0};double* radius=getMemDoublePtr(0x581450,9720);double saved=*radius;*radius=5;
 int count=ref?game3_ref_sub_4A25C0(xy,&head):sub_4A25C0(xy,&head);*radius=saved;
 int okay=count==2&&head.field_0==&nodes[0].list&&head.field_1==&nodes[2].list;
 free(nodes);return okay;
}
static int game3_test_guarded_selection(uintptr_t* address) {
#ifndef _WIN32
 size_t page=(size_t)sysconf(_SC_PAGESIZE);if(page<64)return 0;
 unsigned char* bank=mmap(NULL,page*2,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
 if(bank==MAP_FAILED)return 0;
 if(mprotect(bank+page,page,PROT_NONE)!=0){munmap(bank,page*2);return 0;}
 uint32_t* values=(uint32_t*)(bank+page-4*sizeof(uint32_t));
 memset((unsigned char*)values-32,0xA7,32);for(int i=0;i<4;i++)values[i]=10+i;
 nox_scrollListBox_data* data=calloc(1,sizeof(*data));if(!data){munmap(bank,page*2);return 0;}
 data->count=4;data->field_12=values;*address=(uintptr_t)values;
 short* result=sub_4A3090((short*)data,1);
 int okay=result==(short*)data&&data->field_12==values&&values[0]==10&&values[1]==12&&values[2]==13&&values[3]==UINT32_MAX;
 for(int i=1;i<=32;i++)if(*((unsigned char*)values-i)!=0xA7)okay=0;
 free(data);munmap(bank,page*2);return okay;
#else
 (void)address;return 0;
#endif
}
static void game3_test_selection_ptr(void* p) { sub_4A3090((short*)p,0); }
static int game3_test_layout(size_t out[8]) {
 out[0]=sizeof(nox_window);out[1]=sizeof(nox_window_data);out[2]=sizeof(nox_scrollListBox_data);out[3]=sizeof(nox_scrollListBox_item);
 out[4]=offsetof(nox_window,widget_data);out[5]=offsetof(nox_window,field_100);out[6]=offsetof(nox_scrollListBox_data,field_12);out[7]=sizeof(nox_slider_data_4B4860);return 1;
}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client/gui"
	"runtime"
	"unsafe"
)

func game3WidgetPositionCOwner(ref bool, x, y int32) (out [2]uint32, address uintptr, intact bool) {
	r := C.int(0)
	if ref {
		r = 1
	}
	var nativeAddress C.uintptr_t
	intact = C.game3_test_position_owned(r, C.int32_t(x), C.int32_t(y), (*C.uint32_t)(unsafe.Pointer(&out[0])), &nativeAddress) != 0
	return out, uintptr(nativeAddress), intact
}

func game3WidgetReferencePosition(ref bool, x, y int32) (out [2]uint32) {
	out, _, intact := game3WidgetPositionCOwner(ref, x, y)
	if !intact {
		panic("position output identity or canary changed")
	}
	return out
}

func game3WidgetPositionPinnedOwner(x, y int32) (out [2]uint32, same bool) {
	var pin runtime.Pinner
	pin.Pin(&out[0])
	defer pin.Unpin()
	runtime.GC()
	p := (*C.uint32_t)(unsafe.Pointer(&out[0]))
	returned := C.game3_test_position(0, C.int32_t(x), C.int32_t(y), p)
	same = returned == p
	runtime.KeepAlive(&out)
	return out, same
}
func game3WidgetReferenceDistance(ref bool, x, y uint32, sx, sy int16, radius float64) (bool, uintptr) {
	r := C.int(0)
	if ref {
		r = 1
	}
	var addr C.uintptr_t
	ok := C.game3_test_distance(r, C.uint32_t(x), C.uint32_t(y), C.int16_t(sx), C.int16_t(sy), C.double(radius), &addr) != 0
	return ok, uintptr(addr)
}
func game3WidgetReferenceSlider(ref, vertical, pre bool, ev int, a, b uintptr, lo, hi int32, scale uint32, value int32) (out [20]int64) {
	r, v, p := C.int(0), C.int(0), C.int(0)
	if ref {
		r = 1
	}
	if vertical {
		v = 1
	}
	if pre {
		p = 1
	}
	C.game3_test_slider(r, v, p, C.int(ev), C.uintptr_t(a), C.uintptr_t(b), C.int(lo), C.int(hi), C.uint32_t(scale), C.int(value), (*C.int64_t)(unsafe.Pointer(&out[0])))
	return
}
func game3WidgetReferenceList(ref bool, op, scroll, delta int, update bool) (out [16]uint64) {
	r, u := C.int(0), C.int(0)
	if ref {
		r = 1
	}
	if update {
		u = 1
	}
	C.game3_test_list(r, C.int(op), C.int(scroll), C.int(delta), u, (*C.uint64_t)(unsafe.Pointer(&out[0])))
	return
}
func game3WidgetReferenceColor(i int) uint32 { return uint32(C.game3_test_palette(C.int(i))) }
func game3WidgetSelectionCheck() (uintptr, bool) {
	var p C.uintptr_t
	v := C.game3_test_selection(&p) != 0
	return uintptr(p), v
}
func game3WidgetDestroyCheck(vertical bool) bool {
	v := C.int(0)
	if vertical {
		v = 1
	}
	return C.game3_test_destroy(v) != 0
}
func game3WidgetLayout() (out [8]uintptr) {
	C.game3_test_layout((*C.size_t)(unsafe.Pointer(&out[0])))
	return
}

func game3WidgetPinnedSelection(p unsafe.Pointer) { C.game3_test_selection_ptr(p) }
func game3WidgetCreateSlider(parent *gui.Window, draw *gui.WindowData, opts *game3SliderData) *gui.Window {
	return asWindow(C.nox_gui_newSlider_4B4EE0((*C.nox_window)(parent.C()), 8, 0, 0, 100, 100, (*C.nox_window_data)(draw.C()), (*C.float)(unsafe.Pointer(opts))))
}
func game3WidgetInitList(w *gui.Window) { C.nox_xxx_wndListboxInit_4A3C00((*C.nox_window)(w.C()), nil) }
func game3WidgetListDrawWithoutRenderer(w *gui.Window) int {
	return int(C.nox_xxx_wndListboxDrawNoImage_4A3C50((*C.nox_window)(w.C()), (*C.nox_window_data)(w.DrawData().C())))
}

func game3WidgetServerCount(ref bool) (uintptr, bool) {
	r := C.int(0)
	if ref {
		r = 1
	}
	var p C.uintptr_t
	ok := C.game3_test_server_count(r, &p) != 0
	return uintptr(p), ok
}
func game3WidgetRegistryCheck() (uintptr, bool) {
	p := C.calloc(1, 169)
	if p == nil {
		return 0, false
	}
	defer C.free(p)
	old, oldCount := game3NearbyServers, C.dword_5d4594_1307720
	defer func() { game3NearbyServers = old; C.dword_5d4594_1307720 = oldCount }()
	game3NearbyServers[0] = (*C.nox_gui_server_ent_t)(p)
	C.dword_5d4594_1307720 = 1
	return uintptr(p), unsafe.Pointer(C.sub_4A28C0(0)) == p && C.sub_4A28C0(-1) == nil && C.sub_4A28C0(1) == nil
}

func game3WidgetGuardedSelectionCheck() (uintptr, bool) {
	var address C.uintptr_t
	okay := C.game3_test_guarded_selection(&address) != 0
	return uintptr(address), okay
}
