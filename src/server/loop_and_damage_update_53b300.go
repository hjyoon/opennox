package server

import "github.com/opennox/libs/object"

// LoopAndDamageUpdate53B300 follows GAME.EXE 0053B300. The callback's byte
// return value is ignored by object-update dispatch; only ObjFlags changes.
//
//go:noinline
func LoopAndDamageUpdate53B300(obj *Object) {
	const (
		loopDisabled = object.Flags(0x1000000)
		loopActive   = object.Flags(0x40)
	)
	if obj.ObjFlags&loopDisabled != 0 {
		obj.ObjFlags &^= loopActive
	} else {
		obj.ObjFlags |= loopActive
	}
}
