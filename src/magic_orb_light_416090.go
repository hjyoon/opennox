//go:build !server

package opennox

import "math"

// Both original magic draw routines FSTPS the retained 416090 result under
// gameplay's round-toward-zero control word before setting light intensity.
func magicDrawSpill32_4B98A0(value float64) float32 {
	stored := float32(value)
	if math.Abs(float64(stored)) > math.Abs(value) {
		stored = math.Nextafter32(stored, 0)
	}
	return stored
}
