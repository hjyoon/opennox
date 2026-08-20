package server

type glyphInventoryCountNativeDeps4EF6F0 struct {
	loadCache  func() uint32
	lookupType func(string) uint32
	storeCache func(uint32)
}

func glyphInventoryCountNative4EF6F0(
	owner *Object,
	deps glyphInventoryCountNativeDeps4EF6F0,
) int32 {
	return glyphInventoryCount4EF6F0(owner, glyphInventoryCountHooks4EF6F0[*Object]{
		loadCache:  deps.loadCache,
		lookupType: deps.lookupType,
		storeCache: deps.storeCache,
		loadFirst: func(owner *Object) *Object {
			// Match 004E7980: there is deliberately no nil-owner guard.
			return owner.InvFirstItem
		},
		loadType: func(item *Object) uint16 {
			return item.TypeInd
		},
		loadNext: func(item *Object) *Object {
			return item.InvNextItem
		},
	})
}

func glyphInventoryCountServerDeps4EF6F0(s *Server) glyphInventoryCountNativeDeps4EF6F0 {
	return glyphInventoryCountNativeDeps4EF6F0{
		loadCache: s.Types.playerRespawnGlyphIDCached4EF6F0,
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeCache: s.Types.storePlayerRespawnGlyphID4EF6F0,
	}
}

// GlyphInventoryCount4EF6F0 binds GAME.EXE 004EF6F0 to the function's own
// fixed-width type cache and native-width Object inventory links. TypeInd and
// the wrapping return count retain their original 16- and 32-bit widths.
func (s *Server) GlyphInventoryCount4EF6F0(owner *Object) int32 {
	return glyphInventoryCountNative4EF6F0(
		owner,
		glyphInventoryCountServerDeps4EF6F0(s),
	)
}
