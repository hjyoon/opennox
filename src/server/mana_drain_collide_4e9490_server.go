package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// ManaDrainCollideData is the pointer-independent eight-byte collide record
// registered for ManaDrainCollide. GAME.EXE 00536E50 writes Amount only and
// leaves the remaining seven bytes untouched.
type ManaDrainCollideData struct {
	Amount   uint8
	Reserved [7]uint8
}

type manaDrainNativeDeps4E9490 struct {
	loadFrame   func() uint32
	loadFPS     func() uint32
	godMode     func() bool
	protectMana func(uint32, int16)
	audio       func(*Object)
}

func manaDrainManaSubNative4EEBF0(
	unit *Object,
	amount uint8,
	deps manaDrainNativeDeps4E9490,
) {
	_ = playerManaSubNative4EEBF0(unit, int32(amount), playerManaSubNativeDeps4EEBF0{
		loadEngineGodMode: deps.godMode,
		protectMana: func(token uint32, delta int16) uintptr {
			deps.protectMana(token, delta)
			return 0
		},
	})
}

func manaDrainCollideNative4E9490(
	source, target *Object,
	collision unsafe.Pointer,
	deps manaDrainNativeDeps4E9490,
) {
	manaDrainCollide4E9490(source, target, collision, manaDrainCollideHooks4E9490[
		*Object,
		*PlayerUpdateData,
		*ManaDrainCollideData,
	]{
		classLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadUpdateData: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadManaCurrent: func(update *PlayerUpdateData) uint16 {
			return update.ManaCur
		},
		loadCollideData: func(obj *Object) *ManaDrainCollideData {
			return (*ManaDrainCollideData)(obj.CollideData)
		},
		loadAmount: func(data *ManaDrainCollideData) uint8 {
			return data.Amount
		},
		subtractMana: func(unit *Object, amount uint8) {
			manaDrainManaSubNative4EEBF0(unit, amount, deps)
		},
		loadSharedTimer: func(obj *Object) int16 {
			return int16(obj.Field542)
		},
		loadFrame: deps.loadFrame,
		loadFPS:   deps.loadFPS,
		audio:     deps.audio,
		storeSharedTimer: func(obj *Object, frame uint16) {
			obj.Field542 = frame
		},
	})
}

// ManaDrainCollide4E9490 binds the original callback to native pointer-width
// Object, PlayerUpdateData and Player layouts. protectMana crosses the legacy
// boundary with only the original fixed-width protection token and delta.
func (s *Server) ManaDrainCollide4E9490(
	source, target *Object,
	collision unsafe.Pointer,
	protectMana func(uint32, int16),
) {
	manaDrainCollideNative4E9490(source, target, collision, manaDrainNativeDeps4E9490{
		loadFrame: s.Frame,
		loadFPS:   s.TickRate,
		godMode: func() bool {
			return noxflags.HasEngine(noxflags.EngineGodMode)
		},
		protectMana: protectMana,
		audio: func(obj *Object) {
			s.Audio.EventObj(sound.ID(228), obj, 0, 0)
		},
	})
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(ManaDrainCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ManaDrainCollideData{}.Amount)]
)
