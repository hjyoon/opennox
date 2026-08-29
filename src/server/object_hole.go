package server

import "unsafe"

// HoleCollideData is GAME.EXE's fixed-width 28-byte Hole destination record.
// It contains only wire values; the owning Object and ScriptData pointers stay
// native-width in Object.
type HoleCollideData struct {
	Script             ScriptCallback // 0
	DestinationX       int32          // 8
	DestinationY       int32          // 12
	DestinationExtent  uint32         // 16
	DestinationNetCode uint16         // 20
	Reserved22         uint16         // 22
	Field24            uint32         // 24
}

var (
	_ = [1]struct{}{}[28-unsafe.Sizeof(HoleCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(HoleCollideData{}.Script)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(HoleCollideData{}.DestinationX)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(HoleCollideData{}.DestinationY)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(HoleCollideData{}.DestinationExtent)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(HoleCollideData{}.DestinationNetCode)]
	_ = [1]struct{}{}[22-unsafe.Offsetof(HoleCollideData{}.Reserved22)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(HoleCollideData{}.Field24)]
)
