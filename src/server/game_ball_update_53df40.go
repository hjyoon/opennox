package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const (
	gameBallUpdateDragBits53DF40        = uint32(1008981770)
	gameBallUpdateCarrierDead53DF40     = uint8(0x20)
	gameBallUpdateCarriedFlag53DF40     = uint32(0x40)
	gameBallUpdateBlockedFlags53DF40    = uint32(0x8020)
	gameBallUpdateTimeout53DF40         = uint64(20000)
	gameBallUpdateTraceFlags53DF40      = uint8(5)
	gameBallUpdateOwnerGap53DF40        = float64(10)
	gameBallUpdateReleaseOffset53DF40   = float64(20)
	gameBallUpdateReleaseForce53DF40    = float32(30)
	gameBallUpdateReleaseAudio53DF40    = uint32(926)
	gameBallUpdateReleasedState53DF40   = uint8(1)
	gameBallUpdateReleasedNetCode53DF40 = uint16(0)
)

type gameBallUpdateHooks53DF40[O comparable, D any] struct {
	loadUpdate             func(O) D
	storeDrag              func(O, uint32)
	loadCarrier            func(D) O
	loadFlagsLow           func(O) uint8
	carrierState           func(O, O)
	changeTeam             func(O)
	ballStatus             func(uint8, uint16)
	loadTicks              func() uint64
	loadUpdateTicks        func(D) uint64
	resetBall              func(O)
	loadOwner              func(O) O
	loadFrame              func() uint32
	loadCarrierFrame       func(D) uint32
	loadPossessionDuration func(D) uint32
	loadFlags              func(O) uint32
	storeFlags             func(O, uint32)
	storeUpdateTicks       func(D, uint64)
	clearOwner             func(O)
	loadRadius             func(O) float32
	loadPosX               func(O) float32
	loadPosY               func(O) float32
	loadDirection          func(O) int16
	loadDirectionCos       func(uint8) float32
	loadDirectionSin       func(uint8) float32
	trace                  func(types.Pointf, types.Pointf, uint8) bool
	move                   func(O, types.Pointf)
	storeObj130            func(O, O)
	randomInt              func(int32, int32) int32
	applyForce             func(O, types.Pointf, float32)
	audio                  func(uint32, O)
	loadVelocityX          func(O) float32
	loadVelocityY          func(O) float32
	loadResetVelocity      func(D) float32
	storeCarrier           func(D, O)
}

// gameBallUpdate53DF40 preserves GAME.EXE 0053DF40 while keeping every
// object reference native-width. The update record is intentionally cached
// once; owner and carrier fields are reloaded at the same mutation boundaries
// as the original routine.
func gameBallUpdate53DF40[O comparable, D any](
	ball O,
	hooks gameBallUpdateHooks53DF40[O, D],
) {
	update := hooks.loadUpdate(ball)
	hooks.storeDrag(ball, gameBallUpdateDragBits53DF40)

	var zero O
	carrier := hooks.loadCarrier(update)
	if carrier != zero && hooks.loadFlagsLow(carrier)&gameBallUpdateCarrierDead53DF40 != 0 {
		hooks.carrierState(ball, zero)
		hooks.changeTeam(ball)
		hooks.ballStatus(gameBallUpdateReleasedState53DF40, gameBallUpdateReleasedNetCode53DF40)
	}

	if hooks.loadTicks()-hooks.loadUpdateTicks(update) > gameBallUpdateTimeout53DF40 {
		hooks.resetBall(ball)
		return
	}

	owner := hooks.loadOwner(ball)
	if owner == zero {
		velocityY := hooks.loadVelocityY(ball)
		velocityX := hooks.loadVelocityX(ball)
		flags := hooks.loadFlags(ball)
		hooks.storeFlags(ball, flags&^gameBallUpdateCarriedFlag53DF40)
		speed := math.Sqrt(float64(velocityX)*float64(velocityX) + float64(velocityY)*float64(velocityY))
		if float64(hooks.loadResetVelocity(update)) > speed {
			hooks.storeCarrier(update, zero)
		}
		return
	}

	if owner != hooks.loadCarrier(update) ||
		hooks.loadFrame()-hooks.loadCarrierFrame(update) <= hooks.loadPossessionDuration(update) {
		flags := hooks.loadFlags(ball)
		hooks.storeFlags(ball, flags|gameBallUpdateCarriedFlag53DF40)
		hooks.storeUpdateTicks(update, hooks.loadTicks())

		owner = hooks.loadOwner(ball)
		if hooks.loadFlags(owner)&gameBallUpdateBlockedFlags53DF40 != 0 {
			hooks.clearOwner(ball)
			hooks.carrierState(ball, zero)
			hooks.changeTeam(ball)
			hooks.ballStatus(gameBallUpdateReleasedState53DF40, gameBallUpdateReleasedNetCode53DF40)
			return
		}

		ownerRadius := hooks.loadRadius(owner)
		startX := hooks.loadPosX(owner)
		ballRadius := hooks.loadRadius(ball)
		startY := hooks.loadPosY(owner)
		distance := float64(ownerRadius) + float64(ballRadius) + gameBallUpdateOwnerGap53DF40
		cosine := hooks.loadDirectionCos(uint8(hooks.loadDirection(owner)))
		destinationX := float32(distance*float64(cosine) + float64(hooks.loadPosX(owner)))
		sine := hooks.loadDirectionSin(uint8(hooks.loadDirection(owner)))
		destinationY := float32(distance*float64(sine) + float64(hooks.loadPosY(owner)))
		destination := types.Pointf{X: destinationX, Y: destinationY}
		if hooks.trace(types.Pointf{X: startX, Y: startY}, destination, gameBallUpdateTraceFlags53DF40) {
			hooks.move(ball, destination)
		}
		return
	}

	flags := hooks.loadFlags(ball)
	hooks.storeFlags(ball, flags&^gameBallUpdateCarriedFlag53DF40)
	hooks.storeObj130(ball, zero)
	releaseOwner := hooks.loadOwner(ball)
	direction := uint8(int32(hooks.loadDirection(releaseOwner)) + hooks.randomInt(-32, 32))
	cosine := hooks.loadDirectionCos(direction)
	sine := hooks.loadDirectionSin(direction)
	origin := types.Pointf{
		X: float32(float64(hooks.loadPosX(ball)) - float64(cosine)*gameBallUpdateReleaseOffset53DF40),
		Y: float32(float64(hooks.loadPosY(ball)) - float64(sine)*gameBallUpdateReleaseOffset53DF40),
	}
	hooks.applyForce(ball, origin, gameBallUpdateReleaseForce53DF40)
	hooks.clearOwner(ball)
	hooks.ballStatus(gameBallUpdateReleasedState53DF40, gameBallUpdateReleasedNetCode53DF40)
	hooks.audio(gameBallUpdateReleaseAudio53DF40, ball)
}
