package server

const (
	rewardFieldGuideType4F0D20         = "FieldGuide"
	rewardFieldGuideExplicitFlag4F0D20 = uint8(4)
	rewardFieldGuideCount4F0D20        = 41
)

type rewardFieldGuideHooks4F0D20[M, D, R, O, U any] struct {
	loadInitData       func(M) D
	loadFlags          func(D) uint8
	loadExplicitGuide  func(D, int) uint8
	pickSlots          func(uint32) uint32
	rows               R
	loadRowWeight      func(R, int) uint8
	loadRowGuideID     func(R, int) uint32
	loadRowSlots       func(R, int) uint32
	randomInt          func(int32, int32) int32
	createObjectByType func(string) O
	isNilObject        func(O) bool
	loadUseData        func(O) U
	guideName          func(uint32) string
	storeGuide         func(U, string)
}

func rewardFieldGuideAddWeight4F0D20(total int32, weight uint8) int32 {
	return int32(uint32(total) + uint32(weight))
}

func rewardFieldGuideExplicit4F0D20[M, D, R, O, U any](
	data D,
	hooks rewardFieldGuideHooks4F0D20[M, D, R, O, U],
) (uint32, bool) {
	count := int32(0)
	for index := 0; index < rewardFieldGuideCount4F0D20; index++ {
		if hooks.loadExplicitGuide(data, index) == 1 {
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, count-1)
	ordinal := int32(0)
	for index := 0; index < rewardFieldGuideCount4F0D20; index++ {
		if hooks.loadExplicitGuide(data, index) == 1 {
			if ordinal == draw {
				return uint32(index), true
			}
			ordinal++
		}
	}
	return 0, false
}

func rewardFieldGuideAutomatic4F0D20[M, D, R, O, U any](
	stage uint32,
	hooks rewardFieldGuideHooks4F0D20[M, D, R, O, U],
) (uint32, bool) {
	slots := hooks.pickSlots(stage)
	if hooks.loadRowGuideID(hooks.rows, 0) == 0 {
		return 0, false
	}
	var total int32
	for index := 0; ; index++ {
		if slots&hooks.loadRowSlots(hooks.rows, index) != 0 {
			total = rewardFieldGuideAddWeight4F0D20(total, hooks.loadRowWeight(hooks.rows, index))
		}
		if hooks.loadRowGuideID(hooks.rows, index+1) == 0 {
			break
		}
	}
	if total == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, total-1)
	if hooks.loadRowGuideID(hooks.rows, 0) == 0 {
		return 0, false
	}
	var cumulative int32
	for index := 0; ; index++ {
		if slots&hooks.loadRowSlots(hooks.rows, index) != 0 {
			cumulative = rewardFieldGuideAddWeight4F0D20(cumulative, hooks.loadRowWeight(hooks.rows, index))
			if draw < cumulative {
				return hooks.loadRowGuideID(hooks.rows, index), true
			}
		}
		if hooks.loadRowGuideID(hooks.rows, index+1) == 0 {
			return 0, false
		}
	}
}

// rewardFieldGuide4F0D20 reconstructs GAME.EXE 004F0D20. InitData is cached
// at entry. Explicit guide bytes and automatic table fields are read in two
// live passes around their RNG callbacks. Automatic totals wrap as signed
// int32 values and use a signed draw/cumulative comparison. After successful
// object creation, use data is loaded before the unchecked guide-name lookup.
// There are deliberately no callback, data, row, name, or use-data guards
// beyond the original zero-result branches.
func rewardFieldGuide4F0D20[M, D, R, O, U any](
	marker M,
	stage uint32,
	hooks rewardFieldGuideHooks4F0D20[M, D, R, O, U],
) O {
	var zero O
	data := hooks.loadInitData(marker)
	var guideID uint32
	var selected bool
	if hooks.loadFlags(data)&rewardFieldGuideExplicitFlag4F0D20 != 0 {
		guideID, selected = rewardFieldGuideExplicit4F0D20(data, hooks)
	} else {
		guideID, selected = rewardFieldGuideAutomatic4F0D20(stage, hooks)
	}
	if !selected || guideID == 0 {
		return zero
	}
	result := hooks.createObjectByType(rewardFieldGuideType4F0D20)
	if hooks.isNilObject(result) {
		return zero
	}
	useData := hooks.loadUseData(result)
	name := hooks.guideName(guideID)
	hooks.storeGuide(useData, name)
	return result
}
