package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const (
	moverUpdateTargetRequiredFlag54F740 = uint32(0x00000004)
	moverUpdateTargetBlockedFlag54F740  = uint32(0x00000020)
	moverUpdateActiveFlag54F740         = uint32(0x01000000)
	moverUpdateSpeedScale54F740         = float64(0.25)
	moverUpdateSteerEpsilon54F740       = float32(0.1)
)

type moverUpdateHooks54F740[O, D, W comparable] struct {
	loadUpdateData   func(O) D
	loadTargetExtent func(D) uint32
	loadTarget       func(D) O
	storeTarget      func(D, O)
	objectByExtent   func(uint32) O
	loadWaypointID   func(D, int) uint32
	loadWaypoint     func(D, int) W
	storeWaypoint    func(D, int, W)
	waypointByID     func(uint32) W
	loadFlags        func(O) uint32
	loadState        func(D) uint8
	storeState       func(D, uint8)
	loadSpeedUnits   func(D) int32
	storeSpeedBase   func(O, float32)
	storeSpeedCur    func(O, float32)
	loadSpeedCur     func(O) float32
	loadPosition     func(O) types.Pointf
	loadPosX         func(O) float32
	loadPosY         func(O) float32
	loadVelocityX    func(O) float32
	loadVelocityY    func(O) float32
	storeVelocityX   func(O, float32)
	storeVelocityY   func(O, float32)
	loadWaypointPos  func(W) types.Pointf
	loadWaypointX    func(W) float32
	loadWaypointY    func(W) float32
	waypointPointCnt func(W) uint8
	waypointPoint    func(W, int) W
	randomInt        func(int, int) int
	move             func(O, types.Pointf)
	removeUpdatable  func(O)
}

func moverUpdateNonzero54F740(value float32) bool {
	return value != 0 && !math.IsNaN(float64(value))
}

// moverUpdate54F740 preserves GAME.EXE 0054F740 through its four-entry jump
// table at 0054F990. The 36-byte update record is cached once, while the
// original transient target and waypoint slots are deliberately reloaded at
// the same state-machine boundaries. This matters when Move or RNG callbacks
// mutate the live record.
func moverUpdate54F740[O, D, W comparable](source O, hooks moverUpdateHooks54F740[O, D, W]) {
	data := hooks.loadUpdateData(source)
	extent := hooks.loadTargetExtent(data)
	var zeroObject O
	var zeroWaypoint W
	if extent == 0 {
		hooks.removeUpdatable(source)
		return
	}

	target := hooks.loadTarget(data)
	if target == zeroObject {
		target = hooks.objectByExtent(extent)
		hooks.storeTarget(data, target)
		if target == zeroObject {
			hooks.removeUpdatable(source)
			return
		}
	}

	waypointID := hooks.loadWaypointID(data, 3)
	if waypointID != 0 && hooks.loadWaypoint(data, 3) == zeroWaypoint {
		hooks.storeWaypoint(data, 3, hooks.waypointByID(waypointID))
	}
	waypointID = hooks.loadWaypointID(data, 5)
	if waypointID != 0 && hooks.loadWaypoint(data, 5) == zeroWaypoint {
		hooks.storeWaypoint(data, 5, hooks.waypointByID(waypointID))
	}

	targetFlags := hooks.loadFlags(target)
	if targetFlags&moverUpdateTargetRequiredFlag54F740 == 0 ||
		targetFlags&moverUpdateTargetBlockedFlag54F740 != 0 {
		hooks.storeTarget(data, zeroObject)
		hooks.removeUpdatable(source)
		return
	}

	switch hooks.loadState(data) {
	case 0:
		if hooks.loadFlags(source)&moverUpdateActiveFlag54F740 == 0 {
			return
		}
		waypoint := hooks.waypointByID(uint32(hooks.loadWaypointID(data, 2)))
		if waypoint == zeroWaypoint {
			hooks.storeState(data, 3)
			return
		}
		hooks.move(source, hooks.loadPosition(hooks.loadTarget(data)))
		speedUnits := hooks.loadSpeedUnits(data)
		hooks.storeState(data, 1)
		hooks.storeWaypoint(data, 3, waypoint)
		hooks.storeWaypoint(data, 5, zeroWaypoint)
		speed := float32(float64(speedUnits) * moverUpdateSpeedScale54F740)
		hooks.storeSpeedBase(source, speed)
		hooks.storeSpeedCur(source, speed)
		return

	case 1:
		if hooks.loadFlags(source)&moverUpdateActiveFlag54F740 == 0 {
			hooks.storeState(data, 2)
			return
		}

		velocityX := hooks.loadVelocityX(source)
		if moverUpdateNonzero54F740(velocityX) {
			velocityY := hooks.loadVelocityY(source)
			if moverUpdateNonzero54F740(velocityY) {
				current := hooks.loadWaypoint(data, 3)
				dx := float64(hooks.loadWaypointX(current)) - float64(hooks.loadPosX(source))
				dy := float64(hooks.loadWaypointY(current)) - float64(hooks.loadPosY(source))
				dot := dy*float64(hooks.loadVelocityY(source)) +
					dx*float64(hooks.loadVelocityX(source))
				if dot <= 0 || math.IsNaN(dot) {
					count := hooks.waypointPointCnt(current)
					switch count {
					case 0:
						hooks.storeState(data, 3)
						hooks.move(target, hooks.loadWaypointPos(current))
					case 1:
						hooks.storeWaypoint(data, 5, current)
						hooks.storeWaypoint(data, 3, hooks.waypointPoint(current, 0))
					default:
						var index int
						for {
							index = hooks.randomInt(0, int(hooks.waypointPointCnt(current))-1)
							current = hooks.loadWaypoint(data, 3)
							previous := hooks.loadWaypoint(data, 5)
							if hooks.waypointPoint(current, index) != previous {
								break
							}
						}
						hooks.storeWaypoint(data, 5, current)
						hooks.storeWaypoint(data, 3, hooks.waypointPoint(current, index))
					}
				}
			}
		}

		if hooks.loadState(data) != 1 {
			return
		}
		hooks.move(hooks.loadTarget(data), hooks.loadPosition(source))
		waypoint := hooks.loadWaypoint(data, 3)
		dx := float64(hooks.loadWaypointX(waypoint)) - float64(hooks.loadPosX(source))
		dyExtended := float64(hooks.loadWaypointY(waypoint)) - float64(hooks.loadPosY(source))
		dy := float32(dyExtended)
		denominator := float32(math.Sqrt(dyExtended*float64(dy)+dx*dx) +
			float64(moverUpdateSteerEpsilon54F740))
		hooks.storeVelocityX(
			source,
			float32(dx*float64(hooks.loadSpeedCur(source))/float64(denominator)),
		)
		hooks.storeVelocityY(
			source,
			float32(float64(dy)*float64(hooks.loadSpeedCur(source))/float64(denominator)),
		)
		return

	case 2:
		if hooks.loadFlags(source)&moverUpdateActiveFlag54F740 != 0 {
			hooks.move(source, hooks.loadPosition(target))
			hooks.storeState(data, 1)
		}
		return

	case 3:
		hooks.removeUpdatable(source)
		return

	default:
		return
	}
}
