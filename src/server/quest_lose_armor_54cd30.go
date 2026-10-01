package server

type questLoseArmorHooks54CD30[O comparable] struct {
	first, next   func(O) O
	flags, class  func(O) uint32
	typeIndex     func(O) uint16
	armorFlags    func(uint16) uint32
	randomInt     func(int32, int32) int32
	delayedDelete func(O)
}

// questLoseArmor54CD30 preserves the two live inventory walks in GAME.EXE
// 0054CD30. An empty candidate set does not call the RNG. The second walk
// starts from the fresh head and deletes only its matching eligible ordinal.
func questLoseArmor54CD30[O comparable](unit O, h questLoseArmorHooks54CD30[O]) {
	var zero O
	var count int32
	for item := h.first(unit); item != zero; item = h.next(item) {
		if h.flags(item)&0x100 != 0 && h.class(item)&0x02000000 != 0 && h.armorFlags(h.typeIndex(item))&0x405 == 0 {
			count++
		}
	}
	if count == 0 {
		return
	}
	selected := h.randomInt(0, count-1)
	var ordinal int32
	for item := h.first(unit); item != zero; item = h.next(item) {
		if h.flags(item)&0x100 == 0 || h.class(item)&0x02000000 == 0 || h.armorFlags(h.typeIndex(item))&0x405 != 0 {
			continue
		}
		if ordinal == selected {
			h.delayedDelete(item)
			return
		}
		ordinal++
	}
}
