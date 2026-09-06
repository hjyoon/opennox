package server

const (
	itemCancelDurSpellsClass4FEB60      = uint32(0x1000)
	itemCancelDurSpellsSpell43Bit4FEB60 = uint32(0x40000)
	itemCancelDurSpellsSpell59Bit4FEB60 = uint32(0x4000000)
	itemCancelDurSpellsSpell43ID4FEB60  = int32(43)
	itemCancelDurSpellsSpell59ID4FEB60  = int32(59)
)

// ItemCancelDurSpellsHooks4FEB60 exposes every observable argument and item
// field load plus both cancellation callbacks in GAME.EXE 004FEB60. Item and
// object tokens remain native-width values rather than inheriting PE32 pointer
// truncation.
type ItemCancelDurSpellsHooks4FEB60[Item, Object any] struct {
	LoadItemArg  func() Item
	LoadClass    func(Item) uint32
	LoadSubclass func(Item) uint32
	LoadOwnerArg func() Object
	Cancel       func(int32, Object)
}

// ItemCancelDurSpells4FEB60 preserves GAME.EXE 004FEB60's exact access and
// callback order. The item is dereferenced without a nil guard. Its class is
// tested before either subclass or owner is loaded. Eligible items snapshot
// subclass once for spell 43, cache the owner once, then reload live subclass
// after the first callback before deciding whether to cancel spell 59. A nil
// owner remains a valid identity forwarded to the cancellation routine.
func ItemCancelDurSpells4FEB60[Item, Object any](h ItemCancelDurSpellsHooks4FEB60[Item, Object]) {
	item := h.LoadItemArg()
	if h.LoadClass(item)&itemCancelDurSpellsClass4FEB60 == 0 {
		return
	}

	subclass := h.LoadSubclass(item)
	owner := h.LoadOwnerArg()
	if subclass&itemCancelDurSpellsSpell43Bit4FEB60 != 0 {
		h.Cancel(itemCancelDurSpellsSpell43ID4FEB60, owner)
	}
	if h.LoadSubclass(item)&itemCancelDurSpellsSpell59Bit4FEB60 != 0 {
		h.Cancel(itemCancelDurSpellsSpell59ID4FEB60, owner)
	}
}
