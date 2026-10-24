#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#undef _Static_assert
_Static_assert(sizeof(void*) == 8, "native 64-bit callback probe");
_Static_assert(sizeof(int) == 4, "original integer width");
_Static_assert(sizeof(wchar2_t) == 2, "original wide character width");
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
#include "pickup_default_4f31e0.h"
#include "decay_511660.h"


#include "memmap.h"
static nox_object_t owner,item,other;
static nox_player_update_data_t old_update,new_update;
static int pickup_return,replace_at,state,event_count;
uint32_t dword_5d4594_2488720=103,dword_5d4594_2488724=104;
static uint32_t name_ids[2]={101,102};
uint32_t* mem_getU32Ptr(uintptr_t base,uintptr_t off) {assert(base==0x5D4594 && (off==2488712 || off==2488716));return &name_ids[(off-2488712)/4];}
int nox_xxx_getNameId_4E3AA0(char* name) {(void)name;assert(!"unrepresented cold-cache name lookup");return 0;}
static void argument(uintptr_t value) {if(value==(uintptr_t)&owner)printf("\"owner\"");else if(value==(uintptr_t)&item)printf("\"item\"");else printf("%llu",(unsigned long long)value);}
static void event(const char* name,uintptr_t* args,unsigned n,int result) {assert(event_count<64);printf("%s{\"service\":\"%s\",\"args\":[",event_count++?",":"",name);for(unsigned i=0;i<n;i++){if(i)printf(",");argument(args[i]);}printf("],\"result\":%d}",result);}
bool nox_common_gameFlags_check_40A5C0(unsigned int value) {int result=(value&2048)!=0;uintptr_t args[]={value};event("flags",args,1,result);return result;}
int sub_409F40(int x){(void)x;assert(!"unrepresented game mode");return 0;}
int32_t sub_4E7EC0(const nox_object_t* a,const nox_object_t* b){(void)a;(void)b;assert(!"unrepresented duplicate");return 0;}
int nox_xxx_playerClassCanUseItem_57B3D0(nox_object_t* a,int8_t b){(void)a;(void)b;assert(!"unrepresented class");return 0;}
void nox_xxx_netPriMsgToPlayer_4DA2C0(nox_object_t* a,const char* b,char c){(void)a;(void)b;(void)c;assert(!"unrepresented rejection");}
void nox_xxx_aud_501960(int a,nox_object_t* b,int c,int d){(void)a;(void)b;(void)c;(void)d;assert(!"unrepresented material audio");}
int32_t nox_xxx_pickupDefault_4F31E0(nox_object_t* a,nox_object_t* b,int c,int d){assert(a==&owner&&b==&item&&c==17&&d==23);if(replace_at==1)a->data_update=&new_update;uintptr_t args[]={(uintptr_t)a,(uintptr_t)b,c,d};event("pickup",args,4,pickup_return);return pickup_return;}
nox_object_t* nox_xxx_armorHaveSameSubclass_53E7B0(nox_object_t* a,const nox_object_t* b){assert(a==&owner&&b==&item);if(replace_at==2)a->data_update=&new_update;uintptr_t args[]={(uintptr_t)a,(uintptr_t)b};event("sameSubclass",args,2,0);return NULL;}
int sub_419E60(nox_object_t* a){assert(a==&owner);if(replace_at==3)a->data_update=&new_update;uintptr_t args[]={(uintptr_t)a};event("state",args,1,state);return state;}
int nox_xxx_playerEquipArmor_53E650(nox_object_t* a,nox_object_t* b,int c,int d){assert(a==&owner&&b==&item&&c==23&&d==0);uintptr_t args[]={(uintptr_t)a,(uintptr_t)b,c,d};event("equip",args,4,1);return 1;}
void nox_xxx_decay_5116F0(nox_object_t* a){assert(a==&item);uintptr_t args[]={(uintptr_t)a};event("decay",args,1,0);}
// PRODUCTION_BODY_53E7F0
int main(int argc,char** argv){assert(argc==8);uintptr_t ptrs[]={(uintptr_t)&owner,(uintptr_t)&item,(uintptr_t)&other,(uintptr_t)&old_update,(uintptr_t)&new_update};for(unsigned i=0;i<5;i++)assert(ptrs[i]>UINT32_MAX);
owner.obj_class=strtoul(argv[1],NULL,0);item.obj_subclass=strtoul(argv[2],NULL,0);item.typ_ind=105;pickup_return=atoi(argv[3]);old_update.equipped_weapon=atoi(argv[4])?&other:NULL;new_update.equipped_weapon=atoi(argv[5])?&other:NULL;replace_at=atoi(argv[6]);state=atoi(argv[7]);owner.data_update=&old_update;other.obj_subclass=0x400;
nox_object_t before_owner=owner,before_item=item;printf("{\"pointer_bytes\":8,\"high_native_pointers\":true,\"events\":[");int result=nox_xxx_pickupArmor_53E7F0(&owner,&item,17,23);before_owner.data_update=owner.data_update;assert(memcmp(&before_owner,&owner,sizeof(owner))==0&&memcmp(&before_item,&item,sizeof(item))==0);printf("],\"return_value\":%d,\"live_update\":\"%s\",\"nonoutput_canaries_unchanged\":true}\n",result,owner.data_update==&new_update?"new":"old");return 0;}
