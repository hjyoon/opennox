package client

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

type AnimKind uint32

const (
	AnimOneShot       = AnimKind(0)
	AnimOneShotRemove = AnimKind(1)
	AnimLoop          = AnimKind(2)
	AnimLoopAndFade   = AnimKind(3)
	AnimRandom        = AnimKind(4)
	AnimSlave         = AnimKind(5)
)

func ParseAnimKind(str string) AnimKind {
	switch str {
	case "OneShot":
		return AnimOneShot
	case "OneShotRemove":
		return AnimOneShotRemove
	case "Loop":
		return AnimLoop
	case "LoopAndFade":
		return AnimLoopAndFade
	case "Random":
		return AnimRandom
	case "Slave":
		return AnimSlave
	}
	return AnimOneShot
}

type AnimationVector struct {
	Size   uint32                    // 0, 0
	Frames [9]*noxrender.ImageHandle // 1, 4; [9][]noxrender.ImageHandle
	Cnt40  uint16                    // 10, 40
	Val42  uint16                    // 10, 42
	Kind   AnimKind                  // 11, 44
}

// AnimationStateDrawData stores the three state-dependent vectors used by
// AnimateStateDraw. The original PE32 layout is 148 bytes; native 64-bit
// builds widen the pointer-bearing vectors instead of indexing them with the
// original 48-byte stride.
type AnimationStateDrawData struct {
	Size uint32
	Anim [3]AnimationVector
}

func (a *AnimationVector) FramesSlice(i int) []noxrender.ImageHandle {
	return unsafe.Slice(a.Frames[i], a.Cnt40)
}
