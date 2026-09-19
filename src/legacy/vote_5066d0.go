package legacy

/*
#include "vote_5066d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func voteRecordAlloc5066D0() unsafe.Pointer {
	return unsafe.Pointer(C.nox_vote_record_alloc_5066D0())
}

func voteRecordAdd506AD0(vote unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_voteAddMB_506AD0((*C.nox_vote_5066D0)(vote)))
}

func voteRecordCreate506A20(typ int, player *server.Object) unsafe.Pointer {
	return unsafe.Pointer(C.sub_506A20(C.int(typ), asObjectC(player)))
}

func voteHead5066D0() unsafe.Pointer {
	return unsafe.Pointer(C.nox_vote_head_5066D0())
}

func voteAllocator5066D0() unsafe.Pointer {
	return C.nox_vote_allocator_5066D0()
}

func voteRecordSize5066D0() uintptr {
	return uintptr(C.nox_vote_record_size_5066D0())
}

func voteNext5066D0(vote unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_vote_next_5066D0((*C.nox_vote_5066D0)(vote)))
}

func votePrevious5066D0(vote unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_vote_previous_5066D0((*C.nox_vote_5066D0)(vote)))
}

func voteTeam5066D0(vote unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_vote_team_5066D0((*C.nox_vote_5066D0)(vote)))
}

func voteFrame5066D0(vote unsafe.Pointer) uint32 {
	return uint32(C.nox_vote_frame_5066D0((*C.nox_vote_5066D0)(vote)))
}

func voteVoters5066D0(vote unsafe.Pointer) uint32 {
	return uint32(C.nox_vote_voters_5066D0((*C.nox_vote_5066D0)(vote)))
}

func voteCount5066D0(vote unsafe.Pointer) uint8 {
	return uint8(C.nox_vote_count_5066D0((*C.nox_vote_5066D0)(vote)))
}

func voteThreshold5066D0(vote unsafe.Pointer) uint8 {
	return uint8(C.nox_vote_threshold_5066D0((*C.nox_vote_5066D0)(vote)))
}

func voteSetVoters5066D0(vote unsafe.Pointer, voters uint32, count uint8) {
	C.nox_vote_set_voters_5066D0((*C.nox_vote_5066D0)(vote), C.uint32_t(voters), C.uint8_t(count))
}
