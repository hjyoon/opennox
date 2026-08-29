package server

// OneSecondDieUpdateRuntime53CB60 supplies the deletion service reached once
// an object's unsigned age is at least one server second.
type OneSecondDieUpdateRuntime53CB60 struct {
	DelayedDelete func(*Object)
}

type oneSecondDieUpdateNativeDeps53CB60 struct {
	frame         func() uint32
	fps           func() uint32
	delayedDelete func(*Object)
}

func oneSecondDieUpdateNative53CB60(source *Object, deps oneSecondDieUpdateNativeDeps53CB60) {
	// GAME.EXE loads these values in this order. In particular, the source
	// creation frame is read between the two shared-clock loads.
	frame := deps.frame()
	creationFrame := source.Field32
	fps := deps.fps()
	if frame-creationFrame >= fps {
		deps.delayedDelete(source)
	}
}

// OneSecondDieUpdate53CB60 restores GAME.EXE 0053CB60 without narrowing the
// source object to an ABI32 integer. The subtraction and comparison remain
// uint32 operations, including wraparound at the frame counter boundary.
func (s *Server) OneSecondDieUpdate53CB60(source *Object, runtime OneSecondDieUpdateRuntime53CB60) {
	oneSecondDieUpdateNative53CB60(source, oneSecondDieUpdateNativeDeps53CB60{
		frame:         s.Frame,
		fps:           s.TickRate,
		delayedDelete: runtime.DelayedDelete,
	})
}
