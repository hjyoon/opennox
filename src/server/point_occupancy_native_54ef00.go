package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
)

// PointOccupiedCandidateNative54EF60 has the original void callback contract.
// A rejected candidate leaves the shared DWORD untouched.
func (obj *Object) PointOccupiedCandidateNative54EF60(point *types.Pointf, store func(uint32)) {
	if uint32(obj.ObjClass)&0x20000 != 0 || obj.Field5&0x800 != 0 {
		return
	}
	if obj.PointInShapeNative547DB0(point) {
		store(1)
	}
}

// PointOccupiedNative54EF00 resets the shared scratch before reading point.
// The iterator retains the same point pointer and completes after a hit;
// its final scratch DWORD, including service mutations, is returned unchanged.
func PointOccupiedNative54EF00(point *types.Pointf,
	each func(types.Rectf, func(*Object) bool), load func() uint32, store func(uint32),
) uint32 {
	store(0)
	gap := float64(monsterGeneratorCollisionGap54EF00)
	minX := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(point.X), gap))
	minY := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(point.Y), gap))
	maxX := experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(point.X), gap))
	maxY := experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(point.Y), gap))
	rect := types.Rectf{Min: types.Ptf(minX, minY), Max: types.Ptf(maxX, maxY)}
	each(rect, func(obj *Object) bool {
		obj.PointOccupiedCandidateNative54EF60(point, store)
		return true
	})
	return load()
}

func (s *Server) PointOccupied54EF00(point *types.Pointf) bool {
	return PointOccupiedNative54EF00(point, s.Map.EachObjInRect,
		func() uint32 { return memmap.Uint32(0x5D4594, 2491708) },
		func(value uint32) { *memmap.PtrUint32(0x5D4594, 2491708) = value },
	) != 0
}
