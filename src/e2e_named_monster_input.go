package opennox

import (
	"image"
	"math"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2eNamedMonsterInputTarget struct {
	wire      uint16
	point     image.Point
	movePoint image.Point
	distance  float32
	hovered   bool
	dead      bool
	absent    bool
	health    uint16
	playerHP  uint16
}

// StrikeNamedMonster observes the current world/drawable/viewport and sends
// ordinary mouse movement and button events. The normal game moves the player,
// chooses the cursor target, performs one sword action, and decides its result.
// No object/drawable pointer is retained between input frames.
func (sc *e2eScenario) StrikeNamedMonster(id string, offset image.Point, name string) {
	sc.strikeNamedMonsterInput(name, func() (e2eNamedMonsterInputTarget, bool) {
		if noxServer == nil {
			return e2eNamedMonsterInputTarget{}, false
		}
		return e2eReadMonsterInputTarget(noxServer.Objs.GetObjectByID(id), offset, false)
	}, e2eQueueInput, func(target e2eNamedMonsterInputTarget) {
		e2eLog.Printf("NAMED MONSTER ACTUAL INPUT: id=%q wire=%#x point=%v distance=%g observed_dead=%t frame=%d",
			id, target.wire, target.point, target.distance, target.dead, noxServer.Frame())
	})
}

// Native pointers are used only during this observation. Airborne targets use
// their current sprite height for aiming and their ground point for movement.
func e2eReadMonsterInputTarget(npc *server.Object, offset image.Point, followGround bool) (e2eNamedMonsterInputTarget, bool) {
	if noxServer == nil || noxClient == nil || noxClient.Viewport() == nil || noxClient.ClientPlayerUnit() == nil {
		return e2eNamedMonsterInputTarget{}, false
	}
	unit := noxServer.Players.HostUnit()
	if unit == nil || unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		return e2eNamedMonsterInputTarget{}, false
	}
	if npc == nil || !npc.Class().Has(object.ClassMonster) || npc.HealthData == nil {
		return e2eNamedMonsterInputTarget{}, false
	}
	if npc.HealthData.Cur == 0 && npc.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		return e2eNamedMonsterInputTarget{dead: true, playerHP: unit.HealthData.Cur}, true
	}
	if npc.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || !npc.Flags().Has(object.FlagEnabled) {
		return e2eNamedMonsterInputTarget{}, false
	}
	code := noxServer.GetUnitNetCode(npc)
	if code <= 0 || code > int(^uint16(0)) {
		return e2eNamedMonsterInputTarget{}, false
	}
	dr := noxClient.Objs.ByNetCode(uint16(code))
	if dr == nil {
		return e2eNamedMonsterInputTarget{}, false
	}
	vp := noxClient.Viewport()
	height := 0
	if followGround {
		height = dr.Z() + int(int16(dr.ZVal2))
	}
	point, movePoint := e2eMonsterMousePoints(vp.ToScreenPos(noxClient.ClientPlayerUnit().Pos()), vp.ToScreenPos(dr.Pos()), offset, height, followGround)
	if !point.In(vp.Screen) || !movePoint.In(vp.Screen) {
		return e2eNamedMonsterInputTarget{}, false
	}
	hovered := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == dr
	if hovered {
		point = noxClient.Inp.GetMousePos()
	}
	wx, wy := unit.PosVec.X-npc.PosVec.X, unit.PosVec.Y-npc.PosVec.Y
	return e2eNamedMonsterInputTarget{wire: uint16(code), point: point, movePoint: movePoint,
		distance: float32(math.Hypot(float64(wx), float64(wy))), hovered: hovered,
		health: npc.HealthData.Cur, playerHP: unit.HealthData.Cur}, true
}

func e2eMonsterMousePoints(player, ground, offset image.Point, height int, followGround bool) (point, movePoint image.Point) {
	point = ground.Add(offset).Sub(image.Pt(0, height))
	movePoint = point
	if followGround {
		movePoint = ground
	}
	dx, dy := float64(movePoint.X-player.X), float64(movePoint.Y-player.Y)
	if length := math.Hypot(dx, dy); length > 0 {
		// Ordinary movement toward a point beyond the creature preserves running
		// while the original collision and speed rules determine the approach.
		movePoint = player.Add(image.Pt(int(200*dx/length), int(200*dy/length)))
	}
	return point, movePoint
}

func (sc *e2eScenario) strikeNamedMonsterInput(name string, read func() (e2eNamedMonsterInputTarget, bool), queue func(...seat.InputEvent), clicked func(e2eNamedMonsterInputTarget)) {
	var target e2eNamedMonsterInputTarget
	var moving, pressed bool
	stopMoving := func() {
		if moving {
			queue(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
			moving = false
		}
	}
	sc.addWhen(0, name+" actual approach and cursor", 600, func() bool {
		current, ok := read()
		if !ok {
			stopMoving()
			return false
		}
		if current.dead || current.absent {
			stopMoving()
			target = current
			return true
		}
		if current.distance > 28 {
			queue(&seat.MouseMoveEvent{Pos: current.movePoint, Relative: false})
			if !moving {
				queue(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
				moving = true
			}
			return false
		}
		stopMoving()
		if !current.hovered {
			queue(&seat.MouseMoveEvent{Pos: current.point, Relative: false})
			return false
		}
		target = current
		return true
	}, func() {
		if !target.dead && !target.absent {
			queue(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
			pressed = true
		}
		clicked(target)
	})
	sc.add(1, name+" release actual sword input", func() {
		if pressed {
			queue(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		}
	})
}
