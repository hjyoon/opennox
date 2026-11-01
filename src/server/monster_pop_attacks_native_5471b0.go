package server

// MonsterPopAttackActionsNative5471B0 retains the original void entry and its
// entry-cached MUD across Pop calls. The defined index domain is -1 and 0..23.
func (unit *Object) MonsterPopAttackActionsNative5471B0(pop func(*Object) int) {
	update := (*MonsterUpdateData)(unit.UpdateData)
	for {
		index := int(update.AIStackInd)
		var action uint32
		if index == -1 {
			// Original MUD+528 precedes slot zero by the original 24-byte
			// stack stride; the widened native layout needs a field projection.
			action = update.Field132
		} else {
			action = update.AIStack[index].Action
		}
		switch action {
		case 16, 17, 18, 19, 20, 25, 26, 27, 28:
			pop(unit)
		default:
			return
		}
	}
}
