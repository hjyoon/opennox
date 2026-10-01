package server

// QuestLoseWeapon54CC40 binds the original weapon-loss helper to native
// inventory, modifier and Player pointers. It does not require a Player class
// on unit itself or normalize missing bindings that the original would fault.
func (s *Server) QuestLoseWeapon54CC40(unit *Object, classCanUse func(*Object, uint8) int32, delayedDelete func(*Object)) {
	questLoseWeapon54CC40(unit, questLoseWeaponHooks54CC40[*Object, *PlayerUpdateData, *ModifierInitData, *ModifierEff]{
		updateData: func(unit *Object) *PlayerUpdateData { return (*PlayerUpdateData)(unit.UpdateData) },
		first:      func(unit *Object) *Object { return unit.InvFirstItem },
		next:       func(item *Object) *Object { return item.InvNextItem },
		flags:      func(item *Object) uint32 { return uint32(item.ObjFlags) },
		class:      func(item *Object) uint32 { return uint32(item.ObjClass) },
		subclass:   func(item *Object) uint32 { return uint32(item.ObjSubClass) },
		modifierData: func(item *Object) *ModifierInitData {
			return (*ModifierInitData)(item.InitData)
		},
		modifier:      func(data *ModifierInitData, slot int) *ModifierEff { return data.Modifiers[slot] },
		playerClass:   func(update *PlayerUpdateData) uint8 { return update.Player.info[66] },
		classCanUse:   classCanUse,
		delayedDelete: delayedDelete,
	})
}
