package server

const (
	rewardSpellBookCommonType4F09F0   = "CommonSpellBook"
	rewardSpellBookWizardType4F09F0   = "WizardSpellBook"
	rewardSpellBookConjurerType4F09F0 = "ConjurerSpellBook"
	rewardSpellBookExplicitFlag4F09F0 = uint8(1)
	rewardSpellBookSpellCount4F09F0   = 137
)

type rewardSpellBookHooks4F09F0[M, D, O any] struct {
	loadInitData       func(M) D
	loadFlags          func(D) uint8
	loadExplicitSpell  func(D, int) uint8
	pickSlots          func(uint32) uint32
	rows               []rewardSpellDefinition4F09F0
	randomInt          func(int32, int32) int32
	checkSpellClass    func(uint32, uint32) int32
	createObjectByType func(string) O
	isNilObject        func(O) bool
	storeSpell         func(O, uint8)
}

func rewardSpellBookExplicit4F09F0[M, D, O any](
	data D,
	hooks rewardSpellBookHooks4F09F0[M, D, O],
) (uint32, bool) {
	count := int32(0)
	for index := 0; index < rewardSpellBookSpellCount4F09F0; index++ {
		if hooks.loadExplicitSpell(data, index) == 1 {
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, count-1)
	ordinal := int32(0)
	for index := 0; index < rewardSpellBookSpellCount4F09F0; index++ {
		if hooks.loadExplicitSpell(data, index) == 1 {
			if ordinal == draw {
				return uint32(index), true
			}
			ordinal++
		}
	}
	return 0, false
}

func rewardSpellBookAutomatic4F09F0[M, D, O any](
	stage uint32,
	hooks rewardSpellBookHooks4F09F0[M, D, O],
) (uint32, bool) {
	slots := hooks.pickSlots(stage)
	if hooks.rows[0].SpellID == 0 {
		return 0, false
	}
	var total int32
	for index := 0; ; index++ {
		row := &hooks.rows[index]
		if slots&row.Slots != 0 {
			total += int32(row.Weight)
		}
		if hooks.rows[index+1].SpellID == 0 {
			break
		}
	}
	if total == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, total-1)
	if hooks.rows[0].SpellID == 0 {
		return 0, false
	}
	var cumulative int32
	for index := 0; ; index++ {
		row := &hooks.rows[index]
		if slots&row.Slots != 0 {
			cumulative += int32(row.Weight)
			if draw < cumulative {
				return row.SpellID, true
			}
		}
		if hooks.rows[index+1].SpellID == 0 {
			return 0, false
		}
	}
}

func rewardSpellBookCreate4F09F0[M, D, O any](
	spellID uint32,
	hooks rewardSpellBookHooks4F09F0[M, D, O],
) O {
	var zero O
	bookType := ""
	if hooks.checkSpellClass(1, spellID) == 0 {
		if hooks.checkSpellClass(2, spellID) == 0 {
			bookType = rewardSpellBookCommonType4F09F0
		}
	}
	if bookType == "" {
		if hooks.checkSpellClass(1, spellID) == 0 {
			bookType = rewardSpellBookWizardType4F09F0
		} else {
			if hooks.checkSpellClass(2, spellID) != 0 {
				return zero
			}
			bookType = rewardSpellBookConjurerType4F09F0
		}
	}
	result := hooks.createObjectByType(bookType)
	if hooks.isNilObject(result) {
		return zero
	}
	hooks.storeSpell(result, uint8(spellID))
	return result
}

// rewardSpellBook4F09F0 reconstructs GAME.EXE 004F09F0. InitData is cached
// at entry. Explicit spell bytes and the automatic reward table are read in
// two passes around their respective RNG callbacks. Class checks deliberately
// retain the original repeated short-circuit order, and there are no callback
// or data guards beyond the original zero-result branches.
func rewardSpellBook4F09F0[M, D, O any](
	marker M,
	stage uint32,
	hooks rewardSpellBookHooks4F09F0[M, D, O],
) O {
	var zero O
	data := hooks.loadInitData(marker)
	var spellID uint32
	var selected bool
	if hooks.loadFlags(data)&rewardSpellBookExplicitFlag4F09F0 != 0 {
		spellID, selected = rewardSpellBookExplicit4F09F0(data, hooks)
	} else {
		spellID, selected = rewardSpellBookAutomatic4F09F0(stage, hooks)
	}
	if !selected || spellID == 0 {
		return zero
	}
	return rewardSpellBookCreate4F09F0(spellID, hooks)
}
