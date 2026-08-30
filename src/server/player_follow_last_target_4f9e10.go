package server

const (
	playerFollowRejectedFlags4F9E10 = uint8(0x20)
	playerFollowMonsterClass4F9E10  = uint8(0x02)
	playerFollowPlayerClass4F9E10   = uint8(0x04)
	playerFollowObserverBit4F9E10   = uint32(0x01)
)

// playerFollowLastTargetHooks4F9E10 exposes every observable load and call in
// GAME.EXE 004F9E10 without assuming the width or layout of an Object pointer.
type playerFollowLastTargetHooks4F9E10[O comparable, U, P any] struct {
	loadLastTarget       func(O) O
	findOwnerChainPlayer func(O) O
	loadFlagsByte        func(O) uint8
	loadClassByte        func(O) uint8
	loadUpdateData       func(O) U
	loadPlayer           func(U) P
	loadPlayerStatus     func(P) uint32
	cameraFollow         func(O, O)
}

// playerFollowLastTarget4F9E10 preserves GAME.EXE 004F9E10. It follows the
// player's Obj130 attribution link to its terminal/player owner, rejects
// destroyed, Monster, and observing Player targets, and otherwise delegates
// to the original camera-toggle operation. The conditional update/player
// dereferences deliberately remain unguarded, matching the original fault
// contract for malformed Player objects.
func playerFollowLastTarget4F9E10[O comparable, U, P any](
	unit O,
	hooks playerFollowLastTargetHooks4F9E10[O, U, P],
) int32 {
	var zero O
	if unit == zero {
		return 0
	}
	target := hooks.loadLastTarget(unit)
	if target == zero {
		return 0
	}
	target = hooks.findOwnerChainPlayer(target)
	if hooks.loadFlagsByte(target)&playerFollowRejectedFlags4F9E10 != 0 {
		return 0
	}
	class := hooks.loadClassByte(target)
	if class&playerFollowMonsterClass4F9E10 != 0 {
		return 0
	}
	if class&playerFollowPlayerClass4F9E10 != 0 {
		update := hooks.loadUpdateData(target)
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerStatus(player)&playerFollowObserverBit4F9E10 != 0 {
			return 0
		}
	}
	hooks.cameraFollow(unit, target)
	return 1
}
