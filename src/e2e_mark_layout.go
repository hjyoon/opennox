package opennox

import "github.com/opennox/libs/types"

// The arena has a verified forward 160-unit lane. Keep the idle NPC beyond
// the five player inputs (0..48), including both stock collision radii and
// the same conservative four-unit margin used by the ordinary lane check.
func e2eMarkNPCPosition(origin, direction types.Pointf, hostRadius, npcRadius float32) types.Pointf {
	distance := max(float32(120), 48+hostRadius+npcRadius+4+16)
	return origin.Add(direction.Mul(distance))
}
