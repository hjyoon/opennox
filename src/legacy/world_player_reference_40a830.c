#include <stdlib.h>
#include <stdint.h>
#include <stddef.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "noxstring.h"
#include "operators.h"
#include "common__system__team.h"
#include "world_player_reference_40a830.h"
extern void* dword_5d4594_251560;
extern obj_5D4594_2650668_t** ptr_5D4594_2650668;
extern const int ptr_5D4594_2650668_cap;
extern nox_tileDef_t nox_tile_defs_arr[176];
void* nox_server_wallAttachDoor(void*,nox_object_t*);
extern void* nox_world_player_reference_pointer(int32_t,void*,uint32_t,uint32_t,void*);
extern uint32_t nox_world_player_reference_word(int32_t,void*,uint32_t,uint32_t,void*,uint32_t);

static nox_playerInfo* wp_ref_first(void){return nox_world_player_reference_pointer(1,0,0,0,0);}
static nox_playerInfo* wp_ref_next(nox_playerInfo* p){return nox_world_player_reference_pointer(2,p,0,0,0);}
static nox_object_t* wp_ref_unit_first(void){return nox_world_player_reference_pointer(3,0,0,0,0);}
static nox_object_t* wp_ref_unit_next(const nox_object_t* p){return nox_world_player_reference_pointer(4,(void*)p,0,0,0);}
static nox_team_t* wp_ref_team_first(void){return nox_world_player_reference_pointer(5,0,0,0,0);}
static nox_team_t* wp_ref_team_next(nox_team_t* p){return nox_world_player_reference_pointer(6,p,0,0,0);}
static void* wp_ref_wall_create(int x,int y){return nox_world_player_reference_pointer(7,0,(uint32_t)x,(uint32_t)y,0);}
static void* wp_ref_wall_attach(void* w,nox_object_t* o){return nox_world_player_reference_pointer(8,w,0,0,o);}
static void* wp_ref_armor(int type){return nox_world_player_reference_pointer(9,0,(uint32_t)type,0,0);}
static int wp_ref_contains(nox_object_team_t* p,unsigned char id){return (int)nox_world_player_reference_word(1,p,id,0,0,0);}
static int wp_ref_engine(int v){return (int)nox_world_player_reference_word(2,0,v,0,0,0);}
static int wp_ref_gameplay(int v){return (int)nox_world_player_reference_word(3,0,v,0,0,0);}
static uint32_t wp_ref_flags(uint32_t v){return nox_world_player_reference_word(4,0,v,0,0,0);}
static uint32_t wp_ref_flags_value(void){return nox_world_player_reference_word(5,0,0,0,0,0);}
static int wp_ref_timer_status(void){return (int)nox_world_player_reference_word(6,0,0,0,0,0);}
static unsigned char wp_ref_mode(short mode){return (unsigned char)nox_world_player_reference_word(7,0,(uint16_t)mode,0,0,0);}
static long long wp_ref_timer_start(void){nox_world_player_reference_word(8,0,0,0,0,0);return 0;}
static int wp_ref_timer_set(int v){return (int)nox_world_player_reference_word(9,0,(uint32_t)v,0,0,0);}
static int wp_ref_send(int recipient,const void* payload,int len,nox_object_t* related,int remove){return (int)nox_world_player_reference_word(10,(void*)payload,(uint32_t)recipient,(uint32_t)remove,related,(uint32_t)len);}
static int wp_ref_lesson(nox_object_t* p){nox_world_player_reference_word(11,p,0,0,0,0);return 0;}
static void wp_ref_mark(int index,nox_object_t* p,unsigned int flags){nox_world_player_reference_word(12,p,(uint32_t)index,flags,0,0);}
static void wp_ref_unmark(int index,nox_object_t* p,unsigned int flags){nox_world_player_reference_word(13,p,(uint32_t)index,flags,0,0);}

// Private prototype for the one later-defined callee in the exact include.
int world_old_nox_xxx_netReportPlayerStatus_417630(nox_playerInfo*);

// Service replacements belong only to these private reference copies.
// The exact prior definition include itself is never rewritten.
#include "world_player_reference_renames_40a830.h"
#define nox_common_playerInfoGetFirst_416EA0 wp_ref_first
#define nox_common_playerInfoGetNext_416EE0 wp_ref_next
#define nox_xxx_getFirstPlayerUnit_4DA7C0 wp_ref_unit_first
#define nox_xxx_getNextPlayerUnit_4DA7F0 wp_ref_unit_next
#define nox_server_teamFirst_418B10 wp_ref_team_first
#define nox_server_teamNext_418B60 wp_ref_team_next
#define nox_xxx_teamCompare2_419180 wp_ref_contains
#define nox_common_getEngineFlag wp_ref_engine
#define nox_xxx_CheckGameplayFlags_417DA0 wp_ref_gameplay
#define nox_common_gameFlags_check_40A5C0 wp_ref_flags
#define nox_common_gameFlags_getVal_40A5B0 wp_ref_flags_value
#define sub_40A220 wp_ref_timer_status
#define sub_40A180 wp_ref_mode
#define sub_40A250 wp_ref_timer_start
#define sub_40A1F0 wp_ref_timer_set
#define nox_xxx_netSendPacket1_4E5390 wp_ref_send
#define nox_xxx_netReportLesson_4D8EF0 wp_ref_lesson
#define nox_xxx_netMarkMinimapObject_417190 wp_ref_mark
#define nox_xxx_netUnmarkMinimapObj_417300 wp_ref_unmark
#define nox_xxx_wallCreateAt_410250 wp_ref_wall_create
#define nox_server_wallAttachDoor wp_ref_wall_attach
#define nox_xxx_equipClothFindDefByTT_413270 wp_ref_armor
#include "world_player_preserved_reference_40a830.inc"

// Explicit public calls below cannot be rewritten to private copies.
#undef nox_xxx_countNonEliminatedPlayersInTeam_40A830
#undef nox_xxx_gamePlayIsAnyPlayers_40A8A0
#undef nox_xxx_cliCanTalkMB_4100F0
#undef nox_xxx_doorAttachWall_410360
#undef nox_xxx_wallSecretBlock_410760
#undef nox_xxx_wallSecretGetFirstWall_410780
#undef nox_xxx_wallSecretNext_410790
#undef nox_xxx_tileAlloc_410F60_init
#undef nox_xxx_tileFree_410FC0_free
#undef nox_xxx_tileNFromPoint_411160
#undef nox_xxx_mapTileAllowTeleport_411A90
#undef nox_xxx_modifNext_4133C0
#undef nox_xxx_checkGameFlagPause_413A50
#undef nox_xxx_itemApplyDefendEffect_415C00
#undef nox_xxx_cliGamedataGet_416590
#undef nox_xxx_cliSetSettingsAcquired_4169D0
#undef nox_xxx_playerForceSendLessons_416E50
#undef nox_xxx_playerByName_4170D0
#undef nox_xxx_netUnmarkMinimapSpec_417470
#undef nox_xxx_netMarkMinimapForAll_4174B0
#undef nox_xxx_playerUnsetStatus_417530
#undef nox_xxx_sendAllClientStatus_4175C0
#undef nox_xxx_netReportPlayerStatus_417630

uint32_t world_player_fixture_word(int reference,int entry,void* p,uint32_t arg){
#define WP_CALL(name,...) (reference?world_old_##name(__VA_ARGS__):name(__VA_ARGS__))
 switch(entry){
 case 0:return (uint32_t)WP_CALL(nox_xxx_countNonEliminatedPlayersInTeam_40A830,p);
 case 1:return (uint32_t)WP_CALL(nox_xxx_gamePlayIsAnyPlayers_40A8A0);
 case 2:return (uint32_t)WP_CALL(nox_xxx_cliCanTalkMB_4100F0,p);
 case 3:return (uint32_t)WP_CALL(nox_xxx_tileAlloc_410F60_init);
 case 4:WP_CALL(nox_xxx_tileFree_410FC0_free);return 0;
 case 5:return (uint32_t)WP_CALL(nox_xxx_tileNFromPoint_411160,p);
 case 6:return (uint32_t)WP_CALL(nox_xxx_mapTileAllowTeleport_411A90,p);
 case 7:return (uint32_t)WP_CALL(nox_xxx_checkGameFlagPause_413A50);
 case 8:return (uint32_t)WP_CALL(nox_xxx_cliSetSettingsAcquired_4169D0,(int32_t)arg);
 case 9:return (uint8_t)WP_CALL(nox_xxx_playerUnsetStatus_417530,p,(int32_t)arg);
 case 10:return (uint32_t)WP_CALL(nox_xxx_netReportPlayerStatus_417630,p);
 default:abort();
 }
}
void* world_player_fixture_pointer(int reference,int entry,void* p,void* q,int32_t a,int32_t b){
 switch(entry){
 case 0:return WP_CALL(nox_xxx_doorAttachWall_410360,p,a,b);
 case 1:return WP_CALL(nox_xxx_wallSecretBlock_410760,p);
 case 2:return WP_CALL(nox_xxx_wallSecretGetFirstWall_410780);
 case 3:return WP_CALL(nox_xxx_wallSecretNext_410790,p);
 case 4:return WP_CALL(nox_xxx_modifNext_4133C0,p);
 case 5:return WP_CALL(nox_xxx_cliGamedataGet_416590,a);
 case 6:return WP_CALL(nox_xxx_playerForceSendLessons_416E50,a);
 case 7:return WP_CALL(nox_xxx_playerByName_4170D0,p);
 case 8:return WP_CALL(nox_xxx_netUnmarkMinimapSpec_417470,p,a);
 case 9:return WP_CALL(nox_xxx_netMarkMinimapForAll_4174B0,p,a);
 case 10:return WP_CALL(nox_xxx_sendAllClientStatus_4175C0,p);
 default:abort();
 }
}
double world_player_fixture_armor(int reference,void* p){return reference?world_old_nox_xxx_itemApplyDefendEffect_415C00(p):nox_xxx_itemApplyDefendEffect_415C00(p);}
uintptr_t world_player_fixture_layout(int f){switch(f){
case 0:return sizeof(nox_playerInfo);case 1:return offsetof(nox_playerInfo,playerUnit);case 2:return offsetof(nox_playerInfo,field_3680);
case 3:return offsetof(nox_playerInfo,lessons);case 4:return offsetof(nox_playerInfo,field_2140);case 5:return offsetof(nox_playerInfo,name_final);
case 6:return offsetof(nox_object_t,data_update);case 7:return offsetof(nox_object_t,field_12);case 8:return offsetof(obj_412ae0_t,field_34);
case 9:return sizeof(obj_412ae0_t);case 10:return sizeof(obj_5D4594_2650668_t);case 11:return offsetof(obj_5D4594_2650668_t,field_5);case 12:return offsetof(obj_5D4594_2650668_t,field_10);case 13:return sizeof(nox_tileDef_t);default:abort();}}
void world_player_fixture_tile_owner(void* p,int n){
 // Capacity belongs to GAME2_1.c and is immutable; only replace the row owner.
 if(n!=ptr_5D4594_2650668_cap) abort();
 ptr_5D4594_2650668=p;
}
void world_player_fixture_modifier_next(void* p,void* next){((obj_412ae0_t*)p)->field_34=next;}
static void* wp_seen_effect;static nox_object_t* wp_seen_item;static int wp_seen_zeros;
static void wp_defend(void* e,nox_object_t* item,uintptr_t zero,nox_object_t* same,uintptr_t zero2,float* value){wp_seen_effect=e;wp_seen_item=item;wp_seen_zeros=(zero==0&&zero2==0&&item==same);*value=*value*2.0f+0.25f;}
void* world_player_fixture_defend_callback(void){return (void*)wp_defend;}
int world_player_fixture_defend_seen(void* e,void* item){return wp_seen_effect==e&&wp_seen_item==item&&wp_seen_zeros;}
