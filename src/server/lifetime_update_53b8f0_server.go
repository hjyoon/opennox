package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// LifetimeUpdateData53B8F0 is the four-byte duration record registered for
// LifetimeUpdate in thing.bin.
type LifetimeUpdateData53B8F0 struct {
	Duration uint32
}

var _ = [1]struct{}{}[4-unsafe.Sizeof(LifetimeUpdateData53B8F0{})]

// LifetimeUpdateRuntime53B8F0 supplies the deletion service used when an
// expired object has no death override.
type LifetimeUpdateRuntime53B8F0 struct {
	DelayedDelete func(*Object)
}

type lifetimeUpdateNativeDeps53B8F0 struct {
	frame         func() uint32
	callDeath     func(unsafe.Pointer, *Object)
	delayedDelete func(*Object)
}

func lifetimeUpdateNative53B8F0(source *Object, deps lifetimeUpdateNativeDeps53B8F0) {
	lifetimeUpdate53B8F0(source, lifetimeUpdateHooks53B8F0[
		*Object,
		*LifetimeUpdateData53B8F0,
		unsafe.Pointer,
	]{
		frame: deps.frame,
		loadCreationFrame: func(obj *Object) uint32 {
			return obj.Field32
		},
		loadUpdateData: func(obj *Object) *LifetimeUpdateData53B8F0 {
			return (*LifetimeUpdateData53B8F0)(obj.UpdateData)
		},
		loadDuration: func(data *LifetimeUpdateData53B8F0) uint32 {
			return data.Duration
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

// LifetimeUpdate53B8F0 binds GAME.EXE 0053B8F0 to native-width object,
// update-data, and callback pointers.
func (s *Server) LifetimeUpdate53B8F0(source *Object, runtime LifetimeUpdateRuntime53B8F0) {
	lifetimeUpdateNative53B8F0(source, lifetimeUpdateNativeDeps53B8F0{
		frame: s.Frame,
		callDeath: func(death unsafe.Pointer, obj *Object) {
			CallObjectDeath(death, obj)
		},
		delayedDelete: runtime.DelayedDelete,
	})
}
