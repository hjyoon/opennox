package server

import "math"

func (obj *Object) colliderKind4E7290() uint32 {
	return uint32(obj.Shape.Kind)
}

func (obj *Object) colliderPosXBits4E7290() uint32 {
	return math.Float32bits(obj.PosVec.X)
}

func (obj *Object) colliderPosYBits4E7290() uint32 {
	return math.Float32bits(obj.PosVec.Y)
}

func (obj *Object) colliderNewPosXBits4E7350() uint32 {
	return math.Float32bits(obj.NewPos.X)
}

func (obj *Object) colliderNewPosYBits4E7350() uint32 {
	return math.Float32bits(obj.NewPos.Y)
}

func (obj *Object) colliderRadiusBits4E7290() uint32 {
	return math.Float32bits(obj.Shape.Circle.R)
}

func (obj *Object) colliderBoxMinXBits4E7290() uint32 {
	return math.Float32bits(obj.Shape.Box.LeftBottom2)
}

func (obj *Object) colliderBoxMinYBits4E7290() uint32 {
	return math.Float32bits(obj.Shape.Box.LeftBottom)
}

func (obj *Object) colliderBoxMaxXBits4E7290() uint32 {
	return math.Float32bits(obj.Shape.Box.RightTop)
}

func (obj *Object) colliderBoxMaxYBits4E7290() uint32 {
	return math.Float32bits(obj.Shape.Box.RightTop2)
}

func (obj *Object) colliderStoreMinXBits4E7290(v uint32) {
	obj.CollideP1.X = math.Float32frombits(v)
}

func (obj *Object) colliderStoreMinYBits4E7290(v uint32) {
	obj.CollideP1.Y = math.Float32frombits(v)
}

func (obj *Object) colliderStoreMaxXBits4E7290(v uint32) {
	obj.CollideP2.X = math.Float32frombits(v)
}

func (obj *Object) colliderStoreMaxYBits4E7290(v uint32) {
	obj.CollideP2.Y = math.Float32frombits(v)
}

func (obj *Object) Nox_xxx_objectUnkUpdateCoords_4E7290() *Object {
	return objectUpdateCollider4E7290(obj)
}

func (obj *Object) Sub_4E7350() *Object {
	return objectUpdateCollider4E7350(obj)
}
