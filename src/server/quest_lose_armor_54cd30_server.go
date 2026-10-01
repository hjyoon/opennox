package server

// QuestLoseArmor54CD30 uses native item links, the exact type-WORD armor
// lookup, the original logic RNG and the caller's normal delayed deletion.
// No Player binding is read by this helper, and nil unit remains a fault.
func (s *Server) QuestLoseArmor54CD30(unit *Object, delayedDelete func(*Object)) {
	questLoseArmor54CD30(unit, questLoseArmorHooks54CD30[*Object]{
		first: func(unit *Object) *Object { return unit.InvFirstItem },
		next: func(item *Object) *Object {
			if item == nil {
				return nil
			}
			return item.InvNextItem
		},
		flags:     func(item *Object) uint32 { return uint32(item.ObjFlags) },
		class:     func(item *Object) uint32 { return uint32(item.ObjClass) },
		typeIndex: func(item *Object) uint16 { return item.TypeInd },
		armorFlags: func(typ uint16) uint32 {
			return s.Armor.Sub_415D10(int(typ))
		},
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		delayedDelete: delayedDelete,
	})
}
