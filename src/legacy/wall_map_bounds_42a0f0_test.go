package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestWallMapBounds42A0F0(t *testing.T) {
	minX, maxX := uint32(256), uint32(0)
	minY, maxY := uint32(256), uint32(0)

	if got := wallMapBounds42A0F0(&server.Wall{X5: 91, Y6: 37}, &minX, &maxX, &minY, &maxY); got != 37 {
		t.Fatalf("return = %d, want 37", got)
	}
	if minX != 91 || maxX != 91 || minY != 37 || maxY != 37 {
		t.Fatalf("bounds = (%d..%d, %d..%d), want (91..91, 37..37)", minX, maxX, minY, maxY)
	}

	if got := wallMapBounds42A0F0(&server.Wall{X5: 120, Y6: 12}, &minX, &maxX, &minY, &maxY); got != 12 {
		t.Fatalf("return = %d, want 12", got)
	}
	if minX != 91 || maxX != 120 || minY != 12 || maxY != 37 {
		t.Fatalf("bounds = (%d..%d, %d..%d), want (91..120, 12..37)", minX, maxX, minY, maxY)
	}

	wallMapBounds42A0F0(&server.Wall{X5: 100, Y6: 20}, &minX, &maxX, &minY, &maxY)
	if minX != 91 || maxX != 120 || minY != 12 || maxY != 37 {
		t.Fatalf("interior wall changed bounds to (%d..%d, %d..%d)", minX, maxX, minY, maxY)
	}
}
