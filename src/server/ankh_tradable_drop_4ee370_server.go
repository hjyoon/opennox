package server

import "github.com/opennox/libs/types"

// AnkhTradableDropRuntime4EE370 supplies the already-restored DefaultDrop
// dependency. The thunk itself has no additional Server service dependency.
type AnkhTradableDropRuntime4EE370 struct {
	DefaultDrop func(*Object, *Object, *types.Pointf) int32
}

func ankhTradableDropNative4EE370(
	owner, item *Object,
	point *types.Pointf,
	runtime AnkhTradableDropRuntime4EE370,
) int32 {
	return ankhTradableDrop4EE370(ankhTradableDropHooks4EE370[*Object, *types.Pointf]{
		loadPointArg: func() *types.Pointf {
			return point
		},
		loadItemArg: func() *Object {
			return item
		},
		loadOwnerArg: func() *Object {
			return owner
		},
		defaultDrop: runtime.DefaultDrop,
	})
}

// AnkhTradableDrop4EE370 binds GAME.EXE 004EE370 to native-width Object and
// Pointf pointers while preserving the forwarding thunk's exact result.
func (s *Server) AnkhTradableDrop4EE370(
	owner, item *Object,
	point *types.Pointf,
	runtime AnkhTradableDropRuntime4EE370,
) int32 {
	return ankhTradableDropNative4EE370(owner, item, point, runtime)
}
