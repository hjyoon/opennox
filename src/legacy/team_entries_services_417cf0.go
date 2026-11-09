package legacy

/*
#include <stdlib.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_3.h"
#include "common__system__team.h"
#include "client__gui__servopts__playrlst.h"
extern uint32_t dword_5d4594_526276;
static int32_t team_entries_limit_417cf0(void) { return (int32_t)dword_5d4594_526276; }
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

var teamEntriesDefaults417CF0 = teamEntriesRuntime417CF0{
	RemoveActive: func() int32 { return int32(GetServer().TeamsRemoveActive(false)) },
	Limit:        func() int32 { return int32(C.team_entries_limit_417cf0()) },
	Count:        func() byte { return byte(GetServer().S().Teams.Count()) },
	SetGameplay:  func(v uint32) { noxflags.SetGamePlay(noxflags.GameplayFlag(v)) },
	AutoAssign:   func(v int32) { TeamAutoAssign4181F0(v != 0) },
	FirstTeam:    func() *server.Team { return GetServer().S().Teams.First() },
	NextTeam:     func(t *server.Team) *server.Team { return GetServer().S().Teams.Next(t) },
	FirstUnit:    func() *server.Object { return GetServer().S().Players.FirstUnit() },
	NextUnit:     func(u *server.Object) *server.Object { return GetServer().S().Players.NextUnit(u) },
	ByID:         func(id server.TeamID) *server.Team { return GetServer().S().Teams.ByID(id) },
	ClientCode:   func() uint32 { return uint32(ClientPlayerNetCode()) },
	NoRendering:  func() bool { return noxflags.HasEngine(noxflags.EngineNoRendering) },
	Change: func(value *server.ObjectTeam, team *server.Team, code int32) {
		Sub_4196D0(value.C(), team.C(), int(code), 0)
	},
	Attach: func(id server.TeamID, value *server.ObjectTeam, code int32) {
		Nox_xxx_createAtImpl_4191D0(id, value, 1, int(code), 0)
	},
	Prepare: func(t *server.Team) { C.sub_457230((*C.wchar2_t)(t.C())) },
	Flag512: func() bool { return noxflags.HasGame(noxflags.GameFlag(512)) },
	Alloc:   func(n uintptr) unsafe.Pointer { return C.calloc(1, C.size_t(n)) },
	Free:    func(p unsafe.Pointer) { C.free(p) },
	Send: func(p unsafe.Pointer, n int) {
		C.nox_xxx_netSendPacket1_4E5390(159, p, C.int(n), nil, 1)
	},
}
