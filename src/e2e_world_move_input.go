package opennox

import (
	"image"
	"math"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type e2eWorldMoveInputTarget struct {
	point          image.Point
	serverPos      image.Point
	serverDistance float64
	clientDistance float64
}

// MoveToWorld sends ordinary mouse input toward a requested world position.
// Arrival requires current server and client observations; collision, speed,
// camera, health and path selection remain owned by the normal game.
func (sc *e2eScenario) MoveToWorld(goal image.Point, walk bool, name string) {
	distance := e2eRunDist
	if walk {
		distance = e2eWalkDist
	}
	sc.moveToWorldInput(name, func() (e2eWorldMoveInputTarget, bool) {
		if noxServer == nil || noxClient == nil || noxClient.Viewport() == nil || noxClient.ClientPlayerUnit() == nil {
			return e2eWorldMoveInputTarget{}, false
		}
		unit := noxServer.Players.HostUnit()
		if unit == nil || unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			return e2eWorldMoveInputTarget{}, false
		}
		if unit.UpdateData == nil {
			return e2eWorldMoveInputTarget{}, false
		}
		state := unit.UpdateDataPlayer().State
		if state == server.PlayerState3 || state == server.PlayerState4 {
			return e2eWorldMoveInputTarget{}, false
		}
		clientPos := noxClient.ClientPlayerUnit().Pos()
		target := e2eWorldMoveInputTarget{
			serverPos:      image.Pt(int(unit.PosVec.X), int(unit.PosVec.Y)),
			serverDistance: math.Hypot(float64(goal.X)-float64(unit.PosVec.X), float64(goal.Y)-float64(unit.PosVec.Y)),
			clientDistance: math.Hypot(float64(goal.X-clientPos.X), float64(goal.Y-clientPos.Y)),
		}
		if math.IsNaN(target.serverDistance) || math.IsInf(target.serverDistance, 0) {
			return e2eWorldMoveInputTarget{}, false
		}
		if target.serverDistance <= 12 {
			return target, true
		}
		if target.clientDistance == 0 {
			return e2eWorldMoveInputTarget{}, false
		}
		vp := noxClient.Viewport()
		playerPoint, goalPoint := vp.ToScreenPos(clientPos), vp.ToScreenPos(goal)
		dx, dy := float64(goalPoint.X-playerPoint.X), float64(goalPoint.Y-playerPoint.Y)
		target.point = playerPoint.Add(image.Pt(int(float64(distance)*dx/target.clientDistance), int(float64(distance)*dy/target.clientDistance)))
		if !target.point.In(vp.Screen) {
			return e2eWorldMoveInputTarget{}, false
		}
		return target, true
	}, e2eQueueInput, func(target e2eWorldMoveInputTarget) {
		e2eLog.Printf("WORLD MOVE ACTUAL INPUT: goal=%v server=%v server_distance=%g client_distance=%g walk=%t frame=%d",
			goal, target.serverPos, target.serverDistance, target.clientDistance, walk, noxServer.Frame())
	})
}

func (sc *e2eScenario) moveToWorldInput(name string, read func() (e2eWorldMoveInputTarget, bool), queue func(...seat.InputEvent), arrived func(e2eWorldMoveInputTarget)) {
	var target e2eWorldMoveInputTarget
	moving := false
	stop := func() {
		if moving {
			queue(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
			moving = false
		}
	}
	sc.addWhen(0, name+" observed world arrival", 600, func() bool {
		current, ok := read()
		if !ok {
			stop()
			return false
		}
		if current.serverDistance <= 12 {
			stop()
			if current.clientDistance > 12 {
				return false
			}
			target = current
			return true
		}
		queue(&seat.MouseMoveEvent{Pos: current.point, Relative: false})
		if !moving {
			queue(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
			moving = true
		}
		return false
	}, func() {
		arrived(target)
	})
	sc.Wait(5, name+" settle actual movement release")
}
