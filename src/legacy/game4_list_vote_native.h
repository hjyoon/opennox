#ifndef NOX_GAME4_LIST_VOTE_NATIVE_H
#define NOX_GAME4_LIST_VOTE_NATIVE_H
#include <stdint.h>
#include <stddef.h>
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
extern uint32_t dword_5d4594_3835396;
typedef struct {
	void* allocator;
	uint32_t diagnostic;
	uint32_t minimum_kick;
	uint32_t minimum_reset;
} nox_vote_state_t;
_Static_assert(offsetof(nox_vote_state_t, diagnostic) == sizeof(void*), "vote word offset");
nox_vote_state_t* nox_vote_state_storage(void);

void sub_579E90(nox_waypoint_t* waypoint);

#endif
