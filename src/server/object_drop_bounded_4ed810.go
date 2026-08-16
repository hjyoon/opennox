package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const (
	objectDropBoundedMaxDistance4ED810   = float32(75)
	objectDropBoundedKOTRFlag4ED810      = uint32(16)
	objectDropBoundedRejectAudio4ED810   = uint32(925)
	objectDropBoundedRejectMessage4ED810 = "drop.c:DropNotAllowed"
)

// objectDropBoundedHooks4ED810 exposes the original argument, field, service,
// and shared-cache access order. The owner is cached first. The requested
// point and item remain delayed until their corresponding GAME.EXE loads.
type objectDropBoundedHooks4ED810[O, P any] struct {
	loadOwnerArg func() O
	loadOwnerX   func(O) float32
	loadOwnerY   func(O) float32
	loadPointArg func() P
	loadPointX   func(P) float32
	loadPointY   func(P) float32

	mapTrace        func(origin, target *types.Pointf) int32
	priorityMessage func(O, string, int32)
	loadNetCode     func(O) uint32
	audio           func(uint32, O, int32, uint32)

	gameFlag            func(uint32) int32
	loadItemArg         func() O
	loadCrownTypeCache  func() uint32
	lookupCrownType     func() uint32
	storeCrownTypeCache func(uint32)
	loadTypeIndex       func(O) uint16
	dispatch            func(O, O, *types.Pointf) int32
}

// objectDropBoundedDestination4ED810 reproduces the x87 calculation used by
// GAME.EXE 004ED810. Inputs are binary32. The Y delta and distance are spilled
// to binary32 before their later uses, while the X delta remains unspilled.
func objectDropBoundedDestination4ED810(origin, requested types.Pointf) types.Pointf {
	dx := float64(requested.X) - float64(origin.X)
	dyExtended := float64(requested.Y) - float64(origin.Y)
	dy := float32(dyExtended)
	dySquare := float64(dyExtended * float64(dy))
	dxSquare := float64(dx * dx)
	distanceExtended := math.Sqrt(float64(dySquare + dxSquare))
	distance := float32(distanceExtended)

	if distanceExtended > float64(objectDropBoundedMaxDistance4ED810) {
		requested.X = float32(dx*float64(objectDropBoundedMaxDistance4ED810)/float64(distance) + float64(origin.X))
		requested.Y = float32(float64(dy)*float64(objectDropBoundedMaxDistance4ED810)/float64(distance) + float64(origin.Y))
	}
	return requested
}

// objectDropBounded4ED810 preserves GAME.EXE 004ED810. MapTrace receives a
// local origin/destination pair and gates on the whole EAX value. A rejected
// trace reports the owner's live post-message NetCode. On success, the item is
// loaded only after the KOTR flag callback; KOTR Crown drops are suppressed via
// the shared 32-bit type cache, and every dispatched 32-bit result is kept.
func objectDropBounded4ED810[O, P any](hooks objectDropBoundedHooks4ED810[O, P]) int32 {
	owner := hooks.loadOwnerArg()
	origin := types.Pointf{
		X: hooks.loadOwnerX(owner),
		Y: hooks.loadOwnerY(owner),
	}
	point := hooks.loadPointArg()
	requested := types.Pointf{
		X: hooks.loadPointX(point),
		Y: hooks.loadPointY(point),
	}
	target := objectDropBoundedDestination4ED810(origin, requested)

	if hooks.mapTrace(&origin, &target) == 0 {
		hooks.priorityMessage(owner, objectDropBoundedRejectMessage4ED810, 0)
		code := hooks.loadNetCode(owner)
		hooks.audio(objectDropBoundedRejectAudio4ED810, owner, 2, code)
		return 0
	}

	kotr := hooks.gameFlag(objectDropBoundedKOTRFlag4ED810)
	item := hooks.loadItemArg()
	if kotr == 0 {
		return hooks.dispatch(owner, item, &target)
	}

	crownType := hooks.loadCrownTypeCache()
	if crownType == 0 {
		crownType = hooks.lookupCrownType()
		hooks.storeCrownTypeCache(crownType)
	}
	if uint32(hooks.loadTypeIndex(item)) == crownType {
		return 0
	}
	return hooks.dispatch(owner, item, &target)
}
