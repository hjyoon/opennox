package legacy

/*
#include "common__system__team.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func assignTeamFlags(world *server.Server, rename bool) {
	for flag := world.Objs.First(); flag != nil; flag = flag.Next() {
		if !flag.Class().Has(object.ClassFlag) {
			continue
		}
		team := world.Teams.ByID(flag.TeamVal.ID)
		if team == nil {
			continue
		}
		index := server.TeamMaterialObjectIndex4ECBD0(flag)
		if rename {
			legacy := (*nox_team_t)(team.C())
			if title := nox_server_teamTitle_418C20(int32(index)); title != nil {
				sub_418800(&legacy.name[0], title, 1)
			}
			team.ColorInd = server.TeamColor(index)
			sub4184D0Native(team)
		} else {
			team.ColorInd = server.TeamColor(index)
		}
		team.Field_72 = unsafe.Pointer(flag)
	}
}

//export nox_xxx_wndGuiTeamCreate_4185B0
func nox_xxx_wndGuiTeamCreate_4185B0() C.int {
	world := GetServer().S()
	noxflags.SetGamePlay(noxflags.GameplayFlag4)
	world.Teams.Create(0)
	world.Teams.Create(0)
	assignTeamFlags(world, true)
	return 0
}

//export nox_xxx_teamAssignFlags_418640
func nox_xxx_teamAssignFlags_418640() C.int {
	assignTeamFlags(GetServer().S(), false)
	return 0
}

//export nox_xxx_toggleAllTeamFlags_418690
func nox_xxx_toggleAllTeamFlags_418690(on C.int) *C.char {
	world := GetServer().S()
	for team := world.Teams.First(); team != nil; team = world.Teams.Next(team) {
		if flag := (*nox_object_t)(team.Field_72); flag != nil {
			if on != 0 {
				nox_xxx_objectSetOn_4E75B0(flag)
			} else {
				nox_xxx_objectSetOff_4E7600(flag)
			}
		}
	}
	return nil
}
