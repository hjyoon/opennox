package server

import (
	"math"

	"github.com/opennox/libs/types"
)

type CircleCollisionRuntime550D00 struct {
	PushScale       func() float32
	RandomDirection func(box bool) types.Pointf
	Trace           func(types.Pointf, types.Pointf) bool
	AddCollision    func(*Object, *Object, types.Pointf)
	Wake            func(*Object)
}

// CircleBoxGeometry54A990 is shared by movement and attack scans. Return
// values stay PC53/ToZero-wide; only the original binary32 stores spill.
func CircleBoxGeometry54A990(center types.Pointf, radius float32, box *Object, normal *types.Pointf, random func() types.Pointf) float64 {
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	k := float64(math.Float32frombits(0x3f350481))
	corner := func(x, y float32) types.Pointf {
		return types.Ptf(spill(add(float64(x), memory(box.NewPos.X))), spill(add(float64(y), memory(box.NewPos.Y))))
	}
	lt := corner(box.Shape.Box.LeftTop, box.Shape.Box.LeftBottom)
	lb := corner(box.Shape.Box.LeftBottom2, box.Shape.Box.LeftTop2)
	rt := corner(box.Shape.Box.RightTop, box.Shape.Box.RightBottom)
	rb := corner(box.Shape.Box.RightBottom2, box.Shape.Box.RightTop2)
	diagonal := func(p types.Pointf, sum bool) float32 {
		if sum {
			return spill(mul(sub(sub(add(float64(p.Y), memory(p.X)), memory(center.X)), memory(center.Y)), k))
		}
		return spill(mul(sub(add(sub(float64(p.Y), memory(p.X)), memory(center.X)), memory(center.Y)), k))
	}
	top, bottom, right, left := diagonal(lt, false), diagonal(lb, false), diagonal(rt, true), diagonal(lt, true)
	region := byte(0)
	if top > 0 {
		region = 1
	} else if bottom < 0 || math.IsNaN(float64(bottom)) {
		region = 2
	}
	if right < 0 || math.IsNaN(float64(right)) {
		region |= 8
	} else if left > 0 {
		region |= 4
	}
	switch region {
	case 0:
		delta := types.Ptf(spill(sub(float64(center.X), memory(box.NewPos.X))), spill(sub(float64(center.Y), memory(box.NewPos.Y))))
		if (delta.X == 0 || math.IsNaN(float64(delta.X))) && (delta.Y == 0 || math.IsNaN(float64(delta.Y))) {
			delta = random()
		}
		distance := collisionSqrt53_5488B0(add(mul(float64(delta.Y), memory(delta.Y)), mul(float64(delta.X), memory(delta.X))))
		if distance == 0 || math.IsNaN(distance) {
			distance = float64(math.Float32frombits(0x3dcccccd))
		}
		*normal = types.Ptf(spill(div(float64(delta.X), distance)), spill(div(float64(delta.Y), distance)))
		return distance
	case 1:
		*normal = types.Ptf(float32(k), -float32(k))
		return sub(float64(radius), memory(top))
	case 2:
		*normal = types.Ptf(-float32(k), float32(k))
		return sub(memory(radius), -float64(bottom))
	case 4:
		*normal = types.Ptf(-float32(k), -float32(k))
		return sub(float64(radius), memory(left))
	case 8:
		*normal = types.Ptf(float32(k), float32(k))
		return sub(memory(radius), -float64(right))
	case 5, 6, 9, 10:
		p := lt
		switch region {
		case 6:
			p = lb
		case 9:
			p = rt
		case 10:
			p = rb
		}
		dx := spill(sub(float64(center.X), memory(p.X)))
		dyWide := sub(float64(center.Y), memory(p.Y))
		dy := spill(dyWide)
		distanceWide := collisionSqrt53_5488B0(add(mul(dyWide, memory(dy)), mul(float64(dx), memory(dx))))
		distance := spill(distanceWide)
		if distanceWide == 0 || math.IsNaN(distanceWide) {
			distance = math.Float32frombits(0x3dcccccd)
		}
		result := sub(float64(radius), memory(distance))
		if result >= 0 {
			*normal = types.Ptf(spill(div(float64(dx), memory(distance))), spill(div(float64(dy), memory(distance))))
		}
		return result
	default:
		return -1
	}
}

func circleCollisionWake550D00(first, second *Object, wake func(*Object)) {
	if first.ObjFlags&0x08000000 != 0 {
		wake(first)
		first.ObjFlags &^= 0x08000000
	}
	if second.ObjFlags&0x08000000 != 0 {
		wake(second)
		second.ObjFlags &^= 0x08000000
	}
}

func circleCollisionMass550D00(first, second *Object) float64 {
	// The first call spills binary32; the comparison chooses the second on
	// <= or unordered and the selected object's mass is loaded again.
	firstMass := experienceSpill32TowardZero(float64(first.Mass))
	secondMass := float64(second.Mass)
	if !(secondMass > float64(firstMass)) {
		return float64(second.Mass)
	}
	return float64(first.Mass)
}

func CircleCircleCollision550D00(moving, candidate *Object, r CircleCollisionRuntime550D00) {
	add, sub, mul, div := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80, objectForceDiv53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	normal := types.Ptf(spill(sub(float64(candidate.NewPos.X), memory(moving.NewPos.X))), spill(sub(float64(candidate.NewPos.Y), memory(moving.NewPos.Y))))
	if (normal.X == 0 || math.IsNaN(float64(normal.X))) && (normal.Y == 0 || math.IsNaN(float64(normal.Y))) {
		normal = r.RandomDirection(false)
	}
	distanceWide := collisionSqrt53_5488B0(add(mul(float64(normal.Y), memory(normal.Y)), mul(float64(normal.X), memory(normal.X))))
	distance := spill(distanceWide)
	if distanceWide == 0 || math.IsNaN(distanceWide) {
		distance = math.Float32frombits(0x3c23d70a)
	}
	penetrationWide := sub(add(float64(moving.Shape.Circle.R), memory(candidate.Shape.Circle.R)), memory(distance))
	penetration := spill(penetrationWide)
	if !(penetrationWide > 0) {
		return
	}
	if moving.ObjClass&0x2204 != 0 && candidate.ObjClass&0x2204 != 0 && !r.Trace(moving.PosVec, candidate.PosVec) {
		return
	}
	r.AddCollision(candidate, moving, normal)
	if (moving.ObjClass&6 == 0 || candidate.ObjFlags&0x2000 == 0) && moving.ObjFlags&8 == 0 && candidate.ObjFlags&8 == 0 {
		nx, ny := spill(div(float64(normal.X), memory(distance))), spill(div(float64(normal.Y), memory(distance)))
		rx, ry := spill(sub(float64(moving.VelVec.X), memory(candidate.VelVec.X))), spill(sub(float64(moving.VelVec.Y), memory(candidate.VelVec.Y)))
		mass := circleCollisionMass550D00(moving, candidate)
		push := mul(float64(r.PushScale()), memory(penetration))
		force := types.Ptf(spill(-mul(push, memory(nx))), spill(-mul(push, memory(ny))))
		if moving.ObjClass&6 == 0 || candidate.ObjClass&6 == 0 {
			tx := -float64(ny)
			dotWide := add(mul(tx, memory(rx)), mul(float64(ry), memory(nx)))
			dot := spill(dotWide)
			friction := float64(float32(.7))
			force.X = spill(sub(memory(force.X), mul(mul(mul(dotWide, tx), mass), friction)))
			force.Y = spill(sub(memory(force.Y), mul(mul(mul(float64(dot), mass), memory(nx)), friction)))
		}
		moving.Sub548600(force)
	}
	circleCollisionWake550D00(moving, candidate, r.Wake)
}

func CircleBoxCollision54AD50(circle, box *Object, moveBox int32, r CircleCollisionRuntime550D00) {
	add, sub, mul := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	var normal types.Pointf
	penetrationWide := CircleBoxGeometry54A990(circle.NewPos, circle.Shape.Circle.R, box, &normal, func() types.Pointf { return r.RandomDirection(true) })
	penetration := spill(penetrationWide)
	if penetrationWide < 0 || math.IsNaN(penetrationWide) {
		return
	}
	if circle.ObjClass&0x2204 != 0 && box.ObjClass&0x2204 != 0 && !r.Trace(circle.PosVec, box.PosVec) {
		return
	}
	r.AddCollision(circle, box, normal)
	if (moveBox != 0 || circle.ObjClass&6 == 0 || box.ObjFlags&0x2000 == 0) && circle.ObjFlags&8 == 0 && box.ObjFlags&8 == 0 {
		push := mul(float64(r.PushScale()), memory(penetration))
		force := types.Ptf(spill(mul(float64(normal.X), push)), spill(mul(float64(normal.Y), push)))
		rx, ry := sub(float64(circle.VelVec.X), memory(box.VelVec.X)), sub(float64(circle.VelVec.Y), memory(box.VelVec.Y))
		tx, ty := -normal.Y, normal.X
		dot := spill(add(mul(float64(tx), rx), mul(ry, memory(ty))))
		tangent := mul(circleCollisionMass550D00(circle, box), memory(dot))
		friction := float64(float32(.7))
		force.Y = spill(sub(memory(force.Y), mul(mul(tangent, memory(ty)), friction)))
		force.X = spill(sub(memory(force.X), mul(mul(tangent, memory(tx)), friction)))
		if moveBox != 0 {
			box.Sub548600(types.Ptf(-force.X, -force.Y))
		} else {
			circle.Sub548600(force)
		}
	}
	circleCollisionWake550D00(circle, box, r.Wake)
}
