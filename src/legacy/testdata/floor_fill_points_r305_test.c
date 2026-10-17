// R305 source-bound callback regression. Literal event tables, inputs,
// callback checks, native-token collision and allocation canaries are identical
// to the independently reviewed 15-case private original-instruction fixture.
// The runner inserts the actual production fill typedef/resolver and three
// complete function bodies. It links the actual production registry separately.
// No GAME.EXE execution or full classic-generation equivalence is claimed.
#include <stdbool.h>
#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "mapgen_legacy_ptr.h"
// PRODUCTION_FILL_POINT_TYPES_R305
int sub_524950(int,int,float*,int*);
int sub_5249C0(int,int,float*,int*);
void sub_524B50(int,int,float*,int*);
typedef struct {int kind; uint32_t x,y; int a,b,result;} Event;
typedef struct {const char* name; int fn; uint32_t x,y; int w,h; int null_fill; const Event* events; size_t count;} Case;
typedef nox_mapgen_decor_fill_legacy Fill;
static const Case* current;
static size_t observed;
static const int wanted_layer = 73;
static uint32_t bits(float f) {uint32_t v;memcpy(&v,&f,4);return v;}
static float decode(uint32_t v) {float f;memcpy(&f,&v,4);return f;}
static void check(int ok,const char* message) {
 if (!ok) {fprintf(stderr,"FAIL %s event %zu: %s\n",current?current->name:"setup",observed,message);exit(86);}
}
static Event next_event(int kind,float* p,int a,int b) {
 check(observed<current->count,"unexpected callback");
 Event e=current->events[observed++];
 uint32_t x=0,y=0;
 if(p) {volatile float* q=p;x=bits(q[0]);y=bits(q[1]);}
 if (e.kind!=kind || e.x!=x || e.y!=y || e.a!=a || e.b!=b) {
  fprintf(stderr,"got kind=%d xy=%08x/%08x a/b=%d/%d, want kind=%d xy=%08x/%08x a/b=%d/%d\n",kind,x,y,a,b,e.kind,e.x,e.y,e.a,e.b);
  check(0,"callback order/value mismatch");
 }
 return e;
}
__attribute__((noinline)) int nox_xxx_tileGetDefByName_51D4D0(char* name) {
 check(!strcmp(name,"FixtureTile"),"tile lookup name changed");next_event(1,NULL,0,0);return 177;
}
__attribute__((noinline)) float* sub_5245A0(int layer,float* p,int w,int h) {
 check(layer==wanted_layer,"rectangle layer changed");next_event(2,p,w,h);return p;
}
__attribute__((noinline)) int sub_544020(char* name) {
 check(!strcmp(name,"FixtureEdge"),"edge lookup name changed");next_event(3,NULL,0,0);return 255;
}
__attribute__((noinline)) int sub_543680(float* p) {next_event(4,p,0,0);return 91;}
__attribute__((noinline)) int nox_xxx_mapGenRandFunc_526AC0(int lo,int hi) {
 Event e=next_event(5,NULL,lo,hi);check(e.result>=lo&&e.result<=hi,"invalid controlled RNG value");return e.result;
}
__attribute__((noinline)) float* sub_524610(int layer,float* p,int h) {
 check(layer==wanted_layer,"column layer changed");next_event(6,p,h,0);return p;
}
static const Event events_0[] = {{0,0,0,0,0,0}};
static const Event events_1[] = {{0,0,0,0,0,0}};
static const Event events_2[] = {{0,0,0,0,0,0}};
static const Event events_3[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),4,6,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x43250dc8),UINT32_C(0xc01ad530),0,0,0}};
static const Event events_4[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),0,2,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x42c80000),UINT32_C(0xc286f238),0,0,0}};
static const Event events_5[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x43250dc8),UINT32_C(0xc2c80000),2,1,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc286f238),4,1,0},{2,UINT32_C(0x42c80000),UINT32_C(0xc20bc8e1),6,1,0},{2,UINT32_C(0x43250dc8),UINT32_C(0x41f0dc78),2,1,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc01ad530),4,1,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x434594ab),UINT32_C(0xc1957635),0,0,0}};
static const Event events_6[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x36800000),UINT32_C(0xc2c80000),3,1,0},{2,UINT32_C(0xc2021b8e),UINT32_C(0xc286f238),5,1,0},{2,UINT32_C(0x00000000),UINT32_C(0xc286f238),3,1,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x42432958),UINT32_C(0xc286f238),0,0,0}};
static const Event events_7[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc2c80000),1,1,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x4314ca56),UINT32_C(0xc2e886e4),0,0,0}};
static const Event events_8[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc2c80000),1,1,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x4314ca56),UINT32_C(0xc2c80000),0,0,0}};
static const Event events_9[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc286f238),1,1,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc2c80000),1,1,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc20bc8e2),1,1,0},{2,UINT32_C(0x42c80000),UINT32_C(0xc286f238),1,1,0},{2,UINT32_C(0x43250dc8),UINT32_C(0xc286f238),1,1,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x4314ca56),UINT32_C(0xc24cd6aa),0,0,0}};
static const Event events_10[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc286f238),4,3,0},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,3,1},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,3,2},{2,UINT32_C(0x43250dc8),UINT32_C(0xc2c80000),1,1,0},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,3,1},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,3,2},{2,UINT32_C(0x43250dc8),UINT32_C(0x41f0dc78),1,1,0},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,2,1},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,2,2},{6,UINT32_C(0x42c80000),UINT32_C(0xc20bc8e2),1,0,0},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,2,1},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,2,1},{6,UINT32_C(0x43835139),UINT32_C(0xc286f238),1,0,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x434594ab),UINT32_C(0xc1957635),0,0,0}};
static const Event events_11[] = {{1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc286f238),1,2,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc2c80000),1,1,0},{2,UINT32_C(0x430486e4),UINT32_C(0xc20bc8e2),1,1,0},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,1,1},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,1,1},{6,UINT32_C(0x42c80000),UINT32_C(0xc286f238),1,0,0},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,1,1},{5,UINT32_C(0x00000000),UINT32_C(0x00000000),1,1,1},{6,UINT32_C(0x43250dc8),UINT32_C(0xc286f238),1,0,0},{3,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,0},{4,UINT32_C(0x4314ca56),UINT32_C(0xc20bc8e2),0,0,0}};
static const Event events_12[] = {{0,0,0,0,0,0}};
static const Event events_13[] = {{0,0,0,0,0,0}};
static const Event events_14[] = {{0,0,0,0,0,0}};
static const Case cases[] = {
{"sub_524950-null-fill",0,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,1,events_0,0},
{"sub_5249C0-null-fill",1,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,1,events_1,0},
{"sub_524B50-null-fill",2,UINT32_C(0x00000000),UINT32_C(0x00000000),0,0,1,events_2,0},
{"rectangle-center-negative-tie",0,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),4,6,0,events_3,4},
{"rectangle-zero-dimension-still-calls",0,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),0,2,0,events_4,4},
{"taper-odd-height-row-order",1,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),6,5,0,events_5,8},
{"taper-first-pass-extended-second-pass-staged",1,UINT32_C(0xc2c32956),UINT32_C(0xc2c80000),9,2,0,events_6,6},
{"taper-negative-one-height-truncates-to-zero",1,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),3,-1,0,events_7,4},
{"taper-zero-height-includes-top-row",1,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),3,0,0,events_8,4},
{"irregular-three-square-no-rng",2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),3,3,0,events_9,8},
{"irregular-rng-and-column-order",2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),6,5,0,events_10,16},
{"irregular-three-by-four-mixed-branches",2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),3,4,0,events_11,12},
{"irregular-width-two-skips-all",2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),2,3,0,events_12,0},
{"irregular-height-two-skips-all",2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),3,2,0,events_13,0},
{"irregular-negative-size-skips-all",2,UINT32_C(0x42c80000),UINT32_C(0xc2c80000),-1,3,0,events_14,0},
};
// PRODUCTION_FILL_POINT_BODIES_R305
int main(int argc,char** argv) {
 check(argc==2,"one exact case name required");
 for(size_t i=0;i<sizeof(cases)/sizeof(cases[0]);i++)if(!strcmp(cases[i].name,argv[1]))current=&cases[i];
 check(current!=NULL,"unknown case");
 struct {uint64_t before;float point[2];uint64_t after;} origin={UINT64_C(0x1122334455667788),{decode(current->x),decode(current->y)},UINT64_C(0x8877665544332211)};
 struct {uint64_t before;int dims[2];uint64_t after;} dimensions={UINT64_C(0x99aabbccddeeff00),{current->w,current->h},UINT64_C(0x00ffeeddccbbaa99)};
 struct {uint64_t before;Fill fill;uint64_t after;} *allocation=calloc(1,sizeof(*allocation));
 check(allocation!=NULL,"allocate fixture");
 allocation->before=UINT64_C(0x3141592653589793);allocation->after=UINT64_C(0x2718281828459045);
 strcpy(allocation->fill.object_name,"FixtureEdge");strcpy(allocation->fill.tile_name,"FixtureTile");allocation->fill.next_token=UINT32_C(0xd00d5678);
 unsigned char old_origin[sizeof origin],old_dims[sizeof dimensions],old_fill[sizeof(*allocation)];
 memcpy(old_origin,&origin,sizeof origin);memcpy(old_dims,&dimensions,sizeof dimensions);memcpy(old_fill,allocation,sizeof(*allocation));
 check((uintptr_t)&allocation->fill>UINT32_MAX,"actual fill must exercise >4 GiB pointer");
 // A never-dereferenced opaque alias reserves the same low bits. The real
 // registry then assigns the live native allocation a high-bit PE32 token.
 void* opaque_alias=(void*)((uintptr_t)&allocation->fill^((uintptr_t)UINT32_MAX+1));
 uint32_t alias_token=nox_mapgenLegacyPtrRegister(opaque_alias);check(alias_token!=0,"register opaque alias");
 uint32_t token=nox_mapgenLegacyPtrRegister(&allocation->fill);check(token>=UINT32_C(0xf0000000),"force high-bit legacy token");
 check(nox_mapgenLegacyPtrResolve(token)==&allocation->fill,"real resolver preserves full pointer");
 if(current->null_fill)token=0;
 float* p=current->null_fill?NULL:origin.point;int* d=current->null_fill?NULL:dimensions.dims;
 if(current->fn==0) {int ret=sub_524950(wanted_layer,(int32_t)token,p,d);check(ret==(current->null_fill?0:91),"preserve existing C return contract");}
 else if(current->fn==1) {int ret=sub_5249C0(wanted_layer,(int32_t)token,p,d);check(ret==(current->null_fill?0:91),"preserve existing C return contract");}
 else sub_524B50(wanted_layer,(int32_t)token,p,d);
 check(observed==current->count,"missing callback");
 check(!memcmp(old_origin,&origin,sizeof origin),"input coordinates or neighbor canary changed");
 check(!memcmp(old_dims,&dimensions,sizeof dimensions),"input dimensions or neighbor canary changed");
 check(!memcmp(old_fill,allocation,sizeof(*allocation)),"fill record or neighbor canary changed");
 nox_mapgenLegacyPtrForget(&allocation->fill);nox_mapgenLegacyPtrForget(opaque_alias);free(allocation);
 printf("PASS %s callbacks=%zu high-bit-token/native-pointer/canaries=ok\n",current->name,observed);return 0;
}
