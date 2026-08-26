package server

const rewardMarkerPlusTypeName4F0720 = "RewardMarkerPlus"

// RewardMarkerInitData is the pointer-free 220-byte record registered by
// RewardMarkerInit. GAME.EXE 004F0720 reads CategoryMask at +0 and ChanceMode
// at +212. RewardFlags at +4, Spells at +8, and Abilities at +145 are consumed
// by reward spell-book and ability-book creation at 004F09F0 and 004F0C70.
// Field216 is retained because reward-container callers inspect its high byte
// independently of these routines.
type RewardMarkerInitData struct {
	CategoryMask uint32
	RewardFlags  uint8
	_            [3]byte
	Spells       [137]uint8
	Abilities    [6]uint8
	_            [61]byte
	ChanceMode   uint32
	Field216     uint32
}

type rewardMarkerCategory4F0720 struct {
	Weight uint8
	Field4 uint32
}

// rewardMarkerCategories4F0720 is the native fixed-width meaning of the eight
// original rows at GAME.EXE 005B98C4..005B9903. The first dword's upper three
// bytes are zero; Field4 retains the second dword even though 004F0720 derives
// category masks from row indices and never reads it.
var rewardMarkerCategories4F0720 = [...]rewardMarkerCategory4F0720{
	{Weight: 16, Field4: 2},
	{Weight: 2, Field4: 4},
	{Weight: 2, Field4: 8},
	{Weight: 24, Field4: 16},
	{Weight: 16, Field4: 32},
	{Weight: 23, Field4: 64},
	{Weight: 16, Field4: 128},
	{Weight: 1, Field4: 4},
}

type rewardMarkerDispatch4F0720 uint8

const (
	rewardMarkerDispatchSpellBook4F0720 rewardMarkerDispatch4F0720 = iota
	rewardMarkerDispatchAbilityBook4F0720
	rewardMarkerDispatchFieldGuide4F0720
	rewardMarkerDispatchWeapon4F0720
	rewardMarkerDispatchArmor4F0720
	rewardMarkerDispatchGem4F0720
	rewardMarkerDispatchPotion4F0720
	rewardMarkerDispatchGem2_4F0720
	rewardMarkerDispatchDefaultGem4F0720
)

// rewardMarkerDispatchIndex4F0720 is the meaning of the original selector
// byte table at 004F0970. Only exact one-hot values select the first eight
// entries; zero, non-powers of two, and values above 128 select entry eight,
// which reaches the same gem creator as entry five.
func rewardMarkerDispatchIndex4F0720(selector uint32) rewardMarkerDispatch4F0720 {
	switch selector {
	case 1:
		return rewardMarkerDispatchSpellBook4F0720
	case 2:
		return rewardMarkerDispatchAbilityBook4F0720
	case 4:
		return rewardMarkerDispatchFieldGuide4F0720
	case 8:
		return rewardMarkerDispatchWeapon4F0720
	case 16:
		return rewardMarkerDispatchArmor4F0720
	case 32:
		return rewardMarkerDispatchGem4F0720
	case 64:
		return rewardMarkerDispatchPotion4F0720
	case 128:
		return rewardMarkerDispatchGem2_4F0720
	default:
		return rewardMarkerDispatchDefaultGem4F0720
	}
}

type rewardMarkerActivateHooks4F0720[O, D, R any] struct {
	loadCachedPlusType  func() uint32
	loadInitData        func(O) D
	lookupType          func(string) uint32
	storeCachedPlusType func(uint32)
	loadTypeInd         func(O) uint16
	loadChanceMode      func(D) uint32
	randomInt           func(int32, int32) int32
	loadCategoryMask    func(D) uint32
	dispatch            func(rewardMarkerDispatch4F0720, O, uint32) R
}

// rewardMarkerActivate4F0720 preserves GAME.EXE 004F0720's observable order.
// The RewardMarkerPlus cache is loaded before the marker's InitData pointer;
// a missing cache is resolved only after that pointer has been cached. The
// type comparison may add two to stage with uint32 wrapping semantics.
//
// Chance modes one through four perform one inclusive 0..100 logic-RNG draw
// and reject values above 75, 50, 25, and 5 respectively. Other modes skip
// that draw. The category mask is loaded once for the total-weight pass and
// loaded again after the inclusive 1..total draw for the selection pass. The
// cumulative comparison is unsigned. If the second pass does not select a
// category, the possibly adjusted stage is sent through the original selector
// mapping. There are deliberately no nil, callback, or lookup-result guards.
func rewardMarkerActivate4F0720[O, D, R any](
	marker O,
	stage uint32,
	hooks rewardMarkerActivateHooks4F0720[O, D, R],
) R {
	cachedPlusType := hooks.loadCachedPlusType()
	initData := hooks.loadInitData(marker)
	if cachedPlusType == 0 {
		cachedPlusType = hooks.lookupType(rewardMarkerPlusTypeName4F0720)
		hooks.storeCachedPlusType(cachedPlusType)
	}
	if uint32(hooks.loadTypeInd(marker)) == cachedPlusType {
		stage += 2
	}

	var chanceThreshold int32
	switch hooks.loadChanceMode(initData) {
	case 1:
		chanceThreshold = 75
	case 2:
		chanceThreshold = 50
	case 3:
		chanceThreshold = 25
	case 4:
		chanceThreshold = 5
	default:
		chanceThreshold = -1
	}
	if chanceThreshold >= 0 && hooks.randomInt(0, 100) > chanceThreshold {
		var zero R
		return zero
	}

	firstMask := hooks.loadCategoryMask(initData)
	var total uint32
	for index, row := range rewardMarkerCategories4F0720 {
		categoryMask := uint32(1) << index
		if firstMask&categoryMask != 0 {
			total += uint32(row.Weight)
		}
	}
	if total == 0 {
		var zero R
		return zero
	}

	draw := uint32(hooks.randomInt(1, int32(total)))
	secondMask := hooks.loadCategoryMask(initData)
	selector := stage
	var cumulative uint32
	for index, row := range rewardMarkerCategories4F0720 {
		categoryMask := uint32(1) << index
		if secondMask&categoryMask == 0 {
			continue
		}
		cumulative += uint32(row.Weight)
		if cumulative >= draw {
			selector = categoryMask
			break
		}
	}
	return hooks.dispatch(rewardMarkerDispatchIndex4F0720(selector), marker, stage)
}
