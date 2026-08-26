package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterIdleSoundHooks5469B0 struct {
	frame    func() uint32
	tickRate func() uint32
	random   func(min, max int) int
	play     func(id sound.ID, unit *Object)
}

func monsterIdleSound5469B0(unit *Object, hooks monsterIdleSoundHooks5469B0) {
	if unit == nil || !unit.ObjClass.Has(object.ClassMonster) || unit.UpdateData == nil {
		return
	}
	update := unit.UpdateDataMonster()
	if unit.Nox_xxx_monsterCanAttackAtWill_534390() ||
		update.CurrentEnemy != nil ||
		math.Float32frombits(update.Field131) > 300.0 {
		return
	}
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_IDLE && head.Type() != ai.ACTION_GUARD {
		return
	}
	frame := hooks.frame()
	if frame < update.Field132 {
		return
	}
	tickRate := int(hooks.tickRate())
	update.Field132 = frame + uint32(hooks.random(20*tickRate, 60*tickRate))
	if update.SoundSet122 != nil {
		id := sound.ID(*(*uint32)(unsafe.Add(update.SoundSet122, 16)))
		hooks.play(id, unit)
	}
}

// MonsterIdleSound5469B0 keeps the original idle/guard sound schedule while
// reading MonsterUpdateData and the sound-set pointer at native width.
func (s *Server) MonsterIdleSound5469B0(unit *Object) {
	monsterIdleSound5469B0(unit, monsterIdleSoundHooks5469B0{
		frame:    s.Frame,
		tickRate: s.TickRate,
		random:   s.Rand.Logic.IntClamp,
		play: func(id sound.ID, unit *Object) {
			s.Audio.EventObj(id, unit, 0, 0)
		},
	})
}
