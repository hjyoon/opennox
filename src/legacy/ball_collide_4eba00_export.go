package legacy

/*
#include "ball_collide_4eba00.h"

int sub_4196D0(void* object_team, void* team, int net_code, int flags);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const ballCollideFeedbackFrameOffset4EBA00 = uintptr(1568008)

func ballCollideRuntime4EBA00(s *server.Server) server.BallCollideRuntime4EBA00 {
	return server.BallCollideRuntime4EBA00{
		TeamMemberCount: func(team *server.Team) int32 {
			return flagPickupTeamEligible418BC0(s, team)
		},
		LoadFeedbackFrame: func() uint32 {
			return *memmap.PtrUint32(0x5D4594, ballCollideFeedbackFrameOffset4EBA00)
		},
		StoreFeedback: func(frame uint32) {
			*memmap.PtrUint32(0x5D4594, ballCollideFeedbackFrameOffset4EBA00) = frame
		},
		ChangeTeam: func(value *server.ObjectTeam, team *server.Team, netCode uint32, flags int32) int32 {
			// 004196D0 still owns the broader team-list/network mutation unit.
			// Its native-pointer restoration is tracked separately from this
			// callback boundary.
			return int32(C.sub_4196D0(value.C(), team.C(), C.int(netCode), C.int(flags)))
		},
		CreateTeam: func(id server.TeamID, value *server.ObjectTeam, active int32, netCode uint32, flags int32) {
			// 004191D0 is likewise a separate team-service restoration unit.
			Nox_xxx_createAtImpl_4191D0(id, value, int(active), int(netCode), int(flags))
		},
		BallStatus: Sub_4E8290,
		BuffPurge:  flagPickupBuffPurgeRuntime4EA7A0(s),
	}
}

//export nox_xxx_collideBall_4EBA00
func nox_xxx_collideBall_4EBA00(
	ball, target *C.nox_object_t,
	collision *C.float,
) {
	s := GetServer().S()
	s.BallCollide4EBA00(
		asObjectS((*nox_object_t)(ball)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		ballCollideRuntime4EBA00(s),
	)
}
