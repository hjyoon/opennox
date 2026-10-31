package server

import "unsafe"

// MonsterAnimationUpdateNative50A850 retains the entry update through the
// animation lookup. Attack and completed paths do not request a provider.
// The original char result is a literal byte or the returned record's low byte.
func MonsterAnimationUpdateNative50A850(unit *Object, runtime MonsterAnimationUpdateNativeRuntime50A850) uint8 {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if uint32(unit.ObjSubClass)&0x10 != 0 {
		item := (*AIStackItem)(unsafe.Add(unsafe.Pointer(&update.AIStack[0]), int(update.AIStackInd)*int(unsafe.Sizeof(update.AIStack[0]))))
		if action := item.Action; action == 16 || action == 17 {
			update.Field120_3 = 0
			return uint8(action)
		}
	}
	if done := update.Field120_3; done != 0 {
		return done
	}
	animation := runtime.Animation(unit)
	if animation == nil {
		return 0
	}
	monsterUpdateAnim50A850(update, animation)
	return uint8(uintptr(unsafe.Pointer(animation)))
}

// MonsterAnimationUpdate50A850 uses the ordinary registry and player frames.
// Callers needing class/nil admission retain MonsterUpdateNPCAnim50A850.
func (s *Server) MonsterAnimationUpdate50A850(unit *Object) uint8 {
	return MonsterAnimationUpdateNative50A850(unit, MonsterAnimationUpdateNativeRuntime50A850{
		Animation: s.MonsterNPCAnimationPointer533D00,
	})
}
