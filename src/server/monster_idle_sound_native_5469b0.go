package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
)

// MonsterIdleSoundNative5469B0 retains the original raw void entry contract.
// The valid stack domain is all 24 slots and the original -1 timer-prefix alias.
func (unit *Object) MonsterIdleSoundNative5469B0(frame, tickRate func() uint32,
	random func(int, int) int, audio func(uint32, *Object, uint32, uint32)) {
	class := uint8(unit.ObjClass)
	update := (*MonsterUpdateData)(unit.UpdateData)
	if class&uint8(object.ClassMonster) == 0 {
		return
	}
	if unit.Nox_xxx_monsterCanAttackAtWill_534390() || update.CurrentEnemy != nil || math.Float32frombits(update.Field131) > 300 {
		return
	}
	// The original signed index -1 addresses MUD+528. Project that field by
	// meaning rather than subtracting a widened native AIStackItem stride.
	index := int(update.AIStackInd)
	var action uint32
	if index == -1 {
		action = update.Field132
	} else {
		action = update.AIStack[index].Action
	}
	if action != 0 && action != 4 {
		return
	}
	if frame() < update.Field132 {
		return
	}
	fps := tickRate()
	delay := random(int(int32(fps*20)), int(int32(fps*60)))
	update.Field132 = frame() + uint32(delay)
	// Original 424300 reloads class and the object's current MUD after RNG.
	if uint8(unit.ObjClass)&uint8(object.ClassMonster) != 0 {
		if set := (*MonsterUpdateData)(unit.UpdateData).SoundSet122; set != nil {
			audio(*(*uint32)(unsafe.Add(set, 16)), unit, 0, 0)
		}
	}
}
