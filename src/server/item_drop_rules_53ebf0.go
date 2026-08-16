package server

// itemDropRuleNames53EBF0 is the exact 19-row table at GAME.EXE
// 005CB388. The two Lich orb keys are present in the membership table but
// carry a zero drop mask, matching the original data rows.
var itemDropRuleNames53EBF0 = [...]struct {
	name string
	mask uint32
}{
	{"BookOfOblivion", 1},
	{"OblivionHalberd", 1},
	{"OblivionHeart", 1},
	{"OblivionWierdling", 1},
	{"OblivionOrb", 1},
	{"AmuletOfClarity", 1},
	{"AmuletOfTeleportation", 1},
	{"BridgeGuardsBoots", 1},
	{"Spectacles", 1},
	{"MatildasCloak", 1},
	{"ThaviusStaff", 1},
	{"BlueOrbKeyOfTheLich", 0},
	{"RedOrbKeyOfTheLich", 0},
	{"SilverKey", 1},
	{"GoldKey", 1},
	{"SapphireKey", 1},
	{"RubyKey", 1},
	{"SponsorshipLetter", 1},
	{"MayorsScepter", 1},
}

type itemDropRules53EBF0 struct {
	initialized bool
	typeInd     [len(itemDropRuleNames53EBF0)]uint32
}

// initialize53EC40 preserves the one-shot lookup order of GAME.EXE 0053EC40.
// Lookup results retain all 32 bits even though Object.TypeInd is 16-bit.
func (r *itemDropRules53EBF0) initialize53EC40(lookup func(string) uint32) {
	if r.initialized {
		return
	}
	for i, row := range itemDropRuleNames53EBF0 {
		r.typeInd[i] = lookup(row.name)
	}
	r.initialized = true
}

func (r *itemDropRules53EBF0) itemIsDroppable53EBF0(typeInd uint16) int32 {
	want := uint32(typeInd)
	for _, got := range r.typeInd {
		if got == want {
			return 1
		}
	}
	return 0
}

func (r *itemDropRules53EBF0) itemDropMask53EC80(typeInd uint16, mask uint32) int32 {
	want := uint32(typeInd)
	for i, got := range r.typeInd {
		if got == want {
			if itemDropRuleNames53EBF0[i].mask&mask != 0 {
				return 1
			}
			return 0
		}
	}
	return 0
}

// ItemIsDroppable53EBF0 is the native-width form of GAME.EXE 0053EBF0.
func (s *Server) ItemIsDroppable53EBF0(item *Object) int32 {
	if item == nil {
		return 0
	}
	s.itemDropRules.initialize53EC40(func(name string) uint32 {
		return uint32(s.Types.IndByID(name))
	})
	return s.itemDropRules.itemIsDroppable53EBF0(item.TypeInd)
}

// ItemDropMask53EC80 is the native-width form of GAME.EXE 0053EC80.
func (s *Server) ItemDropMask53EC80(item *Object, mask uint32) int32 {
	if item == nil {
		return 0
	}
	s.itemDropRules.initialize53EC40(func(name string) uint32 {
		return uint32(s.Types.IndByID(name))
	})
	return s.itemDropRules.itemDropMask53EC80(item.TypeInd, mask)
}
