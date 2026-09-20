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
	owner := source.ObjOwner
	if owner == nil || owner.UpdateData == nil {
		deps.delayedDelete(source)
		return
	}
	update := (*PlayerUpdateData)(owner.UpdateData)
	if uint32(owner.ObjFlags)&telekinesisOwnerStopFlags53D330 != 0 ||
		deps.frame()-source.Field32 > telekinesisMaximumLifetimeSecs53D330*deps.fps() {
		snd := deps.spellOffSound(spell.SPELL_TELEKINESIS)
		deps.audioEvent(snd, owner)
		deps.delayedDelete(source)
		// GAME.EXE reloads ObjOwner after delayed deletion instead of reusing
		// the owner cached at entry.
		deps.buffOff(source.ObjOwner, ENCHANT_TELEKINESIS)
		return
	}

	player := update.Player
	if player == nil {
		deps.delayedDelete(source)
		return
	}
	point := types.Ptf(
		float32(int32(player.CursorVec.X)),
		float32(int32(player.CursorVec.Y)),
	)
	if deps.traceRay(owner.PosVec, point, telekinesisTraceFlags53D330) {
		deps.move(source, point)
	}
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
