package server

const (
	pickupFoodUseBypassSubClassLow4F3350 = uint8(0x84)
	pickupFoodDestroyedFlagsLow4F3350    = uint8(0x20)
)

type pickupFoodSoundRule4F3350 struct {
	subClassMask uint32
	materialMask uint16
	sound        uint16
}

var pickupFoodSoundRules4F3350 = [...]pickupFoodSoundRule4F3350{
	{subClassMask: 0x00000000, materialMask: 0x0001, sound: 834},
	{subClassMask: 0x00000002, materialMask: 0x0000, sound: 836},
	{subClassMask: 0x00000004, materialMask: 0x0000, sound: 832},
	{subClassMask: 0x00000080, materialMask: 0x0000, sound: 838},
	{subClassMask: 0x00000000, materialMask: 0x0000, sound: 0},
}

// pickupFoodHooks4F3350 exposes every observable callback, field load, and
// sound-table load in GAME.EXE 004F3350. A rule sound is deliberately loaded
// once as the current/next-row sentinel and again after a match, matching the
// original x86 rather than treating the table as an immutable Go value.
type pickupFoodHooks4F3350[O comparable, U any] struct {
	playerState     func(O) int32
	loadSubClassLow func(O) uint8
	loadUse         func(O) U
	callUse         func(U, O, O) int32
	loadFlagsLow    func(O) uint8
	defaultPickup   func(O, O, int32, int32) int32

	loadRuleSound        func(int) uint16
	loadSubClass         func(O) uint32
	loadRuleSubClassMask func(int) uint32
	loadRuleMaterialMask func(int) uint16
	loadMaterialLow      func(O) uint16
	audio                func(uint32, O, int32, uint32)
}

// pickupFood4F3350 preserves GAME.EXE 004F3350. Null owner and item pointers
// return before any callback or field access. Outside the special player
// state, every food except Jug and Mushroom calls its live Use callback with
// (owner, item); there is no nil-callback guard and the callback result is
// ignored. A live Destroyed low flag then returns canonical one. Otherwise the
// original four callback arguments are forwarded to DefaultPickup and its
// full int32 result is cached and returned exactly. Only a nonzero result scans
// the live sound rules, testing subclass before material and targeting audio
// at the owner.
func pickupFood4F3350[O comparable, U any](
	owner, item O,
	arg3, arg4 int32,
	hooks pickupFoodHooks4F3350[O, U],
) int32 {
	var nilObject O
	if owner == nilObject {
		return 0
	}
	if item == nilObject {
		return 0
	}

	if hooks.playerState(owner) == 0 {
		if hooks.loadSubClassLow(item)&pickupFoodUseBypassSubClassLow4F3350 == 0 {
			use := hooks.loadUse(item)
			_ = hooks.callUse(use, owner, item)
		}
	}

	if hooks.loadFlagsLow(item)&pickupFoodDestroyedFlagsLow4F3350 != 0 {
		return 1
	}

	result := hooks.defaultPickup(owner, item, arg3, arg4)
	if result == 0 {
		return result
	}

	row := 0
	if hooks.loadRuleSound(row) == 0 {
		return result
	}
	for {
		subClass := hooks.loadSubClass(item)
		if subClass&hooks.loadRuleSubClassMask(row) != 0 {
			sound := hooks.loadRuleSound(row)
			hooks.audio(uint32(sound), owner, 0, 0)
			return result
		}

		materialMask := hooks.loadRuleMaterialMask(row)
		if materialMask&hooks.loadMaterialLow(item) != 0 {
			sound := hooks.loadRuleSound(row)
			hooks.audio(uint32(sound), owner, 0, 0)
			return result
		}

		row++
		if hooks.loadRuleSound(row) == 0 {
			return result
		}
	}
}
