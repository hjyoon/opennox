package server

const (
	foodDropCoopFlag4EDE50 = uint32(0x00000800)
	foodDropSeconds4EDE50  = uint32(25)
)

type foodDropSoundRule4EDE50 struct {
	subClassMask uint32
	flagsLowMask uint16
	sound        uint16
}

var foodDropSoundRules4EDE50 = [...]foodDropSoundRule4EDE50{
	{subClassMask: 0x00000000, flagsLowMask: 0x0001, sound: 835},
	{subClassMask: 0x00000002, flagsLowMask: 0x0000, sound: 837},
	{subClassMask: 0x00000004, flagsLowMask: 0x0000, sound: 833},
	{subClassMask: 0x00000080, flagsLowMask: 0x0000, sound: 839},
	{subClassMask: 0x00000000, flagsLowMask: 0x0000, sound: 0},
}

// foodDropHooks4EDE50 exposes GAME.EXE 004EDE50's exact argument, callback,
// object-field, and sound-table read order. Sound is deliberately loaded once
// to test the row/sentinel and again after a match, as in the original x86.
type foodDropHooks4EDE50[O, P comparable] struct {
	loadOwnerArg func() O
	loadFoodArg  func() O
	loadPointArg func() P

	defaultDrop func(O, O, P) int32
	gameFlag    func(uint32) int32
	loadGameFPS func() uint32
	setDecay    func(O, uint32)

	loadRuleSound        func(int) uint16
	loadSubClass         func(O) uint32
	loadRuleSubClassMask func(int) uint32
	loadRuleFlagsLowMask func(int) uint16
	loadFlagsLow         func(O) uint16
	audio                func(uint32, O, int32, uint32)
}

// foodDrop4EDE50 preserves GAME.EXE 004EDE50. Nil arguments return canonical
// zero before DefaultDrop. A nonzero DefaultDrop result, including a
// noncanonical value, is cached and returned exactly. The x86 LEA sequence
// computes 25*FPS modulo 2^32.
func foodDrop4EDE50[O, P comparable](hooks foodDropHooks4EDE50[O, P]) int32 {
	var nilObject O
	owner := hooks.loadOwnerArg()
	if owner == nilObject {
		return 0
	}

	food := hooks.loadFoodArg()
	if food == nilObject {
		return 0
	}

	var nilPoint P
	point := hooks.loadPointArg()
	if point == nilPoint {
		return 0
	}

	result := hooks.defaultDrop(owner, food, point)
	if result == 0 {
		return result
	}

	if hooks.gameFlag(foodDropCoopFlag4EDE50) == 0 {
		hooks.setDecay(food, hooks.loadGameFPS()*foodDropSeconds4EDE50)
	}

	for row := 0; ; row++ {
		if hooks.loadRuleSound(row) == 0 {
			return result
		}

		subClass := hooks.loadSubClass(food)
		if subClass&hooks.loadRuleSubClassMask(row) != 0 {
			sound := hooks.loadRuleSound(row)
			hooks.audio(uint32(sound), owner, 0, 0)
			return result
		}

		flagsLowMask := hooks.loadRuleFlagsLowMask(row)
		if flagsLowMask&hooks.loadFlagsLow(food) != 0 {
			sound := hooks.loadRuleSound(row)
			hooks.audio(uint32(sound), owner, 0, 0)
			return result
		}
	}
}
