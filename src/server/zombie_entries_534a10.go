package server

import (
	"github.com/opennox/libs/object"
	"unsafe"
)

// TypePredicateRuntime534A10 views the three original DWORD caches through
// their native owner. Lookup calls can change both caches and the object.
type TypePredicateRuntime534A10 struct {
	Read   func(int) uint32
	Write  func(int, uint32)
	Lookup func(string) uint32
}

func UnitIsPlant534A10(unit *Object, r TypePredicateRuntime534A10) uint32 {
	value := r.Read(0)
	if value == 0 {
		value = r.Lookup("CarnivorousPlant")
		r.Write(0, value)
	}
	if uint32(unit.TypeInd) == value {
		return 1
	}
	return 0
}

func UnitIsZombie534A40(unit *Object, r TypePredicateRuntime534A10) uint32 {
	if r.Read(1) == 0 {
		r.Write(1, r.Lookup("Zombie"))
		r.Write(2, r.Lookup("VileZombie"))
	}
	typ := uint32(unit.TypeInd)
	if typ == r.Read(1) || typ == r.Read(2) {
		return 1
	}
	return 0
}

func (s *Server) TypePredicateRuntime534A10() TypePredicateRuntime534A10 {
	return TypePredicateRuntime534A10{
		Read: func(index int) uint32 {
			switch index {
			case 0:
				return s.Types.fast.unitIsPlant534A10
			case 1:
				return s.Types.fast.unitIsZombie534A40
			default:
				return s.Types.fast.unitIsVileZombie534A40
			}
		},
		Write: func(index int, value uint32) {
			switch index {
			case 0:
				s.Types.fast.unitIsPlant534A10 = value
			case 1:
				s.Types.fast.unitIsZombie534A40 = value
			default:
				s.Types.fast.unitIsVileZombie534A40 = value
			}
		},
		Lookup: func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
	}
}

// MonsterActionGetUpRaw534A90 retains the original AL result. The action
// dispatcher can ignore it, while the C ABI still observes its pointer/pop byte.
func MonsterActionGetUpRaw534A90(unit *Object, pop func(*Object) uint32) byte {
	result := byte(uintptr(unsafe.Pointer(unit)))
	if (*MonsterUpdateData)(unit.UpdateData).Field120_3 != 0 {
		result = byte(pop(unit))
	}
	return result
}

type MonsterRaiseZombieRawRuntime534AB0 struct {
	IsZombie      func(*Object) uint32
	CurrentAction func(*Object) uint32
	Pop           func(*Object) uint32
	Push          func(*Object, uint32) uintptr
	Audio         func(uint32, *Object)
	Health        func(*Object)
}

func MonsterRaiseZombieRaw534AB0(unit *Object, r MonsterRaiseZombieRawRuntime534AB0) uint32 {
	result := r.IsZombie(unit)
	if result == 0 {
		return result
	}
	result = r.CurrentAction(unit)
	if result != 31 {
		return result
	}
	r.Pop(unit)
	r.Push(unit, 61)
	r.Push(unit, 35)
	r.Audio(469, unit)
	r.Health(unit)
	result = uint32(unit.ObjFlags) & 0xffff7fa7
	unit.ObjFlags = object.Flags(result)
	return result
}
