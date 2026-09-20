package server

import "github.com/opennox/libs/object"

type kotrMapSetupHooks4180D0[O, T comparable] struct {
	firstObject       func() O
	nextObject        func(O) O
	isFlag            func(O) bool
	delayedDelete     func(O)
	firstTeam         func() T
	nextTeam          func(T) T
	clearTeamFlag     func(T)
	isCrown           func(O) bool
	clearPickupTarget func(O)
	teamMode          bool
	teamID            func(O) TeamID
	respawnRemove     func(O)
	markMinimapForAll func(O)
	teamByID          func(TeamID) T
	setTeamFlag       func(T, O)
}

// mapInfoSetKotr4180D0 preserves the two object walks and team reset order of
// GAME.EXE 004180D0 while keeping objects and teams as typed native pointers.
func mapInfoSetKotr4180D0[O, T comparable](hooks kotrMapSetupHooks4180D0[O, T]) bool {
	var zeroObject O
	for obj := hooks.firstObject(); obj != zeroObject; {
		next := hooks.nextObject(obj)
		if hooks.isFlag(obj) {
			hooks.delayedDelete(obj)
		}
		obj = next
	}

	var zeroTeam T
	for team := hooks.firstTeam(); team != zeroTeam; team = hooks.nextTeam(team) {
		hooks.clearTeamFlag(team)
	}

	crowns := 0
	for obj := hooks.firstObject(); obj != zeroObject; {
		next := hooks.nextObject(obj)
		if hooks.isCrown(obj) {
			crowns++
			hooks.clearPickupTarget(obj)
			teamID := hooks.teamID(obj)
			if !hooks.teamMode {
				if teamID != 0 {
					hooks.delayedDelete(obj)
					hooks.respawnRemove(obj)
				} else {
					hooks.markMinimapForAll(obj)
				}
			} else if teamID == 0 {
				hooks.delayedDelete(obj)
				hooks.respawnRemove(obj)
			} else if team := hooks.teamByID(teamID); team != zeroTeam {
				hooks.setTeamFlag(team, obj)
				hooks.markMinimapForAll(obj)
			}
		}
		obj = next
	}
	return crowns != 0
}

// KOTRMapSetupRuntime4180D0 supplies lifecycle and network operations owned
// by the root server package.
type KOTRMapSetupRuntime4180D0 struct {
	DelayedDelete     func(*Object)
	RespawnRemove     func(*Object)
	MarkMinimapForAll func(*Object)
}

// MapInfoSetKotr4180D0 initializes KOTR without routing Object or Team
// pointers through the original PE32 integer ABI.
func (s *Server) MapInfoSetKotr4180D0(
	crownType uint16,
	teamMode bool,
	runtime KOTRMapSetupRuntime4180D0,
) bool {
	return mapInfoSetKotr4180D0(kotrMapSetupHooks4180D0[*Object, *Team]{
		firstObject: s.Objs.First,
		nextObject: func(obj *Object) *Object {
			return obj.Next()
		},
		isFlag: func(obj *Object) bool {
			return obj.Class().Has(object.ClassFlag)
		},
		delayedDelete: runtime.DelayedDelete,
		firstTeam:     s.Teams.First,
		nextTeam:      s.Teams.Next,
		clearTeamFlag: func(team *Team) {
			s.Teams.SetTeamFlag(team, nil)
		},
		isCrown: func(obj *Object) bool {
			return obj.TypeInd == crownType
		},
		clearPickupTarget: func(obj *Object) {
			(*CrownUpdateData)(obj.UpdateData).PickupTarget = nil
		},
		teamMode: teamMode,
		teamID: func(obj *Object) TeamID {
			return obj.TeamVal.ID
		},
		respawnRemove:     runtime.RespawnRemove,
		markMinimapForAll: runtime.MarkMinimapForAll,
		teamByID:          s.Teams.ByID,
		setTeamFlag:       s.Teams.SetTeamFlag,
	})
}
