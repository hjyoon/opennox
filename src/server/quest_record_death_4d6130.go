package server

const questRecordDeathDestroyed4D6130 = uint32(0x20)

type questRecordDeathHooks4D6130[O, U, P comparable] struct {
	loadFlags   func(O) uint32
	loadUpdate  func(O) U
	loadPlayer  func(U) P
	loadDeaths  func(P) uint32
	storeDeaths func(P, uint32)
	loadMask    func(P) uint32
	storeMask   func(P, uint32)
}

// The original EAX contains the input Object on the nil/destroyed paths and
// the second live Player on the counted path. Keep those pointer identities
// separate instead of carrying them through a PE32 integer.
type questRecordDeathResult4D6130[O, P comparable] struct {
	unit     O
	player   P
	isPlayer bool
}

// questRecordDeath4D6130 restores GAME.EXE 004D6130's exact load/store order.
// UpdateData is cached, Player is reloaded after the wrapping DWORD increment,
// and bit 2 is ORed into that second Player's progress mask. Apart from the
// original nil-unit and destroyed-bit checks, no guards are added.
func questRecordDeath4D6130[O, U, P comparable](unit O, h questRecordDeathHooks4D6130[O, U, P]) questRecordDeathResult4D6130[O, P] {
	var zero O
	if unit == zero || h.loadFlags(unit)&questRecordDeathDestroyed4D6130 != 0 {
		return questRecordDeathResult4D6130[O, P]{unit: unit}
	}
	update := h.loadUpdate(unit)
	firstPlayer := h.loadPlayer(update)
	h.storeDeaths(firstPlayer, h.loadDeaths(firstPlayer)+1)
	player := h.loadPlayer(update)
	h.storeMask(player, h.loadMask(player)|2)
	return questRecordDeathResult4D6130[O, P]{player: player, isPlayer: true}
}
