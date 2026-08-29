package server

// ExpireUpdateRuntime53DB00 supplies the deletion service reached when the
// current frame lies outside an object's inclusive lifetime interval.
type ExpireUpdateRuntime53DB00 struct {
	DelayedDelete func(*Object)
}

func expireUpdateNative53DB00(source *Object, frame uint32, runtime ExpireUpdateRuntime53DB00) {
	if source.Field32 > frame || source.Field34 < frame {
		runtime.DelayedDelete(source)
	}
}

// ExpireUpdate53DB00 restores GAME.EXE 0053DB00 without narrowing the source
// object to an ABI32 integer. GAME.EXE caches the frame once before comparing
// the start and end fields, so the server frame is sampled exactly once here.
func (s *Server) ExpireUpdate53DB00(source *Object, runtime ExpireUpdateRuntime53DB00) {
	frame := s.Frame()
	expireUpdateNative53DB00(source, frame, runtime)
}
