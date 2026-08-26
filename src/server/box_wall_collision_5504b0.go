package server

import (
	"image"
	"math"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	boxWallGridScale5504B0 = float32(0.043478262)
	boxWallGridStep5504B0  = float32(23.0)
	boxWallSqrtHalf5504B0  = float32(0.70710677)
	boxWallHalfStep5504B0  = float32(16.263456)
	boxWallSpring5504B0    = float32(100.0)
)

type boxWallDef5504B0 struct {
	firstEnabled  byte
	firstMask     byte
	firstOffset   float32
	firstLength   float32
	secondEnabled byte
	secondMask    byte
	secondOffset  float32
	secondLength  float32
}

type boxWallBounds5504B0 struct {
	minX float32
	minY float32
	maxX float32
	maxY float32
}

type boxWallCollisionHooks5504B0 struct {
	wallType     func(image.Point, byte) int8
	wallDef      func(byte) boxWallDef5504B0
	addCollision func(*Object, types.Pointf)
	openSecret   func(image.Point, *Object)
}

func boxWallRound5504B0(value float32) int {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int(math.RoundToEven(float64(value)))
}

func boxWallQuadrant550CB0(point, wall types.Pointf) byte {
	x := point.X - wall.X
	y := point.Y - wall.Y
	if x <= boxWallHalfStep5504B0 {
		if y <= 0 {
			return 1
		}
		return 8
	}
	if y <= 0 {
		return 2
	}
	return 4
}

func boxWallCollisionSecond550760(obj *Object, current, previous *types.Pointf, bounds *boxWallBounds5504B0, line types.Pointf, length float32, addCollision func(*Object, types.Pointf)) bool {
	low := max(bounds.minX, line.X)
	high := min(bounds.maxX, line.X+length)
	if low > line.X+length || high < line.X {
		return false
	}

	delta := current.Y - line.Y
	previousDelta := previous.Y - line.Y
	if previousDelta*delta < 0 {
		if previousDelta >= 0 {
			current.Y = line.Y + 2
		} else {
			current.Y = line.Y - 2
		}
		obj.NewPos.X = (current.X - current.Y) * boxWallSqrtHalf5504B0
		obj.NewPos.Y = (current.X + current.Y) * boxWallSqrtHalf5504B0
		bounds.minY = current.Y - obj.Shape.Box.H*0.5
		bounds.maxY = current.Y + obj.Shape.Box.H*0.5
		delta = current.Y - line.Y
	}
	if bounds.minY > line.Y || bounds.maxY < line.Y {
		return false
	}

	overlap := (high - low) / (bounds.maxX - bounds.minX)
	velocity := (obj.VelVec.Y - obj.VelVec.X) * boxWallSqrtHalf5504B0
	var penetration, spring float32
	if delta >= 0 {
		penetration = line.Y - bounds.minY
		spring = boxWallSpring5504B0 * penetration
	} else {
		penetration = bounds.maxY - line.Y
		spring = -boxWallSpring5504B0 * penetration
	}
	force := float32((math.Sqrt(float64(obj.Mass*boxWallSpring5504B0*4))*float64(-velocity)*0.5 + float64(spring)) * float64(overlap))
	impulse := types.Ptf(-force*boxWallSqrtHalf5504B0, force*boxWallSqrtHalf5504B0)

	var normal types.Pointf
	if penetration >= 0 {
		normal = types.Ptf(-boxWallSqrtHalf5504B0, boxWallSqrtHalf5504B0)
	} else {
		normal = types.Ptf(boxWallSqrtHalf5504B0, -boxWallSqrtHalf5504B0)
	}
	velocityNormal := normal.X*obj.VelVec.X + normal.Y*obj.VelVec.Y
	if velocityNormal < 0 {
		obj.VelVec.X -= velocityNormal * normal.X
		obj.VelVec.Y -= velocityNormal * normal.Y
	}
	tangent := types.Ptf(-normal.Y, normal.X)
	velocityTangent := tangent.X*obj.VelVec.X + tangent.Y*obj.VelVec.Y
	impulse.X -= obj.Mass * velocityTangent * tangent.X * 0.7
	impulse.Y -= obj.Mass * velocityTangent * tangent.Y * 0.7
	obj.Sub548600(impulse)
	addCollision(obj, impulse)
	return true
}

func boxWallCollisionFirst550A10(obj *Object, current, previous *types.Pointf, bounds *boxWallBounds5504B0, line types.Pointf, length float32, addCollision func(*Object, types.Pointf)) bool {
	low := max(line.Y, bounds.minY)
	high := min(bounds.maxY, line.Y+length)
	if low > line.Y+length || high < line.Y {
		return false
	}

	delta := current.X - line.X
	previousDelta := previous.X - line.X
	if previousDelta*delta < 0 {
		if previousDelta >= 0 {
			current.X = line.X + 2
		} else {
			current.X = line.X - 2
		}
		obj.NewPos.X = (current.X - current.Y) * boxWallSqrtHalf5504B0
		obj.NewPos.Y = (current.Y + current.X) * boxWallSqrtHalf5504B0
		bounds.minX = current.X - obj.Shape.Box.W*0.5
		bounds.maxX = current.X + obj.Shape.Box.W*0.5
		delta = current.X - line.X
	}
	if bounds.minX > line.X || bounds.maxX < line.X {
		return false
	}

	overlap := (high - low) / (bounds.maxY - bounds.minY)
	velocity := (obj.VelVec.X + obj.VelVec.Y) * boxWallSqrtHalf5504B0
	var penetration, spring float32
	if delta >= 0 {
		penetration = line.X - bounds.minX
		spring = boxWallSpring5504B0 * penetration
	} else {
		penetration = bounds.maxX - line.X
		spring = -boxWallSpring5504B0 * penetration
	}
	force := float32((math.Sqrt(float64(obj.Mass*boxWallSpring5504B0*4))*float64(-velocity)*0.5 + float64(spring)) * float64(overlap) * float64(boxWallSqrtHalf5504B0))
	impulse := types.Ptf(force, force)

	var normal types.Pointf
	if penetration >= 0 {
		normal = types.Ptf(boxWallSqrtHalf5504B0, boxWallSqrtHalf5504B0)
	} else {
		normal = types.Ptf(-boxWallSqrtHalf5504B0, -boxWallSqrtHalf5504B0)
	}
	velocityNormal := normal.X*obj.VelVec.X + normal.Y*obj.VelVec.Y
	if velocityNormal < 0 {
		obj.VelVec.X -= velocityNormal * normal.X
		obj.VelVec.Y -= velocityNormal * normal.Y
	}
	tangent := types.Ptf(-normal.Y, normal.X)
	velocityTangent := tangent.X*obj.VelVec.X + tangent.Y*obj.VelVec.Y
	impulse.X -= obj.Mass * velocityTangent * tangent.X * 0.7
	impulse.Y -= obj.Mass * velocityTangent * tangent.Y * 0.7
	obj.Sub548600(impulse)
	addCollision(obj, impulse)
	return true
}

func boxWallCollisionAt550580(grid image.Point, obj *Object, hooks boxWallCollisionHooks5504B0) bool {
	wallType := hooks.wallType(grid, 0x40)
	if wallType == -1 {
		return false
	}
	def := hooks.wallDef(byte(wallType))
	wallX := float32(grid.X) * boxWallGridStep5504B0
	wallY := float32(grid.Y) * boxWallGridStep5504B0
	current := types.Ptf(
		(obj.NewPos.Y+obj.NewPos.X)*boxWallSqrtHalf5504B0,
		(obj.NewPos.Y-obj.NewPos.X)*boxWallSqrtHalf5504B0,
	)
	previous := types.Ptf(
		(obj.PrevPos.Y+obj.PrevPos.X)*boxWallSqrtHalf5504B0,
		(obj.PrevPos.Y-obj.PrevPos.X)*boxWallSqrtHalf5504B0,
	)
	halfW := obj.Shape.Box.W * 0.5
	halfH := obj.Shape.Box.H * 0.5
	bounds := boxWallBounds5504B0{
		minX: current.X - halfW,
		minY: current.Y - halfH,
		maxX: current.X + halfW,
		maxY: current.Y + halfH,
	}
	wall := types.Ptf(
		(wallY+wallX)*boxWallSqrtHalf5504B0,
		(wallY-wallX)*boxWallSqrtHalf5504B0,
	)
	quadrant := boxWallQuadrant550CB0(current, wall)
	collided := false
	if def.firstEnabled != 0 && quadrant&def.firstMask == 0 {
		line := types.Ptf(wall.X+boxWallHalfStep5504B0, wall.Y-boxWallHalfStep5504B0+def.firstOffset)
		if boxWallCollisionFirst550A10(obj, &current, &previous, &bounds, line, def.firstLength, hooks.addCollision) {
			collided = true
		}
	}
	if def.secondEnabled != 0 && quadrant&def.secondMask == 0 {
		line := types.Ptf(wall.X+def.secondOffset, wall.Y)
		if boxWallCollisionSecond550760(obj, &current, &previous, &bounds, line, def.secondLength, hooks.addCollision) {
			collided = true
		}
	}
	return collided
}

// boxWallCollision5504B0 restores GAME.EXE 005504B0 and its box/wall
// narrow-phase routines without indexing Object through PE32 float slots.
func boxWallCollision5504B0(obj *Object, hooks boxWallCollisionHooks5504B0) {
	if obj == nil || obj.ObjClass&0x00400000 != 0 {
		return
	}
	minX := boxWallRound5504B0(obj.CollideP1.X * boxWallGridScale5504B0)
	minY := boxWallRound5504B0(obj.CollideP1.Y * boxWallGridScale5504B0)
	maxX := boxWallRound5504B0(obj.CollideP2.X * boxWallGridScale5504B0)
	maxY := boxWallRound5504B0(obj.CollideP2.Y * boxWallGridScale5504B0)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			grid := image.Pt(x, y)
			if boxWallCollisionAt550580(grid, obj, hooks) {
				hooks.openSecret(grid, obj)
			}
		}
	}
}

// BoxWallCollision5504B0 binds the restored box/wall collision to the live
// server wall lookup and the legacy collision queue callbacks.
func (s *Server) BoxWallCollision5504B0(obj *Object, addCollision func(*Object, types.Pointf), openSecret func(image.Point, *Object)) {
	boxWallCollision5504B0(obj, boxWallCollisionHooks5504B0{
		wallType: s.Sub_57B500,
		wallDef: func(wallType byte) boxWallDef5504B0 {
			offset := uintptr(wallType) * 24
			return boxWallDef5504B0{
				firstEnabled:  memmap.Uint8(0x587000, 292520+offset),
				firstMask:     memmap.Uint8(0x587000, 292521+offset),
				firstOffset:   memmap.Float32(0x587000, 292524+offset),
				firstLength:   memmap.Float32(0x587000, 292528+offset),
				secondEnabled: memmap.Uint8(0x587000, 292532+offset),
				secondMask:    memmap.Uint8(0x587000, 292533+offset),
				secondOffset:  memmap.Float32(0x587000, 292536+offset),
				secondLength:  memmap.Float32(0x587000, 292540+offset),
			}
		},
		addCollision: addCollision,
		openSecret:   openSecret,
	})
}
