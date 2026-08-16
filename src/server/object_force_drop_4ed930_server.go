package server

import "github.com/opennox/libs/types"

// ObjectForceDropRuntime4ED930 supplies the restored 004ED790 dispatcher.
// The random reachable-point dependency is owned by Server and all objects
// and points remain native-width.
type ObjectForceDropRuntime4ED930 struct {
	Dispatch func(*Object, *Object, *types.Pointf) int32
}

type objectForceDropNativeDeps4ED930 struct {
	randomReachable func(float32, *Object, *types.Pointf) *types.Pointf
	dispatch        func(*Object, *Object, *types.Pointf) int32
}

func objectForceDropNative4ED930(
	owner, item *Object,
	deps objectForceDropNativeDeps4ED930,
) int32 {
	return objectForceDrop4ED930(objectForceDropHooks4ED930[
		*Object,
		types.Pointf,
	]{
		loadOwnerArg: func() *Object {
			return owner
		},
		randomReachable: deps.randomReachable,
		loadItemArg: func() *Object {
			return item
		},
		dispatch: deps.dispatch,
	})
}

func objectForceDropServerDeps4ED930(
	s *Server,
	runtime ObjectForceDropRuntime4ED930,
) objectForceDropNativeDeps4ED930 {
	return objectForceDropNativeDeps4ED930{
		randomReachable: func(radius float32, owner *Object, output *types.Pointf) *types.Pointf {
			return s.RandomReachablePointAroundInto4ED970(radius, &owner.PosVec, output)
		},
		dispatch: runtime.Dispatch,
	}
}

// ObjectForceDrop4ED930 selects a reachable point at radius 50 and invokes
// the restored object-drop dispatcher with its local result.
func (s *Server) ObjectForceDrop4ED930(
	owner, item *Object,
	runtime ObjectForceDropRuntime4ED930,
) int32 {
	return objectForceDropNative4ED930(owner, item, objectForceDropServerDeps4ED930(s, runtime))
}
