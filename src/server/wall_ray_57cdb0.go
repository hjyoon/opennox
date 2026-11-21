package server

import (
	"image"
	"math"
	"math/bits"

	"github.com/opennox/libs/types"
)

// WallRayTargetQuadrant57F1D0 classifies the ray end in its 23-unit tile.
func WallRayTargetQuadrant57F1D0(p types.Pointf) uint8 {
	x := uint8(wallRayTruncate419A70(p.X) % 23)
	y := uint8(wallRayTruncate419A70(p.Y) % 23)
	if x >= 12 {
		if y >= 12 {
			return 4
		}
		return 1
	}
	if y >= 12 {
		return 8
	}
	return 2
}

func wallRayTruncate419A70(v float32) int32 {
	if !(v >= -2147483648 && v < 2147483648) {
		return math.MinInt32
	}
	return int32(v)
}

// WallRayRegion57F2A0 classifies the ray origin relative to the hit tile.
// The C entry retains the high bytes of the original converter result.
func WallRayRegion57F2A0(p types.Pointf, tileX, tileY int32) int32 {
	x := wallRayTruncate419A70(pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(float64(p.X), -float64(23*tileX))))
	y := wallRayTruncate419A70(pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(float64(p.Y), -float64(23*tileY))))
	word := uint32(y) &^ 0xff
	if 22-x <= y {
		word |= 1
	}
	if x > y {
		return int32(word + 1)
	}
	return int32(((word - 1) &^ 2) + 3)
}

type wallNormalDirection uint8

const (
	wallNorthWest wallNormalDirection = iota
	wallNorthEast
	wallSouthEast
	wallSouthWest
)

const wallNormalDiagonal = 0.70709997

var wallNormalVectors = [...]types.Pointf{
	wallNorthWest: {X: -wallNormalDiagonal, Y: -wallNormalDiagonal},
	wallNorthEast: {X: wallNormalDiagonal, Y: -wallNormalDiagonal},
	wallSouthEast: {X: wallNormalDiagonal, Y: wallNormalDiagonal},
	wallSouthWest: {X: -wallNormalDiagonal, Y: wallNormalDiagonal},
}

// Indexed by wall shape, origin region and the set bit in the target quadrant.
var wallNormalDirections = [11][4][4]wallNormalDirection{
	0: {
		{wallNorthWest, wallNorthWest, wallNorthWest, wallNorthWest},
		{wallNorthWest, wallNorthWest, wallNorthWest, wallNorthWest},
		{wallSouthEast, wallSouthEast, wallSouthEast, wallSouthEast},
		{wallSouthEast, wallSouthEast, wallSouthEast, wallSouthEast},
	},
	1: {
		{wallSouthWest, wallSouthWest, wallSouthWest, wallSouthWest},
		{wallNorthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallNorthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthWest, wallSouthWest, wallSouthWest, wallSouthWest},
	},
	2: {
		{wallNorthWest, wallSouthWest, wallNorthWest, wallNorthWest},
		{wallNorthWest, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthEast, wallSouthEast, wallSouthWest, wallSouthEast},
	},
	3: {
		{wallNorthWest, wallSouthWest, wallNorthWest, wallNorthWest},
		{wallNorthWest, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthEast, wallSouthEast, wallSouthEast, wallSouthEast},
		{wallSouthEast, wallSouthEast, wallSouthEast, wallSouthEast},
	},
	4: {
		{wallSouthWest, wallSouthWest, wallSouthWest, wallSouthWest},
		{wallNorthWest, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthWest, wallSouthWest, wallSouthWest, wallSouthWest},
	},
	5: {
		{wallNorthWest, wallNorthWest, wallNorthWest, wallNorthWest},
		{wallNorthWest, wallNorthWest, wallNorthWest, wallNorthWest},
		{wallSouthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthEast, wallSouthEast, wallSouthWest, wallSouthEast},
	},
	6: {
		{wallNorthWest, wallSouthWest, wallNorthWest, wallNorthWest},
		{wallNorthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallNorthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthEast, wallSouthEast, wallSouthWest, wallSouthEast},
	},
	7: {
		{wallSouthEast, wallSouthWest, wallSouthWest, wallSouthWest},
		{wallNorthWest, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallSouthEast, wallSouthWest, wallSouthWest, wallSouthWest},
		{wallSouthEast, wallSouthWest, wallSouthWest, wallSouthWest},
	},
	8: {
		{wallNorthWest, wallSouthWest, wallSouthWest, wallSouthWest},
		{wallNorthWest, wallSouthWest, wallSouthWest, wallSouthWest},
		{wallSouthEast, wallNorthEast, wallNorthEast, wallNorthEast},
		{wallNorthWest, wallSouthWest, wallSouthWest, wallSouthWest},
	},
	9: {
		{wallNorthWest, wallNorthWest, wallNorthEast, wallNorthWest},
		{wallNorthWest, wallNorthWest, wallNorthEast, wallNorthWest},
		{wallNorthWest, wallNorthWest, wallNorthEast, wallNorthWest},
		{wallSouthEast, wallSouthEast, wallSouthWest, wallSouthEast},
	},
	10: {
		{wallNorthWest, wallSouthWest, wallNorthWest, wallNorthWest},
		{wallSouthEast, wallNorthEast, wallSouthEast, wallSouthEast},
		{wallSouthEast, wallNorthEast, wallSouthEast, wallSouthEast},
		{wallSouthEast, wallNorthEast, wallSouthEast, wallSouthEast},
	},
}

// ProjectileWallNormal finds the original diagonal normal of the struck wall.
func (s *Server) ProjectileWallNormal(grid image.Point, ray types.Rectf) (types.Pointf, bool) {
	x, y := int32(grid.X), int32(grid.Y)
	region := uint8(WallRayRegion57F2A0(ray.Min, x, y))
	quadrant := bits.TrailingZeros8(WallRayTargetQuadrant57F1D0(ray.Max))
	dir := uint8(s.Sub_57B500(image.Pt(int(x), int(y)), 0x40))
	if dir >= uint8(len(wallNormalDirections)) {
		return types.Pointf{}, false
	}
	return wallNormalVectors[wallNormalDirections[dir][region][quadrant]], true
}
