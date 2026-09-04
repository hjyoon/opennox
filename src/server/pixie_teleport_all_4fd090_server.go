package server

type pixieTeleportAllNativeDeps4FD090 struct {
	loadPixieTypeID func() uint32
	teleport        func(*Object, *Object)
}

func pixieTeleportAllNative4FD090(
	owner *Object,
	deps pixieTeleportAllNativeDeps4FD090,
) {
	pixieTeleportAll4FD090(pixieTeleportAllHooks4FD090[*Object, *PixieUpdateData]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadFirstOwned: func(owner *Object) *Object {
			return owner.Field129
		},
		loadPixieTypeID: deps.loadPixieTypeID,
		loadTypeInd: func(pixie *Object) uint16 {
			return pixie.TypeInd
		},
		loadFlags: func(pixie *Object) uint32 {
			return uint32(pixie.ObjFlags)
		},
		loadUpdateData: func(pixie *Object) *PixieUpdateData {
			return (*PixieUpdateData)(pixie.UpdateData)
		},
		loadTarget: func(updateData *PixieUpdateData) *Object {
			return updateData.Target
		},
		teleport: deps.teleport,
		loadNextOwned: func(pixie *Object) *Object {
			return pixie.Field128
		},
	})
}

// PixieTeleportAll4FD090 binds GAME.EXE 004FD090 to native-width Object and
// PixieUpdateData pointers. The type-ID loader remains live per owned object,
// and teleport receives the current Pixie followed by the cached owner.
//
//go:noinline
func (*Server) PixieTeleportAll4FD090(
	owner *Object,
	loadPixieTypeID func() uint32,
	teleport func(*Object, *Object),
) {
	pixieTeleportAllNative4FD090(owner, pixieTeleportAllNativeDeps4FD090{
		loadPixieTypeID: loadPixieTypeID,
		teleport:        teleport,
	})
}
