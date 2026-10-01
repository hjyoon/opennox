package server

type questLoseGemsHooks54D080[O comparable] struct {
	loadType      func(int) uint32
	storeType     func(int, uint32)
	lookupType    func(string) uint32
	first, next   func(O) O
	typeIndex     func(O) uint16
	price         func(O) int32
	delayedDelete func(O)
	addGold       func(O, int32)
}

// questLoseGems54D080 preserves GAME.EXE 0054D080's lazy DWORD type cache,
// two inventory walks, WORD-to-DWORD comparisons and cached second-walk next.
// Each odd count removes its first gem for a signed half-price gold credit;
// the remaining loss quota is the truncated signed half of the even count.
func questLoseGems54D080[O comparable](unit O, h questLoseGemsHooks54D080[O]) {
	if h.loadType(0) == 0 {
		for slot, name := range [...]string{"Diamond", "Emerald", "Ruby"} {
			h.storeType(slot, h.lookupType(name))
		}
	}
	var zero O
	var counts [3]int32
	for item := h.first(unit); item != zero; item = h.next(item) {
		diamond := h.loadType(0)
		kind := uint32(h.typeIndex(item))
		switch {
		case kind == diamond:
			counts[0]++
		case kind == h.loadType(1):
			counts[1]++
		case kind == h.loadType(2):
			counts[2]++
		}
	}
	var odd uint8
	for slot := range counts {
		if counts[slot]&1 != 0 {
			odd |= 1 << slot
			counts[slot]--
		}
		counts[slot] /= 2
	}
	for item := h.first(unit); item != zero; {
		next := h.next(item) // Before any type read, price, deletion or credit.
		diamond := h.loadType(0)
		kind := uint32(h.typeIndex(item))
		slot := -1
		switch {
		case kind == diamond:
			slot = 0
		case kind == h.loadType(1):
			slot = 1
		case kind == h.loadType(2):
			slot = 2
		}
		if slot >= 0 {
			mask := uint8(1 << slot)
			if odd&mask != 0 {
				price := h.price(item)
				h.delayedDelete(item)
				h.addGold(unit, price/2)
				odd &^= mask
			} else if counts[slot] > 0 {
				h.delayedDelete(item)
				counts[slot]--
			}
		}
		item = next
	}
}
