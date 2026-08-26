package server

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestBreakableWallRetainsNativeWallPointer(t *testing.T) {
	walls := serverWalls{}
	wl, freeWall := alloc.New(Wall{})
	defer freeWall()
	walls.AddBreakable(wl)
	defer walls.ClearBreakable()

	if got := walls.FirstBreakable().Wall; got != wl {
		t.Fatalf("breakable wall pointer = %p, want %p", got, wl)
	}
}
