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
	return int(netFXTruncateSignedDWORD419A70(value))
}

func boxWallGrid5504B0(value float32) int32 {
	return int32(boxWallRound5504B0(experienceSpill32TowardZero(
		objectForceMul53_52DF80(float64(value), float64(boxWallGridScale5504B0)))))
}

func boxWallQuadrant550CB0(point, wall types.Pointf) byte {
	x := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(point.X), objectForceMemory32_52DF80(wall.X)))
	y := objectForceSub53_52DF80(float64(point.Y), objectForceMemory32_52DF80(wall.Y))
	// Original C0/ZF tests send unordered comparisons down the <= arms.
	if !(x > boxWallHalfStep5504B0) {
		if !(y > 0) {
			return 1
		}
		return 8
	}
	if !(y > 0) {
		return 2
	}
	return 4
}

// boxWallResponse550760 retains the wide spring/damping result until its
// component stores. The first wall rotates it before both equal stores;
// the second rotates the unspilled result separately for X and Y.
func boxWallResponse550760(obj *Object, velocity, penetration, spring, overlap float32, first bool, addCollision func(*Object, types.Pointf)) {
	mul, add, sub := objectForceMul53_52DF80, objectForceAdd53_52DF80, objectForceSub53_52DF80
	spill := experienceSpill32TowardZero
	k := float64(boxWallSqrtHalf5504B0)
	force := mul(mul(monsterMoveForceSqrtChop53_50D50C(mul(mul(float64(obj.Mass), float64(boxWallSpring5504B0)), 4)), -float64(velocity)), .5)
	force = mul(add(force, float64(spring)), float64(overlap))
	var impulse types.Pointf
	normal := types.Ptf(-boxWallSqrtHalf5504B0, boxWallSqrtHalf5504B0)
	if first {
		value := spill(mul(force, k))
		impulse = types.Ptf(value, value)
		normal.X = boxWallSqrtHalf5504B0
	} else {
		impulse = types.Ptf(spill(mul(force, -k)), spill(mul(force, k)))
	}
	if !(penetration >= 0) {
		normal.X, normal.Y = -normal.X, -normal.Y
	}
	vn := add(mul(float64(normal.X), float64(obj.VelVec.X)), mul(float64(normal.Y), float64(obj.VelVec.Y)))
	if !(vn >= 0) {
		obj.VelVec.X = spill(sub(float64(obj.VelVec.X), mul(vn, float64(normal.X))))
		obj.VelVec.Y = spill(sub(float64(obj.VelVec.Y), mul(vn, float64(normal.Y))))
	}
	tangent := types.Ptf(-normal.Y, normal.X)
	vt := spill(add(mul(float64(tangent.X), float64(obj.VelVec.X)), mul(float64(tangent.Y), float64(obj.VelVec.Y))))
	frictionX := mul(mul(mul(float64(obj.Mass), float64(vt)), float64(tangent.X)), float64(float32(.7)))
	impulse.X = spill(sub(float64(impulse.X), frictionX))
	frictionY := mul(mul(mul(float64(obj.Mass), float64(vt)), float64(tangent.Y)), float64(float32(.7)))
	impulse.Y = spill(sub(float64(impulse.Y), frictionY))
	obj.Sub548600(impulse)
	addCollision(obj, impulse)
}

func boxWallCollisionSecond550760(obj *Object, current, previous *types.Pointf, bounds *boxWallBounds5504B0, line types.Pointf, length float32, addCollision func(*Object, types.Pointf)) bool {
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	end := add(float64(length), memory(line.X))
	low := line.X
	if bounds.minX > line.X {
		low = bounds.minX
	}
	high := spill(end)
	if float64(bounds.maxX) < end || math.IsNaN(float64(bounds.maxX)) || math.IsNaN(end) {
		high = bounds.maxX
	}
	if float64(low) > end || !(high >= line.X) {
		return false
	}

	delta := sub(float64(current.Y), memory(line.Y))
	previousWide := sub(float64(previous.Y), memory(line.Y))
	previousDelta := spill(previousWide)
	if !(mul(previousWide, delta) >= 0) {
		if previousDelta >= 0 {
			current.Y = spill(add(float64(line.Y), 2))
		} else {
			current.Y = spill(sub(float64(line.Y), 2))
		}
		obj.NewPos.X = spill(mul(sub(float64(current.X), memory(current.Y)), float64(boxWallSqrtHalf5504B0)))
		obj.NewPos.Y = spill(mul(add(float64(current.X), memory(current.Y)), float64(boxWallSqrtHalf5504B0)))
		bounds.minY = spill(sub(float64(current.Y), mul(float64(obj.Shape.Box.H), .5)))
		bounds.maxY = spill(add(mul(float64(obj.Shape.Box.H), .5), memory(current.Y)))
		delta = sub(float64(current.Y), memory(line.Y))
	}
	if bounds.minY > line.Y || !(bounds.maxY >= line.Y) {
		return false
	}

	overlap := spill(div(sub(float64(high), memory(low)), sub(float64(bounds.maxX), memory(bounds.minX))))
	velocity := spill(mul(sub(float64(obj.VelVec.Y), memory(obj.VelVec.X)), float64(boxWallSqrtHalf5504B0)))
	var penetration float32
	if delta >= 0 {
		penetration = spill(sub(float64(line.Y), memory(bounds.minY)))
	} else {
		penetration = spill(sub(float64(bounds.maxY), memory(line.Y)))
	}
	spring := mul(float64(boxWallSpring5504B0), float64(penetration))
	if !(delta >= 0) {
		spring = -spring
	}
	boxWallResponse550760(obj, velocity, penetration, spill(spring), overlap, false, addCollision)
	return true
}

func boxWallCollisionFirst550A10(obj *Object, current, previous *types.Pointf, bounds *boxWallBounds5504B0, line types.Pointf, length float32, addCollision func(*Object, types.Pointf)) bool {
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	end := add(float64(length), memory(line.Y))
	low := line.Y
	if line.Y < bounds.minY || math.IsNaN(float64(line.Y)) || math.IsNaN(float64(bounds.minY)) {
		low = bounds.minY
	}
	high := spill(end)
	if float64(bounds.maxY) < end || math.IsNaN(float64(bounds.maxY)) || math.IsNaN(end) {
		high = bounds.maxY
	}
	if float64(low) > end || !(high >= line.Y) {
		return false
	}

	delta := sub(float64(current.X), memory(line.X))
	previousWide := sub(float64(previous.X), memory(line.X))
	previousDelta := spill(previousWide)
	if !(mul(previousWide, delta) >= 0) {
		if previousDelta >= 0 {
			current.X = spill(add(float64(line.X), 2))
		} else {
			current.X = spill(sub(float64(line.X), 2))
		}
		obj.NewPos.X = spill(mul(sub(float64(current.X), memory(current.Y)), float64(boxWallSqrtHalf5504B0)))
		obj.NewPos.Y = spill(mul(add(float64(current.Y), memory(current.X)), float64(boxWallSqrtHalf5504B0)))
		bounds.minX = spill(sub(float64(current.X), mul(float64(obj.Shape.Box.W), .5)))
		bounds.maxX = spill(add(mul(float64(obj.Shape.Box.W), .5), memory(current.X)))
		delta = sub(float64(current.X), memory(line.X))
	}
	if bounds.minX > line.X || !(bounds.maxX >= line.X) {
		return false
	}

	overlap := spill(div(sub(float64(high), memory(low)), sub(float64(bounds.maxY), memory(bounds.minY))))
	velocity := spill(mul(add(float64(obj.VelVec.X), memory(obj.VelVec.Y)), float64(boxWallSqrtHalf5504B0)))
	var penetration float32
	if delta >= 0 {
		penetration = spill(sub(float64(line.X), memory(bounds.minX)))
	} else {
		penetration = spill(sub(float64(bounds.maxX), memory(line.X)))
	}
	spring := mul(float64(boxWallSpring5504B0), float64(penetration))
	if !(delta >= 0) {
		spring = -spring
	}
	boxWallResponse550760(obj, velocity, penetration, spill(spring), overlap, true, addCollision)
	return true
}

func boxWallCollisionAt550580(grid image.Point, obj *Object, hooks boxWallCollisionHooks5504B0) bool {
	wallType := hooks.wallType(grid, 0x40)
	if wallType == -1 {
		return false
	}
	add, sub, mul := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	k := float64(boxWallSqrtHalf5504B0)
	// The original multiplies signed DWORD grid coordinates by 23 before FILD.
	wallX := float64(int32(grid.X) * int32(boxWallGridStep5504B0))
	wallY := float64(int32(grid.Y) * int32(boxWallGridStep5504B0))
	current := types.Ptf(
		spill(mul(add(float64(obj.NewPos.Y), memory(obj.NewPos.X)), k)),
		spill(mul(sub(float64(obj.NewPos.Y), memory(obj.NewPos.X)), k)),
	)
	previous := types.Ptf(
		spill(mul(add(float64(obj.PrevPos.Y), memory(obj.PrevPos.X)), k)),
		spill(mul(sub(float64(obj.PrevPos.Y), memory(obj.PrevPos.X)), k)),
	)
	halfW := mul(float64(obj.Shape.Box.W), .5)
	halfH := spill(mul(float64(obj.Shape.Box.H), .5))
	bounds := boxWallBounds5504B0{
		minX: spill(sub(float64(current.X), halfW)),
		minY: spill(sub(float64(current.Y), memory(halfH))),
		maxX: spill(add(halfW, memory(current.X))),
		maxY: spill(add(float64(halfH), memory(current.Y))),
	}
	wall := types.Ptf(spill(mul(add(wallY, wallX), k)), spill(mul(sub(wallY, wallX), k)))
	quadrant := boxWallQuadrant550CB0(current, wall)
	collided := false
	def := hooks.wallDef(byte(wallType))
	if def.firstEnabled != 0 && quadrant&def.firstMask == 0 {
		line := types.Ptf(
			spill(add(float64(wall.X), float64(boxWallHalfStep5504B0))),
			spill(add(sub(float64(wall.Y), float64(boxWallHalfStep5504B0)), memory(def.firstOffset))),
		)
		if boxWallCollisionFirst550A10(obj, &current, &previous, &bounds, line, def.firstLength, hooks.addCollision) {
			collided = true
		}
	}
	// Collision services may update the live wall definition before side two.
	def = hooks.wallDef(byte(wallType))
	if def.secondEnabled != 0 && quadrant&def.secondMask == 0 {
		line := types.Ptf(spill(add(float64(wall.X), memory(def.secondOffset))), wall.Y)
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
	minX := boxWallGrid5504B0(obj.CollideP1.X)
	minY := boxWallGrid5504B0(obj.CollideP1.Y)
	maxX := boxWallGrid5504B0(obj.CollideP2.X)
	maxY := boxWallGrid5504B0(obj.CollideP2.Y)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			grid := image.Pt(int(x), int(y))
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
