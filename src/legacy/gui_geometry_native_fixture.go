package legacy

/*
#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__draw__parse__parse.h"
#include "client__gui__window.h"
#include "client__video__draw_common.h"
#include "operators.h"
#include "client__draw__doordraw.h"
typedef nox_window*(*gui_geometry_style_C)(nox_window*,int);
typedef int(*gui_geometry_rect_C)(int2*,int4*);
typedef int(*gui_geometry_float_rect_C)(float2*,float4*);
typedef int(*gui_geometry_screen_C)(nox_draw_viewport_t*,int2*,int2*);
typedef int(*gui_geometry_world_C)(nox_draw_viewport_t*,int2*,uint32_t*);
typedef int(*gui_geometry_door_C)(nox_draw_viewport_t*,nox_drawable*);
_Static_assert(_Generic(&sub_46AE10,gui_geometry_style_C:1,default:0),"nullable native Window return ABI");
_Static_assert(_Generic(&nox_xxx_pointInRect_4281F0,gui_geometry_rect_C:1,default:0),"inclusive integer bounds ABI");
_Static_assert(_Generic(&sub_428220,gui_geometry_float_rect_C:1,default:0),"fixed32 float bounds ABI");
_Static_assert(_Generic(&sub_4739E0,gui_geometry_screen_C:1,default:0),"native viewport screen ABI");
_Static_assert(_Generic(&sub_473A10,gui_geometry_world_C:1,default:0),"native viewport world ABI");
_Static_assert(_Generic(&nox_thing_door_draw,gui_geometry_door_C:1,default:0),"actual DoorDraw native viewport ABI");
static gui_geometry_style_C volatile gui_geometry_style_callback_C=sub_46AE10;
static gui_geometry_rect_C volatile gui_geometry_rect_callback_C=nox_xxx_pointInRect_4281F0;
static gui_geometry_float_rect_C volatile gui_geometry_float_callback_C=sub_428220;
static gui_geometry_screen_C volatile gui_geometry_screen_callback_C=sub_4739E0;
static gui_geometry_world_C volatile gui_geometry_world_callback_C=sub_473A10;
typedef struct gui_geometry_result_C {uint32_t word;void*pointer;}gui_geometry_result_C;
static gui_geometry_result_C gui_geometry_call_C(unsigned route,unsigned index,nox_window*win,int enabled,nox_draw_viewport_t*vp,int2*point,void*bounds,void*out){
 gui_geometry_result_C r={0,NULL};
 if(route==1){switch(index){
 case 0:{gui_geometry_style_C f=gui_geometry_style_callback_C;r.pointer=f(win,enabled);break;}
 case 1:{gui_geometry_rect_C f=gui_geometry_rect_callback_C;r.word=f(point,(int4*)bounds);break;}
 case 2:{gui_geometry_float_rect_C f=gui_geometry_float_callback_C;r.word=f((float2*)point,(float4*)bounds);break;}
 case 3:{gui_geometry_screen_C f=gui_geometry_screen_callback_C;r.word=f(vp,point,(int2*)out);break;}
 case 4:{gui_geometry_world_C f=gui_geometry_world_callback_C;r.word=f(vp,point,(uint32_t*)out);break;}
 default:abort();}return r;}
 switch(index){
 case 0:r.pointer=sub_46AE10(win,enabled);break;
 case 1:r.word=nox_xxx_pointInRect_4281F0(point,(int4*)bounds);break;
 case 2:r.word=sub_428220((float2*)point,(float4*)bounds);break;
 case 3:r.word=sub_4739E0(vp,point,(int2*)out);break;
 case 4:r.word=sub_473A10(vp,point,(uint32_t*)out);break;
 default:abort();}return r;
}
static void gui_geometry_seed_C(void*p,size_t placement,const uint32_t*q,size_t point,size_t bounds,size_t out){
 unsigned char*raw=p;nox_window*win=(nox_window*)(raw+placement+0x5000);nox_draw_viewport_t*vp=(nox_draw_viewport_t*)(raw+placement+0x1000);
 win->draw_data.field_0=q[3];
 vp->x1=(int32_t)q[10];vp->y1=(int32_t)q[11];vp->x2=(int32_t)q[12];vp->y2=(int32_t)q[13];
 vp->field_4=(int32_t)q[14];vp->field_5=(int32_t)q[15];vp->field_6=(int32_t)q[16];vp->field_7=(int32_t)q[17];
 vp->width=(int32_t)q[18];vp->height=(int32_t)q[19];vp->field_10=q[20];vp->field_11=q[21];vp->field_12=(int32_t)q[22];
 memcpy(raw+placement+point,q+4,8);memcpy(raw+placement+bounds,q+6,16);memcpy(raw+placement+out,q+23,8);
}
static void gui_geometry_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_window);v[2]=offsetof(nox_window,draw_data.field_0);v[3]=offsetof(nox_window,flags);v[4]=sizeof(nox_draw_viewport_t);
 const size_t offsets[]={offsetof(nox_draw_viewport_t,x1),offsetof(nox_draw_viewport_t,y1),offsetof(nox_draw_viewport_t,x2),offsetof(nox_draw_viewport_t,y2),offsetof(nox_draw_viewport_t,field_4),offsetof(nox_draw_viewport_t,field_5),offsetof(nox_draw_viewport_t,field_6),offsetof(nox_draw_viewport_t,field_7),offsetof(nox_draw_viewport_t,width),offsetof(nox_draw_viewport_t,height),offsetof(nox_draw_viewport_t,field_10),offsetof(nox_draw_viewport_t,field_11),offsetof(nox_draw_viewport_t,field_12)};
 for(unsigned j=0;j<13;j++)v[5+j]=offsets[j];
 v[18]=sizeof(int2);v[19]=offsetof(int2,field_4);v[20]=sizeof(int4);v[21]=sizeof(float2);v[22]=sizeof(float4);v[23]=sizeof(((nox_window*)0)->draw_data.field_0);v[24]=sizeof(uint32_t);
}
static uintptr_t gui_geometry_door_image_C;static int gui_geometry_door_x_C,gui_geometry_door_y_C,gui_geometry_door_calls_C;static unsigned char gui_geometry_door_light_C[12];
static void gui_geometry_capture_image_C(nox_video_bag_image_t*image,int x,int y){gui_geometry_door_image_C=(uintptr_t)image;gui_geometry_door_x_C=x;gui_geometry_door_y_C=y;gui_geometry_door_calls_C++;}
#define nox_thing_door_draw gui_geometry_bounded_door_draw_C
#define nox_things_door_draw_parse gui_geometry_bounded_unused_door_parse_C
#define nox_xxx_drawObject_4C4770_draw(a,b,c) ((void)0)
#define nox_common_gameFlags_check_40A5C0(a) (1)
#define nox_xxx_gLoadImg_42F970(a) (abort(),(nox_video_bag_image_t*)0)
#define sub_469920(a) ((char*)gui_geometry_door_light_C)
#define sub_4345F0(a) ((void)0)
#define nox_draw_setColorMultAndIntensityRGB_433CD0(a,b,c) ((void)0)
#define nox_client_drawImageAt_47D2C0(image,x,y) gui_geometry_capture_image_C((image),(x),(y))
#define nox_xxx_spriteLoadStaticRandomData_44C000(a,b) ((nox_static_random_draw_data_t*)0)
#include "testdata/retired-game-sources/client__draw__doordraw.c.txt"
static gui_geometry_door_C volatile gui_geometry_door_callback_C=gui_geometry_bounded_door_draw_C;
typedef struct gui_geometry_door_result_C{void*image;int x,y,calls,result;}gui_geometry_door_result_C;
static void gui_geometry_door_seed_C(void*p,size_t placement,int frame,int metal){
 unsigned char*raw=p;nox_drawable*dr=(nox_drawable*)(raw+placement+0x6000);memset(dr,0,sizeof(*dr));dr->pos.x=143;dr->pos.y=279;dr->field_108_1=1;dr->field_108_2=metal;dr->field_74_4=frame;
 dr->field_109=(nox_video_bag_image_t*)(raw+placement+0x7100);dr->field_110=(nox_video_bag_image_t*)(raw+placement+0x7000);dr->field_111=(nox_video_bag_image_t*)(raw+placement+0x7200);dr->field_112=(nox_video_bag_image_t*)(raw+placement+0x7300);
}
static gui_geometry_door_result_C gui_geometry_door_call_C(unsigned route,void*p,size_t placement){
 unsigned char*raw=p;nox_draw_viewport_t*vp=(nox_draw_viewport_t*)(raw+placement+0x1000);nox_drawable*dr=(nox_drawable*)(raw+placement+0x6000);gui_geometry_door_calls_C=0;
 int result;if(route){gui_geometry_door_C f=gui_geometry_door_callback_C;result=f(vp,dr);}else result=gui_geometry_bounded_door_draw_C(vp,dr);
 gui_geometry_door_result_C r={(void*)gui_geometry_door_image_C,gui_geometry_door_x_C,gui_geometry_door_y_C,gui_geometry_door_calls_C,result};return r;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/ntype"
)

func guiGeometryNativeLayoutC() (out [25]uintptr) {
	C.gui_geometry_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func guiGeometryNativeSeedC(raw unsafe.Pointer, placement int, input *[25]uint32, point, bounds, out int) {
	C.gui_geometry_seed_C(raw, C.size_t(placement), (*C.uint32_t)(unsafe.Pointer(&input[0])), C.size_t(point), C.size_t(bounds), C.size_t(out))
}

func guiGeometryNativeCall(route, index int, win unsafe.Pointer, enabled int32, viewport, point, bounds, out unsafe.Pointer) (uint32, unsafe.Pointer) {
	if route != 2 {
		r := C.gui_geometry_call_C(C.uint(route), C.uint(index), (*C.nox_window)(win), C.int(enabled), (*C.nox_draw_viewport_t)(viewport), (*C.int2)(point), bounds, out)
		return uint32(r.word), r.pointer
	}
	switch index {
	case 0:
		return 0, guiGeometryStyle46AE10((*gui.Window)(win), enabled).C()
	case 1:
		return uint32(guiGeometryPointInRect4281F0((*ntype.Point32)(point), (*[4]int32)(bounds))), nil
	case 2:
		return uint32(guiGeometryFloatInRect428220((*[2]float32)(point), (*[4]float32)(bounds))), nil
	case 3:
		return uint32(guiGeometryToScreen4739E0((*noxrender.Viewport)(viewport), (*ntype.Point32)(point), (*ntype.Point32)(out))), nil
	case 4:
		return uint32(guiGeometryToWorld473A10((*noxrender.Viewport)(viewport), (*ntype.Point32)(point), (*ntype.Point32)(out))), nil
	}
	panic("invalid GUI geometry entry")
}

func guiGeometryNativeStyleFacade(win unsafe.Pointer, enabled int32) int32 {
	return g2qStyle((*C.nox_window)(win), enabled)
}

func guiGeometryNativeDoorSeed(raw unsafe.Pointer, placement, frame, metal int) {
	C.gui_geometry_door_seed_C(raw, C.size_t(placement), C.int(frame), C.int(metal))
}

func guiGeometryNativeDoorCall(route int, raw unsafe.Pointer, placement int) (unsafe.Pointer, int32, int32, int, int) {
	r := C.gui_geometry_door_call_C(C.uint(route), raw, C.size_t(placement))
	return r.image, int32(r.x), int32(r.y), int(r.calls), int(r.result)
}
