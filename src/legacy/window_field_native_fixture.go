package legacy

/*
#include "GAME2_1.h"
typedef nox_window*(*window_field_offset_C)(nox_window*,int,int);
typedef int(*window_field_image_C)(nox_window*,nox_video_bag_image_t*);
typedef void*(*window_field_font_C)(nox_window*);
_Static_assert(_Generic(&nox_xxx_wndSetOffsetMB_46AE40,window_field_offset_C:1,default:0),"native window offset return ABI");
_Static_assert(_Generic(&nox_xxx_wndSetIcon_46AE60,window_field_image_C:1,default:0),"window background image ABI");
_Static_assert(_Generic(&nox_xxx_wndSetIconLit_46AEA0,window_field_image_C:1,default:0),"window highlight image ABI");
_Static_assert(_Generic(&sub_46AEC0,window_field_image_C:1,default:0),"window selected image ABI");
_Static_assert(_Generic(&sub_46AF40,window_field_font_C:1,default:0),"native font return ABI");
static window_field_offset_C volatile window_field_offset_callback_C=nox_xxx_wndSetOffsetMB_46AE40;
static window_field_image_C volatile window_field_image_callbacks_C[]={nox_xxx_wndSetIcon_46AE60,nox_xxx_wndSetIconLit_46AEA0,sub_46AEC0};
static window_field_font_C volatile window_field_font_callback_C=sub_46AF40;
typedef struct window_field_result_C {uint32_t word;void* pointer;} window_field_result_C;
static window_field_result_C window_field_call_C(unsigned route,unsigned i,nox_window*win,int x,int y,nox_video_bag_image_t*image){
 window_field_result_C r={0,NULL};
 if(route==1){
  if(i==0){window_field_offset_C f=window_field_offset_callback_C;r.pointer=f(win,x,y);}
  else if(i==4){window_field_font_C f=window_field_font_callback_C;r.pointer=f(win);}
  else if(i>=1&&i<=3){window_field_image_C f=window_field_image_callbacks_C[i-1];r.word=(uint32_t)f(win,image);}
  else abort();
  return r;
 }
 switch(i){
 case 0:r.pointer=nox_xxx_wndSetOffsetMB_46AE40(win,x,y);break;
 case 1:r.word=(uint32_t)nox_xxx_wndSetIcon_46AE60(win,image);break;
 case 2:r.word=(uint32_t)nox_xxx_wndSetIconLit_46AEA0(win,image);break;
 case 3:r.word=(uint32_t)sub_46AEC0(win,image);break;
 case 4:r.pointer=sub_46AF40(win);break;
 default:abort();
 }
 return r;
}
static void window_field_seed_C(void*p,size_t placement,const uint32_t*q){
 unsigned char* raw=p;nox_window*win=(nox_window*)(raw+0x1000+placement);
 void*images[]={NULL,raw+0x3000+placement,raw+0x4000+placement};
 void*fonts[]={NULL,raw+0x6000+placement,raw+0x7000+placement};
 win->draw_data.img_px=(intptr_t)(int32_t)q[0];win->draw_data.img_py=(intptr_t)(int32_t)q[1];win->draw_data.field_0=q[2];
 win->draw_data.bg_image=images[q[3]];win->draw_data.hl_image=images[q[4]];win->draw_data.sel_image=images[q[5]];win->draw_data.font=fonts[q[6]];
}
static void window_field_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_window);v[2]=sizeof(nox_window_data);v[3]=offsetof(nox_window,draw_data);
 v[4]=offsetof(nox_window_data,field_0);v[5]=offsetof(nox_window_data,bg_image);v[6]=offsetof(nox_window_data,hl_image);v[7]=offsetof(nox_window_data,sel_image);
 v[8]=offsetof(nox_window_data,img_px);v[9]=offsetof(nox_window_data,img_py);v[10]=offsetof(nox_window_data,font);
 v[11]=sizeof(((nox_window_data*)0)->field_0);v[12]=sizeof(((nox_window_data*)0)->bg_image);v[13]=sizeof(((nox_window_data*)0)->hl_image);v[14]=sizeof(((nox_window_data*)0)->sel_image);
 v[15]=sizeof(((nox_window_data*)0)->img_px);v[16]=sizeof(((nox_window_data*)0)->img_py);v[17]=sizeof(((nox_window_data*)0)->font);v[18]=sizeof(int);v[19]=sizeof(nox_window*);
}
*/
import "C"

import "unsafe"

func windowFieldNativeLayoutC() (out [20]uintptr) {
	C.window_field_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func windowFieldNativeSeedC(raw unsafe.Pointer, placement int, input *[7]uint32) {
	C.window_field_seed_C(raw, C.size_t(placement), (*C.uint32_t)(unsafe.Pointer(&input[0])))
}

func windowFieldNativeCall(route, index int, owner, image unsafe.Pointer, x, y int32) (uint32, unsafe.Pointer) {
	win := (*C.nox_window)(owner)
	img := (*C.nox_video_bag_image_t)(image)
	if route != 2 {
		r := C.window_field_call_C(C.uint(route), C.uint(index), win, C.int(x), C.int(y), img)
		return uint32(r.word), r.pointer
	}
	switch index {
	case 0:
		return 0, unsafe.Pointer(nox_xxx_wndSetOffsetMB_46AE40(win, C.int(x), C.int(y)))
	case 1:
		return uint32(nox_xxx_wndSetIcon_46AE60(win, img)), nil
	case 2:
		return uint32(nox_xxx_wndSetIconLit_46AEA0(win, img)), nil
	case 3:
		return uint32(sub_46AEC0(win, img)), nil
	case 4:
		return 0, sub_46AF40(win)
	}
	panic("invalid window field entry")
}
