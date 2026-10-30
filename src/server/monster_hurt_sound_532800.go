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
	if hooks.frame() < update.Field133 {
		return
	}
	// Original LEA bounds wrap as DWORDs before the signed C RNG call.
	tickRate := hooks.tickRate()
	delay := hooks.random(int(int32(2*tickRate)), int(int32(4*tickRate)))
	// Keep the entry update cached, then reload frame and the class-gated
	// sound set after RNG, which may replace the live update record.
	update.Field133 = hooks.frame() + uint32(delay)
	if uint8(unit.ObjClass)&uint8(object.ClassMonster) == 0 {
		return
	}
	soundSet := (*MonsterUpdateData)(unit.UpdateData).SoundSet122
	if soundSet != nil {
		id := sound.ID(*(*uint32)(unsafe.Add(soundSet, 2*4)))
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
