package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

type GameBallPlayerDamageRuntime4E1230 struct {
	LoadTypeCache  func() uint32
	StoreTypeCache func(uint32)
	ApplyForce     func(*Object, types.Pointf, float64)
	ChangeTeam     func(*ObjectTeam, *Team, uint32, int32) int32
	CreateTeam     func(TeamID, *ObjectTeam, int32, uint32, int32)
}

func (s *Server) GameBallOnPlayerDamage4E1230(source, target *Object, damage int32, r GameBallPlayerDamageRuntime4E1230) {
	gameBallPlayerDamage4E1230(source, target, damage, gameBallPlayerDamageHooks4E1230[*Object, *ObjectTeam, *Team]{
		loadClassLow:   func(obj *Object) uint8 { return uint8(obj.ObjClass) },
		loadTypeCache:  r.LoadTypeCache,
		lookupType:     func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		storeTypeCache: r.StoreTypeCache,
		firstOwned:     func(obj *Object) *Object { return obj.Field129 },
		nextOwned:      func(obj *Object) *Object { return obj.Field128 },
		loadType:       func(obj *Object) uint16 { return obj.TypeInd },
		loadFlags:      func(obj *Object) uint32 { return uint32(obj.ObjFlags) },
		storeFlags:     func(obj *Object, flags uint32) { obj.ObjFlags = object.Flags(flags) },
		applyForce:     func(victim, ball *Object, force float32) { r.ApplyForce(ball, victim.PosVec, float64(force)) },
		clearOwner:     s.ObjClearOwner,
		carrierState:   func(ball, victim *Object) { s.GameBallCarrierState4EB9B0(ball, victim) },
		loadTeam:       (*Object).TeamPtr,
		hasTeam:        (*ObjectTeam).Has,
		loadTeamID:     func(obj *Object) uint8 { return uint8(obj.TeamVal.ID) },
		findTeam:       func(id uint8) *Team { return s.Teams.ByID(TeamID(id)) },
		loadNetCode:    func(obj *Object) uint32 { return obj.NetCode },
		changeTeam:     func(value *ObjectTeam, team *Team, code uint32, flags int32) { r.ChangeTeam(value, team, code, flags) },
		createTeam: func(id uint8, value *ObjectTeam, active int32, code uint32, flags int32) {
			r.CreateTeam(TeamID(id), value, active, code, flags)
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) { s.Audio.EventObj(sound.ID(id), obj, int(kind), code) },
	})
}
