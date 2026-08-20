package server

const glyphInventoryCountName4EF6F0 = "Glyph"

type glyphInventoryCountHooks4EF6F0[O comparable] struct {
	loadCache  func() uint32
	lookupType func(string) uint32
	storeCache func(uint32)
	loadFirst  func(O) O
	loadType   func(O) uint16
	loadNext   func(O) O
}

func incrementGlyphInventoryCount4EF6F0(count int32) int32 {
	return count + 1
}

// glyphInventoryCount4EF6F0 preserves GAME.EXE 004EF6F0. The global type
// cache is read before any owner or inventory access. Only an entry-zero cache
// invokes the exact "Glyph" lookup and stores its whole 32-bit result; a zero
// lookup result is not retried during the same call.
//
// The first inventory node is loaded without a nil-owner guard. Each visited
// node then reloads the live whole cache before reading the node's
// zero-extended 16-bit TypeInd. Matching nodes increment a wrapping signed
// 32-bit count without reading flags, so Destroyed Glyphs are included. The
// live successor is loaded only after the comparison and possible increment.
func glyphInventoryCount4EF6F0[O comparable](
	owner O,
	hooks glyphInventoryCountHooks4EF6F0[O],
) int32 {
	if hooks.loadCache() == 0 {
		resolved := hooks.lookupType(glyphInventoryCountName4EF6F0)
		hooks.storeCache(resolved)
	}

	var zero O
	count := int32(0)
	for item := hooks.loadFirst(owner); item != zero; item = hooks.loadNext(item) {
		cachedType := hooks.loadCache()
		itemType := hooks.loadType(item)
		if uint32(itemType) == cachedType {
			count = incrementGlyphInventoryCount4EF6F0(count)
		}
	}
	return count
}
