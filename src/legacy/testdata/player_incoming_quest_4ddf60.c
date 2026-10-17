#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME5_2.h"
#include "common__system__team.h"
#include "player_incoming_quest_4ddf60.h"

static nox_playerInfo player;
static nox_object_t unit;
static _Alignas(8) uint8_t update[656];
static uint8_t expected_update[sizeof(update)];
static unsigned game_flags;
static int player_index, initial_unit, helper_calls, quest_prepare_calls, init_calls;
uint32_t dword_5d4594_2649712;
uint32_t nox_server_sendMotd_108752;

static size_t pending_offset(void) { return sizeof(void*) == 8 ? 648 : 552; }
static uint32_t pending_value(void) {
    uint32_t value;
    memcpy(&value, update + pending_offset(), sizeof(value));
    return value;
}
static int wants_pending(void) { return (game_flags & 4096) && player_index != 31 && initial_unit; }
static void check_update(void) {
    if (memcmp(update, expected_update, sizeof(update)) != 0) {
        fprintf(stderr, "incoming Quest update mismatch: pointer_bytes=%zu pending_offset=%zu pending=%08x raw552=%02x%02x%02x%02x index=%d initial_unit=%d flags=%x\n",
            sizeof(void*), pending_offset(), pending_value(), update[552], update[553], update[554], update[555], player_index, initial_unit, game_flags);
        abort();
    }
}

/* Native tail storage is owned by Go in production. The companion CGo test
 * verifies this helper against the complete native PlayerUpdateData value. */
void nox_xxx_playerIncomingQuest_native_4DDF60(nox_object_t* object) {
    assert(object == &unit && object->data_update == update && wants_pending());
    uint32_t one = 1;
    memcpy(update + pending_offset(), &one, sizeof(one));
    helper_calls++;
}
bool nox_common_gameFlags_check_40A5C0(unsigned flags) { return (game_flags & flags) != 0; }
nox_playerInfo* nox_common_playerInfoFromNum_417090(int index) { assert(index == player_index); return &player; }
int sub_4D9CF0(int index) {
    assert(index == player_index && (game_flags & 4096));
    /* The pending store precedes this callback. A missing unit is legal at
     * that guard; attach the fixture's unit here for the ordinary later tail. */
    check_update();
    player.playerUnit = &unit;
    quest_prepare_calls++;
    return 0;
}
int sub_4D6000(nox_object_t* object) { assert(object == &unit); check_update(); return 0; }
int sub_57B920(void* data) { assert(data == player.netData16); check_update(); return 0; }
char* nox_xxx_newPlayerSendAllPlayers_4DE300(int index) { assert(index == player_index); return NULL; }
static void init_unit(nox_object_t* object, uint32_t arg) { assert(object == &unit && arg == 0); check_update(); init_calls++; }
void nox_xxx_playerSendMOTD_4DD140(int index) { abort(); }
nox_playerInfo* nox_common_playerInfoGetFirst_416EA0(void) { return NULL; }
nox_playerInfo* nox_common_playerInfoGetNext_416EE0(nox_playerInfo* p) { abort(); }
void nox_xxx_netMarkMinimapObject_417190(int index, nox_object_t* object, unsigned arg) { abort(); }
int nox_xxx_netSendSimpleObject2_4DF360(int index, nox_object_t* object) { abort(); }
int nox_xxx_netSendTeam_4D8670(int index, uint32_t* object) { abort(); }
int nox_xxx_servMinimapRevealFlag_4DE380(int index) { assert(index == player_index); return 0; }
void sub_4DF2E0(int index) { assert(index == player_index); }
int sub_40AA70(nox_playerInfo* p) { abort(); }
int nox_xxx_netNeedTimestampStatus_4174F0(nox_playerInfo* p, int mask) { abort(); }
nox_object_t* sub_4E8110(int32_t index) { assert(index == player_index); return &unit; }
nox_team_t* nox_server_teamFirst_418B10(void) { abort(); }
nox_team_t* nox_server_teamNext_418B60(nox_team_t* team) { abort(); }
nox_team_flag_status_t* sub_4E8320(uint8_t id) { abort(); }
int nox_xxx_netSendFlagStatus_4D95A0(int index, char id, char state, char flag_index, short carrier) { abort(); }
nox_game_ball_status_t* sub_4E8310(void) { abort(); }
int nox_xxx_netSendBallStatus_4D95F0(int index, char state, short code) { abort(); }
char* nox_xxx_sendAllClientStatus_4175C0(nox_playerInfo* p) { assert(p == &player); return NULL; }
int sub_409F40(int mode) { assert(mode == 0x2000); return 0; }
int nox_xxx_sendAllPlayerIDs_4DE270(nox_playerInfo* p) { abort(); }
nox_object_t* nox_xxx_getFirstPlayerUnit_4DA7C0(void) { return NULL; }
nox_object_t* nox_xxx_getNextPlayerUnit_4DA7F0(const nox_object_t* object) { abort(); }
int sub_4D9D20(int index, nox_object_t* object) { abort(); }
int sub_4D7A60(int index) { assert(index == player_index && (game_flags & 4096)); return 0; }

/* COMPLETE_INCOMING_BODY */

static void run_case(unsigned flags, int index, int present, unsigned pattern) {
    game_flags = flags;
    player_index = index;
    initial_unit = present;
    helper_calls = quest_prepare_calls = init_calls = 0;
    memset(&player, 0, sizeof(player));
    memset(&unit, 0, sizeof(unit));
    memset(update, pattern, sizeof(update));
    uint32_t sentinel = UINT32_C(0x87654321);
    memcpy(update + pending_offset(), &sentinel, sizeof(sentinel));
    memcpy(expected_update, update, sizeof(update));
    if (wants_pending()) {
        uint32_t one = 1;
        memcpy(expected_update + pending_offset(), &one, sizeof(one));
    }
    player.playerUnit = present ? &unit : NULL;
    player.playerInd = index;
    player.field_4700 = 99;
    player.field_3676 = 88;
    unit.data_update = update;
    unit.func_init = (void*)init_unit;
    unit.x = 128.5f;
    unit.y = -63.25f;
    unit.net_code = 0xabcd;
    nox_server_sendMotd_108752 = 0;
    dword_5d4594_2649712 = 0x02468000;
    if (sizeof(void*) == 8) assert((uintptr_t)&unit > UINT32_MAX && (uintptr_t)update > UINT32_MAX);
    assert(nox_xxx_netPlayerIncomingServ_4DDF60(index) == 0xabcd);
    check_update();
    assert(helper_calls == wants_pending());
    assert(quest_prepare_calls == ((flags & 4096) != 0) && init_calls == 1);
    assert(player.field_4700 == 0 && player.field_3676 == 3);
    assert(player.pos_x_3632 == unit.x && player.pos_y_3636 == unit.y);
}
int main(void) {
    unsigned patterns[] = {0x5a, 0xa5, 0xff};
    int cases = 0;
    for (unsigned p = 0; p < sizeof(patterns)/sizeof(patterns[0]); p++) {
        for (int host = 0; host < 2; host++) {
            int index = host ? 31 : 7;
            run_case(4096, index, 1, patterns[p]); cases++;
            run_case(4096, index, 0, patterns[p]); cases++;
            run_case(0, index, 1, patterns[p]); cases++;
        }
    }
    printf("complete incoming Quest body: %d native cases; pending store and all guard controls preserve surrounding bytes\n", cases);
    return 0;
}
