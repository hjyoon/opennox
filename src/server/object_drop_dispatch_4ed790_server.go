package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// ObjectDropRuntime4ED790 contains the two services whose implementations are
// still legacy-facing. Object fields, the Drop callback and Pointf stay at
// native pointer width throughout the dispatcher.
type ObjectDropRuntime4ED790 struct {
	DefaultDrop func(*Object, *Object, *types.Pointf) int32
	RefreshUnit func(*Object)
}

type objectDropDispatchNativeDeps4ED790 struct {
	gameFlag    func(uint32) int32
	defaultDrop func(*Object, *Object, *types.Pointf) int32
	refreshUnit func(*Object)
}

func objectDropDispatchNative4ED790(
	owner, item *Object,
	point *types.Pointf,
	deps objectDropDispatchNativeDeps4ED790,
) int32 {
	return objectDropDispatch4ED790(objectDropDispatchHooks4ED790[
		*Object,
		*types.Pointf,
		DropFunc,
	]{
		loadItemArg: func() *Object {
			return item
		},
		gameFlag: deps.gameFlag,
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		storeFlags: func(obj *Object, flags uint32) {
			obj.ObjFlags = object.Flags(flags)
		},
		refreshUnit: deps.refreshUnit,
		loadDrop: func(obj *Object) DropFunc {
			return obj.Drop.Get()
		},
		hasDrop: func(drop DropFunc) bool {
			return drop != nil
		},
		loadPointArg: func() *types.Pointf {
			return point
		},
		loadOwnerArg: func() *Object {
			return owner
		},
		callDrop: func(drop DropFunc, owner, item *Object, point *types.Pointf) int32 {
			return drop(owner, item, point)
		},
		defaultDrop: deps.defaultDrop,
	})
}

func objectDropDispatchServerDeps4ED790(runtime ObjectDropRuntime4ED790) objectDropDispatchNativeDeps4ED790 {
	return objectDropDispatchNativeDeps4ED790{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		defaultDrop: runtime.DefaultDrop,
		refreshUnit: runtime.RefreshUnit,
	}
}

// ObjectDrop4ED790 dispatches a drop through the dropped item's live handler
// and preserves the original DefaultDrop fallback and full 32-bit result.
func (s *Server) ObjectDrop4ED790(
	owner, item *Object,
	point *types.Pointf,
	runtime ObjectDropRuntime4ED790,
) int32 {
	return objectDropDispatchNative4ED790(owner, item, point, objectDropDispatchServerDeps4ED790(runtime))
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(types.Pointf{})]
)
