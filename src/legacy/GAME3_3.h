#ifndef NOX_PORT_GAME3_3
#define NOX_PORT_GAME3_3

#include "unit_damage_clear_4ee5e0.h"
#include "unit_hp_set_max_4ee6f0.h"
#include "player_hp_init_4ee730.h"
#include "unit_get_hp_4ee780.h"
#include "unit_get_max_hp_4ee7a0.h"
#include "unit_set_max_hp_4ee7c0.h"
#include "poison_state_4ee7e0.h"
#include "player_mana_add_4eeb80.h"
#include "player_mana_sub_4eebf0.h"
#include "unit_get_old_mana_4eec80.h"
#include "player_get_max_mana_4eecb0.h"
#include "player_set_max_mana_4eecd0.h"
#include "player_mana_refresh_4eecf0.h"
#include "fixed_rng_seed_4eed30.h"
#include "ability_grant_4eed40.h"
#include "player_read_values_4eedc0.h"
#include "player_sync_level_4ef140.h"
#include "bolt_damage_4ef1e0.h"
#include "god_mode_controller_4ef500.h"
#include "fixed_rng_seed_4ef560.h"
#include "fixed_rng_seed_4ef570.h"
#include "respawn_weapon_flags_4ef580.h"
#include "player_respawn_item_4ef750.h"
#include "player_make_def_items_4ef7d0.h"
#include "spell_award_all_4efc80.h"
#include "beast_scroll_award_all_4efd80.h"
#include "warrior_ability_award_all_4efe10.h"
#include "player_unit_init_4efe80.h"
#include "player_reset_4eff10.h"
#include "monster_init_4f0040.h"
#include "grunt_init_4f0360.h"
#include "skeleton_init_4f0370.h"
#include "projectile_init_4f0380.h"
#include "spark_init_4f0390.h"
#include "frog_init_4f03b0.h"
#include "chest_init_4f0400.h"
#include "boulder_init_4f0420.h"
#include "tower_init_4f0440.h"
#include "skull_init_4f0450.h"
#include "direction_init_4f0490.h"
#include "gold_init_4f04b0.h"
#include "break_init_4f0570.h"
#include "monster_generator_init_4f0590.h"
#include "fixed_rng_seed_4f0630.h"
#include "reward_marker_activate_4f0720.h"
#include "reward_spell_book_4f09f0.h"
#include "reward_ability_book_4f0c70.h"
#include "reward_field_guide_4f0d20.h"
#include "spell_loss_eligible_4f24e0.h"
#include "field_guide_loss_eligible_4f2530.h"
#include "ability_loss_eligible_4f2570.h"
#include "player_try_equip_4f2f70.h"
#include "player_try_dequip_4f2fb0.h"
#include "item_apply_engage_4f2ff0.h"
#include "item_apply_disengage_4f3030.h"
#include "player_check_strength_4f3180.h"
#include "pickup_default_4f31e0.h"
#include "pickup_food_4f3350.h"
#include "pickup_ammo_4f3b00.h"
#include "pickup_abilitybook_4f3ce0.h"
#include "pickup_ankh_tradable_4f3dd0.h"
#include "inventory_serv_place_4f36f0.h"

#include "defs.h"
#include "default_drop_4ed290.h"
#include "drop_all_items_4eda40.h"
#include "drop_owned_crowns_4ed050.h"
#include "inventory_detach_4ed0c0.h"
#include "object_collide_noop_4e87a0.h"
#include "quest_map_buffer_4e8e50.h"
#include "random_reachable_point_4ed970.h"
#include "unit_adjust_hp_4ee460.h"

int nox_server_handler_PlayerDamage_4E17B0(int a1, int a2, int a3, int a4, int a5);
void nox_xxx_playerDecrementHPMana_4E20F0(int a1, int a2, float a3);
void nox_xxx_playerDamageItems_4E2180(int a1, int a2, int a3, int a4, float a5);
double sub_4E2220(int a1);
int sub_4E22A0(int a1, int a2, int a3, int a4, float a5, int a6);
int sub_4E2330(int a1, int a2, int a3, int a4, float a5, int a6);
int sub_4E23C0(int a1, int a2, int a3, int a4, int a5);
int sub_4E24B0(int a1, int a2, int a3, int a4, int a5);
int sub_4E24E0(int a1, int a2, int a3, int a4, int a5);
int nox_xxx_damageFlammable_4E2520(int a1, int a2, int a3, int a4, int a5);
int nox_xxx_damageBlackPowder_4E2560(int a1, int a2, int a3, int a4, int a5);
void nox_xxx_gameSetWallsDamage_4E25A0(int a1);
void nox_xxx_mapDamageUnitsAround_4E25B0(float* a1, float a2, float a3, int a4, int a5, nox_object_t* a6, nox_object_t* a7);
int nox_xxx_damageMonsterGen_4E27D0(int a1, int a2, int a3, int a4, int a5);
nox_object_t* nox_xxx_newObjectWithTypeInd_4E3450(int ind);
nox_object_t* nox_xxx_newObjectByTypeID_4E3810(char* id);
int nox_xxx_objectFreeMem_4E38A0(nox_object_t* a1);
char* nox_xxx_getUnitName_4E39D0(nox_object_t* a1);
char* nox_xxx_getUnitNameByThingType_4E3A80(int a1);
int nox_xxx_getNameId_4E3AA0(char* a1);
int nox_xxx_unitDefGetCount_4E3AC0();
int sub_4E3AD0(int a1);
int sub_4E3B80(int a1);
int nox_xxx_getUnitDefDd10_4E3BA0(int a1);
double sub_4E3CA0();
int sub_4E3CB0(float a1);
int nox_game_getQuestStage_4E3CC0();
void nox_game_setQuestStage_4E3CD0(int a1);
int nox_xxx_player_4E3CE0();
int sub_4E3D50();
short sub_4E3DD0();
void sub_4E4080(float a1);
double sub_4E40B0();
void sub_4E40C0(float a1);
double sub_4E40F0();
int sub_4E4100();
int sub_4E41B0(char* a1);
int sub_4E42C0(FILE* a1);
int sub_4E4390(FILE* a1);
FILE* sub_4E43F0(char* a1);
void nox_xxx_unitNeedSync_4E44F0(nox_object_t* a1);
int* sub_4E4500(nox_object_t* a1, int a2, int a3, int a4);
int nox_xxx_unitSetHP_4E4560(nox_object_t* obj, unsigned short amount);
int* nox_xxx_unitSetOnOff_4E4670(nox_object_t* obj, int enabled);
void nox_xxx_unitRaise_4E46F0(nox_object_t* obj, float a2);
void nox_xxx_unitUnsetXStatus_4E4780(nox_object_t* a1, unsigned int a2);
void nox_xxx_unitSetXStatus_4E4800(nox_object_t* a1p, unsigned int a2);
int* nox_xxx_servMarkObjAnimFrame_4E4880(nox_object_t* obj, int frame);
int* nox_xxx_setUnitBuffFlags_4E48F0(nox_object_t* obj, unsigned int flags);
void nox_server_setUnitBuffFlags_4E48F0(nox_object_t* obj, unsigned int flags);
uintptr_t nox_xxx_modifSetItemAttrs_4E4990(nox_object_t* obj, const nox_modifier_attrs_t* attrs);
int nox_server_setModifierAttrs_4E4990(nox_object_t* obj, nox_modifier_attrs_t* attrs, unsigned int team_base);
double nox_xxx_objectGetMass_4E4A70(const nox_object_t* obj);
uint32_t nox_xxx_objectGetBuffs_4E4A80(const nox_object_t* obj);
int* nox_xxx_setNPCColor_4E4A90(nox_object_t* obj, unsigned char index, const nox_color3_t* color);
void nox_server_setNPCColor_4E4A90(nox_object_t* obj, unsigned char index, nox_color3_t* color);
int* nox_xxx_npcSetItemEquipFlags_4E4B20(nox_object_t* obj, nox_object_t* item, int equipped);
void nox_server_npcSetItemEquipFlags_4E4B20(nox_object_t* obj, nox_object_t* item, int equipped);
uint32_t nox_xxx_objectGetNetCode_4E4C00(const nox_object_t* item);
uint16_t nox_xxx_objectGetTypeInd_4E4C10(const nox_object_t* item);
void* nox_object_getInitData_4E4C30(const nox_object_t* item);
uint32_t nox_xxx_objectGetInitDataSize_4E4C50(nox_object_t* item);
const char* nox_xxx_objectGetID_4E4C80(const nox_object_t* item);
int nox_xxx_objectHasSyncData_4E4C90(nox_object_t* obj, unsigned int key);
void sub_4E4DC0(void);
void sub_4E4DD0(void);
typedef struct nox_important_rate_control {
	uint8_t resends_per_update;
	uint8_t resend_interval;
	uint8_t update_rate;
	uint8_t reserved_3;
	uint32_t threshold;
	uint32_t lower_threshold;
} nox_important_rate_control_t;
_Static_assert(offsetof(nox_important_rate_control_t, resends_per_update) == 0,
	"wrong offset of nox_important_rate_control_t.resends_per_update");
_Static_assert(offsetof(nox_important_rate_control_t, resend_interval) == 1,
	"wrong offset of nox_important_rate_control_t.resend_interval");
_Static_assert(offsetof(nox_important_rate_control_t, update_rate) == 2,
	"wrong offset of nox_important_rate_control_t.update_rate");
_Static_assert(offsetof(nox_important_rate_control_t, threshold) == 4,
	"wrong offset of nox_important_rate_control_t.threshold");
_Static_assert(offsetof(nox_important_rate_control_t, lower_threshold) == 8,
	"wrong offset of nox_important_rate_control_t.lower_threshold");
_Static_assert(sizeof(nox_important_rate_control_t) == 12, "wrong size of nox_important_rate_control_t");
typedef nox_important_rate_control_t nox_important_rate_controls_t[32];
_Static_assert(sizeof(nox_important_rate_controls_t) == 384, "wrong size of nox_important_rate_controls_t");
typedef uint16_t nox_important_player_counters_t[32];
_Static_assert(sizeof(nox_important_player_counters_t) == 64, "wrong size of nox_important_player_counters_t");
typedef struct nox_important_packet {
	uint32_t created_frame;
	uint32_t last_send_frame[32];
	uint8_t retry_delay[32];
	uint8_t send_count;
	uint8_t reserved_165[3];
	uint32_t acknowledged_mask;
	uint32_t sent_mask;
	uint32_t recipient_mask;
	uint32_t remove_if_disconnected;
	uint8_t sequence_enabled;
	uint8_t reserved_185;
	uint16_t sequence[32];
	int8_t recipient;
	uint8_t payload[150];
	uint8_t payload_size;
	uint8_t reserved_402[2];
	uint32_t legacy_related_object;
	uint32_t legacy_next;
	uint32_t legacy_prev;
#if UINTPTR_MAX > UINT32_MAX
	nox_object_t* native_related_object;
	struct nox_important_packet* native_next;
	struct nox_important_packet* native_prev;
#endif
} nox_important_packet_t;
_Static_assert(offsetof(nox_important_packet_t, created_frame) == 0,
	"wrong offset of nox_important_packet_t.created_frame");
_Static_assert(offsetof(nox_important_packet_t, last_send_frame) == 4,
	"wrong offset of nox_important_packet_t.last_send_frame");
_Static_assert(offsetof(nox_important_packet_t, retry_delay) == 132,
	"wrong offset of nox_important_packet_t.retry_delay");
_Static_assert(offsetof(nox_important_packet_t, send_count) == 164,
	"wrong offset of nox_important_packet_t.send_count");
_Static_assert(offsetof(nox_important_packet_t, acknowledged_mask) == 168,
	"wrong offset of nox_important_packet_t.acknowledged_mask");
_Static_assert(offsetof(nox_important_packet_t, sent_mask) == 172,
	"wrong offset of nox_important_packet_t.sent_mask");
_Static_assert(offsetof(nox_important_packet_t, recipient_mask) == 176,
	"wrong offset of nox_important_packet_t.recipient_mask");
_Static_assert(offsetof(nox_important_packet_t, remove_if_disconnected) == 180,
	"wrong offset of nox_important_packet_t.remove_if_disconnected");
_Static_assert(offsetof(nox_important_packet_t, sequence_enabled) == 184,
	"wrong offset of nox_important_packet_t.sequence_enabled");
_Static_assert(offsetof(nox_important_packet_t, sequence) == 186,
	"wrong offset of nox_important_packet_t.sequence");
_Static_assert(offsetof(nox_important_packet_t, recipient) == 250,
	"wrong offset of nox_important_packet_t.recipient");
_Static_assert(offsetof(nox_important_packet_t, payload) == 251,
	"wrong offset of nox_important_packet_t.payload");
_Static_assert(offsetof(nox_important_packet_t, payload_size) == 401,
	"wrong offset of nox_important_packet_t.payload_size");
_Static_assert(offsetof(nox_important_packet_t, legacy_related_object) == 404,
	"wrong offset of nox_important_packet_t.legacy_related_object");
_Static_assert(offsetof(nox_important_packet_t, legacy_next) == 408,
	"wrong offset of nox_important_packet_t.legacy_next");
_Static_assert(offsetof(nox_important_packet_t, legacy_prev) == 412,
	"wrong offset of nox_important_packet_t.legacy_prev");
#if UINTPTR_MAX == UINT32_MAX
_Static_assert(sizeof(nox_important_packet_t) == 416, "wrong 32-bit size of nox_important_packet_t");
#else
_Static_assert(offsetof(nox_important_packet_t, native_related_object) == 416,
	"wrong offset of nox_important_packet_t.native_related_object");
_Static_assert(offsetof(nox_important_packet_t, native_next) == 424,
	"wrong offset of nox_important_packet_t.native_next");
_Static_assert(offsetof(nox_important_packet_t, native_prev) == 432,
	"wrong offset of nox_important_packet_t.native_prev");
_Static_assert(sizeof(nox_important_packet_t) == 440, "wrong 64-bit size of nox_important_packet_t");
#endif

nox_alloc_class* nox_server_getImportantAllocClass_4E4DE0(void);
void nox_server_setImportantAllocClass_4E4DE0(nox_alloc_class* alloc);
nox_important_packet_t* nox_server_getImportantFirst_4E4F80(void);
nox_important_packet_t* nox_server_getImportantLast_4E4F80(void);
void nox_server_setImportantFirst_4E4F80(nox_important_packet_t* packet);
void nox_server_setImportantLast_4E4F80(nox_important_packet_t* packet);
nox_object_t* nox_server_getImportantRelatedObject_4E5030(const nox_important_packet_t* packet);
void nox_server_setImportantRelatedObject_4E5030(nox_important_packet_t* packet, nox_object_t* object);
nox_important_packet_t* nox_server_getImportantNext_4E4F80(const nox_important_packet_t* packet);
nox_important_packet_t* nox_server_getImportantPrev_4E4F80(const nox_important_packet_t* packet);
void nox_server_setImportantNext_4E4F80(nox_important_packet_t* packet, nox_important_packet_t* next);
void nox_server_setImportantPrev_4E4F80(nox_important_packet_t* packet, nox_important_packet_t* prev);
int sub_4E4DE0(void);
int sub_4E4E50(int a1);
int sub_4E4ED0(void);
int sub_4E4EF0(void);
int sub_4E4F30(int a1);
int nox_xxx_playerResetImportantCtr_4E4F40(int a1);
int sub_4E4F80(void);
void sub_4E4FC0(nox_important_packet_t* packet);
int nox_xxx_netSendPacket_4E5030(
	int recipient, const void* payload, signed int payload_size, nox_object_t* related_object,
	int remove_if_disconnected, char sequence_enabled);
int nox_xxx_importantCheckRate_4E52B0();
int nox_server_playerKickDueToRate_4E5360(int player_index);
int nox_xxx_playerKickDueToRate_4E5360(int player_index);
int nox_xxx_netSendPacket1_4E5390(
	int recipient, const void* payload, int payload_size, nox_object_t* related_object,
	int remove_if_disconnected);
int nox_xxx_netClientSend2_4E53C0(
	int recipient, const void* payload, int payload_size, nox_object_t* related_object,
	int remove_if_disconnected);
int nox_xxx_netSendPacket0_4E5420(
	int recipient, const void* payload, signed int payload_size, nox_object_t* related_object,
	int remove_if_disconnected);
int sub_4E5450(
	int recipient, const void* payload, signed int payload_size, nox_object_t* related_object,
	int remove_if_disconnected);
void sub_4E54D0(uint32_t client_mask, nox_important_packet_t* packet, int player_index);
int nox_net_importantACK_4E55A0(uint8_t player_index, uint32_t frame);
int sub_4E55F0(uint8_t player_index);
uint32_t sub_4E5630(
	uint8_t player_index, uint32_t* threshold, uint32_t* resend_interval, uint32_t* resends_per_update);
void nox_server_importantPlayerLookup_4E5670(uint8_t player_index);
uint32_t nox_server_importantRateGet_4E5670(void);
uint32_t nox_xxx_importantCheckRate2_4E5670(uint8_t player_index);
int nox_server_importantShouldProcess_4E5770(uint8_t player_index);
int nox_server_importantSend_4E5770(
	uint8_t player_index, int message_kind, uint8_t* data, uint32_t size);
int nox_server_importantReplayRead_4E5770(void);
int nox_server_importantGameHost_4E5770(void);
void nox_server_importantRateAdjust_4E5770(uint8_t player_index);
void nox_xxx_netImportant_4E5770(uint8_t player_index, int message_kind);
uint32_t nox_xxx_importantFreeSlots_4E5A90(void);
void nox_xxx_noop_4E5AB0(const void* unused);
void sub_4E5AC0(void);
void nox_xxx_playerRemoveSpawnedStuff_4E5AD0(nox_object_t* a1);
int nox_xxx_isUnit_4E5B50(nox_object_t* obj);
int sub_4E5B80(nox_object_t* obj);
void sub_4E5BF0(int a1);
void nox_xxx_delayedDeleteObject_4E5CC0(nox_object_t* obj);
int sub_4E5F40(nox_object_t* owner);
void sub_4E5FC0(nox_object_t* owner);
void nox_xxx_playerCameraUnlock_4E6040(nox_object_t* player);
void nox_xxx_playerCameraFollow_4E6060(nox_object_t* player, nox_object_t* unitId);
void nox_xxx_updatePlayerObserver_4E62F0(nox_object_t* a1);
int nox_xxx_playerGoObserver_4E6860(nox_playerInfo* pl, int a2, int a3);
void nox_xxx_playerLeaveObserver_0_4E6AA0(nox_playerInfo* pl);
int sub_4E6BD0(nox_object_t* unit);
double nox_xxx_calcDistance_4E6C00(nox_object_t* a1, nox_object_t* a2);
int sub_4E6CE0(float2* a1, float2* a2);
int nox_server_testTwoPointsAndDirection_4E6E50(float2* a1, int a2, float2* a3);
void nox_xxx_unitMove_4E7010(nox_object_t* obj, float2* a2);
void nox_xxx_teleportToMB_4E7190(nox_object_t* obj, float2* pos);
nox_object_t* nox_xxx_objectUnkUpdateCoords_4E7290(nox_object_t* obj);
nox_object_t* sub_4E7350(nox_object_t* obj);
int sub_4E7410(nox_object_t* obj);
void nox_xxx_spawnSomeBarrel_4E7470(nox_object_t* source, float2* pos);
void sub_4E7540(nox_object_t* a1, nox_object_t* a2);
char nox_xxx_objectSetOn_4E75B0(nox_object_t* obj);
int nox_xxx_objectSetOff_4E7600(nox_object_t* obj);
nox_object_t* nox_xxx_inventoryGetFirst_4E7980(nox_object_t* obj);
nox_object_t* nox_xxx_inventoryGetNext_4E7990(nox_object_t* obj);
uint32_t sub_4E79B0(uint32_t value);
uint32_t* nox_xxx_unitFreezeGateRef_4E79B0(void);
char nox_xxx_unitFreeze_4E79C0(nox_object_t* obj, uint32_t source);
char nox_xxx_unitUnFreeze_4E7A60(nox_object_t* obj, uint32_t force);
void nox_xxx_unitBecomePet_4E7B00(nox_object_t* owner, nox_object_t* pet);
void nox_xxx_monsterRemoveMonitors_4E7B60(nox_object_t* a1, nox_object_t* a2);
int sub_4E7BC0(const nox_object_t* obj);
int nox_xxx_unitIsCrown_4E7BE0(const nox_object_t* owner);
int nox_xxx_unitIsGameball_4E7C30(const nox_object_t* owner);
int32_t nox_xxx_unitIsUnitTT_4E7C80(nox_object_t* owner, int32_t type_ind);
int32_t nox_xxx_unitCountSlaves_4E7CF0(const nox_object_t* owner, uint32_t class_mask, uint32_t subclass_mask);
int32_t nox_xxx_inventoryCountObjects_4E7D30(nox_object_t* owner, int32_t type_ind);
int32_t sub_4E7DE0(const nox_object_t* candidate, const nox_object_t* item);
int32_t sub_4E7EC0(const nox_object_t* owner, const nox_object_t* item);
int32_t nox_xxx_unitIsHostileMimic_4E7F90(nox_object_t* a1, nox_object_t* a2);
void nox_xxx_monsterMarkUpdate_4E8020(nox_object_t* a1);
nox_object_t* sub_4E8110(int32_t player_ind);
int32_t sub_4E8290(uint8_t state, uint16_t net_code);
int32_t sub_4E82C0(uint8_t team_id, uint8_t status, uint8_t flag_index, uint16_t carrier_net_code);
nox_game_ball_status_t* sub_4E8310(void);
nox_team_flag_status_t* sub_4E8320(uint8_t team_id);
void nox_xxx_fnFindCloseDoors_4E8340(nox_object_t* door, nox_point* target);
int32_t sub_4E8390(nox_object_t* door);
void* nox_xxx_collideMonsterEventProc_4E83B0(nox_object_t* monster, nox_object_t* other, float* collision);
void* nox_xxx_collideMimic_4E83D0(nox_object_t* mimic, nox_object_t* other, float* collision);
void nox_xxx_collidePlayer_4E8460(nox_object_t* player, nox_object_t* other, float* collision);
void nox_xxx_collideProjectileGeneric_4E87B0(
	nox_object_t* projectile,
	nox_object_t* other,
	float* collision);
void nox_xxx_collideProjectileSpark_4E8880(
	nox_object_t* projectile,
	nox_object_t* other,
	float* collision);
nox_object_t* nox_xxx_doorGetSomeKey_4E8910(nox_object_t* a1, nox_object_t* a2);
void nox_xxx_collideDoor_4E8AC0(
	nox_object_t* door,
	nox_object_t* unit,
	float* collision);
uintptr_t nox_xxx_collidePickup_4E8DF0(
	nox_object_t* item,
	nox_object_t* unit,
	float* collision);
int32_t sub_4E8E60(void);
int32_t nox_server_questMaybeWarp_4E8F60(void);
int32_t sub_4E9010(void);
void nox_xxx_collideExit_4E9090(
	nox_object_t* exit,
	nox_object_t* unit,
	float* collision);
void nox_xxx_collideDamage_4E9430(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
// GAME.EXE 004E9490 is declared by mana_drain_collide_4e9490.h.
// GAME.EXE 004E9500 is declared by spell_projectile_collide_4e9500.h.
// GAME.EXE 004E96F0 is declared by bomb_collide_4e96f0.h.
// GAME.EXE 004E9770 is declared by boom_collide_4e9770.h.
// GAME.EXE 004E99B0 is declared by die_collide_4e99b0.h.
// GAME.EXE 004E9A00 and 004E9A30 are declared by glyph_collide_4e9a00.h.
// GAME.EXE 004E9AC0 is declared by spark_explosion_collide_4e9ac0.h.
// GAME.EXE 004E9C40 is declared by chest_collide_4e9c40.h.
// GAME.EXE 004E9D80 and 004E9E50 are declared by
// wall_reflect_collide_4e9d80.h.
// GAME.EXE 004E9E90 is declared by death_ball_collide_4e9e90.h.
// GAME.EXE 004E9FE0 is declared by death_ball_fragment_collide_4e9fe0.h.
// GAME.EXE 004EA080 is declared by pixie_collide_4ea080.h.
// GAME.EXE 004EA200 is declared by wall_reflect_spark_collide_4ea200.h.
// GAME.EXE 004EA2C0 is declared by own_collide_4ea2c0.h.
// GAME.EXE 004EA300 is declared by spark_collide_4ea300.h.
// GAME.EXE 004EA380 is declared by webbing_collide_4ea380.h.
// GAME.EXE 004EA400 and 004EA7A0 are declared by flag_collide_4ea400.h;
// 004EA490 and 004EA800 are internal typed Go restoration units.
// GAME.EXE 004EAAA0 is declared by barrel_collide_4eaaa0.h.
// GAME.EXE 004EAAD0 is declared by audio_event_collide_4eaad0.h.
// GAME.EXE 004EAB20 is declared by pentagram_collide_4eab20.h.
// GAME.EXE 004EAB40 is declared by sign_collide_4eab40.h.
// GAME.EXE 004EAB60 is declared by trap_door_collide_4eab60.h.
// GAME.EXE 004EACA0 is declared by teleport_collide_4eaca0.h.
// GAME.EXE 004EAD20 is declared by award_spell_collide_4ead20.h.
// GAME.EXE 004EADF0 is declared by fist_collide_4eadf0.h.
// GAME.EXE 004EAE30 is declared by teleport_wake_collide_4eae30.h.
// GAME.EXE 004EAF00 is declared by chakram_collide_4eaf00.h.
int sub_4EB250(int a1);
void sub_4EB340(float* a1, int a2);
void sub_4EB3E0(int a1);
// GAME.EXE 004EB490 is declared by arrow_collide_4eb490.h.
// GAME.EXE 004EB6A0 is declared by harpoon_collide_4eb6a0.h.
// GAME.EXE 004EB800 is declared by monster_arrow_collide_4eb800.h.
// GAME.EXE 004EB890 is declared by bear_trap_collide_4eb890.h.
// GAME.EXE 004EB910 is declared by poison_gas_trap_collide_4eb910.h.
// GAME.EXE 004EB9B0 is declared by game_ball_carrier_state_4eb9b0.h.
// GAME.EXE 004EBA00 is declared by ball_collide_4eba00.h.
// GAME.EXE 004EBB50 is declared by crown_collide_4ebb50.h.
// GAME.EXE 004EBB80 is declared by home_base_collide_4ebb80.h.
// GAME.EXE 004EBD40 is declared by undead_killer_collide_4ebd40.h.
// GAME.EXE 004EBE10 is declared by monster_generator_collide_4ebe10.h.
// GAME.EXE 004EBE40 is declared by soul_gate_collide_4ebe40.h.
// GAME.EXE 004EBF40 is declared by ankh_collide_4ebf40.h.
void nox_xxx_unitSetOwner_4EC290(nox_object_t* obj1, nox_object_t* obj2);
void nox_xxx_unitClearOwner_4EC300(nox_object_t* obj);
// GAME.EXE 004EC3E0 is restored by player_observer_good_slave.go. Its two
// decoded callers are Go-owned, so no pointer-truncating C declaration remains.
// GAME.EXE 004EC420 is restored there as well. Its three decoded callers are
// Go-owned, so no pointer-truncating C declaration remains.
// GAME.EXE 004EC470 is restored by unit_remove_child_4ec470_native.go. Its
// only decoded caller is Go-owned, so no CGo declaration is needed.
// GAME.EXE 004EC4B0 is restored by server/unit_transfer_slaves_4ec4b0_server.go.
// One decoded caller remains in GAME4_1.c and uses the typed Go CGo export.
void nox_xxx_unitTransferSlaves_4EC4B0(nox_object_t* a1);
// GAME.EXE 004EC4F0 is restored by server/unit_has_that_parent_4ec4f0_server.go.
// Four decoded callers remain in C and use this native-pointer, 32-bit C int ABI.
int nox_xxx_unitHasThatParent_4EC4F0(nox_object_t* a1, nox_object_t* a2);
// GAME.EXE 004EC520 is declared by units_same_team_4ec520.h.
nox_object_t* nox_xxx_findParentChainPlayer_4EC580(nox_object_t* unit);
void sub_4EC5B0();
// GAME.EXE 004EC5E0 is restored by respawn_add_4ec5e0.go. Its only decoded
// caller is Go-owned, so no CGo declaration remains.
// GAME.EXE 004EC6A0 is restored by respawn_remove_4ec6a0.go. Three decoded
// callers remain in C and use this native object-pointer CGo export.
void sub_4EC6A0(nox_object_t* obj);
// GAME.EXE 004ECA60 is restored by respawn_allocator_4eca60.go. Its only
// decoded caller is Go-owned, so no CGo declaration remains.
// GAME.EXE 004ECA90 is restored by respawn_allocator_free_4eca90.go. Its only
// decoded caller is Go-owned, so no CGo declaration remains.
// GAME.EXE 004ECBD0 is restored by team_material_flag_index_4ecbd0.go. Four
// decoded callers remain in C and use this native object-pointer CGo export;
// the other three callers are part of the Go-owned CTF path.
int sub_4ECBD0(nox_object_t* obj);
// GAME.EXE 004ECC00 is restored as the private portable lookup behind
// 004ECBD0. Its other decoded caller, 004ECC70, is not retained in C, so no
// pointer-to-pointer C declaration remains.
// GAME.EXE 004ECC70 is restored by team_base_material_index_4ecc70.go. It has
// no decoded caller, jump, or stored function pointer, so no declaration or
// CGo export is invented.
nox_object_t* nox_server_getObjectFromNetCode_4ECCB0(int a1);
int sub_4ECDE0(uint32_t* a1, int a2);
int sub_4ECE10(uint32_t* a1, int a2);
// GAME.EXE 004ECF10 is restored by object_by_script_id_4ecf10.go. Nine
// decoded callers remain in C and use this signed-ID/native-pointer export.
nox_object_t* sub_4ECF10(int32_t script_id);
// GAME.EXE 004ECFA0 is restored by net_code_cache_remove_object_4ecfa0.go.
// Its sole decoded caller is Go-owned, so no CGo declaration remains.
// GAME.EXE 004ED020 is restored by object_by_extent_4ed020.go. Three decoded
// callers remain in C and use this unsigned-Extent/native-pointer export.
nox_object_t* nox_xxx_netGetUnitByExtent_4ED020(uint32_t extent);
// GAME.EXE 004ED050 is restored by drop_owned_crowns_4ed050.go. Its sole
// decoded caller remains in GAME5.c and uses this native-pointer export.
// GAME.EXE 004ED0C0 is restored by inventory_detach_4ed0c0.go. C callers use
// the native-pointer declaration from inventory_detach_4ed0c0.h.
// GAME.EXE 004ED290 is restored by default_drop_4ed290.go. C callers use the
// native-pointer declaration from default_drop_4ed290.h.
// GAME.EXE 004ED500 and 004ED580 are restored by glyph_drop_4ed500.go and
// trap_drop_4ed580.go. Their registered CGo boundaries use the native-pointer
// declarations in the dedicated headers.
// GAME.EXE 004ED5E0 is restored by crown_drop_4ed5e0.go. Its registered CGo
// boundary uses the native-pointer declaration in crown_drop_4ed5e0.h.
// GAME.EXE 004ED710 is restored by treasure_drop_4ed710.go. Its registered CGo
// boundary uses the native-pointer declaration in treasure_drop_4ed710.h.
int nox_xxx_drop_4ED790(nox_object_t* a1, nox_object_t* a2, float2* a3);
// GAME.EXE 004ED810 is restored by object_drop_bounded_4ed810.go. Its two
// decoded C callers use this native-pointer CGo boundary.
int nox_xxx_drop_4ED810(nox_object_t* owner, nox_object_t* item, float2* point);
// GAME.EXE 004ED930 is restored by object_force_drop_4ed930.go. Active C
// callers use this native-pointer CGo boundary.
int nox_xxx_invForceDropItem_4ED930(nox_object_t* owner, nox_object_t* item);
// GAME.EXE 004ED970 is restored by random_reachable_point_4ed970_export.go.
// Its decoded C callers use the native-pointer declaration in the dedicated
// header.
// GAME.EXE 004EDA40 and its private 004EDCD0 predicate are restored by the
// dedicated native-pointer CGo boundary.
// GAME.EXE 004EDDE0 is restored by potion_drop_4edde0.go. Its registered CGo
// boundary uses the native-pointer declaration in potion_drop_4edde0.h.
// GAME.EXE 004EDE50 is restored by food_drop_4ede50.go. Its registered CGo
// boundary uses the native-pointer declaration in food_drop_4ede50.h.
// GAME.EXE 004EDF00 and private helper 004EE2A0 are restored by
// chest_open_4edf00.go. The public entrypoint uses the native-pointer
// declaration in chest_open_4edf00.h.
// GAME.EXE 004EE2F0 is restored by aud_event_drop_4ee2f0.go. Its registered
// CGo boundary uses the native-pointer declaration in
// aud_event_drop_4ee2f0.h.
// GAME.EXE 004EE370 is restored by ankh_tradable_drop_4ee370.go. Its
// registered CGo boundary uses the native-pointer declaration in
// ankh_tradable_drop_4ee370.h.
// GAME.EXE 004EE460/004EE4C0 and their current-HP reporter are restored by
// unit_adjust_hp_4ee460.go. Active C callers use the native-pointer CGo
// boundary declared in unit_adjust_hp_4ee460.h.
// GAME.EXE 004EE5E0 is restored by unit_damage_clear_4ee5e0.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EE6F0 is restored by unit_hp_set_max_4ee6f0.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EE730 is restored by player_hp_init_4ee730.go. Its sole decoded
// C caller uses the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EE780 is restored by unit_get_hp_4ee780.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EE7A0 is restored by unit_get_max_hp_4ee7a0.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EE7C0 is restored by unit_set_max_hp_4ee7c0.go. Its decoded C
// caller uses the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EE7E0..004EEB77 is restored by poison_state_4ee7e0.go. Active C
// callers use the native-pointer CGo declarations in the dedicated header.
// GAME.EXE 004EEB80 is restored by player_mana_add_4eeb80.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EEBF0 is restored by player_mana_sub_4eebf0.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EEC80 is restored by unit_get_old_mana_4eec80.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EECB0 is restored by player_get_max_mana_4eecb0.go. Its decoded
// C callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EECD0 is restored by player_set_max_mana_4eecd0.go. Its sole
// decoded C caller uses the native-pointer CGo declaration in that header.
// GAME.EXE 004EECF0 is restored by player_mana_refresh_4eecf0.go. Its decoded
// C callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EEDC0 is restored by player_read_values_4eedc0.go. Its decoded
// C callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EF140 is restored by player_sync_level_4ef140.go. Its decoded
// C callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EF1E0 is restored by bolt_damage_4ef1e0_export.go. Active C
// callers use the native modifier pointer and scalar declarations above.
// GAME.EXE 004EF410 is restored by player_level_set_4ef410_export.go. Its
// sole decoded C caller uses the native-pointer declaration in that header.
// GAME.EXE 004EF500 is restored by god_mode_controller_4ef500_export.go. Its
// decoded C caller uses the exact int32 declaration in that dedicated header.
// GAME.EXE 004EF580 is restored by respawn_weapon_flags_4ef580_export.go. Its
// sole decoded C caller consumes the exact uint8 declaration in that header.
// GAME.EXE 004EF6F0 is restored by glyph_inventory_count_4ef6f0.go. Its sole
// production caller is Go-owned, so no public C/CGo declaration remains.
// GAME.EXE 004EF750 is restored by player_respawn_item_4ef750_export.go. Its
// decoded C callers use native pointers and exact int32 placement arguments
// through the dedicated CGo declaration above.
// GAME.EXE 004EF7D0 is restored by player_make_def_items_4ef7d0_export.go. Its
// decoded callers use the native object pointer, exact int32 controls, and
// unsigned low-AL result declared in the dedicated header above.
// GAME.EXE 004EFC30 is restored by player_respawn_packet_4efc30.go. Its two
// production callers are Go-owned, so no public C/CGo declaration remains.
// GAME.EXE 004EFC80 is restored by spell_award_all_4efc80.go. Its decoded C
// callers use the native-pointer CGo declaration in the dedicated header.
// GAME.EXE 004EFD80 is restored by beast_scroll_award_all_4efd80.go. Its
// decoded C callers use the native-pointer declaration in the dedicated
// header above.
// GAME.EXE 004EFE10 is restored by warrior_ability_award_all_4efe10.go. Its
// decoded C callers use the native-pointer declaration in the dedicated
// header above.
// GAME.EXE 004EFE80 is restored by player_unit_init_4efe80.go. The object-init
// registry uses its native-pointer, unsigned-byte CGo declaration from the
// dedicated header above.
// GAME.EXE 004F0040 is restored by monster_init_4f0040.go. The MonsterInit
// and ShopkeeperInit registry rows use the native-pointer, void CGo
// declaration from the dedicated header above.
// GAME.EXE 004F0390 is restored by spark_init_4f0390.go. The object-init
// registry uses its native object pointer and typed Spark update-data return
// from the dedicated header above.
// GAME.EXE 004F03B0 is restored by frog_init_4f03b0.go. The object-init
// registry uses its native object pointer and fixed-width int32 return from
// the dedicated header above.
// GAME.EXE 004F0400 is restored by chest_init_4f0400.go. The object-init
// registry and the callback-identity consumer use the native-pointer, void
// declaration from the dedicated header above.
// GAME.EXE 004F0420 is restored by boulder_init_4f0420.go. The object-init
// registry uses its native object-pointer argument and result declaration
// from the dedicated header above.
// GAME.EXE 004F0450 is restored by skull_init_4f0450.go. The object-init
// registry uses its native object-pointer and exact int32_t result declaration
// from the dedicated header above.
// GAME.EXE 004F0490 is restored by direction_init_4f0490.go. The object-init
// registry uses its native object-pointer and exact int32_t result declaration
// from the dedicated header above.
// GAME.EXE 004F04B0 is restored by gold_init_4f04b0.go. The object-init
// registry uses its native object-pointer and exact int32_t result declaration
// from the dedicated header above.
// GAME.EXE 004F0570 is restored by break_init_4f0570.go. The object-init
// registry uses its native object-pointer and void callback declaration from
// the dedicated header above.
// GAME.EXE 004F0590 is restored by monster_generator_init_4f0590.go. The
// object-init registry uses its native object-pointer and exact int32_t result
// declaration from the dedicated header above.
// GAME.EXE 004F0720 is restored by reward_marker_activate_4f0720.go. Its
// decoded C callers use the native object-pointer and exact uint32_t stage
// declaration from the dedicated header above.
// GAME.EXE 004F0D20 is restored by reward_field_guide_4f0d20.go. Its decoded
// C callers use the native object-pointer and exact uint32_t stage declaration
// from the dedicated header above.
// GAME.EXE 004F0E80 is restored by reward_armor_4f0e80.go. Its sole active
// marker dispatch now stays in Go and carries native object pointers.
// GAME.EXE 004F14E0 is restored by reward_weapon_4f14e0.go. Its sole active
// marker dispatch now stays in Go and carries native object pointers.
// GAME.EXE 004F1C40 is restored by reward_potion_4f1c40.go. Its sole active
// marker dispatch now stays in Go and carries native object pointers.
// GAME.EXE 004F1D30 and its exact thin 004F1F00 wrapper are restored by
// reward_gem_4f1d30.go. Their active marker dispatch paths now stay in Go and
// carry native object pointers.
// GAME.EXE 004F2110 is restored by reward_ankh_replace_4f2110.go and is called
// by the native 004F1F20 preprocessing callback.
// GAME.EXE 004F2210 is restored by reward_replenish_4f2210.go and is called by
// that same native preprocessing callback; no active C declaration remains.
// GAME.EXE 004F2530 and 004F2570 are restored by their fixed-width Go files.
// Their decoded C callers use exact scalar declarations from the dedicated
// headers above. GAME.EXE 004F2590..004F2B60 and 004F2C30 are restored by the
// native Go Quest inventory path; their raw int object handles and internal
// helpers have no active C declarations. GAME.EXE 004F2E70 is restored by the
// native Go spellbook path and likewise has no active C declaration. GAME.EXE
// 004F2EF0 is restored by the native Go fieldbook path with the same scalar
// contract, native-width family pointers, and no active C declaration.
// GAME.EXE 004F2F70 is restored by player_try_equip_4f2f70_export.go. Its
// remaining decoded C callers use the exact fixed-width result and native
// object-pointer declaration from the dedicated header above.
// GAME.EXE 004F2FB0 is restored by player_try_dequip_4f2fb0_export.go. Its
// remaining decoded C callers use the const-correct fixed-width declaration
// and adapter from the dedicated header above.
// GAME.EXE 004F2FF0 is restored by item_apply_engage_4f2ff0_export.go. Its
// four decoded C callers discard EAX and use the exact void/native-pointer
// declaration from the dedicated header above.
// GAME.EXE 004F3030 is restored by item_apply_disengage_4f3030_export.go. Its
// four decoded C callers also discard EAX; the dedicated header preserves the
// exact const item/native-pointer public declaration and void contract.
void nox_xxx_inventoryPutImpl_4F3070(nox_object_t* a1, nox_object_t* item, int32_t a3);
// GAME.EXE 004F3180 is restored by player_check_strength_4f3180_export.go.
// Remaining decoded C callers use the exact int32/native-pointer declaration
// from the dedicated header above; the original has no allow-all cheat branch.
// GAME.EXE 004F31E0 is restored by pickup_default_4f31e0_export.go. Its
// registered callback ABI has four arguments; the body reads report and does
// not read the final argument. All decoded C callers use the dedicated header.
// GAME.EXE 004F3350 is restored by pickup_food_4f3350_export.go. Its
// registered callback also has four arguments and forwards both trailing
// int32 values to DefaultPickup; decoded callers use the dedicated header.
// GAME.EXE 004F3400 is declared by crown_pickup_4f3400.h.
// GAME.EXE 004F34D0 is declared by pickup_use_4f34d0.h.
// GAME.EXE 004F3510 is declared by pickup_trap_4f3510.h. The registered
// callback has four arguments even though the old transcription exposed three.
// GAME.EXE 004F3580 is declared by pickup_treasure_4f3580.h. Its registered
// callback also has four arguments; the old three-argument declaration dropped
// the fourth value before forwarding to DefaultPickup.
// GAME.EXE 004F36F0 is restored by inventory_serv_place_4f36f0_export.go.
// All decoded C callers use its native-pointer/fixed-width declaration above.
int nox_xxx_pickupPotion_4F37D0(nox_object_t* a1, nox_object_t* a2, int a3, int a4);
// GAME.EXE 004F3A60 is declared by pickup_gold_4f3a60.h. Its registered
// callback has four arguments; the old transcription exposed three and forced
// the fourth DefaultPickup argument to zero.
// GAME.EXE 004F3B00 is declared by pickup_ammo_4f3b00.h. Both object arguments
// are native pointers and both trailing arguments and the result are int32_t.
// nox_xxx_pickupSpellbook_4F3C60 has a native-width declaration in
// pickup_spellbook_4f3c60.h.
// nox_xxx_pickupAbilitybook_4F3CE0 has a native-width, four-argument
// declaration in pickup_abilitybook_4f3ce0.h.
int nox_xxx_xfer_4F3E30(unsigned short a1, nox_object_t* a2, int a3);
int nox_xxx_servMapLoadPlaceObj_4F3F50(nox_object_t* a1, int a2, void* a3);
char sub_4F40A0(nox_object_t* a1);
int nox_xxx_readObjectOldVer_4F4170(int a1, int a2, int a3);
int nox_xxx_mapReadWriteObjData_4F4530(nox_object_t* a1, int a2);
int nox_xxx_XFerDefault_4F49A0(nox_object_t* a1, void* a2);
int nox_xxx_XFerSpellPagePedistal_4F4A20(int a1);
int nox_xxx_XFerReadable_4F4AB0(nox_object_t* obj);
int nox_xxx_XFerExit_4F4B90(nox_object_t* obj);
int nox_xxx_XFerDoor_4F4CB0(nox_object_t* obj);
int nox_xxx_unitTriggerXfer_4F4E50(nox_object_t* obj);
int nox_xxx_XFerHole_4F51D0(nox_object_t* obj);
int nox_xxx_XFerTransporter_4F5300(nox_object_t* obj);
int nox_xxx_XFerElevator_4F53D0(nox_object_t* obj);
int nox_xxx_XFerElevatorShaft_4F54A0(nox_object_t* obj);
int sub_4F5540(int a1);
int nox_xxx_XFerMover_4F5730(int a1);
int nox_xxx_XFerGlyph_4F5890(int a1);
int nox_xxx_XFerInvLight_4F5AA0(nox_object_t* obj);
int nox_xxx_XFerSentry_4F5E50(int a1);

#endif // NOX_PORT_GAME3_3
