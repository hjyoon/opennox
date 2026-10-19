package opennox

import (
	"image"
	"math"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy"
)

type e2eNamedMonsterInputTarget struct {
	wire      uint16
	point     image.Point
	movePoint image.Point
	distance  float32
	hovered   bool
	dead      bool
}

// StrikeNamedMonster observes the current world/drawable/viewport and sends
// ordinary mouse movement and button events. The normal game moves the player,
// chooses the cursor target, performs one sword action, and decides its result.
// No object/drawable pointer is retained between input frames.
func (sc *e2eScenario) StrikeNamedMonster(id string, offset image.Point, name string) {
	sc.strikeNamedMonsterInput(name, func() (e2eNamedMonsterInputTarget, bool) {
		if noxServer == nil || noxClient == nil || noxClient.Viewport() == nil || noxClient.ClientPlayerUnit() == nil {
			return e2eNamedMonsterInputTarget{}, false
		}
		unit := noxServer.Players.HostUnit()
		if unit == nil || unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			return e2eNamedMonsterInputTarget{}, false
		}
		npc := noxServer.Objs.GetObjectByID(id)
		if npc == nil || !npc.Class().Has(object.ClassMonster) || npc.HealthData == nil {
			return e2eNamedMonsterInputTarget{}, false
		}
		if npc.HealthData.Cur == 0 && npc.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			return e2eNamedMonsterInputTarget{dead: true}, true
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
		point := vp.ToScreenPos(dr.Pos()).Add(offset)
		if !point.In(vp.Screen) {
			return e2eNamedMonsterInputTarget{}, false
		}
		playerPoint := vp.ToScreenPos(noxClient.ClientPlayerUnit().Pos())
		dx, dy := float64(point.X-playerPoint.X), float64(point.Y-playerPoint.Y)
		movePoint := point
		if length := math.Hypot(dx, dy); length > 0 {
			// A genuine mouse position beyond the target keeps running available
			// while approaching a fleeing creature. World movement still follows
			// the ordinary right-button path and all original collisions.
			movePoint = playerPoint.Add(image.Pt(int(200*dx/length), int(200*dy/length)))
		}
		if !movePoint.In(vp.Screen) {
			return e2eNamedMonsterInputTarget{}, false
		}
		hovered := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == dr
		if hovered {
			point = noxClient.Inp.GetMousePos()
		}
		wx, wy := unit.PosVec.X-npc.PosVec.X, unit.PosVec.Y-npc.PosVec.Y
		return e2eNamedMonsterInputTarget{wire: uint16(code), point: point, movePoint: movePoint,
			distance: float32(math.Hypot(float64(wx), float64(wy))), hovered: hovered}, true
	}, e2eQueueInput, func(target e2eNamedMonsterInputTarget) {
		e2eLog.Printf("NAMED MONSTER ACTUAL INPUT: id=%q wire=%#x point=%v distance=%g observed_dead=%t frame=%d",
			id, target.wire, target.point, target.distance, target.dead, noxServer.Frame())
	})
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
		if current.dead {
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
		if !target.dead {
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
