package server

const (
	unitBanishGlyphName5017F0  = "Glyph"
	unitBanishBlueSparks5017F0 = uint8(0x81)
)

// unitBanishHooks5017F0 exposes every observable load and call made by
// GAME.EXE 005017F0. O remains a native-width handle. Object type IDs retain
// the original whole-dword cache and zero-extended word field widths.
type unitBanishHooks5017F0[O comparable] struct {
	loadGlyphCache  func() uint32
	lookupType      func(string) uint32
	storeGlyphCache func(uint32)
	loadFirst       func(O) O
	loadNext        func(O) O
	loadTypeIndex   func(O) uint16
	delayedDelete   func(O)
	sendPointFX     func(uint8, O)
}

// unitBanish5017F0 preserves GAME.EXE 005017F0. The function initializes its
// dedicated Glyph cache before testing a nil unit. Each inventory iteration
// reloads that whole cache, snapshots the successor, then reads TypeInd before
// invoking delayed deletion. Thus deletion callbacks cannot redirect the
// current traversal, but cache changes are visible on the next iteration.
//
// After the traversal it sends blue sparks at the live unit position and
// schedules the unit for deletion. The shipped executable does not dispatch a
// script death event here.
func unitBanish5017F0[O comparable](unit O, hooks unitBanishHooks5017F0[O]) {
	if hooks.loadGlyphCache() == 0 {
		resolved := hooks.lookupType(unitBanishGlyphName5017F0)
		hooks.storeGlyphCache(resolved)
	}

	var nilObject O
	if unit == nilObject {
		return
	}
	for item := hooks.loadFirst(unit); item != nilObject; {
		glyphType := hooks.loadGlyphCache()
		next := hooks.loadNext(item)
		itemType := hooks.loadTypeIndex(item)
		if uint32(itemType) == glyphType {
			hooks.delayedDelete(item)
		}
		item = next
	}
	hooks.sendPointFX(unitBanishBlueSparks5017F0, unit)
	hooks.delayedDelete(unit)
}
