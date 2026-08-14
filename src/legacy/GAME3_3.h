#ifndef NOX_PORT_GAME3_3
#define NOX_PORT_GAME3_3

#include "defs.h"
#include "object_collide_noop_4e87a0.h"
#include "quest_map_buffer_4e8e50.h"

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
int nox_xxx_collideSpellPedestal_4EAD20(int a1, int a2);
void nox_xxx_collideFist_4EADF0(int a1, int a2);
void nox_xxx_collideTeleportWake_4EAE30(int a1, int a2);
void nox_xxx_collideChakram_4EAF00(int a1, int a2, float* a3);
int sub_4EB250(int a1);
void sub_4EB340(float* a1, int a2);
void sub_4EB3E0(int a1);
void nox_xxx_collideArrow_4EB490(int a1, int a2);
void nox_xxx_collideHarpoon_4EB6A0(nox_object_t* a1, nox_object_t* a2);
void nox_xxx_collideMonsterArrow_4EB800(int a1, int a2);
void nox_xxx_collideBearTrap_4EB890(int* a1, int a2);
void nox_xxx_collidePoisonGasTrap_4EB910(int* a1, int a2);
int sub_4EB9B0(int a1, int a2);
void nox_xxx_collideBall_4EBA00(int a1, int a2);
int sub_4EBB50(int a1, int a2);
short nox_xxx_collideHomeBase_4EBB80(int a1, int a2);
void nox_xxx_collideUndeadKiller_4EBD40(int a1, int a2, int a3);
void nox_xxx_collideMonsterGen_4EBE10(int a1, int a2);
void sub_4EBE40(int a1, int a2);
void nox_xxx_collideAnkhQuest_4EBF40(int a1, int a2);
void nox_xxx_unitSetOwner_4EC290(nox_object_t* obj1, nox_object_t* obj2);
void nox_xxx_unitClearOwner_4EC300(nox_object_t* obj);
int nox_xxx_playerObserverFindGoodSlave2_4EC3E0(int a1);
int nox_xxx_playerObserverFindGoodSlave_4EC420(int a1);
void nox_xxx_unitRemoveChild_4EC470(nox_object_t* a1);
void nox_xxx_unitTransferSlaves_4EC4B0(nox_object_t* a1);
int nox_xxx_unitHasThatParent_4EC4F0(nox_object_t* a1, nox_object_t* a2);
// GAME.EXE 004EC520 is declared by units_same_team_4ec520.h.
nox_object_t* nox_xxx_findParentChainPlayer_4EC580(nox_object_t* unit);
void sub_4EC5B0();
nox_respawn_record_t* nox_xxx_respawnAdd_4EC5E0(nox_object_t* obj);
void sub_4EC6A0(nox_object_t* obj);
int nox_xxx_allocItemRespawnArray_4ECA60();
void sub_4ECA90();
int sub_4ECBD0(int a1);
int sub_4ECC00(const char** a1);
nox_object_t* nox_server_getObjectFromNetCode_4ECCB0(int a1);
int nox_server_netCodeCache_lookupObj_4ECD90(int a1);
int sub_4ECDE0(uint32_t* a1, int a2);
int sub_4ECE10(uint32_t* a1, int a2);
int nox_server_netCodeCache_initArray_4ECE50();
int nox_server_netCodeCache_addObj_4ECEA0(int a1);
int nox_server_netCodeCache_nextUnused_4ECEF0();
int sub_4ECF10(int a1);
int sub_4ECFA0(nox_object_t* a1);
int sub_4ECFE0();
int nox_xxx_netGetUnitByExtent_4ED020(int a1);
void sub_4ED050(int a1, int a2);
void sub_4ED0C0(nox_object_t* a1p, nox_object_t* object);
int nox_xxx_dropDefault_4ED290(nox_object_t* a1p, nox_object_t* a2p, float2* a3);
int nox_GlyphDrop_4ED500(int a1, int a2, float2* a3);
int nox_xxx_dropTrap_4ED580(int a1, int a2, float2* a3);
int nox_xxx_dropCrown_4ED5E0(int a1, int a2, int* a3);
int nox_xxx_dropTreasure_4ED710(int a1, int a2, int* a3);
int nox_xxx_drop_4ED790(nox_object_t* a1, nox_object_t* a2, float2* a3);
int nox_xxx_drop_4ED810(int a1, int a2, float* a3);
int nox_xxx_invForceDropItem_4ED930(int a1, uint32_t* a2);
float2* sub_4ED970(float a1, float2* a2, float2* a3);
uint32_t* nox_xxx_dropAllItems_4EDA40(uint32_t* a1);
int nox_xxx_unitDropCheckSome_4EDCD0(int a1, int a2);
int sub_4EDDE0(int a1, uint32_t* a2, int* a3);
int nox_xxx_dropFood_4EDE50(int a1, int a2, int* a3);
void nox_xxx_chest_4EDF00(int a1, int a2);
double sub_4EE2A0(int a1);
int nox_xxx_dropAnkhTradable_4EE370(int a1, int a2, int* a3);
void nox_xxx_unitAdjustHP_4EE460(nox_object_t* unit, int dv);
void nox_xxx_mobInformOwnerHP_4EE4C0(nox_object_t* obj);
void nox_xxx_unitDamageClear_4EE5E0(nox_object_t* unit, int damageAmount);
void nox_xxx_unitHPsetOnMax_4EE6F0(int unit);
void nox_xxx_playerHP_4EE730(int a1);
short nox_xxx_unitGetHP_4EE780(nox_object_t* item);
short nox_xxx_unitGetMaxHP_4EE7A0(int a1);
int nox_xxx_unitSetMaxHP_4EE7C0(int a1, short a2);
int nox_xxx_activatePoison_4EE7E0(int a1, int a2, int a3);
void nox_xxx_updatePoison_4EE8F0(nox_object_t* a1, int a2);
void nox_xxx_removePoison_4EE9D0(nox_object_t* a1);
void nox_xxx_setSomePoisonData_4EEA90(int a1, int a2);
unsigned short nox_xxx_playerManaAdd_4EEB80(nox_object_t* unit, short amount);
uint32_t* nox_xxx_playerManaSub_4EEBF0(int unit, int amount);
short nox_xxx_unitGetOldMana_4EEC80(int unit);
short nox_xxx_playerGetMaxMana_4EECB0(int unit);
int nox_xxx_playerSetMaxMana_4EECD0(int unit, short amount);
uint32_t* nox_xxx_playerManaRefresh_4EECF0(int unit);
void nox_xxx_abilGivePlayerAll_4EED40(int a1, char a2, int a3);
int nox_xxx_plrReadVals_4EEDC0(nox_object_t* a1, int a2);
int sub_4EF140(int a1);
double nox_xxx_calcBoltDamage_4EF1E0(int a1, int a2);
void sub_4EF410(int a1, unsigned char a2);
void nox_xxx_set_god_4EF500(int a1);
char nox_xxx_getRespawnWeaponFlags_4EF580();
int sub_4EF6F0(int a1);
nox_object_t* nox_xxx_playerRespawnItem_4EF750(nox_object_t* player, char* type_id, const nox_modifier_attrs_t* attrs, int a4, int a5);
char nox_xxx_playerMakeDefItems_4EF7D0(int a1, int a2, int a3);
int nox_xxx_netSendPlayerRespawn_4EFC30(int a1, char a2);
void nox_xxx_spellAwardAll2_4EFC80(nox_playerInfo* a1p);
void nox_xxx_spellAwardAll1_4EFD80(nox_playerInfo* a1p);
void nox_xxx_spellAwardAll3_4EFE10(nox_playerInfo* a1p);
char nox_xxx_unitInitPlayer_4EFE80(nox_object_t* a1);
int sub_4EFF10(int a1);
void nox_xxx_unitMonsterInit_4F0040(nox_object_t* a1);
uint32_t* nox_xxx_unitSparkInit_4F0390(int a1);
int nox_xxx_initFrog_4F03B0(int a1);
int* nox_xxx_initChest_4F0400(int a1);
uint32_t* nox_xxx_unitBoulderInit_4F0420(uint32_t* a1);
int sub_4F0450(int a1);
int sub_4F0490(int a1);
int nox_xxx_unitInitGold_4F04B0(int a1);
int* nox_xxx_breakInit_4F0570(int a1);
int nox_xxx_unitInitGenerator_4F0590(int a1);
uint32_t* nox_server_rewardgen_activateMarker_4F0720(int a1, unsigned int a2);
uint32_t* nox_xxx_rewardSpellBook_4F09F0(int a1, unsigned int a2);
int nox_server_rewardGen_pickRandomSlots_4F0B60(unsigned int a1);
uint32_t* nox_xxx_rewardAbilityBook_4F0C70(int a1);
uint32_t* nox_xxx_rewardFieldGuide_4F0D20(int a1, unsigned int a2);
uint32_t* nox_xxx_rewardMakeArmor_4F0E80(int a1, unsigned int a2);
int nox_xxx_rewardMakeWeapon_4F14E0(int a1, unsigned int a2);
uint32_t* nox_xxx_rewardMakePotion_4F1C40(int a1, unsigned int a2);
uint32_t* nox_xxx_createGem_4F1D30(int a1, unsigned int a2);
uint32_t* nox_xxx_createGem2_4F1F00(int a1, unsigned int a2);
void sub_4F2110();
int sub_4F2210();
int sub_4F24E0(int a1);
int sub_4F2530(int a1);
int sub_4F2570(int a1);
int sub_4F2590(int a1);
int sub_4F2700(int a1);
int sub_4F27A0(int a1);
int sub_4F27E0(int a1);
int sub_4F28C0(int a1);
int sub_4F2960(int a1);
int sub_4F2B20(int a1);
int sub_4F2B60(int a1);
int sub_4F2C30(int a1);
int nox_xxx_spell_4F2E70(int a1);
int sub_4F2EF0(int a1);
int nox_xxx_playerTryEquip_4F2F70(nox_object_t* a1, nox_object_t* item);
int nox_xxx_playerTryDequip_4F2FB0(nox_object_t* a1, const nox_object_t* object);
int nox_xxx_itemApplyEngageEffect_4F2FF0(nox_object_t* item, int a2);
int nox_xxx_itemApplyDisengageEffect_4F3030(const nox_object_t* object, int a2);
void nox_xxx_inventoryPutImpl_4F3070(nox_object_t* a1, nox_object_t* item, int a3);
bool nox_xxx_playerCheckStrength_4F3180(nox_object_t* a1, nox_object_t* item);
int nox_xxx_pickupDefault_4F31E0(nox_object_t* a1p, nox_object_t* item, int a3);
int nox_xxx_pickupFood_4F3350(int a1, int a2, int a3);
int sub_4F3400(int a1, int a2, int a3);
int nox_xxx_pickupUse_4F34D0(int a1, int a2, int a3);
int nox_xxx_pickupTrap_4F3510(int a1, int a2, int a3);
int nox_xxx_pickupTreasure_4F3580(int a1, int a2, int a3);
int nox_xxx_inventoryServPlace_4F36F0(nox_object_t* a1p, nox_object_t* a2p, int a3, int a4);
int nox_xxx_pickupPotion_4F37D0(nox_object_t* a1, nox_object_t* a2, int a3);
int nox_xxx_pickupAmmo_4F3B00(int a1, nox_object_t* item, int a3, int a4);
int nox_xxx_pickupSpellbook_4F3C60(int a1, int a2, int a3);
int nox_xxx_pickupAbilitybook_4F3CE0(int a1, int a2, int a3);
int sub_4F3DD0(int a1, int a2);
int nox_xxx_xfer_4F3E30(unsigned short a1, nox_object_t* a2, int a3);
int nox_xxx_servMapLoadPlaceObj_4F3F50(nox_object_t* a1, int a2, void* a3);
char sub_4F40A0(nox_object_t* a1);
int nox_xxx_readObjectOldVer_4F4170(int a1, int a2, int a3);
int nox_xxx_mapReadWriteObjData_4F4530(nox_object_t* a1, int a2);
int nox_xxx_XFerDefault_4F49A0(nox_object_t* a1, void* a2);
int nox_xxx_XFerSpellPagePedistal_4F4A20(int a1);
int nox_xxx_XFerReadable_4F4AB0(int a1);
int nox_xxx_XFerExit_4F4B90(int a1);
int nox_xxx_XFerDoor_4F4CB0(int a1);
int nox_xxx_unitTriggerXfer_4F4E50(float a1);
int nox_xxx_XFerHole_4F51D0(int a1);
int nox_xxx_XFerTransporter_4F5300(int a1);
int nox_xxx_XFerElevator_4F53D0(int a1);
int nox_xxx_XFerElevatorShaft_4F54A0(int a1);
int sub_4F5540(int a1);
int nox_xxx_XFerMover_4F5730(int a1);
int nox_xxx_XFerGlyph_4F5890(int a1);
int nox_xxx_XFerInvLight_4F5AA0(int* a1);
int nox_xxx_XFerSentry_4F5E50(int a1);

#endif // NOX_PORT_GAME3_3
