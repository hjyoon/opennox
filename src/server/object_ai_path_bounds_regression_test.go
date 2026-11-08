package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestAIPathIndexObjectOutsideGridRegression(t *testing.T) {
	paths := &serverAIPaths{mapIndexGen: 77}
	obj := aiPathCircleObject50AFA0(object.ClassObstacle, 0, -1, 1)
	paths.IndexObject(obj)
	if cell := paths.MapIndex(0, 1); cell.IndexGen4 != 0 || cell.Flags8 != 0 {
		t.Fatalf("outside obstacle changed the border cell: %+v", cell)
	}
}
