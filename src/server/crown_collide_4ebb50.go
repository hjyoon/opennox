package server

const (
	crownCollideBlockedFlags4EBB50 = uint32(0x00008020)
	crownCollidePlayerClass4EBB50  = uint8(0x04)
	crownCollidePickupFlag4EBB50   = int32(1)
)

type crownCollideResult4EBB50[O any] struct {
	target          O
	pickupResult    uint32
	pickupAttempted bool
}

type crownCollideHooks4EBB50[O any] struct {
	loadFlags    func(O) uint32
	loadClassLow func(O) uint8
	pickup       func(O, O, int32, int32) uint32
}

// crownCollide4EBB50 preserves GAME.EXE 004EBB50. Guard paths return the
// original target pointer unchanged. Only a non-nil Player target with neither
// blocked flag reaches CrownPickup, which receives both original pickup flags
// as one. The registered collision pointer is not read.
func crownCollide4EBB50[O comparable, C any](
	crown, target O,
	_ C,
	hooks crownCollideHooks4EBB50[O],
) crownCollideResult4EBB50[O] {
	result := crownCollideResult4EBB50[O]{target: target}
	var zero O
	if target == zero {
		return result
	}
	if hooks.loadFlags(target)&crownCollideBlockedFlags4EBB50 != 0 {
		return result
	}
	if hooks.loadClassLow(target)&crownCollidePlayerClass4EBB50 == 0 {
		return result
	}
	result.pickupAttempted = true
	result.pickupResult = hooks.pickup(
		target,
		crown,
		crownCollidePickupFlag4EBB50,
		crownCollidePickupFlag4EBB50,
	)
	return result
}
