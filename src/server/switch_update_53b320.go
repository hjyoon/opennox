package server

import (
	"github.com/opennox/libs/object"
)

type SwitchUpdateRuntime53B320 struct {
	QueueCollision func(*Object)
}

// SwitchUpdate53B320 restores GAME.EXE 0053B320 without invoking the
// PE32-only object updater on native-width Object records.
func (s *Server) SwitchUpdate53B320(unit *Object, runtime SwitchUpdateRuntime53B320) uint8 {
	if unit == nil {
		return 0
	}
	if unit.ObjFlags.Has(object.FlagEnabled) {
		queue := unit.Collide != nil && unit.ObjFlags.Has(object.FlagNoCollide)
		unit.ObjFlags &^= object.FlagNoCollide
		if queue && runtime.QueueCollision != nil {
			runtime.QueueCollision(unit)
		}
	} else {
		if unit.Field33 == 0 {
			unit.NeedSync()
			unit.Field33 = 1
		}
		unit.ObjFlags |= object.FlagNoCollide
	}
	return uint8(unit.ObjFlags)
}
