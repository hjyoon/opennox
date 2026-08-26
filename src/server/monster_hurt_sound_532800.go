package server

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
)

type monsterHurtSoundHooks532800 struct {
	frame    func() uint32
	tickRate func() uint32
	random   func(min, max int) int
	play     func(id sound.ID, unit *Object)
}

// monsterHurtSound532800 preserves GAME.EXE 00532800's hurt-sound cooldown
// while accessing MonsterUpdateData and SoundSet122 at native pointer width.
func monsterHurtSound532800(unit *Object, hooks monsterHurtSoundHooks532800) {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	frame := hooks.frame()
	if frame < update.Field133 {
		return
	}
	tickRate := int(hooks.tickRate())
	update.Field133 = frame + uint32(hooks.random(2*tickRate, 4*tickRate))
	if update.SoundSet122 != nil {
		id := sound.ID(*(*uint32)(unsafe.Add(update.SoundSet122, 2*4)))
		hooks.play(id, unit)
	}
}

func (s *Server) MonsterHurtSound532800(unit *Object) {
	monsterHurtSound532800(unit, monsterHurtSoundHooks532800{
		frame:    s.Frame,
		tickRate: s.TickRate,
		random:   s.Rand.Logic.IntClamp,
		play: func(id sound.ID, unit *Object) {
			s.Audio.EventObj(id, unit, 0, 0)
		},
	})
}
