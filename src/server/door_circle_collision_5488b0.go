package server

import (
	"math"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
)

type doorCircleCollisionHooks5488B0 struct {
	frame           func() uint32
	tickRate        func() uint32
	direction       func(byte) (int32, int32)
	addCollision    func(*Object, *Object, types.Pointf)
	wake            func(*Object)
	queueDoor       func(*DoorUpdateData)
	addToUpdatable  func(*Object)
	priorityMessage func(*Object, string)
}

func doorClosestPoint5488B0(start types.Pointf, directionX, directionY int32, point types.Pointf) types.Pointf {
	end := types.Ptf(start.X+float32(2*directionX), start.Y+float32(2*directionY))
	dx := float64(end.X - start.X)
	dy := float64(end.Y - start.Y)
	dot := dy*float64(point.Y-start.Y) + dx*float64(point.X-start.X)
	result := types.Ptf(
		float32(dot*dx/(32.0*32.0)+float64(start.X)),
		float32(dot*dy/(32.0*32.0)+float64(start.Y)),
	)
	result.X = max(min(result.X, max(start.X, end.X)), min(start.X, end.X))
	result.Y = max(min(result.Y, max(start.Y, end.Y)), min(start.Y, end.Y))
	return result
}

// doorCircleCollision5488B0 restores GAME.EXE 005488B0 while accessing the
// native-width Object.UpdateData pointer through DoorUpdateData. The original
// uses a fixed 32-unit door segment and the integer direction table.
func doorCircleCollision5488B0(door, circle *Object, moveCircle bool, hooks doorCircleCollisionHooks5488B0) {
	if door == nil || circle == nil || door.UpdateData == nil {
		return
	}
	update := door.UpdateDataDoor()
	direction := int(update.FractionalDir) + 160
	if direction >= 256 {
		direction -= 256
	}
	if direction < 0 {
		direction += 256
	}
	directionX, directionY := hooks.direction(byte(direction))
	closest := doorClosestPoint5488B0(door.NewPos, directionX, directionY, circle.NewPos)
	delta := circle.NewPos.Sub(closest)
	distance := math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y))
	if distance == 0 {
		distance = 0.1
	}
	if distance >= float64(circle.Shape.Circle.R) {
		return
	}
	normal := types.Ptf(float32(float64(delta.X)/distance), float32(float64(delta.Y)/distance))
	hooks.addCollision(circle, door, normal)
	update.LastMoveFrame = hooks.frame()

	penetration := float64(circle.Shape.Circle.R) - distance
	if moveCircle {
		velocityAgainstNormal := -float64(normal.Y)*float64(circle.VelVec.Y) - float64(normal.X)*float64(circle.VelVec.X)
		impulse := math.Sqrt(float64(circle.Mass)*100.0*4.0)*velocityAgainstNormal*0.25 + 100.0*penetration
		circle.Sub548600(types.Ptf(float32(impulse*float64(normal.X)), float32(impulse*float64(normal.Y))))
	}
	if circle.ObjFlags&0x08000000 != 0 {
		if circle.ObjFlags&0x8 == 0 {
			hooks.wake(circle)
		}
		circle.ObjFlags &^= 0x08000000
	}
	hooks.wake(door)

	if !door.HasTeam() || update.CurrentDirection != update.TargetDirection || door.TeamPtr().SameAs(circle.TeamPtr()) {
		if !moveCircle && update.LockCode == 0 && (door.ObjOwner == nil || door.ObjOwner == circle) {
			doorForce := penetration * float64(circle.Mass)
			dot := float64(directionX)*float64(circle.NewPos.Y-door.NewPos.Y) -
				float64(directionY)*float64(circle.NewPos.X-door.NewPos.X)
			if dot <= 0 {
				update.AngularVelocity += float32(doorForce)
			} else {
				update.AngularVelocity -= float32(doorForce)
			}
			hooks.queueDoor(update)
			hooks.addToUpdatable(door)
		}
		return
	}
	if hooks.frame() > door.Field34 {
		door.Field34 = hooks.frame() + hooks.tickRate()
		hooks.priorityMessage(circle, "objcoll.c:GateLockedMechanism")
	}
}

// DoorCircleCollision5488B0 binds the restored door/circle collision to the
// live collision queue. Queue ownership remains with the legacy collision
// system until that system is ported as a whole.
func (s *Server) DoorCircleCollision5488B0(door, circle *Object, moveCircle bool, addCollision func(*Object, *Object, types.Pointf), wake func(*Object), queueDoor func(*DoorUpdateData)) {
	doorCircleCollision5488B0(door, circle, moveCircle, doorCircleCollisionHooks5488B0{
		frame:    s.Frame,
		tickRate: s.TickRate,
		direction: func(direction byte) (int32, int32) {
			offset := uintptr(direction) * 8
			return memmap.Int32(0x587000, 192088+offset), memmap.Int32(0x587000, 192092+offset)
		},
		addCollision: addCollision,
		wake:         wake,
		queueDoor:    queueDoor,
		addToUpdatable: func(door *Object) {
			s.Objs.AddToUpdatable(door)
		},
		priorityMessage: func(circle *Object, message string) {
			s.NetPriMsgToPlayer(circle, strman.ID(message), 0)
		},
	})
}
