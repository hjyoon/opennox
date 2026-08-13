package server

import "unsafe"

// DoorTilePoint is the fixed-width tile coordinate passed to GAME.EXE
// 004E8340. It deliberately does not use image.Point, whose fields follow the
// host integer width.
type DoorTilePoint struct {
	X int32
	Y int32
}

var _ = [1]struct{}{}[8-unsafe.Sizeof(DoorTilePoint{})]
var _ = [1]struct{}{}[4-unsafe.Offsetof(DoorTilePoint{}.Y)]
