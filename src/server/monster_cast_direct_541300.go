package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// MonsterMorphToPlayer4FAAF0 switches the original bot's two update records
// without narrowing their links to a PE32 DWORD. All other class bits survive.
func MonsterMorphToPlayer4FAAF0(unit *Object) int8 {
	if unit == nil {
		return 0
	}
	if unit.ObjClass.Has(object.ClassMonster) && unit.UpdateData != nil {
		player := (*MonsterUpdateData)(unit.UpdateData).Field545
		unit.ObjClass = unit.ObjClass&^object.ClassMonster | object.ClassPlayer
		unit.UpdateData = unsafe.Pointer(player)
		unit.ObjSubClass = 0
	}
	return int8(byte(unit.ObjClass))
}

// MonsterMorphFromPlayer4FAAC0 deliberately reads the live player record,
// rather than restoring a cached pointer or the old subclass.
func MonsterMorphFromPlayer4FAAC0(unit *Object) int8 {
	if unit == nil {
		return 0
	}
	if unit.ObjClass.Has(object.ClassPlayer) && unit.UpdateData != nil {
		monster := (*PlayerUpdateData)(unit.UpdateData).Field73
		unit.ObjClass = unit.ObjClass&^object.ClassPlayer | object.ClassMonster
		unit.UpdateData = unsafe.Pointer(monster)
		unit.ObjSubClass = object.SubClass(16)
	}
	return int8(byte(unit.ObjClass))
}

// MonsterCastDirect541300 is the immediate cast service, not MonsterCast's
// animation/action scheduler. GAME.EXE 00541300 caches the monster record,
// optionally morphs a bot, calculates Direction2 at 00533CC0, and calls
// 004FDD20. Its second BOT test rereads that cached record after the cast.
// Cast acceptance is not a success gate: rejected spells still complete this
// service, and MainAI must still charge its cooldown.
func MonsterCastDirect541300(id int32, unit *Object, arg *SpellAcceptArg, cast func(int32, *Object, *SpellAcceptArg) int32) bool {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) || arg == nil || cast == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.StatusFlags.Has(object.MonStatusBot) {
		if update.Field545 == nil {
			return false // contain invalid native metadata before changing class
		}
		MonsterMorphToPlayer4FAAF0(unit)
	}
	unit.Direction2 = DirFromVec(arg.Pos.Sub(unit.PosVec))
	cast(id, unit, arg)
	if update.StatusFlags.Has(object.MonStatusBot) {
		MonsterMorphFromPlayer4FAAC0(unit)
	}
	return true
}
