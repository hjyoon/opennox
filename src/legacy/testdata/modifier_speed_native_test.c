#include <assert.h>
#include <fenv.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#undef _Static_assert
_Static_assert(sizeof(void*) == 8, "native pointers");
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME4.h"
static struct effect_frame { uint64_t before[4]; float coefficient; float unused_disengage; void* engage; uint64_t after[4]; } effect, remaining;
static nox_object_t owner,item;
static nox_player_update_data_t update;
static nox_playerInfo player;
static nox_modifier_attrs_t attrs;
static int reports,sounds,float_getters,function_getters,expected_sound;
static uint32_t packet_bits;
float nox_modifier_effect_getEngageFloat(void* p) {assert(p==&effect);float_getters++;return effect.coefficient;}
void* nox_modifier_effect_getEngageFunc(void* p) {assert(p==&remaining);function_getters++;return remaining.engage;}
int nox_xxx_netReportStatsSpeed_4D9360(int index,nox_object_t* p,char kind,int value) {
 uint32_t actual;memcpy(&actual,&owner.float_138,4);assert(index==player.playerInd&&p==&owner&&kind==0&&(uint32_t)value==actual&&reports==0&&sounds==0&&float_getters==1);packet_bits=(uint32_t)value;reports++;return 0x2468;
}
void nox_xxx_aud_501960(int sound,nox_object_t* p,int kind,int code) {assert(sound==expected_sound&&p==&owner&&kind==0&&code==0&&reports==1&&sounds==0);sounds++;}
void sub_4DFB50(void* a,nox_object_t* b,const nox_object_t* c) {abort();}
void sub_4DFD10(void* a,nox_object_t* b,const nox_object_t* c) {abort();}
void nox_xxx_buff_4DFD80(void* a,nox_object_t* b,const nox_object_t* c) {abort();}
void nox_xxx_checkPoisonProtectEnch_4DFDE0(void* a,nox_object_t* b,const nox_object_t* c) {abort();}
void sub_4E0140(void* a,nox_object_t* b,const nox_object_t* c) {abort();}
// PRODUCTION_SPEED_MATH
// PRODUCTION_INVENTORY_4DFBB0
// PRODUCTION_SPEED_ENGAGE_4DFC30
// PRODUCTION_SPEED_DISENGAGE_4DFCA0
int main(int argc,char** argv) {
 assert(argc==10);int which=atoi(argv[1]),shape=atoi(argv[5]),has=atoi(argv[6]),slot=atoi(argv[7]);assert((which==0||which==1)&&shape>=0&&shape<=2&&slot>=0&&slot<4);
 memset(&effect,0x5a,sizeof(effect));memset(&remaining,0xa6,sizeof(remaining));memset(&owner,0x69,sizeof(owner));memset(&item,0x96,sizeof(item));memset(&update,0x3c,sizeof(update));memset(&player,0xc3,sizeof(player));memset(&attrs,0,sizeof(attrs));
 uint32_t value=(uint32_t)strtoul(argv[2],NULL,10),coefficient=(uint32_t)strtoul(argv[3],NULL,10),flags=(uint32_t)strtoul(argv[4],NULL,10);memcpy(&owner.float_138,&value,4);memcpy(&effect.coefficient,&coefficient,4);uint32_t unused=(uint32_t)strtoul(argv[9],NULL,10);memcpy(&effect.unused_disengage,&unused,4);owner.field_110=flags;owner.obj_class=shape==2?2:4;owner.data_update=&update;update.player=&player;player.playerInd=(uint8_t)strtoul(argv[8],NULL,10);owner.inv_first_item=&item;item.inv_next_item=NULL;item.obj_class=0x1000;item.obj_flags=0x100;item.init_data=&attrs;attrs.modifiers[slot]=&remaining;remaining.engage=has?(void*)nox_xxx_effectSpeedEngage_4DFC30:(void*)sub_4DFD10;expected_sound=which?60:59;
 unsigned char be[sizeof(effect)],br[sizeof(remaining)],bo[sizeof(owner)],bi[sizeof(item)],bu[sizeof(update)],bp[sizeof(player)],ba[sizeof(attrs)];memcpy(be,&effect,sizeof(effect));memcpy(br,&remaining,sizeof(remaining));memcpy(bo,&owner,sizeof(owner));memcpy(bi,&item,sizeof(item));memcpy(bu,&update,sizeof(update));memcpy(bp,&player,sizeof(player));memcpy(ba,&attrs,sizeof(attrs));
 uintptr_t ptrs[]={(uintptr_t)&effect,(uintptr_t)&remaining,(uintptr_t)&owner,(uintptr_t)&item,(uintptr_t)&update,(uintptr_t)&player,(uintptr_t)&attrs,(uintptr_t)&owner.float_138,(uintptr_t)remaining.engage};for(int i=0;i<9;i++)assert(ptrs[i]>UINT32_MAX);
 int rounding=fegetround();nox_object_t* arg=shape==1?NULL:&owner;if(which)nox_xxx_effectSpeedDisengage_4DFCA0(&effect,arg,&item);else nox_xxx_effectSpeedEngage_4DFC30(&effect,arg,&item);
 assert(fegetround()==rounding);memcpy(&value,&owner.float_138,4);flags=owner.field_110;int active=shape==0;assert(reports==active&&sounds==active&&float_getters==active);assert(function_getters==(active&&which?1:0));memcpy(&owner.float_138,bo+offsetof(nox_object_t,float_138),4);memcpy(&owner.field_110,bo+offsetof(nox_object_t,field_110),sizeof(owner.field_110));
 assert(!memcmp(be,&effect,sizeof(effect))&&!memcmp(br,&remaining,sizeof(remaining))&&!memcmp(bo,&owner,sizeof(owner))&&!memcmp(bi,&item,sizeof(item))&&!memcmp(bu,&update,sizeof(update))&&!memcmp(bp,&player,sizeof(player))&&!memcmp(ba,&attrs,sizeof(attrs)));
 printf("{\"value_bits\":%u,\"flags_word\":%u,\"reports\":%d,\"sounds\":%d,\"packet_bits\":%u,\"sound\":%d,\"float_getters\":%d,\"function_getters\":%d,\"rounding_mode_unchanged\":true,\"high_native_pointers\":true,\"nonoutput_canaries_unchanged\":true}\n",value,flags,reports,sounds,packet_bits,sounds?expected_sound:0,float_getters,function_getters);return 0;
}
