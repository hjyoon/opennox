package server

import "github.com/opennox/opennox/v1/internal/legacycrt"

// modifierFloatParameterValue4235C0 preserves the original decimal atof prefix
// and the binary32 store at 4235D3 under startup rounding toward zero.
// A missing token is -1; a present token without a decimal prefix is zero.
func modifierFloatParameterValue4235C0(word string) float32 {
	if word == "" {
		return -1
	}
	return legacycrt.AtofFloat32(word)
}
