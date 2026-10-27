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

	// Startup PC53/ToZero retains both differences. FST stores Y without
	// popping it, and the product uses that stored Y as its memory operand.
	dx := objectForceSub53_52DF80(float64(target.PosVec.X), objectForceMemory32_52DF80(update.Position.X))
	dy := objectForceSub53_52DF80(float64(target.PosVec.Y), objectForceMemory32_52DF80(update.Position.Y))
	dyStored := experienceSpill32TowardZero(dy)
	squareY := objectForceMul53_52DF80(dy, objectForceMemory32_52DF80(dyStored))
	squareX := objectForceMul53_52DF80(dx, dx)
	distanceSquared := objectForceAdd53_52DF80(squareY, squareX)
	// The original C0|C3 gate keeps unordered distances as well as <=.
	if distanceSquared > phantomPlayerMaxDistanceSquared53B860 {
		if delayedDelete != nil {
			delayedDelete(source)
		}
		return
	}

	source.NewPos.X = experienceSpill32TowardZero(objectForceSub53_52DF80(float64(target.PosVec.X), dx))
	source.NewPos.Y = experienceSpill32TowardZero(objectForceSub53_52DF80(float64(update.Target.PosVec.Y), objectForceMemory32_52DF80(dyStored)))
	direction := uint16(update.Target.Direction1) + 128
	if int16(direction) >= 256 {
		direction &= 255
	}
	source.Direction2 = Dir16(direction)
}

// PhantomPlayerUpdate53B860 restores GAME.EXE 0053B860 without truncating the
// tracked object pointer to the PE32 width.
func (s *Server) PhantomPlayerUpdate53B860(source *Object, runtime PhantomPlayerUpdateRuntime53B860) {
	phantomPlayerUpdateNative53B860(source, runtime.DelayedDelete)
}
