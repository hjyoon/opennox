package server

// OpenUpdate53DBB0 restores the object status transition at GAME.EXE
// 0053DBB0 without an indirect C callback. Its fixed-width status fields
// retain their original byte offsets on a native-width Object.
func (s *Server) OpenUpdate53DBB0(obj *Object) {
	if obj == nil {
		return
	}
	switch {
	case obj.Field5&0x2 != 0:
		if uint32(obj.ObjFlags)&0x8000 != 0 {
			obj.UnsetXStatus(0x2)
			obj.SetXStatus(0x4)
			obj.Field34 = s.Frame() + 2*s.TickRate()
		}
	case obj.Field5&0x4 != 0:
		if s.Frame() > obj.Field34 {
			obj.UnsetXStatus(0x4)
			obj.SetXStatus(0x8)
		}
	case obj.Field5&0x8 != 0:
		s.Objs.RemoveFromUpdatable(obj)
	}
}
