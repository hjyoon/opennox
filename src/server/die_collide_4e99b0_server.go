package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

type DieCollideRuntime4E99B0 struct {
	DelayedDelete func(*Object)
}

type dieCollideNativeDeps4E99B0 struct {
	unitsOnSameTeam func(*Object, *Object) int32
	callDeath       func(unsafe.Pointer, *Object)
	delayedDelete   func(*Object)
}

func dieCollideNative4E99B0(
	source, target *Object,
	collision unsafe.Pointer,
	deps dieCollideNativeDeps4E99B0,
) {
	dieCollide4E99B0(source, target, collision, dieCollideHooks4E99B0[
		*Object,
		unsafe.Pointer,
	]{
		unitsOnSameTeam: deps.unitsOnSameTeam,
		classLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadDeath: func(obj *Object) unsafe.Pointer {
			return obj.Death
		},
		storeFlags: func(obj *Object, flags uint32) {
			obj.ObjFlags = object.Flags(flags)
		},
		callDeath:     deps.callDeath,
		delayedDelete: deps.delayedDelete,
	})
}

// DieCollide4E99B0 binds the registered callback to native-width Object and
// function pointers. Its zero-byte collide record and collision argument stay
// unread, matching GAME.EXE.
func (s *Server) DieCollide4E99B0(
	source, target *Object,
	collision unsafe.Pointer,
	runtime DieCollideRuntime4E99B0,
) {
	_ = s
	dieCollideNative4E99B0(source, target, collision, dieCollideNativeDeps4E99B0{
		unitsOnSameTeam: func(first, second *Object) int32 {
			if UnitsHaveSameTeam4EC520(first, second) {
				return 1
			}
			return 0
		},
		callDeath: func(death unsafe.Pointer, obj *Object) {
			CallObjectDeath(death, obj)
		},
		delayedDelete: runtime.DelayedDelete,
	})
}
