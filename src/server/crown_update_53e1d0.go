package server

import "github.com/opennox/libs/types"

const (
	crownUpdateBlockedFlags53E1D0 = uint32(0x00008020)
	crownUpdateDisabledFlag53E1D0 = uint8(0x20)
	crownUpdatePickupFlag53E1D0   = int32(1)
	crownUpdateTraceFlags53E1D0   = uint8(5)
	crownUpdateOwnerGap53E1D0     = float64(10)
)

type crownUpdateHooks53E1D0[O, D any] struct {
	loadUpdate       func(O) D
	loadPickupTarget func(D) O
	loadFlags        func(O) uint32
	pickup           func(O, O, int32, int32) uint32
	loadField0       func(D) O
	loadFlagsLow     func(O) uint8
	clearField0      func(D)
	loadOwner        func(O) O
	clearOwner       func(O)
	loadRadius       func(O) float32
	loadPosX         func(O) float32
	loadPosY         func(O) float32
	loadDirection    func(O) int16
	loadDirectionCos func(int16) float32
	loadDirectionSin func(int16) float32
	trace            func(types.Pointf, types.Pointf, uint8) bool
	move             func(O, types.Pointf)
}

// crownUpdate53E1D0 preserves GAME.EXE 0053E1D0. The update record is cached
// once. A live, unblocked PickupTarget is sent to CrownPickup and returns
// immediately. Otherwise the Crown follows its live owner, clearing disabled
// cached state and blocked ownership along the original branches.
func crownUpdate53E1D0[O comparable, D any](
	crown O,
	hooks crownUpdateHooks53E1D0[O, D],
) {
	update := hooks.loadUpdate(crown)
	target := hooks.loadPickupTarget(update)
	var zero O
	if target != zero && hooks.loadFlags(target)&crownUpdateBlockedFlags53E1D0 == 0 {
		hooks.pickup(
			target,
			crown,
			crownUpdatePickupFlag53E1D0,
			crownUpdatePickupFlag53E1D0,
		)
		return
	}

	field0 := hooks.loadField0(update)
	if field0 != zero && hooks.loadFlagsLow(field0)&crownUpdateDisabledFlag53E1D0 != 0 {
		hooks.clearField0(update)
	}

	owner := hooks.loadOwner(crown)
	if owner == zero {
		return
	}
	if hooks.loadFlags(owner)&crownUpdateBlockedFlags53E1D0 != 0 {
		hooks.clearOwner(crown)
		return
	}

	ownerRadius := hooks.loadRadius(owner)
	startX := hooks.loadPosX(owner)
	crownRadius := hooks.loadRadius(crown)
	startY := hooks.loadPosY(owner)
	distance := float64(ownerRadius) + float64(crownRadius) + crownUpdateOwnerGap53E1D0
	directionX := hooks.loadDirection(owner)
	cosine := hooks.loadDirectionCos(directionX)
	destinationX := float32(distance*float64(cosine) + float64(hooks.loadPosX(owner)))
	directionY := hooks.loadDirection(owner)
	sine := hooks.loadDirectionSin(directionY)
	destinationY := float32(distance*float64(sine) + float64(hooks.loadPosY(owner)))
	destination := types.Pointf{X: destinationX, Y: destinationY}
	if hooks.trace(
		types.Pointf{X: startX, Y: startY},
		destination,
		crownUpdateTraceFlags53E1D0,
	) {
		hooks.move(crown, destination)
	}
}
