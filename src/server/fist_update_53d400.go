package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	fistImpactAudio53D400         = uint32(48)
	fistImpactScorch53D400        = 2
	fistImpactPointFX53D400       = uint8(138)
	fistImpactEarthquake53D400    = 30
	fistDeleteHeight53D400        = float32(200)
	fistMaximumLifetimeSecs53D400 = uint32(3)
)

// FistUpdateRuntime53D400 supplies the world effects reached by the
// native-width restoration of GAME.EXE 0053D400.
type FistUpdateRuntime53D400 struct {
	AudioEvent    func(uint32, *Object)
	MakeScorch    func(types.Pointf, int)
	SendPointFX   func(uint8, types.Pointf)
	Earthquake    func(types.Pointf, int)
	DelayedDelete func(*Object)
}

type fistUpdateDeps53D400 struct {
	frame         func() uint32
	fps           func() uint32
	audioEvent    func(uint32, *Object)
	makeScorch    func(types.Pointf, int)
	sendPointFX   func(uint8, types.Pointf)
	earthquake    func(types.Pointf, int)
	delayedDelete func(*Object)
}

func fistUpdateNative53D400(source *Object, deps fistUpdateDeps53D400) {
	if source.ZVal <= 0 && int32(source.ObjFlags) >= 0 {
		deps.audioEvent(fistImpactAudio53D400, source)
		deps.makeScorch(source.PosVec, fistImpactScorch53D400)

		first := types.Ptf(
			source.PosVec.X+source.Shape.Circle.R,
			source.PosVec.Y,
		)
		source.ObjFlags |= object.FlagMarked
		deps.sendPointFX(fistImpactPointFX53D400, first)

		second := types.Ptf(
			source.PosVec.X-source.Shape.Circle.R,
			source.PosVec.Y,
		)
		deps.sendPointFX(fistImpactPointFX53D400, second)

		third := types.Ptf(
			source.PosVec.X,
			source.PosVec.Y+source.Shape.Circle.R,
		)
		deps.sendPointFX(fistImpactPointFX53D400, third)
		deps.earthquake(source.PosVec, fistImpactEarthquake53D400)
	}
	if source.ZVal >= fistDeleteHeight53D400 && int32(source.ObjFlags) < 0 {
		deps.delayedDelete(source)
	}
	if deps.frame()-source.Field32 > fistMaximumLifetimeSecs53D400*deps.fps() {
		deps.delayedDelete(source)
	}
}

// FistUpdate53D400 restores GAME.EXE 0053D400 without reading a native-width
// Object through PE32 field offsets.
func (s *Server) FistUpdate53D400(source *Object, runtime FistUpdateRuntime53D400) {
	if source == nil {
		return
	}
	fistUpdateNative53D400(source, fistUpdateDeps53D400{
		frame:         s.Frame,
		fps:           s.TickRate,
		audioEvent:    runtime.AudioEvent,
		makeScorch:    runtime.MakeScorch,
		sendPointFX:   runtime.SendPointFX,
		earthquake:    runtime.Earthquake,
		delayedDelete: runtime.DelayedDelete,
	})
}
