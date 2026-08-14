package server

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// AudioEventCollideData is the pointer-independent four-byte record registered
// for AudioEventCollide in GAME.EXE.
type AudioEventCollideData struct {
	Sound uint32
}

type audioEventCollideNativeDeps4EAAD0 struct {
	frame func() uint32
	audio func(uint32, *Object)
}

func audioEventCollideNative4EAAD0(
	source, target *Object,
	collision *types.Pointf,
	deps audioEventCollideNativeDeps4EAAD0,
) {
	audioEventCollide4EAAD0(
		source,
		target,
		collision,
		audioEventCollideHooks4EAAD0[*Object, *AudioEventCollideData]{
			classLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			loadFrame: deps.frame,
			loadLastFrame: func(obj *Object) uint32 {
				return obj.Field34
			},
			storeFrame: func(obj *Object, frame uint32) {
				obj.Field34 = frame
			},
			loadCollideData: func(obj *Object) *AudioEventCollideData {
				return (*AudioEventCollideData)(obj.CollideData)
			},
			loadSound: func(data *AudioEventCollideData) uint32 {
				return data.Sound
			},
			audio: deps.audio,
		},
	)
}

func audioEventCollideServerDeps4EAAD0(s *Server) audioEventCollideNativeDeps4EAAD0 {
	return audioEventCollideNativeDeps4EAAD0{
		frame: s.Frame,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
	}
}

// AudioEventCollide4EAAD0 binds the original registered callback to native
// Object fields and the four-byte sound record.
func (s *Server) AudioEventCollide4EAAD0(
	source, target *Object,
	collision *types.Pointf,
) {
	audioEventCollideNative4EAAD0(
		source,
		target,
		collision,
		audioEventCollideServerDeps4EAAD0(s),
	)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(AudioEventCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(AudioEventCollideData{}.Sound)]
)
