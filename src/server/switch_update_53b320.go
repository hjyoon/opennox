package server

type SwitchUpdateRuntime53B320 struct {
	QueueCollision func(*Object)
}

// SwitchUpdate53B320 restores GAME.EXE 0053B320 without invoking the
// PE32-only object updater on native-width Object records.
func (s *Server) SwitchUpdate53B320(unit *Object, runtime SwitchUpdateRuntime53B320) uint8 {
	if unit == nil {
		return 0
	}
	return SwitchUpdateRaw53B320(unit, SwitchUpdateRawRuntime53B320{
		NeedSync: (*Object).NeedSync,
		QueueCollision: func(o *Object) uint8 {
			if runtime.QueueCollision != nil {
				runtime.QueueCollision(o)
			}
			// Preserve this public API's existing post-service flags result.
			return uint8(o.ObjFlags)
		},
	})
}
