package server

import "unsafe"

// MonsterAnimationNative533790 computes the index before type predicates,
// retaining the entry update and the original predicate short circuits.
func MonsterAnimationNative533790(unit *Object, runtime MonsterAnimationNativeRuntime533790) int32 {
	update := (*MonsterUpdateData)(unit.UpdateData)
	index := int32(8)
	if head := update.AIStackInd; head != -1 {
		item := (*AIStackItem)(unsafe.Add(unsafe.Pointer(&update.AIStack[0]), int(head)*int(unsafe.Sizeof(update.AIStack[0]))))
		switch item.Action {
		case 7, 8, 10, 13, 29, 36, 37:
			index = int32((uint32(update.StatusFlags)&0x4000 | 0x30000) >> 14)
		case 9:
			index = 12
		case 16:
			index = 1
		case 17:
			index = 3
		case 18, 19, 20:
			index = 7
		case 21, 23:
			index = 5
		case 22:
			index = 6
		case 24:
			index = 13
		case 30:
			index = 9
		case 31:
			index = 10
		case 33, 35:
			index = 14
		case 34:
			index = 15
		}
	}
	if runtime.Mimic(unit) != 0 && index == 8 {
		if uint32(update.StatusFlags)&0x40000 != 0 {
			return 0
		}
		return index
	}
	if runtime.Plant(unit) != 0 && index == 8 {
		if update.CurrentEnemy == nil {
			return 14
		}
		return index
	}
	if runtime.Zombie(unit) != 0 && index == 9 && uint32(update.StatusFlags)&0x80000 != 0 {
		return 15
	}
	return index
}

func (s *Server) MonsterAnimationIndex533790(unit *Object) int32 {
	word := func(value bool) uint32 {
		if value {
			return 1
		}
		return 0
	}
	return MonsterAnimationNative533790(unit, MonsterAnimationNativeRuntime533790{
		Mimic:  func(unit *Object) uint32 { return word(s.IsMimic(unit)) },
		Plant:  func(unit *Object) uint32 { return word(s.IsPlant(unit)) },
		Zombie: func(unit *Object) uint32 { return word(s.IsZombie(unit)) },
	})
}
