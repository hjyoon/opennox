package server

import (
	"math"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
)

const doorBoxSqrtHalf551250 = float32(0.70710677)

type doorBoxRect551250 struct {
	minX float32
	minY float32
	maxX float32
	maxY float32
}

type doorBoxSegment551250 struct {
	start types.Pointf
	end   types.Pointf
}

type doorBoxCollisionHooks551250 struct {
	spring          func() float32
	frame           func() uint32
	tickRate        func() uint32
	directionFloat  func(byte) types.Pointf
	directionInt    func(byte) (int32, int32)
	addCollision    func(*Object, *Object, types.Pointf)
	wake            func(*Object)
	queueDoor       func(*DoorUpdateData)
	addToUpdatable  func(*Object)
	priorityMessage func(*Object, string)
}

// Both original axis helpers retain their denominators and spill t and u.
func doorBoxAxisIntersection551780(segment doorBoxSegment551250, edge1, edge2, fixed float32, allowNegative, vertical bool) (types.Pointf, bool) {
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	startFixed, endFixed := segment.start.Y, segment.end.Y
	startOther, endOther := segment.start.X, segment.end.X
	if vertical {
		startFixed, endFixed = segment.start.X, segment.end.X
		startOther, endOther = segment.start.Y, segment.end.Y
	}
	denominator := sub(float64(startFixed), memory(endFixed))
	if denominator == 0 || math.IsNaN(denominator) {
		return types.Pointf{}, false
	}
	t := spill(div(sub(float64(fixed), memory(endFixed)), denominator))
	if (!(t >= 0) && !allowNegative) || t > 1 {
		return types.Pointf{}, false
	}
	denominator = sub(float64(edge1), memory(edge2))
	if denominator == 0 || math.IsNaN(denominator) {
		return types.Pointf{}, false
	}
	projected := add(mul(sub(1, memory(t)), memory(endOther)), mul(float64(t), memory(startOther)))
	u := spill(div(sub(projected, memory(edge2)), denominator))
	if !(u >= 0) || u > 1 {
		return types.Pointf{}, false
	}
	value := spill(add(mul(sub(1, memory(u)), memory(edge2)), mul(float64(u), memory(edge1))))
	if vertical {
		return types.Ptf(fixed, value), true
	}
	return types.Ptf(value, fixed), true
}

func doorBoxHorizontalIntersection551780(segment doorBoxSegment551250, edgeX1, edgeX2, edgeY float32, allowNegative bool) (types.Pointf, bool) {
	return doorBoxAxisIntersection551780(segment, edgeX1, edgeX2, edgeY, allowNegative, false)
}

func doorBoxVerticalIntersection551870(segment doorBoxSegment551250, edgeX, edgeY1, edgeY2 float32, allowNegative bool) (types.Pointf, bool) {
	return doorBoxAxisIntersection551780(segment, edgeY1, edgeY2, edgeX, allowNegative, true)
}

func doorBoxSegmentRectIntersections5516A0(segment doorBoxSegment551250, rect doorBoxRect551250, limit int, allowNegative bool) []types.Pointf {
	if limit <= 0 {
		return nil
	}
	out := make([]types.Pointf, 0, min(limit, 4))
	for edge := 0; edge < 4 && len(out) < limit; edge++ {
		var point types.Pointf
		var ok bool
		switch edge {
		case 0:
			point, ok = doorBoxHorizontalIntersection551780(segment, rect.minX, rect.maxX, rect.minY, allowNegative)
		case 1:
			point, ok = doorBoxVerticalIntersection551870(segment, rect.maxX, rect.minY, rect.maxY, allowNegative)
		case 2:
			point, ok = doorBoxHorizontalIntersection551780(segment, rect.minX, rect.maxX, rect.maxY, allowNegative)
		case 3:
			point, ok = doorBoxVerticalIntersection551870(segment, rect.minX, rect.minY, rect.maxY, allowNegative)
		}
		if ok {
			out = append(out, point)
		}
	}
	return out
}

func doorBoxPointInRect551A90(point types.Pointf, rect doorBoxRect551250) bool {
	return point.X >= rect.minX && point.X <= rect.maxX && point.Y >= rect.minY && point.Y <= rect.maxY
}

func doorBoxCollisionMidpoint551960(segment doorBoxSegment551250, segmentBounds, boxBounds doorBoxRect551250) (types.Pointf, bool) {
	add, mul := objectForceAdd53_52DF80, objectForceMul53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	average := func(a, b types.Pointf) types.Pointf {
		return types.Ptf(spill(mul(add(float64(a.X), memory(b.X)), .5)), spill(mul(add(float64(a.Y), memory(b.Y)), .5)))
	}
	if !(segmentBounds.minX >= boxBounds.minX) || segmentBounds.maxX > boxBounds.maxX ||
		!(segmentBounds.minY >= boxBounds.minY) || segmentBounds.maxY > boxBounds.maxY {
		points := doorBoxSegmentRectIntersections5516A0(segment, boxBounds, 2, false)
		switch len(points) {
		case 2:
			return average(points[1], points[0]), true
		case 1:
			endpoint := segment.end
			if doorBoxPointInRect551A90(segment.start, boxBounds) {
				endpoint = segment.start
			}
			return average(points[0], endpoint), true
		default:
			return types.Pointf{}, false
		}
	}
	return average(segment.end, segment.start), true
}

// doorBoxCollision551250 restores GAME.EXE 00551250 and its four geometric
// helpers while keeping both Object pointers native-width.
func doorBoxCollision551250(door, box *Object, moveBox int32, hooks doorBoxCollisionHooks551250) {
	if door == nil || box == nil || door.UpdateData == nil {
		return
	}
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	k := float64(doorBoxSqrtHalf551250)
	update := door.UpdateDataDoor()
	// 005512EF stores X without popping it before the endpoint calculation.
	startX := mul(add(float64(door.NewPos.Y), memory(door.NewPos.X)), k)
	startY := spill(mul(sub(float64(door.NewPos.Y), memory(door.NewPos.X)), k))
	doorStart := types.Ptf(spill(startX), startY)
	boxCenter := types.Ptf(
		spill(mul(add(float64(box.NewPos.Y), memory(box.NewPos.X)), k)),
		spill(mul(sub(float64(box.NewPos.Y), memory(box.NewPos.X)), k)),
	)
	halfW := mul(float64(box.Shape.Box.W), .5)
	halfH := spill(mul(float64(box.Shape.Box.H), .5))
	boxBounds := doorBoxRect551250{
		minX: spill(sub(float64(boxCenter.X), halfW)),
		minY: spill(sub(float64(boxCenter.Y), memory(halfH))),
		maxX: spill(add(halfW, memory(boxCenter.X))),
		maxY: spill(add(float64(halfH), memory(boxCenter.Y))),
	}
	direction := int(update.FractionalDir) + 128
	if direction >= 256 {
		direction -= 256
	}
	vector := hooks.directionFloat(byte(direction))
	doorEnd := types.Ptf(spill(add(mul(float64(vector.X), 32), startX)), spill(add(mul(float64(vector.Y), 32), memory(startY))))
	segment := doorBoxSegment551250{start: doorStart, end: doorEnd}
	segmentBounds := doorBoxRect551250{minX: doorEnd.X, maxX: doorStart.X, minY: doorEnd.Y, maxY: startY}
	if startX < float64(doorEnd.X) || math.IsNaN(startX) || math.IsNaN(float64(doorEnd.X)) {
		segmentBounds.minX, segmentBounds.maxX = doorStart.X, doorEnd.X
	}
	if startY < doorEnd.Y || math.IsNaN(float64(startY)) || math.IsNaN(float64(doorEnd.Y)) {
		segmentBounds.minY, segmentBounds.maxY = startY, doorEnd.Y
	}
	if segmentBounds.minX > boxBounds.maxX || segmentBounds.minY > boxBounds.maxY ||
		!(segmentBounds.maxX >= boxBounds.minX) || !(segmentBounds.maxY >= boxBounds.minY) {
		return
	}
	midpoint, ok := doorBoxCollisionMidpoint551960(segment, segmentBounds, boxBounds)
	if !ok {
		return
	}
	dx := sub(float64(boxCenter.X), memory(midpoint.X))
	dyWide := sub(float64(boxCenter.Y), memory(midpoint.Y))
	dy := spill(dyWide)
	distanceWide := collisionSqrt53_5488B0(add(mul(dyWide, float64(dy)), mul(dx, dx)))
	distance := spill(distanceWide)
	if distanceWide == 0 || math.IsNaN(distanceWide) {
		return
	}
	diagonalNormal := types.Ptf(spill(div(dx, float64(distance))), spill(div(float64(dy), float64(distance))))
	points := doorBoxSegmentRectIntersections5516A0(doorBoxSegment551250{start: boxCenter, end: midpoint}, boxBounds, 1, true)
	if len(points) != 1 {
		return
	}
	edgeX := sub(float64(points[0].X), memory(boxCenter.X))
	edgeY := sub(float64(points[0].Y), memory(boxCenter.Y))
	// Both edge differences and the root stay wide through subtraction.
	edgeDistance := collisionSqrt53_5488B0(add(mul(edgeY, edgeY), mul(edgeX, edgeX)))
	if edgeDistance == 0 || math.IsNaN(edgeDistance) {
		return
	}
	normal := types.Ptf(
		spill(mul(sub(float64(diagonalNormal.X), memory(diagonalNormal.Y)), k)),
		spill(mul(add(float64(diagonalNormal.Y), memory(diagonalNormal.X)), k)),
	)
	penetrationWide := sub(edgeDistance, float64(distance))
	penetration := spill(penetrationWide)
	if !(penetrationWide > 0) {
		return
	}
	hooks.addCollision(box, door, normal)
	update.LastMoveFrame = hooks.frame()
	if moveBox == 1 {
		applyDoorCollisionImpulse5489C7(box, normal, penetration, hooks.spring)
	}
	if box.ObjFlags&0x08000000 != 0 {
		if box.ObjFlags&8 == 0 {
			hooks.wake(box)
		}
		box.ObjFlags &^= 0x08000000
	}
	hooks.wake(door)
	if !door.HasTeam() || update.CurrentDirection != update.TargetDirection || door.TeamPtr().SameAs(box.TeamPtr()) {
		if moveBox == 0 && update.LockCode == 0 && (door.ObjOwner == nil || door.ObjOwner == box) {
			torqueDirection := direction + 32
			if torqueDirection >= 256 {
				torqueDirection -= 256
			}
			dirX, dirY := hooks.directionInt(byte(torqueDirection))
			dx := sub(float64(box.NewPos.X), memory(door.NewPos.X))
			dy := sub(float64(box.NewPos.Y), memory(door.NewPos.Y))
			dot := sub(mul(float64(dirX), dy), mul(float64(dirY), dx))
			force := mul(float64(penetration), memory(box.Mass))
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
		hooks.priorityMessage(box, "objcoll.c:GateLockedMechanism")
	}
}

func (s *Server) DoorBoxCollision551250(door, box *Object, moveBox int32, spring func() float32, addCollision func(*Object, *Object, types.Pointf), wake func(*Object), queueDoor func(*DoorUpdateData)) {
	doorBoxCollision551250(door, box, moveBox, doorBoxCollisionHooks551250{
		spring:   spring,
		frame:    s.Frame,
		tickRate: s.TickRate,
		directionFloat: func(direction byte) types.Pointf {
			offset := uintptr(direction) * 8
			return types.Ptf(memmap.Float32(0x587000, 194136+offset), memmap.Float32(0x587000, 194140+offset))
		},
		directionInt: func(direction byte) (int32, int32) {
			offset := uintptr(direction) * 8
			return memmap.Int32(0x587000, 192088+offset), memmap.Int32(0x587000, 192092+offset)
		},
		addCollision: addCollision,
		wake:         wake,
		queueDoor:    queueDoor,
		addToUpdatable: func(door *Object) {
			s.Objs.AddToUpdatable(door)
		},
		priorityMessage: func(box *Object, message string) {
			s.NetPriMsgToPlayer(box, strman.ID(message), 0)
		},
	})
}
