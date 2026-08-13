package server

import "math"

func (obj *Object) colliderFlagsByte4E7410() uint8 {
	return uint8(obj.ObjFlags)
}

func (obj *Object) colliderMinXBits4E7410() uint32 {
	return math.Float32bits(obj.CollideP1.X)
}

func (obj *Object) colliderMinYBits4E7410() uint32 {
	return math.Float32bits(obj.CollideP1.Y)
}

func (obj *Object) colliderMaxXBits4E7410() uint32 {
	return math.Float32bits(obj.CollideP2.X)
}

func (obj *Object) colliderMaxYBits4E7410() uint32 {
	return math.Float32bits(obj.CollideP2.Y)
}

func (obj *Object) Sub_4E7410() int {
	if objectColliderAllowed4E7410(obj) {
		return 1
	}
	return 0
}
