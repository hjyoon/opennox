package legacy

type questPreviewSpec44E110 struct {
	slot int
	name string
}

// Slots follow the extracted C symbols 832492..832536, not PE32 pointer
// arithmetic. The original factory visits the exit before the generator.
var questPreviewSpecs44E110 = [...]questPreviewSpec44E110{
	{1, "GauntletExitB"}, {0, "BeholderGenerator"}, {2, "Ankh"},
	{3, "SoulGate"}, {4, "SilverKey"}, {5, "GoldKey"},
	{6, "QuestGoldChest"}, {7, "QuestGoldPile"}, {8, "DunMirChest4"},
	{9, "WarHammer"}, {10, "HastePotion"}, {11, "ConjurerSpellBook"},
}

type questPreviewHooks44E110[D, F comparable] struct {
	loadFont    func() F
	fontByName  func(string) F
	storeFont   func(F)
	load        func(int) D
	thingByName func(string) int32
	create      func(int32) D
	store       func(int, D)
	mark        func(D)
}

// questPreview44E110 is the complete sealed 0044E110..0044E31E body.
// Each slot is loaded only when reached; callbacks can change later caches.
// A newly created value is published before its flag write, including nil.
// Cached drawables get the same flag write on every call.
func questPreview44E110[D, F comparable](h questPreviewHooks44E110[D, F]) D {
	var zeroF F
	if h.loadFont() == zeroF {
		h.storeFont(h.fontByName("default"))
	}
	var zeroD, last D
	for _, spec := range questPreviewSpecs44E110 {
		d := h.load(spec.slot)
		if d == zeroD {
			d = h.create(h.thingByName(spec.name))
			h.store(spec.slot, d)
		}
		h.mark(d)
		last = d
	}
	return last
}
