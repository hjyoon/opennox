package server

type playerQuestKeysReportHooks4D9A3F[O comparable, U any, P comparable] struct {
	loadCachedType  func(kind int) uint32
	lookupType      func(kind int) uint32
	storeCachedType func(kind int, value uint32)
	playerByIndex   func(int32) P
	firstItem       func(O) O
	nextItem        func(O) O
	typeInd         func(O) uint16
	marker          func(U, int, int32) uint8
	storeMarker     func(U, int, int32, uint8)
	report          func(int32, O, int, uint8) int32
}

// playerQuestKeysReport4D9A3F restores the two key loops inside GAME.EXE
// 004D9900 after the Quest flag has already admitted the block. Every slot
// retries a zero type cache before looking up its recipient. Each candidate
// reloads the full cache DWORD before reading the item type WORD. Unlike the
// life loop, a recipient need not have a unit. The computed presence survives
// the send callback and is stored through the entry-cached update pointer.
func playerQuestKeysReport4D9A3F[O comparable, U any, P comparable](unit O, update U, h playerQuestKeysReportHooks4D9A3F[O, U, P]) {
	var zeroPlayer P
	var zeroItem O
	for kind := 0; kind < 2; kind++ {
		for index := int32(0); index < 32; index++ {
			presence := uint8(0)
			if h.loadCachedType(kind) == 0 {
				value := h.lookupType(kind)
				h.storeCachedType(kind, value)
			}
			if h.playerByIndex(index) == zeroPlayer {
				continue
			}
			for item := h.firstItem(unit); item != zeroItem; item = h.nextItem(item) {
				cached := h.loadCachedType(kind)
				if uint32(h.typeInd(item)) == cached {
					presence = 1
					break
				}
			}
			if presence != h.marker(update, kind, index) {
				h.report(index, unit, kind, presence)
				h.storeMarker(update, kind, index, presence)
			}
		}
	}
}
