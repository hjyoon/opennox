package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// MonsterNPCAnimationNative533D00 retains the entry update across the
// selector call, then reloads that update's animation table for the return.
func MonsterNPCAnimationNative533D00(unit *Object, runtime MonsterNPCAnimationNativeRuntime533D00) *MonsterAnim {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if uint32(unit.ObjSubClass)&0x10 != 0 {
		item := (*AIStackItem)(unsafe.Add(unsafe.Pointer(&update.AIStack[0]), int(update.AIStackInd)*int(unsafe.Sizeof(update.AIStack[0]))))
		var frames, duration int32
		switch item.Action {
		case 16, 17:
			var index int32
			if equip := update.WeaponEquipFlags &^ 3; equip != 0 {
				index = runtime.WeaponAnimation(equip)
			} else {
				index = int32(uint8(update.Field517))
			}
			frames, duration = runtime.Frames(index)
		case 18, 19, 20:
			frames, duration = runtime.Frames(21)
		case 30:
			frames, duration = runtime.Frames(1)
		case 31:
			frames, duration = runtime.Frames(2)
		case 21:
			frames, duration = runtime.Frames(40)
		case 23:
			index := int32(47)
			if update.WeaponEquipFlags&0x7ff8000 != 0 {
				index = 30
			}
			frames, duration = runtime.Frames(index)
		}
		result := runtime.Scratch()
		*result = MonsterAnim{frames: byte(frames), field10: byte(duration)}
		return result
	}
	if update.Field119 == nil {
		return nil
	}
	index := runtime.Selector(unit)
	if table := update.Field119; table != nil {
		return (*MonsterAnim)(unsafe.Add(unsafe.Pointer(table), int(index)*int(unsafe.Sizeof(MonsterAnim{}))))
	}
	return monsterNPCAnimationNilTableOffset533D00(index)
}

// If lookup cleared the cached table, the original machine code still
// returns the integer offset as an address without dereferencing it. Keep
// this raw nil-base word conversion separate from checked native table
// arithmetic; an initial nil table continues to return before the selector.
//
//go:noinline
//go:nocheckptr
func monsterNPCAnimationNilTableOffset533D00(index int32) *MonsterAnim {
	return (*MonsterAnim)(unsafe.Pointer(uintptr(uint32(index) << 4)))
}

// MonsterNPCAnimationPointer533D00 uses the ordinary Go type registry and
// player frame table, sharing the original fixed scratch record with C.
func (s *Server) MonsterNPCAnimationPointer533D00(unit *Object) *MonsterAnim {
	return MonsterNPCAnimationNative533D00(unit, MonsterNPCAnimationNativeRuntime533D00{
		WeaponAnimation: playerWeaponAnimationNative4FA280,
		Frames: func(index int32) (int32, int32) {
			frames, duration := s.PlayerAnimFrames(int(index))
			return int32(frames), int32(duration)
		},
		Selector: s.MonsterAnimationIndex533790,
		Scratch: func() *MonsterAnim {
			return (*MonsterAnim)(memmap.PtrOff(0x5D4594, 2487964))
		},
	})
}
