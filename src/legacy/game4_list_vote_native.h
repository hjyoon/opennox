#ifndef NOX_GAME4_LIST_VOTE_NATIVE_H
#define NOX_GAME4_LIST_VOTE_NATIVE_H
#include <stdint.h>
#include <stdlib.h>
#include "map_temp_lists_503f40.h"
#include "map_object_list_5048a0.h"
#include "vote_5066d0.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5_2.h"
#include "player_respawn_4f7ef0.h"
#include "common/alloc/classes/alloc_class.h"

extern uint32_t dword_5d4594_3835368;
extern void* dword_5d4594_1599532;
extern nox_tile_coord_entry_t* dword_5d4594_1599556;
extern void* dword_5d4594_1599548;
extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_1599480;
extern void* dword_5d4594_1599540;
extern uint32_t dword_5d4594_3835396;
extern uint32_t nox_server_kickQuestPlayerMinVotes_229992;
extern uint32_t nox_server_resetQuestMinVotes_229988;
extern void* nox_alloc_vote_1599652;
extern uint32_t dword_5d4594_1599656;
void sub_579E90(nox_waypoint_t* waypoint);

// Only typed external-service forwards. Keep the old K&R names out of new
// cgo call identities; all new private entry prototypes have explicit (void).
static inline nox_waypoint_t* game4_waypoint_alloc_native(void) { return sub_579E70(); }
static inline char* game4_wall_size_native(void) { return nox_xxx_mapGetWallSize_426A70(); }
static inline nox_playerInfo* game4_players_first_native(void) { return nox_common_playerInfoGetFirst_416EA0(); }
static inline nox_object_t* game4_units_first_native(void) { return nox_xxx_getFirstPlayerUnit_4DA7C0(); }
static inline char* game4_quest_map_native(void) { return nox_xxx_getQuestMapFile_4D0F60(); }
#endif
