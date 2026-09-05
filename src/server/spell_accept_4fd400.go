package server

const (
	spellAcceptTargetFlag4FD400 = uint32(0x80)
	spellAcceptUnitMask4FD400   = uint8(0x06)
	spellAcceptFizzle4FD400     = int32(231)
)

type spellAcceptDispatch4FD400 uint8

const (
	spellAcceptDefault4FD400 spellAcceptDispatch4FD400 = iota
	spellAcceptInstant4FD400
	spellAcceptDurationBlink4FD400
	spellAcceptDurationChannel4FD400
	spellAcceptDurationCharm4FD400
	spellAcceptDurationTurnUndead4FD400
	spellAcceptDurationDrainMana4FD400
	spellAcceptDurationLightning4FD400
	spellAcceptDurationFirewalk4FD400
	spellAcceptDurationForceNature4FD400
	spellAcceptDurationGreaterHeal4FD400
	spellAcceptDurationChainLightning4FD400
	spellAcceptDurationShield4FD400
	spellAcceptDurationMoonglow4FD400
	spellAcceptDurationManaBomb4FD400
	spellAcceptDurationPlasma4FD400
	spellAcceptDurationOvalShield4FD400
	spellAcceptDurationSummon4FD400
	spellAcceptDurationSwap4FD400
	spellAcceptDurationTag4FD400
	spellAcceptDurationTeleportMark4FD400
	spellAcceptDurationTeleportPop4FD400
	spellAcceptDurationTeleportTarget4FD400
	spellAcceptDurationWall4FD400
)

type spellAcceptHooks4FD400[Object comparable, AcceptArg comparable] struct {
	loadSpellArg  func() int32
	loadThirdArg  func() Object
	loadSecondArg func() Object
	loadAcceptArg func() AcceptArg

	spellHasFlags func(int32, uint32) int32
	loadTarget    func(AcceptArg) Object
	loadClassLow  func(Object) uint8
	captureMagic  func(int32, Object) int32
	audio         func(int32, Object, int32, int32)

	loadLevelArg  func() int32
	loadFourthArg func() Object
	tickRate      func() uint32
	plasmaTime    func() float64
	instant       func(int32, Object, Object, Object, AcceptArg, int32) int32
	duration      func(spellAcceptDispatch4FD400, int32, Object, Object, Object, AcceptArg, int32, uint32) int32
}

// spellAcceptDispatchFor4FD400 is the decoded 133-byte selector table at
// GAME.EXE 004FDB80 expressed as semantic routes. It deliberately dispatches
// by the signed numeric spell ID, not by a mutable spell definition's Effect.
func spellAcceptDispatchFor4FD400(spellID int32) spellAcceptDispatch4FD400 {
	switch spellID {
	case 4:
		return spellAcceptDurationBlink4FD400
	case 8:
		return spellAcceptDurationChannel4FD400
	case 9:
		return spellAcceptDurationCharm4FD400
	case 21:
		return spellAcceptDurationTurnUndead4FD400
	case 22:
		return spellAcceptDurationDrainMana4FD400
	case 24:
		return spellAcceptDurationLightning4FD400
	case 28:
		return spellAcceptDurationFirewalk4FD400
	case 31:
		return spellAcceptDurationForceNature4FD400
	case 35:
		return spellAcceptDurationGreaterHeal4FD400
	case 43:
		return spellAcceptDurationChainLightning4FD400
	case 51:
		return spellAcceptDurationShield4FD400
	case 54:
		return spellAcceptDurationMoonglow4FD400
	case 56:
		return spellAcceptDurationManaBomb4FD400
	case 59:
		return spellAcceptDurationPlasma4FD400
	case 67:
		return spellAcceptDurationOvalShield4FD400
	case 75, 76, 77, 78,
		80, 81, 82, 83, 84, 85, 86, 87, 88, 89,
		90, 91, 92, 93, 94, 95, 96, 97, 98, 99,
		100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
		110, 111, 112, 113, 114:
		return spellAcceptDurationSummon4FD400
	case 115:
		return spellAcceptDurationSwap4FD400
	case 116:
		return spellAcceptDurationTag4FD400
	case 117, 118, 119, 120, 122, 123, 124, 125:
		return spellAcceptDurationTeleportMark4FD400
	case 121:
		return spellAcceptDurationTeleportPop4FD400
	case 126:
		return spellAcceptDurationTeleportTarget4FD400
	case 132:
		return spellAcceptDurationWall4FD400
	case 1, 2, 3, 5, 6,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		23, 26, 27, 29,
		32, 33, 34, 36, 37, 38, 39, 40, 41, 42,
		44, 45, 46, 47, 48, 49, 50, 52, 53, 55, 57, 58,
		60, 61, 62, 63, 64, 65, 66, 68, 69, 70, 71, 72, 74,
		127, 128, 129, 130, 131, 133:
		return spellAcceptInstant4FD400
	default:
		return spellAcceptDefault4FD400
	}
}

func spellAcceptDuration4FD400[Object comparable, AcceptArg comparable](
	dispatch spellAcceptDispatch4FD400,
	spellID int32,
	second, third Object,
	arg AcceptArg,
	hooks spellAcceptHooks4FD400[Object, AcceptArg],
) int32 {
	var timeout uint32
	var fourth Object
	var level int32

	switch dispatch {
	case spellAcceptDurationFirewalk4FD400:
		// LEA performs the multiply in one wrapping dword before the live
		// fourth-object and level loads.
		timeout = hooks.tickRate() * 3
		fourth = hooks.loadFourthArg()
		level = hooks.loadLevelArg()
	case spellAcceptDurationForceNature4FD400:
		// GAME.EXE wraps the doubled dword and then performs unsigned /3.
		timeout = (hooks.tickRate() * 2) / 3
		fourth = hooks.loadFourthArg()
		level = hooks.loadLevelArg()
	case spellAcceptDurationPlasma4FD400:
		// 00566DCC truncates an x87 value to a signed qword; only EAX is
		// forwarded as the duration dword.
		timeout = uint32(x87TruncSignedQwordLow566DCC(hooks.plasmaTime()))
		fourth = hooks.loadFourthArg()
		level = hooks.loadLevelArg()
	default:
		level = hooks.loadLevelArg()
		fourth = hooks.loadFourthArg()
	}

	thirdForDuration := third
	switch dispatch {
	case spellAcceptDurationShield4FD400,
		spellAcceptDurationMoonglow4FD400,
		spellAcceptDurationOvalShield4FD400:
		// These three branches reload *arg after the fourth-object load.
		thirdForDuration = hooks.loadTarget(arg)
	}

	return hooks.duration(dispatch, spellID, second, thirdForDuration, fourth, arg, level, timeout)
}

// spellAccept4FD400 preserves GAME.EXE 004FD400's guard, reload, callback and
// return-value contract. The spell, second object, third object and argument
// pointer are cached at their original entry loads. The target stored in the
// argument remains live: it is independently reloaded for the class gate,
// capture gate, capture-failure sound and the three target-based duration
// routes. The fourth object and level are delayed until the selected cast.
//
// Spell flags must return exactly one to activate the 0x80 target-class gate.
// Instant and duration results are returned as unmodified signed dwords. Only
// a zero instant result emits sound 231, using the fourth object cached for
// that callback. No level clamp or spell-definition lookup exists here.
func spellAccept4FD400[Object comparable, AcceptArg comparable](
	hooks spellAcceptHooks4FD400[Object, AcceptArg],
) int32 {
	var zeroObject Object
	var zeroArg AcceptArg

	spellID := hooks.loadSpellArg()
	if spellID == 0 {
		return 0
	}
	third := hooks.loadThirdArg()
	if third == zeroObject {
		return 0
	}
	second := hooks.loadSecondArg()
	if second == zeroObject {
		return 0
	}
	arg := hooks.loadAcceptArg()
	if arg == zeroArg {
		return 0
	}

	if hooks.spellHasFlags(spellID, spellAcceptTargetFlag4FD400) == 1 {
		target := hooks.loadTarget(arg)
		if target != zeroObject && hooks.loadClassLow(target)&spellAcceptUnitMask4FD400 == 0 {
			return 0
		}
	}

	target := hooks.loadTarget(arg)
	if target != zeroObject && hooks.captureMagic(spellID, target) == 0 {
		hooks.audio(spellAcceptFizzle4FD400, hooks.loadTarget(arg), 0, 0)
		return 0
	}

	dispatch := spellAcceptDispatchFor4FD400(spellID)
	switch dispatch {
	case spellAcceptDefault4FD400:
		return 1
	case spellAcceptInstant4FD400:
		level := hooks.loadLevelArg()
		fourth := hooks.loadFourthArg()
		result := hooks.instant(spellID, second, third, fourth, arg, level)
		if result == 0 {
			hooks.audio(spellAcceptFizzle4FD400, fourth, 0, 0)
		}
		return result
	default:
		return spellAcceptDuration4FD400(dispatch, spellID, second, third, arg, hooks)
	}
}
