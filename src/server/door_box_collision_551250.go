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

func doorBoxHorizontalIntersection551780(segment doorBoxSegment551250, edgeX1, edgeX2, edgeY float32, allowNegative bool) (types.Pointf, bool) {
	dy := segment.start.Y - segment.end.Y
	if dy == 0 {
		return types.Pointf{}, false
	}
	t := (edgeY - segment.end.Y) / dy
	if t < 0 && !allowNegative || t > 1 {
		return types.Pointf{}, false
	}
	dx := edgeX1 - edgeX2
	if dx == 0 {
		return types.Pointf{}, false
	}
	u := ((1-t)*segment.end.X + t*segment.start.X - edgeX2) / dx
	if u < 0 || u > 1 {
		return types.Pointf{}, false
	}
	return types.Ptf((1-u)*edgeX2+u*edgeX1, edgeY), true
}

func doorBoxVerticalIntersection551870(segment doorBoxSegment551250, edgeX, edgeY1, edgeY2 float32, allowNegative bool) (types.Pointf, bool) {
	dx := segment.start.X - segment.end.X
	if dx == 0 {
		return types.Pointf{}, false
	}
	t := (edgeX - segment.end.X) / dx
	if t < 0 && !allowNegative || t > 1 {
		return types.Pointf{}, false
	}
	dy := edgeY1 - edgeY2
	if dy == 0 {
		return types.Pointf{}, false
	}
	u := ((1-t)*segment.end.Y + t*segment.start.Y - edgeY2) / dy
	if u < 0 || u > 1 {
		return types.Pointf{}, false
	}
	return types.Ptf(edgeX, (1-u)*edgeY2+u*edgeY1), true
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
	if segmentBounds.minX < boxBounds.minX || segmentBounds.maxX > boxBounds.maxX ||
		segmentBounds.minY < boxBounds.minY || segmentBounds.maxY > boxBounds.maxY {
		points := doorBoxSegmentRectIntersections5516A0(segment, boxBounds, 2, false)
		switch len(points) {
		case 2:
			return types.Ptf((points[1].X+points[0].X)*0.5, (points[1].Y+points[0].Y)*0.5), true
		case 1:
			endpoint := segment.end
			if doorBoxPointInRect551A90(segment.start, boxBounds) {
				endpoint = segment.start
			}
			return types.Ptf((points[0].X+endpoint.X)*0.5, (points[0].Y+endpoint.Y)*0.5), true
		default:
			return types.Pointf{}, false
		}
	}
	return types.Ptf((segment.end.X+segment.start.X)*0.5, (segment.end.Y+segment.start.Y)*0.5), true
}

// doorBoxCollision551250 restores GAME.EXE 00551250 and its four geometric
// helpers while keeping both Object pointers native-width.
func doorBoxCollision551250(door, box *Object, moveBox bool, hooks doorBoxCollisionHooks551250) {
	if door == nil || box == nil || door.UpdateData == nil {
		return
	}
	update := door.UpdateDataDoor()
	doorStart := types.Ptf(
		(door.NewPos.Y+door.NewPos.X)*doorBoxSqrtHalf551250,
		(door.NewPos.Y-door.NewPos.X)*doorBoxSqrtHalf551250,
	)
	boxCenter := types.Ptf(
		(box.NewPos.Y+box.NewPos.X)*doorBoxSqrtHalf551250,
		(box.NewPos.Y-box.NewPos.X)*doorBoxSqrtHalf551250,
	)
	boxBounds := doorBoxRect551250{
		minX: boxCenter.X - box.Shape.Box.W*0.5,
		minY: boxCenter.Y - box.Shape.Box.H*0.5,
		maxX: boxCenter.X + box.Shape.Box.W*0.5,
		maxY: boxCenter.Y + box.Shape.Box.H*0.5,
	}
	direction := int(update.FractionalDir) + 128
	if direction >= 256 {
		direction = int(update.FractionalDir) - 128
	}
	directionVector := hooks.directionFloat(byte(direction))
	doorEnd := types.Ptf(doorStart.X+directionVector.X*32, doorStart.Y+directionVector.Y*32)
	segment := doorBoxSegment551250{start: doorStart, end: doorEnd}
	segmentBounds := doorBoxRect551250{
		minX: min(doorStart.X, doorEnd.X),
		minY: min(doorStart.Y, doorEnd.Y),
		maxX: max(doorStart.X, doorEnd.X),
		maxY: max(doorStart.Y, doorEnd.Y),
	}
	if segmentBounds.minX > boxBounds.maxX || segmentBounds.minY > boxBounds.maxY ||
		segmentBounds.maxX < boxBounds.minX || segmentBounds.maxY < boxBounds.minY {
		return
	}
	midpoint, ok := doorBoxCollisionMidpoint551960(segment, segmentBounds, boxBounds)
	if !ok {
		return
	}
	diagonalNormal := boxCenter.Sub(midpoint)
	distance := float32(math.Sqrt(float64(diagonalNormal.X*diagonalNormal.X + diagonalNormal.Y*diagonalNormal.Y)))
	if distance == 0 {
		return
	}
	diagonalNormal = diagonalNormal.Div(distance)
	centerRay := doorBoxSegment551250{start: boxCenter, end: midpoint}
	edgePoints := doorBoxSegmentRectIntersections5516A0(centerRay, boxBounds, 1, true)
	if len(edgePoints) != 1 {
		return
	}
	edgeDelta := edgePoints[0].Sub(boxCenter)
	edgeDistance := float32(math.Sqrt(float64(edgeDelta.X*edgeDelta.X + edgeDelta.Y*edgeDelta.Y)))
	if edgeDistance == 0 {
		return
	}
	normal := types.Ptf(
		(diagonalNormal.X-diagonalNormal.Y)*doorBoxSqrtHalf551250,
		(diagonalNormal.Y+diagonalNormal.X)*doorBoxSqrtHalf551250,
	)
	penetration := edgeDistance - distance
	if penetration <= 0 {
		return
	}
	hooks.addCollision(box, door, normal)
	update.LastMoveFrame = hooks.frame()
	if moveBox {
		velocityAgainstNormal := -normal.Y*box.VelVec.Y - normal.X*box.VelVec.X
		impulse := float32(math.Sqrt(float64(box.Mass*boxWallSpring5504B0*4)))*velocityAgainstNormal*0.25 + penetration*boxWallSpring5504B0
		box.Sub548600(types.Ptf(impulse*normal.X, impulse*normal.Y))
	}
	if box.ObjFlags&0x08000000 != 0 {
		if box.ObjFlags&0x8 == 0 {
			hooks.wake(box)
		}
		box.ObjFlags &^= 0x08000000
	}
	hooks.wake(door)

	if !door.HasTeam() || update.CurrentDirection != update.TargetDirection || door.TeamPtr().SameAs(box.TeamPtr()) {
		if !moveBox && update.LockCode == 0 && (door.ObjOwner == nil || door.ObjOwner == box) {
			torqueDirection := direction + 32
			if torqueDirection >= 256 {
				torqueDirection -= 256
			}
			directionX, directionY := hooks.directionInt(byte(torqueDirection))
			doorForce := penetration * box.Mass
			dot := float32(directionX)*(box.NewPos.Y-door.NewPos.Y) - float32(directionY)*(box.NewPos.X-door.NewPos.X)
			if dot <= 0 {
				update.AngularVelocity += doorForce
			} else {
				update.AngularVelocity -= doorForce
			}
			hooks.queueDoor(update)
			hooks.addToUpdatable(door)
		}
		return
	}
	if hooks.frame() > door.Field34 {
		door.Field34 = hooks.frame() + hooks.tickRate()
		hooks.priorityMessage(box, "objcoll.c:GateLockedMechanism")
	}
}

func (s *Server) DoorBoxCollision551250(door, box *Object, moveBox bool, addCollision func(*Object, *Object, types.Pointf), wake func(*Object), queueDoor func(*DoorUpdateData)) {
	doorBoxCollision551250(door, box, moveBox, doorBoxCollisionHooks551250{
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
