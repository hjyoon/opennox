package legacy

const unitPetSubclassBit4E7B00 uint32 = 0x00000080

type unitPetHooks4E7B00[O comparable, U any, P any] struct {
	subclass    func(O) uint32
	setSubclass func(O, uint32)
	updateData  func(O) U
	player      func(U) P
	playerInd   func(P) byte
	monitor     func(byte, O)
	mark        func(byte, O, uint32)
	setOwner    func(O, O)
	unmonitor   func(byte, O)
	unmark      func(byte, O, uint32)
	clearOwner  func(O)
}

// unitBecomePet4E7B00 preserves GAME.EXE 004E7B00. The owner's update-data
// pointer is cached before the pet subclass write, while the Player pointer
// inside that cached update-data is reloaded after the monitor callback.
func unitBecomePet4E7B00[O comparable, U any, P any](owner, pet O, h unitPetHooks4E7B00[O, U, P]) {
	var zero O
	if owner == zero {
		return
	}
	if pet == zero {
		return
	}

	subclass := h.subclass(pet)
	update := h.updateData(owner)
	h.setSubclass(pet, subclass|unitPetSubclassBit4E7B00)

	player := h.player(update)
	h.monitor(h.playerInd(player), pet)
	player = h.player(update)
	h.mark(h.playerInd(player), pet, 1)
	h.setOwner(owner, pet)
}

// unitBecomeEnemy4E7B60 preserves GAME.EXE 004E7B60. Unlike the pet path,
// the original reads owner update-data before checking owner for null, so a
// null owner faults before the pet argument is considered.
func unitBecomeEnemy4E7B60[O comparable, U any, P any](owner, pet O, h unitPetHooks4E7B00[O, U, P]) {
	update := h.updateData(owner)
	var zero O
	if owner == zero {
		return
	}
	if pet == zero {
		return
	}

	subclass := h.subclass(pet)
	h.setSubclass(pet, subclass&^unitPetSubclassBit4E7B00)

	player := h.player(update)
	h.unmonitor(h.playerInd(player), pet)
	player = h.player(update)
	h.unmark(h.playerInd(player), pet, 1)
	h.clearOwner(pet)
}
