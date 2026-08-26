package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestMapCountWalls51DEA0(t *testing.T) {
	minX, minY := uint32(256), uint32(256)
	mapCountWalls51DEA0(&server.Wall{X5: 91, Y6: 37}, &minX, &minY)
	if minX != 91 || minY != 37 {
		t.Fatalf("minimum = (%d, %d), want (91, 37)", minX, minY)
	}

	mapCountWalls51DEA0(&server.Wall{X5: 120, Y6: 12}, &minX, &minY)
	if minX != 91 || minY != 12 {
		t.Fatalf("minimum = (%d, %d), want (91, 12)", minX, minY)
	}

	mapCountWalls51DEA0(&server.Wall{X5: 91, Y6: 12}, &minX, &minY)
	if minX != 91 || minY != 12 {
		t.Fatalf("equal coordinates changed minimum to (%d, %d)", minX, minY)
	}
}
