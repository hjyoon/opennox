package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// ObjectDropBoundedRuntime4ED810 supplies the shared legacy Crown cache and
// the restored 004ED790 dispatcher. All objects and points remain native-width.
type ObjectDropBoundedRuntime4ED810 struct {
	LoadCrownTypeCache  func() uint32
	StoreCrownTypeCache func(uint32)
	Dispatch            func(*Object, *Object, *types.Pointf) int32
}

type objectDropBoundedNativeDeps4ED810 struct {
	mapTrace            func(*types.Pointf, *types.Pointf) int32
	priorityMessage     func(*Object, string, int32)
	audio               func(uint32, *Object, int32, uint32)
	gameFlag            func(uint32) int32
	loadCrownTypeCache  func() uint32
	lookupCrownType     func() uint32
	storeCrownTypeCache func(uint32)
	dispatch            func(*Object, *Object, *types.Pointf) int32
}

func objectDropBoundedNative4ED810(
	owner, item *Object,
	point *types.Pointf,
	deps objectDropBoundedNativeDeps4ED810,
) int32 {
	return objectDropBounded4ED810(objectDropBoundedHooks4ED810[
		*Object,
		*types.Pointf,
	]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadOwnerX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadOwnerY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadPointArg: func() *types.Pointf {
			return point
		},
		loadPointX: func(point *types.Pointf) float32 {
			return point.X
		},
		loadPointY: func(point *types.Pointf) float32 {
			return point.Y
		},
		mapTrace:        deps.mapTrace,
		priorityMessage: deps.priorityMessage,
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
		audio:               deps.audio,
		gameFlag:            deps.gameFlag,
		loadItemArg:         func() *Object { return item },
		loadCrownTypeCache:  deps.loadCrownTypeCache,
		lookupCrownType:     deps.lookupCrownType,
		storeCrownTypeCache: deps.storeCrownTypeCache,
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		dispatch: deps.dispatch,
	})
}

func objectDropBoundedServerDeps4ED810(
	s *Server,
	runtime ObjectDropBoundedRuntime4ED810,
) objectDropBoundedNativeDeps4ED810 {
	return objectDropBoundedNativeDeps4ED810{
		mapTrace: func(origin, target *types.Pointf) int32 {
			if s.MapTraceRayAt(*origin, *target, nil, nil, 0) {
				return 1
			}
			return 0
		},
		priorityMessage: func(owner *Object, message string, kind int32) {
			s.NetPriMsgToPlayer(owner, strman.ID(message), byte(kind))
		},
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadCrownTypeCache: runtime.LoadCrownTypeCache,
		lookupCrownType: func() uint32 {
			return uint32(s.Types.IndByID("Crown"))
		},
		storeCrownTypeCache: runtime.StoreCrownTypeCache,
		dispatch:            runtime.Dispatch,
	}
}

// ObjectDropBounded4ED810 clamps a requested drop point to 75 units, checks
// line-of-sight, rejects KOTR Crown drops, and otherwise invokes 004ED790.
func (s *Server) ObjectDropBounded4ED810(
	owner, item *Object,
	point *types.Pointf,
	runtime ObjectDropBoundedRuntime4ED810,
) int32 {
	return objectDropBoundedNative4ED810(owner, item, point, objectDropBoundedServerDeps4ED810(s, runtime))
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(types.Pointf{})]
)
