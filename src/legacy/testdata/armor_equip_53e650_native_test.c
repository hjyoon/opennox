#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#undef _Static_assert
_Static_assert(sizeof(void*)==8,"native pointer width");
_Static_assert(sizeof(int)==4,"original integer width");
#include "GAME1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
#include "inventory_detach_4ed0c0.h"
#include "player_check_strength_4f3180.h"
static nox_object_t owner,item,other;
static nox_player_update_data_t old_update,new_update;
static nox_playerInfo old_player,new_player;
static int class_return,strength_return,equipped,equip_flags,special_return,npc_return,replace_at,replace_kind,event_count;
static void replace(int stage){if(replace_at==stage){if(replace_kind==1)owner.data_update=&new_update;else if(replace_kind==2)old_update.player=&new_player;else assert(0);}}
static void argument(uintptr_t a){if(a==(uintptr_t)&owner)printf("\"owner\"");else if(a==(uintptr_t)&item)printf("\"item\"");else if(a==(uintptr_t)&other)printf("\"other-armor\"");else printf("%llu",(unsigned long long)a);}
static void event(const char* name,uintptr_t* a,int n,int result){printf("%s{\"service\":\"%s\",\"args\":[",event_count++?",":"",name);for(int i=0;i<n;i++){if(i)putchar(',');argument(a[i]);}printf("],\"result\":");if(result==-99)printf("\"other-armor\"");else printf("%d",result);putchar('}');}
int nox_xxx_NPCEquipArmor_53E520(nox_object_t* a,nox_object_t* b){uintptr_t v[]={(uintptr_t)a,(uintptr_t)b};event("npcEquip",v,2,npc_return);return npc_return;}
nox_object_t* nox_xxx_armorHaveSameSubclass_53E7B0(nox_object_t* a,const nox_object_t* b){replace(1);uintptr_t v[]={(uintptr_t)a,(uintptr_t)b};event("sameSubclass",v,2,equipped?-99:0);return equipped?&other:NULL;}
int nox_xxx_playerClassCanUseItem_57B3D0(nox_object_t* a,int8_t b){replace(2);uintptr_t v[]={(uintptr_t)a,(uint8_t)b};event("classUse",v,2,class_return);return class_return;}
int32_t nox_xxx_playerCheckStrength_4F3180(nox_object_t* a,nox_object_t* b){replace(3);uintptr_t v[]={(uintptr_t)a,(uintptr_t)b};event("strength",v,2,strength_return);return strength_return;}
int sub_53E430(nox_object_t* a,nox_object_t* b,int c,int d){replace(4);uintptr_t v[]={(uintptr_t)a,(uintptr_t)b,c,d};event("dequip",v,4,0);return 0;}
int nox_xxx_unitArmorInventoryEquipFlags_415C70(nox_object_t* a){replace(5);uintptr_t v[]={(uintptr_t)a};event("armorFlags",v,1,equip_flags);return equip_flags;}
uint32_t* nox_xxx_netReportEquip_4D8540(int a,nox_object_t* b,int c){replace(6);uintptr_t v[]={a,(uintptr_t)b,c};event("report",v,3,0);return NULL;}
int sub_53E2D0(nox_object_t* a){uintptr_t v[]={(uintptr_t)a};event("special",v,1,special_return);return special_return;}
int nox_xxx_recalculateArmorVal_53E300(nox_object_t* a){uintptr_t v[]={(uintptr_t)a};event("recalculate",v,1,0);return 0;}
void nox_xxx_itemApplyEngageEffect_4F2FF0(nox_object_t* a,nox_object_t* b){uintptr_t v[]={(uintptr_t)a,(uintptr_t)b};event("engage",v,2,0);}
void sub_53E600(nox_object_t* a){uintptr_t v[]={(uintptr_t)a};event("shield",v,1,0);}
void nox_xxx_netPriMsgToPlayer_4DA2C0(nox_object_t* a,const char* b,char c){assert(a==&owner);printf("%s{\"service\":\"message\",\"args\":[\"owner\",\"%s\",%d],\"result\":0}",event_count++?",":"",b,c);}
void nox_xxx_aud_501960(int a,nox_object_t* b,int c,int d){uintptr_t v[]={a,(uintptr_t)b,c,d};event("audio",v,4,0);}
// PRODUCTION_BODY_53E650
int main(int argc,char** argv){assert(argc==14);uintptr_t ptrs[]={(uintptr_t)&owner,(uintptr_t)&item,(uintptr_t)&other,(uintptr_t)&old_update,(uintptr_t)&new_update,(uintptr_t)&old_player,(uintptr_t)&new_player};for(int i=0;i<7;i++)assert(ptrs[i]>UINT32_MAX);owner.obj_class=strtoul(argv[1],0,0);item.obj_class=strtoul(argv[2],0,0);item.obj_flags=strtoul(argv[3],0,0);item.obj_subclass=strtoul(argv[4],0,0);class_return=atoi(argv[5]);strength_return=atoi(argv[6]);equipped=atoi(argv[7]);equip_flags=atoi(argv[8]);special_return=atoi(argv[9]);npc_return=atoi(argv[10]);int sound=atoi(argv[11]);replace_at=atoi(argv[12]);replace_kind=atoi(argv[13]);owner.net_code=91;owner.data_update=&old_update;old_update.player=&old_player;new_update.player=&new_player;old_player.field_0=0x1000;new_player.field_0=0x2000;old_player.playerInd=3;new_player.playerInd=7;old_player.info.playerClass=1;new_player.info.playerClass=2;
nox_object_t bo=owner,bi=item,bt=other;nox_player_update_data_t bu=old_update,bn=new_update;nox_playerInfo bp=old_player,bq=new_player;printf("{\"pointer_bytes\":8,\"high_native_pointers\":true,\"nonoutput_canaries_unchanged\":true,\"events\":[");int result=nox_xxx_playerEquipArmor_53E650(&owner,&item,17,sound);bo.data_update=owner.data_update;bi.obj_flags=item.obj_flags;bu.player=old_update.player;bp.field_0=old_player.field_0;bq.field_0=new_player.field_0;assert(!memcmp(&bo,&owner,sizeof(owner))&&!memcmp(&bi,&item,sizeof(item))&&!memcmp(&bt,&other,sizeof(other))&&!memcmp(&bu,&old_update,sizeof(old_update))&&!memcmp(&bn,&new_update,sizeof(new_update))&&!memcmp(&bp,&old_player,sizeof(old_player))&&!memcmp(&bq,&new_player,sizeof(new_player)));printf("],\"return_value\":%d,\"old_player_flags\":%u,\"new_player_flags\":%u,\"item_flags\":%u,\"live_update\":\"%s\",\"cached_update_player\":\"%s\"}\n",result,old_player.field_0,new_player.field_0,item.obj_flags,owner.data_update==&new_update?"new":"old",old_update.player==&new_player?"new":"old");return 0;}
