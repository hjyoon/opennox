package opennox

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type delayedDeleteObject4E5CC0Hooks struct {
	isCreatureMonitored   func(owner, obj *server.Object) bool
	removeMonsterMonitors func(owner, obj *server.Object)
	removeFromInventory   func(holder, obj *server.Object)
	cancelPlayerSpells    func(obj *server.Object)
	questMode             func() bool
	questDeleteMonster    func(obj *server.Object)
	deletePlayer          func(obj *server.Object)
	deletedList           func() *server.Object
	setDeletedList        func(obj *server.Object)
	frame                 func() uint32
	changeTeam            func(team *server.ObjectTeam, netCode uint32)
}

func delayedDeleteObject_4E5CC0(obj *server.Object, hooks delayedDeleteObject4E5CC0Hooks) {
	if obj == nil || obj.Flags().Has(object.FlagDestroyed) {
		return
	}
	if owner := obj.Owner(); owner != nil && owner.Class().Has(object.ClassPlayer) {
		if obj.Class().Has(object.ClassMonster) && !hooks.isCreatureMonitored(owner, obj) &&
			obj.SubClass().AsMonster().Has(object.MonsterMigrate) {
			hooks.removeMonsterMonitors(obj.Owner(), obj)
		}
	}

	if holder := obj.InvHolder; holder != nil {
		hooks.removeFromInventory(holder, obj)
	}
	hooks.cancelPlayerSpells(obj)
	if hooks.questMode() && obj.Class().Has(object.ClassMonster) {
		hooks.questDeleteMonster(obj)
	}
	if obj.Class().Has(object.ClassPlayer) {
		hooks.deletePlayer(obj)
	}

	obj.ObjFlags |= object.FlagDestroyed
	obj.DeletedNext = hooks.deletedList()
	deletedAt := hooks.frame()
	hooks.setDeletedList(obj)
	obj.DeletedAt = deletedAt
	if obj.HasTeam() {
		hooks.changeTeam(obj.TeamPtr(), obj.NetCode)
	}
}
