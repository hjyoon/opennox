package server

type ownCollideNativeDeps4EA2C0 struct {
	frame    func() uint32
	setOwner func(*Object, *Object)
}

func ownCollideNative4EA2C0(source, target *Object, deps ownCollideNativeDeps4EA2C0) {
	ownCollide4EA2C0(source, target, ownCollideHooks4EA2C0[*Object]{
		loadTargetClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadSourceOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadFrame: deps.frame,
		storeSourceFrame: func(obj *Object, frame uint32) {
			obj.Field34 = frame
		},
		setOwner: deps.setOwner,
	})
}

// OwnCollide4EA2C0 binds the original OwnCollide callback to native-width
// Object pointers and the server's live frame and ownership state.
func (s *Server) OwnCollide4EA2C0(source, target *Object) {
	ownCollideNative4EA2C0(source, target, ownCollideNativeDeps4EA2C0{
		frame:    s.Frame,
		setOwner: s.ObjSetOwner,
	})
}
