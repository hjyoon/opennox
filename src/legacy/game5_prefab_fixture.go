package legacy

/*
#include "native_fixture_alloc.h"
#include "GAME4_2.h"
#include "GAME5.h"
#include "mapgen_legacy_ptr.h"
#include "memmap.h"

typedef struct {
 int ok, count, total;
 uintptr_t owners[5];
 uintptr_t neighbors[5][4];
 int32_t geometry[3][5];
} game5_prefab_result;
static game5_prefab_result game5_prefab_probe(int entry,int side,int ax,int ay,int bx,int by,int cx,int cy,int blocked) {
 game5_prefab_result out={0};
 uint8_t* theme=nox_native_fixture_calloc_high(1,0x45c);
 if(!theme)return out;
 uint32_t saved[4];const uint32_t opposite[4]={1,0,3,2};
 for(int i=0;i<4;i++){saved[i]=*getMemU32Ptr(0x587000,254952+4*i);*getMemU32Ptr(0x587000,254952+4*i)=opposite[i];}
 *(uint32_t*)(theme+68)=32;
 if(!sub_520EA0(theme))goto finish;
 uint8_t* first=(uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2,2);
 uint8_t* second=(uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2,2);
 if(!first||!second){if(first)sub_521A10(first);if(second)sub_521A10(second);goto free_grid;}
 float2 pos={0,0};nox_xxx_mapGenSetRoomPos_521880((uint32_t*)first,&pos);nox_xxx_mapGenAddNewRoom_521730((uint32_t*)first);
 pos.field_0=(float)((double)cx*32.526913);pos.field_4=(float)((double)cy*32.526913);
 nox_xxx_mapGenSetRoomPos_521880((uint32_t*)second,&pos);nox_xxx_mapGenAddNewRoom_521730((uint32_t*)second);
 if(blocked){uint8_t* obstacle=(uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(1,1);pos.field_0=0;pos.field_4=(float)(-5.0*32.526913);nox_xxx_mapGenSetRoomPos_521880((uint32_t*)obstacle,&pos);nox_xxx_mapGenAddNewRoom_521730((uint32_t*)obstacle);}
 out.owners[0]=(uintptr_t)first;out.owners[1]=(uintptr_t)second;
 int32_t a[2]={ax,ay},b[2]={bx,by};
 switch(entry){
 case 0:out.ok=sub_54B810((uintptr_t)first,(uintptr_t)second,a,(int2*)b,1);break;
 case 1:out.ok=sub_54BB20((uintptr_t)first,(uintptr_t)second,a,(uint32_t*)b,1);break;
 case 2:out.ok=sub_54BD90((uintptr_t)first,(uintptr_t)second,a,b,1);break;
 case 3:out.ok=sub_54BF20((uintptr_t)first,(uintptr_t)second,a,b,1);break;
 case 4:{int32_t prefab[40]={0};prefab[4*side+20]=ax;prefab[4*side+21]=ay;prefab[4*side+22]=1;prefab[37]=(int32_t)nox_mapgenLegacyPtrRegister(first);out.ok=sub_54B2D0((uint8_t*)prefab,side,second);break;}
 }
 out.count=*getMemU32Ptr(0x5D4594,2491608);
 if(out.ok){for(int i=0;i<out.count&&i<3;i++){uint8_t* room=nox_mapgenLegacyPtrResolve(*getMemU32Ptr(0x5D4594,2491612+4*i));out.owners[2+i]=(uintptr_t)room;for(int k=0;k<5;k++)out.geometry[i][k]=((int32_t*)room)[k];}}
 for(uint8_t* room=nox_xxx_mapGenGetTopRoom_521710();room;room=nox_mapgenRoomNextNative_521720(room))out.total++;
 for(int i=0;i<5;i++){uint8_t* room=(uint8_t*)out.owners[i];if(!room)continue;for(int d=0;d<4;d++)out.neighbors[i][d]=(uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room+88+32*d));}
 nox_xxx_mapGenFreeTopRoom_521A40();
free_grid:sub_520F80();
finish:
 for(int i=0;i<4;i++)*getMemU32Ptr(0x587000,254952+4*i)=saved[i];
 nox_native_fixture_free_high(theme);
 return out;
}
*/
import "C"

type game5PrefabResult struct {
	ok           bool
	count, total int
	owners       [5]uintptr
	neighbors    [5][4]uintptr
	geometry     [3][5]int32
}

func game5PrefabProbe(entry, side int, a, b, c [2]int32, blocked bool) game5PrefabResult {
	restore := mapgenHighFixtureMemory()
	defer restore()
	raw := C.game5_prefab_probe(C.int(entry), C.int(side), C.int(a[0]), C.int(a[1]), C.int(b[0]), C.int(b[1]), C.int(c[0]), C.int(c[1]), C.int(bool2int(blocked)))
	out := game5PrefabResult{ok: raw.ok != 0, count: int(raw.count), total: int(raw.total)}
	for i := range out.owners {
		out.owners[i] = uintptr(raw.owners[i])
		for d := range out.neighbors[i] {
			out.neighbors[i][d] = uintptr(raw.neighbors[i][d])
		}
	}
	for i := range out.geometry {
		for j := range out.geometry[i] {
			out.geometry[i][j] = int32(raw.geometry[i][j])
		}
	}
	return out
}
