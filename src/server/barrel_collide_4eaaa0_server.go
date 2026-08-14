package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type barrelCollideNativeDeps4EAAA0 struct {
	frame func() uint32
	audio func(uint32, *Object)
}

func barrelCollideNative4EAAA0(
	source, target *Object,
	collision *types.Pointf,
	deps barrelCollideNativeDeps4EAAA0,
) {
	barrelCollide4EAAA0(source, target, collision, barrelCollideHooks4EAAA0[*Object]{
		loadFrame: deps.frame,
		loadLastFrame: func(obj *Object) uint32 {
			return obj.Field34
		},
		storeFrame: func(obj *Object, frame uint32) {
			obj.Field34 = frame
		},
		audio: deps.audio,
	})
}

func barrelCollideServerDeps4EAAA0(s *Server) barrelCollideNativeDeps4EAAA0 {
	return barrelCollideNativeDeps4EAAA0{
		frame: s.Frame,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
	}
}

// BarrelCollide4EAAA0 binds the original BarrelCollide registration to the
// native Object timestamp while preserving the three-argument callback ABI.
func (s *Server) BarrelCollide4EAAA0(
	source, target *Object,
	collision *types.Pointf,
) {
	barrelCollideNative4EAAA0(source, target, collision, barrelCollideServerDeps4EAAA0(s))
}
