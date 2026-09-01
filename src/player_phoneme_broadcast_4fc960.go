package opennox

type playerPhonemeBroadcastHooks4FC960[O comparable] struct {
	firstUnit       func() O
	nextUnit        func(O) O
	loadNetCode     func(O) uint32
	spellGetPhoneme func(uint32, int8) int32
	audioEvent      func(int32, O, int32, uint32)
}

// playerPhonemeBroadcast4FC960 restores the observable ordering of
// GAME.EXE 004FC960 without depending on a PE32 object or list layout.
func playerPhonemeBroadcast4FC960[O comparable](
	source O,
	phoneme int8,
	h playerPhonemeBroadcastHooks4FC960[O],
) int32 {
	var nilObject O
	for unit := h.firstUnit(); unit != nilObject; unit = h.nextUnit(unit) {
		if unit == source {
			continue
		}

		listenerNetCode := h.loadNetCode(unit)
		sourceNetCode := h.loadNetCode(source)
		soundID := h.spellGetPhoneme(sourceNetCode, phoneme)
		h.audioEvent(soundID, source, 2, listenerNetCode)
	}
	return 0
}
