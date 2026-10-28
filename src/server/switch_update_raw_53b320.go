package server

import "github.com/opennox/libs/object"

type SwitchUpdateRawRuntime53B320 struct {
	QueueCollision func(*Object) uint8
	NeedSync       func(*Object)
}

// SwitchUpdateRaw53B320 preserves the required object and the actual queue
// helper's return byte, which need not equal the object's resulting flags.
func SwitchUpdateRaw53B320(unit *Object, h SwitchUpdateRawRuntime53B320) uint8 {
	flags := unit.ObjFlags
	if flags.Has(object.FlagEnabled) {
		queue := unit.Collide != nil && flags.Has(object.FlagNoCollide)
		flags &^= object.FlagNoCollide
		unit.ObjFlags = flags
		if queue {
			return h.QueueCollision(unit)
		}
		return uint8(flags)
	}
	if unit.Field33 == 0 {
		h.NeedSync(unit)
		unit.Field33 = 1
	}
	flags = unit.ObjFlags | object.FlagNoCollide
	unit.ObjFlags = flags
	return uint8(flags)
}
