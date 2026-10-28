package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// EachMissileInCirclePtr518170 retains the original borrowed-origin API.
func (s *serverMap) EachMissileInCirclePtr518170(origin *types.Pointf, radius float32, fnc func(*Object) bool) {
	if fnc == nil {
		return
	}
	sx := mapIndexCoord5175E0(experienceSpill32TowardZero(objectForceSub53_52DF80(float64(origin.X), objectForceMemory32_52DF80(radius))))
	sy := mapIndexCoord5175E0(experienceSpill32TowardZero(objectForceSub53_52DF80(float64(origin.Y), objectForceMemory32_52DF80(radius))))
	ex := mapIndexCoord5175E0(experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(radius), objectForceMemory32_52DF80(origin.X))))
	ey := mapIndexCoord5175E0(experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(radius), objectForceMemory32_52DF80(origin.Y))))
	if sx < 0 {
		sx = 0
	}
	if ex >= s.indexSize {
		ex = s.indexSize - 1
	}
	if sy < 0 {
		sy = 0
	}
	if ey >= s.indexSize {
		ey = s.indexSize - 1
	}
	r2 := experienceSpill32TowardZero(objectForceMul53_52DF80(float64(radius), float64(radius)))
	for y := sy; y <= ey; y++ {
		for x := sx; x <= ex; x++ {
			for it := s.bucketByPos[x][y].List; it != nil; it = it.Next {
				obj := it.Object
				if uint8(obj.ObjClass)&1 == 0 {
					continue
				}
				// Bounds are cached, but the original reloads the borrowed origin and
				// the live node link after every callback. No binary32 delta spills occur.
				dx := objectForceSub53_52DF80(float64(origin.X), objectForceMemory32_52DF80(obj.PosVec.X))
				dy := objectForceSub53_52DF80(float64(origin.Y), objectForceMemory32_52DF80(obj.PosVec.Y))
				d2 := objectForceAdd53_52DF80(objectForceMul53_52DF80(dy, dy), objectForceMul53_52DF80(dx, dx))
				if !(d2 > float64(r2)) && !fnc(obj) {
					return
				}
			}
		}
	}
}

// The original 005175E0 multiplies by a binary32 inverse-cell constant,
// then spills (product+8388608) under PC53/ToZero. Registration and every
// native map query use this same conversion through RoundCoord.
func mapIndexCoord5175E0(value float32) int {
	if !(value >= 0) {
		if math.IsNaN(float64(value)) || value < -2147483648 {
			return math.MinInt32
		}
		return int(int32(value))
	}
	v := objectForceAdd53_52DF80(objectForceMul53_52DF80(float64(value), float64(math.Float32frombits(0x3c40c0c1))), 8388608)
	return int(math.Float32bits(experienceSpill32TowardZero(v)) & 0x7fffff)
}
