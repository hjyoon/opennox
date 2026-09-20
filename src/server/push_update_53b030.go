package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// PushUpdateData53B030 is the fixed-width PushUpdate record parsed from
// thing.bin. The parser stores the first radius twice; GAME.EXE 0053B030 only
// consumes Radius and Force.
type PushUpdateData53B030 struct {
	Radius     float32
	RadiusCopy float32
	Force      float32
}

var (
	_ = [1]struct{}{}[12-unsafe.Sizeof(PushUpdateData53B030{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(PushUpdateData53B030{}.Radius)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(PushUpdateData53B030{}.RadiusCopy)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(PushUpdateData53B030{}.Force)]
)

// PushUpdateRuntime53B030 supplies the native radial-push implementation. The
// callback keeps object and update-data pointers in Go instead of reading the
// PE32 Object.UpdateData offset from C.
type PushUpdateRuntime53B030 struct {
	PushUnits func(types.Pointf, float32, float32, float32)
}

// PushUpdate53B030 restores GAME.EXE 0053B030. The original passes zero as the
// inner radius and no source/filter callbacks.
func PushUpdate53B030(source *Object, runtime PushUpdateRuntime53B030) bool {
	if source == nil || source.UpdateData == nil || runtime.PushUnits == nil {
		return false
	}
	update := (*PushUpdateData53B030)(source.UpdateData)
	runtime.PushUnits(source.PosVec, update.Radius, 0, update.Force)
	return true
}
