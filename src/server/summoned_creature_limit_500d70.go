package server

const summonedCreatureLimit500D70 = int32(4)

// summonedCreatureLimitHooks500D70 exposes the two calls made by GAME.EXE
// 00500D70. Object handles remain native-width values while the guide index
// and controlled-creature result preserve their original signed dword ABI.
type summonedCreatureLimitHooks500D70[O comparable] struct {
	loadGuideSize   func(int32) int32
	countControlled func(O) int32
}

// summonedCreatureLimitCheck500D70 preserves the original call order and
// arithmetic: the guide result is narrowed to its low byte, addition wraps as
// a PE32 dword, and that same bit pattern is compared as a signed int32.
func summonedCreatureLimitCheck500D70[O comparable](
	owner O,
	guideIndex int32,
	hooks summonedCreatureLimitHooks500D70[O],
) bool {
	size := uint32(uint8(hooks.loadGuideSize(guideIndex)))
	count := uint32(hooks.countControlled(owner))
	return int32(count+size) <= summonedCreatureLimit500D70
}
