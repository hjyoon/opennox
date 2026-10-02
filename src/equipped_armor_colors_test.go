//go:build !server

package opennox

import "testing"

func TestEquippedArmorColorsMatchGAMEEXE(t *testing.T) {
	checkEquippedColorsMatchGAMEEXE(t, false)
}
