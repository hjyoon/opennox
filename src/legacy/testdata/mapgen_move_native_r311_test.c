#include <assert.h>
#include <stdio.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_2.h"
#include "xfer_door_4f4cb0.h"

static nox_object_t* object;
static float2 input, fixed;
static int valid, door;
static unsigned fix_calls,name_calls,xfer_calls,create_calls,pending_calls;
static uint32_t dword_5d4594_3835392;
static char name[]="owned-object";
int nox_xxx_mapGenFixCoords_4D3D90(float2* from,float2* to) {
 assert(from==&input);++fix_calls;*to=fixed;return valid;
}
char* nox_xxx_getUnitName_4E39D0(nox_object_t* value) {assert(value==object);++name_calls;return name;}
int32_t nox_xxx_XFerDoor_4F4CB0(nox_object_t* value,void* context) {abort();}
void* nox_objectTypeGetXfer(char* id) {assert(id==name);++xfer_calls;return door?(void*)nox_xxx_XFerDoor_4F4CB0:NULL;}
void nox_xxx_createAt_4DAA50(nox_object_t* value,nox_object_t* owner,float x,float y) {
 assert(value==object && !owner && x==object->x && y==object->y);++create_calls;
}
void nox_xxx_unitClearPendingMB_4DB030(void) {++pending_calls;}

// COMPLETE_MAPGEN_MOVE_R311

int main(void) {
 object=malloc(sizeof(*object));assert(object);
#if UINTPTR_MAX>UINT32_MAX
 assert((uintptr_t)object>UINT32_MAX);
#endif
 const struct {float x,y,door_x,door_y;} vectors[]={
  {7.25f,-9.5f,0,0},{24,47,23,46},{-47,-24,-23,0},{0,0,0,0}};
 unsigned cases=0;
 for(unsigned i=0;i<4;++i) for(door=0;door<2;++door) for(valid=0;valid<2;++valid) {
  memset(object,0x5A,sizeof(*object));object->obj_flags=0xA5000003;object->extent=83;
  _Alignas(nox_object_t) unsigned char expected[sizeof(*object)];memcpy(expected,object,sizeof(expected));
  input=(float2){123,456};fixed=(float2){vectors[i].x,vectors[i].y};
  dword_5d4594_3835392=i==3?UINT32_MAX:123;
  uint32_t old=dword_5d4594_3835392;
  fix_calls=name_calls=xfer_calls=create_calls=pending_calls=0;
  if(valid) {
   nox_object_t* want=(nox_object_t*)expected;
   want->extent=old;want->obj_flags|=0x1000000;
   want->x=door?vectors[i].door_x:fixed.field_0;want->y=door?vectors[i].door_y:fixed.field_4;
  }
  assert(nox_xxx_mapGenMoveObject_527A10((float*)object,&input)==(float*)object);
  assert(!memcmp(object,expected,sizeof(expected)) && input.field_0==123 && input.field_4==456);
  assert(fix_calls==1 && name_calls==(unsigned)valid && xfer_calls==(unsigned)valid && create_calls==(unsigned)valid && pending_calls==(unsigned)valid);
  assert(dword_5d4594_3835392==(valid?old+1:old));++cases;
 }
 for(valid=0;valid<2;++valid) {
  fix_calls=name_calls=xfer_calls=create_calls=pending_calls=0;
  assert(nox_xxx_mapGenMoveObject_527A10(NULL,&input)==NULL);
  assert(fix_calls==1 && !name_calls && !xfer_calls && !create_calls && !pending_calls);++cases;
 }
 free(object);printf("Mapgen move native R311: %u original-domain cases PASS\n",cases);return 0;
}
