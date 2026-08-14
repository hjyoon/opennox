package server

import "unsafe"

type UnusedManaTransferRuntime4EAD50 struct {
	TeamContains  func(*ObjectTeam, TeamID) int32
	AddPlayerMana func(*Object, int16) uint16
}

type unusedManaTransferNativeDeps4EAD50 struct {
	findTeamByID  func(TeamID) *Team
	teamContains  func(*ObjectTeam, TeamID) int32
	addPlayerMana func(*Object, int16) uint16
}

func unusedManaTransferNative4EAD50(
	source, target *Object,
	deps unusedManaTransferNativeDeps4EAD50,
) {
	unusedManaTransfer4EAD50(source, target, unusedManaTransferHooks4EAD50[
		*Object,
		*PlayerUpdateData,
		*ObeliskUpdateData,
		*Team,
	]{
		loadClassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadPlayerUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadManaCurrent: func(update *PlayerUpdateData) uint16 {
			return update.ManaCur
		},
		loadSourceUpdate: func(obj *Object) *ObeliskUpdateData {
			return (*ObeliskUpdateData)(obj.UpdateData)
		},
		loadManaMax: func(update *PlayerUpdateData) uint16 {
			return update.ManaMax
		},
		loadSourceMana: func(update *ObeliskUpdateData) int32 {
			return update.Mana
		},
		hasTeam: func(obj *Object) int32 {
			if obj.TeamVal.Has() {
				return 1
			}
			return 0
		},
		loadObjectTeamID: func(obj *Object) uint8 {
			return uint8(obj.TeamVal.ID)
		},
		findTeamByID: func(id uint8) *Team {
			return deps.findTeamByID(TeamID(id))
		},
		loadTeamID: func(team *Team) uint8 {
			return uint8(team.IDVal)
		},
		teamContains: func(obj *Object, id uint8) int32 {
			return deps.teamContains(&obj.TeamVal, TeamID(id))
		},
		addPlayerMana: deps.addPlayerMana,
		storeSourceMana: func(update *ObeliskUpdateData, mana int32) {
			update.Mana = mana
		},
	})
}

// UnusedManaTransfer4EAD50 exposes the unreferenced original routine without
// inventing a registration or C callback edge. The two runtime hooks preserve
// the legacy team-membership and player-mana side effects while all object,
// player-update, obelisk-update and team pointers remain native-width.
func (s *Server) UnusedManaTransfer4EAD50(
	source, target *Object,
	runtime UnusedManaTransferRuntime4EAD50,
) {
	unusedManaTransferNative4EAD50(source, target, unusedManaTransferNativeDeps4EAD50{
		findTeamByID: func(id TeamID) *Team {
			return s.Teams.ByID(id)
		},
		teamContains:  runtime.TeamContains,
		addPlayerMana: runtime.AddPlayerMana,
	})
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(ObeliskUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ObeliskUpdateData{}.Mana)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(PlayerUpdateData{}.ManaCur)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(PlayerUpdateData{}.ManaMax)]
)
