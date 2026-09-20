package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

const phantomPlayerMaxDistanceSquared53B860 = float64(160000)

// PhantomPlayerUpdateData is the native-width form of the three-word PE32
// record consumed by GAME.EXE 0053B860.
type PhantomPlayerUpdateData struct {
	Target   *Object
	Position types.Pointf
}

func (obj *Object) UpdateDataPhantomPlayer() *PhantomPlayerUpdateData {
	return (*PhantomPlayerUpdateData)(obj.UpdateData)
}

var (
	_ = [1]struct{}{}[0-unsafe.Offsetof(PhantomPlayerUpdateData{}.Target)]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Offsetof(PhantomPlayerUpdateData{}.Position)]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))+8-unsafe.Sizeof(PhantomPlayerUpdateData{})]
)

// PhantomPlayerUpdateRuntime53B860 supplies the object lifecycle operation
// reached by the native-width restoration of GAME.EXE 0053B860.
type PhantomPlayerUpdateRuntime53B860 struct {
	DelayedDelete func(*Object)
}

func phantomPlayerUpdateNative53B860(source *Object, delayedDelete func(*Object)) {
	if source == nil {
		return
	}
	if source.UpdateData == nil {
		if delayedDelete != nil {
			delayedDelete(source)
		}
		return
	}

	update := source.UpdateDataPhantomPlayer()
	target := update.Target
	if target == nil {
		if delayedDelete != nil {
			delayedDelete(source)
		}
		return
	}

	// The original x87 code keeps the coordinate differences in extended
	// precision, except for the Y operand that it explicitly rounds to float.
	dx := float64(target.PosVec.X) - float64(update.Position.X)
	dy := float64(target.PosVec.Y) - float64(update.Position.Y)
	dyRounded := float64(float32(dy))
	distanceSquared := dx*dx + dy*dyRounded
	if !(distanceSquared <= phantomPlayerMaxDistanceSquared53B860) {
		if delayedDelete != nil {
			delayedDelete(source)
		}
		return
	}

	source.NewPos.X = float32(float64(target.PosVec.X) - dx)
	source.NewPos.Y = float32(float64(target.PosVec.Y) - dyRounded)
	source.Direction2 = Dir16((uint32(target.Direction1) + 128) & 0xff)
}

// PhantomPlayerUpdate53B860 restores GAME.EXE 0053B860 without truncating the
// tracked object pointer to the PE32 width.
func (s *Server) PhantomPlayerUpdate53B860(source *Object, runtime PhantomPlayerUpdateRuntime53B860) {
	phantomPlayerUpdateNative53B860(source, runtime.DelayedDelete)
}
