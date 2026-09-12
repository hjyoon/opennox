package server

import "github.com/opennox/libs/object"

// SetRoamFlag515C80 is the native-width form of GAME.EXE 00515C80. The
// original writes one byte at MonsterUpdateData+1332, not the whole word.
func (obj *Object) SetRoamFlag515C80(flag uint8) {
	if obj == nil || !obj.Class().Has(object.ClassMonster) {
		return
	}
	update := obj.UpdateDataMonster()
	if update == nil {
		return
	}
	update.Field333 = update.Field333&^uint32(0xff) | uint32(flag)
}
