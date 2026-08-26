package server

const (
	rewardAbilityBookType4F0C70         = "AbilityBook"
	rewardAbilityBookExplicitFlag4F0C70 = uint8(2)
	rewardAbilityBookCount4F0C70        = 6
)

type rewardAbilityBookHooks4F0C70[M, D, O any] struct {
	loadInitData        func(M) D
	loadFlags           func(D) uint8
	loadExplicitAbility func(D, int) uint8
	randomInt           func(int32, int32) int32
	createObjectByType  func(string) O
	isNilObject         func(O) bool
	storeAbility        func(O, uint8)
}

func rewardAbilityBookExplicit4F0C70[M, D, O any](
	data D,
	hooks rewardAbilityBookHooks4F0C70[M, D, O],
) (int32, bool) {
	count := int32(0)
	for index := 0; index < rewardAbilityBookCount4F0C70; index++ {
		if hooks.loadExplicitAbility(data, index) == 1 {
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, count-1)
	ordinal := int32(0)
	for index := 0; index < rewardAbilityBookCount4F0C70; index++ {
		if hooks.loadExplicitAbility(data, index) == 1 {
			if ordinal == draw {
				return int32(index), true
			}
			ordinal++
		}
	}
	return 0, false
}

// rewardAbilityBook4F0C70 reconstructs GAME.EXE 004F0C70. InitData is
// cached at entry. Explicit ability bytes are read in two passes around the
// RNG callback and only exact byte value one is enabled. Automatic selection
// preserves the full signed RNG callback result until the zero-ID gate, then
// stores its low byte after successful object creation. There are deliberately
// no nil, callback, object-data, or range guards beyond original branches.
func rewardAbilityBook4F0C70[M, D, O any](
	marker M,
	hooks rewardAbilityBookHooks4F0C70[M, D, O],
) O {
	var zero O
	data := hooks.loadInitData(marker)
	var abilityID int32
	if hooks.loadFlags(data)&rewardAbilityBookExplicitFlag4F0C70 != 0 {
		selected := false
		abilityID, selected = rewardAbilityBookExplicit4F0C70(data, hooks)
		if !selected {
			return zero
		}
	} else {
		abilityID = hooks.randomInt(1, 5)
	}
	if abilityID == 0 {
		return zero
	}
	result := hooks.createObjectByType(rewardAbilityBookType4F0C70)
	if hooks.isNilObject(result) {
		return zero
	}
	hooks.storeAbility(result, uint8(abilityID))
	return result
}
