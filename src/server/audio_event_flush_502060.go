package server

const (
	audioEventFlushWordCount502060   = int32(0x20)
	audioEventFlushBitsPerWord502060 = int32(0x20)
)

// audioEventFlushHooks502060 describes the observable loads and dispatch made
// by GAME.EXE 00502060. Pointer-shaped listener, descriptor, and event values
// remain native-width; the descriptor limit and event percentage are PE32
// signed dwords.
type audioEventFlushHooks502060[L, D any, E comparable] struct {
	loadBitmapWord func(int32) uint32
	descriptor     func(int32) D
	loadHead       func(D) E
	loadLimit      func(D) int32
	loadPercentage func(E) int32
	send           func(L, E, int32) bool
	loadNext       func(E) E
}

// flushAudioEvents502060 preserves GAME.EXE 00502060's traversal and cache
// boundaries. A bitmap word and each selected descriptor's signed limit are
// cached, while later bitmap words and an event's next link are loaded only
// when the original routine reaches them. The packet-dispatch return value is
// deliberately ignored.
func flushAudioEvents502060[L, D any, E comparable](
	listener L,
	hooks audioEventFlushHooks502060[L, D, E],
) {
	var nilEvent E
	soundBase := int32(0)
	for word := int32(0); word < audioEventFlushWordCount502060; word++ {
		mask := hooks.loadBitmapWord(word)
		if mask != 0 {
			for bit := int32(0); bit < audioEventFlushBitsPerWord502060; bit++ {
				if mask&1 != 0 {
					descriptor := hooks.descriptor(soundBase + bit)
					event := hooks.loadHead(descriptor)
					remaining := hooks.loadLimit(descriptor)
					for event != nilEvent && remaining > 0 {
						remaining--
						percentage := hooks.loadPercentage(event)
						_ = hooks.send(listener, event, percentage)
						event = hooks.loadNext(event)
					}
				}
				mask >>= 1
			}
		}
		soundBase += audioEventFlushBitsPerWord502060
	}
}
