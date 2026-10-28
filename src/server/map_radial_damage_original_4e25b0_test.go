package server_test

import (
	"github.com/opennox/opennox/v1/internal/porttest"
	"testing"
)

func TestMapRadialDamageOriginalGo4E25B0(t *testing.T) {
	porttest.RunRadialOriginal4E25B0(t, "testdata/map_radial_damage_original_4e25b0.json", false, func(s *porttest.RadialScene4E25B0) { s.RunGo() })
}
