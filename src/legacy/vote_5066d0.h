#ifndef NOX_VOTE_5066D0_H
#define NOX_VOTE_5066D0_H

#include <stddef.h>
#include <stdint.h>

#include "common__system__team.h"
#include "defs.h"

typedef struct nox_vote_5066D0 {
	uint32_t type;
	uint8_t count;
	uint8_t reserved_5[3];
	uint32_t voters;
	uint8_t threshold;
	uint8_t reserved_13[3];
	nox_object_team_t* team;
	uint32_t team_mode;
	uint32_t frame;
	nox_object_t* target;
	uint32_t reserved_32[3];
	struct nox_vote_5066D0* next;
	struct nox_vote_5066D0* previous;
} nox_vote_5066D0;

_Static_assert(offsetof(nox_vote_5066D0, type) == 0,
	"wrong native offset of Vote type");
_Static_assert(offsetof(nox_vote_5066D0, count) == 4,
	"wrong native offset of Vote count");
_Static_assert(offsetof(nox_vote_5066D0, voters) == 8,
	"wrong native offset of Vote voter mask");
_Static_assert(offsetof(nox_vote_5066D0, threshold) == 12,
	"wrong native offset of Vote threshold");
_Static_assert(offsetof(nox_vote_5066D0, team) == 16,
	"wrong native offset of Vote team pointer");
_Static_assert(offsetof(nox_vote_5066D0, team_mode) == (sizeof(void*) == 4 ? 20 : 24),
	"wrong native offset of Vote team mode");
_Static_assert(offsetof(nox_vote_5066D0, frame) == (sizeof(void*) == 4 ? 24 : 28),
	"wrong native offset of Vote frame");
_Static_assert(offsetof(nox_vote_5066D0, target) == (sizeof(void*) == 4 ? 28 : 32),
	"wrong native offset of Vote target pointer");
_Static_assert(offsetof(nox_vote_5066D0, next) == (sizeof(void*) == 4 ? 44 : 56),
	"wrong native offset of Vote next pointer");
_Static_assert(offsetof(nox_vote_5066D0, previous) == (sizeof(void*) == 4 ? 48 : 64),
	"wrong native offset of Vote previous pointer");
_Static_assert(sizeof(nox_vote_5066D0) == (sizeof(void*) == 4 ? 52 : 72),
	"wrong native size of Vote record");

int nox_xxx_allocVoteArray_5066D0(void);
void sub_506700(void);
int sub_506720(void);
int sub_506740(nox_object_t* player_unit);
void sub_5067B0(nox_vote_5066D0* vote);
nox_vote_5066D0* sub_506810(nox_vote_5066D0* vote);
int nox_xxx_netSendVote_506840(int player_index);
char sub_506870(int type, nox_object_t* player_unit, wchar2_t* player_name);
char sub_5068E0(int type, nox_object_t* player_unit, wchar2_t* player_name);
nox_vote_5066D0* sub_506A20(int type, nox_object_t* player_unit);
nox_vote_5066D0* nox_xxx_voteAddMB_506AD0(nox_vote_5066D0* vote);
nox_vote_5066D0* sub_506B00(int type, nox_object_t* player_unit);
nox_vote_5066D0* sub_506B80(int type, nox_object_t* player_unit, wchar2_t* player_name);
void sub_506C90(int type, nox_object_t* player_unit, wchar2_t* player_name);
void sub_506D00(nox_object_t* player_unit, wchar2_t* player_name);
void sub_506DE0(nox_object_t* player_unit);
void sub_506E50(nox_object_t* player_unit, wchar2_t* player_name);
void nox_xxx_voteUptate_506F30(void);
void sub_506F80(nox_vote_5066D0* vote);
int sub_507000(nox_vote_5066D0* vote);
void sub_507090(nox_vote_5066D0* vote);
void sub_507100(nox_vote_5066D0* vote);
int sub_507190(int recipient, char active);
int sub_5071C0(void);

// Native-width record accessors used by the Go boundary and regression tests.
nox_vote_5066D0* nox_vote_record_alloc_5066D0(void);
nox_vote_5066D0* nox_vote_head_5066D0(void);
void* nox_vote_allocator_5066D0(void);
size_t nox_vote_record_size_5066D0(void);
nox_vote_5066D0* nox_vote_next_5066D0(nox_vote_5066D0* vote);
nox_vote_5066D0* nox_vote_previous_5066D0(nox_vote_5066D0* vote);
nox_object_team_t* nox_vote_team_5066D0(nox_vote_5066D0* vote);
uint32_t nox_vote_frame_5066D0(nox_vote_5066D0* vote);
uint32_t nox_vote_voters_5066D0(nox_vote_5066D0* vote);
uint8_t nox_vote_count_5066D0(nox_vote_5066D0* vote);
uint8_t nox_vote_threshold_5066D0(nox_vote_5066D0* vote);
void nox_vote_set_voters_5066D0(nox_vote_5066D0* vote, uint32_t voters, uint8_t count);

#endif // NOX_VOTE_5066D0_H
