#ifndef NOX_PORT_GAME4
#define NOX_PORT_GAME4

#include "defs.h"
#include "map_object_list_5048a0.h"
#include "xfer_spell_reward_4f5f30.h"
#include "xfer_ability_reward_4f6240.h"
#include "xfer_field_guide_4f6390.h"
#include "xfer_weapon_4f64a0.h"
#include "xfer_armor_4f6860.h"
#include "xfer_ammo_4f6b20.h"
#include "xfer_team_4f6d20.h"
#include "xfer_gold_4f6ec0.h"
#include "xfer_obelisk_4f6f60.h"
#include "xfer_toxic_cloud_4f70a0.h"
#include "xfer_monster_generator_4f7130.h"
#include "xfer_reward_marker_4f74d0.h"
#include "fixed_rng_seed_4f78d0.h"
#include "inventory_lookup_4f78e0.h"
#include "player_confused_direction_4f7a40.h"
#include "map_find_player_start_4f7ab0.h"
#include "player_sub_stamina_4f7d30.h"
#include "player_adjust_stamina_4f7db0.h"
#include "wink_game_ball_release_4f7df0.h"
#include "weapon_stamina_by_type_4f7e80.h"
#include "player_respawn_4f7ef0.h"
#include "fixed_rng_seed_4fb940.h"

void sub_4F7950(nox_object_t* a1);
void nox_xxx_playerSetCustomWP_4F79A0(nox_object_t* unit, float x, float y);
void nox_xxx_updatePlayer_4F8100(nox_object_t* a1);
int sub_4F9A80(nox_object_t* a1);
int sub_4F9AB0(nox_object_t* a1);
int nox_xxx_playerCanMove_4F9BC0(nox_object_t* a1);
int nox_server_playerCanMove_4F9BC0(nox_object_t* unit);
int nox_xxx_playerCanAttack_4F9C40(nox_object_t* a1);
int nox_server_playerCanAttack_4F9C40(nox_object_t* unit);
void nox_xxx_playerInputAttack_4F9C70(nox_object_t* a1);
int nox_xxx_playerAimsAtEnemy_4F9DC0(nox_object_t* player);
void nox_xxx_animPlayerGetFrameRange_4F9F90(int a1, int* a2, int* a3);
int nox_xxx_unitGetStrength_4F9FD0(nox_object_t* obj);
int nox_xxx_playerSetState_4FA020(nox_object_t* a1, int a2);
int sub_4FA280(int a1);
int nox_common_mapPlrActionToStateId_4FA2B0(nox_object_t* a1);
int nox_xxx_checkInversionEffect_4FA4F0(int a1, int a2);
uint32_t* nox_xxx_playerAddGold_4FA590(int a1, int a2);
uint32_t* nox_xxx_playerSubGold_4FA5D0(int a1, unsigned int a2);
void nox_object_setGold_4FA620(nox_object_t* a1, int a2);
int nox_xxx_playerGetGold_4FA6B0(int a1);
int nox_object_getGold_4FA6D0(nox_object_t* a1);
int nox_xxx_playerBotCreate_4FA700(nox_object_t* a1);
char nox_xxx_mobMorphFromPlayer_4FAAC0(uint32_t* a1);
char nox_xxx_mobMorphToPlayer_4FAAF0(uint32_t* a1);
int nox_xxx_updatePlayerMonsterBot_4FAB20(uint32_t* a1);
char nox_xxx_monsterActionToPlrState_4FABC0(int a1);
int nox_xxx_respawnPlayerBot_4FAC70(int a1);
int nox_xxx_netSendRewardNotify_4FAD50(nox_object_t* a1, int a2, nox_object_t* a3, char a4);
void sub_4FADD0(int a1, const char* a2, char a3);
int sub_4FB000(int a1, int a2);
int sub_4FB050(int a1, int a2, int* a3);
int nox_xxx_playerDoSchedSpell_4FB0E0(nox_object_t* a1, nox_object_t* a2);
int nox_xxx_playerDoSchedSpellQueue_4FB1D0(nox_object_t* a1, nox_object_t* a2);
int32_t sub_4FBE60(nox_object_t* a1, int32_t a2);
int32_t sub_4FBEA0(nox_object_t* a1, int32_t a2, int32_t a3);
int32_t sub_4FC030(nox_object_t* a1, int32_t a2);
void sub_4FC070(nox_object_t* a1, int32_t a2, int32_t a3);
void sub_4FC0B0(nox_object_t* a1, int32_t a2);
void nox_xxx_playerCancelAbils_4FC180(nox_object_t* a1);
int32_t nox_common_playerIsAbilityActive_4FC250(nox_object_t* a1, int32_t a2);
void sub_4FC300(nox_object_t* a1, int32_t a2);
int nox_xxx_probablyWarcryCheck_4FC3E0(nox_object_t* a1, int a2);
void sub_4FC440(nox_object_t* a1, int a2);
void sub_4FC670(int a1);
int sub_4FC960(int a1, char a2);
int nox_xxx_Fn_4FCAC0(int a1, int a2);
void nox_xxx_spellCastByBook_4FCB80();
int sub_4FCEB0(int a1);
int nox_xxx_spellCheckSmth_4FCEF0(int a1, int* a2, int a3);
int sub_4FCF90(nox_object_t* a1, int a2, int a3);
unsigned short sub_4FD030(int a1, short a2);
void nox_xxx_teleportAllPixies_4FD090(nox_object_t* a1);
int sub_4FD0E0(nox_object_t* a1, int a2);
int nox_xxx_checkPlrCantCastSpell_4FD150(nox_object_t* a1, int a2, int a3);
int nox_xxx_spellAccept_4FD400(int a1, nox_object_t* a2, nox_object_t* a3p, nox_object_t* a4p, void* a5p, int a6);
int nox_xxx_castSpellByUser_4FDD20(int a1, nox_object_t* a2, void* a3);
uint32_t* nox_xxx_createSpellFly_4FDDA0(nox_object_t* a1, nox_object_t* a2, int a3);
void nox_xxx_collide_4FDF90(nox_object_t* a1, nox_object_t* a2);
int nox_xxx_spellGetPhoneme_4FE1C0(int a1, char a2);
int nox_xxx_spellByBookInsert_4FE340(int a1, int* a2, int a3, int a4, int a5);
void nox_xxx_spell_4FE680(nox_object_t* a1, float a2);
int nox_xxx_spellGetPower_4FE7B0(int a1, nox_object_t* a2);
void sub_4FE8A0(int a1);
void* nox_xxx_spellCastedFirst_4FE930();
void* nox_xxx_spellCastedNext_4FE940(void* a1);
nox_object_t* nox_xxx_spellCastedCaster_native(void* a1);
nox_object_t* nox_xxx_spellCastedTarget_native(void* a1);
int nox_xxx_spellCastedSpell_native(void* a1);
void sub_4FE980(void* a1);
void nox_xxx_spellCancelSpellDo_4FE9D0(void* a1);
int sub_4FEA70(int a1, float2* a2);
int nox_xxx_playerCancelSpells_4FEAE0(nox_object_t* a1);
void nox_xxx_spellCancelDurSpell_4FEB10(int a1, nox_object_t* a2);
void sub_4FEB60(nox_object_t* owner, const nox_object_t* item);
void nox_xxx_cancelAllSpells_4FEE90(nox_object_t* a1);
void nox_xxx_netStopRaySpell_4FEF90(void* a1, nox_object_t* a2);
char* nox_xxx_netStartDurationRaySpell_4FF130(int a1);
int sub_4FF2D0(int a1, int a2);
int nox_xxx_testUnitBuffs_4FF350(nox_object_t* unit, char buff);
void nox_xxx_buffApplyTo_4FF380(nox_object_t* unit, int buff, short dur, char power);
int nox_xxx_unitGetBuffTimer_4FF550(nox_object_t* unit, int buff);
char nox_xxx_buffGetPower_4FF570(nox_object_t* unit, int buff);
void nox_xxx_unitClearBuffs_4FF580(nox_object_t* unit);
int nox_xxx_spellBuffOff_4FF5B0(nox_object_t* a1, int a2);
void nox_xxx_updateUnitBuffs_4FF620(nox_object_t* a1);
char* nox_xxx_journalQuestSet_500540(char* a1, int a2);
char* nox_xxx_scriptGetJournal_5005E0(char* a1);
char* nox_xxx_journalQuestSetBool_5006B0(char* a1, int a2);
int sub_500750(char* a1);
double sub_500770(char* a1);
void sub_500790(void* lpMem);
char* sub_5007E0(char* a1);
unsigned int sub_5009B0(char* a1);
int sub_500A60();
int sub_500B70();
int nox_xxx_orderUnitLocal_500C70(int owner, int orderType);
int sub_500CA0(int a1, int a2);
int nox_xxx_creatureIsMonitored_500CC0(nox_object_t* a1, nox_object_t* a2);
bool nox_xxx_checkSummonedCreaturesLimit_500D70(nox_object_t* a1, int a2);
int nox_xxx_summonStart_500DA0(int a1);
int sub_500F40(int a1, float a2);
int nox_xxx_summonFinish_5010D0(int a1);
void nox_xxx_summonCancel_5011C0(int a1);
int nox_xxx_charmCreature1_5011F0(int* a1);
int nox_xxx_charmCreatureFinish_5013E0(int* a1);
int nox_xxx_charmCreature2_501690(int a1);
void nox_xxx_banishUnit_5017F0(int unit);
int nox_xxx_getSevenDwords3_501940(int a1);
void nox_xxx_aud_501960(int a1, nox_object_t* a2, int a3, int a4);
void nox_xxx_audCreate_501A30(int a1, float2* a2, int a3, int a4);
void nox_xxx_gameSetAudioFadeoutMb_501AC0(int a1);
char sub_501C00(float* a1, nox_object_t* a2);
void nox_xxx_netUpdateRemotePlr_501CA0(nox_object_t* a1);
int nox_xxx_mapgenMakeScript_502790(FILE* a1, char* a2);
void nox_xxx_mapReset_5028E0();
int sub_5029A0(char* a1);
int sub_5029F0(int a1);
int sub_502A20();
int sub_502A50(char* a1);
int sub_502AB0(char* a1);
int sub_502B10();
int sub_502D70(int a1);
FILE* sub_502DA0(char* a1);
FILE* sub_502DF0();
FILE* sub_502E10(int a1);
double sub_502E70(int a1);
double sub_502EA0(int a1);
int nox_xxx_mapgenSaveMap_503830(int a1);
int sub_503B30(float2* a1);
int sub_503EC0(int a1, float* a2);
void nox_xxx_free_503F40();
nox_tile_coord_entry_t* nox_xxx_tileAllocTileInCoordList_5040A0(int a1, int a2, float a3);
int nox_xxx_tileInit_504150(int a1, int a2);
uint32_t* sub_504290(char a1, char a2);
uint32_t* nox_xxx_cliWallGet_5042F0(int a1, int a2);
int sub_504330(int a1, int a2);
uint32_t* sub_5044B0(int a1, float a2, float a3);
int sub_504560(int a1, int a2);
void sub_504600(char* a1, unsigned int a2, unsigned char a3);
int sub_5046A0(uint32_t* a1, unsigned int a2);
int sub_504720(unsigned int a1, unsigned int a2);
// GAME.EXE 005048A0..00504AA3 is restored by map_object_list_5048a0.c.
// The dedicated header above preserves the three native pointers in each
// temporary node and the native object pointers returned by the traversals.
void* sub_505060();
int nox_server_mapRWMapIntro_505080();
int nox_server_mapRWGroupData_505C30();
int nox_server_mapRWWaypoints_506260(uint32_t* a1);
int nox_xxx_allocVoteArray_5066D0();
int sub_506720();
int sub_506740(nox_object_t* a1);
void sub_5067B0(int a1);
int sub_506810(int a1);
int nox_xxx_netSendVote_506840(int a1);
char sub_506870(int a1, int a2, wchar2_t* a3);
char sub_5068E0(int a1, int a2, wchar2_t* a3);
uint32_t* sub_506A20(int a1, int a2);
int nox_xxx_voteAddMB_506AD0(int a1);
uint32_t* sub_506B00(int a1, int a2);
uint32_t* sub_506B80(int a1, int a2, wchar2_t* a3);
void sub_506C90(int a1, int a2, wchar2_t* a3);
void sub_506D00(int a1, wchar2_t* a2);
void sub_506DE0(int a1);
void sub_506E50(int a1, wchar2_t* a2);
void sub_506F80(int a1);
int sub_507000(int a1);
void sub_507090(int a1);
void sub_507100(int a1);
int sub_507190(int a1, char a2);
int sub_5071C0();
void sub_509120(uint32_t* a1, int a2, const char* a3);
int sub_5095E0();
int sub_5096F0();

void nox_server_scriptExecuteFnForEachGroupObj_502670(unsigned char* groupPtr, int expectedType, void (*a3)(int, int),
													  int a4);

#endif // NOX_PORT_GAME4
