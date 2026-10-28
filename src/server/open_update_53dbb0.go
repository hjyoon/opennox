package server

// OpenUpdate53DBB0 restores the object status transition at GAME.EXE
// 0053DBB0 without an indirect C callback. It uses typed native fields
// and preserves the original status-service ordering.
func (s *Server) OpenUpdate53DBB0(obj *Object) {
	if obj == nil {
		return
	}
	OpenUpdateRaw53DBB0(obj, s.stateUpdateRuntime53DB30())
}
