package blobdata

import "math/bits"

// ApproxSqrt32 uses the original 0048C730 table to approximate a square root.
func ApproxSqrt32(value uint32) uint32 {
	// Normalize by powers of four so each pair of input bits corresponds to
	// one shift of the resulting root.
	power := (bits.Len32(value|1) - 1) / 2
	var index uint32
	if power < 3 {
		index = value << (6 - 2*power)
	} else {
		index = value >> (2*power - 6)
	}
	root := uint32(data587000[155956+index])
	if power < 7 {
		return root >> (7 - power)
	}
	return root << (power - 7)
}
