package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

const (
	telekinesisOwnerStopFlags53D330      = uint32(0x8020)
	telekinesisMaximumLifetimeSecs53D330 = uint32(20)
	telekinesisTraceFlags53D330          = MapTraceFlags(5)
)

// TelekinesisUpdateRuntime53D330 supplies the object operations that remain
// owned by the legacy-facing runtime. Object identities stay native-width.
type TelekinesisUpdateRuntime53D330 struct {
	Move          func(*Object, types.Pointf)
	DelayedDelete func(*Object)
	BuffOff       func(*Object, EnchantID)
}

type telekinesisUpdateDeps53D330 struct {
	frame         func() uint32
	fps           func() uint32
	spellOffSound func(spell.ID) sound.ID
	audioEvent    func(sound.ID, *Object)
	traceRay      func(types.Pointf, types.Pointf, MapTraceFlags) bool
	move          func(*Object, types.Pointf)
	delayedDelete func(*Object)
	buffOff       func(*Object, EnchantID)
}

func telekinesisUpdateNative53D330(source *Object, deps telekinesisUpdateDeps53D330) {
	lifetimeTelekinesis53D330(source, LifetimeEffectRuntime53CC90{
		Frame: deps.frame, FPS: deps.fps, SpellSound: func(id spell.ID, _ int) sound.ID { return deps.spellOffSound(id) }, Audio: deps.audioEvent, DelayedDelete: deps.delayedDelete, BuffOff: deps.buffOff,
		TraceRay: func(r *types.Rectf, f MapTraceFlags) bool { return deps.traceRay(r.Min, r.Max, f) },
		Move:     func(o *Object, p *types.Pointf) { deps.move(o, *p) },
	}, true)
}

func (s *Server) telekinesisUpdateDeps53D330(runtime TelekinesisUpdateRuntime53D330) telekinesisUpdateDeps53D330 {
	return telekinesisUpdateDeps53D330{
		frame: s.Frame,
		fps:   s.TickRate,
		spellOffSound: func(id spell.ID) sound.ID {
			return s.Spells.DefByInd(id).GetOffSound()
		},
		audioEvent: func(id sound.ID, obj *Object) {
			s.Audio.EventObj(id, obj, 0, 0)
		},
		traceRay:      s.MapTraceRay,
		move:          runtime.Move,
		delayedDelete: runtime.DelayedDelete,
		buffOff:       runtime.BuffOff,
	}
}

// TelekinesisUpdate53D330 restores GAME.EXE 0053D330 without reading the
// owner, player, or cursor through PE32 offsets.
func (s *Server) TelekinesisUpdate53D330(source *Object, runtime TelekinesisUpdateRuntime53D330) {
	if source == nil {
		return
	}
	telekinesisUpdateNative53D330(source, s.telekinesisUpdateDeps53D330(runtime))
}
