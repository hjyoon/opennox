package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/ntype"
)

type unitClearOwnerNativeDeps4EC300 struct {
	isMonitored    func(*Object, *Object) bool
	netFxShield    func(uint8, *Object)
	unmarkMinimap  func(uint8, *Object, uint32)
	resetMonster   func(*Object)
	markUnitUpdate func(*Object)
}

func unitClearOwnerNative4EC300(obj *Object, deps unitClearOwnerNativeDeps4EC300) {
	unitClearOwner4EC300(obj, unitClearOwnerHooks4EC300[*Object, *PlayerUpdateData, *Player]{
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		isMonitored: deps.isMonitored,
		loadSubClass: func(obj *Object) uint32 {
			return uint32(obj.ObjSubClass)
		},
		loadPlayerData: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		storeSubClass: func(obj *Object, subClass uint32) {
			obj.ObjSubClass = object.SubClass(subClass)
		},
		loadPlayer: func(data *PlayerUpdateData) *Player {
			return data.Player
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		netFxShield:   deps.netFxShield,
		unmarkMinimap: deps.unmarkMinimap,
		loadFirstOwned: func(owner *Object) *Object {
			return owner.Field129
		},
		loadNextOwned: func(obj *Object) *Object {
			return obj.Field128
		},
		storeNextOwned: func(obj, next *Object) {
			obj.Field128 = next
		},
		storeFirstOwned: func(owner, first *Object) {
			owner.Field129 = first
		},
		storeOwner: func(obj, owner *Object) {
			obj.ObjOwner = owner
		},
		resetMonster:   deps.resetMonster,
		markUnitUpdate: deps.markUnitUpdate,
	})
}

func (s *Server) unitClearOwner4EC300(obj *Object) {
	unitClearOwnerNative4EC300(obj, unitClearOwnerNativeDeps4EC300{
		isMonitored: Nox_xxx_creatureIsMonitored_500CC0,
		netFxShield: func(ind uint8, obj *Object) {
			s.Nox_xxx_netFxShield_0_4D9200(int(ind), obj)
		},
		unmarkMinimap: func(ind uint8, obj *Object, flags uint32) {
			s.Players.Nox_xxx_netUnmarkMinimapObj_417300(ntype.PlayerInd(ind), obj, flags)
		},
		resetMonster: func(obj *Object) {
			obj.Nox_xxx_monsterResetEnemy_5346F0()
		},
		markUnitUpdate: func(obj *Object) {
			obj.Nox_xxx_monsterMarkUpdate_4E8020()
		},
	})
}
