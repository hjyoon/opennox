package server

import (
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"math"
)

type doorCircleCollisionHooks5488B0 struct {
	spring          func() float32
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
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	end := types.Ptf(spill(add(float64(2*directionX), memory(start.X))), spill(add(float64(2*directionY), memory(start.Y))))
	px := sub(float64(point.X), memory(start.X))
	py := sub(float64(point.Y), memory(start.Y))
	dx := spill(sub(float64(end.X), memory(start.X)))
	dyWide := sub(float64(end.Y), memory(start.Y))
	dy := spill(dyWide)
	dot := add(mul(dyWide, py), mul(float64(dx), px))
	// 0057C7DA spills X before adding the origin; Y remains wide.
	x := spill(div(mul(dot, float64(dx)), 32*32))
	y := div(mul(dot, float64(dy)), 32*32)
	result := types.Ptf(spill(add(float64(x), memory(start.X))), spill(add(y, memory(start.Y))))
	minX, maxX := end.X, start.X
	if start.X < end.X || math.IsNaN(float64(start.X)) || math.IsNaN(float64(end.X)) {
		minX, maxX = start.X, end.X
	}
	minY, maxY := end.Y, start.Y
	if start.Y < end.Y || math.IsNaN(float64(start.Y)) || math.IsNaN(float64(end.Y)) {
		minY, maxY = start.Y, end.Y
	}
	if !(result.X >= minX) {
		result.X = minX
	} else if result.X > maxX {
		result.X = maxX
	}
	if minY > result.Y {
		result.Y = minY
	} else if result.Y > maxY {
		result.Y = maxY
	}
	return result
}

// Preserve the original retained differences and startup PC53/ToZero stores.
func doorCircleCollision5488B0(door, circle *Object, moveCircle int32, hooks doorCircleCollisionHooks5488B0) {
	if door == nil || circle == nil || door.UpdateData == nil {
		return
	}
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	update := door.UpdateDataDoor()
	direction := int(update.FractionalDir) + 160
	if direction >= 256 {
		direction -= 256
	}
	directionX, directionY := hooks.direction(byte(direction))
	closest := doorClosestPoint5488B0(door.NewPos, directionX, directionY, circle.NewPos)
	dx := sub(float64(circle.NewPos.X), memory(closest.X))
	dyWide := sub(float64(circle.NewPos.Y), memory(closest.Y))
	dy := spill(dyWide)
	distanceWide := monsterMoveForceSqrtChop53_50D50C(add(mul(dyWide, float64(dy)), mul(dx, dx)))
	distance := spill(distanceWide)
	if distanceWide == 0 || math.IsNaN(distanceWide) {
		distance = math.Float32frombits(0x3dcccccd)
	}
	if distance >= circle.Shape.Circle.R {
		return
	}
	normal := types.Ptf(spill(div(dx, float64(distance))), spill(div(float64(dy), float64(distance))))
	hooks.addCollision(circle, door, normal)
	update.LastMoveFrame = hooks.frame()
	// Exactly one applies linear force; zero selects angular force.
	if moveCircle == 1 {
		penetration := spill(sub(float64(circle.Shape.Circle.R), memory(distance)))
		applyDoorCollisionImpulse5489C7(circle, normal, penetration, hooks.spring)
	}
	if circle.ObjFlags&0x08000000 != 0 {
		if circle.ObjFlags&8 == 0 {
			hooks.wake(circle)
		}
		circle.ObjFlags &^= 0x08000000
	}
	hooks.wake(door)
	if !door.HasTeam() || update.CurrentDirection != update.TargetDirection || door.TeamPtr().SameAs(circle.TeamPtr()) {
		if moveCircle == 0 && update.LockCode == 0 && (door.ObjOwner == nil || door.ObjOwner == circle) {
			// Reload live radius, mass, positions and the cached table addresses
			// after the wakeup services, while retaining the cached update pointer.
			penetration := spill(sub(float64(circle.Shape.Circle.R), memory(distance)))
			directionX, directionY = hooks.direction(byte(direction))
			dx := sub(float64(circle.NewPos.X), memory(door.NewPos.X))
			dy := sub(float64(circle.NewPos.Y), memory(door.NewPos.Y))
			dot := sub(mul(float64(directionX), dy), mul(float64(directionY), dx))
			force := mul(float64(penetration), memory(circle.Mass))
			if !(dot > 0) {
				update.AngularVelocity = spill(add(force, memory(update.AngularVelocity)))
			} else {
				update.AngularVelocity = spill(sub(memory(update.AngularVelocity), force))
			}
			hooks.queueDoor(update)
			hooks.addToUpdatable(door)
		}
		return
	}
	frame := hooks.frame()
	if frame > door.Field34 {
		door.Field34 = frame + hooks.tickRate()
		hooks.priorityMessage(circle, "objcoll.c:GateLockedMechanism")
	}
}

// Direction tables remain mapped data; the spring belongs to live legacy C.
func (s *Server) DoorCircleCollision5488B0(door, circle *Object, moveCircle int32, spring func() float32, addCollision func(*Object, *Object, types.Pointf), wake func(*Object), queueDoor func(*DoorUpdateData)) {
	doorCircleCollision5488B0(door, circle, moveCircle, doorCircleCollisionHooks5488B0{
		spring: spring, frame: s.Frame, tickRate: s.TickRate,
		direction: func(direction byte) (int32, int32) {
			offset := uintptr(direction) * 8
			return memmap.Int32(0x587000, 192088+offset), memmap.Int32(0x587000, 192092+offset)
		},
		addCollision: addCollision, wake: wake, queueDoor: queueDoor,
		addToUpdatable:  func(door *Object) { s.Objs.AddToUpdatable(door) },
		priorityMessage: func(circle *Object, message string) { s.NetPriMsgToPlayer(circle, strman.ID(message), 0) },
	})
}
