package server

import (
	"image"
	"math"

	"github.com/opennox/libs/types"
)

type CircleWallDefinition54FFC0 struct {
	First, Second                        byte
	FirstLo, FirstHi, SecondLo, SecondHi float32
}

type CircleWallRuntime54FEF0 struct {
	Spring       func() float32
	WallType     func(image.Point, byte) int8
	Definition   func(byte) CircleWallDefinition54FFC0
	EdgeMask     func(byte) byte
	WallObject   func(image.Point) *Wall
	IsGameBall   func(*Object) bool
	AddCollision func(*Object, types.Pointf)
	OpenSecret   func(image.Point, *Object)
}

func CircleWallNearest550280(second bool, origin types.Pointf, lo, hi float32, allowLo, allowHi int32, point types.Pointf, result *types.Pointf) bool {
	add, sub, mul := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	k := float64(math.Float32frombits(0x3f350481))
	x := sub(float64(point.X), memory(origin.X))
	y := sub(float64(point.Y), memory(origin.Y))
	value := mul(mul(add(x, y), k), k)
	if second {
		x = sub(memory(point.X), add(float64(origin.X), 23))
		value = sub(23, mul(sub(mul(y, k), mul(x, k)), k))
	}
	if !(value >= float64(lo)) {
		if allowLo == 0 {
			return false
		}
		value = float64(lo)
	}
	if value > float64(hi) {
		if allowHi == 0 {
			return false
		}
		value = float64(hi)
	}
	result.X = spill(add(value, memory(origin.X)))
	if second {
		result.Y = spill(add(sub(23, value), memory(origin.Y)))
	} else {
		result.Y = spill(add(value, memory(origin.Y)))
	}
	return true
}

func CircleWallResponse550380(obj *Object, point types.Pointf, r CircleWallRuntime54FEF0) bool {
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	dx := sub(float64(obj.NewPos.X), memory(point.X))
	dyWide := sub(float64(obj.NewPos.Y), memory(point.Y))
	dy := spill(dyWide)
	distanceWide := collisionSqrt53_5488B0(add(mul(dyWide, memory(dy)), mul(dx, dx)))
	distance := spill(distanceWide)
	if distanceWide == 0 || math.IsNaN(distanceWide) {
		distance = math.Float32frombits(0x3c23d70a)
	}
	if distance >= obj.Shape.Circle.R {
		return false
	}
	nx := spill(div(dx, memory(distance)))
	nyWide := div(float64(dy), memory(distance))
	ny := spill(nyWide)
	normal := types.Ptf(nx, ny)
	velocityWide := add(-mul(nyWide, memory(obj.VelVec.Y)), -mul(float64(nx), memory(obj.VelVec.X)))
	velocity := spill(velocityWide)
	if velocityWide > 0 && !r.IsGameBall(obj) {
		obj.VelVec.X = spill(add(mul(float64(velocity), memory(nx)), memory(obj.VelVec.X)))
		obj.VelVec.Y = spill(add(mul(float64(velocity), memory(ny)), memory(obj.VelVec.Y)))
	}
	force := mul(sub(float64(obj.Shape.Circle.R), memory(distance)), memory(r.Spring()))
	obj.Sub548600(types.Ptf(spill(mul(force, memory(nx))), spill(mul(force, memory(ny)))))
	r.AddCollision(obj, normal)
	return true
}

func CircleWallAt54FFC0(grid image.Point, obj *Object, r CircleWallRuntime54FEF0) bool {
	flags := byte(64)
	if obj.ObjFlags&0x4000 != 0 && obj.Shape.Kind == ShapeKindCircle && !(obj.Shape.Circle.R > 9) {
		flags = 0
	}
	typ := byte(r.WallType(grid, flags))
	if typ == 255 {
		return false
	}
	origin := types.Ptf(experienceSpill32TowardZero(float64(int32(uint32(grid.X)*23))), experienceSpill32TowardZero(float64(int32(uint32(grid.Y)*23))))
	first, second := int32(0), int32(0)
	collided := false
	neighbor := func(x, y int32) image.Point { return image.Pt(int(int32(grid.X)+x), int(int32(grid.Y)+y)) }
	rememberWall := func() {
		if obj.ObjClass&4 != 0 && obj.UpdateData != nil {
			update := obj.UpdateDataPlayer()
			update.CollisionWall = r.WallObject(grid)
		}
	}
	if r.Definition(typ).First != 0 {
		if r.EdgeMask(typ)&2 != 0 && r.WallType(neighbor(-1, -1), flags) == -1 {
			first = 1
		}
		if r.EdgeMask(typ)&4 != 0 && r.WallType(neighbor(1, 1), flags) == -1 {
			second = 1
		}
		if typ == 7 || typ == 10 {
			second = 1
		} else if typ == 8 || typ == 9 {
			first = 1
		}
		d := r.Definition(typ)
		var point types.Pointf
		if CircleWallNearest550280(false, origin, d.FirstLo, d.FirstHi, first, second, obj.NewPos, &point) {
			rememberWall()
			if CircleWallResponse550380(obj, point, r) {
				collided = true
			}
		}
	}
	// The second face reloads its live definition after the first response;
	// the endpoint allowances intentionally carry over from that face.
	if r.Definition(typ).Second != 0 {
		if r.EdgeMask(typ)&8 != 0 && r.WallType(neighbor(-1, 1), flags) == -1 {
			first = 1
		}
		if r.EdgeMask(typ)&1 != 0 && r.WallType(neighbor(1, -1), flags) == -1 {
			second = 1
		}
		if typ == 7 || typ == 8 {
			first = 1
		} else if typ == 9 || typ == 10 {
			second = 1
		}
		d := r.Definition(typ)
		var point types.Pointf
		if CircleWallNearest550280(true, origin, d.SecondLo, d.SecondHi, first, second, obj.NewPos, &point) {
			rememberWall()
			if CircleWallResponse550380(obj, point, r) {
				collided = true
			}
		}
	}
	return collided
}

func CircleWallCollision54FEF0(obj *Object, r CircleWallRuntime54FEF0) {
	if obj.ObjClass&0x400000 != 0 {
		return
	}
	minX, minY := boxWallGrid5504B0(obj.CollideP1.X), boxWallGrid5504B0(obj.CollideP1.Y)
	maxX, maxY := boxWallGrid5504B0(obj.CollideP2.X), boxWallGrid5504B0(obj.CollideP2.Y)
	for y := int64(minY); y <= int64(maxY); y++ {
		for x := int64(minX); x <= int64(maxX); x++ {
			grid := image.Pt(int(x), int(y))
			if CircleWallAt54FFC0(grid, obj, r) {
				r.OpenSecret(grid, obj)
			}
		}
	}
}
