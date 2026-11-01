package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// MorphPlayerBotFromPlayer4FAAC0 preserves the raw public morph entry.
// Read the replacement link before changing any unit state: a nil live
// player record faults even when its replacement link would be nil.
func (unit *Object) MorphPlayerBotFromPlayer4FAAC0() byte {
	class := unit.ObjClass
	if class.Has(object.ClassPlayer) {
		monster := (*PlayerUpdateData)(unit.UpdateData).Field73
		class = class&^object.ClassPlayer | object.ClassMonster
		unit.ObjClass = class
		unit.UpdateData = unsafe.Pointer(monster)
		unit.ObjSubClass = 16
	}
	return byte(class)
}

// MorphPlayerBotToPlayer4FAAF0 restores the live monster record's player
// link, clears the complete subclass word and returns only the original AL.
// Like the original public entry, a nil unit or selected record faults.
func (unit *Object) MorphPlayerBotToPlayer4FAAF0() byte {
	class := unit.ObjClass
	if class.Has(object.ClassMonster) {
		player := (*MonsterUpdateData)(unit.UpdateData).Field545
		class = class&^object.ClassMonster | object.ClassPlayer
		unit.ObjClass = class
		unit.UpdateData = unsafe.Pointer(player)
		unit.ObjSubClass = 0
	}
	return byte(class)
}
