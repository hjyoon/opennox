package server

type questLivesSyncHooks4D99E1[O comparable, U any, P comparable] struct {
	gameFlag      func(uint32) int32
	playerByIndex func(int32) P
	playerUnit    func(P) O
	lives         func(U) uint32
	marker        func(U, int32) uint8
	report        func(int32, O) int32
	lifeByte      func(U) uint8
	storeMarker   func(U, int32, uint8)
}

// questLivesSync4D99E1 is the Quest extra-life loop inside GAME.EXE
// 004D9900. update is the entry-cached pointer, not the sender's live
// UpdateData. Compare the full life DWORD with the zero-extended marker
// BYTE, ignore the send result, then reload the life BYTE for the marker.
func questLivesSync4D99E1[O comparable, U any, P comparable](unit O, update U, h questLivesSyncHooks4D99E1[O, U, P]) {
	if h.gameFlag(0x1000) == 0 {
		return
	}
	var zero P
	var zeroUnit O
	for index := int32(0); index < 32; index++ {
		player := h.playerByIndex(index)
		if player == zero || h.playerUnit(player) == zeroUnit {
			continue
		}
		lives := h.lives(update)
		marker := h.marker(update, index)
		if lives == uint32(marker) {
			continue
		}
		h.report(index, unit)
		value := h.lifeByte(update)
		h.storeMarker(update, index, value)
	}
}
