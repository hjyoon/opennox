package server

const trapDropRejectAudio4ED580 = uint32(925)

// trapDropHooks4ED580 exposes the delayed argument loads in GAME.EXE
// 004ED580. The point pointer is cached before MapTileAllowTeleport, while the
// owner and glyph arguments are loaded only after that callback returns.
type trapDropHooks4ED580[O, P comparable] struct {
	loadPointArg func() P
	mapTile      func(P) int32

	loadOwnerArg func() O
	loadNetCode  func(O) uint32
	audio        func(uint32, O, int32, uint32)

	loadGlyphArg func() O
	defaultDrop  func(O, O, P) int32
	setOwner     func(O, O)
}

// trapDrop4ED580 preserves GAME.EXE 004ED580. Both callback gates test the
// whole EAX value. A forbidden tile reports the owner's live NetCode, while a
// successful DefaultDrop transfers ownership using the cached owner and glyph
// arguments.
func trapDrop4ED580[O, P comparable](hooks trapDropHooks4ED580[O, P]) int32 {
	point := hooks.loadPointArg()
	if hooks.mapTile(point) != 0 {
		owner := hooks.loadOwnerArg()
		code := hooks.loadNetCode(owner)
		hooks.audio(trapDropRejectAudio4ED580, owner, 2, code)
		return 0
	}

	owner := hooks.loadOwnerArg()
	glyph := hooks.loadGlyphArg()
	if hooks.defaultDrop(owner, glyph, point) == 0 {
		return 0
	}
	hooks.setOwner(owner, glyph)
	return 1
}
